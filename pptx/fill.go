package pptx

import "github.com/hurtener/pptx-go/internal/ooxml/slide"

// Fill is a shape's interior fill. Construct with SolidFill or NoFill. The
// interface is sealed (applyFill is unexported); a fill resolves its colors
// against the active theme when applied, so theme tokens re-render after a
// theme swap (P2).
//
// V1 ships SolidFill and NoFill. Gradient, pattern and picture (blip) fills are
// tracked separately — picture fills arrive with media support.
type Fill interface {
	applyFill(sp *slide.XShapeProperties, t *Theme)
}

// solidFill is a single-color fill.
type solidFill struct{ color Color }

// SolidFill returns a solid fill of the given color (literal or theme token).
func SolidFill(c Color) Fill { return solidFill{color: c} }

func (f solidFill) applyFill(sp *slide.XShapeProperties, t *Theme) {
	sp.NoFill = nil
	if f.color == nil {
		return
	}
	sp.SolidFill = &slide.XSolidFill{SrgbClr: srgbFrom(f.color.resolve(t))}
}

// noFill is an explicit "no interior fill".
type noFill struct{}

// NoFill returns an explicit empty fill (<a:noFill/>) — the shape is
// transparent rather than inheriting a style fill.
func NoFill() Fill { return noFill{} }

func (noFill) applyFill(sp *slide.XShapeProperties, _ *Theme) {
	sp.SolidFill = nil
	sp.NoFill = &slide.XNoFill{}
}

// Line is a shape's outline. A zero Line (Width 0, nil Color) leaves the
// outline unset (the shape inherits its style line).
type Line struct {
	// Width is the stroke width in EMU.
	Width EMU
	// Color is the stroke color (literal or theme token); nil leaves it unset.
	Color Color
	// Dash is an optional preset dash style ("dash", "dot", "sysDash", …);
	// empty is solid.
	Dash string
}

// isZero reports whether the line carries no outline information.
func (ln Line) isZero() bool { return ln.Width == 0 && ln.Color == nil && ln.Dash == "" }

func (ln Line) apply(sp *slide.XShapeProperties, t *Theme) {
	if ln.isZero() {
		return
	}
	x := &slide.XLineProperties{Width: int(ln.Width)}
	if ln.Color != nil {
		x.SolidFill = &slide.XSolidFill{SrgbClr: srgbFrom(ln.Color.resolve(t))}
	}
	if ln.Dash != "" {
		x.PrstDash = &slide.XPrstDash{Val: ln.Dash}
	}
	sp.Line = x
}

// srgbFrom bridges a resolved color to the OOXML <a:srgbClr> wire element,
// attaching an <a:alpha> child only when the color is not fully opaque.
func srgbFrom(rc resolvedColor) *slide.XSrgbClr {
	c := &slide.XSrgbClr{Val: string(rc.rgb)}
	if rc.alpha != AlphaOpaque {
		c.Alpha = &slide.XAlpha{Val: rc.alpha}
	}
	return c
}
