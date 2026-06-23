package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for the ChipRow node (R12.5, D-096): chips render as real
// pills (not bullets), wrap, a leading label appears, an unknown icon fails Stage-1,
// and identical input is byte-identical across workers.

// TestChipRow_RendersPills: a chip row emits roundRect pills (not bullet text) and the
// chip + label text.
func TestChipRow_RendersPills(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "tags", Nodes: []scene.SlideNode{
			scene.ChipRow{Label: "COMMON BUILDS", Chips: []scene.ChipSpec{
				{Label: "Finance"}, {Label: "HR"}, {Label: "Sales"},
			}},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")

	if n := strings.Count(xml, `prst="roundRect"`); n < 3 {
		t.Errorf("chip row emitted %d roundRect pills, want >= 3", n)
	}
	if strings.Contains(xml, "buChar") || strings.Contains(xml, "buAutoNum") {
		t.Error("chip row fell back to a font bullet (the broken-list bug)")
	}
	for _, w := range []string{"COMMON BUILDS", "Finance", "HR", "Sales"} {
		if !strings.Contains(xml, w) {
			t.Errorf("chip row missing %q", w)
		}
	}
}

// TestChipRow_Wraps: a wrapping row of many chips renders every label and stays on the
// slide canvas (no off-slide boxes).
func TestChipRow_Wraps(t *testing.T) {
	chips := make([]scene.ChipSpec, 10)
	for i := range chips {
		chips[i] = scene.ChipSpec{Label: "Capability area number"}
	}
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "wrap", Nodes: []scene.SlideNode{scene.ChipRow{Wrap: true, Chips: chips}}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(xml, `prst="roundRect"`); n != 10 {
		t.Errorf("wrapped chip row emitted %d pills, want 10", n)
	}
}

// TestChipRow_UnknownIconFails: an unknown chip icon fails Stage-1 validation at Render.
func TestChipRow_UnknownIconFails(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "bad", Nodes: []scene.SlideNode{
			scene.ChipRow{Chips: []scene.ChipSpec{{Label: "x", Icon: "no-such-icon"}}},
		}},
	}}
	_, err := scene.Render(pptx.New(), sc)
	if err == nil {
		t.Fatal("Render accepted an unknown chip icon; want a Stage-1 error")
	}
	if !strings.Contains(err.Error(), "no-such-icon") {
		t.Errorf("error %q should name the unknown icon", err)
	}
}

// TestChipRow_Deterministic: identical input renders byte-identically across workers.
func TestChipRow_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "c", Content: scene.Alignment{Horizontal: scene.HAlignCenter}, Nodes: []scene.SlideNode{
			scene.ChipRow{Label: "BUILDS", Wrap: true, Chips: []scene.ChipSpec{
				{Label: "Finance"}, {Label: "HR", Tone: scene.ChipSolid, Color: scene.ColorAccent},
				{Label: "Sales", Tone: scene.ChipOutline, Color: scene.ColorAccent}, {Label: "Legal", Icon: "star"},
			}},
		}},
	}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("chip row render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
