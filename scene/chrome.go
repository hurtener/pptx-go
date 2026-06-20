package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// Slide chrome (RFC §10.2, Deckard R3). Opt-in per-slide furniture drawn outside
// the body region: a top section eyebrow + hairline rule and a bottom footer
// (brand slot + "N / total" page number). bodyRegion shrinks by the band heights
// when chrome is enabled, so chrome never overlaps content. Native shapes reuse
// existing theme tokens (P2) — no new builder API, no new token. Chrome is a
// mechanism, not a judgment (D-026): the engine draws what it is handed and
// composes the page-number string, but invents no brand or section names.

// Chrome band geometry. Pinned compile-time EMU literals so output is
// worker-count independent (RFC §10.1). The eyebrow/footer heights also drive
// the bodyRegion shrink (see render.go).
const (
	chromeEyebrowH = pptx.EMU(274320) // ~0.30"; eyebrow caption row height
	chromeFooterH  = pptx.EMU(274320) // ~0.30"; footer row height
	chromeBandGap  = pptx.EMU(91440)  // ~0.10"; gap between a chrome band and the body region
	chromeRuleH    = pptx.EMU(9525)   // 0.75pt; hairline rule thickness
)

// chromeTotalFor resolves the page-number denominator: the caller's explicit
// Chrome.Total when positive, else the slide count.
func chromeTotalFor(s Scene) int {
	if s.Chrome.Total > 0 {
		return s.Chrome.Total
	}
	return len(s.Slides)
}

// chromeNeedsSerial reports whether enabling chrome forces the whole deck to
// compose sequentially. Only a brand *image* registers global media (the brand
// text and page number are native shapes); the same asset on every slide must be
// numbered in scene order for deterministic bytes.
func chromeNeedsSerial(s Scene) bool {
	return s.Chrome.Enabled && s.Chrome.BrandAsset != ""
}

// renderChrome draws the eyebrow band (when the slide sets a Section) and the
// footer band (always, when chrome is enabled). It is inert when chrome is
// disabled. Geometry is deterministic integer EMU; the page-number string is
// composed from the slide's number and the deck total.
func (r *renderer) renderChrome(ps *pptx.Slide, sl *SceneSlide) {
	if !r.chrome.Enabled {
		return
	}
	cx, cy := r.pres.SlideSize()
	left := bodyMargin
	width := pptx.EMU(cx) - 2*bodyMargin
	if width <= 0 {
		return
	}

	// Top eyebrow band: section label + hairline rule beneath it. Drawn only when
	// the slide carries a Section, so a divider/cover can stay eyebrow-free.
	if sl.Section != "" {
		ebBox := pptx.Box{X: left, Y: bodyMargin, W: width, H: chromeEyebrowH}
		tf := ps.AddTextFrame(ebBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(sl.Section, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++

		rule := pptx.Box{X: left, Y: bodyMargin + chromeEyebrowH, W: width, H: chromeRuleH}
		ps.AddShape(pptx.ShapeRect, rule, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))))
		r.stats.Shapes++
	}

	// Bottom footer band: brand slot (left half) + "N / total" page number
	// (right half, right-aligned).
	footerY := pptx.EMU(cy) - bodyMargin - chromeFooterH
	half := width / 2
	r.renderChromeBrand(ps, sl, pptx.Box{X: left, Y: footerY, W: half, H: chromeFooterH})

	page := sl.PageNumber
	if page == 0 {
		page = r.slideIndex + 1 // default: 1-based scene position
	}
	pnBox := pptx.Box{X: left + half, Y: footerY, W: width - half, H: chromeFooterH}
	tf := ps.AddTextFrame(pnBox).Anchor(pptx.AnchorMiddle)
	p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignRight})
	p.AddRun(fmt.Sprintf("%d / %d", page, r.chromeTotal),
		pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
	r.stats.Shapes++
}

// renderChromeBrand draws the footer-left brand slot: a resolved image when
// BrandAsset is set (an unresolved asset degrades to a LayoutWarning — the
// warn-don't-fail asset contract), else the Brand text. Nothing when both are
// empty.
func (r *renderer) renderChromeBrand(ps *pptx.Slide, sl *SceneSlide, box pptx.Box) {
	switch {
	case r.chrome.BrandAsset != "":
		data, ct, err := r.resolve(r.chrome.BrandAsset)
		if err != nil {
			r.warn(sl.ID, fmt.Sprintf("chrome brand asset %q unresolved: %v", r.chrome.BrandAsset, err))
			return
		}
		if _, aerr := ps.AddImage(pptx.ImageBytes(data, ct), box); aerr != nil {
			r.warn(sl.ID, fmt.Sprintf("chrome brand asset %q: %v", r.chrome.BrandAsset, aerr))
			return
		}
		r.stats.Shapes++
		r.stats.Assets++
	case r.chrome.Brand != "":
		tf := ps.AddTextFrame(box).Anchor(pptx.AnchorMiddle)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(r.chrome.Brand, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++
	}
}
