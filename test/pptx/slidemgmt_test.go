package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestOpen_RepopulatesSlides proves an opened deck exposes its slides (G6): the
// model is rebuilt so it can be read, edited, and re-saved losslessly.
func TestOpen_RepopulatesSlides(t *testing.T) {
	p := pptx.New()
	p.AddSlide().AddTextBox(0, 0, 100, 100, "original")
	p.AddSlide()
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	r, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if got := len(r.Slides()); got != 2 {
		t.Fatalf("reopened Slides() = %d, want 2", got)
	}

	// Edit a reopened slide and re-save: the original content survives and the
	// new shape's relationship/embed does not collide with loaded rIds.
	r.Slides()[0].AddShape(pptx.ShapeRect, pptx.Box{X: 0, Y: 0, W: 100, H: 100},
		pptx.WithFill(pptx.SolidFill(pptx.RGB("FF0000"))))
	data2, err := r.WriteToBytes()
	if err != nil {
		t.Fatalf("re-WriteToBytes: %v", err)
	}

	s1 := readZipPart(t, data2, "ppt/slides/slide1.xml")
	if !strings.Contains(s1, "original") || !strings.Contains(s1, "FF0000") {
		t.Errorf("edited slide lost content or new shape:\n%s", s1)
	}
	rep, _ := conformance.ValidateBytes(data2, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if !rep.OK() {
		t.Fatalf("edit-reopen deck failed conformance:\n%s", rep)
	}
}

// TestOpen_AddImageNoCollision proves an image added to a reopened deck that
// already has media gets a fresh part name (the media counter is seeded on
// Open) rather than colliding and being dropped.
func TestOpen_AddImageNoCollision(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	if _, err := s.AddImage(pptx.ImageBytes(pngBytes("first"), "image/png"), imgBox); err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	data, _ := p.WriteToBytes()

	r, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if _, err := r.Slides()[0].AddImage(pptx.ImageBytes(pngBytes("second"), "image/png"), imgBox); err != nil {
		t.Fatalf("AddImage on reopened: %v", err)
	}
	data2, err := r.WriteToBytes()
	if err != nil {
		t.Fatalf("re-WriteToBytes: %v", err)
	}

	names := partNames(t, data2)
	if !names["ppt/media/image1.png"] || !names["ppt/media/image2.png"] {
		t.Errorf("expected both image1.png and image2.png, got: %v", names)
	}
	if got := readZipPart(t, data2, "ppt/media/image2.png"); got != string(pngBytes("second")) {
		t.Errorf("new image bytes not written correctly")
	}
}

// TestOpen_RepopulatesSections proves sections survive a write → Open → write
// cycle (previously dropped because syncSections had nothing to emit).
func TestOpen_RepopulatesSections(t *testing.T) {
	p := pptx.New()
	intro := p.AddSection("Introduction")
	intro.Include(p.AddSlide())
	p.AddSlide()

	data, _ := p.WriteToBytes()
	r, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if got := len(r.Sections()); got == 0 {
		t.Fatalf("reopened Sections() = 0, want the parsed sections")
	}

	data2, _ := r.WriteToBytes()
	pres := readZipPart(t, data2, "ppt/presentation.xml")
	if !strings.Contains(pres, `name="Introduction"`) {
		t.Errorf("section lost on re-save:\n%s", pres)
	}
}

// TestAddSlideAt_Order proves a slide inserted at index 0 is emitted first in
// sldIdLst (presentation order matches the builder).
func TestAddSlideAt_Order(t *testing.T) {
	p := pptx.New()
	p.AddSlide() // slide1.xml — ends up second
	ins, err := p.AddSlideAt(0)
	if err != nil {
		t.Fatalf("AddSlideAt: %v", err)
	}
	ins.AddTextBox(0, 0, 100, 100, "inserted") // slide2.xml — should be first

	data, _ := p.WriteToBytes()
	pres := readZipPart(t, data, "ppt/presentation.xml")
	lst := pres[strings.Index(pres, "<p:sldIdLst>"):strings.Index(pres, "</p:sldIdLst>")]
	// slide2.xml is wired as rId3 (the inserted slide); it must appear before
	// slide1's rId2.
	i3, i2 := strings.Index(lst, `r:id="rId3"`), strings.Index(lst, `r:id="rId2"`)
	if i3 < 0 || i2 < 0 || i3 > i2 {
		t.Errorf("inserted slide not first in sldIdLst: %s", lst)
	}
}

// TestRemoveSlide_NoDanglingRel proves removing a slide drops its
// presentation→slide relationship (so the deck stays conformant) and its notes
// part.
func TestRemoveSlide_NoDanglingRel(t *testing.T) {
	p := pptx.New()
	s0 := p.AddSlide()
	s0.SetSpeakerNotes("notes for slide 0")
	p.AddSlide()

	if err := p.RemoveSlide(0); err != nil {
		t.Fatalf("RemoveSlide: %v", err)
	}

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("deck has a dangling relationship after RemoveSlide:\n%s", rep)
	}
	if names := partNames(t, data); names["ppt/notesSlides/notesSlide1.xml"] {
		t.Errorf("removed slide's notes part still present")
	}
}
