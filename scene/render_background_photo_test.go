package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// photoResolver resolves the single background asset "photo" to a valid PNG.
func photoResolver() scene.AssetResolver {
	return scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "photo" {
			return pngOf(1920, 1080), "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
}

// TestBackground_ScrimSolid is R14.1 acceptance (legibility scrim): a solid scrim
// over a color background draws a second full-slide rect with the scrim's alpha.
func TestBackground_ScrimSolid(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "scrim-solid",
		Background: scene.Background{
			Kind:  scene.BackgroundColor,
			Color: pptx.ColorAccent,
			Scrim: &scene.Scrim{Color: pptx.ColorCanvas, Opacity: 45000},
		},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("scrim solid: unexpected warnings: %+v", stats.Warnings)
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(xml, "<p:sp>"); n < 2 {
		t.Errorf("scrim solid: want >=2 shapes (fill + scrim), got %d:\n%s", n, xml)
	}
	if !strings.Contains(xml, `<a:alpha val="45000"`) {
		t.Errorf("scrim solid: missing scrim alpha 45000:\n%s", xml)
	}
}

// TestBackground_ScrimGradient verifies a gradient scrim emits a linear-gradient
// overlay running transparent → color at the scrim opacity.
func TestBackground_ScrimGradient(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "scrim-grad",
		Background: scene.Background{
			Kind:  scene.BackgroundColor,
			Color: pptx.ColorSurface,
			Scrim: &scene.Scrim{Color: pptx.ColorCanvas, Opacity: 60000, Gradient: true},
		},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("scrim gradient: unexpected warnings: %+v", stats.Warnings)
	}
	pres, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	shapes := pres.Slides()[0].Shapes()
	if len(shapes) < 2 {
		t.Fatalf("scrim gradient: want >=2 shapes, got %d", len(shapes))
	}
	fill := shapes[1].Fill() // the scrim is drawn after the base fill
	if fill == nil || fill.Kind() != pptx.FillGradient {
		t.Errorf("scrim gradient: overlay fill kind = %v, want FillGradient", fill)
	}
}

// TestBackground_ScrimDuotonePhoto is R14.1 acceptance (photo class): a
// full-bleed photo background with a duotone tint and a gradient scrim emits the
// pic, an <a:duotone> recolor, and the scrim overlay, with no warnings.
func TestBackground_ScrimDuotonePhoto(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "photo",
		Background: scene.Background{
			Kind:    scene.BackgroundAsset,
			AssetID: "asset://photo",
			Duotone: &scene.Duotone{Shadow: pptx.ColorAccent, Highlight: pptx.ColorCanvas},
			Scrim:   &scene.Scrim{Color: pptx.ColorSurface, Opacity: 50000, Gradient: true},
		},
		Nodes: []scene.SlideNode{scene.Hero{Title: "Over the photo"}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(photoResolver()))
	if len(stats.Warnings) != 0 {
		t.Errorf("photo class: unexpected warnings: %+v", stats.Warnings)
	}
	if stats.Assets == 0 {
		t.Errorf("photo class: expected an asset to be counted")
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:duotone>") {
		t.Errorf("photo class: missing <a:duotone> recolor:\n%s", xml)
	}
	if !strings.Contains(xml, "<p:pic>") {
		t.Errorf("photo class: missing background pic")
	}
}

// TestBackground_ScrimNilByteIdentical verifies that a nil Scrim + nil Duotone is
// byte-identical to the same background without the new fields (R14.1: zero
// renders exactly as today).
func TestBackground_ScrimNilByteIdentical(t *testing.T) {
	mk := func(scrim *scene.Scrim) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID:         "bg",
			Background: scene.Background{Kind: scene.BackgroundColor, Color: pptx.ColorAccent, Scrim: scrim},
		}}}
	}
	with, _ := render(t, mk(nil))
	plain, _ := render(t, scene.Scene{Slides: []scene.SceneSlide{{
		ID:         "bg",
		Background: scene.Background{Kind: scene.BackgroundColor, Color: pptx.ColorAccent},
	}}})
	if !bytes.Equal(with, plain) {
		t.Errorf("nil Scrim not byte-identical to absent field (%d vs %d bytes)", len(with), len(plain))
	}
}

// TestBackground_ScrimDeterministic guards that the scrim + duotone photo path is
// worker-count independent (byte-identical at 1 vs 8 workers).
func TestBackground_ScrimDeterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "photo",
		Background: scene.Background{
			Kind:    scene.BackgroundAsset,
			AssetID: "asset://photo",
			Duotone: &scene.Duotone{Shadow: pptx.ColorAccent, Highlight: pptx.ColorCanvas},
			Scrim:   &scene.Scrim{Color: pptx.ColorSurface, Opacity: 50000, Gradient: true, GradientAngle: 90},
		},
		Nodes: []scene.SlideNode{scene.Hero{Title: "Over the photo"}},
	}}}
	seq := renderBytes(t, sc, scene.WithAssetResolver(photoResolver()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithAssetResolver(photoResolver()), scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("scrim+duotone photo not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
