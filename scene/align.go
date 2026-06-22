package scene

import "github.com/hurtener/pptx-go/pptx"

// Alignment types for the scene body stack (Phase 13 engine richness). The
// zero values {VAlignTop, HAlignLeft} reproduce the pre-Phase-13 layout
// unchanged — fully backward-compatible.

// HAlign selects horizontal text alignment within the body region.
// The zero value HAlignLeft is the default (left-flush, full-width box).
// Per-node Align fields and the slide Content.Horizontal use this type.
type HAlign int

const (
	// HAlignLeft (zero value) is the default: text leaf nodes span the full
	// body width and render left-aligned paragraphs. Backward-compatible.
	HAlignLeft HAlign = iota
	// HAlignCenter sets paragraph alignment to center on text leaf nodes
	// (Hero, Heading, Prose, Quote). The text box keeps its full body width;
	// each paragraph line is centered within that frame. For Chip nodes the
	// box is physically centered instead (the pill should move, not just its
	// text).
	HAlignCenter
	// HAlignRight sets paragraph alignment to right on text leaf nodes.
	// For Chip nodes the box is physically placed at the body right edge.
	HAlignRight
)

// hAlignToParagraph converts a scene HAlign to a pptx paragraph Alignment.
// HAlignLeft maps to AlignLeft (the OOXML default, emitting no algn attr),
// which preserves byte-identical output for unaligned content.
func hAlignToParagraph(h HAlign) pptx.Alignment {
	switch h {
	case HAlignCenter:
		return pptx.AlignCenter
	case HAlignRight:
		return pptx.AlignRight
	default:
		return pptx.AlignLeft
	}
}

// String returns the horizontal alignment name.
func (h HAlign) String() string {
	switch h {
	case HAlignCenter:
		return "center"
	case HAlignRight:
		return "right"
	default:
		return "left"
	}
}

// VAlign selects vertical alignment of the body stack within the body region.
// The zero value VAlignTop is the default (top-flush stack). The slide
// Content.Vertical field uses this type.
type VAlign int

const (
	// VAlignTop (zero value) is the default: the body stack starts at the body
	// region's top edge. Backward-compatible.
	VAlignTop VAlign = iota
	// VAlignCenter distributes the remaining vertical space equally above and
	// below the body stack; the stack never starts above the top edge.
	VAlignCenter
	// VAlignBottom places the body stack flush with the body region's bottom
	// edge; the stack never starts above the top edge.
	VAlignBottom
	// VAlignJustify distributes the vertical slack evenly into the inter-node
	// gaps. Equivalent to VAlignTop for a single node or when slack ≤ 0.
	VAlignJustify
	// VAlignFill pins fixed leaves at the top (like VAlignTop) and grows the
	// flexible nodes (containers + Image/Chart) to consume the remaining body
	// height, so a sparse slide fills its frame instead of reading thin. With no
	// flexible node, or when slack ≤ 0, it is equivalent to VAlignTop.
	VAlignFill
	// VAlignFit is the compression inverse of VAlignFill: an opt-in fit-to-region
	// mode for over-full slides. When the body stack's preferred height exceeds
	// the region, the renderer applies a single deterministic compression pass —
	// it shrinks the inter-node gaps toward a pinned floor (SpaceXS) and, if still
	// overflowing, proportionally scales every node's slot height toward a pinned
	// ratio floor — so the last node's bottom lands inside the region instead of
	// clipping off-slide. When the content already fits, VAlignFit is
	// byte-identical to VAlignTop (top-pinned, standard gap). All math is integer
	// EMU / basis-point, so the result is deterministic regardless of worker
	// count. The card-interior-padding and display-type-scale sub-steps are
	// layered in by separate engine units.
	VAlignFit
	// VAlignFillCapped is VAlignFill with a ceiling: each flexible node grows by
	// at most a pinned factor of its preferred height, so a near-empty node cannot
	// balloon to consume all the slack. The leftover slack beyond the caps becomes
	// balanced spacing — an even top margin and widened inter-node gaps — instead
	// of inflating one node. With no flexible node, or when the stack already
	// fills the region, it is equivalent to VAlignTop. Deterministic integer EMU.
	VAlignFillCapped
	// VAlignBalanced distributes a sparse stack's slack as an even rhythm — a top
	// margin plus widened inter-node gaps (the slack split across the n+1 spaces of
	// the stack) — with an optical-center bias that seats the stack slightly above
	// geometric center. Unlike VAlignJustify (all slack into gaps, no margins) and
	// VAlignCenter (all slack into equal margins, fixed gaps), it spreads whitespace
	// across both, so a sparse cover or closing reads balanced rather than clustered
	// with a large void. With no slack it is equivalent to VAlignTop. Deterministic
	// integer EMU.
	VAlignBalanced
)

// String returns the vertical alignment name.
func (v VAlign) String() string {
	switch v {
	case VAlignCenter:
		return "center"
	case VAlignBottom:
		return "bottom"
	case VAlignJustify:
		return "justify"
	case VAlignFill:
		return "fill"
	case VAlignFit:
		return "fit"
	case VAlignFillCapped:
		return "fill-capped"
	case VAlignBalanced:
		return "balanced"
	default:
		return "top"
	}
}

// Alignment is the combined body-stack alignment for a SceneSlide. The zero
// value {VAlignTop, HAlignLeft} reproduces the pre-Phase-13 top-left layout
// unchanged (backward-compatible zero value).
type Alignment struct {
	Vertical   VAlign
	Horizontal HAlign
}
