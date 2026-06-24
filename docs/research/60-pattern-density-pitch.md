# Brief 60 — Pattern density / pitch (R13.7)

> Informs Phase 77 (Wave 13). Engine req R13.7
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059). Builds on the
> role-colored recipe (Phase 73 / D-107) and the box-derived starfield
> (Phase 76 / D-110).

## 1. Motivating phase

`grid_dots` is a fixed 6×4 lattice and `noise_overlay` a fixed 12×8, regardless
of box size. Stretched full-bleed (the only way to texture a whole slide), 24
dots smear across 13 inches at a huge pitch — nothing like the fine, consistent
dot-grid texture the reference uses. Phase 77 derives the repeat count from a
caller **pitch** so a texture keeps a consistent visual spacing at any box size.

## 2. Subsystem / files

- `assets/ornaments/patterns.go` — `GridDots`, `NoiseOverlay`, `Starfield`.
- `scene/ornaments/registry.go` — the `Recipe` type.
- `scene/render_decoration.go` — the recipe call + the warn hook.
- `scene/nodes.go` — the `Decoration` struct.

## 3. Findings

- **The pitch must reach the recipe.** The recipe owns the lattice loop, so the
  pitch has to flow through the `Recipe` signature. Add a trailing `pitch
  pptx.EMU` (a second v0.x break this wave, after D-107's `role`) where `pitch ==
  0` means "the recipe's legacy fixed count" (byte-identical) and `pitch > 0`
  means box-derived `cols = box.W/pitch`, `rows = box.H/pitch`. A trailing
  positional param mirrors how D-107 added `role`; a params-struct refactor is
  cleaner long-term but a much larger churn — note it for V2.
- **Legacy byte-identity via `pitch == 0`.** `GridDots` keeps 6×4, `NoiseOverlay`
  12×8, `Starfield` its `starfieldPitch` default when `pitch == 0` — existing
  decks (which never set the new `Decoration.Pitch`) are byte-identical.
- **`Decoration.Pitch pptx.EMU`** carries the caller value into
  `render_decoration`, which passes it to the recipe. The glow/bracket/chevron
  recipes take (and ignore) the param.
- **Cap + warning.** A tiny pitch over a full-bleed box could emit thousands of
  shapes. Cap each pattern recipe at a shared `patternMaxDots`, and have
  `render_decoration` emit one `LayoutWarning` when `Pitch > 0` and the projected
  lattice count exceeds the cap (gated to the pattern presets so a glow with a
  stray `Pitch` doesn't false-warn). The recipe has no `r.warn` hook, so the
  warning lives in `render_decoration` (where `slideID` is in scope) and the cap
  lives in the recipe.
- **Determinism.** Integer division for `cols`/`rows`; the existing fixed
  per-cell offset idiom (no RNG). Worker-independent.
- **No OOXML / `restorenamespaces` change.** Same ellipse shapes.

## 4. Recommendations

- Extend `Recipe` with a trailing `pitch pptx.EMU`; thread it through all seven
  recipes (the three patterns honor it, the four others ignore it).
- Add `Decoration.Pitch pptx.EMU`; `render_decoration` passes it and warns on a
  cap-exceeding pattern projection.
- Cap the three pattern recipes at `patternMaxDots`.
- Tests: `grid_dots` at a 0.4in pitch over a 13in box yields ~32 columns (≫ 6);
  the same pitch on a small box yields proportionally fewer; a `Pitch == 0`
  (legacy) decoration is byte-identical; a tiny pitch warns; determinism.
- CHANGELOG note (2nd `OrnamentRecipe` break), THEME.md, glossary, skill,
  docs/site. D-111.

## 5. Open questions

- A params-struct `Recipe` (absorbing alpha/rotation/role/pitch) would future-
  proof against further additions — deferred to V2 (a one-time clean break).
- Density (dots per area) vs pitch (EMU spacing): pitch is the simpler, directly
  deterministic knob; density is `pitch = sqrt(area/count)` — caller-derivable.
