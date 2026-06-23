package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
)

// White-box tests for the Card.Ribbon field extension (R12.3, D-098): the top-bar band
// reserve (body shifts down), the color resolution, and the per-position shape count.

// TestRibbonReserve: a top bar reserves its band height; corner positions and nil
// reserve nothing.
func TestRibbonReserve(t *testing.T) {
	if got := ribbonReserveOf(cardChrome{ribbon: &Ribbon{Position: RibbonTopBar}}); got != ribbonTopBarH {
		t.Errorf("top-bar reserve = %d, want %d", got, ribbonTopBarH)
	}
	for _, pos := range []RibbonPos{RibbonCornerTL, RibbonCornerTR, RibbonCornerStar} {
		if got := ribbonReserveOf(cardChrome{ribbon: &Ribbon{Position: pos}}); got != 0 {
			t.Errorf("corner ribbon %d reserve = %d, want 0", pos, got)
		}
	}
	if got := ribbonReserveOf(cardChrome{}); got != 0 {
		t.Errorf("no ribbon reserve = %d, want 0", got)
	}
}

// TestRibbonTopBarShiftsBody: a top-bar ribbon pushes the card body start down by exactly
// the band height; a corner ribbon leaves it unchanged.
func TestRibbonTopBarShiftsBody(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(4)}
	base := r.cardHeaderBottom(box, cardChrome{header: "Scale", size: CardSizeMD})
	bar := r.cardHeaderBottom(box, cardChrome{header: "Scale", size: CardSizeMD, ribbon: &Ribbon{Text: "POPULAR", Position: RibbonTopBar}})
	if bar-base != ribbonTopBarH {
		t.Errorf("top-bar body shift = %d, want the band height %d", bar-base, ribbonTopBarH)
	}
	corner := r.cardHeaderBottom(box, cardChrome{header: "Scale", size: CardSizeMD, ribbon: &Ribbon{Position: RibbonCornerStar}})
	if corner != base {
		t.Errorf("corner ribbon shifted the body to %d, want unchanged %d", corner, base)
	}
}

// TestRibbonColors: a nil Color resolves to accent; the default TextColor auto-contrasts
// to inverse on the dark accent fill; an explicit color is honored.
func TestRibbonColors(t *testing.T) {
	r := newTestRenderer(t)
	if got := ribbonColorRole(&Ribbon{}); got != pptx.ColorAccent {
		t.Errorf("nil ribbon color = %v, want ColorAccent", got)
	}
	warm := ColorAccentWarm
	if got := ribbonColorRole(&Ribbon{Color: &warm}); got != pptx.ColorAccentWarm {
		t.Errorf("explicit ribbon color = %v, want ColorAccentWarm", got)
	}
	if got := r.ribbonTextColor(&Ribbon{}, pptx.ColorAccent); got != pptx.TokenTextColor(pptx.TextInverse) {
		t.Errorf("default ribbon text on accent = %#v, want inverse", got)
	}
	if got := r.ribbonTextColor(&Ribbon{TextColor: TextAccent}, pptx.ColorAccent); got != pptx.TokenTextColor(pptx.TextAccent) {
		t.Errorf("explicit ribbon text = %#v, want accent", got)
	}
}

// TestRenderCardRibbon_ShapeCount: a top bar / corner tab emits a tab + text (2 shapes);
// a star emits one glyph.
func TestRenderCardRibbon_ShapeCount(t *testing.T) {
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(3), H: pptx.In(4)}
	cases := []struct {
		name string
		rb   Ribbon
		want int
	}{
		{"topbar", Ribbon{Text: "MOST POPULAR", Position: RibbonTopBar}, 2},
		{"corner-tl", Ribbon{Text: "NEW", Position: RibbonCornerTL}, 2},
		{"corner-tr", Ribbon{Text: "NEW", Position: RibbonCornerTR}, 2},
		{"star", Ribbon{Position: RibbonCornerStar}, 1},
	}
	for _, c := range cases {
		r := newTestRenderer(t)
		r.cfg.icons = icons.Curated()
		ps := r.pres.AddSlide()
		r.renderCardRibbon(ps, box, &c.rb, pptx.In(0.2))
		if r.stats.Shapes != c.want {
			t.Errorf("%s: emitted %d shapes, want %d", c.name, r.stats.Shapes, c.want)
		}
	}
}
