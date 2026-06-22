package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// R11.11 verify-and-close (D-091): a card watermark is drawn behind the body content
// (z-order first) at low opacity, so it never reduces body legibility — already
// implemented by D-054. These guards pin the z-order and the low alpha.

func slideXML(t *testing.T, sc scene.Scene) string {
	t.Helper()
	data, _ := render(t, sc)
	return zipPart(t, data, "ppt/slides/slide1.xml")
}

// TestWatermark_BehindBody: the watermark text frame is emitted before the body
// content in the slide XML, so PowerPoint paints it behind (z-order). A
// content-dense card therefore keeps its body legible over the ghosted watermark.
func TestWatermark_BehindBody(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.Card{
			Header:    "Plan",
			Watermark: "WMARK",
			Body:      []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("BODYTEXT one"), rt("BODYTEXT two"), rt("BODYTEXT three")}}},
		},
	}}}}
	xml := slideXML(t, sc)

	wm := strings.Index(xml, "<a:t>WMARK</a:t>")
	body := strings.Index(xml, "<a:t>BODYTEXT one</a:t>")
	if wm < 0 || body < 0 {
		t.Fatalf("missing watermark (%d) or body (%d) text", wm, body)
	}
	if wm > body {
		t.Errorf("watermark (at %d) should be emitted before the body (at %d) so it sits behind it", wm, body)
	}
}

// TestWatermark_LowAlpha: the watermark run carries a low-opacity alpha (~13%), so
// even where it overlaps body content it does not impair legibility.
func TestWatermark_LowAlpha(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.Card{Header: "Plan", Watermark: "WMARK", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}},
	}}}}
	if xml := slideXML(t, sc); !strings.Contains(xml, `<a:alpha val="13000"/>`) {
		t.Errorf("watermark should emit a low ~13%% alpha; XML had no <a:alpha val=\"13000\"/>")
	}
}

// TestWatermark_OmittedWhenUnset: a card without a watermark emits no alpha run —
// the guard is inert when unused.
func TestWatermark_OmittedWhenUnset(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.Card{Header: "Plan", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}},
	}}}}
	if xml := slideXML(t, sc); strings.Contains(xml, `<a:alpha`) {
		t.Error("no watermark set: should emit no <a:alpha> run")
	}
}
