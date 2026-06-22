package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box tests for the R11.2 auto-contrast mechanism (D-082).

// TestRelLuminance_Ordering guards the sRGB relative-luminance estimator: white is
// max, black is min, and a mid gray sits between — the ordering every downstream
// decision relies on.
func TestRelLuminance_Ordering(t *testing.T) {
	white := relLuminance("FFFFFF")
	black := relLuminance("000000")
	gray := relLuminance("808080")
	if white != 100000 {
		t.Errorf("white luminance = %d, want 100000", white)
	}
	if black != 0 {
		t.Errorf("black luminance = %d, want 0", black)
	}
	if black >= gray || gray >= white {
		t.Errorf("luminance ordering broken: black=%d gray=%d white=%d", black, gray, white)
	}
	// Green is far brighter than blue at equal channel value (the gamma/coefficient
	// property a linear-luma proxy would get wrong).
	if relLuminance("00FF00") <= relLuminance("0000FF") {
		t.Error("green should be more luminous than blue")
	}
}

// TestParseHexRGB guards the parser, including the fail-safe for malformed input
// (treated as fully light so it never forces a dark-surface override).
func TestParseHexRGB(t *testing.T) {
	r, g, b, ok := parseHexRGB("2563EB")
	if !ok || r != 0x25 || g != 0x63 || b != 0xEB {
		t.Errorf("parseHexRGB(2563EB) = %d,%d,%d,%v", r, g, b, ok)
	}
	if _, _, _, ok := parseHexRGB("XYZ"); ok {
		t.Error("malformed hex should report ok=false")
	}
	if l := relLuminance("nothex"); l != 100000 {
		t.Errorf("malformed color luminance = %d, want fail-safe 100000", l)
	}
}

// TestOnCardSurface_ContrastGuarantee is the R11.2 acceptance by construction:
// across a sweep of surfaces (light, dark, brand-saturated), the effective text
// color onCardSurface yields — the light TextInverse token on a dark surface, or
// the inherited near-black default (modeled as 000000) on a light one — always
// clears the 4.5:1 minimum contrast against that surface. The darkSurfaceLumaMax
// threshold is exactly the black/white crossover, so both branches clear ~4.58:1.
func TestOnCardSurface_ContrastGuarantee(t *testing.T) {
	r := newTestRenderer(t)
	// Inject deterministic surface colors via the theme so ResolveColor returns
	// them. We test the helper directly against a range of literal surfaces by
	// overriding ColorInfo as a scratch surface role.
	surfaces := []pptx.RGB{
		"FFFFFF", // white card
		"F1F3F5", // surfaceAlt light
		"111827", // dark navy (variant canvas)
		"1F2937", // dark surface
		"0D9488", // saturated teal band
		"2563EB", // brand blue
		"7C3AED", // brand violet
		"EAB308", // brand yellow (light, saturated)
		"DC2626", // brand red
	}
	for _, s := range surfaces {
		r.theme.Colors.Surfaces[pptx.ColorInfo] = s
		c := r.onCardSurface(pptx.ColorInfo)

		// Model the effective rendered color: nil → the near-black default.
		var effective pptx.RGB = "000000"
		if c != nil {
			effective = r.theme.ResolveTextColor(pptx.TextInverse)
		}
		ratio := contrastRatioT10(relLuminance(effective), relLuminance(s))
		if ratio < accentMinContrastT10 {
			t.Errorf("surface %s: effective text %s contrast %d.%d:1 < 4.5:1 (c=%v)",
				s, effective, ratio/10, ratio%10, c != nil)
		}
	}
}

// TestOnCardSurface_LightIsNil guards the byte-identical lever: a light surface
// returns nil (leave the run unset → inherit the dark default), a dark surface
// returns a non-nil light token. nil on light is what keeps light cards
// byte-identical to the pre-R11.2 output.
func TestOnCardSurface_LightIsNil(t *testing.T) {
	r := newTestRenderer(t)
	r.theme.Colors.Surfaces[pptx.ColorInfo] = "FFFFFF"
	if c := r.onCardSurface(pptx.ColorInfo); c != nil {
		t.Errorf("light surface should yield nil (byte-identical), got %v", c)
	}
	r.theme.Colors.Surfaces[pptx.ColorInfo] = "111827"
	if c := r.onCardSurface(pptx.ColorInfo); c == nil {
		t.Error("dark surface should yield an explicit light token, got nil")
	}
}

// TestAccentLegible guards the eyebrow decision: the default accent on a white
// card clears the threshold (so the common eyebrow stays accent-tinted and
// byte-identical), but the same accent on a same-hue band fails (so it falls back
// to the contrast token instead of going invisible).
func TestAccentLegible(t *testing.T) {
	r := newTestRenderer(t)
	// Default theme: TextAccent 2563EB, accent surface candidates.
	r.theme.Colors.Surfaces[pptx.ColorInfo] = "FFFFFF"
	if !r.accentLegible(pptx.ColorInfo) {
		t.Error("accent on white should be legible (>=4.5:1) — common eyebrow must stay byte-identical")
	}
	// A surface equal to the accent hue: blue accent on blue band → near-zero
	// contrast → not legible.
	r.theme.Colors.Surfaces[pptx.ColorInfo] = "2563EB"
	if r.accentLegible(pptx.ColorInfo) {
		t.Error("accent on a same-hue band should be illegible (falls back to onCardSurface)")
	}
}

// TestDarkSurfaceThreshold pins the crossover constant so a refactor cannot drift
// it: teal (L≈0.23) is treated as light (keeps dark text) and dark navy (L≈0.02)
// as dark (flips to light text).
func TestDarkSurfaceThreshold(t *testing.T) {
	if relLuminance("0D9488") < darkSurfaceLumaMax {
		t.Error("teal band should be above the dark threshold (black text wins)")
	}
	if relLuminance("111827") >= darkSurfaceLumaMax {
		t.Error("dark navy should be below the dark threshold (white text wins)")
	}
}
