# Brief 62 — Focal glow behind a card (R13.10)

> Informs Phase 79 (Wave 13). Engine req R13.10
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059). Composes the
> decoration node (Phase 73 role glows / D-107).

## 1. Motivating phase

The reference puts a soft glow precisely behind a focal element (the center
"operating layer" card sits in a faint halo). Deckard decorations anchor only to
the slide region, not to a computed node box, so you cannot place a glow behind
the middle card across any layout. Phase 79 adds a card-relative backdrop.

## 2. Subsystem / files

- `scene/nodes.go` — the `Card` struct.
- `scene/render_card.go` — `renderCard` (the card's computed box is in hand).
- `scene/render_decoration.go` — `renderDecoration(ps, region, v, slideID)`
  already places a decoration within an arbitrary region box.

## 3. Findings

- **`Card.Backdrop *Decoration` is the simplest additive form** the req names.
  `renderCard` already receives the card's computed, safe-area-clamped box, so
  drawing the backdrop **before** the card chrome with that box as the region
  puts the glow behind the card's fill, on the card's actual box, regardless of
  column layout.
- **`renderDecoration` is reusable as-is.** It takes a `region` and resolves the
  decoration's box from `Anchor`/`Offset`/`Size`/`Bleed` within it. A
  center-anchored, larger-than-card, bleeding `radial_glow` produces a halo that
  spills beyond the card and sits behind it (z-order: backdrop first, chrome on
  top). No renderer change beyond the one call.
- **nil = byte-identical.** A `*Decoration` (nil = none) — `Card` is already
  non-comparable (`Body []SlideNode`), so the pointer adds no constraint.
- **Reuses the role-colored glow + alpha tokens (D-107).** The caller sets the
  glow `Color` and a low `Opacity`; the soul keeps it subtle (R13.13).
- **Warnings.** A glow larger than the card box is "off" the card region, so the
  caller sets `Bleed: true` (the existing off-region warning is suppressed,
  exactly as for slide-region decorations). The R13.7 pitch-cap warning only
  fires for pattern presets, so a glow backdrop never trips it.
- **No OOXML / `restorenamespaces` change.** It is a decoration render.

## 4. Recommendations

- Add `Card.Backdrop *Decoration`; in `renderCard`, if non-nil, call
  `r.renderDecoration(ps, box, *v.Backdrop, slideID)` **before**
  `renderCardChrome`.
- Tests: a card with a `radial_glow` backdrop emits a radial-gradient ellipse
  *before* the card's rounded-rect fill in the slide XML (z-order); a card
  without a backdrop is byte-identical; determinism.
- THEME.md note, glossary, compose-a-scene skill, docs/site. D-113.

## 5. Open questions

- A general node-relative anchor (option (b) in the req) is broader; the
  `Card.Backdrop` form covers the motivating case (a focal card halo). Container/
  Bento-cell backdrops can follow the same pattern if needed.
