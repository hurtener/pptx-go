# Phase 56 — stat-value overflow guard (R11.8)

**Subsystem:** scene — Layer 2 renderer (Stat leaf)
**RFC sections:** §11.1, §10.2
**Deps:** Phase 28 (D-057 Stat), Phase 43 (R10.5 `fitScale`/AutoFit), brief 39.
**Status:** Done

---

## 1. Goal

Keep a Stat value on a single line — via a pinned role ladder (TypeDisplay → H1 →
H2) plus a FontScale floor — so a wide value like "$4,000+" never wraps and crowds
the caption beneath it.

## 2. Why now

R11.8 is a Wave-11 HIGH unit (recreation slide 9). It refines R10.5's Stat AutoFit
shrink into a cleaner role ladder, reusing `wrappedLines` / `naturalWidthAt` /
`fitScale`.

## 3. RFC sections implemented

- `RFC §11.1` — the Stat leaf.
- `RFC §10.2` — content-aware fit (one-line value).

## 4. Brief findings incorporated

- `docs/research/39-stat-value-overflow-guard.md`:
  - "a pinned role ladder, gated on AutoFit" → `statValueFit` walks
    `[TypeDisplay, TypeH1, TypeH2]`, returns the first that fits one line, else the
    floor + `fitScale`; `(TypeDisplay, 0)` when AutoFit is off or the value fits.
  - "gate on AutoFit to preserve D-074" → AutoFit-off and AutoFit-on-fitting are
    byte-identical; the existing AutoFit tests stay green.
  - "stack-height clamp deferred" → needs `slideID` into `renderStat`; the one-line
    value fix removes the reported caption-crowding.

## 5. Findings I'm departing from

*"none"* — the optional stack-height clamp is deferred (documented).

## 6. Decisions referenced

- `D-057` — the Stat leaf.
- `D-074` — `fitScale` / AutoFit (the role ladder refines its Stat path).
- `D-088` — **new** — the role-ladder value guard.

## 7. Architecture

```text
statValueFit(autofit, value, boxW):
  if !autofit || value=="" → (TypeDisplay, 0)              # byte-identical
  for role in [Display, H1, H2]:
    if wrappedLines(value, role, boxW) <= 1 → (role, 0)
  → (TypeH2, fitScale(naturalWidthAt(value, H2), boxW))    # floor + shrink
renderStat → value run uses the returned (role, FontScale)
```

## 8. Files added or changed

```text
scene/render_stat.go               # CHANGED — statValueFit role ladder + value run
scene/render_stat_overflow_test.go # NEW — ladder steps + one-line-at-each-width
scripts/smoke/phase-56.sh          # NEW — phase smoke
docs/research/39-stat-value-overflow-guard.md  # NEW — brief 39
docs/research/INDEX.md             # CHANGED — registers brief 39
docs/plans/phase-56-stat-value-overflow-guard.md  # NEW — this plan
docs/plans/README.md               # CHANGED — Wave 11 Phase 56
docs/decisions.md                  # CHANGED — adds D-088
docs/glossary.md                   # CHANGED — Stat node fit note
```

No new exported symbol.

## 9. Public API surface

None. `statValueFit` / `statValueRoleLadder` are unexported. `Stat.AutoFit` (existing)
now drives a role ladder for the value instead of a within-role font scale.

## 10. Risks

- **R1 — changing the AutoFit-on wide-value bytes.** **Mitigation:** no golden pins the
  exact old AutoFit-on Stat sz; the existing test asserts only "not full display sz",
  which the ladder satisfies. AutoFit-off and AutoFit-on-fitting are byte-identical.
- **R2 — sub-floor overflow in a tiny box.** **Mitigation:** the `fitScale` 0.60 floor
  is a legibility bound (the spec's "or a floor is reached"); the one-line test uses
  widths above the floor.

## 11. Acceptance criteria

1. The value steps down the ladder (`Display → H1 → H2 → floor+scale`) as the box
   narrows; AutoFit-off is `(TypeDisplay, 0)` (`TestStatValueFit_RoleLadder`).
2. At the chosen role/scale the value fits one line across a width sweep (above the
   floor) (`TestStatValueFit_OneLine`).
3. An AutoFit over-wide Stat emits a reduced sz; AutoFit-off keeps the full display
   sz; AutoFit decks are deterministic (existing `TestAutoFit_Stat_EmitsReducedSz`,
   `TestAutoFit_Deterministic`).
4. `go test -race ./...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; ladder unit- and render-covered |

## 13. Smoke check

`scripts/smoke/phase-56.sh` runs the four acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` white-box (`render_stat_overflow_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "Stat node" already defined; its entry gains a value-fit note.)*

## 16. Plan deviations encountered during implementation

- The optional value+label+delta stack-height clamp is deferred (needs `slideID`
  into `renderStat`); the one-line value guarantee is the reported fix. Documented in
  brief 39 and D-088.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-56.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-088).
- [x] Docs site / skill updated (n/a — no surface change; glossary note suffices).
