package pptx

import (
	"math"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// Fill is a shape's interior fill. Construct with SolidFill, NoFill,
// LinearGradient or RadialGradient. The interface is sealed (applyFill is
// unexported); a fill resolves its colors against the active theme when applied,
// so theme tokens re-render after a theme swap (P2).
//
// V1 ships SolidFill, NoFill, and gradient fills (D-041). Pattern and picture
// (blip) shape fills remain tracked separately.
type Fill interface {
	applyFill(sp *slide.XShapeProperties, t *Theme)
}

// solidFill is a single-color fill.
type solidFill struct{ color Color }

// SolidFill returns a solid fill of the given color (literal or theme token).
func SolidFill(c Color) Fill { return solidFill{color: c} }

func (f solidFill) applyFill(sp *slide.XShapeProperties, t *Theme) {
	sp.NoFill = nil
	sp.GradientFill = nil
	if f.color == nil {
		return
	}
	sp.SolidFill = &slide.XSolidFill{SrgbClr: srgbFrom(f.color.resolve(t))}
}

// GradientStop is a color at a position (Pos 0..1) along a gradient.
type GradientStop struct {
	Pos   float64
	Color Color
}

// gradientFill is a multi-stop linear or radial fill.
type gradientFill struct {
	stops  []GradientStop
	angle  float64 // linear only: degrees clockwise from the positive x-axis
	radial bool
}

// LinearGradient returns a linear gradient fill across the stops (Pos 0..1) at
// angleDeg, measured clockwise from the positive x-axis (D-041).
func LinearGradient(angleDeg float64, stops ...GradientStop) Fill {
	return gradientFill{stops: stops, angle: angleDeg}
}

// RadialGradient returns a radial gradient fill (path="circle") from the centre
// outward across the stops — the primitive behind glow ornaments (D-041).
func RadialGradient(stops ...GradientStop) Fill {
	return gradientFill{stops: stops, radial: true}
}

func (f gradientFill) applyFill(sp *slide.XShapeProperties, t *Theme) {
	sp.NoFill = nil
	sp.SolidFill = nil
	if len(f.stops) == 0 {
		sp.GradientFill = nil
		return
	}
	gs := make([]slide.XGradientStop, len(f.stops))
	for i, s := range f.stops {
		stop := slide.XGradientStop{Pos: int(math.Round(clampUnit(s.Pos) * 100000))}
		if s.Color != nil {
			stop.SrgbClr = srgbFrom(s.Color.resolve(t))
		}
		gs[i] = stop
	}
	g := &slide.XGradientFill{GsLst: slide.XGradientStopList{Gs: gs}}
	if f.radial {
		// A centred focal point (50% inset from each edge) — a circular glow.
		g.Path = &slide.XPathGradient{Path: "circle", FillToRect: &slide.XFillToRect{L: 50000, T: 50000, R: 50000, B: 50000}}
	} else {
		g.Lin = &slide.XLinearGradient{Ang: normalizeAngle60k(f.angle)}
	}
	sp.GradientFill = g
}

// clampUnit bounds a value to [0,1].
func clampUnit(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	default:
		return v
	}
}

// normalizeAngle60k converts degrees to OOXML's 1/60000-degree units in
// [0, 360°), deterministically.
func normalizeAngle60k(deg float64) int {
	a := math.Mod(deg, 360)
	if a < 0 {
		a += 360
	}
	return int(math.Round(a * 60000))
}

// noFill is an explicit "no interior fill".
type noFill struct{}

// NoFill returns an explicit empty fill (<a:noFill/>) — the shape is
// transparent rather than inheriting a style fill.
func NoFill() Fill { return noFill{} }

func (noFill) applyFill(sp *slide.XShapeProperties, _ *Theme) {
	sp.SolidFill = nil
	sp.GradientFill = nil
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
