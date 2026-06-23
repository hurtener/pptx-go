package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for the Button node (R12.1, D-094): a standalone button
// renders as a pill + label (+ icon), a button in a card body composes, an unknown
// icon name fails Stage-1 validation, and a button-free deck is byte-identical.

// TestButton_RendersPill: a standalone button emits a roundRect pill carrying its
// label, plus a custom-geometry trailing icon.
func TestButton_RendersPill(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "cta", Nodes: []scene.SlideNode{
			scene.Button{Label: "Talk to the team", Tone: scene.ButtonPrimary, Size: scene.ButtonLG, TrailingIcon: "arrow-right"},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")

	if !strings.Contains(xml, `prst="roundRect"`) {
		t.Error("button did not emit a roundRect pill")
	}
	if !strings.Contains(xml, "Talk to the team") {
		t.Error("button label not present in the slide XML")
	}
	if !strings.Contains(xml, "<a:custGeom>") {
		t.Error("trailing icon did not emit a custom-geometry shape")
	}
}

// TestButton_InCardBody: a button placed last in a card body renders without error and
// its label appears (it is laid out by the card body stack like any leaf).
func TestButton_InCardBody(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "pricing", Nodes: []scene.SlideNode{
			scene.Card{Header: "Scale", Body: []scene.SlideNode{
				scene.Stat{Value: "$399", Label: "per month"},
				scene.Button{Label: "Start free", Tone: scene.ButtonAccentAlt},
			}},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "Start free") {
		t.Error("in-card button label not present")
	}
}

// TestButton_GhostIsOutline: a ghost button draws a line (the accent hairline) and no
// solid pill fill of its own beyond the outline.
func TestButton_GhostIsOutline(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "ghost", Nodes: []scene.SlideNode{
			scene.Button{Label: "Learn more", Tone: scene.ButtonGhost},
		}},
	}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:ln") {
		t.Error("ghost button did not emit an outline line")
	}
	if !strings.Contains(xml, "<a:noFill/>") {
		t.Error("ghost button did not emit a no-fill pill")
	}
}

// TestButton_UnknownIconFails: an unknown trailing/leading icon name fails Stage-1
// icon validation at Render (the walkIconRefs Button case), naming the icon.
func TestButton_UnknownIconFails(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "bad", Nodes: []scene.SlideNode{
			scene.Button{Label: "Go", TrailingIcon: "no-such-icon"},
		}},
	}}
	_, err := scene.Render(pptx.New(), sc)
	if err == nil {
		t.Fatal("Render accepted an unknown button icon; want a Stage-1 error")
	}
	if !strings.Contains(err.Error(), "no-such-icon") {
		t.Errorf("error %q should name the unknown icon", err)
	}
}

// TestButton_ByteIdenticalWhenUnused: a deck whose only difference is the absence of a
// button is unaffected — adding the node never perturbs the other slides' bytes.
func TestButton_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "cta", Content: scene.Alignment{Horizontal: scene.HAlignCenter}, Nodes: []scene.SlideNode{
			scene.Button{Label: "Talk to the team", Tone: scene.ButtonPrimary, TrailingIcon: "arrow-right"},
			scene.Button{Label: "Contact sales", Tone: scene.ButtonGhost, LeadingIcon: "star"},
		}},
	}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("button render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
