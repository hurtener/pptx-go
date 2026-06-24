package scene_test

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// zipNames returns the set of part names in a saved deck.
func zipNames(t *testing.T, data []byte) map[string]bool {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	names := make(map[string]bool, len(zr.File))
	for _, f := range zr.File {
		names[f.Name] = true
	}
	return names
}

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

// TestRenderImage_Crop is acceptance criterion 1: a non-zero Crop emits an
// srcRect with the expected per-edge permille; a zero Crop emits none.
func TestRenderImage_Crop(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "img",
		Nodes: []scene.SlideNode{scene.Image{
			AssetID: "asset://x",
			Crop:    scene.Crop{Left: 0.1, Top: 0.2, Right: 0.05, Bottom: 0.1},
		}},
	}}}
	data, _ := render(t, sc, scene.WithAssetResolver(resolver))
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{`<a:srcRect`, `l="10000"`, `t="20000"`, `r="5000"`, `b="10000"`} {
		if !strings.Contains(slide, want) {
			t.Errorf("cropped image missing %q in:\n%s", want, slide)
		}
	}
}

// TestRenderImage_NoCrop confirms the default (zero) Crop emits no srcRect.
func TestRenderImage_NoCrop(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "img",
		Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x"}},
	}}}
	data, _ := render(t, sc, scene.WithAssetResolver(resolver))
	if slide := zipPart(t, data, "ppt/slides/slide1.xml"); strings.Contains(slide, "<a:srcRect") {
		t.Errorf("uncropped image unexpectedly emitted a srcRect:\n%s", slide)
	}
}

// TestRenderImage_Fit is acceptance criterion 2: FitNone omits the stretch fill;
// FitFill (the default) keeps it.
func TestRenderImage_Fit(t *testing.T) {
	resolver, _ := pngResolver()
	for _, tc := range []struct {
		name       string
		fit        scene.Fit
		wantStruct bool
	}{
		{"fill-default", scene.FitFill, true},
		{"none", scene.FitNone, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc := scene.Scene{Slides: []scene.SceneSlide{{
				ID:    "img",
				Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x", Fit: tc.fit}},
			}}}
			data, _ := render(t, sc, scene.WithAssetResolver(resolver))
			slide := zipPart(t, data, "ppt/slides/slide1.xml")
			has := strings.Contains(slide, "<a:stretch")
			if has != tc.wantStruct {
				t.Errorf("%s: stretch present = %v, want %v:\n%s", tc.name, has, tc.wantStruct, slide)
			}
		})
	}
}

// TestRenderImage_CropWithFrame is acceptance criterion 3: crop + fit compose
// with a frame — the cropped picture is placed inside the bezel interior.
func TestRenderImage_CropWithFrame(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "img",
		Nodes: []scene.SlideNode{scene.Image{
			AssetID: "asset://x",
			Frame:   scene.FrameBrowser,
			Crop:    scene.Crop{Left: 0.05, Right: 0.05},
			Fit:     scene.FitNone,
		}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(resolver))
	if stats.Shapes < 2 {
		t.Fatalf("Stats.Shapes = %d, want >= 2 (bezel + image)", stats.Shapes)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<p:sp>") || !strings.Contains(slide, "<p:pic>") {
		t.Errorf("framed+cropped image missing bezel or pic:\n%s", slide)
	}
	if !strings.Contains(slide, "<a:srcRect") {
		t.Errorf("framed+cropped image missing srcRect:\n%s", slide)
	}
	if strings.Contains(slide, "<a:stretch") {
		t.Errorf("FitNone image unexpectedly kept a stretch fill:\n%s", slide)
	}
}

// TestRenderImage_InvalidCrop is acceptance criterion 4: an out-of-range or
// over-crop fails Stage-1 validation.
func TestRenderImage_InvalidCrop(t *testing.T) {
	resolver, _ := pngResolver()
	for _, tc := range []struct {
		name string
		crop scene.Crop
	}{
		{"edge>1", scene.Crop{Left: 1.5}},
		{"edge<0", scene.Crop{Top: -0.2}},
		{"over-crop-horizontal", scene.Crop{Left: 0.6, Right: 0.6}},
		{"over-crop-vertical", scene.Crop{Top: 0.5, Bottom: 0.5}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sc := scene.Scene{Slides: []scene.SceneSlide{{
				ID:    "img",
				Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://x", Crop: tc.crop}},
			}}}
			_, err := scene.Render(pptx.New(), sc, scene.WithAssetResolver(resolver))
			if err == nil || !strings.Contains(err.Error(), "crop") {
				t.Fatalf("%s: want a crop validation error, got %v", tc.name, err)
			}
		})
	}
}

// TestRenderImage_SceneDedup is acceptance criterion 5: the same asset on two
// slides writes one media part at the scene seam (dedup is preserved).
func TestRenderImage_SceneDedup(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "a", Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://shared"}}},
		{ID: "b", Nodes: []scene.SlideNode{scene.Image{AssetID: "asset://shared"}}},
	}}
	data, stats := render(t, sc, scene.WithAssetResolver(resolver))
	if stats.Assets != 2 {
		t.Errorf("Stats.Assets = %d, want 2 (one per slide)", stats.Assets)
	}
	names := zipNames(t, data)
	if !names["ppt/media/image1.png"] {
		t.Error("expected ppt/media/image1.png to exist")
	}
	if names["ppt/media/image2.png"] {
		t.Error("identical asset bytes were not deduplicated (image2.png exists)")
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

// TestRenderImage_Framing is R13.11 acceptance 2: a scene Image with CornerRadius
// + Elevation set emits a roundRect-clipped pic with a drop shadow, and survives
// a round-trip (D-114).
func TestRenderImage_Framing(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "f",
		Nodes: []scene.SlideNode{scene.Image{
			AssetID: "asset://x", CornerRadius: scene.RadiusMD, Elevation: scene.ElevationRaised,
		}},
	}}}
	data, _ := render(t, sc, scene.WithAssetResolver(resolver))
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `prst="roundRect"`) {
		t.Errorf("framed image missing roundRect geometry:\n%s", slide)
	}
	if !strings.Contains(slide, "<a:outerShdw") {
		t.Errorf("framed image missing drop shadow:\n%s", slide)
	}
	// Round-trip: the framing survives reopen + re-write (G6).
	pres, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	out, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rt := zipPart(t, out, "ppt/slides/slide1.xml")
	if !strings.Contains(rt, `prst="roundRect"`) || !strings.Contains(rt, "<a:outerShdw") {
		t.Errorf("image framing did not survive round-trip")
	}
}

// TestRenderImage_FramingZeroByteIdentical is R13.11 acceptance 3: a scene Image
// with zero CornerRadius/Elevation is byte-identical to a plain image (D-114).
func TestRenderImage_FramingZeroByteIdentical(t *testing.T) {
	resolver, _ := pngResolver()
	mk := func(n scene.Image) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{ID: "z", Nodes: []scene.SlideNode{n}}}}
	}
	withZero, _ := render(t, mk(scene.Image{AssetID: "asset://x", CornerRadius: scene.RadiusNone, Elevation: scene.ElevationFlat}), scene.WithAssetResolver(resolver))
	plain, _ := render(t, mk(scene.Image{AssetID: "asset://x"}), scene.WithAssetResolver(resolver))
	if !bytes.Equal(withZero, plain) {
		t.Errorf("zero-token framed image not byte-identical to plain (%d vs %d bytes)", len(withZero), len(plain))
	}
	if strings.Contains(zipPart(t, withZero, "ppt/slides/slide1.xml"), `prst="roundRect"`) {
		t.Errorf("zero-radius image unexpectedly emitted roundRect")
	}
}
