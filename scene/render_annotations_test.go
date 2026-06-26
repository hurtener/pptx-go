package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func shotResolver() scene.AssetResolver {
	return scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "shot" {
			return pngOf(1600, 900), "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
}

func annotatedScene() scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID: "annotated",
		Nodes: []scene.SlideNode{scene.Image{
			AssetID: "asset://shot",
			Annotations: &scene.ImageAnnotations{
				Pins: []scene.ImagePin{
					{X: 0.2, Y: 0.3, Label: "1", Caption: "The nav bar"},
					{X: 0.8, Y: 0.6, Label: "2", Caption: "The CTA", AccentIndex: 1},
					{X: 0.95, Y: 0.1, Label: "3"},
				},
				Highlights: []scene.ImageHighlight{{X: 0.1, Y: 0.2, W: 0.3, H: 0.2}},
			},
		}},
	}}}
}

// TestImageAnnotations is R14.17 acceptance: an image with numbered pins + leader
// captions + a highlight box renders each pin at its coordinate, conformant.
func TestImageAnnotations(t *testing.T) {
	data, stats := render(t, annotatedScene(), scene.WithAssetResolver(shotResolver()))
	if len(stats.Warnings) != 0 {
		t.Errorf("annotations: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, `prst="ellipse"`); n < 3 {
		t.Errorf("annotations: want >=3 pin discs, got %d", n)
	}
	if !strings.Contains(slide, "<a:t>The nav bar</a:t>") {
		t.Errorf("annotations: missing leader caption")
	}
	if !strings.Contains(slide, "<p:pic>") {
		t.Errorf("annotations: missing the base image")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("annotated deck failed conformance:\n%s", rep)
	}
}

// TestImageAnnotations_InvalidWarns verifies an out-of-range pin fails validation.
func TestImageAnnotations_InvalidWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bad",
		Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://shot", Annotations: &scene.ImageAnnotations{Pins: []scene.ImagePin{{X: 1.5, Y: 0.5, Label: "1"}}}}},
	}}}
	if _, err := scene.Render(pptx.New(), sc, scene.WithAssetResolver(shotResolver())); err == nil {
		t.Errorf("pin at x=1.5 should fail validation")
	}
}

// TestImageAnnotations_NilByteIdentical verifies a nil Annotations is
// byte-identical to an image without the field.
func TestImageAnnotations_NilByteIdentical(t *testing.T) {
	mk := func(a *scene.ImageAnnotations) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{ID: "i", Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://shot", Annotations: a}}}}}
	}
	with, _ := render(t, mk(nil), scene.WithAssetResolver(shotResolver()))
	plain, _ := render(t, scene.Scene{Slides: []scene.SceneSlide{{ID: "i", Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://shot"}}}}}, scene.WithAssetResolver(shotResolver()))
	if !bytes.Equal(with, plain) {
		t.Errorf("nil Annotations not byte-identical (%d vs %d bytes)", len(with), len(plain))
	}
}

// TestImageAnnotations_Deterministic guards worker-count independence.
func TestImageAnnotations_Deterministic(t *testing.T) {
	sc := annotatedScene()
	seq := renderBytes(t, sc, scene.WithAssetResolver(shotResolver()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithAssetResolver(shotResolver()), scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("annotations not deterministic (%d vs %d bytes)", len(seq), len(par))
	}
}
