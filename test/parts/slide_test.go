package parts_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/parts"
	"github.com/hurtener/pptx-go/pptx"
)

// writeSlideToXML is a helper that serializes an XSlide using XMLWriter.
func writeSlideToXML(xs *parts.XSlide) ([]byte, error) {
	xw := parts.NewXMLWriterBuffered(4096)
	if err := xw.Declaration(); err != nil {
		return nil, err
	}
	if err := xs.WriteXML(xw); err != nil {
		return nil, err
	}
	return xw.Bytes(), nil
}

// writeTextBodyToXML is a helper that serializes an XTextBody using XMLWriter.
func writeTextBodyToXML(xtb *parts.XTextBody) ([]byte, error) {
	xw := parts.NewXMLWriterBuffered(4096)
	if err := xw.Declaration(); err != nil {
		return nil, err
	}
	if err := xtb.WriteXML(xw); err != nil {
		return nil, err
	}
	return xw.Bytes(), nil
}

// TestSlideBuilder_AddText tests the text-addition logic of the Slide Builder API.
// It asserts directly on the Go struct without involving XML serialization.
func TestSlideBuilder_AddText(t *testing.T) {
	// Instantiate a blank SlidePart.
	slidePart := parts.NewSlidePart(1)
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
	slidePart := parts.NewSlidePart(1)
	builder := pptx.NewSlideBuilder(slidePart)

	// Add several text boxes and collect the returned shape pointers.
	texts := []string{"Heading Text", "Body Paragraph 1", "Body Paragraph 2"}
	var shapes []*parts.XSp
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
	slidePart := parts.NewSlidePart(1)
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
		para := parts.XTextParagraph{
			TextRuns: []parts.XTextRun{
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
		run := parts.XTextRun{
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
		run := parts.XTextRun{
			Text: "Styled",
			TextProperties: &parts.XTextProperties{
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
		para := parts.XTextParagraph{
			Level:     1,     // set attribute
			Alignment: "ctr", // center alignment
			TextRuns: []parts.XTextRun{
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
		sp := parts.XSp{
			NonVisual: parts.XNonVisualDrawingShape{
				CNvPr: &parts.XNvCxnSpPr{
					ID:   1,
					Name: "Test",
				},
				CNvSpPr: &parts.XNvSpPr{},
			},
			ShapeProperties: &parts.XShapeProperties{
				Transform2D: &parts.XTransform2D{
					Offset: &parts.XOv2DrOffset{X: 0, Y: 0},
					Extent: &parts.XOv2DrExtent{Cx: 100, Cy: 100},
				},
			},
			TextBody: &parts.XTextBody{
				Paragraphs: []parts.XTextParagraph{
					{TextRuns: []parts.XTextRun{{Text: "Test"}}},
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
	// Construct XSlide directly to test namespace serialization.
	xslide := parts.XSlide{
		XmlnsA:    "http://schemas.openxmlformats.org/drawingml/2006/main",
		XmlnsR:    "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		XmlnsP:    "http://schemas.openxmlformats.org/presentationml/2006/main",
		ClrMapOvr: &parts.XClrMapOvr{Accent1: "accent1"},
		CSld: &parts.XCSld{
			SpTree: parts.NewXSpTree(),
		},
	}

	// Serialize using WriteXML (produces OOXML format with namespace prefixes).
	data, err := writeSlideToXML(&xslide)
	if err != nil {
		t.Fatalf("WriteXML failed: %v", err)
	}

	xmlStr := string(data)
	t.Logf("generated XML:\n%s", xmlStr)

	// Key assertions: required namespace declarations must be present.
	requiredNamespaces := []string{
		`xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"`,
		`xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"`,
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`,
	}

	for _, ns := range requiredNamespaces {
		if !strings.Contains(xmlStr, ns) {
			t.Errorf("missing required namespace declaration: %s", ns)
		}
	}

	// Verify root element <p:sld>.
	if !strings.Contains(xmlStr, "<p:sld") {
		t.Error("missing root element <p:sld>")
	}
	if !strings.Contains(xmlStr, "</p:sld>") {
		t.Error("missing root element closing tag </p:sld>")
	}

	// Verify shape tree <p:spTree>.
	if !strings.Contains(xmlStr, "<p:spTree>") {
		t.Error("missing shape tree element <p:spTree>")
	}
	if !strings.Contains(xmlStr, "</p:spTree>") {
		t.Error("missing shape tree closing tag </p:spTree>")
	}

	t.Logf("namespace and root element verification passed")
}

// TestSlide_MarshalFullPage_WithContent tests full-slide serialization with content.
func TestSlide_MarshalFullPage_WithContent(t *testing.T) {
	// Construct an XTextBody containing text.
	textBody := parts.XTextBody{
		Paragraphs: []parts.XTextParagraph{
			{
				TextRuns: []parts.XTextRun{
					{Text: "Presentation Title"},
				},
			},
		},
	}

	// Serialize using WriteXML (produces OOXML format with namespace prefixes).
	data, err := writeTextBodyToXML(&textBody)
	if err != nil {
		t.Fatalf("WriteXML failed: %v", err)
	}

	xmlStr := string(data)
	t.Logf("generated XML:\n%s", xmlStr)

	// Verify text content.
	if !strings.Contains(xmlStr, "Presentation Title") {
		t.Error("missing text content: Presentation Title")
	}

	// Verify <a:t> tag.
	if !strings.Contains(xmlStr, "<a:t>") {
		t.Error("missing text tag <a:t>")
	}

	t.Logf("text content serialization verified")
}
