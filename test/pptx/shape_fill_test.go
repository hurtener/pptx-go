package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// slideXMLWithShape builds a one-slide deck, runs fn to add a shape, and
// returns the emitted slide1.xml.
func slideXMLWithShape(t *testing.T, p *pptx.Presentation, fn func(s *pptx.Slide)) string {
	t.Helper()
	s := p.AddSlide()
	fn(s)
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return readZipPart(t, data, "ppt/slides/slide1.xml")
}

var unitBox = pptx.Box{X: 914400, Y: 914400, W: 2743200, H: 1371600}

func TestAddShape_SolidFill_Literal(t *testing.T) {
	xml := slideXMLWithShape(t, pptx.New(), func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, unitBox, pptx.WithFill(pptx.SolidFill(pptx.RGB("FF0000"))))
	})
	for _, want := range []string{
		`<a:prstGeom prst="rect">`,
		`<a:solidFill><a:srgbClr val="FF0000"/></a:solidFill>`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("missing %q in:\n%s", want, xml)
		}
	}
}

func TestAddShape_SolidFill_Alpha(t *testing.T) {
	xml := slideXMLWithShape(t, pptx.New(), func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeEllipse, unitBox, pptx.WithFill(pptx.SolidFill(pptx.RGBA("00FF00", 50000))))
	})
	if !strings.Contains(xml, `<a:srgbClr val="00FF00"><a:alpha val="50000"/></a:srgbClr>`) {
		t.Errorf("missing alpha-bearing solid fill in:\n%s", xml)
	}
}

func TestAddShape_NoFill(t *testing.T) {
	xml := slideXMLWithShape(t, pptx.New(), func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, unitBox, pptx.WithFill(pptx.NoFill()))
	})
	if !strings.Contains(xml, `<a:noFill/>`) {
		t.Errorf("missing <a:noFill/> in:\n%s", xml)
	}
}

func TestAddShape_Line(t *testing.T) {
	xml := slideXMLWithShape(t, pptx.New(), func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, unitBox, pptx.WithLine(pptx.Line{
			Width: pptx.Pt(2), Color: pptx.RGB("0000FF"), Dash: "dash",
		}))
	})
	for _, want := range []string{
		`<a:ln w="25400">`, // 2pt = 25400 EMU
		`<a:solidFill><a:srgbClr val="0000FF"/></a:solidFill>`,
		`<a:prstDash val="dash"/>`,
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("missing %q in:\n%s", want, xml)
		}
	}
}

// TestAddShape_TokenColor_ThemeSwap is acceptance criterion 7: the same builder
// input (SolidFill(TokenColor(ColorAccent))) re-renders in the active theme's
// palette — token, not literal.
func TestAddShape_TokenColor_ThemeSwap(t *testing.T) {
	render := func(accent pptx.RGB) string {
		th := pptx.DefaultTheme()
		th.Colors.Surfaces[pptx.ColorAccent] = accent
		return slideXMLWithShape(t, pptx.New(pptx.WithTheme(th)), func(s *pptx.Slide) {
			s.AddShape(pptx.ShapeRect, unitBox, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
		})
	}

	a := render("AA0000")
	b := render("00BB00")

	if !strings.Contains(a, `<a:srgbClr val="AA0000"/>`) {
		t.Errorf("theme A accent not resolved into fill:\n%s", a)
	}
	if !strings.Contains(b, `<a:srgbClr val="00BB00"/>`) {
		t.Errorf("theme B accent not resolved into fill:\n%s", b)
	}
	if strings.Contains(a, "00BB00") || strings.Contains(b, "AA0000") {
		t.Error("token color leaked across themes — not resolved against the active theme")
	}
}

// TestShape_Box confirms the opaque handle reports its EMU box.
func TestShape_Box(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	sh := s.AddShape(pptx.ShapeRect, unitBox)
	if got := sh.Box(); got != unitBox {
		t.Errorf("Shape.Box() = %+v, want %+v", got, unitBox)
	}
}
