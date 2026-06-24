# Phase 72 — radial-vignette background

**Subsystem:** `scene` (Layer 2 renderer — Background)
**RFC sections:** §10.1 (backward-compat), §10.2 (warn-don't-fail)
**Deps:** Phase 71 (multi-stop stops resolver, D-105); brief 55.
**Status:** Done

---

## 1. Goal

Add a `BackgroundRadial` kind so a slide can carry a center-out radial
background (spotlight/vignette) via `pptx.RadialGradient`, reusing the Phase-71
stops, byte-identical when unused.

## 2. Why now

Wave 13 backgrounds (`docs/plans/README.md`); the multi-stop stops landed in
Phase 71, so radial is the natural follow-up — both consume the same `Stops`
list. Engine req R13.2 (HIGH · engine; D-059).

## 3. RFC sections implemented

- `RFC §10.1` — additive `BackgroundKind` value; existing slides unaffected.
- `RFC §10.2` — invalid explicit stops degrade to a `LayoutWarning`, never a panic.

## 4. Brief findings incorporated

- `docs/research/55-radial-vignette-background.md` — *"`pptx.RadialGradient` is
  ready"* → scene-side kind + render case, no builder change.
- `55` — *"reuse the Phase-71 stops resolver; refactoring the linear case
  through it is byte-identical"* → shared `backgroundGradientStopsFor`.
- `55` — *"append the kind last → byte-identical"* → `BackgroundRadial` after
  `BackgroundAsset`.
- `55` — *"focal offset is a builder gap; ship center-only, document it"* →
  center-only radial; no new `Background` field; deferral noted in D-106.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine/product split — engine extension.
- `D-041` — gradient mechanism — reuses `pptx.RadialGradient`.
- `D-105` — multi-stop stops — the radial path consumes the same `Stops`.
- `D-106` (new) — files `BackgroundRadial` + the center-only focal deferral.

## 7. Architecture

`BackgroundRadial` joins `BackgroundKind` (last). `renderBackground` resolves a
background's stops via a shared `backgroundGradientStopsFor(bg)` (multi-stop
`Stops` validated, else the legacy 2-role pair) and feeds them to
`pptx.RadialGradient(stops...)`, which emits a centered 50%-inset circular
focal. The existing `BackgroundGradient` case is refactored through the same
resolver — byte-identical. Center-only focal; the focal-offset knob is deferred
(no new field this phase).

```text
Background{Kind: BackgroundRadial,
           Stops: []GradientStop{{0,ColorSurface},{1,ColorCanvas}}}  // dark center→darker edge
  → backgroundGradientStopsFor → pptx.RadialGradient(gs0, gs1)
  → <a:gradFill><a:gsLst>…</a:gsLst><a:path path="circle"><a:fillToRect .../></a:path></a:gradFill>
Stops empty → 2-role Gradient pair at Pos 0/1 (also works for radial)
```

## 8. Files added or changed

```text
scene/background.go           # CHANGED — BackgroundRadial kind + String
scene/render.go               # CHANGED — renderBackground radial case + backgroundGradientStopsFor (shared resolver)
scene/background_test.go      # CHANGED — radial render/round-trip (Radial==true), legacy-2-role radial, invalid-stops, determinism, String
scripts/smoke/phase-72.sh     # NEW — phase smoke
docs/research/55-radial-vignette-background.md  # NEW — brief
docs/research/INDEX.md        # CHANGED — registers brief 55
docs/plans/phase-72-radial-vignette-background.md  # NEW — this plan
docs/plans/README.md          # CHANGED — Phase 72 detail
docs/design/THEME.md          # CHANGED — radial background mechanism note
docs/glossary.md              # CHANGED — radial background term
docs/decisions.md             # CHANGED — adds D-106
docs/site/reference/scene.md  # CHANGED — BackgroundRadial kind
skills/compose-a-scene/SKILL.md  # CHANGED — radial background kind
```

## 9. Public API surface

```go
// scene
const BackgroundRadial BackgroundKind = … // appended after BackgroundAsset; center-out radial fill
```

No prior surface breaks (append-only iota).

## 10. Risks

- **R1 — linear-case refactor drift.** **Mitigation:** the shared resolver
  produces identical stops for the 2-role path; existing
  `TestBackground_GradientSlide` / `TestBackground_LegacyGradientByteIdentical`
  pin it.
- **R2 — focal-offset expectation.** **Mitigation:** center-only is an
  acceptance-valid vignette; the offset knob is explicitly deferred (D-106).

## 11. Acceptance criteria

1. A `BackgroundRadial` slide renders a `<a:gradFill>` with `<a:path path="circle">` (a centered focal); `GradientRead.Radial == true` after `pptx.Open`.
2. Radial works with either an explicit `Stops` list or the legacy 2-role `Gradient`.
3. Invalid explicit stops record one `LayoutWarning` and emit no fill (no panic).
4. `BackgroundRadial` re-render is byte-identical; an unused kind is byte-identical to today (existing linear/legacy paths unchanged).
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; additive kind covered by background tests |

## 13. Smoke check

`scripts/smoke/phase-72.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `BackgroundRadial` kind + `String`.
3. `OK:` radial render case + shared resolver.
4. `OK:` radial render/round-trip test passes.
5. `OK:` legacy-2-role radial + invalid-stops + determinism tests pass.

## 14. Tests

- **Round-trip golden (black-box `scene_test`):** radial (multi-stop) emits
  `<a:path path="circle">` and round-trips with `Radial == true`; legacy 2-role
  radial works; invalid stops warn + skip; determinism guard; `String()` ==
  "radial".
- **Integration / Fuzz:** no.
