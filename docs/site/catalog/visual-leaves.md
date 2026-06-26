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

## Timeline (roadmap)

A roadmap / timeline: a horizontal axis with milestones placed at proportional
positions, optional phase bands behind the axis, and optional swimlanes (rows).
Markers (an accent dot, or a curated `Icon`), the axis line, and labels render
**natively** (no media); labels stagger above/below the axis to avoid collision.
The caller maps dates to `0..1` positions; the engine places the fraction
deterministically.

Empty `Lanes` renders one implicit lane from the top-level `Milestones`; non-empty
`Lanes` renders swimlane rows (each a left-gutter label + its own axis). `Bands`
span the full timeline width behind every lane.

| Field | Type | Meaning |
| --- | --- | --- |
| `Milestones` | `[]Milestone` | Single-lane milestones (used when `Lanes` is empty) |
| `Lanes` | `[]TimelineLane` | Swimlane rows; supersedes `Milestones` when non-empty |
| `Bands` | `[]TimelineBand` | Phase/horizon regions drawn behind the axis |

`Milestone` fields: `Position float64` (0..1 along the axis), `Label`, `Detail`
`string`, `Icon string` (replaces the dot marker), `AccentIndex int` (selects the
marker color from a pinned token cycle; 0 = accent).

`TimelineLane` fields: `Label string`, `Milestones []Milestone`.

`TimelineBand` fields: `From`, `To float64` (0..1), `Label string`, `Fill
ColorRole`.

```go
timeline := scene.Timeline{
	Bands: []scene.TimelineBand{
		{From: 0, To: 0.5, Label: "Now", Fill: pptx.ColorAccent},
		{From: 0.5, To: 1, Label: "Next", Fill: pptx.ColorInfo},
	},
	Lanes: []scene.TimelineLane{
		{Label: "Platform", Milestones: []scene.Milestone{
			{Position: 0.1, Label: "Beta", Icon: "star"},
			{Position: 0.8, Label: "GA", Detail: "Q4"},
		}},
	},
}
```

## DataMark (native micro-chart)

A crisp, brand-colored vector micro-chart drawn entirely from preset shapes —
**native**, no raster, no AssetResolver. `Kind` selects the shape; values are
`0..1`; it sizes to its box and embeds in a Card/Bento cell.

`DataMarkKind` values: `DataMarkBar` (a progress/capacity bar — a track + a fill to
`Value`, with an optional inline `Label`; `Orientation` `FlowHorizontal` (default)
or `FlowVertical`), `DataMarkBars` (a small bar group, one bar per `Values` entry),
`DataMarkSparkline` (a trend polyline through `Values` with an accent end dot),
`DataMarkDonut` (a single-value ring + centered `Label` — a 331° accent arc at
0.92), `DataMarkGauge` (a 270° speedometer). The donut/gauge are native `blockArc`
ring sectors (a value arc + a remainder arc, no hole).

| Field | Type | Meaning |
| --- | --- | --- |
| `Kind` | `DataMarkKind` | The mark shape |
| `Value` | `float64` | 0..1 fraction for `DataMarkBar` |
| `Values` | `[]float64` | 0..1 fractions for `DataMarkBars` / `DataMarkSparkline` |
| `Orientation` | `FlowOrientation` | `DataMarkBar` direction (horizontal default / vertical) |
| `Color` | `*ColorRole` | Mark color (nil = accent); the track is always `ColorSurfaceAlt` |
| `Label` | `string` | Optional inline label (right of a horizontal bar) |

```go
bar := scene.DataMark{Kind: scene.DataMarkBar, Value: 0.92, Label: "92%"}
bars := scene.DataMark{Kind: scene.DataMarkBars, Values: []float64{0.3, 0.6, 0.9, 0.5}}
spark := scene.DataMark{Kind: scene.DataMarkSparkline, Values: []float64{0.2, 0.8, 0.4, 1.0}}
```
