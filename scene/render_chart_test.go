package scene_test

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/scene"
)

// pngOf builds a valid PNG of the given pixel dimensions (header carries the
// dimensions the chart composer reads via image.DecodeConfig).
func pngOf(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func chartResolverOf(b []byte) scene.AssetResolver {
	return scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "chart1" {
			return b, "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
}

// TestRenderChart is criterion 1: a chart with a raster renders a contained pic
// + caption, conformant, no warning for a slot-matching aspect.
func TestRenderChart(t *testing.T) {
	// A very wide image (~4:1) roughly matches the full-width single-node slot.
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "chart",
		Nodes: []scene.SlideNode{scene.Chart{AssetID: "asset://chart1", Caption: "Revenue by quarter"}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(chartResolverOf(pngOf(1600, 400))))
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("chart deck failed conformance:\n%s", rep)
	}
	if stats.Assets != 1 {
		t.Errorf("Stats.Assets = %d, want 1", stats.Assets)
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<p:pic>") {
		t.Errorf("chart missing the raster pic:\n%s", xml)
	}
	if !strings.Contains(xml, "<a:t>Revenue by quarter</a:t>") {
		t.Errorf("chart missing caption")
	}
}

// TestChartAspectWarning is criterion 2: a divergent aspect warns; a matching
// one does not.
func TestChartAspectWarning(t *testing.T) {
	mk := func() scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID:    "c",
			Nodes: []scene.SlideNode{scene.Chart{AssetID: "asset://chart1"}},
		}}}
	}
	// Square image in the wide full-width slot → large divergence → warning.
	_, sq := render(t, mk(), scene.WithAssetResolver(chartResolverOf(pngOf(600, 600))))
	if len(sq.Warnings) == 0 {
		t.Error("square chart in a wide slot should raise an aspect warning")
	} else if !strings.Contains(sq.Warnings[0].Message, "aspect ratio diverges") {
		t.Errorf("unexpected warning: %q", sq.Warnings[0].Message)
	}
	// Wide image (~4:1) close to the slot aspect → no warning.
	_, wide := render(t, mk(), scene.WithAssetResolver(chartResolverOf(pngOf(1600, 400))))
	if len(wide.Warnings) != 0 {
		t.Errorf("wide chart matching the slot should not warn: %+v", wide.Warnings)
	}
}

// TestChartUnresolvedPlaceholder is criterion 3: an unresolved asset renders a
// ChartPlaceholder + a warning (not an error).
func TestChartUnresolvedPlaceholder(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "c",
		Nodes: []scene.SlideNode{scene.Chart{AssetID: "asset://missing", Caption: "TBD"}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(chartResolverOf(pngOf(800, 600))))
	if len(stats.Warnings) == 0 {
		t.Error("unresolved chart should warn")
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, `prst="roundRect"`) || !strings.Contains(xml, "<a:t>Chart</a:t>") {
		t.Errorf("unresolved chart did not render a ChartPlaceholder:\n%s", xml)
	}
	if strings.Contains(xml, "<p:pic>") {
		t.Errorf("unresolved chart should not emit a pic")
	}
}

// TestChartParallel is criterion 5: byte-identical at workers=1 vs N (the dims
// read + warning text are deterministic).
func TestChartParallel(t *testing.T) {
	mk := func() scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{
			{ID: "a", Nodes: []scene.SlideNode{scene.Chart{AssetID: "asset://chart1", Caption: "A"}}},
			{ID: "b", Nodes: []scene.SlideNode{scene.Chart{AssetID: "asset://chart1"}}},
		}}
	}
	seq, _ := render(t, mk(), scene.WithAssetResolver(chartResolverOf(pngOf(600, 600))), scene.WithWorkers(1))
	par, _ := render(t, mk(), scene.WithAssetResolver(chartResolverOf(pngOf(600, 600))), scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Error("chart render differs between workers=1 and workers=4")
	}
}
