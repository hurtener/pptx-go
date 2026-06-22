# Phase 54 — chrome-element anti-collision (R11.6)

**Subsystem:** scene — Layer 2 renderer (card chrome)
**RFC sections:** §12.1
**Deps:** Phase 25 (D-054 status dot / pill), Phase 53 (R11.5 `cardPillWidthOf`),
brief 37.
**Status:** Done

---

## 1. Goal

Place a card's header pill and status dot so their boxes never overlap when both are
set — instead of both anchoring to the same top-right point and colliding
(recreation slide 9).

## 2. Why now

R11.6 is a Wave-11 HIGH chrome-robustness unit; it builds directly on R11.5's fitted
pill width (`cardPillWidthOf`) to know where the pill's left edge is.

## 3. RFC sections implemented

- `RFC §12.1` — card chrome (status dot + header pill placement).

## 4. Brief findings incorporated

- `docs/research/37-chrome-element-anti-collision.md`:
  - "shift the dot left of the pill" → when both set, `dotX = pillX − gapSM −
    cardStatusDotSz` (floored at `innerX`); the dot's right edge sits a gap to the
    left of the pill's left edge.
  - "byte-identical when only one is set" → the shift is guarded by `c.pill != ""`;
    a dot-only card keeps the corner placement.
  - "disjointness by construction" → the dot derives from the same `cardPillWidthOf`
    the pill is drawn with, so they never intersect for any pill length.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-054` — the status dot and header pill chrome.
- `D-085` — `cardPillWidthOf` (the pill's left edge).
- `D-086` — **new** — the anti-collision placement.

## 7. Architecture

```text
status-dot block of renderCardChrome:
  dotX = box.X + box.W − pad − cardStatusDotSz          (corner, default)
  if c.pill != "":
    pillX = innerX + innerW − cardPillWidthOf(theme, pill, innerW)
    dotX  = max(innerX, pillX − gapSM − cardStatusDotSz) (left of the pill)
```

## 8. Files added or changed

```text
scene/render_card.go                  # CHANGED — dot shifts left of the pill when both set
scene/render_anticollision_test.go    # NEW — dot shifts / dot-only unchanged
scripts/smoke/phase-54.sh             # NEW — phase smoke
docs/research/37-chrome-element-anti-collision.md  # NEW — brief 37
docs/research/INDEX.md                # CHANGED — registers brief 37
docs/plans/phase-54-chrome-element-anti-collision.md  # NEW — this plan
docs/plans/README.md                  # CHANGED — Wave 11 Phase 54
docs/decisions.md                     # CHANGED — adds D-086
```

No new exported symbol, no new token, no user-facing surface change.

## 9. Public API surface

None.

## 10. Risks

- **R1 — the shift pushes the dot off-card for a huge pill.** **Mitigation:** `dotX`
  is floored at `innerX`.
- **R2 — vacuous "shift" test.** **Mitigation:** the test compares the dot's x with
  and without a pill; `dotX(pill) = dotX(corner) − pillW − gap < dotX(corner)` always,
  so the assertion is meaningful.

## 11. Acceptance criteria

1. With both a header pill and a status dot, the dot's x-offset is smaller than a
   dot-only card's corner x (the shift fired → disjoint boxes)
   (`TestStatusDot_AntiCollision`).
2. A dot-only card keeps the corner placement and renders stably
   (`TestStatusDot_ByteIdentical_NoPill`); existing rich-visuals goldens (dot, no
   pill) pass unchanged.
3. `go test -race ./...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; render-covered |

## 13. Smoke check

`scripts/smoke/phase-54.sh` runs the two acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` black-box (`render_anticollision_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "Header pill" / "status dot" already defined.)*

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-54.sh` reports `OK ≥ 2` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (n/a — no new term).
- [x] Decision entries added (D-086).
- [x] Docs site / skill updated (n/a — no surface change).
