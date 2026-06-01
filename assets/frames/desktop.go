package frames

import "github.com/hurtener/pptx-go/pptx"

// Desktop draws a monitor bezel: a rounded display panel over a stand neck and
// a foot. The interior is the viewport inside the panel, inset by a thin bezel.
// The stand (neck + foot) sits in the bottom band of the region, so the
// interior never overlaps it.
func Desktop(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int) {
	panelH := region.H * 78 / 100

	// Display panel.
	panel := pptx.Box{X: region.X, Y: region.Y, W: region.W, H: panelH}
	sl.AddShape(pptx.ShapeRoundRect, panel, fill(pptx.ColorSurfaceAlt), pptx.WithRadius(pptx.RadiusMD))
	shapes++

	// Stand neck.
	neckW := maxEMU(region.W*8/100, pptx.In(0.2))
	neckH := region.H * 14 / 100
	neck := pptx.Box{
		X: region.X + (region.W-neckW)/2,
		Y: region.Y + panelH,
		W: neckW,
		H: neckH,
	}
	sl.AddShape(pptx.ShapeRect, neck, fill(pptx.ColorSurfaceAlt))
	shapes++

	// Foot.
	footW := region.W * 30 / 100
	footH := maxEMU(region.H*6/100, pptx.In(0.12))
	foot := pptx.Box{
		X: region.X + (region.W-footW)/2,
		Y: region.Bottom() - footH,
		W: footW,
		H: footH,
	}
	sl.AddShape(pptx.ShapeRoundRect, foot, fill(pptx.ColorSurfaceAlt), pptx.WithRadius(pptx.RadiusSM))
	shapes++

	bezel := maxEMU(region.W*3/100, pptx.In(0.04))
	interior = pptx.Box{
		X: region.X + bezel,
		Y: region.Y + bezel,
		W: region.W - 2*bezel,
		H: panelH - 2*bezel,
	}
	return interior, shapes
}
