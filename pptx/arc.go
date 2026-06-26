package pptx

import (
	"fmt"
	"math"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// AddBlockArc adds a block-arc — an annular (ring) sector — positioned by box, a
// native preset geometry (<a:prstGeom prst="blockArc">). It sweeps clockwise from
// startDeg by sweepDeg degrees (OOXML angle convention: 0° at 3 o'clock,
// increasing clockwise; use 270° for 12 o'clock) at the given innerRatio (the
// ring's inner radius as a fraction of the outer, 0..1 — e.g. 0.6 = a thin ring).
// It is the native primitive behind donut/ring and gauge micro-charts (R14.8).
// Fills/lines resolve against the active theme like AddShape (P2).
func (s *Slide) AddBlockArc(box Box, startDeg, sweepDeg, innerRatio float64, opts ...ShapeOption) *Shape {
	sh := s.AddShape(ShapeGeometry("blockArc"), box, opts...)
	spPr := sh.props()
	if spPr == nil || spPr.PresetGeom == nil {
		return sh
	}
	if innerRatio < 0 {
		innerRatio = 0
	} else if innerRatio > 1 {
		innerRatio = 1
	}
	// blockArc adjust handles: adj1 = start angle, adj2 = end angle (60000ths of a
	// degree), adj3 = inner radius (×100000). The sector fills from adj1 to adj2.
	spPr.PresetGeom.AvLst = &slide.XAvLst{Gd: []slide.XShapeGuide{
		{Name: "adj1", Fmla: fmt.Sprintf("val %d", angle60k(startDeg))},
		{Name: "adj2", Fmla: fmt.Sprintf("val %d", angle60k(startDeg+sweepDeg))},
		{Name: "adj3", Fmla: fmt.Sprintf("val %d", int(math.Round(innerRatio*100000)))},
	}}
	return sh
}

// angle60k normalizes deg to [0, 360) and converts to 60000ths of a degree (the
// OOXML positive-angle unit), deterministically (integer result).
func angle60k(deg float64) int {
	d := math.Mod(deg, 360)
	if d < 0 {
		d += 360
	}
	return int(math.Round(d * 60000))
}
