package pptx_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Per-paragraph BulletIndent (Phase 47, R10.9, D-078): ParagraphOpts.BulletIndent
// overrides the default 0.5" bullet hanging indent, emitted as a:pPr/@marL +
// @indent. Zero is byte-identical.

func bulletSlideXML(t *testing.T, data []byte) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rc, _ := f.Open()
			defer func() { _ = rc.Close() }()
			b, _ := io.ReadAll(rc)
			return string(b)
		}
	}
	t.Fatal("slide1.xml not found")
	return ""
}

func bulletDeck(t *testing.T, opts pptx.ParagraphOpts) []byte {
	t.Helper()
	p := pptx.New()
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 914400, W: 5000000, H: 1000000})
	tf.AddParagraph(opts).AddRun("Item", pptx.RunStyle{TypeRole: pptx.TypeBody})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestBulletIndent_TightEmits: a tighter BulletIndent emits the reduced
// marL/indent (the marker-to-text offset), smaller than the 0.5" default.
func TestBulletIndent_TightEmits(t *testing.T) {
	xml := bulletSlideXML(t, bulletDeck(t, pptx.ParagraphOpts{Bullet: pptx.BulletDisc, BulletIndent: 228600}))
	if !strings.Contains(xml, `marL="228600"`) {
		t.Errorf("tight BulletIndent should emit marL=\"228600\"; xml:\n%s", xml)
	}
	if !strings.Contains(xml, `indent="-228600"`) {
		t.Errorf("tight BulletIndent should emit indent=\"-228600\"; xml:\n%s", xml)
	}
	// Must be tighter than the 0.5" (457200) default.
	if strings.Contains(xml, `marL="457200"`) {
		t.Error("tight BulletIndent should not emit the default marL=\"457200\"")
	}
}

// TestBulletIndent_DefaultByteIdentical: BulletIndent 0 keeps the default 0.5"
// hanging indent — byte-identical to a paragraph with no BulletIndent field.
func TestBulletIndent_DefaultByteIdentical(t *testing.T) {
	bare := bulletDeck(t, pptx.ParagraphOpts{Bullet: pptx.BulletDisc})
	zero := bulletDeck(t, pptx.ParagraphOpts{Bullet: pptx.BulletDisc, BulletIndent: 0})
	if !bytes.Equal(bare, zero) {
		t.Errorf("BulletIndent 0 is not byte-identical to no BulletIndent (%d vs %d)", len(zero), len(bare))
	}
	if !strings.Contains(bulletSlideXML(t, bare), `marL="457200"`) {
		t.Error("default bulleted paragraph should emit the 0.5\" marL=\"457200\"")
	}
}

// TestBulletIndent_RoundTrips (G6): the tight marL/indent survive a reopen, read
// back both as raw XML and via the Paragraph.BulletIndent accessor (the inverse
// of ParagraphOpts.BulletIndent).
func TestBulletIndent_RoundTrips(t *testing.T) {
	data := bulletDeck(t, pptx.ParagraphOpts{Bullet: pptx.BulletDisc, BulletIndent: 228600})
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	// Go-model accessor round-trip.
	tf, ok := re.Slides()[0].Shapes()[0].TextFrame()
	if !ok {
		t.Fatal("reopened shape has no text frame")
	}
	if got := tf.Paragraphs()[0].BulletIndent(); got != 228600 {
		t.Errorf("reopened Paragraph.BulletIndent() = %d, want 228600", got)
	}
	// Raw-XML round-trip (re-emit).
	out, err := re.WriteToBytes()
	if err != nil {
		t.Fatalf("re-WriteToBytes: %v", err)
	}
	if !strings.Contains(bulletSlideXML(t, out), `marL="228600"`) {
		t.Error("reopened deck lost the tight marL=\"228600\"")
	}
}

// TestBulletIndent_AccessorDefault: a default-indent bulleted paragraph reports
// the 0.5" marL; a non-bulleted paragraph reports 0.
func TestBulletIndent_AccessorDefault(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: 5000000, H: 1000000})
	bulleted := tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletDisc})
	plain := tf.AddParagraph(pptx.ParagraphOpts{})
	if got := bulleted.BulletIndent(); got != 457200 {
		t.Errorf("default bulleted BulletIndent() = %d, want 457200 (0.5\")", got)
	}
	if got := plain.BulletIndent(); got != 0 {
		t.Errorf("plain paragraph BulletIndent() = %d, want 0", got)
	}
}
