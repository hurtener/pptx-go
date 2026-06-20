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

// cardChrome is the chrome inputs shared by Card and CardSection.
type cardChrome struct {
	header    string
	eyebrow   string
	icon      string
	pill      string
	fill      ColorRole
	outline   bool
	border    BorderStyle
	size      CardSize
	layout    CardLayout
	elevation ElevationRole
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
	y := box.Y + pad
	headerLeft := innerX
	headerW := innerW

	iconSz := pptx.In(0.45)
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
		pillH := pptx.In(0.3)
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

	// Eyebrow (kicker) above the header.
	if c.eyebrow != "" {
		ebH := pptx.In(0.26)
		ebBox := pptx.Box{X: headerLeft, Y: y, W: headerW, H: ebH}
		tf := ps.AddTextFrame(ebBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(c.eyebrow, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextAccent)})
		r.stats.Shapes++
		y += ebH
	}

	// Header title.
	if c.header != "" {
		hH := pptx.In(0.4)
		hBox := pptx.Box{X: headerLeft, Y: y, W: headerW, H: hH}
		tf := ps.AddTextFrame(hBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(c.header, pptx.RunStyle{TypeRole: pptx.TypeH3, Bold: true})
		r.stats.Shapes++
		y += hH
	}

	// For icon-left, make sure the body starts below the icon too.
	if hasIcon && c.layout != CardLayoutIconTop {
		if iconBottom := box.Y + pad + iconSz; y < iconBottom {
			y = iconBottom
		}
	}

	// 4. Body region: below the header, full inner width, down to the bottom pad.
	bodyY := y + gapSM
	bodyBox := pptx.Box{X: innerX, Y: bodyY, W: innerW, H: box.Bottom() - pad - bodyY}
	if bodyBox.H < 0 {
		bodyBox.H = 0
	}
	return bodyBox
}

func (r *renderer) renderCard(ps *pptx.Slide, box pptx.Box, v Card, slideID string) {
	body := r.renderCardChrome(ps, box, cardChrome{
		header: v.Header, eyebrow: v.Eyebrow, icon: v.Icon, pill: v.HeaderPill,
		fill: v.Fill, outline: v.Outline, border: v.BorderStyle, size: v.Size,
		layout: v.Layout, elevation: v.Elevation,
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
