package slide

import (
	"reflect"
	"testing"
)

// faceRun builds a run carrying a Latin typeface with bold/italic flags.
func faceRun(text, typeface string, bold, italic bool) *XTextRun {
	pr := &XTextProperties{Latin: &XLatinFont{Typeface: typeface}}
	if bold {
		pr.Bold = "1"
	}
	if italic {
		pr.Italic = "1"
	}
	return &XTextRun{Text: text, TextProperties: pr}
}

// textBoxWithRuns builds a text box whose single paragraph holds the runs.
func textBoxWithRuns(id int, runs ...*XTextRun) *XSp {
	content := make([]any, len(runs))
	for i, r := range runs {
		content[i] = r
	}
	return &XSp{
		NonVisual: XNonVisualDrawingShape{
			CNvPr:   &XNvCxnSpPr{ID: id, Name: "TextBox"},
			CNvSpPr: &XNvSpPr{},
		},
		TextBody: &XTextBody{
			Paragraphs: []XTextParagraph{{Content: content}},
		},
	}
}

func TestUsedFontFacesDistinctAndOrdered(t *testing.T) {
	s := NewSlidePart(1)
	s.AppendShapeChild(textBoxWithRuns(2,
		faceRun("Title", "Playfair Display", false, false),
		faceRun("Bold bit", "Playfair Display", true, false),
		faceRun("Body", "Inter", false, false),
		faceRun("Dup body", "Inter", false, false), // duplicate — deduped
	))

	got := s.UsedFontFaces()
	want := []FontFace{
		{Typeface: "Playfair Display", Bold: false, Italic: false},
		{Typeface: "Playfair Display", Bold: true, Italic: false},
		{Typeface: "Inter", Bold: false, Italic: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UsedFontFaces = %#v, want %#v", got, want)
	}
}

func TestUsedFontFacesIgnoresUnsetLatin(t *testing.T) {
	s := NewSlidePart(1)
	s.AppendShapeChild(textBoxWithRuns(2,
		&XTextRun{Text: "no rPr"}, // nil TextProperties
		&XTextRun{Text: "no latin", TextProperties: &XTextProperties{Bold: "1"}}, // no Latin
		faceRun("explicit", "Inter", false, false),
	))
	got := s.UsedFontFaces()
	want := []FontFace{{Typeface: "Inter"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UsedFontFaces = %#v, want %#v (only explicit Latin faces)", got, want)
	}
}

func TestUsedFontFacesWalksTableCells(t *testing.T) {
	cell := XTableCell{TextBody: &XTextBody{
		Paragraphs: []XTextParagraph{{Content: []any{faceRun("cell", "Lora", false, true)}}},
	}}
	gf := &XGraphicFrame{
		Graphic: &XGraphic{
			GraphicData: &XGraphicData{
				Table: &XTable{Rows: []XTableRow{{Cells: []XTableCell{cell}}}},
			},
		},
	}
	s := NewSlidePart(1)
	s.AppendShapeChild(gf)

	got := s.UsedFontFaces()
	want := []FontFace{{Typeface: "Lora", Italic: true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UsedFontFaces (table) = %#v, want %#v", got, want)
	}
}

func TestRewriteFontFaces(t *testing.T) {
	s := NewSlidePart(1)
	s.AppendShapeChild(textBoxWithRuns(2,
		faceRun("a", "Playfair Display", false, false),
		faceRun("b", "Playfair Display", true, false),
		faceRun("c", "Inter", false, false), // untouched
		&XTextRun{Text: "no latin"},         // skipped
	))
	// Table cell run also rewritten.
	cell := XTableCell{TextBody: &XTextBody{
		Paragraphs: []XTextParagraph{{Content: []any{faceRun("d", "Playfair Display", false, false)}}},
	}}
	s.AppendShapeChild(&XGraphicFrame{Graphic: &XGraphic{GraphicData: &XGraphicData{
		Table: &XTable{Rows: []XTableRow{{Cells: []XTableCell{cell}}}},
	}}})

	n := s.RewriteFontFaces(map[string]string{"Playfair Display": "Georgia"})
	if n != 3 {
		t.Errorf("rewrote %d runs, want 3", n)
	}
	got := s.UsedFontFaces()
	want := []FontFace{
		{Typeface: "Georgia", Bold: false},
		{Typeface: "Georgia", Bold: true},
		{Typeface: "Inter"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("after rewrite UsedFontFaces = %#v, want %#v", got, want)
	}

	// Idempotent: a second pass with the same mapping rewrites nothing.
	if n2 := s.RewriteFontFaces(map[string]string{"Playfair Display": "Georgia"}); n2 != 0 {
		t.Errorf("second rewrite changed %d runs, want 0", n2)
	}
	// Nil/empty mapping and nil part are no-ops.
	if n3 := s.RewriteFontFaces(nil); n3 != 0 {
		t.Errorf("nil mapping rewrote %d runs, want 0", n3)
	}
	var nilPart *SlidePart
	if n4 := nilPart.RewriteFontFaces(map[string]string{"x": "y"}); n4 != 0 {
		t.Errorf("nil part rewrote %d runs, want 0", n4)
	}
}

func TestUsedFontFacesEmpty(t *testing.T) {
	s := NewSlidePart(1)
	s.AppendShapeChild(shapeRect(2)) // no text body
	if got := s.UsedFontFaces(); len(got) != 0 {
		t.Errorf("UsedFontFaces = %#v, want empty", got)
	}
	var nilPart *SlidePart
	if got := nilPart.UsedFontFaces(); got != nil {
		t.Errorf("nil SlidePart UsedFontFaces = %#v, want nil", got)
	}
}
