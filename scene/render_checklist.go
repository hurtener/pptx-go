package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// Checklist composer (RFC §11.1 / §12, R12.2, D-095). A Checklist renders a dense
// feature list: rows of a filled status glyph (a curated check / x / dot custGeom,
// NOT an empty font checkbox) before rich text, reflowed row-major into 1–3 columns,
// with the text hanging-indented from the glyph width. Fill distributes inter-row
// slack so a short list spans its box. Media-free, deterministic integer-EMU layout.

// Pinned layout metrics (not theme tokens — they size geometry, not a visual property,
// like buttonMetrics). The glyph colors are tokens (P2); see checklistGlyphColor.
const (
	checklistGlyphSz  = pptx.EMU(146304) // In(0.16); the status glyph box
	checklistGlyphGap = pptx.EMU(91440)  // In(0.10); glyph-to-text hanging gap
	checklistColGap   = pptx.EMU(274320) // In(0.30); gap between columns
	checklistRowGap   = pptx.EMU(91440)  // In(0.10); default inter-row gap
	checklistLineH    = pptx.EMU(292608) // In(0.32); per text line (matches renderList)
)

// checklistCols clamps the requested column count to 1..3 (0 → 1).
func checklistCols(v Checklist) int {
	cols := v.Columns
	if cols < 1 {
		cols = 1
	}
	if cols > 3 {
		cols = 3
	}
	return cols
}

// checklistColW returns the per-column width and the text column width (column minus
// the glyph and its hanging gap) for a box of width boxW split into cols columns.
func checklistColW(boxW pptx.EMU, cols int) (colW, textColW pptx.EMU) {
	colW = (boxW - checklistColGap*pptx.EMU(cols-1)) / pptx.EMU(cols)
	textColW = colW - checklistGlyphSz - checklistGlyphGap
	if textColW < 0 {
		textColW = 0
	}
	return colW, textColW
}

// checklistRowHeights returns the per-row heights: each row is the tallest cell in it
// (the cell's wrapped-line count × the per-line height), measured at the text column
// width. Rows are filled row-major: item i is at (row=i/cols, col=i%cols). A row with
// no measurable text is one line tall.
func checklistRowHeights(v Checklist, textColW pptx.EMU, theme *pptx.Theme) []pptx.EMU {
	cols := checklistCols(v)
	n := len(v.Items)
	rows := (n + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	heights := make([]pptx.EMU, rows)
	for i, it := range v.Items {
		row := i / cols
		lines := wrappedLines(it.Text, pptx.TypeBody, textColW, theme)
		if h := checklistLineH * pptx.EMU(lines); h > heights[row] {
			heights[row] = h
		}
	}
	for r := range heights {
		if heights[r] == 0 {
			heights[r] = checklistLineH
		}
	}
	return heights
}

// checklistPreferredHeight is the node's slot height: the summed per-row heights plus
// the default inter-row gaps. Content-aware (rows grow with wrapped text), deterministic.
func checklistPreferredHeight(v Checklist, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	cols := checklistCols(v)
	_, textColW := checklistColW(avail, cols)
	heights := checklistRowHeights(v, textColW, theme)
	var total pptx.EMU
	for _, h := range heights {
		total += h
	}
	if len(heights) > 1 {
		total += checklistRowGap * pptx.EMU(len(heights)-1)
	}
	return total
}

// checklistGlyphName maps an item to its glyph icon name: the per-item override if set,
// else the state default (CheckDone → check, CheckNo → x, CheckNeutral → dot).
func checklistGlyphName(it ChecklistItem) string {
	if it.Icon != "" {
		return it.Icon
	}
	switch it.State {
	case CheckNo:
		return "x"
	case CheckNeutral:
		return "dot"
	default: // CheckDone
		return "check"
	}
}

// checklistGlyphColor resolves the glyph fill color: the GlyphTone override (all items)
// if set, else the per-state default — CheckDone is accent-tinted, the rest muted.
// Token-bound (P2), so a theme swap re-skins it.
func checklistGlyphColor(v Checklist, st CheckState) pptx.Color {
	if v.GlyphTone != nil {
		return pptx.TokenColor(*v.GlyphTone)
	}
	if st == CheckDone {
		return pptx.TokenColor(pptx.ColorAccent)
	}
	return pptx.TokenTextColor(pptx.TextMuted)
}

func (r *renderer) renderChecklist(ps *pptx.Slide, box pptx.Box, v Checklist) {
	cols := checklistCols(v)
	colW, textColW := checklistColW(box.W, cols)
	heights := checklistRowHeights(v, textColW, r.theme)
	rows := len(heights)

	var total pptx.EMU
	for _, h := range heights {
		total += h
	}

	// Inter-row gap: the pinned default, or — when Fill is set and the box is taller
	// than the content — the slack spread evenly across the gaps so the last row's
	// bottom meets the box bottom (the VAlignJustify primitive, per-row). Integer math.
	rowGap := checklistRowGap
	if v.Fill && rows > 1 {
		if slack := box.H - total; slack > checklistRowGap*pptx.EMU(rows-1) {
			rowGap = slack / pptx.EMU(rows-1)
		}
	}

	// Precompute each row's top Y.
	rowY := make([]pptx.EMU, rows)
	y := box.Y
	for ri := 0; ri < rows; ri++ {
		rowY[ri] = y
		y += heights[ri] + rowGap
	}

	for i, it := range v.Items {
		row := i / cols
		col := i % cols
		cellX := box.X + pptx.EMU(col)*(colW+checklistColGap)
		cellY := rowY[row]

		// Glyph: a filled custGeom, vertically centered on the first text line.
		glyphBox := pptx.Box{X: cellX, Y: cellY + (checklistLineH-checklistGlyphSz)/2, W: checklistGlyphSz, H: checklistGlyphSz}
		r.addChecklistGlyph(ps, glyphBox, checklistGlyphName(it), checklistGlyphColor(v, it.State))

		// Text, hanging-indented past the glyph.
		textBox := pptx.Box{X: cellX + checklistGlyphSz + checklistGlyphGap, Y: cellY, W: textColW, H: heights[row]}
		tf := ps.AddTextFrame(textBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{LineHeight: r.lineH(pptx.TypeBody)})
		r.addRichText(ps, p, it.Text, pptx.TypeBody)
		r.stats.Shapes++
	}
}

// addChecklistGlyph renders a curated check/x/dot (or override) icon as a native
// custGeom glyph filled with color. The name resolved at Stage-1 for overrides; the
// state defaults are curated, so a miss here is a wiring bug surfaced as a warning.
func (r *renderer) addChecklistGlyph(ps *pptx.Slide, box pptx.Box, name string, color pptx.Color) {
	svg, ok := r.cfg.icons.Lookup(name)
	if !ok {
		r.warn("", fmt.Sprintf("checklist glyph %q not found at compose (should have failed Stage-1)", name))
		return
	}
	if _, err := ps.AddIcon(svg, box, pptx.WithFill(pptx.SolidFill(color))); err != nil {
		r.warn("", fmt.Sprintf("checklist glyph %q failed to render: %v", name, err))
		return
	}
	r.stats.Shapes++
}
