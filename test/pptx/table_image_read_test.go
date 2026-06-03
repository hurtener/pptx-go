package pptx_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestTableRead_StructureAndCells is PR#3 acceptance criterion 3 (table): a
// reopened table shape yields the authored row/column counts, column widths,
// header/banding intent, per-cell text, and merge spans.
func TestTableRead_StructureAndCells(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		tbl := s.AddTable(fxBox, 3, 2)
		tbl.SetColumnWidths(pptx.Cm(4), pptx.Cm(6))
		tbl.SetHeaderRow(true)
		tbl.SetBanding(true, false)
		tbl.Cell(0, 0).SetText("H1")
		tbl.Cell(0, 1).SetText("H2")
		tbl.Cell(1, 0).SetText("merged").MergeRight(2)
		tbl.Cell(2, 0).SetText("down").MergeDown(2) // clamps to table bounds
	})
	if len(shapes) != 1 {
		t.Fatalf("Shapes() = %d, want 1", len(shapes))
	}
	tbl, ok := shapes[0].Table()
	if !ok {
		t.Fatal("first shape Table() ok = false, want true")
	}

	if tbl.RowCount() != 3 || tbl.ColCount() != 2 {
		t.Errorf("dimensions = %dx%d, want 3x2", tbl.RowCount(), tbl.ColCount())
	}
	if !tbl.HeaderRow() {
		t.Error("HeaderRow() = false, want true")
	}
	if !tbl.RowBanding() {
		t.Error("RowBanding() = false, want true")
	}
	if got, want := tbl.ColumnWidths(), []pptx.EMU{pptx.Cm(4), pptx.Cm(6)}; !reflect.DeepEqual(got, want) {
		t.Errorf("ColumnWidths() = %v, want %v", got, want)
	}

	// Cell text round-trips through the PR#2 text read model.
	if got := cellText(t, tbl, 0, 0); got != "H1" {
		t.Errorf("cell(0,0) text = %q, want %q", got, "H1")
	}
	if got := cellText(t, tbl, 0, 1); got != "H2" {
		t.Errorf("cell(0,1) text = %q, want %q", got, "H2")
	}

	// Horizontal merge: the anchor spans 2 cols, the covered cell reports Covered.
	if span := tbl.Cell(1, 0).GridSpan(); span != 2 {
		t.Errorf("cell(1,0) GridSpan() = %d, want 2", span)
	}
	if !tbl.Cell(1, 1).Covered() {
		t.Error("cell(1,1) Covered() = false, want true")
	}
	// An unmerged cell reports span 1 and not covered.
	if span := tbl.Cell(0, 0).GridSpan(); span != 1 {
		t.Errorf("cell(0,0) GridSpan() = %d, want 1", span)
	}
	if tbl.Cell(0, 0).Covered() {
		t.Error("cell(0,0) Covered() = true, want false")
	}
}

func cellText(t *testing.T, tbl *pptx.Table, row, col int) string {
	t.Helper()
	paras := tbl.Cell(row, col).TextFrame().Paragraphs()
	if len(paras) == 0 {
		return ""
	}
	runs := paras[0].Runs()
	if len(runs) == 0 {
		return ""
	}
	return runs[0].Text()
}

// TestTableRead_HeaderCellFill is PR#3 acceptance criterion 3 (cell fill): the
// concrete header/banding fill applyStyling emits reopens as a readable solid
// fill on the cell.
func TestTableRead_HeaderCellFill(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		tbl := s.AddTable(fxBox, 2, 1)
		tbl.SetHeaderRow(true) // fills row 0 with SurfaceAlt
		tbl.Cell(0, 0).SetText("head")
	})
	tbl, _ := shapes[0].Table()
	if f := tbl.Cell(0, 0).Fill(); f == nil || f.Kind() != pptx.FillSolid {
		t.Errorf("header cell Fill() = %#v, want a solid fill", f)
	}
	if f := tbl.Cell(1, 0).Fill(); f != nil {
		t.Errorf("body cell Fill() = %#v, want nil", f)
	}
}

// TestImageRead_PropsAndBytes is PR#3 acceptance criterion 3 (image): a reopened
// image yields the authored alt / crop / fit / rotation / opacity and resolves
// the embedded bytes verbatim.
func TestImageRead_PropsAndBytes(t *testing.T) {
	raw := pngBytes("the-payload-bytes")
	crop := pptx.Crop{Left: 0.1, Top: 0.2, Right: 0.05, Bottom: 0.15}
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		img, err := s.AddImage(pptx.ImageBytes(raw, "image/png"), fxBox)
		if err != nil {
			t.Fatalf("AddImage: %v", err)
		}
		img.SetAltText("a logo").SetCrop(crop).SetFit(pptx.FitNone).SetRotation(90).SetOpacity(40000)
	})
	if len(shapes) != 1 {
		t.Fatalf("Shapes() = %d, want 1", len(shapes))
	}
	img, ok := shapes[0].Image()
	if !ok {
		t.Fatal("first shape Image() ok = false, want true")
	}

	if got := img.AltText(); got != "a logo" {
		t.Errorf("AltText() = %q, want %q", got, "a logo")
	}
	if got := img.Crop(); got != crop {
		t.Errorf("Crop() = %+v, want %+v", got, crop)
	}
	if got := img.Fit(); got != pptx.FitNone {
		t.Errorf("Fit() = %v, want FitNone", got)
	}
	if got := img.Rotation(); got != 90 {
		t.Errorf("Rotation() = %v, want 90", got)
	}
	if got := img.Opacity(); got != 40000 {
		t.Errorf("Opacity() = %d, want 40000", got)
	}

	got, err := img.Bytes()
	if err != nil {
		t.Fatalf("Bytes(): %v", err)
	}
	if !bytes.Equal(got, raw) {
		t.Errorf("Bytes() = %d bytes, want the %d authored bytes", len(got), len(raw))
	}
}

// TestImageRead_Defaults is PR#3 acceptance criterion 3 (image defaults): an
// image with no options reopens with FitFill, opaque, uncropped, unrotated, no
// alt text.
func TestImageRead_Defaults(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		if _, err := s.AddImage(pptx.ImageBytes(pngBytes("y"), "image/png"), fxBox); err != nil {
			t.Fatalf("AddImage: %v", err)
		}
	})
	img, _ := shapes[0].Image()
	if img.Fit() != pptx.FitFill {
		t.Errorf("Fit() = %v, want FitFill (default)", img.Fit())
	}
	if img.Opacity() != pptx.AlphaOpaque {
		t.Errorf("Opacity() = %d, want AlphaOpaque", img.Opacity())
	}
	if img.Crop() != (pptx.Crop{}) {
		t.Errorf("Crop() = %+v, want zero", img.Crop())
	}
	if img.Rotation() != 0 {
		t.Errorf("Rotation() = %v, want 0", img.Rotation())
	}
	if img.AltText() != "" {
		t.Errorf("AltText() = %q, want empty", img.AltText())
	}
}

// TestTableImageRead_WrongShape is PR#3 acceptance criterion 3 (negative): a
// plain shape is neither a table nor an image.
func TestTableImageRead_WrongShape(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		s.AddShape(pptx.ShapeRect, fxBox)
	})
	if _, ok := shapes[0].Table(); ok {
		t.Error("Rect Table() ok = true, want false")
	}
	if _, ok := shapes[0].Image(); ok {
		t.Error("Rect Image() ok = true, want false")
	}
}
