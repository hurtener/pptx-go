# Brief 11 — slide-chrome

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-20
**Motivating phase:** Phase 24 — slide chrome (section eyebrow + footer)

## 1. Question

Reference decks read "designed" partly because every content slide carries a
**section eyebrow** + hairline rule at the top (`01 — DIRECTION`) and a **footer**
with a brand mark and an `N / total` page number. The engine has no concept of
chrome — recurring per-slide furniture drawn outside the content. How can the
engine render opt-in chrome that never overlaps content, is driven by simple
fields on `Scene`/`SceneSlide`, and leaves a chrome-free deck byte-identical?

## 2. Prior art surveyed

- **`scene/render.go::composeSlide` + `bodyRegion`.** The body region is a fixed
  margin inset (`bodyMargin` all around); every node lays out inside it. This is
  exactly the seam to shrink: enabling chrome reduces the body region's top and
  bottom so the bands sit in the reclaimed margin, guaranteeing no overlap.
- **PowerPoint placeholders (slide number, footer, date).** PowerPoint models
  footers as master/layout placeholders inherited per slide. pptx-go composes
  native shapes per slide instead (no master-placeholder authoring in V1), which
  keeps chrome a pure function of the scene fields and avoids touching the master
  codec.
- **`render_leaves.go` / `render_card.go` chrome patterns.** Eyebrows
  (`TypeCaption` + accent/muted color), hairline rules (a thin `ShapeRect`), and
  right-aligned labels already exist as composer idioms. Chrome reuses them — no
  new builder primitive, no new token.
- **D-026 (engine, not product).** *Whether* a deck wants chrome, *what* the
  brand mark is, and *what* each section is called are the caller's taste. The
  engine provides the mechanism (a chrome toggle + slots) and renders what it is
  handed; it composes the `N / total` string mechanically but invents no labels.
- **Determinism (RFC §10.1).** Per-slide native shapes are deterministic. The
  only global-media touch is an optional brand *image*; that must force
  sequential composition so its part number is stable.

## 3. Findings

- **Chrome belongs outside the body region, and the body must shrink.** Drawing
  the eyebrow in the top margin and the footer in the bottom margin, and
  shrinking `bodyRegion` by those band heights when chrome is on, makes overlap
  structurally impossible — the body simply never extends into the bands.
- **The field split writes itself.** Deck-level data (the brand slot, the page
  total) lives on `Scene`; per-slide data (the section label, the page index)
  lives on `SceneSlide`. A single `Chrome` struct on `Scene` with an `Enabled`
  master switch keeps the zero value inert (chrome off ⇒ byte-identical).
- **Page numbering should auto-derive but stay overridable.** `Total` defaults
  to `len(Scene.Slides)`; a slide's page number defaults to its scene position
  (1-based) and is overridable per slide. This keeps the common case
  zero-config while letting a caller renumber (skip a cover, restart a section).
- **The eyebrow is per-slide and optional; the footer is consistent.** A slide
  with an empty `Section` draws no eyebrow; the footer (brand + page number)
  draws on every chrome-enabled slide. This matches the reference: a consistent
  footer, section labels that change (or are absent on a divider).
- **Brand slot is text-or-asset, mirroring the existing seam.** A brand *text*
  is a native run; a brand *asset* (an `AssetID` resolved through the existing
  `AssetResolver`) is an image. Only the asset path registers global media, so
  only it forces sequential composition.
- **Tokens, not literals (P2).** Chrome colors resolve through existing roles
  (`TextMuted` for eyebrow/footer text, `ColorSurfaceAlt` for the hairline), so
  a theme swap re-skins chrome and **no new token is introduced** — no
  `THEME.md` taxonomy entry is required.

## 4. Recommendations

1. Add a `Chrome` struct on `Scene`: `Enabled bool`, `Brand string`,
   `BrandAsset AssetID`, `Total int`. Add `Section string` + `PageNumber int` to
   `SceneSlide`. All zero values inert.
2. Shrink `bodyRegion` by a fixed eyebrow-band height (top) and footer-band
   height (bottom) when chrome is enabled; draw the bands in the reclaimed
   margin via a new `renderChrome` composer (native shapes only, except the
   optional brand image).
3. Auto-derive `Total` (← `len(Slides)`) and per-slide page number (← scene
   position), both overridable. Compose the `N / total` string in the engine.
4. Force sequential composition for the whole deck when `Chrome.BrandAsset` is
   set (the only global-media touch); brand-text chrome stays parallel.
5. Degrade an unresolved brand asset to a `LayoutWarning` (never fail) — the
   warn-don't-fail asset contract.

## 5. Open questions

- **Per-slide chrome opt-out.** Chrome is deck-wide when enabled; a cover slide
  still gets a footer page number. A per-slide "suppress chrome" flag (so the
  cover is bare) is a plausible refinement; deferred, since the requirement says
  "on each slide" and a caller can leave `Section` empty.
- **Custom page-number format / separator.** The engine hard-codes `N / total`.
  Exposing a format string is product-flavored and deferred; revisit only if a
  real brand needs `N of total` or roman numerals.
- **Footer date / confidentiality slot.** Reference decks sometimes add a center
  footer (date, "Confidential"). Additional slots are a trivial extension of the
  same band; deferred until asked.
- **Master-placeholder chrome.** Authoring chrome as true master placeholders
  (so PowerPoint's "insert slide number" toggles it) is a V2 codec concern;
  per-slide native shapes are the V1 mechanism.
