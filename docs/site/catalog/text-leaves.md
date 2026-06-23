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

`ListIndent` values: `IndentNormal` (default), `IndentTight`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Kind` | `ListKind` | Marker style (bullet / number / checklist) |
| `Items` | `[]ListItem` | The list entries |
| `Indent` | `ListIndent` | Bullet hanging-indent density; `IndentTight` packs markers tighter to their text |

By default the bullet marker sits a 0.5" hanging indent from its text. Set
`Indent: scene.IndentTight` to halve that gap (to `In(0.25)`) so a list reads
dense and aligned instead of loose — consistently across all items and levels. The
default (`IndentNormal`) is unchanged.

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

## Button

A presentational CTA / action affordance: a content-fit pill with a bold label and
optional leading/trailing icons. Drop it standalone on a closing slide, at the foot of
a pricing card, or inside a banner. It is a shape only — there is no hyperlink or action
wiring (a `.pptx` deck is static).

| Field | Type | Meaning |
| --- | --- | --- |
| `Label` | `string` | The button text (required) |
| `Tone` | `ButtonTone` | Fill treatment: `ButtonPrimary` (accent solid, default), `ButtonAccentAlt`, `ButtonGhost` (outline), `ButtonNeutral` (surface) |
| `Size` | `ButtonSize` | `ButtonMD` (default), `ButtonSM`, `ButtonLG` — scales height, padding, and icon size |
| `LeadingIcon` | `string` | Closed-name registry icon before the label (e.g. `"star"`); `""` = none |
| `TrailingIcon` | `string` | Closed-name registry icon after the label (e.g. `"arrow-right"`); `""` = none |
| `Align` | `HAlign` | Center/right-align the pill within its box; `0` inherits the slide's content alignment |

The width is fit to the label plus any icons and clamped to the available box; a label
too wide for the box is shrunk to one line. Tone colors resolve through the theme, so a
theme swap re-skins every button (`ButtonPrimary` uses the accent token; `ButtonGhost`
is a no-fill pill with an accent hairline). A deck that uses no `Button` is unchanged.

```go
cta := scene.Button{
	Label:        "Talk to the team",
	Tone:         scene.ButtonPrimary,
	Size:         scene.ButtonLG,
	TrailingIcon: "arrow-right",
	Align:        scene.HAlignCenter,
}
```

## Checklist

A dense feature / "what you get" list: rows of a filled status glyph before rich text,
reflowed into 1–3 columns. The glyph is a real filled mark (a check, cross, or dot) —
never the empty square a font checkbox draws — and the text hangs indented past it.

| Field | Type | Meaning |
| --- | --- | --- |
| `Items` | `[]ChecklistItem` | The rows (each `{Text RichText; State CheckState; Icon string}`) |
| `Columns` | `int` | 1–3 columns; items reflow row-major (`0` = 1 column) |
| `GlyphTone` | `*ColorRole` | Override the glyph color for all rows; `nil` = per-state default |
| `Fill` | `bool` | Distribute rows to fill the box height (a short list spans the card) |

`CheckState` selects the glyph: `CheckDone` (a check, accent-tinted), `CheckNo` (a
cross, muted), `CheckNeutral` (a dot, muted). A per-item `Icon` (a closed-name registry
icon) overrides the state glyph. Place a `Checklist` with `Fill: true` in a card whose
body uses `VAlignFill` and the list spreads evenly to the card bottom. A deck that uses
no `Checklist` is unchanged.

```go
list := scene.Checklist{
	Columns: 2,
	Fill:    true,
	Items: []scene.ChecklistItem{
		{Text: scene.RichText{{Text: "Understands your data"}}, State: scene.CheckDone},
		{Text: scene.RichText{{Text: "Follows your rules"}}, State: scene.CheckDone},
		{Text: scene.RichText{{Text: "No training on prompts"}}, State: scene.CheckNo},
	},
}
```

## ChipRow

A horizontal row of tag / category chips — a labeled capability strip ("COMMON BUILDS ·
Finance · HR · …") or a card-footer tag row. Each chip is a real pill sized to its
label; chips lay left-to-right and reflow onto new lines when `Wrap` is set.

| Field | Type | Meaning |
| --- | --- | --- |
| `Label` | `string` | Optional leading caption before the first chip; `""` = none |
| `Chips` | `[]ChipSpec` | The chips (each `{Label string; Tone ChipTone; Color ColorRole; Icon string}`) |
| `Wrap` | `bool` | Reflow chips onto new lines within the width (zero = single line) |
| `Align` | `HAlign` | Align each line's chips left / center / right within the box |

Each `ChipSpec` mirrors the single `Chip`: `Tone` is `ChipTint` (default surface),
`ChipSolid` (the `Color` role), or `ChipOutline` (a `Color` hairline); a leading `Icon`
(a closed-name registry icon) is optional. Set `Wrap: true` for a long strip so it
reflows instead of overflowing.

```go
strip := scene.ChipRow{
	Label: "COMMON BUILDS",
	Wrap:  true,
	Chips: []scene.ChipSpec{
		{Label: "Finance"}, {Label: "HR"}, {Label: "Sales"},
		{Label: "Legal & Compliance"}, {Label: "Operations"},
	},
}
```

## Banner

A full-width filled "big takeaway / promo / call-to-action" strip — a leading icon, a
bold lead phrase + supporting body, and an optional right-aligned embedded action. Use
it across the top or bottom of a slide; it is distinct from the small side-bar `Callout`.

| Field | Type | Meaning |
| --- | --- | --- |
| `Lead` | `RichText` | The bold lead phrase |
| `Body` | `RichText` | The supporting body line |
| `Icon` | `string` | Optional leading registry icon; `""` = none |
| `Fill` | `ColorRole` | Strip fill; the zero value defaults to the accent role |
| `TextColor` | `TextColorRole` | Lead/body color; the zero value auto-contrasts against the fill |
| `Trailing` | `[]SlideNode` | Optional right-aligned children — a `Stat` and/or `Button` |

The strip is a single colored `RadiusLG` rounded rect; the lead/body sit on the left and
are kept legible against the fill automatically (or set `TextColor` explicitly). Embed a
`Button` in `Trailing` for a promo banner's action. A deck that uses no `Banner` is
unchanged.

```go
banner := scene.Banner{
	Lead: scene.RichText{{Text: "Run it internally, sell it externally"}},
	Body: scene.RichText{{Text: "the power of an agentic platform, without building one"}},
	Icon: "star",
	Fill: scene.ColorAccent,
	Trailing: []scene.SlideNode{
		scene.Button{Label: "Talk to the team", TrailingIcon: "arrow-right"},
	},
}
```
