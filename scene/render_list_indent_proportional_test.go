package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box tests for R11.10 proportional list bullet indent (D-090).

// TestListTightIndent_AnchorByteIdentical: at the default 14pt body the IndentTight
// hanging indent is exactly In(0.25) — byte-identical to the R10.9 pinned value.
func TestListTightIndent_AnchorByteIdentical(t *testing.T) {
	r := newTestRenderer(t)
	if got := r.listTightIndent(); got != listTightIndentBase {
		t.Errorf("default-body tight indent = %d, want the In(0.25) anchor %d", got, listTightIndentBase)
	}
	if got := r.listTightIndent(); got != pptx.In(0.25) {
		t.Errorf("default-body tight indent = %d, want In(0.25) = %d", got, pptx.In(0.25))
	}
}

// TestListTightIndent_Proportional: the indent scales with the body type size — a
// larger body yields a proportionally larger indent, a smaller body a smaller one.
func TestListTightIndent_Proportional(t *testing.T) {
	r := newTestRenderer(t)
	base := r.listTightIndent() // 14pt

	// Double the body size → ~double the indent.
	big := r.theme.ResolveType(pptx.TypeBody)
	big.Size = 28
	r.theme.Typography[pptx.TypeBody] = big
	if got := r.listTightIndent(); got <= base {
		t.Errorf("28pt body tight indent %d should exceed the 14pt %d", got, base)
	}
	if want := listTightIndentBase * 2; r.listTightIndent() != want {
		t.Errorf("28pt body tight indent = %d, want 2× the anchor %d", r.listTightIndent(), want)
	}

	// Half the body size → ~half the indent.
	small := r.theme.ResolveType(pptx.TypeBody)
	small.Size = 7
	r.theme.Typography[pptx.TypeBody] = small
	if got := r.listTightIndent(); got >= base {
		t.Errorf("7pt body tight indent %d should be below the 14pt %d", got, base)
	}
}

// TestListTightIndent_GapTight: the bullet-to-text gap (the hanging indent) is
// meaningfully tighter than the builder's 0.5" default — at most ~In(0.3) at the
// default body — so it never reads as the oversized fixed gap the recreation showed.
func TestListTightIndent_GapTight(t *testing.T) {
	r := newTestRenderer(t)
	const builderDefault = pptx.EMU(457200) // In(0.5), the un-tightened default
	gap := r.listTightIndent()
	if gap >= builderDefault {
		t.Errorf("tight bullet gap %d should be smaller than the 0.5\" default %d", gap, builderDefault)
	}
	if gap > pptx.In(0.3) {
		t.Errorf("tight bullet gap %d exceeds the In(0.3) tight target (reads as oversized)", gap)
	}
}
