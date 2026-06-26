package scene

import "github.com/hurtener/pptx-go/pptx"

// Quadrant composer (R14.9, D-124). A 2x2 positioning map: labeled X/Y axes,
// optional per-quadrant tints + titles, and items plotted at (x,y) in [0,1]
// (origin bottom-left). Axes, dividers, dots, and labels are native shapes; all
// geometry is integer EMU (worker-count deterministic).

const (
	qAxisGutterW = pptx.EMU(731520)  // In(0.80); left gutter for Y-axis end labels
	qAxisStripH  = pptx.EMU(274320)  // In(0.30); bottom strip for X-axis end labels
	qDotR        = pptx.EMU(64008)   // In(0.07); a plotted item's dot radius
	qTintAlpha   = 10000             // per-quadrant tint OOXML opacity (subtle)
	qItemLabelW  = pptx.EMU(1188720) // In(1.30); an item label box width
	qItemLabelH  = pptx.EMU(201168)  // In(0.22); an item label line height
)

func (r *renderer) renderQuadrant(ps *pptx.Slide, box pptx.Box, v Quadrant) {
	// Field = box minus the left (Y labels) gutter and the bottom (X labels) strip.
	field := pptx.Box{X: box.X + qAxisGutterW, Y: box.Y, W: box.W - qAxisGutterW, H: box.H - qAxisStripH}
	if field.W <= 0 || field.H <= 0 {
		return
	}
	midX := field.X + field.W/2
	midY := field.Y + field.H/2
	halfW := field.W / 2
	halfH := field.H / 2

	// Per-quadrant tints (0=TL,1=TR,2=BL,3=BR), drawn first (behind).
	cells := [4]pptx.Box{
		{X: field.X, Y: field.Y, W: halfW, H: halfH},               // TL
		{X: midX, Y: field.Y, W: field.W - halfW, H: halfH},        // TR
		{X: field.X, Y: midY, W: halfW, H: field.H - halfH},        // BL
		{X: midX, Y: midY, W: field.W - halfW, H: field.H - halfH}, // BR
	}
	for i, qc := range v.Quadrants {
		if qc.Fill != nil {
			ps.AddShape(pptx.ShapeRect, cells[i], pptx.WithFill(pptx.SolidFill(pptx.TokenColorAlpha(*qc.Fill, qTintAlpha))))
			r.stats.Shapes++
		}
		if qc.Title != "" {
			tf := ps.AddTextFrame(pptx.Box{X: cells[i].X, Y: cells[i].Y, W: cells[i].W, H: qItemLabelH}).Anchor(pptx.AnchorMiddle)
			tp := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
			tp.AddRun(qc.Title, pptx.RunStyle{TypeRole: pptx.TypeCaption, Bold: true, Color: pptx.TokenTextColor(pptx.TextMuted)})
			r.stats.Shapes++
		}
	}

	// Axis dividers (center cross).
	div := pptx.Line{Width: pptx.Pt(1.25), Color: pptx.TokenColor(pptx.ColorSurfaceAlt)}
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: midX, Y: field.Y, W: 1, H: field.H}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(div))
	ps.AddShape(pptx.ShapeLine, pptx.Box{X: field.X, Y: midY, W: field.W, H: 1}, pptx.WithFill(pptx.NoFill()), pptx.WithLine(div))
	r.stats.Shapes += 2

	// Plotted items: a dot at (x,y) (origin bottom-left → invert Y) + a label.
	for i, it := range v.Items {
		x := field.X + pptx.EMU(clampUnit01(it.X)*float64(field.W))
		y := field.Bottom() - pptx.EMU(clampUnit01(it.Y)*float64(field.H))
		accent := timelineAccent(it.AccentIndex)
		ps.AddShape(pptx.ShapeEllipse, pptx.Box{X: x - qDotR, Y: y - qDotR, W: 2 * qDotR, H: 2 * qDotR},
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(accent))))
		r.stats.Shapes++
		if it.Label == "" {
			continue
		}
		lx := x + qDotR + pptx.EMU(27432) // In(0.03) gap
		if lx+qItemLabelW > field.Right() {
			lx = x - qDotR - qItemLabelW // flip the label to the left of the dot
		}
		if lx < field.X {
			lx = field.X
		}
		ly := y - qItemLabelH/2
		lf := ps.AddTextFrame(pptx.Box{X: lx, Y: ly, W: qItemLabelW, H: qItemLabelH}).Anchor(pptx.AnchorMiddle)
		lp := lf.AddParagraph(pptx.ParagraphOpts{})
		lp.AddRun(it.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Color: pptx.TokenTextColor(pptx.TextPrimary)})
		r.stats.Shapes++
		_ = i
	}

	// Axis end labels: X low/high on the bottom strip; Y low/high on the gutter.
	r.quadrantAxisLabel(ps, pptx.Box{X: field.X, Y: field.Bottom(), W: field.W / 2, H: qAxisStripH}, v.AxisX.LowLabel, pptx.AlignLeft)
	r.quadrantAxisLabel(ps, pptx.Box{X: midX, Y: field.Bottom(), W: field.W / 2, H: qAxisStripH}, v.AxisX.HighLabel, pptx.AlignRight)
	r.quadrantAxisLabel(ps, pptx.Box{X: box.X, Y: midY, W: qAxisGutterW, H: halfH}, v.AxisY.HighLabel, pptx.AlignCenter)
	r.quadrantAxisLabel(ps, pptx.Box{X: box.X, Y: field.Y, W: qAxisGutterW, H: halfH}, v.AxisY.LowLabel, pptx.AlignCenter)
}

func (r *renderer) quadrantAxisLabel(ps *pptx.Slide, box pptx.Box, text string, align pptx.Alignment) {
	if text == "" {
		return
	}
	tf := ps.AddTextFrame(box).Anchor(pptx.AnchorMiddle)
	p := tf.AddParagraph(pptx.ParagraphOpts{Align: align})
	p.AddRun(text, pptx.RunStyle{TypeRole: pptx.TypeCaption, Bold: true, Color: pptx.TokenTextColor(pptx.TextSecondary)})
	r.stats.Shapes++
}
