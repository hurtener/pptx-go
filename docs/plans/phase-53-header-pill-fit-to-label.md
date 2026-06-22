# Phase 53 — header-pill fit-to-label (R11.5)

**Subsystem:** scene — Layer 2 renderer (card chrome)
**RFC sections:** §12.1
**Deps:** Phase 22 (`naturalWidth`), Phase 39 (R10.1 header geometry), Phase 43
(R10.5 `fitScale`/`FontScale`), brief 36.
**Status:** Done

---

## 1. Goal

Size a card header pill to its label on a single line — instead of a fixed `In(1.0)`
box that wraps a long label like "CUSTOMIZABLE" — for any caller-supplied pill text.

## 2. Why now

R11.5 is the first of the Wave-11 fit-to-label HIGH units (recreation slide 5: a
"CUSTOMIZABLE" pill wraps inside its chip). It reuses the established `naturalWidth`
estimator and the R10.5 `fitScale` shrink primitive.

## 3. RFC sections implemented

- `RFC §12.1` — card chrome (the header pill).

## 4. Brief findings incorporated

- `docs/research/36-header-pill-fit-to-label.md`:
  - "one shared width function" → `cardPillWidthOf(theme, pill, innerW)` called from
    both `cardHeaderColumnWOf` (the reservation) and `renderCardChrome` (the drawn
    pill), so they never drift.
  - "single-line guarantee via `fitScale`" → the pill run gets
    `FontScale = fitScale(naturalWidth, pillW − 2·padX)` (0 when it already fits).
  - "not byte-identical, by design" → every pill resizes from the fixed `In(1.0)`;
    determinism holds; existing pill tests assert presence/shape, not the fixed width.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-070` — the wrapped-header column (`cardHeaderColumnWOf`) the pill reservation
  feeds; it now reserves the fitted pill width.
- `D-074` — `fitScale` / `RunStyle.FontScale`, reused for the single-line guarantee.
- `D-085` — **new** — the pill fit-to-label mechanism.

## 7. Architecture

```text
cardPillWidthOf(theme, pill, innerW) = clamp(naturalWidth(pill@TypeCaption) + 2·cardPillPadX,
                                             cardPillMinW, innerW)
  ├── cardHeaderColumnWOf → reserves pillW + gapSM from the header text column
  └── renderCardChrome    → draws the pill at pillW; FontScale = fitScale(natW, pillW−2padX)
```

## 8. Files added or changed

```text
scene/render_card.go              # CHANGED — cardPillWidthOf + both pill sites + FontScale
scene/render_pill_fit_test.go     # NEW — width fit/clamp + reservation-matches-drawn
scene/render_parallel_test.go     # CHANGED — TestRenderDeterministic_PillFit
scripts/smoke/phase-53.sh         # NEW — phase smoke
docs/research/36-header-pill-fit-to-label.md  # NEW — brief 36
docs/research/INDEX.md            # CHANGED — registers brief 36
docs/plans/phase-53-header-pill-fit-to-label.md  # NEW — this plan
docs/plans/README.md              # CHANGED — Wave 11 Phase 53
docs/decisions.md                 # CHANGED — adds D-085
docs/glossary.md                  # CHANGED — adds "header pill" fit note
skills/compose-a-scene/SKILL.md   # CHANGED — HeaderPill sizes to its label
```

No new exported symbol (the existing `Card.HeaderPill` field's behavior improves).

## 9. Public API surface

None. `cardPillWidthOf` / `cardPillPadX` / `cardPillMinW` are unexported. The
user-visible effect is that `Card.HeaderPill` / `CardSection` pills size to their
text.

## 10. Risks

- **R1 — reservation/drawn drift.** **Mitigation:** both sites call the same
  `cardPillWidthOf`; a white-box test asserts `headerW == innerW − (pillW + gap)`.
- **R2 — a label too long even at full inner width.** **Mitigation:** clamp to
  `innerW` then `fitScale` shrinks the run to one line (down to the 0.60 floor).

## 11. Acceptance criteria

1. The pill width sizes to `naturalWidth(label) + 2·padX`, floored at the circular
   minimum, clamped to inner width; empty → 0 (`TestCardPillWidth_FitsLabel`).
2. The header-column reservation equals the drawn pill width
   (`TestCardPillWidth_ReservationMatchesDrawn`).
3. Pills render byte-identically across worker counts
   (`TestRenderDeterministic_PillFit`).
4. `go test -race ./...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; pill width unit- and render-covered |

## 13. Smoke check

`scripts/smoke/phase-53.sh` runs the three acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` white-box (`render_pill_fit_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "Header pill" already defined; its entry gains a fit-to-label note.)*

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-53.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-085).
- [x] Docs site updated (n/a — no new surface; behavior note in the skill).
- [x] Affected agent skill(s) updated (compose-a-scene HeaderPill note).
