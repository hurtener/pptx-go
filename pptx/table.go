package pptx

import "github.com/hurtener/pptx-go/internal/ooxml/slide"

// ============================================================================
// Table (RFC §8.5)
// ============================================================================
//
// A Table renders as a native OOXML <a:tbl> inside a graphic frame. Header rows,
// banding, merged cells, per-cell fills/borders, and rich-text cells are all
// first-class. Banding/header are emitted as concrete alternating cell fills
// (deterministic and visible without a table-style part); the tblPr flags carry
// the intent. A table caption is a separate TextFrame the caller positions above
// the table (the scene renderer composes it).

// defaultTableStyleID is a built-in "No Style, Table Grid" table style — a GUID
// PowerPoint resolves without a styles part, so cells get gridlines.
const defaultTableStyleID = "{5940675A-B579-460E-94D1-54222C63F5DA}"

// Table is a handle to a table added to a slide.
type Table struct {
	slide      *Slide
	gf         *slide.XGraphicFrame
	rows, cols int
	headerOn   bool
	bandRowOn  bool
}

// AddTable adds a rows×cols table positioned by box (EMU) with equal column
// widths, and returns a handle (RFC §8.5).
func (s *Slide) AddTable(box Box, rows, cols int) *Table {
	if rows < 1 {
		rows = 1
	}
	if cols < 1 {
		cols = 1
	}
	gf := s.builder.AddTable(int(box.X), int(box.Y), int(box.W), int(box.H), rows, cols)
	t := &Table{slide: s, gf: gf, rows: rows, cols: cols}
	t.tbl().Pr = &slide.XTablePr{TableStyleID: defaultTableStyleID}
	return t
}

func (t *Table) tbl() *slide.XTable { return t.gf.Graphic.GraphicData.Table }

// SetColumnWidths overrides column widths (EMU), left to right.
func (t *Table) SetColumnWidths(widths ...EMU) *Table {
	grid := t.tbl().Grid
	for i := 0; i < len(grid.GridCols) && i < len(widths); i++ {
		grid.GridCols[i].W = int(widths[i])
	}
	return t
}

// Cell returns the cell at (row, col).
func (t *Table) Cell(row, col int) *Cell {
	if row < 0 || row >= t.rows || col < 0 || col >= t.cols {
		return &Cell{}
	}
	return &Cell{slide: t.slide, table: t, row: row, col: col, tc: &t.tbl().Rows[row].Cells[col]}
}

// SetHeaderRow marks (or unmarks) the first row as a header and restyles.
func (t *Table) SetHeaderRow(on bool) *Table {
	t.headerOn = on
	pr := t.ensurePr()
	pr.FirstRow = boolAttr(on)
	t.applyStyling()
	return t
}

// SetBanding toggles row/column banding and restyles. V1 emits alternating row
// fills for rowBand; colBand sets the intent flag only.
func (t *Table) SetBanding(rowBand, colBand bool) *Table {
	t.bandRowOn = rowBand
	pr := t.ensurePr()
	pr.BandRow = boolAttr(rowBand)
	pr.BandCol = boolAttr(colBand)
	t.applyStyling()
	return t
}

func (t *Table) ensurePr() *slide.XTablePr {
	if t.tbl().Pr == nil {
		t.tbl().Pr = &slide.XTablePr{TableStyleID: defaultTableStyleID}
	}
	return t.tbl().Pr
}

// applyStyling sets concrete header/banding cell fills. The header row and every
// odd body row (0-based, below the header) get a SurfaceAlt fill.
func (t *Table) applyStyling() {
	offset := 0
	if t.headerOn {
		offset = 1
	}
	band := SolidFill(TokenColor(ColorSurfaceAlt))
	for r := 0; r < t.rows; r++ {
		fill := false
		switch {
		case t.headerOn && r == 0:
			fill = true
		case t.bandRowOn && (r-offset)%2 == 1:
			fill = true
		}
		if !fill {
			continue
		}
		for c := 0; c < t.cols; c++ {
			t.Cell(r, c).SetFill(band)
		}
	}
}

// Cell is a handle to a table cell.
type Cell struct {
	slide    *Slide
	table    *Table
	row, col int
	tc       *slide.XTableCell
}

func (c *Cell) ok() bool { return c != nil && c.tc != nil }

// TextFrame returns a rich-text frame over the cell's text body (the Phase 04
// model — paragraphs, runs, hyperlinks).
func (c *Cell) TextFrame() *TextFrame {
	if !c.ok() {
		return &TextFrame{}
	}
	if c.tc.TextBody == nil {
		c.tc.TextBody = &slide.XTextBody{BodyPr: &slide.XBodyPr{}, LstStyle: &slide.XTextParagraphList{}}
	}
	return &TextFrame{s: c.slide, body: c.tc.TextBody}
}

// SetText replaces the cell text with a single themed body run.
func (c *Cell) SetText(text string) *Cell {
	if !c.ok() {
		return c
	}
	c.tc.TextBody = &slide.XTextBody{BodyPr: &slide.XBodyPr{}, LstStyle: &slide.XTextParagraphList{}}
	c.TextFrame().AddParagraph(ParagraphOpts{}).AddRun(text, RunStyle{TypeRole: TypeBody})
	return c
}

// SetFill sets the cell's interior fill (resolved against the active theme).
func (c *Cell) SetFill(f Fill) *Cell {
	if !c.ok() || f == nil {
		return c
	}
	tmp := &slide.XShapeProperties{}
	f.applyFill(tmp, c.activeTheme())
	pr := c.ensurePr()
	pr.SolidFill = tmp.SolidFill
	pr.NoFill = tmp.NoFill
	return c
}

// SetBorders sets all four cell borders to line (resolved against the theme).
func (c *Cell) SetBorders(line Line) *Cell {
	if !c.ok() || line.isZero() {
		return c
	}
	tmp := &slide.XShapeProperties{}
	line.apply(tmp, c.activeTheme())
	pr := c.ensurePr()
	pr.LnL, pr.LnR, pr.LnT, pr.LnB = tmp.Line, tmp.Line, tmp.Line, tmp.Line
	return c
}

// MergeRight spans the cell across n columns (n ≥ 2), marking the covered cells.
func (c *Cell) MergeRight(n int) *Cell {
	if !c.ok() || c.table == nil || n < 2 {
		return c
	}
	c.tc.GridSpan = n
	tbl := c.table.tbl()
	for j := 1; j < n && c.col+j < c.table.cols; j++ {
		tbl.Rows[c.row].Cells[c.col+j].HMerge = "1"
	}
	return c
}

// MergeDown spans the cell across n rows (n ≥ 2), marking the covered cells.
func (c *Cell) MergeDown(n int) *Cell {
	if !c.ok() || c.table == nil || n < 2 {
		return c
	}
	c.tc.RowSpan = n
	tbl := c.table.tbl()
	for i := 1; i < n && c.row+i < c.table.rows; i++ {
		tbl.Rows[c.row+i].Cells[c.col].VMerge = "1"
	}
	return c
}

func (c *Cell) ensurePr() *slide.XTableCellProps {
	if c.tc.Pr == nil {
		c.tc.Pr = &slide.XTableCellProps{}
	}
	return c.tc.Pr
}

func (c *Cell) activeTheme() *Theme {
	if c.slide != nil {
		return c.slide.activeTheme()
	}
	return DefaultTheme()
}

// ============================================================================
// Read accessors (RFC §16) — the read inverse of the table authoring API.
// ============================================================================

// RowCount returns the table's number of rows — the read inverse of AddTable's
// rows argument.
func (t *Table) RowCount() int { return t.rows }

// ColCount returns the table's number of columns — the read inverse of
// AddTable's cols argument.
func (t *Table) ColCount() int { return t.cols }

// ColumnWidths returns the table's column widths (EMU), left to right — the read
// inverse of SetColumnWidths.
func (t *Table) ColumnWidths() []EMU {
	grid := t.tbl().Grid
	if grid == nil {
		return nil
	}
	ws := make([]EMU, len(grid.GridCols))
	for i, c := range grid.GridCols {
		ws[i] = EMU(c.W)
	}
	return ws
}

// HeaderRow reports whether the first row is marked as a header — the read
// inverse of SetHeaderRow.
func (t *Table) HeaderRow() bool { return t.headerOn }

// RowBanding reports whether row banding is enabled — the read inverse of
// SetBanding's rowBand argument.
func (t *Table) RowBanding() bool { return t.bandRowOn }

// GridSpan returns the number of columns the cell spans (1 for an unmerged
// cell) — the read inverse of MergeRight.
func (c *Cell) GridSpan() int {
	if !c.ok() || c.tc.GridSpan == 0 {
		return 1
	}
	return c.tc.GridSpan
}

// RowSpan returns the number of rows the cell spans (1 for an unmerged cell) —
// the read inverse of MergeDown.
func (c *Cell) RowSpan() int {
	if !c.ok() || c.tc.RowSpan == 0 {
		return 1
	}
	return c.tc.RowSpan
}

// Covered reports whether the cell is covered by a neighboring cell's merge
// (hMerge or vMerge) — i.e. it is not the merge anchor and renders no content.
func (c *Cell) Covered() bool {
	return c.ok() && (c.tc.HMerge == "1" || c.tc.VMerge == "1")
}

// Fill returns the cell's interior fill, or nil when the cell has no explicit
// fill — the read inverse of SetFill.
func (c *Cell) Fill() Fill {
	if !c.ok() || c.tc.Pr == nil {
		return nil
	}
	switch {
	case c.tc.Pr.SolidFill != nil:
		return solidFill{color: colorFromSrgb(c.tc.Pr.SolidFill.SrgbClr)}
	case c.tc.Pr.NoFill != nil:
		return noFill{}
	default:
		return nil
	}
}

// boolAttr renders an OOXML boolean attribute ("1" when true, omitted otherwise).
func boolAttr(b bool) string {
	if b {
		return "1"
	}
	return ""
}
