package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestBackgroundGradientStops covers the multi-stop validator directly (D-105):
// 2..8 ascending stops in [0,1] map 1:1; any violation returns (nil, false) so
// renderBackground warns and skips (RFC §10.2, no panic).
func TestBackgroundGradientStops(t *testing.T) {
	stop := func(p float64, r pptx.ColorRole) GradientStop { return GradientStop{Pos: p, Color: r} }
	tests := []struct {
		name string
		in   []GradientStop
		want int // expected stop count; 0 means invalid (ok == false)
	}{
		{"two ok", []GradientStop{stop(0, pptx.ColorAccent), stop(1, pptx.ColorCanvas)}, 2},
		{"three ok", []GradientStop{stop(0, pptx.ColorAccent), stop(0.5, pptx.ColorAccentAlt), stop(1, pptx.ColorCanvas)}, 3},
		{"interior only", []GradientStop{stop(0.2, pptx.ColorAccent), stop(0.8, pptx.ColorCanvas)}, 2},
		{"too few", []GradientStop{stop(0, pptx.ColorAccent)}, 0},
		{"too many", make([]GradientStop, 9), 0},
		{"descending", []GradientStop{stop(0.6, pptx.ColorAccent), stop(0.2, pptx.ColorCanvas)}, 0},
		{"equal pos", []GradientStop{stop(0.5, pptx.ColorAccent), stop(0.5, pptx.ColorCanvas)}, 0},
		{"above one", []GradientStop{stop(0, pptx.ColorAccent), stop(1.2, pptx.ColorCanvas)}, 0},
		{"below zero", []GradientStop{stop(-0.1, pptx.ColorAccent), stop(1, pptx.ColorCanvas)}, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := backgroundGradientStops(tc.in)
			if tc.want == 0 {
				if ok {
					t.Errorf("%s: expected invalid, got ok with %d stops", tc.name, len(got))
				}
				return
			}
			if !ok {
				t.Fatalf("%s: expected valid, got invalid", tc.name)
			}
			if len(got) != tc.want {
				t.Errorf("%s: stops = %d, want %d", tc.name, len(got), tc.want)
			}
			for i, s := range got {
				if s.Color == nil {
					t.Errorf("%s: stop %d has nil color", tc.name, i)
				}
			}
		})
	}
}
