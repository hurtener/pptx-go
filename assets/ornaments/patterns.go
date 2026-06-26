package ornaments

import "github.com/hurtener/pptx-go/pptx"

// PatternMaxDots caps the dots a pattern recipe emits, so a fine pitch over a
// full-bleed box cannot explode the part size (R13.7, D-111). It is the single
// source of truth for the cap: scene/render_decoration.go's ornamentPatternCap
// references it so the warn-threshold can never drift from the recipe cap (D-115).
const PatternMaxDots = 2000

// patternMaxDots is the internal alias used by the recipes below.
const patternMaxDots = PatternMaxDots

// patternDims returns the column and row counts for a pattern lattice: a
// box-derived count (cols = box.W/pitch) when pitch > 0, else the legacy fixed
// (defCols, defRows) so a pitch-unset decoration is byte-identical (R13.7, D-111).
func patternDims(box pptx.Box, pitch pptx.EMU, defCols, defRows int) (cols, rows int) {
	if pitch <= 0 {
		return defCols, defRows
	}
	cols = int(box.W / pitch)
	rows = int(box.H / pitch)
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	return cols, rows
}

// GridDots draws a regular dotted texture: a lattice of small role-colored dots
// centered in their cells. The lattice is 6×4 by default, or box-derived at the
// caller pitch (R13.7) so a full-bleed texture keeps a consistent spacing.
// Deterministic; rotation is ignored (a dot grid is symmetric and the builder has
// no group transform — D-041). Capped at patternMaxDots.
func GridDots(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole, pitch pptx.EMU) int {
	cols, rows := patternDims(box, pitch, 6, 4)
	cellW := box.W / pptx.EMU(cols)
	cellH := box.H / pptx.EMU(rows)
	dot := minEMU(cellW, cellH) / 5
	if dot < pptx.Pt(2) {
		dot = pptx.Pt(2)
	}
	n := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if n >= patternMaxDots {
				return n
			}
			x := box.X + cellW*pptx.EMU(c) + cellW/2 - dot/2
			y := box.Y + cellH*pptx.EMU(r) + cellH/2 - dot/2
			sl.AddShape(pptx.ShapeEllipse, pptx.Box{X: x, Y: y, W: dot, H: dot}, roleFill(role, alpha))
			n++
		}
	}
	return n
}

// NoiseOverlay draws a subtle grain: a deterministic sparse scatter of tiny
// low-alpha role-colored dots over a lattice (no per-pixel noise natively — a
// documented approximation, D-041). The lattice is 12×8 by default, or box-derived
// at the caller pitch (R13.7). The opacity is divided down so the grain stays
// faint. Capped at patternMaxDots.
func NoiseOverlay(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole, pitch pptx.EMU) int {
	cols, rows := patternDims(box, pitch, 12, 8)
	a := alpha / 3
	if a < 1 {
		a = 1
	}
	dot := pptx.Pt(2)
	cellW := box.W / pptx.EMU(cols)
	cellH := box.H / pptx.EMU(rows)
	n := 0
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if n >= patternMaxDots {
				return n
			}
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

// starfieldPitch is the default lattice spacing when no caller pitch is set; the
// dot count then scales with the box (R13.6, D-110).
const starfieldPitch = pptx.EMU(457200) // In(0.5)

// Starfield draws an organic scatter of role-colored dots — a sparse, irregular
// starfield with per-dot size and alpha variance (D-110). Placement is a lattice
// (the caller pitch, or starfieldPitch when 0 — R13.7) perturbed by a fixed
// integer hash of the cell index, so the dot count scales with the box (a
// full-bleed box gets a dense field, a small box a sparse one) and two renders
// are byte-identical (no RNG/clock — D-035). Rotation is ignored (a scatter is
// symmetric). The total is capped at patternMaxDots to protect the part size.
func Starfield(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole, pitch pptx.EMU) int {
	return scatter(sl, box, alpha, role, pitch, scatterDot)
}

// scatterShape draws one scatter mark of side d at (x,y) in role at alpha a.
type scatterShape func(sl *pptx.Slide, x, y, d pptx.EMU, role pptx.ColorRole, a int)

func scatterDot(sl *pptx.Slide, x, y, d pptx.EMU, role pptx.ColorRole, a int) {
	sl.AddShape(pptx.ShapeEllipse, pptx.Box{X: x, Y: y, W: d, H: d}, roleFill(role, a))
}

func scatterStar(sl *pptx.Slide, x, y, d pptx.EMU, role pptx.ColorRole, a int) {
	sl.AddShape(pptx.ShapeGeometry("star5"), pptx.Box{X: x, Y: y, W: d, H: d}, roleFill(role, a))
}

func scatterPlus(sl *pptx.Slide, x, y, d pptx.EMU, role pptx.ColorRole, a int) {
	sl.AddShape(pptx.ShapeGeometry("mathPlus"), pptx.Box{X: x, Y: y, W: d, H: d}, roleFill(role, a))
}

func scatterRing(sl *pptx.Slide, x, y, d pptx.EMU, role pptx.ColorRole, a int) {
	sl.AddShape(pptx.ShapeEllipse, pptx.Box{X: x, Y: y, W: d, H: d},
		pptx.WithFill(pptx.NoFill()), pptx.WithLine(pptx.Line{Width: pptx.Pt(1.25), Color: pptx.TokenColorAlpha(role, a)}))
}

// ScatterDot / ScatterStar / ScatterPlus / ScatterRing are the scatter ornament
// FAMILY (R14.20, D-131): one deterministic hash-of-index placement engine
// (shared with Starfield) parameterized by mark shape, so a starfield, a dust
// field, a confetti of plusses, or a bokeh of rings all draw from one recipe.
func ScatterDot(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole, pitch pptx.EMU) int {
	return scatter(sl, box, alpha, role, pitch, scatterDot)
}

func ScatterStar(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole, pitch pptx.EMU) int {
	return scatter(sl, box, alpha, role, pitch, scatterStar)
}

func ScatterPlus(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole, pitch pptx.EMU) int {
	return scatter(sl, box, alpha, role, pitch, scatterPlus)
}

func ScatterRing(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole, pitch pptx.EMU) int {
	return scatter(sl, box, alpha, role, pitch, scatterRing)
}

// scatter is the shared scatter-family placement engine (R14.20, generalizing
// D-110's starfield): a lattice (caller pitch, or starfieldPitch when 0)
// perturbed by a fixed integer hash of the cell index, with per-mark size + alpha
// variance, ~20% of cells sieved empty for irregularity. No RNG/clock → two
// renders are byte-identical (D-035). Capped at patternMaxDots. The shape draws
// the mark; `scatterDot` reproduces the original starfield exactly.
func scatter(sl *pptx.Slide, box pptx.Box, alpha int, role pptx.ColorRole, pitch pptx.EMU, shape scatterShape) int {
	p := pitch
	if p <= 0 {
		p = starfieldPitch
	}
	cols := int(box.W / p)
	rows := int(box.H / p)
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
			if n >= patternMaxDots {
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
			ox := cellW * pptx.EMU((h/3)%5) / 6
			oy := cellH * pptx.EMU((h/11)%5) / 6
			x := box.X + cellW*pptx.EMU(c) + ox
			y := box.Y + cellH*pptx.EMU(r) + oy
			shape(sl, x, y, dot, role, a)
			n++
		}
	}
	return n
}
