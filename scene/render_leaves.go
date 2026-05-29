package scene

import (
	"fmt"

	"github.com/hurtener/pptx-go/pptx"
)

// Per-leaf composers (RFC §11.1 / §12). Each maps a node to builder calls
// following its intrinsic policy — native shapes, except code_block (an image).
// No product behavior (D-026): typography is the node's theme role, verbatim.

func (r *renderer) renderHero(ps *pptx.Slide, box pptx.Box, v Hero) {
	tf := ps.AddTextFrame(box).Anchor(pptx.AnchorMiddle)
	r.plainPara(tf, v.Eyebrow, pptx.TypeCaption, pptx.ParagraphOpts{})
	r.plainPara(tf, v.Title, pptx.TypeDisplay, pptx.ParagraphOpts{})
	r.plainPara(tf, v.Subtitle, pptx.TypeBody, pptx.ParagraphOpts{})
	r.stats.Shapes++
}

func (r *renderer) renderProse(ps *pptx.Slide, box pptx.Box, v Prose) {
	tf := ps.AddTextFrame(box)
	for _, para := range v.Paragraphs {
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		r.addRichText(ps, p, para, pptx.TypeBody)
	}
	r.stats.Shapes++
}

func (r *renderer) renderHeading(ps *pptx.Slide, box pptx.Box, v Heading) {
	tf := ps.AddTextFrame(box)
	p := tf.AddParagraph(pptx.ParagraphOpts{})
	r.addRichText(ps, p, v.Text, headingRole(v.Level))
	r.stats.Shapes++
}

func (r *renderer) renderList(ps *pptx.Slide, box pptx.Box, v List) {
	tf := ps.AddTextFrame(box)
	bullet := listBullet(v.Kind)
	for _, item := range v.Items {
		p := tf.AddParagraph(pptx.ParagraphOpts{Bullet: bullet, Level: item.Level})
		r.addRichText(ps, p, item.Text, pptx.TypeBody)
	}
	r.stats.Shapes++
}

func (r *renderer) renderDivider(ps *pptx.Slide, box pptx.Box, _ Divider) {
	rule := pptx.Box{X: box.X, Y: box.Y + box.H/2, W: box.W, H: pptx.Pt(1.5)}
	ps.AddShape(pptx.ShapeRect, rule, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))))
	r.stats.Shapes++
}

func (r *renderer) renderQuote(ps *pptx.Slide, box pptx.Box, v Quote) {
	tf := ps.AddTextFrame(box)
	p := tf.AddParagraph(pptx.ParagraphOpts{})
	r.addRichText(ps, p, v.Text, pptx.TypeH3)
	if v.Attribution != "" {
		r.plainPara(tf, "— "+v.Attribution, pptx.TypeCaption, pptx.ParagraphOpts{})
	}
	r.stats.Shapes++
}

func (r *renderer) renderCallout(ps *pptx.Slide, box pptx.Box, v Callout) {
	// Accent side-bar + a text block inset to its right.
	bar := pptx.Box{X: box.X, Y: box.Y, W: pptx.Pt(4), H: box.H}
	ps.AddShape(pptx.ShapeRect, bar, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(calloutColor(v.Kind)))))
	r.stats.Shapes++

	textBox := pptx.Box{X: box.X + pptx.In(0.2), Y: box.Y, W: box.W - pptx.In(0.2), H: box.H}
	tf := ps.AddTextFrame(textBox)
	if v.Title != "" {
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun(v.Title, pptx.RunStyle{TypeRole: pptx.TypeBody, Bold: true})
	}
	if len(v.Body) > 0 {
		p := tf.AddParagraph(pptx.ParagraphOpts{})
		r.addRichText(ps, p, v.Body, pptx.TypeBody)
	}
	r.stats.Shapes++
}

func (r *renderer) renderChip(ps *pptx.Slide, box pptx.Box, v Chip) {
	chip := pptx.Box{X: box.X, Y: box.Y, W: box.W, H: box.H}
	var opts []pptx.ShapeOption
	switch v.Tone {
	case ChipSolid:
		opts = append(opts, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(v.Color))))
	case ChipOutline:
		opts = append(opts, pptx.WithFill(pptx.NoFill()),
			pptx.WithLine(pptx.Line{Width: pptx.Pt(1), Color: pptx.TokenColor(v.Color)}))
	default: // ChipTint
		opts = append(opts, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt))))
	}
	ps.AddShape(pptx.ShapeRoundRect, chip, opts...)
	r.stats.Shapes++

	tf := ps.AddTextFrame(chip).Anchor(pptx.AnchorMiddle)
	p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	p.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall})
	r.stats.Shapes++
}

func (r *renderer) renderArrow(ps *pptx.Slide, box pptx.Box, v Arrow) {
	ps.AddShape(arrowGeom(v.Direction), box,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
	r.stats.Shapes++
	if v.Label != "" {
		tf := ps.AddTextFrame(box).Anchor(pptx.AnchorMiddle)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeBodySmall, Color: pptx.TokenTextColor(pptx.TextInverse)})
		r.stats.Shapes++
	}
}

func (r *renderer) renderCodeBlock(ps *pptx.Slide, box pptx.Box, v CodeBlock, slideID string) {
	imgBox := box
	if v.Caption != "" {
		imgBox.H = box.H - pptx.In(0.4)
	}
	data, ct, err := r.resolve(v.AssetID)
	if err != nil {
		r.warn(slideID, fmt.Sprintf("code_block asset %q unresolved: %v", v.AssetID, err))
	} else if _, aerr := ps.AddImage(pptx.ImageBytes(data, ct), imgBox); aerr != nil {
		r.warn(slideID, fmt.Sprintf("code_block image %q: %v", v.AssetID, aerr))
	} else {
		r.stats.Shapes++
		r.stats.Assets++
	}
	if v.Caption != "" {
		capBox := pptx.Box{X: box.X, Y: imgBox.Y + imgBox.H, W: box.W, H: pptx.In(0.4)}
		tf := ps.AddTextFrame(capBox)
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.Caption, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextMuted)})
		r.stats.Shapes++
	}
}

func (r *renderer) renderSectionDivider(ps *pptx.Slide, box pptx.Box, v SectionDivider) {
	// Full-bleed background fill + centered label.
	ps.AddShape(pptx.ShapeRect, box, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
	r.stats.Shapes++

	tf := ps.AddTextFrame(box).Anchor(pptx.AnchorMiddle)
	if v.Eyebrow != "" {
		p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
		p.AddRun(v.Eyebrow, pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextInverse)})
	}
	p := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	p.AddRun(v.Label, pptx.RunStyle{TypeRole: pptx.TypeDisplay, Color: pptx.TokenTextColor(pptx.TextInverse)})
	r.stats.Shapes++
}

// ---- helpers --------------------------------------------------------------

func (r *renderer) plainPara(tf *pptx.TextFrame, text string, role pptx.TypeRole, opts pptx.ParagraphOpts) {
	if text == "" {
		return
	}
	tf.AddParagraph(opts).AddRun(text, pptx.RunStyle{TypeRole: role})
}

func (r *renderer) resolve(id AssetID) ([]byte, string, error) {
	if r.cfg.resolver == nil {
		return nil, "", ErrAssetNotFound
	}
	return r.cfg.resolver.Resolve(r.ctx, id)
}

func headingRole(level int) pptx.TypeRole {
	switch level {
	case 1:
		return pptx.TypeH1
	case 2:
		return pptx.TypeH2
	case 3:
		return pptx.TypeH3
	case 4:
		return pptx.TypeH4
	default:
		return pptx.TypeH5
	}
}

func listBullet(k ListKind) pptx.BulletKind {
	switch k {
	case ListNumber:
		return pptx.BulletNumber
	case ListChecklist:
		return pptx.BulletCheckbox
	default:
		return pptx.BulletDisc
	}
}

func calloutColor(k CalloutKind) pptx.ColorRole {
	switch k {
	case CalloutWarning:
		return pptx.ColorWarning
	case CalloutTip:
		return pptx.ColorSuccess
	case CalloutImportant:
		return pptx.ColorAccent
	default: // CalloutNote
		return pptx.ColorInfo
	}
}

func arrowGeom(d ArrowDirection) pptx.ShapeGeometry {
	switch d {
	case ArrowLeft:
		return pptx.ShapeGeometry("leftArrow")
	case ArrowUp:
		return pptx.ShapeGeometry("upArrow")
	case ArrowDown:
		return pptx.ShapeGeometry("downArrow")
	default:
		return pptx.ShapeRightArrow
	}
}
