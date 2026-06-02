package ornaments

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestCornerArms_OrientToCorner checks each rotation quadrant puts the L's
// corner (the junction of the two bars) at the matching box corner.
func TestCornerArms_OrientToCorner(t *testing.T) {
	box := pptx.Box{X: pptx.In(2), Y: pptx.In(1), W: pptx.In(2), H: pptx.In(2)}
	thick := pptx.In(0.2)
	arm := pptx.In(1)

	// The junction is where both bars share a corner.
	cases := []struct {
		name   string
		q      int
		jx, jy pptx.EMU // expected junction (corner of the L)
	}{
		{"top-left", 0, box.X, box.Y},
		{"top-right", 1, box.Right(), box.Y},
		{"bottom-right", 2, box.Right(), box.Bottom()},
		{"bottom-left", 3, box.X, box.Bottom()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, v := cornerArms(box, thick, arm, tc.q)
			// Both bars must touch the junction corner.
			if !touches(h, tc.jx, tc.jy) || !touches(v, tc.jx, tc.jy) {
				t.Errorf("q=%d: bars do not meet at (%d,%d); h=%+v v=%+v", tc.q, tc.jx, tc.jy, h, v)
			}
			// Both bars must stay within the box.
			for _, b := range []pptx.Box{h, v} {
				if b.X < box.X || b.Y < box.Y || b.Right() > box.Right() || b.Bottom() > box.Bottom() {
					t.Errorf("q=%d: bar %+v escapes box %+v", tc.q, b, box)
				}
			}
		})
	}
}

// touches reports whether the box has a corner at (x,y).
func touches(b pptx.Box, x, y pptx.EMU) bool {
	cornerX := b.X == x || b.Right() == x
	cornerY := b.Y == y || b.Bottom() == y
	return cornerX && cornerY
}

// TestQuadrant_Snapping checks rotation snaps to the nearest 90°.
func TestQuadrant_Snapping(t *testing.T) {
	cases := map[float64]int{0: 0, 44: 0, 46: 1, 90: 1, 135: 2, 180: 2, 270: 3, 360: 0, -90: 3}
	for deg, want := range cases {
		if got := quadrant(deg); got != want {
			t.Errorf("quadrant(%v) = %d, want %d", deg, got, want)
		}
	}
}
