package slide

import (
	"strings"
	"testing"
)

// shapeRect builds a rectangle auto-shape with the given id.
func shapeRect(id int) *XSp {
	return &XSp{
		NonVisual: XNonVisualDrawingShape{
			CNvPr:   &XNvCxnSpPr{ID: id, Name: "rect"},
			CNvSpPr: &XNvSpPr{},
		},
		ShapeProperties: &XShapeProperties{
			Transform2D: &XTransform2D{
				Offset: &XOv2DrOffset{X: 914400, Y: 914400},
				Extent: &XOv2DrExtent{Cx: 2743200, Cy: 1371600},
			},
			PresetGeom: &XPresetGeometry{Prst: "rect", AvLst: &XAvLst{}},
		},
	}
}

// shapeTextBox builds a text box carrying a single run of text.
func shapeTextBox(id int, text string) *XSp {
	return &XSp{
		NonVisual: XNonVisualDrawingShape{
			CNvPr:   &XNvCxnSpPr{ID: id, Name: "TextBox"},
			CNvSpPr: &XNvSpPr{},
		},
		ShapeProperties: &XShapeProperties{
			Transform2D: &XTransform2D{
				Offset: &XOv2DrOffset{X: 914400, Y: 2743200},
				Extent: &XOv2DrExtent{Cx: 4572000, Cy: 914400},
			},
		},
		TextBody: &XTextBody{
			BodyPr:     &XBodyPr{},
			LstStyle:   &XTextParagraphList{},
			Paragraphs: []XTextParagraph{{TextRuns: []XTextRun{{Text: text}}}},
		},
	}
}

// shapePicture builds a picture referencing the given image relationship.
func shapePicture(id int, rID string) *XPicture {
	return &XPicture{
		NonVisual: XNonVisualDrawingPic{
			CNvPr:    &XNvCxnSpPr{ID: id, Name: "Picture"},
			CNvPicPr: &XNvPicPr{},
		},
		BlipFill: &XBlipFillProperties{
			Blip:    &XBlip{Embed: rID},
			Stretch: &XStretchProperties{FillRect: &XFillRectProperties{}},
		},
		ShapeProperties: &XShapeProperties{
			Transform2D: &XTransform2D{
				Offset: &XOv2DrOffset{X: 0, Y: 0},
				Extent: &XOv2DrExtent{Cx: 100, Cy: 100},
			},
		},
	}
}

// TestSlideRoundTrip proves a self-authored slide (rect + text box + picture)
// round-trips losslessly through ToXML → FromXML (D-032; G6). It is the codec
// half of the Phase 03 round-trip requirement; the builder-facing half is in
// test/parts.
func TestSlideRoundTrip(t *testing.T) {
	src := NewSlidePart(1)
	src.AppendShapeChild(shapeRect(2))
	src.AppendShapeChild(shapeTextBox(3, "Hello round-trip"))
	src.AppendShapeChild(shapePicture(4, "rId1"))

	data, err := src.ToXML()
	if err != nil {
		t.Fatalf("ToXML: %v", err)
	}

	dst := NewSlidePart(1)
	if err := dst.FromXML(data); err != nil {
		t.Fatalf("FromXML: %v", err)
	}

	children := dst.SpTree().Children
	if len(children) != 3 {
		t.Fatalf("child count = %d, want 3 (XML:\n%s)", len(children), data)
	}

	rect, ok := children[0].(*XSp)
	if !ok {
		t.Fatalf("child[0] type = %T, want *XSp", children[0])
	}
	if rect.NonVisual.CNvPr.ID != 2 || rect.NonVisual.CNvPr.Name != "rect" {
		t.Errorf("rect cNvPr = %+v, want id=2 name=rect", rect.NonVisual.CNvPr)
	}
	if rect.ShapeProperties.PresetGeom == nil || rect.ShapeProperties.PresetGeom.Prst != "rect" {
		t.Errorf("rect preset geometry not preserved: %+v", rect.ShapeProperties.PresetGeom)
	}
	off := rect.ShapeProperties.Transform2D.Offset
	ext := rect.ShapeProperties.Transform2D.Extent
	if off.X != 914400 || off.Y != 914400 || ext.Cx != 2743200 || ext.Cy != 1371600 {
		t.Errorf("rect transform not preserved: off=%+v ext=%+v", off, ext)
	}

	tb, ok := children[1].(*XSp)
	if !ok {
		t.Fatalf("child[1] type = %T, want *XSp", children[1])
	}
	if got := tb.TextBody.Paragraphs[0].TextRuns[0].Text; got != "Hello round-trip" {
		t.Errorf("text run = %q, want %q", got, "Hello round-trip")
	}

	pic, ok := children[2].(*XPicture)
	if !ok {
		t.Fatalf("child[2] type = %T, want *XPicture", children[2])
	}
	if pic.BlipFill.Blip.Embed != "rId1" {
		t.Errorf("blip embed = %q, want rId1", pic.BlipFill.Blip.Embed)
	}
}

// TestPictureMediaRoundTrip proves a picture's alt text (cNvPr/@descr) and crop
// (srcRect) survive ToXML → FromXML (Chunk C; G6).
func TestPictureMediaRoundTrip(t *testing.T) {
	pic := shapePicture(2, "rId2")
	pic.NonVisual.CNvPr.Descr = "Acme logo"
	pic.BlipFill.SrcRect = &XSrcRect{L: 10000, T: 20000, R: 10000, B: 20000}

	src := NewSlidePart(1)
	src.AppendShapeChild(pic)

	data, err := src.ToXML()
	if err != nil {
		t.Fatalf("ToXML: %v", err)
	}
	dst := NewSlidePart(1)
	if err := dst.FromXML(data); err != nil {
		t.Fatalf("FromXML: %v", err)
	}

	got, ok := dst.SpTree().Children[0].(*XPicture)
	if !ok {
		t.Fatalf("child[0] type = %T, want *XPicture", dst.SpTree().Children[0])
	}
	if got.NonVisual.CNvPr.Descr != "Acme logo" {
		t.Errorf("alt text not preserved: %q", got.NonVisual.CNvPr.Descr)
	}
	if got.BlipFill.SrcRect == nil {
		t.Fatalf("srcRect not preserved")
	}
	if sr := got.BlipFill.SrcRect; sr.L != 10000 || sr.T != 20000 || sr.R != 10000 || sr.B != 20000 {
		t.Errorf("srcRect not preserved: %+v", sr)
	}
}

// TestShapeFillRoundTrip proves a shape's solid fill (with alpha) and outline
// survive ToXML → FromXML (Chunk B; G6).
func TestShapeFillRoundTrip(t *testing.T) {
	rect := shapeRect(2)
	rect.ShapeProperties.SolidFill = &XSolidFill{
		SrgbClr: &XSrgbClr{Val: "FF0000", Alpha: &XAlpha{Val: 50000}},
	}
	rect.ShapeProperties.Line = &XLineProperties{
		Width:     25400,
		SolidFill: &XSolidFill{SrgbClr: &XSrgbClr{Val: "0000FF"}},
		PrstDash:  &XPrstDash{Val: "dash"},
	}

	src := NewSlidePart(1)
	src.AppendShapeChild(rect)

	data, err := src.ToXML()
	if err != nil {
		t.Fatalf("ToXML: %v", err)
	}
	dst := NewSlidePart(1)
	if err := dst.FromXML(data); err != nil {
		t.Fatalf("FromXML: %v", err)
	}

	got, ok := dst.SpTree().Children[0].(*XSp)
	if !ok {
		t.Fatalf("child[0] type = %T, want *XSp", dst.SpTree().Children[0])
	}
	sp := got.ShapeProperties
	if sp.SolidFill == nil || sp.SolidFill.SrgbClr == nil || sp.SolidFill.SrgbClr.Val != "FF0000" {
		t.Fatalf("solid fill color not preserved: %+v", sp.SolidFill)
	}
	if sp.SolidFill.SrgbClr.Alpha == nil || sp.SolidFill.SrgbClr.Alpha.Val != 50000 {
		t.Errorf("fill alpha not preserved: %+v", sp.SolidFill.SrgbClr.Alpha)
	}
	if sp.Line == nil || sp.Line.Width != 25400 {
		t.Fatalf("line width not preserved: %+v", sp.Line)
	}
	if sp.Line.SolidFill == nil || sp.Line.SolidFill.SrgbClr.Val != "0000FF" {
		t.Errorf("line color not preserved: %+v", sp.Line.SolidFill)
	}
	if sp.Line.PrstDash == nil || sp.Line.PrstDash.Val != "dash" {
		t.Errorf("line dash not preserved: %+v", sp.Line.PrstDash)
	}
}

// TestSlideToXMLStructure pins the key structural properties of the emitted
// slide XML: a namespaced root, the cSld envelope, attributes serialized as
// attributes (not element text — the bug that motivated D-032), and the
// r: namespace appearing only when a relationship attribute (picture) is used.
func TestSlideToXMLStructure(t *testing.T) {
	sp := NewSlidePart(1)
	sp.AppendShapeChild(shapeRect(2))
	sp.AppendShapeChild(shapePicture(3, "rId1"))

	data, err := sp.ToXML()
	if err != nil {
		t.Fatalf("ToXML: %v", err)
	}
	xmlStr := string(data)

	for _, want := range []string{
		`<p:sld `,
		`xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"`,
		`xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"`,
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"`,
		`<p:cSld><p:spTree>`,
		`<p:cNvPr id="2" name="rect"/>`,
		`<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`,
		`<a:blip r:embed="rId1"/>`,
		`<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>`,
	} {
		if !strings.Contains(xmlStr, want) {
			t.Errorf("missing %q in:\n%s", want, xmlStr)
		}
	}

	// Attributes must not leak into element text content.
	if strings.Contains(xmlStr, `<p:cNvPr>`) {
		t.Errorf("cNvPr serialized with text content instead of attributes:\n%s", xmlStr)
	}
}
