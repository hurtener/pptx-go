package integration

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestConformance_BuilderOutput gates the structural soundness (OPC layer,
// D-031) of decks the public builder emits: content-type coverage, resolved
// relationships, no dangling rIds, valid pack URIs, and — since Phase 03
// Chunk A2 — full-deck completeness (master/layout/theme present and wired).
func TestConformance_BuilderOutput(t *testing.T) {
	pres := pptx.New()
	s := pres.AddSlide()
	s.AddRectangle(914400, 914400, 2743200, 1371600)
	s.AddTextBox(914400, 2743200, 4572000, 914400, "conformance")
	pres.AddSlide()

	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	// A complete deck (A2): presentation + slides + master + blank layout +
	// theme, every relationship resolved and every root namespaced (D-032).
	opts := conformance.Options{
		RequiredParts: []string{
			"/ppt/presentation.xml",
			"/ppt/slides/slide1.xml",
			"/ppt/slideMasters/slideMaster1.xml",
			"/ppt/slideLayouts/slideLayout1.xml",
			"/ppt/theme/theme1.xml",
		},
	}
	rep, err := conformance.ValidateBytes(data, opts)
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("builder output failed full-deck OPC conformance:\n%s", rep)
	}
}
