# Phase 33 — display font

**Subsystem:** pptx — Layer 1 builder (theme/typography)
**RFC sections:** §7 (theme tokens / font scheme)
**Deps:** Phase 02 (theme). External: none.
**Status:** In progress

---

## 1. Goal

Add a first-class display font face (`Theme.DisplayFont`) so the `TypeDisplay`
role can use a distinct editorial face independent of `HeadingFont` — additive,
order-independent with `WithFonts`, byte-identical when omitted.

## 2. Why now

Wave 9 unit (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.2, HIGH; engine half — D-059).
Pro decks pair a serif display with a separate sans heading; the theme had only
heading + body faces, so a brand could not express it.

## 3. RFC sections implemented

- `RFC §7` — extends the theme font scheme with a third (display) face that the
  `TypeDisplay` role resolves against.

## 4. Brief findings incorporated

No new informing brief — a direct theme font-scheme field addition. The R9
typography context (the family system) is surveyed in
`docs/research/17-type-detail-tokens.md`.

## 5. Findings I'm departing from

None.

## 6. Decisions referenced

- `D-059` — Wave-2 engine scope (engine half of R9.2). Files **D-063**.

## 7. Architecture

```text
pptx/theme.go  Theme += DisplayFont string; WithDisplayFont(family) ThemeOption;
               WithFonts made DisplayFont-aware (TypeDisplay = DisplayFont when set,
               else heading) so the two options are order-independent.
```

The family flows through the existing run `a:latin` emit — a display run renders
(and round-trips) with the display typeface; no OOXML or scene change.

## 8. Files added or changed

```text
pptx/theme.go                        # CHANGED — DisplayFont field + WithDisplayFont + WithFonts
pptx/theme_display_test.go           # NEW — resolution, omit-inherits, order-independence, render
scripts/smoke/phase-33.sh            # NEW — phase smoke
docs/plans/phase-33-display-font.md  # NEW — this plan
docs/decisions.md                    # CHANGED — adds D-063
docs/glossary.md                     # CHANGED — adds "Display font"
docs/design/THEME.md                 # CHANGED — display face note
docs/site/guide/theme.md             # CHANGED — WithDisplayFont (§19)
skills/define-a-theme/SKILL.md       # CHANGED — WithDisplayFont (§19)
```

## 9. Public API surface

```go
// pptx
type Theme struct { /* … */ DisplayFont string }
func WithDisplayFont(family string) ThemeOption
```

`WithFonts(heading, body string)` keeps its signature (no break); it becomes
`DisplayFont`-aware so order does not matter. New theme font-scheme field ⇒
`docs/design/THEME.md` entry (P2).

## 10. Risks

- **R1 — order dependence / clobber.** **Mitigation:** `WithFonts` reads
  `DisplayFont` when rewriting `TypeDisplay`; a test asserts both orders yield the
  same resolution.
- **R2 — byte-identity.** **Mitigation:** empty `DisplayFont` ⇒ `TypeDisplay`
  inherits `HeadingFont`; `DefaultTheme().DisplayFont` is empty; a test asserts
  it; the scene/integration suites (default theme) stay green.

## 11. Acceptance criteria

1. `WithFonts("Y","Z")` + `WithDisplayFont("X")` ⇒ `TypeDisplay` family X,
   `TypeH2`–`TypeH5` Y, body Z.
2. Omitting `WithDisplayFont` ⇒ `TypeDisplay` inherits `HeadingFont`
   (byte-identical); `DefaultTheme().TypeDisplay` family == `HeadingFont`.
3. `WithDisplayFont` and `WithFonts` are order-independent.
4. A `TypeDisplay` run renders with the display typeface.
5. `make coverage` ≥ band; `make preflight` + `make lint` pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default new builder API |

## 13. Smoke check

`scripts/smoke/phase-33.sh` — builds; DisplayFont resolves on TypeDisplay only
(+ order-independent); omit inherits HeadingFont; display run renders the face.

## 14. Tests

- **Unit:** `pptx` — resolution, omit-inherits, order-independence, render.

## 15. Vocabulary added

- `Display font` — the optional third theme font-scheme face used by
  `TypeDisplay` (`Theme.DisplayFont` / `WithDisplayFont`).

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `pptx`.
- [ ] `scripts/smoke/phase-33.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] `make lint` clean.
- [ ] Glossary + `docs/design/THEME.md` updated.
- [ ] Decision entry D-063 added.
- [ ] Docs site + `define-a-theme` skill updated (§19).
