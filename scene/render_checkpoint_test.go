package scene_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Wave-11 §17 checkpoint backfill tests (D-093): coverage of shipped D-082/D-086/
// D-092 behavior the per-phase tests missed — the pill and join-badge auto-contrast
// colors, the narrow-card status-dot clamp, nested-container overflow warnings, and
// CardSection header auto-contrast.

// M4 — the header pill run auto-contrasts: on a dark-variant slide the pill label
// (a separate shape from the header title) is light, not black-on-dark.
func TestCardPill_AutoContrast_DarkVariant(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:      "s1",
		Variant: scene.VariantDark,
		Nodes:   []scene.SlideNode{scene.Card{Header: "Plan", HeaderPill: "POPULAR", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
	}}}
	xml := slideXML(t, sc)
	run := headerRun(t, xml, "POPULAR")
	if !strings.Contains(run, `val="FFFFFF"`) {
		t.Errorf("dark-variant pill label should be white; run = %s", run)
	}
}

// M4 — the join-badge label is light on the default (dark) accent fill (the prior
// hardcoded TextInverse, now via onCardSurface — byte-identical here).
func TestJoinBadge_AutoContrast_DefaultAccent(t *testing.T) {
	xml := slideXML(t, joinScene("vs"))
	run := headerRun(t, xml, "vs")
	if !strings.Contains(run, `val="FFFFFF"`) {
		t.Errorf("join-badge label on the default accent should be white; run = %s", run)
	}
}

// L1 — the status-dot anti-collision clamp keeps the dot on the card in a narrow
// card with both a pill and a dot (exercises the dotX < innerX clamp, D-086).
func TestStatusDot_AntiCollision_NarrowCard(t *testing.T) {
	dot := scene.ColorSuccess
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "s1",
		// A single narrow card via a 4-column grid cell.
		Nodes: []scene.SlideNode{scene.Grid{Columns: 4, Cells: []scene.SlideNode{
			scene.Card{Header: "Scale", HeaderPill: "POPULAR", StatusDot: &dot, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}},
			scene.Card{Header: "x"}, scene.Card{Header: "y"}, scene.Card{Header: "z"},
		}}},
	}}}
	xml := slideXML(t, sc)
	x, cx := ellipseOffX(t, xml)
	// The dot stays on its card (x within the slide body, not off the left edge) and
	// fully on-canvas.
	if x < int(pptx.In(0.5)) {
		t.Errorf("narrow-card dot x=%d should be at/after the body margin (clamped on-card)", x)
	}
	if slideW, _ := pptx.New().SlideSize(); x+cx > slideW {
		t.Errorf("narrow-card dot right edge %d exceeds the canvas %d", x+cx, slideW)
	}
}

// L2 — a nested overflowing container (Bento of Cards taller than the slide) warns
// via the safe-area clamp and emits no off-canvas box.
func TestNestedOverflow_WarnsAndOnCanvas(t *testing.T) {
	var rows []scene.BentoRow
	for i := 0; i < 8; i++ {
		rows = append(rows, scene.BentoRow{Cells: []scene.BentoCell{
			{Span: 1, Node: scene.Card{Header: "left", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
			{Span: 1, Node: scene.Card{Header: "right", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}}},
		}})
	}
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{scene.Bento{Columns: 2, Rows: rows}}}}}
	data, stats := render(t, sc)

	var warned bool
	for _, w := range stats.Warnings {
		if strings.Contains(w.Message, "exceeds the slide safe area") {
			warned = true
		}
	}
	if !warned {
		t.Errorf("nested overflow should warn; got %+v", stats.Warnings)
	}

	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	slideW, slideH := pptx.New().SlideSize()
	offs := offRe.FindAllStringSubmatch(xml, -1)
	exts := extRe.FindAllStringSubmatch(xml, -1)
	n := len(offs)
	if len(exts) < n {
		n = len(exts)
	}
	for k := 0; k < n; k++ {
		x, _ := strconv.Atoi(offs[k][1])
		y, _ := strconv.Atoi(offs[k][2])
		cx, _ := strconv.Atoi(exts[k][1])
		cy, _ := strconv.Atoi(exts[k][2])
		if x < 0 || y < 0 || x+cx > slideW || y+cy > slideH {
			t.Errorf("nested overflow: box [x=%d y=%d cx=%d cy=%d] off-canvas", x, y, cx, cy)
		}
	}
}

// L3 — CardSection shares renderCardChrome, so its header inherits D-082 auto-
// contrast: on a dark-variant slide the section header is light.
func TestCardSectionHeader_AutoContrast_DarkVariant(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:      "s1",
		Variant: scene.VariantDark,
		Nodes: []scene.SlideNode{scene.CardSection{Header: "SECTION", Body: []scene.SlideNode{
			scene.Grid{Columns: 2, Cells: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("a")}}, scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}},
		}}},
	}}}
	xml := slideXML(t, sc)
	run := headerRun(t, xml, "SECTION")
	if !strings.Contains(run, `val="FFFFFF"`) {
		t.Errorf("dark-variant CardSection header should be white; run = %s", run)
	}
}
