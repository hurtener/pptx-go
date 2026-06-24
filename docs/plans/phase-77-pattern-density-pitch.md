# Phase 77 — pattern density / pitch

**Subsystem:** `assets/ornaments` (recipes) + `scene` (decoration)
**RFC sections:** §14.2 (ornaments), §10.1 (backward-compat), §10.2 (warn-don't-fail)
**Deps:** Phase 73 (role recipe, D-107), Phase 76 (starfield, D-110); brief 60.
**Status:** Done

---

## 1. Goal

Derive the pattern ornaments' repeat count from a caller pitch (EMU spacing) so a
texture keeps a consistent visual density at any box size — fixing the
"24 dots smeared across a full-bleed slide" gap — defaulting to today's fixed
counts (byte-identical).

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); a full-slide dot texture is
unusable at the fixed 6×4/12×8 counts. Engine req R13.7 (MED · engine; D-059).
Closes the box-derived-density story the starfield (R13.6) opened.

## 3. RFC sections implemented

- `RFC §14.2` — pattern recipes gain a pitch parameter.
- `RFC §10.1` — `pitch == 0` reproduces the legacy fixed counts exactly.
- `RFC §10.2` — a cap-exceeding pitch degrades to a `LayoutWarning`, never a
  file-size blowup or panic.

## 4. Brief findings incorporated

- `docs/research/60-pattern-density-pitch.md` — *"the pitch must reach the
  recipe (a 2nd v0.x break)"* → trailing `pitch pptx.EMU`, `0` = legacy.
- `60` — *"legacy byte-identity via `pitch == 0`"* → fixed counts when unset.
- `60` — *"cap in the recipe, warn in render_decoration"* → `patternMaxDots` +
  a projection warning gated to the pattern presets.
- `60` — *"params-struct refactor deferred to V2"* → minimal trailing param now.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-107` — the `role` recipe param (the precedent for a trailing positional add).
- `D-110` — the starfield's box-derived pitch (now caller-settable).
- `D-035` — deterministic, no RNG.
- `D-111` (new) — files the pitch param + the 2nd `Recipe` break.

## 7. Architecture

`Recipe` gains a trailing `pitch pptx.EMU`. The three pattern recipes
(`GridDots`, `NoiseOverlay`, `Starfield`) compute `cols = box.W/pitch`, `rows =
box.H/pitch` when `pitch > 0`, else their legacy fixed counts (6×4 / 12×8 /
`starfieldPitch`). Each caps at `patternMaxDots`. `Decoration.Pitch pptx.EMU`
flows through `render_decoration`, which also emits one `LayoutWarning` when
`Pitch > 0`, the preset is a pattern, and the projected lattice exceeds the cap.
The four non-pattern recipes take and ignore the param.

```text
Decoration{Preset: "grid_dots", Pitch: In(0.4), Bleed: true, Size: full}
  → GridDots(..., pitch=In(0.4)) → cols=13in/0.4in≈32, rows≈18 → a fine grid
Decoration{Preset: "grid_dots"}  (Pitch 0) → 6×4 (byte-identical)
```

## 8. Files added or changed

```text
assets/ornaments/patterns.go     # CHANGED — GridDots/NoiseOverlay/Starfield honor pitch; patternMaxDots cap
assets/ornaments/glows.go        # CHANGED — RadialGlow/GlowRing take (ignore) pitch
assets/ornaments/accents.go      # CHANGED — CornerBracket/ChevronArrow take (ignore) pitch
assets/ornaments/ornaments_test.go  # CHANGED — recipe type alias + call sites
scene/ornaments/registry.go      # CHANGED — Recipe signature + doc
scene/ornaments/registry_test.go # CHANGED — noopRecipe signature
scene/nodes.go                   # CHANGED — Decoration.Pitch field
scene/render_decoration.go       # CHANGED — pass pitch + cap-projection warning
scene/render_decoration_test.go  # CHANGED — pitch density + legacy byte-identical + cap warn + determinism
scripts/smoke/phase-77.sh        # NEW — phase smoke
docs/research/60-pattern-density-pitch.md  # NEW — brief
docs/research/INDEX.md           # CHANGED — registers brief 60
docs/plans/phase-77-pattern-density-pitch.md  # NEW — this plan
docs/plans/README.md             # CHANGED — Phase 77 detail
docs/design/THEME.md             # CHANGED — pattern pitch mechanism note
docs/glossary.md                 # CHANGED — pattern pitch term
docs/decisions.md                # CHANGED — adds D-111
CHANGELOG.md                     # CHANGED — 2nd OrnamentRecipe signature break
docs/site/reference/scene.md     # CHANGED — OrnamentRecipe signature + Decoration.Pitch
skills/compose-a-scene/SKILL.md  # CHANGED — Decoration.Pitch
```

## 9. Public API surface

```go
// scene/ornaments (and scene.OrnamentRecipe)
type Recipe func(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64, role pptx.ColorRole, pitch pptx.EMU) int // +pitch
// scene
// Decoration gains: Pitch pptx.EMU // 0 = legacy fixed pattern count
```

**Break:** `OrnamentRecipe`/`Recipe` gains a 2nd parameter this wave — caller
ornament extensions add the `pitch` arg (they may ignore it). v0.x; CHANGELOG.

## 10. Risks

- **R1 — legacy drift.** **Mitigation:** `pitch == 0` keeps the fixed counts; a
  byte-identity test pins each pattern.
- **R2 — file-size blowup.** **Mitigation:** `patternMaxDots` cap + a warning.
- **R3 — 2nd break churn.** **Mitigation:** mechanical positional add; the four
  non-pattern recipes ignore it; documented in CHANGELOG + D-111.

## 11. Acceptance criteria

1. `grid_dots` at a 0.4in pitch over a 13in box yields ≫ 6 columns (~32); the same pitch on a small box yields proportionally fewer dots at the same spacing.
2. A `Pitch == 0` (legacy) pattern decoration is byte-identical to the pre-Phase-77 build, per pattern preset.
3. A tiny pitch over a full-bleed box records exactly one `LayoutWarning` (cap) and emits ≤ the cap.
4. A pitched pattern re-renders deterministically.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `assets/ornaments` | 85% | pattern recipe coverage |
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-77.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `Recipe` signature carries `pitch pptx.EMU`.
3. `OK:` `Decoration.Pitch` field + render passes it.
4. `OK:` pitch-density test (many columns).
5. `OK:` legacy byte-identical + cap-warn tests.

## 14. Tests

- **Black-box (`scene_test`):** pitched `grid_dots` emits ≫ legacy count; a
  smaller box proportionally fewer; `Pitch == 0` byte-identical per pattern; a
  tiny pitch warns + caps; determinism.
- **Integration / Fuzz:** no.
