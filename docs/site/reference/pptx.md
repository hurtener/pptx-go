# `pptx` package reference

The `pptx` package is Layer 1: a general-purpose, theme-aware PowerPoint
builder. It is the primary entry point for both human developers and AI
callers. This page lists the exported surface, organized by area, with the
exact Go signature and a one-line note for each symbol.

For a narrative introduction see the [builder guide](/guide/builder); for
the design rationale behind the choices here see the
[decisions reference](/reference/decisions).

## Presentation lifecycle

A `Presentation` is the top-level deck facade. Create one, add slides,
then save or stream it out.

### Constructors and options

```go
func New(opts ...Option) *Presentation
```
Creates an empty presentation, configured by zero or more `Option`s.

```go
func NewFromBytes(data []byte, opts ...Option) (*Presentation, error)
func NewFromFile(path string, opts ...Option) (*Presentation, error)
func OpenStream(path string, opts ...Option) (*Presentation, error)
```
Open an existing deck: from an in-memory byte slice, from a file path, or
lazily via the streaming path. See the [reading guide](/guide/reading).

```go
func NewWithTemplate(name TemplateType) (*Presentation, error)
```
Create a deck seeded from a registered starter template.

```go
type Option func(*Presentation)

func WithFormat(f Format) Option        // slide canvas format (16:9, 4:3)
func WithTheme(t *Theme) Option         // active theme
func WithLogger(l *slog.Logger) Option  // structured logging (no global logger)
func WithFontSource(src FontSource) Option // font byte resolver for EmbedFont
func WithFontEmbedding() Option          // auto-embed every used face at save (needs a FontSource)
func WithReadPartLimit(n int64) Option  // per-part read ceiling (n <= 0 = unlimited)
func FromTemplate(brand *Presentation) Option // adopt a brand kit (clone + strip slides)
```
`FromTemplate` clones a brand deck's OPC package and strips its slides,
preserving theme, masters, layouts, and rels (D-037).
`WithReadPartLimit` maps to the `internal/opc` per-part memory bound
(D-049).

### Saving and serializing

```go
func (p *Presentation) Save(path string) error
func (p *Presentation) Write(w io.Writer) error
func (p *Presentation) WriteToBytes() ([]byte, error)
func (p *Presentation) SaveStream(path string) error
```
Persist the deck: to a file, to any `io.Writer`, to a byte slice, or via
the streaming write path.

### Slides and sections

```go
func (p *Presentation) AddSlide(layout ...string) *Slide
func (p *Presentation) AddSlideAt(index int, layout ...string) (*Slide, error)
func (p *Presentation) GetSlide(index int) (*Slide, error)
func (p *Presentation) RemoveSlide(index int) error
func (p *Presentation) Slides() []*Slide
func (p *Presentation) SlideCount() int
func (p *Presentation) AddSection(name string) *Section
func (p *Presentation) Sections() []*Section
```
Add, fetch, remove, and enumerate slides; group slides into named
sections.

### Theme, layouts, and metadata

```go
func (p *Presentation) Theme() *Theme
func (p *Presentation) SetTheme(t *Theme)
func (p *Presentation) Masters() []*Master
func (p *Presentation) HasLayout(name string) bool
func (p *Presentation) SetMetadata(m Metadata)
func (p *Presentation) SetSlideSize(cx, cy int)
func (p *Presentation) SetSlideSizeStandard(name string)
func (p *Presentation) SlideSize() (int, int)
```
Read/replace the active theme, enumerate read-only masters, set core
document metadata, and control the slide canvas size.

### Fonts, reading, and lifecycle

```go
func (p *Presentation) EmbedFont(name, style string, weight int) error
func (p *Presentation) SetFontSource(src FontSource)
func (p *Presentation) ReadWarnings() []ReadWarning
func (p *Presentation) Clone() (*Presentation, error)
func (p *Presentation) Close() error
```
`EmbedFont` embeds a single font face resolved via the configured `FontSource`.
For a deck themed with a brand display/heading face, prefer
`WithFontEmbedding()`: at save it walks every slide's runs, collects the distinct
used faces, and `EmbedFont`s each via the registered `FontSource` — a no-op
without a source, warn-don't-fail on a missing face, idempotent against manual
`EmbedFont`, and byte-identical when off.
`ReadWarnings` returns the non-fatal degradations noted while opening a
(third-party) deck (D-048). `Clone` deep-copies the deck.

## Slide

A `Slide` is the high-level slide object returned by `AddSlide`. Its
builder methods place content; its read accessors inspect a reopened
slide.

```go
func (s *Slide) AddShape(geom ShapeGeometry, box Box, opts ...ShapeOption) *Shape
func (s *Slide) AddTextFrame(box Box) *TextFrame
func (s *Slide) AddTable(box Box, rows, cols int) *Table
func (s *Slide) AddImage(src ImageSource, box Box) (*Image, error)
func (s *Slide) AddIcon(svg []byte, box Box, opts ...ShapeOption) (*Shape, error)
func (s *Slide) ChartPlaceholder(box Box, opts ...ShapeOption) *Shape
func (s *Slide) Shapes() []*Shape
```
Add a preset-geometry shape, a rich-text frame, a table, an image, an icon
(SVG translated to native geometry), or a chart placeholder; `Shapes`
enumerates the slide's shapes (including those recovered on open, RFC §16).

```go
func (s *Slide) SetSpeakerNotes(text string)
func (s *Slide) SpeakerNotes() *TextFrame
func (s *Slide) HasSpeakerNotes() bool
func (s *Slide) Index() int
func (s *Slide) Layout() string
func (s *Slide) SetLayout(layoutName string) bool
```
Set/read speaker notes (round-trip on open, D-022/D-050) and inspect the
slide's index and layout.

## Shapes and geometry

### Shape handle

```go
type Shape struct { /* opaque */ }

func (sh *Shape) Box() Box                    // position + size
func (sh *Shape) Geometry() ShapeGeometry     // preset geometry
func (sh *Shape) Rotation() float64           // rotation in degrees
func (sh *Shape) Fill() Fill                  // interior fill
func (sh *Shape) Line() Line                  // outline
func (sh *Shape) Shadow() (Elevation, bool)   // shadow, if set
func (sh *Shape) TextFrame() (*TextFrame, bool) // text body, if any
func (sh *Shape) Table() (*Table, bool)       // table, if a graphic frame
func (sh *Shape) Image() (*Image, bool)       // picture, if a pic shape
```
A `Shape` is a handle to a shape — one the builder added or one recovered
by `Open`. Read accessors map recovered wire fields back to the public
types (no raw OOXML, P3).

### Shape options and geometries

```go
type ShapeOption func(*shapeConfig)

func WithFill(f Fill) ShapeOption
func WithLine(l Line) ShapeOption
func WithRadius(role RadiusRole) ShapeOption
func WithRotation(deg float64) ShapeOption
func WithElevation(role ElevationRole) ShapeOption
func WithShadow(e Elevation) ShapeOption
func WithImageFill(src ImageSource) ShapeOption  // cover-fit photo surface fill (a:blipFill, D-117)
```
Configure a shape at creation: fill, outline, corner radius token,
rotation, elevation token, an explicit shadow, or a cover-fit image
surface fill (the photo is center-cropped to cover the shape; it wins
over `WithFill` and the geometry still clips it).

```go
type ShapeGeometry string

const (
	ShapeRect          ShapeGeometry = "rect"
	ShapeRoundRect     ShapeGeometry = "roundRect"
	ShapeEllipse       ShapeGeometry = "ellipse"
	ShapeTriangle      ShapeGeometry = "triangle"
	ShapeDiamond       ShapeGeometry = "diamond"
	ShapeParallelogram ShapeGeometry = "parallelogram"
	ShapeHexagon       ShapeGeometry = "hexagon"
	ShapeChevron       ShapeGeometry = "chevron"
	ShapeRightArrow    ShapeGeometry = "rightArrow"
	ShapeLine          ShapeGeometry = "line"
)
```
A curated subset of the OOXML preset geometries.

### Geometry value types

```go
type Box struct{ X, Y, W, H EMU }       // positioned rectangle (offset + extent)
type Position struct{ X, Y EMU }        // a point on the canvas
type Size struct{ W, H EMU }            // a width/height extent
type Inset struct{ Top, Right, Bottom, Left EMU } // per-edge padding

func UniformInset(v EMU) Inset          // equal padding on all edges
func (b Box) Position() Position
func (b Box) Size() Size
func (b Box) Right() EMU
func (b Box) Bottom() EMU
func (b Box) Inset(in Inset) Box        // shrink a box by an inset

type Anchor int
const (
	AnchorTopLeft Anchor = iota
	AnchorTopCenter
	AnchorTopRight
	AnchorCenterLeft
	AnchorCenter
	AnchorCenterRight
	AnchorBottomLeft
	AnchorBottomCenter
	AnchorBottomRight
)
func (a Anchor) Point(b Box) Position    // EMU coordinate of an anchor on a box
```

### Units (EMU)

```go
type EMU int64

const (
	Slide16x9Width  EMU = 12192000
	Slide16x9Height EMU = 6858000
	Slide4x3Width   EMU = 9144000
	Slide4x3Height  EMU = 6858000
)

func In(in float64) EMU   // inches -> EMU
func Cm(cm float64) EMU   // centimeters -> EMU
func Pt(pt float64) EMU   // points -> EMU (1 pt = 1/72 in)
func Px(px float64) EMU   // pixels -> EMU at the 96-DPI reference

func (e EMU) Inches() float64
func (e EMU) Centimeters() float64
func (e EMU) Points() float64
func (e EMU) Pixels() float64

func EMUToPx(emu int) int  // EMU -> pixels (int)
func PxToEMU(px int) int   // pixels -> EMU (int)
```
All geometry is EMU; the constructors convert from human units at the call
site.

## Color and fill

```go
type Color interface { /* sealed */ }

type RGB string // 6-hex-digit string without '#', e.g. "2563EB"

func RGBA(hex RGB, alpha int) Color                  // literal color + alpha
func TokenColor(role ColorRole) Color                // surface-color token
func TokenColorAlpha(role ColorRole, alpha int) Color // surface token + alpha
func TokenTextColor(role TextColorRole) Color        // text-color token
```
`Color` is a sealed interface resolved at write time. `RGB` itself
implements it, so `pptx.RGB("2563EB")` is both a value and a literal color.
Tokens resolve against the active theme (P2; D-033). The literal alpha
helpers use the `AlphaOpaque = 100000` scale.

```go
type Fill interface {
	Kind() FillKind
	SolidColor() (Color, bool)
	Gradient() (GradientRead, bool)
}

func SolidFill(c Color) Fill
func NoFill() Fill
func LinearGradient(angleDeg float64, stops ...GradientStop) Fill
func RadialGradient(stops ...GradientStop) Fill

type FillKind int
const (
	FillSolid FillKind = iota
	FillNone
	FillGradient
)

type GradientStop struct {
	Pos   float64 // 0..1 along the gradient
	Color Color
}
```
`Fill` is a shape's interior. Construct with the four constructors; the
read accessors inspect a fill recovered from a reopened deck (D-041).

```go
type Line struct {
	Width EMU    // stroke width
	Color Color  // stroke color; nil leaves it unset
	Dash  string // preset dash ("dash", "dot", "sysDash", …); empty = solid
}
```
A shape's outline; a zero `Line` leaves the outline unset.

## Rich text

```go
type TextFrame struct { /* opaque */ }

func (tf *TextFrame) AddParagraph(opts ParagraphOpts) *Paragraph
func (tf *TextFrame) Paragraphs() []*Paragraph
func (tf *TextFrame) Clear() *TextFrame
func (tf *TextFrame) AutoFit(mode AutoFitMode) *TextFrame      // setter (chainable)
func (tf *TextFrame) AutoFitMode() AutoFitMode                 // reader
func (tf *TextFrame) Anchor(v TextAnchor) *TextFrame           // vertical anchor setter
func (tf *TextFrame) VerticalAnchor() TextAnchor               // reader
func (tf *TextFrame) Margins(top, right, bottom, left EMU) *TextFrame // setter
func (tf *TextFrame) MarginInsets() (top, right, bottom, left EMU)    // reader
```
A shape-level rich-text container. Setters chain; the parallel readers
inspect a reopened frame.

```go
type Paragraph struct { /* opaque */ }

func (p *Paragraph) AddRun(text string, style RunStyle) *Run
func (p *Paragraph) AddHyperlink(text, target string, style RunStyle) *Run
func (p *Paragraph) AddBreak()
func (p *Paragraph) Align(a Alignment) *Paragraph
func (p *Paragraph) Alignment() Alignment
func (p *Paragraph) Bullet(kind BulletKind) *Paragraph
func (p *Paragraph) BulletStyle() BulletKind
func (p *Paragraph) BulletIndent() EMU
func (p *Paragraph) Indent(level int) *Paragraph
func (p *Paragraph) Level() int
func (p *Paragraph) LineHeight() float64
func (p *Paragraph) Runs() []*Run
```

```go
type Run struct { /* opaque */ }

func (r *Run) Text() string
func (r *Run) Bold() bool
func (r *Run) Italic() bool
func (r *Run) Underline() Underline
func (r *Run) Strike() Strike
func (r *Run) Baseline() BaselineShift
func (r *Run) Code() bool
func (r *Run) Color() (Color, bool)
func (r *Run) Font() string
func (r *Run) FontSize() float64
func (r *Run) Hyperlink() (string, bool)
```
`Run` read accessors over a styled span.

```go
type RunStyle struct {
	TypeRole    TypeRole
	Color       Color
	Bold        bool
	Italic      bool
	Underline   Underline
	Strike      Strike
	BaselineRel BaselineShift
	Code        bool      // inline code: mono + subtle tint (D-013)
	Tracking    *float64  // per-run letter-spacing override, pt; nil inherits the role (D-060)
	Case        *TextCase // per-run case override; nil inherits the role (D-062)
	FontScale   float64   // per-run multiplier on the role size; 0/1 = unchanged (D-074)
}

type ParagraphOpts struct {
	Align        Alignment
	Level        int
	Bullet       BulletKind
	LineHeight   float64 // line spacing as a percent of single; 0/100 = unchanged (D-061)
	BulletIndent EMU     // bullet hanging indent (marker-to-text offset); 0 = default 0.5" (D-078)
}
```

Each added field is additive: its zero value emits nothing and reproduces the
prior output byte-for-byte.

### Text enums

```go
type AutoFitMode int
const (
	AutoFitNone   AutoFitMode = iota // fixed sizes (may overflow)
	AutoFitNormal                    // shrink font to fit shape
	AutoFitShape                     // grow shape to fit text
)

type TextAnchor int
const ( AnchorTop TextAnchor = iota; AnchorMiddle; AnchorBottom )

type Alignment int
const ( AlignLeft Alignment = iota; AlignCenter; AlignRight; AlignJustify )

type BulletKind int
const ( BulletNone BulletKind = iota; BulletDisc; BulletNumber; BulletCheckbox )

type Underline int
const ( UnderlineNone Underline = iota; UnderlineSingle; UnderlineDouble )

type Strike int
const ( StrikeNone Strike = iota; StrikeSingle; StrikeDouble )

type BaselineShift int
const ( BaselineNone BaselineShift = iota; Superscript; Subscript )
```

## Table

```go
type Table struct { /* opaque */ }

func (t *Table) Cell(row, col int) *Cell
func (t *Table) RowCount() int
func (t *Table) ColCount() int
func (t *Table) SetColumnWidths(widths ...EMU) *Table
func (t *Table) ColumnWidths() []EMU
func (t *Table) SetHeaderRow(on bool) *Table
func (t *Table) HeaderRow() bool
func (t *Table) SetBanding(rowBand, colBand bool) *Table
func (t *Table) RowBanding() bool
```

```go
type Cell struct { /* opaque */ }

func (c *Cell) SetText(text string) *Cell
func (c *Cell) TextFrame() *TextFrame
func (c *Cell) SetFill(f Fill) *Cell
func (c *Cell) Fill() Fill
func (c *Cell) SetBorders(line Line) *Cell
func (c *Cell) MergeRight(n int) *Cell
func (c *Cell) MergeDown(n int) *Cell
func (c *Cell) GridSpan() int   // horizontal span (reader)
func (c *Cell) RowSpan() int    // vertical span (reader)
func (c *Cell) Covered() bool   // true if covered by a merge
```

## Media

```go
type ImageSource interface { /* sealed */ }

func ImageFile(path string) ImageSource
func ImageBytes(data []byte, mime string) ImageSource
func ImageReader(r io.Reader, mime string) ImageSource
```
Sealed image input for `AddImage`; new backends land behind the same seam.

```go
type Image struct { /* opaque */ }

func (im *Image) SetAltText(text string) *Image
func (im *Image) AltText() string
func (im *Image) SetCrop(c Crop) *Image
func (im *Image) Crop() Crop
func (im *Image) SetFit(f Fit) *Image
func (im *Image) Fit() Fit
func (im *Image) SetRotation(deg float64) *Image
func (im *Image) Rotation() float64
func (im *Image) SetOpacity(alpha int) *Image
func (im *Image) Opacity() int
func (im *Image) SetCornerRadius(role RadiusRole) *Image  // roundRect-clip the picture (D-114)
func (im *Image) SetElevation(role ElevationRole) *Image  // soft drop shadow on the picture (D-114)
func (im *Image) SetDuotone(shadow, highlight Color) *Image  // two-tone recolor via <a:duotone> (D-116)
func (im *Image) Duotone() (shadow, highlight RGB, ok bool)  // read inverse of SetDuotone
func (im *Image) Bytes() ([]byte, error)
```
Typed mutators and read accessors over an inserted image (P3).

```go
type Crop struct{ Left, Top, Right, Bottom float64 } // per-edge fraction (0..1)

type Fit int
const ( FitFill Fit = iota; FitNone )
```

## Theme and tokens

```go
type Theme struct {
	Name        string
	HeadingFont string
	BodyFont    string
	Colors      ColorPalette
	Typography  Typography
	Spacing     Spacing
	Radii       Radii
	Elevations  Elevations
}

func DefaultTheme() *Theme
func NewTheme(opts ...ThemeOption) *Theme
func LoadTheme(path string) (*Theme, error)
func LoadThemeFromBytes(data []byte) (*Theme, error)
func (t *Theme) Clone() *Theme
func (t *Theme) ResolveColor(role ColorRole) RGB
func (t *Theme) ResolveTextColor(role TextColorRole) RGB
func (t *Theme) ResolveType(role TypeRole) FontSpec
func (t *Theme) ResolveSpace(role SpaceRole) EMU
func (t *Theme) ResolveRadius(role RadiusRole) EMU
func (t *Theme) ResolveElevation(role ElevationRole) Elevation
func (t *Theme) ThemeXML() ([]byte, error)
```
The semantic visual contract. Tokens resolve at apply time; theme1.xml
token emission is a follow-up (D-033). See the [theme guide](/guide/theme).

```go
type ThemeOption func(*Theme)

func WithName(name string) ThemeOption
func WithAccent(c RGB) ThemeOption
func WithPaper(c RGB) ThemeOption    // off-white "paper" canvas tint (ColorPaper, D-104)
func WithFonts(heading, body string) ThemeOption
```

### Role enums

```go
type ColorRole int      // surface colors
const (
	ColorCanvas ColorRole = iota
	ColorSurface
	ColorSurfaceAlt
	ColorAccent
	ColorAccentAlt
	ColorAccentWarm
	ColorSuccess
	ColorWarning
	ColorError
	ColorInfo
	ColorPaper // off-white "paper" canvas; defaults to ColorCanvas, set via WithPaper (D-104)
)

type TextColorRole int  // text colors
const (
	TextPrimary TextColorRole = iota
	TextSecondary
	TextTertiary
	TextInverse
	TextMuted
	TextAccent
	TextAccentAlt
	TextSuccess
	TextWarning
	TextError
)

type TypeRole int       // typography scale
const (
	TypeDisplay TypeRole = iota
	TypeH1
	TypeH2
	TypeH3
	TypeH4
	TypeH5
	TypeBody
	TypeBodySmall
	TypeCaption
	TypeMono
	TypeCode
)
```
`SpaceRole`, `RadiusRole`, and `ElevationRole` follow the same pattern for
spacing, corner radius, and elevation tokens (see the
[theme guide](/guide/theme)).

```go
type FontSpec struct {
	Family string
	Size   float64 // points
	Weight int     // 100–900, 400 = regular, 700 = bold
	Italic bool
}
func (f FontSpec) Bold() bool
```

```go
type Metadata struct {
	Title   string
	Author  string // dc:creator
	Subject string
}
```

## Templates

```go
func FromTemplate(brand *Presentation) Option // (see Presentation lifecycle)

type Master struct { /* opaque, read-only */ }
func (m *Master) Name() string
func (m *Master) Layouts() []*Layout
func (m *Master) Layout(name string) (*Layout, bool)

type Layout struct { /* opaque, read-only */ }
func (l *Layout) Name() string
```
A `Master` is a read-only view of one slide master; a `Layout` is one of
its layouts. See the [reading guide](/guide/reading).

## Sections

```go
type Section struct { /* opaque */ }
func (sec *Section) Name() string
func (sec *Section) Include(s *Slide)
```
A named, ordered grouping of slides. Create with
`Presentation.AddSection` and assign with `Include`.

## Read warnings

```go
type ReadWarning struct {
	Kind    ReadWarningKind
	Part    string // part URI, e.g. "/ppt/slides/slide2.xml"
	Element string // element local-name (WarnDroppedElement); else empty
	Detail  string // human-readable context
}

type ReadWarningKind int
const (
	WarnDroppedElement ReadWarningKind = iota // unrecognized element ignored at parse
	WarnUnreadablePart                        // referenced part missing/unparseable, skipped
)
func (k ReadWarningKind) String() string
```
One non-fatal degradation noted while reading a (third-party) deck,
returned by `Presentation.ReadWarnings()` (D-048). See the
[reading guide](/guide/reading).

## Sentinel errors

```go
var ErrUnknownImageFormat = errors.New("pptx: unrecognized or malformed image data")
var ErrImageMIMEMismatch  = errors.New("pptx: declared image MIME does not match content")
var ErrImagePartMissing   = errors.New("pptx: image media part not found")
var ErrNoFontSource       = errors.New("pptx: no font source registered (use SetFontSource)")
var ErrFontNotFound       = errors.New("pptx: font not found")
var ErrThemeNotFound      = errors.New("pptx: no theme part in package")
```
Match these with `errors.Is`. On the read path, a part exceeding the
configured per-part limit surfaces `opc.ErrPartTooLarge`
(`internal/opc`, D-049) — match it via `errors.Is`.
