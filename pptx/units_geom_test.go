package pptx_test

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func TestUnitConstructors(t *testing.T) {
	cases := []struct {
		name string
		got  pptx.EMU
		want pptx.EMU
	}{
		{"1in", pptx.In(1), 914400},
		{"2in", pptx.In(2), 1828800},
		{"1cm", pptx.Cm(1), 360000},
		{"1mm", pptx.Cm(0.1), 36000},
		{"72pt", pptx.Pt(72), 914400},
		{"12pt", pptx.Pt(12), 152400},
		{"96px", pptx.Px(96), 914400},
		{"half-inch", pptx.In(0.5), 457200},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %d EMU, want %d", c.name, c.got, c.want)
		}
	}
}

func TestUnitRoundTrip(t *testing.T) {
	e := pptx.In(2)
	if got := e.Inches(); got != 2 {
		t.Errorf("Inches: got %v want 2", got)
	}
	if got := pptx.Pt(18).Points(); got != 18 {
		t.Errorf("Points: got %v want 18", got)
	}
	if got := pptx.Cm(3).Centimeters(); got != 3 {
		t.Errorf("Centimeters: got %v want 3", got)
	}
	if got := pptx.Px(48).Pixels(); got != 48 {
		t.Errorf("Pixels: got %v want 48", got)
	}
}

func TestSlideSizeConstants(t *testing.T) {
	if pptx.Slide16x9Width != pptx.In(13.333333333) && pptx.Slide16x9Width != 12192000 {
		t.Errorf("16x9 width: got %d", pptx.Slide16x9Width)
	}
	if pptx.Slide4x3Width != 9144000 {
		t.Errorf("4x3 width: got %d", pptx.Slide4x3Width)
	}
}

func TestBoxGeometry(t *testing.T) {
	b := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(4), H: pptx.In(2)}
	if b.Right() != pptx.In(5) {
		t.Errorf("Right: got %d want %d", b.Right(), pptx.In(5))
	}
	if b.Bottom() != pptx.In(3) {
		t.Errorf("Bottom: got %d want %d", b.Bottom(), pptx.In(3))
	}
	if b.Position() != (pptx.Position{X: pptx.In(1), Y: pptx.In(1)}) {
		t.Errorf("Position: %+v", b.Position())
	}
	if b.Size() != (pptx.Size{W: pptx.In(4), H: pptx.In(2)}) {
		t.Errorf("Size: %+v", b.Size())
	}
}

func TestBoxInset(t *testing.T) {
	b := pptx.Box{X: 0, Y: 0, W: pptx.In(4), H: pptx.In(4)}
	got := b.Inset(pptx.UniformInset(pptx.In(1)))
	want := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(2), H: pptx.In(2)}
	if got != want {
		t.Errorf("Inset: got %+v want %+v", got, want)
	}
}

func TestAnchorPoint(t *testing.T) {
	b := pptx.Box{X: 0, Y: 0, W: 1000, H: 800}
	cases := map[pptx.Anchor]pptx.Position{
		pptx.AnchorTopLeft:      {X: 0, Y: 0},
		pptx.AnchorCenter:       {X: 500, Y: 400},
		pptx.AnchorBottomRight:  {X: 1000, Y: 800},
		pptx.AnchorTopRight:     {X: 1000, Y: 0},
		pptx.AnchorBottomCenter: {X: 500, Y: 800},
		pptx.AnchorCenterLeft:   {X: 0, Y: 400},
	}
	for a, want := range cases {
		if got := a.Point(b); got != want {
			t.Errorf("anchor %d: got %+v want %+v", a, got, want)
		}
	}
}
