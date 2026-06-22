# Phase 43 — display text shrink-to-fit

**Subsystem:** scene — Layer 2 renderer (+ a `pptx` run override)
**RFC sections:** §8.4 (rich text), §10.2 (content-aware metrics)
**Deps:** Phase 22 (`naturalWidth`), Phase 28 (Stat), Phase 40 (quantized scale)
**Status:** Done

---

## 1. Goal

Add an opt-in `AutoFit` on the display nodes (`Hero`, `Stat`, `Heading`) that
deterministically downscales the display run's font size — via a new per-run
`FontScale` — so a too-wide title or price fits its box on one line instead of
wrapping, within a pinned minimum ratio.

## 2. Why now

R10.5 is the next Wave-10 unit. It fixes the recreation's "$4,000+" wrapping in a
narrow pricing column and the wrapped titles on slides 4/5 (DECKARD R10.5 gap). It
reuses the `naturalWidth` estimator (Phase 22) and the quantized basis-point scale
pattern from R10.2/R10.3.

## 3. RFC sections implemented

- `RFC §8.4` — a per-run font-size modifier (`FontScale`) on the builder.
- `RFC §10.2` — content-aware sizing: the renderer uses `naturalWidth` to fit a
  display run to its box width.

## 4. Brief findings incorporated

- `docs/research/26-display-text-shrink-to-fit.md` — *the width estimator exists;
  add a builder seam + a pure scale function* → `RunStyle.FontScale` +
  `fitScale(natW, boxW)`.
- `docs/research/26-display-text-shrink-to-fit.md` — *a multiplier, not an
  absolute size, keeps P2* → `FontScale` scales the resolved role size; a theme
  swap still re-skins the base.
- `docs/research/26-display-text-shrink-to-fit.md` — *quantize down + pinned floor
  guarantees fit and determinism* → `fitScale` floors `boxW·10000/natW` to a
  0.025 step, floored at 0.60, returns 0 when it already fits.
- `docs/research/26-display-text-shrink-to-fit.md` — *only the display run scales;
  Heading is multi-run* → `AutoFit` on Hero/Stat/Heading; `addRichTextScaled`
  applies one scale across a heading's runs.

## 5. Findings I'm departing from

- The brief notes `Chip`/`Arrow`/table-cell labels could also AutoFit. This plan
  scopes `AutoFit` to the **display class** (`Hero`, `Stat`, `Heading`) the spec
  names. **Departing because** those are the gap's cases; `fitScale`/`FontScale`
  remain reusable if a later req wants the others. (§4.3.)

## 6. Decisions referenced

- `D-064` — per-face width metric — `naturalWidth` (which AutoFit consults).
- `D-071` / `D-072` — quantized basis-point scale toward a pinned floor — the
  shape `fitScale` mirrors.
- `D-026` — engine, not product — AutoFit is an opt-in mechanism; the engine never
  decides on its own to shrink text.
- **New:** `D-074` — display text shrink-to-fit — filed in this PR.

## 7. Architecture

```text
pptx:  RunStyle.FontScale float64   // 0 = role size; >0 multiplies spec.Size in toProps → @sz
scene: fitScale(natW, boxW) float64 // 0 when fits; else floor(boxW·10000/natW) → 0.025 step, ≥0.60
       Hero/Stat/Heading.AutoFit bool
       renderHero/renderStat: display run RunStyle.FontScale = fitScale(naturalWidth(text@role), box.W)
       renderHeading: addRichTextScaled(..., fitScale(...)) applies one scale to all runs
```

`FontScale` is byte-identical when 0 (`size = spec.Size` exactly). `fitScale`
returns 0 when the text fits, so AutoFit-off and already-fitting content are
unchanged.

## 8. Files added or changed

```text
pptx/text.go                                    # CHANGED — RunStyle.FontScale field + godoc
pptx/text_layout.go                             # CHANGED — toProps applies FontScale to spec.Size
pptx/text_scale_test.go                         # NEW — toProps @sz + round-trip via Run.FontSize
scene/metrics.go                                # CHANGED — fitScale helper
scene/nodes.go                                  # CHANGED — Hero/Stat/Heading AutoFit bool
scene/render_leaves.go                          # CHANGED — Hero/Heading AutoFit wiring; addRichTextScaled
scene/render.go                                 # CHANGED — addRichText delegates to addRichTextScaled
scene/render_stat.go                            # CHANGED — Stat AutoFit wiring
scene/render_autofit_test.go                    # NEW — fitScale unit + per-node shrink/byte-identical/sz + determinism
scripts/smoke/phase-43.sh                       # NEW — phase smoke
docs/research/26-display-text-shrink-to-fit.md  # NEW — brief 26
docs/research/INDEX.md                          # CHANGED — register brief 26
docs/plans/phase-43-display-text-shrink-to-fit.md # NEW — this plan
docs/plans/README.md                            # CHANGED — Wave 10 phase index row
docs/decisions.md                               # CHANGED — adds D-074
docs/glossary.md                                # CHANGED — AutoFit, FontScale terms
docs/design/THEME.md                            # CHANGED — FontScale run-level size modifier note
docs/site/catalog/text-leaves.md                # CHANGED — document AutoFit
skills/compose-a-scene/SKILL.md                 # CHANGED — AutoFit in the display-node notes
```

## 9. Public API surface

```go
// pptx
type RunStyle struct {
    // …
    FontScale float64 // multiplies the resolved type-role size for this run
                      // (0/unset = the role size; >0 and <1 shrinks). Emitted as
                      // a:rPr/@sz; the result round-trips via Run.FontSize().
}

// scene
type Hero struct    { /* … */ AutoFit bool } // shrink the Title to fit one line
type Stat struct    { /* … */ AutoFit bool } // shrink the Value to fit its box
type Heading struct { /* … */ AutoFit bool } // shrink the Text to fit one line
```

Additive: `FontScale` 0 and `AutoFit` false reproduce the current output exactly.

## 10. Risks

- **R1 — byte-identical regression.** **Mitigation:** `FontScale=0` leaves `size =
  spec.Size`; `fitScale` returns 0 for fitting text, so AutoFit-off and fitting
  content are byte-identical; a test asserts an AutoFit-off node and a fitting
  AutoFit-on node match the non-AutoFit render.
- **R2 — non-deterministic scale.** **Mitigation:** integer / quantized
  basis-point math; a determinism guard renders an AutoFit deck at 1 and 8
  workers.
- **R3 — over-shrink hides content.** **Mitigation:** the 0.60 pinned floor caps
  the reduction; a test asserts the scale never drops below the floor.

## 11. Acceptance criteria

1. For a `Stat`/price value whose `naturalWidth` exceeds its column, `AutoFit`
   yields a single-line render (estimated width ≤ boxW) at a font no smaller than
   `ratioMin × base`.
2. Text that already fits is byte-identical with `AutoFit` on or off.
3. `AutoFit=false` is byte-identical to the current output.
4. A scaled run emits the expected `a:rPr/@sz` and round-trips via
   `Run.FontSize()`.
5. Identical inputs yield identical sizes (deterministic at any worker count).
6. `make coverage` keeps `pptx` and `scene` ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `pptx` | 85% | default; the change is a small toProps branch + test |
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-43.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` an over-wide AutoFit value fits its box at ≥ ratioMin
   (`TestFitScale_OverflowFitsAtOrAboveFloor`).
3. `OK:` AutoFit-off / fitting text byte-identical
   (`TestAutoFit_OffByteIdentical`).
4. `OK:` a scaled run round-trips its size (`TestRunFontScale_RoundTrip`).
5. `OK:` AutoFit render stays deterministic (`TestAutoFit_Deterministic`).

## 14. Tests

- **Unit:** `scene` — `fitScale` (fits→0, overflow→quantized fitting scale, floor
  cap, determinism); per-node AutoFit shrink + off byte-identical.
- **Codec / round-trip:** `pptx` — a `FontScale` run emits the expected `@sz` and
  `Run.FontSize()` reads it back (G6).
- **Integration / Fuzz / Bench:** none.

## 15. Vocabulary added

- `AutoFit` — the opt-in display-node flag (`Hero`/`Stat`/`Heading`) that shrinks
  the display run to fit its box width on one line.
- `FontScale` — the per-run `RunStyle.FontScale` multiplier on the resolved
  type-role size.

## 16. Plan deviations encountered during implementation

- *(none beyond the §5 scope: AutoFit limited to the display class.)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-43.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-074).
- [x] Docs site + THEME.md updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
