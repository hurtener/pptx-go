package frames

import "github.com/hurtener/pptx-go/pptx"

// Laptop draws a laptop bezel: a rounded screen panel over a wide, thin base
// deck (the keyboard half). The interior is the viewport inside the screen
// panel, inset by a thin bezel. The base spans the region width so every bezel
// shape stays within the region.
func Laptop(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int) {
	screenH := region.H * 88 / 100

	// Screen panel.
	screen := pptx.Box{X: region.X, Y: region.Y, W: region.W, H: screenH}
	sl.AddShape(pptx.ShapeRoundRect, screen, fill(pptx.ColorSurfaceAlt), pptx.WithRadius(pptx.RadiusMD))
	shapes++

	// Base deck (full width, thin).
	baseH := region.H - screenH
	base := pptx.Box{X: region.X, Y: region.Y + screenH, W: region.W, H: baseH}
	sl.AddShape(pptx.ShapeRoundRect, base, fill(pptx.ColorSurfaceAlt), pptx.WithRadius(pptx.RadiusSM))
	shapes++

	bezel := maxEMU(region.W*3/100, pptx.In(0.04))
	interior = pptx.Box{
		X: region.X + bezel,
		Y: region.Y + bezel,
		W: region.W - 2*bezel,
		H: screenH - 2*bezel,
	}
	return interior, shapes
}
