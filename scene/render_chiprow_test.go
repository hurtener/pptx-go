package scene

import (
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene/icons"
)

// White-box tests for the ChipRow composer (R12.5, D-096): content-fit chip width, the
// greedy wrap packer, the leading-label accounting, and the per-row shape count.

func chipSpecs(n int, label string) []ChipSpec {
	cs := make([]ChipSpec, n)
	for i := range cs {
		cs[i] = ChipSpec{Label: label}
	}
	return cs
}

// TestChipWidthOf: a chip's width grows with its label and with a leading icon.
func TestChipWidthOf(t *testing.T) {
	r := newTestRenderer(t)
	short := chipWidthOf(r.theme, ChipSpec{Label: "HR"})
	long := chipWidthOf(r.theme, ChipSpec{Label: "Legal & Compliance"})
	if long <= short {
		t.Errorf("longer chip label did not widen: short=%d long=%d", short, long)
	}
	withIcon := chipWidthOf(r.theme, ChipSpec{Label: "HR", Icon: "star"})
	if withIcon <= short {
		t.Errorf("a chip icon did not widen the chip: plain=%d icon=%d", short, withIcon)
	}
}

// TestChipRowLines_Wrap: with Wrap, chips that exceed the width pack onto multiple
// lines; without Wrap they stay on one line.
func TestChipRowLines_Wrap(t *testing.T) {
	r := newTestRenderer(t)
	v := ChipRow{Wrap: true, Chips: chipSpecs(12, "Capability")}
	box := pptx.In(4) // narrow, forces wrapping
	lines := chipRowLines(v, box, r.theme)
	if len(lines) < 2 {
		t.Errorf("Wrap: 12 chips in a narrow box packed into %d lines, want >= 2", len(lines))
	}
	// Every line must fit the box width.
	for i, ln := range lines {
		if w := chipRowLineWidth(v, ln, i, r.theme); w > box {
			t.Errorf("line %d width %d exceeds box %d", i, w, box)
		}
	}

	one := ChipRow{Wrap: false, Chips: chipSpecs(12, "Capability")}
	if got := chipRowLines(one, box, r.theme); len(got) != 1 {
		t.Errorf("Wrap=false packed into %d lines, want 1", len(got))
	}
}

// TestChipRowLines_LabelOnLine0: a leading label consumes width on the first line, so a
// labeled row wraps at least as early as an unlabeled one.
func TestChipRowLines_LabelOnLine0(t *testing.T) {
	r := newTestRenderer(t)
	box := pptx.In(4)
	labeled := ChipRow{Wrap: true, Label: "COMMON BUILDS", Chips: chipSpecs(6, "Finance")}
	unlabeled := ChipRow{Wrap: true, Chips: chipSpecs(6, "Finance")}
	if chipRowLineWidth(labeled, chipRowLines(labeled, box, r.theme)[0], 0, r.theme) == 0 {
		t.Error("labeled line 0 has zero width")
	}
	if len(chipRowLines(labeled, box, r.theme)) < len(chipRowLines(unlabeled, box, r.theme)) {
		t.Error("the labeled row should wrap no later than the unlabeled row")
	}
}

// TestRenderChipRow_ShapeCount: each chip emits a pill + a label (2 shapes); a leading
// label adds one shape.
func TestRenderChipRow_ShapeCount(t *testing.T) {
	r := newTestRenderer(t)
	r.cfg.icons = icons.Curated()
	ps := r.pres.AddSlide()
	box := pptx.Box{X: 0, Y: 0, W: pptx.In(9), H: pptx.In(1)}
	v := ChipRow{Label: "TAGS", Wrap: true, Chips: chipSpecs(3, "tag")}
	r.renderChipRow(ps, box, v, HAlignLeft)
	// 1 label + 3 chips × (pill + text) = 7 shapes.
	if r.stats.Shapes != 7 {
		t.Errorf("emitted %d shapes, want 7 (warnings: %v)", r.stats.Shapes, r.stats.Warnings)
	}
}
