package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// IconRows composer (RFC §11.1 / §12, R12.7, D-100). A vertical stack of [icon | label |
// optional right-aligned meta] rows, each optionally framed by a SurfaceAlt pill, with a
// Fill mode that distributes inter-row slack so the rows span the box. Mirrors the
// Checklist row engine (Phase 62). Media-free, deterministic integer-EMU layout.

// Pinned layout metrics (not theme tokens — geometry).
const (
	iconRowsIconSz  = pptx.EMU(201168) // In(0.22); leading icon
	iconRowsGap     = pptx.EMU(109728) // In(0.12); icon-to-label gap
	iconRowsRowGap  = pptx.EMU(91440)  // In(0.10); default inter-row gap
	iconRowsLineH   = pptx.EMU(311112) // In(0.34); per label line
	iconRowsPillPad = pptx.EMU(109728) // In(0.12); RowPill horizontal inset
	iconRowsMetaGap = pptx.EMU(109728) // In(0.12); label-to-meta gap
)

// iconRowsGlyphColorRole resolves the icon tint: the caller's GlyphColor, or ColorAccent
// when it is the zero value (ColorCanvas) — a canvas-colored glyph would be invisible.
func iconRowsGlyphColorRole(v IconRows) pptx.ColorRole {
	if v.GlyphColor == ColorCanvas {
		return pptx.ColorAccent
	}
	return v.GlyphColor
}

// iconRowContentX returns the row's content inset (the RowPill frame pad) and width.
func iconRowContentX(box pptx.Box, tone RowTone) (x, w pptx.EMU) {
	if tone == RowPill {
		return box.X + iconRowsPillPad, box.W - 2*iconRowsPillPad
	}
	return box.X, box.W
}

// iconRowMetaW returns the right-aligned meta column width (0 when no meta), content-fit
// and clamped to a third of the content width.
func iconRowMetaW(meta RichText, contentW pptx.EMU, theme *pptx.Theme) pptx.EMU {
	if len(meta) == 0 {
		return 0
	}
	w := naturalWidthAt(meta, pptx.TypeCaption, theme) + iconRowsMetaGap
	if max := contentW / 3; w > max {
		w = max
	}
	return w
}

// iconRowsLabelW returns the label column width for a row: the content width minus the
// icon column and the meta column.
func iconRowsLabelW(box pptx.Box, row IconRow, theme *pptx.Theme) pptx.EMU {
	_, contentW := iconRowContentX(box, row.Tone)
	w := contentW
	if row.Icon != "" {
		w -= iconRowsIconSz + iconRowsGap
	}
	if mw := iconRowMetaW(row.Meta, contentW, theme); mw > 0 {
		w -= mw + iconRowsMetaGap
	}
	if w < 0 {
		w = 0
	}
	return w
}

// iconRowsRowHeights returns the per-row heights: the wrapped label line count × the line
// height, floored at the icon size so a single-line row still fits its glyph.
func iconRowsRowHeights(v IconRows, box pptx.Box, theme *pptx.Theme) []pptx.EMU {
	heights := make([]pptx.EMU, len(v.Rows))
	for i, row := range v.Rows {
		lines := wrappedLines(row.Label, pptx.TypeBody, iconRowsLabelW(box, row, theme), theme)
		h := iconRowsLineH * pptx.EMU(lines)
		if h < iconRowsIconSz {
			h = iconRowsIconSz
		}
		heights[i] = h
	}
	return heights
}

// iconRowsPreferredHeight is the node's slot height: the summed per-row heights plus the
// default inter-row gaps. Content-aware, deterministic.
func iconRowsPreferredHeight(v IconRows, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	heights := iconRowsRowHeights(v, pptx.Box{W: avail}, theme)
	var total pptx.EMU
	for _, h := range heights {
		total += h
	}
	if len(heights) > 1 {
		total += iconRowsRowGap * pptx.EMU(len(heights)-1)
	}
	return total
}

func (r *renderer) renderIconRows(ps *pptx.Slide, box pptx.Box, v IconRows) {
	heights := iconRowsRowHeights(v, box, r.theme)
	rows := len(heights)
	var total pptx.EMU
	for _, h := range heights {
		total += h
	}

	// Fill: spread the slack across the inter-row gaps so the last row meets the bottom
	// (the VAlignJustify primitive, per-row); off → the pinned default gap.
	rowGap := iconRowsRowGap
	if v.Fill && rows > 1 {
		if slack := box.H - total; slack > iconRowsRowGap*pptx.EMU(rows-1) {
			rowGap = slack / pptx.EMU(rows-1)
		}
	}

	glyphColor := pptx.TokenColor(iconRowsGlyphColorRole(v))
	y := box.Y
	for i, row := range v.Rows {
		rowH := heights[i]
		rowBox := pptx.Box{X: box.X, Y: y, W: box.W, H: rowH}

		if row.Tone == RowPill {
			ps.AddShape(pptx.ShapeRoundRect, rowBox,
				pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))),
				pptx.WithRadius(pptx.RadiusMD))
			r.stats.Shapes++
		}

		contentX, contentW := iconRowContentX(rowBox, row.Tone)
		x := contentX
		if row.Icon != "" {
			iconBox := pptx.Box{X: contentX, Y: y + (iconRowsLineH-iconRowsIconSz)/2, W: iconRowsIconSz, H: iconRowsIconSz}
			r.addIconRowGlyph(ps, iconBox, row.Icon, glyphColor)
			x += iconRowsIconSz + iconRowsGap
		}

		metaW := iconRowMetaW(row.Meta, contentW, r.theme)
		if metaW > 0 {
			metaBox := pptx.Box{X: contentX + contentW - metaW, Y: y, W: metaW, H: rowH}
			tf := ps.AddTextFrame(metaBox).Anchor(pptx.AnchorMiddle)
			p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignRight})
			r.addRichText(ps, p, row.Meta, pptx.TypeCaption)
			r.stats.Shapes++
		}

		labelW := contentX + contentW - x
		if metaW > 0 {
			labelW -= metaW + iconRowsMetaGap
		}
		if labelW < 0 {
			labelW = 0
		}
		tf := ps.AddTextFrame(pptx.Box{X: x, Y: y, W: labelW, H: rowH}).Anchor(pptx.AnchorMiddle)
		p := tf.AddParagraph(pptx.ParagraphOpts{LineHeight: r.lineH(pptx.TypeBody)})
		r.addRichText(ps, p, row.Label, pptx.TypeBody)
		r.stats.Shapes++

		y += rowH + rowGap
	}
}

// addIconRowGlyph renders a row's leading icon as a native custGeom glyph filled with
// color. The name resolved at Stage-1 (walkIconRefs), so a miss is a wiring bug warned on.
func (r *renderer) addIconRowGlyph(ps *pptx.Slide, box pptx.Box, name string, color pptx.Color) {
	svg, ok := r.cfg.icons.Lookup(name)
	if !ok {
		r.warn("", fmt.Sprintf("icon-row icon %q not found at compose (should have failed Stage-1)", name))
		return
	}
	if _, err := ps.AddIcon(svg, box, pptx.WithFill(pptx.SolidFill(color))); err != nil {
		r.warn("", fmt.Sprintf("icon-row icon %q failed to render: %v", name, err))
		return
	}
	r.stats.Shapes++
}
