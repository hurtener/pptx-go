# Visual leaf nodes

Visual leaves draw structure and ornament. Every node on this page renders as
**native** PowerPoint shapes resolved from theme tokens. (A `Decoration` whose
`Kind` is `DecorationAsset` renders as a picture instead — see
[asset leaves](/catalog/asset-leaves).)

See the [index](/catalog/) for the shared render boilerplate.

## Divider

A horizontal rule with surrounding spacing. `Spacing` selects the vertical
breathing room from the spacing token scale. Render policy: **native**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Spacing` | `SpaceRole` | Vertical spacing around the rule |

```go
divider := scene.Divider{
	Spacing: scene.SpaceLG,
}
```

## SectionDivider

A full-bleed chapter break that occupies a whole slide: an optional eyebrow
kicker over a large label. Render policy: **native**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Eyebrow` | `string` | Small kicker label above the chapter label |
| `Label` | `string` | The chapter / section title |

```go
section := scene.SectionDivider{
	Eyebrow: "Part two",
	Label:   "Financials",
}
```

## Table

Headered tabular data; every cell is `RichText`. A non-empty `Caption` renders
as a separate text shape above the table. Render policy: **native**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Headers` | `[]RichText` | Header-row cells |
| `Rows` | `[][]RichText` | Body rows; each inner slice is a row of cells |
| `Caption` | `string` | Optional caption above the table |

```go
cell := func(s string) scene.RichText {
	return scene.RichText{{Text: s, Style: scene.RunStyle{TypeRole: scene.TypeBody}}}
}

table := scene.Table{
	Caption: "Revenue by region",
	Headers: []scene.RichText{cell("Region"), cell("Q1"), cell("Q2")},
	Rows: [][]scene.RichText{
		{cell("EMEA"), cell("4.1M"), cell("4.6M")},
		{cell("APAC"), cell("2.8M"), cell("3.2M")},
	},
}
```

## Flow

A sequential step pipeline. `Orientation` selects the layout direction, `Steps`
holds the pipeline entries, and `Connector` selects the inter-step glyph. The
`Connector` zero value (`ConnectorArrow`) keeps a solid-arrow pipeline (D-044). A
flow with no steps fails Stage-1 validation. Render policy: **native**.

`FlowOrientation` values: `FlowHorizontal`, `FlowVertical`.

`ConnectorKind` values: `ConnectorArrow` (solid arrow, default),
`ConnectorArrowDashed` (dashed line + chevron head), `ConnectorCycle` (arrows +
a trailing return arrow), `ConnectorPlus` (a plus glyph between steps).

| Field | Type | Meaning |
| --- | --- | --- |
| `Orientation` | `FlowOrientation` | Horizontal or vertical pipeline |
| `Steps` | `[]FlowStep` | The pipeline steps |
| `Connector` | `ConnectorKind` | Inter-step glyph |

`FlowStep` fields:

| Field | Type | Meaning |
| --- | --- | --- |
| `Label` | `RichText` | The step's primary label |
| `Detail` | `RichText` | Optional second line |
| `Icon` | `string` | Optional curated/extension icon name (closed-name; Stage-1 validated) |

```go
flow := scene.Flow{
	Orientation: scene.FlowHorizontal,
	Connector:   scene.ConnectorArrow,
	Steps: []scene.FlowStep{
		{Label: scene.RichText{{Text: "Draft", Style: scene.RunStyle{TypeRole: scene.TypeBody}}}},
		{Label: scene.RichText{{Text: "Review", Style: scene.RunStyle{TypeRole: scene.TypeBody}}}},
		{Label: scene.RichText{{Text: "Publish", Style: scene.RunStyle{TypeRole: scene.TypeBody}}}},
	},
}
```

## Decoration (preset)

An anchored ornament. With `Kind: DecorationPreset` it renders a curated
ornament **natively** (an SVG translated to a preset geometry or path); `Preset`
names the curated ornament and is required for the preset path (Stage-1
validation). The same struct renders a caller-supplied asset as a picture when
`Kind` is `DecorationAsset` — see [asset leaves](/catalog/asset-leaves).

`DecorationKind` values: `DecorationPreset`, `DecorationAsset`.

`Layer` values: `LayerBackground` (behind body content), `LayerForeground`
(above it).

The placement box aligns the box point that corresponds to `Anchor` to that
anchor point on the slide, shifted by `Offset` and sized by `Size` (a zero
`Size` uses a default). `Bleed` permits the box to extend past the slide edge
without a warning. `Opacity` (0..1; 0 = fully opaque) dims the ornament, and
`Rotation` (degrees clockwise) rotates a single-shape ornament (a multi-shape
ornament cannot rotate as a unit in V1 — D-041).

| Field | Type | Meaning |
| --- | --- | --- |
| `Kind` | `DecorationKind` | Preset (native) or asset (picture) source |
| `Preset` | `string` | Curated ornament name (used when `Kind == DecorationPreset`) |
| `AssetID` | `AssetID` | Asset reference (used only when `Kind == DecorationAsset`) |
| `Layer` | `Layer` | Background or foreground z-order |
| `Anchor` | `Anchor` | Reference point on the slide |
| `Offset` | `Position` | EMU shift from the anchor point |
| `Size` | `Size` | Ornament box extent; zero = a default size |
| `Bleed` | `bool` | Allow the box to extend past the slide edge |
| `Opacity` | `float64` | 0..1; 0 = fully opaque |
| `Rotation` | `float64` | Degrees clockwise |

```go
decoration := scene.Decoration{
	Kind:    scene.DecorationPreset,
	Preset:  "corner-chevron",
	Layer:   scene.LayerBackground,
	Anchor:  scene.AnchorTopRight,
	Offset:  scene.Position{X: pptx.In(-0.5), Y: pptx.In(0.5)},
	Size:    scene.Size{W: pptx.In(2), H: pptx.In(2)},
	Opacity: 0.2,
}
```
