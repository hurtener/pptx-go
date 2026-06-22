package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box tests for R11.5 header-pill fit-to-label (D-085).

// TestCardPillWidth_FitsLabel: the pill width tracks the label — a long label is
// wider than a short one, both are floored at the circular minimum, and the width
// is clamped to the card inner width.
func TestCardPillWidth_FitsLabel(t *testing.T) {
	r := newTestRenderer(t)
	innerW := pptx.In(3)

	if w := cardPillWidthOf(r.theme, "", innerW); w != 0 {
		t.Errorf("empty pill width = %d, want 0", w)
	}

	short := cardPillWidthOf(r.theme, "X", innerW)
	if short != cardPillMinW {
		t.Errorf("one-char pill width = %d, want the circular minimum %d", short, cardPillMinW)
	}

	long := cardPillWidthOf(r.theme, "FULLY CUSTOMIZABLE", innerW)
	if long <= short {
		t.Errorf("long pill width %d should exceed short %d", long, short)
	}
	// It sizes to naturalWidth + padding.
	want := naturalWidth(RichText{{Text: "FULLY CUSTOMIZABLE", Style: RunStyle{TypeRole: pptx.TypeCaption}}}, r.theme) + 2*cardPillPadX
	if long != want {
		t.Errorf("long pill width = %d, want naturalWidth+2pad = %d", long, want)
	}

	// Clamped to inner width when the label is enormous.
	clamped := cardPillWidthOf(r.theme, "A REALLY LONG PILL LABEL THAT EXCEEDS THE CARD", pptx.In(1))
	if clamped != pptx.In(1) {
		t.Errorf("over-long pill width = %d, want clamp to innerW %d", clamped, pptx.In(1))
	}
}

// TestCardPillWidth_ReservationMatchesDrawn: the header-column reservation
// (cardHeaderColumnWOf) shrinks by exactly the same pill width the renderer draws,
// so the reserved and drawn widths never drift.
func TestCardPillWidth_ReservationMatchesDrawn(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(4), H: pptx.In(3)}
	c := cardChrome{header: "Title", pill: "CUSTOMIZABLE", size: CardSizeMD}

	pad := cardPaddingScaled(r.theme, c)
	innerW := box.W - cardStripeW - 2*pad
	pillW := cardPillWidthOf(r.theme, c.pill, innerW)
	gapSM := r.theme.ResolveSpace(pptx.SpaceSM)

	headerW := r.cardHeaderColumnW(box, c)
	// With a title and a pill (no icon): headerW == innerW − (pillW + gapSM).
	if want := innerW - (pillW + gapSM); headerW != want {
		t.Errorf("header column width = %d, want innerW − (pillW+gap) = %d", headerW, want)
	}
}
