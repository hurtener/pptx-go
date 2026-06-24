package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
	"github.com/hurtener/pptx-go/scene/ornaments"
)

// TestDecoration_CuratedOrnaments is acceptance criterion 6 (PR #2): each
// curated ornament renders at least one shape with no validation error.
func TestDecoration_CuratedOrnaments(t *testing.T) {
	for _, name := range ornaments.Curated().Names() {
		t.Run(name, func(t *testing.T) {
			sc := scene.Scene{Slides: []scene.SceneSlide{{
				ID:    "d",
				Nodes: []scene.SlideNode{scene.Decoration{Kind: scene.DecorationPreset, Preset: name, Anchor: scene.AnchorCenter}},
			}}}
			_, stats := render(t, sc)
			if stats.Shapes < 1 {
				t.Errorf("ornament %q rendered %d shapes, want >= 1", name, stats.Shapes)
			}
			for _, w := range stats.Warnings {
				t.Errorf("unexpected warning for %q: %s", name, w.Message)
			}
		})
	}
}

// TestDecoration_LayerZOrder is acceptance criterion 7: a background decoration's
// shapes precede the body's, and a foreground decoration's follow (RFC §10.2).
func TestDecoration_LayerZOrder(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "z",
		Nodes: []scene.SlideNode{
			// Authored fg-first to prove the renderer reorders by Layer, not IR order.
			scene.Decoration{Kind: scene.DecorationPreset, Preset: "chevron_arrow", Layer: scene.LayerForeground, Anchor: scene.AnchorCenter},
			scene.Prose{Paragraphs: []scene.RichText{rt("BODYTEXT")}},
			scene.Decoration{Kind: scene.DecorationPreset, Preset: "grid_dots", Layer: scene.LayerBackground, Anchor: scene.AnchorCenter},
		},
	}}}
	data, _ := render(t, sc)
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	bg := strings.Index(slide, `prst="ellipse"`) // grid_dots (background)
	body := strings.Index(slide, "BODYTEXT")     // prose body
	fg := strings.Index(slide, `prst="chevron"`) // chevron (foreground)
	if bg < 0 || body < 0 || fg < 0 {
		t.Fatalf("missing markers: bg=%d body=%d fg=%d\n%s", bg, body, fg, slide)
	}
	if bg >= body || body >= fg {
		t.Errorf("z-order wrong: background(%d) < body(%d) < foreground(%d) expected", bg, body, fg)
	}
}

// TestDecoration_BleedSuppressesWarning is acceptance criterion 6 (bleed): a
// bleed decoration placed off-canvas renders without a warning; the same
// placement without Bleed warns.
func TestDecoration_BleedSuppressesWarning(t *testing.T) {
	mk := func(bleed bool) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID: "b",
			Nodes: []scene.SlideNode{scene.Decoration{
				Kind:   scene.DecorationPreset,
				Preset: "radial_glow",
				Anchor: scene.AnchorTopLeft,
				Offset: scene.Position{X: -pptx.In(1), Y: -pptx.In(1)}, // pushes off the top-left
				Size:   scene.Size{W: pptx.In(2), H: pptx.In(2)},
				Bleed:  bleed,
			}},
		}}}
	}
	if _, stats := render(t, mk(true)); len(stats.Warnings) != 0 {
		t.Errorf("bleed decoration should not warn, got: %+v", stats.Warnings)
	}
	_, stats := render(t, mk(false))
	var found bool
	for _, w := range stats.Warnings {
		if strings.Contains(w.Message, "past the slide") {
			found = true
		}
	}
	if !found {
		t.Errorf("non-bleed off-canvas decoration should warn, got: %+v", stats.Warnings)
	}
}

// TestDecoration_PresetOpacity checks a preset decoration's Opacity flows
// through to the ornament's accent alpha (a solid ornament dims via alpha).
func TestDecoration_PresetOpacity(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "op",
		Nodes: []scene.SlideNode{scene.Decoration{Kind: scene.DecorationPreset, Preset: "grid_dots", Anchor: scene.AnchorCenter, Opacity: 0.5}},
	}}}
	data, _ := render(t, sc)
	if slide := zipPart(t, data, "ppt/slides/slide1.xml"); !strings.Contains(slide, `<a:alpha val="50000"`) {
		t.Errorf("preset decoration opacity did not reach the accent alpha:\n%s", slide)
	}
}

// TestDecoration_OrnamentExtension is acceptance criterion 8: a caller ornament
// registered via WithOrnamentExtension renders.
func TestDecoration_OrnamentExtension(t *testing.T) {
	custom := func(sl *pptx.Slide, box pptx.Box, alpha int, _ float64, role pptx.ColorRole) int {
		sl.AddShape(pptx.ShapeRect, box, pptx.WithFill(pptx.SolidFill(pptx.TokenColorAlpha(role, alpha))))
		return 1
	}
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "e",
		Nodes: []scene.SlideNode{scene.Decoration{Kind: scene.DecorationPreset, Preset: "spark", Anchor: scene.AnchorCenter}},
	}}}
	_, stats := render(t, sc, scene.WithOrnamentExtension("spark", custom))
	if stats.Shapes != 1 {
		t.Fatalf("Stats.Shapes = %d, want 1 (the extension ornament)", stats.Shapes)
	}
	if len(stats.Warnings) != 0 {
		t.Errorf("unexpected warnings: %+v", stats.Warnings)
	}
}

// TestDecoration_UnknownOrnament is the Stage-1 closed-name check: an
// unregistered ornament name fails validation.
func TestDecoration_UnknownOrnament(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "u",
		Nodes: []scene.SlideNode{scene.Decoration{Kind: scene.DecorationPreset, Preset: "ghost", Anchor: scene.AnchorCenter}},
	}}}
	if _, err := scene.Render(pptx.New(), sc); err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("unknown ornament not rejected; err = %v", err)
	}
}

// TestDecoration_OpacityRange is the Stage-1 opacity check.
func TestDecoration_OpacityRange(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "o",
		Nodes: []scene.SlideNode{scene.Decoration{Kind: scene.DecorationPreset, Preset: "radial_glow", Opacity: 1.5}},
	}}}
	if _, err := scene.Render(pptx.New(), sc); err == nil || !strings.Contains(err.Error(), "opacity") {
		t.Fatalf("out-of-range opacity not rejected; err = %v", err)
	}
}

// TestDecoration_Asset renders an asset-kind decoration as a picture.
func TestDecoration_Asset(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "a",
		Nodes: []scene.SlideNode{scene.Decoration{Kind: scene.DecorationAsset, AssetID: "asset://x", Anchor: scene.AnchorCenter}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(resolver))
	if stats.Assets != 1 {
		t.Errorf("Stats.Assets = %d, want 1", stats.Assets)
	}
	if slide := zipPart(t, data, "ppt/slides/slide1.xml"); !strings.Contains(slide, "<p:pic>") {
		t.Errorf("asset decoration missing pic:\n%s", slide)
	}
}

// TestDecoration_AssetRotationOpacity checks an asset decoration honors Rotation
// and Opacity (the Phase-13 audit wiring — previously dropped on the asset path).
func TestDecoration_AssetRotationOpacity(t *testing.T) {
	resolver, _ := pngResolver()
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "a",
		Nodes: []scene.SlideNode{scene.Decoration{
			Kind: scene.DecorationAsset, AssetID: "asset://x", Anchor: scene.AnchorCenter,
			Rotation: 45, Opacity: 0.5,
		}},
	}}}
	data, _ := render(t, sc, scene.WithAssetResolver(resolver))
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `rot="2700000"`) { // 45 × 60000
		t.Errorf("asset decoration missing rotation:\n%s", slide)
	}
	if !strings.Contains(slide, `<a:alphaModFix amt="50000"`) {
		t.Errorf("asset decoration missing opacity (alphaModFix):\n%s", slide)
	}
}

// TestDecoration_ColorRole is R13.5 acceptance 1: a decoration with an explicit
// Color renders a different srgbClr than the accent default for the same preset
// (D-107).
func TestDecoration_ColorRole(t *testing.T) {
	role := pptx.ColorError // DC2626, distinct from the default accent 2563EB
	mk := func(color *pptx.ColorRole) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID: "c",
			Nodes: []scene.SlideNode{scene.Decoration{
				Kind: scene.DecorationPreset, Preset: ornaments.NameGridDots,
				Anchor: scene.AnchorCenter, Color: color,
			}},
		}}}
	}
	accent := zipPart(t, mustRender(t, mk(nil)), "ppt/slides/slide1.xml")
	custom := zipPart(t, mustRender(t, mk(&role)), "ppt/slides/slide1.xml")

	if !strings.Contains(accent, "2563EB") {
		t.Errorf("default decoration did not use the accent color 2563EB")
	}
	if !strings.Contains(custom, "DC2626") {
		t.Errorf("Color=ColorError decoration did not use DC2626")
	}
	if strings.Contains(custom, "2563EB") {
		t.Errorf("Color=ColorError decoration leaked the accent color 2563EB")
	}
}

// TestDecoration_ColorNilByteIdentical is R13.5 acceptance 2: a nil Color
// decoration is byte-identical to itself across renders (and uses the accent
// default), for every curated preset (D-107).
func TestDecoration_ColorNilByteIdentical(t *testing.T) {
	for _, preset := range ornaments.Curated().Names() {
		t.Run(preset, func(t *testing.T) {
			sc := scene.Scene{Slides: []scene.SceneSlide{{
				ID: "n",
				Nodes: []scene.SlideNode{scene.Decoration{
					Kind: scene.DecorationPreset, Preset: preset, Anchor: scene.AnchorCenter,
				}},
			}}}
			a := mustRender(t, sc)
			b := mustRender(t, sc)
			if !bytes.Equal(a, b) {
				t.Errorf("%s nil-Color decoration not deterministic (%d vs %d bytes)", preset, len(a), len(b))
			}
		})
	}
}

func mustRender(t *testing.T, sc scene.Scene) []byte {
	t.Helper()
	data, _ := render(t, sc)
	return data
}

// TestDecoration_TextWatermark is R13.9 acceptance 1: a DecorationText renders one
// run carrying the text and a low <a:alpha> (from Opacity) (D-109).
func TestDecoration_TextWatermark(t *testing.T) {
	grey := pptx.ColorSurfaceAlt
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "w",
		Nodes: []scene.SlideNode{scene.Decoration{
			Kind: scene.DecorationText, Text: "03", Color: &grey, Opacity: 0.08,
			Anchor: scene.AnchorCenter, Size: scene.Size{W: pptx.In(6), H: pptx.In(6)},
		}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("text watermark: unexpected warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, ">03<") {
		t.Errorf("text watermark missing the text run:\n%s", slide)
	}
	if !strings.Contains(slide, `<a:alpha val="8000"`) {
		t.Errorf("text watermark missing the low alpha (8000):\n%s", slide)
	}
}

// TestDecoration_TextWatermarkEmpty is R13.9 acceptance 2: an empty Text fails
// Stage-1 validation (D-109).
func TestDecoration_TextWatermarkEmpty(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "we",
		Nodes: []scene.SlideNode{scene.Decoration{Kind: scene.DecorationText, Anchor: scene.AnchorCenter}},
	}}}
	if _, err := scene.Render(pptx.New(), sc); err == nil || !strings.Contains(err.Error(), "text") {
		t.Fatalf("empty text watermark not rejected; err = %v", err)
	}
}

// TestDecoration_TextWatermarkDeterministic is R13.9 acceptance 3: a text
// watermark re-renders byte-identically (D-109).
func TestDecoration_TextWatermarkDeterministic(t *testing.T) {
	grey := pptx.ColorSurfaceAlt
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "wd",
		Nodes: []scene.SlideNode{scene.Decoration{
			Kind: scene.DecorationText, Text: "02", Color: &grey, Opacity: 0.1,
			Anchor: scene.AnchorTopRight,
		}},
	}}}
	a, _ := render(t, sc)
	b, _ := render(t, sc)
	if !bytes.Equal(a, b) {
		t.Errorf("text watermark not deterministic (%d vs %d bytes)", len(a), len(b))
	}
}
