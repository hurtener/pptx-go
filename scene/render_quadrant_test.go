package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func quadrantScene() scene.Scene {
	tl, tr := scene.ColorAccent, scene.ColorInfo
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID: "quadrant",
		Nodes: []scene.SlideNode{scene.Quadrant{
			AxisX: scene.QuadrantAxis{LowLabel: "Low effort", HighLabel: "High effort"},
			AxisY: scene.QuadrantAxis{LowLabel: "Low impact", HighLabel: "High impact"},
			Quadrants: [4]scene.QuadrantCell{
				{Title: "Quick wins", Fill: &tl}, {Title: "Big bets", Fill: &tr},
				{Title: "Fill-ins"}, {Title: "Thankless"},
			},
			Items: []scene.QuadrantItem{
				{X: 0.2, Y: 0.8, Label: "Onboarding"},
				{X: 0.75, Y: 0.7, Label: "Platform", AccentIndex: 1},
				{X: 0.5, Y: 0.5, Label: "Center"},
				{X: 0.1, Y: 0.2, Label: "Cleanup"},
				{X: 0.9, Y: 0.3, Label: "Migration", AccentIndex: 2},
				{X: 0.95, Y: 0.95, Label: "Edge"},
			},
		}},
	}}}
}

// TestQuadrant is R14.9 acceptance: a 2x2 with 6 items renders axes + dividers +
// 4 tints + item dots, conformant, no warnings.
func TestQuadrant(t *testing.T) {
	data, stats := render(t, quadrantScene())
	if len(stats.Warnings) != 0 {
		t.Errorf("quadrant: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, `prst="line"`); n < 2 {
		t.Errorf("quadrant: want >=2 divider lines, got %d", n)
	}
	if n := strings.Count(slide, `prst="ellipse"`); n < 6 {
		t.Errorf("quadrant: want >=6 item dots, got %d", n)
	}
	if !strings.Contains(slide, "<a:t>High impact</a:t>") || !strings.Contains(slide, "<a:t>Quick wins</a:t>") {
		t.Errorf("quadrant: missing axis/quadrant labels")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("quadrant deck failed conformance:\n%s", rep)
	}
}

// TestQuadrant_InvalidWarns verifies an out-of-range coordinate fails validation.
func TestQuadrant_InvalidWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bad",
		Nodes: []scene.SlideNode{scene.Quadrant{Items: []scene.QuadrantItem{{X: 1.5, Y: 0.5, Label: "x"}}}},
	}}}
	if _, err := scene.Render(pptx.New(), sc); err == nil {
		t.Errorf("quadrant with x=1.5 should fail validation")
	}
}

// TestQuadrant_Deterministic guards worker-count independence.
func TestQuadrant_Deterministic(t *testing.T) {
	sc := quadrantScene()
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("quadrant not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
