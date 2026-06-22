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
		{Typeface: "Playfair Display", Bold: false, Italic: false, Weight: 400},
		{Typeface: "Playfair Display", Bold: true, Italic: false, Weight: 700},
		{Typeface: "Inter", Bold: false, Italic: false, Weight: 400},
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
	want := []FontFace{{Typeface: "Inter", Weight: 400}}
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
	want := []FontFace{{Typeface: "Lora", Italic: true, Weight: 400}}
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

	playfairToGeorgia := func(typeface string, _, _ bool) string {
		if typeface == "Playfair Display" {
			return "Georgia"
		}
		return ""
	}
	n := s.RewriteFontFaces(playfairToGeorgia)
	if n != 3 {
		t.Errorf("rewrote %d runs, want 3", n)
	}
	got := s.UsedFontFaces()
	want := []FontFace{
		{Typeface: "Georgia", Bold: false, Weight: 400},
		{Typeface: "Georgia", Bold: true, Weight: 700},
		{Typeface: "Inter", Weight: 400},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("after rewrite UsedFontFaces = %#v, want %#v", got, want)
	}

	// Idempotent: a second pass with the same resolver rewrites nothing.
	if n2 := s.RewriteFontFaces(playfairToGeorgia); n2 != 0 {
		t.Errorf("second rewrite changed %d runs, want 0", n2)
	}
	// Nil resolver and nil part are no-ops.
	if n3 := s.RewriteFontFaces(nil); n3 != 0 {
		t.Errorf("nil resolver rewrote %d runs, want 0", n3)
	}
	var nilPart *SlidePart
	if n4 := nilPart.RewriteFontFaces(playfairToGeorgia); n4 != 0 {
		t.Errorf("nil part rewrote %d runs, want 0", n4)
	}
}

func TestRewriteFontFacesItalicAware(t *testing.T) {
	s := NewSlidePart(1)
	s.AppendShapeChild(textBoxWithRuns(2,
		faceRun("upright", "Display", false, false),
		faceRun("emph", "Display", false, true),
	))
	// Only the italic run of "Display" is rewritten.
	resolve := func(typeface string, _, italic bool) string {
		if typeface == "Display" && italic {
			return "Georgia"
		}
		return ""
	}
	if n := s.RewriteFontFaces(resolve); n != 1 {
		t.Fatalf("rewrote %d runs, want 1 (italic only)", n)
	}
	got := s.UsedFontFaces()
	want := []FontFace{
		{Typeface: "Display", Italic: false, Weight: 400},
		{Typeface: "Georgia", Italic: true, Weight: 400},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("after italic-aware rewrite = %#v, want %#v", got, want)
	}
}

func TestUsedFontFacesCarriesWeight(t *testing.T) {
	s := NewSlidePart(1)
	withW := &XTextProperties{Latin: &XLatinFont{Typeface: "Inter"}, Weight: 500}
	boldNoW := &XTextProperties{Latin: &XLatinFont{Typeface: "Inter"}, Bold: "1"} // Weight 0 → infer 700
	regNoW := &XTextProperties{Latin: &XLatinFont{Typeface: "Inter"}}             // Weight 0 → infer 400
	s.AppendShapeChild(textBoxWithRuns(2,
		&XTextRun{Text: "m", TextProperties: withW},
		&XTextRun{Text: "b", TextProperties: boldNoW},
		&XTextRun{Text: "r", TextProperties: regNoW},
	))
	got := s.UsedFontFaces()
	want := []FontFace{
		{Typeface: "Inter", Weight: 500},
		{Typeface: "Inter", Bold: true, Weight: 700},
		{Typeface: "Inter", Weight: 400},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UsedFontFaces weights = %#v, want %#v", got, want)
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
