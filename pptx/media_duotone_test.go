package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestImageDuotone_RoundTrip is R14.1 acceptance (duotone deterministic, G6): a
// builder image given two-tone shadow/highlight colors emits an <a:duotone> blip
// effect with both resolved srgbClr values, and the recolor survives a write →
// reopen → re-write cycle.
func TestImageDuotone_RoundTrip(t *testing.T) {
	const shadow, highlight = "0B1F3A", "E8F0FF"
	p := pptx.New()
	sl := p.AddSlide("")
	img, err := sl.AddImage(pptx.ImageBytes(tinyPNG(), "image/png"), pptx.Box{X: pptx.In(0), Y: pptx.In(0), W: pptx.In(10), H: pptx.In(7.5)})
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	img.SetDuotone(pptx.RGB(shadow), pptx.RGB(highlight))

	// Read accessor reflects the two tones.
	gotS, gotH, ok := img.Duotone()
	if !ok || string(gotS) != shadow || string(gotH) != highlight {
		t.Errorf("Duotone() = (%q,%q,%v), want (%q,%q,true)", gotS, gotH, ok, shadow, highlight)
	}

	xml := slideXML(t, p)
	if !strings.Contains(xml, "<a:duotone>") {
		t.Errorf("duotone image missing <a:duotone>:\n%s", xml)
	}
	if !strings.Contains(xml, shadow) || !strings.Contains(xml, highlight) {
		t.Errorf("duotone image missing tone colors %s/%s:\n%s", shadow, highlight, xml)
	}

	// Reopen + re-write: the duotone must persist losslessly.
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	rt := slideXML(t, re)
	if !strings.Contains(rt, "<a:duotone>") || !strings.Contains(rt, shadow) || !strings.Contains(rt, highlight) {
		t.Errorf("duotone did not survive round-trip:\n%s", rt)
	}
}

// TestImageDuotone_TokenResolves verifies SetDuotone resolves theme tokens to the
// active theme's palette (P2): the emitted colors equal the resolved accent and
// canvas, and a theme swap would re-tint.
func TestImageDuotone_TokenResolves(t *testing.T) {
	th := pptx.NewTheme(pptx.WithAccent("123456"))
	p := pptx.New(pptx.WithTheme(th))
	sl := p.AddSlide("")
	img, err := sl.AddImage(pptx.ImageBytes(tinyPNG(), "image/png"), pptx.Box{X: 0, Y: 0, W: pptx.In(4), H: pptx.In(3)})
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	img.SetDuotone(pptx.TokenColor(pptx.ColorAccent), pptx.TokenColor(pptx.ColorCanvas))
	s, h, ok := img.Duotone()
	if !ok || string(s) != "123456" {
		t.Errorf("token duotone shadow = %q (ok=%v), want resolved accent 123456", s, ok)
	}
	if string(h) != "FFFFFF" {
		t.Errorf("token duotone highlight = %q, want canvas FFFFFF", h)
	}
}

// TestImageDuotone_NilByteIdentical verifies that an image with no duotone (or a
// nil tone) is byte-identical to one that never calls SetDuotone.
func TestImageDuotone_NilByteIdentical(t *testing.T) {
	build := func(call bool) []byte {
		p := pptx.New()
		sl := p.AddSlide("")
		img, err := sl.AddImage(pptx.ImageBytes(tinyPNG(), "image/png"), pptx.Box{X: 0, Y: 0, W: pptx.In(4), H: pptx.In(3)})
		if err != nil {
			t.Fatalf("AddImage: %v", err)
		}
		if call {
			img.SetDuotone(nil, pptx.RGB("FFFFFF")) // nil shadow → no-op
		}
		data, err := p.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}
	if a, b := build(false), build(true); len(a) != len(b) {
		t.Errorf("nil-tone duotone not byte-identical (%d vs %d bytes)", len(a), len(b))
	}
	if _, _, ok := (&pptx.Image{}).Duotone(); ok {
		t.Errorf("zero-value Image.Duotone() ok = true, want false")
	}
}
