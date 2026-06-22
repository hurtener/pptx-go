# Brief 35 — content-region-reserves-chrome (R11.4 verify-and-close)

**Subsystem:** scene — Layer 2 renderer (body region / chrome)
**Authored:** 2026-06-22
**Motivating phase:** Phase 52 — content-region-reserves-chrome (R11.4, HIGH · engine)

## 1. Question

R11.4 requires the body content region handed to the node tree to subtract the top
section-eyebrow band and the bottom footer/page-number band when chrome is enabled,
so no body content can occupy the footer/eyebrow rows (recreation slides 6, 7 show
content drawn over the footer). Is this already satisfied, and if so what is the
minimal close?

## 2. Prior art surveyed

- **`scene/render.go bodyRegion()`** — when `r.chrome.Enabled`, already adds
  `chromeEyebrowH + chromeBandGap` to the top inset and `chromeFooterH +
  chromeBandGap` to the bottom inset (the constants `chrome.go` draws the bands
  with — single source of truth). This is exactly R11.4's "subtract the eyebrow
  band from the top and the footer band from the bottom".
- **`scene/render.go layout()`** — the body stack is placed into `r.bodyRegion()`
  (line 322), so the reduced region *is* the region the node tree flows inside.
- **`scene/chrome.go renderChrome`** — draws the eyebrow band at
  `[bodyMargin, bodyMargin + chromeEyebrowH]` (+ a `chromeRuleH` hairline) and the
  footer band at `[cy − bodyMargin − chromeFooterH, cy − bodyMargin]`.
- **R11.3 / D-083 (Phase 51)** — `clampToSafeArea` (safe area `= bodyRegion()`) caps
  every container's box to the reserved region, so even an over-full stack cannot
  push a container below it into the footer.

## 3. Findings

- **R11.4 is already implemented by D-053 (Phase 24).** `bodyRegion()` reserves both
  chrome bands using the chrome composer's own constants, and the body stack is laid
  out inside it. The geometry is disjoint by construction: the body top
  (`bodyMargin + chromeEyebrowH + chromeBandGap`) sits below the eyebrow band *and*
  its rule (because `chromeBandGap = 91440 EMU ≈ 0.10"` exceeds
  `chromeRuleH = 9525 EMU ≈ 0.0075"`), and the body bottom
  (`cy − bodyMargin − chromeFooterH − chromeBandGap`) sits above the footer band top
  (`cy − bodyMargin − chromeFooterH`).
- **The overflow hole that defeated the reservation is closed by R11.3.** The
  recreation's footer overlap happened not because the region wasn't reserved, but
  because an over-full body stack placed nodes *below* the reserved region into the
  footer. R11.3's `clampToSafeArea` (safe area = `bodyRegion()`) now caps containers
  to the reserved region, and `VAlignFit` reflows over-full stacks — so the
  reservation is now actually honored under hostile content.
- **The open gap is the acceptance test, not mechanism.** R11.4's acceptance: "On
  any chromed slide, no body shape/text box intersects the eyebrow band (top) or
  footer band (bottom). A test renders a content-maximal slide with chrome on and
  asserts zero intersection between body boxes and the reserved bands; chrome-off
  slides unchanged." The close is that assertion: (a) the reserved `bodyRegion`
  (chrome on) is disjoint from both bands computed from the chrome constants;
  (b) chrome-off `bodyRegion` is the plain margin box (no reservation →
  byte-identical); (c) with R11.3, a clamped container on a chromed slide stays
  above the footer band.

## 4. Recommendation

Ship Phase 52 as a verify-and-close: a white-box acceptance test asserting the
chrome-on `bodyRegion` is disjoint from the eyebrow and footer bands (and that a
clamped container stays above the footer), plus a chrome-off byte-identical check.
D-084 records R11.4 closed by D-053 (reservation) + D-083 (the clamp that makes the
reservation hold under overflow). No renderer change.

## 5. Open questions

- `bodyRegion` reserves the top eyebrow band whenever chrome is enabled, even on a
  slide with no `Section` (no eyebrow drawn) — a minor, intentional
  over-reservation (pre-existing, D-053). Not changed here.
