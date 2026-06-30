package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// borderedCard is a Card with a neutral hairline border (ColorSurfaceAlt) — the
// border R8.7 says must dark-resolve, not inherit the light value.
func borderedCard(style scene.BorderStyle) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID:      "card",
		Variant: scene.VariantDark,
		Nodes: []scene.SlideNode{scene.Card{
			Header:      "Dark card",
			BorderStyle: style,
			Body:        []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}},
		}},
	}}}
}

// TestDarkExtensions_DefaultBorderDarkResolves is R8.7 acceptance (a): a dark
// card's neutral hairline resolves to the DARK ColorSurfaceAlt (#374151), never
// the light SurfaceAlt (#F1F3F5) — the engine's borders dark-resolve, so there is
// no light-on-dark "cream border" artifact.
func TestDarkExtensions_DefaultBorderDarkResolves(t *testing.T) {
	data, _ := render(t, borderedCard(scene.BorderSolid))
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, "374151") {
		t.Errorf("dark card border did not resolve to the dark SurfaceAlt 374151:\n%s", slideXML)
	}
	if strings.Contains(slideXML, "F1F3F5") {
		t.Errorf("dark card border leaked the LIGHT SurfaceAlt F1F3F5 (R8.7 regression):\n%s", slideXML)
	}
}

// TestDarkExtensions_AccentBorderOverridable is R8.7 acceptance (b): a soul's
// DarkColors.Surfaces[ColorAccent] re-tints a dark accent border to a
// dark-variant value distinct from the light accent — the per-variant accent
// override mechanism (the Phase-97 DarkColors overlay) covers accent surfaces.
func TestDarkExtensions_AccentBorderOverridable(t *testing.T) {
	const darkAccent = "5EEAD4" // brand jade for the dark variant
	th := pptx.NewTheme(pptx.WithDarkSurface(pptx.ColorAccent, darkAccent))

	// With the override: the dark accent border uses the dark accent hue.
	data, _ := render(t, borderedCard(scene.BorderAccent), scene.WithTheme(th))
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, darkAccent) {
		t.Errorf("dark accent border did not use the soul dark accent %s:\n%s", darkAccent, slideXML)
	}
	// The default light accent (#2563EB) must NOT appear — it was overridden.
	if strings.Contains(slideXML, "2563EB") {
		t.Errorf("dark accent border leaked the light accent 2563EB despite a dark override:\n%s", slideXML)
	}

	// Without the override: brand identity survives the swap (the light accent is
	// preserved on dark — correct default; the soul opts into a dark accent).
	plain, _ := render(t, borderedCard(scene.BorderAccent))
	plainXML := zipPart(t, plain, "ppt/slides/slide1.xml")
	if !strings.Contains(plainXML, "2563EB") {
		t.Errorf("default dark accent border should preserve the brand light accent 2563EB:\n%s", plainXML)
	}
}

// TestDarkExtensions_AccentTextOverridable is R8.7 acceptance (c): a soul's
// DarkColors.Text[TextAccent] re-tints accent text for the dark variant.
func TestDarkExtensions_AccentTextOverridable(t *testing.T) {
	const darkAccentText = "7DD3FC"
	th := pptx.NewTheme(pptx.WithDarkText(pptx.TextAccent, darkAccentText))
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		Variant: scene.VariantDark,
		Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{
			{{Text: "accent word", Color: scene.TokenTextColor(scene.TextAccent)}},
		}}},
	}}}
	data, _ := render(t, sc, scene.WithTheme(th))
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, darkAccentText) {
		t.Errorf("dark accent text did not use the soul dark TextAccent %s:\n%s", darkAccentText, slideXML)
	}
}

// TestDarkExtensions_NilByteIdentical is R8.7 acceptance (d): a dark slide with no
// DarkColors is byte-for-byte identical to the same slide on the default theme —
// the per-variant override is purely additive (R8.7 is byte-identical for
// light-only decks / souls without dark extensions).
func TestDarkExtensions_NilByteIdentical(t *testing.T) {
	sc := borderedCard(scene.BorderAccent)
	dDefault := renderBytes(t, sc)
	dExplicit := renderBytes(t, sc, scene.WithTheme(pptx.NewTheme()))
	if !bytes.Equal(dDefault, dExplicit) {
		t.Errorf("no-DarkColors dark card not byte-identical to the default theme (%d vs %d bytes)", len(dExplicit), len(dDefault))
	}
}
