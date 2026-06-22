package scene_test

import (
	"bytes"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Cross-feature interaction (Wave 10 §17 checkpoint, NH9). Each test compounds two
// or more Wave-10 mechanisms on one node tree and asserts the render is
// deterministic across worker counts (and does not panic). These guard the
// untested interactions the checkpoint flagged.

func crossDeterministic(t *testing.T, name string, nodes []scene.SlideNode) {
	t.Helper()
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: name, Nodes: nodes}}}
	seq, _ := render(t, sc, scene.WithWorkers(1))
	par, _ := render(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Fatalf("%s: parallel render differs from sequential (%d vs %d bytes)", name, len(par), len(seq))
	}
	if len(seq) == 0 {
		t.Fatalf("%s: empty render", name)
	}
}

// TestCross_CardFitPaddingWrappedHeader: a card compounding BodyVAlign=VAlignFit +
// PaddingScale + a wrapped multi-line header (cardPaddingFor shrinks the body box,
// the wrapped header inflates the chrome, alignedStackIn(Fit) compresses inside).
func TestCross_CardFitPaddingWrappedHeader(t *testing.T) {
	card := scene.Card{
		Header:       "Platform White Label Operating Layer Console",
		Size:         scene.CardSizeLG,
		PaddingScale: 5000,
		BodyVAlign:   scene.VAlignFit,
		Body: []scene.SlideNode{
			scene.Prose{Paragraphs: []scene.RichText{rt("one"), rt("two"), rt("three"), rt("four")}},
			scene.List{Items: []scene.ListItem{{Text: rt("a")}, {Text: rt("b")}, {Text: rt("c")}}},
		},
	}
	crossDeterministic(t, "cross-a", []scene.SlideNode{scene.Grid{Columns: 3, Cells: []scene.SlideNode{card, scene.Card{Header: "b"}, scene.Card{Header: "c"}}}})
}

// TestCross_AutoFitStatInWeightedBento: an AutoFit Stat inside a WeightedRows bento
// cell (the weighted row measures the cell at full size; renderStat then shrinks).
func TestCross_AutoFitStatInWeightedBento(t *testing.T) {
	bento := scene.Bento{Columns: 3, WeightedRows: true, Rows: []scene.BentoRow{
		{Label: "Metrics", Cells: []scene.BentoCell{
			{Span: 1, Node: scene.Stat{Value: "$4,000,000+", Label: "ARR", AutoFit: true}},
			{Span: 2, Node: scene.Prose{Paragraphs: []scene.RichText{rt("dense capability description across the row")}}},
		}},
		{Cells: []scene.BentoCell{{Span: 1, Node: scene.Stat{Value: "38%", Label: "Margin", AutoFit: true}}}},
	}}
	crossDeterministic(t, "cross-b", []scene.SlideNode{bento})
}

// TestCross_CardBodyCappedAndBalanced: VAlignFillCapped and VAlignBalanced card
// bodies (the 4 of 8 modes the per-phase card-body tests did not cover).
func TestCross_CardBodyCappedAndBalanced(t *testing.T) {
	capped := scene.Card{Header: "Capped", BodyVAlign: scene.VAlignFillCapped, Body: []scene.SlideNode{
		scene.Grid{Columns: 2, Cells: []scene.SlideNode{scene.Card{Header: "i1"}, scene.Card{Header: "i2"}}},
		scene.Prose{Paragraphs: []scene.RichText{rt("note")}},
	}}
	balanced := scene.Card{Header: "Balanced", BodyVAlign: scene.VAlignBalanced, Body: []scene.SlideNode{
		scene.Prose{Paragraphs: []scene.RichText{rt("lead")}},
		scene.List{Items: []scene.ListItem{{Text: rt("x")}, {Text: rt("y")}}},
	}}
	crossDeterministic(t, "cross-c", []scene.SlideNode{scene.Grid{Columns: 2, Cells: []scene.SlideNode{capped, balanced}}})
}
