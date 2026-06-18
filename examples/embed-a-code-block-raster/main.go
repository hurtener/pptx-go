// Command embed-a-code-block-raster demonstrates placing a source-code listing
// in a deck as a caller-rendered raster.
//
// PowerPoint cannot preserve monospace metrics and whitespace fidelity, so a
// code block is NOT drawn as native text — it is a picture built from bytes you
// supply (D-014). The engine never syntax-highlights and never rasterizes
// (D-026): you render the listing to an image (a headless highlighter, a
// terminal screenshotter, an HTML-to-PNG step) and hand the bytes to the
// renderer through an AssetResolver. The engine then places the image and draws
// an optional language badge and caption natively over and below it (D-045).
//
// Run it:
//
//	go run ./examples/embed-a-code-block-raster
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

// mapResolver is a minimal map-backed AssetResolver: it looks an AssetID up in
// an in-memory map and returns the bytes plus a content-type hint. A miss
// returns scene.ErrAssetNotFound, the sentinel scene treats as "skip + warn".
//
// A real caller would back this with a code-to-image step: highlight the source
// (e.g. Chroma), render it to a bitmap, and cache the PNG by AssetID.
type mapResolver struct {
	assets map[scene.AssetID][]byte
}

func (r mapResolver) Resolve(_ context.Context, id scene.AssetID) ([]byte, string, error) {
	b, ok := r.assets[id]
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
	// 1. The caller owns rendering source code to pixels. Here we stand in for a
	//    syntax highlighter with a solid PNG; a real caller supplies a rendered
	//    listing instead.
	snippetPNG, err := makePNG(960, 540, color.RGBA{R: 0x1E, G: 0x1E, B: 0x2E, A: 0xFF})
	if err != nil {
		return fmt.Errorf("rasterize snippet: %w", err)
	}

	resolver := mapResolver{assets: map[scene.AssetID][]byte{
		"snippet": snippetPNG, // registered: this AssetID resolves to bytes
	}}

	// 2. A scene with one slide carrying a CodeBlock. AssetID points at the
	//    rendered image; Language draws a badge top-right; Caption draws a
	//    centered caption below the image (D-045).
	s := scene.Scene{
		Meta: scene.Metadata{Title: "Embed a code-block raster", Author: "pptx-go examples"},
		Slides: []scene.SceneSlide{
			{
				ID:     "slide-code",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					scene.CodeBlock{
						AssetID:  "snippet",
						Language: "go",
						Caption:  "main.go",
					},
				},
			},
		},
	}

	// 3. Wire the resolver into Render with the WithAssetResolver option.
	pres := pptx.New()
	stats, err := scene.Render(pres, s, scene.WithAssetResolver(resolver))
	if err != nil {
		// A resolver miss is NOT an error — reaching here means a real failure.
		return fmt.Errorf("render: %w", err)
	}

	// 4. Inspect Stats. stats.Assets counts pictures actually embedded; any
	//    unresolved CodeBlock would show up in stats.Warnings, not as an error.
	fmt.Printf("slides rendered:    %d\n", stats.Slides)
	fmt.Printf("assets embedded:    %d\n", stats.Assets)
	fmt.Printf("layout warnings:    %d\n", len(stats.Warnings))
	for _, w := range stats.Warnings {
		fmt.Printf("  - warning on slide %q: %s\n", w.SlideID, w.Message)
	}

	if stats.Assets != 1 {
		return fmt.Errorf("expected exactly 1 embedded asset (the code raster), got %d", stats.Assets)
	}
	if len(stats.Warnings) != 0 {
		return fmt.Errorf("expected no warnings, got %d", len(stats.Warnings))
	}

	// 5. Serialize the deck to bytes so the example produces a real artifact.
	deck, err := pres.WriteToBytes()
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	out := filepath.Join(os.TempDir(), "embed-a-code-block-raster.pptx")
	if err := os.WriteFile(out, deck, 0o644); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Printf("OK — wrote %s (%d bytes): 1 code raster embedded with a 'go' badge and 'main.go' caption\n", out, len(deck))
	return nil
}

// makePNG rasterizes a solid-color rectangle to PNG bytes using stdlib only.
// This stands in for whatever the caller uses to turn a source listing into an
// image — a syntax highlighter, a headless browser, a terminal screenshotter.
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
