# Brief 71 — Native dataviz arcs: donut + gauge (R14.8 part 2)

> Informs Phase 88 (Wave 14). Engine req R14.8
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine; D-059). Completes R14.8 after
> Phase 87 (bar / bars / sparkline).

## 1. Motivating phase

R14.8's acceptance names a "donut at 0.92 [that] renders a 331-degree arc with
'92%' centered" and a gauge. The bar family (Phase 87) is rect/line-based; the
donut and gauge need a native arc, which the builder did not expose. Phase 88 adds
the arc seam and the two arc marks.

## 2. Subsystem / files

- `pptx/arc.go` (new) — `AddBlockArc` (the arc seam).
- `pptx/shape.go` — `Shape.props()` (to set the adjust guides).
- `scene/nodes.go` — `DataMarkKind` gains `Donut`/`Gauge`.
- `scene/render_datamark.go` — the donut/gauge composers.

## 3. Findings

- **`blockArc` is the right native primitive.** It is an annular (ring) sector
  with three adjust handles — `adj1` (start angle), `adj2` (end angle, both
  60000ths of a degree, 0° at 3 o'clock, clockwise), `adj3` (inner radius ×100000).
  It is in the vendored transitional XSD and round-trips (the `XAvLst`/`XShapeGuide`
  structs the `roundRect` adjust already uses). A new builder `AddBlockArc(box,
  startDeg, sweepDeg, innerRatio, opts)` sets the three guides — a genuinely new
  OOXML capability, so a new builder method is justified (P1); the scene composes
  it (no raw OOXML in `scene`, P3).
- **Value + remainder arcs avoid a hole.** A donut renders as a value `blockArc`
  (accent) + a remainder `blockArc` (track) that together close the ring — no
  center-hole ellipse, so no dependency on the surface color behind the mark (the
  leaf-surface limitation). A gauge is the same over a 270° range. The center label
  sits in the (transparent) hole via a centered text frame.
- **Angles are pinned + deterministic.** The donut starts at 270° (12 o'clock);
  the gauge opens at 135° over a 270° sweep. `angle60k` normalizes to `[0,360)` and
  converts to 60000ths (integer) — byte-stable. The value sweep is `value*360`
  (donut) or `value*270` (gauge).
- **Edge cases.** value=1 draws only the value arc (no remainder); value=0 draws
  only the track. Both render without panic.
- **Squared box.** The mark squares its box (largest centered square) so the ring
  is circular regardless of the slot aspect.
- **Verification posture.** The exact filled direction of `blockArc` follows the
  vendored OOXML semantics; the test pins the emitted adjust values (e.g. the value
  arc's `adj1 = 16200000` = 270°) and conformance, not a PowerPoint screenshot
  (consistent with the structural-only in-suite posture). A wrong direction would
  be a one-line angle fix.

## 4. Recommendations

- Builder: `pptx.AddBlockArc(box, startDeg, sweepDeg, innerRatio, opts)` +
  `angle60k`.
- `DataMarkKind` += `DataMarkDonut`, `DataMarkGauge`; composers draw value +
  remainder arcs + a centered label; `dataMarkPreferredHeight` returns a square-ish
  slot; `validate` treats Donut/Gauge as single-value (`Value` in `[0,1]`).
- Tests: a donut emits ≥2 `blockArc`s + `adj1=16200000` + a centered "92%"
  (conformant); a gauge; full/empty edge cases; determinism; an `AddBlockArc`
  builder round-trip. Catalog stays 30 (Kind values, not nodes). Glossary +
  visual-leaves + skill add the two kinds; `pptx.md` adds `AddBlockArc`. D-123.

## 5. Open questions

- A gauge needle / tick marks → V1.x (the filled value arc conveys the reading).
- A multi-segment donut (stacked categories) → V1.x (the single-value ring is the
  acceptance).
