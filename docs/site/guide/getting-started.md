# Getting started

**pptx-go** (`github.com/hurtener/pptx-go`, Apache-2.0) is a pure-Go library for
authoring and reading PowerPoint (PPTX / Open Office XML) files. It has no CGo
and no third-party runtime dependencies — the shipped artifact compiles
`CGO_ENABLED=0` and cross-compiles anywhere Go does.

## Install

```bash
go get github.com/hurtener/pptx-go
```

The module requires Go 1.24+. Import `github.com/hurtener/pptx-go/pptx` for the
builder and `github.com/hurtener/pptx-go/scene` for the optional renderer.

## The two layers

pptx-go exposes exactly two public packages:

- **`pptx`** — the builder. A general-purpose, theme-aware API for creating
  slides, shapes, rich text, tables, images, speaker notes, and sections, and
  for reading a deck back into a navigable model.
- **`scene`** — the renderer. Describe a deck declaratively as a typed `Scene`
  of nodes (`Hero`, `Card`, `Grid`, `Table`, `Flow`, …) and `Render` it onto a
  builder. `scene` composes `pptx`; nothing in `scene` reaches under it.

The two consumer paths are independent. A direct consumer of [the builder](/guide/builder)
writes generic Go and gets a production-grade deck. A consumer of [the scene
renderer](/guide/scene) builds a typed IR and gets a PPTX with no opinionated
boilerplate.

## Author, save, read back

A complete round trip: build a deck, write it to bytes, and reopen it into the
same navigable model.

```go
package main

import (
	"fmt"
	"log"

	"github.com/hurtener/pptx-go/pptx"
)

func main() {
	// Author. New() yields a valid 16:9 deck themed with the default theme.
	p := pptx.New()
	s := p.AddSlide()

	box := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(8), H: pptx.In(1.5)}
	tf := s.AddTextFrame(box)
	tf.AddParagraph(pptx.ParagraphOpts{}).
		AddRun("Hello, pptx-go", pptx.RunStyle{
			TypeRole: pptx.TypeH1,
			Color:    pptx.TokenTextColor(pptx.TextPrimary),
		})

	// Save to bytes (use p.Save("deck.pptx") to write a file).
	data, err := p.WriteToBytes()
	if err != nil {
		log.Fatal(err)
	}

	// Read back. A self-authored deck round-trips losslessly.
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		log.Fatal(err)
	}
	for _, sl := range re.Slides() {
		for _, sh := range sl.Shapes() {
			if frame, ok := sh.TextFrame(); ok {
				for _, para := range frame.Paragraphs() {
					for _, run := range para.Runs() {
						fmt.Println(run.Text())
					}
				}
			}
		}
	}
}
```

See [reading decks](/guide/reading) for the full read surface, and
[themes & tokens](/guide/theme) for token-driven styling.

## When to use `pptx` vs `scene`

Reach for **`pptx`** when you want imperative control: you know where each shape
goes, you are editing an existing deck, or you are reading a deck back. Reach for
**`scene`** when you want to describe *what* a slide contains and let the layout
engine place it — a hero, a two-column split, a card grid — without computing
EMU coordinates yourself.

They are not mutually exclusive: `scene.Render` takes a `*pptx.Presentation`, so
you can render a scene and then drop down to the builder on the same deck.

## The binding properties

pptx-go holds itself to four properties that shape the public API:

1. **Two layers, one library.** `pptx` and `scene` are the only two public
   layers. `scene` composes `pptx`; nothing in `scene` reaches under it.
2. **Tokens, not literals.** Every visual property (color, typography, spacing,
   radius, elevation) flows through a `Theme` whose semantic tokens map to
   concrete values. Literals (`pptx.RGB`, `pptx.Pt`) are an escape hatch; the
   documented default is tokens. Swap the theme and the same input re-renders in
   the new visual language.
3. **OOXML by isolation.** Raw XML wire types never appear in the `pptx` or
   `scene` signatures. You work with Go domain types, not `encoding/xml`.
4. **No CGo, stdlib-only runtime.** The shipped artifact imports stdlib only.

## Engine, not product

pptx-go converts a builder call or a typed scene into PPTX, and nothing else
(D-026). Product behavior — render modes, legibility heuristics, markdown
ingestion, comments, "make this slide pretty" — lives in your code, not the
library. What pptx-go gives you is a set of *mechanisms* you drive: theme tokens,
asset resolution, font embedding, slide grouping, speaker notes. There is no
deck-wide "raster everything" switch and no opinion about who the audience is.
The library is the engine; you are the product.
