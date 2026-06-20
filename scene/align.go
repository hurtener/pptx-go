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
