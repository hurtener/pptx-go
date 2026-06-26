---
name: define-a-theme
description: Define and customize a pptx-go Theme — the semantic token system (P2). Use when setting a deck's brand colors, typography, spacing, corner radii, or elevations; when choosing between token and literal colors; or when you want a single theme swap to re-skin the same builder/scene input. Covers DefaultTheme, NewTheme options, Clone-and-mutate, the full token taxonomy with default values, and the resolution model.
---

# Define a Theme

## Overview

A `*pptx.Theme` is pptx-go's single source of visual truth. It maps **semantic
tokens** — color, text color, typography, spacing, radius, elevation — to
concrete OOXML values at render time. Every builder call that takes a color, a
font, a radius, or an elevation takes a *token*; the active theme decides what
that token looks like. Set a theme on a `Presentation`, author with tokens, and
the whole deck speaks one visual language. Swap the theme and the same input
re-renders in the new language.

A theme is a plain, reusable value. `DefaultTheme()` and `Clone()` each return a
fresh deep copy you may mutate freely; the same `*Theme` is safe to share across
concurrently-built presentations as long as you do not mutate it while it is in
use.

## Why tokens (P2)

Tokens, not literals, are the **documented path**. Property by property —
color, typography, spacing, radius, elevation — the builder API accepts a
semantic token (`pptx.TokenColor(pptx.ColorAccent)`, `RunStyle{TypeRole:
pptx.TypeH1}`, `pptx.WithRadius(pptx.RadiusLG)`, `pptx.WithElevation(
pptx.ElevationRaised)`). Literals (`pptx.RGB("2563EB")`, `pptx.RGBA`,
`pptx.Pt(...)`) are an **escape hatch** for the rare one-off that has no
business in the theme.

The payoff is the **theme swap**: because a token resolves against the active
theme, the same builder or scene input re-skins when you render it under a
different theme. A literal is baked and ignores the theme — that is exactly why
it is the exception, not the rule.

## The token taxonomy

The complete V1 taxonomy, with the values `DefaultTheme()` ships. Authoritative
catalog: `docs/design/THEME.md`.

### Surface colors — `ColorRole` → `Theme.Colors.Surfaces[role]`

Fill colors for shapes and surfaces. Construct with `pptx.TokenColor(role)` (or
`pptx.TokenColorAlpha(role, alpha)` to dim it).

| Role | Default RGB | Meaning |
|---|---|---|
| `ColorCanvas` | `FFFFFF` | Page background |
| `ColorSurface` | `FFFFFF` | Card / panel surface |
| `ColorSurfaceAlt` | `F1F3F5` | Secondary surface |
| `ColorAccent` | `2563EB` | Primary brand accent |
| `ColorAccentAlt` | `7C3AED` | Secondary accent |
| `ColorAccentWarm` | `EA580C` | Warm accent |
| `ColorSuccess` | `16A34A` | Success state |
| `ColorWarning` | `D97706` | Warning state |
| `ColorError` | `DC2626` | Error state |
| `ColorInfo` | `0EA5E9` | Informational state |
| `ColorPaper` | `FFFFFF` | Off-white "paper" canvas — a tinted alternative to pure white; defaults to `ColorCanvas`, set via `WithPaper` |

`ColorPaper` is a faintly tinted off-white canvas distinct from pure white, for
a designed paper tone on content slides. It defaults to white (= `ColorCanvas`)
so it is invisible until you set it, e.g. `pptx.NewTheme(pptx.WithPaper(pptx.RGB("FAFAF8")))`,
then point a slide's background at it (`scene.Background{Kind: scene.BackgroundColor, Color: pptx.ColorPaper}`).
It has no theme1.xml slot, so a re-opened deck's theme reads it back at its
default — set it on the theme you author with (the resolved background color
itself round-trips on the slide).

### Text colors — `TextColorRole` → `Theme.Colors.Text[role]`

Colors for text runs. Construct with `pptx.TokenTextColor(role)`.

| Role | Default RGB | Meaning |
|---|---|---|
| `TextPrimary` | `111827` | Body / heading text |
| `TextSecondary` | `374151` | Secondary text |
| `TextTertiary` | `6B7280` | Tertiary text |
| `TextInverse` | `FFFFFF` | Text on a dark/accent surface |
| `TextMuted` | `9CA3AF` | Muted / disabled text |
| `TextAccent` | `2563EB` | Accent-colored text |
| `TextAccentAlt` | `7C3AED` | Secondary accent text |
| `TextSuccess` | `16A34A` | Success text |
| `TextWarning` | `D97706` | Warning text |
| `TextError` | `DC2626` | Error text |

### Typography — `TypeRole` → `Theme.Typography[role]` (`FontSpec`)

A `FontSpec` is `{Family string; Size float64; Weight int; Italic bool; Tracking
float64; LineHeight float64; Case TextCase; AvgCharWidth float64; Fallback
[]string}`. Weight
is 100–900 (400 = regular, 700 = bold); `FontSpec.Bold()` reports `Weight >=
600`. `Tracking` is letter-spacing in points (signed): positive opens glyphs
apart (wide-tracked eyebrows), negative tightens (display headlines), emitted as
OOXML `a:rPr/@spc`; `0` emits nothing, and a `RunStyle.Tracking *float64`
overrides it per run. `LineHeight` is line spacing as a percent of single
(tight display ~100–105, body ~120–135), applied to a node's paragraphs and
emitted as `a:pPr/a:lnSpc/a:spcPct`; `0`/`100` emit nothing, and
`ParagraphOpts.LineHeight` overrides per paragraph. `FontSpec.Case`
(`CaseNone`/`CaseUpper`/`CaseSmallCaps`) is a case transform rendered via
`a:rPr/@cap` — the run text stays original-case while the display is cased
(`RunStyle.Case` overrides per run); pairs with `Tracking` for tracked-caps
eyebrows. `AvgCharWidth` is the face's average glyph advance (fraction of size)
for the wrap/overflow estimator only — set a measured factor for a serif/display
face; `0` uses the `~0.5` sans fallback (it never renders). `Fallback` is an
ordered substitute chain: when a `FontSource` is registered and it cannot resolve
the role's primary `Family`, the run's typeface is rewritten at save to the first
family in `[Family] + Fallback` the source resolves (a controlled near-match, not
a host default); empty or no `FontSource` is byte-identical. Resolution is
italic-aware — an italic run whose family lacks an italic cut falls back to an
italic-capable face (not a faux-italic), while upright runs keep the primary.
Select a role via
`RunStyle{TypeRole: role}`. Defaults
below use heading font `Calibri Light`, body font `Calibri`, mono font `Consolas`.

| Role | Family | Size (pt) | Weight |
|---|---|---|---|
| `TypeDisplay` | Calibri Light | 40 | 700 |
| `TypeH1` | Calibri Light | 32 | 700 |
| `TypeH2` | Calibri Light | 28 | 600 |
| `TypeH3` | Calibri Light | 24 | 600 |
| `TypeH4` | Calibri Light | 20 | 600 |
| `TypeH5` | Calibri Light | 16 | 600 |
| `TypeBody` | Calibri | 14 | 400 |
| `TypeBodySmall` | Calibri | 12 | 400 |
| `TypeCaption` | Calibri | 10 | 400 |
| `TypeMono` | Consolas | 13 | 400 |
| `TypeCode` | Consolas | 12 | 400 |

### Spacing — `SpaceRole` → `Theme.Spacing[role]` (`EMU`)

| Role | Default |
|---|---|
| `SpaceXS` | `Pt(2)` |
| `SpaceSM` | `Pt(4)` |
| `SpaceMD` | `Pt(8)` |
| `SpaceLG` | `Pt(16)` |
| `SpaceXL` | `Pt(24)` |
| `Space2XL` | `Pt(40)` |

### Radii — `RadiusRole` → `Theme.Radii[role]` (`EMU`)

Applied to `ShapeRoundRect` via `pptx.WithRadius(role)`.

| Role | Default |
|---|---|
| `RadiusNone` | `0` |
| `RadiusSM` | `Pt(2)` |
| `RadiusMD` | `Pt(6)` |
| `RadiusLG` | `Pt(12)` |
| `RadiusFull` | `Pt(7200)` (pill at slide scale) |

### Elevations — `ElevationRole` → `Theme.Elevations[role]` (`Elevation`)

An `Elevation` is `{Blur, OffsetX, OffsetY EMU; Color RGB; Alpha int}` (Alpha is
OOXML 0–100000). Applied to a shape via `pptx.WithElevation(role)`. A flat
elevation (`IsFlat()`) emits no shadow.

| Role | Default |
|---|---|
| `ElevationFlat` | `{}` (no shadow) |
| `ElevationRaised` | `{Blur: Pt(4), OffsetY: Pt(1), Color: "000000", Alpha: 25000}` |
| `ElevationElevated` | `{Blur: Pt(12), OffsetY: Pt(4), Color: "000000", Alpha: 35000}` |

## Creating a Theme

There are three constructors and one mutator, in increasing order of control.

**1. `DefaultTheme()`** — a complete, legible light theme with no font
embedding. Every `pptx.New()` deck already uses it; call it directly only when
you want a base to clone.

```go
base := pptx.DefaultTheme()
```

**2. `NewTheme(opts ...ThemeOption)`** — starts from `DefaultTheme()` and
applies functional options. Unset roles keep their defaults. The options:

- `pptx.WithName(string)` — sets `Theme.Name`.
- `pptx.WithAccent(pptx.RGB)` — overrides `ColorAccent`.
- `pptx.WithPaper(pptx.RGB)` — sets `ColorPaper`, the off-white "paper" canvas
  tint (defaults to white = `ColorCanvas`). Pair it with a `BackgroundColor`
  background to give content slides a designed paper tone.
- `pptx.WithFonts(heading, body string)` — sets `HeadingFont`/`BodyFont` **and**
  rewrites the `Typography` families: heading roles (`TypeDisplay`–`TypeH5`) get
  `heading`, mono roles (`TypeMono`, `TypeCode`) are left on the mono face, the
  rest get `body`.
- `pptx.WithDisplayFont(family string)` — sets `Theme.DisplayFont`, a distinct
  face for the `TypeDisplay` role (the big editorial face), independent of
  `heading`. Order-independent with `WithFonts`; omitting it leaves `TypeDisplay`
  on `HeadingFont` (byte-identical). Lets a brand pair a serif display with a
  separate sans heading.
- `pptx.WithDarkSurface(role pptx.ColorRole, c pptx.RGB)` /
  `pptx.WithDarkText(role pptx.TextColorRole, c pptx.RGB)` — set a brand's
  **dark palette** (`Theme.DarkColors`): the colors a `scene.VariantDark` slide
  resolves to, instead of the engine's pinned neutral gray. Each call overrides
  one surface/text role for the dark variant only; unset roles keep the pinned
  dark default. Setting none leaves the pinned gray (byte-identical). Composable
  and order-independent. See **Dark palette** below.

```go
theme := pptx.NewTheme(
    pptx.WithName("Acme"),
    pptx.WithAccent(pptx.RGB("DB2777")),
    pptx.WithFonts("Georgia", "Verdana"),
)
```

**3. `Clone()` + direct struct edits** — full control. `Clone()` is a deep copy
(every map is reallocated), so mutating the clone never affects the source. Edit
any token map directly:

```go
t := pptx.DefaultTheme().Clone()
t.Name = "Acme Dark"
t.Colors.Surfaces[pptx.ColorCanvas]  = pptx.RGB("0B1220")
t.Colors.Text[pptx.TextPrimary]      = pptx.RGB("E5E7EB")
h1 := t.Typography[pptx.TypeH1]
h1.Size = 36
t.Typography[pptx.TypeH1] = h1
```

You can also build a `pptx.Theme{...}` struct literal field-by-field, but
cloning the default and editing the few roles you care about is safer — you
inherit complete maps and cannot leave a role unset.

**Apply the theme to a presentation** at construction or any time before
authoring the content that should use it:

```go
pres := pptx.New(pptx.WithTheme(theme)) // at construction
// or
pres.SetTheme(theme)                    // later; a nil theme is ignored
```

## Dark palette (`scene.VariantDark`)

When a scene slide sets `Variant: scene.VariantDark`, the renderer derives a
dark theme for that slide. By default it uses a pinned neutral-gray palette
(`#111827`/`#1F2937`/`#374151` surfaces, `#F9FAFB`/`#E5E7EB`/`#9CA3AF` text). To
render a brand's *own* dark side (e.g. deep navy) instead, give the theme a
**dark palette** — `Theme.DarkColors`, set with `WithDarkSurface` /
`WithDarkText`:

```go
theme := pptx.NewTheme(
    pptx.WithName("Acme"),
    pptx.WithDarkSurface(pptx.ColorCanvas, pptx.RGB("0A0E1A")),  // brand navy canvas
    pptx.WithDarkSurface(pptx.ColorSurface, pptx.RGB("14182B")), // brand dark card
    pptx.WithDarkText(pptx.TextPrimary, pptx.RGB("F4F6FF")),     // brand light text
)
```

Each call overrides one role **for the dark variant only**; roles you don't set
keep the pinned dark default. The renderer writes the pinned grays first, then
overlays your dark palette role-by-role — so you can override as much or as
little as you want. Any surface or text role is overridable (including accent and
semantic roles). A theme that sets **no** dark palette renders the pinned gray —
byte-identical to the engine default.

The dark palette has **no theme1.xml slot**: it is consumed only to derive the
dark variant and is never serialized. The resolved dark colors a slide renders
with still round-trip (and are reported per-slide via `scene.Render`'s
`Stats.Colors`, so you can verify exactly what each dark slide resolved to). You
can also build the whole palette at once via the struct:
`theme.DarkColors = &pptx.DarkPalette{Surfaces: …, Text: …}`.

## Token resolution model

Resolution is **deterministic** and happens at **apply time** — the moment you
call a builder method like `AddShape` or `AddRun`, the token is resolved against
the presentation's *active* theme and the concrete value is written into the
slide (D-033). The same token against the same theme always yields the same
value; that determinism is what the theme-swap guarantee rests on.

Two practical consequences:

- **Set the theme before you author.** Because resolution is at apply time, a
  shape added under theme A keeps theme A's colors even if you later call
  `SetTheme(B)`. To re-skin the same content under a new theme, run the same
  builder calls under the new active theme (see the example) — that *is* the
  theme swap.
- **Resolution never panics.** A token a theme leaves unset falls back to a safe
  neutral rather than failing across the public API (surfaces → `FFFFFF`, text →
  `000000`, type → `Calibri 14/400`, spacing/radius → `0`, elevation → flat). A
  `Clone()`'d default theme has every role set, so fallbacks only bite a
  hand-built partial theme.

The resolver methods are public if you need a value directly:
`(*Theme).ResolveColor(ColorRole) RGB`,
`ResolveTextColor(TextColorRole) RGB`,
`ResolveType(TypeRole) FontSpec`,
`ResolveSpace(SpaceRole) EMU`,
`ResolveRadius(RadiusRole) EMU`,
`ResolveElevation(ElevationRole) Elevation`.

**Color constructors** at a glance:

| Constructor | Result |
|---|---|
| `pptx.TokenColor(role)` | Surface token, opaque |
| `pptx.TokenColorAlpha(role, alpha)` | Surface token at OOXML alpha 0–100000 |
| `pptx.TokenTextColor(role)` | Text token, opaque |
| `pptx.RGB("2563EB")` | Literal opaque color (escape hatch) |
| `pptx.RGBA("2563EB", alpha)` | Literal color at OOXML alpha (escape hatch) |

`Color` is a sealed interface (D-033): the only colors are the ones the package
defines, so you cannot hand the codec a color it can't emit.

## Complete example

A runnable program lives at `examples/define-a-theme/main.go`. It builds a theme
two ways (`NewTheme` with options, and `Clone()` + direct mutation), renders an
identical token-filled card and token-typed heading under each theme, and shows
that the two outputs differ — the theme swap re-skinning the same builder input.
Run it with:

```sh
go run ./examples/define-a-theme
```

## What the caller owns

pptx-go owns the *mechanism* (the token taxonomy and the resolver); the **brand
identity is yours**. You decide the accent palette, the type scale and font
faces, the spacing rhythm, the corner language, and the elevation depth — by
constructing the `Theme` that encodes them. The library does not ship opinions
about what a brand should look like beyond a neutral, legible default; it
guarantees only that whatever tokens you define resolve consistently and that
one theme swap re-skins everything authored through tokens.

## See also

- **scaffold-a-presentation** — create a `Presentation` and attach the theme
  with `WithTheme` / `SetTheme`.
- **load-a-brand-template** — seed a deck's theme, masters, and layouts from an
  existing brand `.pptx` instead of building a `Theme` by hand.
- **compose-a-scene** — drive the same token system from the scene (Layer 2) IR.
