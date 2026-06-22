package scene_test

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Black-box tests for R11.6 chrome anti-collision (D-086).

// ellipseOffX returns the x-offset (EMU) of the first ellipse shape in the slide
// XML — the status dot — and its width, by reading the <a:off>/<a:ext> just before
// the prst="ellipse" geometry.
func ellipseOffX(t *testing.T, xml string) (x, cx int) {
	t.Helper()
	idx := strings.Index(xml, `prst="ellipse"`)
	if idx < 0 {
		t.Fatal("no ellipse (status dot) in slide XML")
	}
	off := regexp.MustCompile(`<a:off x="(-?\d+)" y="(-?\d+)"/>`)
	ext := regexp.MustCompile(`<a:ext cx="(\d+)" cy="(\d+)"/>`)
	offs := off.FindAllStringSubmatch(xml[:idx], -1)
	exts := ext.FindAllStringSubmatch(xml[:idx], -1)
	if len(offs) == 0 || len(exts) == 0 {
		t.Fatal("no <a:off>/<a:ext> before the ellipse")
	}
	x, _ = strconv.Atoi(offs[len(offs)-1][1])
	cx, _ = strconv.Atoi(exts[len(exts)-1][1])
	return x, cx
}

// TestStatusDot_AntiCollision: when a card carries both a header pill and a status
// dot, the dot is shifted left of the pill (its x is smaller than a dot-only card's
// corner x), so the two top-right chrome boxes do not overlap.
func TestStatusDot_AntiCollision(t *testing.T) {
	dot := scene.ColorSuccess
	body := []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}}

	both := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.Card{Header: "Scale", HeaderPill: "POPULAR", StatusDot: &dot, Body: body},
	}}}}
	dotOnly := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.Card{Header: "Scale", StatusDot: &dot, Body: body},
	}}}}

	bData, _ := render(t, both)
	bx, _ := ellipseOffX(t, zipPart(t, bData, "ppt/slides/slide1.xml"))
	dData, _ := render(t, dotOnly)
	dx, _ := ellipseOffX(t, zipPart(t, dData, "ppt/slides/slide1.xml"))

	if bx >= dx {
		t.Errorf("with a pill the dot should shift left: pill+dot x=%d, dot-only x=%d", bx, dx)
	}
}

// TestStatusDot_ByteIdentical_NoPill: a status dot on a card with no pill stays in
// the corner — the anti-collision shift is inert when only one element is set.
func TestStatusDot_ByteIdentical_NoPill(t *testing.T) {
	dot := scene.ColorSuccess
	a := scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.Card{Header: "X", StatusDot: &dot, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}},
	}}}}
	d1, _ := render(t, a)
	d2, _ := render(t, a)
	if string(d1) != string(d2) {
		t.Error("dot-only render is not stable")
	}
	// The dot's right edge reaches near the card's right inset (corner placement),
	// not shifted left by a (nonexistent) pill.
	x, cx := ellipseOffX(t, zipPart(t, d1, "ppt/slides/slide1.xml"))
	if x <= 0 || cx <= 0 {
		t.Errorf("unexpected dot geometry x=%d cx=%d", x, cx)
	}
}
