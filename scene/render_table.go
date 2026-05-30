package scene

import "github.com/hurtener/pptx-go/pptx"

// Table composer (RFC §11.1 / §12 → native tbl). A header row is emitted when
// Headers is non-empty (with banding); each cell's RichText maps through the
// shared text mapping; a non-empty Caption renders as a separate text shape
// above the table (RFC §8.5).

func (r *renderer) renderTable(ps *pptx.Slide, box pptx.Box, v Table, slideID string) {
	cols := tableColumns(v)
	if cols == 0 {
		return
	}
	hasHeader := len(v.Headers) > 0
	rows := len(v.Rows)
	if hasHeader {
		rows++
	}
	if rows == 0 {
		return
	}

	tableBox := box
	if v.Caption != "" {
		capH := pptx.In(0.4)
		tf := ps.AddTextFrame(pptx.Box{X: box.X, Y: box.Y, W: box.W, H: capH})
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(v.Caption, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++
		tableBox = pptx.Box{X: box.X, Y: box.Y + capH, W: box.W, H: box.H - capH}
	}

	t := ps.AddTable(tableBox, rows, cols)

	rowIdx := 0
	if hasHeader {
		for c := 0; c < cols; c++ {
			if c < len(v.Headers) {
				r.fillCell(ps, t.Cell(0, c), v.Headers[c], true)
			}
		}
		t.SetHeaderRow(true).SetBanding(true, false)
		rowIdx = 1
	}
	for _, row := range v.Rows {
		for c := 0; c < cols; c++ {
			if c < len(row) {
				r.fillCell(ps, t.Cell(rowIdx, c), row[c], false)
			}
		}
		rowIdx++
	}
	r.stats.Shapes++ // the table (graphic frame) itself
	_ = slideID
}

// fillCell replaces a cell's text with the given RichText.
func (r *renderer) fillCell(ps *pptx.Slide, cell *pptx.Cell, rt RichText, header bool) {
	p := cell.TextFrame().Clear().AddParagraph(pptx.ParagraphOpts{})
	role := pptx.TypeBody
	r.addRichText(ps, p, rt, role)
	if header {
		// header cells are emphasized by the builder's header styling; nothing
		// extra to do here (engine, not product).
		_ = header
	}
}

// tableColumns returns the column count (max of header width and any row width).
func tableColumns(v Table) int {
	cols := len(v.Headers)
	for _, row := range v.Rows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	return cols
}
