package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestSections_RoundTrip is acceptance criterion 5: a section of slides is
// emitted into presentation.xml's p14 section list, keeps the deck valid, and
// round-trips through pptx.Open (the presentation part re-parses the sections).
func TestSections_RoundTrip(t *testing.T) {
	p := pptx.New()
	intro := p.AddSection("Introduction")
	intro.Include(p.AddSlide())

	body := p.AddSection("Body")
	body.Include(p.AddSlide())
	body.Include(p.AddSlide())

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
		t.Fatalf("sectioned deck failed conformance:\n%s", rep)
	}

	pres := readZipPart(t, data, "ppt/presentation.xml")
	for _, want := range []string{
		`<p:ext uri="{521415D9-36F7-43E2-AB2F-B90AF26B5E84}">`,
		`xmlns:p14="http://schemas.microsoft.com/office/powerpoint/2010/main"`,
		`<p14:section name="Introduction"`,
		`<p14:section name="Body"`,
		`<p14:sldId id="257"/>`, // first slide (IDs start at 257)
	} {
		if !strings.Contains(pres, want) {
			t.Errorf("presentation.xml missing %q in:\n%s", want, pres)
		}
	}

	// Round-trip: reopen and confirm the presentation part re-parsed the
	// sections (names + membership).
	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	secs := reopened.PresentationPart().Sections()
	if len(secs) != 2 {
		t.Fatalf("parsed %d sections, want 2: %+v", len(secs), secs)
	}
	if secs[0].Name != "Introduction" || len(secs[0].SlideIDs) != 1 {
		t.Errorf("section[0] = %+v, want Introduction with 1 slide", secs[0])
	}
	if secs[1].Name != "Body" || len(secs[1].SlideIDs) != 2 {
		t.Errorf("section[1] = %+v, want Body with 2 slides", secs[1])
	}
}

// TestSections_DefaultCoversUnassigned proves unassigned slides land in a
// leading implicit default section so the section list spans every slide.
func TestSections_DefaultCoversUnassigned(t *testing.T) {
	p := pptx.New()
	p.AddSlide() // unassigned
	body := p.AddSection("Body")
	body.Include(p.AddSlide())

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	secs := reopened.PresentationPart().Sections()
	if len(secs) != 2 {
		t.Fatalf("parsed %d sections, want 2 (default + Body)", len(secs))
	}
	if secs[0].Name != pptx.DefaultSectionName || len(secs[0].SlideIDs) != 1 {
		t.Errorf("leading default section = %+v, want %q with 1 slide", secs[0], pptx.DefaultSectionName)
	}
}
