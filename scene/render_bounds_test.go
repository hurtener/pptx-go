package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// White-box tests for the R11.3 container-slide-bounds clamp (D-083).

// TestClampToSafeArea_ShrinksOverflow: a box whose bottom runs past the safe area
// is shrunk so its bottom == safeArea.Bottom(), and a warning is logged.
func TestClampToSafeArea_ShrinksOverflow(t *testing.T) {
	r := newTestRenderer(t)
	sa := r.safeArea()
	// A box that starts inside the region but extends an inch past the bottom.
	over := pptx.Box{X: sa.X, Y: sa.Y, W: sa.W, H: sa.H + pptx.In(1)}
	got := r.clampToSafeArea(over, "s1")
	if got.Bottom() != sa.Bottom() {
		t.Errorf("clamped bottom = %d, want safeArea bottom %d", got.Bottom(), sa.Bottom())
	}
	if got.H >= over.H {
		t.Errorf("clamped H %d should be < original %d", got.H, over.H)
	}
	if len(r.stats.Warnings) == 0 {
		t.Error("expected a clamp warning")
	}
}

// TestClampToSafeArea_FitsByteIdentical: a box already within the safe area is
// returned unchanged with no warning (the byte-identical path).
func TestClampToSafeArea_FitsByteIdentical(t *testing.T) {
	r := newTestRenderer(t)
	sa := r.safeArea()
	fits := pptx.Box{X: sa.X, Y: sa.Y, W: sa.W, H: sa.H - pptx.In(1)}
	got := r.clampToSafeArea(fits, "s1")
	if got != fits {
		t.Errorf("fitting box should be unchanged: got %+v want %+v", got, fits)
	}
	// A box whose bottom is exactly the safe-area bottom is also unchanged.
	exact := pptx.Box{X: sa.X, Y: sa.Y, W: sa.W, H: sa.H}
	if got := r.clampToSafeArea(exact, "s1"); got != exact {
		t.Errorf("exact-fit box should be unchanged: got %+v want %+v", got, exact)
	}
	if len(r.stats.Warnings) != 0 {
		t.Errorf("fitting boxes should log no warning, got %d", len(r.stats.Warnings))
	}
}

// TestBentoBoxesWithinSafeArea: after the clamp, every bento cell box a tall bento
// emits sits inside the slide safe area (no off-slide cell).
func TestBentoBoxesWithinSafeArea(t *testing.T) {
	r := newTestRenderer(t)
	sa := r.safeArea()
	// A box taller than the safe area (as an over-full stack would hand a bento).
	box := pptx.Box{X: sa.X, Y: sa.Y, W: sa.W, H: sa.H * 2}
	box = r.clampToSafeArea(box, "s1")

	v := Bento{Columns: 2, Rows: []BentoRow{
		{Cells: []BentoCell{{Span: 1, Node: Card{Header: "a"}}, {Span: 1, Node: Card{Header: "b"}}}},
		{Cells: []BentoCell{{Span: 1, Node: Card{Header: "c"}}, {Span: 1, Node: Card{Header: "d"}}}},
		{Cells: []BentoCell{{Span: 1, Node: Card{Header: "e"}}, {Span: 1, Node: Card{Header: "f"}}}},
	}}
	gap := r.theme.ResolveSpace(pptx.SpaceMD)
	_, _, _, cells := bentoGeometry(box, v, gap, nil)
	for ri := range cells {
		for ci := range cells[ri] {
			if cells[ri][ci].Bottom() > sa.Bottom() {
				t.Errorf("cell [%d][%d] bottom %d exceeds safe area bottom %d", ri, ci, cells[ri][ci].Bottom(), sa.Bottom())
			}
		}
	}
}
