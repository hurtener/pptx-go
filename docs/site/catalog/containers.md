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

```go
grid := scene.Grid{
	Columns: 3,
	Gap:     scene.SpaceMD,
	Cells: []scene.SlideNode{
		scene.Card{Header: "Fast"},
		scene.Card{Header: "Safe"},
		scene.Card{Header: "Simple"},
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

`HeaderFill` and `StatusDot` are `*ColorRole` (not `ColorRole`) so their `nil`
zero value means "omit" — a plain `ColorRole` zero is a real color (`canvas`).
Take the address of a role to set them: `hf := scene.ColorAccent; card.HeaderFill
= &hf`. All three reuse theme tokens, so a theme swap re-skins them.

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
