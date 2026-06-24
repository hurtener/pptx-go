# Phase 71 — multi-stop background gradient

**Subsystem:** `scene` (Layer 2 renderer — Background)
**RFC sections:** §10.1 (backward-compat), §10.2 (warn-don't-fail)
**Deps:** none (extends the existing Background); brief 54.
**Status:** Done

---

## 1. Goal

Extend the scene `Background` from a fixed 2-role linear gradient to an
arbitrary-N (2..8) stop gradient at caller-chosen positions, so a brand's
multi-hue hero wash is expressible, keeping the legacy 2-role form
byte-identical.

## 2. Why now

Wave 13's background work (`docs/plans/README.md`); the multi-stop field is
the foundation R13.2 radial (Phase 72) builds on — both consume the same
`Stops` list. Engine req R13.3 (MED · engine; D-059).

## 3. RFC sections implemented

- `RFC §10.1` — additive field; empty `Stops` reproduces prior output exactly.
- `RFC §10.2` — invalid stops degrade to a `LayoutWarning`, never a panic.

## 4. Brief findings incorporated

- `docs/research/54-multistop-background-gradient.md` — *"`pptx.LinearGradient`
  is already variadic"* → scene-side field extension only, no builder change.
- `54` — *"empty-slice fallback keeps the 2-role path byte-identical"* →
  `Stops` empty → legacy `Gradient[2]` path verbatim.
- `54` — *"validation is render-time (D-026 warn-don't-fail)"* →
  `backgroundGradientStops` validates at render; invalid → warn + skip.
- `54` — *"the slice makes `Background` non-comparable"* → no `==` on
  `Background` anywhere; tests use byte-comparison / `reflect.DeepEqual`.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine/product split — engine extension; the soul choosing hues is
  Deckard's.
- `D-041` — gradient/rotation mechanism — reuses `pptx.LinearGradient`.
- `D-105` (new) — files the multi-stop `Background.Stops` field + validation.

## 7. Architecture

`Background.Stops []GradientStop` (a new scene type `{Pos float64; Color
pptx.ColorRole}`) supersedes `Gradient [2]ColorRole` when non-empty. A
`backgroundGradientStops` validator maps it to `[]pptx.GradientStop`
(2..8 ascending in [0,1]); `renderBackground` feeds them to the already-variadic
`pptx.LinearGradient(angle, stops...)`. Empty `Stops` keeps the legacy
two-`TokenColor` path → byte-identical. No new `BackgroundKind`, no OOXML
element, no `restorenamespaces` change.

```text
Background{Kind: BackgroundGradient,
           Stops: []GradientStop{{0,ColorAccent},{0.5,ColorAccentAlt},{1,ColorCanvas}},
           Angle: 45}
  → backgroundGradientStops → pptx.LinearGradient(45, gs0, gs1, gs2)
  → <a:gradFill><a:gsLst> 3×<a:gs> </a:gsLst><a:lin ang=…></a:gradFill>
Stops empty → LinearGradient(angle, Gradient[0], Gradient[1])  (byte-identical)
```

## 8. Files added or changed

```text
scene/background.go           # CHANGED — GradientStop type + Background.Stops field
scene/render.go               # CHANGED — renderBackground multi-stop path + backgroundGradientStops
scene/background_test.go      # CHANGED — multi-stop render/round-trip/validation/determinism + legacy byte-identical
scripts/smoke/phase-71.sh     # NEW — phase smoke
docs/research/54-multistop-background-gradient.md  # NEW — brief
docs/research/INDEX.md        # CHANGED — registers brief 54
docs/plans/phase-71-multistop-background-gradient.md  # NEW — this plan
docs/plans/README.md          # CHANGED — Phase 71 detail
docs/design/THEME.md          # CHANGED — multi-stop gradient mechanism note
docs/glossary.md              # CHANGED — multi-stop background gradient term
docs/decisions.md             # CHANGED — adds D-105
docs/site/reference/scene.md  # CHANGED — Background.Stops + GradientStop
skills/compose-a-scene/SKILL.md  # CHANGED — multi-stop background field
```

## 9. Public API surface

```go
// scene
type GradientStop struct { Pos float64; Color pptx.ColorRole } // one stop in a multi-stop background gradient
// Background gains:
//   Stops []GradientStop  // 2..8 ascending stops in [0,1]; supersedes Gradient when non-empty
```

`Background` becomes non-comparable (slice field) — a v0.x-acceptable change;
no exported API relied on `Background ==`.

## 10. Risks

- **R1 — legacy 2-role drift.** **Mitigation:** empty `Stops` takes the
  unchanged path; a byte-identity test pins it.
- **R2 — non-comparable `Background` breaks a test.** **Mitigation:** `grep`
  confirms no `==`; build + `-race` catch any.
- **R3 — invalid stops panic.** **Mitigation:** validator returns `(nil,
  false)` → warn + skip; tested with `<2`, `>8`, out-of-range, descending.

## 11. Acceptance criteria

1. A 3-stop `BackgroundGradient` renders a `<a:gradFill>` with exactly 3 `<a:gs>` stops at the specified positions; round-trips through `pptx.Open`.
2. Out-of-range / descending / `<2` / `>8` stops record exactly one `LayoutWarning` and emit no gradient shape (no panic).
3. A legacy 2-role `Gradient` background (empty `Stops`) is byte-identical to the pre-Phase-71 build.
4. Re-render of a multi-stop background is byte-identical (determinism).
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; additive field covered by background tests |

## 13. Smoke check

`scripts/smoke/phase-71.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `GradientStop` + `Background.Stops` present.
3. `OK:` `backgroundGradientStops` validator present.
4. `OK:` multi-stop render/round-trip test passes.
5. `OK:` invalid-stops warning test passes.
6. `OK:` legacy byte-identical test passes.

## 14. Tests

- **Unit (white-box `scene`):** `backgroundGradientStops` accepts 3-stop, rejects
  `<2`/`>8`/out-of-range/descending.
- **Round-trip golden (black-box `scene_test`):** 3-stop gradient emits 3 `<a:gs>`,
  survives `pptx.Open`; invalid → warning + no shape; legacy 2-role byte-identical.
- **Determinism:** render-twice byte-equality guard.
- **Integration / Fuzz:** no.
