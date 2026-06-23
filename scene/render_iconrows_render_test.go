package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for IconRows (R12.7, D-100): icon + label + meta rows, the
// RowPill frame, an unknown icon failing Stage-1, and determinism.

// TestIconRows_RendersRows: rows render their icons (custGeom), labels, and meta.
func TestIconRows_RendersRows(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "caps", Nodes: []scene.SlideNode{
			scene.IconRows{Rows: []scene.IconRow{
				{Icon: "star", Label: rt("Chat & Q&A"), Meta: rt("core")},
				{Icon: "check", Label: rt("Specialized agents")},
			}},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:custGeom>") {
		t.Error("icon-rows did not emit custGeom glyphs")
	}
	for _, w := range []string{"core", "Specialized agents"} {
		if !strings.Contains(xml, w) {
			t.Errorf("icon-rows missing %q", w)
		}
	}
	// The ampersand in "Chat & Q&A" is XML-escaped in the run text.
	if !strings.Contains(xml, "Chat &amp; Q&amp;A") {
		t.Error("icon-rows label ampersand not escaped in the run text")
	}
}

// TestIconRows_RowPillFrame: a RowPill row emits a roundRect frame.
func TestIconRows_RowPillFrame(t *testing.T) {
	none, _ := render(t, scene.Scene{Slides: []scene.SceneSlide{
		{ID: "p", Nodes: []scene.SlideNode{scene.IconRows{Rows: []scene.IconRow{{Label: rt("plain")}}}}},
	}})
	pill, _ := render(t, scene.Scene{Slides: []scene.SceneSlide{
		{ID: "p", Nodes: []scene.SlideNode{scene.IconRows{Rows: []scene.IconRow{{Label: rt("framed"), Tone: scene.RowPill}}}}},
	}})
	xn := zipPart(t, none, "ppt/slides/slide1.xml")
	xp := zipPart(t, pill, "ppt/slides/slide1.xml")
	if strings.Count(xp, `prst="roundRect"`) <= strings.Count(xn, `prst="roundRect"`) {
		t.Error("RowPill should add a roundRect frame")
	}
}

// TestIconRows_UnknownIconFails: an unknown row icon fails Stage-1 validation at Render.
func TestIconRows_UnknownIconFails(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "bad", Nodes: []scene.SlideNode{
			scene.IconRows{Rows: []scene.IconRow{{Icon: "no-such-icon", Label: rt("x")}}},
		}},
	}}
	if _, err := scene.Render(pptx.New(), sc); err == nil || !strings.Contains(err.Error(), "no-such-icon") {
		t.Fatalf("want a Stage-1 error naming the icon, got %v", err)
	}
}

// TestIconRows_Deterministic: identical input renders byte-identically across workers.
func TestIconRows_Deterministic(t *testing.T) {
	rows := scene.IconRows{Fill: true, GlyphColor: scene.ColorAccent, Rows: []scene.IconRow{
		{Icon: "star", Label: rt("Salesforce · Slack"), Meta: rt("CRM"), Tone: scene.RowPill},
		{Icon: "check", Label: rt("Microsoft 365 · Workspace")},
		{Icon: "dot", Label: rt("Custom integrations")},
	}}
	card := scene.Card{Header: "Integrations", BodyVAlign: scene.VAlignFill, Body: []scene.SlideNode{rows}}
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "c", Nodes: []scene.SlideNode{card}}}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("icon-rows render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
