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
| Typography | `TypeRole` | `FontSpec` (family, size pt, weight, italic) |
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

### Spacing / Radius / Elevation

- Spacing: `SpaceXS`, `SpaceSM`, `SpaceMD`, `SpaceLG`, `SpaceXL`, `Space2XL`.
- Radius: `RadiusNone`, `RadiusSM`, `RadiusMD`, `RadiusLG`, `RadiusFull`.
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

## Font embedding (mechanism, no default — D-019)

A theme references font *names*; PowerPoint renders them only if installed
or embedded. pptx-go embeds on demand and never automatically:

```go
pres.SetFontSource(src)              // caller-injected FontSource
pres.EmbedFont("Inter", "bold", 700) // explicit, per face
```

`EmbedFont` writes a `*.fntdata` part, relates it to `presentation.xml`, and
records it in `<p:embeddedFontLst>`. With no `EmbedFont` call, nothing is
embedded. Subsetting (embed only used glyphs) is V1.x.

> The lazy `Color` interface and the `pptx.TokenColor(role)` / `pptx.RGB(...)`
> builder constructors arrive with the builder spine (D-030); until then the
> resolver (`Theme.Resolve*`) returns concrete values.
