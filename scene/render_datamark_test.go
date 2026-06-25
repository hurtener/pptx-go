package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestDataMark_Bar is R14.8 acceptance (bar): a progress bar fills to its value
// in soul colors as native rounded rects (track + fill), conformant, no warnings.
func TestDataMark_Bar(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bar",
		Nodes: []scene.SlideNode{scene.DataMark{Kind: scene.DataMarkBar, Value: 0.92, Label: "92%"}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("data mark bar: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, `prst="roundRect"`); n < 2 {
		t.Errorf("bar: want >=2 roundRects (track + fill), got %d", n)
	}
	if !strings.Contains(slide, "<a:t>92%</a:t>") {
		t.Errorf("bar: missing inline label")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("bar deck failed conformance:\n%s", rep)
	}
}

// TestDataMark_BarsAndSparkline renders a bar group and a sparkline (with an
// upward segment exercising the flipV path), conformant.
func TestDataMark_BarsAndSparkline(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "viz",
		Nodes: []scene.SlideNode{
			scene.DataMark{Kind: scene.DataMarkBars, Values: []float64{0.3, 0.6, 0.9, 0.5}},
			scene.DataMark{Kind: scene.DataMarkSparkline, Values: []float64{0.2, 0.8, 0.4, 1.0}},
		},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("bars/sparkline: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `prst="line"`) {
		t.Errorf("sparkline: missing line segments")
	}
	if !strings.Contains(slide, `flipV="true"`) {
		t.Errorf("sparkline: an upward segment should emit flipV")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("bars/sparkline deck failed conformance:\n%s", rep)
	}
}

// TestDataMark_InCard verifies a DataMark embeds in a card body without overflow
// (no warnings) — R14.8 acceptance.
func TestDataMark_InCard(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "card-bar",
		Nodes: []scene.SlideNode{scene.Card{Header: "Capacity", Body: []scene.SlideNode{
			scene.DataMark{Kind: scene.DataMarkBar, Value: 0.7, Label: "70%"},
		}}},
	}}}
	_, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("data mark in card: warnings: %+v", stats.Warnings)
	}
}

// TestDataMark_InvalidWarns verifies an out-of-range value fails Stage-1 validation.
func TestDataMark_InvalidWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bad",
		Nodes: []scene.SlideNode{scene.DataMark{Kind: scene.DataMarkBar, Value: 1.5}},
	}}}
	if _, err := scene.Render(pptx.New(), sc); err == nil {
		t.Errorf("data mark with value 1.5 should fail validation")
	}
}

// TestDataMark_Deterministic guards worker-count independence.
func TestDataMark_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "viz",
		Nodes: []scene.SlideNode{
			scene.DataMark{Kind: scene.DataMarkBar, Value: 0.92, Label: "92%"},
			scene.DataMark{Kind: scene.DataMarkBars, Values: []float64{0.3, 0.6, 0.9}},
			scene.DataMark{Kind: scene.DataMarkSparkline, Values: []float64{0.2, 0.8, 0.4, 1.0}},
		},
	}}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("data marks not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
