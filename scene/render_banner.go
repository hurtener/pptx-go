package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// Banner composer (RFC §11.1 / §12, R12.6, D-097). A Banner renders a full-width
// RadiusLG filled strip with a leading icon + bold lead + body on the left and optional
// right-aligned Trailing children. Distinct from the side-bar Callout. The strip + icon
// + text are native (media-free); Trailing children render per their own policy.

// Pinned layout metrics (not theme tokens — geometry, not a visual property).
const (
	bannerPadX     = pptx.EMU(274320) // In(0.30); horizontal inset
	bannerPadY     = pptx.EMU(182880) // In(0.20); vertical inset
	bannerIconSz   = pptx.EMU(365760) // In(0.40); leading icon
	bannerIconGap  = pptx.EMU(164592) // In(0.18); icon-to-text gap
	bannerMinH     = pptx.EMU(914400) // In(1.0); minimum strip height
	bannerLeadLine = pptx.EMU(411480) // In(0.45); per lead line (TypeH3-ish)
	bannerBodyLine = pptx.EMU(292608) // In(0.32); per body line
)

// bannerFillRole resolves the strip fill: the caller's Fill, or ColorAccent when Fill is
// the zero value (ColorCanvas) — a banner is always a filled strip (D-097).
func bannerFillRole(v Banner) pptx.ColorRole {
	if v.Fill == ColorCanvas {
		return pptx.ColorAccent
	}
	return v.Fill
}

// bannerTextColor resolves the lead/body color: an explicit non-default TextColor
// verbatim, else auto-contrast against the fill (light on a dark fill, the default on
// light) so the banner is legible out of the box.
func (r *renderer) bannerTextColor(v Banner, fillRole pptx.ColorRole) pptx.Color {
	if v.TextColor != TextPrimary {
		return pptx.TokenTextColor(v.TextColor)
	}
	if c := r.onCardSurface(fillRole); c != nil {
		return c
	}
	return pptx.TokenTextColor(pptx.TextPrimary)
}

// bannerRegions splits the inner band width into the left text-column width (after the
// trailing region and the leading icon) and the right trailing-region width (0 when
// there are no trailing children). Shared by the composer and the slot estimator.
func bannerRegions(innerW pptx.EMU, v Banner, theme *pptx.Theme) (leftTextW, rightW pptx.EMU) {
	leftW := innerW
	if len(v.Trailing) > 0 {
		rightW = innerW * 30 / 100
		if rightW < pptx.In(1.8) {
			rightW = pptx.In(1.8)
		}
		if rightW > pptx.In(3.5) {
			rightW = pptx.In(3.5)
		}
		if rightW > innerW/2 {
			rightW = innerW / 2
		}
		leftW = innerW - rightW - theme.ResolveSpace(pptx.SpaceMD)
	}
	leftTextW = leftW
	if v.Icon != "" {
		leftTextW -= bannerIconSz + bannerIconGap
	}
	if leftTextW < 0 {
		leftTextW = 0
	}
	return leftTextW, rightW
}

// bannerPreferredHeight is the strip's slot height: the taller of the wrapped lead+body
// text and the stacked trailing children, plus vertical padding, floored at a minimum.
func bannerPreferredHeight(v Banner, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	innerW := avail - 2*bannerPadX
	leftTextW, rightW := bannerRegions(innerW, v, theme)
	textH := bannerLeadLine*pptx.EMU(wrappedLines(v.Lead, pptx.TypeH3, leftTextW, theme)) +
		bannerBodyLine*pptx.EMU(wrappedLines(v.Body, pptx.TypeBody, leftTextW, theme))
	trailingH := nodesHeight(v.Trailing, rightW, theme)
	h := textH
	if trailingH > h {
		h = trailingH
	}
	h += 2 * bannerPadY
	if h < bannerMinH {
		h = bannerMinH
	}
	return h
}

func (r *renderer) renderBanner(ps *pptx.Slide, box pptx.Box, v Banner, slideID string) {
	fillRole := bannerFillRole(v)
	ps.AddShape(pptx.ShapeRoundRect, box,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(fillRole))),
		pptx.WithRadius(pptx.RadiusLG))
	r.stats.Shapes++

	textColor := r.bannerTextColor(v, fillRole)
	inner := pptx.Box{X: box.X + bannerPadX, Y: box.Y + bannerPadY, W: box.W - 2*bannerPadX, H: box.H - 2*bannerPadY}
	leftTextW, rightW := bannerRegions(inner.W, v, r.theme)

	// Right region: the trailing children, stacked (the card-body mechanism).
	if rightW > 0 {
		rightRegion := pptx.Box{X: inner.Right() - rightW, Y: inner.Y, W: rightW, H: inner.H}
		for _, pl := range r.stackIn(rightRegion, v.Trailing, slideID) {
			r.renderNode(ps, pl.box, pl.node, slideID, HAlignLeft)
		}
	}

	// Left region: leading icon + bold lead + body.
	textX := inner.X
	if v.Icon != "" {
		iconBox := pptx.Box{X: inner.X, Y: inner.Y + (inner.H-bannerIconSz)/2, W: bannerIconSz, H: bannerIconSz}
		r.addBannerIcon(ps, iconBox, v.Icon, textColor)
		textX += bannerIconSz + bannerIconGap
	}
	tf := ps.AddTextFrame(pptx.Box{X: textX, Y: inner.Y, W: leftTextW, H: inner.H}).Anchor(pptx.AnchorMiddle)
	if len(v.Lead) > 0 {
		p := tf.AddParagraph(pptx.ParagraphOpts{LineHeight: r.lineH(pptx.TypeH3)})
		r.addBannerRuns(p, v.Lead, pptx.TypeH3, textColor, true)
	}
	if len(v.Body) > 0 {
		p := tf.AddParagraph(pptx.ParagraphOpts{LineHeight: r.lineH(pptx.TypeBody)})
		r.addBannerRuns(p, v.Body, pptx.TypeBody, textColor, false)
	}
	r.stats.Shapes++
}

// addBannerRuns adds rt to p at role, forcing the banner's contrast color on every run
// (so the lead/body stay legible on the fill) and bold for the lead phrase.
func (r *renderer) addBannerRuns(p *pptx.Paragraph, rt RichText, role pptx.TypeRole, color pptx.Color, bold bool) {
	for _, run := range rt {
		p.AddRun(run.Text, pptx.RunStyle{
			TypeRole: role,
			Bold:     bold || run.Style.Bold,
			Italic:   run.Style.Italic,
			Color:    color,
		})
	}
}

// addBannerIcon renders the leading icon as a native custGeom glyph filled with color.
func (r *renderer) addBannerIcon(ps *pptx.Slide, box pptx.Box, name string, color pptx.Color) {
	svg, ok := r.cfg.icons.Lookup(name)
	if !ok {
		r.warn("", fmt.Sprintf("banner icon %q not found at compose (should have failed Stage-1)", name))
		return
	}
	if _, err := ps.AddIcon(svg, box, pptx.WithFill(pptx.SolidFill(color))); err != nil {
		r.warn("", fmt.Sprintf("banner icon %q failed to render: %v", name, err))
		return
	}
	r.stats.Shapes++
}
