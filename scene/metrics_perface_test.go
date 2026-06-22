package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestNaturalWidth_PerFaceFactor (Phase 34, R9.5, D-064): the wrap estimator uses
// a role's AvgCharWidth when set, and falls back to the built-in sans factor
// (byte-identical) when unset.
func TestNaturalWidth_PerFaceFactor(t *testing.T) {
	rt := RichText{{Text: "abcdefghij", Style: RunStyle{TypeRole: pptx.TypeBody}}}

	def := pptx.DefaultTheme() // TypeBody AvgCharWidth 0 → built-in fallback
	wDef := naturalWidth(rt, def)

	wide := pptx.DefaultTheme().Clone()
	b := wide.ResolveType(pptx.TypeBody)
	b.AvgCharWidth = 0.7 // a wider face
	wide.Typography[pptx.TypeBody] = b
	if wWide := naturalWidth(rt, wide); wWide <= wDef {
		t.Errorf("a wider AvgCharWidth should widen the estimate: default=%d wide=%d", wDef, wWide)
	}

	// An explicit 0.5 equals the built-in fallback — the unset face is byte-identical.
	half := pptx.DefaultTheme().Clone()
	h := half.ResolveType(pptx.TypeBody)
	h.AvgCharWidth = 0.5
	half.Typography[pptx.TypeBody] = h
	if got := naturalWidth(rt, half); got != wDef {
		t.Errorf("explicit 0.5 = %d, want the fallback estimate %d", got, wDef)
	}
}
