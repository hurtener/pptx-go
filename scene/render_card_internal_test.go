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

// TestCardHeaderColumn_PillReservation guards the Wave-11 checkpoint H1 fix (D-093):
// the header column reserves the pill width unconditionally. A normal pill is
// byte-identical (headerW == innerW − (pillW + gapSM)); a pill clamped to the whole
// inner width collapses the header column to 0 instead of leaving it at full width
// (which let the title overlap the pill).
func TestCardHeaderColumn_PillReservation(t *testing.T) {
	r := newTestRenderer(t)
	gapSM := r.theme.ResolveSpace(pptx.SpaceSM)

	// Normal pill in a roomy card: byte-identical reservation.
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(4), H: pptx.In(3)}
	c := cardChrome{header: "Title", pill: "NEW", size: CardSizeMD}
	pad := cardPaddingScaled(r.theme, c)
	innerW := box.W - cardStripeW - 2*pad
	pillW := cardPillWidthOf(r.theme, c.pill, innerW)
	if got, want := r.cardHeaderColumnW(box, c), innerW-(pillW+gapSM); got != want {
		t.Errorf("normal-pill header column = %d, want innerW − (pillW+gap) = %d", got, want)
	}

	// A pill label far wider than the whole card → pillW clamps to innerW → the
	// header column collapses to 0 (no overlap), where the old conditional left it
	// at innerW.
	wide := cardChrome{header: "Title", pill: "AN ABSURDLY LONG PILL LABEL THAT EXCEEDS THE ENTIRE CARD WIDTH MANY TIMES OVER", size: CardSizeMD}
	narrow := pptx.Box{X: 0, Y: 0, W: pptx.In(1.2), H: pptx.In(3)}
	if got := r.cardHeaderColumnW(narrow, wide); got != 0 {
		t.Errorf("full-width-pill header column = %d, want 0 (collapsed, no overlap)", got)
	}
}

// TestCardBodyBelowWrappedHeader_AllCombos is the R11.1 acceptance golden: a
// deliberately long, wrapping header across every CardSize × CardLayout
// combination must (a) advance the body region top to at or below the header
// band bottom — no header/body overlap — and (b) when a HeaderFill band is
// enabled, size the band to exactly the header bottom so the band fully contains
// the wrapped header. A single-line header stays byte-identical to the legacy
// fixed advance for every combo. (R11.1, closed-by-D-070; verified by D-081.)
func TestCardBodyBelowWrappedHeader_AllCombos(t *testing.T) {
	const longHeader = "A deliberately long card header that cannot possibly fit on a single line in a narrow third-width card column"

	teal := ColorAccent
	sizes := []struct {
		name string
		size CardSize
	}{
		{"MD", CardSizeMD},
		{"SM", CardSizeSM},
		{"LG", CardSizeLG},
	}
	layouts := []struct {
		name   string
		layout CardLayout
	}{
		{"Default", CardLayoutDefault},
		{"IconTop", CardLayoutIconTop},
	}

	for _, sz := range sizes {
		for _, ly := range layouts {
			t.Run(sz.name+"/"+ly.name, func(t *testing.T) {
				r := newTestRenderer(t)
				pres := pptx.New()
				ps := pres.AddSlide()
				// A narrow (1/3-width) card so the long header wraps.
				box := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(3)}

				long := cardChrome{
					header:     longHeader,
					eyebrow:    "SECTION",
					icon:       "star", // exercise the icon-band advance (esp. IconTop)
					fill:       ColorSurface,
					headerFill: &teal, // exercise the D-054 header band containment
					size:       sz.size,
					layout:     ly.layout,
				}

				// (4) The test must not be vacuous: the long header must actually wrap.
				_, titleH := r.cardHeaderRowHeights(box, long)
				if lines := titleH / cardTitleRowH; lines < 2 {
					t.Fatalf("long header wrapped to %d line(s); expected >= 2 (test is vacuous)", lines)
				}

				// bandBottom is both the body-region top (cardHeaderBottom) and the
				// bottom of the D-054 header band: renderCardChrome draws the band at
				// height cardHeaderBottom(box,c) - box.Y, so the band bottom IS this
				// value by construction.
				bandBottom := r.cardHeaderBottom(box, long)
				body := r.renderCardChrome(ps, box, long, "s1")

				// (1) No header/body overlap AND no drift: the composed body top equals
				// cardHeaderBottom exactly, because renderCardChrome advances through
				// the same shared cardHeaderRowHeights. Body top == band bottom means
				// the header band meets the body with the wrapped header fully inside it.
				if body.Y != bandBottom {
					t.Errorf("body top %d != header band bottom %d (R11.1: band/body must agree, no overlap/gap)", body.Y, bandBottom)
				}

				// (3) A single-line header is byte-identical to the legacy fixed advance.
				single := long
				single.header = "Short"
				single.eyebrow = ""
				_, singleTitleH := r.cardHeaderRowHeights(box, single)
				if singleTitleH != cardTitleRowH {
					t.Errorf("single-line titleH = %d, want the legacy fixed %d (byte-identical, R11.1)", singleTitleH, cardTitleRowH)
				}
			})
		}
	}
}
