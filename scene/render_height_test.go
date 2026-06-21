package scene

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Content-aware text height (Phase 22, R1). These white-box tests exercise the
// unexported wrappedLines / preferredHeight path: a node's slot grows with the
// lines its text wraps to, stacked nodes stop overlapping, overflow is reported
// truthfully, and single-line content reduces to the prior fixed heights.

// longProse returns a Prose whose single paragraph is long enough to wrap to
// several lines at body width.
func longProse(words int) Prose {
	return Prose{Paragraphs: []RichText{{{Text: strings.TrimSpace(strings.Repeat("lorem ipsum ", words))}}}}
}

// TestWrappedLines_Monotonic: longer text never yields fewer lines.
func TestWrappedLines_Monotonic(t *testing.T) {
	theme := pptx.DefaultTheme()
	avail := pptx.In(4)
	short := RichText{{Text: "hi"}}
	long := RichText{{Text: strings.Repeat("a longer line of body text ", 12)}}
	ls := wrappedLines(short, pptx.TypeBody, avail, theme)
	ll := wrappedLines(long, pptx.TypeBody, avail, theme)
	if ls != 1 {
		t.Errorf("short text: want 1 line, got %d", ls)
	}
	if ll <= ls {
		t.Errorf("wrappedLines not monotonic: short=%d, long=%d", ls, ll)
	}
}

// TestWrappedLines_Ceil: a paragraph whose natural width is just over the
// available width takes 2 lines (ceil, not floor).
func TestWrappedLines_Ceil(t *testing.T) {
	theme := pptx.DefaultTheme()
	rt := RichText{{Text: "some body text", Style: RunStyle{TypeRole: pptx.TypeBody}}}
	w := naturalWidth(rt, theme)
	// Available width one EMU under the natural width must force a second line.
	if got := wrappedLines(rt, pptx.TypeBody, w-1, theme); got != 2 {
		t.Errorf("ceil: avail just under natural width should give 2 lines, got %d", got)
	}
	// Exactly the natural width fits on one line.
	if got := wrappedLines(rt, pptx.TypeBody, w, theme); got != 1 {
		t.Errorf("avail == natural width should give 1 line, got %d", got)
	}
}

// TestWrappedLines_Fallbacks: empty text, non-positive width, and nil theme all
// fall back to 1 (the byte-identical fixed-height path).
func TestWrappedLines_Fallbacks(t *testing.T) {
	theme := pptx.DefaultTheme()
	rt := RichText{{Text: "anything"}}
	cases := []struct {
		name  string
		rt    RichText
		avail pptx.EMU
		theme *pptx.Theme
	}{
		{"empty text", RichText{{Text: ""}}, pptx.In(4), theme},
		{"nil richtext", nil, pptx.In(4), theme},
		{"zero width", rt, 0, theme},
		{"negative width", rt, -100, theme},
		{"nil theme", rt, pptx.In(4), nil},
	}
	for _, c := range cases {
		if got := wrappedLines(c.rt, pptx.TypeBody, c.avail, c.theme); got != 1 {
			t.Errorf("%s: want 1, got %d", c.name, got)
		}
	}
}

// TestPreferredHeight_SingleLineReducesToFixed is acceptance criterion 4 at the
// unit level: single-line content yields exactly the pre-Phase-22 fixed slot
// heights, and the avail<=0 / nil-theme fallback yields the same — the basis of
// the byte-identity guarantee.
func TestPreferredHeight_SingleLineReducesToFixed(t *testing.T) {
	theme := pptx.DefaultTheme()
	wide := pptx.In(9) // generous: every fixture below is one line
	cases := []struct {
		name string
		node SlideNode
		want pptx.EMU
	}{
		{"heading", Heading{Text: RichText{{Text: "Short"}}, Level: 2}, pptx.In(0.6)},
		{"prose 1 para", Prose{Paragraphs: []RichText{{{Text: "One line."}}}}, pptx.In(0.4)},
		{"prose 3 paras", Prose{Paragraphs: []RichText{{{Text: "a"}}, {{Text: "b"}}, {{Text: "c"}}}}, pptx.In(0.4) * 3},
		{"list 2 items", List{Items: []ListItem{{Text: RichText{{Text: "x"}}}, {Text: RichText{{Text: "y"}}}}}, pptx.In(0.32) * 2},
		{"quote", Quote{Text: RichText{{Text: "Be."}}}, pptx.In(1.1)},
		{"callout", Callout{Body: RichText{{Text: "Note."}}}, pptx.In(1.0)},
	}
	for _, c := range cases {
		if got := preferredHeight(c.node, wide, theme); got != c.want {
			t.Errorf("%s: content-aware single-line height = %d, want %d", c.name, got, c.want)
		}
		// Fallback path (no width / no theme) must match the same fixed height.
		if got := preferredHeight(c.node, 0, nil); got != c.want {
			t.Errorf("%s: fallback height = %d, want %d", c.name, got, c.want)
		}
	}
}

// TestPreferredHeight_ProseGrowsWithWrap is acceptance criterion 1: a paragraph
// that wraps to N lines is allotted at least N line-heights.
func TestPreferredHeight_ProseGrowsWithWrap(t *testing.T) {
	theme := pptx.DefaultTheme()
	r := newTestRenderer(t)
	avail := r.bodyRegion().W
	p := longProse(40)
	n := wrappedLines(p.Paragraphs[0], pptx.TypeBody, avail, theme)
	if n < 2 {
		t.Fatalf("fixture did not wrap (got %d lines); widen the paragraph", n)
	}
	got := preferredHeight(p, avail, theme)
	want := pptx.In(0.4) * pptx.EMU(n)
	if got != want {
		t.Errorf("wrapped prose height = %d, want %d (= %d lines × 0.4in)", got, want, n)
	}
	// And strictly taller than the single-line allotment.
	if got <= pptx.In(0.4) {
		t.Errorf("wrapped prose (%d) should exceed one line-height (%d)", got, pptx.In(0.4))
	}
}

// TestLayout_NoOverlapMultiLine is acceptance criterion 2: the node stacked
// below a multi-line paragraph starts at or below that paragraph's bottom.
func TestLayout_NoOverlapMultiLine(t *testing.T) {
	r := newTestRenderer(t)
	nodes := []SlideNode{
		longProse(40),
		Heading{Text: RichText{{Text: "After the prose"}}, Level: 2},
	}
	pls := r.layout(nodes, "s", Alignment{})
	if len(pls) != 2 {
		t.Fatalf("expected 2 placements, got %d", len(pls))
	}
	prose, heading := pls[0], pls[1]
	if _, ok := heading.node.(Heading); !ok {
		t.Fatalf("placement order unexpected: %T then %T", prose.node, heading.node)
	}
	if heading.box.Y < prose.box.Bottom() {
		t.Errorf("overlap: heading Y (%d) is above prose bottom (%d)", heading.box.Y, prose.box.Bottom())
	}
}

// TestOverflow_FiresOnWrappedContent is acceptance criterion 3: a slide whose
// real wrapped content exceeds the body region emits the overflow warning
// (which the fixed-height model did not).
func TestOverflow_FiresOnWrappedContent(t *testing.T) {
	// One very long paragraph: ~many wrapped lines, taller than the body region.
	sc := Scene{Slides: []SceneSlide{{
		ID:    "overflowing",
		Nodes: []SlideNode{longProse(400)},
	}}}
	pres := pptx.New()
	stats, err := Render(pres, sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !hasOverflowWarning(stats, "overflowing") {
		t.Errorf("expected overflow warning for a slide of wrapped content exceeding the body; warnings=%v", stats.Warnings)
	}

	// A single short line must NOT warn (no false positive).
	scOK := Scene{Slides: []SceneSlide{{
		ID:    "fine",
		Nodes: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "Short."}}}}},
	}}}
	statsOK, err := Render(pptx.New(), scOK)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if hasOverflowWarning(statsOK, "fine") {
		t.Errorf("unexpected overflow warning for a single short line: %v", statsOK.Warnings)
	}
}

// TestPreferredHeight_NonProseGrow asserts the content-aware growth arithmetic
// for the other text families (Quote, Callout, List, Table) at lines>1 — Prose is
// covered by TestPreferredHeight_ProseGrowsWithWrap, but the per-line constants,
// the callout inset width adjustment, and the table per-row model are otherwise
// only exercised at single-line (where the lines-1 term is zero).
func TestPreferredHeight_NonProseGrow(t *testing.T) {
	theme := pptx.DefaultTheme()
	avail := pptx.In(9)
	long := RichText{{Text: strings.TrimSpace(strings.Repeat("lorem ipsum dolor ", 60))}}

	// Quote: In(1.1) + quoteLineEst*(lines-1), wrapped at TypeH3.
	qLines := wrappedLines(long, pptx.TypeH3, avail, theme)
	if qLines < 2 {
		t.Fatalf("quote fixture did not wrap (%d lines)", qLines)
	}
	gotQ := preferredHeight(Quote{Text: long}, avail, theme)
	if want := pptx.In(1.1) + quoteLineEst*pptx.EMU(qLines-1); gotQ != want {
		t.Errorf("quote height = %d, want %d", gotQ, want)
	}
	if gotQ <= pptx.In(1.1) {
		t.Errorf("quote should grow beyond the single-line baseline %d, got %d", pptx.In(1.1), gotQ)
	}

	// Callout: In(1.0) + calloutLineEst*(lines-1), wrapped at the inset width.
	cLines := wrappedLines(long, pptx.TypeBody, avail-calloutInsetEst, theme)
	gotC := preferredHeight(Callout{Body: long}, avail, theme)
	if want := pptx.In(1.0) + calloutLineEst*pptx.EMU(cLines-1); gotC != want {
		t.Errorf("callout height = %d, want %d (inset-adjusted lines=%d)", gotC, want, cLines)
	}
	if gotC <= pptx.In(1.0) {
		t.Errorf("callout should grow beyond the single-line baseline %d, got %d", pptx.In(1.0), gotC)
	}

	// List: a single long item grows to its wrapped line count × In(0.32).
	lLines := wrappedLines(long, pptx.TypeBody, avail, theme)
	gotL := preferredHeight(List{Items: []ListItem{{Text: long}}}, avail, theme)
	if want := pptx.In(0.32) * pptx.EMU(lLines); gotL != want {
		t.Errorf("list height = %d, want %d", gotL, want)
	}
	if gotL <= pptx.In(0.32) {
		t.Errorf("list item should grow with wrap, got %d", gotL)
	}

	// Table: a long cell makes its body row taller than one line-height; height is
	// the header row (1 line) plus the body row (the tallest cell's line count).
	tbl := Table{Headers: []RichText{{{Text: "H"}}}, Rows: [][]RichText{{long}}}
	tLines := wrappedLines(long, pptx.TypeBody, avail /*colW = avail/1*/, theme)
	if tLines < 2 {
		t.Fatalf("table cell did not wrap (%d lines)", tLines)
	}
	gotT := preferredHeight(tbl, avail, theme)
	if want := pptx.In(0.4) + pptx.In(0.4)*pptx.EMU(tLines); gotT != want {
		t.Errorf("table height = %d, want %d (header + %d-line body row)", gotT, want, tLines)
	}
}

func hasOverflowWarning(s Stats, slideID string) bool {
	for _, w := range s.Warnings {
		if w.SlideID == slideID && strings.Contains(w.Message, "overflows its region") {
			return true
		}
	}
	return false
}
