package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

func statSlide(nodes ...scene.SlideNode) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: nodes}}}
}

// TestStat_RendersValueLabelDelta is acceptance criterion 1: the value, label,
// and delta text all reach the slide.
func TestStat_RendersValueLabelDelta(t *testing.T) {
	data, _ := render(t, statSlide(scene.Stat{Value: "$2,200", Label: "ARR", Delta: "+12%", DeltaTone: scene.DeltaUp}))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{"$2,200", "ARR", "+12%"} {
		if !strings.Contains(xml, "<a:t>"+want+"</a:t>") {
			t.Errorf("stat slide missing text %q", want)
		}
	}
}

// TestStat_NoDeltaOmitsLine: a Stat without a Delta renders value + label only.
func TestStat_NoDeltaOmitsLine(t *testing.T) {
	data, _ := render(t, statSlide(scene.Stat{Value: "42", Label: "NPS"}))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:t>42</a:t>") || !strings.Contains(xml, "<a:t>NPS</a:t>") {
		t.Error("stat missing value/label")
	}
}

// TestStat_GridStrip is acceptance criterion 2: a Grid of Stats renders a strip
// (one stat per cell).
func TestStat_GridStrip(t *testing.T) {
	grid := scene.Grid{Columns: 3, Cells: []scene.SlideNode{
		scene.Stat{Value: "$2,200", Label: "ARR", Delta: "+12%", DeltaTone: scene.DeltaUp},
		scene.Stat{Value: "38%", Label: "Margin", Delta: "-3%", DeltaTone: scene.DeltaDown},
		scene.Stat{Value: "4.8", Label: "NPS"},
	}}
	data, _ := render(t, statSlide(grid))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{"$2,200", "38%", "4.8", "ARR", "Margin", "NPS"} {
		if !strings.Contains(xml, "<a:t>"+want+"</a:t>") {
			t.Errorf("stat strip missing text %q", want)
		}
	}
}

// TestStat_Validation is acceptance criterion 3: Stage-1 rejects an empty value.
func TestStat_Validation(t *testing.T) {
	if err := scene.ValidateScene(statSlide(scene.Stat{Value: "1"})); err != nil {
		t.Errorf("valid stat rejected: %v", err)
	}
	if err := scene.ValidateScene(statSlide(scene.Stat{Label: "no value"})); err == nil {
		t.Error("a stat with an empty Value should fail Stage-1")
	}
}

// TestStat_Deterministic is acceptance criterion 5: a stat deck renders
// byte-identical across worker counts.
func TestStat_Deterministic(t *testing.T) {
	sc := scene.Scene{}
	tones := []scene.DeltaTone{scene.DeltaUp, scene.DeltaDown, scene.DeltaNeutral}
	for i := 0; i < 12; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + i)),
			Nodes: []scene.SlideNode{scene.Grid{Columns: 3, Cells: []scene.SlideNode{
				scene.Stat{Value: "1", Label: "a", Delta: "+1", DeltaTone: tones[i%3]},
				scene.Stat{Value: "2", Label: "b"},
				scene.Stat{Value: "3", Label: "c", Delta: "-1", DeltaTone: scene.DeltaDown},
			}}},
		})
	}
	seq, _ := render(t, sc, scene.WithWorkers(1))
	par, _ := render(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("stat deck: parallel render differs from sequential (%d vs %d bytes)", len(par), len(seq))
	}
}
