---
name: extend-the-icon-set
description: >-
  Use when a scene needs a custom icon beyond the curated set — a brand mark or
  glyph to drop into a Card or a Flow step. Teaches how to register a
  caller-supplied single-path SVG with scene.WithIconExtension (per-render), the
  exact SVG translator constraints icons must satisfy (D-040), and how to
  validate early with scene.ValidateIcon. Reach for it whenever a Card.Icon or
  FlowStep.Icon name isn't in the curated set.
---

# Extend the icon set

## Overview

Icons in pptx-go are a **closed-name curated set plus a per-render extension
seam** (RFC §14.1/§14.4, D-005, D-040). A node references an icon by **name**
(`Card.Icon`, `FlowStep.Icon`) — never by bytes. The name must resolve to a
curated icon or one you register for that render, or Stage-1 validation fails.

Icons render as **native custom geometry** (`<a:custGeom>`), not as a `pic`
image. The translator turns a single-path SVG into a DrawingML path, and the
shape is filled with the **accent token by default** (P2) — the SVG's own fill
color is discarded, so a theme swap re-tints every icon. There is no
rasterization and no embedded image (contrast with `register-an-asset`, which is
the seam for byte-backed nodes like Image/Chart/CodeBlock).

The extension is **per-render**, not global: `scene.WithIconExtension` registers
the icon for one `Render` call only, and registering a curated name overrides it
for that render alone. Concurrent renders with different extensions do not
interfere.

## The curated set

`scene/icons.Curated().Names()` returns the shipped names (sorted). The set is
seeded from the embedded `assets/icons/*.svg`:

```
arrow-down  arrow-left  arrow-right  arrow-up
chevron-down  chevron-left  chevron-right  chevron-up
check  circle  diamond  dot  minus  plus  square  star  triangle  x
```

If you need anything outside this list, register it (below). Don't invent a
name and hope — an unknown name is a Stage-1 error.

## Registering a custom icon

`scene.WithIconExtension(name string, svg []byte) scene.RenderOption` registers
a caller SVG under `name` for this render:

```go
stats, err := scene.Render(pres, sc, scene.WithIconExtension("brand-mark", svg))
```

The SVG is validated when the option is applied — an SVG outside the translator
subset fails the render with a Stage-1 error (at registration, not at compose).
A blank name or nil SVG is ignored. Passing a curated name overrides that icon
for this render only.

## SVG constraints

The translator (`internal/render/svgpath.go`) accepts a deliberately small SVG
subset. An SVG must be:

- **Exactly one `<path>` element.** No `<circle>`, `<rect>`, `<line>`,
  `<polyline>`, `<polygon>`, or `<ellipse>` — author those shapes as a path.
  Zero paths or more than one path is rejected.
- **Solid-filled.** `fill="none"` is rejected (the path must be filled), and a
  `fill="url(#…)"` gradient/pattern reference is rejected. `<linearGradient>`,
  `<radialGradient>`, and `<pattern>` elements are rejected. The fill *color*
  itself is ignored — icons are accent-tinted at draw time.
- **A `viewBox`** (or usable positive `width`/`height`) so coordinates have a
  window.

The path `d` data may use only these commands, **absolute or relative**:

```
M / m   moveto
L / l   lineto
H / h   horizontal lineto
V / v   vertical lineto
C / c   cubic Bézier
S / s   smooth cubic (reflects the previous control point)
Q / q   quadratic Bézier
T / t   smooth quadratic
Z / z   closepath
```

**Elliptical-arc commands `A` / `a` are NOT supported.** Author rounded curves
as Béziers instead. Any other command letter, or a malformed number, is
rejected.

## Validating early

`scene.ValidateIcon(svg []byte) error` (a re-export of `pptx.ValidateIcon`)
runs the exact same translator check without drawing anything. Call it at your
own registration point — when a user uploads an icon, when you load it from
disk — so a bad SVG fails fast with a clear error instead of surfacing at
render:

```go
if err := scene.ValidateIcon(svg); err != nil {
    return fmt.Errorf("icon rejected: %w", err) // e.g. "...elliptical-arc command \"A\" is not supported..."
}
```

`WithIconExtension` runs this check too, so validation is not optional — calling
`ValidateIcon` yourself just moves the failure earlier.

## Using an icon

Both the icon-bearing nodes take a **closed-name string**, Stage-1 validated:

```go
// In a Card:
scene.Card{
    Header: "Custom",
    Icon:   "brand-mark", // curated name or a WithIconExtension name
    Body:   []scene.SlideNode{ /* ... */ },
}

// In a Flow step:
scene.Flow{
    Steps: []scene.FlowStep{
        {Label: scene.RichText{{Text: "Ship"}}, Icon: "brand-mark"},
    },
}
```

The icon's zero value (empty string) renders the node without an icon. An icon
name that resolves to neither a curated nor a registered icon fails
`ValidateScene` / `Render` before any slide composes.

## Complete example

A runnable program lives at `examples/extend-the-icon-set/main.go`. It defines
a valid single-path triangle SVG, confirms `scene.ValidateIcon` accepts it,
renders a Card and a Flow step that use it via `WithIconExtension`, prints the
stats and warning count (0), and then shows `ValidateIcon` **rejecting** an SVG
that uses an elliptical-arc `A` command. Run it:

```bash
go run ./examples/extend-the-icon-set
```

## What the caller owns

You own **the SVG**. pptx-go gives you the translator and the accent fill, not a
drawing tool:

- Produce (or accept) a single `<path>` with a solid fill and a `viewBox`.
- Convert any non-path primitive (circle, rect, polygon) and any elliptical arc
  into path commands / Béziers before registering.
- Keep the bytes stable for a given name so re-renders stay byte-identical
  (Render is deterministic).

## See also

- `compose-a-scene` — building the Scene, Cards, and Flows the icon lands in.
- `register-an-asset` — the *other* extension seam, for byte-backed nodes
  (Image / Chart / CodeBlock) that render as a `pic` instead of native geometry.
