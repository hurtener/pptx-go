package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Resolved per-slide colors (Phase 29, R7). Stats.Colors reports the canvas /
// surface / primary-text the engine rendered each slide with — the derived dark
// palette for VariantDark — so a caller can compute its own contrast.

func headingSlide(id string, variant Variant) SceneSlide {
	return SceneSlide{ID: id, Variant: variant, Nodes: []SlideNode{Heading{Text: RichText{{Text: "h"}}, Level: 1}}}
}

// TestStatsColors_PerSlideSceneOrder is acceptance criterion 1: one entry per
// slide, in scene order, keyed by SlideID.
func TestStatsColors_PerSlideSceneOrder(t *testing.T) {
	sc := Scene{Slides: []SceneSlide{headingSlide("a", VariantLight), headingSlide("b", VariantLight), headingSlide("c", VariantLight)}}
	stats, err := Render(pptx.New(), sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(stats.Colors) != 3 {
		t.Fatalf("Colors len = %d, want 3", len(stats.Colors))
	}
	for i, want := range []string{"a", "b", "c"} {
		if stats.Colors[i].SlideID != want {
			t.Errorf("Colors[%d].SlideID = %q, want %q (scene order)", i, stats.Colors[i].SlideID, want)
		}
	}
}

// TestStatsColors_LightMatchesTheme is acceptance criterion 2 (light): a light
// slide's resolved colors equal the active theme's.
func TestStatsColors_LightMatchesTheme(t *testing.T) {
	pres := pptx.New()
	th := pres.Theme()
	stats, err := Render(pres, Scene{Slides: []SceneSlide{headingSlide("a", VariantLight)}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	c := stats.Colors[0]
	if c.Canvas != th.ResolveColor(pptx.ColorCanvas) {
		t.Errorf("Canvas = %q, want %q", c.Canvas, th.ResolveColor(pptx.ColorCanvas))
	}
	if c.Surface != th.ResolveColor(pptx.ColorSurface) {
		t.Errorf("Surface = %q, want %q", c.Surface, th.ResolveColor(pptx.ColorSurface))
	}
	if c.PrimaryText != th.ResolveTextColor(pptx.TextPrimary) {
		t.Errorf("PrimaryText = %q, want %q", c.PrimaryText, th.ResolveTextColor(pptx.TextPrimary))
	}
}

// TestStatsColors_DarkPalette is acceptance criterion 2 (dark): a VariantDark
// slide's resolved colors equal the derived dark palette and differ from a light
// slide's (canvas darker, primary text lighter).
func TestStatsColors_DarkPalette(t *testing.T) {
	pres := pptx.New()
	dark := darkThemeFrom(pres.Theme())
	sc := Scene{Slides: []SceneSlide{headingSlide("light", VariantLight), headingSlide("dark", VariantDark)}}
	stats, err := Render(pres, sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	light, darkC := stats.Colors[0], stats.Colors[1]

	if darkC.Canvas != dark.ResolveColor(pptx.ColorCanvas) {
		t.Errorf("dark Canvas = %q, want derived dark %q", darkC.Canvas, dark.ResolveColor(pptx.ColorCanvas))
	}
	if darkC.PrimaryText != dark.ResolveTextColor(pptx.TextPrimary) {
		t.Errorf("dark PrimaryText = %q, want derived dark %q", darkC.PrimaryText, dark.ResolveTextColor(pptx.TextPrimary))
	}
	if darkC.Canvas == light.Canvas {
		t.Errorf("dark canvas %q should differ from light canvas %q", darkC.Canvas, light.Canvas)
	}
	if darkC.PrimaryText == light.PrimaryText {
		t.Errorf("dark primary text %q should differ from light %q", darkC.PrimaryText, light.PrimaryText)
	}
}

// TestStatsColors_Deterministic is acceptance criterion 3: Colors is identical
// across worker counts (built in scene order, independent of scheduling).
func TestStatsColors_Deterministic(t *testing.T) {
	sc := Scene{}
	for i := 0; i < 12; i++ {
		variant := VariantLight
		if i%2 == 1 {
			variant = VariantDark
		}
		sc.Slides = append(sc.Slides, headingSlide(string(rune('A'+i)), variant))
	}
	seq, err := Render(pptx.New(), sc, WithWorkers(1))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	par, err := Render(pptx.New(), sc, WithWorkers(8))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(seq.Colors) != len(par.Colors) {
		t.Fatalf("Colors len differs: %d vs %d", len(seq.Colors), len(par.Colors))
	}
	for i := range seq.Colors {
		if seq.Colors[i] != par.Colors[i] {
			t.Errorf("Colors[%d] differs across workers: %+v vs %+v", i, seq.Colors[i], par.Colors[i])
		}
	}
}
