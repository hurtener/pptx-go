package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// R11.4 verify-and-close (D-084): the chrome-aware body region reserves the
// eyebrow and footer bands, so body content never intersects them — already
// implemented by D-053 (the bodyRegion shrink) and made overflow-proof by D-083
// (the safe-area clamp). These guards pin the disjointness.

// eyebrowBandBottom / footerBandTop recompute the band edges from the same chrome
// constants the composer (chrome.go) draws with, so the test and the renderer share
// one source of truth.
func eyebrowBandBottom() pptx.EMU { return bodyMargin + chromeEyebrowH + chromeRuleH }
func footerBandTop(cy pptx.EMU) pptx.EMU {
	return cy - bodyMargin - chromeFooterH
}

// TestBodyRegionReservesChrome is the R11.4 acceptance: on a chromed slide the body
// region is disjoint from both reserved bands — its top is at or below the eyebrow
// band (incl. the rule) and its bottom is at or above the footer band top.
func TestBodyRegionReservesChrome(t *testing.T) {
	r := newTestRenderer(t)
	r.chrome = Chrome{Enabled: true}
	_, cy := r.pres.SlideSize()

	body := r.bodyRegion()
	if body.Y < eyebrowBandBottom() {
		t.Errorf("body top %d intersects the eyebrow band (bottom %d)", body.Y, eyebrowBandBottom())
	}
	if body.Bottom() > footerBandTop(pptx.EMU(cy)) {
		t.Errorf("body bottom %d intersects the footer band (top %d)", body.Bottom(), footerBandTop(pptx.EMU(cy)))
	}
}

// TestBodyRegionChromeOff_ByteIdentical: with chrome off, the body region is the
// plain margin box (no band reservation) — the byte-identical, pre-chrome layout.
func TestBodyRegionChromeOff_ByteIdentical(t *testing.T) {
	r := newTestRenderer(t)
	cx, cy := r.pres.SlideSize()
	want := pptx.Box{X: bodyMargin, Y: bodyMargin, W: pptx.EMU(cx) - 2*bodyMargin, H: pptx.EMU(cy) - 2*bodyMargin}
	if got := r.bodyRegion(); got != want {
		t.Errorf("chrome-off body region = %+v, want the plain margin box %+v", got, want)
	}
}

// TestClampedContainerStaysAboveFooter combines R11.4 + R11.3: a container handed a
// box that overflows a chromed slide is clamped to the reserved region, so its
// bottom stays at or above the footer band top.
func TestClampedContainerStaysAboveFooter(t *testing.T) {
	r := newTestRenderer(t)
	r.chrome = Chrome{Enabled: true}
	_, cy := r.pres.SlideSize()
	sa := r.safeArea()

	overflow := pptx.Box{X: sa.X, Y: sa.Y, W: sa.W, H: sa.H * 2}
	clamped := r.clampToSafeArea(overflow, "s1")
	if clamped.Bottom() > footerBandTop(pptx.EMU(cy)) {
		t.Errorf("clamped container bottom %d intersects the footer band (top %d)", clamped.Bottom(), footerBandTop(pptx.EMU(cy)))
	}
}
