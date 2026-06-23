package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
)

// White-box tests for the Lockup composer (R12.9, D-102): the logo-height default, the
// asset-vs-icon classification, and the icon-path shape count.

// TestLockupLogoH: MaxHeight when set, else the pinned default.
func TestLockupLogoH(t *testing.T) {
	if got := lockupLogoH(Lockup{}); got != lockupDefaultH {
		t.Errorf("default logo height = %d, want %d", got, lockupDefaultH)
	}
	if got := lockupLogoH(Lockup{MaxHeight: pptx.In(0.6)}); got != pptx.In(0.6) {
		t.Errorf("MaxHeight logo height = %d, want In(0.6)", got)
	}
}

// TestLockup_UsesAssets: only the asset variant registers media (renders serially).
func TestLockup_UsesAssets(t *testing.T) {
	if nodeUsesAssets(Lockup{Icon: "star"}) {
		t.Error("an icon lockup should be media-free")
	}
	if !nodeUsesAssets(Lockup{AssetID: "a"}) {
		t.Error("an asset lockup should report media use")
	}
}

// TestRenderLockup_IconShapeCount: an icon lockup emits a caption text + an icon glyph (2).
func TestRenderLockup_IconShapeCount(t *testing.T) {
	r := newTestRenderer(t)
	r.cfg.icons = icons.Curated()
	ps := r.pres.AddSlide()
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(8), H: pptx.In(0.6)}
	r.renderLockup(ps, box, Lockup{Caption: "POWERED BY", Icon: "star"}, "s1", HAlignCenter)
	if r.stats.Shapes != 2 {
		t.Errorf("icon lockup emitted %d shapes, want 2 (caption + glyph)", r.stats.Shapes)
	}
}
