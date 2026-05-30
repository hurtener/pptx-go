package layout

import "github.com/hurtener/pptx-go/pptx"

// Grid splits parent into a row-major grid of count cells: cols columns (widths
// from colWeights, or equal when empty) and rows = ⌈count/cols⌉ equal-height
// rows, separated by gap on both axes. Returns one Box per cell, in row-major
// order, len == count. The last row's height absorbs any rounding remainder.
func Grid(parent pptx.Box, cols int, colWeights []int, gap pptx.EMU, count int) []pptx.Box {
	if count <= 0 {
		return nil
	}
	if cols < 1 {
		cols = 1
	}
	rows := (count + cols - 1) / cols

	weights := colWeights
	if len(weights) != cols {
		weights = equalWeights(cols) // empty or mismatched → equal columns
	}
	colBoxes := Columns(parent, weights, gap)
	if len(colBoxes) == 0 {
		return nil
	}

	availH := parent.H - gap*pptx.EMU(rows-1)
	if availH < 0 {
		availH = 0
	}

	out := make([]pptx.Box, 0, count)
	for i := 0; i < count; i++ {
		rIdx := i / cols
		cIdx := i % cols
		cb := colBoxes[cIdx]

		var rowH pptx.EMU
		if rIdx == rows-1 {
			rowH = availH - pptx.EMU(rows-1)*(availH/pptx.EMU(rows)) // remainder row
		} else {
			rowH = availH / pptx.EMU(rows)
		}
		y := parent.Y + pptx.EMU(rIdx)*(availH/pptx.EMU(rows)+gap)

		out = append(out, pptx.Box{X: cb.X, Y: y, W: cb.W, H: rowH})
	}
	return out
}
