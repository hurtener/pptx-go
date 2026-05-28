package integration

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestConformance_BuilderOutput gates the structural soundness (OPC layer,
// D-031) of decks the public builder emits: content-type coverage, resolved
// relationships, no dangling rIds, valid pack URIs.
//
// RequiredParts (full-deck completeness: master/layout/theme/etc.) is NOT
// asserted here — the builder is reorganized but not yet rewritten, so a
// complete deck is a Phase 03 deliverable. Phase 03 turns the completeness
// gate (and the xmllint/LibreOffice layers) on.
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

	rep, err := conformance.ValidateBytes(data, conformance.Options{})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("builder output failed OPC conformance:\n%s", rep)
	}
}
