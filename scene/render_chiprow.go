package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// ChipRow composer (RFC §11.1 / §12, R12.5, D-096). A ChipRow lays a sequence of
// content-fit chip pills left-to-right, wrapping onto new lines when Wrap is set, with
// an optional leading label on the first line. Each chip reuses the single Chip's
// tone→fill treatment. Greedy integer-EMU packing → deterministic; media-free (native
// rounded-rect pills + optional custGeom icons).

// Pinned layout metrics (not theme tokens — they size geometry, like buttonMetrics).
const (
	chipRowChipH   = pptx.EMU(311112) // In(0.34); pill height
	chipRowPadX    = pptx.EMU(109728) // In(0.12); horizontal padding each side of a chip
	chipRowIconSz  = pptx.EMU(128016) // In(0.14); leading chip icon
	chipRowIconGap = pptx.EMU(54864)  // In(0.06); icon-to-label gap
	chipRowGap     = pptx.EMU(91440)  // In(0.10); gap between chips and between lines
)

// chipRowLine is the chip indices packed onto one line.
type chipRowLine struct{ idxs []int }

// chipWidthOf is a single chip's content-fit width: the label at TypeBodySmall plus
// padding, plus a leading icon when present. Deterministic (pure integer naturalWidth).
func chipWidthOf(theme *pptx.Theme, c ChipSpec) pptx.EMU {
	w := naturalWidth(RichText{{Text: c.Label, Style: RunStyle{TypeRole: pptx.TypeBodySmall}}}, theme) + 2*chipRowPadX
	if c.Icon != "" {
		w += chipRowIconSz + chipRowIconGap
	}
	return w
}

// chipRowLabelW is the leading label's text width (TypeCaption), or 0 when no label.
func chipRowLabelW(theme *pptx.Theme, label string) pptx.EMU {
	if label == "" {
		return 0
	}
	return naturalWidth(RichText{{Text: label, Style: RunStyle{TypeRole: pptx.TypeCaption}}}, theme)
}

// chipRowLines greedily packs the chips into lines that each fit boxW. The leading
// label occupies the start of line 0. When Wrap is false everything stays on one line.
// Pure integer arithmetic over the ordered slice → deterministic.
func chipRowLines(v ChipRow, boxW pptx.EMU, theme *pptx.Theme) []chipRowLine {
	lines := []chipRowLine{{}}
	li := 0
	curX := chipRowLabelW(theme, v.Label)
	for i, c := range v.Chips {
		w := chipWidthOf(theme, c)
		hasContent := len(lines[li].idxs) > 0 || (li == 0 && v.Label != "")
		gap := pptx.EMU(0)
		if hasContent {
			gap = chipRowGap
		}
		if v.Wrap && hasContent && curX+gap+w > boxW {
			lines = append(lines, chipRowLine{})
			li++
			curX = 0
			gap = 0
		}
		lines[li].idxs = append(lines[li].idxs, i)
		curX += gap + w
	}
	return lines
}

// chipRowLineWidth recomputes a packed line's total width (label on line 0 + chips +
// inter-element gaps), matching chipRowLines' accounting, for the per-line HAlign offset.
func chipRowLineWidth(v ChipRow, line chipRowLine, lineIndex int, theme *pptx.Theme) pptx.EMU {
	var w pptx.EMU
	labeled := lineIndex == 0 && v.Label != ""
	if labeled {
		w += chipRowLabelW(theme, v.Label)
	}
	for k, idx := range line.idxs {
		if k > 0 || labeled {
			w += chipRowGap
		}
		w += chipWidthOf(theme, v.Chips[idx])
	}
	return w
}

// chipRowPreferredHeight is the node's slot height: one chip height per packed line plus
// the inter-line gaps. Content-aware (wrapping grows the line count), deterministic.
func chipRowPreferredHeight(v ChipRow, avail pptx.EMU, theme *pptx.Theme) pptx.EMU {
	n := len(chipRowLines(v, avail, theme))
	if n < 1 {
		n = 1
	}
	return pptx.EMU(n)*chipRowChipH + pptx.EMU(n-1)*chipRowGap
}

func (r *renderer) renderChipRow(ps *pptx.Slide, box pptx.Box, v ChipRow, hAlign HAlign) {
	lines := chipRowLines(v, box.W, r.theme)
	y := box.Y
	for li, line := range lines {
		lineW := chipRowLineWidth(v, line, li, r.theme)
		offX := pptx.EMU(0)
		if lineW < box.W {
			switch hAlign {
			case HAlignCenter:
				offX = (box.W - lineW) / 2
			case HAlignRight:
				offX = box.W - lineW
			}
		}
		x := box.X + offX

		labeled := li == 0 && v.Label != ""
		if labeled {
			lblW := chipRowLabelW(r.theme, v.Label)
			tf := ps.AddTextFrame(pptx.Box{X: x, Y: y, W: lblW, H: chipRowChipH}).Anchor(pptx.AnchorMiddle)
			p := tf.AddParagraph(pptx.ParagraphOpts{})
			p.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
			r.stats.Shapes++
			x += lblW
		}
		for k, idx := range line.idxs {
			if k > 0 || labeled {
				x += chipRowGap
			}
			c := v.Chips[idx]
			w := chipWidthOf(r.theme, c)
			r.drawChip(ps, pptx.Box{X: x, Y: y, W: w, H: chipRowChipH}, c)
			x += w
		}
		y += chipRowChipH + chipRowGap
	}
}

// drawChip renders one chip pill: a RadiusFull rounded-rect with the tone fill, an
// optional leading custGeom icon, and a centered TypeBodySmall label. Shared by ChipRow.
func (r *renderer) drawChip(ps *pptx.Slide, box pptx.Box, c ChipSpec) {
	var opts []pptx.ShapeOption
	switch c.Tone {
	case ChipSolid:
		opts = append(opts, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(c.Color))))
	case ChipOutline:
		opts = append(opts, pptx.WithFill(pptx.NoFill()),
			pptx.WithLine(pptx.Line{Width: pptx.Pt(1), Color: pptx.TokenColor(c.Color)}))
	default: // ChipTint
		opts = append(opts, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))))
	}
	opts = append(opts, pptx.WithRadius(pptx.RadiusFull))
	ps.AddShape(pptx.ShapeRoundRect, box, opts...)
	r.stats.Shapes++

	labelColor := r.chipLabelColor(c)
	labelX := box.X + chipRowPadX
	labelW := box.W - 2*chipRowPadX
	if c.Icon != "" {
		iconColor := labelColor
		if iconColor == nil {
			iconColor = pptx.TokenTextColor(pptx.TextPrimary)
		}
		iconBox := pptx.Box{X: labelX, Y: box.Y + (box.H-chipRowIconSz)/2, W: chipRowIconSz, H: chipRowIconSz}
		r.addChipIcon(ps, iconBox, c.Icon, iconColor)
		labelX += chipRowIconSz + chipRowIconGap
		labelW -= chipRowIconSz + chipRowIconGap
	}
	if labelW < 0 {
		labelW = 0
	}
	tf := ps.AddTextFrame(pptx.Box{X: labelX, Y: box.Y, W: labelW, H: box.H}).Anchor(pptx.AnchorMiddle)
	p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	p.AddRun(c.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Color: labelColor})
	r.stats.Shapes++
}

// chipLabelColor auto-contrasts a solid chip's label against its fill (R11.2); tint and
// outline chips sit on a light surface, so nil (the default text color) is byte-clean.
func (r *renderer) chipLabelColor(c ChipSpec) pptx.Color {
	if c.Tone == ChipSolid {
		return r.onCardSurface(c.Color)
	}
	return nil
}

// addChipIcon renders a closed-name icon as a native custGeom glyph filled with color.
func (r *renderer) addChipIcon(ps *pptx.Slide, box pptx.Box, name string, color pptx.Color) {
	svg, ok := r.cfg.icons.Lookup(name)
	if !ok {
		r.warn("", fmt.Sprintf("chip icon %q not found at compose (should have failed Stage-1)", name))
		return
	}
	if _, err := ps.AddIcon(svg, box, pptx.WithFill(pptx.SolidFill(color))); err != nil {
		r.warn("", fmt.Sprintf("chip icon %q failed to render: %v", name, err))
		return
	}
	r.stats.Shapes++
}
