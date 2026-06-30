# Phase 102 — soul→engine roundtrip verification (Wave-15 capstone)

**Subsystem:** `scene` (`Stats.Colors` hook + fidelity test)
**RFC sections:** §7.1 (color roles), §13.3 (variants), §8 (Stats observability)
**Deps:** Phases 97–101 (the full theme/soul surface); brief 85
**Status:** Done

---

## 1. Goal

Extend the per-slide resolved-color hook (`Stats.Colors`) to report the full
surface/accent/text set per variant, and prove with a capstone test that a
non-default brand soul (light + dark palette + multi-accent + gradient)
round-trips to the rendered colors on both variants — so brand drift is catchable.

## 2. Why now

The Wave-15 capstone, after the theme/soul mechanisms (R8.3–R8.7) it verifies. It
is the theme analogue of the R14.19 conformance corpus: a generalizable check
that what a soul declares is what the engine renders, across variants
(`DECKARD-PRODUCT-REQUIREMENTS.md` R8.10). D-059 puts the resolved-color hook +
the round-trip proof on the engine; the soul-fidelity *gate* (intended-vs-resolved)
is Deckard's.

## 3. RFC sections implemented

- `RFC §8` — extends the `Stats.Colors` observability hook (D-058) with more
  resolved roles.
- `RFC §7.1` / `§13.3` — verifies surface/accent/text roles resolve to the soul's
  per-variant values.

## 4. Brief findings incorporated

- `docs/research/85-soul-roundtrip-verification.md` — *`SlideColors` must stay
  comparable (the determinism test uses `==`); add scalar RGB fields only* →
  `SurfaceAlt` / `Accent` / `AccentAlt` / `TextAccent` (no slice).
- `docs/research/85-soul-roundtrip-verification.md` — *the new fields dark-resolve
  for free (composeOne captures from the post-swap dark theme)* → just more
  `ResolveColor` reads in the existing capture block.
- `docs/research/85-soul-roundtrip-verification.md` — *the fidelity comparison is
  Deckard's; the engine reports resolved colors + proves resolved == theme* → a
  round-trip test, no public fidelity helper.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-058` — `SlideColors` / `Stats.Colors` — the hook this phase extends.
- `D-135` / `D-136` / `D-137` — the dark palette / multi-accent / gradient surface
  the capstone fixture exercises.
- `D-026` / `D-059` — the engine reports resolved colors; Deckard owns the
  intended-vs-resolved gate.
- New decision **D-140** (+ the Wave-15 close: R8.1/.2/.8/.9 product) filed in
  this PR.

## 7. Architecture

```text
SlideColors  (+ SurfaceAlt, Accent, AccentAlt, TextAccent; all scalar RGB → comparable)
composeOne:  capture them from sr.theme (already the dark theme for VariantDark)

capstone test:  brand soul (WithAccents + WithPaper + DarkColors + WithGradient)
  → render light + dark slides
  → assert every SlideColors field == variant theme's token
       (active theme for light, darkThemeFrom(theme) for dark)
  → deliberate mismatch fails; identical across worker counts
```

`SlideColors` is pure metadata (never emitted) → all rendered bytes unchanged.

## 8. Files added or changed

```text
scene/scene.go                         # CHANGED — SlideColors gains SurfaceAlt/Accent/AccentAlt/TextAccent + doc
scene/render.go                        # CHANGED — composeOne captures the new resolved roles
scene/render_fidelity_test.go          # NEW — Wave-15 fidelity capstone (light+dark roundtrip, mismatch, determinism)
scripts/smoke/phase-102.sh             # NEW — phase smoke
docs/research/85-soul-roundtrip-verification.md  # NEW — brief
docs/research/INDEX.md                 # CHANGED — registers brief 85
docs/plans/phase-102-soul-roundtrip-verification.md  # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 15 phase entry + Wave-15 close
docs/decisions.md                      # CHANGED — adds D-140 (+ Wave-15 close)
docs/glossary.md                       # CHANGED — SlideColors updated
docs/site/reference/scene.md           # CHANGED — SlideColors new fields
skills/compose-a-scene/SKILL.md        # CHANGED — Stats.Colors fidelity note (if present)
```

## 9. Public API surface

```go
// scene — SlideColors gains (additive):
//   SurfaceAlt, Accent, AccentAlt, TextAccent pptx.RGB
```

Additive struct fields; `SlideColors` stays comparable. No break.

## 10. Risks

- **R1 — breaking comparability** — a slice field would break the determinism
  test's `==`. **Mitigation:** scalar RGB fields only; the existing determinism
  test passes and a new one covers the brand soul.
- **R2 — capture reads the wrong theme on dark** — **Mitigation:** `composeOne`
  captures after the dark-theme swap; the capstone asserts the dark slide against
  `darkThemeFrom(theme)` and that dark ≠ light.

## 11. Acceptance criteria

1. For a non-default brand soul, every light slide's resolved
   canvas/surface/surfaceAlt/accent/accentAlt/primaryText/textAccent equals the
   active theme's token, and every dark slide's equals the derived dark theme's.
2. The dark variant re-resolves the soul (dark canvas/accent/text differ from
   light; the dark accent is the soul's dark accent).
3. A deliberate token mismatch is caught by the fidelity comparison.
4. The extended `SlideColors` is identical across worker counts.
5. The full existing scene suite passes unchanged (additive metadata).

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-102.sh` verifies the extended `SlideColors` fields exist and
the fidelity capstone tests pass.

## 14. Tests

- **Unit / round-trip:** `scene` white-box fidelity capstone (light+dark
  roundtrip, mismatch negative, determinism). The existing scene goldens are the
  byte-identity proof.
- **Integration / Fuzz / Benchmark:** none.

## 15. Vocabulary added

none (extends the existing `SlideColors` / `Stats.Colors`).

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-102.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-140 + Wave-15 close).
- [x] Docs site / skill updated.
