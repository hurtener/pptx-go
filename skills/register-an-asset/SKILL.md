---
name: register-an-asset
description: >-
  Use when a scene node needs caller-supplied bytes — an Image, Chart,
  CodeBlock, or asset-kind Decoration that references content by AssetID. Teaches
  how to implement an AssetResolver, wire it with scene.WithAssetResolver, and
  understand the warn-don't-fail behavior when an asset is missing. Reach for it
  whenever you need to provide image/chart/code bytes that scene nodes reference
  by AssetID.
---

# Register an asset

## Overview

pptx-go renders each scene node under a **per-node policy** that is intrinsic to
the node type, not a per-deck switch (D-011, D-018). A node renders one of two
ways:

- **Native PPTX shapes** — most nodes (Hero, Prose, List, Table, Flow, Card,
  preset Decoration, …). pptx-go draws these from theme tokens.
- **A `pic` shape built from YOUR bytes** — every node whose IR struct carries
  an `AssetID` field: `Image`, `Chart`, `CodeBlock`, and an asset-kind
  `Decoration`.

The engine **never rasterizes** (D-026). It does not render a chart, syntax-
highlight a code block, or fetch a URL. You pre-rasterize the content to image
bytes and hand them to the renderer through an `AssetResolver`. pptx-go embeds
the bytes verbatim. That division — engine renders the deck, caller produces the
pixels — is the whole point of this seam.

## The AssetResolver interface

A resolver maps an `AssetID` to bytes plus a content-type hint:

```go
type AssetID string

type AssetResolver interface {
    // Resolve returns the asset's bytes and a content-type hint
    // ("image/png", "image/jpeg", "image/svg+xml", …). A missing asset
    // returns (nil, "", scene.ErrAssetNotFound).
    Resolve(ctx context.Context, id AssetID) ([]byte, string, error)
}
```

A minimal map-backed resolver:

```go
type mapResolver struct{ assets map[scene.AssetID][]byte }

func (r mapResolver) Resolve(_ context.Context, id scene.AssetID) ([]byte, string, error) {
    b, ok := r.assets[id]
    if !ok {
        return nil, "", scene.ErrAssetNotFound
    }
    return b, "image/png", nil
}
```

The `ctx` is the render context (`scene.WithContext`, default
`context.Background()`); honor cancellation if your resolver does real I/O.

## AssetID conventions

`AssetID` is **free-form** — pptx-go imposes no scheme. Use bare keys
(`"logo"`), content hashes, UUIDs, paths — whatever your store keys on.

For the common `asset://<uuid>` convention there is a helper:

```go
// URIAssetResolver strips an "asset://" prefix and passes the bare uuid to fn.
// A non-asset:// id is passed through unchanged.
resolver := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
    return blobStore.Get(uuid) // returns (bytes, "image/png", error)
})
```

## Wiring it

Pass the resolver to `Render` with the `WithAssetResolver` option:

```go
stats, err := scene.Render(pres, myScene, scene.WithAssetResolver(resolver))
```

With **no** resolver registered, every asset-bearing node is treated as
unresolved (the renderer returns `ErrAssetNotFound` internally) — see failure
behavior below.

## Failure behavior

A missing asset is **non-fatal** — this is the engine's "warn, don't fail"
contract (D-036). When `Resolve` returns an error (including
`ErrAssetNotFound`):

- `Render` still returns `(Stats, nil)` — **no error**.
- The node is **skipped** (not drawn).
- A `LayoutWarning` is appended to `Stats.Warnings`, naming the slide and the
  unresolved AssetID.
- An injected logger (`scene.WithLogger`) emits a `Warn` event per warning.

There is **no strict mode in V1** (D-036). If you want a miss to be fatal,
inspect `Stats.Warnings` yourself after `Render` and fail at the call site.

One node type degrades visibly rather than silently: an **unresolved `Chart`**
draws a labeled `ChartPlaceholder` (a placeholder rect + label) in its slot
instead of leaving a blank gap, and still records the warning.

Verified empirically (see the example): a two-slide scene with one resolved and
one missing `Image` renders with `Slides=2`, `Assets=1`, and one warning —
`image asset "missing-banner" unresolved: scene: asset not found`.

## Content-type hints

The second return value is a MIME hint that flows into `pptx.ImageBytes`. Return
the type that matches your bytes: `"image/png"`, `"image/jpeg"`,
`"image/gif"`, or `"image/svg+xml"`. pptx-go verifies the bytes match the
declared type and rejects obviously malformed images (e.g. a missing PNG
signature), but it does **not** parse pixel data — a malicious image is the
caller's problem at display time (§7). Chart and image nodes read dimensions
from the format header only (`image.DecodeConfig`) to fit the slot.

## Complete example

A runnable, self-contained program lives at
`examples/register-an-asset/main.go`. It implements a map-backed resolver,
generates a real PNG at runtime with stdlib `image`/`image/png`, renders an
`Image` that resolves and an `Image` that does not, and prints
`stats.Assets` and `stats.Warnings` to show the warn-don't-fail behavior. Run
it:

```bash
go run ./examples/register-an-asset
```

## What the caller owns

You own **turning content into bytes**. pptx-go gives you the embedding
mechanism, not the rasterizer:

- **Image** — provide the encoded image (PNG/JPEG/GIF/SVG) you want placed.
- **Chart** — render your chart to an image (a charting lib, a headless
  browser, a server-side renderer) and supply those bytes; native `c:chart`
  is V2 (D-004).
- **CodeBlock** — syntax-highlight and rasterize the listing to an image
  (D-014), then supply the bytes.
- **Decoration** (`Kind: DecorationAsset`) — supply the ornament image.

The resolver is the single seam. Keep it deterministic for a given AssetID so
re-renders stay byte-identical (Render is idempotent).

## See also

- `compose-a-scene` — building the Scene and slides these nodes live in.
- `embed-a-chart-raster` — producing Chart bytes for a `Chart` node.
- `embed-a-code-block-raster` — producing CodeBlock bytes for a `CodeBlock`
  node.
