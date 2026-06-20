package scene

import "github.com/hurtener/pptx-go/pptx"

// CardSection composer (RFC §11.2 / §12, D-043). A CardSection is a top-level
// card whose body accepts containers (grid / two_column / nested cards) rather
// than leaves. It shares the card chrome (renderCardChrome) and lays its body
// out by stacking each container through the normal dispatch, so nesting — a
// card_section of a grid of cards — composes via the existing layout engine.
func (r *renderer) renderCardSection(ps *pptx.Slide, box pptx.Box, v CardSection, slideID string) {
	body := r.renderCardChrome(ps, box, cardChrome{header: v.Header}, slideID)
	for _, pl := range r.stackIn(body, v.Body, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
}
