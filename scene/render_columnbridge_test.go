package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box tests for the TwoColumn column-bridge (R12.8, D-101): the band reserve, the
// per-configuration shape count, and the seam default's byte-identity.

// TestColumnBridge_ShapeCount: a labeled bridge emits a spanning line + two stubs + a
// pill + its text (5); an unlabeled bridge emits just the bracket (3).
func TestColumnBridge_ShapeCount(t *testing.T) {
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(9), H: pptx.In(4)}
	left := pptx.Box{X: 0, Y: pptx.In(0.5), W: pptx.In(4), H: pptx.In(3)}
	right := pptx.Box{X: pptx.In(5), Y: pptx.In(0.5), W: pptx.In(4), H: pptx.In(3)}

	r := newTestRenderer(t)
	ps := r.pres.AddSlide()
	r.renderColumnBridge(ps, box, left, right, TwoColumn{Join: JoinBadge, JoinLabel: "One agent", JoinPosition: JoinTopBridge})
	if r.stats.Shapes != 5 {
		t.Errorf("labeled bridge emitted %d shapes, want 5", r.stats.Shapes)
	}

	r2 := newTestRenderer(t)
	ps2 := r2.pres.AddSlide()
	r2.renderColumnBridge(ps2, box, left, right, TwoColumn{Join: JoinBadge, JoinPosition: JoinBottomBridge})
	if r2.stats.Shapes != 3 {
		t.Errorf("unlabeled bridge emitted %d shapes, want 3 (line + 2 stubs)", r2.stats.Shapes)
	}
}

// TestTwoColumn_BridgeReservesBand: a top/bottom bridge grows the TwoColumn slot by the
// band height; JoinSeam reserves nothing.
func TestTwoColumn_BridgeReservesBand(t *testing.T) {
	r := newTestRenderer(t)
	avail := pptx.In(9)
	base := TwoColumn{Join: JoinBadge, JoinLabel: "x", Left: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "a"}}}}}, Right: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "b"}}}}}}
	seam := preferredHeight(base, avail, r.theme)
	bridged := base
	bridged.JoinPosition = JoinTopBridge
	if got := preferredHeight(bridged, avail, r.theme); got-seam != bridgeBandH {
		t.Errorf("bridge band reserve = %d, want %d", got-seam, bridgeBandH)
	}
}
