package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestDataMark_Donut is R14.8 part 2 acceptance: a donut at 0.92 renders a value
// arc starting at 12 o'clock (270°) with "92%" centered, as native blockArc
// shapes in soul colors, conformant, no warnings.
func TestDataMark_Donut(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "donut",
		Nodes: []scene.SlideNode{scene.DataMark{Kind: scene.DataMarkDonut, Value: 0.92, Label: "92%"}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("donut: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, `prst="blockArc"`); n < 2 {
		t.Errorf("donut: want >=2 blockArcs (value + track), got %d", n)
	}
	// Value arc starts at 270° (= 16200000 in 60000ths) — top of the ring.
	if !strings.Contains(slide, `<a:gd name="adj1" fmla="val 16200000"/>`) {
		t.Errorf("donut: value arc should start at 270°:\n%s", slide)
	}
	if !strings.Contains(slide, "<a:t>92%</a:t>") {
		t.Errorf("donut: missing centered label")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("donut deck failed conformance:\n%s", rep)
	}
}

// TestDataMark_Gauge renders a gauge and checks it emits native blockArcs + label.
func TestDataMark_Gauge(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "gauge",
		Nodes: []scene.SlideNode{scene.DataMark{Kind: scene.DataMarkGauge, Value: 0.5, Label: "50"}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("gauge: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `prst="blockArc"`) {
		t.Errorf("gauge: missing blockArc")
	}
	if !strings.Contains(slide, "<a:t>50</a:t>") {
		t.Errorf("gauge: missing label")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("gauge deck failed conformance:\n%s", rep)
	}
}

// TestDataMark_DonutFullAndEmpty verifies the value=1 (no track arc) and value=0
// (no value arc) edge cases render without panic.
func TestDataMark_DonutFullAndEmpty(t *testing.T) {
	for _, val := range []float64{0, 1} {
		sc := scene.Scene{Slides: []scene.SceneSlide{{
			ID:    "donut-edge",
			Nodes: []scene.SlideNode{scene.DataMark{Kind: scene.DataMarkDonut, Value: val}},
		}}}
		if _, stats := render(t, sc); len(stats.Warnings) != 0 {
			t.Errorf("donut value=%g: warnings: %+v", val, stats.Warnings)
		}
	}
}

// TestDataMark_ArcDeterministic guards worker-count independence of the arc marks.
func TestDataMark_ArcDeterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "arcs",
		Nodes: []scene.SlideNode{
			scene.DataMark{Kind: scene.DataMarkDonut, Value: 0.92, Label: "92%"},
			scene.DataMark{Kind: scene.DataMarkGauge, Value: 0.5, Label: "50"},
		},
	}}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("arc marks not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}

// TestBlockArc_Builder is the builder-level check: AddBlockArc emits a blockArc
// with adj1/adj2/adj3 and survives a write → reopen → re-write (G6, structural).
func TestBlockArc_Builder(t *testing.T) {
	p := pptx.New()
	sl := p.AddSlide("")
	sl.AddBlockArc(pptx.Box{X: 0, Y: 0, W: pptx.In(2), H: pptx.In(2)}, 270, 180, 0.6,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if x := zipPart(t, data, "ppt/slides/slide1.xml"); !strings.Contains(x, `prst="blockArc"`) || !strings.Contains(x, `name="adj3"`) {
		t.Fatalf("blockArc not emitted on first write:\n%s", x)
	}
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	out, err := re.WriteToBytes()
	if err != nil {
		t.Fatalf("re-write: %v", err)
	}
	if x := zipPart(t, out, "ppt/slides/slide1.xml"); !strings.Contains(x, `prst="blockArc"`) || !strings.Contains(x, `name="adj3"`) {
		t.Errorf("blockArc adjust guides did not survive round-trip:\n%s", x)
	}
}
