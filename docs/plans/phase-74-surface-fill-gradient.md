# Phase 74 — surface fill gradient

**Subsystem:** `scene` (Layer 2 — Card surface)
**RFC sections:** §11.1 (card), §7.1 (token color), §10.1 (backward-compat)
**Deps:** none; brief 57.
**Status:** Done

---

## 1. Goal

Add an optional 2-stop gradient fill to the `Card` surface so a card can carry
the subtle top-to-bottom depth shift pro decks use, defaulting to the solid
`Fill` (byte-identical).

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); flat card swatches are a visible
gap vs the reference's lit surfaces. Engine req R13.8 (MED · engine; D-059).

## 3. RFC sections implemented

- `RFC §11.1` — the card surface gains a gradient fill option.
- `RFC §7.1` — both stops are surface token roles (P2).
- `RFC §10.1` — unset gradient = solid fill = byte-identical.

## 4. Brief findings incorporated

- `docs/research/57-surface-fill-gradient.md` — *"one fill site in
  `renderCardChrome`"* → a single branch swaps `SolidFill` for `LinearGradient`.
- `57` — *"`*GradientFill` keeps byte-identity + unambiguous unset"* →
  `Card.FillGradient *GradientFill`, nil = solid.
- `57` — *"no engine auto-tint (D-026)"* → both `From`/`To` explicit; the
  soul owns any darker-`To` convenience.
- `57` — *"CardSection/Bento/Container deferred"* → Card-only field this phase.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-041` — gradient mechanism (`pptx.LinearGradient`).
- `D-026` — auto-tint is taste → soul, not engine.
- `D-108` (new) — files `Card.FillGradient`.

## 7. Architecture

`GradientFill{From, To pptx.ColorRole; Angle int}` + `Card.FillGradient
*GradientFill` thread into `cardChrome.fillGradient`. `renderCardChrome` picks
the surface fill: a `LinearGradient(Angle, {0,From},{1,To})` when set, else the
unchanged `SolidFill(TokenColor(c.fill))`. No new `BackgroundKind`, no OOXML
element, no `restorenamespaces` change.

```text
Card{Fill: ColorSurface, FillGradient: &GradientFill{From: ColorSurface, To: ColorSurfaceAlt, Angle: 90}}
  renderCardChrome → WithFill(LinearGradient(90, {0,ColorSurface},{1,ColorSurfaceAlt}))
Card{Fill: ColorSurface}  (FillGradient nil) → WithFill(SolidFill(ColorSurface))  (byte-identical)
```

## 8. Files added or changed

```text
scene/nodes.go                # CHANGED — GradientFill type + Card.FillGradient field
scene/render_card.go          # CHANGED — cardChrome.fillGradient + renderCardChrome fill branch
scene/render_card_test.go     # CHANGED/NEW — gradient card emits <a:gradFill>; solid byte-identical; determinism
scripts/smoke/phase-74.sh     # NEW — phase smoke
docs/research/57-surface-fill-gradient.md  # NEW — brief
docs/research/INDEX.md        # CHANGED — registers brief 57
docs/plans/phase-74-surface-fill-gradient.md  # NEW — this plan
docs/plans/README.md          # CHANGED — Phase 74 detail
docs/design/THEME.md          # CHANGED — surface gradient mechanism note
docs/glossary.md              # CHANGED — surface fill gradient term
docs/decisions.md             # CHANGED — adds D-108
docs/site/reference/scene.md  # CHANGED — Card.FillGradient + GradientFill
skills/compose-a-scene/SKILL.md  # CHANGED — Card.FillGradient
```

## 9. Public API surface

```go
// scene
type GradientFill struct { From, To pptx.ColorRole; Angle int } // 2-stop linear surface fill
// Card gains: FillGradient *GradientFill // nil = solid Fill
```

No prior surface breaks (additive pointer field; `Card` already non-comparable).

## 10. Risks

- **R1 — solid-fill drift.** **Mitigation:** nil `FillGradient` keeps the
  unchanged `SolidFill` branch; a byte-identity test pins it.
- **R2 — gradient over headerFill/ribbon interplay.** **Mitigation:** the
  gradient is the base surface fill only; header band / ribbon draw over it
  exactly as over a solid fill — covered by the existing card chrome tests.

## 11. Acceptance criteria

1. A card with `FillGradient` set renders a `<a:gradFill>` surface with the `From`/`To` role colors; a top-vs-bottom luminance delta is present.
2. A card with `FillGradient == nil` is byte-identical to the pre-Phase-74 build.
3. A gradient card re-renders byte-identically (determinism).
4. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-74.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `GradientFill` type + `Card.FillGradient` field.
3. `OK:` `renderCardChrome` gradient branch.
4. `OK:` gradient card emits `<a:gradFill>` test.
5. `OK:` solid card byte-identical test.

## 14. Tests

- **Black-box (`scene_test`):** a gradient card's surface XML has `<a:gradFill>`
  with the two role colors; a solid card is byte-identical; determinism guard.
- **Integration / Fuzz:** no.
