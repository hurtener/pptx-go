# Phase 103 — estimator inter-node gap derives from the theme SpaceMD token

> Supersedes the "`estGap` stays pinned" clause of phase-48 (D-079) for the
> inter-node gap half only. `cardChromeEst` remains a pinned deterministic literal
> (unchanged). Closes Deckard R15.2.

**Subsystem:** scene — Layer 2 renderer (estimator / layout)
**RFC sections:** §10.2 (layout policy / overflow reporting — estimator parity)
**Deps:** Phase 22 (content-aware `preferredHeight`), Phase 39 (R10.1 wrapped
header, D-070), Phase 41 (Bento span geometry, D-072), Phase 48 (D-079 — the
plan we partially supersede), Wave-11 checkpoint (the H3 "false positive" note
this decision re-engages with)
**Status:** Done

---

## 1. Goal

Make the deterministic height estimators (`preferredHeight`, `nodesHeight`)
derive the inter-node gap from the same `SpaceMD` token the renderer uses, so
estimates match composed heights and `VAlignFill` / `VAlignFit` /
`VAlignFillCapped` / weighted Bento allocate accurately.

## 2. Why now

R15.2 (Deckard Product Requirements; documented in `DECKARD-PRODUCT-REQUIREMENTS.md`
§R15, found by the 2026-07-07 Pengui dogfood). The estimator's pinned
`estGap = 137160 EMU` (~0.15") diverges from the renderer's
`theme.ResolveSpace(pptx.SpaceMD) = 101600 EMU` (~0.08") on the default theme
by ~0.07" per inter-node gap. The Wave-11 checkpoint called this a "false
positive" because the overflow warning alone was safe (the estimate
over-counted); that holds for the warning, but **the allocation paths that USE
the estimate to render geometry were biased** — `distributeFill`/`distributeFillCapped`
grow flexible nodes proportional to `preferredHeight`, `VAlignFit` truncates by it,
weighted Bento sizes by it. A 2-node column overflows the box it was sized for
because the install includes an over-counted gap. R15.2 is the deliberate
correctness fix; mirrors R15.1 / R1's "deliberate correctness, goldens regenerate"
model.

## 3. RFC sections implemented

- `RFC §10.2` — the deterministic height estimators match the composed geometry
  (parity criterion now holds for the inter-node gap, not only chrome).
- `RFC §10.2` — overflow reporting is trustworthy (the warning fires iff the
  composed content actually exceeds the region; it no longer says "no overflow"
  when a `VAlignFill` would have allocated beyond the region).

## 4. Brief findings incorporated

- `docs/research/31-estimate-actual-parity.md` — R10.10's parity finding:
  *alignment with composed geometry is the goal* → this phase extends parity to
  the inter-node gap too.
- `docs/research/22-content-aware-height.md` (Phase 22) — the deterministic
  `wrappedLines` model + the integer-EMU discipline → the new helper
  `estGapOf(theme)` resolves an integer EMU from the theme token, preserving
  pure-integer determinism.

## 5. Findings I'm departing from

- `docs/research/31-estimate-actual-parity.md` — D-079 / phase-48 / Wave-11
  H3 "conservative estimator constant ... only over-estimates (safe)". **Departing
  because** that rationale is sound for *overflow warnings alone* but ignores the
  *allocation paths* (`VAlignFill`, `VAlignFit`, weighted Bento) where the biased
  estimate flows into real rendered geometry. D-142 supersedes the
  "`estGap` stays pinned" clause for the inter-node gap half; `cardChromeEst`
  remains pinned (the chrome height is a different mechanism, D-079 stays
  correct there).
- The DECKARD invariant "`Additive + backward-compatible. New capabilities are
  new optional fields whose zero value reproduces today's render byte-for-byte.
  (The ONE exception is R1)`". R15.2 is not R1 from the deck's own enumeration,
  but the spec literal ("the estimator must derive the inter-node gap from the
  same theme spacing token the renderer uses") is an unconditional correctness
  change. **Departing from byte-identity** by design: any container with ≥2
  stacked nodes / any `VAlignFill` / `VAlignFit` / `VAlignFillCapped` / weighted
  Bento will render with shifted geometry because the share-of-region math is now
  correct. Single-line content, span-1 Bento cells, single-node stacks,
  equal-mode Bentos with no fill, and `VAlignTop` non-fill slides stay
  **byte-identical** (their geometry did not flow through the estimator).
  Goldens regenerate + eyeball per R15.1's precedent.

## 6. Decisions referenced

- `D-026` — engine, not product — accurate estimators are a mechanism, not taste.
- `D-061` — LineHeight per-paragraph metric — the byte-identity-when-unset
  pattern the new helper mirrors (deterministic, integer-EMU).
- `D-064` — per-face `AvgCharWidth` on `FontSpec` — paired seam (the renderer
  reads what the soul populates; here the estimator reads what the theme defines).
- `D-070` — content-aware card header height.
- `D-072` — content-weighted bento rows — the path that uses the estimator to
  drive allocation; supersedes the "H3 false positive" call.
- `D-076` — `cardPaddingFor` — unchanged.
- `D-079` — estimate/actual parity — partially superseded (chrome half stands;
  gap half superseded); the new D-142 supersedes the gap half.
- **New:** `D-142` — estimator inter-node gap derives from the theme SpaceMD
  token — filed in this PR.

## 7. Architecture

```text
scene/render.go
  // const renamed: estGap → estGapFallback (the documented nil-theme fallback only)
  estGapFallback = pptx.EMU(137160)  // ~0.15"; nil-theme fallback for estGapOf

  // NEW helper: resolve the theme's inter-node gap (the same token the renderer uses).
  estGapOf(theme *pptx.Theme) pptx.EMU {
      if theme == nil { return estGapFallback }
      return theme.ResolveSpace(pptx.SpaceMD)
  }

  // Estimators now read estGapOf(theme) instead of the pinned literal
  // (9 sites: nodesHeight:1034 + 8 inside preferredHeight).
  // Renderers unchanged (they already read theme.ResolveSpace(SpaceMD)).
```

A subtle implication: when accepting `theme == nil` only as the fallback, every
caller of `preferredHeight`/`nodesHeight` already passes `r.theme` (non-nil in
practice) — see `nodesHeight(...)`/`preferredHeight(...)` callers.

## 8. Files added or changed

```text
scene/render.go                          # CHANGED — rename estGap→estGapFallback; add estGapOf; 9 estimator sites use estGapOf(theme)
scene/render_parity_test.go              # CHANGED — single-line and span-1 byte-identity baselines recompute via theme.ResolveSpace(SpaceMD), not estGap
scripts/smoke/phase-103.sh               # NEW — phase smoke
docs/plans/phase-103-estimator-gap-token.md  # NEW — this plan
docs/plans/phase-48-estimate-actual-parity.md  # CHANGED — §9 + §16 reflect D-142 supersession
docs/decisions.md                        # CHANGED — adds D-142
```

No new public surface. No `pptx` or `scene` public API changes (the helper and
the renamed const are unexported). No skill / no docs-site update required by
§14 (no new visual-property token, no new IR node, no new theme token) — but
the §19 forbidden-vocabulary hook is honored (no "phase 103" wording in
`README.md` / `CHANGELOG.md` / `docs/site/**/*.md` / `examples/**/README.md`).
A `CHANGELOG.md` entry under the next version describing the internal-correctness
change is appropriate. Verify `Makefile` rolling-cl test surface still passes.

## 9. Public API surface

None. Internal estimator-only change. No exports added, removed, or changed.
`estGap` (the symbol) is unexported, so renaming to `estGapFallback` is
internal-only.

## 10. Risks

- **R1 — goldens shift.** Any deck with ≥2 stacked nodes OR a fill-mode slide
  will render with shifted geometry because the gap is now accurate.
  **Mitigation:** (a) goldens regenerate and are eyeballed in the same PR (R15.1
  precedent); (b) acceptance criterion 1 explicitly asserts *single-line* +
  *span-1* + *single-node* + *non-fill* decks stay byte-identical (they did not
  flow through the estimator); (c) the existing determinism
  (`WithWorkers(1)` == `WithWorkers(8)`) and `TestStatsColors_Deterministic`
  guards catch any non-pure-introduced variance.
- **R2 — nil-theme edge case.** A `theme == nil` would panic on `ResolveSpace`.
  **Mitigation:** the helper short-circuits to `estGapFallback` on nil — the
  only theme-==-nil paths today are parity tests with `pptx.DefaultTheme()`
  (non-nil) and the renderer's `r.theme` (always non-nil).
- **R3 — determinism regression.** `ResolveSpace` is a pure map lookup against
  the theme, integer-EMU. The helper is a function of `theme` alone. **Mitigation:**
  the existing worker-count determinism arms (`render_parallel_test.go`) cover
  the scene-wide byte-equality invariant; we extend them with a 2-node column +
  `VAlignFill` case to lock the new parity in.
- **R4 — Wave-11 H3 false-positive note drift.** The Wave-11 checkpoint
  canonicalizes (`docs/decisions.md` ~L3009) "H3 is a false positive" as a
  documented intentional behavior. **Mitigation:** D-142 explicitly supersedes
  that note with the allocation-path argument (the overflow-warning path was safe;
  the allocation path was biased), and the same `chore(checkpoint)` PR rationale
  paragraph D-142 cites keeps auditability.
- **R5 — phase-48 plan drift.** Phase 48 §9 + §16 claim "estGap stays pinned".
  **Mitigation:** edited in the same PR (§4.3) — the two plan files now agree,
  D-142 carries the supersession rationale, `make drift-audit` passes.

## 11. Acceptance criteria

1. **Estimated height matches composed height for N-node stacks.** For `N ∈ {1, 2, 3, 4}`,
   `nodesHeight` of an N-node stack of `Prose` is within ±1 EMU of the composed
   height emitted by `stackIn` for the same default-theme slide (the R15.2 accept
   criterion "estimated == composed"); an explicit gap assertion guards the math.
2. **2-node column no longer overflows the box it was sized for.** A
   `TwoColumn{L:{Prose}, R:{Prose}}` whose `preferredHeight` is exactly the box
   `H` does *not* emit an overflow warning (the install no longer steals
   `(estGap − SpaceMD)` from the box).
3. **Wrappers stay byte-identical when the estimator is unused.**
   `TestPreferredHeight_SingleLineCardUnchanged` and
   `TestPreferredHeight_BentoSpanOneByteIdentical` pass after their
   `want`-baseline is rebuilt against `theme.ResolveSpace(SpaceMD)` (the
   byte-identity property they assert — wrapped increment == 0, span-1 == unit
   width — is preserved by the new code).
4. **Determinism preserved.** `go test -race ./...` clean; `WithWorkers(1)`
   produces byte-identical output to `WithWorkers(8)` for a scene containing at
   least one 2-node column under `VAlignFill` (covers the allocation path).
5. **No new public API.** No symbols added, removed, or renamed in `pptx` or
   `scene` exports; `go test ./...` covers the same surface area as before.
6. **`make coverage` keeps `scene` ≥ its band.**
7. **No §19 forbidden-name leak** in `README.md` / `CHANGELOG.md` /
   `docs/site/**/*.md` / `examples/**/README.md` (the drift-audit §19 guard).
8. Adversal review: `scene/render_adversarial_test.go`'s
   `adversarialScene()` continues to satisfy the R11.12 invariants (no box off-
   canvas, no header/body overlap, fit text one-line, contrast pass).

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package |
| `scene (render.go)` | existing | `estGapOf` is small but exercised by every estimator call; no override |

## 13. Smoke check

`scripts/smoke/phase-103.sh` (each prints exactly one OK/FAIL; phase is done
when OK ≥ 7 and FAIL == 0):

1. `OK:` library builds CGo-free.
2. `OK:` `nodesHeight` of an N-node stack equals composed `stackIn` height
   within ±1 EMU for `N ∈ {1, 2, 3, 4}` on the default theme
   (`TestNodesHeightMatchesComposed`).
3. `OK:` 2-node `TwoColumn` no longer overflows the box it was sized for
   (`TestTwoColumnPrefitFitsItsBox`).
4. `OK:` `TestPreferredHeight_SingleLineCardUnchanged` passes — wrapped increment
   is 0 (byte-identity property preserved).
5. `OK:` `TestPreferredHeight_BentoSpanOneByteIdentical` passes — span-1
   byte-identity property preserved.
6. `OK:` worker-count determinism (`go test ./scene -run
   TestRenderDeterministic_ParallelMatchesSequential`) stays green.
7. `OK:` `make drift-audit` clean.
8. `OK:` `make check-mirror` clean (`AGENTS.md` == `CLAUDE.md`).

## 14. Tests

- **Unit (white-box, `scene` package):**
  - `scene/render_parity_test.go` — keep the existing 4 tests
    (`TestPreferredHeight_WrappedCardGrows`,
    `TestPreferredHeight_BingleLineCardUnchanged`,
    `TestPreferredHeight_BentoSpanWidth`,
    `TestOverflow_WrappedHeaderCardFires`,
    `TestCardHeaderExtraHeight_Eyebrow`,
    `TestPreferredHeight_BentoSpanOneByteIdentical`).
    **Update `TestPreferredHeight_SingleLineCardUnchanged` + `TestPreferredHeight_BentoSpanOneByteIdentical`** to use
    `theme.ResolveSpace(pptx.SpaceMD)` in `want` (`estGapFallback` is gone from
    the test source); the byte-identity properties they assert survive.
  - **NEW** `TestNodesHeightMatchesComposed_N1_N4` — assert
    `nodesHeight(stack of N prose) ≈ stackIn-emitted-height` for N ∈ 1..4.
  - **NEW** `TestTwoColumnPrefitFitsItsBox` — a 2-node TwoColumn with
    preferredHeight == box.H does not emit a LayoutWarning.
- **Determinism (black-box/`scene_test`):**
  - `render_parallel_test.go` `TestRenderDeterministic_ParallelMatchesSequential`
    (existing) — extended with a 2-node column-under-fill case.
- **Adversarial (black-box/`scene_test`):** `render_adversarial_test.go`'s
  `adversarialScene()` keeps passing the R11.12 invariants. The wave-15 close
  asserted no broken W14/W15 invariants; Phase 103 must keep that.
- **Round-trip golden:** no new public API ⇒ no new round-trip golden; the
  existing golden corpus (if any shifts under fill modes) regenerates.
- **Integration (`test/integration/`):** no cross-subsystem seam opened.
- **Fuzz / Bench:** none.

## 15. Vocabulary added

- *(none — refines `preferredHeight` / `nodesHeight` / the
  `Content-aware text height` glossary entry.)*

## 16. Plan deviations encountered during implementation

- **No golden regeneration required in the existing suite.** D-142 was authored as a
  deliberate correctness change mirroring R15.1/R1, but the shipped scene suite's
  byte output stayed unchanged: the existing fixtures do not thread the estimator
  through the box-edge overflow boundary the old pinned gap biased. The new
  white-box parity tests carry the acceptance instead.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for `scene` (band-met).
- [x] `scripts/smoke/phase-103.sh` reports `OK ≥ 7` and `FAIL == 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (no new terms — confirmed).
- [x] Decision entries added (D-142).
- [x] (Phase 21+) Docs site updated for user-facing surface changes —
      **N/A** (no user-facing change).
- [x] (Phase 21+) Affected agent skill(s) updated — **N/A** (no user-facing
      change; the `define-a-theme` skill is untouched by an internal estimator
      change).
- [x] §19 drift-audit guard clean.

---

*Replaces the template footer; structure mirrors `docs/plans/_template.md`.*
