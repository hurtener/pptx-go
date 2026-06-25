package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// roadmapScene builds a 6-milestone, 3-band, 2-lane roadmap (R14.4 acceptance).
func roadmapScene() scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID: "roadmap",
		Nodes: []scene.SlideNode{scene.Timeline{
			Bands: []scene.TimelineBand{
				{From: 0, To: 0.34, Label: "Now", Fill: scene.ColorAccent},
				{From: 0.34, To: 0.67, Label: "Next", Fill: scene.ColorInfo},
				{From: 0.67, To: 1, Label: "Later", Fill: scene.ColorSuccess},
			},
			Lanes: []scene.TimelineLane{
				{Label: "Platform", Milestones: []scene.Milestone{
					{Position: 0.1, Label: "Beta", Icon: "star", AccentIndex: 0},
					{Position: 0.5, Label: "GA", Detail: "general availability", AccentIndex: 1},
					{Position: 0.9, Label: "Scale", AccentIndex: 2},
				}},
				{Label: "Go-to-market", Milestones: []scene.Milestone{
					{Position: 0.2, Label: "Pilot"},
					{Position: 0.55, Label: "Launch", Detail: "press + demand-gen"},
					{Position: 0.85, Label: "Expand"},
				}},
			},
		}},
	}}}
}

// TestTimeline_Roadmap is R14.4 acceptance: a 6-milestone / 3-band / 2-lane
// roadmap renders within the safe area (axis lines + marker dots/icons + band
// fills + labels), conformant, with no warnings.
func TestTimeline_Roadmap(t *testing.T) {
	data, stats := render(t, roadmapScene())
	if len(stats.Warnings) != 0 {
		t.Errorf("roadmap: unexpected warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	// Two lane axes are straight lines; band fills + marker dots are ellipses/rects.
	if n := strings.Count(slide, `prst="line"`); n < 2 {
		t.Errorf("roadmap: want >=2 axis lines (one per lane), got %d", n)
	}
	if !strings.Contains(slide, `prst="ellipse"`) {
		t.Errorf("roadmap: missing marker dots (ellipse)")
	}
	if !strings.Contains(slide, "<a:t>Now</a:t>") || !strings.Contains(slide, "<a:t>Platform</a:t>") {
		t.Errorf("roadmap: missing band/lane labels:\n%s", slide)
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("roadmap deck failed conformance:\n%s", rep)
	}
}

// TestTimeline_SingleLane verifies the implicit single-lane path (top-level
// Milestones, no Lanes) renders an axis with markers.
func TestTimeline_SingleLane(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "tl",
		Nodes: []scene.SlideNode{scene.Timeline{Milestones: []scene.Milestone{
			{Position: 0, Label: "Start"}, {Position: 1, Label: "End"},
		}}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("single-lane timeline: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `prst="line"`) || !strings.Contains(slide, "<a:t>Start</a:t>") {
		t.Errorf("single-lane timeline missing axis/label:\n%s", slide)
	}
}

// TestTimeline_InvalidWarns verifies an out-of-range milestone position fails
// Stage-1 validation (Render returns an error).
func TestTimeline_InvalidWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bad",
		Nodes: []scene.SlideNode{scene.Timeline{Milestones: []scene.Milestone{{Position: 1.5, Label: "x"}}}},
	}}}
	if _, err := scene.Render(pptx.New(), sc); err == nil {
		t.Errorf("timeline with position 1.5 should fail validation")
	}
}

// TestTimeline_Deterministic guards worker-count independence.
func TestTimeline_Deterministic(t *testing.T) {
	sc := roadmapScene()
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("timeline not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
