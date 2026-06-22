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

// TestCardHeaderBottom_WrappedTitle guards R10.1: a header that wraps to N lines
// at its column width makes the header bottom advance by N title rows, so the
// body begins below the wrapped header (no overlap). A single-line header is
// byte-identical to the legacy fixed row.
func TestCardHeaderBottom_WrappedTitle(t *testing.T) {
	r := newTestRenderer(t)
	// A 1/3-width card (the R10.1 acceptance scenario).
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(3)}

	short := r.cardHeaderBottom(box, cardChrome{header: "Short", size: CardSizeMD})
	long := r.cardHeaderBottom(box, cardChrome{
		header: "A deliberately long card header that cannot possibly fit on a single line in a narrow third-width card column",
		size:   CardSizeMD,
	})
	if long <= short {
		t.Fatalf("wrapped header bottom %d should exceed single-line %d (R10.1)", long, short)
	}

	// The advance is an exact multiple of the per-line row height.
	_, titleH := r.cardHeaderRowHeights(box, cardChrome{
		header: "A deliberately long card header that cannot possibly fit on a single line in a narrow third-width card column",
		size:   CardSizeMD,
	})
	if titleH%cardTitleRowH != 0 {
		t.Errorf("wrapped titleH %d is not a multiple of the per-line row %d", titleH, cardTitleRowH)
	}
	if lines := titleH / cardTitleRowH; lines < 2 {
		t.Errorf("expected the long header to wrap to >=2 lines, got %d", lines)
	}

	// Single-line header is byte-identical to the legacy fixed advance.
	_, singleH := r.cardHeaderRowHeights(box, cardChrome{header: "Short", size: CardSizeMD})
	if singleH != cardTitleRowH {
		t.Errorf("single-line titleH = %d, want the legacy fixed %d (byte-identical)", singleH, cardTitleRowH)
	}
}

// TestCardBodyBelowWrappedHeader is the R10.1 acceptance: the composed body
// region's top is at or below the rendered (wrapped) header's bottom — no
// vertical overlap — for a long header in a narrow card.
func TestCardBodyBelowWrappedHeader(t *testing.T) {
	r := newTestRenderer(t)
	pres := pptx.New()
	ps := pres.AddSlide()
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(3)}
	c := cardChrome{
		header: "A deliberately long card header that wraps to multiple lines in a narrow third-width card",
		size:   CardSizeMD,
	}
	body := r.renderCardChrome(ps, box, c, "s1")
	if body.Y < r.cardHeaderBottom(box, c) {
		t.Errorf("body top %d is above the wrapped header bottom %d (overlap, R10.1)", body.Y, r.cardHeaderBottom(box, c))
	}
}
