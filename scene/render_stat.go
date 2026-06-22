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
	vp.AddRun(v.Value, pptx.RunStyle{TypeRole: pptx.TypeDisplay, Bold: true, FontScale: scale})

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
