package scene

import "github.com/hurtener/pptx-go/pptx"

// The scene node catalog (RFC §11). SlideNode is a sealed union: an unexported
// marker keeps the set closed to this package, and NodeKind discriminates for
// validation and (later) rendering. Leaf nodes hold content; container nodes
// hold child SlideNodes and introduce sub-layouts.
//
// Exactly the nodes that render as an image (RFC §12) carry an AssetID field:
// Image, Chart, CodeBlock, and asset-kind Decoration. policy.go encodes that
// mapping and policy_test.go asserts it against the structs.

// NodeKind discriminates a SlideNode.
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
	KindBento
	KindStat
)

// String returns the node kind's IR name.
func (k NodeKind) String() string {
	switch k {
	case KindHero:
		return "hero"
	case KindProse:
		return "prose"
	case KindHeading:
		return "heading"
	case KindList:
		return "list"
	case KindDivider:
		return "divider"
	case KindQuote:
		return "quote"
	case KindCallout:
		return "callout"
	case KindImage:
		return "image"
	case KindChip:
		return "chip"
	case KindArrow:
		return "arrow"
	case KindCodeBlock:
		return "code_block"
	case KindChart:
		return "chart"
	case KindTable:
		return "table"
	case KindFlow:
		return "flow"
	case KindDecoration:
		return "decoration"
	case KindSectionDivider:
		return "section_divider"
	case KindTwoColumn:
		return "two_column"
	case KindGrid:
		return "grid"
	case KindCard:
		return "card"
	case KindCardSection:
		return "card_section"
	case KindBento:
		return "bento"
	case KindStat:
		return "stat"
	default:
		return "unknown"
	}
}

// SlideNode is the sealed scene IR union. Construct one of the concrete node
// types in this package; the set is closed (isSlideNode is unexported).
type SlideNode interface {
	NodeKind() NodeKind
	isSlideNode()
}

// node is embedded by every concrete node to satisfy the sealed marker.
type node struct{}

func (node) isSlideNode() {}

// ============================================================================
// Leaf nodes (RFC §11.1)
// ============================================================================

// Hero is a cover-slide title block: eyebrow + title + optional subtitle.
// Align overrides the slide's Content.Horizontal for this node; the zero
// value (HAlignLeft) inherits the slide default.
type Hero struct {
	node
	Eyebrow  string
	Title    string
	Subtitle string
	Align    HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (Hero) NodeKind() NodeKind { return KindHero }

// Prose is one or more body paragraphs.
// Align overrides the slide's Content.Horizontal for this node; the zero
// value (HAlignLeft) inherits the slide default.
type Prose struct {
	node
	Paragraphs []RichText
	Align      HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (Prose) NodeKind() NodeKind { return KindProse }

// Heading is a section heading at the given level (1–6).
// Align overrides the slide's Content.Horizontal for this node; the zero
// value (HAlignLeft) inherits the slide default.
type Heading struct {
	node
	Text  RichText
	Level int
	Align HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (Heading) NodeKind() NodeKind { return KindHeading }

// ListKind selects a list's marker style.
type ListKind int

const (
	ListBullet ListKind = iota
	ListNumber
	ListChecklist
)

// ListItem is one entry in a List.
type ListItem struct {
	Text    RichText
	Level   int
	Checked bool // checklist items
}

// List is a bullet / numbered / checklist block.
type List struct {
	node
	Kind  ListKind
	Items []ListItem
}

func (List) NodeKind() NodeKind { return KindList }

// Divider is a horizontal rule with surrounding spacing.
type Divider struct {
	node
	Spacing SpaceRole
}

func (Divider) NodeKind() NodeKind { return KindDivider }

// Quote is a pull quote with optional attribution.
// Align overrides the slide's Content.Horizontal for this node; the zero
// value (HAlignLeft) inherits the slide default.
type Quote struct {
	node
	Text        RichText
	Attribution string
	Align       HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (Quote) NodeKind() NodeKind { return KindQuote }

// CalloutKind selects a callout's tone.
type CalloutKind int

const (
	CalloutNote CalloutKind = iota
	CalloutWarning
	CalloutTip
	CalloutImportant
)

// Callout is a colored side-bar note.
type Callout struct {
	node
	Kind  CalloutKind
	Title string
	Body  RichText
}

func (Callout) NodeKind() NodeKind { return KindCallout }

// FrameKind selects optional device-frame chrome around an image.
type FrameKind int

const (
	FrameNone FrameKind = iota
	FrameBrowser
	FramePhone
	FrameDesktop
	FrameLaptop
)

// Crop is a per-edge fractional image crop (0..1 trimmed from each edge),
// re-exported from the builder so the IR uses the same vocabulary (D-039). It
// drives the OOXML srcRect; the zero value is no crop.
type Crop = pptx.Crop

// Fit is the image fill mode, re-exported from the builder (D-039). V1 ships
// FitFill (the default — stretches to fill the box) and FitNone; aspect-aware
// cover/contain are not in V1 (they need pixel dimensions, forbidden by §7).
type Fit = pptx.Fit

const (
	// FitFill stretches the image to fill its box (the zero value / default).
	FitFill = pptx.FitFill
	// FitNone places the image without a stretch fill mode.
	FitNone = pptx.FitNone
)

// Image is an asset image with optional frame chrome (renders as a pic shape).
//
// Frame selects one of the curated device frames by enum; FrameName selects a
// frame by name and, when non-empty, takes precedence over Frame — it is the
// seam for a caller frame registered via scene.WithFrameExtension (D-038). With
// both unset (FrameNone, "") the image renders without a bezel.
//
// Crop trims the source image per edge; Fit selects the fill mode. Both are
// mechanism exposure of the builder's crop/fit (D-039); their zero values
// (Crop{}, FitFill) render the image uncropped and stretched.
type Image struct {
	node
	AssetID   AssetID
	Alt       string
	Frame     FrameKind
	FrameName string
	Crop      Crop
	Fit       Fit
}

func (Image) NodeKind() NodeKind { return KindImage }

// ChipTone selects a chip's fill treatment.
type ChipTone int

const (
	ChipTint ChipTone = iota
	ChipSolid
	ChipOutline
)

// Chip is an inline pill.
// Align overrides the slide's Content.Horizontal for this node; the zero
// value (HAlignLeft) inherits the slide default.
type Chip struct {
	node
	Label string
	Tone  ChipTone
	Color ColorRole
	Align HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (Chip) NodeKind() NodeKind { return KindChip }

// ArrowDirection selects an arrow's direction.
type ArrowDirection int

const (
	ArrowRight ArrowDirection = iota
	ArrowLeft
	ArrowUp
	ArrowDown
)

// Arrow is an inline directional connector with an optional label.
type Arrow struct {
	node
	Direction ArrowDirection
	Label     string
}

func (Arrow) NodeKind() NodeKind { return KindArrow }

// CodeBlock is block-level code, rendered as a caller-rasterized pic (D-014).
type CodeBlock struct {
	node
	AssetID  AssetID
	Language string
	Caption  string
}

func (CodeBlock) NodeKind() NodeKind { return KindCodeBlock }

// Chart is an image-shape chart in V1 (native c:chart is V2; D-004).
type Chart struct {
	node
	AssetID AssetID
	Caption string
}

func (Chart) NodeKind() NodeKind { return KindChart }

// Table is headered tabular data; every cell is RichText. A non-empty Caption
// renders as a separate text shape above the table.
type Table struct {
	node
	Headers []RichText
	Rows    [][]RichText
	Caption string
}

func (Table) NodeKind() NodeKind { return KindTable }

// FlowOrientation selects a flow's direction.
type FlowOrientation int

const (
	FlowHorizontal FlowOrientation = iota
	FlowVertical
)

// ConnectorKind selects a flow's inter-step glyph (D-044). The zero value is
// ConnectorArrow, so an existing Flow with no Connector keeps a solid-arrow
// pipeline. Connectors compose preset shapes — no anchored AddConnector.
type ConnectorKind int

const (
	ConnectorArrow       ConnectorKind = iota // solid arrow (default)
	ConnectorArrowDashed                      // dashed line + chevron head
	ConnectorCycle                            // arrows + a trailing return arrow
	ConnectorPlus                             // a mathPlus glyph between steps
)

// FlowStep is one step in a Flow: a label, optional detail line, and optional
// icon. Icon is a closed-name curated/extension icon (Stage-1 validated), like
// a card's (D-044); its zero value renders a plain pill.
type FlowStep struct {
	Label  RichText
	Detail RichText
	Icon   string
}

// Flow is a sequential step pipeline. Connector is additive (D-044): its zero
// value (ConnectorArrow) preserves a solid-arrow flow.
type Flow struct {
	node
	Orientation FlowOrientation
	Steps       []FlowStep
	Connector   ConnectorKind
}

func (Flow) NodeKind() NodeKind { return KindFlow }

// DecorationKind selects how a Decoration is sourced.
type DecorationKind int

const (
	// DecorationPreset renders a curated ornament natively (SVG → preset/path).
	DecorationPreset DecorationKind = iota
	// DecorationAsset renders caller-supplied bytes as a pic.
	DecorationAsset
)

// Layer selects whether a decoration renders behind or above body content.
type Layer int

const (
	LayerBackground Layer = iota
	LayerForeground
)

// Decoration is an anchored ornament: a curated preset (native) or a
// caller-supplied asset (image). The AssetID is used only when Kind is
// DecorationAsset.
//
// The placement box aligns the box's anchor-corresponding point (its top-left
// for a top-left Anchor, its center for a center Anchor, …) to that anchor point
// on the slide, shifted by Offset and sized by Size (a zero Size uses a
// default). Bleed permits the box to extend past the slide edge (negative
// offsets, RFC §14.2) without a warning. Opacity (0..1; 0 = opaque) dims the
// decoration and Rotation (degrees) rotates it — both honored for asset
// decorations and single-shape ornaments (chevron); a multi-shape ornament
// cannot rotate as a unit in V1 (no group transform — D-041). Layer selects
// z-order: background renders behind body content, foreground above it
// (RFC §10.2).
type Decoration struct {
	node
	Kind     DecorationKind
	Preset   string // curated ornament name (Kind == DecorationPreset)
	AssetID  AssetID
	Layer    Layer
	Anchor   Anchor
	Offset   Position // EMU shift from the anchor point
	Size     Size     // ornament box; zero = a default size
	Bleed    bool     // allow the box to extend past the slide edge
	Opacity  float64  // 0..1; 0 = fully opaque
	Rotation float64  // degrees clockwise
}

func (Decoration) NodeKind() NodeKind { return KindDecoration }

// SectionDivider is a full-bleed chapter break (a whole slide).
// Align overrides the slide's Content.Horizontal for this node; the zero
// value (HAlignLeft) inherits the slide default.
type SectionDivider struct {
	node
	Eyebrow string
	Label   string
	Align   HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (SectionDivider) NodeKind() NodeKind { return KindSectionDivider }

// DeltaTone selects the color direction of a Stat's delta (D-057). The zero
// value DeltaNeutral is muted, so a delta with no tone set reads as neutral.
type DeltaTone int

const (
	DeltaNeutral DeltaTone = iota // muted (zero value)
	DeltaUp                       // positive — success color
	DeltaDown                     // negative — error color
)

// Stat is a hero big-number metric: a display-scale Value with a Label and an
// optional directional Delta (e.g. "$2,200" / "ARR" / "+12%"). A row of Stats
// inside a Grid forms a metric/pricing strip. The engine renders Value/Delta
// verbatim — it formats no numbers (D-026).
type Stat struct {
	node
	Value     string
	Label     string
	Delta     string    // "" = no delta line
	DeltaTone DeltaTone // color direction of Delta
}

func (Stat) NodeKind() NodeKind { return KindStat }

// ============================================================================
// Container nodes (RFC §11.2)
// ============================================================================

// ColumnRatio selects a two_column split.
type ColumnRatio int

const (
	Ratio11 ColumnRatio = iota // 1:1
	Ratio12                    // 1:2
	Ratio21                    // 2:1
)

// ColumnJoin is the optional element a TwoColumn draws centered on its seam
// (D-055). JoinNone (zero value) draws nothing, so an existing TwoColumn renders
// byte-for-byte unchanged.
type ColumnJoin int

const (
	JoinNone  ColumnJoin = iota // default: nothing between the columns
	JoinBadge                   // a circular text badge (JoinLabel), e.g. "VS"
	JoinArrow                   // a right-arrow connector between the columns
)

// TwoColumn splits the body into left/right regions with leaf children. Join /
// JoinLabel are additive (D-055): their zero values draw no inter-column element.
type TwoColumn struct {
	node
	Ratio     ColumnRatio
	Left      []SlideNode
	Right     []SlideNode
	Join      ColumnJoin // optional element centered on the column seam; JoinNone = none
	JoinLabel string     // badge text when Join == JoinBadge (e.g. "VS")
}

func (TwoColumn) NodeKind() NodeKind { return KindTwoColumn }

// Grid is a 2/3/4-column layout with weighted ratios and one child per cell.
type Grid struct {
	node
	Columns int
	Ratio   []int // per-column weights; empty = equal
	Gap     SpaceRole
	Cells   []SlideNode
}

func (Grid) NodeKind() NodeKind { return KindGrid }

// BentoCell is one cell of a BentoRow: its content and how many of the bento's
// column units it spans (>= 1).
type BentoCell struct {
	Span int
	Node SlideNode
}

// BentoRow is one row of a Bento: an optional left-gutter label and a left-to-
// right sequence of span-weighted cells.
type BentoRow struct {
	Label string // "" = no label for this row
	Cells []BentoCell
}

// Bento is a row-labeled grid (D-056): rows that each carry an optional left
// label and cells of variable column span, measured against Columns shared
// column units (a span-S cell occupies S units, so columns align across rows).
// A row's spans sum to <= Columns. It is a container — its cells render per their
// own policy — and is distinct from Grid (uniform columns, one child per cell).
type Bento struct {
	node
	Columns int // shared column-unit count a row's spans are measured against (>= 1)
	Rows    []BentoRow
}

func (Bento) NodeKind() NodeKind { return KindBento }

// cellNodes returns every cell's node in row-major order. Used by the Stage-1
// validation and the asset/icon/image/decoration walks so a Bento recurses like
// any other container.
func (b Bento) cellNodes() []SlideNode {
	out := make([]SlideNode, 0, len(b.Rows))
	for _, row := range b.Rows {
		for _, cell := range row.Cells {
			out = append(out, cell.Node)
		}
	}
	return out
}

// BodyLayout selects how a card stacks its children.
type BodyLayout int

const (
	BodyVertical BodyLayout = iota
	BodyHorizontal
)

// BorderStyle selects a card's border treatment (D-043). BorderDefault (zero)
// defers to the legacy Outline bool, so an existing Card{…, Outline:…} renders
// byte-identically; an explicit style overrides Outline.
type BorderStyle int

const (
	BorderDefault BorderStyle = iota // defer to Outline
	BorderNone                       // no border (even if Outline is true)
	BorderSolid                      // neutral hairline border
	BorderAccent                     // accent-colored border
)

// CardSize scales a card's interior padding. CardSizeMD (zero) preserves the
// default padding.
type CardSize int

const (
	CardSizeMD CardSize = iota
	CardSizeSM
	CardSizeLG
)

// CardLayout arranges a card's header region. CardLayoutDefault (zero) places
// the icon to the left of the eyebrow/header stack (a header row); IconTop
// stacks the icon above the text. (Further v4 header variants are deferred —
// RFC §11.3; plan §16.)
type CardLayout int

const (
	CardLayoutDefault CardLayout = iota // icon left of the eyebrow/header stack
	CardLayoutIconTop                   // icon above the eyebrow/header stack
)

// Card is an accent card: chrome (rounded rect + accent stripe + optional
// icon/eyebrow/header/header-pill) over a body of leaf children. All fields
// beyond Header/Body/BodyLayout/Fill/Outline/Elevation are additive (D-043):
// their zero values reproduce the pre-Phase-14 render byte-for-byte.
type Card struct {
	node
	Header      string
	Eyebrow     string // kicker label above the header
	Icon        string // curated/extension icon name (closed-name; Stage-1 validated)
	HeaderPill  string // pill badge text, right of the header row
	Body        []SlideNode
	BodyLayout  BodyLayout
	Fill        ColorRole
	Outline     bool        // legacy border shorthand; see BorderStyle (D-043)
	BorderStyle BorderStyle // explicit border; BorderDefault defers to Outline
	Size        CardSize    // interior padding scale
	Layout      CardLayout  // header arrangement
	Elevation   ElevationRole
	// Rich visuals (D-054). Optional; each zero value (nil / "") omits its
	// element, so a card that sets none renders byte-for-byte as before. The two
	// colors are *ColorRole, not ColorRole, because ColorRole's zero value is a
	// real color (ColorCanvas) and cannot signal "unset".
	HeaderFill *ColorRole // banded header region color (body keeps Fill); nil = no band
	StatusDot  *ColorRole // small status dot, top-right corner; nil = no dot
	Watermark  string     // large, low-opacity label drawn behind the body; "" = none
}

func (Card) NodeKind() NodeKind { return KindCard }

// CardSection is a top-level card that accepts grid / two_column / nested cards.
type CardSection struct {
	node
	Header string
	Body   []SlideNode
}

func (CardSection) NodeKind() NodeKind { return KindCardSection }
