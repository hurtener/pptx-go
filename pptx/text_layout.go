package pptx

import "github.com/hurtener/pptx-go/internal/ooxml/slide"

// Text enum → OOXML token mappings, and the RunStyle → run-properties
// translation. Kept apart from the builder types in text.go.

// alignString maps an Alignment to its OOXML algn value ("" = left default).
func alignString(a Alignment) string {
	switch a {
	case AlignCenter:
		return "ctr"
	case AlignRight:
		return "r"
	case AlignJustify:
		return "just"
	default:
		return ""
	}
}

// anchorString maps a TextAnchor to its OOXML anchor value ("" = top default).
func anchorString(v TextAnchor) string {
	switch v {
	case AnchorMiddle:
		return "ctr"
	case AnchorBottom:
		return "b"
	default:
		return ""
	}
}

// toProps translates a RunStyle to run character properties, resolving the type
// role and color against t. It returns nil when the style carries nothing to
// emit, so an unstyled run inherits its placeholder/master styling.
func (rs RunStyle) toProps(t *Theme) *slide.XTextProperties {
	spec := t.ResolveType(rs.TypeRole)
	p := &slide.XTextProperties{}
	set := false

	if spec.Size > 0 {
		p.FontSize = int(spec.Size * 100) // OOXML sz is in 1/100 pt
		set = true
	}
	family := spec.Family
	if rs.Code {
		family = t.ResolveType(TypeMono).Family // inline code is monospace (D-013)
	}
	if family != "" {
		p.Latin = &slide.XLatinFont{Typeface: family}
		set = true
	}
	if rs.Bold || spec.Bold() {
		p.Bold = "1"
		set = true
	}
	if rs.Italic || spec.Italic {
		p.Italic = "1"
		set = true
	}
	switch rs.Underline {
	case UnderlineSingle:
		p.Underline, set = "sng", true
	case UnderlineDouble:
		p.Underline, set = "dbl", true
	}
	switch rs.Strike {
	case StrikeSingle:
		p.Strike, set = "sngStrike", true
	case StrikeDouble:
		p.Strike, set = "dblStrike", true
	}
	switch rs.BaselineRel {
	case Superscript:
		p.Baseline, set = 30000, true
	case Subscript:
		p.Baseline, set = -25000, true
	}
	if rs.Color != nil {
		p.SolidFill = &slide.XSolidFill{SrgbClr: srgbFrom(rs.Color.resolve(t))}
		set = true
	}

	if !set {
		return nil
	}
	return p
}
