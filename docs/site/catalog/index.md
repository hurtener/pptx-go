# Scene node catalog

The `scene` package describes a deck declaratively as a typed `Scene` of
**nodes**. Each node is a value of a concrete Go type that implements the sealed
`SlideNode` interface. The set is closed: you construct one of the catalog types
in the `scene` package, and `NodeKind()` discriminates it for validation and
rendering.

```go
// SlideNode is the sealed scene IR union.
type SlideNode interface {
	NodeKind() NodeKind
	isSlideNode()
}
```

`NodeKind` is an enum with one value per node type (`KindHero`, `KindProse`,
`KindHeading`, `KindList`, `KindDivider`, `KindQuote`, `KindCallout`,
`KindImage`, `KindChip`, `KindArrow`, `KindCodeBlock`, `KindChart`, `KindTable`,
`KindFlow`, `KindDecoration`, `KindSectionDivider`, `KindTwoColumn`, `KindGrid`,
`KindCard`, `KindCardSection`).

## Render policy: native vs picture

Every node renders one of two ways, and the choice is intrinsic to the node
type — there is no deck-wide toggle (D-011, D-018):

- **Picture (`pic`) nodes** carry an `AssetID` field. The caller pre-rasterizes
  the content and supplies the bytes through an
  [`AssetResolver`](/guide/assets); pptx-go embeds them verbatim as a picture
  shape. These are `Image`, `Chart`, `CodeBlock`, and a `Decoration` whose
  `Kind` is `DecorationAsset`. See [asset leaves](/catalog/asset-leaves).
- **Native nodes** render as native PowerPoint shapes (rectangles, text bodies,
  lines, tables, preset geometries) resolved from theme tokens. Every other node
  is native, including a `Decoration` whose `Kind` is `DecorationPreset`.

Container nodes (`TwoColumn`, `Grid`, `Card`, `CardSection`) do not have a
policy of their own: they render their children, and **each child renders per
its own policy**. A native `Card` can hold an `Image` child that renders as a
picture.

## Categories

| Category | Nodes |
| --- | --- |
| [Text leaves](/catalog/text-leaves) | Hero, Prose, Heading, List, Quote, Callout, Chip, Arrow |
| [Visual leaves](/catalog/visual-leaves) | Divider, SectionDivider, Table, Flow, Decoration (preset) |
| [Asset leaves](/catalog/asset-leaves) | Image, CodeBlock, Chart, Decoration (asset) |
| [Containers](/catalog/containers) | TwoColumn, Grid, Card, CardSection |

## How to render

Build a `Scene`, then `Render` it onto a `*pptx.Presentation`. The per-node
snippets on the category pages show only node construction; this is the full
surrounding boilerplate.

```go
package main

import (
	"log"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func main() {
	sc := scene.Scene{
		Meta: scene.Metadata{Title: "Quarterly review"},
		Slides: []scene.SceneSlide{
			{
				ID:     "cover",
				Layout: scene.LayoutCover,
				Nodes: []scene.SlideNode{
					scene.Hero{
						Eyebrow:  "FY26",
						Title:    "Quarterly review",
						Subtitle: "Results and outlook",
					},
				},
			},
		},
	}

	pres := pptx.New()
	if _, err := scene.Render(pres, sc); err != nil {
		log.Fatal(err)
	}
	if err := pres.Save("review.pptx"); err != nil {
		log.Fatal(err)
	}
}
```

`Render` returns a [`Stats`](/reference/scene) struct (slide/shape/asset counts,
per-slide timings, and non-fatal warnings) and an error. Picture nodes require a
resolver passed via `scene.WithAssetResolver` — see [assets](/guide/assets). For
a fuller tour of `Scene`, `SceneSlide`, layouts, and render options, see the
[scene guide](/guide/scene).
