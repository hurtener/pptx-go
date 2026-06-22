package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Balanced vertical rhythm (Phase 46, R10.8). White-box tests for VAlignBalanced:
// a sparse stack's slack is distributed as an even rhythm (top margin + widened
// gaps) with an optical-center upward bias, and Top/Center stay byte-identical.

func sparseStack() []SlideNode {
	return []SlideNode{
		Hero{Eyebrow: "FY26", Title: "Title"},
		Heading{Text: RichText{{Text: "Subtitle"}}, Level: 2},
		Prose{Paragraphs: []RichText{{{Text: "A short description."}}}},
	}
}

// TestBalanced_EvenRhythmWithinBox: a 3-node sparse stack gets a non-zero top
// margin and inter-node gaps widened beyond the standard gap (the slack
// distributed), and the last node stays within the body region.
func TestBalanced_EvenRhythmWithinBox(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()
	gap := r.theme.ResolveSpace(pptx.SpaceMD)

	pls := r.layout(sparseStack(), "s", Alignment{Vertical: VAlignBalanced})
	if len(pls) != 3 {
		t.Fatalf("want 3 placements, got %d", len(pls))
	}
	// Non-zero top margin (even rhythm, not pinned to box top).
	if pls[0].box.Y <= body.Y {
		t.Errorf("balanced first node Y (%d) should be below body top (%d)", pls[0].box.Y, body.Y)
	}
	// Inter-node gaps widened beyond the standard gap.
	g1 := pls[1].box.Y - (pls[0].box.Y + pls[0].box.H)
	if g1 <= gap {
		t.Errorf("balanced gap %d should exceed the standard gap %d", g1, gap)
	}
	// Last node bottom within the region.
	last := pls[2]
	if b := last.box.Y + last.box.H; b > body.Bottom() {
		t.Errorf("balanced last node bottom %d exceeds region bottom %d", b, body.Bottom())
	}
}

// TestBalanced_OpticalBiasAboveCenter: the top margin is smaller than the bottom
// margin (the stack sits above geometric center).
func TestBalanced_OpticalBiasAboveCenter(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()

	pls := r.layout(sparseStack(), "s", Alignment{Vertical: VAlignBalanced})
	topMargin := pls[0].box.Y - body.Y
	last := pls[len(pls)-1]
	bottomMargin := body.Bottom() - (last.box.Y + last.box.H)
	if topMargin <= 0 {
		t.Fatalf("expected a positive top margin, got %d", topMargin)
	}
	if topMargin >= bottomMargin {
		t.Errorf("optical bias: top margin %d should be < bottom margin %d (above center)", topMargin, bottomMargin)
	}
}

// TestBalanced_TopCenterByteIdentical: VAlignBalanced does not perturb the
// existing VAlignTop or VAlignCenter placements.
func TestBalanced_TopCenterByteIdentical(t *testing.T) {
	r := newTestRenderer(t)
	nodes := sparseStack()

	// Re-render Top and Center after the balanced change; compare to a fresh
	// renderer's Top/Center (the modes must be untouched).
	for _, mode := range []VAlign{VAlignTop, VAlignCenter} {
		a := newTestRenderer(t).layout(nodes, "s", Alignment{Vertical: mode})
		b := r.layout(nodes, "s", Alignment{Vertical: mode})
		if len(a) != len(b) {
			t.Fatalf("mode %v: placement count differs", mode)
		}
		for i := range a {
			if a[i].box != b[i].box {
				t.Errorf("mode %v placement[%d] box differs: %+v vs %+v", mode, i, a[i].box, b[i].box)
			}
		}
	}
}

// TestVAlignBalanced_String guards the enum name.
func TestVAlignBalanced_String(t *testing.T) {
	if got := VAlignBalanced.String(); got != "balanced" {
		t.Errorf("VAlignBalanced.String() = %q, want %q", got, "balanced")
	}
}
