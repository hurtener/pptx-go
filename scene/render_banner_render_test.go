package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for the Banner node (R12.6, D-097): a filled strip with
// lead/body, an embedded trailing button, an accent fill by default, an unknown icon
// failing Stage-1, and determinism.

// TestBanner_FilledStrip: a banner emits a roundRect strip and its lead + body text.
func TestBanner_FilledStrip(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "promo", Nodes: []scene.SlideNode{
			scene.Banner{
				Lead: rt("Run it internally, sell it externally"),
				Body: rt("the power of an agentic platform"),
				Icon: "star", Fill: scene.ColorAccent,
			},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, `prst="roundRect"`) {
		t.Error("banner did not emit a roundRect strip")
	}
	for _, w := range []string{"Run it internally", "agentic platform"} {
		if !strings.Contains(xml, w) {
			t.Errorf("banner missing text %q", w)
		}
	}
	if !strings.Contains(xml, "<a:custGeom>") {
		t.Error("banner leading icon did not emit a custGeom glyph")
	}
}

// TestBanner_TrailingButton: an embedded Trailing button renders its label and pill.
func TestBanner_TrailingButton(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "cta", Nodes: []scene.SlideNode{
			scene.Banner{Lead: rt("$0 to start"), Fill: scene.ColorAccent, Trailing: []scene.SlideNode{
				scene.Button{Label: "Start free", Tone: scene.ButtonNeutral, TrailingIcon: "arrow-right"},
			}},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "Start free") {
		t.Error("embedded trailing button label missing")
	}
	// strip pill + button pill = at least 2 roundRects.
	if n := strings.Count(xml, `prst="roundRect"`); n < 2 {
		t.Errorf("got %d roundRects, want >= 2 (strip + button)", n)
	}
}

// TestBanner_DefaultFillAccent: a Banner with no Fill still renders a filled strip (the
// zero ColorCanvas maps to accent), not an invisible/absent fill.
func TestBanner_DefaultFillAccent(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "b", Nodes: []scene.SlideNode{scene.Banner{Lead: rt("Takeaway")}}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, `prst="roundRect"`) {
		t.Error("default-fill banner did not emit a strip")
	}
}

// TestBanner_UnknownIconFails: an unknown banner (or trailing child) icon fails Stage-1.
func TestBanner_UnknownIconFails(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "bad", Nodes: []scene.SlideNode{scene.Banner{Lead: rt("x"), Icon: "no-such-icon"}}},
	}}
	if _, err := scene.Render(pptx.New(), sc); err == nil || !strings.Contains(err.Error(), "no-such-icon") {
		t.Fatalf("want a Stage-1 error naming the icon, got %v", err)
	}
	sc2 := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "bad2", Nodes: []scene.SlideNode{scene.Banner{Lead: rt("x"), Trailing: []scene.SlideNode{
			scene.Button{Label: "go", LeadingIcon: "nope-icon"},
		}}}},
	}}
	if _, err := scene.Render(pptx.New(), sc2); err == nil || !strings.Contains(err.Error(), "nope-icon") {
		t.Fatalf("want a Stage-1 error naming the trailing-child icon, got %v", err)
	}
}

// TestBanner_Deterministic: identical input renders byte-identically across workers.
func TestBanner_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "b", Variant: scene.VariantDark, Nodes: []scene.SlideNode{
			scene.Banner{Lead: rt("Lead phrase"), Body: rt("supporting body"), Icon: "star",
				Fill: scene.ColorAccent, Trailing: []scene.SlideNode{scene.Button{Label: "Go", TrailingIcon: "arrow-right"}}},
		}},
	}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("banner render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
