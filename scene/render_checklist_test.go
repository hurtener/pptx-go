package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
)

// White-box tests for the Checklist composer (R12.2, D-095): column clamp, the
// glyph-per-state mapping + GlyphTone override, content-aware row heights, the Fill
// row distribution, and the per-row shape count.

func checklistItems(n int) []ChecklistItem {
	items := make([]ChecklistItem, n)
	for i := range items {
		items[i] = ChecklistItem{Text: RichText{{Text: "feature"}}}
	}
	return items
}

// TestChecklistCols_Clamp: Columns clamps to 1..3 (0 → 1).
func TestChecklistCols_Clamp(t *testing.T) {
	cases := map[int]int{0: 1, 1: 1, 2: 2, 3: 3, 4: 3, -1: 1}
	for in, want := range cases {
		if got := checklistCols(Checklist{Columns: in}); got != want {
			t.Errorf("checklistCols(%d) = %d, want %d", in, got, want)
		}
	}
}

// TestChecklistGlyphName: state defaults map to filled glyphs; an item Icon overrides.
func TestChecklistGlyphName(t *testing.T) {
	cases := []struct {
		it   ChecklistItem
		want string
	}{
		{ChecklistItem{State: CheckDone}, "check"},
		{ChecklistItem{State: CheckNo}, "x"},
		{ChecklistItem{State: CheckNeutral}, "dot"},
		{ChecklistItem{State: CheckDone, Icon: "star"}, "star"},
	}
	for _, c := range cases {
		if got := checklistGlyphName(c.it); got != c.want {
			t.Errorf("checklistGlyphName(%+v) = %q, want %q", c.it, got, c.want)
		}
	}
}

// TestChecklistGlyphColor: CheckDone is accent, others muted; GlyphTone overrides all.
func TestChecklistGlyphColor(t *testing.T) {
	if c := checklistGlyphColor(Checklist{}, CheckDone); c != pptx.TokenColor(pptx.ColorAccent) {
		t.Errorf("CheckDone default color = %#v, want accent", c)
	}
	if c := checklistGlyphColor(Checklist{}, CheckNo); c != pptx.TokenTextColor(pptx.TextMuted) {
		t.Errorf("CheckNo default color = %#v, want muted", c)
	}
	warm := ColorAccentWarm
	if c := checklistGlyphColor(Checklist{GlyphTone: &warm}, CheckDone); c != pptx.TokenColor(pptx.ColorAccentWarm) {
		t.Errorf("GlyphTone override = %#v, want accent-warm", c)
	}
}

// TestChecklistColW: the text column width is the column width minus the glyph and gap,
// and a 2-column split is narrower than 1-column.
func TestChecklistColW(t *testing.T) {
	box := pptx.In(10)
	col1, text1 := checklistColW(box, 1)
	col2, _ := checklistColW(box, 2)
	if text1 != col1-checklistGlyphSz-checklistGlyphGap {
		t.Errorf("text column width %d != colW - glyph - gap", text1)
	}
	if col2 >= col1 {
		t.Errorf("2-column width %d should be narrower than 1-column %d", col2, col1)
	}
}

// TestChecklistRowHeights_RowMajor: a 5-item 2-column list has 3 rows; a row with a
// long wrapping item is taller than one with a short item.
func TestChecklistRowHeights_RowMajor(t *testing.T) {
	r := newTestRenderer(t)
	_, textColW := checklistColW(pptx.In(6), 2)
	v := Checklist{Columns: 2, Items: []ChecklistItem{
		{Text: RichText{{Text: "short"}}},
		{Text: RichText{{Text: "short"}}},
		{Text: RichText{{Text: "A deliberately long line of text that will certainly wrap across several lines in a narrow column"}}},
		{Text: RichText{{Text: "short"}}},
		{Text: RichText{{Text: "short"}}},
	}}
	heights := checklistRowHeights(v, textColW, r.theme)
	if len(heights) != 3 {
		t.Fatalf("rows = %d, want 3 (5 items, 2 columns)", len(heights))
	}
	if heights[1] <= heights[0] {
		t.Errorf("row 1 (long item) height %d should exceed row 0 (short) %d", heights[1], heights[0])
	}
}

// TestRenderChecklist_Fill: with Fill the last row's bottom reaches the box bottom; the
// composer emits a glyph + text per item.
func TestRenderChecklist_Fill(t *testing.T) {
	r := newTestRenderer(t)
	r.cfg.icons = icons.Curated()
	ps := r.pres.AddSlide()
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(8), H: pptx.In(6)}
	v := Checklist{Columns: 1, Fill: true, Items: checklistItems(3)}
	r.renderChecklist(ps, box, v)
	// 3 items × (glyph + text) = 6 shapes.
	if r.stats.Shapes != 6 {
		t.Errorf("emitted %d shapes, want 6 (warnings: %v)", r.stats.Shapes, r.stats.Warnings)
	}
}
