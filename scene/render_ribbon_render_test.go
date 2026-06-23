package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for Card.Ribbon (R12.3, D-098): a top bar with text, a corner
// star glyph, byte-identity when nil, position validation, and determinism.

func ribbonCard(rb *scene.Ribbon) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{
		{ID: "tier", Nodes: []scene.SlideNode{
			scene.Card{Header: "Scale", Eyebrow: "TIER", Ribbon: rb,
				Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("up to 100 agents")}}}},
		}},
	}}
}

// TestRibbon_TopBarText: a top-bar ribbon emits its label and an extra roundRect tab.
func TestRibbon_TopBarText(t *testing.T) {
	withBar, _ := render(t, ribbonCard(&scene.Ribbon{Text: "MOST POPULAR", Position: scene.RibbonTopBar}))
	none, _ := render(t, ribbonCard(nil))
	xb := zipPart(t, withBar, "ppt/slides/slide1.xml")
	xn := zipPart(t, none, "ppt/slides/slide1.xml")
	if !strings.Contains(xb, "MOST POPULAR") {
		t.Error("top-bar ribbon label missing")
	}
	if strings.Count(xb, `prst="roundRect"`) <= strings.Count(xn, `prst="roundRect"`) {
		t.Error("ribboned card should have more roundRects than a plain one (the badge)")
	}
}

// TestRibbon_CornerStar: a star-corner ribbon emits a custGeom glyph.
func TestRibbon_CornerStar(t *testing.T) {
	data, _ := render(t, ribbonCard(&scene.Ribbon{Position: scene.RibbonCornerStar}))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:custGeom>") {
		t.Error("corner-star ribbon did not emit a custGeom glyph")
	}
}

// TestRibbon_NilDistinct: a card with no ribbon has fewer shapes than a ribboned one, and
// a ribbon-free render is stable (byte-identical to itself) — the field is additive.
func TestRibbon_NilDistinct(t *testing.T) {
	a := renderBytes(t, ribbonCard(nil))
	b := renderBytes(t, ribbonCard(nil))
	if string(a) != string(b) {
		t.Fatal("ribbon-free card is not stable across renders")
	}
	withBar := renderBytes(t, ribbonCard(&scene.Ribbon{Text: "NEW", Position: scene.RibbonTopBar}))
	if string(withBar) == string(a) {
		t.Error("a ribboned card should differ from a ribbon-free one")
	}
}

// TestRibbon_PositionValidation: an out-of-range ribbon position fails Stage-1.
func TestRibbon_PositionValidation(t *testing.T) {
	sc := ribbonCard(&scene.Ribbon{Text: "x", Position: scene.RibbonPos(99)})
	if err := scene.ValidateScene(sc); err == nil {
		t.Fatal("out-of-range ribbon position passed validation")
	}
}

// TestRibbon_Deterministic: a ribboned card renders byte-identically across workers.
func TestRibbon_Deterministic(t *testing.T) {
	tr := scene.ColorAccentWarm
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "g", Nodes: []scene.SlideNode{scene.Grid{Columns: 2, Cells: []scene.SlideNode{
			scene.Card{Header: "Scale", Ribbon: &scene.Ribbon{Text: "MOST POPULAR", Position: scene.RibbonTopBar, Color: &tr},
				Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("a")}}}},
			scene.Card{Header: "Pro", Ribbon: &scene.Ribbon{Position: scene.RibbonCornerStar},
				Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}},
		}}}},
	}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("ribbon render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
