package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// renderFramedDeck builds and renders a single-slide deck with a browser-framed
// image, returning the saved bytes. It is called twice to prove determinism.
func renderFramedDeck(t *testing.T) []byte {
	t.Helper()
	png := append([]byte("\x89PNG\r\n\x1a\n"), []byte("framed-shot")...)
	resolver := scene.URIAssetResolver(func(string) ([]byte, string, error) {
		return png, "image/png", nil
	})
	sc := scene.Scene{
		Meta: scene.Metadata{Title: "Framed"},
		Slides: []scene.SceneSlide{{
			ID:    "shot",
			Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://hero", Alt: "product UI", Frame: scene.FrameBrowser}},
		}},
	}
	pres := pptx.New()
	if _, err := scene.Render(pres, sc, scene.WithAssetResolver(resolver)); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestFrameImage_RoundTripAndDeterminism gates the Phase 10 seam: a framed-image
// deck is OOXML-conformant, re-renders byte-identically (D-035), and reopens
// through pptx round-trip with the picture, its alt text, and the bezel shapes
// intact. Deps name Phase 09 (a different subsystem's shipped phase) and this
// phase opens the curated-asset extension seam Phases 12–13 build on (§17).
func TestFrameImage_RoundTripAndDeterminism(t *testing.T) {
	data := renderFramedDeck(t)

	// Conformance: the framed deck and its embedded media validate (D-031).
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{
			"/ppt/presentation.xml",
			"/ppt/slides/slide1.xml",
			"/ppt/media/image1.png",
		},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("framed-image deck failed conformance:\n%s", rep)
	}

	// Byte-identical idempotency (D-035): a second independent render matches.
	if again := renderFramedDeck(t); !bytes.Equal(data, again) {
		t.Fatalf("framed-image render is not byte-identical (%d vs %d bytes)", len(data), len(again))
	}

	// The slide carries the bezel (native shapes), the picture, and the alt text.
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<p:sp>") {
		t.Errorf("framed slide missing bezel shapes:\n%s", slide)
	}
	if !strings.Contains(slide, "<p:pic>") || !strings.Contains(slide, "r:embed=") {
		t.Errorf("framed slide missing pic/embed:\n%s", slide)
	}
	if !strings.Contains(slide, "product UI") {
		t.Errorf("framed slide missing alt text:\n%s", slide)
	}

	// Round-trip: the self-authored framed deck reopens cleanly (RFC §16, G6).
	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("reopen framed deck: %v", err)
	}
	if reopened.SlideCount() != 1 {
		t.Errorf("reopened slide count = %d, want 1", reopened.SlideCount())
	}
}
