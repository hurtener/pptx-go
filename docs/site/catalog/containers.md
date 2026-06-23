# Container nodes

Container nodes hold child `SlideNode`s and introduce sub-layouts. A container
has no render policy of its own: it lays out its children, and **each child
renders per its own policy** (D-011, D-018). A native `Card` can hold an `Image`
child that renders as a picture. See the [index](/catalog/) for the shared
render boilerplate.

## TwoColumn

Splits the body into left and right regions, each a list of leaf children.
`Ratio` selects the split. Render policy: **container** (children render per
their own policy).

`ColumnRatio` values: `Ratio11` (1:1), `Ratio12` (1:2), `Ratio21` (2:1).

| Field | Type | Meaning |
| --- | --- | --- |
| `Ratio` | `ColumnRatio` | Left/right width split |
| `Left` | `[]SlideNode` | Children in the left region |
| `Right` | `[]SlideNode` | Children in the right region |
| `Join` | `ColumnJoin` | Centered seam element: `JoinNone` (default), `JoinBadge`, `JoinArrow` |
| `JoinLabel` | `string` | Badge text when `Join == JoinBadge` (e.g. `"VS"`) |

`Join` draws an optional element straddling the seam between the columns: a
circular "VS"-style badge (`JoinBadge` + `JoinLabel`) for comparing two cards, or
a right-arrow connector (`JoinArrow`). `JoinNone` (the zero value) draws nothing.

```go
two := scene.TwoColumn{
	Ratio: scene.Ratio12,
	Left: []scene.SlideNode{
		scene.Heading{Text: scene.RichText{{Text: "Context", Style: scene.RunStyle{TypeRole: scene.TypeH3}}}, Level: 3},
	},
	Right: []scene.SlideNode{
		scene.Prose{Paragraphs: []scene.RichText{
			{{Text: "Details go here.", Style: scene.RunStyle{TypeRole: scene.TypeBody}}},
		}},
	},
}
```

## Grid

A 2/3/4-column layout with weighted ratios and one child per cell. `Columns`
must be 2, 3, or 4. `Ratio` holds per-column weights (empty = equal columns).
`Gap` selects inter-cell spacing from the spacing token scale. Render policy:
**container**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Columns` | `int` | Number of columns (2..4) |
| `Ratio` | `[]int` | Per-column weights; empty = equal |
| `Gap` | `SpaceRole` | Spacing between cells |
| `Cells` | `[]SlideNode` | One child per cell |
| `Connectors` | `[]GridConnector` | Inter-column gutter glyphs (`{Between [2]int; Kind ConnectorKind; Label string}`); empty = none |

`Connectors` draw a glyph in the gutter between two adjacent columns, so an
architecture / pipeline grid reads as flow. `Between` is the two adjacent column
indices (e.g. `{0, 1}`); `Kind` reuses the flow connectors (`ConnectorArrow`,
`ConnectorBiArrow`, …); an optional `Label` sits below the glyph. An empty slice is
byte-identical.

```go
grid := scene.Grid{
	Columns: 3,
	Gap:     scene.SpaceMD,
	Cells: []scene.SlideNode{
		scene.Card{Header: "People"},
		scene.Card{Header: "Operating layer"},
		scene.Card{Header: "Knowledge"},
	},
	Connectors: []scene.GridConnector{
		{Between: [2]int{0, 1}, Kind: scene.ConnectorArrow},
		{Between: [2]int{1, 2}, Kind: scene.ConnectorBiArrow, Label: "feeds"},
	},
}
```

## Bento

A row-labeled grid: rows that each carry an optional left label and cells of
variable column span on a shared column grid. Unlike `Grid` (uniform columns,
one child per cell), a `Bento` row can mix wide and narrow cells, and the spans
align across rows. Render policy: **container**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Columns` | `int` | Shared column units a row's spans are measured against (≥ 1) |
| `Rows` | `[]BentoRow` | The rows (`{Label string; Cells []BentoCell}`) |
| `WeightedRows` | `bool` | Opt-in: size each row to its content (clamped to fit) instead of equal rows |

Each `BentoCell` is `{Span int; Node SlideNode}` — a span-S cell occupies S of
the `Columns` units (so a span-2 cell is twice a span-1 cell). A row's spans sum
to ≤ `Columns`. The left-label gutter is reserved only when at least one row sets
a `Label`.

By default every bento row gets an equal share of the height. When a sparse row
and a dense row share a bento that wastes the sparse row's band and starves the
dense one. Set `WeightedRows: true` to size each row to its content's preferred
height instead — the dense row grows, the sparse row shrinks, and if the rows
would overflow they are scaled down together so the bento always fits its region.
The default (equal rows) is unchanged.

```go
bento := scene.Bento{
	Columns: 3,
	Rows: []scene.BentoRow{
		{Label: "Revenue", Cells: []scene.BentoCell{
			{Span: 2, Node: scene.Card{Header: "ARR"}},
			{Span: 1, Node: scene.Card{Header: "Growth"}},
		}},
		{Label: "Costs", Cells: []scene.BentoCell{
			{Span: 1, Node: scene.Card{Header: "COGS"}},
			{Span: 1, Node: scene.Card{Header: "Opex"}},
			{Span: 1, Node: scene.Card{Header: "Margin"}},
		}},
	},
}
```

## Card

An accent card: chrome (a rounded rectangle + accent stripe + optional
icon/eyebrow/header/header-pill) over a body of leaf children. All fields beyond
`Header`/`Body`/`BodyLayout`/`Fill`/`Outline`/`Elevation` are additive (D-043):
their zero values reproduce the earlier render byte-for-byte. Render policy:
**container**.

`BodyLayout` values: `BodyVertical`, `BodyHorizontal`.

`BorderStyle` values: `BorderDefault` (defer to `Outline`), `BorderNone` (no
border even if `Outline` is true), `BorderSolid` (neutral hairline),
`BorderAccent` (accent-colored). An explicit style overrides `Outline` (D-043).

`CardSize` values: `CardSizeMD` (default), `CardSizeSM`, `CardSizeLG` — interior
padding scale.

`CardLayout` values: `CardLayoutDefault` (icon left of the eyebrow/header
stack), `CardLayoutIconTop` (icon above it).

| Field | Type | Meaning |
| --- | --- | --- |
| `Header` | `string` | Card header text |
| `Eyebrow` | `string` | Kicker label above the header |
| `Icon` | `string` | Curated/extension icon name (closed-name; Stage-1 validated) |
| `HeaderPill` | `string` | Pill badge text, right of the header row |
| `Body` | `[]SlideNode` | Child nodes inside the card |
| `BodyLayout` | `BodyLayout` | Vertical or horizontal child stacking |
| `Fill` | `ColorRole` | Card fill color role |
| `Outline` | `bool` | Legacy border shorthand; see `BorderStyle` |
| `BorderStyle` | `BorderStyle` | Explicit border; `BorderDefault` defers to `Outline` |
| `Size` | `CardSize` | Interior padding scale |
| `Layout` | `CardLayout` | Header arrangement |
| `Elevation` | `ElevationRole` | Shadow/elevation role |
| `HeaderFill` | `*ColorRole` | Colored header band (body keeps `Fill`); `nil` = none |
| `StatusDot` | `*ColorRole` | Small status dot, top-right corner; `nil` = none |
| `Watermark` | `string` | Large, low-opacity label behind the body; `""` = none |
| `BodyVAlign` | `VAlign` | Vertical distribution of the body within the card (`Top`/`Center`/`Bottom`/`Justify`/`Fill`/`FillCapped`/`Balanced`/`Fit`); zero `Top` = top-anchored |
| `PaddingScale` | `int` | Basis-point multiplier on the size-resolved interior padding (0/10000 = unchanged; tighten a dense card, floored at a minimum) |
| `Ribbon` | `*Ribbon` | Pinned emphasis badge (`{Text; Position; Color *ColorRole; TextColor}`); `nil` = none |

By default a card body is top-anchored, so a short body floats in the upper card
with empty space below. Set `BodyVAlign` to distribute it: `VAlignBottom` pins
secondary content (a badge, a CTA) to the card bottom, `VAlignJustify` spreads the
items to fill, `VAlignFill`/`VAlignFillCapped` grow flexible body nodes (capped
turns the leftover into even spacing), `VAlignBalanced` distributes a sparse body
as an even, optically-centered rhythm, and `VAlignFit` compresses an over-full
body. The zero value (`VAlignTop`) is byte-identical to the prior layout. Applies
to the vertical body only (not `BodyHorizontal`).

`HeaderFill` and `StatusDot` are `*ColorRole` (not `ColorRole`) so their `nil`
zero value means "omit" — a plain `ColorRole` zero is a real color (`canvas`).
Take the address of a role to set them: `hf := scene.ColorAccent; card.HeaderFill
= &hf`. All three reuse theme tokens, so a theme swap re-skins them.

`Ribbon` pins an emphasis badge outside the header text flow — distinct from the
in-row `HeaderPill` — to single one card out of a row. `RibbonTopBar` is a
full-width tab across the top that reserves a band (the body shifts down below it);
`RibbonCornerStar` is a star glyph in the top-right corner; `RibbonCornerTL`/
`RibbonCornerTR` are content-fit corner text tabs. Its `Color` (`*ColorRole`, nil =
accent) and `TextColor` (auto-contrast by default) are theme tokens. `nil` = no
ribbon.

```go
card := scene.Card{
	Eyebrow:    "Speed",
	Header:     "Fast by default",
	Icon:       "zap",
	HeaderPill: "New",
	Fill:       scene.ColorSurface,
	BorderStyle: scene.BorderAccent,
	Size:       scene.CardSizeMD,
	Layout:     scene.CardLayoutDefault,
	Elevation:  scene.ElevationRaised,
	BodyLayout: scene.BodyVertical,
	Body: []scene.SlideNode{
		scene.Prose{Paragraphs: []scene.RichText{
			{{Text: "Renders in milliseconds.", Style: scene.RunStyle{TypeRole: scene.TypeBody}}},
		}},
	},
}
```

## CardSection

A top-level card that accepts grids, two-column splits, or nested cards as its
body. Render policy: **container**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Header` | `string` | Section header text |
| `Body` | `[]SlideNode` | Child nodes (grid / two_column / nested cards) |

```go
section := scene.CardSection{
	Header: "Highlights",
	Body: []scene.SlideNode{
		scene.Grid{
			Columns: 2,
			Gap:     scene.SpaceMD,
			Cells: []scene.SlideNode{
				scene.Card{Header: "Revenue"},
				scene.Card{Header: "Margin"},
			},
		},
	},
}
```
