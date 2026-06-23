package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Black-box render tests for Lockup (R12.9, D-102): the icon path (media-free), the asset
// path (a pic via the resolver), validation, and determinism.

// TestLockup_IconPath: an icon lockup renders the caption + a custGeom glyph, no media.
func TestLockup_IconPath(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "cover", Nodes: []scene.SlideNode{
			scene.Lockup{Caption: "POWERED BY CLEAR TECH", Icon: "star", AssetSide: scene.TrailCaption},
		}},
	}}
	data, stats := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "POWERED BY CLEAR TECH") {
		t.Error("lockup caption missing")
	}
	if !strings.Contains(xml, "<a:custGeom>") {
		t.Error("icon lockup did not emit a custGeom glyph")
	}
	if stats.Assets != 0 {
		t.Errorf("icon lockup registered %d assets, want 0", stats.Assets)
	}
}

// TestLockup_AssetPath: an asset lockup resolves the logo and registers a pic.
func TestLockup_AssetPath(t *testing.T) {
	resolver := scene.URIAssetResolver(func(string) ([]byte, string, error) {
		return append([]byte("\x89PNG\r\n\x1a\n"), []byte("logo")...), "image/png", nil
	})
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "cover", Nodes: []scene.SlideNode{
			scene.Lockup{Caption: "IN PARTNERSHIP WITH", AssetID: "asset://logo"},
		}},
	}}
	pres := pptx.New()
	stats, err := scene.Render(pres, sc, scene.WithAssetResolver(resolver))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if stats.Assets != 1 {
		t.Errorf("asset lockup registered %d assets, want 1", stats.Assets)
	}
}

// TestLockup_Validation: a lockup with neither / both of asset+icon fails Stage-1.
func TestLockup_Validation(t *testing.T) {
	none := scene.Scene{Slides: []scene.SceneSlide{{ID: "x", Nodes: []scene.SlideNode{scene.Lockup{Caption: "x"}}}}}
	if err := scene.ValidateScene(none); err == nil {
		t.Error("lockup with no asset/icon passed validation")
	}
	both := scene.Scene{Slides: []scene.SceneSlide{{ID: "x", Nodes: []scene.SlideNode{scene.Lockup{Caption: "x", AssetID: "a", Icon: "star"}}}}}
	if err := scene.ValidateScene(both); err == nil {
		t.Error("lockup with both asset and icon passed validation")
	}
}

// TestLockup_UnknownIconFails: an unknown lockup icon fails Stage-1 at Render.
func TestLockup_UnknownIconFails(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "bad", Nodes: []scene.SlideNode{scene.Lockup{Caption: "x", Icon: "no-such-icon"}}},
	}}
	if _, err := scene.Render(pptx.New(), sc); err == nil || !strings.Contains(err.Error(), "no-such-icon") {
		t.Fatalf("want a Stage-1 error naming the icon, got %v", err)
	}
}

// TestLockup_Deterministic: an icon lockup renders byte-identically across workers.
func TestLockup_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "c", Content: scene.Alignment{Horizontal: scene.HAlignCenter}, Nodes: []scene.SlideNode{
			scene.Lockup{Caption: "POWERED BY CLEAR TECH", Icon: "star", MaxHeight: pptx.In(0.5)},
		}},
	}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("lockup render: parallel (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}
}
