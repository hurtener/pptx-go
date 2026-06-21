package scene

import (
	"github.com/hurtener/pptx-go/pptx"
)

// Bento composer (RFC §11.2, D-056). A Bento subdivides its slot into
// equal-height rows; each row carries an optional left-gutter label and a
// left-to-right sequence of cells whose widths come from their column span on a
// shared column grid (a span-S cell is S unit-widths, so columns align across
// rows). Like other containers it emits no shape of its own beyond the labels —
// each cell renders through the normal dispatch. Deterministic integer-EMU
// geometry (D-035).

// bentoGutterW is the fixed width reserved for the left row-label gutter (only
// when at least one row is labelled).
const bentoGutterW = pptx.EMU(1097280) // ~1.2"

// bentoGeometry computes a bento's deterministic layout: the left-gutter width
// (non-zero iff any row is labelled), the equal row height, and one Box per cell
// (row-major). It is a pure function — no shapes emitted — so the span/alignment
// geometry is unit-testable. Returns nil cells for a degenerate bento.
func bentoGeometry(box pptx.Box, v Bento, gap pptx.EMU) (gutterW, rowH pptx.EMU, cells [][]pptx.Box) {
	cols := v.Columns
	nRows := len(v.Rows)
	if cols < 1 || nRows == 0 {
		return 0, 0, nil
	}
	for _, row := range v.Rows {
		if row.Label != "" {
			gutterW = bentoGutterW
			break
		}
	}
	contentX := box.X
	contentW := box.W
	if gutterW > 0 {
		contentX = box.X + gutterW + gap
		contentW = box.W - gutterW - gap
		if contentW < 0 {
			contentW = 0
		}
	}
	rowH = (box.H - gap*pptx.EMU(nRows-1)) / pptx.EMU(nRows)
	if rowH < 0 {
		rowH = 0
	}
	unitW := (contentW - gap*pptx.EMU(cols-1)) / pptx.EMU(cols)
	if unitW < 0 {
		unitW = 0
	}

	cells = make([][]pptx.Box, nRows)
	for ri, row := range v.Rows {
		rowY := box.Y + pptx.EMU(ri)*(rowH+gap)
		x := contentX
		boxes := make([]pptx.Box, 0, len(row.Cells))
		for _, cell := range row.Cells {
			span := cell.Span
			if span < 1 {
				span = 1
			}
			cw := pptx.EMU(span)*unitW + gap*pptx.EMU(span-1)
			boxes = append(boxes, pptx.Box{X: x, Y: rowY, W: cw, H: rowH})
			x += cw + gap
		}
		cells[ri] = boxes
	}
	return gutterW, rowH, cells
}

func (r *renderer) renderBento(ps *pptx.Slide, box pptx.Box, v Bento, slideID string) {
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	gutterW, rowH, cells := bentoGeometry(box, v, gap)
	if cells == nil {
		return
	}
	for ri, row := range v.Rows {
		rowY := box.Y + pptx.EMU(ri)*(rowH+gap)
		if row.Label != "" && gutterW > 0 {
			lb := pptx.Box{X: box.X, Y: rowY, W: gutterW, H: rowH}
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
