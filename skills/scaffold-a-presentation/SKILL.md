---
name: scaffold-a-presentation
description: Build a PowerPoint deck from scratch with the pptx (Layer 1) builder — create a presentation, add slides, place shapes, rich text, tables, and images styled with theme tokens, attach speaker notes and sections, save to a file or bytes, and read a deck back. Use this when an agent needs to author a .pptx programmatically in Go without a scene IR, or to inspect/round-trip an existing deck.
---

# Scaffold a presentation

## Overview

`github.com/hurtener/pptx-go/pptx` is the Layer 1 builder: general-purpose,
theme-aware PowerPoint (PPTX / Open Office XML) authoring in pure Go. You
create a `*Presentation`, add `*Slide`s, place primitives (shapes, text frames,
tables, images) positioned in EMU, style them through semantic theme tokens,
then write the deck to a file, an `io.Writer`, or a `[]byte`. The same package
reads a deck back: a deck pptx-go authored round-trips losslessly.

This skill is the builder quickstart. For typed scene composition use
`compose-a-scene` (Layer 2); for defining a custom palette/typography use
`define-a-theme`.

A complete, runnable program lives at
`examples/scaffold-a-presentation/main.go`.

## Binding properties you must know

- **P1 — Two layers, one library.** `pptx` is the builder; `scene` is the
  optional renderer that composes it. Stay in `pptx` for direct authoring;
  never reach under it into `internal/...`.
- **P2 — Tokens, not literals (the default).** Every visual property flows
  through the active `*Theme` via a semantic token: colors via
  `TokenColor(ColorRole)` / `TokenTextColor(TextColorRole)`, typography via
  `RunStyle.TypeRole`, radius via `WithRadius(RadiusRole)`, elevation via
  `WithElevation(ElevationRole)`. A theme swap re-renders the same builder
  input in the new palette. `RGB("2563EB")` / `RGBA` and `Pt`/`In`/`Cm`/`Px`
  are escape hatches — reach for them only when a literal is genuinely
  required.
- **P4 — stdlib-only runtime.** The library imports only the Go standard
  library and compiles `CGO_ENABLED=0`. Your code that uses it should do the
  same; do not pull third-party deps to drive the builder.

## Core API

All geometry is `pptx.Box{X, Y, W, H pptx.EMU}` measured from the slide's
top-left origin. Convert human units with `pptx.In/Cm/Pt/Px(float64) EMU`
(e.g. `pptx.In(0.8)`). Canvas constants: `pptx.Slide16x9Width/Height`,
`pptx.Slide4x3Width/Height`.

### Create a presentation

```go
func New(opts ...Option) *Presentation
```

Options (all `pptx.Option`):

```go
WithFormat(f Format)            // Slides16x9 (default) | Slides4x3
WithTheme(t *Theme)             // active theme (default DefaultTheme())
WithLogger(l *slog.Logger)      // structured events; nil = no logs
WithFontSource(src FontSource)  // resolves bytes for embedded fonts
WithReadPartLimit(n int64)      // read-only: per-part size ceiling (default 100 MB)
FromTemplate(brand *Presentation) // seed theme + masters + layouts from a brand deck
```

`New()` with no options is a 16:9 deck themed with `DefaultTheme()` and a valid
scaffold (master + layout + theme), ready for slides immediately.

### Add slides

```go
func (p *Presentation) AddSlide(layout ...string) *Slide
func (p *Presentation) AddSlideAt(index int, layout ...string) (*Slide, error)
func (p *Presentation) Slides() []*Slide
func (p *Presentation) SlideCount() int
func (p *Presentation) GetSlide(index int) (*Slide, error)
func (p *Presentation) RemoveSlide(index int) error
```

The optional `layout` name selects a named layout from the active template
(e.g. when seeded via `FromTemplate`); omit it for the default blank layout.

### Shapes

```go
func (s *Slide) AddShape(geom ShapeGeometry, box Box, opts ...ShapeOption) *Shape
```

Geometries (`ShapeGeometry`): `ShapeRect`, `ShapeRoundRect`, `ShapeEllipse`,
`ShapeTriangle`, `ShapeDiamond`, `ShapeParallelogram`, `ShapeHexagon`,
`ShapeChevron`, `ShapeRightArrow`, `ShapeLine`.

Options (`ShapeOption`):

```go
WithFill(f Fill)               // SolidFill(Color) | NoFill() | LinearGradient(...) | RadialGradient(...)
WithLine(l Line)               // outline; Line{Width EMU, Color Color, Dash string}
WithRadius(role RadiusRole)    // ShapeRoundRect only; RadiusNone/SM/MD/LG/Full
WithElevation(role ElevationRole) // drop shadow token: ElevationFlat/Raised/Elevated
WithShadow(e Elevation)        // literal-shadow escape hatch for WithElevation
WithRotation(deg float64)      // clockwise degrees, normalized to [0,360)
```

Fills and colors:

```go
func SolidFill(c Color) Fill
func NoFill() Fill
func LinearGradient(angleDeg float64, stops ...GradientStop) Fill
func RadialGradient(stops ...GradientStop) Fill

func TokenColor(role ColorRole) Color          // surface roles (P2 default)
func TokenColorAlpha(role ColorRole, alpha int) Color
func TokenTextColor(role TextColorRole) Color  // text roles (P2 default)
func RGB(hex string) RGB                        // literal escape hatch (also a Color)
func RGBA(hex RGB, alpha int) Color             // literal with OOXML alpha 0..100000
```

`ColorRole`: `ColorCanvas`, `ColorSurface`, `ColorSurfaceAlt`, `ColorAccent`,
`ColorAccentAlt`, `ColorAccentWarm`, `ColorSuccess`, `ColorWarning`,
`ColorError`, `ColorInfo`.

### Text

```go
func (s *Slide) AddTextFrame(box Box) *TextFrame
func (tf *TextFrame) AddParagraph(opts ParagraphOpts) *Paragraph
func (tf *TextFrame) AutoFit(mode AutoFitMode) *TextFrame   // AutoFitNone | AutoFitNormal | AutoFitShape
func (tf *TextFrame) Anchor(v TextAnchor) *TextFrame        // AnchorTop | AnchorMiddle | AnchorBottom
func (tf *TextFrame) Margins(top, right, bottom, left EMU) *TextFrame
func (tf *TextFrame) Clear() *TextFrame

func (p *Paragraph) AddRun(text string, style RunStyle) *Run
func (p *Paragraph) AddBreak()
func (p *Paragraph) Align(a Alignment) *Paragraph          // AlignLeft/Center/Right/Justify
func (p *Paragraph) Indent(level int) *Paragraph
func (p *Paragraph) Bullet(kind BulletKind) *Paragraph     // BulletNone/Disc/Number/Checkbox
```

`ParagraphOpts{Align Alignment; Level int; Bullet BulletKind}`.

`RunStyle` is token-typed:

```go
type RunStyle struct {
    TypeRole    TypeRole       // typography scale (size + family)
    Color       Color          // theme token (default path) or literal
    Bold, Italic bool
    Underline   Underline      // UnderlineNone/Single/Double
    Strike      Strike         // StrikeNone/Single/Double
    BaselineRel BaselineShift  // BaselineNone/Superscript/Subscript
    Code        bool           // inline code: monospace + subtle tint
}
```

`TypeRole`: `TypeDisplay`, `TypeH1`..`TypeH5`, `TypeBody`, `TypeBodySmall`,
`TypeCaption`, `TypeMono`, `TypeCode`. `TextColorRole`: `TextPrimary`,
`TextSecondary`, `TextTertiary`, `TextInverse`, `TextMuted`, `TextAccent`,
`TextAccentAlt`, `TextSuccess`, `TextWarning`, `TextError`.

### Tables

```go
func (s *Slide) AddTable(box Box, rows, cols int) *Table
func (t *Table) Cell(row, col int) *Cell
func (t *Table) SetHeaderRow(on bool) *Table
func (t *Table) SetBanding(rowBand, colBand bool) *Table
func (t *Table) SetColumnWidths(widths ...EMU) *Table

func (c *Cell) SetText(text string) *Cell           // single themed body run
func (c *Cell) TextFrame() *TextFrame               // full rich-text control
func (c *Cell) SetFill(f Fill) *Cell
func (c *Cell) SetBorders(line Line) *Cell
func (c *Cell) MergeRight(n int) *Cell
func (c *Cell) MergeDown(n int) *Cell
```

`rows`/`cols` are clamped to a minimum of 1; columns start equal-width.
`SetHeaderRow` / `SetBanding(rowBand, …)` emit concrete alternating cell fills
(visible without a table-style part).

### Images

```go
func (s *Slide) AddImage(src ImageSource, box Box) (*Image, error)

func ImageBytes(data []byte, mime string) ImageSource  // mime verified against bytes
func ImageFile(path string) ImageSource                // format sniffed from bytes
func ImageReader(r io.Reader, mime string) ImageSource

func (im *Image) SetAltText(text string) *Image
func (im *Image) SetCrop(c Crop) *Image                // per-edge fractions 0..1
func (im *Image) SetFit(f Fit) *Image                  // FitFill | FitNone
func (im *Image) SetRotation(deg float64) *Image
func (im *Image) SetOpacity(alpha int) *Image          // 0..100000
```

Identical bytes across the deck are written to the package once (dedup).
Recognized formats: PNG, JPEG, GIF, BMP, WebP — recognized by magic bytes, not
extension; mismatched/malformed bytes are rejected (`ErrUnknownImageFormat`,
`ErrImageMIMEMismatch`). Pixel data is never parsed (security §7).

### Speaker notes & sections

```go
func (s *Slide) SpeakerNotes() *TextFrame           // creates on first use; author like any frame
func (s *Slide) SetSpeakerNotes(text string)        // convenience: single plain paragraph
func (s *Slide) HasSpeakerNotes() bool

func (p *Presentation) AddSection(name string) *Section
func (p *Presentation) Sections() []*Section
func (sec *Section) Include(s *Slide)                // assign a slide (idempotent)
func (sec *Section) Name() string
```

Once any section exists, slides left unassigned fall into an implicit leading
"Default Section" so every slide is covered (PowerPoint requires it).

### Save

```go
func (p *Presentation) Save(path string) error
func (p *Presentation) Write(w io.Writer) error      // e.g. an HTTP response
func (p *Presentation) WriteToBytes() ([]byte, error)
func (p *Presentation) SaveStream(path string) error // streaming OPC writer
```

Every write path runs the always-on repair-prompt hygiene pass so the deck
opens in PowerPoint without a "needs repair" prompt.

## Reading a deck back

```go
func NewFromBytes(data []byte, opts ...Option) (*Presentation, error)
func NewFromFile(path string, opts ...Option) (*Presentation, error)
func OpenStream(path string, opts ...Option) (*Presentation, error) // lazy per-part read
```

Read-relevant options: `WithLogger` (each non-fatal degradation is logged at
Warn) and `WithReadPartLimit(n)` (per-part decompressed ceiling; default
100 MB; a larger part is rejected wrapping `opc.ErrPartTooLarge` — the memory
bound, D-049; `n <= 0` disables it).

Inspect via the same handles you author with:

```go
for _, s := range pres.Slides() {
    for _, sh := range s.Shapes() {
        sh.Geometry(); sh.Box(); sh.Fill(); sh.Line(); sh.Rotation()
        if tf, ok := sh.TextFrame(); ok { /* tf.Paragraphs() -> p.Runs() */ }
        if t, ok := sh.Table(); ok { /* t.RowCount(), t.Cell(r,c) */ }
        if im, ok := sh.Image(); ok { im.AltText(); im.Bytes() }
    }
    if s.HasSpeakerNotes() { /* s.SpeakerNotes().Paragraphs() */ }
}
```

Read accessors surface **resolved literals**, not the originating tokens: theme
tokens resolve to concrete sRGB / point sizes at write time (D-030, D-033), so
a reopened run reports its resolved font/size/color, not its `TypeRole`.

- **G6 — round-trip fidelity.** Every shape, run, fill, line, table, and image
  pptx-go emits round-trips losslessly through the read path, and such a deck
  reports **no** warnings.
- **D-048 — best-effort external read.** Opening a deck pptx-go did *not*
  author is best-effort: unrecognized content is dropped at parse rather than
  failing the open, and each degradation is reported.

```go
for _, w := range pres.ReadWarnings() { // []ReadWarning, stable order; nil for self-authored decks
    // w.Kind (WarnDroppedElement | WarnUnreadablePart), w.Part, w.Element, w.Detail
}
```

## Complete example

The full program is `examples/scaffold-a-presentation/main.go` (run it with
`go run ./examples/scaffold-a-presentation`). The load-bearing core:

```go
pres := pptx.New(pptx.WithFormat(pptx.Slides16x9))

title := pres.AddSlide()
title.AddTextFrame(pptx.Box{X: pptx.In(0.8), Y: pptx.In(0.7), W: pptx.In(11), H: pptx.In(1.4)}).
    AddParagraph(pptx.ParagraphOpts{}).
    AddRun("Quarterly Review", pptx.RunStyle{
        TypeRole: pptx.TypeDisplay,
        Color:    pptx.TokenTextColor(pptx.TextPrimary),
        Bold:     true,
    })

title.AddShape(pptx.ShapeRoundRect,
    pptx.Box{X: pptx.In(0.8), Y: pptx.In(3.1), W: pptx.In(4), H: pptx.In(0.18)},
    pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))),
    pptx.WithRadius(pptx.RadiusFull),
    pptx.WithElevation(pptx.ElevationRaised),
)
title.SetSpeakerNotes("Open with the headline number, then walk the table.")

data := pres.AddSlide()
tbl := data.AddTable(pptx.Box{X: pptx.In(0.8), Y: pptx.In(1.8), W: pptx.In(7), H: pptx.In(2.5)}, 3, 2)
tbl.SetHeaderRow(true)
tbl.SetBanding(true, false)
tbl.Cell(0, 0).SetText("Region")
tbl.Cell(0, 1).SetText("Revenue")

sec := pres.AddSection("Overview")
for _, s := range pres.Slides() {
    sec.Include(s)
}

bytesOut, err := pres.WriteToBytes()
// ... reopen:
reopened, err := pptx.NewFromBytes(bytesOut)
_ = reopened.Slides()        // 2
_ = reopened.ReadWarnings()  // nil — self-authored decks round-trip losslessly
```

## What the caller owns

pptx-go is the engine, not the product (D-026). It converts your builder calls
into valid OOXML and nothing more. **Content and layout decisions are yours:**
what the slides say, where shapes sit on the canvas, which palette is on-brand,
how data maps to a table, whether an image is appropriate. There is no
deck-wide "render mode", no legibility heuristic, no auto-layout in Layer 1 —
position every `Box` yourself (or move up to `scene` / `compose-a-scene`, whose
layout engine does the placement).

## See also

- `define-a-theme` — build a custom `*Theme` (palette, typography, radii,
  elevations) so your tokens render in a brand language.
- `compose-a-scene` — the Layer 2 typed scene IR with an automatic layout
  engine, when you want structured content instead of manual `Box` placement.
- `load-a-brand-template` — seed a deck's theme + masters from an existing
  `.pptx` via `FromTemplate`.
