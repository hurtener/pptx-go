package scene

import "github.com/hurtener/pptx-go/pptx"

// Stat composer (RFC §11.1, D-057). A Stat renders as native text — one anchored
// frame stacking the display-scale value, a muted label, and (when set) a
// directional delta colored by its tone. No product behavior (D-026): the engine
// renders Value/Delta verbatim and only maps the tone to a token color.

func (r *renderer) renderStat(ps *pptx.Slide, box pptx.Box, v Stat) {
	tf := ps.AddTextFrame(box).Anchor(pptx.AnchorMiddle)

	vp := tf.AddParagraph(pptx.ParagraphOpts{})
	scale := r.displayRunScale(v.AutoFit, v.Value, pptx.TypeDisplay, box.W)
	// Auto-contrast the (otherwise uncolored, near-black) value against the slide
	// variant surface (R11.2, D-082): nil on a light slide → byte-identical; a light
	// token on a dark-variant slide so the value is never black-on-dark. A Stat
	// placed on a strongly-colored card resolves against the slide surface, not the
	// card fill (leaf nodes do not receive their container surface — a documented
	// follow-up); callers needing that drive it via the surrounding card.
	vp.AddRun(v.Value, pptx.RunStyle{TypeRole: pptx.TypeDisplay, Bold: true, FontScale: scale, Color: r.onCardSurface(pptx.ColorCanvas)})

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
