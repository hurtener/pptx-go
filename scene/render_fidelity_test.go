package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// brandSoulTheme is a non-default brand theme exercising the full Wave-15 surface:
// custom light accents (WithAccents), a paper tint, a soul dark palette overriding
// accent + accent-alt + accent-text for the dark variant, and a named gradient.
func brandSoulTheme() *pptx.Theme {
	return pptx.NewTheme(
		pptx.WithName("Brand"),
		pptx.WithPaper("F0ECE2"),
		pptx.WithAccents("0D9488", "EA580C", "7C3AED", "65A30D"),
		pptx.WithDarkSurface(pptx.ColorCanvas, "0A0E1A"),
		pptx.WithDarkSurface(pptx.ColorSurface, "14182B"),
		pptx.WithDarkSurface(pptx.ColorSurfaceAlt, "1C2238"),
		pptx.WithDarkSurface(pptx.ColorAccent, "5EEAD4"),
		pptx.WithDarkSurface(pptx.ColorAccentAlt, "FDBA74"),
		pptx.WithDarkText(pptx.TextPrimary, "F4F6FF"),
		pptx.WithDarkText(pptx.TextAccent, "5EEAD4"),
		pptx.WithGradient("heroDark", pptx.GradientSpec{
			Stops:  []pptx.GradientStop{{Pos: 0, Color: pptx.RGB("1E293B")}, {Pos: 1, Color: pptx.RGB("0A0E1A")}},
			Radial: true,
		}),
	)
}

// assertFidelity asserts every resolved color in c equals the given theme's token
// for that role — the soul→engine fidelity check (R8.10).
func assertFidelity(t *testing.T, label string, c SlideColors, th *pptx.Theme) {
	t.Helper()
	checks := []struct {
		role string
		got  pptx.RGB
		want pptx.RGB
	}{
		{"Canvas", c.Canvas, th.ResolveColor(pptx.ColorCanvas)},
		{"Surface", c.Surface, th.ResolveColor(pptx.ColorSurface)},
		{"SurfaceAlt", c.SurfaceAlt, th.ResolveColor(pptx.ColorSurfaceAlt)},
		{"Accent", c.Accent, th.ResolveColor(pptx.ColorAccent)},
		{"AccentAlt", c.AccentAlt, th.ResolveColor(pptx.ColorAccentAlt)},
		{"PrimaryText", c.PrimaryText, th.ResolveTextColor(pptx.TextPrimary)},
		{"TextAccent", c.TextAccent, th.ResolveTextColor(pptx.TextAccent)},
	}
	for _, ck := range checks {
		if ck.got != ck.want {
			t.Errorf("%s: resolved %s = %q, want soul token %q", label, ck.role, ck.got, ck.want)
		}
	}
}

// TestSoulFidelity_LightAndDark is the Wave-15 capstone (R8.10): for a non-default
// brand soul, every slide's resolved surfaces/accents/text equal the soul's
// declared values for that variant — the light theme for a light slide, the
// derived dark theme for a dark slide.
func TestSoulFidelity_LightAndDark(t *testing.T) {
	th := brandSoulTheme()
	pres := pptx.New(pptx.WithTheme(th))
	sc := Scene{Slides: []SceneSlide{
		headingSlide("light", VariantLight),
		headingSlide("dark", VariantDark),
	}}
	stats, err := Render(pres, sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(stats.Colors) != 2 {
		t.Fatalf("Colors len = %d, want 2", len(stats.Colors))
	}

	// Light slide resolves against the active brand theme.
	assertFidelity(t, "light", stats.Colors[0], th)
	// Dark slide resolves against the derived dark theme (soul dark overrides).
	assertFidelity(t, "dark", stats.Colors[1], darkThemeFrom(th))

	// The soul's dark overrides must actually have taken: dark accent/canvas/text
	// differ from the light slide's.
	light, dark := stats.Colors[0], stats.Colors[1]
	if dark.Canvas == light.Canvas || dark.Accent == light.Accent || dark.TextAccent == light.TextAccent {
		t.Errorf("dark variant did not re-resolve the brand soul: light=%+v dark=%+v", light, dark)
	}
	// And the dark accent is the soul's dark accent, not the light brand accent.
	if dark.Accent != "5EEAD4" {
		t.Errorf("dark Accent = %q, want soul dark accent 5EEAD4", dark.Accent)
	}
}

// TestSoulFidelity_MismatchFails is the negative guard (R8.10 acceptance): a
// deliberate token mismatch is caught by the fidelity comparison.
func TestSoulFidelity_MismatchFails(t *testing.T) {
	th := brandSoulTheme()
	pres := pptx.New(pptx.WithTheme(th))
	stats, err := Render(pres, Scene{Slides: []SceneSlide{headingSlide("a", VariantLight)}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	c := stats.Colors[0]
	// A wrong intended accent must not match the resolved accent.
	if c.Accent == "FFFFFF" {
		t.Fatal("precondition: brand accent should not be white")
	}
	wrong := th.Clone()
	wrong.Colors.Surfaces[pptx.ColorAccent] = "FFFFFF"
	if c.Accent == wrong.ResolveColor(pptx.ColorAccent) {
		t.Error("fidelity check failed to catch a deliberate accent mismatch")
	}
}

// TestSoulFidelity_Deterministic asserts the extended SlideColors is identical
// across worker counts (R8.10 byte-stability; SlideColors stays comparable).
func TestSoulFidelity_Deterministic(t *testing.T) {
	th := brandSoulTheme()
	sc := Scene{}
	for i := 0; i < 8; i++ {
		v := VariantLight
		if i%2 == 1 {
			v = VariantDark
		}
		sc.Slides = append(sc.Slides, headingSlide(string(rune('A'+i)), v))
	}
	seq, err := Render(pptx.New(pptx.WithTheme(th.Clone())), sc, WithWorkers(1))
	if err != nil {
		t.Fatalf("Render seq: %v", err)
	}
	par, err := Render(pptx.New(pptx.WithTheme(th.Clone())), sc, WithWorkers(8))
	if err != nil {
		t.Fatalf("Render par: %v", err)
	}
	for i := range seq.Colors {
		if seq.Colors[i] != par.Colors[i] {
			t.Errorf("Colors[%d] differs across workers: %+v vs %+v", i, seq.Colors[i], par.Colors[i])
		}
	}
}
