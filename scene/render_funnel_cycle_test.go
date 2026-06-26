package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestFunnel is R14.11 acceptance (funnel): a 4-stage funnel tapers monotonically
// with values labeled, native rects, conformant, no warnings.
func TestFunnel(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "funnel", Nodes: []scene.SlideNode{scene.Funnel{Stages: []scene.FunnelStage{
		{Label: "Visitors", Value: "100k"}, {Label: "Signups", Value: "12k"}, {Label: "Trials", Value: "3k"}, {Label: "Paid", Value: "800"},
	}}}}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("funnel: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, `prst="roundRect"`); n < 4 {
		t.Errorf("funnel: want >=4 bands, got %d", n)
	}
	if !strings.Contains(slide, "<a:t>100k</a:t>") {
		t.Errorf("funnel: missing value label")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("funnel deck failed conformance:\n%s", rep)
	}
}

// TestCycle is R14.11 acceptance (cycle): a 5-stage cycle places stage cards on a
// ring with directional connectors, conformant, no warnings.
func TestCycle(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "cycle", Nodes: []scene.SlideNode{scene.Cycle{Stages: []scene.CycleStage{
		{Label: "Plan"}, {Label: "Build"}, {Label: "Measure"}, {Label: "Learn"}, {Label: "Iterate"},
	}}}}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("cycle: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, `prst="roundRect"`); n < 5 {
		t.Errorf("cycle: want >=5 stage cards, got %d", n)
	}
	if !strings.Contains(slide, `prst="chevron"`) {
		t.Errorf("cycle: missing directional chevron arrowheads")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("cycle deck failed conformance:\n%s", rep)
	}
}

// TestFunnelCycle_InvalidWarns verifies empty funnel/cycle fail validation.
func TestFunnelCycle_InvalidWarns(t *testing.T) {
	for _, n := range []scene.SlideNode{scene.Funnel{}, scene.Cycle{}} {
		sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "x", Nodes: []scene.SlideNode{n}}}}
		if _, err := scene.Render(pptx.New(), sc); err == nil {
			t.Errorf("%T with no stages should fail validation", n)
		}
	}
}

// TestFunnelCycle_Deterministic guards worker-count independence (incl. the cycle
// trig + rotation math).
func TestFunnelCycle_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "fc", Nodes: []scene.SlideNode{
		scene.Funnel{Stages: []scene.FunnelStage{{Label: "a", Value: "1"}, {Label: "b"}, {Label: "c"}}},
		scene.Cycle{Stages: []scene.CycleStage{{Label: "p"}, {Label: "q"}, {Label: "r"}, {Label: "s"}}},
	}}}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("funnel/cycle not deterministic (%d vs %d bytes)", len(seq), len(par))
	}
}
