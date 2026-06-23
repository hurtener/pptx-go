package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// Lockup composer (RFC §11.1 / §12, R12.9, D-102). A compact attribution mark: a caption
// paired with a small partner logo (an asset pic or a curated icon) composed as one
// centered inline group. The asset variant resolves through the AssetResolver and renders
// a pic (so the slide composes serially for deterministic part numbering); the icon
// variant is media-free. Deterministic integer-EMU layout.

// Pinned layout metrics (not theme tokens — geometry).
const (
	lockupDefaultH = pptx.EMU(365760) // In(0.40); default logo height when MaxHeight is 0
	lockupGap      = pptx.EMU(109728) // In(0.12); caption-to-logo gap
	lockupPadY     = pptx.EMU(45720)  // In(0.05); vertical padding in the slot estimate
)

// lockupLogoH returns the logo height: MaxHeight when set, else the pinned default.
func lockupLogoH(v Lockup) pptx.EMU {
	if v.MaxHeight > 0 {
		return v.MaxHeight
	}
	return lockupDefaultH
}

// lockupPreferredHeight is the node's slot height: the logo height plus a little vertical
// padding (the caption sits at a smaller type role, so the logo bounds the height).
func lockupPreferredHeight(v Lockup) pptx.EMU {
	return lockupLogoH(v) + 2*lockupPadY
}

func (r *renderer) renderLockup(ps *pptx.Slide, box pptx.Box, v Lockup, slideID string, hAlign HAlign) {
	logoH := lockupLogoH(v)
	if logoH > box.H {
		logoH = box.H
	}
	logoW := logoH // square logo box — no pixel aspect available (§7); Fit handles the image
	captionW := pptx.EMU(0)
	if v.Caption != "" {
		captionW = naturalWidth(RichText{{Text: v.Caption, Style: RunStyle{TypeRole: pptx.TypeCaption}}}, r.theme)
	}

	gap := pptx.EMU(0)
	if v.Caption != "" {
		gap = lockupGap
	}
	groupW := captionW + gap + logoW

	// Center / align the whole group within the box.
	offsetX := pptx.EMU(0)
	if groupW < box.W {
		switch hAlign {
		case HAlignCenter:
			offsetX = (box.W - groupW) / 2
		case HAlignRight:
			offsetX = box.W - groupW
		}
	}
	x := box.X + offsetX
	y := box.Y + (box.H-logoH)/2

	drawCaption := func(at pptx.EMU) {
		if v.Caption == "" {
			return
		}
		tf := ps.AddTextFrame(pptx.Box{X: at, Y: box.Y, W: captionW, H: box.H}).Anchor(pptx.AnchorMiddle)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.Caption, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++
	}
	drawLogo := func(at pptx.EMU) {
		logoBox := pptx.Box{X: at, Y: y, W: logoW, H: logoH}
		if v.AssetID != "" {
			data, ct, err := r.resolve(v.AssetID)
			if err != nil {
				r.warn(slideID, fmt.Sprintf("lockup asset %q unresolved: %v", v.AssetID, err))
				return
			}
			if _, aerr := ps.AddImage(pptx.ImageBytes(data, ct), logoBox); aerr != nil {
				r.warn(slideID, fmt.Sprintf("lockup image %q: %v", v.AssetID, aerr))
				return
			}
			r.stats.Shapes++
			r.stats.Assets++
			return
		}
		// Icon variant (media-free).
		if svg, ok := r.cfg.icons.Lookup(v.Icon); ok {
			if _, err := ps.AddIcon(svg, logoBox); err == nil {
				r.stats.Shapes++
			}
		} else {
			r.warn(slideID, fmt.Sprintf("lockup icon %q not found at compose (should have failed Stage-1)", v.Icon))
		}
	}

	if v.AssetSide == TrailCaption {
		// logo then caption
		drawLogo(x)
		drawCaption(x + logoW + gap)
	} else {
		// caption then logo (LeadCaption, default)
		drawCaption(x)
		drawLogo(x + captionW + gap)
	}
}
