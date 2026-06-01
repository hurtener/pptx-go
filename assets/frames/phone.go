package frames

import "github.com/hurtener/pptx-go/pptx"

// Phone draws a phone bezel: a tall rounded slab with a status strip reserved
// at the top and a home-indicator pill near the bottom. The interior is the
// central viewport between them, inset on the sides by the side bezel. The
// status "notch" is approximated by reserved space (a true subtractive cutout
// needs the Phase 12 path translator).
func Phone(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int) {
	// Device slab (heavily rounded).
	sl.AddShape(pptx.ShapeRoundRect, region, fill(pptx.ColorSurfaceAlt), pptx.WithRadius(pptx.RadiusLG))
	shapes++

	statusH := maxEMU(region.H*6/100, pptx.In(0.12))
	bottomH := maxEMU(region.H*8/100, pptx.In(0.16))
	sideInset := maxEMU(region.W*4/100, pptx.In(0.06))

	// Home-indicator pill, centered in the bottom band.
	pillW := region.W * 30 / 100
	pillH := maxEMU(region.H*1/100, pptx.Pt(3))
	pill := pptx.Box{
		X: region.X + (region.W-pillW)/2,
		Y: region.Bottom() - bottomH + (bottomH-pillH)/2,
		W: pillW,
		H: pillH,
	}
	sl.AddShape(pptx.ShapeRoundRect, pill, fill(pptx.ColorSurface), pptx.WithRadius(pptx.RadiusFull))
	shapes++

	interior = pptx.Box{
		X: region.X + sideInset,
		Y: region.Y + statusH,
		W: region.W - 2*sideInset,
		H: region.H - statusH - bottomH,
	}
	return interior, shapes
}
