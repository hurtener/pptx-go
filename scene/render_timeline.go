package scene

import "github.com/hurtener/pptx-go/pptx"

// Timeline composer (R14.4, D-119). A horizontal axis with milestones at
// caller-specified proportional positions, optional phase bands behind the axis,
// and optional swimlanes (rows). Markers, the axis line, and labels compose from
// native preset shapes (no media). Labels stagger above / below the axis to avoid
// collision. All geometry is integer EMU, so output is worker-count deterministic.

// Pinned timeline geometry (not theme tokens — layout metrics; colors are tokens).
const (
	tlBandLabelH = pptx.EMU(274320)  // In(0.30); the band-label strip at the top
	tlLaneLabelW = pptx.EMU(1097280) // In(1.20); the left gutter for lane labels
	tlMarkerR    = pptx.EMU(64008)   // In(0.07); milestone marker dot radius
	tlIconSz     = pptx.EMU(228600)  // In(0.25); milestone icon box side
	tlLabelW     = pptx.EMU(1463040) // In(1.60); a milestone label box width
	tlLabelH     = pptx.EMU(228600)  // In(0.25); a milestone label line height
	tlDetailH    = pptx.EMU(201168)  // In(0.22); a milestone detail line height
	tlGap        = pptx.EMU(45720)   // In(0.05); marker↔label gap
	tlBandAlpha  = 14000             // band fill OOXML opacity (subtle)
	tlLaneMinH   = pptx.EMU(1280160) // In(1.40); a lane's minimum height estimate
)

// timelinePreferredHeight returns the timeline's slot height: a band-label strip
// (if any band is labeled) plus one pinned lane height per lane (>=1).
func timelinePreferredHeight(v Timeline) pptx.EMU {
	lanes := len(v.Lanes)
	if lanes < 1 {
		lanes = 1
	}
	h := tlLaneMinH * pptx.EMU(lanes)
	if timelineHasBandLabel(v) {
		h += tlBandLabelH
	}
	return h
}

func timelineHasBandLabel(v Timeline) bool {
	for _, b := range v.Bands {
		if b.Label != "" {
			return true
		}
	}
	return false
}

// timelineAccent maps a milestone's AccentIndex to a token color role, cycling a
// pinned set so per-phase markers read distinctly; the soul drives which index
// each milestone uses (D-026). All entries are existing theme roles (P2). It is
// the fallback cycle the accent resolvers use when the theme defines no brand
// accent palette (R8.4).
func timelineAccent(idx int) pptx.ColorRole {
	roles := []pptx.ColorRole{pptx.ColorAccent, pptx.ColorAccentAlt, pptx.ColorInfo, pptx.ColorSuccess, pptx.ColorWarning}
	if idx < 0 {
		idx = 0
	}
	return roles[idx%len(roles)]
}

// accentColorAt returns the fill/stroke Color for an element's accent index
// (R8.4). When the active theme defines a brand-accent palette (Theme.Accents),
// the index cycles those literal hues; otherwise it falls back to the pinned
// five-role cycle resolved as a token — byte-identical to the pre-R8.4 output.
// The soul drives which index each element uses (D-026).
func (r *renderer) accentColorAt(idx int) pptx.Color {
	if n := len(r.theme.Accents); n > 0 {
		if idx < 0 {
			idx = 0
		}
		return r.theme.Accents[idx%n]
	}
	return pptx.TokenColor(timelineAccent(idx))
}

// accentRGBAt returns the resolved RGB for an element's accent index (R8.4), for
// computing auto-contrast text on an accent fill. It mirrors accentColorAt: the
// brand palette when defined, else the pinned role cycle resolved against the
// active (possibly dark-variant) theme — byte-identical when no palette is set.
func (r *renderer) accentRGBAt(idx int) pptx.RGB {
	if n := len(r.theme.Accents); n > 0 {
		if idx < 0 {
			idx = 0
		}
		return r.theme.Accents[idx%n]
	}
	return r.theme.ResolveColor(timelineAccent(idx))
}

func (r *renderer) renderTimeline(ps *pptx.Slide, box pptx.Box, v Timeline, slideID string) {
	// Reserve a top strip for band labels and a left gutter for lane labels.
	bandLabelH := pptx.EMU(0)
	if timelineHasBandLabel(v) {
		bandLabelH = tlBandLabelH
	}
	laneLabelW := pptx.EMU(0)
	hasLaneLabel := false
	for _, ln := range v.Lanes {
		if ln.Label != "" {
			hasLaneLabel = true
			break
		}
	}
	if hasLaneLabel {
		laneLabelW = tlLaneLabelW
	}

	axis := pptx.Box{X: box.X + laneLabelW, Y: box.Y + bandLabelH, W: box.W - laneLabelW, H: box.H - bandLabelH}
	if axis.W <= 0 || axis.H <= 0 {
		return
	}

	// Bands behind everything: a low-alpha region spanning [From,To] of the axis
	// width across the full lane area, labeled at the top.
	for _, b := range v.Bands {
		bx := axis.X + pptx.EMU(b.From*float64(axis.W))
		bw := pptx.EMU((b.To - b.From) * float64(axis.W))
		if bw <= 0 {
			continue
		}
		fill := b.Fill
		ps.AddShape(pptx.ShapeRect, pptx.Box{X: bx, Y: axis.Y, W: bw, H: axis.H},
			pptx.WithFill(pptx.SolidFill(pptx.TokenColorAlpha(fill, tlBandAlpha))))
		r.stats.Shapes++
		if b.Label != "" {
			lf := ps.AddTextFrame(pptx.Box{X: bx, Y: box.Y, W: bw, H: bandLabelH}).Anchor(pptx.AnchorMiddle)
			lp := lf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
			lp.AddRun(b.Label, pptx.RunStyle{TypeRole: pptx.TypeCaption, Bold: true, Color: pptx.TokenTextColor(pptx.TextSecondary)})
			r.stats.Shapes++
		}
	}

	// Lanes: explicit swimlanes, or one implicit lane from the top-level milestones.
	lanes := v.Lanes
	if len(lanes) == 0 {
		lanes = []TimelineLane{{Milestones: v.Milestones}}
	}
	laneH := axis.H / pptx.EMU(len(lanes))
	for i, ln := range lanes {
		laneBox := pptx.Box{X: axis.X, Y: axis.Y + laneH*pptx.EMU(i), W: axis.W, H: laneH}
		if ln.Label != "" {
			lf := ps.AddTextFrame(pptx.Box{X: box.X, Y: laneBox.Y, W: laneLabelW, H: laneBox.H}).Anchor(pptx.AnchorMiddle)
			lp := lf.AddParagraph(pptx.ParagraphOpts{})
			lp.AddRun(ln.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: pptx.TokenTextColor(pptx.TextSecondary)})
			r.stats.Shapes++
		}
		r.renderTimelineLane(ps, laneBox, ln.Milestones, slideID)
	}
	_ = slideID
}

// renderTimelineLane draws one lane's axis line + markers + staggered labels.
func (r *renderer) renderTimelineLane(ps *pptx.Slide, lane pptx.Box, ms []Milestone, slideID string) {
	axisY := lane.Y + lane.H/2
	// Axis line.
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: lane.X, Y: axisY, W: lane.W, H: 1},
		pptx.WithFill(pptx.NoFill()),
		pptx.WithLine(pptx.Line{Width: pptx.Pt(1.5), Color: pptx.TokenColor(pptx.ColorSurfaceAlt)}))
	r.stats.Shapes++

	for i, m := range ms {
		pos := m.Position
		if pos < 0 {
			pos = 0
		} else if pos > 1 {
			pos = 1
		}
		cx := lane.X + pptx.EMU(pos*float64(lane.W))
		accent := r.accentColorAt(m.AccentIndex)

		// Marker: a curated icon when set, else an accent dot.
		if m.Icon != "" {
			ib := pptx.Box{X: cx - tlIconSz/2, Y: axisY - tlIconSz/2, W: tlIconSz, H: tlIconSz}
			if svg, ok := r.cfg.icons.Lookup(m.Icon); !ok {
				r.warn(slideID, "timeline milestone icon "+m.Icon+" not found at compose (should have failed Stage-1)")
			} else if _, err := ps.AddIcon(svg, ib, pptx.WithFill(pptx.SolidFill(accent))); err == nil {
				r.stats.Shapes++
			}
		} else {
			ps.AddShape(pptx.ShapeEllipse, pptx.Box{X: cx - tlMarkerR, Y: axisY - tlMarkerR, W: 2 * tlMarkerR, H: 2 * tlMarkerR},
				pptx.WithFill(pptx.SolidFill(accent)))
			r.stats.Shapes++
		}

		// Label staggered above (even index) / below (odd index) the axis.
		above := i%2 == 0
		lx := cx - tlLabelW/2
		if lx < lane.X {
			lx = lane.X
		}
		if lx+tlLabelW > lane.Right() {
			lx = lane.Right() - tlLabelW
		}
		blockH := tlLabelH
		if m.Detail != "" {
			blockH += tlDetailH
		}
		var ly pptx.EMU
		if above {
			ly = axisY - tlMarkerR - tlGap - blockH
			if ly < lane.Y {
				ly = lane.Y
			}
		} else {
			ly = axisY + tlMarkerR + tlGap
		}
		lf := ps.AddTextFrame(pptx.Box{X: lx, Y: ly, W: tlLabelW, H: tlLabelH}).Anchor(pptx.AnchorMiddle)
		lp := lf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		lp.AddRun(m.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: pptx.TokenTextColor(pptx.TextPrimary)})
		r.stats.Shapes++
		if m.Detail != "" {
			df := ps.AddTextFrame(pptx.Box{X: lx, Y: ly + tlLabelH, W: tlLabelW, H: tlDetailH}).Anchor(pptx.AnchorMiddle)
			dp := df.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
			dp.AddRun(m.Detail, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
			r.stats.Shapes++
		}
	}
}
