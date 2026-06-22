# Phase 41 — content-weighted bento rows

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §11.2 (Bento), §10 (layout engine)
**Deps:** Phase 27 (Bento, D-056), Phase 40 (R10.2 fit primitive, D-071)
**Status:** Done

---

## 1. Goal

Add an opt-in `Bento.WeightedRows` mode that sizes each bento row to its content's
preferred height — clamped so the rows always fit the region — so a dense row no
longer shares an equal band with a sparse one.

## 2. Why now

R10.3 is the next Wave-10 unit after the two CRITICALs (R10.1 header height,
R10.2 fit compression). It is the direct fix for the recreation's slide-6 bento
where the sparse top "Canvas" row occupies a huge band while the dense bottom
rows starve and overflow (DECKARD R10.3 gap). It composes the R10.2 fit primitive
(D-071) and the Bento node (D-056).

## 3. RFC sections implemented

- `RFC §11.2` — the Bento container gains content-proportional row heights
  (opt-in); the equal-row default is unchanged.
- `RFC §10` — content-aware vertical distribution within a container, clamped to
  the region (a partial; the Grid analog and estimator parity are deferred).

## 4. Brief findings incorporated

- `docs/research/24-content-weighted-bento-rows.md` — *the fix is purely vertical
  distribution; factor out the shared horizontal math* → `bentoColumns` +
  `cellWidth` extracted; `bentoGeometry` returns per-row Y/H.
- `docs/research/24-content-weighted-bento-rows.md` — *per-row height = max cell
  preferred height at span width, computed by a renderer method; geometry stays
  pure* → `r.bentoWeightedRowHeights` computes the slice, `bentoGeometry` accepts
  it.
- `docs/research/24-content-weighted-bento-rows.md` — *the overflow clamp must
  guarantee fit (no 0.60 floor), unlike fitCompress* → a single basis-point scale
  `sBP = avail·10000/Σpref`, flooring guarantees `Σ ≤ avail`.
- `docs/research/24-content-weighted-bento-rows.md` — *byte-identical default; the
  per-row-array refactor reproduces today's boxes and label Ys* → equal mode
  fills every row with the same `rowH` and accumulates `rowY` identically.

## 5. Findings I'm departing from

- The R10.3 spec title names *bento/grid*. This plan implements **Bento only**;
  the **Grid analog is deferred**. **Departing because** the acceptance criterion
  is bento-specific ("a bento with one 1-line row and one 4-line row…"), Grid
  cells are single children laid out by the pure `layout.Grid` (no theme), and
  content-weighting Grid is a separable change better folded into R10.10 or a
  follow-up. Documented as a §4.3 deviation (see §16).
- The spec offers a *proportional* row-fill of leftover slack. This plan defaults
  weighted mode to **top-align** (rows at preferred height, slack as bottom
  whitespace) — the spec lists top-align as an acceptable option and it already
  satisfies "a sparse row does not steal space from a dense one".

## 6. Decisions referenced

- `D-056` — Bento node — the container this extends.
- `D-071` — fit-to-region compression — the basis-point "scale toward fit"
  primitive the overflow clamp mirrors (without the ratio floor).
- `D-026` — engine, not product — content-proportional rows are an opt-in
  mechanism, not a taste decision.
- **New:** `D-072` — content-weighted bento rows — filed in this PR.

## 7. Architecture

`bentoGeometry` is refactored: the shared horizontal math moves to `bentoColumns`
(gutter width, content origin X, unit column width) and `cellWidth(span)`; the
function now returns per-row `rowYs` and `rowHs` slices plus the cell boxes, and
accepts an optional `rowHs []EMU` (nil ⇒ equal mode). `renderBento`, when
`v.WeightedRows`, computes the per-row heights via `r.bentoWeightedRowHeights`
and threads them in; both the gutter labels and the cells use the per-row Y/H.

```text
bentoWeightedRowHeights(box, v, gap):
  unitW = bentoColumns(...).unitW
  pref_i = max over cells in row i of preferredHeight(cell, cellWidth(span), theme)
  avail  = box.H - gap*(nRows-1)
  if Σpref > avail:  sBP = avail*10000/Σpref;  h_i = pref_i*sBP/10000   // fits
  else:              h_i = pref_i                                       // top-align, slack at bottom
```

## 8. Files added or changed

```text
scene/nodes.go                                  # CHANGED — Bento.WeightedRows bool
scene/render_bento.go                           # CHANGED — bentoColumns/cellWidth; bentoGeometry per-row Y/H + rowHs arg; bentoWeightedRowHeights; renderBento wiring
scene/render_bento_test.go                      # CHANGED — geometry call sites; weighted-row white-box tests
scene/render_bento_render_test.go               # CHANGED — weighted determinism / byte-identical render tests
scripts/smoke/phase-41.sh                       # NEW — phase smoke
docs/research/24-content-weighted-bento-rows.md # NEW — brief 24
docs/research/INDEX.md                          # CHANGED — register brief 24
docs/plans/phase-41-content-weighted-bento-rows.md # NEW — this plan
docs/plans/README.md                            # CHANGED — Wave 10 phase index row
docs/decisions.md                               # CHANGED — adds D-072
docs/glossary.md                                # CHANGED — Bento weighted rows term
docs/site/guide/scene.md                        # CHANGED — document WeightedRows
skills/compose-a-scene/SKILL.md                 # CHANGED — WeightedRows in the bento section
```

## 9. Public API surface

```go
// scene
type Bento struct {
    // …
    WeightedRows bool // opt-in: size each row to its content's preferred height
                      // (clamped to fit the region) instead of equal rows.
}
```

No breaking change: an additive bool field; the zero value reproduces the equal-
row layout exactly.

## 10. Risks

- **R1 — byte-identical regression in equal mode.** The geometry refactor could
  perturb existing bento output. **Mitigation:** equal mode fills every `rowH`
  with the same value and accumulates `rowY` identically; a render test asserts a
  default bento is byte-identical to the pre-refactor output.
- **R2 — overflow in weighted mode.** **Mitigation:** the basis-point scale floors
  so `Σ rows ≤ avail`; a test asserts `Σ rowHs + gaps ≤ box.H`.
- **R3 — determinism under parallel render.** **Mitigation:** integer/basis-point
  math; a determinism guard renders the same weighted-bento deck at 1 and 8
  workers.

## 11. Acceptance criteria

1. A bento with one 1-line row and one 4-line row, in `WeightedRows` mode, renders
   the dense row taller than the sparse row, with `Σ row heights + gaps ≤ box.H`
   (no overflow).
2. An equal-mode (`WeightedRows=false`) bento — and a single-density bento — is
   byte-identical to the current output.
3. Identical inputs yield identical EMU geometry; a weighted-bento deck rendered
   at 1 and 8 workers is byte-identical.
4. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package (no override) |

## 13. Smoke check

`scripts/smoke/phase-41.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` weighted rows size dense > sparse and fit
   (`TestBentoWeightedRows_DenseTallerAndFits`).
3. `OK:` equal-mode bento byte-identical
   (`TestBentoGeometry_EqualModeByteIdentical`).
4. `OK:` weighted bento render stays deterministic
   (`TestBentoWeighted_Deterministic`).

## 14. Tests

- **Unit:** `scene` (white-box) — `bentoWeightedRowHeights` dense>sparse + fits +
  clamp; `bentoGeometry` equal-mode byte-identical; per-row Y/H.
- **Round-trip golden:** n/a (scene layout change, no new OOXML).
- **Integration:** no (scene-internal).
- **Fuzz / Benchmark:** none.

## 15. Vocabulary added

- `Bento weighted rows` — the opt-in `Bento.WeightedRows` mode that sizes each
  bento row to its content's preferred height, clamped to fit the region.

## 16. Plan deviations encountered during implementation

- **Grid analog deferred.** Phase 41 implements content-weighted rows for
  `Bento` only; the Grid analog is deferred (separable; `layout.Grid` is a pure
  theme-free function). Restates AC1 to bento. (§4.3.)
- **Top-align default for leftover slack** (vs proportional fill), per §5.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-41.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-072).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
