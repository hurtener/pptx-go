package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// triangleSVG is a minimal valid single-path filled icon.
const triangleSVG = `<svg viewBox="0 0 24 24"><path d="M12 2 L22 22 L2 22 Z" fill="black"/></svg>`

// TestAddIcon_EmitsCustGeom is acceptance criterion 1: AddIcon renders a native
// custom-geometry path shape (not a pic) filled with the accent token.
func TestAddIcon_EmitsCustGeom(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	if _, err := s.AddIcon([]byte(triangleSVG), pptx.Box{X: 914400, Y: 914400, W: 457200, H: 457200}); err != nil {
		t.Fatalf("AddIcon: %v", err)
	}
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
		t.Fatalf("icon deck failed conformance:\n%s", rep)
	}

	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{"<a:custGeom>", "<a:pathLst>", "<a:path ", "<a:moveTo>", "<a:pt ", "2563EB"} {
		if !strings.Contains(slide, want) {
			t.Errorf("icon slide missing %q in:\n%s", want, slide)
		}
	}
	if strings.Contains(slide, "<p:pic>") {
		t.Errorf("icon unexpectedly rendered as a picture:\n%s", slide)
	}
	if strings.Contains(slide, "<a:prstGeom") {
		t.Errorf("icon unexpectedly used preset geometry:\n%s", slide)
	}
}

// TestAddIcon_RoundTrip is acceptance criterion 2/6: an icon deck reopens and
// re-emits its custom geometry, and two independent renders are byte-identical.
func TestAddIcon_RoundTrip(t *testing.T) {
	build := func() []byte {
		p := pptx.New()
		s := p.AddSlide()
		if _, err := s.AddIcon([]byte(triangleSVG), pptx.Box{X: 914400, Y: 914400, W: 457200, H: 457200}); err != nil {
			t.Fatalf("AddIcon: %v", err)
		}
		data, err := p.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}

	a := build()
	if b := build(); string(a) != string(b) {
		t.Fatalf("icon render is not byte-identical (%d vs %d bytes)", len(a), len(b))
	}

	reopened, err := pptx.NewFromBytes(a)
	if err != nil {
		t.Fatalf("reopen icon deck: %v", err)
	}
	if reopened.SlideCount() != 1 {
		t.Errorf("reopened slide count = %d, want 1", reopened.SlideCount())
	}
	again, err := reopened.WriteToBytes()
	if err != nil {
		t.Fatalf("re-save reopened deck: %v", err)
	}
	if slide := readZipPart(t, again, "ppt/slides/slide1.xml"); !strings.Contains(slide, "<a:custGeom>") {
		t.Errorf("custom geometry did not survive round-trip:\n%s", slide)
	}
}

// TestAddIcon_Invalid is acceptance criterion 4: an SVG outside the subset is
// rejected by AddIcon and by ValidateIcon (registration-time failure).
func TestAddIcon_Invalid(t *testing.T) {
	bad := []byte(`<svg viewBox="0 0 24 24"><path d="M0 0 A5 5 0 0 1 10 10"/></svg>`)
	p := pptx.New()
	s := p.AddSlide()
	if _, err := s.AddIcon(bad, pptx.Box{W: 457200, H: 457200}); err == nil {
		t.Error("AddIcon accepted an arc SVG; want an error")
	}
	if err := pptx.ValidateIcon(bad); err == nil {
		t.Error("ValidateIcon accepted an arc SVG; want an error")
	}
	if err := pptx.ValidateIcon([]byte(triangleSVG)); err != nil {
		t.Errorf("ValidateIcon rejected a valid icon: %v", err)
	}
}
