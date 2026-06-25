# Brief 68 — Quote / testimonial enrichment (R14.5)

> Informs Phase 85 (Wave 14). Engine req R14.5
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine; D-059).

## 1. Motivating phase

Sales/investor decks lean on social proof: a large editorial pull-quote with an
oversized quotation glyph, an author avatar, structured name/role/company
attribution, and a customer logo. The `Quote` node is minimal (Text +
Attribution + Align), so every testimonial reads as plain centered text. Phase 85
extends `Quote` additively into a designed testimonial.

## 2. Subsystem / files

- `scene/nodes.go` — the `Quote` node.
- `scene/render_leaves.go` — `renderQuote`.
- `scene/render.go` — `preferredHeight` (Quote) + `nodeUsesAssets` + the dispatch.
- `pptx/media.go` — `(*Image).SetCornerRadius` (rounded avatar, D-114).

## 3. Findings

- **Additive fields, a branched renderer.** `Mark bool`, `AvatarAssetID`,
  `AttributionName/Role/Company`, `LogoAssetID`. `renderQuote` keeps the existing
  plain path verbatim when no enrichment is set (byte-identical) and branches to
  `renderTestimonial` otherwise (a `Quote.enriched()` predicate gates it).
- **Avatar/logo make the Quote asset-bearing.** `nodeUsesAssets(Quote)` returns
  true when `AvatarAssetID`/`LogoAssetID` is set, so the slide composes serially
  and media part numbering stays deterministic. A plain quote stays parallel-safe.
- **Policy stays `HasAsset:false`.** The new asset fields are named
  `AvatarAssetID`/`LogoAssetID`, not `AssetID`, and the Quote still renders as
  native text + embedded pics (not a single pic node), so `TestPolicy_MatchesStructs`
  (which keys on a field literally named `AssetID`) is unaffected — `KindQuote`
  is unchanged. No new node; the catalog stays 29.
- **The avatar reuses `SetCornerRadius(RadiusFull)`** (D-114) for a circular crop
  of a square box.
- **The quotation mark is a font glyph, not a checkbox.** `“` (U+201C) renders in
  every standard face; the D-095 empty-box risk was specific to checkbox
  characters, not the quote mark. It is drawn first (behind) at a low alpha
  (`TokenColorAlpha`) so it never harms legibility.
- **Logo cover-fit.** The logo is height-bounded with a 2.5:1 width reserve; the
  builder stretches it (a follow-up could route it through `WithImageFill`
  cover-crop, but a logo is usually pre-trimmed).
- **`preferredHeight` reserves the strip.** The enriched quote adds the
  attribution strip (+ half the mark) to the slot estimate so the R11.3 clamp
  stays truthful.

## 4. Recommendations

- Node: `Quote{… Mark bool; AvatarAssetID AssetID; AttributionName, Role,
  Company string; LogoAssetID AssetID}` + an `enriched()` predicate.
- `renderTestimonial`: optional oversized mark (behind), the quote text (TypeH3),
  then a bottom strip `[rounded avatar | name(bold)/role·company(muted) | logo]`.
  Assets via `r.resolve` (warn + omit on miss). `nodeUsesAssets` + `preferredHeight`
  updated.
- Tests: a full testimonial renders (2 pics, rounded avatar, role·company,
  conformant), a plain quote is byte-identical (no pic), a missing avatar warns,
  determinism; an adversarial enriched-quote slide (mark + structured attribution,
  no assets). THEME.md (mark color mechanism), glossary, compose-a-scene skill,
  docs/site text-leaves. D-120.

## 5. Open questions

- Logo cover-crop via `WithImageFill` → a follow-up (logos are usually trimmed).
- A horizontal avatar-left layout variant → the vertical strip covers the
  acceptance; a side layout is V1.x.
