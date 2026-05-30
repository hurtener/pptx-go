package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/scene"
)

// TestGridCompleteness is acceptance criterion 3: a grid whose cell count is not
// a multiple of columns is a Stage 1 validation error.
func TestGridCompleteness(t *testing.T) {
	complete := scene.Grid{Columns: 3, Cells: []scene.SlideNode{
		scene.Prose{}, scene.Prose{}, scene.Prose{}, scene.Prose{}, scene.Prose{}, scene.Prose{},
	}}
	if err := scene.ValidateScene(sceneWith(complete)); err != nil {
		t.Errorf("complete 3×2 grid rejected: %v", err)
	}
	partial := scene.Grid{Columns: 3, Cells: []scene.SlideNode{scene.Prose{}, scene.Prose{}, scene.Prose{}, scene.Prose{}}}
	if err := scene.ValidateScene(sceneWith(partial)); err == nil {
		t.Error("partial grid (4 cells / 3 columns) accepted; expected a validation error")
	}
}

// TestRenderContainer_TwoColumn is acceptance criterion 4: a two_column renders
// its children as native shapes in their slots and stays conformant.
func TestRenderContainer_TwoColumn(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "tc",
		Nodes: []scene.SlideNode{scene.TwoColumn{
			Ratio: scene.Ratio12,
			Left:  []scene.SlideNode{scene.Heading{Text: rt("Left"), Level: 3}},
			Right: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("Right body")}}},
		}},
	}}}

	data, stats := render(t, sc)
	if stats.Shapes < 2 {
		t.Errorf("two_column rendered %d shapes, want >= 2 (one per child)", stats.Shapes)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<a:t>Left</a:t>") || !strings.Contains(slide, "<a:t>Right body</a:t>") {
		t.Errorf("two_column children missing from slide:\n%s", slide)
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("two_column deck failed conformance:\n%s", rep)
	}
}

// TestRenderGrid renders a grid of leaves and checks all cells appear.
func TestRenderGrid(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "g",
		Nodes: []scene.SlideNode{scene.Grid{Columns: 2, Cells: []scene.SlideNode{
			scene.Heading{Text: rt("A"), Level: 4},
			scene.Heading{Text: rt("B"), Level: 4},
			scene.Heading{Text: rt("C"), Level: 4},
			scene.Heading{Text: rt("D"), Level: 4},
		}}},
	}}}
	data, _ := render(t, sc)
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, label := range []string{"A", "B", "C", "D"} {
		if !strings.Contains(slide, "<a:t>"+label+"</a:t>") {
			t.Errorf("grid cell %q missing:\n%s", label, slide)
		}
	}
}

// TestRenderContainer_Nesting is acceptance criterion 5: a grid inside a
// two_column column composes.
func TestRenderContainer_Nesting(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "nest",
		Nodes: []scene.SlideNode{scene.TwoColumn{
			Ratio: scene.Ratio11,
			Left:  []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("L")}}},
			Right: []scene.SlideNode{scene.Grid{Columns: 2, Cells: []scene.SlideNode{
				scene.Heading{Text: rt("g1"), Level: 5},
				scene.Heading{Text: rt("g2"), Level: 5},
			}}},
		}},
	}}}
	data, _ := render(t, sc)
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, label := range []string{"L", "g1", "g2"} {
		if !strings.Contains(slide, "<a:t>"+label+"</a:t>") {
			t.Errorf("nested content %q missing:\n%s", label, slide)
		}
	}
}
