package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for the R11.3 container-slide-bounds clamp (D-083).

// tallBento builds a Bento whose preferred height far exceeds a slide's body
// region, so its slot overflows the safe area and the clamp must fire.
func tallBento() scene.Bento {
	var rows []scene.BentoRow
	for i := 0; i < 8; i++ {
		rows = append(rows, scene.BentoRow{Cells: []scene.BentoCell{
			{Span: 1, Node: scene.Card{Header: "left", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body text")}}}}},
			{Span: 1, Node: scene.Card{Header: "right", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body text")}}}}},
		}})
	}
	return scene.Bento{Columns: 2, Rows: rows}
}

// TestContainerOverflow_Warns: a deck whose sole node is an over-tall Bento logs
// the safe-area clamp warning.
func TestContainerOverflow_Warns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{tallBento()}}}}
	_, stats := render(t, sc)
	var found bool
	for _, w := range stats.Warnings {
		if strings.Contains(w.Message, "exceeds the slide safe area") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a safe-area clamp warning; got %+v", stats.Warnings)
	}
}

// TestContainerFits_NoWarn: a deck whose container comfortably fits logs no clamp
// warning (the byte-identical path).
func TestContainerFits_NoWarn(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.Grid{Columns: 2, Cells: []scene.SlideNode{scene.Card{Header: "a"}, scene.Card{Header: "b"}}},
	}}}}
	_, stats := render(t, sc)
	for _, w := range stats.Warnings {
		if strings.Contains(w.Message, "exceeds the slide safe area") {
			t.Errorf("fitting container should not warn; got %q", w.Message)
		}
	}
}
