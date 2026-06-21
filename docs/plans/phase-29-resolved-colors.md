# Phase 29 — resolved colors

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §10.1 (Render + Stats), §13.3 (theme variants)
**Deps:** Phase 05–06 (scene spine + Stats), the `VariantDark` derived palette.
External: none.
**Status:** In progress

---

## 1. Goal

Expose, per slide in `Stats`, the canvas / surface / primary-text colors the
engine actually resolved — including a `VariantDark` slide's derived dark
palette — so a caller can compute true contrast against the real background. The
engine performs no contrast logic.

## 2. Why now

Final unit of **Wave 8** (`DECKARD-PRODUCT-REQUIREMENTS.md` R7, LOW). The product
validates text/surface contrast but cannot see the colors the engine resolved per
slide; for `VariantDark` it checks against the light theme and false-flags
white-on-dark. R7 closes that visibility gap. Picked up after R1–R6 per the
one-requirement-per-PR cadence; completing it finishes Wave 8 (R1–R7).

## 3. RFC sections implemented

- `RFC §10.1` — extends the `Stats` observability surface with a per-slide
  resolved-color slice (no event protocol — D-016).
- `RFC §13.3` — surfaces the per-slide theme variant's *resolved* palette
  (the derived dark theme for `VariantDark`).

## 4. Brief findings incorporated

- `docs/research/16-resolved-colors.md` — *capture from the per-slide theme,
  after compose* → `composeOne` reads `sr.theme` after `composeSlide` returns
  (dark for dark slides), with no new plumbing.
- `docs/research/16-resolved-colors.md` — *canvas / surface / primary-text are
  the right three* → `SlideColors{Canvas, Surface, PrimaryText}`.
- `docs/research/16-resolved-colors.md` — *additive and output-invariant* →
  `Stats.Colors` is pure metadata; rendered bytes are byte-identical.
- `docs/research/16-resolved-colors.md` — *deterministic and scene-ordered* →
  merged in `Render`'s scene-order results loop (like `Timings`).
- `docs/research/16-resolved-colors.md` — *no contrast logic* → the engine
  returns RGBs only; ratios/thresholds stay in the caller.

## 5. Findings I'm departing from

None. The brief's open-questions (more roles, per-node colors, a query API
instead of a `Stats` field) are explicitly deferred there.

## 6. Decisions referenced

- `D-026` — *Engine, not product.* Contrast is a caller judgment; the engine
  exposes resolved values and computes no ratio or pass/fail.
- `D-016` — `Stats` is the observability surface (no event protocol); the
  resolved colors land there.
- This plan files **D-058 — resolved per-slide colors** in `docs/decisions.md`.

## 7. Architecture

```text
scene/scene.go    SlideColors{SlideID, Canvas, Surface, PrimaryText pptx.RGB}   NEW
                  Stats.Colors []SlideColors                                     NEW field
                  Render: append results[i].colors in the scene-order merge loop

scene/render.go   slideResult: + colors SlideColors
                  composeOne: after composeSlide, capture from sr.theme —
                    Canvas=ResolveColor(ColorCanvas), Surface=ResolveColor(ColorSurface),
                    PrimaryText=ResolveTextColor(TextPrimary)
```

`sr.theme` is the per-slide theme the slide rendered with: `composeSlide` sets it
to the derived dark theme for `VariantDark` and leaves it as the active theme
otherwise, so the captured RGBs are exactly what the codec emitted with.

## 8. Files added or changed

```text
scene/scene.go                       # CHANGED — SlideColors type, Stats.Colors, Render merge
scene/render.go                      # CHANGED — slideResult.colors + composeOne capture
scene/render_colors_test.go          # NEW — light vs dark resolved colors, order, byte-invariant, determinism
scripts/smoke/phase-29.sh            # NEW — phase smoke
docs/research/16-resolved-colors.md  # NEW — informing brief
docs/research/INDEX.md               # CHANGED — registers brief 16
docs/plans/phase-29-resolved-colors.md # NEW — this plan
docs/plans/README.md                 # CHANGED — adds Phase 29 to Wave 8
docs/decisions.md                    # CHANGED — adds D-058
docs/glossary.md                     # CHANGED — adds "SlideColors"
docs/site/guide/scene.md             # CHANGED — Stats.Colors doc (§19)
skills/compose-a-scene/SKILL.md      # CHANGED — Stats.Colors note (§19)
```

## 9. Public API surface

```go
// scene (scene.go)
type SlideColors struct {
    SlideID     string   // the slide's ID, in scene order
    Canvas      pptx.RGB // resolved ColorCanvas (the slide's base background)
    Surface     pptx.RGB // resolved ColorSurface
    PrimaryText pptx.RGB // resolved TextPrimary
}

type Stats struct {
    // ... existing ...
    Colors []SlideColors // per-slide resolved colors, scene order (incl. dark variant)
}
```

Additive to `Stats` only — no new builder API, no new node, no new token, and no
change to emitted bytes. A new manifest/observability field ⇒ a smoke check lands
in this PR (§4.2).

## 10. Risks

- **R1 — wrong theme captured for dark slides.** **Mitigation:** capture from
  `sr.theme` (which `composeSlide` leaves as the dark theme), not the
  presentation theme (restored by defer); a test asserts a dark slide's resolved
  colors equal `darkThemeFrom(theme)`'s and differ from a light slide's.
- **R2 — output drift.** **Mitigation:** `Stats.Colors` is never emitted; a test
  asserts the rendered bytes are identical with and without reading `Colors`, and
  determinism holds across workers.
- **R3 — order/race.** **Mitigation:** merged in the scene-order results loop
  (like `Timings`); the existing parallel determinism guards cover it, plus a
  `Colors` order assertion.

## 11. Acceptance criteria

1. After `Render`, `Stats.Colors` has one entry per slide, in scene order, each
   carrying the slide's resolved `Canvas`, `Surface`, and `PrimaryText`.
2. A light slide's resolved colors equal the active theme's; a `VariantDark`
   slide's equal the derived dark palette and differ from the light slide's
   (canvas darker, primary text lighter).
3. Adding the field does not change the rendered `.pptx` bytes; `Colors` is
   deterministic across worker counts.
4. `make coverage` shows `scene` ≥ its band; `make preflight` + `make lint` pass.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry; new branches covered by
`render_colors_test.go`.

## 13. Smoke check

`scripts/smoke/phase-29.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` Stats.Colors has one scene-ordered entry per slide (criterion 1).
3. `OK:` dark-variant resolved colors differ from light + match the dark palette
   (criterion 2).
4. `OK:` Colors is deterministic across workers (criterion 3).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — light vs dark resolved colors (white-box vs
  `darkThemeFrom`), scene-order, byte-invariance, determinism (white/black box).
- **Round-trip golden:** N/A — no builder primitive / node / emitted shape.
- **Integration:** no — internal to `scene` Stats.
- **Fuzz / Benchmark:** no.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `SlideColors` — the per-slide resolved canvas/surface/primary-text colors the
  engine reports in `Stats.Colors`, so a caller can compute its own contrast
  (the derived dark palette for a dark-variant slide).

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-29.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] `make lint` clean.
- [ ] Glossary updated.
- [ ] Decision entry D-058 added.
- [ ] Docs site updated for `Stats.Colors` (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
