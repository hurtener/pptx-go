package ornaments

import "github.com/hurtener/pptx-go/pptx"

// CornerBracket draws an L-shaped bracket at the box's top-left: a horizontal
// and a vertical accent bar. Rotation is ignored (a two-bar bracket has no clean
// per-shape rotation without a group transform — D-041).
func CornerBracket(sl *pptx.Slide, box pptx.Box, alpha int, _ float64) int {
	thick := minEMU(box.W, box.H) / 12
	if thick < pptx.Pt(2) {
		thick = pptx.Pt(2)
	}
	arm := minEMU(box.W, box.H) / 2
	sl.AddShape(pptx.ShapeRect, pptx.Box{X: box.X, Y: box.Y, W: arm, H: thick}, accent(alpha))
	sl.AddShape(pptx.ShapeRect, pptx.Box{X: box.X, Y: box.Y, W: thick, H: arm}, accent(alpha))
	return 2
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
