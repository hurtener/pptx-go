package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestDarkPalette_SoulDrivenColors is the R8.3 acceptance: a theme carrying a
// soul-driven dark palette renders every VariantDark slide's resolved
// canvas/surface/primary-text to the supplied hexes (reported via stats.Colors,
// D-058) and emits them into the slide XML — instead of the pinned Tailwind grays.
func TestDarkPalette_SoulDrivenColors(t *testing.T) {
	const (
		navyCanvas  = "0A0E1A" // brand deep-navy canvas (not the pinned 111827)
		navySurface = "14182B" // brand dark surface (not the pinned 1F2937)
		darkText    = "F4F6FF" // brand light text (not the pinned F9FAFB)
	)
	th := pptx.NewTheme(
		pptx.WithDarkSurface(pptx.ColorCanvas, navyCanvas),
		pptx.WithDarkSurface(pptx.ColorSurface, navySurface),
		pptx.WithDarkText(pptx.TextPrimary, darkText),
	)
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:      "dark",
		Variant: scene.VariantDark,
		Nodes:   []scene.SlideNode{scene.Heading{Text: rt("Dark slide"), Level: 1}},
	}}}
	data, stats := render(t, sc, scene.WithTheme(th))
	if len(stats.Warnings) != 0 {
		t.Errorf("soul dark palette: unexpected warnings: %+v", stats.Warnings)
	}

	// stats.Colors reports what the slide actually resolved (the dark palette).
	if len(stats.Colors) != 1 {
		t.Fatalf("stats.Colors len = %d, want 1", len(stats.Colors))
	}
	got := stats.Colors[0]
	if string(got.Canvas) != navyCanvas {
		t.Errorf("resolved dark canvas = %q, want %q", got.Canvas, navyCanvas)
	}
	if string(got.Surface) != navySurface {
		t.Errorf("resolved dark surface = %q, want %q", got.Surface, navySurface)
	}
	if string(got.PrimaryText) != darkText {
		t.Errorf("resolved dark primary text = %q, want %q", got.PrimaryText, darkText)
	}

	// The emitted slide XML carries the brand navy canvas, not the pinned gray.
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, navyCanvas) {
		t.Errorf("slide missing brand navy canvas %s:\n%s", navyCanvas, slideXML)
	}
	if strings.Contains(slideXML, "111827") {
		t.Errorf("slide still carries the pinned gray canvas 111827 — soul override did not win:\n%s", slideXML)
	}
}

// TestDarkPalette_NilByteIdentical is the byte-identity guard: a VariantDark
// slide on a theme with NO dark palette is byte-for-byte identical to the same
// slide on the default theme — the pinned-gray fallback is unchanged (R8.3 acc 2).
func TestDarkPalette_NilByteIdentical(t *testing.T) {
	nodes := []scene.SlideNode{scene.Heading{Text: rt("Dark"), Level: 1}}
	sc := scene.Scene{Slides: []scene.SceneSlide{{Variant: scene.VariantDark, Nodes: nodes}}}

	// Default theme (no dark palette) vs an explicitly-constructed theme that
	// also sets no dark palette — both must hit the pinned-gray path.
	dDefault := renderBytes(t, sc)
	dExplicit := renderBytes(t, sc, scene.WithTheme(pptx.NewTheme()))
	if !bytes.Equal(dDefault, dExplicit) {
		t.Errorf("nil DarkColors not byte-identical to the default theme (%d vs %d bytes)", len(dExplicit), len(dDefault))
	}

	// And the pinned gray canvas (111827) is still what a no-dark-palette slide emits.
	slideXML := zipPart(t, dDefault, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "111827") {
		t.Errorf("default-theme dark slide lost the pinned gray canvas 111827:\n%s", slideXML)
	}
}

// TestDarkPalette_Determinism proves a soul-dark deck renders byte-identically
// across worker counts (the overlay is order-independent — RFC §10.1).
func TestDarkPalette_Determinism(t *testing.T) {
	th := pptx.NewTheme(
		pptx.WithDarkSurface(pptx.ColorCanvas, "0A0E1A"),
		pptx.WithDarkSurface(pptx.ColorSurfaceAlt, "1C2238"),
		pptx.WithDarkText(pptx.TextSecondary, "C8D0E8"),
	)
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{Variant: scene.VariantDark, Nodes: []scene.SlideNode{scene.Heading{Text: rt("A"), Level: 1}}},
		{Variant: scene.VariantLight, Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("B.")}}}},
		{Variant: scene.VariantDark, Nodes: []scene.SlideNode{scene.Heading{Text: rt("C"), Level: 2}}},
	}}
	seq := renderBytes(t, sc, scene.WithTheme(th.Clone()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithTheme(th.Clone()), scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Errorf("soul-dark deck not deterministic across workers (%d vs %d bytes)", len(seq), len(par))
	}
}
