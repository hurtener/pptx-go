package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// cardPhotoResolver resolves "card-photo" to a valid PNG.
func cardPhotoResolver() scene.AssetResolver {
	return scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "card-photo" {
			return pngOf(1200, 800), "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
}

// TestCardImageFill is R14.1 part 2 (D-117): a Card with ImageFill resolves a
// photo and fills its surface with a cover-fit <a:blipFill> (replacing the solid
// fill), with no warnings and the asset counted.
func TestCardImageFill(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "card-fill",
		Nodes: []scene.SlideNode{scene.Card{
			Header:    "Photo card",
			ImageFill: "asset://card-photo",
			Body:      []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("Over the photo")}}},
		}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(cardPhotoResolver()))
	if len(stats.Warnings) != 0 {
		t.Errorf("card image fill: unexpected warnings: %+v", stats.Warnings)
	}
	if stats.Assets == 0 {
		t.Errorf("card image fill: expected an asset to be counted")
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:blipFill>") {
		t.Errorf("card image fill: missing <a:blipFill>:\n%s", xml)
	}
	if strings.Contains(xml, "<p:blipFill>") {
		t.Errorf("card image fill wrongly emitted <p:blipFill> (namespace)")
	}
}

// TestCardImageFill_MissingWarns verifies an unresolvable ImageFill warns and
// falls back to the solid fill (no panic, RFC §10.2).
func TestCardImageFill_MissingWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "card-fill-missing",
		Nodes: []scene.SlideNode{scene.Card{Header: "X", ImageFill: "asset://nope"}},
	}}}
	_, stats := render(t, sc, scene.WithAssetResolver(cardPhotoResolver()))
	if len(stats.Warnings) == 0 {
		t.Errorf("card image fill missing: expected a warning")
	}
}

// TestCardImageFill_EmptyByteIdentical verifies a Card with ImageFill == "" is
// byte-identical to a Card without the field (R14.1: zero renders as today).
func TestCardImageFill_EmptyByteIdentical(t *testing.T) {
	mk := func(id scene.AssetID) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID:    "card",
			Nodes: []scene.SlideNode{scene.Card{Header: "Plain", ImageFill: id, Fill: pptx.ColorSurface}},
		}}}
	}
	with, _ := render(t, mk(""), scene.WithAssetResolver(cardPhotoResolver()))
	plain, _ := render(t, scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "card",
		Nodes: []scene.SlideNode{scene.Card{Header: "Plain", Fill: pptx.ColorSurface}},
	}}})
	if !bytes.Equal(with, plain) {
		t.Errorf("empty ImageFill not byte-identical to absent field (%d vs %d bytes)", len(with), len(plain))
	}
}

// TestCardImageFill_Deterministic guards that the card image-fill path is
// worker-count independent.
func TestCardImageFill_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "card-fill",
		Nodes: []scene.SlideNode{scene.Card{
			Header:    "Photo card",
			ImageFill: "asset://card-photo",
			Body:      []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("Over the photo")}}},
		}},
	}}}
	seq := renderBytes(t, sc, scene.WithAssetResolver(cardPhotoResolver()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithAssetResolver(cardPhotoResolver()), scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("card image fill not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
