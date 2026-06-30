package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// heroDarkTheme registers a deep-navy radial "heroDark" brand gradient with RGB
// stops (variant-independent exact hues).
func heroDarkTheme() *pptx.Theme {
	return pptx.NewTheme(pptx.WithGradient("heroDark", pptx.GradientSpec{
		Stops: []pptx.GradientStop{
			{Pos: 0, Color: pptx.RGB("1E293B")},
			{Pos: 1, Color: pptx.RGB("0A0E1A")},
		},
		Radial: true,
	}))
}

func namedGradientScene() scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID:         "hero",
		Background: scene.Background{Kind: scene.BackgroundGradient, GradientName: "heroDark"},
		Nodes:      []scene.SlideNode{scene.Heading{Text: rt("The Shift"), Level: 1}},
	}}}
}

// TestNamedGradient_BrandWashRendered is the R8.5 acceptance: a named brand
// gradient renders its exact stop hues into the slide as a radial gradient fill.
func TestNamedGradient_BrandWashRendered(t *testing.T) {
	data, stats := render(t, namedGradientScene(), scene.WithTheme(heroDarkTheme()))
	if len(stats.Warnings) != 0 {
		t.Errorf("named gradient: unexpected warnings: %+v", stats.Warnings)
	}
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, hue := range []string{"1E293B", "0A0E1A"} {
		if !strings.Contains(slideXML, hue) {
			t.Errorf("slide missing brand gradient hue %s:\n%s", hue, slideXML)
		}
	}
	// Radial spec → a circular path gradient.
	if !strings.Contains(slideXML, "<a:gradFill") || !strings.Contains(slideXML, "path=\"circle\"") {
		t.Errorf("named radial gradient did not emit a circular gradFill:\n%s", slideXML)
	}
}

// TestNamedGradient_LinearAngle verifies a non-radial named gradient renders as a
// linear gradient at the spec's angle (a <a:lin> element, no circular path).
func TestNamedGradient_LinearAngle(t *testing.T) {
	th := pptx.NewTheme(pptx.WithGradient("brandLinear", pptx.GradientSpec{
		Stops: []pptx.GradientStop{{Pos: 0, Color: pptx.RGB("FF0000")}, {Pos: 1, Color: pptx.RGB("0000FF")}},
		Angle: 90,
	}))
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		Background: scene.Background{Kind: scene.BackgroundGradient, GradientName: "brandLinear"},
		Nodes:      []scene.SlideNode{scene.Heading{Text: rt("Linear"), Level: 1}},
	}}}
	data, _ := render(t, sc, scene.WithTheme(th))
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "<a:lin") || strings.Contains(slideXML, "path=\"circle\"") {
		t.Errorf("named linear gradient should emit <a:lin>, not a circular path:\n%s", slideXML)
	}
}

// TestNamedGradient_MissingWarns verifies that requesting an unregistered name
// records a LayoutWarning and skips the fill (RFC §10.2 degrade), not a panic.
func TestNamedGradient_MissingWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:         "miss",
		Background: scene.Background{Kind: scene.BackgroundGradient, GradientName: "nope"},
		Nodes:      []scene.SlideNode{scene.Heading{Text: rt("x"), Level: 1}},
	}}}
	_, stats := render(t, sc) // default theme registers no gradients
	var found bool
	for _, w := range stats.Warnings {
		if strings.Contains(w.Message, "nope") && strings.Contains(w.Message, "not registered") {
			found = true
		}
	}
	if !found {
		t.Errorf("missing named gradient did not warn: %+v", stats.Warnings)
	}
}

// TestNamedGradient_InvalidStopsWarn verifies a registered spec with out-of-range
// / non-ascending stops warns and skips (not a panic).
func TestNamedGradient_InvalidStopsWarn(t *testing.T) {
	th := pptx.NewTheme(pptx.WithGradient("bad", pptx.GradientSpec{
		Stops: []pptx.GradientStop{{Pos: 0.8, Color: pptx.RGB("111111")}, {Pos: 0.2, Color: pptx.RGB("222222")}}, // descending
	}))
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:         "bad",
		Background: scene.Background{Kind: scene.BackgroundGradient, GradientName: "bad"},
		Nodes:      []scene.SlideNode{scene.Heading{Text: rt("x"), Level: 1}},
	}}}
	_, stats := render(t, sc, scene.WithTheme(th))
	var found bool
	for _, w := range stats.Warnings {
		if strings.Contains(w.Message, "bad") && strings.Contains(w.Message, "invalid stops") {
			found = true
		}
	}
	if !found {
		t.Errorf("invalid named gradient stops did not warn: %+v", stats.Warnings)
	}
}

// TestNamedGradient_EmptyByteIdentical is the byte-identity guard: a gradient
// background with no GradientName (using the legacy 2-role pair) is byte-for-byte
// identical whether or not the theme registers unrelated named gradients (R8.5).
func TestNamedGradient_EmptyByteIdentical(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		Background: scene.Background{
			Kind:     scene.BackgroundGradient,
			Gradient: [2]pptx.ColorRole{pptx.ColorAccent, pptx.ColorAccentAlt},
			Angle:    90,
		},
		Nodes: []scene.SlideNode{scene.Heading{Text: rt("Legacy"), Level: 1}},
	}}}
	plain := renderBytes(t, sc)
	withGradients := renderBytes(t, sc, scene.WithTheme(heroDarkTheme())) // registers heroDark, unused here
	if !bytes.Equal(plain, withGradients) {
		t.Errorf("legacy 2-role gradient not byte-identical when a named gradient is registered but unused (%d vs %d bytes)", len(withGradients), len(plain))
	}
}

// TestNamedGradient_Determinism proves a named-gradient deck renders
// byte-identically across worker counts (RFC §10.1).
func TestNamedGradient_Determinism(t *testing.T) {
	sc := namedGradientScene()
	seq := renderBytes(t, sc, scene.WithTheme(heroDarkTheme()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithTheme(heroDarkTheme()), scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Errorf("named-gradient deck not deterministic across workers (%d vs %d bytes)", len(seq), len(par))
	}
}
