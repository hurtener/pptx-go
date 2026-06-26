# Phase 97 — soul-driven dark palette

**Subsystem:** `pptx` (Layer 1 builder — Theme) + `scene` (Layer 2 — dark
variant derivation)
**RFC sections:** §7.1 (semantic color roles), §13.3 (theme variants)
**Deps:** none (foundational for Wave 15; Phase 100/R8.7 builds on it)
**Status:** Done

---

## 1. Goal

A theme can carry its own VariantDark palette so dark slides render in the
brand's deep hues (e.g. navy) instead of the pinned neutral gray — and a theme
that sets none is byte-identical to today.

## 2. Why now

First phase of Wave 15 (R8 theme/soul engine bits), the final wave of the
R1–R14 engine campaign. R8.3 is CRITICAL and foundational: the dark-palette
override seam it adds is what R8.7 (dark accent/extension overrides) extends,
and it is the largest single "template not designed" tell on dark slides
(`DECKARD-PRODUCT-REQUIREMENTS.md` R8.3). D-059 puts the override *mechanism*
here; the soul's dark-palette *derivation* is Deckard's half.

## 3. RFC sections implemented

- `RFC §7.1` — extends the semantic surface/text color roles with an optional
  per-variant (dark) override set.
- `RFC §13.3` — VariantDark theme derivation: `darkThemeFrom` now consumes a
  theme-supplied dark palette before falling back to the pinned default.

## 4. Brief findings incorporated

- `docs/research/80-soul-driven-dark-palette.md` — *overlay soul overrides
  after the pinned grays; nil = byte-identical* → `darkThemeFrom` keeps the six
  pinned assignments verbatim, then overlays `base.DarkColors` maps when
  non-nil.
- `docs/research/80-soul-driven-dark-palette.md` — *model the dark palette as
  the same dynamic map types so it overrides any role and pre-builds the R8.7
  seam* → `DarkPalette{Surfaces map[ColorRole]RGB; Text map[TextColorRole]RGB}`.
- `docs/research/80-soul-driven-dark-palette.md` — *the field needs no
  theme1.xml slot; G6 is about emitted output; `stats.Colors` already verifies
  the resolved dark palette* → acceptance asserts via the `stats.Colors` hook,
  no codec change.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-058` — `SlideColors` / `Stats.Colors` — the resolved-color hook that
  reports the dark palette a VariantDark slide actually rendered with; the
  acceptance vehicle here (and the R8.10 capstone gate).
- `D-104` — `ColorPaper` (role without a theme1.xml slot) — the precedent that
  a theme color field need not be serialized; the resolved RGB round-trips, the
  field does not.
- `D-026` — engine exposes a mechanism, not taste — the engine takes the dark
  palette as an explicit input and falls back to the pinned gray; deriving a
  brand dark palette from the light one is the soul's job.
- `D-059` — engine/product scope split — R8.3 is `both`; this phase ships the
  engine override mechanism, Deckard derives + seeds the dark palette.
- New decision **D-135** filed in this PR.

## 7. Architecture

```text
pptx.Theme
  + DarkColors *DarkPalette          // nil = pinned-gray fallback (byte-identical)
        DarkPalette{ Surfaces map[ColorRole]RGB; Text map[TextColorRole]RGB }
  + WithDarkSurface(role, c) / WithDarkText(role, c)   // lazily alloc DarkColors
  Clone()  // deep-copies DarkColors when non-nil

scene.darkThemeFrom(base):
  dark := base.Clone()
  <pinned darkCanvas/darkSurface/darkSurfaceAlt + dark text>   // unchanged
  if base.DarkColors != nil {                                  // R8.3 overlay
      overlay Surfaces / Text role-by-role
  }
```

The dark palette is consumed only by `scene` to derive a per-slide VariantDark
theme; it is never emitted to theme1.xml. P3 holds (no raw OOXML leaves
`internal/ooxml`); P2 holds (colors are tokens, documented in THEME.md).

## 8. Files added or changed

```text
pptx/theme.go                          # CHANGED — DarkPalette type, Theme.DarkColors, WithDarkSurface/WithDarkText, Clone deep-copy
pptx/theme_test.go                     # CHANGED — option + Clone deep-copy + -race concurrent reuse
scene/background.go                    # CHANGED — darkThemeFrom overlays base.DarkColors
scene/background_darkpalette_internal_test.go # NEW — white-box darkThemeFrom overlay / nil fallback
scene/render_darkpalette_test.go       # NEW — black-box stats.Colors == soul dark; nil byte-identical; determinism guard
scene/render_corpus_test.go            # CHANGED — a soul-dark corpus variant (on-canvas/conformant/deterministic)
scripts/smoke/phase-97.sh              # NEW — phase smoke
docs/research/80-soul-driven-dark-palette.md   # NEW — brief
docs/research/INDEX.md                 # CHANGED — registers brief 80
docs/plans/phase-97-soul-driven-dark-palette.md # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 15 detail + phase entry
docs/decisions.md                      # CHANGED — adds D-135
docs/glossary.md                       # CHANGED — DarkPalette / dark palette
docs/design/THEME.md                   # CHANGED — dark-palette mechanism + tokens
docs/site/reference/pptx.md            # CHANGED — DarkColors / WithDark* surface
skills/define-a-theme/SKILL.md         # CHANGED — dark-palette section
```

## 9. Public API surface

```go
// pptx
type DarkPalette struct {
    Surfaces map[ColorRole]RGB     // VariantDark surface overrides
    Text     map[TextColorRole]RGB // VariantDark text overrides
}
// Theme gains: DarkColors *DarkPalette
func WithDarkSurface(role ColorRole, c RGB) ThemeOption     // lazily allocs DarkColors
func WithDarkText(role TextColorRole, c RGB) ThemeOption     // lazily allocs DarkColors
```

No prior public surface breaks (additive field + new options).

## 10. Risks

- **R1 — byte-identity regression** — a non-nil-but-empty `DarkColors` must
  behave exactly like nil. **Mitigation:** the overlay iterates the maps, so an
  empty/nil map writes nothing; a byte-negative test asserts a default-theme
  dark slide is unchanged.
- **R2 — shallow clone aliases the dark maps** — `Clone()`'s `c := *t` copies
  the pointer. **Mitigation:** deep-copy `DarkColors` (both maps) when non-nil;
  a `-race` concurrent-reuse test proves no shared mutable state.
- **R3 — determinism** — overlay order must not matter. **Mitigation:** each
  role is written exactly once (pinned default then single overlay); a
  determinism guard renders a dark-palette slide across worker counts.

## 11. Acceptance criteria

1. A theme with `WithDarkSurface(ColorCanvas, navy)` renders every VariantDark
   slide's resolved canvas to `navy` (via `stats.Colors`); `ColorSurface` /
   `TextPrimary` overrides likewise resolve to the supplied hexes.
2. A theme with **no** `DarkColors` reproduces the current pinned-gray dark
   output byte-for-byte (a default-theme dark slide is unchanged).
3. `Theme.Clone()` deep-copies `DarkColors`: mutating a clone's dark maps does
   not affect the original (asserted under `-race` concurrent reuse).
4. The dark-palette slide renders byte-identically across worker counts
   (determinism guard).
5. `make coverage` shows `pptx` and `scene` ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default for `pptx` |
| `scene` | 80% | default for `scene` |

## 13. Smoke check

`scripts/smoke/phase-97.sh` verifies:

1. `OK:` `pptx.DarkPalette` + `Theme.DarkColors` exist (grep `pptx/theme.go`).
2. `OK:` `WithDarkSurface` / `WithDarkText` options exist.
3. `OK:` `darkThemeFrom` overlays `base.DarkColors` (grep `scene/background.go`).
4. `OK:` `go test ./pptx/ ./scene/` for the new tests pass.

## 14. Tests

- **Unit:** `pptx` (option + Clone deep-copy + `-race` concurrent reuse),
  `scene` white-box (`darkThemeFrom` overlay / nil byte-identical).
- **Round-trip golden:** yes — a VariantDark slide built on a soul dark palette
  is re-opened and its resolved dark canvas RGB asserted (via `stats.Colors` +
  byte assertion), plus the nil-byte-identical negative.
- **Integration:** no new cross-subsystem seam (the existing `scene`→`pptx`
  theme seam is exercised by the black-box render test).
- **Fuzz:** none (no new parse/decode surface).
- **Benchmark:** none.

## 15. Vocabulary added

- `DarkPalette` — a theme's optional VariantDark surface/text overrides;
  consumed by the scene renderer's `darkThemeFrom`, nil = the pinned gray
  default.

## 16. Plan deviations encountered during implementation

- The determinism guard ships in the dedicated black-box file
  `scene/render_darkpalette_test.go` (`TestDarkPalette_Determinism`) and the
  soul-dark corpus variant (`TestCorpus_SoulDarkPalette`), rather than as an
  edit to `scene/render_parallel_test.go`. Same acceptance criterion (4),
  cleaner colocation with the feature's other tests. The white-box overlay /
  nil-fallback test lives in a new `scene/background_darkpalette_internal_test.go`
  (`package scene`) since `darkThemeFrom` is unexported.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-97.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-135).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (define-a-theme).
