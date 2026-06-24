# Brief 59 — Starfield scatter ornament (R13.6)

> Informs Phase 76 (Wave 13). Engine req R13.6
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine; D-059). Builds on the
> decoration color role (Phase 73 / D-107).

## 1. Motivating phase

Reference dark slides are covered in an irregular, sparse starfield — dots of
varying size and opacity scattered organically for depth. Deckard's only scatter
is `noise_overlay`: a regular lattice of identical 2pt dots at uniform alpha.
Phase 76 adds a curated `starfield` ornament with deterministic
pseudo-random placement and per-dot size/alpha variance.

## 2. Subsystem / files

- `assets/ornaments/patterns.go` — `GridDots`/`NoiseOverlay` (the fixed-lattice
  pattern recipes + the deterministic per-cell offset idiom `(c*7+r*3)%5`).
- `scene/ornaments/registry.go` — the closed curated-name set + `Curated()`.
- `scene/ornaments/registry_test.go` — `TestCurated_HasSixOrnaments` (the
  closed-name assertion to extend to seven).

## 3. Findings

- **Determinism via a fixed integer hash, not RNG (D-035).** `NoiseOverlay`
  already proves the pattern: a fixed per-cell arithmetic offset, no
  `math/rand`, no clock. Index the size and alpha tables by the same hash so the
  whole field is pure integer-EMU and reproducible.
- **The `Recipe` signature has no density/pitch param.** `Recipe func(sl, box,
  alpha, rotationDeg, role) int` is fixed (changing it again would be a third
  break this wave). Derive the dot count from the **box size** at a fixed
  internal pitch: `cols = box.W / pitch`, `rows = box.H / pitch`. A full-bleed
  box gets a dense field; a small box a sparse one — so the caller controls
  density by sizing the decoration (and `Bleed` for full coverage). This is the
  "scaling with box" the req calls for; an explicit caller pitch/density is
  R13.7's job (a separate phase).
- **Variance from small fixed tables.** Per-dot size from `{1,2,3}pt` and
  per-dot alpha from `{35,60,100}%` of the caller alpha, each indexed by the
  hash → ≥2 distinct sizes and ≥2 distinct alphas, irregular. ~20% of cells
  empty (hash sieve) for sparseness.
- **Cap for file size.** Bound the total dots (a few thousand) so a huge
  full-bleed box cannot explode the part. The recipe has no `r.warn` hook, so
  cap silently and document (the warning-past-cap is R13.7's pitch concern).
- **Multi-hue `Decoration.Palette` deferred.** R13.5 deferred a multi-hue
  scatter palette here, but the `Recipe` signature cannot carry a `[]ColorRole`
  without another break. Ship the single-`role` starfield; multi-hue confetti is
  a later phase / V2 (note it).
- **No OOXML / `restorenamespaces` change.** Same `a:prstGeom` ellipses +
  `a:solidFill`/`a:alpha` the other patterns emit.

## 4. Recommendations

- Add `Starfield(sl, box, alpha, rotationDeg, role) int` in
  `assets/ornaments/patterns.go`: box-derived lattice, hash-perturbed placement,
  size/alpha from fixed tables, ~20% empty cells, capped total.
- Register `NameStarfield = "starfield"` in `scene/ornaments/registry.go` +
  `Curated()`; extend `TestCurated_HasSixOrnaments` to seven.
- Tests: a starfield emits ≥2 distinct dot sizes and ≥2 distinct alphas; a
  bigger box yields more dots than a smaller one; two renders byte-identical;
  the role colors the dots (via D-107). THEME.md note, glossary, compose-a-scene
  skill, docs/site curated names. D-110.

## 5. Open questions

- Explicit caller pitch/density (and the past-cap warning) → R13.7.
- Multi-hue `Decoration.Palette` scatter → deferred (signature constraint).
