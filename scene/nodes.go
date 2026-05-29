package scene

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
type Hero struct {
	node
	Eyebrow  string
	Title    string
	Subtitle string
}

func (Hero) NodeKind() NodeKind { return KindHero }

// Prose is one or more body paragraphs.
type Prose struct {
	node
	Paragraphs []RichText
}

func (Prose) NodeKind() NodeKind { return KindProse }

// Heading is a section heading at the given level (1–6).
type Heading struct {
	node
	Text  RichText
	Level int
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
type Quote struct {
	node
	Text        RichText
	Attribution string
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

// Image is an asset image with optional frame chrome (renders as a pic shape).
type Image struct {
	node
	AssetID AssetID
	Alt     string
	Frame   FrameKind
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
type Chip struct {
	node
	Label string
	Tone  ChipTone
	Color ColorRole
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

// Table is headered tabular data; every cell is RichText.
type Table struct {
	node
	Headers []RichText
	Rows    [][]RichText
}

func (Table) NodeKind() NodeKind { return KindTable }

// FlowOrientation selects a flow's direction.
type FlowOrientation int

const (
	FlowHorizontal FlowOrientation = iota
	FlowVertical
)

// FlowStep is one step in a Flow.
type FlowStep struct {
	Label  RichText
	Detail RichText
}

// Flow is a sequential step pipeline.
type Flow struct {
	node
	Orientation FlowOrientation
	Steps       []FlowStep
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
type Decoration struct {
	node
	Kind    DecorationKind
	Preset  string // curated ornament name (Kind == DecorationPreset)
	AssetID AssetID
	Layer   Layer
	Anchor  Anchor
}

func (Decoration) NodeKind() NodeKind { return KindDecoration }

// SectionDivider is a full-bleed chapter break (a whole slide).
type SectionDivider struct {
	node
	Eyebrow string
	Label   string
}

func (SectionDivider) NodeKind() NodeKind { return KindSectionDivider }

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

// TwoColumn splits the body into left/right regions with leaf children.
type TwoColumn struct {
	node
	Ratio ColumnRatio
	Left  []SlideNode
	Right []SlideNode
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

// BodyLayout selects how a card stacks its children.
type BodyLayout int

const (
	BodyVertical BodyLayout = iota
	BodyHorizontal
)

// Card is an accent card with a body of leaf children.
type Card struct {
	node
	Header     string
	Body       []SlideNode
	BodyLayout BodyLayout
	Fill       ColorRole
	Outline    bool
	Elevation  ElevationRole
}

func (Card) NodeKind() NodeKind { return KindCard }

// CardSection is a top-level card that accepts grid / two_column / nested cards.
type CardSection struct {
	node
	Header string
	Body   []SlideNode
}

func (CardSection) NodeKind() NodeKind { return KindCardSection }
