---
name: embed-a-code-block-raster
description: >-
  Use when you need to show source code in a deck. PowerPoint cannot preserve
  monospace metrics or whitespace, so a code listing is supplied as a
  pre-rendered image (a raster) referenced by AssetID, and the engine places it
  with an optional native language badge and caption. Reach for it whenever a
  slide should display a code snippet: render the code to an image yourself, wire
  an AssetResolver, and drop a scene.CodeBlock node.
---

# Embed a code-block raster

## Overview

A code listing is **not** drawn as native PowerPoint text. PowerPoint cannot
faithfully reproduce monospace metrics, tab stops, or exact whitespace, and it
has no concept of syntax highlighting. So pptx-go renders a code block as a
**caller-rasterized picture** (a `pic` built from your bytes, D-014) — exactly
the same per-node policy as `Image` and `Chart`.

The engine **never rasterizes and never highlights** (D-026). It does not parse
your source, does not pick a color scheme, and does not lay out glyphs. You
render the listing to an image — with a headless highlighter, a terminal
screenshotter, an HTML-to-PNG step — and hand the bytes to the renderer through
an `AssetResolver`. pptx-go embeds the bytes verbatim and draws a thin layer of
**native** chrome around them: an optional language badge and an optional
caption (D-045). The division is the whole point: the engine renders the deck,
the caller produces the pixels.

## The CodeBlock node

`scene.CodeBlock` carries an asset reference plus two optional native overlays:

```go
type CodeBlock struct {
    AssetID  scene.AssetID // required — the rendered code image
    Language string        // optional — drawn as a badge pill, top-right
    Caption  string        // optional — drawn as a centered caption, below
}
```

A minimal node:

```go
scene.CodeBlock{
    AssetID:  "snippet",
    Language: "go",
    Caption:  "main.go",
}
```

**Validation (Stage 1):** `AssetID` must be non-empty. An empty AssetID fails
the render with `code_block requires an asset id`. `Language` and `Caption` are
free-form and may be empty.

## Wiring the asset (AssetResolver + WithAssetResolver)

The node references its image by `AssetID`; the bytes come from an
`AssetResolver` you wire into `Render`:

```go
type AssetResolver interface {
    // Resolve returns the asset's bytes and a content-type hint
    // ("image/png", "image/jpeg", "image/svg+xml", …). A miss returns
    // (nil, "", scene.ErrAssetNotFound).
    Resolve(ctx context.Context, id scene.AssetID) ([]byte, string, error)
}
```

A minimal map-backed resolver, wired with `scene.WithAssetResolver`:

```go
type mapResolver struct{ assets map[scene.AssetID][]byte }

func (r mapResolver) Resolve(_ context.Context, id scene.AssetID) ([]byte, string, error) {
    b, ok := r.assets[id]
    if !ok {
        return nil, "", scene.ErrAssetNotFound
    }
    return b, "image/png", nil
}

stats, err := scene.Render(pres, s, scene.WithAssetResolver(resolver))
```

The resolver is the single seam for code bytes. Keep it deterministic for a
given AssetID so re-renders stay byte-identical (Render is idempotent). See the
`register-an-asset` skill for the resolver contract in full.

## Language badge & caption (native overlays — D-045)

The badge and caption are drawn natively — they are real PowerPoint shapes and
text, editable in PowerPoint, not baked into your raster:

- **Language badge.** When `Language != ""` *and* the image resolved, the engine
  draws a small rounded pill in the image's **top-right corner**, inset slightly
  from the edge, filled with the theme's `ColorSurfaceAlt` token and rounded to
  `RadiusFull`. The language string is centered inside it in the `TypeCaption`
  type role with the `TextSecondary` color. The pill is drawn **after** the
  picture, so shape-tree order puts it on top (z-order). An empty `Language`
  emits no badge. A badge is **never** drawn over a missing image.
- **Caption.** When `Caption != ""`, the image area is shrunk by a fixed caption
  height (0.4") and the caption is drawn as a centered line **below** the image,
  in the `TypeCaption` role with the `TextMuted` color. The caption renders
  whether or not the image resolved (it is positioned relative to the reserved
  image box).

All chrome flows through theme tokens (P2) — a theme swap restyles the badge and
caption without touching this code.

## Unresolved behavior (missing asset → LayoutWarning + node skipped — D-036)

A missing code image is **non-fatal**. If the resolver returns
`scene.ErrAssetNotFound` (or any error), the render still succeeds: the engine
skips the picture, records a `LayoutWarning` in `Stats.Warnings`, and moves on.
There is no strict mode in V1 — a caller that wants a miss to be fatal inspects
`Stats.Warnings` itself (D-036).

```go
stats, err := scene.Render(pres, s, scene.WithAssetResolver(resolver))
// err is nil even when "snippet" did not resolve.
for _, w := range stats.Warnings {
    // w.Message ~ `code_block asset "snippet" unresolved: scene: asset not found`
}
```

`stats.Assets` counts pictures **actually embedded**, so an unresolved
CodeBlock leaves `Assets` unincremented and adds one warning. Because the badge
is drawn only over a rendered image, a missing asset yields no badge either —
but the caption (if set) still renders.

## Complete example

A runnable, self-contained program lives at
`examples/embed-a-code-block-raster/main.go`. It implements a map-backed
resolver, generates a real PNG at runtime with stdlib `image`/`image/png`,
renders a `CodeBlock{AssetID:"snippet", Language:"go", Caption:"main.go"}`,
prints `stats.Assets` and `stats.Warnings`, and serializes the deck with
`WriteToBytes`. Run it:

```bash
go run ./examples/embed-a-code-block-raster
```

Expected output: 1 slide, 1 asset embedded, 0 warnings, and a written `.pptx`.

## What the caller owns

You own **turning code into an image**. pptx-go gives you the embedding
mechanism and the native chrome, not the renderer:

- **Highlight** the source with whatever you like — Chroma, Pygments, a
  language server, a hand-rolled tokenizer.
- **Render** the highlighted listing to a bitmap — a headless browser
  (HTML/CSS → PNG), a terminal screenshotter, an SVG-to-PNG step, or a direct
  raster.
- **Supply** those bytes through the resolver, keyed by the `AssetID` your
  `CodeBlock` references, with a correct content-type hint.

Keep rendering deterministic for a given AssetID so re-renders stay
byte-identical.

## See also

- `register-an-asset` — the AssetResolver contract and warn-don't-fail behavior,
  shared by every asset-bearing node.
- `compose-a-scene` — building the Scene and slides the CodeBlock lives in.
- `embed-a-chart-raster` — the sibling pattern for a `Chart` node (also a
  caller-rendered raster).
