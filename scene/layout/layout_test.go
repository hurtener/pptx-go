package layout_test

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/layout"
)

var parent = pptx.Box{X: 0, Y: 0, W: pptx.In(10), H: pptx.In(6)}

// TestColumns_Ratios is acceptance criterion 1: 1:1 / 1:2 / 2:1 produce the
// expected widths and fit inside the parent.
func TestColumns_Ratios(t *testing.T) {
	gap := pptx.In(0.2)
	tests := []struct {
		name    string
		weights []int
		ratio   float64 // right / left
	}{
		{"1:1", []int{1, 1}, 1},
		{"1:2", []int{1, 2}, 2},
		{"2:1", []int{2, 1}, 0.5},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cols := layout.Columns(parent, tc.weights, gap)
			if len(cols) != 2 {
				t.Fatalf("got %d columns, want 2", len(cols))
			}
			// Columns fit the parent exactly (widths + gap == parent width).
			if cols[0].W+cols[1].W+gap != parent.W {
				t.Errorf("widths %d+%d+gap %d != parent %d", cols[0].W, cols[1].W, gap, parent.W)
			}
			// Right edge of col1 == parent right.
			if cols[1].Right() != parent.Right() {
				t.Errorf("col1 right %d != parent right %d", cols[1].Right(), parent.Right())
			}
			// Ratio holds (within 1 EMU rounding).
			got := float64(cols[1].W) / float64(cols[0].W)
			if got < tc.ratio-0.001 || got > tc.ratio+0.001 {
				t.Errorf("ratio = %.3f, want %.3f", got, tc.ratio)
			}
		})
	}
}

// TestGrid_Dims is acceptance criterion 2: a weighted multi-column grid produces
// the expected cell widths and row count.
func TestGrid_Dims(t *testing.T) {
	gap := pptx.In(0.2)
	cells := layout.Grid(parent, 3, []int{1, 1, 2}, gap, 6) // 6 cells → 2 rows
	if len(cells) != 6 {
		t.Fatalf("got %d cells, want 6", len(cells))
	}
	// Column widths follow 1:1:2.
	if cells[0].W != cells[1].W {
		t.Errorf("col0 (%d) != col1 (%d)", cells[0].W, cells[1].W)
	}
	if cells[2].W != 2*cells[0].W {
		t.Errorf("col2 width %d != 2× col0 %d", cells[2].W, cells[0].W)
	}
	// Cell 0 (row 0) and cell 3 (row 1, same column) share a column.
	if cells[0].X != cells[3].X || cells[0].W != cells[3].W {
		t.Errorf("cell 0 and 3 not column-aligned: %+v / %+v", cells[0], cells[3])
	}
	// Row 1 is below row 0.
	if cells[3].Y <= cells[0].Y {
		t.Errorf("row 1 Y %d not below row 0 Y %d", cells[3].Y, cells[0].Y)
	}
	// Cells fit within the parent.
	for i, c := range cells {
		if c.Right() > parent.Right()+1 || c.Bottom() > parent.Bottom()+1 {
			t.Errorf("cell %d overflows parent: %+v", i, c)
		}
	}
}

// TestColumns_EqualFallback proves a non-positive weight set falls back to equal.
func TestColumns_EqualFallback(t *testing.T) {
	cols := layout.Columns(parent, []int{0, 0, 0}, 0)
	if len(cols) != 3 {
		t.Fatalf("got %d columns, want 3", len(cols))
	}
	if cols[0].W != cols[1].W || cols[1].W != cols[2].W {
		t.Errorf("equal fallback produced unequal widths: %+v", cols)
	}
}
