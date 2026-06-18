# Assets, icons & rasters

Some scene nodes render as an image rather than as native shapes. Those nodes
carry an `AssetID` and resolve their bytes through an `AssetResolver` at render
time. pptx-go never rasterizes for you — you pre-rasterize and supply the bytes.
This page covers the asset seam, curated icons, and the raster-bearing nodes.

## Which nodes carry assets

A node renders as a `pic` shape (with caller-supplied bytes) exactly when its IR
carries an `AssetID` field. That set is:

- **`Image`** — an asset image, optionally wrapped in device-frame chrome.
- **`Chart`** — a chart. In V1 a chart is an image-shape: you render the chart
  to a PNG/SVG elsewhere and supply the bytes (native `c:chart` is V2, D-004).
- **`CodeBlock`** — block-level code, rendered as a caller-side raster (D-014).
  You syntax-highlight and rasterize; pptx-go places the image and an optional
  caption.
- **`Decoration` with `Kind: DecorationAsset`** — caller-supplied ornament bytes.

Every other node renders as native PPTX shapes. The choice is intrinsic to the
node type — there is no deck-wide "raster everything" toggle.

```go
scene.Image{
	AssetID: "asset://hero-shot",
	Alt:     "Product screenshot",
	Frame:   scene.FrameBrowser,
}
scene.Chart{AssetID: "asset://revenue", Caption: "Revenue by quarter"}
scene.CodeBlock{AssetID: "asset://snippet-1", Language: "go", Caption: "main.go"}
```

## The AssetResolver seam

An `AssetResolver` maps an `AssetID` to bytes and a content-type hint:

```go
type AssetResolver interface {
	Resolve(ctx context.Context, id AssetID) ([]byte, string, error)
}
```

`AssetID` is free-form — pptx-go imposes no scheme. Register your resolver with
`scene.WithAssetResolver`:

```go
resolver := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
	b, err := os.ReadFile("assets/" + uuid + ".png")
	if err != nil {
		return nil, "", scene.ErrAssetNotFound
	}
	return b, "image/png", nil
})

stats, err := scene.Render(p, sc, scene.WithAssetResolver(resolver))
```

`URIAssetResolver` is a convenience that accepts `asset://<uuid>` ids and calls
your function with the bare uuid (a non-`asset://` id is passed through
unchanged). A resolver that has no bytes for an id returns
`(nil, "", scene.ErrAssetNotFound)`.

### Warn, don't fail

An unresolved asset is **not** a render error (D-036). The renderer places a
labeled placeholder where the asset would go and records a `LayoutWarning` in
`Stats.Warnings` — a deck with one missing chart still renders the other 49
slides. Inspect `stats.Warnings` to find what didn't resolve.

## Curated icons

Several nodes (a `Card`, a `FlowStep`) accept a closed-name icon. The curated set
ships embedded; reference an icon by name:

```go
scene.Card{
	Header: "Fast",
	Icon:   "check", // a curated icon name
	Body:   []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{{{Text: "Sub-second renders."}}}}},
}
```

Curated names include `check`, `x`, `plus`, `minus`, `star`, `circle`, `square`,
`triangle`, `diamond`, `dot`, and the `arrow-*` / `chevron-*` directions.

### Caller icon extensions

Register your own icon for a render with `scene.WithIconExtension`. The SVG is
validated when the option is applied — an SVG outside the translator subset fails
the render with a Stage-1 error, at registration, not at compose:

```go
stats, err := scene.Render(p, sc,
	scene.WithIconExtension("logo", logoSVG),
)
```

The icon translator accepts a constrained SVG subset:

- a **single path**,
- a **solid fill** (not `fill="none"`, not a `url(#…)` gradient/pattern
  reference),
- **no gradients**,
- **no elliptical arcs** (`A`/`a`) — author curves with Béziers instead.

Registering a curated name overrides that icon for the one render only;
extensions are per-render state, so concurrent renders with different extensions
do not interfere. Validate an SVG ahead of time with `scene.ValidateIcon(svg)`.

Frames and ornaments follow the same per-render extension pattern via
`scene.WithFrameExtension(name, recipe)` and
`scene.WithOrnamentExtension(name, recipe)`.

## Charts and code blocks are caller rasters

Because charts (D-004) and code blocks (D-014) are images in V1, the workflow is
the same for both: produce the bytes however you like (a charting library, a
syntax highlighter), register them through your resolver under an `AssetID`, and
reference that id from the node. pptx-go handles placement, optional captions,
and dedup — not pixels.
