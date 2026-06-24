# Phase 79 — focal glow behind a card

**Subsystem:** `scene` (Layer 2 — Card)
**RFC sections:** §11.2 (card), §14.2 (decoration), §10.1 (backward-compat)
**Deps:** Phase 73 (role glows, D-107); brief 62.
**Status:** Done

---

## 1. Goal

Add `Card.Backdrop *Decoration` so a focal card can carry a soft glow/halo
behind it, positioned on the card's actual box across any layout, byte-identical
when unset.

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); decorations anchor only to the
slide region today, so a halo behind the middle card is unreachable. Engine req
R13.10 (MED · engine; D-059). Composes the role glows (R13.5).

## 3. RFC sections implemented

- `RFC §11.2` — a card-relative backdrop decoration.
- `RFC §14.2` — reuses the decoration node + recipes.
- `RFC §10.1` — nil backdrop = byte-identical.

## 4. Brief findings incorporated

- `docs/research/62-focal-glow-backdrop.md` — *"`Card.Backdrop *Decoration`, drawn
  before the chrome with the card box as the region"* → adopted.
- `62` — *"`renderDecoration` is reusable as-is"* → one call, no renderer change.
- `62` — *"nil = byte-identical; Card already non-comparable"* → pointer field.
- `62` — *"caller sets Bleed for the halo; pitch-cap warning doesn't apply"* →
  documented.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-107` — the role-colored glow the backdrop uses.
- `D-113` (new) — files `Card.Backdrop`.

## 7. Architecture

`Card.Backdrop *Decoration`; `renderCard` draws it via `r.renderDecoration(ps,
box, *v.Backdrop, slideID)` **before** `renderCardChrome`, with the card's
computed (safe-area-clamped) box as the region. A center-anchored, larger,
bleeding `radial_glow` becomes a halo behind the card's fill. nil → nothing.

```text
Card{Backdrop: &Decoration{Preset: "radial_glow", Color: &accent, Opacity: 0.18,
                           Anchor: AnchorCenter, Size: {cardW+2in, cardH+2in}, Bleed: true}}
  renderCard: renderDecoration(cardBox) → radial ellipse behind, then card chrome on top
```

## 8. Files added or changed

```text
scene/nodes.go             # CHANGED — Card.Backdrop *Decoration
scene/render_card.go       # CHANGED — renderCard draws Backdrop before chrome
scene/render_card_test.go  # CHANGED/NEW — backdrop glow before fill (z-order); nil byte-identical; determinism
scripts/smoke/phase-79.sh  # NEW — phase smoke
docs/research/62-focal-glow-backdrop.md  # NEW — brief
docs/research/INDEX.md     # CHANGED — registers brief 62
docs/plans/phase-79-focal-glow-backdrop.md  # NEW — this plan
docs/plans/README.md       # CHANGED — Phase 79 detail
docs/design/THEME.md       # CHANGED — backdrop glow mechanism note
docs/glossary.md           # CHANGED — card backdrop term
docs/decisions.md          # CHANGED — adds D-113
docs/site/reference/scene.md  # CHANGED — Card.Backdrop (catalog note)
skills/compose-a-scene/SKILL.md  # CHANGED — Card.Backdrop
```

## 9. Public API surface

```go
// scene
// Card gains: Backdrop *Decoration // nil = none; a glow/halo behind the card box
```

Additive pointer field; no break.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** nil `Backdrop` skips the call; a
  byte-identity test pins it.
- **R2 — backdrop drawn over the card.** **Mitigation:** it is drawn *before*
  the chrome, so the card fill sits on top; a z-order test asserts the glow
  precedes the rounded-rect fill.

## 11. Acceptance criteria

1. A card with a `radial_glow` `Backdrop` emits a radial-gradient ellipse *before* the card's rounded-rect fill in the slide XML (behind it in z-order), bleeding beyond the card box.
2. A card with `Backdrop == nil` is byte-identical to the pre-Phase-79 build.
3. A backdrop card re-renders deterministically.
4. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-79.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `Card.Backdrop` field + render wiring.
3. `OK:` backdrop-before-fill (z-order) test.
4. `OK:` nil-backdrop byte-identical + determinism tests.

## 14. Tests

- **Black-box (`scene_test`):** a backdrop glow's radial ellipse precedes the
  card's rounded-rect fill; a nil-backdrop card byte-identical; determinism.
- **Integration / Fuzz:** no.
