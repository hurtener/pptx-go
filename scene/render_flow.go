package scene

import "github.com/hurtener/pptx-go/pptx"

// Flow composer (RFC §11.1 / §12, D-044). A Flow renders a sequence of step
// pills joined by connector glyphs, horizontal or vertical. It composes the
// public builder only — pills are roundRects, connectors are preset shapes
// (rightArrow/downArrow, chevron, mathPlus) placed in the gaps; cycle adds a
// return arrow in a reserved band looping back to the first step. There is no
// anchored AddConnector. A step's optional icon resolves through the render's
// icon registry as a native custGeom, so a flow is media-free.

const (
	flowGap        = pptx.EMU(457200)  // 0.5" reserved per connector gap
	flowReturnBand = pptx.EMU(548640)  // 0.6" band reserved for the cycle return arrow
	flowMaxPillW   = pptx.EMU(5486400) // 6" — vertical pills are centered in a column, not full-bleed
)

func (r *renderer) renderFlow(ps *pptx.Slide, box pptx.Box, v Flow, slideID string) {
	if len(v.Steps) == 0 {
		return
	}
	if v.Orientation == FlowVertical {
		r.renderFlowVertical(ps, box, v, slideID)
		return
	}
	r.renderFlowHorizontal(ps, box, v, slideID)
}

func (r *renderer) renderFlowHorizontal(ps *pptx.Slide, box pptx.Box, v Flow, slideID string) {
	n := len(v.Steps)
	region := box
	if v.Connector == ConnectorCycle {
		region.H -= flowReturnBand // pills sit above the return band
	}
	pillW := region.W - pptx.EMU(n-1)*flowGap
	if pillW < 0 {
		pillW = 0
	}
	pillW /= pptx.EMU(n)

	x := region.X
	for i, step := range v.Steps {
		r.renderFlowStep(ps, pptx.Box{X: x, Y: region.Y, W: pillW, H: region.H}, step, slideID)
		x += pillW
		if i < n-1 {
			r.renderConnector(ps, pptx.Box{X: x, Y: region.Y, W: flowGap, H: region.H}, v.Connector, false)
			x += flowGap
		}
	}
	if v.Connector == ConnectorCycle && n >= 2 {
		firstCx := region.X + pillW/2
		lastCx := region.X + pptx.EMU(n-1)*(pillW+flowGap) + pillW/2
		r.renderReturnLoopH(ps, region, firstCx, lastCx)
	}
}

// renderReturnLoopH draws a feedback loop under a horizontal row: down from the
// last step, back across to the first, then up into it with an arrowhead. Thin
// accent lines + a small chevron — reads as a cycle, not a heavy block arrow.
func (r *renderer) renderReturnLoopH(ps *pptx.Slide, region pptx.Box, firstCx, lastCx pptx.EMU) {
	line := pptx.Line{Width: pptx.Pt(1.5), Color: pptx.TokenColor(pptx.ColorAccent)}
	accent := pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))
	top := region.Bottom()
	midY := top + flowReturnBand/2
	head := pptx.In(0.16)
	// down stub at the last step, across to the first, up stub into the first
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: lastCx, Y: top, W: 1, H: midY - top}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: firstCx, Y: midY, W: lastCx - firstCx, H: 1}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: firstCx, Y: top + head, W: 1, H: midY - top - head}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	ps.AddShape(pptx.ShapeChevron, pptx.Box{X: firstCx - head/2, Y: top, W: head, H: head}, pptx.WithFill(accent), pptx.WithRotation(270))
	r.stats.Shapes += 4
}

func (r *renderer) renderFlowVertical(ps *pptx.Slide, box pptx.Box, v Flow, slideID string) {
	n := len(v.Steps)
	// Center the pill column (not full-bleed); cycle reserves a right band.
	availW := box.W
	if v.Connector == ConnectorCycle {
		availW -= flowReturnBand
	}
	pillW := availW
	if pillW > flowMaxPillW {
		pillW = flowMaxPillW
	}
	pillX := box.X + (availW-pillW)/2

	pillH := box.H - pptx.EMU(n-1)*flowGap
	if pillH < 0 {
		pillH = 0
	}
	pillH /= pptx.EMU(n)

	y := box.Y
	for i, step := range v.Steps {
		r.renderFlowStep(ps, pptx.Box{X: pillX, Y: y, W: pillW, H: pillH}, step, slideID)
		y += pillH
		if i < n-1 {
			r.renderConnector(ps, pptx.Box{X: pillX, Y: y, W: pillW, H: flowGap}, v.Connector, true)
			y += flowGap
		}
	}
	if v.Connector == ConnectorCycle && n >= 2 {
		firstCy := box.Y + pillH/2
		lastCy := box.Y + pptx.EMU(n-1)*(pillH+flowGap) + pillH/2
		r.renderReturnLoopV(ps, box, pillX+pillW, firstCy, lastCy)
	}
}

// renderReturnLoopV draws a feedback loop to the right of a vertical column:
// right from the last step, up to the first, then left into it with an
// arrowhead. rightEdge is the pills' right edge; the loop lives in the band
// between it and the slot's right side.
func (r *renderer) renderReturnLoopV(ps *pptx.Slide, box pptx.Box, rightEdge, firstCy, lastCy pptx.EMU) {
	line := pptx.Line{Width: pptx.Pt(1.5), Color: pptx.TokenColor(pptx.ColorAccent)}
	accent := pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))
	midX := rightEdge + flowReturnBand/2
	head := pptx.In(0.16)
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: rightEdge, Y: lastCy, W: midX - rightEdge, H: 1}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: midX, Y: firstCy, W: 1, H: lastCy - firstCy}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: rightEdge + head, Y: firstCy, W: midX - rightEdge - head, H: 1}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	ps.AddShape(pptx.ShapeChevron, pptx.Box{X: rightEdge, Y: firstCy - head/2, W: head, H: head}, pptx.WithFill(accent), pptx.WithRotation(180))
	r.stats.Shapes += 4
}

// renderFlowStep draws one pill: a rounded rect with the icon + label + detail
// group centered vertically.
func (r *renderer) renderFlowStep(ps *pptx.Slide, box pptx.Box, step FlowStep, slideID string) {
	ps.AddShape(pptx.ShapeRoundRect, box,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))),
		pptx.WithRadius(pptx.RadiusMD))
	r.stats.Shapes++

	pad := r.theme.ResolveSpace(pptx.SpaceSM)
	innerX := box.X + pad
	innerW := box.W - 2*pad
	if innerW < 0 {
		innerW = 0
	}

	const (
		iconSz  = pptx.EMU(457200) // 0.5"
		labelH  = pptx.EMU(310896) // ~0.34"
		detailH = pptx.EMU(237744) // ~0.26"
		vgap    = pptx.EMU(54864)  // ~0.06"
	)
	hasIcon := step.Icon != ""
	hasDetail := len(step.Detail) > 0

	// Measure the content group and center it vertically within the pill.
	contentH := labelH
	if hasIcon {
		contentH += iconSz + vgap
	}
	if hasDetail {
		contentH += vgap + detailH
	}
	y := box.Y + (box.H-contentH)/2
	if min := box.Y + pad; y < min {
		y = min
	}

	if hasIcon {
		ib := pptx.Box{X: innerX + (innerW-iconSz)/2, Y: y, W: iconSz, H: iconSz}
		if svg, ok := r.cfg.icons.Lookup(step.Icon); !ok {
			r.warn(slideID, "flow step icon "+step.Icon+" not found at compose (should have failed Stage-1)")
		} else if _, err := ps.AddIcon(svg, ib); err == nil {
			r.stats.Shapes++
		}
		y += iconSz + vgap
	}

	lf := ps.AddTextFrame(pptx.Box{X: innerX, Y: y, W: innerW, H: labelH}).Anchor(pptx.AnchorMiddle)
	lp := lf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	r.addRichText(ps, lp, step.Label, pptx.TypeBody)
	r.stats.Shapes++
	y += labelH

	if hasDetail {
		y += vgap
		df := ps.AddTextFrame(pptx.Box{X: innerX, Y: y, W: innerW, H: detailH}).Anchor(pptx.AnchorMiddle)
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
		sz := pptx.In(0.24)
		ps.AddShape(pptx.ShapeGeometry("mathPlus"), centerIn(gap, sz, sz), pptx.WithFill(accent))
		r.stats.Shapes++
	case ConnectorArrowDashed:
		r.renderDashedArrow(ps, gap, vertical)
	default: // ConnectorArrow and ConnectorCycle inter-pair glyph
		if vertical {
			ps.AddShape(pptx.ShapeGeometry("downArrow"), centerIn(gap, pptx.In(0.3), gap.H*7/10), pptx.WithFill(accent))
		} else {
			ps.AddShape(pptx.ShapeRightArrow, centerIn(gap, gap.W*7/10, pptx.In(0.3)), pptx.WithFill(accent))
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

// centerIn returns a w×h box centered in parent.
func centerIn(parent pptx.Box, w, h pptx.EMU) pptx.Box {
	return pptx.Box{X: parent.X + (parent.W-w)/2, Y: parent.Y + (parent.H-h)/2, W: w, H: h}
}
