# Brief 39 — stat-value-overflow-guard

**Subsystem:** scene — Layer 2 renderer (Stat leaf)
**Authored:** 2026-06-22
**Motivating phase:** Phase 56 — stat-value-overflow-guard (R11.8, HIGH · engine)

## 1. Question

A Stat renders its value at a fixed `TypeDisplay` with no width awareness; a wide
value like "$4,000+" wraps to "$4,000 / +", and the stray line pushes down and
crowds the caption beneath it (recreation slide 9). How can the value render on a
single line — and stay byte-identical when it already fits and when AutoFit is off?

## 2. Prior art surveyed

- **`scene/render_stat.go`** — the value run uses `RunStyle{TypeRole: TypeDisplay,
  Bold, FontScale: displayRunScale(v.AutoFit, …)}`: R10.5's within-`TypeDisplay`
  shrink, gated on `AutoFit`.
- **`scene/metrics.go wrappedLines` / `naturalWidthAt` / `fitScale`** — the
  deterministic one-line measure and shrink primitives.
- **R10.5 / D-074** — `AutoFit` is an *opt-in* (a fitting value, and AutoFit-off, are
  byte-identical); `render_autofit_render_test.go` pins "AutoFit-off Stat emits the
  full display sz".
- DECKARD R11.8 spec: before emitting the value, `wrappedLines(value @ TypeDisplay,
  box.W)`; if > 1 step the value's effective size down `TypeDisplay → TypeH1 →
  TypeH2` (pinned ladder) until it fits one line or the floor is reached, then emit
  at the chosen role with no-wrap; optionally clamp the value+label+delta stack to
  box.H; byte-identical when the value already fits at TypeDisplay.

## 3. Findings

- **A pinned role ladder, gated on AutoFit.** R11.8 refines R10.5's Stat shrink:
  instead of scaling the font *within* `TypeDisplay`, walk the pinned ladder
  `[TypeDisplay, TypeH1, TypeH2]` and pick the first role whose value fits one line
  (`wrappedLines == 1`); if even the `TypeH2` floor wraps, return the floor plus a
  `fitScale` shrink to one line. Stepping through real type roles keeps the value on
  clean typographic steps (40 → 32 → 28 pt) before any sub-role scaling.
- **Gate on `AutoFit` to preserve the D-074 contract.** Making the ladder always-on
  would change an AutoFit-off wide value (the existing test pins it to the full
  `TypeDisplay` sz, and D-074 made shrink opt-in). Gating keeps AutoFit-off
  byte-identical and AutoFit-on-fitting byte-identical (`(TypeDisplay, 0)` either
  way), and uses the ladder only for AutoFit-on wide values — exactly the cases the
  caller asked to fit. The product (D-026) opts its Stats into AutoFit to get the
  guard.
- **Existing tests stay green.** `TestAutoFit_Stat_EmitsReducedSz`: AutoFit-off → full
  display ✓ (ladder gated off); AutoFit-on over-wide → not full display ✓ (steps to
  H1/H2). `TestAutoFit_OffByteIdentical`: "42" fits → `(TypeDisplay, 0)` on and off ✓.
  `TestAutoFit_Deterministic`: the ladder is pure integer → deterministic ✓.
- **The 0.60 floor is a legibility bound.** In a sub-~0.85" box even `TypeH2` ×0.60
  cannot fit a 7-char value on one line; reaching the floor (the spec's "or a floor
  is reached") and a residual overflow is accepted rather than shrinking to
  illegibility.
- **Stack-height clamp deferred.** The optional value+label+delta stack clamp needs a
  `slideID` to warn (`renderStat` does not receive one) and the one-line value fix
  already removes the reported caption-crowding (the wrapped value was the cause).
  Deferred as a documented follow-up.

## 4. Recommendation

Add `statValueFit(autofit, value, boxW) (TypeRole, float64)` (the ladder + `fitScale`
floor, gated on AutoFit, `(TypeDisplay, 0)` when off/fitting) and emit the value at
the returned role/scale. White-box tests: the ladder steps `Display → H1 → H2 →
floor+scale` as the box narrows, the value is one line at each width (above the
floor), AutoFit-off is `(TypeDisplay, 0)`. The existing AutoFit render + determinism
tests cover the rendered path. D-088 records the role ladder and the deferred stack
clamp.

## 5. Open questions

- Stack-height clamp: deferred (needs `slideID` plumbing into `renderStat`). The
  one-line value guarantee is the reported fix.
