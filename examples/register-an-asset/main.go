// Command register-an-asset demonstrates supplying caller-rasterized bytes for
// asset-bearing scene nodes (Image, Chart, CodeBlock, asset Decoration).
//
// A scene node that carries an AssetID renders as a picture built from YOUR
// bytes — pptx-go never rasterizes (D-026). You supply the bytes through an
// AssetResolver wired in with scene.WithAssetResolver. A missing asset is
// non-fatal: the render succeeds, the node is skipped, and the miss surfaces as
// a LayoutWarning in Stats (warn-don't-fail; there is no strict mode in V1,
// D-036).
//
// Run it:
//
//	go run ./examples/register-an-asset
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
// A real caller would back this with a blob store, a cache, or an on-the-fly
// rasterizer — the contract is the same: (bytes, contentTypeHint, error).
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
	// 1. The caller owns rasterization: produce real PNG bytes for "logo".
	logoPNG, err := makePNG(480, 240, color.RGBA{R: 0x2E, G: 0x6F, B: 0xF2, A: 0xFF})
	if err != nil {
		return fmt.Errorf("rasterize logo: %w", err)
	}

	resolver := mapResolver{assets: map[scene.AssetID][]byte{
		"logo": logoPNG, // registered: this AssetID resolves to bytes
		// note: no entry for "missing-banner" — that AssetID will not resolve
	}}

	// 2. A scene with two slides: one references a registered asset, the other
	//    references an AssetID the resolver does not know about.
	s := scene.Scene{
		Meta: scene.Metadata{Title: "Register an asset", Author: "pptx-go examples"},
		Slides: []scene.SceneSlide{
			{
				ID:     "slide-resolved",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					scene.Image{AssetID: "logo", Alt: "Company logo"},
				},
			},
			{
				ID:     "slide-missing",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					// This AssetID has no bytes: the render skips the node and
					// records a LayoutWarning instead of failing.
					scene.Image{AssetID: "missing-banner", Alt: "Banner (unresolved)"},
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

	// 4. Inspect Stats. stats.Assets counts pictures actually embedded; the
	//    unresolved node shows up in stats.Warnings, not as an error.
	fmt.Printf("slides rendered:    %d\n", stats.Slides)
	fmt.Printf("assets embedded:    %d\n", stats.Assets)
	fmt.Printf("layout warnings:    %d\n", len(stats.Warnings))
	for _, w := range stats.Warnings {
		fmt.Printf("  - warning on slide %q: %s\n", w.SlideID, w.Message)
	}

	if stats.Assets != 1 {
		return fmt.Errorf("expected exactly 1 embedded asset (the resolved logo), got %d", stats.Assets)
	}
	if len(stats.Warnings) == 0 {
		return fmt.Errorf("expected a warning for the unresolved asset, got none")
	}

	// 5. Save the deck so the example produces a real, openable artifact.
	out := filepath.Join(os.TempDir(), "register-an-asset.pptx")
	if err := pres.Save(out); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	fmt.Printf("OK — wrote %s (1 resolved asset embedded, 1 unresolved asset skipped with a warning)\n", out)
	return nil
}

// makePNG rasterizes a solid-color rectangle to PNG bytes using stdlib only.
// This stands in for whatever the caller uses to turn its content (a logo, a
// chart, a code listing) into image bytes.
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
