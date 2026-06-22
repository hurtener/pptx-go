package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Density-aware card padding (Phase 45, R10.7). White-box tests for
// cardPaddingFor: a tighter PaddingScale shrinks the inset (and grows the body),
// the SpaceXS floor caps an extreme scale, and the default is byte-identical.

func paddingBox() pptx.Box { return pptx.Box{X: 0, Y: 0, W: pptx.In(4), H: pptx.In(3)} }

// TestCardPaddingScale_DefaultByteIdentical: scale 0 and 10000 both return the
// size-resolved base padding unchanged.
func TestCardPaddingScale_DefaultByteIdentical(t *testing.T) {
	r := newTestRenderer(t)
	base := r.cardPadding(CardSizeMD)
	if got := r.cardPaddingFor(cardChrome{size: CardSizeMD, paddingScale: 0}); got != base {
		t.Errorf("scale 0: padding = %d, want base %d", got, base)
	}
	if got := r.cardPaddingFor(cardChrome{size: CardSizeMD, paddingScale: 10000}); got != base {
		t.Errorf("scale 10000: padding = %d, want base %d", got, base)
	}
}

// TestCardPaddingScale_TighterReducesInset: a 5000 scale halves the inset (still
// above the floor) and, routed through cardHeaderBottom, grows the body region.
func TestCardPaddingScale_TighterReducesInset(t *testing.T) {
	r := newTestRenderer(t)
	box := paddingBox()
	base := r.cardPadding(CardSizeLG) // SpaceXL, comfortably above SpaceXS

	tight := r.cardPaddingFor(cardChrome{size: CardSizeLG, paddingScale: 5000})
	if tight >= base {
		t.Errorf("tighter scale: padding %d should be < base %d", tight, base)
	}
	if want := base * 5000 / 10000; tight != want {
		t.Errorf("tighter scale: padding = %d, want %d (50%%)", tight, want)
	}
	floor := r.theme.ResolveSpace(pptx.SpaceXS)
	if tight < floor {
		t.Errorf("tighter padding %d below the floor %d", tight, floor)
	}
	// The body starts higher (smaller top inset) with a tighter scale → more body.
	full := r.cardHeaderBottom(box, cardChrome{header: "H", size: CardSizeLG})
	tighter := r.cardHeaderBottom(box, cardChrome{header: "H", size: CardSizeLG, paddingScale: 5000})
	if tighter >= full {
		t.Errorf("tighter card body should start higher: tighter=%d full=%d", tighter, full)
	}
}

// TestCardPaddingScale_FloorsAtMin: an extreme tighten floors the inset at SpaceXS.
func TestCardPaddingScale_FloorsAtMin(t *testing.T) {
	r := newTestRenderer(t)
	floor := r.theme.ResolveSpace(pptx.SpaceXS)
	got := r.cardPaddingFor(cardChrome{size: CardSizeSM, paddingScale: 1}) // 0.01% → would be ~0
	if got != floor {
		t.Errorf("extreme tighten: padding = %d, want the floor %d", got, floor)
	}
}

// TestCardPaddingScale_Loosens: a scale above 10000 increases the inset.
func TestCardPaddingScale_Loosens(t *testing.T) {
	r := newTestRenderer(t)
	base := r.cardPadding(CardSizeMD)
	got := r.cardPaddingFor(cardChrome{size: CardSizeMD, paddingScale: 15000})
	if want := base * 15000 / 10000; got != want {
		t.Errorf("loosen scale: padding = %d, want %d (150%%)", got, want)
	}
	if got <= base {
		t.Errorf("loosen scale: padding %d should exceed base %d", got, base)
	}
}
