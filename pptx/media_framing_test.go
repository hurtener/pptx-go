package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// tinyPNG is a minimal byte slice with a valid PNG signature (AddImage verifies
// the signature, not the pixels).
func tinyPNG() []byte {
	return append([]byte("\x89PNG\r\n\x1a\n"), []byte("framing")...)
}

// TestImageFraming_RoundTrip is R13.11 acceptance 1 (D-114): a builder image with
// a radius + elevation token emits a roundRect-clipped pic with a drop shadow,
// and both survive a write → reopen → re-write cycle (G6, structural).
func TestImageFraming_RoundTrip(t *testing.T) {
	p := pptx.New()
	sl := p.AddSlide("")
	img, err := sl.AddImage(pptx.ImageBytes(tinyPNG(), "image/png"), pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(4), H: pptx.In(3)})
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	img.SetCornerRadius(pptx.RadiusMD).SetElevation(pptx.ElevationRaised)

	xml := slideXML(t, p)
	if !strings.Contains(xml, `prst="roundRect"`) {
		t.Errorf("framed image missing roundRect:\n%s", xml)
	}
	if !strings.Contains(xml, "<a:outerShdw") {
		t.Errorf("framed image missing outerShdw:\n%s", xml)
	}

	// Reopen + re-write: the framing must persist losslessly.
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	rt := slideXML(t, re)
	if !strings.Contains(rt, `prst="roundRect"`) || !strings.Contains(rt, "<a:outerShdw") {
		t.Errorf("image framing did not survive round-trip:\n%s", rt)
	}
}

// TestImageFraming_ZeroByteIdentical is R13.11 acceptance 3 (D-114): an image with
// RadiusNone/ElevationFlat is byte-identical to an image with no framing calls.
func TestImageFraming_ZeroByteIdentical(t *testing.T) {
	build := func(frame bool) []byte {
		p := pptx.New()
		sl := p.AddSlide("")
		img, err := sl.AddImage(pptx.ImageBytes(tinyPNG(), "image/png"), pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(4), H: pptx.In(3)})
		if err != nil {
			t.Fatalf("AddImage: %v", err)
		}
		if frame {
			img.SetCornerRadius(pptx.RadiusNone).SetElevation(pptx.ElevationFlat)
		}
		data, err := p.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}
	if a, b := build(true), build(false); string(a) != string(b) {
		t.Errorf("zero-token framing not byte-identical (%d vs %d bytes)", len(a), len(b))
	}
}
