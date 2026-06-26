package scene

import (
	"math"

	"github.com/hurtener/pptx-go/pptx"
)

// Funnel + Cycle composers (R14.11, D-128). Funnel = a vertical stack of bands
// tapering in width (a stepped funnel) with optional per-stage values. Cycle =
// stages placed evenly on a ring with directional connectors. Both are native
// preset shapes; pure integer-EMU → byte-identical / worker-count deterministic.

const (
	funnelGap     = pptx.EMU(45720)   // In(0.05); vertical gap between bands
	funnelMinFrac = 35                // narrowest band as a percent of the widest
	cycleNodeW    = pptx.EMU(1280160) // In(1.40); a ring stage card width
	cycleNodeH    = pptx.EMU(548640)  // In(0.60); a ring stage card height
	cycleIconSz   = pptx.EMU(228600)  // In(0.25)
)

// renderFunnel draws a stepped funnel: N centered bands of linearly decreasing
// width (widest at top), each with a centered label + optional value.
func (r *renderer) renderFunnel(ps *pptx.Slide, box pptx.Box, v Funnel) {
	n := len(v.Stages)
	if n == 0 || box.W <= 0 || box.H <= 0 {
		return
	}
	bandH := (box.H - funnelGap*pptx.EMU(n-1)) / pptx.EMU(n)
	if bandH <= 0 {
		bandH = box.H / pptx.EMU(n)
	}
	for i, st := range v.Stages {
		// Width tapers from 100% (top) to funnelMinFrac% (bottom), linearly.
		frac := 100
		if n > 1 {
			frac = 100 - (100-funnelMinFrac)*i/(n-1)
		}
		w := box.W * pptx.EMU(frac) / 100
		x := box.X + (box.W-w)/2
		y := box.Y + (bandH+funnelGap)*pptx.EMU(i)
		accent := r.accentColorAt(st.AccentIndex)
		bandBox := pptx.Box{X: x, Y: y, W: w, H: bandH}
		ps.AddShape(pptx.ShapeRoundRect, bandBox, pptx.WithRadius(pptx.RadiusSM), pptx.WithFill(pptx.SolidFill(accent)))
		r.stats.Shapes++
		// Centered label (+ value) in contrast text.
		tc := r.cellTextOnColor(r.accentRGBAt(st.AccentIndex))
		tf := ps.AddTextFrame(bandBox).Anchor(pptx.AnchorMiddle)
		lp := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		lp.AddRun(st.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: tc})
		if st.Value != "" {
			vp := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
			vp.AddRun(st.Value, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: tc})
		}
		r.stats.Shapes++
	}
}

// renderCycle places stages evenly on a ring (clockwise from the top) and draws a
// directional connector between consecutive stages (last → first), with a chevron
// arrowhead rotated to the chord.
func (r *renderer) renderCycle(ps *pptx.Slide, box pptx.Box, v Cycle, slideID string) {
	n := len(v.Stages)
	if n == 0 || box.W <= 0 || box.H <= 0 {
		return
	}
	sq := squareCentered(box)
	cx := sq.X + sq.W/2
	cy := sq.Y + sq.H/2
	// Ring radius leaves room for the stage cards.
	radius := (sq.W - cycleNodeW) / 2
	if h := (sq.H - cycleNodeH) / 2; h < radius {
		radius = h
	}
	if radius <= 0 {
		return
	}
	centers := make([]pptx.Position, n)
	for i := range v.Stages {
		// Clockwise from the top: angle = -90° + i*360/n.
		ang := -math.Pi/2 + 2*math.Pi*float64(i)/float64(n)
		centers[i] = pptx.Position{
			X: cx + pptx.EMU(math.Round(math.Cos(ang)*float64(radius))),
			Y: cy + pptx.EMU(math.Round(math.Sin(ang)*float64(radius))),
		}
	}
	// Connectors between consecutive stages (drawn behind the cards).
	line := pptx.Line{Width: pptx.Pt(1.5), Color: pptx.TokenColor(pptx.ColorSurfaceAlt)}
	for i := 0; i < n; i++ {
		from := centers[i]
		to := centers[(i+1)%n]
		if n == 1 {
			break
		}
		r.cycleConnector(ps, from, to, line)
	}
	// Stage cards.
	for i, st := range v.Stages {
		c := centers[i]
		nb := pptx.Box{X: c.X - cycleNodeW/2, Y: c.Y - cycleNodeH/2, W: cycleNodeW, H: cycleNodeH}
		accent := r.accentColorAt(st.AccentIndex)
		ps.AddShape(pptx.ShapeRoundRect, nb,
			pptx.WithRadius(pptx.RadiusMD),
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurface))),
			pptx.WithLine(pptx.Line{Width: pptx.Pt(1.25), Color: accent}))
		r.stats.Shapes++
		tx, tw := nb.X, nb.W
		if st.Icon != "" {
			ib := pptx.Box{X: nb.X + funnelGap, Y: nb.Y + (nb.H-cycleIconSz)/2, W: cycleIconSz, H: cycleIconSz}
			if svg, ok := r.cfg.icons.Lookup(st.Icon); ok {
				if _, err := ps.AddIcon(svg, ib, pptx.WithFill(pptx.SolidFill(accent))); err == nil {
					r.stats.Shapes++
				}
			} else {
				r.warn(slideID, "cycle stage icon "+st.Icon+" not found at compose")
			}
			tx += cycleIconSz + funnelGap
			tw -= cycleIconSz + funnelGap
		}
		tf := ps.AddTextFrame(pptx.Box{X: tx, Y: nb.Y, W: tw, H: nb.H}).Anchor(pptx.AnchorMiddle)
		lp := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		lp.AddRun(st.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: pptx.TokenTextColor(pptx.TextPrimary)})
		r.stats.Shapes++
	}
}

// cycleConnector draws a straight line between two ring points (flip-aware for an
// upward run) plus a chevron arrowhead at the target, rotated to the chord.
func (r *renderer) cycleConnector(ps *pptx.Slide, from, to pptx.Position, line pptx.Line) {
	dx := to.X - from.X
	dy := to.Y - from.Y
	if dy >= 0 {
		ps.AddShape(pptx.ShapeLine, pptx.Box{X: from.X, Y: from.Y, W: dx, H: dy}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
	} else {
		// Upward run: top-anchored positive box + flipV (draws BL→TR).
		bx := from.X
		bw := dx
		if dx < 0 {
			bx = to.X
			bw = -dx
		}
		ps.AddShape(pptx.ShapeLine, pptx.Box{X: bx, Y: to.Y, W: bw, H: -dy}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(line), pptx.WithFlipV(true))
	}
	r.stats.Shapes++
	// Chevron arrowhead at the midpoint, rotated to the chord direction.
	midX := (from.X + to.X) / 2
	midY := (from.Y + to.Y) / 2
	deg := math.Atan2(float64(dy), float64(dx)) * 180 / math.Pi
	head := pptx.EMU(137160) // In(0.15)
	ps.AddShape(pptx.ShapeChevron, pptx.Box{X: midX - head/2, Y: midY - head/2, W: head, H: head},
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))), pptx.WithRotation(deg))
	r.stats.Shapes++
}
