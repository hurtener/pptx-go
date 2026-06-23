package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// Button composer (RFC §11.1 / §12, R12.1, D-094). A Button renders native chrome —
// a RadiusFull rounded-rect filled per its tone (ghost = no fill + an accent
// hairline) over a middle-anchored bold TypeBody label flanked by optional native
// custGeom icons. It is presentational only: no hyperlink/action wiring (the deck is
// static), so it adds no builder capability (P1) and registers no media (D-026). The
// width is content-fit (label + icons + padding) clamped to the box; Align offsets
// the pill within the box.

// buttonMetric is the pinned geometry for a button size: the pill height, the
// horizontal padding each side of the content, the icon-to-label gap, and the icon
// size. A pinned layout metric (it sizes geometry, not a visual property) — not a
// theme token. The padX is generous enough to absorb a text frame's default inset so
// the measured label fits on one line (mirrors cardPillPadX's rationale).
type buttonMetric struct {
	height pptx.EMU
	padX   pptx.EMU
	gap    pptx.EMU
	iconSz pptx.EMU
}

// buttonMetrics returns the pinned geometry for a button size. ButtonMD is the
// default; SM/LG step it down/up. Pure function of the size (deterministic).
func buttonMetrics(size ButtonSize) buttonMetric {
	switch size {
	case ButtonSM:
		return buttonMetric{height: pptx.In(0.34), padX: pptx.In(0.14), gap: pptx.In(0.08), iconSz: pptx.In(0.16)}
	case ButtonLG:
		return buttonMetric{height: pptx.In(0.56), padX: pptx.In(0.22), gap: pptx.In(0.12), iconSz: pptx.In(0.26)}
	default: // ButtonMD
		return buttonMetric{height: pptx.In(0.44), padX: pptx.In(0.18), gap: pptx.In(0.10), iconSz: pptx.In(0.20)}
	}
}

// buttonWidthOf returns the content-fit width of a button: the label's natural width
// at TypeBody plus each present icon (iconSz + gap) plus padding each side, floored at
// the pill height (so a one-character button stays a proper capsule) and clamped to
// boxW. Shared by the composer; deterministic (pure integer naturalWidth). When the
// fit width exceeds boxW it is clamped and the label is shrunk to one line by fitScale.
func buttonWidthOf(theme *pptx.Theme, v Button, m buttonMetric, boxW pptx.EMU) pptx.EMU {
	w := naturalWidth(RichText{{Text: v.Label, Style: RunStyle{TypeRole: pptx.TypeBody}}}, theme) + 2*m.padX
	if v.LeadingIcon != "" {
		w += m.iconSz + m.gap
	}
	if v.TrailingIcon != "" {
		w += m.iconSz + m.gap
	}
	if w < m.height {
		w = m.height // circular floor — a tiny label stays a proper pill
	}
	if w > boxW {
		w = boxW
	}
	return w
}

// buttonToneStyle maps a tone to its fill options and label/icon color. Token-bound
// (P2), so a theme swap re-skins every button. Ghost is an outline (no fill + an
// accent hairline); the solid tones use an inverse label, neutral the default text.
func buttonToneStyle(tone ButtonTone) ([]pptx.ShapeOption, pptx.Color) {
	switch tone {
	case ButtonAccentAlt:
		return []pptx.ShapeOption{pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccentAlt)))}, pptx.TokenTextColor(pptx.TextInverse)
	case ButtonGhost:
		return []pptx.ShapeOption{
			pptx.WithFill(pptx.NoFill()),
			pptx.WithLine(pptx.Line{Width: pptx.Pt(1), Color: pptx.TokenColor(pptx.ColorAccent)}),
		}, pptx.TokenTextColor(pptx.TextAccent)
	case ButtonNeutral:
		return []pptx.ShapeOption{pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt)))}, pptx.TokenTextColor(pptx.TextPrimary)
	default: // ButtonPrimary
		return []pptx.ShapeOption{pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent)))}, pptx.TokenTextColor(pptx.TextInverse)
	}
}

func (r *renderer) renderButton(ps *pptx.Slide, box pptx.Box, v Button, hAlign HAlign) {
	m := buttonMetrics(v.Size)
	w := buttonWidthOf(r.theme, v, m, box.W)

	// Vertically center the pill within its slot (the slot is the button height in the
	// body/card stack, so this is a no-op there; it keeps the pill centered if a
	// container hands it a taller box).
	pillH := m.height
	if pillH > box.H {
		pillH = box.H
	}
	pillY := box.Y + (box.H-pillH)/2

	// Horizontal offset within the box per the effective alignment (the body stack
	// passes the slide/per-node alignment; a container cell passes HAlignLeft).
	offsetX := pptx.EMU(0)
	if w < box.W {
		switch hAlign {
		case HAlignCenter:
			offsetX = (box.W - w) / 2
		case HAlignRight:
			offsetX = box.W - w
		}
	}
	pillBox := pptx.Box{X: box.X + offsetX, Y: pillY, W: w, H: pillH}

	fillOpts, textColor := buttonToneStyle(v.Tone)
	opts := append([]pptx.ShapeOption{pptx.WithRadius(pptx.RadiusFull)}, fillOpts...)
	ps.AddShape(pptx.ShapeRoundRect, pillBox, opts...)
	r.stats.Shapes++

	// Lay out the content row inside the pill: [leading icon] label [trailing icon].
	// The label occupies the region between any icons and is centered there; icons are
	// vertically centered in the pill.
	iconY := pillBox.Y + (pillH-m.iconSz)/2
	labelX := pillBox.X + m.padX
	labelW := pillBox.W - 2*m.padX

	if v.LeadingIcon != "" {
		iconBox := pptx.Box{X: labelX, Y: iconY, W: m.iconSz, H: m.iconSz}
		r.addButtonIcon(ps, iconBox, v.LeadingIcon, textColor)
		labelX += m.iconSz + m.gap
		labelW -= m.iconSz + m.gap
	}
	if v.TrailingIcon != "" {
		iconBox := pptx.Box{X: pillBox.X + pillBox.W - m.padX - m.iconSz, Y: iconY, W: m.iconSz, H: m.iconSz}
		r.addButtonIcon(ps, iconBox, v.TrailingIcon, textColor)
		labelW -= m.iconSz + m.gap
	}
	if labelW < 0 {
		labelW = 0
	}

	// Single-line guarantee: if the label cannot fit its region (only when the pill was
	// clamped to the box), shrink it to one line via FontScale (the R10.5 fit primitive).
	scale := fitScale(naturalWidthAt(RichText{{Text: v.Label}}, pptx.TypeBody, r.theme), labelW)
	tf := ps.AddTextFrame(pptx.Box{X: labelX, Y: pillBox.Y, W: labelW, H: pillH}).Anchor(pptx.AnchorMiddle)
	p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	p.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeBody, Bold: true, Color: textColor, FontScale: scale})
	r.stats.Shapes++
}

// addButtonIcon renders a closed-name icon as a native custGeom glyph filled with the
// button's label color. The name resolved at Stage-1 (validateIconRefs via the
// walkIconRefs Button case), so a miss here is a wiring bug surfaced as a warning.
func (r *renderer) addButtonIcon(ps *pptx.Slide, box pptx.Box, name string, color pptx.Color) {
	svg, ok := r.cfg.icons.Lookup(name)
	if !ok {
		r.warn("", fmt.Sprintf("button icon %q not found at compose (should have failed Stage-1)", name))
		return
	}
	if _, err := ps.AddIcon(svg, box, pptx.WithFill(pptx.SolidFill(color))); err != nil {
		r.warn("", fmt.Sprintf("button icon %q failed to render: %v", name, err))
		return
	}
	r.stats.Shapes++
}
