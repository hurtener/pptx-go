package scene

import "github.com/hurtener/pptx-go/pptx"

// Flow composer (RFC §11.1 / §12, D-044). A Flow renders a sequence of step
// pills joined by connector glyphs, horizontal or vertical. It composes the
// public builder only — pills are roundRects, connectors are preset shapes
// (rightArrow/downArrow, chevron, mathPlus, circularArrow) placed in the gaps;
// there is no anchored AddConnector. A step's optional icon resolves through the
// render's icon registry as a native custGeom, so a flow is media-free.

// flowGap is the extent reserved for each connector glyph between pills (and for
// the cycle return arrow).
const flowGap = pptx.EMU(457200) // 0.5"

func (r *renderer) renderFlow(ps *pptx.Slide, box pptx.Box, v Flow, slideID string) {
	n := len(v.Steps)
	if n == 0 {
		return
	}
	vertical := v.Orientation == FlowVertical
	gaps := n - 1
	if v.Connector == ConnectorCycle {
		gaps++ // a trailing slot for the return arrow
	}

	if vertical {
		pillH := box.H - pptx.EMU(gaps)*flowGap
		if pillH < 0 {
			pillH = 0
		}
		pillH /= pptx.EMU(n)
		y := box.Y
		for i, step := range v.Steps {
			r.renderFlowStep(ps, pptx.Box{X: box.X, Y: y, W: box.W, H: pillH}, step, slideID)
			y += pillH
			if i < n-1 {
				r.renderConnector(ps, pptx.Box{X: box.X, Y: y, W: box.W, H: flowGap}, v.Connector, true)
				y += flowGap
			}
		}
		if v.Connector == ConnectorCycle {
			r.renderReturnArrow(ps, pptx.Box{X: box.X, Y: y, W: box.W, H: flowGap}, true)
		}
		return
	}

	pillW := box.W - pptx.EMU(gaps)*flowGap
	if pillW < 0 {
		pillW = 0
	}
	pillW /= pptx.EMU(n)
	x := box.X
	for i, step := range v.Steps {
		r.renderFlowStep(ps, pptx.Box{X: x, Y: box.Y, W: pillW, H: box.H}, step, slideID)
		x += pillW
		if i < n-1 {
			r.renderConnector(ps, pptx.Box{X: x, Y: box.Y, W: flowGap, H: box.H}, v.Connector, false)
			x += flowGap
		}
	}
	if v.Connector == ConnectorCycle {
		r.renderReturnArrow(ps, pptx.Box{X: x, Y: box.Y, W: flowGap, H: box.H}, false)
	}
}

// renderFlowStep draws one pill: a rounded rect with an optional top-center
// icon, a centered label, and an optional detail line below.
func (r *renderer) renderFlowStep(ps *pptx.Slide, box pptx.Box, step FlowStep, slideID string) {
	ps.AddShape(pptx.ShapeRoundRect, box,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))),
		pptx.WithRadius(pptx.RadiusMD))
	r.stats.Shapes++

	pad := r.theme.ResolveSpace(pptx.SpaceSM)
	inner := pptx.Box{X: box.X + pad, Y: box.Y + pad, W: box.W - 2*pad, H: box.H - 2*pad}
	if inner.W < 0 {
		inner.W = 0
	}
	y := inner.Y

	if step.Icon != "" {
		iconSz := pptx.In(0.35)
		ib := pptx.Box{X: inner.X + (inner.W-iconSz)/2, Y: y, W: iconSz, H: iconSz}
		if svg, ok := r.cfg.icons.Lookup(step.Icon); !ok {
			r.warn(slideID, "flow step icon "+step.Icon+" not found at compose (should have failed Stage-1)")
		} else if _, err := ps.AddIcon(svg, ib); err == nil {
			r.stats.Shapes++
			y += iconSz + pad
		}
	}

	labelH := pptx.In(0.32)
	tf := ps.AddTextFrame(pptx.Box{X: inner.X, Y: y, W: inner.W, H: labelH}).Anchor(pptx.AnchorMiddle)
	p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	r.addRichText(ps, p, step.Label, pptx.TypeBody)
	r.stats.Shapes++
	y += labelH

	if len(step.Detail) > 0 {
		df := ps.AddTextFrame(pptx.Box{X: inner.X, Y: y, W: inner.W, H: inner.Bottom() - y})
		dp := df.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		r.addRichText(ps, dp, step.Detail, pptx.TypeCaption)
		r.stats.Shapes++
	}
}

// renderConnector draws one inter-step glyph centered in gap. ConnectorCycle's
// inter-pair glyph is a plain arrow; its return arrow is drawn separately.
func (r *renderer) renderConnector(ps *pptx.Slide, gap pptx.Box, kind ConnectorKind, vertical bool) {
	accent := pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))
	switch kind {
	case ConnectorPlus:
		sz := pptx.In(0.22)
		b := centerIn(gap, sz, sz)
		ps.AddShape(pptx.ShapeGeometry("mathPlus"), b, pptx.WithFill(accent))
		r.stats.Shapes++
	case ConnectorArrowDashed:
		r.renderDashedArrow(ps, gap, vertical)
	default: // ConnectorArrow and ConnectorCycle inter-pair glyph
		if vertical {
			b := centerIn(gap, pptx.In(0.3), gap.H*7/10)
			ps.AddShape(pptx.ShapeGeometry("downArrow"), b, pptx.WithFill(accent))
		} else {
			b := centerIn(gap, gap.W*7/10, pptx.In(0.3))
			ps.AddShape(pptx.ShapeRightArrow, b, pptx.WithFill(accent))
		}
		r.stats.Shapes++
	}
}

// renderDashedArrow draws a dashed line through the gap plus a small solid
// chevron head at the leading edge (D-044 — arrow_dashed has no one-shape form).
func (r *renderer) renderDashedArrow(ps *pptx.Slide, gap pptx.Box, vertical bool) {
	line := pptx.Line{Width: pptx.Pt(1.5), Color: pptx.TokenColor(pptx.ColorAccent), Dash: "dash"}
	accent := pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))
	headSz := pptx.In(0.2)
	if vertical {
		cx := gap.X + gap.W/2
		ps.AddShape(pptx.ShapeLine, pptx.Box{X: cx, Y: gap.Y, W: 1, H: gap.H - headSz},
			pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
		r.stats.Shapes++
		head := pptx.Box{X: cx - headSz/2, Y: gap.Bottom() - headSz, W: headSz, H: headSz}
		ps.AddShape(pptx.ShapeChevron, head, pptx.WithFill(accent), pptx.WithRotation(90))
		r.stats.Shapes++
		return
	}
	cy := gap.Y + gap.H/2
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: gap.X, Y: cy, W: gap.W - headSz, H: 1},
		pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	r.stats.Shapes++
	head := pptx.Box{X: gap.Right() - headSz, Y: cy - headSz/2, W: headSz, H: headSz}
	ps.AddShape(pptx.ShapeChevron, head, pptx.WithFill(accent))
	r.stats.Shapes++
}

// renderReturnArrow draws the cycle return glyph (a circular arrow) in the
// trailing slot after the last step.
func (r *renderer) renderReturnArrow(ps *pptx.Slide, gap pptx.Box, vertical bool) {
	sz := pptx.In(0.35)
	var b pptx.Box
	if vertical {
		b = centerIn(gap, sz, gap.H*7/10)
	} else {
		b = centerIn(gap, gap.W*7/10, sz)
	}
	ps.AddShape(pptx.ShapeGeometry("circularArrow"), b,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
	r.stats.Shapes++
}

// centerIn returns a w×h box centered in parent.
func centerIn(parent pptx.Box, w, h pptx.EMU) pptx.Box {
	return pptx.Box{X: parent.X + (parent.W-w)/2, Y: parent.Y + (parent.H-h)/2, W: w, H: h}
}
