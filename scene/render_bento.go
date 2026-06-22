package scene

import (
	"github.com/hurtener/pptx-go/pptx"
)

// Bento composer (RFC §11.2, D-056, D-072). A Bento subdivides its slot into
// rows; each row carries an optional left-gutter label and a left-to-right
// sequence of cells whose widths come from their column span on a shared column
// grid (a span-S cell is S unit-widths, so columns align across rows). Rows are
// equal-height by default; with WeightedRows they size to their content's
// preferred height, clamped to fit the region (D-072). Like other containers it
// emits no shape of its own beyond the labels — each cell renders through the
// normal dispatch. Deterministic integer-EMU geometry (D-035).

// bentoGutterW is the fixed width reserved for the left row-label gutter (only
// when at least one row is labeled).
const bentoGutterW = pptx.EMU(1097280) // ~1.2"

// bentoColumns computes a bento's shared horizontal geometry: the left-gutter
// width (non-zero iff any row is labeled), the content origin X, and the unit
// column width. Pure integer-EMU; shared by bentoGeometry and the weighted
// row-height pass so the two never drift.
func bentoColumns(box pptx.Box, v Bento, gap pptx.EMU) (gutterW, contentX, unitW pptx.EMU) {
	cols := v.Columns
	if cols < 1 {
		cols = 1
	}
	for _, row := range v.Rows {
		if row.Label != "" {
			gutterW = bentoGutterW
			break
		}
	}
	contentX = box.X
	contentW := box.W
	if gutterW > 0 {
		contentX = box.X + gutterW + gap
		contentW = box.W - gutterW - gap
		if contentW < 0 {
			contentW = 0
		}
	}
	unitW = (contentW - gap*pptx.EMU(cols-1)) / pptx.EMU(cols)
	if unitW < 0 {
		unitW = 0
	}
	return gutterW, contentX, unitW
}

// cellWidth returns the EMU width of a span-S cell: S unit columns plus the
// inter-unit gaps they bridge. A span < 1 is treated as 1.
func cellWidth(span int, unitW, gap pptx.EMU) pptx.EMU {
	if span < 1 {
		span = 1
	}
	return pptx.EMU(span)*unitW + gap*pptx.EMU(span-1)
}

// bentoGeometry computes a bento's deterministic layout: the left-gutter width,
// the per-row top Y, the per-row height, and one Box per cell (row-major). It is
// a pure function — no shapes emitted — so the span/alignment geometry is
// unit-testable. rowHs supplies per-row heights (len == nRows) for the
// content-weighted mode; when nil/mismatched, every row gets the equal height
// (box.H − gaps)/nRows, reproducing the pre-R10.3 layout byte-for-byte. Returns
// nil for a degenerate bento.
func bentoGeometry(box pptx.Box, v Bento, gap pptx.EMU, rowHs []pptx.EMU) (gutterW pptx.EMU, rowYs, heights []pptx.EMU, cells [][]pptx.Box) {
	cols := v.Columns
	nRows := len(v.Rows)
	if cols < 1 || nRows == 0 {
		return 0, nil, nil, nil
	}
	gutterW, contentX, unitW := bentoColumns(box, v, gap)

	heights = make([]pptx.EMU, nRows)
	if len(rowHs) == nRows {
		copy(heights, rowHs)
	} else {
		rowH := (box.H - gap*pptx.EMU(nRows-1)) / pptx.EMU(nRows)
		if rowH < 0 {
			rowH = 0
		}
		for ri := range heights {
			heights[ri] = rowH
		}
	}

	rowYs = make([]pptx.EMU, nRows)
	cells = make([][]pptx.Box, nRows)
	y := box.Y
	for ri, row := range v.Rows {
		rowYs[ri] = y
		x := contentX
		boxes := make([]pptx.Box, 0, len(row.Cells))
		for _, cell := range row.Cells {
			cw := cellWidth(cell.Span, unitW, gap)
			boxes = append(boxes, pptx.Box{X: x, Y: y, W: cw, H: heights[ri]})
			x += cw + gap
		}
		cells[ri] = boxes
		y += heights[ri] + gap
	}
	return gutterW, rowYs, heights, cells
}

// bentoWeightedRowHeights returns the per-row heights for the content-weighted
// mode: each row's height is the preferred height of its tallest cell at that
// cell's span width. When the rows plus gaps would overflow the region, every
// row is scaled down by a single basis-point factor so the total fits exactly
// (flooring guarantees Σ ≤ avail — no row clips off-slide). When they fit, rows
// keep their preferred height (top-aligned; leftover slack is bottom whitespace).
// Deterministic integer / basis-point math.
func (r *renderer) bentoWeightedRowHeights(box pptx.Box, v Bento, gap pptx.EMU) []pptx.EMU {
	nRows := len(v.Rows)
	if nRows == 0 {
		return nil
	}
	_, _, unitW := bentoColumns(box, v, gap)

	pref := make([]pptx.EMU, nRows)
	var sum pptx.EMU
	for ri, row := range v.Rows {
		var h pptx.EMU
		for _, cell := range row.Cells {
			if ph := preferredHeight(cell.Node, cellWidth(cell.Span, unitW, gap), r.theme); ph > h {
				h = ph
			}
		}
		pref[ri] = h
		sum += h
	}

	avail := box.H - gap*pptx.EMU(nRows-1)
	if avail < 0 {
		avail = 0
	}
	if sum > avail && sum > 0 {
		const full = 10000
		sBP := avail * full / sum // floors → Σ(pref·sBP/full) ≤ avail
		if sBP < 1 {
			sBP = 1
		}
		for ri := range pref {
			pref[ri] = pref[ri] * sBP / full
		}
	}
	return pref
}

func (r *renderer) renderBento(ps *pptx.Slide, box pptx.Box, v Bento, slideID string) {
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	var rowHs []pptx.EMU
	if v.WeightedRows {
		rowHs = r.bentoWeightedRowHeights(box, v, gap)
	}
	gutterW, rowYs, heights, cells := bentoGeometry(box, v, gap, rowHs)
	if cells == nil {
		return
	}
	for ri, row := range v.Rows {
		if row.Label != "" && gutterW > 0 {
			lb := pptx.Box{X: box.X, Y: rowYs[ri], W: gutterW, H: heights[ri]}
			tf := ps.AddTextFrame(lb).Anchor(pptx.AnchorMiddle)
			p := tf.AddParagraph(pptx.ParagraphOpts{})
			p.AddRun(row.Label, pptx.RunStyle{TypeRole: pptx.TypeCaption, Bold: true, Color: pptx.TokenTextColor(pptx.TextMuted)})
			r.stats.Shapes++
		}
		for ci, cell := range row.Cells {
			r.renderNode(ps, cells[ri][ci], cell.Node, slideID, HAlignLeft)
		}
	}
}
