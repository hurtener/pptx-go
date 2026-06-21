package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box guards for the card header geometry (Wave 8 checkpoint).

// TestCardHeaderConstants guards the header-row constant extraction (D-054): the
// constants must stay value-identical to their In() literals, or the refactored
// bare-card path silently drifts (there is no byte-golden for a bare card).
func TestCardHeaderConstants(t *testing.T) {
	cases := []struct {
		name      string
		got, want pptx.EMU
	}{
		{"cardIconSz", cardIconSz, pptx.In(0.45)},
		{"cardEyebrowRowH", cardEyebrowRowH, pptx.In(0.26)},
		{"cardTitleRowH", cardTitleRowH, pptx.In(0.40)},
		{"cardPillH", cardPillH, pptx.In(0.30)},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want the In() literal %d", c.name, c.got, c.want)
		}
	}
}

// TestCardHeaderBottom_PillRow guards the checkpoint N1 fix: a pill-only card's
// header bottom sits below the pill, so the body (and the D-054 header band) does
// not overlap it.
func TestCardHeaderBottom_PillRow(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(3)}
	pad := r.cardPadding(CardSizeMD)

	pill := r.cardHeaderBottom(box, cardChrome{pill: "NEW", size: CardSizeMD})
	if pillBottom := box.Y + pad + cardPillH; pill < pillBottom {
		t.Errorf("pill-only header bottom %d should be >= pill bottom %d (body would overlap the pill)", pill, pillBottom)
	}

	// The pill clause is what lifts it: a card with no header content at all stops
	// higher than the pill card.
	bare := r.cardHeaderBottom(box, cardChrome{size: CardSizeMD})
	if bare >= pill {
		t.Errorf("the pill should advance the header bottom: bare=%d pill=%d", bare, pill)
	}

	// A pill alongside a title (taller than the pill) is unaffected — the clamp
	// only matters when the pill is the tallest header element.
	withTitle := r.cardHeaderBottom(box, cardChrome{pill: "NEW", header: "Title", size: CardSizeMD})
	noPillTitle := r.cardHeaderBottom(box, cardChrome{header: "Title", size: CardSizeMD})
	if withTitle != noPillTitle {
		t.Errorf("pill+title header bottom %d should equal title-only %d (title is taller than the pill)", withTitle, noPillTitle)
	}
}
