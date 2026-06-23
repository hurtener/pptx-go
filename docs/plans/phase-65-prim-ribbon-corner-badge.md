# Phase 65 — prim ribbon corner badge

**Subsystem:** scene — Layer 2 renderer (Card field extension)
**RFC sections:** §11.2 (Card), §12 (policy)
**Deps:** Phase 39/48 (card header geometry, D-070/D-079), Phase 64 (auto-contrast color
pattern), Phase 25 (rich card visuals `*ColorRole`, D-054); brief 48.
**Status:** Done

---

## 1. Goal

Add a `Card.Ribbon` field — a pinned emphasis badge (a "MOST POPULAR" top bar, a corner
tab, or a star) that sits outside the header text flow — so a deck can single one card
out of a row.

## 2. Why now

R12.3 is a HIGH Wave-12 primitive. `Card.HeaderPill` is an in-row pill that overlapped
the tier name in the recreation; there is no pinned badge distinct from it. See
`docs/plans/README.md` Wave 12 and D-059 (engine-tagged).

## 3. RFC sections implemented

- `RFC §11.2` — extends `Card` with the additive `Ribbon` field.
- `RFC §12` — native policy (filled tab/bar + text, or a star custGeom); no new asset.

## 4. Brief findings incorporated

- `docs/research/48-prim-ribbon-corner-badge.md` — *"RibbonTopBar reserves a band; the
  body shifts down"* → `ribbonReserveOf` threaded through `cardHeaderBottom` /
  `renderCardChrome` / `cardHeaderExtraHeight`.
- `48-...md` — *"the diagonal corner ribbon is deferred; a corner tab carries the
  label"* → CornerTL/TR render content-fit horizontal tabs; CornerStar a star glyph.
- `48-...md` — *"colors are tokens (`*ColorRole` + auto-contrast)"* → `ribbonColorRole`
  / `ribbonTextColor`.

## 5. Findings I'm departing from

- The spec's *"diagonal corner ribbon = a rotated rect clipped to the corner"* with the
  label on it. **Departing because** the builder has no rotated-text primitive, so a
  text-bearing diagonal band is not expressible in V1. CornerTL/TR render a horizontal
  content-fit corner tab instead (the label is the point); the rotated-band visual waits
  on a builder text-rotation enhancement. Documented in D-098.

## 6. Decisions referenced

- `D-054` — rich card visuals — the `*ColorRole` "nil = default" pattern for `Color`.
- `D-070` / `D-079` — card header geometry — the shared `cardHeaderBottom` /
  `cardHeaderExtraHeight` the top-bar reserve threads through.
- `D-026` — engine not product — the ribbon is a mechanism; the soul decides when to use
  it. New: `D-098` — prim-ribbon-corner-badge (filed in this PR).

## 7. Architecture

```text
Card.Ribbon *Ribbon{Text, Position, Color *ColorRole, TextColor}
  ribbonReserveOf(c) = ribbonTopBarH (TopBar) | 0 (corner)
  cardHeaderBottom / renderCardChrome:  header top = box.Y + ribbonReserveOf(c) + pad
  cardHeaderExtraHeight += ribbonReserveOf(c)            // slot estimate
  renderCardRibbon (drawn last, on top):
    TopBar  -> full-width RadiusSM tab + centered text
    CornerTL/TR -> content-fit horizontal corner tab + text
    CornerStar  -> curated star custGeom in the top-right corner
```

## 8. Files added or changed

```text
scene/nodes.go                # CHANGED — RibbonPos, Ribbon; Card.Ribbon *Ribbon
scene/validate.go             # CHANGED — Card ribbon position range
scene/render.go               # CHANGED — Card preferredHeight passes ribbon
scene/render_card.go          # CHANGED — cardChrome.ribbon; ribbonReserveOf + color helpers; cardHeaderBottom / renderCardChrome shifts; renderCardRibbon; cardHeaderExtraHeight
scene/render_ribbon_test.go   # NEW — white-box: top-bar shifts body, reserve, color helpers, shape counts
scene/render_ribbon_render_test.go # NEW — black-box: top bar + text, corner star glyph, nil byte-identical, position validation, determinism
scene/render_adversarial_test.go   # CHANGED — a top-bar-ribboned card in the grid
scripts/smoke/phase-65.sh     # NEW — phase smoke
docs/research/48-...md + INDEX.md ; docs/plans/phase-65-...md + README.md
docs/glossary.md ; docs/design/THEME.md ; docs/site/catalog/containers.md ; skills/compose-a-scene/SKILL.md ; docs/decisions.md (D-098)
```

## 9. Public API surface

```go
// scene
type RibbonPos int
const ( RibbonTopBar RibbonPos = iota; RibbonCornerTL; RibbonCornerTR; RibbonCornerStar )
type Ribbon struct { Text string; Position RibbonPos; Color *ColorRole; TextColor TextColorRole }
// Card gains: Ribbon *Ribbon  (nil = none)
```

## 10. Risks

- **R1 — byte-identity when nil.** **Mitigation:** `ribbonReserveOf(nil-ribbon) == 0`;
  the unchanged existing card golden/round-trip tests pass, plus a dedicated nil-vs-absent
  byte test.
- **R2 — top-bar overlaps the header.** **Mitigation:** the reserve shifts the header text
  and body down; a white-box test asserts `cardHeaderBottom(withTopBar) >
  cardHeaderBottom(without)` by exactly the band height.
- **R3 — invisible default color.** **Mitigation:** nil `Color` → `ColorAccent`; a test
  asserts the resolved role.

## 11. Acceptance criteria

1. A `RibbonTopBar` never overlaps the eyebrow/title — it pushes the body down by the
   band height; a corner ribbon stays in the corner.
2. The highlighted card is visually distinguishable (an extra filled tab/bar or star).
3. A card with no `Ribbon` renders byte-identically to before; repeated renders are
   byte-identical.
4. An out-of-range `Ribbon.Position` fails Stage-1 validation.
5. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-65.sh` greps the new surface (Ribbon struct, `Card.Ribbon`,
`ribbonReserveOf`, `renderCardRibbon`) and runs the white/black-box tests; SKIPs
gracefully before the code exists. OK ≥ the acceptance-criteria count, FAIL = 0.
