---
name: load-a-brand-template
description: Start a new deck from an existing .pptx brand kit so it inherits the brand's theme and slide masters/layouts. Use when a design team has supplied a branded .pptx (logo, colors, fonts, named layouts) and you want every new deck to adopt that look instead of the built-in default theme. Covers opening the brand file, seeding a deck with pptx.FromTemplate, and selecting layouts by name.
---

# Load a brand template

## Overview

A *brand kit* is an ordinary `.pptx` that carries a populated theme plus one or
more slide masters and their named layouts — the file a design team hands you so
decks look on-brand. This skill ingests that file so a new deck adopts its theme
and masters/layouts, then builds fresh slides on top. The library stays an
engine: it adopts what the brand file encodes and adds nothing opinionated
(D-026).

Use this when you have a brand `.pptx` on hand. To build a theme from scratch in
code instead, see `define-a-theme`; to start from one of the built-in templates,
see `scaffold-a-presentation`.

## How brand ingestion works

1. **Open the brand deck.** `pptx.NewFromFile` / `pptx.NewFromBytes` parse the
   `.pptx`. On open, the deck's theme (`theme1.xml`) and its master/layout
   registry are extracted, so the opened deck can serve as a brand kit
   (RFC §13.1).
2. **Seed a new deck from it.** `pptx.New(pptx.FromTemplate(brand))` clones the
   brand's package — theme, masters, layouts, and the auxiliary parts PowerPoint
   expects — and strips any slides, so the new deck starts slide-free with the
   brand's look. The brand presentation is **cloned, not retained**: it is not
   mutated, and you can close it right after `New` returns (D-037).
3. **Build on it.** `deck.AddSlide("Layout Name")` resolves the name against the
   adopted registry; new slides inherit the brand's theme and master.

Cloning the brand's already-valid relationship graph is what keeps ingestion
free of the "PowerPoint needs to repair this file" class of bug — ingestion
never hand-rewires masters.

## API

```go
// Open a brand kit (extracts its theme + master/layout registry on open).
func pptx.NewFromFile(path string, opts ...pptx.Option) (*pptx.Presentation, error)
func pptx.NewFromBytes(data []byte, opts ...pptx.Option) (*pptx.Presentation, error)

// Seed a new, slide-free deck from the brand kit. FromTemplate is an Option.
// A nil brand is ignored (the deck falls back to the default scaffold).
func pptx.FromTemplate(brand *pptx.Presentation) pptx.Option
deck := pptx.New(pptx.FromTemplate(brand))

// The adopted theme is the brand's; override it explicitly if you must.
deck.Theme() *pptx.Theme               // active theme (brand's, after ingestion)
pptx.New(pptx.FromTemplate(brand), pptx.WithTheme(custom)) // override
```

For starting from a **built-in** template instead of a brand file, use the named
templates: `pptx.NewWithTemplate(name pptx.TemplateType)` with
`pptx.TemplateBlank`, `pptx.TemplateDefault`, `pptx.TemplateWide`, or
`pptx.TemplateStandard`.

## Selecting layouts by name

The adopted registry is read-only and OOXML-free. Discover names, then select:

```go
for _, m := range deck.Masters() {          // []*pptx.Master
    fmt.Println("master:", m.Name())
    for _, l := range m.Layouts() {         // []*pptx.Layout
        fmt.Println("  layout:", l.Name())
    }
}

if deck.HasLayout("Title and Content") {    // resolves across all masters
    deck.AddSlide("Title and Content")      // named layout from the registry
} else {
    deck.AddSlide()                         // blank fallback
}

// A master can resolve a single layout too:
if l, ok := deck.Masters()[0].Layout("Section Header"); ok {
    _ = l.Name()
}
```

`AddSlide(name)` falls back to a blank layout when the name does not resolve, so
`HasLayout` is the way to know up front whether a selection will land. Layout
names come from the brand file (the PowerPoint layout-picker names); a layout
authored without a name is reachable only by iterating the registry.

## Complete example

See `examples/load-a-brand-template/main.go` for a runnable version. It
synthesizes a brand kit in code (since there is no file on disk), then ingests
it:

```go
brand, err := pptx.NewFromFile("acme-brand.pptx") // your design team's .pptx
if err != nil {
    log.Fatal(err)
}
defer brand.Close()

deck := pptx.New(pptx.FromTemplate(brand)) // theme + masters/layouts adopted

if deck.HasLayout("Title Slide") {
    deck.AddSlide("Title Slide")
} else {
    deck.AddSlide()
}

fmt.Println("masters adopted:", len(deck.Masters()))
fmt.Println("adopted theme:  ", deck.Theme().Name)

if err := deck.Save("on-brand.pptx"); err != nil {
    log.Fatal(err)
}
```

**A caveat worth knowing.** A brand authored *in code* with `WithTheme` does not
yet persist its custom color/font tokens into `theme1.xml` (token emission to
`theme1.xml` is pending). So a deck you build in code, save, and re-open adopts
the default scaffold theme, not the in-memory `WithTheme` values. A brand
`.pptx` designed in PowerPoint carries a real `theme1.xml` and is adopted in
full — which is the intended path for this skill. The runnable example prints
its adopted theme so this is visible.

## What the caller owns

- **The brand `.pptx` itself.** pptx-go ingests it; it does not author brand
  assets, validate that the file is "on-brand", or fetch it from anywhere. You
  supply the bytes or path.
- **Choosing layouts.** The library resolves a layout name against the registry
  or falls back to blank; deciding *which* layout each slide should use is your
  call.
- **Theme overrides.** The brand's theme is adopted by default; pass
  `WithTheme` if you deliberately want to render the brand's masters/layouts
  under a different theme.

## See also

- `define-a-theme` — build a `Theme` from semantic tokens in code (no brand
  file).
- `scaffold-a-presentation` — start a deck from scratch or a built-in template.
- `compose-a-scene` — render a typed `scene.Scene`; pair with the scene
  renderer's `WithLayoutMap` to map scene slides onto the brand's named layouts
  (use `HasLayout` to confirm the names resolve first).
