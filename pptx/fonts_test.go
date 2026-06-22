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
	// The <p:font> child must carry the p: prefix in the emitted bytes — a bare
	// <font> is invalid OOXML and PowerPoint cannot bind the embedded face (the
	// reader matches by local name and would hide this).
	if !strings.Contains(xml, "<p:font typeface=") {
		t.Errorf("embedded font element not p:-prefixed (bare <font> is invalid):\n%s", xml)
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

// ----------------------------------------------------------------------------
// Phase 36 — font fallback chain (R9.6, D-066)
// ----------------------------------------------------------------------------

// fallbackTheme builds a theme whose TypeH1 family is heading with the given
// fallback chain and body role family.
func fallbackTheme(heading, body string, fallback ...string) *pptx.Theme {
	theme := pptx.NewTheme(pptx.WithFonts(heading, body))
	spec := theme.Typography[pptx.TypeH1]
	spec.Fallback = fallback
	theme.Typography[pptx.TypeH1] = spec
	return theme
}

// fallbackDeck builds a one-slide deck with an H1 run in the heading face.
func fallbackDeck(t *testing.T, theme *pptx.Theme, opts ...pptx.Option) *pptx.Presentation {
	t.Helper()
	all := append([]pptx.Option{pptx.WithTheme(theme)}, opts...)
	pres := pptx.New(all...)
	s := pres.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Title", pptx.RunStyle{TypeRole: pptx.TypeH1})
	return pres
}

func slideXML(t *testing.T, pres *pptx.Presentation) string {
	t.Helper()
	pkg := reopenPackage(t, pres)
	defer func() { _ = pkg.Close() }()
	part := pkg.GetPartByStr("/ppt/slides/slide1.xml")
	if part == nil {
		t.Fatal("slide1.xml missing")
	}
	return string(part.Blob())
}

func TestFontFallbackSubstitutesUnavailablePrimary(t *testing.T) {
	// Source has the fallback (Georgia) but not the primary (Playfair Display).
	src := mapFontSource{"Georgia": []byte("GEORGIA")}
	pres := fallbackDeck(t, fallbackTheme("Playfair Display", "Inter", "Georgia"),
		pptx.WithFontSource(src))

	xml := slideXML(t, pres)
	if !strings.Contains(xml, `typeface="Georgia"`) {
		t.Errorf("run not substituted to the fallback face:\n%s", xml)
	}
	if strings.Contains(xml, `typeface="Playfair Display"`) {
		t.Errorf("unavailable primary still emitted:\n%s", xml)
	}
	// Embedding off: no font parts shipped.
	if got := countFontParts(t, pres); got != 0 {
		t.Errorf("fallback (embedding off) shipped %d font parts, want 0", got)
	}
}

func TestFontFallbackPrimaryWinsWhenAvailable(t *testing.T) {
	// Source resolves the primary — it must win, no substitution.
	src := mapFontSource{"Playfair Display": []byte("PLAYFAIR"), "Georgia": []byte("GEORGIA")}
	pres := fallbackDeck(t, fallbackTheme("Playfair Display", "Inter", "Georgia"),
		pptx.WithFontSource(src))

	xml := slideXML(t, pres)
	if !strings.Contains(xml, `typeface="Playfair Display"`) {
		t.Errorf("available primary not kept:\n%s", xml)
	}
	if strings.Contains(xml, `typeface="Georgia"`) {
		t.Errorf("substituted to fallback though the primary resolved:\n%s", xml)
	}
}

func TestFontFallbackByteIdenticalWhenUnused(t *testing.T) {
	baseline, err := fallbackDeck(t, fallbackTheme("Playfair Display", "Inter")).WriteToBytes()
	if err != nil {
		t.Fatalf("baseline: %v", err)
	}
	// (a) Fallback declared but NO FontSource → byte-identical.
	noSrc, err := fallbackDeck(t, fallbackTheme("Playfair Display", "Inter", "Georgia")).WriteToBytes()
	if err != nil {
		t.Fatalf("no-source: %v", err)
	}
	if !bytes.Equal(baseline, noSrc) {
		t.Error("a declared fallback with no FontSource changed output")
	}
	// (b) FontSource registered but NO fallback declared → byte-identical.
	src := mapFontSource{"Georgia": []byte("GEORGIA")}
	noFb, err := fallbackDeck(t, fallbackTheme("Playfair Display", "Inter"), pptx.WithFontSource(src)).WriteToBytes()
	if err != nil {
		t.Fatalf("no-fallback: %v", err)
	}
	if !bytes.Equal(baseline, noFb) {
		t.Error("a registered FontSource with no declared fallback changed output")
	}
}

func TestFontFallbackDeterministicIdempotent(t *testing.T) {
	src := mapFontSource{"Georgia": []byte("GEORGIA")}
	pres := fallbackDeck(t, fallbackTheme("Playfair Display", "Inter", "Georgia"),
		pptx.WithFontSource(src))
	a, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("save 1: %v", err)
	}
	b, err := pres.WriteToBytes() // second save of the same (mutated) deck
	if err != nil {
		t.Fatalf("save 2: %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Errorf("two saves differ (not idempotent): %d vs %d bytes", len(a), len(b))
	}
}

// ----------------------------------------------------------------------------
// Phase 37 — italic-aware font fallback (R9.7, D-067)
// ----------------------------------------------------------------------------

// styleFontSource resolves per (family, style) — so it can model a family that
// ships a regular cut but no italic cut. Key is "family|style" (style "" or
// "italic").
type styleFontSource map[string][]byte

func (m styleFontSource) Resolve(name, style string, weight int) ([]byte, error) {
	if b, ok := m[name+"|"+style]; ok {
		return b, nil
	}
	return nil, pptx.ErrFontNotFound
}

func TestEmphasisItalicFallback(t *testing.T) {
	// "Display" ships regular but NOT italic; "Georgia" ships italic.
	src := styleFontSource{
		"Display|":       []byte("DISPLAY-REG"),
		"Georgia|italic": []byte("GEORGIA-ITALIC"),
		"Georgia|":       []byte("GEORGIA-REG"),
	}
	theme := fallbackTheme("Display", "Inter", "Georgia")
	pres := pptx.New(pptx.WithTheme(theme), pptx.WithFontSource(src))
	s := pres.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	p := tf.AddParagraph(pptx.ParagraphOpts{})
	p.AddRun("Heading ", pptx.RunStyle{TypeRole: pptx.TypeH1})               // upright
	p.AddRun("emphasis", pptx.RunStyle{TypeRole: pptx.TypeH1, Italic: true}) // italic

	xml := slideXML(t, pres)
	// Upright run keeps the primary (its regular cut resolves)…
	if !strings.Contains(xml, `typeface="Display"`) {
		t.Errorf("upright run did not keep the primary face:\n%s", xml)
	}
	// …while the italic run falls back to a face with an italic cut.
	if !strings.Contains(xml, `typeface="Georgia"`) {
		t.Errorf("italic run did not fall back to an italic-capable face:\n%s", xml)
	}
}

func TestEmphasisDisplayItalicEmbedded(t *testing.T) {
	// The already-satisfied guarantee (D-063 + D-065): an italic run at the
	// display role embeds the display face's italic cut.
	src := styleFontSource{"Cardo|": []byte("CARDO-REG"), "Cardo|italic": []byte("CARDO-ITALIC")}
	theme := pptx.NewTheme(pptx.WithDisplayFont("Cardo"))
	pres := pptx.New(pptx.WithTheme(theme), pptx.WithFontSource(src), pptx.WithFontEmbedding())
	s := pres.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	tf.AddParagraph(pptx.ParagraphOpts{}).
		AddRun("Editorial", pptx.RunStyle{TypeRole: pptx.TypeDisplay, Italic: true})

	xml := presXML(t, pres)
	if !strings.Contains(xml, `typeface="Cardo"`) {
		t.Errorf("display face not embedded:\n%s", xml)
	}
	// The display role is bold by default, so an italic display run embeds the
	// boldItalic bucket — either way an italic cut is present.
	if !strings.Contains(strings.ToLower(xml), "italic") {
		t.Errorf("italic cut not present in embeddedFontLst:\n%s", xml)
	}
}

func TestFontFallbackEmbedsResolvedFace(t *testing.T) {
	// Fallback + embedding: the resolved fallback face is embedded, not the primary.
	src := mapFontSource{"Georgia": []byte("GEORGIA"), "Inter": []byte("INTER")}
	pres := fallbackDeck(t, fallbackTheme("Playfair Display", "Inter", "Georgia"),
		pptx.WithFontSource(src), pptx.WithFontEmbedding())

	if got := countFontParts(t, pres); got != 1 {
		t.Fatalf("font parts = %d, want 1 (Georgia)", got)
	}
	xml := presXML(t, pres)
	if !strings.Contains(xml, `typeface="Georgia"`) {
		t.Errorf("resolved fallback face not embedded:\n%s", xml)
	}
	if strings.Contains(xml, `typeface="Playfair Display"`) {
		t.Errorf("unavailable primary embedded:\n%s", xml)
	}
}
