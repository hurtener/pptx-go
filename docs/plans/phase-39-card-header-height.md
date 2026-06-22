# Phase 39 — content-aware card header height

**Subsystem:** scene — Layer 2 renderer (card chrome / layout)
**RFC sections:** §10.1, §12.1
**Deps:** brief 09 (`wrappedLines`); the D-054 card chrome.
**Status:** Done

---

## 1. Goal

A card whose header (or eyebrow) wraps to N lines begins its body below the
*actual wrapped* header — no overlap — and sizes the D-054 header band to the
wrapped height.

## 2. Why now

R10.1 is one of the two CRITICAL off-slide/overlap bugs that open Wave 10
(content fit & density). The card header overlap is a visible defect on dense
slides; the fix reuses the existing `wrappedLines` estimator and unblocks R10.2
(fit-to-region) and R10.10 (estimate-actual parity).

## 3. RFC sections implemented

- `RFC §10.1` — deterministic layout geometry. Header rows become content-aware.
- `RFC §12.1` — card chrome (the header band tracks the wrapped header).

## 4. Brief findings incorporated

- `docs/research/22-card-header-height.md` — *both paths must wrap at the same
  width* → shared `cardHeaderColumnW` + `cardHeaderRowHeights`.
- `22-card-header-height.md` — *single-line byte-identical* → `wrappedLines`
  returns 1 for fitting text, so the legacy fixed advance is reproduced exactly.
- `22-card-header-height.md` — *estimator parity deferred* → `cardChromeEst`
  stays fixed; R10.10 owns it.

## 5. Findings I'm departing from

None.

## 6. Decisions referenced

- `D-054` — card chrome + the header band sized off `cardHeaderBottom`.
- `D-061` — the visual-first / estimator-later split this phase mirrors.
- New: `D-070` — content-aware card header height (this phase).

## 7. Architecture

```text
cardHeaderColumnW(box,c)   = innerW − icon-left shift − pill reservation
cardHeaderRowHeights(box,c):
    eyebrowH = cardEyebrowRowH × wrappedLines(eyebrow, TypeCaption, headerW)
    titleH   = cardTitleRowH   × wrappedLines(header,  TypeH3,      headerW)
cardHeaderBottom  → advances by eyebrowH/titleH  (body Y; D-054 band height)
renderCardChrome  → emits the eyebrow/title frames at eyebrowH/titleH
```

## 8. Files added or changed

```text
scene/render_card.go                       # CHANGED — cardHeaderColumnW + cardHeaderRowHeights; wrapped advance
scene/render_card_internal_test.go         # CHANGED — wrapped-title + no-overlap tests
scripts/smoke/phase-39.sh                  # NEW — phase smoke
docs/research/22-card-header-height.md      # NEW — brief
docs/research/INDEX.md                      # CHANGED — register brief 22
docs/plans/phase-39-card-header-height.md   # NEW — this plan
docs/plans/README.md                        # CHANGED — Wave 10 phase index
docs/decisions.md                           # CHANGED — adds D-070
docs/site/guide/scene.md / skills/compose-a-scene # CHANGED — wrapped-header note (if present)
```

## 9. Public API surface

No new public symbol. `Card.Header`/`Card.Eyebrow` rendering becomes
wrapped-aware (internal geometry only).

## 10. Risks

- **R1 — Drift between body-Y and emitted header.** **Mitigation:** both paths
  route through the same `cardHeaderRowHeights` (and `cardHeaderColumnW`).
- **R2 — Single-line regression.** **Mitigation:** `wrappedLines==1` → exact
  legacy heights; covered by a byte-identical assertion + the existing card
  goldens.
- **R3 — Non-determinism.** **Mitigation:** `wrappedLines` is pure integer ceil
  division; the scene determinism guard still passes.

## 11. Acceptance criteria

1. A long header in a 1/3-width card advances the body top below the wrapped
   header bottom (no vertical overlap).
2. The header band height equals the wrapped header height.
3. A single-line header is byte-identical to the prior output.
4. Deterministic; `make coverage` ≥ bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default new scene band |

## 13. Smoke check

`scripts/smoke/phase-39.sh`: the wrapped-title body-below-header guard, the
single-line byte-identical guard, and the scene determinism guard.

## 14. Tests

- **Unit (white-box):** `cardHeaderBottom`/`cardHeaderRowHeights` wrapped vs
  single-line; `renderCardChrome` body top ≥ wrapped header bottom.
- **Golden:** existing card goldens stay byte-identical (single-line).
- **Integration / Fuzz / Benchmark:** none.

## 15. Vocabulary added

- (none new) — reuses `wrappedLines`, card chrome vocabulary.

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-39.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (n/a — no new term).
- [x] Decision entries added (D-070).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated.
