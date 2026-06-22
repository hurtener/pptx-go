package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Card body vertical distribution (Phase 42, R10.4). White-box tests for the
// engine the card body now routes through (alignedStackIn): bottom pins the last
// node to the body bottom, justify spreads the gaps, and the zero value
// (VAlignTop) is placement-identical to the top-anchored stackIn the card used
// before.

// cardBodyBox is a tall card body region with a short body, leaving vertical
// slack so the distribution modes are observable.
func cardBodyBox() pptx.Box {
	return pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(4), H: pptx.In(5)}
}

func shortCardBody() []SlideNode {
	return []SlideNode{
		Heading{Text: RichText{{Text: "Plan"}}, Level: 3},
		List{Items: []ListItem{{Text: RichText{{Text: "alpha"}}}, {Text: RichText{{Text: "bravo"}}}}},
	}
}

// TestCardBodyVAlign_TopByteIdentical: the zero BodyVAlign (VAlignTop) routes
// through alignedStackIn and must yield placements identical to the legacy
// top-anchored stackIn — the byte-identical guarantee.
func TestCardBodyVAlign_TopByteIdentical(t *testing.T) {
	r := newTestRenderer(t)
	box := cardBodyBox()
	body := shortCardBody()

	top := r.alignedStackIn(box, body, "s", Alignment{Vertical: VAlignTop})
	plain := r.stackIn(box, body, "s")
	if len(top) != len(plain) {
		t.Fatalf("placement count differs: aligned=%d stackIn=%d", len(top), len(plain))
	}
	for i := range plain {
		if top[i].box != plain[i].box {
			t.Errorf("placement[%d] box differs: aligned=%+v stackIn=%+v", i, top[i].box, plain[i].box)
		}
	}
}

// TestCardBodyVAlign_BottomPinsLastNode: BodyVAlign=Bottom places the last body
// node's bottom at the card body bottom.
func TestCardBodyVAlign_BottomPinsLastNode(t *testing.T) {
	r := newTestRenderer(t)
	box := cardBodyBox()
	body := shortCardBody()

	pls := r.alignedStackIn(box, body, "s", Alignment{Vertical: VAlignBottom})
	last := pls[len(pls)-1]
	if got := last.box.Y + last.box.H; got != box.Bottom() {
		t.Errorf("last body node bottom = %d, want card body bottom %d", got, box.Bottom())
	}
	// And it must sit strictly below where Top would place it (slack pushed it down).
	top := r.alignedStackIn(box, body, "s", Alignment{Vertical: VAlignTop})
	if pls[len(pls)-1].box.Y <= top[len(top)-1].box.Y {
		t.Errorf("bottom-aligned last node Y (%d) should exceed top-aligned (%d)", pls[len(pls)-1].box.Y, top[len(top)-1].box.Y)
	}
}

// TestCardBodyVAlign_JustifySpreadsGaps: BodyVAlign=Justify widens the inter-node
// gap to consume the slack (the second node starts lower than under Top).
func TestCardBodyVAlign_JustifySpreadsGaps(t *testing.T) {
	r := newTestRenderer(t)
	box := cardBodyBox()
	body := shortCardBody()

	just := r.alignedStackIn(box, body, "s", Alignment{Vertical: VAlignJustify})
	top := r.alignedStackIn(box, body, "s", Alignment{Vertical: VAlignTop})
	if len(just) < 2 {
		t.Fatal("need >= 2 body nodes to observe gap spreading")
	}
	// First node stays at the top in both modes; the second is pushed down by the
	// widened gap under Justify.
	if just[0].box.Y != top[0].box.Y {
		t.Errorf("justify moved the first node: %d vs %d", just[0].box.Y, top[0].box.Y)
	}
	if just[1].box.Y <= top[1].box.Y {
		t.Errorf("justify second node Y (%d) should exceed top (%d)", just[1].box.Y, top[1].box.Y)
	}
}
