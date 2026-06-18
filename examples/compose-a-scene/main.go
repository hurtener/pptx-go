// Command compose-a-scene builds a typed scene IR and renders it to a deck.
//
// It shows the Layer 2 path: instead of imperatively placing shapes with the
// pptx builder, you describe slides declaratively as a tree of typed nodes
// (a scene.Scene) and let scene.Render compose them onto a presentation.
//
// This example uses only native nodes (Hero, Heading, Prose, List, Callout,
// TwoColumn, Card). Asset-backed nodes (Image, Chart, CodeBlock) need an
// AssetResolver and are covered by sibling skills.
//
// Run: go run ./examples/compose-a-scene
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func main() {
	// A Scene is a theme + ordered slides + deck metadata. A nil Theme uses the
	// builder's default theme; every color/type below is a semantic token, so a
	// theme swap re-renders the same IR in a new visual language (P2).
	sc := scene.Scene{
		Meta: scene.Metadata{
			Title:   "Quarterly Review",
			Author:  "Platform Team",
			Subject: "Q2 results and roadmap",
		},
		Slides: []scene.SceneSlide{
			// Cover slide: a single Hero block (eyebrow + title + subtitle).
			{
				ID:     "cover",
				Layout: scene.LayoutCover,
				Nodes: []scene.SlideNode{
					scene.Hero{
						Eyebrow:  "Q2 2026",
						Title:    "Quarterly Review",
						Subtitle: "What shipped, what's next",
					},
				},
				Notes: scene.RichText{
					{Text: "Welcome the room; keep the cover up while people settle."},
				},
			},
			// Content slide: a heading, a paragraph, a checklist, a callout, and a
			// two-column split of cards. All native nodes — no assets needed.
			{
				ID:     "highlights",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					scene.Heading{
						Level: 2,
						Text: scene.RichText{
							{Text: "Highlights"},
						},
					},
					scene.Prose{
						Paragraphs: []scene.RichText{
							{
								{Text: "We focused on "},
								{Text: "reliability", Style: scene.RunStyle{Bold: true}},
								{Text: " this quarter, with "},
								{
									Text:  "measurable wins",
									Color: scene.TokenTextColor(scene.TextAccent),
								},
								{Text: " across the board."},
							},
						},
					},
					scene.List{
						Kind: scene.ListChecklist,
						Items: []scene.ListItem{
							{Text: scene.RichText{{Text: "Cut p99 latency by 38%"}}, Checked: true},
							{Text: scene.RichText{{Text: "Zero Sev-1 incidents"}}, Checked: true},
							{Text: scene.RichText{{Text: "Migrate the last legacy service"}}, Checked: false},
						},
					},
					scene.Callout{
						Kind:  scene.CalloutTip,
						Title: "Takeaway",
						Body: scene.RichText{
							{Text: "Investing in test coverage paid off — fewer rollbacks, faster ships."},
						},
					},
					scene.TwoColumn{
						Ratio: scene.Ratio11,
						Left: []scene.SlideNode{
							scene.Card{
								Eyebrow: "Now",
								Header:  "Stabilize",
								Fill:    scene.ColorSurface,
								Body: []scene.SlideNode{
									scene.Prose{Paragraphs: []scene.RichText{
										{{Text: "Lock in the reliability gains and document the runbooks."}},
									}},
								},
							},
						},
						Right: []scene.SlideNode{
							scene.Card{
								Eyebrow:   "Next",
								Header:    "Expand",
								Fill:      scene.ColorSurfaceAlt,
								Elevation: scene.ElevationRaised,
								Body: []scene.SlideNode{
									scene.Prose{Paragraphs: []scene.RichText{
										{{Text: "Open the platform to two new internal teams."}},
									}},
								},
							},
						},
					},
				},
			},
		},
	}

	// Validation runs automatically inside Render, but you can call it early to
	// fail fast before touching a presentation.
	if err := scene.ValidateScene(sc); err != nil {
		log.Fatalf("invalid scene: %v", err)
	}

	// scene.Render composes the IR onto a fresh presentation (P1: scene drives
	// the pptx builder). It returns Stats — counts, per-slide timings, and any
	// non-fatal layout warnings (warn-don't-fail).
	pres := pptx.New()
	stats, err := scene.Render(pres, sc)
	if err != nil {
		log.Fatalf("render: %v", err)
	}

	data, err := pres.WriteToBytes()
	if err != nil {
		log.Fatalf("write: %v", err)
	}
	out := filepath.Join(os.TempDir(), "compose-a-scene.pptx")
	if err := os.WriteFile(out, data, 0o644); err != nil {
		log.Fatalf("save: %v", err)
	}

	fmt.Printf("OK wrote %s — slides=%d shapes=%d assets=%d warnings=%d (%d bytes)\n",
		out, stats.Slides, stats.Shapes, stats.Assets, len(stats.Warnings), len(data))
}
