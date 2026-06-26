package scene

import "github.com/hurtener/pptx-go/pptx"

// DataMark composer (R14.8, D-122). Native (no-raster) micro-charts — a progress
// bar, a small bar group, or a sparkline — drawn entirely from preset rects and
// lines in theme colors. All geometry is integer EMU, so output is worker-count
// deterministic; no AssetResolver is touched.

// Pinned DataMark geometry (layout metrics, not tokens; colors are tokens).
const (
	dmBarH      = pptx.EMU(219456) // In(0.24); a horizontal progress-bar track height
	dmLabelW    = pptx.EMU(640080) // In(0.70); the inline label reserve (right of a bar)
	dmGap       = pptx.EMU(54864)  // In(0.06); inter-bar gap / bar↔label gap
	dmBarsMinW  = pptx.EMU(91440)  // In(0.10); a single bar's minimum width in a group
	dmSparkDotR = pptx.EMU(45720)  // In(0.05); the sparkline end-dot radius
)

// dataMarkPreferredHeight returns the mark's slot height: a thin slot for a single
// bar, a taller slot for a bar group / sparkline.
func dataMarkPreferredHeight(v DataMark) pptx.EMU {
	switch {
	case v.Kind == DataMarkBar && v.Orientation != FlowVertical:
		return pptx.In(0.4)
	case v.Kind == DataMarkDonut || v.Kind == DataMarkGauge:
		return pptx.In(1.6)
	default:
		return pptx.In(1.2)
	}
}

func (r *renderer) renderDataMark(ps *pptx.Slide, box pptx.Box, v DataMark) {
	switch v.Kind {
	case DataMarkBar:
		r.renderDataMarkBar(ps, box, v)
	case DataMarkBars:
		r.renderDataMarkBars(ps, box, v)
	case DataMarkSparkline:
		r.renderDataMarkSparkline(ps, box, v)
	case DataMarkDonut:
		r.renderDataMarkDonut(ps, box, v)
	case DataMarkGauge:
		r.renderDataMarkGauge(ps, box, v)
	}
}

// Pinned arc-mark geometry (R14.8 part 2).
const (
	dmInnerRatio = 0.62  // ring inner radius as a fraction of the outer
	dmDonutStart = 270.0 // 12 o'clock (OOXML 0° = 3 o'clock, clockwise)
	dmGaugeStart = 135.0 // lower-left opening of a 270° speedometer
	dmGaugeRange = 270.0 // total gauge sweep
)

// squareCentered returns the largest centered square box that fits box (so an arc
// mark is circular regardless of the slot aspect).
func squareCentered(box pptx.Box) pptx.Box {
	d := box.W
	if box.H < d {
		d = box.H
	}
	return pptx.Box{X: box.X + (box.W-d)/2, Y: box.Y + (box.H-d)/2, W: d, H: d}
}

// renderDataMarkDonut draws a single-value ring: an accent value arc + a track
// arc that together close the ring, with the Label centered in the hole.
func (r *renderer) renderDataMarkDonut(ps *pptx.Slide, box pptx.Box, v DataMark) {
	val := clampUnit01(v.Value)
	sq := squareCentered(box)
	accent := pptx.TokenColor(v.markColor())
	track := pptx.TokenColor(pptx.ColorSurfaceAlt)
	valSweep := val * 360
	// Value arc (accent), then the remainder arc (track) completing the ring.
	if valSweep > 0 {
		ps.AddBlockArc(sq, dmDonutStart, valSweep, dmInnerRatio, pptx.WithFill(pptx.SolidFill(accent)))
		r.stats.Shapes++
	}
	if val < 1 {
		ps.AddBlockArc(sq, dmDonutStart+valSweep, 360-valSweep, dmInnerRatio, pptx.WithFill(pptx.SolidFill(track)))
		r.stats.Shapes++
	}
	if v.Label != "" {
		lf := ps.AddTextFrame(sq).Anchor(pptx.AnchorMiddle)
		lp := lf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		lp.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeH2, Bold: true, Color: pptx.TokenTextColor(pptx.TextPrimary)})
		r.stats.Shapes++
	}
}

// renderDataMarkGauge draws a 270° speedometer: an accent value arc + a track arc
// over the gauge range, with the Label centered.
func (r *renderer) renderDataMarkGauge(ps *pptx.Slide, box pptx.Box, v DataMark) {
	val := clampUnit01(v.Value)
	sq := squareCentered(box)
	accent := pptx.TokenColor(v.markColor())
	track := pptx.TokenColor(pptx.ColorSurfaceAlt)
	valSweep := val * dmGaugeRange
	if valSweep > 0 {
		ps.AddBlockArc(sq, dmGaugeStart, valSweep, dmInnerRatio, pptx.WithFill(pptx.SolidFill(accent)))
		r.stats.Shapes++
	}
	if val < 1 {
		ps.AddBlockArc(sq, dmGaugeStart+valSweep, dmGaugeRange-valSweep, dmInnerRatio, pptx.WithFill(pptx.SolidFill(track)))
		r.stats.Shapes++
	}
	if v.Label != "" {
		lf := ps.AddTextFrame(sq).Anchor(pptx.AnchorMiddle)
		lp := lf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		lp.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeH3, Bold: true, Color: pptx.TokenTextColor(pptx.TextPrimary)})
		r.stats.Shapes++
	}
}

func clampUnit01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// renderDataMarkBar draws a single progress/capacity bar (a track + a fill to
// Value) with an optional inline label.
func (r *renderer) renderDataMarkBar(ps *pptx.Slide, box pptx.Box, v DataMark) {
	val := clampUnit01(v.Value)
	accent := pptx.TokenColor(v.markColor())
	track := pptx.TokenColor(pptx.ColorSurfaceAlt)

	if v.Orientation == FlowVertical {
		// Vertical bar: track full height, fill from the bottom.
		ps.AddShape(pptx.ShapeRoundRect, box, pptx.WithRadius(pptx.RadiusFull), pptx.WithFill(pptx.SolidFill(track)))
		r.stats.Shapes++
		fillH := pptx.EMU(val * float64(box.H))
		if fillH > 0 {
			fb := pptx.Box{X: box.X, Y: box.Bottom() - fillH, W: box.W, H: fillH}
			ps.AddShape(pptx.ShapeRoundRect, fb, pptx.WithRadius(pptx.RadiusFull), pptx.WithFill(pptx.SolidFill(accent)))
			r.stats.Shapes++
		}
		return
	}

	// Horizontal bar centered vertically in the box; optional label to the right.
	trackW := box.W
	if v.Label != "" {
		trackW = box.W - dmLabelW - dmGap
	}
	if trackW < 0 {
		trackW = 0
	}
	y := box.Y + (box.H-dmBarH)/2
	tb := pptx.Box{X: box.X, Y: y, W: trackW, H: dmBarH}
	ps.AddShape(pptx.ShapeRoundRect, tb, pptx.WithRadius(pptx.RadiusFull), pptx.WithFill(pptx.SolidFill(track)))
	r.stats.Shapes++
	fillW := pptx.EMU(val * float64(trackW))
	if fillW > 0 {
		fb := pptx.Box{X: box.X, Y: y, W: fillW, H: dmBarH}
		ps.AddShape(pptx.ShapeRoundRect, fb, pptx.WithRadius(pptx.RadiusFull), pptx.WithFill(pptx.SolidFill(accent)))
		r.stats.Shapes++
	}
	if v.Label != "" {
		lf := ps.AddTextFrame(pptx.Box{X: box.Right() - dmLabelW, Y: box.Y, W: dmLabelW, H: box.H}).Anchor(pptx.AnchorMiddle)
		lp := lf.AddParagraph(pptx.ParagraphOpts{})
		lp.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: pptx.TokenTextColor(pptx.TextPrimary)})
		r.stats.Shapes++
	}
}

// renderDataMarkBars draws a small vertical bar group, one bar per value.
func (r *renderer) renderDataMarkBars(ps *pptx.Slide, box pptx.Box, v DataMark) {
	n := len(v.Values)
	if n == 0 {
		return
	}
	accent := pptx.TokenColor(v.markColor())
	// Total gap = (n-1)*dmGap; each bar gets an equal share of the remainder.
	totalGap := dmGap * pptx.EMU(n-1)
	barW := (box.W - totalGap) / pptx.EMU(n)
	if barW < dmBarsMinW {
		barW = dmBarsMinW
	}
	for i, raw := range v.Values {
		val := clampUnit01(raw)
		x := box.X + (barW+dmGap)*pptx.EMU(i)
		h := pptx.EMU(val * float64(box.H))
		if h <= 0 {
			continue
		}
		b := pptx.Box{X: x, Y: box.Bottom() - h, W: barW, H: h}
		ps.AddShape(pptx.ShapeRoundRect, b, pptx.WithRadius(pptx.RadiusSM), pptx.WithFill(pptx.SolidFill(accent)))
		r.stats.Shapes++
	}
}

// renderDataMarkSparkline draws a trend polyline through the values as connected
// straight segments (a stroke, so line segments — not a filled custGeom).
func (r *renderer) renderDataMarkSparkline(ps *pptx.Slide, box pptx.Box, v DataMark) {
	n := len(v.Values)
	accent := pptx.TokenColor(v.markColor())
	line := pptx.Line{Width: pptx.Pt(2), Color: accent}
	if n == 1 {
		// A single point: just the end dot.
		r.sparkDot(ps, box.X, box.Bottom()-pptx.EMU(clampUnit01(v.Values[0])*float64(box.H)), accent)
		return
	}
	pt := func(i int) (pptx.EMU, pptx.EMU) {
		x := box.X + pptx.EMU(int64(box.W)*int64(i)/int64(n-1))
		y := box.Bottom() - pptx.EMU(clampUnit01(v.Values[i])*float64(box.H))
		return x, y
	}
	for i := 0; i < n-1; i++ {
		x0, y0 := pt(i)
		x1, y1 := pt(i + 1)
		// A line draws TL→BR of its (positive-extent) box; an upward segment
		// (y1 < y0) uses a top-anchored box + WithFlipV so it draws BL→TR.
		if y1 >= y0 {
			ps.AddShape(pptx.ShapeLine, pptx.Box{X: x0, Y: y0, W: x1 - x0, H: y1 - y0},
				pptx.WithFill(pptx.NoFill()), pptx.WithLine(line))
		} else {
			ps.AddShape(pptx.ShapeLine, pptx.Box{X: x0, Y: y1, W: x1 - x0, H: y0 - y1},
				pptx.WithFill(pptx.NoFill()), pptx.WithLine(line), pptx.WithFlipV(true))
		}
		r.stats.Shapes++
	}
	// Accent dot at the last point.
	lx, ly := pt(n - 1)
	r.sparkDot(ps, lx, ly, accent)
}

func (r *renderer) sparkDot(ps *pptx.Slide, cx, cy pptx.EMU, color pptx.Color) {
	ps.AddShape(pptx.ShapeEllipse, pptx.Box{X: cx - dmSparkDotR, Y: cy - dmSparkDotR, W: 2 * dmSparkDotR, H: 2 * dmSparkDotR},
		pptx.WithFill(pptx.SolidFill(color)))
	r.stats.Shapes++
}
