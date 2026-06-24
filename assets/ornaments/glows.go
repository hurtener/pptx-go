package ornaments

import "github.com/hurtener/pptx-go/pptx"

// RadialGlow draws a soft circular glow: one ellipse filled with a radial
// gradient from role (center, at the caller opacity) to fully transparent
// at the edge. Rotation is moot for a circle and ignored.
func RadialGlow(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole) int {
	sl.AddShape(pptx.ShapeEllipse, box, pptx.WithFill(pptx.RadialGradient(
		pptx.GradientStop{Pos: 0, Color: pptx.TokenColorAlpha(role, alpha)},
		pptx.GradientStop{Pos: 1, Color: pptx.TokenColorAlpha(role, 0)},
	)))
	return 1
}

// GlowRing draws a halo ring: one ellipse with a radial gradient that is
// transparent at the center, bright (role at the caller opacity) in a band
// near the rim, and transparent again at the edge — a glowing ring.
func GlowRing(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole) int {
	sl.AddShape(pptx.ShapeEllipse, box, pptx.WithFill(pptx.RadialGradient(
		pptx.GradientStop{Pos: 0, Color: pptx.TokenColorAlpha(role, 0)},
		pptx.GradientStop{Pos: 0.55, Color: pptx.TokenColorAlpha(role, 0)},
		pptx.GradientStop{Pos: 0.78, Color: pptx.TokenColorAlpha(role, alpha)},
		pptx.GradientStop{Pos: 1, Color: pptx.TokenColorAlpha(role, 0)},
	)))
	return 1
}
