package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// pngResolver returns a resolver that serves a tiny valid PNG for any id, plus
// the bytes it serves.
func pngResolver() (scene.AssetResolver, []byte) {
	png := append([]byte("\x89PNG\r\n\x1a\n"), []byte("framed")...)
	return scene.URIAssetResolver(func(string) ([]byte, string, error) {
		return png, "image/png", nil
	}), png
}

// TestRenderImage_CuratedFrame is acceptance criterion 1: each curated frame
// (selected by the FrameKind enum) renders the image inside a bezel — the slide
// carries both the picture and the frame's native shapes, the asset is counted,
// and the alt text is emitted.
func TestRenderImage_CuratedFrame(t *testing.T) {
	resolver, png := pngResolver()
	for _, fr := range []struct {
		name string
		kind scene.FrameKind
	}{
		{"browser", scene.FrameBrowser},
		{"phone", scene.FramePhone},
		{"desktop", scene.FrameDesktop},
		{"laptop", scene.FrameLaptop},
	} {
		t.Run(fr.name, func(t *testing.T) {
			sc := scene.Scene{Slides: []scene.SceneSlide{{
				ID:    "img",
				Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x", Alt: "a chart", Frame: fr.kind}},
			}}}
			data, stats := render(t, sc, scene.WithAssetResolver(resolver))
			if stats.Assets != 1 {
				t.Fatalf("Stats.Assets = %d, want 1", stats.Assets)
			}
			// Bezel shapes + the image: more than one shape total.
			if stats.Shapes < 2 {
				t.Fatalf("Stats.Shapes = %d, want >= 2 (bezel + image)", stats.Shapes)
			}
			slide := zipPart(t, data, "ppt/slides/slide1.xml")
			if !strings.Contains(slide, "<p:pic>") {
				t.Errorf("%s: slide missing the image pic:\n%s", fr.name, slide)
			}
			if !strings.Contains(slide, "<p:sp>") {
				t.Errorf("%s: slide missing the bezel shapes:\n%s", fr.name, slide)
			}
			if !strings.Contains(slide, "a chart") {
				t.Errorf("%s: alt text not emitted:\n%s", fr.name, slide)
			}
			if zipPart(t, data, "ppt/media/image1.png") != string(png) {
				t.Errorf("%s: image bytes not embedded verbatim", fr.name)
			}
		})
	}
}

// TestRenderImage_FrameExtension is acceptance criterion 2: a caller frame
// registered via WithFrameExtension renders an image through its recipe, and
// FrameName takes precedence over the FrameKind enum (D-038).
func TestRenderImage_FrameExtension(t *testing.T) {
	resolver, _ := pngResolver()
	// A custom recipe that draws one rect and insets the interior.
	custom := func(sl *pptx.Slide, region pptx.Box) (pptx.Box, int) {
		sl.AddShape(pptx.ShapeRect, region, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
		in := region.Inset(pptx.UniformInset(pptx.In(0.1)))
		return in, 1
	}

	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "img",
		// Frame enum says Browser, but FrameName must win (precedence).
		Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x", Frame: scene.FrameBrowser, FrameName: "retro"}},
	}}}

	data, stats := render(t, sc,
		scene.WithAssetResolver(resolver),
		scene.WithFrameExtension("retro", custom))

	if len(stats.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %+v", stats.Warnings)
	}
	// Exactly one bezel shape (the custom rect) + the image.
	if stats.Shapes != 2 {
		t.Fatalf("Stats.Shapes = %d, want 2 (1 custom bezel + image)", stats.Shapes)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<p:pic>") || !strings.Contains(slide, "<p:sp>") {
		t.Errorf("extension frame did not render bezel + image:\n%s", slide)
	}
}

// TestRenderImage_FrameNone is acceptance criterion 3: an unframed image (no
// FrameKind, no FrameName) renders just the picture — no bezel shapes.
func TestRenderImage_FrameNone(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "img",
		Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x"}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(resolver))
	if stats.Shapes != 1 {
		t.Fatalf("Stats.Shapes = %d, want 1 (just the image, no bezel)", stats.Shapes)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<p:pic>") {
		t.Errorf("unframed image missing pic:\n%s", slide)
	}
	if strings.Contains(slide, "<p:sp>") {
		t.Errorf("unframed image unexpectedly drew a bezel shape:\n%s", slide)
	}
}

// TestRenderImage_UnknownFrame is acceptance criterion 4: an Image whose
// resolved FrameName is neither curated nor registered fails Stage-1 validation
// (closed-name, §14.4) — Render returns an error naming the frame.
func TestRenderImage_UnknownFrame(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "img",
		Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x", FrameName: "nope"}},
	}}}
	pres := pptx.New()
	_, err := scene.Render(pres, sc, scene.WithAssetResolver(resolver))
	if err == nil {
		t.Fatal("Render accepted an unknown frame name; want a validation error")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error %q does not name the unknown frame", err)
	}
}

// TestRenderImage_UnknownFrameNested checks the frame-name validation recurses
// into container children.
func TestRenderImage_UnknownFrameNested(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "grid",
		Nodes: []scene.SlideNode{scene.Grid{
			Columns: 2,
			Cells: []scene.SlideNode{
				scene.Prose{Paragraphs: []scene.RichText{rt("ok")}},
				scene.Image{AssetID: "asset://x", FrameName: "ghost"},
			},
		}},
	}}}
	_, err := scene.Render(pptx.New(), sc, scene.WithAssetResolver(resolver))
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("nested unknown frame not rejected; err = %v", err)
	}
}
