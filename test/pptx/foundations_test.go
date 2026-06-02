package pptx_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

var fxBox = pptx.Box{X: 914400, Y: 914400, W: 1828800, H: 1828800}

// TestGradientFill_EmitsAndConforms is acceptance criterion 1 (PR #1): linear &
// radial gradient fills emit the gradFill wire and produce a conformant deck.
func TestGradientFill_EmitsAndConforms(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	// Radial glow: accent (centre, opaque) → accent (edge, transparent).
	s.AddShape(pptx.ShapeEllipse, fxBox, pptx.WithFill(pptx.RadialGradient(
		pptx.GradientStop{Pos: 0, Color: pptx.TokenColor(pptx.ColorAccent)},
		pptx.GradientStop{Pos: 1, Color: pptx.TokenColorAlpha(pptx.ColorAccent, pptx.AlphaTransparent)},
	)))
	// Linear gradient.
	s.AddShape(pptx.ShapeRect, fxBox, pptx.WithFill(pptx.LinearGradient(90,
		pptx.GradientStop{Pos: 0, Color: pptx.RGB("2563EB")},
		pptx.GradientStop{Pos: 1, Color: pptx.RGB("FFFFFF")},
	)))

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("gradient deck failed conformance:\n%s", rep)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{`<a:gradFill>`, `<a:path path="circle">`, `<a:fillToRect`, `<a:lin ang="5400000"`, `<a:alpha val="0"`} {
		if !strings.Contains(slide, want) {
			t.Errorf("gradient slide missing %q in:\n%s", want, slide)
		}
	}
	// Round-trip fidelity (G6): the gradient survives Open → re-save, not just
	// reopen-without-error.
	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("reopen gradient deck: %v", err)
	}
	resaved, err := reopened.WriteToBytes()
	if err != nil {
		t.Fatalf("re-save reopened gradient deck: %v", err)
	}
	if rs := readZipPart(t, resaved, "ppt/slides/slide1.xml"); !strings.Contains(rs, "<a:gradFill>") || !strings.Contains(rs, `path="circle"`) {
		t.Errorf("gradient did not survive round-trip through Open:\n%s", rs)
	}
}

// TestWithRotation is acceptance criterion 2 (PR #1): WithRotation sets the
// xfrm rot attribute (degrees × 60000) and round-trips.
func TestWithRotation(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	s.AddShape(pptx.ShapeChevron, fxBox, pptx.WithRotation(45),
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `rot="2700000"`) { // 45 × 60000
		t.Errorf("rotated shape missing rot=2700000 in:\n%s", slide)
	}
}

// TestTokenColorAlpha is acceptance criterion 3 (PR #1): a token color at a
// caller alpha emits the token's RGB with an alpha child.
func TestTokenColorAlpha(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	s.AddShape(pptx.ShapeRect, fxBox, pptx.WithFill(pptx.SolidFill(pptx.TokenColorAlpha(pptx.ColorAccent, 30000))))
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `2563EB`) || !strings.Contains(slide, `<a:alpha val="30000"`) {
		t.Errorf("token-alpha fill missing accent + alpha in:\n%s", slide)
	}
}

// TestSetMetadata is acceptance criterion 4 (PR #1): SetMetadata writes escaped
// core properties, round-trips, and is byte-identical across renders.
func TestSetMetadata(t *testing.T) {
	build := func() []byte {
		p := pptx.New()
		p.AddSlide()
		p.SetMetadata(pptx.Metadata{Title: "Q3 & Q4 <Review>", Author: "Acme", Subject: "Results"})
		data, err := p.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}
	data := build()
	core := readZipPart(t, data, "docProps/core.xml")
	for _, want := range []string{
		"<dc:title>Q3 &amp; Q4 &lt;Review&gt;</dc:title>",
		"<dc:creator>Acme</dc:creator>",
		"<dc:subject>Results</dc:subject>",
	} {
		if !strings.Contains(core, want) {
			t.Errorf("core.xml missing %q in:\n%s", want, core)
		}
	}
	// No timestamps (determinism).
	if strings.Contains(core, "dcterms:created") || strings.Contains(core, "dcterms:modified") {
		t.Errorf("core.xml carries a timestamp (breaks determinism):\n%s", core)
	}
	if again := build(); !bytes.Equal(data, again) {
		t.Errorf("metadata deck is not byte-identical across renders")
	}
	if _, err := pptx.NewFromBytes(data); err != nil {
		t.Fatalf("reopen metadata deck: %v", err)
	}
}

// TestImage_RotationAndOpacity checks the image mutators emit the picture
// rotation and the blip alphaModFix, and round-trip (the asset-decoration
// rotation/opacity wiring from the Phase-13 audit).
func TestImage_RotationAndOpacity(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	img, err := s.AddImage(pptx.ImageBytes(pngBytes("x"), "image/png"), fxBox)
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	img.SetRotation(30).SetOpacity(40000)
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `rot="1800000"`) { // 30 × 60000
		t.Errorf("image missing rotation in:\n%s", slide)
	}
	if !strings.Contains(slide, `<a:alphaModFix amt="40000"`) {
		t.Errorf("image missing alphaModFix in:\n%s", slide)
	}
	again, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	resaved, err := again.WriteToBytes()
	if err != nil {
		t.Fatalf("re-save: %v", err)
	}
	if rs := readZipPart(t, resaved, "ppt/slides/slide1.xml"); !strings.Contains(rs, `<a:alphaModFix amt="40000"`) {
		t.Errorf("opacity did not survive round-trip:\n%s", rs)
	}
}

// TestWithLogger_Builder is acceptance criterion 5 (PR #1): the builder emits a
// write-boundary event when a logger is set; no logger is silent.
func TestWithLogger_Builder(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	p := pptx.New(pptx.WithLogger(logger))
	p.AddSlide()
	if _, err := p.WriteToBytes(); err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if !strings.Contains(buf.String(), "prepared deck for write") {
		t.Errorf("builder did not emit a write event:\n%s", buf.String())
	}
	// No logger: silent + no panic.
	p2 := pptx.New()
	p2.AddSlide()
	if _, err := p2.WriteToBytes(); err != nil {
		t.Fatalf("WriteToBytes (no logger): %v", err)
	}
}
