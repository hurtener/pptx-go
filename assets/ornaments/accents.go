package ornaments

import (
	"math"

	"github.com/hurtener/pptx-go/pptx"
)

// CornerBracket draws an L-shaped bracket hugging one corner of the box. Rotation
// orients it (snapped to 0/90/180/270 — the bracket is two bars repositioned, so
// the orientation is exact without a group transform): 0 = top-left, 90 =
// top-right, 180 = bottom-right, 270 = bottom-left. Pair the rotation with the
// Decoration's corner anchor to frame a slide (e.g. AnchorBottomRight + 180).
func CornerBracket(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64) int {
	thick := minEMU(box.W, box.H) / 12
	if thick < pptx.Pt(2) {
		thick = pptx.Pt(2)
	}
	arm := minEMU(box.W, box.H) / 2
	h, v := cornerArms(box, thick, arm, quadrant(rotationDeg))
	sl.AddShape(pptx.ShapeRect, h, accent(alpha))
	sl.AddShape(pptx.ShapeRect, v, accent(alpha))
	return 2
}

// quadrant snaps a rotation to one of 0,1,2,3 (× 90° clockwise).
func quadrant(deg float64) int {
	d := math.Mod(deg, 360)
	if d < 0 {
		d += 360
	}
	return int(math.Round(d/90)) % 4
}

// cornerArms returns the horizontal and vertical bar boxes for an L bracket
// whose corner sits at the box corner selected by q (0 TL, 1 TR, 2 BR, 3 BL).
func cornerArms(box pptx.Box, thick, arm pptx.EMU, q int) (h, v pptx.Box) {
	switch q {
	case 1: // top-right
		h = pptx.Box{X: box.Right() - arm, Y: box.Y, W: arm, H: thick}
		v = pptx.Box{X: box.Right() - thick, Y: box.Y, W: thick, H: arm}
	case 2: // bottom-right
		h = pptx.Box{X: box.Right() - arm, Y: box.Bottom() - thick, W: arm, H: thick}
		v = pptx.Box{X: box.Right() - thick, Y: box.Bottom() - arm, W: thick, H: arm}
	case 3: // bottom-left
		h = pptx.Box{X: box.X, Y: box.Bottom() - thick, W: arm, H: thick}
		v = pptx.Box{X: box.X, Y: box.Bottom() - arm, W: thick, H: arm}
	default: // 0: top-left
		h = pptx.Box{X: box.X, Y: box.Y, W: arm, H: thick}
		v = pptx.Box{X: box.X, Y: box.Y, W: thick, H: arm}
	}
	return h, v
}

// ChevronArrow draws a directional chevron accent. It is a single shape, so it
// honors the caller rotation (a chevron is the ornament rotation is meant for).
func ChevronArrow(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64) int {
	opts := []pptx.ShapeOption{accent(alpha)}
	if rotationDeg != 0 {
		opts = append(opts, pptx.WithRotation(rotationDeg))
	}
	sl.AddShape(pptx.ShapeChevron, box, opts...)
	return 1
}
