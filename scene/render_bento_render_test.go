package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Bento render, validation, and determinism (Phase 27, R5 c).

func bentoSlide(b scene.Bento) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{b}}}}
}

// TestBento_RendersLabelsAndCells: the row labels and cell content reach the
// slide.
func TestBento_RendersLabelsAndCells(t *testing.T) {
	b := scene.Bento{Columns: 3, Rows: []scene.BentoRow{
		{Label: "Row A", Cells: []scene.BentoCell{
			{Span: 2, Node: scene.Prose{Paragraphs: []scene.RichText{rt("wide cell")}}},
			{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt("narrow")}}},
		}},
		{Label: "Row B", Cells: []scene.BentoCell{
			{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt("bee")}}},
		}},
	}}
	data, _ := render(t, bentoSlide(b))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{"Row A", "Row B", "wide cell", "narrow", "bee"} {
		if !strings.Contains(xml, "<a:t>"+want+"</a:t>") {
			t.Errorf("bento slide missing text %q", want)
		}
	}
}

// TestBento_Validation is acceptance criterion 3: Stage-1 rejects malformed
// bento and accepts a well-formed one.
func TestBento_Validation(t *testing.T) {
	ok := scene.Bento{Columns: 2, Rows: []scene.BentoRow{{Cells: []scene.BentoCell{{Span: 1, Node: scene.Prose{}}}}}}
	if err := scene.ValidateScene(bentoSlide(ok)); err != nil {
		t.Errorf("valid bento rejected: %v", err)
	}
	bad := []struct {
		name string
		b    scene.Bento
	}{
		{"zero columns", scene.Bento{Columns: 0, Rows: []scene.BentoRow{{Cells: []scene.BentoCell{{Span: 1, Node: scene.Prose{}}}}}}},
		{"no rows", scene.Bento{Columns: 2}},
		{"empty row", scene.Bento{Columns: 2, Rows: []scene.BentoRow{{}}}},
		{"zero span", scene.Bento{Columns: 2, Rows: []scene.BentoRow{{Cells: []scene.BentoCell{{Span: 0, Node: scene.Prose{}}}}}}},
		{"nil node", scene.Bento{Columns: 2, Rows: []scene.BentoRow{{Cells: []scene.BentoCell{{Span: 1, Node: nil}}}}}},
		{"spans exceed columns", scene.Bento{Columns: 2, Rows: []scene.BentoRow{{Cells: []scene.BentoCell{{Span: 2, Node: scene.Prose{}}, {Span: 1, Node: scene.Prose{}}}}}}},
	}
	for _, c := range bad {
		if err := scene.ValidateScene(bentoSlide(c.b)); err == nil {
			t.Errorf("%s: expected a validation error, got nil", c.name)
		}
	}
}

// TestBento_Deterministic is acceptance criterion 6: a bento deck renders
// byte-identical across worker counts.
func TestBento_Deterministic(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 12; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + i)),
			Nodes: []scene.SlideNode{scene.Bento{Columns: 3, Rows: []scene.BentoRow{
				{Label: "R", Cells: []scene.BentoCell{
					{Span: 2, Node: scene.Prose{Paragraphs: []scene.RichText{rt("x")}}},
					{Span: 1, Node: scene.Card{Header: "c"}},
				}},
			}}},
		})
	}
	seq, _ := render(t, sc, scene.WithWorkers(1))
	par, _ := render(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("bento deck: parallel render differs from sequential (%d vs %d bytes)", len(par), len(seq))
	}
}
