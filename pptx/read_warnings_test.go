package pptx

import (
	"archive/zip"
	"bytes"
	"io"
	"sort"
	"strings"
	"testing"
)

// rwBox is a reusable placement box for the read-warning tests.
var rwBox = Box{X: 914400, Y: 914400, W: 1828800, H: 914400}

// unzipParts reads every zip entry into a name→bytes map.
func unzipParts(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	m := make(map[string][]byte, len(zr.File))
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		m[f.Name] = b
	}
	return m
}

// rezipParts writes the parts back into a zip in a deterministic order.
func rezipParts(t *testing.T, m map[string][]byte) []byte {
	t.Helper()
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, n := range names {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatalf("zip create %s: %v", n, err)
		}
		if _, err := w.Write(m[n]); err != nil {
			t.Fatalf("zip write %s: %v", n, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// authoredDeck builds a minimal one-slide deck and returns its bytes.
func authoredDeck(t *testing.T) []byte {
	t.Helper()
	p := New()
	s := p.AddSlide()
	s.AddShape(ShapeRect, rwBox, WithFill(SolidFill(RGB("2563EB"))))
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestReadWarnings_DedupAndOrder unit-tests the warning accumulator: duplicates
// per (Kind, Part, Element) collapse, and the surfaced order is by part, then
// element, then kind.
func TestReadWarnings_DedupAndOrder(t *testing.T) {
	p := &Presentation{}
	p.addReadWarning(ReadWarning{Kind: WarnDroppedElement, Part: "/ppt/slides/slide2.xml", Element: "grpSp"})
	p.addReadWarning(ReadWarning{Kind: WarnDroppedElement, Part: "/ppt/slides/slide2.xml", Element: "grpSp"}) // dup
	p.addReadWarning(ReadWarning{Kind: WarnDroppedElement, Part: "/ppt/slides/slide1.xml", Element: "AlternateContent"})
	p.addReadWarning(ReadWarning{Kind: WarnUnreadablePart, Part: "/ppt/slides/slide3.xml"})
	p.sortReadWarnings()

	got := p.ReadWarnings()
	if len(got) != 3 {
		t.Fatalf("ReadWarnings() = %d, want 3 (one dup collapsed)", len(got))
	}
	if got[0].Part != "/ppt/slides/slide1.xml" || got[0].Element != "AlternateContent" {
		t.Errorf("first = %+v, want slide1/AlternateContent", got[0])
	}
	if got[1].Part != "/ppt/slides/slide2.xml" || got[1].Element != "grpSp" {
		t.Errorf("second = %+v, want slide2/grpSp", got[1])
	}
	if got[2].Kind != WarnUnreadablePart || got[2].Part != "/ppt/slides/slide3.xml" {
		t.Errorf("third = %+v, want unreadable slide3", got[2])
	}
}

// TestReadWarnings_AuthoredDeckIsClean is acceptance criterion 1 (the negative):
// a pptx-go-authored deck reopens with no read warnings.
func TestReadWarnings_AuthoredDeckIsClean(t *testing.T) {
	re, err := NewFromBytes(authoredDeck(t))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if w := re.ReadWarnings(); len(w) != 0 {
		t.Errorf("authored deck ReadWarnings() = %+v, want none", w)
	}
}

// TestReadWarnings_DroppedElement is acceptance criterion 1: a slide carrying an
// unrecognized shape-tree element reopens with a WarnDroppedElement naming the
// part and element, and the open does not fail.
func TestReadWarnings_DroppedElement(t *testing.T) {
	parts := unzipParts(t, authoredDeck(t))
	const slide = "ppt/slides/slide1.xml"
	// A group shape pptx-go does not model; p: is already declared on the slide
	// root, so the codec sees the bare local-name "grpSp" and ignores it.
	const grpSp = `<p:grpSp><p:nvGrpSpPr><p:cNvPr id="42" name="grp"/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:grpSp>`
	xml := string(parts[slide])
	if !strings.Contains(xml, "</p:spTree>") {
		t.Fatalf("slide XML missing </p:spTree>:\n%s", xml)
	}
	parts[slide] = []byte(strings.Replace(xml, "</p:spTree>", grpSp+"</p:spTree>", 1))

	re, err := NewFromBytes(rezipParts(t, parts))
	if err != nil {
		t.Fatalf("NewFromBytes on deck with unknown element: %v", err)
	}
	ws := re.ReadWarnings()
	var found bool
	for _, w := range ws {
		if w.Kind == WarnDroppedElement && w.Element == "grpSp" && w.Part == "/ppt/slides/slide1.xml" {
			found = true
		}
	}
	if !found {
		t.Errorf("WarnDroppedElement{grpSp} not surfaced; got %+v", ws)
	}
	// The known shapes still reconstruct — the unknown element is dropped, not fatal.
	if n := len(re.Slides()[0].Shapes()); n != 1 {
		t.Errorf("reopened slide has %d navigable shapes, want 1 (the authored rect)", n)
	}
}

// TestReadWarnings_PartPassThrough is acceptance criterion 3: a part pptx-go does
// not model survives NewFromBytes → WriteToBytes byte-for-byte (the OPC
// pass-through; D-035). Unrecognized PARTS are preserved even though
// unrecognized shapes are not (D-048).
func TestReadWarnings_PartPassThrough(t *testing.T) {
	parts := unzipParts(t, authoredDeck(t))
	const custom = "customXml/item1.xml"
	customBytes := []byte(`<?xml version="1.0" encoding="UTF-8"?><root>external-unmodeled-payload</root>`)
	parts[custom] = customBytes
	// Register a content type for the new part so the package loads it.
	ct := string(parts["[Content_Types].xml"])
	override := `<Override PartName="/customXml/item1.xml" ContentType="application/xml"/></Types>`
	parts["[Content_Types].xml"] = []byte(strings.Replace(ct, "</Types>", override, 1))

	re, err := NewFromBytes(rezipParts(t, parts))
	if err != nil {
		t.Fatalf("NewFromBytes with unmodeled part: %v", err)
	}
	out, err := re.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	got := unzipParts(t, out)[custom]
	if !bytes.Equal(got, customBytes) {
		t.Errorf("unmodeled part not preserved: got %q, want %q", got, customBytes)
	}
}
