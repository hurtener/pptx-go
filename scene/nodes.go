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
	KindButton
	KindChecklist
	KindChipRow
	KindBanner
	KindIconRows
	KindLockup
	KindTimeline
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
	case KindButton:
		return "button"
	case KindChecklist:
		return "checklist"
	case KindChipRow:
		return "chip_row"
	case KindBanner:
		return "banner"
	case KindIconRows:
		return "icon_rows"
	case KindLockup:
		return "lockup"
	case KindTimeline:
		return "timeline"
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
	// AutoFit opts the Title (the display run) into shrink-to-fit: when its
	// estimated width exceeds the box, the engine downscales the Title font so it
	// fits one line, within a pinned minimum ratio. Zero = off, byte-identical.
	// (D-074.)
	AutoFit bool
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
	// AutoFit opts the heading text into shrink-to-fit: when its estimated width
	// exceeds the box, the engine downscales every run by one shared factor so it
	// fits one line, within a pinned minimum ratio. Zero = off, byte-identical.
	// (D-074.)
	AutoFit bool
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

// ListIndent selects a list's bullet hanging-indent density (the marker-to-text
// offset). The zero value IndentNormal preserves the default; IndentTight packs
// the markers closer to their text for dense lists. (D-078.)
type ListIndent int

const (
	IndentNormal ListIndent = iota
	IndentTight
)

// List is a bullet / numbered / checklist block.
type List struct {
	node
	Kind  ListKind
	Items []ListItem
	// Indent selects the bullet hanging-indent density. IndentNormal (zero) is
	// byte-identical to the pre-R10.9 render; IndentTight tightens the
	// marker-to-text offset consistently across all items and levels. (D-078.)
	Indent ListIndent
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
	// Testimonial enrichments (R14.5, D-120). Each zero value omits its element, so
	// a Quote with only Text+Attribution renders byte-for-byte as before. When any
	// of these is set the enriched testimonial layout runs: an optional oversized
	// quotation Mark behind the text, an optional rounded Avatar, a structured
	// attribution (Name / Role / Company), and an optional customer Logo.
	Mark bool // draw a large, low-emphasis quotation glyph behind the quote text
	// AvatarAssetID is the author's avatar (resolved via the AssetResolver, drawn
	// as a rounded picture); "" = no avatar.
	AvatarAssetID AssetID
	// AttributionName / Role / Company are the structured attribution; when Name
	// is set they supersede the flat Attribution string in the enriched layout.
	AttributionName    string
	AttributionRole    string
	AttributionCompany string
	// LogoAssetID is the customer/brand logo (resolved via the AssetResolver); ""
	// = no logo.
	LogoAssetID AssetID
}

func (Quote) NodeKind() NodeKind { return KindQuote }

// enriched reports whether a Quote uses any testimonial enrichment (R14.5) and so
// renders via the enriched layout rather than the plain text path.
func (q Quote) enriched() bool {
	return q.Mark || q.AvatarAssetID != "" || q.LogoAssetID != "" ||
		q.AttributionName != "" || q.AttributionRole != "" || q.AttributionCompany != ""
}

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
	// CornerRadius rounds the picture's corners from a theme radius token (D-114).
	// RadiusNone (the zero value) leaves the picture rectangular — byte-identical.
	CornerRadius RadiusRole
	// Elevation casts a soft drop shadow on the picture from a theme elevation
	// token (D-114). ElevationFlat (the zero value) emits no shadow —
	// byte-identical. Matches the card/surface finish.
	Elevation ElevationRole
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
	// Style, when non-nil, applies comparison-matrix styling — a header band,
	// zebra body striping, a highlighted column, an emphasized row-label column,
	// and grouped header spans — all from theme tokens (R14.3, D-118). nil leaves
	// the plain banded table (byte-identical). A non-nil Style controls every cell
	// fill explicitly (it does not use the builder's default header/row banding).
	Style *TableStyle
}

func (Table) NodeKind() NodeKind { return KindTable }

// TableStyle is the additive visual styling for a comparison-matrix Table
// (R14.3, D-118). Every field's zero value reproduces an unstyled column, so a
// caller turns features on one at a time. Colors resolve from theme tokens (P2):
// the header band and highlighted column use ColorAccent, zebra and the row-label
// column use ColorSurfaceAlt. Cell-value glyphs (check / cross / dot / mini-bar)
// are intentionally not a Table feature — a native OOXML table cell holds only a
// text body (no shape children), so they are composed instead with a Bento of
// Checklist / IconRows cells (the glyph nodes already shipped — D-095/D-100).
type TableStyle struct {
	// HeaderFill fills the header row with the accent band (contrast text).
	HeaderFill bool
	// Zebra alternates a subtle SurfaceAlt fill on odd body rows.
	Zebra bool
	// HighlightCol is the 1-based column to emphasize (accent tint + heavier
	// accent border) — e.g. a "recommended" plan column. 0 (the zero value) = none.
	HighlightCol int
	// RowLabelCol emphasizes the first column as row labels (SurfaceAlt fill + bold).
	RowLabelCol bool
	// HeaderGroups, when non-empty, adds a grouped header row above the headers:
	// each group's Label spans Span columns (merged), laid left-to-right from
	// column 0. The spans should sum to the column count.
	HeaderGroups []HeaderGroup
}

// HeaderGroup is one merged span in a Table's grouped header row (D-118).
type HeaderGroup struct {
	// Label is the group heading (e.g. "Enterprise").
	Label string
	// Span is the number of columns the group covers (>= 1).
	Span int
}

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
	ConnectorBiArrow                          // a bidirectional left-right / up-down arrow (R12.4)
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
	// DecorationText renders a large, low-opacity text watermark (an oversized
	// ghost number/word behind the body) from Decoration.Text (D-109). Appended
	// last so existing DecorationKind values are unchanged (byte-identical).
	DecorationText
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
	// Color overrides the ornament's color role (D-107). nil = ColorAccent
	// (byte-identical to pre-D-107 output) — a pointer because ColorRole's zero
	// value is ColorCanvas, a real color (the D-054 pattern). Set it to render a
	// neutral-grey paper grain, an inverse-white starfield, or any brand-role
	// texture/glow. Applies to DecorationPreset; an asset decoration ignores it.
	// For DecorationText it colors the watermark glyph (nil = ColorAccent).
	Color *pptx.ColorRole
	// Text is the watermark string for DecorationText (an oversized ghost
	// number/word, e.g. "03"); ignored by other kinds. Empty fails validation
	// for DecorationText (D-109).
	Text string
	// FontSize is the watermark text size in points for DecorationText; 0 uses a
	// box-height "fill the box" default. Ignored by other kinds (D-109).
	FontSize float64
	// Pitch is the lattice spacing (EMU) for the pattern ornaments (grid_dots /
	// noise_overlay / starfield): their dot count derives from the box at this
	// pitch, so a full-bleed texture keeps a consistent visual density. 0 (the
	// zero value) keeps each pattern's legacy fixed count — byte-identical.
	// Ignored by non-pattern presets and other kinds (D-111).
	Pitch pptx.EMU
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
	// AutoFit opts the Value (the display run) into shrink-to-fit: when its
	// estimated width exceeds the box, the engine downscales the Value font so a
	// long number/price fits one line, within a pinned minimum ratio. Zero = off,
	// byte-identical. (D-074.)
	AutoFit bool
}

func (Stat) NodeKind() NodeKind { return KindStat }

// ButtonTone selects a Button's fill treatment (R12.1, D-094). Each tone maps to
// theme color tokens (P2), so a theme swap re-skins every button. The zero value
// ButtonPrimary is a solid accent pill — the default "do this next" affordance.
type ButtonTone int

const (
	ButtonPrimary   ButtonTone = iota // solid ColorAccent fill, inverse label (zero value)
	ButtonAccentAlt                   // solid ColorAccentAlt fill, inverse label
	ButtonGhost                       // no fill + an accent hairline outline, accent label
	ButtonNeutral                     // solid ColorSurfaceAlt fill, default label
)

// ButtonSize scales a Button's height, interior padding, and icon size. The zero
// value ButtonMD is the default; SM/LG step it down/up. A pinned layout metric, not
// a theme token (it sizes geometry, not a visual property).
type ButtonSize int

const (
	ButtonMD ButtonSize = iota // default
	ButtonSM
	ButtonLG
)

// Button is a presentational CTA / action affordance (R12.1, D-094): a content-fit
// RadiusFull pill with a label and optional leading/trailing icons, droppable
// standalone (a closing slide), inside a card body (a pricing card), or inside a
// banner. It is a shape only — no hyperlink/action wiring (the deck is static).
//
// Width is content-fit (label + icons + padding) clamped to its box; Align offsets
// the pill within the box (zero = inherit the slide's Content.Horizontal). Tone
// selects the token fill (ghost = outline); Size scales the geometry. LeadingIcon /
// TrailingIcon are closed-name curated/extension icons (Stage-1 validated); their
// zero value ("") renders no glyph. Additive: absent ⇒ byte-identical.
type Button struct {
	node
	Label        string
	Tone         ButtonTone
	Size         ButtonSize
	LeadingIcon  string // closed-name curated/extension icon; "" = none
	TrailingIcon string // closed-name curated/extension icon; "" = none
	Align        HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (Button) NodeKind() NodeKind { return KindButton }

// CheckState selects a checklist item's status glyph (R12.2, D-095). The zero value
// CheckDone is a filled affirmative check (the common "you get this" row).
type CheckState int

const (
	CheckDone    CheckState = iota // a filled check glyph (default), accent-tinted
	CheckNo                        // a filled cross glyph, muted
	CheckNeutral                   // a filled dot glyph, muted
)

// ChecklistItem is one row of a Checklist: rich text, a status, and an optional icon
// name that overrides the state's default glyph (a closed-name curated/extension icon,
// Stage-1 validated). The zero value renders a filled check glyph before the text.
type ChecklistItem struct {
	Text  RichText
	State CheckState
	Icon  string // optional glyph override; "" = the state's default glyph
}

// Checklist is a dense feature/"what you get" list (R12.2, D-095): rows of a filled
// status glyph (check / cross / dot) before rich text, reflowed row-major into 1–3
// balanced columns, with the text hanging-indented from the glyph width. The glyph is a
// true filled custGeom (the curated check/x/dot icon), never an empty font checkbox.
//
// GlyphTone overrides the per-state glyph color for every item; it is a *ColorRole so
// nil selects the per-state default (CheckDone → accent, others → muted) — ColorRole's
// zero value is a real color (ColorCanvas) and cannot signal "unset" (D-054 pattern).
// Fill distributes inter-row slack so a short list spans the box height (the last row
// meets the bottom); the zero value top-aligns the rows. Additive: a deck with no
// Checklist is byte-identical (the List/BulletCheckbox path is untouched).
type Checklist struct {
	node
	Items     []ChecklistItem
	Columns   int        // 1..3 column reflow (row-major); 0 = 1 column
	GlyphTone *ColorRole // glyph color override; nil = per-state default
	Fill      bool       // distribute rows to fill the box height (like VAlignFill)
}

func (Checklist) NodeKind() NodeKind { return KindChecklist }

// ChipSpec is one chip in a ChipRow: a label, a tone, the tone's color role, and an
// optional leading icon (a closed-name curated/extension icon, Stage-1 validated). It
// mirrors the single Chip node's vocabulary (ChipTone / ColorRole). For ChipTint the
// Color is ignored (the chip uses ColorSurfaceAlt); ChipSolid/ChipOutline use Color.
type ChipSpec struct {
	Label string
	Tone  ChipTone
	Color ColorRole
	Icon  string // optional leading glyph; "" = none
}

// ChipRow is a horizontal, wrap-to-next-line row of content-fit chip pills with an
// optional leading label (R12.5, D-096): a tag / category / capability strip. Each
// chip sizes to its label (plus an optional leading icon); chips lay left-to-right and,
// when Wrap is set, reflow onto new lines when the row width is exceeded.
//
// Wrap is the engine mechanism: the zero value lays all chips on a single line (the
// minimal behavior); a product that wants a reflowing strip sets Wrap true (D-026). A
// non-empty Label renders as a leading TypeCaption label before the first chip. Align
// offsets each line's chips (zero = inherit the slide's Content.Horizontal). Additive:
// a deck with no ChipRow is byte-identical.
type ChipRow struct {
	node
	Label string
	Chips []ChipSpec
	Wrap  bool   // wrap chips onto new lines (zero = single line)
	Align HAlign // per-node horizontal alignment override; 0 = inherit slide
}

func (ChipRow) NodeKind() NodeKind { return KindChipRow }

// Banner is a full-width filled "big takeaway / promo / CTA" strip (R12.6, D-097): a
// leading icon + a bold lead phrase + a supporting body on the left, with optional
// right-aligned Trailing children (a Stat and/or a Button). Distinct from the side-bar
// Callout — the banner is a wide, full-fill band.
//
// Fill is the strip color; its zero value (ColorCanvas) is treated as ColorAccent (a
// banner is always a filled strip — a canvas-colored one would be invisible). TextColor
// colors the lead/body; its zero value (TextPrimary) auto-contrasts against Fill (light
// on a dark fill), and any explicit non-default value is honored. Trailing children
// render in a right region per their own policy. Additive: a deck with no Banner is
// byte-identical.
type Banner struct {
	node
	Lead      RichText
	Body      RichText
	Icon      string        // leading curated/extension icon; "" = none
	Fill      ColorRole     // strip fill; zero (ColorCanvas) = ColorAccent
	TextColor TextColorRole // lead/body color; zero (TextPrimary) = auto-contrast on Fill
	Trailing  []SlideNode   // right-aligned children (e.g. Stat/Button/Lockup); nil = none
}

func (Banner) NodeKind() NodeKind { return KindBanner }

// RowTone selects an IconRow's framing (R12.7, D-100). The zero value RowPlain draws no
// frame; RowPill wraps the row in a SurfaceAlt rounded-rect.
type RowTone int

const (
	RowPlain RowTone = iota // no frame (zero value)
	RowPill                 // a SurfaceAlt rounded-rect frame around the row
)

// IconRow is one row of an IconRows: a leading icon, a rich label, and an optional
// right-aligned meta. Icon is a closed-name curated/extension icon (Stage-1 validated);
// "" renders no glyph (the label starts at the left).
type IconRow struct {
	Icon  string
	Label RichText
	Meta  RichText // optional, right-aligned; nil = none
	Tone  RowTone
}

// IconRows is a vertical stack of [icon | label | optional meta] rows (R12.7, D-100): the
// "integrations / capabilities / sources" list that reads as designed rows rather than
// bullets. Fill distributes inter-row spacing so the rows span the box height (like
// VAlignFill); GlyphColor tints every row's icon (its zero value, ColorCanvas, defaults to
// ColorAccent — a canvas-colored glyph would be invisible). Additive: a deck with no
// IconRows is byte-identical.
type IconRows struct {
	node
	Rows       []IconRow
	Fill       bool      // distribute rows to fill the box height
	GlyphColor ColorRole // icon tint; zero (ColorCanvas) = ColorAccent
}

func (IconRows) NodeKind() NodeKind { return KindIconRows }

// AssetSide selects where a Lockup's logo sits relative to its caption (R12.9, D-102).
// The zero value LeadCaption places the caption first (caption leads, logo trails);
// TrailCaption places the logo first.
type AssetSide int

const (
	LeadCaption  AssetSide = iota // caption then logo (zero value)
	TrailCaption                  // logo then caption
)

// Lockup is a compact "powered by / in partnership with" attribution mark (R12.9, D-102):
// a caption paired with a small partner logo composed as one inline, centerable unit. The
// mark is either an AssetID (a partner logo, resolved via the AssetResolver — renders as a
// pic) or an Icon (a curated/extension glyph — media-free); exactly one is set. AssetSide
// places the logo before or after the caption; MaxHeight height-bounds the logo (a pinned
// default when 0); Align positions the whole group (zero = inherit the slide). Additive: a
// deck with no Lockup is byte-identical.
type Lockup struct {
	node
	Caption   string
	AssetID   AssetID   // the partner logo (resolved via AssetResolver); "" = use Icon
	Icon      string    // a curated/extension glyph instead of an asset; "" = use AssetID
	AssetSide AssetSide // logo before (TrailCaption) or after (LeadCaption) the caption
	MaxHeight pptx.EMU  // logo height bound; 0 = a pinned default
	Align     HAlign    // per-node horizontal alignment override; 0 = inherit slide
}

func (Lockup) NodeKind() NodeKind { return KindLockup }

// Timeline is a roadmap / timeline node (R14.4, D-119): a horizontal axis with
// milestones placed at caller-specified proportional positions, optional phase
// bands behind them, and optional swimlanes (rows). Markers, the axis line, and
// labels compose from native preset shapes (no media). Labels stagger above /
// below the axis to avoid collision. Additive: a deck with no Timeline is
// byte-identical (it is a new node — unused means absent).
//
// Either Milestones (a single implicit lane) or Lanes (explicit swimlanes) drives
// the markers; Lanes, when non-empty, supersedes Milestones. Bands span the full
// timeline width behind every lane.
type Timeline struct {
	node
	// Milestones is the single-lane milestone list, used when Lanes is empty.
	Milestones []Milestone
	// Lanes are swimlanes (rows), each with its own milestones; supersedes
	// Milestones when non-empty.
	Lanes []TimelineLane
	// Bands are optional phase/horizon regions drawn behind the axis, each
	// spanning [From,To] of the timeline width.
	Bands []TimelineBand
}

func (Timeline) NodeKind() NodeKind { return KindTimeline }

// Milestone is one point on a Timeline axis (D-119). Position is the proportional
// location along the axis in [0,1]; Label is the marker heading, Detail an
// optional sub-line; Icon (optional, curated/extension) replaces the dot marker;
// AccentIndex selects the marker color from a pinned token cycle (0 = ColorAccent).
type Milestone struct {
	Position    float64
	Label       string
	Detail      string
	Icon        string
	AccentIndex int
}

// TimelineLane is one swimlane (row) of a Timeline (D-119): a left-gutter Label
// and its own milestones placed along the lane's axis.
type TimelineLane struct {
	Label      string
	Milestones []Milestone
}

// TimelineBand is a phase/horizon region behind a Timeline axis (D-119): it spans
// [From,To] (each in [0,1]) of the timeline width, filled with Fill (a surface
// role, low-alpha) and labeled at the top.
type TimelineBand struct {
	From  float64
	To    float64
	Label string
	Fill  ColorRole
}

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

// JoinPosition selects where a TwoColumn's Join element sits (R12.8, D-101). The zero
// value JoinSeam centers it on the vertical seam between the columns (the D-055 default);
// JoinTopBridge / JoinBottomBridge draw a horizontal accent bracket spanning both columns'
// combined width at the top / bottom edge, with the JoinLabel as a centered pill on it —
// the "one X, two ways" header used on option/path slides.
type JoinPosition int

const (
	JoinSeam         JoinPosition = iota // centered on the seam (zero value, D-055)
	JoinTopBridge                        // a bracket spanning both column tops
	JoinBottomBridge                     // a bracket spanning both column bottoms
)

// TwoColumn splits the body into left/right regions with leaf children. Join /
// JoinLabel / JoinPosition are additive (D-055, D-101): their zero values draw no
// inter-column element (or, for a non-None Join, the centered-seam element).
type TwoColumn struct {
	node
	Ratio        ColumnRatio
	Left         []SlideNode
	Right        []SlideNode
	Join         ColumnJoin   // optional element centered on the column seam; JoinNone = none
	JoinLabel    string       // badge / bridge text when Join != JoinNone (e.g. "VS", "One agent")
	JoinPosition JoinPosition // JoinSeam (default) / JoinTopBridge / JoinBottomBridge (R12.8)
}

func (TwoColumn) NodeKind() NodeKind { return KindTwoColumn }

// GridConnector draws a connector glyph in the gutter between two adjacent columns of
// a Grid (R12.4, D-099), so an architecture / pipeline grid reads as data flow, not
// just adjacency. Between holds the two adjacent column indices ({c, c+1}); Kind reuses
// the Flow connector set (plus ConnectorBiArrow); Label is an optional caption.
type GridConnector struct {
	Between [2]int        // adjacent column indices, e.g. {0, 1}
	Kind    ConnectorKind // glyph; ConnectorBiArrow = a bidirectional arrow
	Label   string        // optional caption in the gutter; "" = none
}

// Grid is a 2/3/4-column layout with weighted ratios and one child per cell.
// Connectors (additive, R12.4) draw glyphs in the gutters between adjacent columns;
// an empty Connectors slice renders byte-identically.
type Grid struct {
	node
	Columns    int
	Ratio      []int // per-column weights; empty = equal
	Gap        SpaceRole
	Cells      []SlideNode
	Connectors []GridConnector // inter-column gutter glyphs; empty = none
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
	// WeightedRows opts a bento into content-proportional row heights: each row
	// sizes to the preferred height of its tallest cell (at that cell's span
	// width), clamped by a single deterministic scale so the rows always fit the
	// region. The zero value (false) keeps equal-height rows — byte-identical to
	// the pre-R10.3 layout. (D-072.)
	WeightedRows bool
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
	Header     string
	Eyebrow    string // kicker label above the header
	Icon       string // curated/extension icon name (closed-name; Stage-1 validated)
	HeaderPill string // pill badge text, right of the header row
	Body       []SlideNode
	BodyLayout BodyLayout
	Fill       ColorRole
	// FillGradient, when non-nil, replaces the solid Fill with a 2-stop linear
	// gradient surface (From→To at Angle degrees clockwise from +x) for a subtle
	// top-to-bottom depth shift (D-108). nil = solid Fill (byte-identical). Both
	// stops resolve against the active theme (P2); a darker-To auto-tint is the
	// soul's choice (D-026), not the engine's.
	FillGradient *GradientFill
	Outline      bool        // legacy border shorthand; see BorderStyle (D-043)
	BorderStyle  BorderStyle // explicit border; BorderDefault defers to Outline
	Size         CardSize    // interior padding scale
	Layout       CardLayout  // header arrangement
	Elevation    ElevationRole
	// Rich visuals (D-054). Optional; each zero value (nil / "") omits its
	// element, so a card that sets none renders byte-for-byte as before. The two
	// colors are *ColorRole, not ColorRole, because ColorRole's zero value is a
	// real color (ColorCanvas) and cannot signal "unset".
	HeaderFill *ColorRole // banded header region color (body keeps Fill); nil = no band
	StatusDot  *ColorRole // small status dot, top-right corner; nil = no dot
	Watermark  string     // large, low-opacity label drawn behind the body; "" = none
	// BodyVAlign selects the vertical distribution of the card body within the
	// card body region (the same VAlign modes as the slide body stack:
	// Center / Bottom / Justify / Fill / Fit). The zero value VAlignTop is
	// top-anchored — byte-identical to the pre-R10.4 render. Applies to the
	// vertical body layout only (BodyLayout != BodyHorizontal). (D-073.)
	BodyVAlign VAlign
	// PaddingScale is a basis-point multiplier on the card's size-resolved
	// interior padding (Size → SpaceSM/MD/XL): 0 and 10000 leave it unchanged
	// (byte-identical), a value below 10000 tightens a dense card (floored at a
	// pinned minimum so the inset never collapses), above 10000 loosens it. It
	// resolves through theme spacing tokens — no literals (P2). (D-076.)
	PaddingScale int
	// Ribbon is an optional pinned emphasis badge (R12.3, D-098) — a "MOST POPULAR"
	// top bar or a corner badge that singles this card out of a row. It sits outside
	// the header text flow (distinct from HeaderPill); a RibbonTopBar shifts the body
	// down. nil = no ribbon, byte-identical.
	Ribbon *Ribbon
	// Backdrop is an optional decoration drawn behind the card's computed box,
	// before its fill (D-113) — a focal glow/halo that tracks the card across any
	// layout. Typically a center-anchored, bleeding radial_glow. nil = none,
	// byte-identical. The card box is passed as the decoration's region, so the
	// glow centers on the card and (with Bleed) spills beyond it behind the fill.
	Backdrop *Decoration
	// ImageFill fills the card surface with a cover-fit photo (resolved via the
	// render's AssetResolver) instead of the solid Fill / FillGradient — the
	// image-as-surface treatment for photographic cards (R14.1, D-117). The card's
	// rounded corners still clip the image. "" = the solid/gradient Fill
	// (byte-identical). A missing resolver or unresolvable ID records a
	// LayoutWarning and falls back to the Fill (RFC §10.2 — degrade, no panic).
	// The field is named ImageFill, not AssetID, because the card still renders as
	// native chrome (not a pic), so its policy stays HasAsset:false.
	ImageFill AssetID
}

func (Card) NodeKind() NodeKind { return KindCard }

// GradientFill is an optional 2-stop linear surface fill for a Card (D-108): a
// From→To linear gradient at Angle degrees clockwise from the positive x-axis
// (0 = left→right, 90 = top→bottom). Both stops are surface color roles resolved
// against the active theme, so a theme swap re-paints both (P2). A nil
// *GradientFill leaves the card on its solid Fill (byte-identical).
type GradientFill struct {
	From  ColorRole
	To    ColorRole
	Angle int
}

// RibbonPos selects where a Card.Ribbon is pinned (R12.3, D-098). The zero value
// RibbonTopBar is a full-width tab across the card's top edge that reserves its own band
// (the card body shifts down below it); the corner positions are overlays that do not
// shift the body.
type RibbonPos int

const (
	RibbonTopBar     RibbonPos = iota // full-width tab across the top (reserves a band)
	RibbonCornerTL                    // a text tab pinned in the top-left corner
	RibbonCornerTR                    // a text tab pinned in the top-right corner
	RibbonCornerStar                  // a star glyph in the top-right corner (Text ignored)
)

// Ribbon is a pinned emphasis badge on a Card (R12.3, D-098): a "MOST POPULAR" /
// "RECOMMENDED" / "NEW" marker that singles one card out of a row. It sits OUTSIDE the
// header text flow — distinct from Card.HeaderPill (an in-row pill). A RibbonTopBar
// reserves a band so the card body shifts down (cardHeaderBottom accounts for it); the
// corner positions are overlays.
//
// Color is the badge fill; nil selects ColorAccent (ColorRole's zero value is the real
// color ColorCanvas, so the override is a pointer — the D-054 pattern). TextColor colors
// the label; its zero value (TextPrimary) auto-contrasts against Color.
type Ribbon struct {
	Text      string
	Position  RibbonPos
	Color     *ColorRole    // nil = ColorAccent
	TextColor TextColorRole // zero (TextPrimary) = auto-contrast on Color
}

// CardSection is a top-level card that accepts grid / two_column / nested cards.
type CardSection struct {
	node
	Header string
	Body   []SlideNode
}

func (CardSection) NodeKind() NodeKind { return KindCardSection }
