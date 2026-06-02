package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestDecorationBox_AnchorAlignment checks the box aligns its anchor-matching
// point to the region's anchor point — a top-left anchor sits in the corner
// (not centered half off-canvas), center centers, bottom-right hugs the corner.
func TestDecorationBox_AnchorAlignment(t *testing.T) {
	region := pptx.Box{X: 0, Y: 0, W: pptx.In(10), H: pptx.In(7.5)}
	size := pptx.Size{W: pptx.In(2), H: pptx.In(2)}

	cases := []struct {
		name   string
		anchor Anchor
		offset Position
		wantX  pptx.EMU
		wantY  pptx.EMU
	}{
		{"top-left", AnchorTopLeft, Position{}, 0, 0},
		{"top-left+offset", AnchorTopLeft, Position{X: pptx.In(1), Y: pptx.In(1)}, pptx.In(1), pptx.In(1)},
		{"center", AnchorCenter, Position{}, pptx.In(5) - pptx.In(1), pptx.In(3.75) - pptx.In(1)},
		{"bottom-right", AnchorBottomRight, Position{}, pptx.In(10) - pptx.In(2), pptx.In(7.5) - pptx.In(2)},
		{"top-right", AnchorTopRight, Position{}, pptx.In(10) - pptx.In(2), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			box := decorationBox(region, Decoration{Anchor: tc.anchor, Offset: tc.offset, Size: size})
			if box.X != tc.wantX || box.Y != tc.wantY {
				t.Errorf("box origin = (%d,%d), want (%d,%d)", box.X, box.Y, tc.wantX, tc.wantY)
			}
			if box.W != size.W || box.H != size.H {
				t.Errorf("box size = (%d,%d), want (%d,%d)", box.W, box.H, size.W, size.H)
			}
		})
	}

	// A bottom-right anchor keeps the box fully on-canvas (the bug it fixes:
	// centring would push it half off the bottom-right edge).
	box := decorationBox(region, Decoration{Anchor: AnchorBottomRight, Size: size})
	if box.Right() != region.Right() || box.Bottom() != region.Bottom() {
		t.Errorf("bottom-right box should hug the corner: right=%d bottom=%d", box.Right(), box.Bottom())
	}
}
