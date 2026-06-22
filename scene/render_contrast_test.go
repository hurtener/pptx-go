package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for R11.2 auto-contrast (D-082).

// headerRun returns the <a:r> run substring whose text is exactly text, so a test
// can inspect that run's rPr (the color fill) in isolation.
func headerRun(t *testing.T, xml, text string) string {
	t.Helper()
	marker := "<a:t>" + text + "</a:t>"
	idx := strings.Index(xml, marker)
	if idx < 0 {
		t.Fatalf("run text %q not found in slide XML", text)
	}
	start := strings.LastIndex(xml[:idx], "<a:r>")
	if start < 0 {
		t.Fatalf("no <a:r> before %q", text)
	}
	return xml[start : idx+len(marker)]
}

// TestCardHeader_AutoContrast_DarkVariant: on a dark-variant slide the card header
// run carries an explicit white (FFFFFF) color, so it is never black-on-dark.
func TestCardHeader_AutoContrast_DarkVariant(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:      "s1",
		Variant: scene.VariantDark,
		Nodes:   []scene.SlideNode{scene.Card{Header: "Contrast", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
	}}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	run := headerRun(t, xml, "Contrast")
	if !strings.Contains(run, `val="FFFFFF"`) {
		t.Errorf("dark-variant card header should be white; run = %s", run)
	}
}

// TestCardHeader_AutoContrast_LightByteIdentical: on a light slide with a light
// card fill, the header run carries NO explicit color fill (it inherits the dark
// default) — the byte-identical pre-R11.2 behavior.
func TestCardHeader_AutoContrast_LightByteIdentical(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "s1",
		Nodes: []scene.SlideNode{scene.Card{Header: "Contrast", Fill: scene.ColorSurface, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
	}}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	run := headerRun(t, xml, "Contrast")
	if strings.Contains(run, "<a:solidFill>") {
		t.Errorf("light-card header should have no explicit color (byte-identical); run = %s", run)
	}
}

// TestCardHeader_AutoContrast_DarkFill: a dark card Fill on a *light* slide still
// flips the header to a light color — contrast follows the surface, not the
// variant.
func TestCardHeader_AutoContrast_DarkFill(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "s1",
		Nodes: []scene.SlideNode{scene.Card{Header: "OnDark", Fill: scene.ColorAccent, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
	}}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	run := headerRun(t, xml, "OnDark")
	if !strings.Contains(run, `val="FFFFFF"`) {
		t.Errorf("dark-fill card header should be white on a light slide too; run = %s", run)
	}
}

// TestCardEyebrow_AccentFallback: an eyebrow on a same-hue (accent) header band
// drops its near-invisible accent tint; on a normal light card it keeps the accent
// (byte-identical). Verified via the rendered eyebrow run color.
func TestCardEyebrow_AccentFallback(t *testing.T) {
	accent := scene.ColorAccent
	// Eyebrow on an accent-colored header band: accent-on-accent fails contrast →
	// must NOT be the accent color (2563EB).
	band := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "s1",
		Nodes: []scene.SlideNode{scene.Card{Eyebrow: "VISION", Header: "X", HeaderFill: &accent, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
	}}}
	data, _ := render(t, band)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	run := headerRun(t, xml, "VISION")
	if strings.Contains(run, `val="2563EB"`) {
		t.Errorf("eyebrow on a same-hue band must drop the accent tint; run = %s", run)
	}

	// Eyebrow on a normal light card: keeps the accent (byte-identical).
	plain := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "s1",
		Nodes: []scene.SlideNode{scene.Card{Eyebrow: "VISION", Header: "X", Fill: scene.ColorSurface, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
	}}}
	pdata, _ := render(t, plain)
	pxml := zipPart(t, pdata, "ppt/slides/slide1.xml")
	prun := headerRun(t, pxml, "VISION")
	if !strings.Contains(prun, `val="2563EB"`) {
		t.Errorf("eyebrow on a light card should keep the accent tint (byte-identical); run = %s", prun)
	}
}
