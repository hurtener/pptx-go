# Brief 54 — Multi-stop background gradient field (R13.3)

> Informs Phase 71 (Wave 13). Engine req R13.3
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059). Foundation for
> R13.2 radial backgrounds (Phase 72).

## 1. Motivating phase

Phase 71 extends the scene `Background` from a fixed 2-role linear gradient
(`Gradient [2]pptx.ColorRole`) to an arbitrary-N (2..8) stop gradient at
caller-chosen positions, so a brand's hero wash (2, 3, or 4 hues) is
expressible. The pro reference cover is a soft multi-hue wash; today
`Background.Gradient` caps it at two stops, so the recreation renders flat.

## 2. Subsystem / files

- `scene/background.go` — `Background` struct + a new scene `GradientStop`.
- `scene/render.go` — `renderBackground` `BackgroundGradient` case.
- `pptx/fill.go` — `LinearGradient(angle, ...GradientStop)` **already**
  accepts variadic stops and clamps `Pos` to `[0,1]`; no builder change.

## 3. Findings

- **The builder is already N-stop.** `pptx.LinearGradient` /
  `pptx.RadialGradient` take `...GradientStop`; the scene `Background` is the
  only thing capping it at two. The phase is a scene-side field extension, not
  a builder change (P1 — no new OOXML capability needed).
- **Back-compat via an empty-slice fallback.** Adding `Stops []GradientStop`
  (a scene type `{Pos float64; Color pptx.ColorRole}`) keeps `Gradient
  [2]pptx.ColorRole` intact. When `Stops` is empty (the zero value),
  `renderBackground` uses the legacy 2-role path verbatim → byte-identical.
  When `Stops` is non-empty it supersedes `Gradient`.
- **Validation is render-time, not Stage-1/2.** `scene/validate.go` does not
  validate backgrounds today (the asset-unresolved case already warns at
  render time in `renderBackground`). Per D-026, invalid stops (`<2`, `>8`,
  out of `[0,1]`, or not strictly ascending) record a `LayoutWarning` and skip
  the fill — no panic, consistent with the existing background-asset warning.
- **`Background` becomes non-comparable.** A slice field means `==` no longer
  works on `Background`/`SceneSlide`; any test comparing them must use
  `reflect.DeepEqual` (rendered-byte comparison is unaffected). `grep` shows
  no `==` on `Background` in non-test code.
- **Determinism.** Mapping `Stops` 1:1 into `pptx.GradientStop` is pure
  integer-EMU/alpha through the existing fill path (`Pos*100000` rounded);
  worker-count-independent. A determinism guard (render twice, bytes equal)
  covers it.
- **No new OOXML element.** The fill emits `<a:gradFill>`/`<a:gs>`, already
  registered (the legacy 2-stop path emits them); no `restorenamespaces`
  change. Verify emitted bytes: assert `strings.Count(xml, "<a:gs ")` == N.

## 4. Recommendations

- Add `scene.GradientStop{Pos float64; Color pptx.ColorRole}` and
  `Background.Stops []GradientStop`.
- Add a `backgroundGradientStops(in) ([]pptx.GradientStop, bool)` validator:
  `2..8` stops, each `Pos ∈ [0,1]`, strictly ascending → map to
  `pptx.TokenColor(role)`; else `false`.
- `renderBackground`: `Stops` non-empty → validate → `LinearGradient(angle,
  stops...)` or warn+skip; empty → legacy path (byte-identical).
- No adversarial-fixture extension: a full-bleed background gradient cannot
  overflow or mis-contrast content (that harness is content-fit/contrast). Add
  a determinism guard + black-box stop-count/round-trip tests instead.
- Document: THEME.md gradient mechanism note, glossary, compose-a-scene skill,
  docs/site catalog. D-105.

## 5. Open questions

- Per-stop `Pos` need not pin the endpoints to exactly 0/1 — `[0,1]` ascending
  is enough; PowerPoint renders interior-only stops fine. Adopted.
- Radial reuse (R13.2) consumes the same `Stops` field + a new
  `BackgroundRadial` kind → Phase 72.
