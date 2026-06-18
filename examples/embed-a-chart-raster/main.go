// Command embed-a-chart-raster demonstrates placing a chart in a deck as a
// caller-rasterized image. In V1, a chart is an image-shape (native c:chart is
// V2 — D-004): pptx-go never rasterizes (D-026), so YOU render the chart to a
// PNG/SVG and the engine positions those bytes.
//
// It shows both paths in one program:
//
//	(a) Scene path   — a scene.Chart node resolves its bytes through an
//	                   AssetResolver; the renderer contains-to-fit the image in
//	                   its slot, draws a caption, and warns when the image's
//	                   aspect ratio diverges from the slot (D-046).
//	(b) Builder path — on a raw pptx deck, ChartPlaceholder draws a labeled empty
//	                   slot, and AddImage drops a rendered chart picture in.
//
// Run it:
//
//	go run ./examples/embed-a-chart-raster
package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// chartResolver is a minimal map-backed AssetResolver: it maps an AssetID to
// pre-rendered chart bytes plus a content-type hint. A miss returns
// scene.ErrAssetNotFound — the sentinel scene degrades to a LayoutWarning +
// ChartPlaceholder (warn-don't-fail; there is no strict mode in V1, D-036).
//
// A real caller would back this with whatever turns its data into a chart image:
// a Go charting library, a headless browser screenshot, a cached blob store.
type chartResolver struct {
	charts map[scene.AssetID][]byte
}

func (r chartResolver) Resolve(_ context.Context, id scene.AssetID) ([]byte, string, error) {
	b, ok := r.charts[id]
	if !ok {
		return nil, "", scene.ErrAssetNotFound
	}
	return b, "image/png", nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// The caller owns chart rendering: produce real PNG bytes for the chart.
	// Here it is a solid rectangle; in a real program it would be the rasterized
	// chart of your data.
	revenuePNG, err := makePNG(960, 540, color.RGBA{R: 0x2E, G: 0x6F, B: 0xF2, A: 0xFF})
	if err != nil {
		return fmt.Errorf("rasterize revenue chart: %w", err)
	}

	if err := scenePath(revenuePNG); err != nil {
		return fmt.Errorf("scene path: %w", err)
	}
	if err := builderPath(revenuePNG); err != nil {
		return fmt.Errorf("builder path: %w", err)
	}
	return nil
}

// scenePath renders a scene.Chart node whose bytes come from an AssetResolver,
// plus a second slide whose chart AssetID does not resolve (to show the
// placeholder-and-warn fallback).
func scenePath(chartPNG []byte) error {
	resolver := chartResolver{charts: map[scene.AssetID][]byte{
		"revenue-chart": chartPNG, // registered: resolves to bytes
		// note: no entry for "forecast-chart" — that AssetID will not resolve.
	}}

	s := scene.Scene{
		Meta: scene.Metadata{Title: "Embed a chart raster", Author: "pptx-go examples"},
		Slides: []scene.SceneSlide{
			{
				ID:     "slide-chart",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					scene.Chart{AssetID: "revenue-chart", Caption: "Q3 revenue"},
				},
			},
			{
				ID:     "slide-unresolved",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					// No bytes for this id: the renderer draws a ChartPlaceholder
					// (a labeled dashed slot) and records a LayoutWarning — it does
					// not leave a blank gap and it does not fail (D-046, D-036).
					scene.Chart{AssetID: "forecast-chart", Caption: "Q4 forecast"},
				},
			},
		},
	}

	pres := pptx.New()
	stats, err := scene.Render(pres, s, scene.WithAssetResolver(resolver))
	if err != nil {
		// A resolver miss is NOT an error — reaching here means a real failure.
		return fmt.Errorf("render: %w", err)
	}

	fmt.Println("[scene path]")
	fmt.Printf("  slides rendered: %d\n", stats.Slides)
	fmt.Printf("  charts embedded: %d\n", stats.Assets)
	fmt.Printf("  layout warnings: %d\n", len(stats.Warnings))
	for _, w := range stats.Warnings {
		fmt.Printf("    - %s: %s\n", w.SlideID, w.Message)
	}

	if stats.Assets != 1 {
		return fmt.Errorf("expected exactly 1 embedded chart (the resolved one), got %d", stats.Assets)
	}
	if len(stats.Warnings) == 0 {
		return fmt.Errorf("expected a warning for the unresolved chart, got none")
	}

	out := filepath.Join(os.TempDir(), "embed-a-chart-raster-scene.pptx")
	if err := pres.Save(out); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	fmt.Printf("  OK — wrote %s\n", out)
	return nil
}

// builderPath drops a chart into a raw pptx deck two ways on one slide: a
// labeled empty ChartPlaceholder slot, and a rendered chart picture via
// AddImage.
func builderPath(chartPNG []byte) error {
	pres := pptx.New()
	sl := pres.AddSlide()

	// A labeled empty slot — the visible stand-in for a chart whose bytes are
	// not yet committed. Fills/border resolve against the active theme (P2).
	sl.ChartPlaceholder(pptx.Box{X: pptx.In(0.5), Y: pptx.In(1.5), W: pptx.In(5.5), H: pptx.In(4)})

	// A rendered chart picture. The caller owns the bytes; ImageBytes verifies
	// the declared MIME against them.
	if _, err := sl.AddImage(
		pptx.ImageBytes(chartPNG, "image/png"),
		pptx.Box{X: pptx.In(6.5), Y: pptx.In(1.5), W: pptx.In(6), H: pptx.In(3.375)},
	); err != nil {
		return fmt.Errorf("add chart image: %w", err)
	}

	data, err := pres.WriteToBytes()
	if err != nil {
		return fmt.Errorf("write deck: %w", err)
	}

	out := filepath.Join(os.TempDir(), "embed-a-chart-raster-builder.pptx")
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Println("[builder path]")
	fmt.Printf("  OK — wrote %s (%d bytes: 1 placeholder slot + 1 chart picture)\n", out, len(data))
	return nil
}

// makePNG rasterizes a solid-color rectangle to PNG bytes using stdlib only.
// This stands in for whatever the caller uses to turn its data into a chart
// image — pptx-go never rasterizes (D-026).
func makePNG(w, h int, c color.Color) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
