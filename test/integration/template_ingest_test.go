package integration

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// zipPart returns the named part's bytes from a .pptx byte slice.
func zipPart(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, _ := f.Open()
			defer func() { _ = rc.Close() }()
			b, _ := io.ReadAll(rc)
			return string(b)
		}
	}
	t.Fatalf("part %s not found", name)
	return ""
}

// TestTemplateIngest_EndToEnd closes the template → builder → scene seam opened
// by Phase 09 (Deps: Phases 02/03/05). It opens a brand kit, seeds a new
// presentation from it (FromTemplate), renders a scene against a brand theme and
// a layout map, then reopens the result and asserts the brand identity survived
// — all through real OPC writes + encoding/xml decode, under conformance.
func TestTemplateIngest_EndToEnd(t *testing.T) {
	// A brand kit: any valid deck, reopened so its master/layout registry and
	// theme are populated (the scaffold contributes a "Blank" layout).
	seed, err := pptx.New().WriteToBytes()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	brand, err := pptx.NewFromBytes(seed)
	if err != nil {
		t.Fatalf("open brand: %v", err)
	}
	if !brand.HasLayout("Blank") {
		t.Fatal("brand kit missing its Blank layout")
	}

	// Seed a new presentation from the brand, then render a scene against an
	// explicit brand theme + layout map.
	brandTheme := pptx.NewTheme(pptx.WithName("Acme"), pptx.WithAccent("AB12CD"))
	pres := pptx.New(pptx.FromTemplate(brand))

	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "cover", Layout: scene.LayoutBlank, Nodes: []scene.SlideNode{scene.SectionDivider{Label: "Acme FY25"}}},
	}}
	stats, err := scene.Render(pres, sc,
		scene.WithTheme(brandTheme),
		scene.WithLayoutMap(scene.LayoutMap{scene.LayoutBlank: "Blank"}),
	)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(stats.Warnings) != 0 {
		t.Fatalf("unexpected warnings (the mapped layout exists): %+v", stats.Warnings)
	}

	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	rep, err := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"}})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("template-ingested deck failed conformance:\n%s", rep)
	}

	// The brand accent reached the slide, and the deck reopens cleanly.
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "AB12CD") {
		t.Errorf("brand accent not present in the rendered slide:\n%s", slideXML)
	}
	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if reopened.SlideCount() != 1 {
		t.Errorf("reopened slide count = %d, want 1", reopened.SlideCount())
	}
}
