package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// logoResolver resolves "logoN" ids to real PNGs of varied aspect.
func logoResolver() scene.AssetResolver {
	return scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if strings.HasPrefix(uuid, "logo") {
			return pngOf(400, 200), "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
}

func logoWallScene(tone scene.LogoToneKind, n int) scene.Scene {
	logos := make([]scene.LogoEntry, n)
	for i := range logos {
		logos[i] = scene.LogoEntry{AssetID: scene.AssetID("asset://logo" + string(rune('0'+i%10)))}
	}
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "wall",
		Nodes: []scene.SlideNode{scene.LogoWall{Caption: "Trusted by", Columns: 4, Tone: tone, Logos: logos}},
	}}}
}

// TestLogoWall is R14.7 acceptance: a 12-logo mono wall renders pics 4-up, each
// contained (rounded? no — a pic), recolored via duotone, conformant, no warnings.
func TestLogoWall(t *testing.T) {
	data, stats := render(t, logoWallScene(scene.LogoToneMono, 12), scene.WithAssetResolver(logoResolver()))
	if len(stats.Warnings) != 0 {
		t.Errorf("logo wall: warnings: %+v", stats.Warnings)
	}
	if stats.Assets != 12 {
		t.Errorf("logo wall: want 12 assets, got %d", stats.Assets)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, "<p:pic>"); n != 12 {
		t.Errorf("logo wall: want 12 pics, got %d", n)
	}
	if !strings.Contains(slide, "<a:duotone>") {
		t.Errorf("logo wall mono tone: missing duotone recolor")
	}
	if !strings.Contains(slide, "<a:t>Trusted by</a:t>") {
		t.Errorf("logo wall: missing caption")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("logo wall deck failed conformance:\n%s", rep)
	}
}

// TestLogoWall_NoneTone verifies LogoToneNone emits no recolor (plain pics).
func TestLogoWall_NoneTone(t *testing.T) {
	data, _ := render(t, logoWallScene(scene.LogoToneNone, 3), scene.WithAssetResolver(logoResolver()))
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if strings.Contains(slide, "<a:duotone>") {
		t.Errorf("LogoToneNone should emit no duotone")
	}
}

// TestLogoWall_MissingWarns verifies an unresolvable logo warns and is skipped.
func TestLogoWall_MissingWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "wall",
		Nodes: []scene.SlideNode{scene.LogoWall{Logos: []scene.LogoEntry{{AssetID: "asset://nope"}, {AssetID: "asset://logo1"}}}},
	}}}
	_, stats := render(t, sc, scene.WithAssetResolver(logoResolver()))
	if len(stats.Warnings) == 0 {
		t.Errorf("logo wall with a missing logo: expected a warning")
	}
}

// TestLogoWall_Deterministic guards worker-count independence.
func TestLogoWall_Deterministic(t *testing.T) {
	sc := logoWallScene(scene.LogoToneBrand, 7)
	seq := renderBytes(t, sc, scene.WithAssetResolver(logoResolver()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithAssetResolver(logoResolver()), scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("logo wall not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}

// TestLogoWall_EmptyFails verifies an empty logo wall fails Stage-1 validation.
func TestLogoWall_EmptyFails(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "w", Nodes: []scene.SlideNode{scene.LogoWall{}}}}}
	if _, err := scene.Render(pptx.New(), sc, scene.WithAssetResolver(logoResolver())); err == nil {
		t.Errorf("empty logo wall should fail validation")
	}
}
