# Phase 59 — decoration/watermark anti-collision (R11.11 verify-and-close)

**Subsystem:** scene — Layer 2 renderer (card watermark / decorations)
**RFC sections:** §10.2, §12.1
**Deps:** Phase 25 (D-054 watermark), brief 42.
**Status:** Done

---

## 1. Goal

Verify that the card watermark (and background decorations) never reduce body
legibility — they are drawn behind the body content at low opacity — and close
R11.11 with its acceptance.

## 2. Why now

R11.11 is the LOW Wave-11 unit. Its z-order + low-opacity guarantee already landed
with D-054; the R11.11 acceptance is an explicit OR whose z-order branch the engine
already satisfies.

## 3. RFC sections implemented

- `RFC §10.2` — z-order (decorations behind / on top of the body).
- `RFC §12.1` — card chrome (the watermark).

## 4. Brief findings incorporated

- `docs/research/42-decoration-watermark-anticollision.md`:
  - "the acceptance's z-order branch is already satisfied (D-054)" → the watermark is
    emitted before the body (z-order behind) at ~13% opacity; background decorations
    likewise.
  - "the residual-region restriction is the optional alternative, not required" → not
    adopted (it would couple chrome to the body estimate and change positions for a
    LOW cosmetic gain).
  - "the close is the acceptance test" → z-order-behind, low-alpha, inert-when-unset.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-054` — the watermark (z-order-behind + low alpha) and the layer z-order.
- `D-091` — **new** — records R11.11 closed by D-054 + the acceptance.

## 7. Architecture

No production change.

```text
renderCardChrome: … step 6 watermark (last chrome shape, before returning bodyBox)
renderCard:       body content rendered after  →  watermark is behind (z-order first)
layout():         LayerBackground decorations before the body; LayerForeground after
watermark color = TokenColorAlpha(ColorAccent, cardWatermarkAlpha=13000)  // ~13%
```

## 8. Files added or changed

```text
scene/render_watermark_zorder_test.go  # NEW — z-order-behind / low-alpha / inert-when-unset
scripts/smoke/phase-59.sh              # NEW — phase smoke
docs/research/42-decoration-watermark-anticollision.md  # NEW — brief 42
docs/research/INDEX.md                 # CHANGED — registers brief 42
docs/plans/phase-59-decoration-watermark-anticollision.md  # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 11 Phase 59
docs/decisions.md                      # CHANGED — adds D-091
```

No public API change, no new token, no user-facing surface change.

## 9. Public API surface

None.

## 10. Risks

- **R1 — the order test is brittle to refactors.** **Mitigation:** it asserts a
  semantic invariant (watermark text before body text), which is the legibility
  guarantee; a refactor that reorders them would be a real regression the test
  catches.

## 11. Acceptance criteria

1. The watermark text frame is emitted before the body content in the slide XML
   (z-order behind) (`TestWatermark_BehindBody`).
2. The watermark run carries a low ~13% alpha (`TestWatermark_LowAlpha`).
3. A card without a watermark emits no alpha run (inert / byte-identical)
   (`TestWatermark_OmittedWhenUnset`).
4. `go test -race ./...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | unchanged; test-only addition |

## 13. Smoke check

`scripts/smoke/phase-59.sh` runs the three acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` black-box (`render_watermark_zorder_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "Watermark" / "Header band" already defined.)*

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-59.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (n/a — no new term).
- [x] Decision entries added (D-091).
- [x] Docs site / skill updated (n/a — no surface change).
