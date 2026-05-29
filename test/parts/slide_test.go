package parts_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/pptx"
)

// slidePartToXML builds a SlidePart whose shape tree is populated by fn and
// returns the emitted slide XML (xml.Marshal + ooxml.RestoreNamespaces, D-032).
func slidePartToXML(t *testing.T, fn func(b *pptx.SlideBuilder)) string {
	t.Helper()
	sp := slide.NewSlidePart(1)
	if fn != nil {
		fn(pptx.NewSlideBuilder(sp))
	}
	data, err := sp.ToXML()
	if err != nil {
		t.Fatalf("SlidePart.ToXML: %v", err)
	}
	return string(data)
}

// TestSlideBuilder_AddText tests the text-addition logic of the Slide Builder API.
// It asserts directly on the Go struct without involving XML serialization.
func TestSlideBuilder_AddText(t *testing.T) {
	// Instantiate a blank SlidePart.
	slidePart := slide.NewSlidePart(1)
	if slidePart == nil {
		t.Fatal("NewSlidePart returned nil")
	}

	// Add text via the SlideBuilder API.
	testText := "Test Title"
	x, y, cx, cy := 914400, 457200, 9144000, 1143000 // EMU units
	builder := pptx.NewSlideBuilder(slidePart)
	sp := builder.AddTextBox(x, y, cx, cy, testText)

	// The returned shape must not be nil (proves ShapeTree gained a shape).
	if sp == nil {
		t.Fatal("AddTextBox returned nil")
	}

	// Verify the shape's TextBody -> Paragraph -> Run -> Text value.
	textBody := sp.TextBody
	if textBody == nil {
		t.Fatal("shape TextBody is nil")
	}

	if len(textBody.Paragraphs) == 0 {
		t.Fatal("TextBody.Paragraphs length is 0")
	}

	para := textBody.Paragraphs[0]
	if len(para.TextRuns) == 0 {
		t.Fatal("Paragraph.TextRuns length is 0")
	}

	run := para.TextRuns[0]
	if run.Text != testText {
		t.Errorf("Text = %q, want %q", run.Text, testText)
	}

	t.Logf("text content verified successfully: %q", run.Text)
}

// TestSlideBuilder_AddTextBox_Multiple tests adding multiple text boxes.
func TestSlideBuilder_AddTextBox_Multiple(t *testing.T) {
	slidePart := slide.NewSlidePart(1)
	builder := pptx.NewSlideBuilder(slidePart)

	// Add several text boxes and collect the returned shape pointers.
	texts := []string{"Heading Text", "Body Paragraph 1", "Body Paragraph 2"}
	var shapes []*slide.XSp
	for i, text := range texts {
		y := 457200 + i*914400 // incrementing Y coordinate
		sp := builder.AddTextBox(914400, y, 9144000, 457200, text)
		if sp == nil {
			t.Fatalf("AddTextBox #%d returned nil", i+1)
		}
		shapes = append(shapes, sp)
	}

	// Verify the number of returned shapes.
	if len(shapes) != len(texts) {
		t.Errorf("shape count = %d, want %d", len(shapes), len(texts))
	}

	// Verify the text content of each shape.
	for i, sp := range shapes {
		if sp.TextBody == nil || len(sp.TextBody.Paragraphs) == 0 {
			t.Errorf("shape #%d TextBody structure is incomplete", i+1)
			continue
		}
		para := sp.TextBody.Paragraphs[0]
		if len(para.TextRuns) == 0 {
			t.Errorf("shape #%d has no text runs", i+1)
			continue
		}
		if para.TextRuns[0].Text != texts[i] {
			t.Errorf("shape #%d text = %q, want %q", i+1, para.TextRuns[0].Text, texts[i])
		}
	}

	t.Logf("successfully added %d text boxes", len(shapes))
}

// TestSlideBuilder_AddTextBox_VerifyStructure tests the full structure created by AddTextBox.
func TestSlideBuilder_AddTextBox_VerifyStructure(t *testing.T) {
	slidePart := slide.NewSlidePart(1)
	builder := pptx.NewSlideBuilder(slidePart)

	testText := "Full Structure Test"
	x, y, cx, cy := 1000000, 2000000, 3000000, 4000000
	sp := builder.AddTextBox(x, y, cx, cy, testText)

	// Verify non-visual properties of the shape.
	if sp.NonVisual.CNvPr == nil {
		t.Fatal("NonVisual.CNvPr is nil")
	}
	if sp.NonVisual.CNvPr.ID == 0 {
		t.Error("shape ID is 0")
	}
	t.Logf("shape ID: %d, Name: %s", sp.NonVisual.CNvPr.ID, sp.NonVisual.CNvPr.Name)

	// Verify position and size in shape properties.
	if sp.ShapeProperties == nil {
		t.Fatal("ShapeProperties is nil")
	}
	if sp.ShapeProperties.Transform2D == nil {
		t.Fatal("Transform2D is nil")
	}
	if sp.ShapeProperties.Transform2D.Offset == nil {
		t.Fatal("Offset is nil")
	}
	if sp.ShapeProperties.Transform2D.Extent == nil {
		t.Fatal("Extent is nil")
	}

	// Verify coordinates.
	offset := sp.ShapeProperties.Transform2D.Offset
	if offset.X != x || offset.Y != y {
		t.Errorf("Offset = (%d, %d), want (%d, %d)", offset.X, offset.Y, x, y)
	}

	extent := sp.ShapeProperties.Transform2D.Extent
	if extent.Cx != cx || extent.Cy != cy {
		t.Errorf("Extent = (%d, %d), want (%d, %d)", extent.Cx, extent.Cy, cx, cy)
	}

	t.Logf("position verified: offset=(%d, %d), extent=(%d, %d)", offset.X, offset.Y, extent.Cx, extent.Cy)
}

// TestSlide_MarshalComponents tests XML serialization of low-level components.
// It verifies that omitempty tags suppress zero-value attributes correctly.
func TestSlide_MarshalComponents(t *testing.T) {
	// Test XTextParagraph serialization.
	// Note: XTextParagraph has no XMLName field, so the root tag is the struct name.
	t.Run("XTextParagraph_OmitEmpty", func(t *testing.T) {
		// Construct a minimal XTextParagraph with a single XTextRun containing "Hello".
		// Leave all optional fields at their zero values.
		para := slide.XTextParagraph{
			TextRuns: []slide.XTextRun{
				{Text: "Hello"},
			},
		}

		// Serialize with xml.Marshal.
		data, err := xml.Marshal(&para)
		if err != nil {
			t.Fatalf("xml.Marshal failed: %v", err)
		}

		xmlStr := string(data)
		t.Logf("generated XML: %s", xmlStr)

		// Core assertion: no empty attribute tags must be present.
		// Level, Indent, Alignment are omitempty and must not appear at zero value.
		forbiddenPatterns := []string{
			`lvl="0"`,
			`indent="0"`,
			`algn=""`,
		}

		for _, pattern := range forbiddenPatterns {
			if strings.Contains(xmlStr, pattern) {
				t.Errorf("XML contains zero-value attribute: %s (should be suppressed by omitempty)", pattern)
			}
		}

		// Verify text content is serialized correctly.
		if !strings.Contains(xmlStr, "Hello") {
			t.Error("should contain text 'Hello'")
		}
	})

	// Test XTextRun serialization — verify omitempty on TextProperties.
	t.Run("XTextRun_OmitEmpty", func(t *testing.T) {
		// TextProperties is not set.
		run := slide.XTextRun{
			Text: "World",
		}

		data, err := xml.Marshal(&run)
		if err != nil {
			t.Fatalf("xml.Marshal failed: %v", err)
		}

		xmlStr := string(data)
		t.Logf("generated XML: %s", xmlStr)

		// The rPr tag must not appear (TextProperties is nil and omitempty).
		if strings.Contains(xmlStr, "<a:rPr>") || strings.Contains(xmlStr, "<a:rPr/>") {
			t.Error("XTextRun should not contain an empty <a:rPr> tag when TextProperties is nil")
		}

		// Verify text content.
		if !strings.Contains(xmlStr, "World") {
			t.Error("should contain text 'World'")
		}
	})

	// Test XTextRun with properties — verify attribute-level omitempty.
	t.Run("XTextRun_WithProperties", func(t *testing.T) {
		run := slide.XTextRun{
			Text: "Styled",
			TextProperties: &slide.XTextProperties{
				FontSize: 2400, // 24pt
				Bold:     true,
			},
		}

		data, err := xml.Marshal(&run)
		if err != nil {
			t.Fatalf("xml.Marshal failed: %v", err)
		}

		xmlStr := string(data)
		t.Logf("generated XML: %s", xmlStr)

		// The set attributes should be present.
		if !strings.Contains(xmlStr, `sz="2400"`) {
			t.Error("should contain sz=\"2400\" attribute")
		}
		// Note: Go's xml package serializes bool as "true"/"false".
		if !strings.Contains(xmlStr, `b="true"`) {
			t.Error("should contain b=\"true\" attribute")
		}

		// Unset attributes (Italic, Underline, FontFace, Color) must not appear.
		unexpectedAttrs := []string{"i=", "u=", "typeface=", "solidFill"}
		for _, attr := range unexpectedAttrs {
			if strings.Contains(xmlStr, attr) {
				t.Errorf("should not contain unset attribute: %s", attr)
			}
		}
	})

	// Test XTextParagraph with attributes — verify omitempty applies to attributes too.
	t.Run("XTextParagraph_WithAttributes", func(t *testing.T) {
		para := slide.XTextParagraph{
			Level:     1,     // set attribute
			Alignment: "ctr", // center alignment
			TextRuns: []slide.XTextRun{
				{Text: "Centered"},
			},
		}

		data, err := xml.Marshal(&para)
		if err != nil {
			t.Fatalf("xml.Marshal failed: %v", err)
		}

		xmlStr := string(data)
		t.Logf("generated XML: %s", xmlStr)

		// lvl and algn attributes must be present.
		if !strings.Contains(xmlStr, `lvl="1"`) {
			t.Error("should contain lvl=\"1\" attribute")
		}
		if !strings.Contains(xmlStr, `algn="ctr"`) {
			t.Error("should contain algn=\"ctr\" attribute")
		}

		// indent must not appear (zero value, omitempty).
		if strings.Contains(xmlStr, "indent") {
			t.Error("should not contain indent attribute (zero value with omitempty)")
		}
	})

	// Test XShapeProperties omitempty.
	t.Run("XShapeProperties_OmitEmpty", func(t *testing.T) {
		sp := slide.XSp{
			NonVisual: slide.XNonVisualDrawingShape{
				CNvPr: &slide.XNvCxnSpPr{
					ID:   1,
					Name: "Test",
				},
				CNvSpPr: &slide.XNvSpPr{},
			},
			ShapeProperties: &slide.XShapeProperties{
				Transform2D: &slide.XTransform2D{
					Offset: &slide.XOv2DrOffset{X: 0, Y: 0},
					Extent: &slide.XOv2DrExtent{Cx: 100, Cy: 100},
				},
			},
			TextBody: &slide.XTextBody{
				Paragraphs: []slide.XTextParagraph{
					{TextRuns: []slide.XTextRun{{Text: "Test"}}},
				},
			},
		}

		data, err := xml.Marshal(&sp)
		if err != nil {
			t.Fatalf("xml.Marshal failed: %v", err)
		}

		xmlStr := string(data)
		t.Logf("generated XML (first 200 chars): %.200s...", xmlStr)

		// Rotation, FlipH, FlipV and similar omitempty attributes must not appear.
		rotationPatterns := []string{`rot="0"`, `flipH="false"`, `flipV="false"`}
		for _, pattern := range rotationPatterns {
			if strings.Contains(xmlStr, pattern) {
				t.Errorf("should not contain zero-value attribute: %s", pattern)
			}
		}
	})
}

// TestSlide_MarshalFullPage is a full-slide serialization smoke test.
// It verifies namespace declarations, which are critical for PowerPoint to open the file.
func TestSlide_MarshalFullPage(t *testing.T) {
	xmlStr := slidePartToXML(t, func(b *pptx.SlideBuilder) {
		b.AddAutoShape(914400, 914400, 2743200, 1371600, "rect")
	})
	t.Logf("generated XML:\n%s", xmlStr)

	// Key assertions: the namespaces the shape tree actually uses (p: and a:)
	// must be declared on the root. r: is only declared when a relationship
	// attribute is present (e.g. a picture), so it is not required here.
	requiredNamespaces := []string{
		`xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"`,
		`xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"`,
	}
	for _, ns := range requiredNamespaces {
		if !strings.Contains(xmlStr, ns) {
			t.Errorf("missing required namespace declaration: %s", ns)
		}
	}

	// Verify the structural envelope: <p:sld><p:cSld><p:spTree>.
	for _, want := range []string{"<p:sld", "</p:sld>", "<p:cSld>", "<p:spTree>", "</p:spTree>"} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("missing %q", want)
		}
	}

	// The shape's preset geometry must serialize (regression: prstGeom was
	// previously dropped via the xml:"-" tag).
	if !strings.Contains(xmlStr, `<a:prstGeom prst="rect">`) {
		t.Errorf("missing preset geometry <a:prstGeom prst=\"rect\">")
	}

	// Attributes must serialize as attributes, not as element text content
	// (the retired hand-rolled writer emitted `<p:cNvPr>2 name="..."`).
	if strings.Contains(xmlStr, `<p:cNvPr>2`) {
		t.Errorf("cNvPr id leaked into element text instead of an attribute:\n%s", xmlStr)
	}
	if !strings.Contains(xmlStr, `id="2"`) {
		t.Errorf("missing shape id attribute id=\"2\"")
	}
}

// TestSlide_MarshalFullPage_WithContent tests full-slide serialization with text content.
func TestSlide_MarshalFullPage_WithContent(t *testing.T) {
	xmlStr := slidePartToXML(t, func(b *pptx.SlideBuilder) {
		b.AddTextBox(914400, 914400, 4572000, 914400, "Presentation Title")
	})
	t.Logf("generated XML:\n%s", xmlStr)

	if !strings.Contains(xmlStr, "Presentation Title") {
		t.Error("missing text content: Presentation Title")
	}
	if !strings.Contains(xmlStr, "<a:t>") {
		t.Error("missing text tag <a:t>")
	}
}
