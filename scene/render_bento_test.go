package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Bento row-labeled grid (Phase 27, R5 c). White-box geometry tests for span
// alignment and gutter reservation; the render/validation/determinism black-box
// tests live in render_bento_render_test.go (package scene_test).

func bentoBox() pptx.Box { return pptx.Box{X: 0, Y: 0, W: pptx.In(12), H: pptx.In(6)} }

// TestBentoGeometry_SpanWidths is acceptance criterion 1: a span-2 cell is about
// twice a span-1 cell (plus the inter-unit gap), so columns align across rows.
func TestBentoGeometry_SpanWidths(t *testing.T) {
	theme := pptx.DefaultTheme()
	gap := theme.ResolveSpace(pptx.SpaceMD)
	v := Bento{Columns: 3, Rows: []BentoRow{
		{Label: "A", Cells: []BentoCell{{Span: 2, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
		{Label: "B", Cells: []BentoCell{{Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
	}}
	gutterW, _, cells := bentoGeometry(bentoBox(), v, gap)
	if gutterW == 0 {
		t.Fatal("labeled bento should reserve a gutter")
	}
	if len(cells) != 2 || len(cells[0]) != 2 || len(cells[1]) != 3 {
		t.Fatalf("unexpected cell shape: %v", cells)
	}
	span1 := cells[1][0].W // a span-1 unit
	span2 := cells[0][0].W // a span-2 cell
	want := 2*span1 + gap  // span-2 = 2 units + the gap between them
	if span2 != want {
		t.Errorf("span-2 width = %d, want %d (2×span-1 + gap)", span2, want)
	}
	// Columns align: row B's three span-1 cells each equal the unit width.
	for i, c := range cells[1] {
		if c.W != span1 {
			t.Errorf("row B cell %d width = %d, want unit %d", i, c.W, span1)
		}
	}
	// Cells start after the gutter.
	if cells[0][0].X != gutterW+gap {
		t.Errorf("first cell X = %d, want gutter+gap %d", cells[0][0].X, gutterW+gap)
	}
}

// TestBentoGeometry_NoGutterWhenUnlabeled is acceptance criterion 2: a bento
// with no labeled row reserves no gutter and uses the full width.
func TestBentoGeometry_NoGutterWhenUnlabeled(t *testing.T) {
	theme := pptx.DefaultTheme()
	gap := theme.ResolveSpace(pptx.SpaceMD)
	v := Bento{Columns: 2, Rows: []BentoRow{
		{Cells: []BentoCell{{Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
	}}
	gutterW, _, cells := bentoGeometry(bentoBox(), v, gap)
	if gutterW != 0 {
		t.Errorf("unlabeled bento reserved a gutter: %d", gutterW)
	}
	if cells[0][0].X != 0 {
		t.Errorf("first cell X = %d, want 0 (no gutter)", cells[0][0].X)
	}
}

// TestBento_IsFlexible: a bento grows under VAlignFill (it is a container).
func TestBento_IsFlexible(t *testing.T) {
	if !isFlexible(Bento{}) {
		t.Error("Bento should be flexible")
	}
}
