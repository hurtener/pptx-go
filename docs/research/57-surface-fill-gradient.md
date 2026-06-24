# Brief 57 — Surface fill gradient (R13.8)

> Informs Phase 74 (Wave 13). Engine req R13.8
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059).

## 1. Motivating phase

Reference cards use gradient fills for depth (ref-05 "way" cards fade
vertically; dark cards have subtle top-to-bottom shifts). Deckard's cards are
flat single-color fills, so they read as solid swatches. Phase 74 adds an
optional 2-stop gradient fill to the `Card` surface, defaulting to the solid
`Fill` (byte-identical).

## 2. Subsystem / files

- `scene/nodes.go` — the `Card` struct.
- `scene/render_card.go` — `cardChrome` + `renderCardChrome` (the surface
  rounded-rect fill at `pptx.SolidFill(pptx.TokenColor(c.fill))`).
- `pptx/fill.go` — `LinearGradient(angle, ...GradientStop)` (already exists).

## 3. Findings

- **One fill site.** The card surface fill is a single
  `WithFill(SolidFill(TokenColor(c.fill)))` in `renderCardChrome`. Swapping it
  for a `LinearGradient` when a gradient is set is a one-line branch.
- **`*GradientFill` keeps byte-identity + comparability.** Add
  `Card.FillGradient *GradientFill` (`{From, To pptx.ColorRole; Angle int}`).
  nil = solid `Fill` (byte-identical). A pointer (not a value) so "unset" is
  unambiguous; `Card` is already non-comparable (it has `Body []SlideNode`), so
  no new constraint. A distinct 2-role `{From,To,Angle}` type is the simplest
  surface API (vs the N-stop `Background.Stops`) — a card depth shift is 2-stop.
- **No auto-tint.** R13.8 mentions a convenience where `To` defaults to a
  slightly-darker role. That is a *taste* opinion (which direction, how much) —
  D-026 puts it in the soul, not the engine. Ship both `From` and `To` explicit;
  document the auto-tint as the soul's job. (No zero-value remap: `ColorRole`
  zero is `ColorCanvas`, so when `FillGradient` is set both roles are explicit.)
- **CardSection / Bento / Container deferred.** `cardChrome` is shared by Card
  and CardSection, but only `Card` gets the field this phase; Bento cell and
  Container fills are separate code paths. Card is the dominant case and
  satisfies the acceptance criterion; the `GradientFill` type is reusable when
  those follow (§4.3 deviation, noted).
- **No OOXML / `restorenamespaces` change.** Emits `<a:gradFill>`/`<a:gs>` via
  the existing fill path (the background gradient already proves it). Verify
  bytes: a gradient card's surface has `<a:gradFill`; a solid card does not.
- **Determinism.** Pure 2-stop mapping; worker-independent.

## 4. Recommendations

- Add `GradientFill{From, To pptx.ColorRole; Angle int}` + `Card.FillGradient
  *GradientFill`; thread it into `cardChrome.fillGradient`.
- `renderCardChrome`: `fill := SolidFill(TokenColor(c.fill)); if
  c.fillGradient != nil { fill = LinearGradient(angle, {0,From},{1,To}) }`.
- Tests: a gradient card's surface emits `<a:gradFill>` with a top-vs-bottom
  role pair; a solid card is byte-identical; determinism guard.
- THEME.md surface-gradient mechanism note, glossary, compose-a-scene skill,
  docs/site catalog. D-108.

## 5. Open questions

- CardSection/Bento/Container gradient fills → a later phase or V2 if needed;
  the type is reusable.
- Soul-controlled auto-tint of `To` → product (Deckard), D-026.
