// Command extend-the-icon-set registers a caller-supplied SVG icon and uses it
// in a scene.
//
// Icons in pptx-go are a closed-name curated set plus a per-render extension
// seam (RFC §14.1, D-005, D-040). A node references an icon by name (Card.Icon,
// FlowStep.Icon); scene.WithIconExtension binds a caller SVG to a name for one
// render. The SVG is translated to native custom geometry (<a:custGeom>) and
// filled with the accent token by default (P2) — it is not embedded as an image.
//
// The translator accepts a small SVG subset: exactly one solid-filled <path>
// whose d data uses M/L/H/V/C/S/Q/T/Z (absolute or relative). Elliptical-arc
// commands (A/a) are rejected — author curves as Béziers. This program shows a
// valid icon being accepted and used, and an arc-bearing icon being rejected.
//
// Run: go run ./examples/extend-the-icon-set
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// validIcon is a single-path, solid-filled triangle using only M/L/Z commands
// over a 24x24 viewBox. The fill color is discarded — the icon renders in the
// active accent token.
const validIcon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">` +
	`<path fill="#000" d="M12 3 L21 20 L3 20 Z"/></svg>`

// invalidIcon uses an elliptical-arc command (A), which the translator rejects.
const invalidIcon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">` +
	`<path fill="#000" d="M12 12 A6 6 0 1 0 12 11.99 Z"/></svg>`

func main() {
	// 1) Validate the custom icon early. WithIconExtension runs the same check at
	// render time, but validating here fails fast with a clear error.
	if err := scene.ValidateIcon([]byte(validIcon)); err != nil {
		log.Fatalf("valid icon was rejected: %v", err)
	}

	// 2) Build a scene that references the custom icon by name. The name
	// ("my-icon") must resolve to a curated or registered icon, or Stage-1
	// validation fails. Here a Card and a Flow step both use it.
	sc := scene.Scene{
		Meta: scene.Metadata{Title: "Custom icon demo"},
		Slides: []scene.SceneSlide{
			{
				ID:     "icons",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					scene.Heading{
						Level: 2,
						Text:  scene.RichText{{Text: "Custom icons"}},
					},
					scene.Card{
						Header: "Custom",
						Icon:   "my-icon", // registered below via WithIconExtension
						Body: []scene.SlideNode{
							scene.Prose{Paragraphs: []scene.RichText{
								{{Text: "This card's glyph is a caller-supplied SVG, accent-tinted."}},
							}},
						},
					},
					scene.Flow{
						Steps: []scene.FlowStep{
							{Label: scene.RichText{{Text: "Plan"}}, Icon: "my-icon"},
							{Label: scene.RichText{{Text: "Build"}}, Icon: "check"}, // a curated name
							{Label: scene.RichText{{Text: "Ship"}}, Icon: "my-icon"},
						},
					},
				},
			},
		},
	}

	// 3) Render, registering the custom icon for THIS render only.
	pres := pptx.New()
	stats, err := scene.Render(pres, sc, scene.WithIconExtension("my-icon", []byte(validIcon)))
	if err != nil {
		log.Fatalf("render: %v", err)
	}

	data, err := pres.WriteToBytes()
	if err != nil {
		log.Fatalf("write: %v", err)
	}
	out := filepath.Join(os.TempDir(), "extend-the-icon-set.pptx")
	if err := os.WriteFile(out, data, 0o644); err != nil {
		log.Fatalf("save: %v", err)
	}

	// 4) Show the translator rejecting an out-of-subset SVG (an elliptical arc).
	arcErr := scene.ValidateIcon([]byte(invalidIcon))
	if arcErr == nil {
		log.Fatal("expected the arc icon to be rejected, but it passed")
	}

	fmt.Printf("OK wrote %s — slides=%d shapes=%d assets=%d warnings=%d (%d bytes)\n",
		out, stats.Slides, stats.Shapes, stats.Assets, len(stats.Warnings), len(data))
	fmt.Printf("arc icon correctly rejected: %v\n", arcErr)
}
