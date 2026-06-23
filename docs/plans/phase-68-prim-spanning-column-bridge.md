# Phase 68 — prim spanning column bridge

**Subsystem:** scene — Layer 2 renderer (TwoColumn field extension)
**RFC sections:** §11.2 (TwoColumn), §12 (policy)
**Deps:** Phase 26 (column join, D-055), Phase 65 (ribbon band-reserve pattern); brief 51.
**Status:** Done

---

## 1. Goal

Extend `TwoColumn.Join` with a `JoinPosition` so the join can be a horizontal accent
bracket spanning the top (or bottom) of both columns — the "one X, two ways" header —
instead of only a centered seam element.

## 2. Why now

R12.8 is a MED Wave-12 primitive. The recreation collapsed the spanning bridge into a tiny
seam circle with "One age nt" wrapped mid-word. See `docs/plans/README.md` Wave 12 and
D-059 (engine-tagged).

## 3. RFC sections implemented

- `RFC §11.2` — extends `TwoColumn` with the additive `JoinPosition` field.
- `RFC §12` — native policy (accent rects + a label pill); no new asset.

## 4. Brief findings incorporated

- `docs/research/51-prim-spanning-column-bridge.md` — *"a `JoinPosition` field; the bridge
  reserves a band (ribbon-style); the bracket = line + 2 stubs + a content-fit label pill,
  no mid-word wrap"* → `renderColumnBridge` + the band reserve.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-055` — column join — the seam element `JoinSeam` preserves.
- `D-098` — ribbon — the band-reserve-shifts-content pattern reused.
- `D-026` — engine not product. New: `D-101` — prim-spanning-column-bridge (this PR).

## 7. Architecture

```text
TwoColumn.JoinPosition (JoinSeam=today / JoinTopBridge / JoinBottomBridge)
  bridging: colBox insets by bridgeBandH at the top/bottom edge; columns lay out there
  preferredHeight += bridgeBandH when bridging
  renderColumnBridge: accent spanning line (left.X .. right.Right()) + 2 end stubs
                      + content-fit RadiusFull label pill centered on the line (fitScale tail)
```

## 8. Files added or changed

```text
scene/nodes.go                # CHANGED — JoinPosition; TwoColumn.JoinPosition
scene/validate.go             # CHANGED — TwoColumn join position range
scene/render.go               # CHANGED — TwoColumn preferredHeight += band when bridging
scene/render_container.go     # CHANGED — renderTwoColumn band reserve + dispatch; renderColumnBridge
scene/render_columnbridge_test.go ; render_columnbridge_render_test.go # NEW
scene/render_adversarial_test.go     # CHANGED — a top-bridge TwoColumn
scripts/smoke/phase-68.sh     # NEW
docs/research/51-...md + INDEX.md ; docs/plans/phase-68-...md + README.md
docs/glossary.md ; docs/site/catalog/containers.md ; skills/compose-a-scene/SKILL.md ; docs/decisions.md (D-101)
```

## 9. Public API surface

```go
// scene
type JoinPosition int
const ( JoinSeam JoinPosition = iota; JoinTopBridge; JoinBottomBridge )
// TwoColumn gains: JoinPosition JoinPosition
```

## 10. Risks

- **R1 — byte-identity for the seam default.** **Mitigation:** `JoinSeam` (zero) keeps the
  D-055 path; the existing column-join tests pass unchanged.
- **R2 — bridge overlaps column content.** **Mitigation:** the band reserve shifts the
  columns; the adversarial fixture's on-canvas invariant covers it.
- **R3 — mid-word label wrap.** **Mitigation:** the pill is content-fit with a `fitScale`
  tail; a test asserts the label is one intact run.

## 11. Acceptance criteria

1. A top-bridge spans both columns with a single-line label pill (never mid-word wrap) and
   a bracket connecting to both column tops.
2. A `JoinSeam` (default) render is byte-identical to the prior column-join output.
3. An out-of-range `JoinPosition` fails Stage-1 validation.
4. Repeated renders are byte-identical.
5. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-68.sh` greps the new surface (JoinPosition, `JoinTopBridge`,
`renderColumnBridge`, the band reserve) and runs the white/black-box tests; SKIPs gracefully
before the code exists. OK ≥ the acceptance-criteria count, FAIL = 0.
