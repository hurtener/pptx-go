package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// TestFootnotes is R14.12 acceptance: a slide with 2 source lines + a superscript
// marker on a Stat renders the sources in the bottom band (muted) and the marker
// as a superscript, with no warnings.
func TestFootnotes(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "cited",
		Nodes: []scene.SlideNode{
			scene.Stat{Value: "92%", Label: "adoption"},
			scene.Prose{Paragraphs: []scene.RichText{{
				{Text: "Adoption is up"},
				{Text: "1", Style: scene.RunStyle{Superscript: true}},
			}}},
		},
		Footnotes: []scene.RichText{
			rt("Source: internal telemetry, 2026."),
			rt("1. Measured across active accounts."),
		},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("footnotes: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "Source: internal telemetry, 2026.") {
		t.Errorf("footnotes: source line missing:\n%s", slide)
	}
	if !strings.Contains(slide, "baseline=") {
		t.Errorf("footnotes: superscript marker should emit a baseline shift")
	}
}

// TestFootnotes_EmptyByteIdentical verifies a slide with no Footnotes is
// byte-identical to one without the field.
func TestFootnotes_EmptyByteIdentical(t *testing.T) {
	mk := func(fn []scene.RichText) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID:        "s",
			Nodes:     []scene.SlideNode{scene.Hero{Title: "Title"}},
			Footnotes: fn,
		}}}
	}
	a, _ := render(t, mk(nil))
	b, _ := render(t, scene.Scene{Slides: []scene.SceneSlide{{ID: "s", Nodes: []scene.SlideNode{scene.Hero{Title: "Title"}}}}})
	if !bytes.Equal(a, b) {
		t.Errorf("nil Footnotes not byte-identical to absent field (%d vs %d bytes)", len(a), len(b))
	}
}

// TestFootnotes_CapWarns verifies footnotes past the region cap are dropped + warned.
func TestFootnotes_CapWarns(t *testing.T) {
	fns := make([]scene.RichText, 6)
	for i := range fns {
		fns[i] = rt("source line")
	}
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s", Nodes: []scene.SlideNode{scene.Hero{Title: "T"}}, Footnotes: fns}}}
	_, stats := render(t, sc)
	if len(stats.Warnings) == 0 {
		t.Errorf("6 footnotes (> cap): expected a warning")
	}
}

// TestFootnotes_Deterministic guards worker-count independence.
func TestFootnotes_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:        "s",
		Nodes:     []scene.SlideNode{scene.Hero{Title: "T"}},
		Footnotes: []scene.RichText{rt("Source: x"), rt("Source: y")},
	}}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("footnotes not deterministic (%d vs %d bytes)", len(seq), len(par))
	}
}
