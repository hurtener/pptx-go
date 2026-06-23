package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
)

// White-box tests for the Button composer (R12.1, D-094): the size scale, the
// content-fit width geometry, the tone→style mapping, and the per-button shape count.

// TestButtonMetrics_SizeScale: SM < MD < LG on height, and every metric is positive.
func TestButtonMetrics_SizeScale(t *testing.T) {
	sm, md, lg := buttonMetrics(ButtonSM), buttonMetrics(ButtonMD), buttonMetrics(ButtonLG)
	if sm.height >= md.height || md.height >= lg.height {
		t.Errorf("button heights not monotonically increasing: SM=%d MD=%d LG=%d", sm.height, md.height, lg.height)
	}
	for name, m := range map[string]buttonMetric{"SM": sm, "MD": md, "LG": lg} {
		if m.height <= 0 || m.padX <= 0 || m.gap <= 0 || m.iconSz <= 0 {
			t.Errorf("%s metric has a non-positive field: %+v", name, m)
		}
	}
}

// TestButtonWidthOf_ContentFit: the width grows with the label and with each icon,
// is floored at the pill height (circular minimum), and is clamped to the box.
func TestButtonWidthOf_ContentFit(t *testing.T) {
	r := newTestRenderer(t)
	m := buttonMetrics(ButtonMD)
	box := pptx.In(10)

	short := buttonWidthOf(r.theme, Button{Label: "Go"}, m, box)
	long := buttonWidthOf(r.theme, Button{Label: "Talk to the whole team today"}, m, box)
	if long <= short {
		t.Errorf("a longer label did not widen the button: short=%d long=%d", short, long)
	}

	withIcon := buttonWidthOf(r.theme, Button{Label: "Go", TrailingIcon: "arrow-right"}, m, box)
	if withIcon <= short {
		t.Errorf("a trailing icon did not widen the button: plain=%d icon=%d", short, withIcon)
	}
	bothIcons := buttonWidthOf(r.theme, Button{Label: "Go", LeadingIcon: "star", TrailingIcon: "arrow-right"}, m, box)
	if bothIcons <= withIcon {
		t.Errorf("a second icon did not widen the button: one=%d two=%d", withIcon, bothIcons)
	}

	// Circular floor: a one-character label stays at least pill-height wide.
	if tiny := buttonWidthOf(r.theme, Button{Label: "x"}, m, box); tiny < m.height {
		t.Errorf("tiny label width %d is below the circular floor %d", tiny, m.height)
	}

	// Clamp: an enormous label cannot exceed the box.
	huge := buttonWidthOf(r.theme, Button{Label: "an extremely long label that vastly exceeds any reasonable button box width"}, m, pptx.In(2))
	if huge > pptx.In(2) {
		t.Errorf("clamped width %d exceeds the box %d", huge, pptx.In(2))
	}
}

// TestButtonToneStyle: ghost is an outline (two shape options: no-fill + a line); the
// solid tones carry exactly one fill option; every tone yields a non-nil label color.
func TestButtonToneStyle(t *testing.T) {
	cases := []struct {
		tone     ButtonTone
		wantOpts int
	}{
		{ButtonPrimary, 1},
		{ButtonAccentAlt, 1},
		{ButtonNeutral, 1},
		{ButtonGhost, 2},
	}
	for _, c := range cases {
		opts, color := buttonToneStyle(c.tone)
		if len(opts) != c.wantOpts {
			t.Errorf("tone %d: got %d shape options, want %d", c.tone, len(opts), c.wantOpts)
		}
		if color == nil {
			t.Errorf("tone %d: nil label color", c.tone)
		}
	}
}

// TestRenderButton_ShapeCount: a plain button emits a pill + a label (2 shapes); each
// icon adds one. The composer is deterministic about how many shapes it draws.
func TestRenderButton_ShapeCount(t *testing.T) {
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(6), H: pptx.In(0.5)}
	cases := []struct {
		name string
		v    Button
		want int
	}{
		{"plain", Button{Label: "Go"}, 2},
		{"trailing", Button{Label: "Go", TrailingIcon: "arrow-right"}, 3},
		{"both", Button{Label: "Go", LeadingIcon: "star", TrailingIcon: "arrow-right"}, 4},
	}
	for _, c := range cases {
		r := newTestRenderer(t)
		r.cfg.icons = icons.Curated() // Render builds this; the bare test renderer does not
		ps := r.pres.AddSlide()
		r.renderButton(ps, box, c.v, HAlignLeft)
		if r.stats.Shapes != c.want {
			t.Errorf("%s: emitted %d shapes, want %d (warnings: %v)", c.name, r.stats.Shapes, c.want, r.stats.Warnings)
		}
	}
}
