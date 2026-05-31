package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestRender_WithTheme_BrandAccent is acceptance criterion 3: a scene rendered
// with WithTheme(brand) emits the brand's accent color. A section_divider fills
// its background with the ColorAccent token, so the resolved accent must appear
// as an srgbClr in the slide XML.
func TestRender_WithTheme_BrandAccent(t *testing.T) {
	brand := pptx.NewTheme(pptx.WithName("Acme"), pptx.WithAccent("AB12CD"))

	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "cover", Nodes: []scene.SlideNode{scene.SectionDivider{Label: "Welcome"}}},
	}}

	pres := pptx.New()
	if _, err := scene.Render(pres, sc, scene.WithTheme(brand)); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "AB12CD") {
		t.Errorf("slide does not carry the brand accent AB12CD:\n%s", slideXML)
	}
	// The default accent must not leak through when a brand theme is applied.
	if strings.Contains(slideXML, "2563EB") {
		t.Errorf("slide carries the default accent instead of the brand accent:\n%s", slideXML)
	}
}

// TestRender_WithLayoutMap is acceptance criterion 4: a mapped layout the
// template defines resolves silently; a mapped layout it lacks falls back to
// blank and records exactly one LayoutWarning (no error).
func TestRender_WithLayoutMap(t *testing.T) {
	// A deck reopened from bytes has its master/layout registry populated; the
	// scaffold contributes a layout named "Blank".
	seed, err := pptx.New().WriteToBytes()
	if err != nil {
		t.Fatalf("seed WriteToBytes: %v", err)
	}
	brand, err := pptx.NewFromBytes(seed)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if !brand.HasLayout("Blank") {
		t.Fatal("expected the template to define a Blank layout")
	}

	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "ok", Layout: scene.LayoutBlank, Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("a")}}}},
		{ID: "miss", Layout: scene.LayoutCover, Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}},
	}}

	pres := pptx.New(pptx.FromTemplate(brand))
	stats, err := scene.Render(pres, sc, scene.WithLayoutMap(scene.LayoutMap{
		scene.LayoutBlank: "Blank",          // resolves
		scene.LayoutCover: "No Such Layout", // missing → warn + blank fallback
	}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(stats.Warnings) != 1 {
		t.Fatalf("warnings = %d, want 1 (the missing layout): %+v", len(stats.Warnings), stats.Warnings)
	}
	if stats.Warnings[0].SlideID != "miss" {
		t.Errorf("warning slide = %q, want miss", stats.Warnings[0].SlideID)
	}
	if !strings.Contains(stats.Warnings[0].Message, "No Such Layout") {
		t.Errorf("warning message = %q, want it to name the missing layout", stats.Warnings[0].Message)
	}
}
