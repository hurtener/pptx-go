package scene

import (
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/layout"
)

// Container composers (RFC §11.2 / §12). A container emits no shape of its own:
// it subdivides its slot via the scene/layout geometry engine and renders each
// child into a sub-slot through the normal dispatch, so nesting composes.

func (r *renderer) renderTwoColumn(ps *pptx.Slide, box pptx.Box, v TwoColumn, slideID string) {
	gap := r.theme.ResolveSpace(pptx.SpaceMD)

	// A top/bottom bridge (R12.8) reserves a band at that edge so the bracket spans above
	// (or below) the columns without overlapping their content; the columns lay out in the
	// remaining region. JoinSeam (default) reserves nothing — byte-identical (D-101).
	colBox := box
	bridging := v.Join != JoinNone && v.JoinPosition != JoinSeam
	if bridging {
		switch v.JoinPosition {
		case JoinTopBridge:
			colBox.Y += bridgeBandH
			colBox.H -= bridgeBandH
		case JoinBottomBridge:
			colBox.H -= bridgeBandH
		}
	}

	cols := layout.Columns(colBox, ratioWeights(v.Ratio), gap)
	if len(cols) != 2 {
		return
	}
	for _, pl := range r.stackIn(cols[0], v.Left, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
	for _, pl := range r.stackIn(cols[1], v.Right, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
	// Inter-column element (D-055 / D-101), drawn after the column content so it sits on
	// top of both columns. Inert when Join == JoinNone.
	if v.Join != JoinNone {
		if bridging {
			r.renderColumnBridge(ps, box, cols[0], cols[1], v)
		} else {
			r.renderColumnJoin(ps, box, cols[0], cols[1], v)
		}
	}
}

// Column-bridge geometry (R12.8, D-101). Pinned layout metrics — the reserved band, the
// stub length, the bracket stroke, and the label-pill padding.
const (
	bridgeBandH    = pptx.EMU(457200) // In(0.50); the reserved top/bottom band
	bridgeStubLen  = pptx.EMU(146304) // In(0.16); the down/up stubs at each end
	bridgeStroke   = pptx.EMU(19050)  // Pt(1.5); the bracket line thickness
	bridgePillPadX = pptx.EMU(91440)  // In(0.10); the label-pill horizontal padding
	bridgePillH    = pptx.EMU(274320) // In(0.30); the label-pill height
)

// renderColumnBridge draws a horizontal accent bracket (two end stubs + a spanning line)
// across the top or bottom edge of both columns, with the JoinLabel as a content-fit
// centered pill on the line — the "one X, two ways" header (R12.8, D-101). The label
// never wraps mid-word (the pill grows to the text, shrinking via FontScale only if it
// would exceed the span). Deterministic integer-EMU; accent-token colors.
func (r *renderer) renderColumnBridge(ps *pptx.Slide, box, left, right pptx.Box, v TwoColumn) {
	accent := pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))
	bracketLeft := left.X
	bracketRight := right.X + right.W
	span := bracketRight - bracketLeft
	if span <= 0 {
		return
	}

	// The bracket line sits in the reserved band; the stubs reach toward the columns.
	var lineY, stubY pptx.EMU
	if v.JoinPosition == JoinBottomBridge {
		lineY = box.Y + box.H - bridgeBandH/2
		stubY = lineY // stubs go down toward... actually up to the columns above
	} else {
		lineY = box.Y + bridgeBandH/2
		stubY = lineY
	}

	// Horizontal spanning line.
	ps.AddShape(pptx.ShapeRect, pptx.Box{X: bracketLeft, Y: lineY - bridgeStroke/2, W: span, H: bridgeStroke}, pptx.WithFill(accent))
	r.stats.Shapes++
	// End stubs: from the line toward the columns (down for a top bridge, up for a bottom).
	stubTopY := stubY
	if v.JoinPosition == JoinBottomBridge {
		stubTopY = stubY - bridgeStubLen
	}
	for _, sx := range []pptx.EMU{bracketLeft, bracketRight - bridgeStroke} {
		ps.AddShape(pptx.ShapeRect, pptx.Box{X: sx, Y: stubTopY, W: bridgeStroke, H: bridgeStubLen}, pptx.WithFill(accent))
		r.stats.Shapes++
	}

	// Content-fit label pill centered on the line.
	if v.JoinLabel != "" {
		natW := naturalWidth(RichText{{Text: v.JoinLabel, Style: RunStyle{TypeRole: pptx.TypeBodySmall}}}, r.theme)
		pillW := natW + 2*bridgePillPadX
		if pillW > span {
			pillW = span
		}
		pillBox := pptx.Box{X: bracketLeft + (span-pillW)/2, Y: lineY - bridgePillH/2, W: pillW, H: bridgePillH}
		ps.AddShape(pptx.ShapeRoundRect, pillBox, pptx.WithFill(accent), pptx.WithRadius(pptx.RadiusFull))
		r.stats.Shapes++
		jc := r.onCardSurface(pptx.ColorAccent)
		if jc == nil {
			jc = pptx.TokenTextColor(pptx.TextPrimary)
		}
		tf := ps.AddTextFrame(pillBox).Anchor(pptx.AnchorMiddle)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.JoinLabel, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: jc, FontScale: fitScale(natW, pillW-2*bridgePillPadX)})
		r.stats.Shapes++
	}
}

// Column-join geometry (D-055). Pinned EMU literals so output is worker-count
// independent (RFC §10.1).
const (
	joinBadgeSz = pptx.EMU(566928) // In(0.62); base badge diameter
	joinArrowW  = pptx.EMU(457200) // In(0.50); connector arrow width
	joinArrowH  = pptx.EMU(274320) // In(0.30); connector arrow height
)

// Join-badge fit-to-label geometry (R11.7, D-087). The badge diameter grows to
// contain its label (up to a cap); a label too long even at the cap is shrunk to
// one line. Pinned layout metrics, not tokens.
const (
	joinBadgePadX  = pptx.EMU(109728)  // In(0.12); horizontal padding inside the badge
	joinBadgeMaxSz = pptx.EMU(1371600) // In(1.50); cap on the auto-grown diameter
)

// renderColumnJoin draws the TwoColumn seam element centered on the boundary
// between the left and right columns: a "VS"-style accent badge (ellipse +
// centered inverse label) or an accent right-arrow connector.
func (r *renderer) renderColumnJoin(ps *pptx.Slide, box, left, right pptx.Box, v TwoColumn) {
	seamX := (left.X + left.W + right.X) / 2 // midpoint of the inter-column gap
	centerY := box.Y + box.H/2
	switch v.Join {
	case JoinBadge:
		// Fit-to-label (R11.7): grow the badge diameter to contain the label (up to
		// joinBadgeMaxSz), so a multi-word label like "One agent" is not broken
		// mid-word inside the fixed In(0.62) ellipse. A label that still does not fit
		// at the cap is shrunk to one line via FontScale. A short label (e.g. "vs")
		// keeps the base diameter and full size → byte-identical.
		badgeSz := joinBadgeSz
		var labelScale float64
		if v.JoinLabel != "" {
			natW := naturalWidth(RichText{{Text: v.JoinLabel, Style: RunStyle{TypeRole: pptx.TypeBodySmall}}}, r.theme)
			if needed := natW + 2*joinBadgePadX; needed > badgeSz {
				badgeSz = needed
				if badgeSz > joinBadgeMaxSz {
					badgeSz = joinBadgeMaxSz
				}
			}
			labelScale = fitScale(natW, badgeSz-2*joinBadgePadX)
		}
		bb := pptx.Box{X: seamX - badgeSz/2, Y: centerY - badgeSz/2, W: badgeSz, H: badgeSz}
		ps.AddShape(pptx.ShapeEllipse, bb, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
		r.stats.Shapes++
		if v.JoinLabel != "" {
			// Auto-contrast against the badge's accent fill (R11.2, D-082): a dark
			// accent yields the light TextInverse token (byte-identical to the prior
			// hardcoded inverse), a light-accent theme yields a dark text token.
			jc := r.onCardSurface(pptx.ColorAccent)
			if jc == nil {
				jc = pptx.TokenTextColor(pptx.TextPrimary)
			}
			tf := ps.AddTextFrame(bb).Anchor(pptx.AnchorMiddle)
			p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
			p.AddRun(v.JoinLabel, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Bold: true, Color: jc, FontScale: labelScale})
			r.stats.Shapes++
		}
	case JoinArrow:
		ab := pptx.Box{X: seamX - joinArrowW/2, Y: centerY - joinArrowH/2, W: joinArrowW, H: joinArrowH}
		ps.AddShape(pptx.ShapeRightArrow, ab, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
		r.stats.Shapes++
	}
}

func (r *renderer) renderGrid(ps *pptx.Slide, box pptx.Box, v Grid, slideID string) {
	// box is already clamped to the safe area by renderNode (R11.3/R11.12) before it
	// is subdivided, so cells never tile off the slide bottom.
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	cells := layout.Grid(box, v.Columns, v.Ratio, gap, len(v.Cells))
	for i, n := range v.Cells {
		if i >= len(cells) {
			break
		}
		r.renderNode(ps, cells[i], n, slideID, HAlignLeft)
	}
	// Inter-column connectors (R12.4, D-099): a glyph centered in the gutter between
	// two adjacent columns' first-row cells, spanning the grid height. Empty = no draw
	// (byte-identical). The gutter is derived from the cell boxes, so it scales with
	// the column layout deterministically.
	r.renderGridConnectors(ps, box, v, cells)
}

// renderGridConnectors draws each GridConnector in the gutter between the named adjacent
// columns. The gutter box is bounded by the right edge of the left column's first-row
// cell and the left edge of the right column's, spanning the full grid height. A Label
// (when set) sits below the glyph in the gutter. Validated adjacent/in-range at Stage-1.
func (r *renderer) renderGridConnectors(ps *pptx.Slide, box pptx.Box, v Grid, cells []pptx.Box) {
	if len(v.Connectors) == 0 || len(cells) < v.Columns {
		return
	}
	for _, gc := range v.Connectors {
		c0, c1 := gc.Between[0], gc.Between[1]
		if c1 != c0+1 || c0 < 0 || c1 >= v.Columns || c1 >= len(cells) {
			continue // defensive; Stage-1 already rejected bad indices
		}
		left, right := cells[c0], cells[c1]
		gutter := pptx.Box{X: left.Right(), Y: box.Y, W: right.X - left.Right(), H: box.H}
		if gutter.W <= 0 {
			continue
		}
		glyphGutter := gutter
		if gc.Label != "" {
			// Reserve the lower third of the gutter for the caption; the glyph centers
			// in the upper part.
			glyphGutter.H = gutter.H * 2 / 3
			labelBox := pptx.Box{X: gutter.X, Y: gutter.Y + glyphGutter.H, W: gutter.W, H: gutter.H - glyphGutter.H}
			tf := ps.AddTextFrame(labelBox).Anchor(pptx.AnchorTop)
			p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
			p.AddRun(gc.Label, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
			r.stats.Shapes++
		}
		r.renderConnector(ps, glyphGutter, gc.Kind, false)
	}
}

// ratioWeights maps a two_column ratio to per-column weights.
func ratioWeights(rt ColumnRatio) []int {
	switch rt {
	case Ratio12:
		return []int{1, 2}
	case Ratio21:
		return []int{2, 1}
	default: // Ratio11
		return []int{1, 1}
	}
}
