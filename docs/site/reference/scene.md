# `scene` package reference

The `scene` package is Layer 2: a typed scene IR and a `Render` entry
point that composes the `pptx` builder (Layer 1). A caller builds a
`Scene` — an ordered list of `SceneSlide`s, each a list of typed
`SlideNode`s — and `Render` turns it into a `*pptx.Presentation`.
`scene` composes `pptx`; it never reaches under the builder (P1). Its
token enums are aliases of `pptx`'s, so callers use one vocabulary.

For a walkthrough see the [scene guide](/guide/scene); for per-node
detail see the [scene catalog](/catalog/).

## Scene structure

```go
type Scene struct {
	Theme  *pptx.Theme // optional; the builder's default theme if nil
	Slides []SceneSlide
	Meta   Metadata
}

type SceneSlide struct {
	ID         string
	Layout     LayoutKind
	Nodes      []SlideNode
	Notes      RichText
	Variant    Variant
	Background  Background // full-bleed slide background; zero value draws nothing
}

type Background struct {
	Kind     BackgroundKind   // None | Color | Gradient | Asset | Radial
	Color    pptx.ColorRole   // solid fill (Kind == BackgroundColor)
	Gradient [2]pptx.ColorRole // legacy 2-role linear gradient (used when Stops empty)
	Stops    []GradientStop   // multi-stop gradient: 2..8 ascending stops in [0,1]; supersedes Gradient
	Angle    int              // linear gradient angle, degrees clockwise from +x
	AssetID  AssetID          // full-bleed picture (Kind == BackgroundAsset)
}

type GradientStop struct {
	Pos   float64        // [0,1]
	Color pptx.ColorRole
}

type Metadata struct {
	Title   string
	Author  string
	Subject string
}
```
A `Scene` is the input to `Render`; a `SceneSlide` is one slide: a layout
intent, the top-level node list, optional speaker notes, a theme variant, and
an optional full-bleed `Background`. A zero `Background` (`BackgroundNone`)
draws nothing. `BackgroundGradient` takes either the legacy two-role `Gradient`
pair or, for a richer multi-hue wash, a `Stops` list of 2–8 ascending
`GradientStop`s in `[0,1]` (the `Stops` list supersedes `Gradient` when set;
invalid stops degrade to a warning and skip the fill — see the decisions
reference, D-105). `BackgroundRadial` (D-106) draws the same stops as a
center-out radial spotlight/vignette (centered focal). Point a `BackgroundColor`
at `ColorPaper` (D-104) for a tinted off-white paper canvas.

```go
type LayoutKind int
const (
	LayoutCover LayoutKind = iota
	LayoutTitleContent
	LayoutTwoColumn
	LayoutCardGrid
	LayoutFullBleed
	LayoutBlank
)

type Variant int
const (
	VariantLight Variant = iota
	VariantDark
	VariantPrint
)
func (v Variant) String() string
```
`LayoutKind` names a slide's structural intent (maps to a master layout at
render time). `Variant` selects a theme variant; a non-default variant
currently renders with the active theme and surfaces a `LayoutWarning`
(not silently dropped).

## Render

```go
func Render(pres *pptx.Presentation, s Scene, opts ...RenderOption) (Stats, error)
```
Renders a `Scene` into a presentation, returning `Stats`.

```go
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []LayoutWarning
	Timings  []SlideTiming
}

type LayoutWarning struct {
	SlideID string
	Node    string
	Message string
}

type SlideTiming struct {
	SlideID  string
	Duration time.Duration
}
```
`Stats` is the library's observability surface — counts, per-slide
timings, and non-fatal warnings; there is no event protocol (D-016). A
caller that wants warnings to be fatal inspects `Stats.Warnings` itself —
there is no strict mode.

### Render options

```go
type RenderOption func(*renderConfig)

func WithTheme(t *pptx.Theme) RenderOption          // override the scene/builder theme
func WithLayoutMap(m LayoutMap) RenderOption        // LayoutKind -> master layout name
func WithLogger(l *slog.Logger) RenderOption        // structured logging
func WithContext(ctx context.Context) RenderOption  // cancellation
func WithWorkers(n int) RenderOption                // parallel-compose worker count
func WithAssetResolver(r AssetResolver) RenderOption // resolve AssetIDs to bytes
func WithIconExtension(name string, svg []byte) RenderOption       // add a caller icon
func WithFrameExtension(name string, recipe FrameRecipe) RenderOption    // add a caller frame
func WithOrnamentExtension(name string, recipe OrnamentRecipe) RenderOption // add a caller ornament
```

```go
type LayoutMap map[LayoutKind]string
func DefaultLayoutMap() LayoutMap
```
Maps each `LayoutKind` to a master layout name; the default covers the
shipped layouts.

## Slide nodes

```go
type SlideNode interface {
	NodeKind() NodeKind
	// sealed: isSlideNode is unexported
}

type NodeKind int
const (
	KindHero NodeKind = iota
	KindProse
	KindHeading
	KindList
	KindDivider
	KindQuote
	KindCallout
	KindImage
	KindChip
	KindArrow
	KindCodeBlock
	KindChart
	KindTable
	KindFlow
	KindDecoration
	KindSectionDivider
	KindTwoColumn
	KindGrid
	KindCard
	KindCardSection
)
func (k NodeKind) String() string
```
`SlideNode` is the sealed scene IR union; the concrete node types are
closed. The shipped node structs (`Hero`, `Prose`, `Heading`, `List`,
`Divider`, `Quote`, `Callout`, `Image`, `Chip`, `Arrow`, `CodeBlock`,
`Chart`, `Table`, `Flow`, `Decoration`, `SectionDivider`, `TwoColumn`,
`Grid`, `Card`, `CardSection`) each map to a `NodeKind`. See the catalog
for per-node fields:
[text leaves](/catalog/text-leaves),
[visual leaves](/catalog/visual-leaves),
[asset leaves](/catalog/asset-leaves), and
[containers](/catalog/containers).

## Rich text

```go
type RichText []TextRun

type TextRun struct {
	Text  string
	Style RunStyle
	Color TextColor
}

type RunStyle struct {
	TypeRole  TypeRole // typography scale
	Bold      bool
	Italic    bool
	Underline bool
	Strike    bool
	Code      bool   // inline code: mono + tint (D-013)
	Link      bool   // marks the run as a hyperlink
	Href      string // hyperlink target when Link is set
}
```

```go
type TextColor struct { /* opaque; zero value = token TextPrimary */ }

func TokenTextColor(role TextColorRole) TextColor // theme-bound (default path)
func LiteralColor(hex string) TextColor           // literal RGB (escape hatch)
func (c TextColor) IsLiteral() bool
func (c TextColor) Literal() pptx.RGB
func (c TextColor) Role() TextColorRole
```
`TextColor` is a run color: a `TextColorRole` token (the default path) or a
literal RGB.

## Token re-exports

```go
type ColorRole     = pptx.ColorRole
type TextColorRole = pptx.TextColorRole
type TypeRole      = pptx.TypeRole
type SpaceRole     = pptx.SpaceRole
type RadiusRole    = pptx.RadiusRole
type ElevationRole = pptx.ElevationRole
type Anchor        = pptx.Anchor
type Position      = pptx.Position
type Size          = pptx.Size
type Crop          = pptx.Crop
type Fit           = pptx.Fit
```
The token enums and a few geometry types are type aliases of `pptx`'s, plus
matching constant re-exports (`ColorCanvas`, `TextPrimary`, `TypeDisplay`,
`SpaceXS`, `RadiusNone`, `ElevationFlat`, `AnchorTopLeft`, `FitFill`, …) —
so callers use one vocabulary across both layers.

## Validation

```go
func ValidateScene(s Scene) error
```
Runs structural validation over a `Scene` (Stage 1) before render.

## Assets

```go
type AssetID string

type AssetResolver interface {
	Resolve(ctx context.Context, id AssetID) ([]byte, string, error)
}

func URIAssetResolver(fn func(uuid string) ([]byte, string, error)) AssetResolver

var ErrAssetNotFound = errors.New("scene: asset not found")
```
`AssetID` is a free-form reference (pptx-go imposes no scheme, D-024). An
`AssetResolver` maps it to bytes and a content-type hint; a missing asset
returns `(nil, "", ErrAssetNotFound)`. Every unresolved asset degrades to
a `LayoutWarning` and the node is skipped (D-036). `URIAssetResolver`
adapts an `asset://<uuid>` helper function. See the
[assets guide](/guide/assets).

## Curated registries and extension

```go
type FrameRecipe    = frames.Recipe    // func(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int)
type OrnamentRecipe = ornaments.Recipe // func(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64, role pptx.ColorRole) int

func ValidateIcon(svg []byte) error
```
A `FrameRecipe` draws a device bezel and returns the interior box; an
`OrnamentRecipe` draws an ornament at an opacity, rotation, and color role
(`Decoration.Color`, default `ColorAccent` — D-107). Both compose
the public `pptx` builder only (P1). `ValidateIcon` checks caller SVG bytes
against the translator constraints (single path, solid fill, no gradients)
at registration time. Register caller assets with `WithIconExtension`,
`WithFrameExtension`, and `WithOrnamentExtension` (see above).

### Curated names

```go
// scene/icons — curated icon names (assets/icons):
"arrow-down", "arrow-left", "arrow-right", "arrow-up", "check",
"chevron-down", "chevron-left", "chevron-right", "chevron-up",
"circle", "diamond", "dot", "minus", "plus", "square", "star",
"triangle", "x"

// scene/ornaments — curated ornament names:
const (
	NameGlowRing      = "glow_ring"
	NameRadialGlow    = "radial_glow"
	NameGridDots      = "grid_dots"
	NameCornerBracket = "corner_bracket"
	NameChevronArrow  = "chevron_arrow"
	NameNoiseOverlay  = "noise_overlay"
)

// scene/frames — curated device-frame names:
const (
	NameBrowser = "browser"
	NamePhone   = "phone"
	NameDesktop = "desktop"
	NameLaptop  = "laptop"
)
```
The reserved curated names; each registry is a closed name set plus caller
extension (RFC §14). See the [assets guide](/guide/assets).
