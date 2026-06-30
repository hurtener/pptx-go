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
//
// The read accessors (Kind / SolidColor / Gradient) let a fill recovered from a
// reopened deck be inspected (RFC §16). A reopened fill surfaces resolved
// literal colors — theme tokens are resolved to sRGB at write time (D-030), so
// the slide carries no token to reconstruct.
type Fill interface {
	applyFill(sp *slide.XShapeProperties, t *Theme)

	// Kind reports the fill variety, for reading a (reopened) deck.
	Kind() FillKind
	// SolidColor returns the fill color and true when Kind is FillSolid.
	SolidColor() (Color, bool)
	// Gradient returns the gradient description and true when Kind is FillGradient.
	Gradient() (GradientRead, bool)
}

// FillKind discriminates a Fill read back from a deck.
type FillKind int

const (
	// FillSolid is a single-color fill (SolidFill).
	FillSolid FillKind = iota
	// FillNone is an explicit empty fill (NoFill / <a:noFill/>).
	FillNone
	// FillGradient is a linear or radial gradient fill.
	FillGradient
)

// GradientRead is the readable description of a gradient fill (LinearGradient or
// RadialGradient) recovered from a deck.
type GradientRead struct {
	// Stops are the gradient color stops in document order (Pos 0..1).
	Stops []GradientStop
	// Angle is the linear gradient angle in degrees clockwise from the positive
	// x-axis; it is 0 for a radial gradient.
	Angle float64
	// Radial reports whether the gradient is radial (RadialGradient).
	Radial bool
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

func (solidFill) Kind() FillKind                 { return FillSolid }
func (f solidFill) SolidColor() (Color, bool)    { return f.color, true }
func (solidFill) Gradient() (GradientRead, bool) { return GradientRead{}, false }

// GradientStop is a color at a position (Pos 0..1) along a gradient.
type GradientStop struct {
	Pos   float64
	Color Color
}

// GradientSpec is a named brand gradient (R8.5): an ordered stop list plus a
// linear angle and a linear/radial flag, stored on a Theme under a name and
// requested by a scene Background's GradientName. Each stop's Color is a
// pptx.Color, so a soul can pin an exact brand hue with an RGB literal
// (variant-independent) or follow the active theme with a TokenColor. Angle is
// the linear gradient angle in degrees clockwise from the positive x-axis and is
// ignored when Radial is true. It has no theme1.xml slot — the resolved gradient
// fill round-trips, the named spec does not (like DarkColors / Accents).
type GradientSpec struct {
	Stops  []GradientStop
	Angle  int
	Radial bool
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

func (gradientFill) Kind() FillKind            { return FillGradient }
func (gradientFill) SolidColor() (Color, bool) { return nil, false }
func (f gradientFill) Gradient() (GradientRead, bool) {
	return GradientRead{Stops: f.stops, Angle: f.angle, Radial: f.radial}, true
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

func (noFill) Kind() FillKind                 { return FillNone }
func (noFill) SolidColor() (Color, bool)      { return nil, false }
func (noFill) Gradient() (GradientRead, bool) { return GradientRead{}, false }

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

// colorFromSrgb is srgbFrom's read inverse: it reconstructs a literal Color from
// a reopened <a:srgbClr>. A reopened deck carries resolved sRGB (theme tokens
// resolve at write time, D-030), so the read model surfaces literals — a
// fully-opaque color as a bare RGB, a translucent one as a literalColor carrying
// the alpha, mirroring how SolidFill(RGB(...)) and SolidFill(RGBA(...)) author
// them so a reopened fill compares field-equal to the authored one.
func colorFromSrgb(x *slide.XSrgbClr) Color {
	if x == nil {
		return nil
	}
	hex := RGB(x.Val)
	if x.Alpha == nil || x.Alpha.Val == AlphaOpaque {
		return hex
	}
	return literalColor{rgb: hex, alpha: x.Alpha.Val}
}

// fillFromX reconstructs the public Fill from a reopened shape's <spPr>, or nil
// when the shape carries no explicit fill (it inherits its style fill). It is
// the read inverse of the applyFill family.
func fillFromX(spPr *slide.XShapeProperties) Fill {
	if spPr == nil {
		return nil
	}
	switch {
	case spPr.SolidFill != nil:
		return solidFill{color: colorFromSrgb(spPr.SolidFill.SrgbClr)}
	case spPr.GradientFill != nil:
		return gradientFromX(spPr.GradientFill)
	case spPr.NoFill != nil:
		return noFill{}
	default:
		return nil
	}
}

// gradientFromX reconstructs a gradientFill from a reopened <a:gradFill>: the
// OOXML stop positions (×100000) and linear angle (1/60000°) map back to the
// Pos 0..1 / degrees the builder authored, so a reopened gradient compares
// field-equal to LinearGradient / RadialGradient input.
func gradientFromX(g *slide.XGradientFill) Fill {
	f := gradientFill{stops: make([]GradientStop, len(g.GsLst.Gs))}
	for i, s := range g.GsLst.Gs {
		f.stops[i] = GradientStop{
			Pos:   float64(s.Pos) / 100000.0,
			Color: colorFromSrgb(s.SrgbClr),
		}
	}
	switch {
	case g.Path != nil:
		f.radial = true
	case g.Lin != nil:
		f.angle = float64(g.Lin.Ang) / 60000.0
	}
	return f
}
