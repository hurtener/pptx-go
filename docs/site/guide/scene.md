# The scene renderer

The `scene` package is the optional, IR-driven renderer. Instead of placing
shapes at EMU coordinates, you describe a deck as a typed `Scene` of nodes and
`Render` it onto a `*pptx.Presentation`. The layout engine computes placement;
you describe intent. `scene` composes [the builder](/guide/builder) — it never
reaches under it.

## The Scene model

```go
type Scene struct {
	Theme  *pptx.Theme // optional; the builder's default theme if nil
	Slides []SceneSlide
	Meta   Metadata    // deck core properties: Title, Author, Subject
}

type SceneSlide struct {
	ID      string      // a stable identifier, surfaced in Stats and warnings
	Layout  LayoutKind  // structural intent, mapped to a master layout
	Nodes   []SlideNode // the top-level node list
	Notes   RichText    // optional speaker notes
	Variant Variant     // optional theme variant
}
```

`LayoutKind` names a slide's structural intent: `LayoutCover`,
`LayoutTitleContent`, `LayoutTwoColumn`, `LayoutCardGrid`, `LayoutFullBleed`, and
`LayoutBlank`.

## A first render

```go
package main

import (
	"log"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func main() {
	sc := scene.Scene{
		Meta: scene.Metadata{Title: "Q3 Review", Author: "Acme"},
		Slides: []scene.SceneSlide{
			{
				ID:     "cover",
				Layout: scene.LayoutCover,
				Nodes: []scene.SlideNode{
					scene.Hero{
						Eyebrow:  "FY2026",
						Title:    "Quarterly Review",
						Subtitle: "Revenue, growth, and what's next",
					},
				},
			},
			{
				ID:     "intro",
				Layout: scene.LayoutTitleContent,
				Nodes: []scene.SlideNode{
					scene.Heading{Text: scene.RichText{{Text: "Highlights"}}, Level: 2},
					scene.Prose{Paragraphs: []scene.RichText{
						{{Text: "Revenue grew "}, {Text: "18%", Style: scene.RunStyle{Bold: true}}},
					}},
				},
				Notes: scene.RichText{{Text: "Pause for questions."}},
			},
		},
	}

	p := pptx.New()
	stats, err := scene.Render(p, sc, scene.WithWorkers(4))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("rendered %d slides, %d shapes, %d warnings",
		stats.Slides, stats.Shapes, len(stats.Warnings))

	if err := p.Save("review.pptx"); err != nil {
		log.Fatal(err)
	}
}
```

The full node catalog — text leaves, visual leaves, asset leaves, and containers
— is documented in the [scene catalog](/catalog/). See
[text leaves](/catalog/text-leaves), [visual leaves](/catalog/visual-leaves),
[asset leaves](/catalog/asset-leaves), and [containers](/catalog/containers).

## Rich text in the scene

A scene `RichText` is an ordered list of `TextRun`s, each plain text plus an
inline style and color:

```go
scene.RichText{
	{Text: "See the "},
	{
		Text:  "report",
		Style: scene.RunStyle{Link: true, Href: "https://example.com"},
		Color: scene.TokenTextColor(scene.TextAccent),
	},
}
```

`scene.RunStyle` carries `TypeRole`, `Bold`, `Italic`, `Underline`, `Strike`,
`Code` (inline mono), and `Link`/`Href`. Colors are `scene.TokenTextColor(role)`
(theme-bound, the default) or `scene.LiteralColor(hex)` (the escape hatch); the
zero color is the `TextPrimary` token.

## Rendering

```go
func Render(pres *pptx.Presentation, s Scene, opts ...RenderOption) (Stats, error)
```

`Render` validates the scene (Stage 1), applies the theme, lays out each slide's
nodes, and composes them onto the builder. It is deterministic: the same scene
and theme produce byte-identical output regardless of worker count.

### Render options

| Option | Purpose |
| --- | --- |
| `WithTheme(*pptx.Theme)` | Active theme for the render. Takes precedence over `Scene.Theme`. |
| `WithWorkers(int)` | Slides composed concurrently. Default `GOMAXPROCS`; `1` forces sequential. |
| `WithLogger(*slog.Logger)` | Structured render diagnostics (render-boundary summary + a `Warn` per layout warning). |
| `WithLayoutMap(LayoutMap)` | Maps each `LayoutKind` to a named layout in the active template's master. |
| `WithAssetResolver(AssetResolver)` | Resolves asset bytes for asset-bearing nodes. See [assets](/guide/assets). |
| `WithContext(context.Context)` | Context for the resolver and inter-slide cancellation. |
| `WithIconExtension(name, svg)` | Registers a caller icon for this render. See [assets](/guide/assets). |
| `WithFrameExtension(name, recipe)` | Registers a caller device-frame recipe for this render. |
| `WithOrnamentExtension(name, recipe)` | Registers a caller ornament recipe for this render. |

```go
stats, err := scene.Render(p, sc,
	scene.WithTheme(brand),
	scene.WithWorkers(8),
	scene.WithContext(ctx),
	scene.WithLayoutMap(scene.DefaultLayoutMap()),
)
```

## Validation

`ValidateScene(s Scene) error` runs Stage-1 structural validation and returns a
joined error reporting every problem at once (a heading level out of range, a
table row whose width doesn't match the header, an image with no asset id, …).
`Render` calls it for you; call it directly to validate ahead of time.

```go
if err := scene.ValidateScene(sc); err != nil {
	log.Fatal(err)
}
```

## Stats

`Render` returns a `Stats` value — the library's observability surface (there is
no event protocol):

```go
type Stats struct {
	Slides   int
	Shapes   int
	Assets   int
	Warnings []LayoutWarning // non-fatal layout / asset / token issues
	Timings  []SlideTiming   // per-slide compose duration, in scene order
}
```

`LayoutWarning` carries `SlideID`, `Node`, and `Message`. There is no strict
mode: a caller that wants warnings to be fatal inspects `Stats.Warnings` itself.

### Layout sizing and overflow

The renderer sizes each node's slot to its content: a text node's height grows
with the number of lines its text wraps to in the available width, so stacked
nodes don't overlap when a paragraph runs long. This estimate is deterministic
(it never depends on worker count) and is an *allotment*, not a prediction of
PowerPoint's exact on-screen reflow.

Because the height is content-aware, a slide whose text genuinely exceeds the
body region records a `content overflows its region` `LayoutWarning` — that is
the signal to inspect when you want to flag a slide as too full. Short,
single-line content is allotted the same compact height as before and never
falsely warns.
