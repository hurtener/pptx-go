# Theme & token catalog

> The canonical token taxonomy and default-theme values for pptx-go (P2,
> RFC §7, D-003). A new visual property added to the builder lands a token
> entry here in the same PR (`CLAUDE.md §6`, §20).

The `Theme` is the single source of visual truth at write time. Builder and
scene calls take **semantic tokens**; the resolver materializes the concrete
OOXML value against the active theme. A theme swap re-renders the same input
in the new visual language.

## Token roles

| Token type | Go type | Resolves to |
|---|---|---|
| Surface color | `ColorRole` | `RGB` (6-hex) |
| Text color | `TextColorRole` | `RGB` (6-hex) |
| Typography | `TypeRole` | `FontSpec` (family, size pt, weight, italic, tracking, line-height, case, avg-char-width, fallback) |
| Spacing | `SpaceRole` | `EMU` |
| Radius | `RadiusRole` | `EMU` |
| Elevation | `ElevationRole` | `Elevation` (blur, offset, color, alpha) |

### Surface colors (`ColorRole`)

`ColorCanvas`, `ColorSurface`, `ColorSurfaceAlt`, `ColorAccent`,
`ColorAccentAlt`, `ColorAccentWarm`, `ColorSuccess`, `ColorWarning`,
`ColorError`, `ColorInfo`.

### Text colors (`TextColorRole`)

`TextPrimary`, `TextSecondary`, `TextTertiary`, `TextInverse`, `TextMuted`,
`TextAccent`, `TextAccentAlt`, `TextSuccess`, `TextWarning`, `TextError`.

### Typography (`TypeRole`)

`TypeDisplay`, `TypeH1`–`TypeH5`, `TypeBody`, `TypeBodySmall`, `TypeCaption`,
`TypeMono`, `TypeCode`.

A rich-text `Run`'s typography comes from its `RunStyle.TypeRole` (size +
family). The theme carries three font-scheme faces: `HeadingFont` (display +
headings), `BodyFont`, and the optional `DisplayFont` (D-063) — when set,
`TypeDisplay` uses `DisplayFont` (the big editorial face) instead of
`HeadingFont`, so a brand can pair a serif display with a separate sans for
headings. `WithFonts(heading, body)` + `WithDisplayFont(family)` set them
(order-independent); omitting `DisplayFont` leaves `TypeDisplay` on `HeadingFont`
(byte-identical). **Inline code** (`RunStyle.Code = true`, D-013) is not a new token —
it composes existing ones: the run's family switches to `TypeMono` and a subtle
background tint is drawn from `ColorSurfaceAlt`. Swap either token and inline
code re-renders accordingly.

**Tracking** (letter-spacing, D-060): `FontSpec.Tracking` is a per-type-role
value in points (signed) — positive opens glyphs apart (wide-tracked eyebrows/
labels), negative tightens them (display headlines). It resolves as part of the
role's `FontSpec` and is emitted as the OOXML `a:rPr/@spc` attribute (1/100 pt);
an optional `RunStyle.Tracking *float64` overrides it per run. The zero value
emits nothing (byte-identical to an untracked run).

**Font scale** (shrink-to-fit, D-074): `RunStyle.FontScale` is a per-run
multiplier on the resolved type-role size — the run-level escape hatch the scene
shrink-to-fit (`AutoFit`) path uses. The role's size token stays the source of
truth (a theme swap re-skins the base, then this scales it), so it does not weaken
P2; there is no per-role `FontScale` token. The zero value (and 1) leaves the size
unchanged (byte-identical); a value in (0,1) emits the reduced `a:rPr/@sz`, which
round-trips via `Run.FontSize`. Quantized and floored deterministically by the
scene `AutoFit` mechanism — the engine never shrinks on its own.

**Line height** (leading, D-061): `FontSpec.LineHeight` is a per-type-role line
spacing as a percent of single (100 = single, 120 = 1.2×) — tight display
(~100–105), readable body (~120–135). The scene renderer applies a node's role
line-height to its paragraphs, emitted as OOXML `a:pPr/a:lnSpc/a:spcPct`
(1/1000 percent); `pptx.ParagraphOpts.LineHeight` is the builder-level override.
0 or 100 emit nothing (byte-identical). (Estimator-accuracy — feeding leading
into the wrapped-height model — is a later refinement; this token delivers the
visual leading.)

**Case** (case transform, D-062): `FontSpec.Case` is a per-type-role case
transform (`CaseNone` / `CaseUpper` / `CaseSmallCaps`) — pairs with tracking for
the canonical tracked-caps eyebrow. It is rendered via the OOXML `a:rPr/@cap`
attribute (`all` / `small`), so the run **text stays original-case** (and
round-trips) while the display is cased; an optional `RunStyle.Case *TextCase`
overrides per run. `CaseNone` emits nothing (byte-identical). The engine provides
the mechanism only — making the default caption role uppercase is the soul's
choice (D-026), not the engine default.

**Font fallback chain** (D-066): `FontSpec.Fallback []string` is a per-type-role
ordered list of substitute families. When a `FontSource` is registered and it
cannot resolve the role's primary `Family`, the write-time fallback pass rewrites
the run's single-valued `a:latin` typeface to the first family in `[Family]` +
`Fallback` the source can resolve — a controlled near-match instead of an
arbitrary host default. Empty (the zero value) and "no `FontSource`" are
byte-identical; resolution is deterministic and idempotent. The chain *contents*
are the soul's choice; the engine carries and resolves it. A type-scale config
input, not a persisted OOXML field (the *resolved* face round-trips via the run's
`a:latin`).

**Average char width** (estimator metric, D-064): `FontSpec.AvgCharWidth` is the
role face's average glyph advance as a fraction of font size, used **only by the
deterministic wrap/overflow estimator** (it never renders). A soul sets a
measured factor for its bundled face (serif/display faces advance differently
from the default sans); `0` uses the built-in `~0.5` sans fallback —
byte-identical. A layout-estimator input on the type scale, not a visual token.

### Spacing / Radius / Elevation

- Spacing: `SpaceXS`, `SpaceSM`, `SpaceMD`, `SpaceLG`, `SpaceXL`, `Space2XL`.
  A scene `Card` resolves its interior padding from these (`CardSize` →
  `SpaceSM/MD/XL`); the opt-in `Card.PaddingScale` (basis points, D-076) scales
  that resolved value to tighten a dense card, floored at the `SpaceXS` minimum —
  a token-bound density control, no literal.
- Radius: `RadiusNone`, `RadiusSM`, `RadiusMD`, `RadiusLG`, `RadiusFull`.
  Consumed by `Slide.AddShape(ShapeRoundRect, box, WithRadius(role))`: the
  absolute radius token resolves against the active theme and is converted to
  the OOXML `roundRect` adjust (a fraction of the shorter side, capped at the
  50% full-capsule). `RadiusFull` yields a pill; the option is ignored on
  non-`roundRect` geometries.
- Elevation: `ElevationFlat`, `ElevationRaised`, `ElevationElevated`.

## Default theme

The V1 default (`pptx.DefaultTheme()`) is a light surface, a neutral
palette, and a system font stack that renders every node legibly with no
font embedding (RFC §7.5). It is emitted to `templates/_default-theme.pptx`
(regenerate with `go run ./_gen/gentheme`).

| Role | Value | | Role | Value |
|---|---|---|---|---|
| `ColorCanvas` | `FFFFFF` | | `ColorAccent` | `2563EB` |
| `ColorSurface` | `FFFFFF` | | `ColorAccentAlt` | `7C3AED` |
| `ColorSurfaceAlt` | `F1F3F5` | | `ColorAccentWarm` | `EA580C` |
| `ColorSuccess` | `16A34A` | | `ColorWarning` | `D97706` |
| `ColorError` | `DC2626` | | `ColorInfo` | `0EA5E9` |
| `TextPrimary` | `111827` | | `TextSecondary` | `374151` |
| `TextTertiary` | `6B7280` | | `TextInverse` | `FFFFFF` |

Fonts: heading **Calibri Light**, body **Calibri**, mono **Consolas**.
Spacing (pt): XS 2, SM 4, MD 8, LG 16, XL 24, 2XL 40.

## Theme ↔ theme1.xml mapping

PowerPoint's theme is a positional 12-color scheme plus a major/minor font
pair. The semantic palette maps onto it by convention — each OOXML slot has
one canonical semantic owner for writing; each semantic role reads back from
its slot. Roles without a slot keep their default after a load.

| OOXML slot | written from | read back into |
|---|---|---|
| `lt1` | `ColorSurface` | `ColorCanvas`, `ColorSurface`, `TextInverse` |
| `lt2` | `ColorSurfaceAlt` | `ColorSurfaceAlt` |
| `dk1` | `TextPrimary` | `TextPrimary` |
| `dk2` | `TextSecondary` | `TextSecondary` |
| `accent1`–`accent6` | `ColorAccent`, `ColorAccentAlt`, `ColorAccentWarm`, `ColorSuccess`, `ColorWarning`, `ColorError` | same |
| `hlink` | `ColorInfo` | `ColorInfo` |
| `folHlink` | `TextAccentAlt` | `TextAccentAlt` |
| major font | `Theme.HeadingFont` | heading typography |
| minor font | `Theme.BodyFont` | body typography |

## Font embedding (mechanism — D-019, D-065)

A theme references font *names*; PowerPoint renders them only if installed
or embedded. pptx-go embeds on demand from a caller-injected `FontSource`,
either one face at a time or — opt-in — every face a deck uses:

```go
pres.SetFontSource(src)              // caller-injected FontSource
pres.EmbedFont("Inter", "bold", 700) // explicit, per face

// or, automatically, at save (D-065):
pptx.New(pptx.WithFontSource(src), pptx.WithFontEmbedding())
```

`EmbedFont` writes a `*.fntdata` part, relates it to `presentation.xml`, and
records it in `<p:embeddedFontLst>`. `WithFontEmbedding()` runs a save-time
pass that walks every run, collects the distinct used faces (family, weight,
italic) in a stable sorted order, and `EmbedFont`s each — a no-op without a
`FontSource`, warn-don't-fail on a face that can't resolve, idempotent
against manual `EmbedFont`, and byte-identical when off. It is weight-aware
(D-068): it embeds the actual resolved weight file per OOXML bucket (the four
regular/bold/italic/boldItalic cuts), so a medium (500) role ships the medium
file, not a synthetic 400. PowerPoint exposes only four cuts per family, so the
engine embeds one file per bucket (a caller whose rasterizer needs finer cuts
calls `EmbedFont` directly). Subsetting (embed only used glyphs) is V1.x.

> The lazy `Color` interface and the `pptx.TokenColor(role)` / `pptx.RGB(...)`
> builder constructors arrive with the builder spine (D-030); until then the
> resolver (`Theme.Resolve*`) returns concrete values.

## Gradient / rotation / opacity (mechanisms, no new token — D-041)

Gradient fills (`pptx.LinearGradient` / `pptx.RadialGradient`), shape rotation
(`pptx.WithRotation`), and token opacity (`pptx.TokenColorAlpha`) are builder
**mechanisms**, not new theme tokens. They *consume* the existing color tokens:
a gradient stop's color is any `Color` (typically `TokenColor(role)` or
`TokenColorAlpha(role, alpha)`), so a theme swap re-renders a glow in the new
accent. No new token role is introduced; the token taxonomy above is unchanged.

## Elevation / shadow (mechanism, no new token — D-043)

The drop-shadow primitive `pptx.WithElevation(role)` / `pptx.WithShadow(e)` is
a builder **mechanism**, not a new theme token. It *consumes* the existing
`Elevation` token (the `ElevationRole` → `Elevation{Blur, OffsetX, OffsetY,
Color, Alpha}` already in the taxonomy above): `WithElevation(role)` resolves
the role against the active theme at `AddShape` time and emits
`<a:effectLst><a:outerShdw>`, so a theme swap re-renders the same shape with
the brand's elevation. `WithShadow(e)` is the literal escape hatch (P2 — the
documented path is `WithElevation`). A flat elevation
(`Elevation.IsFlat()`) emits no effect. No new token role is introduced.
