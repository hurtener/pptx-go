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
	gutterW, _, _, cells := bentoGeometry(bentoBox(), v, gap, nil)
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
	gutterW, _, _, cells := bentoGeometry(bentoBox(), v, gap, nil)
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

// TestBentoGeometry_EqualModeByteIdentical guards R10.3's byte-identical default:
// with rowHs == nil, every row gets the same equal height and the row Ys / cell
// boxes match the pre-refactor formula exactly.
func TestBentoGeometry_EqualModeByteIdentical(t *testing.T) {
	theme := pptx.DefaultTheme()
	gap := theme.ResolveSpace(pptx.SpaceMD)
	box := bentoBox()
	v := Bento{Columns: 2, Rows: []BentoRow{
		{Label: "A", Cells: []BentoCell{{Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
		{Label: "B", Cells: []BentoCell{{Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
		{Label: "C", Cells: []BentoCell{{Span: 1, Node: Prose{}}, {Span: 1, Node: Prose{}}}},
	}}
	_, rowYs, heights, cells := bentoGeometry(box, v, gap, nil)
	nRows := len(v.Rows)
	wantRowH := (box.H - gap*pptx.EMU(nRows-1)) / pptx.EMU(nRows)
	for ri := range v.Rows {
		if heights[ri] != wantRowH {
			t.Errorf("row %d height = %d, want equal %d", ri, heights[ri], wantRowH)
		}
		wantY := box.Y + pptx.EMU(ri)*(wantRowH+gap)
		if rowYs[ri] != wantY {
			t.Errorf("row %d Y = %d, want %d", ri, rowYs[ri], wantY)
		}
		for ci, c := range cells[ri] {
			if c.Y != wantY || c.H != wantRowH {
				t.Errorf("row %d cell %d Y/H = %d/%d, want %d/%d", ri, ci, c.Y, c.H, wantY, wantRowH)
			}
		}
	}
}

// TestBentoWeightedRows_DenseTallerAndFits is the R10.3 acceptance: a sparse row
// and a dense row in weighted mode get content-proportional heights (dense >
// sparse) that fit the region.
func TestBentoWeightedRows_DenseTallerAndFits(t *testing.T) {
	r := newTestRenderer(t)
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	box := bentoBox()
	sparse := Prose{Paragraphs: []RichText{{{Text: "one line"}}}}
	dense := List{Items: []ListItem{
		{Text: RichText{{Text: "alpha"}}},
		{Text: RichText{{Text: "bravo"}}},
		{Text: RichText{{Text: "charlie"}}},
		{Text: RichText{{Text: "delta"}}},
	}}
	v := Bento{Columns: 1, WeightedRows: true, Rows: []BentoRow{
		{Cells: []BentoCell{{Span: 1, Node: sparse}}},
		{Cells: []BentoCell{{Span: 1, Node: dense}}},
	}}
	hs := r.bentoWeightedRowHeights(box, v, gap)
	if len(hs) != 2 {
		t.Fatalf("want 2 row heights, got %d", len(hs))
	}
	if hs[1] <= hs[0] {
		t.Errorf("dense row (%d) should be taller than sparse row (%d)", hs[1], hs[0])
	}
	if total := hs[0] + hs[1] + gap; total > box.H {
		t.Errorf("weighted rows overflow: %d > box.H %d", total, box.H)
	}
	// Geometry honors the weighted heights and keeps the last row inside the box.
	_, _, _, cells := bentoGeometry(box, v, gap, hs)
	last := cells[len(cells)-1][0]
	if b := last.Y + last.H; b > box.Bottom() {
		t.Errorf("last row bottom %d exceeds box bottom %d", b, box.Bottom())
	}
}

// TestBentoWeightedRows_ClampsToFit: when the content's preferred row heights
// would overflow, the single basis-point scale clamps the rows to fit exactly.
func TestBentoWeightedRows_ClampsToFit(t *testing.T) {
	r := newTestRenderer(t)
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(12), H: pptx.In(1)} // deliberately short
	dense := List{Items: []ListItem{
		{Text: RichText{{Text: "a"}}}, {Text: RichText{{Text: "b"}}},
		{Text: RichText{{Text: "c"}}}, {Text: RichText{{Text: "d"}}},
	}}
	v := Bento{Columns: 1, WeightedRows: true, Rows: []BentoRow{
		{Cells: []BentoCell{{Span: 1, Node: dense}}},
		{Cells: []BentoCell{{Span: 1, Node: dense}}},
	}}
	hs := r.bentoWeightedRowHeights(box, v, gap)
	if total := hs[0] + hs[1] + gap; total > box.H {
		t.Errorf("clamp failed: rows + gap %d > box.H %d", total, box.H)
	}
	// Equal-density rows scale to (nearly) equal heights.
	if d := hs[0] - hs[1]; d > 1 || d < -1 {
		t.Errorf("equal-density rows diverged after clamp: %d vs %d", hs[0], hs[1])
	}
}
