package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
)

// White-box tests for the IconRows composer (R12.7, D-100): the glyph-color default, the
// label/meta width split, content-aware row heights, and the per-row shape count.

// TestIconRowsGlyphColor: a zero GlyphColor (ColorCanvas) resolves to accent; explicit
// is honored.
func TestIconRowsGlyphColor(t *testing.T) {
	if got := iconRowsGlyphColorRole(IconRows{}); got != pptx.ColorAccent {
		t.Errorf("zero GlyphColor = %v, want ColorAccent", got)
	}
	if got := iconRowsGlyphColorRole(IconRows{GlyphColor: ColorSuccess}); got != pptx.ColorSuccess {
		t.Errorf("explicit GlyphColor = %v, want ColorSuccess", got)
	}
}

// TestIconRowsLabelW: a row with an icon and a meta has a narrower label column than a
// bare row.
func TestIconRowsLabelW(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.Box{W: pptx.In(8)}
	bare := iconRowsLabelW(box, IconRow{Label: RichText{{Text: "x"}}}, r.theme)
	full := iconRowsLabelW(box, IconRow{Icon: "star", Label: RichText{{Text: "x"}}, Meta: RichText{{Text: "core"}}}, r.theme)
	if full >= bare {
		t.Errorf("icon+meta row label width %d should be narrower than a bare row %d", full, bare)
	}
}

// TestIconRowsRowHeights: a row whose label wraps is taller than a short one; every row is
// at least the icon size.
func TestIconRowsRowHeights(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.Box{W: pptx.In(3)}
	v := IconRows{Rows: []IconRow{
		{Icon: "star", Label: RichText{{Text: "short"}}},
		{Icon: "star", Label: RichText{{Text: "A deliberately long row label that wraps across several lines in a narrow card column"}}},
	}}
	heights := iconRowsRowHeights(v, box, r.theme)
	if heights[1] <= heights[0] {
		t.Errorf("wrapping row %d should be taller than short row %d", heights[1], heights[0])
	}
	if heights[0] < iconRowsIconSz {
		t.Errorf("row height %d below the icon size %d", heights[0], iconRowsIconSz)
	}
}

// TestRenderIconRows_ShapeCount: a RowPill row with an icon + meta emits frame + icon +
// meta text + label text (4); a plain row with just a label emits 1.
func TestRenderIconRows_ShapeCount(t *testing.T) {
	r := newTestRenderer(t)
	r.cfg.icons = icons.Curated()
	ps := r.pres.AddSlide()
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(8), H: pptx.In(4)}
	v := IconRows{Rows: []IconRow{
		{Icon: "star", Label: RichText{{Text: "Chat"}}, Meta: RichText{{Text: "core"}}, Tone: RowPill},
		{Label: RichText{{Text: "plain"}}},
	}}
	r.renderIconRows(ps, box, v)
	// row0: pill(1)+icon(1)+meta(1)+label(1)=4 ; row1: label(1)=1 → 5.
	if r.stats.Shapes != 5 {
		t.Errorf("emitted %d shapes, want 5 (warnings: %v)", r.stats.Shapes, r.stats.Warnings)
	}
}
