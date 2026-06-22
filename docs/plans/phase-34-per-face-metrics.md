# Phase 34 — per-face width metrics

**Subsystem:** pptx (theme) + scene (estimator) — `RFC §3.3`, §7, §10.2
**RFC sections:** §7 (theme/FontSpec), §10.2 (content-bbox-driven layout)
**Deps:** Phase 02 (FontSpec), Phase 22 (wrap estimator). External: none.
**Status:** Done

---

## 1. Goal

Make the wrap/overflow estimator face-aware: a `FontSpec.AvgCharWidth` per role
(set by the soul for its bundled face) feeds `naturalWidth`/`wrappedLines`, with
the built-in `~0.5` sans factor as the fallback — so serif/display faces estimate
correctly while the default sans stays byte-identical.

## 2. Why now

Wave 9 unit (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.5, HIGH · engine; D-059). The
estimator pinned `avgCharWidthFactor=0.5` for the default sans; a non-sans face
mis-estimates advance widths, contributing to title/body overlaps (R9.5/R10
fit). Foundational for the R10 fit work.

## 3. RFC sections implemented

- `RFC §7` — carries a per-role face-advance metric on `FontSpec`.
- `RFC §10.2` — the content-bbox estimator consults it.

## 4. Brief findings incorporated

`docs/research/17-type-detail-tokens.md` (R9 type-system context). The estimator
becoming face-aware is the R9.5 piece of the same family system.

## 5. Findings I'm departing from

The spec calls for a "per-family table keyed by font family." I implement a
per-role `FontSpec.AvgCharWidth` field instead (D-064): deterministic (no mutable
global registry), theme-scoped, no invented engine factor table, consistent with
the other R9 `FontSpec` tokens. A built-in family→factor table can layer on later.

## 6. Decisions referenced

- `D-059` — Wave-2 engine scope. Files **D-064**.

## 7. Architecture

```text
pptx/theme.go    FontSpec += AvgCharWidth float64 (fraction of size; 0 = ~0.5 fallback)
scene/metrics.go naturalWidth: factor = spec.AvgCharWidth (>0) else avgCharWidthFactor
```

## 8. Files added or changed

```text
pptx/theme.go                          # CHANGED — FontSpec.AvgCharWidth
scene/metrics.go                       # CHANGED — naturalWidth uses per-face factor
scene/metrics_perface_test.go          # NEW — wider factor widens estimate; 0/0.5 = fallback
scripts/smoke/phase-34.sh              # NEW — phase smoke
docs/plans/phase-34-per-face-metrics.md# NEW — this plan
docs/decisions.md                      # CHANGED — adds D-064
docs/glossary.md                       # CHANGED — adds "Average char width"
docs/design/THEME.md                   # CHANGED — estimator metric note
skills/define-a-theme/SKILL.md         # CHANGED — FontSpec.AvgCharWidth (§19)
```

## 9. Public API surface

```go
// pptx
type FontSpec struct { /* … */ AvgCharWidth float64 } // estimator only; 0 = ~0.5 fallback
```

Not a visual token (never rendered) — a layout-estimator input on the type scale;
documented in `docs/design/THEME.md`.

## 10. Risks

- **R1 — byte-identity / estimate drift.** **Mitigation:** `AvgCharWidth <= 0`
  uses the 0.5 fallback; a test asserts explicit 0.5 == fallback and the default
  theme is unchanged (scene/integration suites green).
- **R2 — determinism.** **Mitigation:** soul-time constant, pure integer-truncated
  math (unchanged from today).

## 11. Acceptance criteria

1. A role with a larger `AvgCharWidth` yields a wider `naturalWidth` estimate.
2. `AvgCharWidth == 0` (or explicit 0.5) equals the built-in fallback estimate;
   the default sans is byte-identical.
3. `make coverage` ≥ band; `make preflight` + `make lint` pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default |
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-34.sh` — builds; estimator uses per-face `AvgCharWidth` +
fallback.

## 14. Tests

- **Unit:** `scene` (white-box) — wider factor widens the estimate; 0/0.5 = fallback.

## 15. Vocabulary added

- `Average char width` — `FontSpec.AvgCharWidth`, the per-face wrap-estimator metric.

## 16. Plan deviations encountered during implementation

- Per-role `FontSpec` field instead of a global family table (D-064 / §5).

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean.
- [x] `scripts/smoke/phase-34.sh` reports `OK ≥ 1` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] `make lint` clean.
- [x] Glossary + `docs/design/THEME.md` updated.
- [x] Decision entry D-064 added.
- [x] `define-a-theme` skill updated (§19).
