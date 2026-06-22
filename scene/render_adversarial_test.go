package scene_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// R11.12 adversarial content-fit fixtures (D-092). A reusable torture suite that
// renders every component under hostile content — multi-line headers, over-long
// pill / badge / row labels, oversized stat values, many list items, dark variant +
// dark fills — and asserts the structural invariant that no emitted shape/text box
// ever falls off the slide canvas. Catches the whole class of fixed-size regressions
// (any component that assumes a fixed width/height and draws off-slide) for any
// future component, in one place.

const longText = "A deliberately long stretch of content that wraps across several lines under any reasonable column width"

// adversarialScene builds a deck exercising every component with hostile content,
// in both the light and dark variants.
func adversarialScene() scene.Scene {
	accent := scene.ColorAccent
	dotColor := scene.ColorSuccess
	body := func() []scene.SlideNode {
		return []scene.SlideNode{scene.List{Items: []scene.ListItem{
			{Text: rt(longText)}, {Text: rt(longText)}, {Text: rt(longText)}, {Text: rt(longText)},
		}, Indent: scene.IndentTight}}
	}
	card := func(variantDark bool) scene.Card {
		fill := scene.ColorSurface
		if variantDark {
			fill = scene.ColorAccent
		}
		return scene.Card{
			Eyebrow: "VISION", Header: longText, HeaderPill: "FULLY CUSTOMIZABLE",
			HeaderFill: &accent, StatusDot: &dotColor, Watermark: "01", Fill: fill,
			Body: body(),
		}
	}

	var slides []scene.SceneSlide
	for _, variant := range []scene.Variant{scene.VariantLight, scene.VariantDark} {
		dark := variant == scene.VariantDark
		slides = append(slides,
			// Cards (every size × layout) under a long wrapping header.
			scene.SceneSlide{ID: fmt.Sprintf("cards-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Grid{Columns: 3, Cells: []scene.SlideNode{card(dark), card(dark), card(dark)}},
			}},
			// A dense bento with long row labels + an over-wide stat.
			scene.SceneSlide{ID: fmt.Sprintf("bento-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Bento{Columns: 2, WeightedRows: true, Rows: []scene.BentoRow{
					{Label: "Control plane orchestration", Cells: []scene.BentoCell{{Span: 1, Node: scene.Stat{Value: "$4,000,000+", Label: "per month", AutoFit: true}}, {Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}}},
					{Label: "The core memory layer", Cells: []scene.BentoCell{{Span: 2, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}}},
					{Label: "Run", Cells: []scene.BentoCell{{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}, {Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}}},
				}},
			}},
			// TwoColumn with a long join badge label.
			scene.SceneSlide{ID: fmt.Sprintf("cols-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.TwoColumn{Join: scene.JoinBadge, JoinLabel: "One agent for everything",
					Left:  []scene.SlideNode{scene.Heading{Text: rt(longText), Level: 1, AutoFit: true}, scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}},
					Right: []scene.SlideNode{scene.Stat{Value: "$4,000+", Label: "up to 100 agents", AutoFit: true}, scene.Prose{Paragraphs: []scene.RichText{rt(longText)}}}},
			}},
			// A tall stat strip + hero that would overflow without the guards.
			scene.SceneSlide{ID: fmt.Sprintf("strip-%v", variant), Variant: variant, Nodes: []scene.SlideNode{
				scene.Hero{Title: longText, AutoFit: true},
				scene.Grid{Columns: 4, Cells: []scene.SlideNode{
					scene.Stat{Value: "$4,000,000+", Label: "a", AutoFit: true}, scene.Stat{Value: "99.999%", Label: "b", AutoFit: true},
					scene.Stat{Value: "1,234,567", Label: "c", AutoFit: true}, scene.Stat{Value: "$12.5M", Label: "d", AutoFit: true},
				}},
			}},
		)
	}
	return scene.Scene{Slides: slides}
}

var offRe = regexp.MustCompile(`<a:off x="(-?\d+)" y="(-?\d+)"/>`)
var extRe = regexp.MustCompile(`<a:ext cx="(\d+)" cy="(\d+)"/>`)

// TestAdversarial_AllBoxesOnCanvas is the R11.12 invariant: across every slide of
// the hostile fixture, every emitted shape/text box lies fully within the slide
// canvas (no off-slide / clipped content). A component that reintroduces a
// fixed-size assumption and draws off-slide fails here.
func TestAdversarial_AllBoxesOnCanvas(t *testing.T) {
	sc := adversarialScene()
	data, _ := render(t, sc)
	slideW, slideH := pptx.New().SlideSize()

	for i := range sc.Slides {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		xml := zipPart(t, data, part)
		offs := offRe.FindAllStringSubmatch(xml, -1)
		exts := extRe.FindAllStringSubmatch(xml, -1)
		if len(offs) == 0 {
			t.Errorf("%s (slide %q): no shapes emitted", part, sc.Slides[i].ID)
			continue
		}
		// Pair each <a:off> with the <a:ext> that follows it (xfrm order).
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
				t.Errorf("%s (slide %q): box [x=%d y=%d cx=%d cy=%d] exceeds the %dx%d canvas",
					part, sc.Slides[i].ID, x, y, cx, cy, slideW, slideH)
			}
		}
	}
}

// TestAdversarial_Renders is a smoke guard that the hostile fixture renders without
// error and produces every slide (so the on-canvas suite has parts to check).
func TestAdversarial_Renders(t *testing.T) {
	sc := adversarialScene()
	data, stats := render(t, sc)
	if stats.Slides != len(sc.Slides) {
		t.Errorf("rendered %d slides, want %d", stats.Slides, len(sc.Slides))
	}
	if len(data) == 0 {
		t.Error("empty render")
	}
}

// TestAdversarial_Deterministic: the hostile fixture (every component + the
// safe-area clamp on its overflowing content) renders byte-identically across worker
// counts — the clamp is pure integer math, so it never depends on scheduling.
func TestAdversarial_Deterministic(t *testing.T) {
	sc := adversarialScene()
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("adversarial: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
