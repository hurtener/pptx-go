package scene

import (
	"math"

	"github.com/hurtener/pptx-go/pptx"
)

// Auto-contrast mechanism (R11.2, D-082). A chrome text run left uncolored renders
// as the slide's near-black placeholder default, which is illegible on a dark card
// fill or a dark-variant surface. onCardSurface returns the explicit light text
// token a run needs to stay legible on a dark surface, or nil to leave the run
// unset — the pre-R11.2 default, which is already legible (and byte-identical) on a
// light surface.
//
// This is a MECHANISM, not a policy (D-026): the luminance rule is fixed and
// pinned, and a caller can always override it by supplying an explicit Color. It
// reconciles with D-058 (the engine ships no contrast *logic*, only resolved
// colors) — onCardSurface is a deterministic token picker, the color analog of
// deltaToneColor, not an opinion about what a deck should look like.
//
// All math is pure integer per call (the sRGB gamma curve is precomputed once into
// an integer table at init), so the color choice is identical regardless of
// worker count — the determinism the parallel render path requires.

// srgbLinear[i] is the sRGB-linearized value of the 8-bit channel i, scaled to
// [0, 100000]. Built once at init from the WCAG gamma-expansion curve; the build
// uses math.Pow (deterministic), every per-call use is an integer lookup.
var srgbLinear [256]int

func init() {
	for i := 0; i < 256; i++ {
		c := float64(i) / 255
		var lin float64
		if c <= 0.04045 {
			lin = c / 12.92
		} else {
			lin = math.Pow((c+0.055)/1.055, 2.4)
		}
		srgbLinear[i] = int(lin*100000 + 0.5)
	}
}

// darkSurfaceLumaMax is the relative-luminance threshold (×100000) below which a
// surface is "dark" — i.e. light text out-contrasts dark text on it. It is the
// black/white crossover L* = √(1.05·0.05) − 0.05 ≈ 0.17912, the luminance at which
// contrast(black, L) == contrast(white, L). A surface at or above it (e.g. a
// saturated teal band, L ≈ 0.23) keeps the dark default; below it (a dark navy
// card, L ≈ 0.02) flips to light text.
const darkSurfaceLumaMax = 17912

// accentMinContrastT10 is the minimum contrast ratio (×10) an accent-tinted run
// must clear against its surface to keep the accent, i.e. 4.5:1 (WCAG AA for small
// text). The default accent 2563EB on a white card clears it (5.17:1) so the
// common eyebrow is byte-identical; a same-hue header band fails it and falls back.
const accentMinContrastT10 = 45

// contrastOffset is the WCAG 0.05 black-floor offset, scaled to the [0,100000]
// luminance range used here.
const contrastOffset = 5000

// parseHexRGB splits a 6-hex RGB string ("2563EB") into its 0..255 channels. A
// malformed value reports ok=false; callers treat that as "leave unchanged" so a
// bad color never forces an override (fail safe / byte-identical).
func parseHexRGB(c pptx.RGB) (r, g, b int, ok bool) {
	s := string(c)
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	var v [6]int
	for i := 0; i < 6; i++ {
		ch := s[i]
		switch {
		case ch >= '0' && ch <= '9':
			v[i] = int(ch - '0')
		case ch >= 'a' && ch <= 'f':
			v[i] = int(ch-'a') + 10
		case ch >= 'A' && ch <= 'F':
			v[i] = int(ch-'A') + 10
		default:
			return 0, 0, 0, false
		}
	}
	return v[0]*16 + v[1], v[2]*16 + v[3], v[4]*16 + v[5], true
}

// relLuminance returns the WCAG sRGB relative luminance of c, scaled to
// [0, 100000]. A malformed color returns 100000 (treated as fully light) so the
// fail-safe path never forces a dark-surface override.
func relLuminance(c pptx.RGB) int {
	r, g, b, ok := parseHexRGB(c)
	if !ok {
		return 100000
	}
	// WCAG coefficients (0.2126, 0.7152, 0.0722) ×10000; the /10000 keeps the
	// ×100000 scale of srgbLinear.
	return (2126*srgbLinear[r] + 7152*srgbLinear[g] + 722*srgbLinear[b]) / 10000
}

// contrastRatioT10 returns the WCAG contrast ratio between two surfaces, ×10
// (so 45 == 4.5:1), from their relative luminances (each in [0,100000]).
func contrastRatioT10(lumA, lumB int) int {
	hi, lo := lumA, lumB
	if lo > hi {
		hi, lo = lo, hi
	}
	return (hi + contrastOffset) * 10 / (lo + contrastOffset)
}

// onCardSurface returns the explicit text Color a chrome run should use to stay
// legible on a surface of role bg, or nil to leave the run's Color unset. It
// returns the light TextInverse token only when bg resolves dark enough that the
// inherited dark default would be illegible; on any light/medium surface it returns
// nil, reproducing the pre-R11.2 output byte-for-byte. Resolves bg against the
// active (possibly dark-variant) theme, so it is correct on every variant.
func (r *renderer) onCardSurface(bg pptx.ColorRole) pptx.Color {
	if relLuminance(r.theme.ResolveColor(bg)) < darkSurfaceLumaMax {
		return pptx.TokenTextColor(pptx.TextInverse)
	}
	return nil
}

// accentLegible reports whether the theme's TextAccent clears the minimum contrast
// ratio against a surface of role bg — i.e. whether an accent-tinted run stays
// legible there. Used by the eyebrow: keep the accent when it passes (byte-
// identical on the default light card), else fall back to onCardSurface.
func (r *renderer) accentLegible(bg pptx.ColorRole) bool {
	la := relLuminance(r.theme.ResolveTextColor(pptx.TextAccent))
	ls := relLuminance(r.theme.ResolveColor(bg))
	return contrastRatioT10(la, ls) >= accentMinContrastT10
}
