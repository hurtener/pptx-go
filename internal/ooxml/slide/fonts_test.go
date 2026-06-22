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
		&XTextRun{Text: "no rPr"},                                 // nil TextProperties
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
