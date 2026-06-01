package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
	"github.com/hurtener/pptx-go/scene/icons"
)

const validIconSVG = `<svg viewBox="0 0 24 24"><path d="M12 2 L22 22 L2 22 Z"/></svg>`
const arcIconSVG = `<svg viewBox="0 0 24 24"><path d="M0 0 A5 5 0 0 1 10 10"/></svg>`

// TestWithIconExtension_Valid is acceptance criterion 3: a valid caller icon
// registers without error.
func TestWithIconExtension_Valid(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "s", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}},
	}}
	_, err := scene.Render(pptx.New(), sc, scene.WithIconExtension("brand-mark", []byte(validIconSVG)))
	if err != nil {
		t.Fatalf("Render with a valid icon extension: %v", err)
	}
}

// TestWithIconExtension_Invalid is acceptance criterion 4: an icon SVG outside
// the translator subset fails at registration (a Stage-1 render error that names
// the icon), not silently at compose.
func TestWithIconExtension_Invalid(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "s", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}},
	}}
	_, err := scene.Render(pptx.New(), sc, scene.WithIconExtension("bad", []byte(arcIconSVG)))
	if err == nil {
		t.Fatal("Render accepted an invalid icon extension; want a Stage-1 error")
	}
	if !strings.Contains(err.Error(), "bad") || !strings.Contains(err.Error(), "arc") {
		t.Errorf("error %q should name the icon and the violation", err)
	}
}

// TestValidateIcon checks the public registration-time validator.
func TestValidateIcon(t *testing.T) {
	if err := scene.ValidateIcon([]byte(validIconSVG)); err != nil {
		t.Errorf("ValidateIcon rejected a valid icon: %v", err)
	}
	if err := scene.ValidateIcon([]byte(arcIconSVG)); err == nil {
		t.Error("ValidateIcon accepted an arc icon")
	}
}

// TestCuratedIconsValidateAtSceneLayer is a belt-and-suspenders check that the
// curated set is reachable and valid from the scene layer.
func TestCuratedIconsValidateAtSceneLayer(t *testing.T) {
	for _, name := range icons.Curated().Names() {
		svg, _ := icons.Curated().Lookup(name)
		if err := scene.ValidateIcon(svg); err != nil {
			t.Errorf("curated icon %q invalid at scene layer: %v", name, err)
		}
	}
}
