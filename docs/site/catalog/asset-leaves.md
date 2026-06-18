# Asset leaf nodes

Asset leaves carry an `AssetID` and render as a **picture (`pic`) shape**
(D-011, D-018). pptx-go never rasterizes: the caller pre-rasterizes the content
and supplies the bytes through an [`AssetResolver`](/guide/assets), which maps an
`AssetID` to bytes plus a content-type hint at render time. Register one with
`scene.WithAssetResolver` when you call `Render`.

The `AssetID` is **required** for every node on this page — an empty id fails
Stage-1 validation before any rendering happens. `AssetID` is a free-form
`string`; pptx-go imposes no scheme (callers choose, e.g. `asset://<uuid>`).

See the [index](/catalog/) for the shared render boilerplate and
[assets](/guide/assets) for resolver setup.

## Image

An asset image with optional device-frame chrome. Render policy: **picture**.

`Frame` selects a curated device frame by enum; `FrameName` selects a frame by
name and, when non-empty, takes precedence over `Frame` — it is the seam for a
caller frame registered via `scene.WithFrameExtension` (D-038). With both unset
(`FrameNone`, `""`) the image renders without a bezel. `Crop` trims the source
per edge and `Fit` selects the fill mode; both are mechanism exposure of the
builder's crop/fit (D-039), and their zero values (`Crop{}`, `FitFill`) render
the image uncropped and stretched.

`FrameKind` values: `FrameNone`, `FrameBrowser`, `FramePhone`, `FrameDesktop`,
`FrameLaptop`.

`Fit` values: `FitFill` (stretch to fill the box, the default), `FitNone`.

| Field | Type | Meaning |
| --- | --- | --- |
| `AssetID` | `AssetID` | Asset reference (required) |
| `Alt` | `string` | Alt text |
| `Frame` | `FrameKind` | Curated device-frame chrome |
| `FrameName` | `string` | Named frame; takes precedence over `Frame` when non-empty |
| `Crop` | `Crop` | Per-edge fractional crop (0..1 trimmed per edge) |
| `Fit` | `Fit` | Fill mode |

```go
image := scene.Image{
	AssetID: "asset://hero-screenshot",
	Alt:     "Product dashboard",
	Frame:   scene.FrameBrowser,
	Fit:     scene.FitFill,
}
```

## CodeBlock

Block-level code, rendered as a caller-rasterized picture (D-014). The caller
renders the syntax-highlighted code to image bytes; `Language` and `Caption` are
metadata. Render policy: **picture**.

| Field | Type | Meaning |
| --- | --- | --- |
| `AssetID` | `AssetID` | Asset reference to the rasterized code (required) |
| `Language` | `string` | Source language label |
| `Caption` | `string` | Optional caption |

```go
code := scene.CodeBlock{
	AssetID:  "asset://snippet-main-go",
	Language: "go",
	Caption:  "main.go",
}
```

## Chart

An image-shape chart. In V1 a chart is a caller-rasterized picture; a native
`c:chart` is V2 (D-004). Render policy: **picture**.

| Field | Type | Meaning |
| --- | --- | --- |
| `AssetID` | `AssetID` | Asset reference to the rasterized chart (required) |
| `Caption` | `string` | Optional caption |

```go
chart := scene.Chart{
	AssetID: "asset://revenue-bar-chart",
	Caption: "Revenue by quarter",
}
```

## Decoration (asset)

The `Decoration` node renders a caller-supplied asset as a picture when its
`Kind` is `DecorationAsset`; the `AssetID` is required in that case (Stage-1
validation). With `Kind: DecorationPreset` it renders a curated ornament
natively instead — that path and the full field table are documented under
[visual leaves](/catalog/visual-leaves). Render policy (this `Kind`):
**picture**.

```go
decoration := scene.Decoration{
	Kind:    scene.DecorationAsset,
	AssetID: "asset://watermark-logo",
	Layer:   scene.LayerBackground,
	Anchor:  scene.AnchorBottomRight,
	Offset:  scene.Position{X: pptx.In(-0.4), Y: pptx.In(-0.4)},
	Size:    scene.Size{W: pptx.In(1.5), H: pptx.In(1.5)},
	Opacity: 0.15,
}
```
