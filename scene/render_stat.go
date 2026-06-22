package scene

import "github.com/hurtener/pptx-go/pptx"

// Stat composer (RFC §11.1, D-057). A Stat renders as native text — one anchored
// frame stacking the display-scale value, a muted label, and (when set) a
// directional delta colored by its tone. No product behavior (D-026): the engine
// renders Value/Delta verbatim and only maps the tone to a token color.

// statValueRoleLadder is the pinned shrink ladder for a Stat value (R11.8): try the
// display role, then H1, then H2. Stepping through real type roles keeps the value
// on a clean typographic step before any sub-role font scaling.
var statValueRoleLadder = []pptx.TypeRole{pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2}

// statValueFit returns the type role and FontScale a Stat value should render at to
// stay on a single line in boxW (R11.8, D-088). When AutoFit is off (or the value is
// empty) it returns (TypeDisplay, 0) — byte-identical to the pre-R11.8 render. When
// AutoFit is on it walks the role ladder and returns the first role whose value fits
// one line; if even the floor (TypeH2) wraps, it returns the floor plus a FontScale
// that shrinks it to one line. Pure / deterministic (integer wrappedLines/fitScale).
func (r *renderer) statValueFit(autofit bool, value string, boxW pptx.EMU) (pptx.TypeRole, float64) {
	if !autofit || value == "" {
		return pptx.TypeDisplay, 0
	}
	for _, role := range statValueRoleLadder {
		if wrappedLines(RichText{{Text: value}}, role, boxW, r.theme) <= 1 {
			return role, 0
		}
	}
	floor := pptx.TypeH2
	return floor, fitScale(naturalWidthAt(RichText{{Text: value}}, floor, r.theme), boxW)
}

func (r *renderer) renderStat(ps *pptx.Slide, box pptx.Box, v Stat) {
	tf := ps.AddTextFrame(box).Anchor(pptx.AnchorMiddle)

	vp := tf.AddParagraph(pptx.ParagraphOpts{})
	// Value overflow guard (R11.8): keep the value on one line via the pinned role
	// ladder (TypeDisplay → H1 → H2) plus a FontScale floor, so a wide value like
	// "$4,000+" never wraps and pushes the trailing caption down. Gated on AutoFit
	// (the D-074 opt-in); off → full TypeDisplay, byte-identical.
	valueRole, scale := r.statValueFit(v.AutoFit, v.Value, box.W)
	// Auto-contrast the (otherwise uncolored, near-black) value against the slide
	// variant surface (R11.2, D-082): nil on a light slide → byte-identical; a light
	// token on a dark-variant slide so the value is never black-on-dark. A Stat
	// placed on a strongly-colored card resolves against the slide surface, not the
	// card fill (leaf nodes do not receive their container surface — a documented
	// follow-up); callers needing that drive it via the surrounding card.
	vp.AddRun(v.Value, pptx.RunStyle{TypeRole: valueRole, Bold: true, FontScale: scale, Color: r.onCardSurface(pptx.ColorCanvas)})

	if v.Label != "" {
		lp := tf.AddParagraph(pptx.ParagraphOpts{})
		lp.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
	}

	if v.Delta != "" {
		dp := tf.AddParagraph(pptx.ParagraphOpts{})
		dp.AddRun(v.Delta, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: deltaToneColor(v.DeltaTone)})
	}

	r.stats.Shapes++
}

// deltaToneColor maps a delta direction to a token color: up → success, down →
// error, neutral → muted. Token-bound (P2), so a theme swap re-skins it.
func deltaToneColor(t DeltaTone) pptx.Color {
	switch t {
	case DeltaUp:
		return pptx.TokenColor(pptx.ColorSuccess)
	case DeltaDown:
		return pptx.TokenColor(pptx.ColorError)
	default: // DeltaNeutral
		return pptx.TokenTextColor(pptx.TextMuted)
	}
}
