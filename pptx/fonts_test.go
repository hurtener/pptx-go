package pptx_test

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
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

// ----------------------------------------------------------------------------
// Phase 35 — automatic font-embedding pass (R9.1, D-065)
// ----------------------------------------------------------------------------

// mapFontSource resolves per-family bytes; an unknown family returns
// ErrFontNotFound so the warn-don't-fail path is exercised.
type mapFontSource map[string][]byte

func (m mapFontSource) Resolve(name, style string, weight int) ([]byte, error) {
	if b, ok := m[name]; ok {
		return b, nil
	}
	return nil, pptx.ErrFontNotFound
}

// themedDeck builds a one-slide deck whose H1 run resolves to heading and body
// run to body (via WithFonts), applying the given extra options.
func themedDeck(t *testing.T, heading, body string, opts ...pptx.Option) *pptx.Presentation {
	t.Helper()
	theme := pptx.NewTheme(pptx.WithFonts(heading, body))
	all := append([]pptx.Option{pptx.WithTheme(theme)}, opts...)
	pres := pptx.New(all...)
	s := pres.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	p := tf.AddParagraph(pptx.ParagraphOpts{})
	p.AddRun("Title", pptx.RunStyle{TypeRole: pptx.TypeH1})
	p.AddRun("body text", pptx.RunStyle{TypeRole: pptx.TypeBody})
	return pres
}

func countFontParts(t *testing.T, pres *pptx.Presentation) int {
	t.Helper()
	pkg := reopenPackage(t, pres)
	defer func() { _ = pkg.Close() }()
	n := 0
	for i := 1; ; i++ {
		if pkg.GetPartByStr(fmt.Sprintf("/ppt/fonts/font%d.fntdata", i)) == nil {
			break
		}
		n++
	}
	return n
}

func presXML(t *testing.T, pres *pptx.Presentation) string {
	t.Helper()
	pkg := reopenPackage(t, pres)
	defer func() { _ = pkg.Close() }()
	part := pkg.GetPartByStr("/ppt/presentation.xml")
	if part == nil {
		t.Fatal("presentation.xml missing")
	}
	return string(part.Blob())
}

func TestAutoEmbedShipsUsedFaces(t *testing.T) {
	src := mapFontSource{"Playfair Display": []byte("PLAYFAIR"), "Inter": []byte("INTER")}
	pres := themedDeck(t, "Playfair Display", "Inter",
		pptx.WithFontSource(src), pptx.WithFontEmbedding())

	if got := countFontParts(t, pres); got != 2 {
		t.Fatalf("font parts = %d, want 2 (Playfair Display + Inter)", got)
	}
	xml := presXML(t, pres)
	for _, face := range []string{"Playfair Display", "Inter"} {
		if !strings.Contains(xml, `typeface="`+face+`"`) {
			t.Errorf("embeddedFontLst missing %q:\n%s", face, xml)
		}
	}
}

func TestAutoEmbedDeterministic(t *testing.T) {
	build := func() []byte {
		src := mapFontSource{"Playfair Display": []byte("PLAYFAIR"), "Inter": []byte("INTER")}
		pres := themedDeck(t, "Playfair Display", "Inter",
			pptx.WithFontSource(src), pptx.WithFontEmbedding())
		data, err := pres.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}
	if a, b := build(), build(); !bytes.Equal(a, b) {
		t.Errorf("two embedded saves differ: %d vs %d bytes", len(a), len(b))
	}
}

func TestAutoEmbedOffByteIdentical(t *testing.T) {
	src := mapFontSource{"Playfair Display": []byte("PLAYFAIR"), "Inter": []byte("INTER")}
	// Flag off but a source registered.
	off := themedDeck(t, "Playfair Display", "Inter", pptx.WithFontSource(src))
	// No source, no flag — the pre-change baseline.
	plain := themedDeck(t, "Playfair Display", "Inter")

	offBytes, err := off.WriteToBytes()
	if err != nil {
		t.Fatalf("off WriteToBytes: %v", err)
	}
	plainBytes, err := plain.WriteToBytes()
	if err != nil {
		t.Fatalf("plain WriteToBytes: %v", err)
	}
	if !bytes.Equal(offBytes, plainBytes) {
		t.Errorf("flag-off output not byte-identical to the no-source baseline (%d vs %d bytes)", len(offBytes), len(plainBytes))
	}
	if got := countFontParts(t, off); got != 0 {
		t.Errorf("flag off shipped %d font parts, want 0", got)
	}
}

func TestAutoEmbedIdempotentWithManual(t *testing.T) {
	src := mapFontSource{"Playfair Display": []byte("PLAYFAIR"), "Inter": []byte("INTER")}
	pres := pptx.New(pptx.WithTheme(pptx.NewTheme(pptx.WithFonts("Playfair Display", "Inter"))),
		pptx.WithFontSource(src), pptx.WithFontEmbedding())
	s := pres.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	p := tf.AddParagraph(pptx.ParagraphOpts{})
	p.AddRun("Title", pptx.RunStyle{TypeRole: pptx.TypeH1})  // Playfair Display, regular
	p.AddRun("body", pptx.RunStyle{TypeRole: pptx.TypeBody}) // Inter, regular

	// Pre-embed one of the faces by hand; the pass must not duplicate it.
	if err := pres.EmbedFont("Inter", "regular", 400); err != nil {
		t.Fatalf("manual EmbedFont: %v", err)
	}

	// Two distinct faces total (Inter + Playfair); Inter embedded once.
	if got := countFontParts(t, pres); got != 2 {
		t.Fatalf("font parts = %d, want 2 (no duplicate Inter)", got)
	}
	if n := strings.Count(presXML(t, pres), `typeface="Inter"`); n != 1 {
		t.Errorf("Inter declared %d times, want 1 (idempotent)", n)
	}
}

func TestAutoEmbedWarnsOnMissing(t *testing.T) {
	// Source resolves Inter but not Playfair Display.
	src := mapFontSource{"Inter": []byte("INTER")}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	pres := themedDeck(t, "Playfair Display", "Inter",
		pptx.WithFontSource(src), pptx.WithFontEmbedding(), pptx.WithLogger(logger))

	// Save still succeeds, embedding the face that resolved.
	if got := countFontParts(t, pres); got != 1 {
		t.Fatalf("font parts = %d, want 1 (Inter resolved, Playfair missing)", got)
	}
	xml := presXML(t, pres)
	if !strings.Contains(xml, `typeface="Inter"`) {
		t.Error("Inter not embedded though it resolved")
	}
	if strings.Contains(xml, `typeface="Playfair Display"`) {
		t.Error("Playfair Display embedded though the source could not resolve it")
	}
	if log := buf.String(); !strings.Contains(log, "font embedding skipped face") || !strings.Contains(log, "Playfair Display") {
		t.Errorf("missing-face warning not logged:\n%s", log)
	}
}
