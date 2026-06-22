package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Fit-to-region compression (Phase 40, R10.2). White-box tests for VAlignFit:
// fitCompress shrinks inter-node gaps toward the SpaceXS floor, then scales slot
// heights toward the pinned 0.60 ratio floor, all in deterministic integer /
// basis-point math; the placement path lands an over-full stack inside its
// region; and a stack that already fits is byte-identical to VAlignTop.

// fitBox is a fixed test region: 10" wide × 5" tall, origin at zero.
func fitBox() pptx.Box {
	return pptx.Box{X: 0, Y: 0, W: pptx.In(10), H: pptx.In(5)}
}

// TestFitCompress_GapOnlyFit: a stack that overflows by a little is fitted by
// shrinking the inter-node gaps alone — the slot heights are left untouched.
func TestFitCompress_GapOnlyFit(t *testing.T) {
	r := newTestRenderer(t)
	box := fitBox()
	gap := r.theme.ResolveSpace(pptx.SpaceMD) // Pt(8)

	heights := []pptx.EMU{pptx.In(1.6), pptx.In(1.6), pptx.In(1.6)}
	orig := append([]pptx.EMU(nil), heights...)
	bodyH := orig[0] + orig[1] + orig[2] // In(4.8)

	// totalH = bodyH + gap*2 > box.H, but bodyH + gapMin*2 < box.H, so gaps fix it.
	if bodyH+gap*2 <= box.H {
		t.Fatalf("test setup: stack does not overflow (bodyH=%d, box.H=%d)", bodyH, box.H)
	}
	effGap := r.fitCompress(heights, bodyH, gap, box)

	// Heights unchanged (gap step alone sufficed).
	for i := range heights {
		if heights[i] != orig[i] {
			t.Errorf("height[%d] changed: %d → %d (gap step should suffice)", i, orig[i], heights[i])
		}
	}
	// effGap within [gapMin, gap) and the realized total fits.
	gapMin := r.theme.ResolveSpace(pptx.SpaceXS)
	if effGap < gapMin || effGap >= gap {
		t.Errorf("effGap %d not in [%d, %d)", effGap, gapMin, gap)
	}
	if got := bodyH + effGap*2; got > box.H {
		t.Errorf("did not fit: realized total %d > box.H %d", got, box.H)
	}
}

// TestFitCompress_GapPlusHeightFit: a stack that still overflows at the gap floor
// is fitted by additionally scaling slot heights toward the ratio floor.
func TestFitCompress_GapPlusHeightFit(t *testing.T) {
	r := newTestRenderer(t)
	box := fitBox()
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	gapMin := r.theme.ResolveSpace(pptx.SpaceXS)

	heights := []pptx.EMU{pptx.In(2), pptx.In(2), pptx.In(2)}
	bodyH := pptx.In(6)

	effGap := r.fitCompress(heights, bodyH, gap, box)

	// Gaps bottomed out at the floor.
	if effGap != gapMin {
		t.Errorf("effGap = %d, want gapMin %d (gaps must hit the floor first)", effGap, gapMin)
	}
	// Heights were scaled down (compression engaged) but above the 0.60 floor.
	var newBody pptx.EMU
	for i, h := range heights {
		if h >= pptx.In(2) {
			t.Errorf("height[%d] not compressed: %d", i, h)
		}
		if h < pptx.In(2)*6000/10000 {
			t.Errorf("height[%d] below the 0.60 floor: %d", i, h)
		}
		newBody += h
	}
	// The realized total lands inside the region (no off-box content).
	if got := newBody + effGap*2; got > box.H {
		t.Errorf("did not fit: realized total %d > box.H %d", got, box.H)
	}
}

// TestFitCompress_FloorCapped: an extreme overflow compresses to the pinned 0.60
// ratio floor and no further — residual overflow remains (and the placement path
// will still warn), proving the floor caps compression.
func TestFitCompress_FloorCapped(t *testing.T) {
	r := newTestRenderer(t)
	box := fitBox()
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	gapMin := r.theme.ResolveSpace(pptx.SpaceXS)

	heights := []pptx.EMU{pptx.In(5), pptx.In(5)}
	bodyH := pptx.In(10)

	effGap := r.fitCompress(heights, bodyH, gap, box)

	if effGap != gapMin {
		t.Errorf("effGap = %d, want gapMin %d", effGap, gapMin)
	}
	// Each height pinned to exactly 0.60 of preferred.
	for i, h := range heights {
		if want := pptx.In(5) * 6000 / 10000; h != want {
			t.Errorf("height[%d] = %d, want the 0.60 floor %d", i, h, want)
		}
	}
	// Residual overflow remains (the floor caps compression).
	if got := heights[0] + heights[1] + effGap; got <= box.H {
		t.Errorf("expected residual overflow at the floor; realized %d <= box.H %d", got, box.H)
	}
}

// TestFitCompress_Deterministic: identical inputs yield identical outputs.
func TestFitCompress_Deterministic(t *testing.T) {
	r := newTestRenderer(t)
	box := fitBox()
	gap := r.theme.ResolveSpace(pptx.SpaceMD)

	mk := func() ([]pptx.EMU, pptx.EMU) {
		h := []pptx.EMU{pptx.In(2), pptx.In(1.7), pptx.In(2.3), pptx.In(1.1)}
		return h, h[0] + h[1] + h[2] + h[3]
	}
	h1, body1 := mk()
	h2, body2 := mk()
	g1 := r.fitCompress(h1, body1, gap, box)
	g2 := r.fitCompress(h2, body2, gap, box)
	if g1 != g2 {
		t.Fatalf("non-deterministic gap: %d vs %d", g1, g2)
	}
	for i := range h1 {
		if h1[i] != h2[i] {
			t.Fatalf("non-deterministic height[%d]: %d vs %d", i, h1[i], h2[i])
		}
	}
}

// TestFitCompress_SingleNode (checkpoint NH1): the n==1 path (gapCount==0) — a
// single over-tall node is compressed to fit by the height step alone.
func TestFitCompress_SingleNode(t *testing.T) {
	r := newTestRenderer(t)
	box := fitBox() // H = In(5)
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	heights := []pptx.EMU{pptx.In(8)}
	effGap := r.fitCompress(heights, pptx.In(8), gap, box)
	if effGap != gap {
		t.Errorf("single node: effGap = %d, want unchanged gap %d (no inter-node gaps)", effGap, gap)
	}
	if heights[0] > box.H {
		t.Errorf("single node not compressed to fit: %d > box.H %d", heights[0], box.H)
	}
	if heights[0] >= pptx.In(8) {
		t.Errorf("single node should have shrunk from In(8), got %d", heights[0])
	}
}

// TestFitCompress_TwentyFivePercentBand (checkpoint NH3): the R10.2 headline
// acceptance — a stack ~25% over the region fits at the pinned steps.
func TestFitCompress_TwentyFivePercentBand(t *testing.T) {
	r := newTestRenderer(t)
	box := fitBox() // H = In(5)
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	// bodyH ≈ 1.25 × box.H across 3 nodes.
	heights := []pptx.EMU{pptx.In(2.1), pptx.In(2.1), pptx.In(2.05)} // Σ = In(6.25) = 1.25·In(5)
	bodyH := heights[0] + heights[1] + heights[2]
	effGap := r.fitCompress(heights, bodyH, gap, box)
	var total pptx.EMU
	for _, h := range heights {
		total += h
	}
	total += effGap * 2
	if total > box.H {
		t.Errorf("~25%% overflow did not fit: post-compression total %d > box.H %d", total, box.H)
	}
}

// TestFitCompress_ExtremeOverflowStillWarns (checkpoint NH2): a VAlignFit stack
// over-full beyond the 0.60 floor still fires the overflow warning (the floor
// caps compression, so residual overflow surfaces honestly).
func TestFitCompress_ExtremeOverflowStillWarns(t *testing.T) {
	r := newTestRenderer(t)
	nodes := make([]SlideNode, 14)
	for i := range nodes {
		nodes[i] = Callout{Body: RichText{{Text: "row"}}}
	}
	_ = r.layout(nodes, "fit", Alignment{Vertical: VAlignFit})
	if !hasOverflowWarning(r.stats, "fit") {
		t.Error("an extreme-overflow VAlignFit stack should still warn (floor caps compression)")
	}
}

// TestFitCompress_PlacementFitsWithinRegion is the R10.2 acceptance: an over-full
// (≤~25%) body stack rendered with VAlignFit places its last node bottom at or
// above the region bottom (no off-slide content), and the overflow warning is
// suppressed once it fits.
func TestFitCompress_PlacementFitsWithinRegion(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()
	gap := r.theme.ResolveSpace(pptx.SpaceMD)

	// Grow a uniform text stack until its preferred height overflows the region.
	mkNodes := func(n int) []SlideNode {
		out := make([]SlideNode, n)
		for i := range out {
			out[i] = Callout{Body: RichText{{Text: "row"}}}
		}
		return out
	}
	var nodes []SlideNode
	for n := 3; n <= 40; n++ {
		cand := mkNodes(n)
		var total pptx.EMU
		for _, nd := range cand {
			total += preferredHeight(nd, body.W, r.theme)
		}
		total += gap * pptx.EMU(n-1)
		if total > body.H && total <= body.H*125/100 {
			nodes = cand
			break
		}
	}
	if nodes == nil {
		t.Skip("could not build a moderate-overflow stack at this slide size")
	}

	pls := r.layout(nodes, "fit", Alignment{Vertical: VAlignFit})
	var maxBottom pptx.EMU
	for _, p := range pls {
		if b := p.box.Bottom(); b > maxBottom {
			maxBottom = b
		}
	}
	if maxBottom > body.Bottom() {
		t.Errorf("VAlignFit content spills off-region: last bottom %d > region bottom %d", maxBottom, body.Bottom())
	}
	// Fit succeeded → no overflow warning.
	if hasOverflowWarning(r.stats, "fit") {
		t.Errorf("overflow warning fired even though the stack was fitted")
	}
}

// TestFitCompress_ByteIdenticalWhenFits: with content that already fits, VAlignFit
// produces placements identical to VAlignTop (the compression branch is skipped).
func TestFitCompress_ByteIdenticalWhenFits(t *testing.T) {
	nodes := []SlideNode{
		Heading{Text: RichText{{Text: "One"}}, Level: 1},
		Prose{Paragraphs: []RichText{{{Text: "a"}}}},
		List{Items: []ListItem{{Text: RichText{{Text: "x"}}}}},
	}

	rTop := newTestRenderer(t)
	top := rTop.layout(nodes, "s", Alignment{Vertical: VAlignTop})

	rFit := newTestRenderer(t)
	fit := rFit.layout(nodes, "s", Alignment{Vertical: VAlignFit})

	if len(top) != len(fit) {
		t.Fatalf("placement count differs: top=%d fit=%d", len(top), len(fit))
	}
	for i := range top {
		if top[i].box != fit[i].box {
			t.Errorf("placement[%d] box differs: top=%+v fit=%+v", i, top[i].box, fit[i].box)
		}
	}
	// A fitting stack must not warn under either mode.
	if hasOverflowWarning(rFit.stats, "s") {
		t.Errorf("fitting VAlignFit stack should not warn")
	}
}

// TestVAlignFit_String guards the enum name.
func TestVAlignFit_String(t *testing.T) {
	if got := VAlignFit.String(); got != "fit" {
		t.Errorf("VAlignFit.String() = %q, want %q", got, "fit")
	}
}
