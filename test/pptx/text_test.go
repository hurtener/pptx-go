package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/pptx"
)

// TestTextFrame_StyledRun_RoundTrip is acceptance criterion 1: a TextFrame with
// multiple paragraphs and styled runs round-trips losslessly through pptx.Open
// and keeps the deck conformant.
func TestTextFrame_StyledRun_RoundTrip(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(6), H: pptx.In(2)})
	tf.Anchor(pptx.AnchorMiddle).AutoFit(pptx.AutoFitShape)

	h := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	h.AddRun("Title", pptx.RunStyle{TypeRole: pptx.TypeH1, Bold: true})

	body := tf.AddParagraph(pptx.ParagraphOpts{})
	body.AddRun("normal ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	body.AddRun("italic-underlined", pptx.RunStyle{
		TypeRole:  pptx.TypeBody,
		Italic:    true,
		Underline: pptx.UnderlineSingle,
		Strike:    pptx.StrikeSingle,
		Color:     pptx.RGB("DC2626"),
	})
	body.AddBreak()
	body.AddRun("second line", pptx.RunStyle{TypeRole: pptx.TypeBody})

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	xml := readZipPart(t, data, "ppt/slides/slide1.xml")
	// Spot-check the frame + run styling that must appear.
	for _, want := range []string{
		`anchor="ctr"`,
		`<a:spAutoFit/>`,
		`algn="ctr"`,
		`b="1"`,
		`i="1"`,
		`u="sng"`,
		`strike="sngStrike"`,
		`<a:srgbClr val="DC2626"/>`,
		`<a:br`,
		`<a:t>second line</a:t>`,
		`typeface="`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("slide1.xml missing %q in:\n%s", want, xml)
		}
	}

	rep, _ := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if !rep.OK() {
		t.Fatalf("text deck failed conformance:\n%s", rep)
	}

	// Model round-trip through Open: the slide repopulates with the two
	// paragraphs and their runs.
	r, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	sp := findTextShape(t, r.Slides()[0].Part())
	if n := len(sp.TextBody.Paragraphs); n != 2 {
		t.Fatalf("reopened paragraph count = %d, want 2", n)
	}
	if got := sp.TextBody.Paragraphs[0].Runs()[0].Text; got != "Title" {
		t.Errorf("reopened title run = %q, want %q", got, "Title")
	}
	// The break is preserved in document order (run, run, break, run).
	if len(sp.TextBody.Paragraphs[1].Content) != 4 {
		t.Errorf("reopened body content count = %d, want 4", len(sp.TextBody.Paragraphs[1].Content))
	}
}

// TestTextFrame_Bullets covers disc / numbered / checklist paragraphs.
func TestTextFrame_Bullets(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(6), H: pptx.In(3)})
	tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletDisc}).AddRun("disc", pptx.RunStyle{TypeRole: pptx.TypeBody})
	tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletNumber}).AddRun("num", pptx.RunStyle{TypeRole: pptx.TypeBody})
	tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletCheckbox}).AddRun("check", pptx.RunStyle{TypeRole: pptx.TypeBody})

	data, _ := p.WriteToBytes()
	xml := readZipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{
		`<a:buChar char="`,
		`<a:buAutoNum type="arabicPeriod"/>`,
		`marL="457200"`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("slide1.xml missing %q in:\n%s", want, xml)
		}
	}
}

// TestRunColor_ThemeSwap is acceptance criterion 5: a token run color resolves
// against the active theme set before authoring (token, not literal).
func TestRunColor_ThemeSwap(t *testing.T) {
	render := func(accent pptx.RGB) string {
		th := pptx.DefaultTheme()
		th.Colors.Surfaces[pptx.ColorAccent] = accent
		p := pptx.New(pptx.WithTheme(th))
		s := p.AddSlide()
		tf := s.AddTextFrame(pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(4), H: pptx.In(1)})
		tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("accent text",
			pptx.RunStyle{TypeRole: pptx.TypeBody, Color: pptx.TokenColor(pptx.ColorAccent)})
		data, err := p.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return readZipPart(t, data, "ppt/slides/slide1.xml")
	}

	a := render("AA0000")
	b := render("00BB00")
	if !strings.Contains(a, `<a:srgbClr val="AA0000"/>`) {
		t.Errorf("theme A accent not resolved into run color:\n%s", a)
	}
	if !strings.Contains(b, `<a:srgbClr val="00BB00"/>`) {
		t.Errorf("theme B accent not resolved into run color:\n%s", b)
	}
	if strings.Contains(a, "00BB00") || strings.Contains(b, "AA0000") {
		t.Error("token run color leaked across themes — not resolved against the active theme")
	}
}

// findTextShape returns the first shape carrying a text body.
func findTextShape(t *testing.T, part *slide.SlidePart) *slide.XSp {
	t.Helper()
	for _, c := range part.SpTree().Children {
		if sp, ok := c.(*slide.XSp); ok && sp.TextBody != nil && len(sp.TextBody.Paragraphs) > 0 {
			return sp
		}
	}
	t.Fatal("no text shape found on slide")
	return nil
}
