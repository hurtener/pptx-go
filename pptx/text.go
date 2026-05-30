package pptx

import (
	"fmt"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// ============================================================================
// Rich text model (RFC §8.4, §9)
// ============================================================================
//
// TextFrame → Paragraph → Run is the shared rich-text model. Runs are styled
// through tokens (RunStyle.TypeRole, RunStyle.Color) that resolve against the
// slide's active theme when the run is added — the same call-time resolution
// shapes use for fills (D-033), so a theme set before authoring colors the
// text. The model maps onto the internal text wire types (P3 keeps those
// internal).

// AutoFitMode controls how a TextFrame fits text to its shape (RFC §8.4).
type AutoFitMode int

const (
	// AutoFitNone keeps the text and shape sizes fixed (text may overflow).
	AutoFitNone AutoFitMode = iota
	// AutoFitNormal shrinks the font to fit the shape.
	AutoFitNormal
	// AutoFitShape grows the shape to fit the text.
	AutoFitShape
)

// TextAnchor is a TextFrame's vertical text anchor.
type TextAnchor int

const (
	AnchorTop TextAnchor = iota
	AnchorMiddle
	AnchorBottom
)

// Alignment is a paragraph's horizontal alignment.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
	AlignJustify
)

// Underline is a run's underline style.
type Underline int

const (
	UnderlineNone Underline = iota
	UnderlineSingle
	UnderlineDouble
)

// Strike is a run's strikethrough style.
type Strike int

const (
	StrikeNone Strike = iota
	StrikeSingle
	StrikeDouble
)

// BaselineShift raises or lowers a run relative to the baseline.
type BaselineShift int

const (
	BaselineNone BaselineShift = iota
	Superscript
	Subscript
)

// BulletKind is a paragraph's bullet style.
type BulletKind int

const (
	BulletNone BulletKind = iota
	BulletDisc
	BulletNumber
	BulletCheckbox
)

// RunStyle is the token-typed styling of a Run (RFC §8.4). TypeRole selects the
// typography scale (size + family); Color is a theme token or literal. The
// zero RunStyle uses the theme's TypeDisplay role — set TypeRole (e.g.
// TypeBody) for body text.
type RunStyle struct {
	TypeRole    TypeRole
	Color       Color
	Bold        bool
	Italic      bool
	Underline   Underline
	Strike      Strike
	BaselineRel BaselineShift
	Code        bool // inline code: monospace + a subtle tint (D-013; Chunk B)
}

// ParagraphOpts configures a paragraph at creation time.
type ParagraphOpts struct {
	Align  Alignment
	Level  int
	Bullet BulletKind
}

// TextFrame is a shape-level rich-text container (RFC §8.4). Create one with
// Slide.AddTextFrame.
type TextFrame struct {
	s    *Slide
	body *slide.XTextBody
}

// Paragraph is a line block within a TextFrame.
type Paragraph struct {
	tf  *TextFrame
	idx int
}

// Run is a styled text span within a Paragraph.
type Run struct {
	run *slide.XTextRun
}

// AddTextFrame adds an (initially empty) text-box shape positioned by box (EMU)
// and returns its TextFrame for authoring (RFC §8.4).
func (s *Slide) AddTextFrame(box Box) *TextFrame {
	id := int(s.part.AllocateShapeID())
	sp := &slide.XSp{
		NonVisual: slide.XNonVisualDrawingShape{
			CNvPr:   &slide.XNvCxnSpPr{ID: id, Name: fmt.Sprintf("TextFrame %d", id)},
			CNvSpPr: &slide.XNvSpPr{},
		},
		ShapeProperties: &slide.XShapeProperties{
			Transform2D: &slide.XTransform2D{
				Offset: &slide.XOv2DrOffset{X: int(box.X), Y: int(box.Y)},
				Extent: &slide.XOv2DrExtent{Cx: int(box.W), Cy: int(box.H)},
			},
			PresetGeom: &slide.XPresetGeometry{Prst: "rect", AvLst: &slide.XAvLst{}},
		},
		TextBody: &slide.XTextBody{
			BodyPr:   &slide.XBodyPr{Wrap: "square"},
			LstStyle: &slide.XTextParagraphList{},
		},
	}
	s.part.AppendShapeChild(sp)
	return &TextFrame{s: s, body: sp.TextBody}
}

// Clear removes all paragraphs from the frame and returns it (e.g. to refill a
// table cell's default body).
func (tf *TextFrame) Clear() *TextFrame {
	tf.body.Paragraphs = nil
	return tf
}

// AddParagraph appends a paragraph and returns it.
func (tf *TextFrame) AddParagraph(opts ParagraphOpts) *Paragraph {
	tf.body.Paragraphs = append(tf.body.Paragraphs, slide.XTextParagraph{})
	p := &Paragraph{tf: tf, idx: len(tf.body.Paragraphs) - 1}
	if opts.Align != AlignLeft {
		p.Align(opts.Align)
	}
	if opts.Level != 0 {
		p.Indent(opts.Level)
	}
	if opts.Bullet != BulletNone {
		p.Bullet(opts.Bullet)
	}
	return p
}

// x returns the addressable underlying paragraph (re-derived each call so it
// survives slice reallocation).
func (p *Paragraph) x() *slide.XTextParagraph { return &p.tf.body.Paragraphs[p.idx] }

// pr ensures and returns the paragraph's property element.
func (p *Paragraph) pr() *slide.XParaProps {
	xp := p.x()
	if xp.Pr == nil {
		xp.Pr = &slide.XParaProps{}
	}
	return xp.Pr
}

// AddRun appends a styled run and returns it. Token color and typography
// resolve against the slide's active theme now.
func (p *Paragraph) AddRun(text string, style RunStyle) *Run {
	run := &slide.XTextRun{Text: text, TextProperties: style.toProps(p.tf.s.activeTheme())}
	xp := p.x()
	xp.Content = append(xp.Content, run)
	return &Run{run: run}
}

// AddBreak appends a line break to the paragraph.
func (p *Paragraph) AddBreak() {
	xp := p.x()
	xp.Content = append(xp.Content, &slide.XTextBreak{})
}

// Align sets the paragraph's horizontal alignment and returns it.
func (p *Paragraph) Align(a Alignment) *Paragraph {
	p.pr().Alignment = alignString(a)
	return p
}

// Indent sets the paragraph's outline/indent level (0-based) and returns it.
func (p *Paragraph) Indent(level int) *Paragraph {
	if level < 0 {
		level = 0
	}
	p.pr().Level = level
	return p
}

// Bullet sets the paragraph's bullet style and returns it. A non-none bullet
// also sets a hanging indent so the marker renders outside the text.
func (p *Paragraph) Bullet(kind BulletKind) *Paragraph {
	pr := p.pr()
	pr.BuNone, pr.BuChar, pr.BuAutoNum = nil, nil, nil
	switch kind {
	case BulletNone:
		pr.BuNone = &slide.XEmptyElem{}
	case BulletDisc:
		pr.BuChar = &slide.XBuChar{Char: "•"}
	case BulletNumber:
		pr.BuAutoNum = &slide.XBuAutoNum{Type: "arabicPeriod"}
	case BulletCheckbox:
		pr.BuChar = &slide.XBuChar{Char: "☐"} // ☐ ballot box
	}
	if kind != BulletNone {
		pr.MarL = 457200    // 0.5"
		pr.Indent = -457200 // hanging
	}
	return p
}

// AutoFit sets the frame's auto-fit behavior and returns the frame.
func (tf *TextFrame) AutoFit(mode AutoFitMode) *TextFrame {
	bp := tf.bodyPr()
	bp.NoAutofit, bp.NormAutofit, bp.SpAutoFit = nil, nil, nil
	switch mode {
	case AutoFitNone:
		bp.NoAutofit = &slide.XEmptyElem{}
	case AutoFitNormal:
		bp.NormAutofit = &slide.XNormAutofit{}
	case AutoFitShape:
		bp.SpAutoFit = &slide.XEmptyElem{}
	}
	return tf
}

// Anchor sets the frame's vertical text anchor and returns the frame.
func (tf *TextFrame) Anchor(v TextAnchor) *TextFrame {
	tf.bodyPr().Anchor = anchorString(v)
	return tf
}

// Margins sets the frame's internal insets (EMU) and returns the frame.
func (tf *TextFrame) Margins(top, right, bottom, left EMU) *TextFrame {
	bp := tf.bodyPr()
	t, r, b, l := int(top), int(right), int(bottom), int(left)
	bp.TIns, bp.RIns, bp.BIns, bp.LIns = &t, &r, &b, &l
	return tf
}

// bodyPr ensures and returns the frame's body-properties element.
func (tf *TextFrame) bodyPr() *slide.XBodyPr {
	if tf.body.BodyPr == nil {
		tf.body.BodyPr = &slide.XBodyPr{}
	}
	return tf.body.BodyPr
}
