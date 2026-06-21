package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Rich card visuals (Phase 25, R4): a colored header band, a top-right status
// dot, and a ghosted watermark — each opt-in, omitted when unset, byte-identical
// when none are set.

func richCardSlide(c scene.Card) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{c}}}}
}

func countSubstr(s, sub string) int { return strings.Count(s, sub) }

// TestCardRichVisuals_AllThree is acceptance criterion 1: a card with all three
// set emits a header-band rounded rect, a status-dot ellipse, and a low-opacity
// watermark run carrying the label.
func TestCardRichVisuals_AllThree(t *testing.T) {
	hf := scene.ColorAccent
	sd := scene.ColorSuccess
	card := scene.Card{
		Header:     "Revenue",
		Body:       []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}},
		HeaderFill: &hf,
		StatusDot:  &sd,
		Watermark:  "01",
	}
	bare := scene.Card{Header: "Revenue", Body: card.Body}

	data, _ := render(t, richCardSlide(card))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	bareData, _ := render(t, richCardSlide(bare))
	bareXML := zipPart(t, bareData, "ppt/slides/slide1.xml")

	// Header band: exactly one more roundRect than the bare card (bg + band).
	if got, want := countSubstr(xml, `prst="roundRect"`), countSubstr(bareXML, `prst="roundRect"`)+1; got != want {
		t.Errorf("header band: roundRect count = %d, want %d (one more than bare)", got, want)
	}
	if !strings.Contains(xml, `prst="ellipse"`) {
		t.Error("status dot: expected an ellipse shape")
	}
	if !strings.Contains(xml, "<a:alpha") {
		t.Error("watermark: expected a low-opacity (<a:alpha>) run")
	}
	if !strings.Contains(xml, "<a:t>01</a:t>") {
		t.Error("watermark: expected the label text '01'")
	}
}

// TestCardRichVisuals_OmittedWhenUnset is acceptance criterion 2: a card with no
// rich fields emits none of the three (no extra band, no ellipse, no watermark).
func TestCardRichVisuals_OmittedWhenUnset(t *testing.T) {
	card := scene.Card{
		Header: "Plain",
		Body:   []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}},
	}
	data, _ := render(t, richCardSlide(card))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if strings.Contains(xml, `prst="ellipse"`) {
		t.Error("no StatusDot set: should emit no ellipse")
	}
	if strings.Contains(xml, "<a:alpha") {
		t.Error("no Watermark / no elevation: should emit no <a:alpha>")
	}
	if strings.Contains(xml, "<a:t>01</a:t>") {
		t.Error("no Watermark set: should emit no watermark text")
	}
}

// TestCardRichVisuals_IndividuallyOptional checks each field renders its element
// independently (criterion 2, the dual of all-three).
func TestCardRichVisuals_IndividuallyOptional(t *testing.T) {
	body := []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}
	sd := scene.ColorSuccess

	dotOnly, _ := render(t, richCardSlide(scene.Card{Header: "H", Body: body, StatusDot: &sd}))
	x := zipPart(t, dotOnly, "ppt/slides/slide1.xml")
	if !strings.Contains(x, `prst="ellipse"`) {
		t.Error("StatusDot only: expected ellipse")
	}
	if strings.Contains(x, "<a:t>WM</a:t>") {
		t.Error("StatusDot only: unexpected watermark text")
	}

	wmOnly, _ := render(t, richCardSlide(scene.Card{Header: "H", Body: body, Watermark: "WM"}))
	x = zipPart(t, wmOnly, "ppt/slides/slide1.xml")
	if !strings.Contains(x, "<a:t>WM</a:t>") || !strings.Contains(x, "<a:alpha") {
		t.Error("Watermark only: expected a faint 'WM' run")
	}
	if strings.Contains(x, `prst="ellipse"`) {
		t.Error("Watermark only: unexpected ellipse")
	}
}

// TestCardRichVisuals_Deterministic is acceptance criterion 4: a deck of rich
// cards renders byte-identically across worker counts.
func TestCardRichVisuals_Deterministic(t *testing.T) {
	hf := scene.ColorAccent
	sd := scene.ColorSuccess
	sc := scene.Scene{}
	for i := 0; i < 12; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + i)),
			Nodes: []scene.SlideNode{scene.Card{
				Header:     "Card",
				Body:       []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("body")}}},
				HeaderFill: &hf, StatusDot: &sd, Watermark: "0" + string(rune('1'+i%9)),
			}},
		})
	}
	seq, _ := render(t, sc, scene.WithWorkers(1))
	par, _ := render(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("rich-card deck: parallel render differs from sequential (%d vs %d bytes)", len(par), len(seq))
	}
}
