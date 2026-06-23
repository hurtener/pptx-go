package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for the Checklist node (R12.2, D-095): a filled glyph is
// emitted (not an empty font checkbox), the text renders, columns reflow, an unknown
// item icon fails Stage-1, and identical input is byte-identical across workers.

// TestChecklist_FilledGlyphNotCheckbox: the checklist emits a custGeom glyph (the
// filled check), and does NOT fall back to a PPTX checkbox-bullet autonumber.
func TestChecklist_FilledGlyphNotCheckbox(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "feat", Nodes: []scene.SlideNode{
			scene.Checklist{Items: []scene.ChecklistItem{
				{Text: rt("Understands your data"), State: scene.CheckDone},
				{Text: rt("Will not do that"), State: scene.CheckNo},
			}},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")

	if !strings.Contains(xml, "<a:custGeom>") {
		t.Error("checklist did not emit a filled custGeom glyph")
	}
	if strings.Contains(xml, "buChar") || strings.Contains(xml, "buAutoNum") {
		t.Error("checklist fell back to a font bullet/checkbox (the empty-box bug)")
	}
	if !strings.Contains(xml, "Understands your data") {
		t.Error("checklist item text missing")
	}
}

// TestChecklist_Columns: a 2-column checklist renders every item's text.
func TestChecklist_Columns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "two", Nodes: []scene.SlideNode{
			scene.Checklist{Columns: 2, Items: []scene.ChecklistItem{
				{Text: rt("alpha")}, {Text: rt("bravo")}, {Text: rt("charlie")}, {Text: rt("delta")},
			}},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, w := range []string{"alpha", "bravo", "charlie", "delta"} {
		if !strings.Contains(xml, w) {
			t.Errorf("2-column checklist missing item %q", w)
		}
	}
}

// TestChecklist_UnknownIconFails: an unknown per-item icon override fails Stage-1
// validation at Render, naming the icon.
func TestChecklist_UnknownIconFails(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "bad", Nodes: []scene.SlideNode{
			scene.Checklist{Items: []scene.ChecklistItem{
				{Text: rt("ok"), Icon: "no-such-glyph"},
			}},
		}},
	}}
	_, err := scene.Render(pptx.New(), sc)
	if err == nil {
		t.Fatal("Render accepted an unknown checklist icon; want a Stage-1 error")
	}
	if !strings.Contains(err.Error(), "no-such-glyph") {
		t.Errorf("error %q should name the unknown icon", err)
	}
}

// TestChecklist_Deterministic: identical input renders byte-identically across workers.
func TestChecklist_Deterministic(t *testing.T) {
	warm := scene.ColorAccentWarm
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "c", Nodes: []scene.SlideNode{
			scene.Checklist{Columns: 3, Fill: true, GlyphTone: &warm, Items: []scene.ChecklistItem{
				{Text: rt("one"), State: scene.CheckDone}, {Text: rt("two"), State: scene.CheckNo},
				{Text: rt("three"), State: scene.CheckNeutral}, {Text: rt("four")},
			}},
		}},
	}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("checklist render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
