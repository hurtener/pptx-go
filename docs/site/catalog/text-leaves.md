# Text leaf nodes

Text leaves carry styled text and render as **native** PowerPoint shapes (text
bodies, pills, lines). Text fields are `scene.RichText` — an ordered list of
styled runs — so inline color and typography flow through theme tokens (P2). A
minimal run is:

```go
scene.RichText{{Text: "…", Style: scene.RunStyle{TypeRole: scene.TypeBody}}}
```

See the [index](/catalog/) for the shared render boilerplate, and the
[scene guide](/guide/scene) for `RichText` details.

## Hero

A cover-slide title block: an optional eyebrow kicker, a title, and an optional
subtitle. Render policy: **native**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Eyebrow` | `string` | Small kicker label above the title |
| `Title` | `string` | The main cover title |
| `Subtitle` | `string` | Optional supporting line below the title |
| `AutoFit` | `bool` | Shrink the title to fit one line when it would overflow |

```go
hero := scene.Hero{
	Eyebrow:  "FY26 Q2",
	Title:    "Quarterly review",
	Subtitle: "Results, risks, and outlook",
}
```

Set `AutoFit` to keep a long title on one line: when the title's estimated width
exceeds the box, the engine downscales its font (to no less than 60% of the role
size) instead of letting it wrap. A title that already fits is unchanged.

## Prose

One or more body paragraphs. Render policy: **native**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Paragraphs` | `[]RichText` | One body paragraph per element |

```go
prose := scene.Prose{
	Paragraphs: []scene.RichText{
		{{Text: "Revenue grew 12% year over year.", Style: scene.RunStyle{TypeRole: scene.TypeBody}}},
		{{Text: "Margins held steady across all regions.", Style: scene.RunStyle{TypeRole: scene.TypeBody}}},
	},
}
```

## Heading

A section heading at a level from 1 to 6 (Stage-1 validation rejects out-of-range
levels). Render policy: **native**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Text` | `RichText` | The heading text |
| `Level` | `int` | Heading level, 1..6 |
| `AutoFit` | `bool` | Shrink the heading to fit one line when it would overflow |

```go
heading := scene.Heading{
	Text:  scene.RichText{{Text: "Regional breakdown", Style: scene.RunStyle{TypeRole: scene.TypeH2}}},
	Level: 2,
}
```

## List

A bullet, numbered, or checklist block. `Kind` selects the marker style; each
`ListItem` carries its own indent `Level` and (for checklists) a `Checked` flag.
A list with no items fails Stage-1 validation. Render policy: **native**.

`ListKind` values: `ListBullet`, `ListNumber`, `ListChecklist`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Kind` | `ListKind` | Marker style (bullet / number / checklist) |
| `Items` | `[]ListItem` | The list entries |

`ListItem` fields:

| Field | Type | Meaning |
| --- | --- | --- |
| `Text` | `RichText` | The item text |
| `Level` | `int` | Indent depth |
| `Checked` | `bool` | Checked state (checklist items) |

```go
list := scene.List{
	Kind: scene.ListChecklist,
	Items: []scene.ListItem{
		{Text: scene.RichText{{Text: "Close the books", Style: scene.RunStyle{TypeRole: scene.TypeBody}}}, Checked: true},
		{Text: scene.RichText{{Text: "Ship the report", Style: scene.RunStyle{TypeRole: scene.TypeBody}}}, Level: 1},
	},
}
```

## Quote

A pull quote with optional attribution. Render policy: **native**.

| Field | Type | Meaning |
| --- | --- | --- |
| `Text` | `RichText` | The quoted text |
| `Attribution` | `string` | Optional attribution line |

```go
quote := scene.Quote{
	Text:        scene.RichText{{Text: "Simplicity is the ultimate sophistication.", Style: scene.RunStyle{TypeRole: scene.TypeH3, Italic: true}}},
	Attribution: "Leonardo da Vinci",
}
```

## Callout

A colored side-bar note. `Kind` selects the tone (and the accent color). Render
policy: **native**.

`CalloutKind` values: `CalloutNote`, `CalloutWarning`, `CalloutTip`,
`CalloutImportant`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Kind` | `CalloutKind` | Tone of the callout |
| `Title` | `string` | Optional callout title |
| `Body` | `RichText` | The callout body text |

```go
callout := scene.Callout{
	Kind:  scene.CalloutWarning,
	Title: "Heads up",
	Body:  scene.RichText{{Text: "Figures are preliminary and unaudited.", Style: scene.RunStyle{TypeRole: scene.TypeBody}}},
}
```

## Chip

An inline pill. `Tone` selects the fill treatment and `Color` selects the
semantic color role. Render policy: **native**.

`ChipTone` values: `ChipTint`, `ChipSolid`, `ChipOutline`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Label` | `string` | The pill text |
| `Tone` | `ChipTone` | Fill treatment (tint / solid / outline) |
| `Color` | `ColorRole` | Semantic color role driving the pill |

```go
chip := scene.Chip{
	Label: "New",
	Tone:  scene.ChipSolid,
	Color: scene.ColorAccent,
}
```

## Arrow

An inline directional connector with an optional label. `Direction` selects the
arrowhead orientation. Render policy: **native**.

`ArrowDirection` values: `ArrowRight`, `ArrowLeft`, `ArrowUp`, `ArrowDown`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Direction` | `ArrowDirection` | Which way the arrow points |
| `Label` | `string` | Optional label alongside the arrow |

```go
arrow := scene.Arrow{
	Direction: scene.ArrowRight,
	Label:     "then",
}
```

## Stat

A hero big-number metric: a display-scale value with a label and an optional
directional delta. A `Grid` of `Stat`s forms a metric/pricing strip. The engine
renders the value and delta verbatim — it formats no numbers.

| Field | Type | Meaning |
| --- | --- | --- |
| `Value` | `string` | The big number, rendered at display scale (e.g. `"$2,200"`) |
| `Label` | `string` | Caption below the value |
| `Delta` | `string` | Optional delta (e.g. `"+12%"`); `""` = no delta line |
| `DeltaTone` | `DeltaTone` | Delta color direction: `DeltaUp` (success), `DeltaDown` (error), `DeltaNeutral` (muted, default) |
| `AutoFit` | `bool` | Shrink the value to fit its column when a long number/price would overflow |

In a narrow pricing column a long value like `"$4,000+"` can wrap to two lines.
Set `AutoFit` and the engine downscales the value font (to no less than 60% of the
display size) so it fits one line. A value that already fits is unchanged.

```go
strip := scene.Grid{
	Columns: 3,
	Cells: []scene.SlideNode{
		scene.Stat{Value: "$2,200", Label: "ARR", Delta: "+12%", DeltaTone: scene.DeltaUp},
		scene.Stat{Value: "38%", Label: "Margin", Delta: "-3%", DeltaTone: scene.DeltaDown},
		scene.Stat{Value: "4.8", Label: "NPS"},
	},
}
```
