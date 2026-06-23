package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for Grid.Connectors (R12.4, D-099): gutter glyphs between
// adjacent columns, the bidirectional arrow, additivity, validation, and determinism.

func connGrid(conns []scene.GridConnector) scene.Scene {
	cell := func(s string) scene.SlideNode {
		return scene.Card{Header: s, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{{{Text: s}}}}}}
	}
	return scene.Scene{Slides: []scene.SceneSlide{
		{ID: "arch", Nodes: []scene.SlideNode{
			scene.Grid{Columns: 3, Connectors: conns, Cells: []scene.SlideNode{cell("People"), cell("Operate"), cell("Knowledge")}},
		}},
	}}
}

// TestGridConnectors_RenderGlyphs: connectors between adjacent columns draw glyphs; a
// bidirectional connector emits a leftRightArrow.
func TestGridConnectors_RenderGlyphs(t *testing.T) {
	data, _ := render(t, connGrid([]scene.GridConnector{
		{Between: [2]int{0, 1}, Kind: scene.ConnectorArrow},
		{Between: [2]int{1, 2}, Kind: scene.ConnectorBiArrow, Label: "feeds"},
	}))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, `prst="rightArrow"`) {
		t.Error("arrow connector did not emit a rightArrow")
	}
	if !strings.Contains(xml, `prst="leftRightArrow"`) {
		t.Error("bi-arrow connector did not emit a leftRightArrow")
	}
	if !strings.Contains(xml, "feeds") {
		t.Error("connector label missing")
	}
}

// TestGridConnectors_Additive: a grid with connectors has more shapes than the same grid
// without; an empty Connectors slice adds nothing.
func TestGridConnectors_Additive(t *testing.T) {
	none, _ := render(t, connGrid(nil))
	withConns, _ := render(t, connGrid([]scene.GridConnector{{Between: [2]int{0, 1}}, {Between: [2]int{1, 2}}}))
	xn := zipPart(t, none, "ppt/slides/slide1.xml")
	xw := zipPart(t, withConns, "ppt/slides/slide1.xml")
	if strings.Count(xw, "<p:sp>") <= strings.Count(xn, "<p:sp>") {
		t.Error("connectors should add shapes to the grid")
	}
}

// TestGridConnectors_Validation: a non-adjacent or out-of-range connector fails Stage-1.
func TestGridConnectors_Validation(t *testing.T) {
	if err := scene.ValidateScene(connGrid([]scene.GridConnector{{Between: [2]int{0, 2}}})); err == nil {
		t.Error("non-adjacent connector {0,2} passed validation")
	}
	if err := scene.ValidateScene(connGrid([]scene.GridConnector{{Between: [2]int{2, 3}}})); err == nil {
		t.Error("out-of-range connector {2,3} on a 3-column grid passed validation")
	}
	if err := scene.ValidateScene(connGrid([]scene.GridConnector{{Between: [2]int{0, 1}, Kind: scene.ConnectorBiArrow}})); err != nil {
		t.Errorf("valid adjacent connector rejected: %v", err)
	}
}

// TestGridConnectors_Deterministic: a connectored grid renders byte-identically.
func TestGridConnectors_Deterministic(t *testing.T) {
	sc := connGrid([]scene.GridConnector{
		{Between: [2]int{0, 1}, Kind: scene.ConnectorArrow, Label: "to"},
		{Between: [2]int{1, 2}, Kind: scene.ConnectorBiArrow},
	})
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("grid connector render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
