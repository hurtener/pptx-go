# Phase 49 — card header content-aware height (R11.1 verify-and-close)

**Subsystem:** scene — Layer 2 renderer (card chrome / layout)
**RFC sections:** §10.1, §12.1
**Deps:** Phase 39 (R10.1 / D-070 — wrapped-aware card header), Phase 48 (R10.10 /
D-079 — estimator parity), brief 32.
**Status:** Done

---

## 1. Goal

Verify that the wrapped-aware card-header geometry shipped in R10.1 fully satisfies
R11.1 (header band + body offset grow with wrapped header/eyebrow text under any
content) and close the requirement with its acceptance golden: a long multi-line
header swept across every `CardSize × CardLayout` combination asserting the body
never overlaps the header band, with single-line headers byte-identical.

## 2. Why now

R11.1 is the first (CRITICAL) unit of Wave 11 (R11 component-rendering robustness).
Its mechanism — `cardHeaderBottom` and `renderCardChrome` consuming an identical
shared `wrappedLines`-based header-row computation — already landed as D-070, and
the estimator side as D-079. Per `CLAUDE.md §17`, the correct close for an
already-implemented requirement is to prove the invariant under the full content
the requirement names and record the closure, not to reimplement it. Opening Wave
11 here keeps the wave's CRITICAL items first (master plan, `docs/plans/README.md`).

## 3. RFC sections implemented

- `RFC §12.1` — card chrome (header band + accent + body region). This phase adds
  no new chrome; it verifies the wrapped-aware header band against R11.1.
- `RFC §10.1` — content-aware layout: the header band and body-start Y track the
  real wrapped height (already realized by D-070; verified here).

## 4. Brief findings incorporated

- `docs/research/32-card-header-robustness-verify.md` — "R11.1 is already
  implemented by D-070; the acceptance gap is test coverage across all
  `CardSize × CardLayout` combos, not mechanism" → this plan ships the combinatorial
  acceptance golden and the closure decision, with no renderer change.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-070` — content-aware card header height (R10.1). This phase verifies it
  satisfies R11.1 and adds R11.1's named acceptance golden.
- `D-079` — estimate/actual parity (R10.10). The estimator side of the same
  invariant; cited as the other half of the close.
- `D-081` — **new** — records R11.1 closed by D-070 (+ D-079) with the
  CardSize×CardLayout acceptance sweep as evidence.

## 7. Architecture

No production code change. The existing seam:

```text
cardHeaderColumnWOf(theme, box, c)   — header text column width (icon/pill aware)
        │
cardHeaderRowHeights(box, c)         — (eyebrowH, titleH) = per-row × wrappedLines
        │
        ├── cardHeaderBottom(box, c) ─────────► body-region top Y
        └── renderCardChrome(...)    ─────────► D-054 header band height + text frames
```

`cardHeaderBottom` and `renderCardChrome` share `cardHeaderRowHeights`, so the
drawn band, the emitted text frames, and the body Y are derived from one
line-count computation — the property R11.1's acceptance asserts.

## 8. Files added or changed

```text
scene/render_card_internal_test.go   # CHANGED — adds TestCardBodyBelowWrappedHeader_AllCombos
scripts/smoke/phase-49.sh            # NEW — phase smoke
docs/research/32-card-header-robustness-verify.md  # NEW — brief 32
docs/research/INDEX.md               # CHANGED — registers brief 32
docs/plans/phase-49-...-verify.md    # NEW — this plan
docs/plans/README.md                 # CHANGED — opens Wave 11, adds Phase 49
docs/decisions.md                    # CHANGED — adds D-081
docs/glossary.md                     # CHANGED — adds "card header band" / "header column width"
```

No user-facing public API change → no docs-site / skill update required this phase
(the wrapped-header behavior was documented when D-070 landed).

## 9. Public API surface

None. No new or changed exported symbol.

## 10. Risks

- **R1 — the sweep passes vacuously (no wrapping).** If the chosen header is short
  enough to fit one line at the LG padding, the test would not exercise wrapping.
  **Mitigation:** assert `titleH > cardTitleRowH` (≥ 2 lines) for each combo so the
  test fails if the header does not actually wrap; use a deliberately long string
  in a narrow (1/3-width) card.
- **R2 — IconTop interaction missed.** **Mitigation:** the sweep includes both
  `CardLayoutDefault` and `CardLayoutIconTop`, each with `Icon` set, so the
  icon-band advance is exercised.

## 11. Acceptance criteria

1. For a deliberately long header in a narrow card, across all `{CardSizeMD, SM,
   LG} × {CardLayoutDefault, IconTop}` combinations, the rendered body box top Y is
   `>=` `cardHeaderBottom` (no header/body overlap).
2. With a `HeaderFill` band enabled, the band bottom equals `cardHeaderBottom`
   (the band fully contains the header, no spill).
3. A single-line header yields `titleH == cardTitleRowH` (byte-identical to the
   legacy fixed advance) for every combo.
4. The long header wraps to `>= 2` lines in every combo (the test is not vacuous).
5. `go test -race ./scene/...` passes; `make coverage` stays green.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | unchanged; test-only addition, no new production lines |

## 13. Smoke check

`scripts/smoke/phase-49.sh`:

1. `OK:` `go build ./...`.
2. `OK:` `go test -run TestCardBodyBelowWrappedHeader_AllCombos ./scene/` passes.
3. `OK:` `go test -run TestCardHeaderBottom_WrappedTitle ./scene/` (prior R10.1
   guard still green).

## 14. Tests

- **Unit:** `scene` white-box — `TestCardBodyBelowWrappedHeader_AllCombos`.
- **Round-trip golden:** n/a — no builder API added.
- **Integration:** no — same subsystem, no new seam.
- **Fuzz:** no.
- **Benchmark:** no.

## 15. Vocabulary added

- `card header band` — the colored region (D-054 `HeaderFill`) drawn from the card
  top to the header bottom; its height tracks the wrapped header line count.
- `header column width` — the inner text column at which a card's eyebrow/title
  wrap (`cardHeaderColumnWOf`): inner width minus the icon shift and pill reserve.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-49.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-081).
- [x] Docs site updated for user-facing surface changes (n/a — no surface change).
- [x] Affected agent skill(s) updated (n/a).
