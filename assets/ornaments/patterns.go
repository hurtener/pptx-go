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

// Starfield geometry (R13.6, D-110). Pinned layout metrics — a base lattice
// pitch, the per-dot size table, the per-dot alpha-percent table, and a dot cap.
// Not theme tokens (geometry); the dot color is a token role.
var (
	starfieldSizes    = []pptx.EMU{pptx.Pt(1), pptx.Pt(2), pptx.Pt(3)} // >=2 distinct dot sizes
	starfieldAlphaPct = []int{35, 60, 100}                             // >=2 distinct dot alphas (% of caller alpha)
)

const (
	starfieldPitch   = pptx.EMU(457200) // In(0.5); base lattice spacing — count scales with the box
	starfieldMaxDots = 2000             // cap to protect file size on a huge full-bleed box
)

// Starfield draws an organic scatter of role-colored dots — a sparse, irregular
// starfield with per-dot size and alpha variance (D-110). Placement is a
// box-derived lattice (pitch starfieldPitch) perturbed by a fixed integer hash of
// the cell index, so the dot count scales with the box (a full-bleed box gets a
// dense field, a small box a sparse one) and two renders are byte-identical (no
// RNG/clock — D-035). Rotation is ignored (a scatter is symmetric). The total is
// capped at starfieldMaxDots to protect the part size.
func Starfield(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole) int {
	cols := int(box.W / starfieldPitch)
	rows := int(box.H / starfieldPitch)
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	cellW := box.W / pptx.EMU(cols)
	cellH := box.H / pptx.EMU(rows)
	n := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if n >= starfieldMaxDots {
				return n
			}
			h := (r*131 + c*61 + r*c*17) & 0x7fffffff
			if h%5 == 0 { // ~20% of cells empty → irregular sparseness
				continue
			}
			dot := starfieldSizes[h%len(starfieldSizes)]
			a := alpha * starfieldAlphaPct[(h/7)%len(starfieldAlphaPct)] / 100
			if a < 1 {
				a = 1
			}
			// Fixed per-cell offset (no RNG) so output is byte-identical.
			ox := cellW * pptx.EMU((h/3)%5) / 6
			oy := cellH * pptx.EMU((h/11)%5) / 6
			x := box.X + cellW*pptx.EMU(c) + ox
			y := box.Y + cellH*pptx.EMU(r) + oy
			sl.AddShape(pptx.ShapeEllipse, pptx.Box{X: x, Y: y, W: dot, H: dot}, roleFill(role, a))
			n++
		}
	}
	return n
}
