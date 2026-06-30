# Brief 82 — Named brand gradients (R8.5)

> Informs Phase 99 (Wave 15 — theme/soul engine bits). Engine side of the
> `both`-tagged requirement R8.5 (`DECKARD-PRODUCT-REQUIREMENTS.md`,
> HIGH · both; D-059 puts the named-gradient *mechanism* here, the soul
> *defining* the brand gradients in Deckard).

## 1. Motivating phase

Phase 99 lets a theme store **named brand gradients** (e.g. `hero`, `heroDark`)
as explicit stop lists with an angle and a linear/radial flag, and lets a slide
request one by name. Today the scene `Background` gradient takes raw `ColorRole`
stops resolved through the active theme, so a brand's signature deep-navy radial
hero wash cannot be named, stored, or reused — and on a dark slide the role stops
resolve through the (pinned or soul) dark theme rather than a fixed brand wash.
The gap: gradients are ad-hoc role pairs, not reusable brand tokens.

## 2. Subsystem / files

- `pptx/fill.go` — `GradientStop{Pos, Color}`, `LinearGradient`,
  `RadialGradient`, `Fill`. `GradientStop.Color` is a `pptx.Color` (RGB literal
  *or* token), which is exactly the "RGB|ColorRole" the spec wants.
- `pptx/theme.go` — `Theme` struct, `Clone()`, the option pattern.
- `scene/background.go` — `Background` struct + `BackgroundKind`.
- `scene/render.go` — `drawBackgroundFill` (the `BackgroundGradient` case +
  `backgroundGradientStopsFor`).

## 3. Findings

- **`pptx.GradientStop` already models "RGB|ColorRole".** Its `Color` field is a
  `pptx.Color` interface satisfied by both `RGB` literals (variant-independent,
  pins an exact brand hue) and `TokenColor(role)` (follows the active/dark
  theme). So a named gradient's stops are just `[]pptx.GradientStop` — no new
  stop type, and variant handling is automatic: a token stop re-resolves under a
  VariantDark theme, an RGB stop stays fixed. This satisfies "RGB stops let a
  soul pin exact brand hues independent of variant; ColorRole stops still follow
  the variant" with zero special-casing.
- **A named gradient owns its linear/radial choice.** The requirement's
  `GradientSpec` carries `Radial bool`, so one `Background.GradientName` field
  feeds either `pptx.LinearGradient(angle, stops…)` or
  `pptx.RadialGradient(stops…)` per the named spec — the kind stays
  `BackgroundGradient`, the spec decides the shape. This avoids a second
  background kind.
- **Resolution slots in front of the existing path, byte-identically.** In the
  `BackgroundGradient` case, `GradientName != ""` resolves the named spec from
  the active theme and supersedes `Stops` / the legacy 2-role `Gradient`. Empty
  `GradientName` runs the existing `backgroundGradientStopsFor` path unchanged
  (byte-identical). A name not found in the theme → `LayoutWarning` + skip
  (RFC §10.2 degrade, like an unresolved asset).
- **Validation reuses the stop invariants.** A named spec's stops get the same
  2..8-ascending-in-[0,1] check the multi-stop path uses; invalid → warn + skip.
  The check operates on `pptx.GradientStop` (Pos only — Color is opaque).
- **No theme1.xml slot.** theme1.xml has no gradient slot; a named-gradient map
  is consumed only by the scene background renderer and is never serialized —
  like `DarkColors`/`Accents`/`ColorPaper`, the *resolved* gradient fill
  round-trips (the `gradFill` element survives `pptx.Open`), the named map does
  not. `Clone()` must deep-copy the map (and each spec's stop slice).

## 4. Recommendations

- Add `pptx.GradientSpec{Stops []GradientStop; Angle int; Radial bool}` (in
  `pptx/fill.go`, beside the gradient primitives) + `Theme.Gradients
  map[string]GradientSpec` + a `WithGradient(name, spec)` option + a
  `Theme.Gradient(name) (GradientSpec, bool)` accessor. `Clone()` deep-copies the
  map and stop slices.
- Add `Background.GradientName string` (scene). In the `BackgroundGradient` case
  of `drawBackgroundFill`: if `GradientName != ""`, resolve the named spec; warn
  + skip on miss / invalid stops; else `spec.Radial` picks
  `RadialGradient(stops…)` vs `LinearGradient(spec.Angle, stops…)`. Empty name =
  the existing path (byte-identical).
- Document in THEME.md (a brand gradient is a visual property, P2) + glossary +
  the define-a-theme skill + docs/site + the compose-a-scene skill (the
  `Background.GradientName` field). No theme1.xml slot, no codec/restorenamespaces
  change (the gradient fill primitives already emit + round-trip).

## 5. Open questions

- Should `GradientName` also be honored under `BackgroundRadial`? No — the named
  spec's `Radial` flag already picks the shape, so a single `GradientName` under
  `BackgroundGradient` covers both; `BackgroundRadial` stays the role-based
  center-out path (back-compat). Note in the field doc.
- Should the named gradient round-trip through theme1.xml? No — no gradient slot
  exists; the resolved fill round-trips, the named map does not (consistent with
  `DarkColors`/`Accents`). V2 backlog only if a real read-back case appears.
- Defining the brand gradients (`hero`/`heroDark` from a brand source) is the
  soul's product half (D-059) — out of scope; the engine stores + resolves them.
