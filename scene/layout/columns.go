// Package layout is the scene geometry engine (RFC §10.2): deterministic slot
// division of a parent box into columns and grids. It is pure geometry — it
// imports only pptx and knows nothing about node types, so the scene renderer
// (and later phases) compose it freely.
package layout

import "github.com/hurtener/pptx-go/pptx"

// Columns splits parent into len(weights) side-by-side columns separated by gap,
// each the full parent height, with widths proportional to weights. A
// non-positive or mismatched weight set falls back to equal columns. Returns one
// Box per column, left to right. Width uses floor division (deterministic; the
// last column absorbs any rounding remainder so the columns fit the parent).
func Columns(parent pptx.Box, weights []int, gap pptx.EMU) []pptx.Box {
	n := len(weights)
	if n == 0 {
		return nil
	}
	sum := 0
	for _, w := range weights {
		if w > 0 {
			sum += w
		}
	}
	if sum <= 0 {
		weights = equalWeights(n)
		sum = n
	}

	avail := parent.W - gap*pptx.EMU(n-1)
	if avail < 0 {
		avail = 0
	}

	out := make([]pptx.Box, 0, n)
	x := parent.X
	usedW := pptx.EMU(0)
	for i, w := range weights {
		if w < 0 {
			w = 0
		}
		var cw pptx.EMU
		if i == n-1 {
			cw = avail - usedW // last column takes the remainder
		} else {
			cw = avail * pptx.EMU(w) / pptx.EMU(sum)
		}
		out = append(out, pptx.Box{X: x, Y: parent.Y, W: cw, H: parent.H})
		x += cw + gap
		usedW += cw
	}
	return out
}

func equalWeights(n int) []int {
	w := make([]int, n)
	for i := range w {
		w[i] = 1
	}
	return w
}
