package pptx_test

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/pptx"
)

var tblBox = pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(8), H: pptx.In(3)}

// firstTable returns the table from the first slide of a reopened deck.
func firstTable(t *testing.T, data []byte) *slide.XTable {
	t.Helper()
	r, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	for _, c := range r.Slides()[0].Part().SpTree().Children {
		if gf, ok := c.(*slide.XGraphicFrame); ok && gf.Graphic != nil && gf.Graphic.GraphicData != nil {
			return gf.Graphic.GraphicData.Table
		}
	}
	t.Fatal("no table found on slide")
	return nil
}

// TestTableMerge_RoundTrip is acceptance criterion 1: a merged cell round-trips
// losslessly through pptx.Open and the deck stays conformant.
func TestTableMerge_RoundTrip(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	tbl := s.AddTable(tblBox, 3, 3)
	tbl.Cell(0, 0).SetText("spans two").MergeRight(2)
	tbl.Cell(1, 0).SetText("tall").MergeDown(2)
	tbl.Cell(0, 2).SetText("c")

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("table deck failed conformance:\n%s", rep)
	}

	got := firstTable(t, data)
	if got.Rows[0].Cells[0].GridSpan != 2 {
		t.Errorf("gridSpan = %d, want 2", got.Rows[0].Cells[0].GridSpan)
	}
	if got.Rows[0].Cells[1].HMerge != "1" {
		t.Errorf("covered cell hMerge = %q, want \"1\"", got.Rows[0].Cells[1].HMerge)
	}
	if got.Rows[1].Cells[0].RowSpan != 2 {
		t.Errorf("rowSpan = %d, want 2", got.Rows[1].Cells[0].RowSpan)
	}
	if got.Rows[2].Cells[0].VMerge != "1" {
		t.Errorf("covered cell vMerge = %q, want \"1\"", got.Rows[2].Cells[0].VMerge)
	}
}

// TestTableBand is acceptance criterion 2: a banded table alternates row fills.
func TestTableBand(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	tbl := s.AddTable(tblBox, 4, 2)
	for r := 0; r < 4; r++ {
		for c := 0; c < 2; c++ {
			tbl.Cell(r, c).SetText("x")
		}
	}
	tbl.SetBanding(true, false) // no header → rows 1 and 3 banded

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	got := firstTable(t, data)

	filled := func(r int) bool {
		pr := got.Rows[r].Cells[0].Pr
		return pr != nil && pr.SolidFill != nil
	}
	if filled(0) || filled(2) {
		t.Errorf("even rows should be unfilled (band starts on odd rows)")
	}
	if !filled(1) || !filled(3) {
		t.Errorf("odd rows should carry the band fill")
	}
}

// TestTable_HeaderRow proves a header row marks tblPr firstRow and fills row 0.
func TestTable_HeaderRow(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	tbl := s.AddTable(tblBox, 2, 2)
	tbl.Cell(0, 0).SetText("H1")
	tbl.Cell(0, 1).SetText("H2")
	tbl.SetHeaderRow(true)

	data, _ := p.WriteToBytes()
	got := firstTable(t, data)
	if got.Pr == nil || got.Pr.FirstRow != "1" {
		t.Errorf("tblPr firstRow not set: %+v", got.Pr)
	}
	if pr := got.Rows[0].Cells[0].Pr; pr == nil || pr.SolidFill == nil {
		t.Errorf("header cell not filled: %+v", pr)
	}
}
