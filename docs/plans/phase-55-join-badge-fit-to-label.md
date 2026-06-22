# Phase 55 — join-badge fit-to-label (R11.7)

**Subsystem:** scene — Layer 2 renderer (TwoColumn join)
**RFC sections:** §11.2
**Deps:** Phase 26 (D-055 column join), Phase 43 (D-074 `fitScale`), Phase 53 (R11.5
fit pattern), brief 38.
**Status:** Done

---

## 1. Goal

Grow the TwoColumn join badge to contain its label on a single line — instead of a
fixed `In(0.62)` ellipse that breaks "One agent" mid-word — for any connector label.

## 2. Why now

R11.7 is a Wave-11 HIGH fit-to-label unit (recreation slide 8), mirroring R11.5's
pill fit with the established `naturalWidth` / `fitScale` primitives.

## 3. RFC sections implemented

- `RFC §11.2` — TwoColumn join element.

## 4. Brief findings incorporated

- `docs/research/38-join-badge-fit-to-label.md`:
  - "grow to fit, cap, then shrink" → `badgeSz = clamp(naturalWidth(label) + 2·padX,
    joinBadgeSz, joinBadgeMaxSz)`, then `FontScale = fitScale(natW, badgeSz − 2·padX)`.
  - "the inter-column gap is the wrong cap axis" → a pinned `joinBadgeMaxSz = In(1.5)`
    cap (the badge deliberately overlaps both columns).
  - "byte-identical for 'vs'" → a short label keeps the base diameter and full size.

## 5. Findings I'm departing from

*"none"* (the spec's "clamp to inter-column gap" is replaced by a pinned max cap, for
the reason in §4 — documented in the brief and D-087).

## 6. Decisions referenced

- `D-055` — the column join badge.
- `D-074` — `fitScale` / `FontScale`, reused.
- `D-087` — **new** — the join-badge fit-to-label.

## 7. Architecture

```text
JoinBadge case of renderColumnJoin:
  natW    = naturalWidth(label @ TypeBodySmall)
  badgeSz = clamp(natW + 2·joinBadgePadX, joinBadgeSz, joinBadgeMaxSz)
  scale   = fitScale(natW, badgeSz − 2·joinBadgePadX)   // 0 when it fits
  ellipse(badgeSz) + centered label{FontScale: scale}
```

## 8. Files added or changed

```text
scene/render_container.go         # CHANGED — fit-to-label badge + cap + FontScale
scene/render_join_badge_test.go   # NEW — grows / vs base / caps+shrinks
scene/render_parallel_test.go     # CHANGED — TestRenderDeterministic_JoinBadgeFit
scripts/smoke/phase-55.sh         # NEW — phase smoke
docs/research/38-join-badge-fit-to-label.md  # NEW — brief 38
docs/research/INDEX.md            # CHANGED — registers brief 38
docs/plans/phase-55-join-badge-fit-to-label.md  # NEW — this plan
docs/plans/README.md              # CHANGED — Wave 11 Phase 55
docs/decisions.md                 # CHANGED — adds D-087
docs/glossary.md                  # CHANGED — Column join fit note
```

No new exported symbol.

## 9. Public API surface

None. `joinBadgePadX` / `joinBadgeMaxSz` are unexported. The user-visible effect is
that `TwoColumn.JoinLabel` renders intact at any short length.

## 10. Risks

- **R1 — a pathological label still overflows.** **Mitigation:** capped at
  `joinBadgeMaxSz` then `fitScale` shrinks to the 0.60 floor; out-of-scope beyond.
- **R2 — "vs" not byte-identical.** **Mitigation:** `needed < joinBadgeSz` so the base
  diameter and `FontScale = 0` are kept; a test asserts the "vs" diameter == base.

## 11. Acceptance criteria

1. A multi-word label grows the badge diameter beyond the base `In(0.62)`; "vs" keeps
   the base (`TestJoinBadge_GrowsToLabel`).
2. An over-long label caps at `joinBadgeMaxSz` (`TestJoinBadge_CapsAndShrinks`).
3. Join badges render byte-identically across worker counts
   (`TestRenderDeterministic_JoinBadgeFit`).
4. `go test -race ./...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; render-covered |

## 13. Smoke check

`scripts/smoke/phase-55.sh` runs the three acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` black-box (`render_join_badge_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "Column join" already defined; its entry gains a fit note.)*

## 16. Plan deviations encountered during implementation

- The spec's "clamp to the inter-column gap" is replaced by a pinned `joinBadgeMaxSz`
  cap (the badge intentionally overlaps the columns, so the gap is not the right
  bound). Documented in brief 38 and D-087.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-55.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-087).
- [x] Docs site / skill updated (n/a — no surface change; glossary note suffices).
