# Brief 47 — prim-callout-banner

**Subsystem:** scene — Layer 2 renderer (new IR node with trailing children)
**Authored:** 2026-06-23
**Motivating phase:** Phase 64 — prim-callout-banner (R12.6, HIGH · engine)

## 1. Question

A full-width filled "big takeaway / promo / CTA" band — a lime closing banner, a dark
"$0 · Start free →" promo strip — is a staple top/bottom-of-slide element. The IR's
`Callout` is only a small left-bar note; the recreation rendered the banner as plain
overlapping black text with no fill. What primitive gives a wide, filled banner with a
leading icon, a bold lead + body, and an optional right-aligned embedded action?

## 2. Prior art surveyed

- **`scene/render_leaves.go renderCallout`** — the side-bar note (a 4pt accent bar +
  inset text); the banner is the wide, full-fill sibling, distinct enough to be its own
  node rather than a `Callout.Style`.
- **`scene/render_leaves.go renderSectionDivider`** — the precedent for a token-filled
  full-bleed rect + a `TextInverse` label; the banner reuses the fill + auto-contrast
  text idea on a `RadiusLG` strip.
- **`scene/render.go stackIn` + `renderNode`** — the card-body pattern (lay children in
  a sub-region, then `renderNode` each) the banner reuses for its `Trailing` children in
  a right region.
- **`scene/contrast.go onCardSurface`** — variant-aware auto-contrast for the lead/body
  on the (typically dark/accent) fill.
- **Container walks** — `validateChildren`, `walkIconRefs`, `walkImages`,
  `nodeUsesAssets`, and the integration `collectKinds` all recurse named containers; the
  banner's `Trailing []SlideNode` joins them.
- **D-059:** R12.6 is `engine`. The node + render is the whole requirement here.

## 3. Findings

- **Banner is a node with trailing children** (`Trailing []SlideNode`, a `Stat` and/or
  `Button`), so it recurses like a container in every walk. It carries no `AssetID`
  itself; `nodeUsesAssets` defers to `Trailing` (a `Button`/`Stat` is media-free, so a
  typical banner stays parallel-safe, but the recursion is correct if a caller embeds an
  asset node).
- **Fill defaults to accent.** `Fill ColorRole`'s zero value is `ColorCanvas` (a real
  color), and a banner is always a filled strip, so the renderer treats the zero value
  as `ColorAccent` (the spec default). A literal canvas banner is not expressible — it
  would be invisible anyway (use a `BackgroundColor`). Documented as a deviation.
- **Text auto-contrasts by default, explicit override honored.** The lead/body resolve
  to `onCardSurface(fill)` when `TextColor` is the zero value (`TextPrimary`) — inverse
  on a dark fill, the default on light — so a banner is legible out of the box; any
  explicit non-default `TextColor` is honored verbatim.
- **Trailing is right-aligned in its own region.** When `Trailing` is non-empty the inner
  band splits into a left text region and a right region; the children stack in the right
  region via `stackIn` + `renderNode` (the card-body mechanism). No `Trailing` → the text
  spans the full inner width.
- **`RadiusLG` rounded rect, pinned padding/icon metrics, token fill.** The strip is one
  `ShapeRoundRect` + `WithRadius(RadiusLG)`; the icon (curated custGeom) and the gaps are
  pinned EMU; the fill and text colors are tokens (P2). Additive: no `Banner` ⇒
  byte-identical.

## 4. Recommendation

Add `KindBanner` + a `Banner` node `{Lead RichText; Body RichText; Icon string; Fill
ColorRole; TextColor TextColorRole; Trailing []SlideNode}` and a `scene/render_banner.go`
composer: a full-width `RadiusLG` filled strip (fill defaults to accent), a left region
with the leading icon + bold lead + body (auto-contrast text), and an optional right
region stacking the `Trailing` children. Full new-node wiring including the container
recursions (`validateChildren`/`walkIconRefs`/`walkImages`/`nodeUsesAssets` over
`Trailing`, integration `collectKinds`), `preferredHeight` (max of the text stack and the
trailing stack + padding), `isFlexible` false. Extend the R11.12 adversarial fixture with
a hostile banner (long lead/body, dark variant, an embedded `Button`). Additive ⇒
byte-identical when unused.
