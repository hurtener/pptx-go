package scene

import (
	"errors"
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
	"github.com/hurtener/pptx-go/scene/layout"
)

// validateIconRefs is the registry-aware half of Stage-1 validation (RFC
// §14.1/§14.4, D-043): every Card's non-empty Icon name must resolve in the
// render's icon registry (curated ∪ extensions). It runs in Render, before any
// slide composes, and recurses container children. Mirrors validateOrnamentRefs.
func validateIconRefs(s Scene, reg *icons.Registry) error {
	var errs []error
	for i := range s.Slides {
		sl := &s.Slides[i]
		where := sl.ID
		if where == "" {
			where = fmt.Sprintf("#%d", i)
		}
		check := func(name, kind string) {
			if name == "" {
				return
			}
			if _, ok := reg.Lookup(name); !ok {
				errs = append(errs, fmt.Errorf(
					"slide %s: %s icon %q is not a curated or registered icon (have %v)",
					where, kind, name, reg.Names()))
			}
		}
		walkIconRefs(sl.Nodes, check)
	}
	return errors.Join(errs...)
}

// walkIconRefs visits every icon name in a node tree — a Card's Icon and each
// FlowStep's Icon — recursing into container children (CardSection bodies,
// nested cards, columns, grid cells). The callback receives the name and the
// owning node kind (for the error message).
func walkIconRefs(nodes []SlideNode, fn func(name, kind string)) {
	for _, n := range nodes {
		switch v := n.(type) {
		case Card:
			fn(v.Icon, "card")
			walkIconRefs(v.Body, fn)
		case CardSection:
			walkIconRefs(v.Body, fn)
		case Flow:
			for _, st := range v.Steps {
				fn(st.Icon, "flow step")
			}
		case TwoColumn:
			walkIconRefs(v.Left, fn)
			walkIconRefs(v.Right, fn)
		case Grid:
			walkIconRefs(v.Cells, fn)
		case Bento:
			walkIconRefs(v.cellNodes(), fn)
		}
	}
}

// Card composer (RFC §11.2 / §12, D-043). A Card renders native chrome — a
// rounded-rect background, a left accent stripe, and an optional header row
// (icon + eyebrow + header + right-aligned pill) — then lays out its leaf body
// in the region below the header per BodyLayout. CardSection shares the chrome
// (render_card_section.go). The icon, when present, resolves through the render's
// closed-name icon registry and renders as a native custGeom shape (not media),
// so a plain card stays parallel-safe.

// cardChrome is the chrome inputs shared by Card and CardSection. The rich-visual
// fields (headerFill/statusDot/watermark, D-054) are Card-only — CardSection
// leaves them at their zero values.
type cardChrome struct {
	header     string
	eyebrow    string
	icon       string
	pill       string
	fill       ColorRole
	outline    bool
	border     BorderStyle
	size       CardSize
	layout     CardLayout
	elevation  ElevationRole
	headerFill *ColorRole // banded header region; nil = no band
	statusDot  *ColorRole // top-right status dot; nil = no dot
	watermark  string     // faint label behind the body; "" = none
}

// cardPadding returns the interior inset for a card size.
func (r *renderer) cardPadding(size CardSize) pptx.EMU {
	switch size {
	case CardSizeSM:
		return r.theme.ResolveSpace(pptx.SpaceSM)
	case CardSizeLG:
		return r.theme.ResolveSpace(pptx.SpaceXL)
	default:
		return r.theme.ResolveSpace(pptx.SpaceMD)
	}
}

const cardStripeW = pptx.EMU(45720) // 4pt accent stripe

// Header-row geometry, shared by renderCardChrome (emission) and cardHeaderBottom
// (the pure header-bottom computation) so the two never drift. Values are the
// pre-Phase-25 literals, extracted verbatim (value-identical → byte-identical).
const (
	cardIconSz      = pptx.EMU(411480) // In(0.45); icon box side
	cardEyebrowRowH = pptx.EMU(237744) // In(0.26); eyebrow (kicker) row height
	cardTitleRowH   = pptx.EMU(365760) // In(0.40); header title row height
	cardPillH       = pptx.EMU(274320) // In(0.30); header-pill height
)

// Rich-visual geometry (D-054).
const (
	cardStatusDotSz    = pptx.EMU(146304) // In(0.16); status-dot diameter
	cardWatermarkAlpha = 13000            // ~13% OOXML opacity; the ghosted watermark
)

// cardHeaderColumnW returns the true header text column width — the inner width
// minus the icon-left shift and the header-pill reservation — at which the
// eyebrow and title wrap. cardHeaderRowHeights and renderCardChrome share it so
// the wrapped-line counts (and thus the body Y, the header band, and the emitted
// text frames) never drift.
func (r *renderer) cardHeaderColumnW(box pptx.Box, c cardChrome) pptx.EMU {
	pad := r.cardPadding(c.size)
	gapSM := r.theme.ResolveSpace(pptx.SpaceSM)
	innerW := box.W - cardStripeW - 2*pad
	if innerW < 0 {
		innerW = 0
	}
	headerW := innerW
	if c.icon != "" && c.layout != CardLayoutIconTop {
		headerW = innerW - cardIconSz - gapSM
	}
	if c.pill != "" {
		pillW := pptx.In(1.0)
		if pillW > innerW {
			pillW = innerW
		}
		if reserve := pillW + gapSM; headerW > reserve {
			headerW -= reserve
		}
	}
	if headerW < 0 {
		headerW = 0
	}
	return headerW
}

// cardHeaderRowHeights returns the wrapped heights of the eyebrow and title rows
// (R10.1, D-070): each is the per-line constant times the number of lines the
// text wraps to at the header column width, so a header that wraps to N lines no
// longer collides with the body. A single-line eyebrow/title yields exactly the
// legacy fixed row height (byte-identical). Either is 0 when its text is empty.
func (r *renderer) cardHeaderRowHeights(box pptx.Box, c cardChrome) (eyebrowH, titleH pptx.EMU) {
	headerW := r.cardHeaderColumnW(box, c)
	if c.eyebrow != "" {
		lines := wrappedLines(RichText{{Text: c.eyebrow}}, pptx.TypeCaption, headerW, r.theme)
		eyebrowH = cardEyebrowRowH * pptx.EMU(lines)
	}
	if c.header != "" {
		lines := wrappedLines(RichText{{Text: c.header}}, pptx.TypeH3, headerW, r.theme)
		titleH = cardTitleRowH * pptx.EMU(lines)
	}
	return eyebrowH, titleH
}

// cardHeaderBottom returns the Y at which a card's body region begins (the
// header's bottom). It mirrors the vertical advance in renderCardChrome exactly
// — using the same shared row heights (wrapped-aware, R10.1) — so the header
// band (drawn before the header text) ends precisely where the body starts.
func (r *renderer) cardHeaderBottom(box pptx.Box, c cardChrome) pptx.EMU {
	pad := r.cardPadding(c.size)
	gapSM := r.theme.ResolveSpace(pptx.SpaceSM)
	eyebrowH, titleH := r.cardHeaderRowHeights(box, c)
	y := box.Y + pad
	hasIcon := c.icon != ""
	if hasIcon && c.layout == CardLayoutIconTop {
		y += cardIconSz + gapSM
	}
	if c.eyebrow != "" {
		y += eyebrowH
	}
	if c.header != "" {
		y += titleH
	}
	if hasIcon && c.layout != CardLayoutIconTop {
		if iconBottom := box.Y + pad + cardIconSz; y < iconBottom {
			y = iconBottom
		}
	}
	// The header pill shares the top header row; ensure the body starts below it
	// too, so a pill-only (or pill-without-title) card does not stack its body
	// over the pill (and the D-054 header band is sized to include the pill).
	if c.pill != "" {
		if pillBottom := box.Y + pad + cardPillH; y < pillBottom {
			y = pillBottom
		}
	}
	return y + gapSM
}

// renderCardChrome draws the background, accent stripe, and header row, and
// returns the body region (inset by padding, below the header). It is
// deterministic: integer-EMU geometry, no map iteration (D-035).
func (r *renderer) renderCardChrome(ps *pptx.Slide, box pptx.Box, c cardChrome, slideID string) pptx.Box {
	// 1. Background rounded-rect: fill + border + elevation shadow.
	opts := []pptx.ShapeOption{
		pptx.WithRadius(pptx.RadiusLG),
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(c.fill))),
	}
	switch c.border {
	case BorderNone:
		// explicit no border
	case BorderSolid:
		opts = append(opts, pptx.WithLine(pptx.Line{Width: pptx.Pt(1), Color: pptx.TokenColor(pptx.ColorSurfaceAlt)}))
	case BorderAccent:
		opts = append(opts, pptx.WithLine(pptx.Line{Width: pptx.Pt(1.5), Color: pptx.TokenColor(pptx.ColorAccent)}))
	default: // BorderDefault: defer to the legacy Outline bool (D-043)
		if c.outline {
			opts = append(opts, pptx.WithLine(pptx.Line{Width: pptx.Pt(1), Color: pptx.TokenColor(pptx.ColorSurfaceAlt)}))
		}
	}
	if c.elevation != pptx.ElevationFlat {
		opts = append(opts, pptx.WithElevation(c.elevation))
	}
	ps.AddShape(pptx.ShapeRoundRect, box, opts...)
	r.stats.Shapes++

	// 1b. Header band (D-054): a colored region across the top of the card, from
	// the card top down to where the body begins, while the body keeps Fill.
	// Drawn on top of the background but before the header text. Inert when unset.
	if c.headerFill != nil {
		if bandH := r.cardHeaderBottom(box, c) - box.Y; bandH > 0 {
			band := pptx.Box{X: box.X, Y: box.Y, W: box.W, H: bandH}
			ps.AddShape(pptx.ShapeRoundRect, band,
				pptx.WithRadius(pptx.RadiusLG),
				pptx.WithFill(pptx.SolidFill(pptx.TokenColor(*c.headerFill))))
			r.stats.Shapes++
		}
	}

	// 2. Left accent stripe (the card's accent marker, RFC §12.1). Inset
	// vertically by the corner radius so its square corners stay within the
	// card's rounded outline (a full-height stripe pokes past the rounding).
	// Omitted for BorderAccent: the full accent border already supplies the
	// accent, so a stripe there is a redundant doubly-blue left edge.
	if c.border != BorderAccent {
		rad := r.theme.ResolveRadius(pptx.RadiusLG)
		if half := box.H / 2; rad > half {
			rad = half
		}
		stripe := pptx.Box{X: box.X, Y: box.Y + rad, W: cardStripeW, H: box.H - 2*rad}
		if stripe.H > 0 {
			ps.AddShape(pptx.ShapeRect, stripe, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
			r.stats.Shapes++
		}
	}

	// 3. Header region, inset by padding (past the stripe).
	pad := r.cardPadding(c.size)
	gapSM := r.theme.ResolveSpace(pptx.SpaceSM)
	innerX := box.X + cardStripeW + pad
	innerW := box.W - cardStripeW - 2*pad
	if innerW < 0 {
		innerW = 0
	}
	// Wrapped header row heights (R10.1) — shared with cardHeaderBottom so the
	// emitted text frames, the header band, and the body Y agree.
	eyebrowH, titleH := r.cardHeaderRowHeights(box, c)
	y := box.Y + pad
	headerLeft := innerX
	headerW := innerW

	iconSz := cardIconSz
	hasIcon := c.icon != ""
	if hasIcon {
		iconBox := pptx.Box{X: innerX, Y: y, W: iconSz, H: iconSz}
		// The name resolved at Stage-1 (validateIconRefs) and curated/extension
		// SVGs are pre-validated, so both paths below are defensive: a miss here
		// is a wiring bug, surfaced as a warning rather than a silent drop.
		if svg, ok := r.cfg.icons.Lookup(c.icon); !ok {
			r.warn(slideID, fmt.Sprintf("card icon %q not found at compose (should have failed Stage-1)", c.icon))
		} else if _, err := ps.AddIcon(svg, iconBox); err != nil {
			r.warn(slideID, fmt.Sprintf("card icon %q failed to render: %v", c.icon, err))
		} else {
			r.stats.Shapes++
		}
		switch c.layout {
		case CardLayoutIconTop:
			y += iconSz + gapSM
		default: // icon-left: shift the header text right of the icon
			headerLeft = innerX + iconSz + gapSM
			headerW = innerW - iconSz - gapSM
		}
	}

	// Header pill, right-aligned in the header row (drawn before the text so the
	// header width can reserve space for it).
	if c.pill != "" {
		pillW := pptx.In(1.0)
		pillH := cardPillH
		if pillW > innerW {
			pillW = innerW
		}
		pillBox := pptx.Box{X: innerX + innerW - pillW, Y: y, W: pillW, H: pillH}
		ps.AddShape(pptx.ShapeRoundRect, pillBox,
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))),
			pptx.WithRadius(pptx.RadiusFull))
		r.stats.Shapes++
		ptf := ps.AddTextFrame(pillBox).Anchor(pptx.AnchorMiddle)
		pp := ptf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		pp.AddRun(c.pill, pptx.RunStyle{TypeRole: pptx.TypeCaption})
		r.stats.Shapes++
		if reserve := pillW + gapSM; headerW > reserve {
			headerW -= reserve
		}
	}

	// Eyebrow (kicker) above the header — sized to its wrapped line count (R10.1).
	if c.eyebrow != "" {
		ebBox := pptx.Box{X: headerLeft, Y: y, W: headerW, H: eyebrowH}
		tf := ps.AddTextFrame(ebBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(c.eyebrow, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextAccent)})
		r.stats.Shapes++
		y += eyebrowH
	}

	// Header title — sized to its wrapped line count so a long title never
	// collides with the body (R10.1).
	if c.header != "" {
		hBox := pptx.Box{X: headerLeft, Y: y, W: headerW, H: titleH}
		tf := ps.AddTextFrame(hBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(c.header, pptx.RunStyle{TypeRole: pptx.TypeH3, Bold: true})
		r.stats.Shapes++
		y += titleH
	}

	// For icon-left, make sure the body starts below the icon too.
	if hasIcon && c.layout != CardLayoutIconTop {
		if iconBottom := box.Y + pad + iconSz; y < iconBottom {
			y = iconBottom
		}
	}
	// The header pill shares the top header row; ensure the body starts below it
	// (mirrors cardHeaderBottom) so a pill-only / pill-without-title card does not
	// overlap its body onto the pill.
	if c.pill != "" {
		if pillBottom := box.Y + pad + cardPillH; y < pillBottom {
			y = pillBottom
		}
	}

	// 4. Body region: below the header, full inner width, down to the bottom pad.
	bodyY := y + gapSM
	bodyBox := pptx.Box{X: innerX, Y: bodyY, W: innerW, H: box.Bottom() - pad - bodyY}
	if bodyBox.H < 0 {
		bodyBox.H = 0
	}

	// 5. Status dot (D-054): a small filled dot in the top-right corner, inset by
	// the padding. Drawn over the header band/text; inert when unset.
	if c.statusDot != nil {
		dot := pptx.Box{X: box.X + box.W - pad - cardStatusDotSz, Y: box.Y + pad, W: cardStatusDotSz, H: cardStatusDotSz}
		ps.AddShape(pptx.ShapeEllipse, dot, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(*c.statusDot))))
		r.stats.Shapes++
	}

	// 6. Watermark (D-054): a large, low-opacity label drawn in the body region.
	// It is the last chrome shape, so it sits behind the body content the caller
	// stacks next. Token-bound faintness via TokenColorAlpha (P2); inert when "".
	if c.watermark != "" && bodyBox.H > 0 {
		tf := ps.AddTextFrame(bodyBox).Anchor(pptx.AnchorBottom)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignRight})
		p.AddRun(c.watermark, pptx.RunStyle{
			TypeRole: pptx.TypeDisplay,
			Color:    pptx.TokenColorAlpha(pptx.ColorAccent, cardWatermarkAlpha),
		})
		r.stats.Shapes++
	}

	return bodyBox
}

func (r *renderer) renderCard(ps *pptx.Slide, box pptx.Box, v Card, slideID string) {
	body := r.renderCardChrome(ps, box, cardChrome{
		header: v.Header, eyebrow: v.Eyebrow, icon: v.Icon, pill: v.HeaderPill,
		fill: v.Fill, outline: v.Outline, border: v.BorderStyle, size: v.Size,
		layout: v.Layout, elevation: v.Elevation,
		headerFill: v.HeaderFill, statusDot: v.StatusDot, watermark: v.Watermark,
	}, slideID)

	if v.BodyLayout == BodyHorizontal && len(v.Body) > 0 {
		gap := r.theme.ResolveSpace(pptx.SpaceMD)
		weights := make([]int, len(v.Body))
		for i := range weights {
			weights[i] = 1
		}
		cols := layout.Columns(body, weights, gap)
		for i, n := range v.Body {
			if i < len(cols) {
				r.renderNode(ps, cols[i], n, slideID, HAlignLeft)
			}
		}
		return
	}
	for _, pl := range r.stackIn(body, v.Body, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
}
