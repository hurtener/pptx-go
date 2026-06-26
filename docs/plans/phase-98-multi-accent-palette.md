# Phase 98 — multi-accent brand palette

**Subsystem:** `pptx` (Theme) + `scene` (accent cycle + contrast cores)
**RFC sections:** §7.1 (semantic color roles)
**Deps:** Phase 97 (D-135 dark-palette seam pattern); brief 81
**Status:** Done

---

## 1. Goal

A theme can carry an ordered brand-accent palette (4+ coordinated hues) that the
scene renderer's per-element accent cycle reads, so a deck renders each timeline
phase / quadrant point / tree node / funnel stage / image pin in its true brand
hue — and a theme that sets none is byte-identical to today's pinned cycle.

## 2. Why now

Second phase of Wave 15. R8.4 is HIGH and unblocks the deferred R14.2 ChartStyle
palette and R14.14 per-section accent — both want an ordered accent palette to
index into. The reference deck rotates four-plus accents by meaning; the engine
exposes only three accent roles, so a multi-hue brand collapses to two-tone
(`DECKARD-PRODUCT-REQUIREMENTS.md` R8.4). D-059 puts the ordered-palette
mechanism + cycle here; the soul populating it from a brand source is Deckard's.

## 3. RFC sections implemented

- `RFC §7.1` — extends the semantic accent surface roles with an optional
  ordered brand-accent palette the per-element accent cycle resolves by index.

## 4. Brief findings incorporated

- `docs/research/81-multi-accent-palette.md` — *model the palette as ordered
  literal RGB indexed by position (index→hex), since a brand's 4th/5th hue has no
  spare role/slot* → `Theme.Accents []RGB`.
- `docs/research/81-multi-accent-palette.md` — *byte-identity pivots on "palette
  empty"; the new path is taken only when len(Accents) > 0* → `accentColorAt` /
  `accentRGBAt` fall back to `TokenColor`/`ResolveColor(timelineAccent(idx))`.
- `docs/research/81-multi-accent-palette.md` — *the contrast path needs a
  resolved-RGB core so a literal-hue accent stays legible* → factored
  `onSurfaceRGB` / `cellTextOnColor` out of `onCardSurface` / `cellTextOn`.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-135` — soul-driven dark palette — the precedent for a slot-less theme color
  field consumed only by the scene renderer (the accent palette follows it).
- `D-082` — `onCardSurface` auto-contrast — extended with an RGB-keyed core so a
  brand-accent fill (no `ColorRole`) gets the same legible text.
- `D-026` / `D-059` — the engine exposes the ordered-palette mechanism + index
  cycle; the soul populates the palette from a brand source (Deckard's half).
- New decision **D-136** filed in this PR.

## 7. Architecture

```text
pptx.Theme
  + Accents []RGB                  // ordered brand palette; empty = pinned cycle (byte-identical)
  + WithAccents(...RGB)            // copies the slice
  Clone()                         // deep-copies Accents

scene renderer:
  accentColorAt(idx) pptx.Color   // Accents[idx%n] (literal) else TokenColor(timelineAccent(idx))
  accentRGBAt(idx)   pptx.RGB     // Accents[idx%n] else ResolveColor(timelineAccent(idx))
  timelineAccent(idx) ColorRole   // unchanged — the pinned fallback cycle

  onCardSurface(role) → onSurfaceRGB(ResolveColor(role))   // RGB-keyed core
  cellTextOn(role)    → cellTextOnColor(ResolveColor(role))

  call sites (timeline / funnel / cycle / quadrant / tree / image-pin) route
  fill/stroke through accentColorAt, contrast text through cellTextOnColor(accentRGBAt).
```

## 8. Files added or changed

```text
pptx/theme.go                          # CHANGED — Theme.Accents, WithAccents, Clone deep-copy
pptx/theme_test.go                     # CHANGED — option + Clone + -race concurrent reuse
scene/render_timeline.go               # CHANGED — accentColorAt/accentRGBAt resolvers; timeline call site
scene/render_funnel_cycle.go           # CHANGED — funnel band + cycle node call sites
scene/render_image.go                  # CHANGED — highlight/pin call sites
scene/render_quadrant.go               # CHANGED — quadrant dot call site
scene/render_tree.go                   # CHANGED — tree node call site
scene/contrast.go                      # CHANGED — onSurfaceRGB RGB-keyed core
scene/render_table.go                  # CHANGED — cellTextOnColor RGB-keyed core
scene/render_accents_internal_test.go  # NEW — white-box resolver cycle / pinned fallback
scene/render_accents_test.go           # NEW — black-box brand hues / nil byte-identical / determinism / contrast-from-hue
scripts/smoke/phase-98.sh              # NEW — phase smoke
docs/research/81-multi-accent-palette.md       # NEW — brief
docs/research/INDEX.md                 # CHANGED — registers brief 81
docs/plans/phase-98-multi-accent-palette.md    # NEW — this plan
docs/plans/README.md                   # CHANGED — Wave 15 phase entry
docs/decisions.md                      # CHANGED — adds D-136
docs/glossary.md                       # CHANGED — Brand-accent palette
docs/design/THEME.md                   # CHANGED — accent palette mechanism
docs/site/reference/pptx.md            # CHANGED — Accents / WithAccents surface
skills/define-a-theme/SKILL.md         # CHANGED — multi-accent section
```

## 9. Public API surface

```go
// pptx
// Theme gains: Accents []RGB
func WithAccents(palette ...RGB) ThemeOption   // empty = no-op (byte-identical)
```

No prior public surface breaks (additive field + new option; the scene resolver
methods are unexported).

## 10. Risks

- **R1 — byte-identity regression in the contrast refactor** — routing
  `onCardSurface`/`cellTextOn` through RGB cores must preserve output.
  **Mitigation:** the role wrappers call the core with `ResolveColor(role)`, the
  exact value they used before; the existing timeline/funnel/quadrant/tree/table
  goldens pass unchanged.
- **R2 — shallow clone aliases the slice** — `Clone()`'s `c := *t` copies the
  slice header. **Mitigation:** deep-copy `Accents` when non-empty; a `-race`
  concurrent-reuse test proves no shared backing array.
- **R3 — palette path leaks into the empty case** — **Mitigation:** the
  resolvers branch on `len(Accents) > 0`; a nil-byte-identical render guard +
  white-box pinned-fallback test pin it.

## 11. Acceptance criteria

1. A theme with a four-hue `Accents` palette renders all four brand hues into a
   four-phase timeline's slide XML (each phase marker in its own hue), beyond the
   three accent roles.
2. A theme with no `Accents` reproduces the pinned five-role cycle output
   byte-for-byte (a default-theme deck is unchanged).
3. A funnel band filled with a dark brand accent hue gets inverse (light)
   contrast text — auto-contrast works from a literal hue (D-082 × R8.4).
4. `Theme.Clone()` deep-copies `Accents` (race-safe); a brand-palette deck is
   byte-identical across worker counts.
5. `make coverage` shows `pptx` and `scene` ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default |
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-98.sh` verifies:

1. `OK:` `Theme.Accents` + `WithAccents` exist (grep `pptx/theme.go`).
2. `OK:` `accentColorAt` / `accentRGBAt` resolvers exist (grep
   `scene/render_timeline.go`).
3. `OK:` the contrast RGB cores exist (`onSurfaceRGB`, `cellTextOnColor`).
4. `OK:` the brand-hue / nil-byte-identical / contrast / determinism tests pass.

## 14. Tests

- **Unit:** `pptx` (option + Clone + `-race`); `scene` white-box resolver cycle +
  pinned fallback.
- **Round-trip golden:** the existing timeline/funnel/cycle/quadrant/tree goldens
  pin byte-identity for the no-palette path; the new black-box tests assert the
  brand hues reach the bytes and are deterministic.
- **Integration:** no new cross-subsystem seam.
- **Fuzz / Benchmark:** none.

## 15. Vocabulary added

- `Brand-accent palette` — a theme's optional ordered list of accent hues
  (`Theme.Accents`) the scene accent cycle reads by index; empty = the pinned
  five-role cycle.

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-98.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-136).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (define-a-theme).
