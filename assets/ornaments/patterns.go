package ornaments

import "github.com/hurtener/pptx-go/pptx"

// GridDots draws a regular dotted texture: a 6×4 lattice of small role-colored
// dots centered in their cells. Deterministic; rotation is ignored (a dot grid is
// symmetric and the builder has no group transform — D-041).
func GridDots(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole) int {
	const cols, rows = 6, 4
	cellW := box.W / cols
	cellH := box.H / rows
	dot := minEMU(cellW, cellH) / 5
	if dot < pptx.Pt(2) {
		dot = pptx.Pt(2)
	}
	n := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			x := box.X + cellW*pptx.EMU(c) + cellW/2 - dot/2
			y := box.Y + cellH*pptx.EMU(r) + cellH/2 - dot/2
			sl.AddShape(pptx.ShapeEllipse, pptx.Box{X: x, Y: y, W: dot, H: dot}, roleFill(role, alpha))
			n++
		}
	}
	return n
}

// NoiseOverlay draws a subtle grain: a deterministic sparse scatter of tiny
// low-alpha role-colored dots over a fixed lattice (no per-pixel noise natively —
// a documented approximation, D-041). The opacity is divided down so the grain
// stays faint.
func NoiseOverlay(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole) int {
	const cols, rows = 12, 8
	a := alpha / 3
	if a < 1 {
		a = 1
	}
	dot := pptx.Pt(2)
	cellW := box.W / cols
	cellH := box.H / rows
	n := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if (r*cols+c)%2 == 0 { // sparse: ~half the cells
				continue
			}
			// Fixed per-cell offset sequence (no RNG) so output is byte-identical.
			ox := cellW * pptx.EMU((c*7+r*3)%5) / 6
			oy := cellH * pptx.EMU((r*5+c*2)%5) / 6
			x := box.X + cellW*pptx.EMU(c) + ox
			y := box.Y + cellH*pptx.EMU(r) + oy
			sl.AddShape(pptx.ShapeEllipse, pptx.Box{X: x, Y: y, W: dot, H: dot}, roleFill(role, a))
			n++
		}
	}
	return n
}
