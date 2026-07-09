package scene

import (
	"bytes"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestGridFillFillsBodyRegion(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(10), H: pptx.In(6)}
	grid := Grid{Columns: 2, Fill: true, Cells: []SlideNode{
		Card{Header: "A", Body: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "one"}}}}}},
		Card{Header: "B", Body: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "two"}}}}}},
		Card{Header: "C", Body: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "three"}}}}}},
		Card{Header: "D", Body: []SlideNode{Prose{Paragraphs: []RichText{{{Text: "four"}}}}}},
	}}
	nodes := []SlideNode{Heading{Text: RichText{{Text: "Deck"}}, Level: 2}, grid}
	pl := r.alignedStackIn(box, nodes, "grid-fill", Alignment{Vertical: VAlignTop})
	if len(pl) != 2 {
		t.Fatalf("placements = %d, want 2", len(pl))
	}
	if got, want := pl[1].box.Bottom(), box.Bottom(); got != want {
		t.Fatalf("grid bottom = %d, want body bottom %d", got, want)
	}
	nat := preferredHeight(grid, box.W, r.theme)
	if pl[1].box.H <= nat {
		t.Fatalf("grid fill height = %d, want > natural %d", pl[1].box.H, nat)
	}
}

func TestBentoFillFillsBodyRegion(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(10), H: pptx.In(6)}
	bento := Bento{Columns: 2, WeightedRows: true, Fill: true, Rows: []BentoRow{
		{Label: "Control plane", Cells: []BentoCell{{Span: 1, Node: Prose{Paragraphs: []RichText{{{Text: "one one one one one one one"}}}}}, {Span: 1, Node: Prose{Paragraphs: []RichText{{{Text: "two two two two two two two"}}}}}}},
		{Label: "The core", Cells: []BentoCell{{Span: 2, Node: Prose{Paragraphs: []RichText{{{Text: "three three three three three three three"}}}}}}},
	}}
	nodes := []SlideNode{Heading{Text: RichText{{Text: "Deck"}}, Level: 2}, bento}
	pl := r.alignedStackIn(box, nodes, "bento-fill", Alignment{Vertical: VAlignTop})
	if len(pl) != 2 {
		t.Fatalf("placements = %d, want 2", len(pl))
	}
	if got, want := pl[1].box.Bottom(), box.Bottom(); got != want {
		t.Fatalf("bento bottom = %d, want body bottom %d", got, want)
	}
	nat := preferredHeight(bento, box.W, r.theme)
	if pl[1].box.H <= nat {
		t.Fatalf("bento fill height = %d, want > natural %d", pl[1].box.H, nat)
	}
}

func TestGridBentoFillFalseByteIdentical(t *testing.T) {
	render := func(nodes []SlideNode) []byte {
		t.Helper()
		pres := pptx.New()
		_, err := Render(pres, Scene{Slides: []SceneSlide{{ID: "A", Nodes: nodes}}})
		if err != nil {
			t.Fatalf("Render: %v", err)
		}
		b, err := pres.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return b
	}
	baseGrid := Grid{Columns: 2, Cells: []SlideNode{Card{Header: "A"}, Card{Header: "B"}}}
	falseGrid := Grid{Columns: 2, Fill: false, Cells: []SlideNode{Card{Header: "A"}, Card{Header: "B"}}}
	baseBento := Bento{Columns: 2, Rows: []BentoRow{{Cells: []BentoCell{{Span: 1, Node: Prose{Paragraphs: []RichText{{{Text: "one"}}}}}, {Span: 1, Node: Prose{Paragraphs: []RichText{{{Text: "two"}}}}}}}}}
	falseBento := Bento{Columns: 2, Fill: false, Rows: []BentoRow{{Cells: []BentoCell{{Span: 1, Node: Prose{Paragraphs: []RichText{{{Text: "one"}}}}}, {Span: 1, Node: Prose{Paragraphs: []RichText{{{Text: "two"}}}}}}}}}
	base := render([]SlideNode{Heading{Text: RichText{{Text: "Deck"}}, Level: 2}, baseGrid, baseBento})
	withFalse := render([]SlideNode{Heading{Text: RichText{{Text: "Deck"}}, Level: 2}, falseGrid, falseBento})
	if !bytes.Equal(base, withFalse) {
		t.Fatal("Fill=false changed rendered bytes")
	}
}
