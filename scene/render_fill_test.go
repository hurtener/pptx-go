package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Vertical fill / grow-to-fit (Phase 23, R2). White-box tests for VAlignFill:
// fixed leaves stay at preferred height pinned to the top, flexible nodes grow
// to consume the leftover body height, the distribution is proportional and
// deterministic, and every non-fill path is untouched.

// TestIsFlexible covers the intrinsic flexible set: containers + Image/Chart
// grow; text leaves, atoms, and CodeBlock do not.
func TestIsFlexible(t *testing.T) {
	flexible := []SlideNode{
		Grid{}, TwoColumn{}, Card{}, CardSection{}, Table{}, Chart{}, Image{},
	}
	for _, n := range flexible {
		if !isFlexible(n) {
			t.Errorf("%T: want flexible", n)
		}
	}
	fixed := []SlideNode{
		Hero{}, Heading{}, Prose{}, List{}, Quote{}, Callout{}, Divider{},
		Chip{}, Arrow{}, SectionDivider{}, Flow{}, CodeBlock{},
	}
	for _, n := range fixed {
		if isFlexible(n) {
			t.Errorf("%T: want fixed (not flexible)", n)
		}
	}
}

// TestDistributeFill_Proportional: slack is shared in proportion to preferred
// height, and the last flexible node absorbs the rounding remainder so the
// added heights sum to exactly slack.
func TestDistributeFill_Proportional(t *testing.T) {
	// Heading (fixed), Image (flex, 2"), Image (flex, 1"); slack 3".
	nodes := []SlideNode{Heading{}, Image{}, Image{}}
	heights := []pptx.EMU{pptx.In(0.6), pptx.In(2), pptx.In(1)}
	distributeFill(nodes, heights, pptx.In(3))
	// idx1 gets 3 × 2/3 = 2"; idx2 (last) gets the remainder 3 − 2 = 1".
	if heights[0] != pptx.In(0.6) {
		t.Errorf("fixed node grew: %d", heights[0])
	}
	if heights[1] != pptx.In(4) { // 2" + 2"
		t.Errorf("flex[0]: want %d, got %d", pptx.In(4), heights[1])
	}
	if heights[2] != pptx.In(2) { // 1" + 1"
		t.Errorf("flex[1]: want %d, got %d", pptx.In(2), heights[2])
	}
}

// TestDistributeFill_EqualWhenZero: when the flexible nodes' preferred heights
// sum to zero, the slack is split equally (last absorbs the remainder).
func TestDistributeFill_EqualWhenZero(t *testing.T) {
	nodes := []SlideNode{Image{}, Image{}}
	heights := []pptx.EMU{0, 0}
	distributeFill(nodes, heights, pptx.In(3))
	if heights[0] != pptx.In(1.5) || heights[1] != pptx.In(1.5) {
		t.Errorf("equal split: got %d, %d; want %d each", heights[0], heights[1], pptx.In(1.5))
	}
}

// TestDistributeFill_NoFlexNoOp: with no flexible node, nothing grows.
func TestDistributeFill_NoFlexNoOp(t *testing.T) {
	nodes := []SlideNode{Heading{}, Prose{}}
	heights := []pptx.EMU{pptx.In(0.6), pptx.In(0.4)}
	distributeFill(nodes, heights, pptx.In(3))
	if heights[0] != pptx.In(0.6) || heights[1] != pptx.In(0.4) {
		t.Errorf("no-flex distributeFill mutated heights: %v", heights)
	}
}

// TestVAlignFill_GridFillsToBottom is acceptance criteria 1 & 2: the heading is
// pinned at the top and the grid's slot grows so its bottom reaches the body
// region's bottom margin — and the grown slot is taller than its preferred
// height, so the grid's cells fill the extra space.
func TestVAlignFill_GridFillsToBottom(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()
	grid := Grid{Columns: 2, Cells: []SlideNode{Divider{}, Divider{}, Divider{}, Divider{}}}
	nodes := []SlideNode{Heading{Text: RichText{{Text: "Top"}}, Level: 1}, grid}

	pls := stackedPlacements(r.layout(nodes, "s", Alignment{Vertical: VAlignFill}))
	if len(pls) != 2 {
		t.Fatalf("want 2 stacked placements, got %d", len(pls))
	}
	heading, gridPl := pls[0], pls[1]

	if heading.box.Y != body.Y {
		t.Errorf("heading not pinned at top: Y=%d, want %d", heading.box.Y, body.Y)
	}
	if gridPl.box.Bottom() != body.Bottom() {
		t.Errorf("grid did not fill to bottom margin: bottom=%d, want %d", gridPl.box.Bottom(), body.Bottom())
	}
	if pref := preferredHeight(grid, body.W, r.theme); gridPl.box.H <= pref {
		t.Errorf("grid slot did not grow: H=%d, preferred=%d", gridPl.box.H, pref)
	}
}

// TestVAlignFill_NoFlexMatchesTop is acceptance criterion 4: a VAlignFill slide
// with no flexible node lays out identically to VAlignTop (nothing to grow).
func TestVAlignFill_NoFlexMatchesTop(t *testing.T) {
	r := newTestRenderer(t)
	nodes := []SlideNode{
		Heading{Text: RichText{{Text: "H"}}, Level: 1},
		Prose{Paragraphs: []RichText{{{Text: "Body."}}}},
	}
	top := stackedPlacements(r.layout(nodes, "s", Alignment{Vertical: VAlignTop}))
	fill := stackedPlacements(r.layout(nodes, "s", Alignment{Vertical: VAlignFill}))
	if len(top) != len(fill) {
		t.Fatalf("placement count differs: top=%d fill=%d", len(top), len(fill))
	}
	for i := range top {
		if top[i].box != fill[i].box {
			t.Errorf("node %d box differs: top=%+v fill=%+v", i, top[i].box, fill[i].box)
		}
	}
}

// TestVAlignFill_OverflowStillWarns: when content already overflows (slack ≤ 0),
// fill grows nothing and the overflow warning still fires (composes with R1).
func TestVAlignFill_OverflowStillWarns(t *testing.T) {
	// A grid whose preferred height alone exceeds the body region. Cards are
	// flexible and need no asset (unlike Image/Chart, which fail Stage-1).
	cells := make([]SlideNode, 40)
	for i := range cells {
		cells[i] = Card{Header: "x"} // each ~1.2"+ preferred
	}
	sc := Scene{Slides: []SceneSlide{{
		ID:      "tall",
		Content: Alignment{Vertical: VAlignFill},
		Nodes:   []SlideNode{Grid{Columns: 2, Cells: cells}},
	}}}
	stats, err := Render(pptx.New(), sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	found := false
	for _, w := range stats.Warnings {
		if w.SlideID == "tall" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected overflow warning under VAlignFill when content exceeds body; got %v", stats.Warnings)
	}
}

// Fill cap (Phase 44, R10.6). White-box tests for VAlignFillCapped:
// distributeFillCapped bounds each flexible node's growth, and the alignedStackIn
// capped branch turns the residual slack into even spacing within the box.

// TestDistributeFillCapped_BoundsAndResidual: a sparse + dense flexible stack —
// each node grows by at most fillGrowthMaxBP × its preferred height, and the
// total growth used is less than the slack (the residual becomes spacing).
func TestDistributeFillCapped_BoundsAndResidual(t *testing.T) {
	// Two flexible nodes (Card): sparse 1", dense 2". Big slack so caps bind.
	nodes := []SlideNode{Card{}, Card{}}
	heights := []pptx.EMU{pptx.In(1), pptx.In(2)}
	orig := append([]pptx.EMU(nil), heights...)
	slack := pptx.In(10)

	used := distributeFillCapped(nodes, heights, slack)

	// Each node grew by no more than 1.0× its preferred height (capped).
	for i := range heights {
		grew := heights[i] - orig[i]
		if cap := orig[i] * fillGrowthMaxBP / 10000; grew > cap {
			t.Errorf("node %d grew %d, exceeds cap %d", i, grew, cap)
		}
	}
	// used = sum of capped growth = orig[0]+orig[1] (each doubled) = In(3) < slack.
	if used >= slack {
		t.Errorf("used %d should be < slack %d (residual must remain for spacing)", used, slack)
	}
	if want := orig[0] + orig[1]; used != want {
		t.Errorf("used = %d, want %d (each node doubled at the 1.0× cap)", used, want)
	}
}

// TestDistributeFillCapped_NoFlexNoOp: with no flexible node nothing grows.
func TestDistributeFillCapped_NoFlexNoOp(t *testing.T) {
	nodes := []SlideNode{Heading{}, Prose{}}
	heights := []pptx.EMU{pptx.In(0.6), pptx.In(0.4)}
	if used := distributeFillCapped(nodes, heights, pptx.In(5)); used != 0 {
		t.Errorf("no-flex distributeFillCapped used %d, want 0", used)
	}
	if heights[0] != pptx.In(0.6) || heights[1] != pptx.In(0.4) {
		t.Errorf("no-flex distributeFillCapped mutated heights: %v", heights)
	}
}

// TestFillCapped_EvenSpacingWithinBox: the alignedStackIn capped branch places a
// sparse+dense stack within the box, with the residual slack as even spacing (a
// non-zero top margin and widened gaps), not a single ballooned node.
func TestFillCapped_EvenSpacingWithinBox(t *testing.T) {
	r := newTestRenderer(t)
	box := r.bodyRegion() // the layout uses the body region as the stack box
	// Two flexible Grids with small preferred bodies (sparse), so the caps bind
	// and a large residual remains.
	nodes := []SlideNode{
		Grid{Columns: 1, Cells: []SlideNode{Card{Header: "a"}}},
		Grid{Columns: 1, Cells: []SlideNode{Card{Header: "b"}}},
	}
	pls := stackedPlacements(r.layout(nodes, "s", Alignment{Vertical: VAlignFillCapped}))
	if len(pls) != 2 {
		t.Fatalf("want 2 placements, got %d", len(pls))
	}
	// Top margin: the first node starts strictly below the box top (even spacing).
	if pls[0].box.Y <= box.Y {
		t.Errorf("capped fill first node Y (%d) should be below box top (%d) — top margin", pls[0].box.Y, box.Y)
	}
	// Last node bottom stays within the box.
	last := pls[len(pls)-1]
	if b := last.box.Y + last.box.H; b > box.Bottom() {
		t.Errorf("capped fill last node bottom %d exceeds box bottom %d", b, box.Bottom())
	}
	// The inter-node gap is widened beyond the standard gap (residual spread).
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	if actualGap := pls[1].box.Y - (pls[0].box.Y + pls[0].box.H); actualGap <= gap {
		t.Errorf("capped fill gap %d should exceed the standard gap %d (even spacing)", actualGap, gap)
	}
}

// TestVAlignFillCapped_String guards the enum name.
func TestVAlignFillCapped_String(t *testing.T) {
	if got := VAlignFillCapped.String(); got != "fill-capped" {
		t.Errorf("VAlignFillCapped.String() = %q, want %q", got, "fill-capped")
	}
}

// stackedPlacements keeps only the body-stack placements (drops decoration and
// section-divider full-slide overlays).
func stackedPlacements(pls []placement) []placement {
	var out []placement
	for _, pl := range pls {
		switch pl.node.(type) {
		case Decoration, SectionDivider:
		default:
			out = append(out, pl)
		}
	}
	return out
}
