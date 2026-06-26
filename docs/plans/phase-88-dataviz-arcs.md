# Phase 88 — native dataviz arcs: donut + gauge (R14.8 part 2)

**Subsystem:** `pptx` (an `AddBlockArc` arc seam) + `scene` (DataMark donut/gauge)
**RFC sections:** §8.2/§8.3 (shape geometry), §11/§12 (IR node), §10.1 (backward-compat), §7.1 (tokens)
**Deps:** D-122 (DataMark); brief 71.
**Status:** Done

---

## 1. Goal

Complete `DataMark` with the two arc-based marks — a single-value donut/ring (the
"92%" headline) and a gauge — drawn as native `blockArc` ring sectors, no raster.

## 2. Why now

Finishes R14.8 (HIGH · engine) after Phase 87's bar family; the donut is the
requirement's headline acceptance. The arc-based marks needed a builder arc seam
the bar family did not (the §4.3 split queued in D-122).

## 3. RFC sections implemented

- `RFC §8.2/§8.3` — a native `blockArc` ring sector via adjust guides.
- `RFC §11/§12` — `DataMark` gains the `Donut`/`Gauge` kinds (native, no media).
- `RFC §10.1` — additive Kind values; existing DataMarks unaffected.
- `RFC §7.1` — arc + track colors are theme tokens (P2).

## 4. Brief findings incorporated

- `docs/research/71-dataviz-arcs.md` — *"blockArc is the right native primitive"*
  → `AddBlockArc` (adj1/adj2/adj3); it round-trips.
- `71` — *"value + remainder arcs avoid a hole"* → no center-hole ellipse, no
  surface-color dependency.
- `71` — *"angles are pinned + deterministic"* → `angle60k`; donut 270°, gauge
  135°/270°.
- `71` — *"verification posture"* → the test pins the emitted adjust values +
  conformance, not a PowerPoint screenshot.

## 5. Findings I'm departing from

- A gauge needle / tick marks and a multi-segment donut are deferred to V1.x (the
  filled value arc + centered label satisfy the acceptance).

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-122` — DataMark (the bar family) + the split that queued this.
- `D-040` — why arcs are a builder seam, not the icon translator (no arcs there).
- `D-123` (new) — files `AddBlockArc` + the donut/gauge kinds.

## 7. Architecture

Builder: `pptx.AddBlockArc(box, startDeg, sweepDeg, innerRatio, opts)` emits
`<a:prstGeom prst="blockArc">` with `adj1` = start angle, `adj2` = end angle
(60000ths of a degree, via `angle60k`), `adj3` = inner radius (×100000). Scene:
`DataMarkKind` += `DataMarkDonut`, `DataMarkGauge`. The donut draws a value
`blockArc` (accent, start 270°, sweep `value*360`) + a remainder `blockArc` (track)
closing the ring, with the `Label` centered in the hole; the gauge is the same over
a 270° range from 135°. Both square their box. `validate` treats Donut/Gauge as
single-value (`Value` in `[0,1]`); `dataMarkPreferredHeight` returns a square slot.

```text
DataMark{Donut, Value:0.92, Label:"92%"} → value blockArc(270°,331.2°) + track + "92%"
DataMark{Gauge, Value:0.5, Label:"50"}   → value blockArc(135°,135°) + track over 270°
```

## 8. Files added or changed

```text
pptx/arc.go                          # NEW — AddBlockArc + angle60k
scene/nodes.go                       # CHANGED — DataMarkKind += Donut, Gauge
scene/render_datamark.go             # CHANGED — donut/gauge composers + arc geom consts + dispatch + preferredHeight
scene/validate.go                    # CHANGED — Donut/Gauge are single-value
scene/render_datamark_arc_test.go    # NEW — donut/gauge/edge/determinism + AddBlockArc round-trip
scene/render_adversarial_test.go     # CHANGED — donut + gauge in the dataviz card
test/integration/roundtrip_test.go   # CHANGED — donut + gauge in the all-kinds fixture
scripts/smoke/phase-88.sh            # NEW — phase smoke
docs/research/71-dataviz-arcs.md     # NEW — brief
docs/research/INDEX.md               # CHANGED — registers brief 71
docs/plans/phase-88-dataviz-arcs.md  # NEW — this plan
docs/plans/README.md                 # CHANGED — Phase 88 detail
docs/glossary.md                     # CHANGED — donut/gauge in the DataMark term
docs/decisions.md                    # CHANGED — adds D-123
docs/site/catalog/visual-leaves.md   # CHANGED — DataMark donut/gauge kinds
docs/site/reference/pptx.md          # CHANGED — AddBlockArc
skills/compose-a-scene/SKILL.md      # CHANGED — DataMark donut/gauge
```

## 9. Public API surface

```go
// pptx
func (s *Slide) AddBlockArc(box Box, startDeg, sweepDeg, innerRatio float64, opts ...ShapeOption) *Shape
// scene — DataMarkKind gains: DataMarkDonut, DataMarkGauge
```

Additive; no break.

## 10. Risks

- **R1 — arc validity/round-trip.** **Mitigation:** `blockArc` is in the vendored
  XSD; a builder round-trip test asserts the adjust guides survive reopen; the
  schema check validates the emitted bytes.
- **R2 — determinism.** **Mitigation:** integer `angle60k`; a 1-vs-8-worker test.
- **R3 — arc direction.** **Mitigation:** the test pins the adjust values; a wrong
  filled direction is a one-line angle fix.

## 11. Acceptance criteria

1. A donut at 0.92 renders ≥2 native `blockArc`s with the value arc starting at
   270° (`adj1=16200000`) and "92%" centered; conformant, no warnings.
2. A gauge renders native `blockArc`s + a label; full/empty edge cases render.
3. `AddBlockArc` round-trips (the adjust guides survive write → reopen → re-write).
4. The arc marks are worker-count deterministic.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | the AddBlockArc seam |
| `scene` | 80% | the donut/gauge composers |

## 13. Smoke check

`scripts/smoke/phase-88.sh`: builds CGo-free; `AddBlockArc` / `DataMarkDonut` /
`DataMarkGauge` / the composers present; donut / gauge / edge / determinism /
blockArc-round-trip tests pass.

## 14. Tests

- **Black-box (`scene_test`):** a donut (≥2 blockArcs, `adj1=16200000`, centered
  label, conformant); a gauge; full/empty edge cases; determinism; an `AddBlockArc`
  builder round-trip.
- **Adversarial:** a donut + gauge in the dataviz card.
- **Integration:** donut + gauge in the all-kinds fixture (round-trips the
  blockArc adjust guides).
