# Brief 56 — Decoration color role (R13.5)

> Informs Phase 73 (Wave 13). Engine req R13.5
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, CRITICAL · engine; D-059).

## 1. Motivating phase

Every curated ornament (`grid_dots`, `noise_overlay`, `radial_glow`,
`glow_ring`, `corner_bracket`, `chevron_arrow`) hard-codes `pptx.ColorAccent`.
The reference uses decoration in *many* hues — neutral-grey paper grain, a
white/pale starfield on dark slides, multi-color confetti — so with a single
forced accent Deckard cannot reproduce them and omits decoration entirely.
Phase 73 lets the caller pick the decoration color, defaulting to accent
(byte-identical).

## 2. Subsystem / files

- `assets/ornaments/{ornaments,glows,patterns,accents}.go` — the 6 curated
  recipes + the `accent(alpha)` helper.
- `scene/ornaments/registry.go` — the `Recipe` type + `Curated`/`With`/`Lookup`.
- `scene/render_decoration.go` — calls `recipe(ps, box, alpha, rotation)`.
- `scene/nodes.go` — the `Decoration` struct.
- `scene/scene.go` — `OrnamentRecipe = ornaments.Recipe` alias +
  `WithOrnamentExtension(name, recipe)` (public caller extension).

## 3. Findings

- **The recipe color is centralized.** Every recipe fills via either
  `accent(alpha)` (`SolidFill(TokenColorAlpha(ColorAccent, alpha))`) or, for
  glows, `TokenColorAlpha(ColorAccent, …)` stops. Threading one `role
  pptx.ColorRole` through the recipe replaces all of them.
- **The `Recipe` signature must change.** `Recipe func(sl, box, alpha,
  rotationDeg) int` → add a trailing `role pptx.ColorRole`. This is a public
  break: `ornaments.Recipe` is aliased as `scene.OrnamentRecipe` and callers
  register via `scene.WithOrnamentExtension`. Pre-V1 v0.x makes the break
  acceptable; the spec explicitly calls for it ("change the signature … to
  accept a ColorRole"). Document the break in the plan/decision.
- **`Decoration.Color` must be a pointer.** `pptx.ColorRole`'s zero value is
  `ColorCanvas` (a real color), not accent. To keep "unset = accent =
  byte-identical", use `Color *pptx.ColorRole` (nil = `ColorAccent`) — the
  established D-054 pointer pattern (HeaderFill/StatusDot, `Ribbon.Color`). This
  departs from the spec's literal "`Decoration.Color pptx.ColorRole` (zero =
  ColorAccent)" wording for the same reason D-054 did; note it as a §4.3
  deviation.
- **Byte-identity.** `render_decoration` resolves `role := ColorAccent; if
  v.Color != nil { role = *v.Color }`, then passes it. A nil `Color` →
  `ColorAccent` → every recipe fills exactly as today → byte-identical.
- **Determinism.** No geometry change; only the fill role. Pure, worker-safe.
- **No OOXML / `restorenamespaces` change.** Same `<a:solidFill>` /
  `<a:gradFill>` shapes, only the resolved `srgbClr` differs.
- **`Decoration.Palette` (multi-hue scatter) is deferred** to the starfield
  req (R13.6, a later phase) — R13.5 ships the single `Color` mechanism the
  scatter will cycle.

## 4. Recommendations

- Add `role pptx.ColorRole` as the trailing `Recipe` param; rename `accent` →
  `roleFill(role, alpha)`; thread `role` through all 6 recipes (solid fills and
  glow gradient stops).
- Add `Decoration.Color *pptx.ColorRole` (nil = `ColorAccent`); resolve it in
  `render_decoration` and pass to the recipe.
- Update the recipe unit tests (assets/ornaments) and the registry test to the
  new signature; add a scene black-box test: a `Color`-set decoration (the shipped
  test uses `ColorError` = `DC2626`, a surface role distinct from the accent
  `2563EB`) emits a different `srgbClr` than the accent default, and an unset
  `Color` is byte-identical to the pre-change build.
- THEME.md mechanism note, glossary, compose-a-scene skill, docs/site catalog.
  D-107. CHANGELOG note for the `OrnamentRecipe` signature break.

## 5. Open questions

- `Decoration.Palette []pptx.ColorRole` for multi-hue scatter → R13.6 starfield.
- Should caller extensions that ignore the role still compile? Yes — the param
  is positional; an extension simply takes (and may ignore) the role.
