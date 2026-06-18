---
name: embed-a-chart-raster
description: >-
  Use to put a chart in a deck by supplying a pre-rendered image. In V1 a chart
  is an image-shape (native c:chart is V2 — D-004): the caller rasterizes the
  chart to a PNG/SVG and the engine positions it with a caption and a
  contains-to-fit aspect warning. Covers both the scene Chart node (resolved
  through an AssetResolver) and the pptx-level ChartPlaceholder + AddImage path.
  Reach for it whenever you have chart bytes (or a labeled empty slot) to place.
---

# Embed a chart raster

## Overview

In V1, **a chart is an image** (D-004). pptx-go ships no native `c:chart`
renderer — that is V2. Equally, pptx-go **never rasterizes** (D-026): turning
your data into a chart picture is the caller's job. You render the chart to image
bytes (PNG/JPEG/SVG); the engine's only job is to position those bytes, add a
caption, and warn when the image does not fit its slot cleanly.

There are two ways to get a chart picture into a deck:

- **Scene path** — a `scene.Chart` node whose bytes come from an `AssetResolver`
  at render time. Use this when you are building a deck from the scene IR.
- **Builder path** — `(*Slide).ChartPlaceholder` for a labeled empty slot, and
  `(*Slide).AddImage` for a rendered chart. Use this when you are placing shapes
  imperatively on a `pptx.Presentation`.

## Scene path (Chart node + AssetResolver)

The `Chart` node carries the chart by reference, not by value:

```go
type Chart struct {
    AssetID scene.AssetID // the id your resolver maps to bytes
    Caption string        // optional caption rendered below the chart
}
```

The bytes are supplied through an `AssetResolver`, wired in with
`scene.WithAssetResolver`:

```go
type AssetResolver interface {
    Resolve(ctx context.Context, id scene.AssetID) ([]byte, string, error)
}
```

`Resolve` returns `(bytes, contentTypeHint, error)` — the content-type hint is
something like `"image/png"`, `"image/jpeg"`, or `"image/svg+xml"`. A miss
returns `scene.ErrAssetNotFound`.

When the asset resolves, the renderer:

1. reads the image's pixel dimensions from its **format header only**
   (`image.DecodeConfig` — not pixel data, §7/D-046);
2. **contains-to-fit** the image into its slot — the largest box at the image's
   aspect ratio that fits, centered (aspect preserved, never stretched);
3. emits a `LayoutWarning` when the image's aspect ratio diverges from the slot
   by more than ~15% (so a 16:9 chart dropped into a tall slot warns you it will
   sit letterboxed);
4. draws the caption (if any) as a muted, centered caption-role text frame below
   the chart.

See `compose-a-scene` for building the `Scene` and slides these nodes live in,
and `register-an-asset` for the resolver contract in full.

## Builder path (ChartPlaceholder + AddImage)

On a raw `pptx.Presentation`, two builder calls cover the chart slot:

```go
// A labeled empty slot — a rounded rect with a dashed accent border and a
// centered "Chart" label. It commits no bytes; it is the visible stand-in for a
// chart whose raster is unresolved or not yet rendered. Fills/border resolve
// against the active theme (P2); pass ShapeOptions to override.
func (s *Slide) ChartPlaceholder(box pptx.Box, opts ...pptx.ShapeOption) *pptx.Shape

// A rendered chart picture from caller bytes. ImageBytes verifies the declared
// MIME against the bytes; the box is the picture's position and size.
func (s *Slide) AddImage(src pptx.ImageSource, box pptx.Box) (*pptx.Image, error)
```

Supply the bytes with `pptx.ImageBytes(data, "image/png")` (or
`pptx.ImageFile(path)`):

```go
sl := pres.AddSlide()
sl.ChartPlaceholder(pptx.Box{X: pptx.In(0.5), Y: pptx.In(1.5), W: pptx.In(5.5), H: pptx.In(4)})
sl.AddImage(
    pptx.ImageBytes(chartPNG, "image/png"),
    pptx.Box{X: pptx.In(6.5), Y: pptx.In(1.5), W: pptx.In(6), H: pptx.In(3.375)},
)
```

Note the builder path does **not** auto-fit: the picture fills the box you give
it. Pick a box at the chart's aspect ratio yourself (the 16:9 box above keeps a
960x540 chart undistorted). The contains-to-fit behavior is the scene renderer's,
not the builder's.

## Unresolved chart behavior

A missing `Chart` asset is **not** an error and **not** a blank gap. When
`Resolve` returns `scene.ErrAssetNotFound` (or any error), the renderer:

- draws a `pptx.ChartPlaceholder` (the labeled dashed slot) in the chart's
  slot — the same visible stand-in the builder path exposes; and
- records a `LayoutWarning` in `Stats.Warnings` naming the slide and the
  unresolved id.

The render still succeeds — V1 degrades every asset-resolution failure to a
warning (D-036); there is no strict mode. The caption, if any, still renders
below the placeholder. Inspect `Stats.Warnings` yourself if you want a miss to be
fatal in your pipeline (that is a caller policy, D-026).

The aspect-divergence warning is independent: a chart that **does** resolve but
whose aspect ratio is far from its slot also produces a `LayoutWarning` (the
image is still placed, contained-to-fit).

## Complete example

`examples/embed-a-chart-raster/main.go` runs both paths in one program. It
implements a map-backed resolver returning a runtime-generated PNG for
`"revenue-chart"`, renders a two-slide scene (one resolved `Chart`, one
unresolved), prints `Stats`, then on a raw deck draws a `ChartPlaceholder` and
an `AddImage` chart picture.

```bash
go run ./examples/embed-a-chart-raster
```

Observed output (abridged):

```
[scene path]
  slides rendered: 2
  charts embedded: 1
  layout warnings: 2
    - slide-chart: chart "revenue-chart" aspect ratio diverges from its slot by ~167%; fit within the slot
    - slide-unresolved: chart asset "forecast-chart" unresolved: scene: asset not found
[builder path]
  OK — wrote ...embed-a-chart-raster-builder.pptx (... bytes: 1 placeholder slot + 1 chart picture)
```

The first warning is the aspect-divergence notice for the resolved 16:9 chart in
a narrow slot (it is still embedded, contained-to-fit); the second is the
unresolved-asset placeholder fallback.

## What the caller owns

- **The chart image bytes.** pptx-go never rasterizes (D-026). Render your chart
  with whatever you like — a Go charting library, a headless-browser screenshot,
  a cached blob — and hand the engine the resulting PNG/JPEG/SVG bytes.
- **The slot's aspect ratio on the builder path.** `AddImage` fills the box you
  pass; size it to the chart's aspect ratio to avoid distortion.
- **Whether a miss is fatal.** Inspect `Stats.Warnings` — the engine warns, it
  does not fail.

## See also

- `register-an-asset` — the AssetResolver contract in full (Image, Chart,
  CodeBlock, asset-kind Decoration).
- `compose-a-scene` — building the Scene and slides the Chart node lives in.
