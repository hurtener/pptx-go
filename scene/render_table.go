package scene

import "github.com/hurtener/pptx-go/pptx"

// Table composer (RFC §11.1 / §12 → native tbl). A header row is emitted when
// Headers is non-empty (with banding); each cell's RichText maps through the
// shared text mapping; a non-empty Caption renders as a separate text shape
// above the table (RFC §8.5).
//
// A non-nil Table.Style switches to the comparison-matrix styled path (R14.3,
// D-118): a header band, zebra striping, a highlighted column, a row-label
// column, and grouped header spans — all from theme tokens, controlling every
// cell fill explicitly. A nil Style keeps the plain banded table (byte-identical).

func (r *renderer) renderTable(ps *pptx.Slide, box pptx.Box, v Table, slideID string) {
	cols := tableColumns(v)
	if cols == 0 {
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

	if v.Style != nil {
		r.renderStyledTable(ps, tableBox, v, cols)
		_ = slideID
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

// renderStyledTable renders a Table with a non-nil Style (R14.3, D-118). It sets
// every cell fill explicitly (no builder header/row banding, which would overwrite
// them via applyStyling), so the output is deterministic and token-driven.
func (r *renderer) renderStyledTable(ps *pptx.Slide, box pptx.Box, v Table, cols int) {
	st := v.Style
	hasHeader := len(v.Headers) > 0
	hasGroups := len(st.HeaderGroups) > 0

	rows := len(v.Rows)
	if hasHeader {
		rows++
	}
	if hasGroups {
		rows++
	}
	if rows == 0 {
		return
	}

	t := ps.AddTable(box, rows, cols)
	rowIdx := 0

	// Grouped header row: each group's label spans Span columns, accent band.
	if hasGroups {
		col := 0
		for _, g := range st.HeaderGroups {
			if col >= cols {
				break
			}
			span := g.Span
			if span < 1 {
				span = 1
			}
			if col+span > cols {
				span = cols - col
			}
			cell := t.Cell(0, col)
			cell.SetFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent)))
			r.fillCellColored(cell, RichText{{Text: g.Label}}, r.cellTextOn(pptx.ColorAccent), true)
			if span >= 2 {
				cell.MergeRight(span)
			}
			col += span
		}
		rowIdx++
	}

	// Header row.
	if hasHeader {
		hr := rowIdx
		for c := 0; c < cols; c++ {
			if c >= len(v.Headers) {
				continue
			}
			cell := t.Cell(hr, c)
			if st.HeaderFill {
				cell.SetFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent)))
				r.fillCellColored(cell, v.Headers[c], r.cellTextOn(pptx.ColorAccent), true)
			} else {
				// Bold header text on the default surface (no band).
				r.fillCellColored(cell, v.Headers[c], nil, true)
			}
		}
		rowIdx++
	}

	// Body rows.
	for bi, row := range v.Rows {
		for c := 0; c < cols; c++ {
			if c >= len(row) {
				continue
			}
			cell := t.Cell(rowIdx, c)
			// Row-label column (col 0) is bold; other cells are plain body text.
			if st.RowLabelCol && c == 0 {
				r.fillCellColored(cell, row[c], nil, true)
			} else {
				r.fillCell(ps, cell, row[c], false)
			}
			r.styleBodyCell(cell, bi, c, st)
		}
		rowIdx++
	}
	r.stats.Shapes++ // the table (graphic frame) itself
}

// styleBodyCell applies the zebra / row-label / highlight-column fills to a body
// cell. Highlight wins over row-label, which wins over zebra (applied last wins,
// since SetFill overwrites the cell fill). bodyRow is the 0-based body row index.
func (r *renderer) styleBodyCell(cell *pptx.Cell, bodyRow, col int, st *TableStyle) {
	switch {
	case st.HighlightCol >= 1 && col == st.HighlightCol-1:
		cell.SetFill(pptx.SolidFill(pptx.TokenColorAlpha(pptx.ColorAccent, tableHighlightAlpha)))
		cell.SetBorders(pptx.Line{Width: pptx.Pt(1.5), Color: pptx.TokenColor(pptx.ColorAccent)})
	case st.RowLabelCol && col == 0:
		cell.SetFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt)))
	case st.Zebra && bodyRow%2 == 1:
		cell.SetFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt)))
	}
}

// tableHighlightAlpha is the OOXML opacity of the highlighted column's accent
// tint — a subtle wash (not a full accent fill) so the column reads as emphasized
// without overpowering the matrix. Pinned (not a token); the color is a token.
const tableHighlightAlpha = 16000

// cellTextOn returns the auto-contrast text color for a cell filled with role
// (D-082): the inverse text token on a dark fill, else the primary text token.
func (r *renderer) cellTextOn(role pptx.ColorRole) pptx.Color {
	return r.cellTextOnColor(r.theme.ResolveColor(role))
}

// cellTextOnColor is the resolved-RGB core of cellTextOn (R8.4): the inverse text
// token on a dark fill, else the primary text token. Keyed on a literal RGB so a
// brand-accent fill (multi-accent palette, no ColorRole) gets auto-contrast text
// too. cellTextOn is the role-keyed wrapper; byte-identical for a role argument.
func (r *renderer) cellTextOnColor(bg pptx.RGB) pptx.Color {
	if c := r.onSurfaceRGB(bg); c != nil {
		return c
	}
	return pptx.TokenTextColor(pptx.TextPrimary)
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

// fillCellColored replaces a cell's text, forcing a text color (nil = each run's
// own color) and optionally bolding every run — the styled-table path (D-118).
func (r *renderer) fillCellColored(cell *pptx.Cell, rt RichText, color pptx.Color, bold bool) {
	p := cell.TextFrame().Clear().AddParagraph(pptx.ParagraphOpts{})
	for _, run := range rt {
		style := pptx.RunStyle{
			TypeRole: pptx.TypeBody,
			Bold:     bold || run.Style.Bold,
			Italic:   run.Style.Italic,
			Code:     run.Style.Code,
		}
		if color != nil {
			style.Color = color
		} else {
			style.Color = colorFor(run.Color)
		}
		p.AddRun(run.Text, style)
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
