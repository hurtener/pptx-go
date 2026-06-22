package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// R11.12 white-box invariants (D-092): on the same hostile inputs the black-box
// on-canvas suite renders, assert the other three structural invariants directly via
// the engine's fit helpers — (1) header band ≤ body top, (3) fit-required text is one
// line, (4) chrome text clears the contrast minimum. Together with
// TestAdversarial_AllBoxesOnCanvas (invariant 2) this is the reusable torture harness.

const advLong = "A deliberately long stretch of content that wraps across several lines under any reasonable column width"

// TestAdversarial_HeaderBandBelowBody: invariant (1) — for a hostile long header in
// a narrow card, the body region starts at or below the header band bottom.
func TestAdversarial_HeaderBandBelowBody(t *testing.T) {
	r := newTestRenderer(t)
	pres := pptx.New()
	ps := pres.AddSlide()
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(6)}
	teal := ColorAccent
	c := cardChrome{eyebrow: "VISION", header: advLong, pill: "FULLY CUSTOMIZABLE", headerFill: &teal, size: CardSizeMD}
	band := r.cardHeaderBottom(box, c)
	body := r.renderCardChrome(ps, box, c, "s1")
	if body.Y < band {
		t.Errorf("body top %d is above the header band bottom %d (overlap)", body.Y, band)
	}
}

// TestAdversarial_FitTextOneLine: invariant (3) — the fit-required chrome text (pill,
// stat value, join-badge) resolves to a single line under hostile labels/values.
func TestAdversarial_FitTextOneLine(t *testing.T) {
	r := newTestRenderer(t)

	// Pill: the fitted width holds the label at full size, or it is shrunk to fit.
	innerW := pptx.In(3)
	pillW := cardPillWidthOf(r.theme, "FULLY CUSTOMIZABLE", innerW)
	pillNat := naturalWidth(RichText{{Text: "FULLY CUSTOMIZABLE", Style: RunStyle{TypeRole: pptx.TypeCaption}}}, r.theme)
	scale := fitScale(pillNat, pillW-2*cardPillPadX)
	eff := pillNat
	if scale > 0 {
		eff = pptx.EMU(float64(pillNat) * scale)
	}
	if eff > pillW-2*cardPillPadX {
		t.Errorf("pill label effective width %d exceeds the pill text width %d (would wrap)", eff, pillW-2*cardPillPadX)
	}

	// Stat value: the role ladder keeps a wide value on one line in a narrow box.
	role, vscale := r.statValueFit(true, "$4,000,000+", pptx.In(2))
	vnat := naturalWidthAt(RichText{{Text: "$4,000,000+"}}, role, r.theme)
	veff := vnat
	if vscale > 0 {
		veff = pptx.EMU(float64(vnat) * vscale)
	}
	if veff > pptx.In(2) {
		t.Errorf("stat value effective width %d exceeds the box %d (would wrap)", veff, pptx.In(2))
	}
}

// TestAdversarial_ContrastPasses: invariant (4) — chrome text on hostile surfaces
// (dark fill, dark variant) clears the 4.5:1 contrast minimum.
func TestAdversarial_ContrastPasses(t *testing.T) {
	r := newTestRenderer(t)
	for _, bg := range []pptx.RGB{"111827", "1F2937", "0D9488", "2563EB", "DC2626"} {
		r.theme.Colors.Surfaces[pptx.ColorInfo] = bg
		c := r.onCardSurface(pptx.ColorInfo)
		var effective pptx.RGB = "000000"
		if c != nil {
			effective = r.theme.ResolveTextColor(pptx.TextInverse)
		}
		if ratio := contrastRatioT10(relLuminance(effective), relLuminance(bg)); ratio < accentMinContrastT10 {
			t.Errorf("surface %s: chrome text contrast %d.%d:1 < 4.5:1", bg, ratio/10, ratio%10)
		}
	}
}
