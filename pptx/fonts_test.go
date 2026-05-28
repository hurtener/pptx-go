package pptx_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
	"github.com/hurtener/pptx-go/pptx"
)

// stubFontSource returns fixed bytes for any font (or an error when empty).
type stubFontSource struct {
	data []byte
	err  error
}

func (s stubFontSource) Resolve(name, style string, weight int) ([]byte, error) {
	return s.data, s.err
}

func reopenPackage(t *testing.T, p *pptx.Presentation) *opc.Package {
	t.Helper()
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	pkg, err := opc.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	return pkg
}

func TestEmbedFontShipsBytesAndDeclaration(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	pres.SetFontSource(stubFontSource{data: []byte("FAKE-TTF-BYTES")})

	if err := pres.EmbedFont("Inter", "regular", 400); err != nil {
		t.Fatalf("EmbedFont: %v", err)
	}
	if err := pres.EmbedFont("Inter", "bold", 700); err != nil {
		t.Fatalf("EmbedFont bold: %v", err)
	}

	pkg := reopenPackage(t, pres)
	defer func() { _ = pkg.Close() }()

	// The font-data bytes shipped in the package (acceptance §11.5).
	fontPart := pkg.GetPartByStr("/ppt/fonts/font1.fntdata")
	if fontPart == nil {
		t.Fatal("font1.fntdata not present after round-trip")
	}
	if !bytes.Equal(fontPart.Blob(), []byte("FAKE-TTF-BYTES")) {
		t.Errorf("font bytes not preserved: got %q", fontPart.Blob())
	}

	// presentation.xml declares the embedded font.
	presPart := pkg.GetPartByStr("/ppt/presentation.xml")
	if presPart == nil {
		t.Fatal("presentation.xml missing")
	}
	xml := string(presPart.Blob())
	if !strings.Contains(xml, "embeddedFontLst") {
		t.Error("presentation.xml has no embeddedFontLst")
	}
	if !strings.Contains(xml, `typeface="Inter"`) {
		t.Errorf("embeddedFontLst missing Inter typeface:\n%s", xml)
	}
}

func TestNoEmbedShipsNoFonts(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	pres.SetFontSource(stubFontSource{data: []byte("X")}) // registered but unused

	pkg := reopenPackage(t, pres)
	defer func() { _ = pkg.Close() }()

	if pkg.GetPartByStr("/ppt/fonts/font1.fntdata") != nil {
		t.Error("no EmbedFont call, yet a font part was shipped (acceptance §11.6)")
	}
	presPart := pkg.GetPartByStr("/ppt/presentation.xml")
	if presPart != nil && strings.Contains(string(presPart.Blob()), "embeddedFontLst") {
		t.Error("no EmbedFont call, yet presentation.xml declares embedded fonts")
	}
}

func TestEmbedFontNoSource(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	if err := pres.EmbedFont("Inter", "regular", 400); !errors.Is(err, pptx.ErrNoFontSource) {
		t.Fatalf("expected ErrNoFontSource, got %v", err)
	}
}

func TestEmbedFontNotFound(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	pres.SetFontSource(stubFontSource{err: pptx.ErrFontNotFound})
	if err := pres.EmbedFont("Missing", "regular", 400); !errors.Is(err, pptx.ErrFontNotFound) {
		t.Fatalf("expected ErrFontNotFound, got %v", err)
	}
	// Empty bytes also count as not-found.
	pres.SetFontSource(stubFontSource{data: nil})
	if err := pres.EmbedFont("Empty", "regular", 400); !errors.Is(err, pptx.ErrFontNotFound) {
		t.Fatalf("expected ErrFontNotFound for empty bytes, got %v", err)
	}
}
