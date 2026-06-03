package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// CodeBlock composer (RFC §11.1 / §12, D-014/D-045). A code_block renders as a
// caller-rasterized image (a pic, D-014) with an optional caption below and an
// optional language badge overlaid top-right. The badge is a native pill drawn
// after the pic so it sits on top (shape-tree order = z-order); an empty
// Language emits no badge. Caller code rendering is the caller's job — pptx-go
// embeds the bytes (D-014) and degrades a missing asset to a warning (D-036).

const (
	codeCaptionH = pptx.EMU(365760) // 0.4"
	codeBadgeH   = pptx.EMU(274320) // 0.3"
	codeBadgeW   = pptx.EMU(777240) // ~0.85"
	codeBadgePad = pptx.EMU(109728) // ~0.12" inset from the image corner
)

func (r *renderer) renderCodeBlock(ps *pptx.Slide, box pptx.Box, v CodeBlock, slideID string) {
	imgBox := box
	if v.Caption != "" {
		imgBox.H = box.H - codeCaptionH
	}

	imageOK := false
	data, ct, err := r.resolve(v.AssetID)
	if err != nil {
		r.warn(slideID, fmt.Sprintf("code_block asset %q unresolved: %v", v.AssetID, err))
	} else if _, aerr := ps.AddImage(pptx.ImageBytes(data, ct), imgBox); aerr != nil {
		r.warn(slideID, fmt.Sprintf("code_block image %q: %v", v.AssetID, aerr))
	} else {
		r.stats.Shapes++
		r.stats.Assets++
		imageOK = true
	}

	// Language badge: a small overlay pill in the image's top-right (D-045).
	// Drawn only over a rendered image, and only when Language is set.
	if imageOK && v.Language != "" {
		badge := pptx.Box{
			X: imgBox.Right() - codeBadgePad - codeBadgeW,
			Y: imgBox.Y + codeBadgePad,
			W: codeBadgeW,
			H: codeBadgeH,
		}
		if badge.X < imgBox.X {
			badge.X = imgBox.X
		}
		ps.AddShape(pptx.ShapeRoundRect, badge,
			pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))),
			pptx.WithRadius(pptx.RadiusFull))
		r.stats.Shapes++
		tf := ps.AddTextFrame(badge).Anchor(pptx.AnchorMiddle)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.Language, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextSecondary)})
		r.stats.Shapes++
	}

	if v.Caption != "" {
		capBox := pptx.Box{X: box.X, Y: imgBox.Y + imgBox.H, W: box.W, H: codeCaptionH}
		tf := ps.AddTextFrame(capBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.Caption, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++
	}
}
