# Brief 20 — emphasis-as-italic-display

**Subsystem:** pptx — Layer 1 builder (theme/typography + font embedding)
**Authored:** 2026-06-22
**Motivating phase:** Phase 37 — italic-aware font fallback

## 1. Question

Editorial decks express emphasis with an *italic of the display serif* — "agentic
company." set in serif italic teal — rather than a heavy sans bold. For the
engine the question is: when an italic run renders at a display/heading role, does
it resolve to the **italic cut of that role's face**, and what happens when no
italic cut is available? `DECKARD-PRODUCT-REQUIREMENTS.md` R9.7
(`emphasis-as-italic-display`, MED · both). Per D-059 the engine half is the font
resolution guarantee; the soul's emphasis-treatment choice (italic | bold |
accent-color) is Deckard's product half.

## 2. Prior art surveyed

- **D-063 (display-face role).** `TypeDisplay`'s family already resolves to
  `Theme.DisplayFont` (the serif). An italic run at `TypeDisplay` therefore
  already emits `a:latin = <serif>` + `i="1"` — it carries (serif family,
  italic=true).
- **D-065 (font-embedding pass).** The save-time pass already collects
  `(family, bold, italic)` from each run's `rPr` and embeds the italic bucket via
  `FontSource.Resolve(family, "italic", …)`. So an italic display run's italic cut
  is already collected and embedded.
- **D-066 (font fallback chain).** `resolveFontFallbacks` rewrites a run's
  `a:latin` to a fallback when the source can't resolve the primary — but it
  probes and substitutes at the **regular cut / family** level only: it does not
  distinguish "the italic cut is missing" from "the family is missing." So an
  italic run of a family that ships *regular but not italic* is left on the
  primary (Phase 36 says "primary wins" because the regular cut resolves), and the
  consumer faux-italicizes the upright — exactly what R9.7 wants to avoid.

## 3. Findings

- **The display-italic guarantee is already delivered.** D-063 + D-065 mean an
  italic run at a display/heading role renders in (and embeds) that face's italic
  cut. R9.7's first acceptance ("renders in the serif italic cut … in the
  embedded-font list") is satisfied by the existing surface — a *verification*
  test, not new code.
- **The incremental engine work is italic-aware fallback.** Make the fallback
  resolution per **(family, italic)** instead of per family: probe the italic cut
  (`Resolve(family, "italic", 400)`) for italic runs and the regular cut for
  upright runs, and substitute each independently. An italic run whose family's
  italic cut is unavailable then falls back to the first chain family whose italic
  cut resolves — a controlled near-match instead of a faux-italic sans — while the
  upright runs of the same family keep the primary when its regular cut resolves.
- **Generalize the codec rewrite to a resolver.** Replace
  `SlidePart.RewriteFontFaces(map[string]string)` with a resolver callback
  `RewriteFontFaces(func(typeface string, bold, italic bool) string)` so the
  rewrite can key on the run's italic flag (and, later, weight — R9.8). It returns
  the chosen face ("" / unchanged = no rewrite).
- **Additive and deterministic.** When the source resolves both cuts of the
  primary, both win → no substitution → identical to Phase 36. When no `Fallback`
  is declared or no `FontSource` is registered → byte-identical to the true
  baseline. The only new behavior is the "regular present, italic absent" case,
  which is R9.7's target, not a regression. Resolution stays a pure function of
  (theme, source), iterated in fixed role order.
- **The emphasis-treatment switch is Deckard's.** Deciding that emphasis *means*
  display-italic (vs bold vs accent) is taste (D-026); the engine guarantees the
  resolution once the caller places an italic run at a display role.

## 4. Recommendations

1. `internal/ooxml/slide`: change `RewriteFontFaces` to take a resolver
   `func(typeface string, bold, italic bool) string`.
2. `pptx/fonts.go`: make `resolveFontFallbacks` italic-aware — build a
   `(family, italic) → resolved` map (probe the italic cut for italic, the
   regular cut for upright); pass a closure to `RewriteFontFaces`.
3. Tests: an italic run of a regular-only family falls back to a fallback with an
   italic cut, while its upright runs keep the primary; an italic display run
   embeds the display italic cut (the already-satisfied guarantee); byte-identical
   when unused; deterministic.
4. Docs: extend the fallback-chain glossary/THEME entries to note italic-cut
   awareness; record what R9.7's engine half already satisfied (D-063/D-065).

## 5. Open questions

- **Bold-cut fallback.** This phase makes fallback italic-aware, not bold-aware
  (a bold run uses the regular-cut resolution). A family shipping regular+italic
  but no bold is an edge case; bold-specific fallback can layer on the same
  resolver later. Documented.
- **Weight-aware embedding (R9.8).** The resolver callback already carries `bold`
  and is shaped to extend to numeric weight when weight is tracked per run.
