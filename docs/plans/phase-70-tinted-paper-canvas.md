# Phase 70 — tinted paper canvas

**Subsystem:** `pptx` (Layer 1 builder — theme tokens)
**RFC sections:** §7.1 (token taxonomy), §7.3 (theme ↔ theme1.xml)
**Deps:** none (foundational token addition; opens Wave 13)
**Status:** Done

---

## 1. Goal

Add a `ColorPaper` surface token so a deck can paint a faintly tinted
off-white "paper" canvas distinct from pure white, settable per theme while
defaulting to the existing canvas (byte-identical).

## 2. Why now

Opens **Wave 13 — backgrounds & finish** (`docs/plans/README.md`). Every
later R13 background/decoration req composes the surface palette; the paper
token is the smallest, most foundational piece and the engine half of R13.1
(HIGH · both; D-059 — the soul auto-applying it is Deckard's product half).
Pro reference decks never use flat `#FFFFFF` for content; the warm/cool paper
tone is part of what reads "designed".

## 3. RFC sections implemented

- `RFC §7.1` — extends the `ColorRole` surface-token taxonomy with
  `ColorPaper`.
- `RFC §7.3` — documents `ColorPaper` as a role **without** a theme1.xml slot
  (keeps its default on read-back, the existing `TextMuted` convention).

## 4. Brief findings incorporated

- `docs/research/53-tinted-paper-canvas.md` — *"`Colors.Surfaces` is a map,
  not a fixed array; appending a role is safe"* → append `ColorPaper` after
  `ColorInfo`; `Clone`/`ResolveColor` need no change.
- `53` — *"no spare OOXML slot; `ColorPaper` keeps its default on read-back"*
  → no `writeSlots`/`themeFromPart` change; documented in THEME.md.
- `53` — *"G6 round-trip is about emitted output, not the token"* → round-trip
  test asserts the reopened full-slide rect carries the off-white RGB.
- `53` — *"default `FFFFFF` = canvas keeps decks byte-identical"* → adopted.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — Wave 2 scope split — the `ColorPaper` token + `WithPaper` option is
  the engine half; the soul emitting a paper tint and the composer setting
  `Background = {BackgroundColor, ColorPaper}` on light slides is Deckard's.
- `D-104` (new) — files the `ColorPaper` token + no-slot read-back rationale.

## 7. Architecture

`ColorPaper` is a pure semantic-token addition: it appends to the `ColorRole`
iota, gets a `DefaultTheme` value (= canvas), a `WithPaper(RGB)` option, and a
THEME.md/glossary entry. `BackgroundColor` already resolves any `ColorRole`
via `pptx.TokenColor(bg.Color)`, so no scene/render change is needed — a
`Background{Kind: BackgroundColor, Color: ColorPaper}` slide works the moment
the token exists. No new `BackgroundKind`, no OOXML element, no
`restorenamespaces` change.

```text
pptx.ColorRole: …ColorInfo, ColorPaper(new, last)
DefaultTheme().Colors.Surfaces[ColorPaper] = "FFFFFF"   (= canvas)
WithPaper(c) sets Surfaces[ColorPaper] = c
ResolveColor(ColorPaper) → map lookup (off-white when set)
BackgroundColor{Color: ColorPaper} → solidFill srgbClr (round-trips)
```

## 8. Files added or changed

```text
pptx/theme.go                 # CHANGED — ColorPaper iota value, DefaultTheme value, WithPaper option
pptx/theme_test.go            # CHANGED — ColorPaper default + WithPaper + Clone tests
scene/background_test.go      # CHANGED — ColorPaper background round-trip (off-white RGB survives Open)
scripts/smoke/phase-70.sh     # NEW — phase smoke
docs/research/53-tinted-paper-canvas.md  # NEW — brief
docs/research/INDEX.md        # CHANGED — registers brief 53
docs/plans/phase-70-tinted-paper-canvas.md  # NEW — this plan
docs/plans/README.md          # CHANGED — opens Wave 13 detail section
docs/design/THEME.md          # CHANGED — ColorPaper surface role + default + no-slot note
docs/glossary.md              # CHANGED — ColorPaper term
docs/decisions.md             # CHANGED — adds D-104
docs/site/reference/pptx.md   # CHANGED — ColorPaper in the surface-role list
skills/define-a-theme/SKILL.md  # CHANGED — ColorPaper in the token taxonomy
```

## 9. Public API surface

```go
// pptx
const ColorPaper ColorRole = … // appended after ColorInfo; off-white paper canvas
func WithPaper(c RGB) ThemeOption // sets the ColorPaper surface tint
```

No prior public surface breaks (append-only iota; new option).

## 10. Risks

- **R1 — iota reorder breaks byte-identity.** **Mitigation:** append
  `ColorPaper` *after* `ColorInfo` (last), so every existing `ColorRole` value
  is unchanged; covered by the existing background byte-identity tests.
- **R2 — role-indexed arrays break.** **Mitigation:** `grep` confirms no
  `[N]ColorRole` arrays / range-over-fixed-roles; `Surfaces` is a dynamic map.

## 11. Acceptance criteria

1. `pptx.ColorPaper` exists as a `ColorRole`; `DefaultTheme().ResolveColor(ColorPaper)` == `RGB("FFFFFF")` (= `ColorCanvas`).
2. `pptx.WithPaper(RGB("FAFAF8"))` sets the token; `ResolveColor(ColorPaper)` returns it; `Clone` carries it.
3. A `Background{Kind: BackgroundColor, Color: ColorPaper}` slide on a theme with `ColorPaper = FAFAF8` renders a full-slide rect whose `solidFill` is `FAFAF8`, and that color survives `pptx.Open` (G6 round-trip).
4. A `BackgroundColor{Color: ColorPaper}` slide on the default theme is byte-identical to a `ColorCanvas` one.
5. `make coverage` keeps `pptx` ≥ band; `make lint`/`preflight`/`check-mirror` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default; additive token, covered by theme + background tests |

## 13. Smoke check

`scripts/smoke/phase-70.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `ColorPaper` present in `pptx/theme.go`.
3. `OK:` `WithPaper` option present.
4. `OK:` `ColorPaper` default in `DefaultTheme`.
5. `OK:` `TestColorPaper*` pass.
6. `OK:` `TestBackground*Paper*` round-trip passes.

## 14. Tests

- **Unit:** `pptx` — `ColorPaper` default == canvas, `WithPaper` sets it,
  `Clone` carries it.
- **Round-trip golden:** yes — `scene/background_test.go` renders a
  `ColorPaper` (off-white) background, reopens, asserts the rect fill RGB.
- **Integration:** no (no cross-subsystem seam; pure token).
- **Fuzz:** no.
- **Determinism:** covered by existing background byte-identity test.
