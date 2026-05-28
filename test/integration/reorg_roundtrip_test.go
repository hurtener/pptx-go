// Package integration holds cross-subsystem tests that exercise real drivers
// on a seam (CLAUDE.md §17). reorg_roundtrip_test is the Phase 01 spot-check:
// it drives the builder (pptx) → OOXML codecs (internal/ooxml) → OPC writer
// (internal/opc) and back, proving the reorg preserved the author→save→reopen
// path end to end.
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
	"github.com/hurtener/pptx-go/pptx"
)

func TestReorg_RoundTrip(t *testing.T) {
	// Author a 2-slide deck with a shape, exercising the slide + presentation
	// codecs that moved into internal/ooxml.
	pres := pptx.New()
	s1 := pres.AddSlide()
	s1.AddRectangle(914400, 914400, 2743200, 1371600)
	s1.AddTextBox(914400, 2743200, 4572000, 914400, "reorg round-trip")
	pres.AddSlide()

	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("WriteToBytes returned no bytes")
	}

	// 1. Reopen through the builder: it parses our own output without error.
	//    (Full builder-model reconstruction — e.g. slide count — is the
	//    round-trip read milestone of a later phase, not this move.)
	if _, err := pptx.NewFromBytes(data); err != nil {
		t.Fatalf("NewFromBytes on self-authored deck: %v", err)
	}

	// 2. Reopen through the OPC layer (real ZIP + content-types + rels) and
	//    assert the expected parts are present — the Phase 01 write-path
	//    invariant through the reorganized OOXML + OPC codecs.
	tmp := filepath.Join(t.TempDir(), "roundtrip.pptx")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		t.Fatal(err)
	}
	pkg, err := opc.OpenFile(tmp)
	if err != nil {
		t.Fatalf("opc.OpenFile: %v", err)
	}
	defer func() { _ = pkg.Close() }()

	for _, uri := range []string{
		"/ppt/presentation.xml",
		"/ppt/slides/slide1.xml",
		"/ppt/slides/slide2.xml",
	} {
		if !pkg.ContainsPart(opc.NewPackURI(uri)) {
			t.Errorf("round-tripped package missing part %s", uri)
		}
	}
}
