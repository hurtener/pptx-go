package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for the TwoColumn column-bridge (R12.8, D-101): a top bridge
// with a label pill, the seam default's stability, validation, and determinism.

func bridgeScene(pos scene.JoinPosition, label string) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{
		{ID: "paths", Nodes: []scene.SlideNode{
			scene.TwoColumn{
				Join: scene.JoinBadge, JoinLabel: label, JoinPosition: pos,
				Left:  []scene.SlideNode{scene.Heading{Text: rt("Build it"), Level: 2}, scene.Prose{Paragraphs: []scene.RichText{rt("internally")}}},
				Right: []scene.SlideNode{scene.Heading{Text: rt("Buy it"), Level: 2}, scene.Prose{Paragraphs: []scene.RichText{rt("externally")}}},
			},
		}},
	}}
}

// TestColumnBridge_TopRenders: a top bridge emits its label pill (roundRect + text) and
// the columns' content.
func TestColumnBridge_TopRenders(t *testing.T) {
	data, _ := render(t, bridgeScene(scene.JoinTopBridge, "One agent, two ways"))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "One agent, two ways") {
		t.Error("bridge label missing")
	}
	if !strings.Contains(xml, `prst="roundRect"`) {
		t.Error("bridge label pill (roundRect) missing")
	}
	for _, w := range []string{"Build it", "Buy it"} {
		if !strings.Contains(xml, w) {
			t.Errorf("column content %q missing", w)
		}
	}
}

// TestColumnBridge_LabelIntact: the bridge label renders as a single run (no mid-word
// wrap) — the run text is present verbatim.
func TestColumnBridge_LabelIntact(t *testing.T) {
	data, _ := render(t, bridgeScene(scene.JoinTopBridge, "One agent"))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:t>One agent</a:t>") {
		t.Error("bridge label should be one intact run, not split mid-word")
	}
}

// TestColumnBridge_Validation: an out-of-range JoinPosition fails Stage-1.
func TestColumnBridge_Validation(t *testing.T) {
	if err := scene.ValidateScene(bridgeScene(scene.JoinPosition(99), "x")); err == nil {
		t.Error("out-of-range join position passed validation")
	}
}

// TestColumnBridge_Deterministic: a top and a bottom bridge render byte-identically.
func TestColumnBridge_Deterministic(t *testing.T) {
	for _, pos := range []scene.JoinPosition{scene.JoinTopBridge, scene.JoinBottomBridge} {
		sc := bridgeScene(pos, "One agent, purpose-built")
		seq := renderBytes(t, sc, scene.WithWorkers(1))
		par := renderBytes(t, sc, scene.WithWorkers(8))
		if string(seq) != string(par) {
			t.Fatalf("bridge pos %d: parallel (%d bytes) differs from sequential (%d bytes)", pos, len(par), len(seq))
		}
	}
}
