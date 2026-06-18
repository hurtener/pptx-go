---
layout: home

hero:
  name: pptx-go
  text: PowerPoint, in pure Go.
  tagline: Author and read PPTX decks with a theme-aware builder and an optional typed scene renderer. No CGo, stdlib-only runtime.
  actions:
    - theme: brand
      text: Get started
      link: /guide/getting-started
    - theme: alt
      text: Scene catalog
      link: /catalog/
    - theme: alt
      text: API reference
      link: /reference/pptx

features:
  - title: Two layers, one library
    details: A general-purpose builder (pptx) for imperative control, and an optional IR-driven renderer (scene) that composes the builder. Use either.
  - title: Tokens, not literals
    details: Color, typography, spacing, radius, and elevation flow through a Theme. Swap the theme and the same input re-renders in a new visual language.
  - title: No CGo, stdlib only
    details: The shipped artifact compiles CGO_ENABLED=0 and pulls zero third-party runtime dependencies. It cross-compiles anywhere Go does.
  - title: Round-trips losslessly
    details: Every shape, run, fill, line, table, and image pptx-go emits reopens into the same navigable model. Third-party decks open best-effort with warnings.
---

## What it is

**pptx-go** (`github.com/hurtener/pptx-go`, Apache-2.0) is a pure-Go library for
authoring and reading PowerPoint (PPTX / Open Office XML) files. It has two
public layers:

- **`pptx`** — the builder. Create slides, shapes, rich text, tables, images,
  speaker notes, and sections; then save, or read a deck back into a navigable
  model.
- **`scene`** — the renderer. Describe a deck declaratively as a typed `Scene`
  of nodes (`Hero`, `Card`, `Grid`, `Table`, `Flow`, …) and `Render` it onto a
  builder. `scene` composes `pptx`; nothing in `scene` reaches under it.

## Install

```bash
go get github.com/hurtener/pptx-go
```

## Hello, deck

```go
package main

import (
	"log"

	"github.com/hurtener/pptx-go/pptx"
)

func main() {
	p := pptx.New() // 16:9, default theme, valid out of the box
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(8), H: pptx.In(1.5)})
	tf.AddParagraph(pptx.ParagraphOpts{}).
		AddRun("Hello, pptx-go", pptx.RunStyle{TypeRole: pptx.TypeH1})

	if err := p.Save("hello.pptx"); err != nil {
		log.Fatal(err)
	}
}
```

Then read it back:

```go
re, _ := pptx.NewFromBytes(data)
for _, sl := range re.Slides() {
	for _, sh := range sl.Shapes() {
		// sh.Geometry(), sh.Fill(), sh.TextFrame(), sh.Table(), sh.Image() …
	}
}
```

See the [getting-started guide](/guide/getting-started) for the full tour.
