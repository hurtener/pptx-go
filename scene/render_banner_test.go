package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
)

// White-box tests for the Banner composer (R12.6, D-097): the fill default, the
// auto-contrast text color, the left/right region split, and the shape count.

// TestBannerFillRole: a zero Fill (ColorCanvas) becomes ColorAccent; an explicit fill
// is honored.
func TestBannerFillRole(t *testing.T) {
	if got := bannerFillRole(Banner{}); got != pptx.ColorAccent {
		t.Errorf("zero Fill = %v, want ColorAccent", got)
	}
	if got := bannerFillRole(Banner{Fill: ColorSuccess}); got != pptx.ColorSuccess {
		t.Errorf("explicit Fill = %v, want ColorSuccess", got)
	}
}

// TestBannerTextColor: a default TextColor auto-contrasts to inverse on the dark accent
// fill; an explicit TextColor is honored verbatim.
func TestBannerTextColor(t *testing.T) {
	r := newTestRenderer(t)
	if got := r.bannerTextColor(Banner{}, pptx.ColorAccent); got != pptx.TokenTextColor(pptx.TextInverse) {
		t.Errorf("default text on accent fill = %#v, want inverse (auto-contrast)", got)
	}
	if got := r.bannerTextColor(Banner{TextColor: TextAccent}, pptx.ColorAccent); got != pptx.TokenTextColor(pptx.TextAccent) {
		t.Errorf("explicit TextColor = %#v, want accent", got)
	}
}

// TestBannerRegions: trailing children reserve a right region and shrink the left text
// width; with no trailing the left text spans the inner width (minus an icon).
func TestBannerRegions(t *testing.T) {
	r := newTestRenderer(t)
	innerW := pptx.In(10)
	noTrail, rightW0 := bannerRegions(innerW, Banner{}, r.theme)
	if rightW0 != 0 {
		t.Errorf("no trailing should reserve no right region, got %d", rightW0)
	}
	if noTrail != innerW {
		t.Errorf("no-icon no-trailing left width = %d, want innerW %d", noTrail, innerW)
	}
	withTrail, rightW := bannerRegions(innerW, Banner{Trailing: []SlideNode{Button{Label: "Go"}}}, r.theme)
	if rightW <= 0 || withTrail >= noTrail {
		t.Errorf("trailing should reserve a right region and shrink left: left=%d right=%d", withTrail, rightW)
	}
}

// TestRenderBanner_ShapeCount: strip + text frame (+ icon) + a trailing button's shapes.
func TestRenderBanner_ShapeCount(t *testing.T) {
	r := newTestRenderer(t)
	r.cfg.icons = icons.Curated()
	ps := r.pres.AddSlide()
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(12), H: pptx.In(1.2)}
	v := Banner{Lead: RichText{{Text: "Lead"}}, Body: RichText{{Text: "body"}}, Icon: "star",
		Trailing: []SlideNode{Button{Label: "Go"}}}
	r.renderBanner(ps, box, v, "s1")
	// strip(1) + icon(1) + text frame(1) + button [pill(1)+label(1)] = 5.
	if r.stats.Shapes != 5 {
		t.Errorf("emitted %d shapes, want 5 (warnings: %v)", r.stats.Shapes, r.stats.Warnings)
	}
}
