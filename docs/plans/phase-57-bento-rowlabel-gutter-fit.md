# Phase 57 — bento row-label gutter fit (R11.9)

**Subsystem:** scene — Layer 2 renderer (Bento container)
**RFC sections:** §11.2
**Deps:** Phase 27 (D-056 Bento), Phase 22 (`naturalWidth`), brief 40.
**Status:** Done

---

## 1. Goal

Size the bento row-label gutter to its widest label (clamped to a min/max) instead
of a fixed `In(1.2)`, so labels like "Control plane" don't wrap awkwardly and short
labels don't waste a wide gutter.

## 2. Why now

R11.9 is a Wave-11 MED unit (recreation slide 6), mirroring the R11.5/R11.7
fit-to-label pattern with `naturalWidth`.

## 3. RFC sections implemented

- `RFC §11.2` — the Bento container (row-label gutter).

## 4. Brief findings incorporated

- `docs/research/40-bento-rowlabel-gutter-fit.md`:
  - "one shared fit function" → `bentoGutterWidthOf(theme, v)` used by both
    `bentoColumns` (the drawn gutter) and the `preferredHeight` Bento estimate.
  - "thread theme into the bento geometry" → `bentoColumns`/`bentoGeometry` gain a
    `theme` parameter.
  - "not byte-identical, by design" → the gutter resizes for most label sets;
    existing bento tests assert gutter-presence / span ratios / equal heights
    (gutter-width-independent), so they pass.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-056` — the Bento container.
- `D-072` — the weighted-row geometry that shares `bentoColumns`.
- `D-089` — **new** — the fit-to-label gutter.

## 7. Architecture

```text
bentoGutterWidthOf(theme, v) = labeled ? clamp(max naturalWidth(label@TypeCaption) + 2·padX,
                                               bentoGutterMinW, bentoGutterMaxW) : 0
  ├── bentoColumns(box, v, gap, theme)        → the drawn gutter
  └── preferredHeight (Bento)                 → the slot estimate (same gutter → parity)
```

## 8. Files added or changed

```text
scene/render_bento.go             # CHANGED — bentoGutterWidthOf + theme-threaded bentoColumns/bentoGeometry
scene/render.go                   # CHANGED — Bento preferredHeight uses the fitted gutter
scene/render_bento_test.go        # CHANGED — bentoGeometry calls pass theme
scene/render_bounds_test.go       # CHANGED — bentoGeometry call passes theme
scene/render_bento_gutter_test.go # NEW — gutter fit/clamp + geometry-uses-fit
scripts/smoke/phase-57.sh         # NEW — phase smoke
docs/research/40-bento-rowlabel-gutter-fit.md  # NEW — brief 40
docs/research/INDEX.md            # CHANGED — registers brief 40
docs/plans/phase-57-bento-rowlabel-gutter-fit.md  # NEW — this plan
docs/plans/README.md              # CHANGED — Wave 11 Phase 57
docs/decisions.md                 # CHANGED — adds D-089
docs/glossary.md                  # CHANGED — Bento gutter fit note
```

No new exported symbol (`bentoGeometry`/`bentoColumns` are unexported).

## 9. Public API surface

None. The user-visible effect is that a `BentoRow.Label` gutter sizes to its text.

## 10. Risks

- **R1 — estimator/layout drift.** **Mitigation:** both route through
  `bentoGutterWidthOf`; a white-box test asserts the geometry gutter ==
  `bentoGutterWidthOf`.
- **R2 — a label longer than the cap.** **Mitigation:** clamped to `bentoGutterMaxW`;
  out-of-scope beyond (the labels render anchor-middle and may wrap within the row).

## 11. Acceptance criteria

1. The gutter is 0 unlabeled, the minimum for a short label, `naturalWidth + 2·padX`
   for a medium one, the cap for a very long one (`TestBentoGutterWidth_FitsLabels`).
2. The geometry reserves exactly `bentoGutterWidthOf`, and the widest label fits
   inside it (`TestBentoGutter_GeometryUsesFit`).
3. Labeled bentos render deterministically (existing `TestBento_Deterministic`).
4. `go test -race ./...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; gutter unit-covered |

## 13. Smoke check

`scripts/smoke/phase-57.sh` runs the three acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` white-box (`render_bento_gutter_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "Bento" already defined; its entry gains a gutter-fit note.)*

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-57.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-089).
- [x] Docs site / skill updated (n/a — no surface change; glossary note suffices).
