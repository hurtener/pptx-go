package pptx

import (
	"fmt"
	"math"

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
	// Tracking optionally overrides the type role's letter-spacing for this run,
	// in points (signed). nil inherits the role's FontSpec.Tracking; a non-nil
	// value (including 0) wins over the role. Emitted as a:rPr/@spc (D-060).
	Tracking *float64
}

// ParagraphOpts configures a paragraph at creation time.
type ParagraphOpts struct {
	Align  Alignment
	Level  int
	Bullet BulletKind
	// LineHeight is the paragraph's line spacing as a percent of single
	// (100 = single, 120 = 1.2×); 0 and 100 emit nothing (byte-identical).
	// Emitted as OOXML a:pPr/a:lnSpc/a:spcPct (D-061).
	LineHeight float64
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
	tf  *TextFrame // owning frame (read side: resolves the hyperlink relationship)
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
	// Line spacing: emit a:lnSpc/a:spcPct only when set to a non-single value, so
	// the default (0 or 100) stays byte-identical (D-061).
	if opts.LineHeight != 0 && opts.LineHeight != 100 {
		p.pr().LnSpc = &slide.XLnSpc{SpcPct: &slide.XSpcPct{Val: int(math.Round(opts.LineHeight * 1000))}}
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
	return &Run{tf: p.tf, run: run}
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

// AutoFitMode returns the frame's auto-fit behavior — the read inverse of
// AutoFit. A frame with no explicit autofit child reports AutoFitNone.
func (tf *TextFrame) AutoFitMode() AutoFitMode {
	if tf == nil || tf.body == nil || tf.body.BodyPr == nil {
		return AutoFitNone
	}
	switch bp := tf.body.BodyPr; {
	case bp.NormAutofit != nil:
		return AutoFitNormal
	case bp.SpAutoFit != nil:
		return AutoFitShape
	default:
		return AutoFitNone
	}
}

// VerticalAnchor returns the frame's vertical text anchor — the read inverse of
// Anchor. A frame with no explicit anchor reports AnchorTop (the OOXML default).
func (tf *TextFrame) VerticalAnchor() TextAnchor {
	if tf == nil || tf.body == nil || tf.body.BodyPr == nil {
		return AnchorTop
	}
	switch tf.body.BodyPr.Anchor {
	case "ctr":
		return AnchorMiddle
	case "b":
		return AnchorBottom
	default:
		return AnchorTop
	}
}

// MarginInsets returns the frame's internal insets (EMU) — the read inverse of
// Margins. An unset inset (the attribute is omitted) reads as 0.
func (tf *TextFrame) MarginInsets() (top, right, bottom, left EMU) {
	if tf == nil || tf.body == nil || tf.body.BodyPr == nil {
		return 0, 0, 0, 0
	}
	bp := tf.body.BodyPr
	emu := func(p *int) EMU {
		if p == nil {
			return 0
		}
		return EMU(*p)
	}
	return emu(bp.TIns), emu(bp.RIns), emu(bp.BIns), emu(bp.LIns)
}

// ============================================================================
// Read accessors (RFC §16) — the read inverse of the authoring API. Reopened
// runs surface resolved character properties: token typography/color resolve to
// concrete sizes and sRGB at write time (D-033), so the read model reports the
// resolved family / size / color, not the originating TypeRole / token.
// ============================================================================

// Paragraphs returns the frame's paragraphs in document order — the read-side
// enumerator. Each is addressable through the same Paragraph handle the builder
// hands out, so its read accessors (Runs / Alignment / Level / BulletStyle) and
// authoring methods share one type.
func (tf *TextFrame) Paragraphs() []*Paragraph {
	if tf == nil || tf.body == nil {
		return nil
	}
	ps := make([]*Paragraph, len(tf.body.Paragraphs))
	for i := range tf.body.Paragraphs {
		ps[i] = &Paragraph{tf: tf, idx: i}
	}
	return ps
}

// Runs returns the paragraph's text runs in document order — the read inverse of
// AddRun / AddHyperlink. Line breaks (AddBreak) carry no text and are not
// returned.
func (p *Paragraph) Runs() []*Run {
	xp := p.x()
	runs := make([]*Run, 0, len(xp.Content))
	for _, c := range xp.Content {
		if r, ok := c.(*slide.XTextRun); ok {
			runs = append(runs, &Run{tf: p.tf, run: r})
		}
	}
	return runs
}

// Alignment returns the paragraph's horizontal alignment — the read inverse of
// Align (AlignLeft when unset).
func (p *Paragraph) Alignment() Alignment {
	if pr := p.x().Pr; pr != nil {
		return alignFrom(pr.Alignment)
	}
	return AlignLeft
}

// Level returns the paragraph's outline/indent level (0-based) — the read
// inverse of Indent.
func (p *Paragraph) Level() int {
	if pr := p.x().Pr; pr != nil {
		return pr.Level
	}
	return 0
}

// LineHeight returns the paragraph's line spacing as a percent of single, or 0
// when none is set — the read inverse of ParagraphOpts.LineHeight (D-061).
func (p *Paragraph) LineHeight() float64 {
	if pr := p.x().Pr; pr != nil && pr.LnSpc != nil && pr.LnSpc.SpcPct != nil {
		return float64(pr.LnSpc.SpcPct.Val) / 1000.0
	}
	return 0
}

// BulletStyle returns the paragraph's bullet style — the read inverse of Bullet
// (BulletNone when unset or explicitly suppressed).
func (p *Paragraph) BulletStyle() BulletKind {
	pr := p.x().Pr
	if pr == nil {
		return BulletNone
	}
	switch {
	case pr.BuAutoNum != nil:
		return BulletNumber
	case pr.BuChar != nil:
		return bulletFromChar(pr.BuChar.Char)
	default:
		return BulletNone
	}
}

// Text returns the run's literal text.
func (r *Run) Text() string { return r.run.Text }

// Font returns the run's Latin typeface, or "" when unset (it inherits).
func (r *Run) Font() string {
	if pr := r.run.TextProperties; pr != nil && pr.Latin != nil {
		return pr.Latin.Typeface
	}
	return ""
}

// FontSize returns the run's font size in points, or 0 when unset.
func (r *Run) FontSize() float64 {
	if pr := r.run.TextProperties; pr != nil {
		return float64(pr.FontSize) / 100.0
	}
	return 0
}

// Bold reports whether the run is bold.
func (r *Run) Bold() bool {
	pr := r.run.TextProperties
	return pr != nil && pr.Bold == "1"
}

// Italic reports whether the run is italic.
func (r *Run) Italic() bool {
	pr := r.run.TextProperties
	return pr != nil && pr.Italic == "1"
}

// Underline returns the run's underline style — the read inverse of
// RunStyle.Underline.
func (r *Run) Underline() Underline {
	if pr := r.run.TextProperties; pr != nil {
		return underlineFrom(pr.Underline)
	}
	return UnderlineNone
}

// Strike returns the run's strikethrough style — the read inverse of
// RunStyle.Strike.
func (r *Run) Strike() Strike {
	if pr := r.run.TextProperties; pr != nil {
		return strikeFrom(pr.Strike)
	}
	return StrikeNone
}

// Baseline returns the run's baseline shift — the read inverse of
// RunStyle.BaselineRel.
func (r *Run) Baseline() BaselineShift {
	if pr := r.run.TextProperties; pr != nil {
		return baselineFrom(pr.Baseline)
	}
	return BaselineNone
}

// Tracking returns the run's letter-spacing in points (signed), or 0 when the
// run carries no explicit spacing — the read inverse of the resolved
// FontSpec.Tracking / RunStyle.Tracking (D-060).
func (r *Run) Tracking() float64 {
	if pr := r.run.TextProperties; pr != nil {
		return float64(pr.Spc) / 100.0
	}
	return 0
}

// Color returns the run's text color and true, or nil and false when the run has
// no explicit color (it inherits). A reopened color is a resolved literal (D-030).
func (r *Run) Color() (Color, bool) {
	pr := r.run.TextProperties
	if pr == nil || pr.SolidFill == nil {
		return nil, false
	}
	return colorFromSrgb(pr.SolidFill.SrgbClr), true
}

// Code reports whether the run is styled as inline code — detected by the
// subtle background tint pptx-go applies for code and nothing else (D-013).
func (r *Run) Code() bool {
	pr := r.run.TextProperties
	return pr != nil && pr.Highlight != nil
}
