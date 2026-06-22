package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// List bullet indent density (Phase 47, R10.9). Black-box: an IndentTight List
// emits a smaller, consistent marker-to-text offset; the default (IndentNormal)
// is byte-identical; the render is deterministic.

func listIndentScene(ind scene.ListIndent) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID: "s",
		Nodes: []scene.SlideNode{scene.List{
			Indent: ind,
			Items: []scene.ListItem{
				{Text: rt("Understand")},
				{Text: rt("Operate"), Level: 1},
				{Text: rt("Execute")},
			},
		}},
	}}}
}

// TestListIndent_TightSmallerOffset: a tight list emits the In(0.25) marL across
// all items (consistent), and not the 0.5" default.
func TestListIndent_TightSmallerOffset(t *testing.T) {
	data, _ := render(t, listIndentScene(scene.IndentTight))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(xml, `marL="228600"`); n != 3 {
		t.Errorf("tight list should emit marL=\"228600\" on all 3 items, got %d", n)
	}
	if strings.Contains(xml, `marL="457200"`) {
		t.Error("tight list should not emit the default 0.5\" marL=\"457200\"")
	}
}

// TestListIndent_DefaultByteIdentical: IndentNormal is byte-identical to a list
// that leaves Indent unset.
func TestListIndent_DefaultByteIdentical(t *testing.T) {
	normal, _ := render(t, listIndentScene(scene.IndentNormal))
	unset := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "s",
		Nodes: []scene.SlideNode{scene.List{Items: []scene.ListItem{
			{Text: rt("Understand")},
			{Text: rt("Operate"), Level: 1},
			{Text: rt("Execute")},
		}}},
	}}}
	unsetData, _ := render(t, unset)
	if !bytes.Equal(normal, unsetData) {
		t.Error("IndentNormal is not byte-identical to leaving Indent unset")
	}
	// And the default emits the 0.5" offset.
	if !strings.Contains(zipPart(t, normal, "ppt/slides/slide1.xml"), `marL="457200"`) {
		t.Error("default list should emit the 0.5\" marL=\"457200\"")
	}
}

// TestListIndent_Deterministic: a tight-list deck renders byte-identically across
// worker counts.
func TestListIndent_Deterministic(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 12; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:    string(rune('A' + i)),
			Nodes: listIndentScene(scene.IndentTight).Slides[0].Nodes,
		})
	}
	seq, _ := render(t, sc, scene.WithWorkers(1))
	par, _ := render(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("tight-list deck: parallel render differs from sequential (%d vs %d bytes)", len(par), len(seq))
	}
}
