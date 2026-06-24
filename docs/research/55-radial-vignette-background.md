# Brief 55 — Radial-vignette background (R13.2)

> Informs Phase 72 (Wave 13). Engine req R13.2
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine; D-059). Builds on R13.3
> multi-stop stops (Phase 71 / D-105).

## 1. Motivating phase

Phase 72 adds a center-out radial background so a dark hero/section/closing
slide gets a subtle spotlight/vignette (center fractionally lighter, edges
falling off) instead of a flat fill. The pptx `RadialGradient` primitive
already exists; the scene `Background` only exposed a 2-stop *linear* fill, so
radial backgrounds were unreachable.

## 2. Subsystem / files

- `scene/background.go` — `BackgroundKind` iota + `String`.
- `scene/render.go` — `renderBackground`; `backgroundGradientStops` (Phase 71).
- `pptx/fill.go` — `RadialGradient(...GradientStop)` emits
  `<a:gradFill>…<a:path path="circle"><a:fillToRect L/T/R/B=50000/></a:path>`
  (a centered 50%-inset focal). Already variadic; no builder change.

## 3. Findings

- **The builder is ready.** `pptx.RadialGradient` takes the same
  `...GradientStop` as `LinearGradient` and emits a centered circular focal.
  Phase 72 is a scene-side `BackgroundKind` addition + a render case — no
  builder change (P1).
- **Reuse the Phase-71 stops resolver.** Both linear and radial can take either
  the multi-stop `Stops` list or the legacy 2-role `Gradient` pair. Factor a
  shared `backgroundGradientStopsFor(bg)` that returns `Stops` (validated) when
  present, else the 2-role pair at Pos 0/1. Refactoring the existing
  `BackgroundGradient` case through it is **byte-identical** (same stop values).
- **Append the kind last.** `BackgroundRadial` after `BackgroundAsset` keeps
  every existing `BackgroundKind` value unchanged → byte-identical; existing
  slides never set it.
- **Focal offset is a builder gap.** Biasing the focal point needs a
  focal-rect knob on `pptx.RadialGradient` (it hard-codes the centered 50%
  inset). The req explicitly allows *"otherwise document center-only"*. V1
  ships **center-only** (the common vignette/spotlight case) and defers the
  focal-offset knob — a documented deviation (mirrors the Phase-65 ribbon
  diagonal-rotation deferral). No new `Background` field this phase.
- **No new OOXML element.** `<a:gradFill>`/`<a:path>`/`<a:gs>` already emit via
  the radial glow ornaments (D-041); no `restorenamespaces` change. Verify
  emitted bytes: assert `<a:path path="circle"` present and `GradientRead.Radial
  == true` after round-trip.
- **Determinism.** Pure integer-EMU/alpha through the existing fill path;
  worker-independent. A render-twice guard covers it.

## 4. Recommendations

- Add `BackgroundRadial` to `BackgroundKind` (last) + `String() == "radial"`.
- `renderBackground` `case BackgroundRadial`: resolve stops via
  `backgroundGradientStopsFor`; invalid explicit `Stops` → warn + skip (D-026);
  emit `ps.AddShape(ShapeRect, full, WithFill(pptx.RadialGradient(stops...)))`.
- Refactor `BackgroundGradient` through the same resolver (byte-identical).
- Document center-only focal (defer the offset knob); THEME.md note, glossary,
  compose-a-scene skill, docs/site scene.md. D-106.

## 5. Open questions

- Focal-offset knob (`pptx.RadialGradient` focal-rect parameter) → deferred to a
  later R13 phase or V2 if a real off-center-spotlight case appears; note in the
  decision.
