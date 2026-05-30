package pptx_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// pngBytes returns bytes with a valid PNG signature plus a payload tag, so two
// payloads hash differently while both sniff as PNG. pptx-go never parses
// pixels (§7); only the signature matters.
func pngBytes(payload string) []byte {
	return append([]byte("\x89PNG\r\n\x1a\n"), []byte(payload)...)
}

func partNames(t *testing.T, data []byte) map[string]bool {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	return names
}

var imgBox = pptx.Box{X: 914400, Y: 914400, W: 2743200, H: 1371600}

// TestAddImage_RoundTrip is acceptance criterion 3: a one-slide deck with a
// rect + image is complete (passes conformance), writes the image bytes once,
// wires the image relationship, and references it from the slide.
func TestAddImage_RoundTrip(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	s.AddRectangle(914400, 914400, 2743200, 1371600)

	png := pngBytes("logo")
	img, err := s.AddImage(pptx.ImageBytes(png, "image/png"), imgBox)
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	img.SetAltText("Acme logo").SetCrop(pptx.Crop{Left: 0.1, Top: 0.2, Right: 0.1, Bottom: 0.2})

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	// The complete deck still passes the full-deck conformance gate, now with a
	// resolvable image relationship.
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{
			"/ppt/presentation.xml",
			"/ppt/slides/slide1.xml",
			"/ppt/media/image1.png",
			"/ppt/theme/theme1.xml",
		},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("image deck failed conformance:\n%s", rep)
	}

	// The image bytes are written verbatim, exactly once.
	if got := readZipPart(t, data, "ppt/media/image1.png"); got != string(png) {
		t.Errorf("media bytes not preserved: got %q want %q", got, string(png))
	}

	// The slide references the embed and carries alt text + crop.
	slideXML := readZipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{
		`r:embed="rId2"`,
		`descr="Acme logo"`,
		`<a:srcRect`,
		// A picture needs a geometry or renderers (Quick Look, Keynote,
		// LibreOffice) draw nothing — the blip has no region to fill.
		`<a:prstGeom prst="rect">`,
	} {
		if !strings.Contains(slideXML, want) {
			t.Errorf("slide1.xml missing %q in:\n%s", want, slideXML)
		}
	}

	// The slide's rels wire both the layout (rId1) and the image (rId2).
	rels := readZipPart(t, data, "ppt/slides/_rels/slide1.xml.rels")
	if !strings.Contains(rels, "../media/image1.png") {
		t.Errorf("slide1 rels missing the image relationship:\n%s", rels)
	}
}

// TestAddImage_Dedup proves identical bytes used on two slides are written once
// (the upstream MediaManager dedup is preserved).
func TestAddImage_Dedup(t *testing.T) {
	p := pptx.New()
	png := pngBytes("shared")

	s1 := p.AddSlide()
	if _, err := s1.AddImage(pptx.ImageBytes(png, "image/png"), imgBox); err != nil {
		t.Fatalf("AddImage s1: %v", err)
	}
	s2 := p.AddSlide()
	if _, err := s2.AddImage(pptx.ImageBytes(png, "image/png"), imgBox); err != nil {
		t.Fatalf("AddImage s2: %v", err)
	}

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	names := partNames(t, data)
	if !names["ppt/media/image1.png"] {
		t.Errorf("expected image1.png to exist")
	}
	if names["ppt/media/image2.png"] {
		t.Errorf("identical bytes were not deduplicated (image2.png exists)")
	}

	// Both slides reference the single media part.
	for _, rels := range []string{
		"ppt/slides/_rels/slide1.xml.rels",
		"ppt/slides/_rels/slide2.xml.rels",
	} {
		if !strings.Contains(readZipPart(t, data, rels), "../media/image1.png") {
			t.Errorf("%s does not reference the shared image", rels)
		}
	}
}

// TestAddImage_Verification covers the §7 security checks: unrecognized bytes
// and a declared/actual MIME mismatch are rejected.
func TestAddImage_Verification(t *testing.T) {
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}

	tests := []struct {
		name string
		src  pptx.ImageSource
		want error
	}{
		{"malformed", pptx.ImageBytes([]byte("not an image"), "image/png"), pptx.ErrUnknownImageFormat},
		{"mime mismatch", pptx.ImageBytes(jpeg, "image/png"), pptx.ErrImageMIMEMismatch},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := pptx.New()
			s := p.AddSlide()
			if _, err := s.AddImage(tc.src, imgBox); !errors.Is(err, tc.want) {
				t.Errorf("AddImage error = %v, want %v", err, tc.want)
			}
		})
	}
}
