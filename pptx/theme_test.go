package pptx_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestDefaultThemeResolvesEveryRole(t *testing.T) {
	th := pptx.DefaultTheme()

	surfaces := []pptx.ColorRole{
		pptx.ColorCanvas, pptx.ColorSurface, pptx.ColorSurfaceAlt, pptx.ColorAccent,
		pptx.ColorAccentAlt, pptx.ColorAccentWarm, pptx.ColorSuccess, pptx.ColorWarning,
		pptx.ColorError, pptx.ColorInfo, pptx.ColorPaper,
	}
	for _, r := range surfaces {
		if v := th.ResolveColor(r); len(v) != 6 {
			t.Errorf("ColorRole %d resolved to %q (want 6-hex)", r, v)
		}
	}
	texts := []pptx.TextColorRole{
		pptx.TextPrimary, pptx.TextSecondary, pptx.TextTertiary, pptx.TextInverse,
		pptx.TextMuted, pptx.TextAccent, pptx.TextAccentAlt, pptx.TextSuccess,
		pptx.TextWarning, pptx.TextError,
	}
	for _, r := range texts {
		if v := th.ResolveTextColor(r); len(v) != 6 {
			t.Errorf("TextColorRole %d resolved to %q (want 6-hex)", r, v)
		}
	}
	types := []pptx.TypeRole{
		pptx.TypeDisplay, pptx.TypeH1, pptx.TypeH2, pptx.TypeH3, pptx.TypeH4, pptx.TypeH5,
		pptx.TypeBody, pptx.TypeBodySmall, pptx.TypeCaption, pptx.TypeMono, pptx.TypeCode,
	}
	for _, r := range types {
		if spec := th.ResolveType(r); spec.Family == "" || spec.Size == 0 {
			t.Errorf("TypeRole %d resolved to %+v", r, spec)
		}
	}
	for _, r := range []pptx.SpaceRole{pptx.SpaceXS, pptx.SpaceSM, pptx.SpaceMD, pptx.SpaceLG, pptx.SpaceXL, pptx.Space2XL} {
		if th.ResolveSpace(r) <= 0 {
			t.Errorf("SpaceRole %d resolved to 0", r)
		}
	}
	// Radii: RadiusNone is legitimately 0.
	if th.ResolveRadius(pptx.RadiusMD) <= 0 {
		t.Error("RadiusMD resolved to 0")
	}
	if !th.ResolveElevation(pptx.ElevationFlat).IsFlat() {
		t.Error("ElevationFlat should be flat")
	}
	if th.ResolveElevation(pptx.ElevationRaised).IsFlat() {
		t.Error("ElevationRaised should not be flat")
	}
}

func TestResolutionDeterministic(t *testing.T) {
	th := pptx.DefaultTheme()
	for i := 0; i < 3; i++ {
		if th.ResolveColor(pptx.ColorAccent) != th.ResolveColor(pptx.ColorAccent) {
			t.Fatal("ResolveColor not deterministic")
		}
		if !reflect.DeepEqual(th.ResolveType(pptx.TypeH1), th.ResolveType(pptx.TypeH1)) {
			t.Fatal("ResolveType not deterministic")
		}
	}
}

func TestThemeSwapReResolves(t *testing.T) {
	a := pptx.DefaultTheme()
	b := pptx.NewTheme(pptx.WithAccent("FF0000"))
	if a.ResolveColor(pptx.ColorAccent) == b.ResolveColor(pptx.ColorAccent) {
		t.Fatal("theme swap: ColorAccent should differ between themes")
	}
	if b.ResolveColor(pptx.ColorAccent) != "FF0000" {
		t.Errorf("WithAccent: got %q want FF0000", b.ResolveColor(pptx.ColorAccent))
	}
	// The default theme must be unaffected by NewTheme's mutation (Clone).
	if a.ResolveColor(pptx.ColorAccent) == "FF0000" {
		t.Fatal("NewTheme leaked a mutation into the default theme")
	}
}

// TestColorPaperDefaultIsCanvas verifies ColorPaper defaults to the canvas value
// (white) so a deck using it is byte-identical until a theme overrides the tint
// (D-104, R13.1 acceptance 1).
func TestColorPaperDefaultIsCanvas(t *testing.T) {
	th := pptx.DefaultTheme()
	if got := th.ResolveColor(pptx.ColorPaper); got != th.ResolveColor(pptx.ColorCanvas) {
		t.Errorf("ColorPaper default = %q, want = ColorCanvas %q", got, th.ResolveColor(pptx.ColorCanvas))
	}
	if got := th.ResolveColor(pptx.ColorPaper); got != "FFFFFF" {
		t.Errorf("ColorPaper default = %q, want FFFFFF", got)
	}
}

// TestWithPaper verifies WithPaper sets the off-white tint, it resolves, and
// Clone carries it (D-104, R13.1 acceptance 2).
func TestWithPaper(t *testing.T) {
	th := pptx.NewTheme(pptx.WithPaper("FAFAF8"))
	if got := th.ResolveColor(pptx.ColorPaper); got != "FAFAF8" {
		t.Errorf("WithPaper: ColorPaper = %q, want FAFAF8", got)
	}
	// ColorCanvas stays white — the paper tint is a distinct token.
	if got := th.ResolveColor(pptx.ColorCanvas); got != "FFFFFF" {
		t.Errorf("WithPaper leaked into ColorCanvas = %q, want FFFFFF", got)
	}
	if got := th.Clone().ResolveColor(pptx.ColorPaper); got != "FAFAF8" {
		t.Errorf("Clone dropped ColorPaper = %q, want FAFAF8", got)
	}
	// The default theme must be unaffected by NewTheme's mutation (Clone).
	if got := pptx.DefaultTheme().ResolveColor(pptx.ColorPaper); got != "FFFFFF" {
		t.Errorf("WithPaper leaked into the default theme = %q, want FFFFFF", got)
	}
}

func TestResolveUnsetFallback(t *testing.T) {
	empty := &pptx.Theme{Colors: pptx.ColorPalette{Surfaces: map[pptx.ColorRole]pptx.RGB{}, Text: map[pptx.TextColorRole]pptx.RGB{}}}
	if empty.ResolveColor(pptx.ColorAccent) != "FFFFFF" {
		t.Errorf("unset surface fallback: got %q", empty.ResolveColor(pptx.ColorAccent))
	}
	if empty.ResolveTextColor(pptx.TextPrimary) != "000000" {
		t.Errorf("unset text fallback: got %q", empty.ResolveTextColor(pptx.TextPrimary))
	}
	if empty.ResolveType(pptx.TypeBody).Family == "" {
		t.Error("unset type fallback should have a family")
	}
}

func TestCloneIndependence(t *testing.T) {
	a := pptx.DefaultTheme()
	c := a.Clone()
	c.Colors.Surfaces[pptx.ColorAccent] = "00FF00"
	if a.ResolveColor(pptx.ColorAccent) == "00FF00" {
		t.Fatal("Clone shares the surfaces map with the original")
	}
}

// TestWithDarkColors verifies WithDarkSurface/WithDarkText populate DarkColors,
// the default theme carries no dark palette, and Clone deep-copies it (R8.3).
func TestWithDarkColors(t *testing.T) {
	// The default theme has no dark palette → byte-identical pinned-gray fallback.
	if pptx.DefaultTheme().DarkColors != nil {
		t.Fatal("DefaultTheme should have a nil DarkColors (pinned-gray fallback)")
	}

	th := pptx.NewTheme(
		pptx.WithDarkSurface(pptx.ColorCanvas, "0A0E1A"),
		pptx.WithDarkSurface(pptx.ColorSurface, "14182B"),
		pptx.WithDarkText(pptx.TextPrimary, "F4F6FF"),
	)
	if th.DarkColors == nil {
		t.Fatal("WithDarkSurface/WithDarkText did not allocate DarkColors")
	}
	if got := th.DarkColors.Surfaces[pptx.ColorCanvas]; got != "0A0E1A" {
		t.Errorf("dark canvas = %q, want 0A0E1A", got)
	}
	if got := th.DarkColors.Surfaces[pptx.ColorSurface]; got != "14182B" {
		t.Errorf("dark surface = %q, want 14182B", got)
	}
	if got := th.DarkColors.Text[pptx.TextPrimary]; got != "F4F6FF" {
		t.Errorf("dark primary text = %q, want F4F6FF", got)
	}
	// NewTheme must not have mutated the package default theme.
	if pptx.DefaultTheme().DarkColors != nil {
		t.Error("WithDarkSurface leaked a DarkColors into the default theme")
	}
}

// TestCloneDarkColorsIndependence proves Clone deep-copies DarkColors so a clone
// can be mutated without aliasing the original (themes are reusable, §5; R8.3).
func TestCloneDarkColorsIndependence(t *testing.T) {
	a := pptx.NewTheme(pptx.WithDarkSurface(pptx.ColorCanvas, "0A0E1A"))
	c := a.Clone()
	if c.DarkColors == nil || c.DarkColors == a.DarkColors {
		t.Fatal("Clone shares (or dropped) the DarkColors pointer")
	}
	c.DarkColors.Surfaces[pptx.ColorCanvas] = "FFFFFF"
	if a.DarkColors.Surfaces[pptx.ColorCanvas] != "0A0E1A" {
		t.Fatal("Clone shares the DarkColors.Surfaces map with the original")
	}
	// A nil DarkColors clones to nil (byte-identical fallback preserved).
	if pptx.DefaultTheme().Clone().DarkColors != nil {
		t.Error("Clone of a nil-DarkColors theme should stay nil")
	}
}

// TestDarkColorsConcurrentReuse proves a theme carrying a dark palette is safe
// for concurrent Clone+read under -race (a Theme is a reusable artifact, §5).
func TestDarkColorsConcurrentReuse(t *testing.T) {
	base := pptx.NewTheme(
		pptx.WithDarkSurface(pptx.ColorCanvas, "0A0E1A"),
		pptx.WithDarkText(pptx.TextPrimary, "F4F6FF"),
	)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c := base.Clone()
			c.DarkColors.Surfaces[pptx.ColorCanvas] = "112233" // mutate the clone only
			if base.DarkColors.Surfaces[pptx.ColorCanvas] != "0A0E1A" {
				t.Error("concurrent Clone aliased the base dark palette")
			}
		}()
	}
	wg.Wait()
}

func TestWithFonts(t *testing.T) {
	th := pptx.NewTheme(pptx.WithFonts("Inter", "Inter"))
	if th.HeadingFont != "Inter" || th.BodyFont != "Inter" {
		t.Errorf("WithFonts: heading=%q body=%q", th.HeadingFont, th.BodyFont)
	}
	if th.ResolveType(pptx.TypeH1).Family != "Inter" {
		t.Errorf("H1 family: got %q", th.ResolveType(pptx.TypeH1).Family)
	}
	if th.ResolveType(pptx.TypeBody).Family != "Inter" {
		t.Errorf("Body family: got %q", th.ResolveType(pptx.TypeBody).Family)
	}
	// Mono stays mono.
	if th.ResolveType(pptx.TypeMono).Family != "Consolas" {
		t.Errorf("Mono family should stay Consolas, got %q", th.ResolveType(pptx.TypeMono).Family)
	}
}
