package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestDisplayFont_Resolution is acceptance criterion 1: DisplayFont/HeadingFont/
// BodyFont land on TypeDisplay / H2-H5 / body respectively.
func TestDisplayFont_Resolution(t *testing.T) {
	th := pptx.NewTheme(pptx.WithFonts("Heading", "Body"), pptx.WithDisplayFont("Display"))
	if got := th.ResolveType(pptx.TypeDisplay).Family; got != "Display" {
		t.Errorf("TypeDisplay family = %q, want Display", got)
	}
	if got := th.ResolveType(pptx.TypeH2).Family; got != "Heading" {
		t.Errorf("TypeH2 family = %q, want Heading", got)
	}
	if got := th.ResolveType(pptx.TypeBody).Family; got != "Body" {
		t.Errorf("TypeBody family = %q, want Body", got)
	}
}

// TestDisplayFont_OmittedInheritsHeading is acceptance criterion 2: with no
// DisplayFont, TypeDisplay inherits HeadingFont (byte-identical to a 2-font theme).
func TestDisplayFont_OmittedInheritsHeading(t *testing.T) {
	th := pptx.NewTheme(pptx.WithFonts("Heading", "Body"))
	if got := th.ResolveType(pptx.TypeDisplay).Family; got != "Heading" {
		t.Errorf("without DisplayFont, TypeDisplay = %q, want HeadingFont Heading", got)
	}
	if d := pptx.DefaultTheme(); d.ResolveType(pptx.TypeDisplay).Family != d.HeadingFont {
		t.Errorf("default TypeDisplay = %q, want HeadingFont %q", d.ResolveType(pptx.TypeDisplay).Family, d.HeadingFont)
	}
}

// TestDisplayFont_OrderIndependent: WithDisplayFont before or after WithFonts
// yields the same resolution (TypeDisplay = Display, not clobbered by WithFonts).
func TestDisplayFont_OrderIndependent(t *testing.T) {
	after := pptx.NewTheme(pptx.WithFonts("H", "B"), pptx.WithDisplayFont("D"))
	before := pptx.NewTheme(pptx.WithDisplayFont("D"), pptx.WithFonts("H", "B"))
	for _, th := range []*pptx.Theme{after, before} {
		if th.ResolveType(pptx.TypeDisplay).Family != "D" || th.ResolveType(pptx.TypeH1).Family != "H" {
			t.Errorf("order dependence: display=%q h1=%q, want D/H",
				th.ResolveType(pptx.TypeDisplay).Family, th.ResolveType(pptx.TypeH1).Family)
		}
	}
}

// TestDisplayFont_Render: a display run renders with the DisplayFont typeface.
func TestDisplayFont_Render(t *testing.T) {
	th := pptx.NewTheme(pptx.WithFonts("Heading Sans", "Body Sans"), pptx.WithDisplayFont("Playfair Display"))
	p := pptx.New(pptx.WithTheme(th))
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 914400, W: 5000000, H: 1500000})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Hero Title", pptx.RunStyle{TypeRole: pptx.TypeDisplay})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if !strings.Contains(trackSlideXML(t, data), `typeface="Playfair Display"`) {
		t.Error("a TypeDisplay run should render with the DisplayFont typeface")
	}
}
