# Phase 64 — prim callout banner

**Subsystem:** scene — Layer 2 renderer (new IR node with trailing children)
**RFC sections:** §11.1 (catalog), §12 (policy)
**Deps:** Phase 61 (Button, the typical Trailing child), Phase 50 (auto-contrast, D-082),
the card-body `stackIn` + `renderNode` pattern; brief 47.
**Status:** Done

---

## 1. Goal

Add a `Banner` scene node — a full-width filled strip with a leading icon, a bold lead +
body, and an optional right-aligned embedded action — for the "big takeaway / promo /
CTA" band a deck opens or closes on.

## 2. Why now

R12.6 is a HIGH Wave-12 primitive. The recreation rendered the closing/promo banner as
plain overlapping text with no fill; `Callout` is only a small side-bar note. See
`docs/plans/README.md` Wave 12 and D-059 (engine-tagged).

## 3. RFC sections implemented

- `RFC §11.1` — extends the catalog with `Banner` (a node carrying `Trailing` children).
- `RFC §12` — native policy (filled rect + text + custGeom icon; children per their own
  policy), no `AssetID` on the banner itself.

## 4. Brief findings incorporated

- `docs/research/47-prim-callout-banner.md` — *"Banner is a node with trailing children;
  recurse like a container"* → the walks + integration `collectKinds` recurse `Trailing`.
- `47-...md` — *"Fill defaults to accent; text auto-contrasts by default"* → the zero
  `Fill` (`ColorCanvas`) maps to `ColorAccent`; default `TextColor` uses `onCardSurface`.
- `47-...md` — *"Trailing right-aligned in its own region via the card-body mechanism"* →
  `stackIn` + `renderNode` in a right sub-region.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-094` — Button — the typical `Trailing` child + new-node pattern.
- `D-082` — auto-contrast — `onCardSurface` for the lead/body on the fill.
- `D-026` — engine not product — the banner is a filled-strip mechanism; no opinion on
  content. New: `D-097` — prim-callout-banner (filed in this PR).

## 7. Architecture

```text
Banner{Lead, Body, Icon, Fill, TextColor, Trailing[]}
  fill = Fill==ColorCanvas ? ColorAccent : Fill ;  text = TextColor==TextPrimary ? onCardSurface(fill) : TextColor
  full-width RadiusLG rounded-rect fill
  left region:  [icon] + lead(bold) + body  (auto-contrast)
  right region (if Trailing): stackIn(Trailing) + renderNode each
```

## 8. Files added or changed

```text
scene/nodes.go               # CHANGED — KindBanner + String; Banner struct
scene/policy.go              # CHANGED — KindBanner policy (native, no asset)
scene/validate.go            # CHANGED — Banner case (recurse Trailing)
scene/render.go              # CHANGED — dispatch + preferredHeight + isFlexible(false) + nodeUsesAssets(Trailing)
scene/render_card.go         # CHANGED — walkIconRefs case Banner (Icon + Trailing)
scene/render_image.go        # CHANGED — walkImages case Banner (Trailing)
scene/render_banner.go       # NEW — banner composer
scene/render_banner_test.go  # NEW — white-box: fill default, text contrast, trailing region, shape count
scene/render_banner_render_test.go # NEW — black-box: filled strip + lead/body, embedded button, unknown icon fails, determinism
scene/scene_test.go          # CHANGED — allNodes + catalog 25 -> 26
scene/render_adversarial_test.go   # CHANGED — hostile banner
test/integration/roundtrip_test.go # CHANGED — everyNodeScene + collectKinds(Banner) + kind loop -> KindBanner + slide counts
scripts/smoke/phase-64.sh    # NEW — phase smoke
docs/research/47-...md + INDEX.md   # brief 47
docs/plans/phase-64-...md + README.md
docs/glossary.md ; docs/design/THEME.md ; docs/site/catalog/text-leaves.md ; skills/compose-a-scene/SKILL.md ; docs/decisions.md (D-097)
```

## 9. Public API surface

```go
// scene
type Banner struct {
    Lead      RichText
    Body      RichText
    Icon      string        // leading custGeom icon; "" = none
    Fill      ColorRole     // strip fill; zero (ColorCanvas) = ColorAccent
    TextColor TextColorRole // lead/body color; zero (TextPrimary) = auto-contrast on Fill
    Trailing  []SlideNode   // right-aligned children (Stat/Button); nil = none
}
func (Banner) NodeKind() NodeKind // KindBanner
const KindBanner NodeKind = ... // appended after KindChipRow
```

## 10. Risks

- **R1 — Trailing children not walked.** A child icon/asset must be validated.
  **Mitigation:** `Banner` joins `validateChildren`/`walkIconRefs`/`walkImages`/
  `nodeUsesAssets`/`collectKinds`; a test embeds a `Button` with an icon and asserts an
  unknown one fails `Render`.
- **R2 — invisible canvas banner.** **Mitigation:** zero `Fill` maps to `ColorAccent`; a
  test asserts a `Banner{}` emits an accent fill.
- **R3 — illegible text on a dark fill.** **Mitigation:** default `TextColor`
  auto-contrasts via `onCardSurface`; the adversarial fixture covers the dark variant.
- **R4 — catalog/kind drift.** **Mitigation:** count 25 → 26 and the integration loop →
  `KindBanner`.

## 11. Acceptance criteria

1. A `Banner` fills its band width with a single colored rect; the lead text is legible
   against the fill (auto-contrast or the caller's `TextColor`).
2. An embedded `Button`/`Stat` in `Trailing` sits in the right region without overlapping
   the lead/body text.
3. A `Banner{}` with no `Fill` renders an accent strip (not an invisible canvas one).
4. An unknown `Icon` or `Trailing` child icon fails icon validation at `Render`.
5. Identical input is byte-identical across worker counts.
6. `scene` catalog count is 26; the integration kind-range loop covers `KindBanner`.
7. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-64.sh` greps the new surface (KindBanner, the struct, the composer,
policy/validate/walk entries, catalog 26) and runs the white/black-box tests; SKIPs
gracefully before the code exists. OK ≥ the acceptance-criteria count, FAIL = 0.
