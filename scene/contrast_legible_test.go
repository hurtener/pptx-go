package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// ratioOn is the WCAG contrast ratio (×10) of fg against bg, via the package's
// own luminance math — so the test asserts against the same scale LegibleTextOn
// targets.
func ratioOn(fg, bg pptx.RGB) int {
	return contrastRatioT10(relLuminance(fg), relLuminance(bg))
}

// TestLegibleTextOn_ClearsTargetPerBackground verifies the nudge lightens an
// accent that fails on a dark surface and darkens one that fails on a light
// surface, each clearing the target ratio, and that deriving the same accent
// against a dark vs a light background yields different results (R8.6 acceptance).
func TestLegibleTextOn_ClearsTargetPerBackground(t *testing.T) {
	const target = 45 // 4.5:1

	// Lighten path: a dark slate accent on a dark surface fails → lighten.
	const (
		darkAccent = "334155"
		darkBg     = "1F2937"
	)
	if ratioOn(darkAccent, darkBg) >= target {
		t.Fatalf("precondition: dark accent already legible on the dark surface (%d)", ratioOn(darkAccent, darkBg))
	}
	lightened := LegibleTextOn(darkAccent, darkBg, target)
	if got := ratioOn(lightened, darkBg); got < target {
		t.Errorf("lighten: %q on %q ratio %d < target %d", lightened, darkBg, got, target)
	}
	if relLuminance(lightened) <= relLuminance(darkAccent) {
		t.Errorf("lighten path should brighten the accent: %q (lum %d) vs %q (lum %d)", lightened, relLuminance(lightened), darkAccent, relLuminance(darkAccent))
	}

	// Darken path: a teal accent on a cream surface fails → darken.
	const (
		teal  = "0D9488"
		cream = "F0ECE2"
	)
	if ratioOn(teal, cream) >= target {
		t.Fatalf("precondition: teal already legible on cream (%d)", ratioOn(teal, cream))
	}
	darkened := LegibleTextOn(teal, cream, target)
	if got := ratioOn(darkened, cream); got < target {
		t.Errorf("darken: %q on %q ratio %d < target %d", darkened, cream, got, target)
	}
	if relLuminance(darkened) >= relLuminance(teal) {
		t.Errorf("darken path should dim the accent: %q (lum %d) vs %q (lum %d)", darkened, relLuminance(darkened), teal, relLuminance(teal))
	}

	// Per-variant: the same accent derived against a dark vs a light surface
	// differs (legible jade-on-navy vs deeper jade-on-cream).
	if onDark, onLight := LegibleTextOn(teal, darkBg, target), LegibleTextOn(teal, cream, target); onDark == onLight {
		t.Errorf("per-variant derivation should differ: dark=%q light=%q", onDark, onLight)
	}
}

// TestLegibleTextOn_AlreadyLegibleUnchanged verifies an accent that already clears
// the ratio is returned verbatim (byte-identical for the common case).
func TestLegibleTextOn_AlreadyLegibleUnchanged(t *testing.T) {
	// The default accent 2563EB clears 4.5:1 on white (≈5.17:1).
	if got := LegibleTextOn("2563EB", "FFFFFF", 45); got != "2563EB" {
		t.Errorf("already-legible accent changed: %q, want 2563EB", got)
	}
}

// TestLegibleTextOn_Pure verifies the helper is a pure function (same inputs →
// same output), the determinism the parallel render path requires.
func TestLegibleTextOn_Pure(t *testing.T) {
	a := LegibleTextOn("0D9488", "0A0E1A", 45)
	b := LegibleTextOn("0D9488", "0A0E1A", 45)
	if a != b {
		t.Errorf("LegibleTextOn is not pure: %q vs %q", a, b)
	}
}

// TestLegibleTextOn_MalformedFailsSafe verifies a malformed fg or bg returns fg
// unchanged rather than panicking or forcing an override.
func TestLegibleTextOn_MalformedFailsSafe(t *testing.T) {
	if got := LegibleTextOn("ZZZ", "FFFFFF", 45); got != "ZZZ" {
		t.Errorf("malformed fg: got %q, want ZZZ unchanged", got)
	}
	if got := LegibleTextOn("2563EB", "GG", 45); got != "2563EB" {
		t.Errorf("malformed bg: got %q, want 2563EB unchanged", got)
	}
}

// TestLegibleTextOn_DarkenPreservesHue verifies the darken path (blend toward
// black) preserves hue exactly — the result's channel ratios match the source.
func TestLegibleTextOn_DarkenPreservesHue(t *testing.T) {
	const accent = "0D9488"
	out := LegibleTextOn(accent, "F0ECE2", 45) // cream bg → darken toward black
	sr, sg, sb, _ := parseHexRGB(accent)
	or, og, ob, _ := parseHexRGB(out)
	// Scaling toward black keeps cross-channel products equal (r1*g2 == r2*g1),
	// within rounding. Check the dominant pair (g and b are the large channels).
	if abs(og*sb-ob*sg) > 255 {
		t.Errorf("darken did not preserve hue: src=%q (%d,%d,%d) out=%q (%d,%d,%d)", accent, sr, sg, sb, out, or, og, ob)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// TestLegibleTextOn_DisplayTargetWeaker verifies the target ratio is honored: a
// looser large-text target (3:1) yields a less-extreme nudge than the body target
// (4.5:1) for the same accent/background — the caller controls the threshold.
func TestLegibleTextOn_DisplayTargetWeaker(t *testing.T) {
	const (
		accent = "334155"
		bg     = "1F2937"
	)
	body := LegibleTextOn(accent, bg, 45)
	large := LegibleTextOn(accent, bg, 30)
	if ratioOn(large, bg) < 30 {
		t.Errorf("large-text target not met: %q ratio %d < 30", large, ratioOn(large, bg))
	}
	// The 3:1 result needs no more lightening than the 4.5:1 result.
	if relLuminance(large) > relLuminance(body) {
		t.Errorf("looser target should not over-nudge: large %q (lum %d) brighter than body %q (lum %d)", large, relLuminance(large), body, relLuminance(body))
	}
}
