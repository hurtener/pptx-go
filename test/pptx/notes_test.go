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

// TestSpeakerNotes_ReadBack proves the read accessor reconstructs notes on a
// reopened deck (not just that the bytes pass through), and guards the data-loss
// footgun: inspecting SpeakerNotes() on a reopened deck and then re-saving must
// preserve the notes rather than overwrite them with an empty frame.
func TestSpeakerNotes_ReadBack(t *testing.T) {
	const notes = "Talking point: keep it under five minutes."

	p := pptx.New()
	s := p.AddSlide()
	s.AddRectangle(914400, 914400, 2743200, 1371600)
	s.SetSpeakerNotes(notes)
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	rs := reopened.Slides()
	if len(rs) != 1 {
		t.Fatalf("reopened slides = %d, want 1", len(rs))
	}
	// The read accessor reconstructs a navigable notes TextFrame.
	if !rs[0].HasSpeakerNotes() {
		t.Fatal("HasSpeakerNotes() = false on a reopened deck that had notes")
	}
	paras := rs[0].SpeakerNotes().Paragraphs()
	if len(paras) == 0 || len(paras[0].Runs()) == 0 {
		t.Fatalf("reopened notes have no runs: %d paragraphs", len(paras))
	}
	if got := paras[0].Runs()[0].Text(); got != notes {
		t.Errorf("reopened notes text = %q, want %q", got, notes)
	}

	// Data-loss guard: inspecting then re-saving must keep the notes. (Before the
	// read-back fix, SpeakerNotes() created an empty frame that overwrote them.)
	_ = rs[0].SpeakerNotes()
	data2, err := reopened.WriteToBytes()
	if err != nil {
		t.Fatalf("re-WriteToBytes: %v", err)
	}
	if got := readZipPart(t, data2, "ppt/notesSlides/notesSlide1.xml"); !strings.Contains(got, notes) {
		t.Errorf("notes destroyed by inspect-then-save:\n%s", got)
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

// TestNotesMaster_OwnTheme proves the notes master references its own distinct
// theme part (theme2.xml), not the slide master's theme1.xml. PowerPoint repairs
// a deck whose notes master shares the slide master's theme — it splits off a
// theme2.xml on open — so we seed it ourselves. This guards that regression.
func TestNotesMaster_OwnTheme(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	s.SetSpeakerNotes("notes drive the notes master + its theme")

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	// A distinct notes theme part exists alongside the slide master's theme1.
	names := partNames(t, data)
	if !names["ppt/theme/theme1.xml"] || !names["ppt/theme/theme2.xml"] {
		t.Fatalf("expected both theme1.xml and theme2.xml, got: %v", names)
	}

	// The notes master relates to theme2, not theme1.
	rels := readZipPart(t, data, "ppt/notesMasters/_rels/notesMaster1.xml.rels")
	if !strings.Contains(rels, "../theme/theme2.xml") {
		t.Errorf("notes master does not reference its own theme2.xml:\n%s", rels)
	}
	if strings.Contains(rels, "../theme/theme1.xml") {
		t.Errorf("notes master still shares the slide master's theme1.xml (PowerPoint repairs this):\n%s", rels)
	}

	// theme2.xml carries the theme content type (the package adds the override).
	if ct := readZipPart(t, data, "[Content_Types].xml"); !strings.Contains(ct, `PartName="/ppt/theme/theme2.xml"`) {
		t.Errorf("theme2.xml missing its content-type override:\n%s", ct)
	}
}
