# The builder (`pptx`)

The `pptx` package is the general-purpose, theme-aware PPTX builder. You create a
`*Presentation`, add slides, place shapes and text and tables and images on them,
then write the result. Styling flows through [theme tokens](/guide/theme) by
default; literals are an escape hatch.

All geometry is in EMU (English Metric Units, OOXML's canonical integer length;
1 inch = 914400 EMU). Construct EMU values with the converters `pptx.In`,
`pptx.Cm`, `pptx.Pt`, `pptx.Px`, and position shapes with a `pptx.Box`:

```go
box := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(4), H: pptx.In(2)}
```

## Creating a presentation

`pptx.New` returns a valid, ready-to-edit deck. With no options it is a 16:9
widescreen deck themed with the default theme.

```go
p := pptx.New(
	pptx.WithFormat(pptx.Slides16x9),       // or pptx.Slides4x3
	pptx.WithTheme(pptx.DefaultTheme()),    // active theme for token resolution
)
```

The construction options:

| Option | Purpose |
| --- | --- |
| `WithFormat(Format)` | Slide canvas aspect ratio: `Slides16x9` (default) or `Slides4x3`. |
| `WithTheme(*Theme)` | Active theme that drives token resolution. See [themes & tokens](/guide/theme). |
| `WithLogger(*slog.Logger)` | Structured logger. Emits a write-boundary event on save; emits read degradations on the read constructors. No logger = no logs. |
| `WithFontSource(FontSource)` | Registers the source `EmbedFont` (and the auto font-embedding pass) resolves font bytes from. |
| `WithFontEmbedding()` | At save, automatically embed every font face the deck uses via the registered `FontSource`. No-op without a source; byte-identical when off. See [embedding fonts](#embedding-fonts). |
| `WithReadPartLimit(int64)` | Per-part decompressed size ceiling for the read constructors (no-op on `New`). See [reading decks](/guide/reading). |
| `FromTemplate(*Presentation)` | Seeds the deck from a brand-kit template: its theme, masters, and layouts are adopted. |

`FromTemplate` adopts an opened deck as a brand kit:

```go
brand, err := pptx.OpenStream("brand-template.pptx")
if err != nil {
	log.Fatal(err)
}
defer brand.Close()

p := pptx.New(pptx.FromTemplate(brand)) // theme + masters + layouts adopted, starts empty
```

## Adding slides

```go
s := p.AddSlide()              // blank layout
title := p.AddSlide("Title Slide") // a named layout from the active template
```

`AddSlide` appends and returns a `*Slide`. There is also `AddSlideAt(index, …)`
to insert, `RemoveSlide(index)`, `GetSlide(index)`, and `Slides()`.

## Shapes

`Slide.AddShape` is the token-aware shape API. It takes a preset geometry, a
`Box`, and zero or more `ShapeOption`s, and returns a `*Shape` handle.

```go
s.AddShape(pptx.ShapeRoundRect,
	pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(3), H: pptx.In(1.5)},
	pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))),
	pptx.WithLine(pptx.Line{Width: pptx.Pt(1), Color: pptx.TokenColor(pptx.ColorAccentAlt)}),
	pptx.WithRadius(pptx.RadiusLG),
	pptx.WithElevation(pptx.ElevationRaised),
	pptx.WithRotation(3),
)
```

Preset geometries (`ShapeGeometry`) include `ShapeRect`, `ShapeRoundRect`,
`ShapeEllipse`, `ShapeTriangle`, `ShapeDiamond`, `ShapeParallelogram`,
`ShapeHexagon`, `ShapeChevron`, `ShapeRightArrow`, and `ShapeLine`.

The shape options:

| Option | Effect |
| --- | --- |
| `WithFill(Fill)` | Interior fill: `SolidFill(Color)`, `NoFill()`, `LinearGradient(angle, …stops)`, `RadialGradient(…stops)`. |
| `WithLine(Line)` | Outline: `Line{Width, Color, Dash}`. |
| `WithRadius(RadiusRole)` | Rounded-corner radius from a theme token. Applies to `ShapeRoundRect` only. |
| `WithElevation(ElevationRole)` | Drop shadow from the theme's elevation token (the documented path). |
| `WithRotation(deg float64)` | Clockwise rotation about the shape's center, normalized to `[0, 360°)`. |

Fills and lines accept a `Color`: a literal (`pptx.RGB("2563EB")`,
`pptx.RGBA(hex, alpha)`) or a token (`pptx.TokenColor(role)`,
`pptx.TokenTextColor(role)`). Tokens resolve against the active theme at
`AddShape` time, so a theme swap re-renders the same shape.

## Rich text

A `TextFrame` is a shape-level rich-text container: `TextFrame → Paragraph →
Run`. Create one with `Slide.AddTextFrame`, then add paragraphs and styled runs.

```go
tf := s.AddTextFrame(pptx.Box{X: pptx.In(1), Y: pptx.In(2), W: pptx.In(8), H: pptx.In(3)})
tf.AutoFit(pptx.AutoFitNormal).Anchor(pptx.AnchorTop)

h := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignLeft})
h.AddRun("Quarterly results", pptx.RunStyle{TypeRole: pptx.TypeH2})

body := tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletDisc, Level: 0})
body.AddRun("Revenue up ", pptx.RunStyle{TypeRole: pptx.TypeBody})
body.AddRun("18%", pptx.RunStyle{
	TypeRole: pptx.TypeBody,
	Bold:     true,
	Color:    pptx.TokenTextColor(pptx.TextSuccess),
})
body.AddHyperlink(" — see report", "https://example.com/q3",
	pptx.RunStyle{TypeRole: pptx.TypeBody, Color: pptx.TokenTextColor(pptx.TextAccent)})
```

`RunStyle` carries `TypeRole` (a typography token — `TypeDisplay`, `TypeH1`…`H5`,
`TypeBody`, `TypeBodySmall`, `TypeCaption`, `TypeMono`, `TypeCode`), `Color`, and
the inline toggles `Bold`, `Italic`, `Underline`, `Strike`, `BaselineRel`, and
`Code` (inline monospace). `ParagraphOpts` carries `Align`, `Level`, and
`Bullet` (`BulletNone`, `BulletDisc`, `BulletNumber`, `BulletCheckbox`). The
frame supports `AutoFit(mode)`, `Anchor(v)`, and `Margins(top, right, bottom,
left)`.

## Tables

`Slide.AddTable` adds a native table positioned by a `Box`, with equal column
widths by default. Cells carry rich text, fills, borders, and merges.

```go
t := s.AddTable(pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(8), H: pptx.In(3)}, 3, 3)
t.SetHeaderRow(true)
t.SetBanding(true, false) // alternating row fills
t.SetColumnWidths(pptx.In(3), pptx.In(2.5), pptx.In(2.5))

t.Cell(0, 0).SetText("Region")
t.Cell(0, 1).SetText("Revenue")
t.Cell(0, 2).SetText("Growth")

t.Cell(1, 0).SetText("EMEA")
t.Cell(1, 1).SetFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt)))

// Merges: span columns or rows from the anchor cell.
t.Cell(2, 0).MergeRight(2)
t.Cell(1, 2).MergeDown(2)
```

`Cell.TextFrame()` returns the cell's full rich-text frame when you need more
than `SetText`; `SetBorders(Line)` sets all four edges.

## Images

Images enter through an `ImageSource` — `ImageBytes`, `ImageFile`, or
`ImageReader`. `Slide.AddImage` places one in a `Box` and returns an `*Image`
handle for alt text, crop, fit, rotation, and opacity. Identical bytes across
the deck are written to the package once.

```go
img, err := s.AddImage(pptx.ImageFile("chart.png"),
	pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(5), H: pptx.In(3)})
if err != nil {
	log.Fatal(err)
}
img.SetAltText("Q3 revenue chart").
	SetCrop(pptx.Crop{Top: 0.05, Bottom: 0.05}).
	SetFit(pptx.FitFill)

// Or from bytes with a declared MIME type, verified against the bytes:
_, _ = s.AddImage(pptx.ImageBytes(pngBytes, "image/png"), box)
```

pptx-go verifies the byte signature matches a known image type (PNG, JPEG, GIF,
BMP, WebP) but never parses pixel data.

## Speaker notes

Every slide can carry speaker notes, authored through the same rich-text frame:

```go
s.SetSpeakerNotes("Pause here for questions.")    // convenience: one plain paragraph
tf := s.SpeakerNotes()                            // or author rich notes
tf.AddParagraph(pptx.ParagraphOpts{}).
	AddRun("Emphasize the 18% number.", pptx.RunStyle{TypeRole: pptx.TypeBody, Bold: true})
```

## Sections

Sections are named, ordered slide groupings (shown in PowerPoint's slide sorter).

```go
intro := p.AddSection("Introduction")
intro.Include(s)
```

Slides left unassigned fall into an implicit leading "Default Section" so the
section list spans every slide.

## Embedding fonts

A deck themed with a brand display or heading face only renders with that face on
machines where it is installed — unless the face's bytes ship inside the `.pptx`.
Register a `FontSource` (it resolves a `(name, style, weight)` to font bytes) and
turn on the automatic embedding pass:

```go
p := pptx.New(
	pptx.WithTheme(brandTheme),       // names e.g. "Playfair Display" / "Inter"
	pptx.WithFontSource(myFontStore), // resolves family bytes
	pptx.WithFontEmbedding(),         // embed every used face at save
)
// …add slides…
p.Save("deck.pptx") // every face the deck actually uses is embedded
```

At save the pass walks every run, collects the distinct used faces — by family,
bold, and italic — in a stable sorted order, and embeds each via the source. It
is:

- **a no-op without a `FontSource`** (and byte-identical to the prior output when
  `WithFontEmbedding` is off);
- **warn-don't-fail** — a face the source cannot resolve logs a warning and is
  skipped; the save still succeeds;
- **idempotent** — a face you embedded by hand with `EmbedFont(name, style,
  weight)` is not embedded twice;
- **deterministic** — two saves of the same deck are byte-identical.

For a single face, call `EmbedFont(name, style, weight)` directly.

## Saving

| Method | Output |
| --- | --- |
| `Save(path string) error` | Writes a `.pptx` file. |
| `Write(w io.Writer) error` | Streams to any writer (e.g. an HTTP response). |
| `WriteToBytes() ([]byte, error)` | Returns the package as a byte slice. |
| `SaveStream(path string) error` | Writes through the OPC streaming writer (lazy per-part). |

```go
if err := p.Save("deck.pptx"); err != nil {
	log.Fatal(err)
}
```

Every write path runs the same always-on hygiene pass, so the deck opens in
PowerPoint without a repair prompt.
