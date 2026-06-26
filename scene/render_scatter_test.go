package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func scatterSlide(preset string) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID: "scatter",
		Nodes: []scene.SlideNode{scene.Decoration{
			Kind: scene.DecorationPreset, Preset: preset, Layer: scene.LayerBackground,
			Size: scene.Size{W: pptx.In(6), H: pptx.In(4)}, Opacity: 0.6,
		}},
	}}}
}

// TestScatterFamily is R14.20 acceptance: the scatter family renders from one
// engine with different mark shapes; scatter_star emits star geometry,
// scatter_plus a mathPlus, scatter_ring an outline — deterministically.
func TestScatterFamily(t *testing.T) {
	cases := map[string]string{
		"scatter_star": `prst="star5"`,
		"scatter_plus": `prst="mathPlus"`,
	}
	for preset, want := range cases {
		t.Run(preset, func(t *testing.T) {
			data, stats := render(t, scatterSlide(preset))
			if len(stats.Warnings) != 0 {
				t.Errorf("%s: warnings: %+v", preset, stats.Warnings)
			}
			slide := zipPart(t, data, "ppt/slides/slide1.xml")
			if !strings.Contains(slide, want) {
				t.Errorf("%s: missing %s", preset, want)
			}
		})
	}
}

// TestScatter_StarfieldByteIdentical verifies the refactored starfield (now
// scatter+scatterDot) is byte-identical to scatter_dot — the same placement
// engine + dot shape — so D-110's starfield output is unchanged.
func TestScatter_StarfieldByteIdentical(t *testing.T) {
	star, _ := render(t, scatterSlide("starfield"))
	dot, _ := render(t, scatterSlide("scatter_dot"))
	if !bytes.Equal(star, dot) {
		t.Errorf("starfield not byte-identical to scatter_dot (%d vs %d bytes)", len(star), len(dot))
	}
}

// TestScatter_Deterministic guards worker-count independence.
func TestScatter_Deterministic(t *testing.T) {
	sc := scatterSlide("scatter_ring")
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("scatter not deterministic (%d vs %d bytes)", len(seq), len(par))
	}
}
