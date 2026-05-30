package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestSpeakerNotes_RoundTrip is acceptance criterion 6: speaker-notes text is
// emitted into a notesSlide part (with a notes master wired from
// presentation.xml) and survives a write → Open → write cycle.
func TestSpeakerNotes_RoundTrip(t *testing.T) {
	const notes = "Talking point: keep it under five minutes."

	p := pptx.New()
	s := p.AddSlide()
	s.AddRectangle(914400, 914400, 2743200, 1371600)
	s.SetSpeakerNotes(notes)

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	// Complete + valid, with the notes parts present.
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{
			"/ppt/presentation.xml",
			"/ppt/slides/slide1.xml",
			"/ppt/notesSlides/notesSlide1.xml",
			"/ppt/notesMasters/notesMaster1.xml",
		},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("notes deck failed conformance:\n%s", rep)
	}

	// The notes text is in the notesSlide; presentation.xml wires the master;
	// the slide relates to its notesSlide.
	if got := readZipPart(t, data, "ppt/notesSlides/notesSlide1.xml"); !strings.Contains(got, notes) {
		t.Errorf("notesSlide1.xml missing notes text in:\n%s", got)
	}
	if pres := readZipPart(t, data, "ppt/presentation.xml"); !strings.Contains(pres, "<p:notesMasterIdLst>") {
		t.Errorf("presentation.xml missing notesMasterIdLst:\n%s", pres)
	}
	// CT_NotesMasterIdList entries are <p:notesMasterId r:id="…"/> with NO id
	// attribute — not <p:sldMasterId> (a schema violation PowerPoint repairs).
	if pres := readZipPart(t, data, "ppt/presentation.xml"); !strings.Contains(pres, "<p:notesMasterId r:id=") {
		t.Errorf("notesMasterIdLst entry is not <p:notesMasterId r:id=…>:\n%s", pres)
	} else if i := strings.Index(pres, "<p:notesMasterIdLst>"); i >= 0 && strings.Contains(pres[i:strings.Index(pres, "</p:notesMasterIdLst>")], "sldMasterId") {
		t.Errorf("notesMasterIdLst contains a sldMasterId (wrong element):\n%s", pres)
	}
	if rels := readZipPart(t, data, "ppt/slides/_rels/slide1.xml.rels"); !strings.Contains(rels, "../notesSlides/notesSlide1.xml") {
		t.Errorf("slide1 rels missing the notesSlide relationship:\n%s", rels)
	}

	// Round-trip: reopen and re-save; the notes survive (pptx.Open).
	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	data2, err := reopened.WriteToBytes()
	if err != nil {
		t.Fatalf("re-WriteToBytes: %v", err)
	}
	if got := readZipPart(t, data2, "ppt/notesSlides/notesSlide1.xml"); !strings.Contains(got, notes) {
		t.Errorf("notes text lost across Open round-trip:\n%s", got)
	}
}

// TestSpeakerNotes_None confirms a deck with no notes emits no notes parts.
func TestSpeakerNotes_None(t *testing.T) {
	p := pptx.New()
	p.AddSlide()
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	names := partNames(t, data)
	if names["ppt/notesMasters/notesMaster1.xml"] || names["ppt/notesSlides/notesSlide1.xml"] {
		t.Errorf("notes parts emitted for a deck with no speaker notes")
	}
}
