# Phase 60 — adversarial content-fit fixtures (R11.12)

**Subsystem:** scene — Layer 2 renderer (test harness + safe-area clamp)
**RFC sections:** §10, §10.1
**Deps:** Phases 49–59 (the Wave-11 per-component guards), brief 43.
**Status:** Done

---

## 1. Goal

Ship a reusable, deterministic adversarial harness that renders every component
under hostile content and asserts the structural invariants — every box on the
canvas, header band ≤ body top, fit-required text one line, chrome contrast passes —
catching the whole class of fixed-size regressions; and fix the off-canvas
card-body-leaf overflow it surfaces.

## 2. Why now

R11.12 is the Wave-11 capstone (HIGH · both). It ties together the per-component
guards (R11.1–R11.11) into one torture suite, and is the CI backstop for the
fixed-size regression class.

## 3. RFC sections implemented

- `RFC §10`, `§10.1` — content layout within bounded regions, deterministic.

## 4. Brief findings incorporated

- `docs/research/43-adversarial-content-fit-fixtures.md`:
  - "a box recorder is unnecessary" → invariant (2) parses every `<a:off>`/`<a:ext>`
    from the rendered XML and checks it lies within the canvas.
  - "the suite surfaced a real bug: card-body leaf overflow" → fixed by generalizing
    the R11.3 clamp to `renderNode` (exempting full-slide overlays), removing the
    three redundant container clamps.
  - "invariants (1)/(3)/(4) asserted white-box on the same inputs" → via
    `cardHeaderBottom`/`renderCardChrome`, `cardPillWidthOf`+`statValueFit`+`fitScale`,
    `onCardSurface`+`contrastRatioT10`.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-071` (`VAlignFit`), `D-073` (`Card.BodyVAlign`) — the opt-in density-compression
  the clamp complements (the clamp caps the box; legible compression is opt-in).
- `D-083` — the R11.3 safe-area clamp this phase generalizes to `renderNode`.
- `D-092` — **new** — the adversarial harness + the leaf-overflow clamp.

## 7. Architecture

```text
renderNode(box, n):
  if n is Decoration | SectionDivider:  pass            # full-slide overlay / bleed
  else:  box = clampToSafeArea(box)                       # caps leaves AND containers
  dispatch(box, n)
# the per-container clamps (renderBento/renderGrid/renderCard) are removed — subsumed.

adversarialScene(): every component × {light, dark} × hostile content
  TestAdversarial_AllBoxesOnCanvas   (2) parse every box, assert on-canvas
  TestAdversarial_HeaderBandBelowBody (1) cardHeaderBottom ≤ body.Y
  TestAdversarial_FitTextOneLine     (3) pill / stat value fit one line
  TestAdversarial_ContrastPasses     (4) onCardSurface clears 4.5:1
  TestAdversarial_Deterministic      byte-identical across worker counts
```

## 8. Files added or changed

```text
scene/render.go                              # CHANGED — renderNode clamps every content node (R11.3 generalized)
scene/render_bento.go                        # CHANGED — drop the redundant container clamp (subsumed)
scene/render_container.go                    # CHANGED — drop the redundant grid clamp
scene/render_card.go                         # CHANGED — drop the redundant card clamp
scene/render_adversarial_test.go             # NEW — hostile fixture + on-canvas parse + determinism
scene/render_adversarial_invariants_test.go  # NEW — header/body, fit-one-line, contrast
scripts/smoke/phase-60.sh                    # NEW — phase smoke
docs/research/43-adversarial-content-fit-fixtures.md  # NEW — brief 43
docs/research/INDEX.md                       # CHANGED — registers brief 43
docs/plans/phase-60-adversarial-content-fit-fixtures.md  # NEW — this plan
docs/plans/README.md                         # CHANGED — Wave 11 Phase 60
docs/decisions.md                            # CHANGED — adds D-092
```

No new exported symbol.

## 9. Public API surface

None. The clamp is internal; the only user-observable change is that an over-full
card body / stack leaf is capped to the safe area (its box no longer draws off-slide)
and an overflow `LayoutWarning` fires.

## 10. Risks

- **R1 — the clamp move changes byte output.** **Mitigation:** the clamp is a no-op
  when the box fits, so fitting decks are byte-identical (the full existing golden
  suite passes); only over-full content is capped.
- **R2 — clamping a full-slide overlay.** **Mitigation:** `Decoration` and
  `SectionDivider` are exempt (they span the slide / bleed by design).
- **R3 — the off↔ext pairing misaligns.** **Mitigation:** each shape's `spPr/xfrm`
  emits exactly one `<a:off>`+`<a:ext>` in order; the suite renders cleanly and the
  invariant holds, so the pairing is sound for the emitted shapes.

## 11. Acceptance criteria

1. Across the hostile fixture (every component × light/dark), every emitted box lies
   within the slide canvas (`TestAdversarial_AllBoxesOnCanvas`).
2. The header band stays above the body for a hostile long header
   (`TestAdversarial_HeaderBandBelowBody`).
3. Fit-required chrome text (pill, stat value) resolves to one line
   (`TestAdversarial_FitTextOneLine`).
4. Chrome text clears 4.5:1 on hostile surfaces (`TestAdversarial_ContrastPasses`).
5. The fixture renders byte-identically across worker counts
   (`TestAdversarial_Deterministic`); the full existing golden suite passes.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; the clamp generalization is covered, overflow + fitting |

## 13. Smoke check

`scripts/smoke/phase-60.sh` runs the five acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` black-box + white-box (the two adversarial files).
- **Integration:** the on-canvas suite is a cross-component integration assertion.
- **Round-trip golden / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — reuses "safe area", "auto-contrast", "Header pill", etc.)*

## 16. Plan deviations encountered during implementation

- The suite surfaced an off-canvas card-body-leaf overflow; fixed in this PR by
  generalizing the R11.3 clamp to `renderNode` (the §17 "fix the root cause in the
  same PR" rule). Documented in brief 43 and D-092.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-60.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (n/a — no new term).
- [x] Decision entries added (D-092).
- [x] Docs site / skill updated (n/a — no surface change).
