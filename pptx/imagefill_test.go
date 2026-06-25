package pptx_test

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// realPNG builds a valid PNG of the given pixel dimensions (the format header
// carries the dimensions coverSrcRect reads via image.DecodeConfig).
func realPNG(w, h int) []byte {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h)))
	return buf.Bytes()
}

// TestImageFill_Emits is R14.1 part 2 (D-117): a shape with WithImageFill emits an
// <a:blipFill> (not <p:blipFill>) with an embedded image, replacing the solid
// fill, and the fill survives write → reopen → re-write (G6, structural).
func TestImageFill_Emits(t *testing.T) {
	p := pptx.New()
	sl := p.AddSlide("")
	box := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(4), H: pptx.In(3)}
	sl.AddShape(pptx.ShapeRoundRect, box,
		pptx.WithRadius(pptx.RadiusLG),
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurface))),
		pptx.WithImageFill(pptx.ImageBytes(realPNG(800, 600), "image/png")))

	xml := slideXML(t, p)
	if !strings.Contains(xml, "<a:blipFill>") {
		t.Errorf("image-fill shape missing <a:blipFill>:\n%s", xml)
	}
	if strings.Contains(xml, "<p:blipFill>") {
		t.Errorf("shape image fill wrongly emitted as <p:blipFill> (namespace):\n%s", xml)
	}
	if strings.Contains(xml, "<a:solidFill>") {
		t.Errorf("image fill should replace the solid fill, but solidFill present:\n%s", xml)
	}
	// Round-trip: reopen + re-write; the blip fill must persist.
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if rt := slideXML(t, re); !strings.Contains(rt, "<a:blipFill>") {
		t.Errorf("image fill did not survive round-trip:\n%s", rt)
	}
}

// TestImageFill_CoverCrop verifies the cover-fit center-crop: a wide image in a
// square box crops left/right; a tall image crops top/bottom; a matching aspect
// emits no srcRect.
func TestImageFill_CoverCrop(t *testing.T) {
	square := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(3)}
	mk := func(w, h int, box pptx.Box) string {
		p := pptx.New()
		sl := p.AddSlide("")
		sl.AddShape(pptx.ShapeRect, box, pptx.WithImageFill(pptx.ImageBytes(realPNG(w, h), "image/png")))
		return slideXML(t, p)
	}
	// Wide (2:1) into a square → crop left/right (l/r), not top/bottom.
	wide := mk(800, 400, square)
	if !strings.Contains(wide, `<a:srcRect l="25000" r="25000"/>`) {
		t.Errorf("wide cover crop want srcRect l=r=25000 (crop 50%% width):\n%s", wide)
	}
	// Tall (1:2) into a square → crop top/bottom.
	tall := mk(400, 800, square)
	if !strings.Contains(tall, `<a:srcRect t="25000" b="25000"/>`) {
		t.Errorf("tall cover crop want srcRect t=b=25000:\n%s", tall)
	}
	// Matching aspect → no srcRect at all.
	match := mk(900, 900, square)
	if strings.Contains(match, "<a:srcRect") {
		t.Errorf("matching-aspect cover should emit no srcRect:\n%s", match)
	}
}

// TestImageFill_NilNoChange verifies a nil source (or an unreadable image) leaves
// the prior fill in place: a shape with WithImageFill(nil-equivalent) keeps its
// solid fill, byte-identical to one with no image-fill option.
func TestImageFill_NilNoChange(t *testing.T) {
	build := func(withBad bool) []byte {
		p := pptx.New()
		sl := p.AddSlide("")
		box := pptx.Box{X: 0, Y: 0, W: pptx.In(4), H: pptx.In(3)}
		opts := []pptx.ShapeOption{pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurface)))}
		if withBad {
			// A nil ImageSource is a no-op; the solid fill remains.
			opts = append(opts, pptx.WithImageFill(nil))
		}
		sl.AddShape(pptx.ShapeRect, box, opts...)
		data, err := p.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}
	if a, b := build(false), build(true); !bytes.Equal(a, b) {
		t.Errorf("nil ImageFill not byte-identical (%d vs %d bytes)", len(a), len(b))
	}
}
