package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestRenderCard is PR#2 acceptance criterion 4: a Card renders native chrome
// (background rounded-rect + accent stripe + header) plus its body, stays
// conformant, and emits no pic for a plain (text) card.
func TestRenderCard(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "card",
		Nodes: []scene.SlideNode{scene.Card{
			Header: "Revenue",
			Body:   []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("Up 24% YoY.")}}},
		}},
	}}}

	data, stats := render(t, sc)
	if stats.Shapes < 3 { // background + stripe + header + body
		t.Errorf("card rendered %d shapes, want >= 3", stats.Shapes)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("card deck failed conformance:\n%s", rep)
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:prstGeom prst=\"roundRect\"") {
		t.Errorf("card missing rounded-rect chrome:\n%s", xml)
	}
	if !strings.Contains(xml, "<a:t>Revenue</a:t>") || !strings.Contains(xml, "<a:t>Up 24% YoY.</a:t>") {
		t.Errorf("card missing header/body text:\n%s", xml)
	}
	if strings.Contains(xml, "<p:pic>") {
		t.Errorf("plain card unexpectedly contains a pic shape")
	}
}

// TestCardKnobs is PR#2 acceptance criterion 5: each card knob renders.
// Elevation emits an outerShdw; an accent border emits a themed line; eyebrow,
// header-pill, and a horizontal body all render.
func TestCardKnobs(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "knobs",
		Nodes: []scene.SlideNode{scene.Card{
			Eyebrow:     "Q3 · FY25",
			Header:      "Elevated card",
			HeaderPill:  "NEW",
			Fill:        scene.ColorSurfaceAlt,
			BorderStyle: scene.BorderAccent,
			Size:        scene.CardSizeLG,
			Elevation:   scene.ElevationElevated,
			BodyLayout:  scene.BodyHorizontal,
			Body: []scene.SlideNode{
				scene.Prose{Paragraphs: []scene.RichText{rt("left")}},
				scene.Prose{Paragraphs: []scene.RichText{rt("right")}},
			},
		}},
	}}}

	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("unexpected warnings: %+v", stats.Warnings)
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{
		"<a:outerShdw",                        // elevation
		"<a:t>Q3 · FY25</a:t>",                // eyebrow
		"<a:t>NEW</a:t>",                      // header pill
		"<a:t>left</a:t>", "<a:t>right</a:t>", // horizontal body
	} {
		if !strings.Contains(xml, want) {
			t.Errorf("card knobs missing %q in:\n%s", want, xml)
		}
	}
}

// TestCardIcon is PR#2 acceptance criterion 6: a Card with a curated icon places
// a native custGeom shape; an unknown icon name fails Stage-1 before compose.
func TestCardIcon(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "icon",
		Nodes: []scene.SlideNode{scene.Card{
			Icon:   "star",
			Header: "Featured",
			Body:   []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}},
		}},
	}}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:custGeom>") {
		t.Errorf("card icon did not render a custGeom shape:\n%s", xml)
	}

	// Unknown icon name → Stage-1 error, before any slide composes.
	bad := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bad",
		Nodes: []scene.SlideNode{scene.Card{Icon: "definitely-not-an-icon", Header: "x"}},
	}}}
	if _, err := scene.Render(pptx.New(), bad); err == nil {
		t.Error("unknown card icon name accepted; expected a Stage-1 render error")
	}
}

// TestRenderCardSection is PR#2 acceptance criterion 7: a card_section of a grid
// of cards (card-of-cards) renders the section chrome plus each nested card.
func TestRenderCardSection(t *testing.T) {
	card := func(h string) scene.Card {
		return scene.Card{Header: h, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt(h + " body")}}}}
	}
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "section",
		Nodes: []scene.SlideNode{scene.CardSection{
			Header: "Pillars",
			Body: []scene.SlideNode{scene.Grid{
				Columns: 2,
				Cells:   []scene.SlideNode{card("A"), card("B")},
			}},
		}},
	}}}

	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("unexpected warnings: %+v", stats.Warnings)
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	// Section chrome + two nested card chromes: at least three roundRects.
	if n := strings.Count(xml, "prst=\"roundRect\""); n < 3 {
		t.Errorf("card-of-cards rendered %d roundRects, want >= 3:\n%s", n, xml)
	}
	for _, want := range []string{"<a:t>Pillars</a:t>", "<a:t>A</a:t>", "<a:t>B</a:t>"} {
		if !strings.Contains(xml, want) {
			t.Errorf("card_section missing %q", want)
		}
	}
}

// TestCardSectionMixedCode is PR#2 acceptance criterion 8: a card_section
// containing a code_block renders native chrome + one pic (mixed policy,
// RFC §12.2).
func TestCardSectionMixedCode(t *testing.T) {
	png := append([]byte("\x89PNG\r\n\x1a\n"), []byte("code")...)
	resolver := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "c1" {
			return png, "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "mixed",
		Nodes: []scene.SlideNode{scene.CardSection{
			Header: "Snippet",
			Body:   []scene.SlideNode{scene.CodeBlock{AssetID: "asset://c1", Language: "go"}},
		}},
	}}}

	data, _ := render(t, sc, scene.WithAssetResolver(resolver))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "prst=\"roundRect\"") {
		t.Errorf("card_section missing native chrome:\n%s", xml)
	}
	if !strings.Contains(xml, "<p:pic>") {
		t.Errorf("card_section code_block did not render a pic:\n%s", xml)
	}
}

// TestCardParallel is PR#2 acceptance criterion 9: a card/icon scene renders
// byte-identically at workers=1 and workers=N (idempotency under parallelism,
// D-035/D-015).
func TestCardParallel(t *testing.T) {
	mk := func() scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{
			{ID: "a", Nodes: []scene.SlideNode{scene.Card{Icon: "check", Header: "One", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("a")}}}}}},
			{ID: "b", Nodes: []scene.SlideNode{scene.Card{Header: "Two", Elevation: scene.ElevationRaised, Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}}}},
			{ID: "c", Nodes: []scene.SlideNode{scene.CardSection{Header: "Sec", Body: []scene.SlideNode{scene.Grid{Columns: 2, Cells: []scene.SlideNode{scene.Card{Header: "x"}, scene.Card{Header: "y"}}}}}}},
		}}
	}
	seq, _ := render(t, mk(), scene.WithWorkers(1))
	par, _ := render(t, mk(), scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Error("card scene render differs between workers=1 and workers=4 (idempotency broken)")
	}
}

// TestCardAccentBorderDropsStripe guards the accent-border fix: a BorderAccent
// card omits the redundant left accent stripe (the border is the accent) and
// emits the accent border line, while other border styles keep their stripe.
func TestCardAccentBorderDropsStripe(t *testing.T) {
	mk := func(b scene.BorderStyle) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID: "b",
			Nodes: []scene.SlideNode{scene.Card{
				Header:      "H",
				BorderStyle: b,
				Body:        []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("x")}}},
			}},
		}}}
	}
	_, solid := render(t, mk(scene.BorderSolid))
	_, accent := render(t, mk(scene.BorderAccent))
	if accent.Shapes != solid.Shapes-1 {
		t.Errorf("BorderAccent shapes=%d, want one fewer than BorderSolid=%d (the stripe is dropped)", accent.Shapes, solid.Shapes)
	}
	data, _ := render(t, mk(scene.BorderAccent))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, `<a:ln w="19050"`) { // 1.5pt accent border
		t.Errorf("BorderAccent card missing the accent border line:\n%s", xml)
	}
}
