# Phase 76 — starfield scatter ornament

**Subsystem:** `assets/ornaments` (curated recipe) + `scene/ornaments` (registry)
**RFC sections:** §14.2 (ornaments), §14.4 (curated names), §10.1 (backward-compat)
**Deps:** Phase 73 (Decoration.Color / role recipe, D-107); brief 59.
**Status:** Done

---

## 1. Goal

Add a curated `starfield` ornament — a deterministic, irregular scatter of
role-colored dots with per-dot size and opacity variance — for the atmospheric
depth the reference's dark slides use.

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); `noise_overlay` (a uniform
lattice) cannot produce the organic starfield look. Engine req R13.6 (HIGH ·
engine; D-059). Builds on the role-colored recipes from Phase 73.

## 3. RFC sections implemented

- `RFC §14.2` — a new curated ornament recipe.
- `RFC §14.4` — extends the closed curated-name set.
- `RFC §10.1` — additive; existing ornaments unchanged.

## 4. Brief findings incorporated

- `docs/research/59-starfield-scatter-ornament.md` — *"determinism via a fixed
  integer hash, not RNG (D-035)"* → hash-perturbed lattice, no `math/rand`.
- `59` — *"the `Recipe` signature has no density param; derive count from the
  box at a fixed pitch"* → `cols = box.W/pitch`; the box is the density driver.
- `59` — *"variance from small fixed tables; cap for file size"* → size/alpha
  tables indexed by the hash; a capped total.
- `59` — *"multi-hue `Palette` deferred (signature constraint)"* → single role.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-035` — deterministic, no RNG/clock.
- `D-107` — the role-colored recipe signature the starfield uses.
- `D-110` (new) — files the `starfield` ornament + the box-as-density rationale.

## 7. Architecture

`Starfield(sl, box, alpha, rotationDeg, role) int` lays a lattice at a fixed
pitch (`cols = box.W/pitch`, `rows = box.H/pitch`), hashes each cell index to
perturb the dot position, pick its size (`{1,2,3}pt`) and its alpha
(`{35,60,100}%` of the caller alpha), and skip ~20% of cells for sparseness,
capped at a few thousand dots. Registered as `starfield` in the curated set.
Deterministic integer-EMU; no new OOXML.

```text
Decoration{Preset: "starfield", Color: &white, Opacity: 0.5, Bleed: true, Size: full}
  → Starfield(box, alpha, role=white): box-derived dots, hash size/alpha variance
  → many a:prstGeom=ellipse + a:solidFill/a:alpha at irregular spacing
```

## 8. Files added or changed

```text
assets/ornaments/patterns.go     # CHANGED — Starfield recipe + pitch/size/alpha consts
assets/ornaments/ornaments_test.go  # CHANGED — starfield in the curated map
scene/ornaments/registry.go      # CHANGED — NameStarfield + Curated()
scene/ornaments/registry_test.go # CHANGED — curated set is seven
scene/render_decoration_test.go  # CHANGED/NEW — starfield size+alpha variance, box-density, determinism
scripts/smoke/phase-76.sh        # NEW — phase smoke
docs/research/59-starfield-scatter-ornament.md  # NEW — brief
docs/research/INDEX.md           # CHANGED — registers brief 59
docs/plans/phase-76-starfield-scatter-ornament.md  # NEW — this plan
docs/plans/README.md             # CHANGED — Phase 76 detail
docs/design/THEME.md             # CHANGED — starfield ornament mechanism note
docs/glossary.md                 # CHANGED — starfield term
docs/decisions.md                # CHANGED — adds D-110
docs/site/reference/scene.md     # CHANGED — starfield in the curated ornament names
skills/compose-a-scene/SKILL.md  # CHANGED — starfield ornament
```

## 9. Public API surface

```go
// assets/ornaments
func Starfield(sl *pptx.Slide, box pptx.Box, alpha int, rotationDeg float64, role pptx.ColorRole) int
// scene/ornaments
const NameStarfield = "starfield"
```

Additive; no break.

## 10. Risks

- **R1 — non-determinism.** **Mitigation:** fixed integer hash, no RNG/clock; a
  determinism guard pins byte-identity.
- **R2 — file-size blowup on a huge box.** **Mitigation:** a hard dot cap; the
  recipe stops past it (documented; the past-cap warning is R13.7).

## 11. Acceptance criteria

1. A `starfield` decoration over a full-bleed box emits dots of ≥2 distinct sizes and ≥2 distinct alphas at irregular spacing.
2. A larger box yields more dots than a smaller box (box-as-density).
3. Two renders of the same starfield are byte-identical; the `role` colors the dots.
4. The curated set now has seven ornaments incl. `starfield`.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `assets/ornaments` | 85% | curated recipe coverage |
| `scene/ornaments` | 100% | registry stays fully covered |

## 13. Smoke check

`scripts/smoke/phase-76.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `Starfield` recipe + `NameStarfield`.
3. `OK:` curated set is seven.
4. `OK:` starfield size+alpha variance test.
5. `OK:` box-density + determinism tests.

## 14. Tests

- **Unit (assets/ornaments):** `starfield` resolves in the curated map and emits
  dots (the existing emit-shapes test covers it).
- **Black-box (`scene_test`):** ≥2 distinct dot sizes and alphas; bigger box →
  more dots; determinism; role color present.
- **Registry:** the curated set is the seven sorted names.
- **Integration / Fuzz:** no.
