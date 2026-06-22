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
	cols := layout.Columns(box, ratioWeights(v.Ratio), gap)
	if len(cols) != 2 {
		return
	}
	for _, pl := range r.stackIn(cols[0], v.Left, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
	for _, pl := range r.stackIn(cols[1], v.Right, slideID) {
		r.renderNode(ps, pl.box, pl.node, slideID, pl.hAlign)
	}
	// Inter-column element (D-055), drawn after the column content so it sits on
	// top of both columns, centered on the seam. Inert when Join == JoinNone.
	if v.Join != JoinNone {
		r.renderColumnJoin(ps, box, cols[0], cols[1], v)
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
