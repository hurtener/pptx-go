# Themes & tokens

The `Theme` is pptx-go's single source of visual truth. Every visual property —
color, text color, typography, spacing, radius, elevation — flows through a
semantic *token* that the theme maps to a concrete value. Author against tokens
and a theme swap re-renders the same input in a new visual language; literals
(`pptx.RGB`, `pptx.Pt`) are an escape hatch, not the default path.

## Constructing a theme

```go
base := pptx.DefaultTheme() // a fresh copy you may mutate freely

brand := pptx.NewTheme(
	pptx.WithName("Acme"),
	pptx.WithAccent("2563EB"),                  // override the accent surface color
	pptx.WithFonts("Poppins", "Inter"),         // heading + body families
	pptx.WithDisplayFont("Playfair Display"),   // optional distinct display face for TypeDisplay
)

clone := brand.Clone() // deep copy; themes are reusable, mutate the copy
```

`DefaultTheme` returns the V1 default: a light surface, a neutral palette, and a
system font stack (Calibri / Calibri Light / Consolas) that renders legibly with
no font embedding. Apply a theme at construction with `pptx.WithTheme(brand)` or
later with `p.SetTheme(brand)`; read the active one with `p.Theme()`.

## Token taxonomy

The default values below are verified against the shipped default theme.

### Surface colors — `ColorRole`

| Role | Default |
| --- | --- |
| `ColorCanvas` | `FFFFFF` |
| `ColorSurface` | `FFFFFF` |
| `ColorSurfaceAlt` | `F1F3F5` |
| `ColorAccent` | `2563EB` |
| `ColorAccentAlt` | `7C3AED` |
| `ColorAccentWarm` | `EA580C` |
| `ColorSuccess` | `16A34A` |
| `ColorWarning` | `D97706` |
| `ColorError` | `DC2626` |
| `ColorInfo` | `0EA5E9` |

### Text colors — `TextColorRole`

| Role | Default |
| --- | --- |
| `TextPrimary` | `111827` |
| `TextSecondary` | `374151` |
| `TextTertiary` | `6B7280` |
| `TextInverse` | `FFFFFF` |
| `TextMuted` | `9CA3AF` |
| `TextAccent` | `2563EB` |
| `TextAccentAlt` | `7C3AED` |
| `TextSuccess` | `16A34A` |
| `TextWarning` | `D97706` |
| `TextError` | `DC2626` |

### Typography — `TypeRole`

Each role resolves to a font family, size (points), weight, an optional
letter-spacing (`FontSpec.Tracking`, points, signed) — positive opens glyphs
apart (wide-tracked eyebrows), negative tightens (display headlines); `0` emits
nothing and a `RunStyle.Tracking` overrides it per run — and an optional
line-height (`FontSpec.LineHeight`, percent of single; tight display ~100–105,
body ~120–135). Role `LineHeight` is applied by the **scene** layer to a node's
paragraphs; a direct `pptx` builder user sets it per paragraph via
`ParagraphOpts.LineHeight` (a paragraph has no type role). `0`/`100` emit nothing.
A role may also carry an average-advance metric for the wrap estimator
(`FontSpec.AvgCharWidth`, fraction of font size; `0` uses the built-in ~0.5 sans
factor) — it tunes content-aware height for serif/display faces and never renders.
A role may also declare a case
transform (`FontSpec.Case`: `CaseUpper`/`CaseSmallCaps`) rendered via `a:rPr/@cap`
— the run text stays original-case while the display is cased; pairs with
tracking for tracked-caps eyebrows. A role may also declare an ordered fallback
chain (`FontSpec.Fallback []string`): when a `FontSource` is registered and it
cannot resolve the role's primary family, the run's typeface is rewritten at save
to the first family in `[Family] + Fallback` the source resolves — a controlled
near-match rather than an arbitrary host default; empty (or no `FontSource`) is
byte-identical. Resolution is italic-aware: an italic emphasis run whose family
lacks an italic cut falls back to an italic-capable face instead of a faux-italic,
while upright runs keep the primary.

| Role | Family | Size | Weight |
| --- | --- | --- | --- |
| `TypeDisplay` | heading | 40 | 700 |
| `TypeH1` | heading | 32 | 700 |
| `TypeH2` | heading | 28 | 600 |
| `TypeH3` | heading | 24 | 600 |
| `TypeH4` | heading | 20 | 600 |
| `TypeH5` | heading | 16 | 600 |
| `TypeBody` | body | 14 | 400 |
| `TypeBodySmall` | body | 12 | 400 |
| `TypeCaption` | body | 10 | 400 |
| `TypeMono` | mono | 13 | 400 |
| `TypeCode` | mono | 12 | 400 |

The default families are `Calibri Light` (heading), `Calibri` (body), and
`Consolas` (mono).

### Spacing — `SpaceRole`

| Role | Default |
| --- | --- |
| `SpaceXS` | 2 pt |
| `SpaceSM` | 4 pt |
| `SpaceMD` | 8 pt |
| `SpaceLG` | 16 pt |
| `SpaceXL` | 24 pt |
| `Space2XL` | 40 pt |

### Radius — `RadiusRole`

| Role | Default |
| --- | --- |
| `RadiusNone` | 0 |
| `RadiusSM` | 2 pt |
| `RadiusMD` | 6 pt |
| `RadiusLG` | 12 pt |
| `RadiusFull` | effectively pill-shaped at slide scale |

### Elevation — `ElevationRole`

| Role | Default |
| --- | --- |
| `ElevationFlat` | no shadow |
| `ElevationRaised` | blur 4 pt, offset-y 1 pt, black @ 25% |
| `ElevationElevated` | blur 12 pt, offset-y 4 pt, black @ 35% |

## Using tokens

Colors are constructed through the sealed `Color` interface (D-033):

```go
pptx.TokenColor(pptx.ColorAccent)        // surface token, resolves at apply time
pptx.TokenTextColor(pptx.TextPrimary)    // text token
pptx.RGB("2563EB")                       // literal — the escape hatch
pptx.RGBA("2563EB", 50000)               // literal with OOXML alpha (0..100000)
```

`pptx.RGB` is both the theme's color value type and a literal `Color`, so
`pptx.RGB("2563EB")` is usable wherever a color is expected. The token
constructors carry a role and resolve against the active theme when applied.

## Resolution is apply-time

Token resolution happens **when the builder call runs**, not at save time
(D-033). Set the theme *before* authoring the content that should pick it up:

```go
p := pptx.New(pptx.WithTheme(brand)) // theme first
s := p.AddSlide()
// every shape/run added now resolves its tokens against `brand`
s.AddShape(pptx.ShapeRect, box, pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
```

If you author content, then swap the theme, already-authored shapes keep the
colors they resolved to. Re-rendering the same input under a different active
theme is the supported theme-swap path — re-run the builder calls (or
`scene.Render`) with the new theme.

## Font embedding

pptx-go provides the mechanism to embed fonts; whether to embed is your
distribution decision. Register a `FontSource` and call `EmbedFont` per face:

```go
p := pptx.New(pptx.WithFontSource(mySource))
if err := p.EmbedFont("Inter", "regular", 400); err != nil {
	log.Fatal(err)
}
```

## Known limitation: in-code themes and round-trip

A `Theme` you set with `WithTheme` resolves tokens to concrete sRGB and font
sizes that are baked into each shape, so the **visuals** round-trip — a reopened
deck looks identical. However, the deck's `theme1.xml` part is still the static
scaffold theme: token-to-`theme1.xml` *emission* is a follow-up (D-033). The
practical consequence is that an in-code `WithTheme` brand does **not** yet
round-trip as a *theme object*: reopen the deck and `p.Theme()` reflects the
scaffold's theme, not the brand you set in code. The shape-level colors and
fonts are preserved; the recoverable `Theme` is not. We state this honestly
rather than imply full theme round-trip. A theme that lives in a loaded
`.pptx` template (via [`FromTemplate`](/guide/builder)) is read from that
template's `theme1.xml` and is recovered on open.
