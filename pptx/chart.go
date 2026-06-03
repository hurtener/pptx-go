package pptx

// Chart V1 is image-shape (D-004): the scene chart node renders as a pic from
// caller bytes (see scene/render_chart.go). This builder helper draws a chart
// *slot* — a labeled placeholder for a chart whose bytes are not (yet) committed
// (RFC §15.1, D-046).

// ChartPlaceholder draws a labeled chart slot at box — a rounded rect with a
// dashed accent border and a centered "Chart" label — and returns the slot
// shape. It commits no chart bytes; it is the visible stand-in for a chart whose
// raster is unresolved or not yet rendered. Fills/line resolve against the
// active theme (P2); pass ShapeOptions to override the default border/fill.
func (s *Slide) ChartPlaceholder(box Box, opts ...ShapeOption) *Shape {
	base := []ShapeOption{
		WithFill(SolidFill(TokenColor(ColorSurfaceAlt))),
		WithLine(Line{Width: Pt(1), Color: TokenColor(ColorAccent), Dash: "dash"}),
		WithRadius(RadiusMD),
	}
	base = append(base, opts...)
	sh := s.AddShape(ShapeRoundRect, box, base...)

	tf := s.AddTextFrame(box).Anchor(AnchorMiddle)
	p := tf.AddParagraph(ParagraphOpts{Align: AlignCenter})
	p.AddRun("Chart", RunStyle{TypeRole: TypeCaption, Color: TokenTextColor(TextMuted)})
	return sh
}
