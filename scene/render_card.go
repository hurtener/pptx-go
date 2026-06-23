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
		case Button:
			fn(v.LeadingIcon, "button leading")
			fn(v.TrailingIcon, "button trailing")
		case Checklist:
			for _, it := range v.Items {
				fn(it.Icon, "checklist item")
			}
		case ChipRow:
			for _, c := range v.Chips {
				fn(c.Icon, "chip")
			}
		case Banner:
			fn(v.Icon, "banner")
			walkIconRefs(v.Trailing, fn)
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
	header       string
	eyebrow      string
	icon         string
	pill         string
	fill         ColorRole
	outline      bool
	border       BorderStyle
	size         CardSize
	layout       CardLayout
	elevation    ElevationRole
	headerFill   *ColorRole // banded header region; nil = no band
	statusDot    *ColorRole // top-right status dot; nil = no dot
	watermark    string     // faint label behind the body; "" = none
	paddingScale int        // basis-point multiplier on the size padding; 0/10000 = unchanged (D-076)
	ribbon       *Ribbon    // pinned emphasis badge; nil = none (R12.3, D-098)
}

// Ribbon geometry (R12.3, D-098). Pinned layout metrics — a top-bar band height, a
// corner-tab height, and the corner star size. Not theme tokens (geometry).
const (
	ribbonTopBarH = pptx.EMU(311112) // In(0.34); the reserved RibbonTopBar band
	ribbonTabH    = pptx.EMU(274320) // In(0.30); a corner text tab
	ribbonStarSz  = pptx.EMU(365760) // In(0.40); the corner star glyph
	ribbonTabPadX = pptx.EMU(91440)  // In(0.10); corner-tab / top-bar label padding
	ribbonTabMinW = pptx.EMU(274320) // In(0.30); a corner tab's minimum width
)

// ribbonReserveOf returns the vertical band a RibbonTopBar reserves at the card top so
// the header (and body) shift down below it; 0 for the corner positions (overlays) and
// when there is no ribbon. Shared by cardHeaderBottom, renderCardChrome, and the slot
// estimate so the reserved band, the header text, and the body Y all agree.
func ribbonReserveOf(c cardChrome) pptx.EMU {
	if c.ribbon != nil && c.ribbon.Position == RibbonTopBar {
		return ribbonTopBarH
	}
	return 0
}

// ribbonColor / ribbonTextColor resolve a ribbon's fill and label colors: the fill is
// the caller's Color or ColorAccent (nil); the label is an explicit non-default TextColor
// or auto-contrast against the fill (the banner/card pattern). Token-bound (P2).
func ribbonColorRole(rb *Ribbon) pptx.ColorRole {
	if rb.Color != nil {
		return *rb.Color
	}
	return pptx.ColorAccent
}

func (r *renderer) ribbonTextColor(rb *Ribbon, fillRole pptx.ColorRole) pptx.Color {
	if rb.TextColor != TextPrimary {
		return pptx.TokenTextColor(rb.TextColor)
	}
	if c := r.onCardSurface(fillRole); c != nil {
		return c
	}
	return pptx.TokenTextColor(pptx.TextPrimary)
}

// cardPaddingBase returns the size-resolved interior inset for a card size (the
// base, before any PaddingScale). Token-bound (P2). Free function so the slot
// estimators (preferredHeight) share it with the composer (R10.10).
func cardPaddingBase(theme *pptx.Theme, size CardSize) pptx.EMU {
	switch size {
	case CardSizeSM:
		return theme.ResolveSpace(pptx.SpaceSM)
	case CardSizeLG:
		return theme.ResolveSpace(pptx.SpaceXL)
	default:
		return theme.ResolveSpace(pptx.SpaceMD)
	}
}

// cardPaddingScaled returns a card's interior inset: the size-resolved base
// padding scaled by c.paddingScale (basis points; 0 or 10000 = unchanged), floored
// at a pinned minimum (SpaceXS) so a tightened card never collapses its inset.
// Deterministic integer math; base and floor both resolve through theme spacing
// tokens — no literals (P2, D-076).
func cardPaddingScaled(theme *pptx.Theme, c cardChrome) pptx.EMU {
	base := cardPaddingBase(theme, c.size)
	if c.paddingScale <= 0 || c.paddingScale == 10000 {
		return base
	}
	pad := base * pptx.EMU(c.paddingScale) / 10000
	if min := theme.ResolveSpace(pptx.SpaceXS); pad < min {
		pad = min
	}
	return pad
}

// cardPadding / cardPaddingFor are the renderer-method wrappers used by the
// composer and the internal tests (so those call sites stay unchanged).
func (r *renderer) cardPadding(size CardSize) pptx.EMU   { return cardPaddingBase(r.theme, size) }
func (r *renderer) cardPaddingFor(c cardChrome) pptx.EMU { return cardPaddingScaled(r.theme, c) }

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

// Header-pill fit-to-label geometry (R11.5, D-085). A pinned layout metric (not a
// token): the horizontal padding each side of the pill label, sized to absorb a
// text frame's default inset so the measured label fits on one line; and a circular
// minimum so a one-character pill stays a proper chip.
const (
	cardPillPadX = pptx.EMU(91440)  // In(0.10); padding each side of the pill label
	cardPillMinW = pptx.EMU(274320) // In(0.30); minimum pill width (== cardPillH)
)

// cardPillWidthOf returns the width a header pill needs to fit its label on a single
// line: naturalWidth(label @ TypeCaption) + 2·cardPillPadX, floored at the circular
// minimum and clamped to the card inner width. Shared by cardHeaderColumnWOf (the
// header-width reservation) and renderCardChrome (the drawn pill) so the reserved
// and drawn widths never drift. Returns 0 for an empty pill. Deterministic (pure
// integer naturalWidth).
func cardPillWidthOf(theme *pptx.Theme, pill string, innerW pptx.EMU) pptx.EMU {
	if pill == "" {
		return 0
	}
	w := naturalWidth(RichText{{Text: pill, Style: RunStyle{TypeRole: pptx.TypeCaption}}}, theme) + 2*cardPillPadX
	if w < cardPillMinW {
		w = cardPillMinW
	}
	if w > innerW {
		w = innerW
	}
	return w
}

// cardHeaderColumnW returns the true header text column width — the inner width
// minus the icon-left shift and the header-pill reservation — at which the
// eyebrow and title wrap. cardHeaderRowHeights and renderCardChrome share it so
// the wrapped-line counts (and thus the body Y, the header band, and the emitted
// text frames) never drift.
func cardHeaderColumnWOf(theme *pptx.Theme, box pptx.Box, c cardChrome) pptx.EMU {
	pad := cardPaddingScaled(theme, c)
	gapSM := theme.ResolveSpace(pptx.SpaceSM)
	innerW := box.W - cardStripeW - 2*pad
	if innerW < 0 {
		innerW = 0
	}
	headerW := innerW
	if c.icon != "" && c.layout != CardLayoutIconTop {
		headerW = innerW - cardIconSz - gapSM
	}
	if c.pill != "" {
		// Reserve the pill width unconditionally (Wave-11 checkpoint H1, D-093): a
		// conditional `if headerW > reserve` left headerW at full width when a pill
		// label clamped to the whole inner width (pillW == innerW), so the title
		// overlapped the pill. Subtracting always collapses the header column to 0 in
		// that degenerate case instead of overlapping. Byte-identical on every normal
		// deck (headerW > pillW+gapSM there); floored at 0 below.
		headerW -= cardPillWidthOf(theme, c.pill, innerW) + gapSM
	}
	if headerW < 0 {
		headerW = 0
	}
	return headerW
}

func (r *renderer) cardHeaderColumnW(box pptx.Box, c cardChrome) pptx.EMU {
	return cardHeaderColumnWOf(r.theme, box, c)
}

// cardHeaderExtraHeight returns the slot height a card's header needs *beyond* the
// single-line baseline: the eyebrow/title lines past the first, each at its
// per-row constant, measured at the card's true header column width for a card of
// total width avail. It is 0 for a single-line (or empty) header, so the
// preferredHeight slot estimate stays byte-identical for single-line cards while a
// wrapped multi-line header grows the slot (R10.10, closing the R10.1 deferral).
func cardHeaderExtraHeight(theme *pptx.Theme, avail pptx.EMU, c cardChrome) pptx.EMU {
	headerW := cardHeaderColumnWOf(theme, pptx.Box{W: avail}, c)
	var extra pptx.EMU
	if c.eyebrow != "" {
		extra += cardEyebrowRowH * pptx.EMU(wrappedLines(RichText{{Text: c.eyebrow}}, pptx.TypeCaption, headerW, theme)-1)
	}
	if c.header != "" {
		extra += cardTitleRowH * pptx.EMU(wrappedLines(RichText{{Text: c.header}}, pptx.TypeH3, headerW, theme)-1)
	}
	// A RibbonTopBar reserves a band beyond the fixed chrome baseline (R12.3), so the
	// slot grows to keep the body below it; corner ribbons reserve nothing.
	extra += ribbonReserveOf(c)
	return extra
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
	pad := r.cardPaddingFor(c)
	gapSM := r.theme.ResolveSpace(pptx.SpaceSM)
	eyebrowH, titleH := r.cardHeaderRowHeights(box, c)
	// A RibbonTopBar reserves a band at the top, so the header (and body) start below
	// it (R12.3); corner ribbons reserve nothing.
	top := box.Y + ribbonReserveOf(c)
	y := top + pad
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
		if iconBottom := top + pad + cardIconSz; y < iconBottom {
			y = iconBottom
		}
	}
	// The header pill shares the top header row; ensure the body starts below it
	// too, so a pill-only (or pill-without-title) card does not stack its body
	// over the pill (and the D-054 header band is sized to include the pill).
	if c.pill != "" {
		if pillBottom := top + pad + cardPillH; y < pillBottom {
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
	pad := r.cardPaddingFor(c)
	gapSM := r.theme.ResolveSpace(pptx.SpaceSM)
	innerX := box.X + cardStripeW + pad
	innerW := box.W - cardStripeW - 2*pad
	if innerW < 0 {
		innerW = 0
	}
	// Wrapped header row heights (R10.1) — shared with cardHeaderBottom so the
	// emitted text frames, the header band, and the body Y agree.
	eyebrowH, titleH := r.cardHeaderRowHeights(box, c)
	// A RibbonTopBar reserves a band at the top (R12.3); the header starts below it.
	top := box.Y + ribbonReserveOf(c)
	y := top + pad
	headerLeft := innerX
	headerW := innerW

	// Auto-contrast surface for the header runs (R11.2, D-082): the eyebrow and
	// title overlap the header band when a HeaderFill is set, otherwise the card
	// Fill. onCardSurface returns a light token on a dark surface and nil (the
	// pre-R11.2 default) on a light one, so a light card is byte-identical.
	headerSurface := c.fill
	if c.headerFill != nil {
		headerSurface = *c.headerFill
	}

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
	// header width can reserve space for it). The pill sizes to its label on one
	// line (R11.5) — a fixed box wrapped a long label like "CUSTOMIZABLE" — sharing
	// cardPillWidthOf with the header-width reservation above so they agree.
	if c.pill != "" {
		pillW := cardPillWidthOf(r.theme, c.pill, innerW)
		pillH := cardPillH
		pillBox := pptx.Box{X: innerX + innerW - pillW, Y: y, W: pillW, H: pillH}
		ps.AddShape(pptx.ShapeRoundRect, pillBox,
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))),
			pptx.WithRadius(pptx.RadiusFull))
		r.stats.Shapes++
		// Single-line guarantee: if the label still cannot fit the pill's inner text
		// width (only when the pill was clamped to innerW), shrink it to one line via
		// FontScale (reusing the R10.5 fit primitive); 0 = no scale when it fits.
		pillNatW := naturalWidth(RichText{{Text: c.pill, Style: RunStyle{TypeRole: pptx.TypeCaption}}}, r.theme)
		pillScale := fitScale(pillNatW, pillW-2*cardPillPadX)
		ptf := ps.AddTextFrame(pillBox).Anchor(pptx.AnchorMiddle)
		pp := ptf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		// The pill sits on its own ColorSurfaceAlt fill — auto-contrast against it
		// (R11.2), not the card surface.
		pp.AddRun(c.pill, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: r.onCardSurface(pptx.ColorSurfaceAlt), FontScale: pillScale})
		r.stats.Shapes++
		// Reserve the pill width unconditionally (checkpoint H1, D-093) — mirrors
		// cardHeaderColumnWOf, so the reserved and drawn header columns agree even for
		// a full-inner-width pill (the header collapses rather than overlapping it).
		headerW -= pillW + gapSM
		if headerW < 0 {
			headerW = 0
		}
	}

	// Eyebrow (kicker) above the header — sized to its wrapped line count (R10.1).
	// The eyebrow keeps its accent tint when the accent clears a minimum contrast
	// against the surface behind it (byte-identical on the default light card),
	// else falls back to the auto-contrast token so it never goes invisible on a
	// same-hue header band (R11.2, D-082).
	if c.eyebrow != "" {
		var ebColor pptx.Color
		if r.accentLegible(headerSurface) {
			ebColor = pptx.TokenTextColor(pptx.TextAccent)
		} else {
			ebColor = r.onCardSurface(headerSurface)
		}
		ebBox := pptx.Box{X: headerLeft, Y: y, W: headerW, H: eyebrowH}
		tf := ps.AddTextFrame(ebBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(c.eyebrow, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: ebColor})
		r.stats.Shapes++
		y += eyebrowH
	}

	// Header title — sized to its wrapped line count so a long title never
	// collides with the body (R10.1). Auto-contrast against the surface (R11.2):
	// nil (the dark default) on a light card → byte-identical; a light token on a
	// dark card / dark variant so the title is never black-on-dark.
	if c.header != "" {
		hBox := pptx.Box{X: headerLeft, Y: y, W: headerW, H: titleH}
		tf := ps.AddTextFrame(hBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(c.header, pptx.RunStyle{TypeRole: pptx.TypeH3, Bold: true, Color: r.onCardSurface(headerSurface)})
		r.stats.Shapes++
		y += titleH
	}

	// For icon-left, make sure the body starts below the icon too.
	if hasIcon && c.layout != CardLayoutIconTop {
		if iconBottom := top + pad + iconSz; y < iconBottom {
			y = iconBottom
		}
	}
	// The header pill shares the top header row; ensure the body starts below it
	// (mirrors cardHeaderBottom) so a pill-only / pill-without-title card does not
	// overlap its body onto the pill.
	if c.pill != "" {
		if pillBottom := top + pad + cardPillH; y < pillBottom {
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
		dotX := box.X + box.W - pad - cardStatusDotSz
		// Anti-collision (R11.6, D-086): the dot and the header pill both anchor to
		// the top-right and would overlap. When both are set, place the dot to the
		// left of the pill (a gap apart) so their boxes are disjoint. Byte-identical
		// when only one of the two is set (dot stays in the corner).
		if c.pill != "" {
			pillX := innerX + innerW - cardPillWidthOf(r.theme, c.pill, innerW)
			dotX = pillX - gapSM - cardStatusDotSz
			if dotX < innerX {
				dotX = innerX
			}
		}
		dot := pptx.Box{X: dotX, Y: box.Y + pad, W: cardStatusDotSz, H: cardStatusDotSz}
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

	// 7. Ribbon (R12.3, D-098): a pinned emphasis badge drawn last (on top). A
	// RibbonTopBar's band was already reserved (the header shifted down); the corner
	// positions are overlays. Inert when unset.
	if c.ribbon != nil {
		r.renderCardRibbon(ps, box, c.ribbon, pad)
	}

	return bodyBox
}

// renderCardRibbon draws a Card.Ribbon (R12.3, D-098). RibbonTopBar is a full-width tab
// across the top with centered text; RibbonCornerTL/TR are content-fit text tabs pinned
// in the top corner; RibbonCornerStar is a star glyph. The diagonal rotated-band variant
// is deferred (the builder has no rotated-text primitive) — a horizontal corner tab
// carries the label instead. Deterministic integer-EMU; token-bound colors.
func (r *renderer) renderCardRibbon(ps *pptx.Slide, box pptx.Box, rb *Ribbon, pad pptx.EMU) {
	fillRole := ribbonColorRole(rb)
	textColor := r.ribbonTextColor(rb, fillRole)

	drawTab := func(tab pptx.Box) {
		ps.AddShape(pptx.ShapeRoundRect, tab,
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(fillRole))),
			pptx.WithRadius(pptx.RadiusSM))
		r.stats.Shapes++
		tf := ps.AddTextFrame(tab).Anchor(pptx.AnchorMiddle)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(rb.Text, pptx.RunStyle{TypeRole: pptx.TypeCaption, Bold: true, Color: textColor})
		r.stats.Shapes++
	}

	// tabWidth is the content-fit width of a corner tab's label, clamped to the card.
	tabWidth := func() pptx.EMU {
		w := naturalWidth(RichText{{Text: rb.Text, Style: RunStyle{TypeRole: pptx.TypeCaption}}}, r.theme) + 2*ribbonTabPadX
		if w < ribbonTabMinW {
			w = ribbonTabMinW
		}
		if w > box.W {
			w = box.W
		}
		return w
	}

	switch rb.Position {
	case RibbonCornerStar:
		starBox := pptx.Box{X: box.X + box.W - pad - ribbonStarSz, Y: box.Y + pad, W: ribbonStarSz, H: ribbonStarSz}
		if svg, ok := r.cfg.icons.Lookup("star"); ok {
			if _, err := ps.AddIcon(svg, starBox, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(fillRole)))); err == nil {
				r.stats.Shapes++
			}
		}
	case RibbonCornerTL:
		drawTab(pptx.Box{X: box.X, Y: box.Y, W: tabWidth(), H: ribbonTabH})
	case RibbonCornerTR:
		w := tabWidth()
		drawTab(pptx.Box{X: box.X + box.W - w, Y: box.Y, W: w, H: ribbonTabH})
	default: // RibbonTopBar — full-width tab across the reserved band
		drawTab(pptx.Box{X: box.X, Y: box.Y, W: box.W, H: ribbonTopBarH})
	}
}

func (r *renderer) renderCard(ps *pptx.Slide, box pptx.Box, v Card, slideID string) {
	// box is already clamped to the safe area by renderNode (R11.3/R11.12), so a tall
	// card never draws past the safe area.
	body := r.renderCardChrome(ps, box, cardChrome{
		header: v.Header, eyebrow: v.Eyebrow, icon: v.Icon, pill: v.HeaderPill,
		fill: v.Fill, outline: v.Outline, border: v.BorderStyle, size: v.Size,
		layout: v.Layout, elevation: v.Elevation,
		headerFill: v.HeaderFill, statusDot: v.StatusDot, watermark: v.Watermark,
		paddingScale: v.PaddingScale, ribbon: v.Ribbon,
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
	// Vertical body: route through the body-stack alignment engine so a card can
	// distribute its body (center / bottom / justify / fill / fit) within the
	// card body box instead of always top-anchoring (R10.4). BodyVAlign's zero
	// value (VAlignTop) reproduces the top-anchored stackIn layout byte-for-byte
	// when no body node carries a per-node Align override (alignedStackIn honors
	// those where the old stackIn ignored them — a deliberate improvement, D-073).
	for _, pl := range r.alignedStackIn(body, v.Body, slideID, Alignment{Vertical: v.BodyVAlign}) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
}
