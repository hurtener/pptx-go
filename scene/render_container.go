package scene

import (
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/layout"
)

// Container composers (RFC §11.2 / §12). A container emits no shape of its own:
// it subdivides its slot via the scene/layout geometry engine and renders each
// child into a sub-slot through the normal dispatch, so nesting composes.

func (r *renderer) renderTwoColumn(ps *pptx.Slide, box pptx.Box, v TwoColumn, slideID string) {
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	cols := layout.Columns(box, ratioWeights(v.Ratio), gap)
	if len(cols) != 2 {
		return
	}
	for _, pl := range r.stackIn(cols[0], v.Left, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
	for _, pl := range r.stackIn(cols[1], v.Right, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
}

func (r *renderer) renderGrid(ps *pptx.Slide, box pptx.Box, v Grid, slideID string) {
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	cells := layout.Grid(box, v.Columns, v.Ratio, gap, len(v.Cells))
	for i, n := range v.Cells {
		if i >= len(cells) {
			break
		}
		r.renderNode(ps, cells[i], n, slideID, HAlignLeft)
	}
}

// ratioWeights maps a two_column ratio to per-column weights.
func ratioWeights(rt ColumnRatio) []int {
	switch rt {
	case Ratio12:
		return []int{1, 2}
	case Ratio21:
		return []int{2, 1}
	default: // Ratio11
		return []int{1, 1}
	}
}
