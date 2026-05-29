package integration

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestConformance_RichText gates a deck exercising the full rich-text surface
// together — a styled+hyperlinked+bulleted text frame plus rich speaker notes —
// proving the text→relationships seam (hyperlink + notes rels mirrored onto the
// package parts) stays structurally sound (D-031, Phase 04).
func TestConformance_RichText(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()

	tf := s.AddTextFrame(pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(8), H: pptx.In(4)})
	tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter}).
		AddRun("Heading", pptx.RunStyle{TypeRole: pptx.TypeH1, Bold: true})
	body := tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletDisc})
	body.AddRun("see ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	body.AddHyperlink("the docs", "https://example.com", pptx.RunStyle{TypeRole: pptx.TypeBody})
	body.AddRun(" and run ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	body.AddRun("go build", pptx.RunStyle{TypeRole: pptx.TypeBody, Code: true})

	notes := s.SpeakerNotes()
	notes.AddParagraph(pptx.ParagraphOpts{}).AddRun("Remember the demo.", pptx.RunStyle{TypeRole: pptx.TypeBody})

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

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
		t.Fatalf("rich-text deck failed conformance:\n%s", rep)
	}
}
