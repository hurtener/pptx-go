package pptx_test

import (
	"reflect"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// firstTextFrame reopens a deck built by build and returns the TextFrame of its
// first shape — the author → save → Open round trip every text read accessor is
// asserted over.
func firstTextFrame(t *testing.T, build func(s *pptx.Slide)) *pptx.TextFrame {
	t.Helper()
	shapes := reopenShapes(t, build)
	if len(shapes) == 0 {
		t.Fatal("reopened deck has no shapes")
	}
	tf, ok := shapes[0].TextFrame()
	if !ok {
		t.Fatal("first shape has no TextFrame")
	}
	return tf
}

// TestTextRead_ParagraphsRuns is PR#2 acceptance criterion 2 (enumeration): a
// reopened text shape yields its paragraphs → runs in document order with the
// authored text.
func TestTextRead_ParagraphsRuns(t *testing.T) {
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		p0 := f.AddParagraph(pptx.ParagraphOpts{})
		p0.AddRun("Hello ", pptx.RunStyle{TypeRole: pptx.TypeBody})
		p0.AddRun("world", pptx.RunStyle{TypeRole: pptx.TypeBody, Bold: true})
		p1 := f.AddParagraph(pptx.ParagraphOpts{})
		p1.AddRun("second", pptx.RunStyle{TypeRole: pptx.TypeBody})
	})
	paras := tf.Paragraphs()
	if len(paras) != 2 {
		t.Fatalf("Paragraphs() = %d, want 2", len(paras))
	}
	r0 := paras[0].Runs()
	if len(r0) != 2 {
		t.Fatalf("paragraph 0 Runs() = %d, want 2", len(r0))
	}
	if r0[0].Text() != "Hello " || r0[1].Text() != "world" {
		t.Errorf("paragraph 0 run texts = %q, %q", r0[0].Text(), r0[1].Text())
	}
	if r1 := paras[1].Runs(); len(r1) != 1 || r1[0].Text() != "second" {
		t.Errorf("paragraph 1 runs = %+v, want one %q", r1, "second")
	}
}

// TestTextRead_RunStyle is PR#2 acceptance criterion 2 (style + color): a
// reopened run reports the authored bold / italic / underline / strike /
// baseline / color, and the resolved family / size of its TypeRole.
func TestTextRead_RunStyle(t *testing.T) {
	color := pptx.RGB("CC0000")
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		p := f.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun("styled", pptx.RunStyle{
			TypeRole:    pptx.TypeBody,
			Bold:        true,
			Italic:      true,
			Underline:   pptx.UnderlineSingle,
			Strike:      pptx.StrikeSingle,
			BaselineRel: pptx.Superscript,
			Color:       color,
		})
	})
	run := tf.Paragraphs()[0].Runs()[0]

	if !run.Bold() {
		t.Error("Bold() = false, want true")
	}
	if !run.Italic() {
		t.Error("Italic() = false, want true")
	}
	if run.Underline() != pptx.UnderlineSingle {
		t.Errorf("Underline() = %v, want UnderlineSingle", run.Underline())
	}
	if run.Strike() != pptx.StrikeSingle {
		t.Errorf("Strike() = %v, want StrikeSingle", run.Strike())
	}
	if run.Baseline() != pptx.Superscript {
		t.Errorf("Baseline() = %v, want Superscript", run.Baseline())
	}
	if got, ok := run.Color(); !ok || !reflect.DeepEqual(got, color) {
		t.Errorf("Color() = %#v, %v; want %#v, true", got, ok, color)
	}

	// Family / size resolve from the TypeRole at write time (D-033); the read
	// model reports the resolved values.
	spec := pptx.DefaultTheme().ResolveType(pptx.TypeBody)
	if run.Font() != spec.Family {
		t.Errorf("Font() = %q, want %q", run.Font(), spec.Family)
	}
	if run.FontSize() != spec.Size {
		t.Errorf("FontSize() = %v, want %v", run.FontSize(), spec.Size)
	}
}

// TestTextRead_Hyperlink is PR#2 acceptance criterion 2 (hyperlink): a reopened
// hyperlink run resolves its relationship back to the authored URL.
func TestTextRead_Hyperlink(t *testing.T) {
	const url = "https://example.com/docs?a=1&b=2"
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		p := f.AddParagraph(pptx.ParagraphOpts{})
		p.AddHyperlink("click", url, pptx.RunStyle{TypeRole: pptx.TypeBody})
	})
	run := tf.Paragraphs()[0].Runs()[0]
	got, ok := run.Hyperlink()
	if !ok {
		t.Fatal("Hyperlink() ok = false, want true")
	}
	if got != url {
		t.Errorf("Hyperlink() = %q, want %q", got, url)
	}
	// A plain run reports no hyperlink.
	plain := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		f.AddParagraph(pptx.ParagraphOpts{}).AddRun("x", pptx.RunStyle{})
	})
	if _, ok := plain.Paragraphs()[0].Runs()[0].Hyperlink(); ok {
		t.Error("plain run Hyperlink() ok = true, want false")
	}
}

// TestTextRead_ParagraphProps is PR#2 acceptance criterion 2 (bullet + align +
// level): reopened paragraphs report the authored alignment, indent level, and
// bullet style.
func TestTextRead_ParagraphProps(t *testing.T) {
	bullets := []pptx.BulletKind{pptx.BulletDisc, pptx.BulletNumber, pptx.BulletCheckbox, pptx.BulletNone}
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		f.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter, Level: 2, Bullet: pptx.BulletDisc})
		f.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignRight, Bullet: pptx.BulletNumber})
		f.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignJustify, Bullet: pptx.BulletCheckbox})
		f.AddParagraph(pptx.ParagraphOpts{})
	})
	paras := tf.Paragraphs()
	if len(paras) != len(bullets) {
		t.Fatalf("Paragraphs() = %d, want %d", len(paras), len(bullets))
	}
	for i, want := range bullets {
		if got := paras[i].BulletStyle(); got != want {
			t.Errorf("paragraph %d BulletStyle() = %v, want %v", i, got, want)
		}
	}
	if got := paras[0].Alignment(); got != pptx.AlignCenter {
		t.Errorf("paragraph 0 Alignment() = %v, want AlignCenter", got)
	}
	if got := paras[0].Level(); got != 2 {
		t.Errorf("paragraph 0 Level() = %v, want 2", got)
	}
	if got := paras[1].Alignment(); got != pptx.AlignRight {
		t.Errorf("paragraph 1 Alignment() = %v, want AlignRight", got)
	}
	if got := paras[3].Alignment(); got != pptx.AlignLeft {
		t.Errorf("paragraph 3 Alignment() = %v, want AlignLeft (default)", got)
	}
}

// TestTextRead_Code is PR#2 acceptance criterion 2 (inline code): a run authored
// as inline code reopens with Code() true, a plain run with Code() false.
func TestTextRead_Code(t *testing.T) {
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		p := f.AddParagraph(pptx.ParagraphOpts{})
		p.AddRun("plain ", pptx.RunStyle{TypeRole: pptx.TypeBody})
		p.AddRun("code", pptx.RunStyle{TypeRole: pptx.TypeBody, Code: true})
	})
	runs := tf.Paragraphs()[0].Runs()
	if runs[0].Code() {
		t.Error("plain run Code() = true, want false")
	}
	if !runs[1].Code() {
		t.Error("code run Code() = false, want true")
	}
}

// TestTextRead_NoTextFrame is PR#2 acceptance criterion 2 (negative): a shape
// with no text body reports TextFrame ok = false.
func TestTextRead_NoTextFrame(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, fxBox)
	})
	if _, ok := shapes[0].TextFrame(); ok {
		t.Error("shapeless-text Rect TextFrame() ok = true, want false")
	}
}
