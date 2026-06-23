# Brief 48 — prim-ribbon-corner-badge

**Subsystem:** scene — Layer 2 renderer (Card field extension)
**Authored:** 2026-06-23
**Motivating phase:** Phase 65 — prim-ribbon-corner-badge (R12.3, HIGH · engine)

## 1. Question

To single one card out of a row — "MOST POPULAR" across a tier's top, a star on a
highlighted feature — a deck needs a pinned emphasis badge that sits OUTSIDE the header
text flow. The only existing tool is `Card.HeaderPill`, an in-row pill that jams into the
title row (ref-09's "POPULAR" overlapped the tier name). What field gives a ribbon /
corner badge distinct from the inline pill?

## 2. Prior art surveyed

- **`scene/render_card.go` chrome** — `cardHeaderBottom` and `renderCardChrome` both
  compute the header top from `box.Y + pad` and share `cardHeaderRowHeights`; the
  D-054 `headerFill` band height is `cardHeaderBottom − box.Y`. A top-bar ribbon must add
  a reserve to that top so the band, header text, and body Y all shift down together.
- **`cardHeaderExtraHeight`** — the R10.10 slot estimate (extra wrapped lines beyond the
  fixed `cardChromeEst`); the top-bar reserve is another additive term there.
- **`scene/render_banner.go` (Phase 64)** — the `*ColorRole` fill + auto-contrast text
  pattern (`onCardSurface`), reused for the ribbon's color/text.
- **Builder limits** — `AddShape` supports `WithRotation`, but text frames have no
  rotation knob (`pptx/text.go`); a true diagonal ribbon with rotated text is not
  expressible in V1.
- **D-054** — the `*ColorRole` "nil = default" pattern for `HeaderFill`/`StatusDot`.

## 3. Findings

- **`Card.Ribbon *Ribbon`, a field extension — not a new node.** No catalog/kind change,
  no `walkIconRefs` (the star uses the curated `star` glyph, no caller icon). `nil` ⇒
  byte-identical (confirmed by the unchanged existing card golden/round-trip tests).
- **RibbonTopBar reserves a band; the body shifts down.** `ribbonReserveOf(c)` returns the
  band height for a top-bar (0 for corner positions). Both `cardHeaderBottom` and
  `renderCardChrome` start the header at `box.Y + ribbonReserveOf(c) + pad`, and
  `cardHeaderExtraHeight` adds it to the slot estimate — so the reserved band, the header
  text, the D-054 band, the body Y, and the `preferredHeight` slot all agree.
- **The diagonal corner ribbon is deferred; a corner tab carries the label.** Because the
  builder has no rotated-text primitive, `RibbonCornerTL`/`TR` render a content-fit
  horizontal text tab pinned in the top corner (an overlay, no body shift) rather than a
  rotated diagonal band. `RibbonCornerStar` renders the curated `star` custGeom in the
  top-right corner. The top-bar (the primary ref-09 case) and the star (ref-06) are
  exact; the diagonal-band-with-text variant waits on a builder text-rotation enhancement
  (documented in D-098). All positions pass the acceptance (distinct, in-corner / top,
  no header-text overlap because the top-bar shifts the body and corner badges sit in the
  empty corner).
- **Colors are tokens.** `Color *ColorRole` (nil = `ColorAccent`); `TextColor` defaults to
  auto-contrast against the fill, with explicit overrides honored. The band/tab/star
  metrics are pinned EMU.

## 4. Recommendation

Add `Card.Ribbon *Ribbon{Text; Position RibbonPos(TopBar/CornerTL/CornerTR/CornerStar);
Color *ColorRole; TextColor TextColorRole}`, drawn last in `renderCardChrome` (on top),
with `ribbonReserveOf` threaded through `cardHeaderBottom` / `renderCardChrome` /
`cardHeaderExtraHeight` so a top-bar shifts the body down deterministically. Corner TL/TR
render content-fit horizontal tabs (the rotated-diagonal variant deferred per the builder
text-rotation limit); CornerStar renders the curated star glyph. Validate the position
range. Additive ⇒ byte-identical when nil. Extend the R11.12 adversarial fixture with a
top-bar-ribboned card in a grid.
