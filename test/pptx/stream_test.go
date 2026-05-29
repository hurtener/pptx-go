package pptx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

func validatePPTXFile(t *testing.T, path string, required ...string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{RequiredParts: required})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("%s failed conformance:\n%s", path, rep)
	}
}

// TestSaveStream_Valid proves the streaming write path emits a complete, valid
// deck (with media, notes and sections) — and applies the same hygiene as Save.
func TestSaveStream_Valid(t *testing.T) {
	p := pptx.New()
	intro := p.AddSection("Intro")
	s := p.AddSlide()
	intro.Include(s)
	s.AddRectangle(914400, 914400, 2743200, 1371600)
	if _, err := s.AddImage(pptx.ImageBytes(pngBytes("stream"), "image/png"), imgBox); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	s.SetSpeakerNotes("streamed notes")

	out := filepath.Join(t.TempDir(), "stream.pptx")
	if err := p.SaveStream(out); err != nil {
		t.Fatalf("SaveStream: %v", err)
	}

	validatePPTXFile(t, out,
		"/ppt/presentation.xml",
		"/ppt/slides/slide1.xml",
		"/ppt/media/image1.png",
		"/ppt/notesSlides/notesSlide1.xml",
		"/ppt/notesMasters/notesMaster1.xml",
		"/ppt/theme/theme1.xml",
	)
}

// TestOpenStream_RoundTrip writes a sectioned deck, reopens it through the
// streaming reader, re-saves it, and confirms it stays valid and keeps its
// sections (an I/O round-trip through the streaming path).
func TestOpenStream_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.pptx")

	p := pptx.New()
	sec := p.AddSection("Chapter One")
	sec.Include(p.AddSlide())
	if err := p.Save(src); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reopened, err := pptx.OpenStream(src)
	if err != nil {
		t.Fatalf("OpenStream: %v", err)
	}
	secs := reopened.PresentationPart().Sections()
	if len(secs) != 1 || secs[0].Name != "Chapter One" {
		t.Fatalf("sections not preserved through OpenStream: %+v", secs)
	}

	dst := filepath.Join(dir, "dst.pptx")
	if err := reopened.SaveStream(dst); err != nil {
		t.Fatalf("SaveStream: %v", err)
	}
	validatePPTXFile(t, dst, "/ppt/presentation.xml", "/ppt/slides/slide1.xml")
}
