# Brief 16 — resolved-colors

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-21
**Motivating phase:** Phase 29 — expose resolved per-slide colors

## 1. Question

A caller (the product built on pptx-go) validates text/surface contrast, but it
cannot see the colors the engine actually *resolved* per slide — especially for
`VariantDark`, where the engine swaps to a **derived dark palette**
(`darkThemeFrom`). The caller's validator checks contrast against the light theme
and false-flags white-on-dark. How can the engine expose, per slide, the resolved
canvas / surface / primary-text colors it rendered with — so the caller computes
true contrast itself — without the engine taking on any contrast logic?

This is `DECKARD-PRODUCT-REQUIREMENTS.md` R7 (LOW), the final Wave 8 unit.

## 2. Prior art surveyed

- **`scene.Stats`.** Already the library's observability surface (per-slide
  timings, shape/asset counts, warnings — no event protocol, D-016). A per-slide
  resolved-colors slice is a natural additive field on it, merged in scene order
  like `Timings`.
- **`composeSlide` theme swap.** For `VariantDark`, `composeSlide` derives a dark
  theme (`darkThemeFrom(orig)`), sets `r.theme = dark`, and defers restoring the
  *presentation* theme — but the per-slide renderer's `r.theme` field stays the
  dark theme for the duration. So **after `composeSlide` returns, `sr.theme` is
  exactly the theme the slide rendered with** (dark for dark slides, the active
  theme otherwise).
- **`Theme.ResolveColor` / `ResolveTextColor`.** Exported, returning `pptx.RGB` —
  the same resolution the codec used when emitting fills. Calling them on the
  per-slide theme yields the *actual* rendered values, dark palette included.
- **D-026 (engine, not product).** Contrast is a product judgment — the caller's.
  The engine must expose the *resolved values* (a mechanism) and compute no
  ratio, threshold, or pass/fail. R7 is explicitly mechanism-only.

## 3. Findings

- **Capture from the per-slide theme, after compose.** `composeOne` already
  builds the `slideResult` after `composeSlide` returns; reading `sr.theme`
  there gives the resolved palette — dark for dark slides — with no new plumbing
  and no second theme derivation.
- **Canvas / surface / primary-text are the right three.** The requirement names
  canvas (the slide's base background), surface (cards/panels), and primary text.
  Resolving those three per slide lets a caller compute background-vs-text
  contrast against the real palette (and apply its own large-text thresholds).
- **Additive and output-invariant.** A new `Stats.Colors []SlideColors` field is
  pure metadata — it is never emitted into the `.pptx`, so the rendered bytes are
  byte-identical; only the returned `Stats` grows. Existing callers ignore the
  field.
- **Deterministic and scene-ordered.** Resolution is pure (theme → RGB);
  merging the per-slide results in scene order (like `Timings`) keeps `Colors`
  worker-count independent.
- **No contrast logic.** The engine returns RGBs only. WCAG ratios, large-text
  thresholds, and pass/fail stay entirely in the caller (D-026).

## 4. Recommendations

1. Add a `SlideColors{SlideID string, Canvas, Surface, PrimaryText pptx.RGB}`
   struct and a `Stats.Colors []SlideColors` field.
2. Capture in `composeOne` from `sr.theme` after `composeSlide`:
   `Canvas = ResolveColor(ColorCanvas)`, `Surface = ResolveColor(ColorSurface)`,
   `PrimaryText = ResolveTextColor(TextPrimary)`; carry it on `slideResult` and
   append in `Render`'s scene-order merge loop.
3. Document that the dark-variant slice carries the derived dark palette, and
   that the engine performs no contrast computation.

## 5. Open questions

- **More roles (accent, secondary text, surface-alt).** Only the three named
  roles are exposed. A caller needing more could get a fuller palette snapshot;
  deferred — the three cover background-vs-text contrast, the stated need.
- **Resolved per-node colors.** This exposes the per-slide *theme* palette, not a
  per-node literal-color override. A node that sets a `LiteralColor` is the
  caller's own value already; deferred.
- **A query API vs a `Stats` field.** The requirement allows either; a `Stats`
  field is the lighter, already-established shape and is chosen here. A
  post-`Render` query API is deferred unless a caller needs colors without the
  rest of `Stats`.
