package pptx

import (
	"math"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

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

// alignFrom is alignString's read inverse: an OOXML algn value → Alignment.
func alignFrom(s string) Alignment {
	switch s {
	case "ctr":
		return AlignCenter
	case "r":
		return AlignRight
	case "just":
		return AlignJustify
	default:
		return AlignLeft
	}
}

// underlineFrom maps an OOXML u value back to an Underline.
func underlineFrom(s string) Underline {
	switch s {
	case "sng":
		return UnderlineSingle
	case "dbl":
		return UnderlineDouble
	default:
		return UnderlineNone
	}
}

// strikeFrom maps an OOXML strike value back to a Strike.
func strikeFrom(s string) Strike {
	switch s {
	case "sngStrike":
		return StrikeSingle
	case "dblStrike":
		return StrikeDouble
	default:
		return StrikeNone
	}
}

// baselineFrom maps an OOXML baseline percentage back to a BaselineShift, by
// sign (toProps emits +30000 for superscript, -25000 for subscript).
func baselineFrom(v int) BaselineShift {
	switch {
	case v > 0:
		return Superscript
	case v < 0:
		return Subscript
	default:
		return BaselineNone
	}
}

// bulletFromChar maps a bullet character back to its BulletKind (the read
// inverse of Bullet's BuChar branches; an unrecognized char is a disc).
func bulletFromChar(char string) BulletKind {
	switch char {
	case "☐":
		return BulletCheckbox
	default:
		return BulletDisc
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
	// Letter-spacing (tracking): a per-run override wins over the role's value.
	// Emitted as a:rPr/@spc in signed 1/100 pt; 0 emits nothing (D-060).
	tracking := spec.Tracking
	if rs.Tracking != nil {
		tracking = *rs.Tracking
	}
	if tracking != 0 {
		p.Spc = int(math.Round(tracking * 100))
		if p.Spc != 0 {
			set = true
		}
	}
	family := spec.Family
	if rs.Code {
		family = t.ResolveType(TypeMono).Family // inline code is monospace (D-013)
	}
	if family != "" {
		p.Latin = &slide.XLatinFont{Typeface: family}
		set = true
	}
	if rs.Code {
		// A subtle background tint sourced from the theme (D-013).
		p.Highlight = &slide.XHighlight{
			SrgbClr: srgbFrom(resolvedColor{rgb: normalizeHex(t.ResolveColor(ColorSurfaceAlt)), alpha: AlphaOpaque}),
		}
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
