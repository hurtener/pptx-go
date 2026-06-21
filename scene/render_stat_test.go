package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Stat leaf node (Phase 28, R6) — white-box: the delta tone→color mapping and the
// not-flexible classification. Render/strip/validation/determinism are black-box
// in render_stat_render_test.go.

func TestDeltaToneColor(t *testing.T) {
	if deltaToneColor(DeltaUp) != pptx.TokenColor(pptx.ColorSuccess) {
		t.Error("DeltaUp should map to the success token color")
	}
	if deltaToneColor(DeltaDown) != pptx.TokenColor(pptx.ColorError) {
		t.Error("DeltaDown should map to the error token color")
	}
	if deltaToneColor(DeltaNeutral) != pptx.TokenTextColor(pptx.TextMuted) {
		t.Error("DeltaNeutral should map to the muted text color")
	}
}

// TestStat_NotFlexible is acceptance criterion 4 (part): a Stat is a fixed number
// block and does not stretch under VAlignFill, but a Grid of stats still grows.
func TestStat_NotFlexible(t *testing.T) {
	if isFlexible(Stat{}) {
		t.Error("Stat should not be flexible")
	}
	if !isFlexible(Grid{}) {
		t.Error("a Grid (of stats) should still be flexible")
	}
}
