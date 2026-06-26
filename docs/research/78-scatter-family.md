# Brief 78 — Scatter ornament family (R14.20)

> Informs Phase 95 (Wave 14). Engine half of req R14.20
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · both; D-059). The pricing-card-recipe
> half is product (Deckard).

## 1. Motivating phase
R14.20 asks to restate the sample-specific `starfield` ornament as a general
*scatter/particle* FAMILY (starfield, dust, confetti, bokeh, grid-of-dots) sharing
one deterministic placement engine with shape/density params.

## 2. Findings
- **One engine, parameterized by mark shape.** The starfield's lattice +
  hash-of-index placement + per-mark size/alpha variance is extracted into a shared
  `scatter(sl, box, alpha, role, pitch, shape)` (in `assets/ornaments`); `Starfield`
  now calls `scatter(scatterDot)` — **byte-identical** to its prior output. New
  recipes `ScatterDot`/`ScatterStar`/`ScatterPlus`/`ScatterRing` pass a different
  `scatterShape` (ellipse / `star5` prst / `mathPlus` prst / ring outline). The
  `ornaments.Recipe` signature is unchanged (the shape is fixed per named recipe),
  so the closed-name registry pattern holds — each family member is a curated name.
- **Registered + capped.** `scatter_dot/_star/_plus/_ring` join `Curated()` (now
  11 ornaments) and `isPatternPreset` (so the `Decoration.Pitch` cap warning
  applies). Density is the existing `Decoration.Pitch`; color the existing role.

## 3. Recommendations
- Shared `scatter` engine + 4 family recipes + registration. Tests: family shapes
  (`star5`/`mathPlus`), `starfield == scatter_dot` byte-identity, determinism;
  curated count 11. Glossary, decision, skill (Decoration presets). D-131.

## 4. Open questions
- A `Decoration.Scatter` param struct (shape/density/sizeVar/alphaVar) → V1.x; the
  named-recipe family + Pitch satisfy the "family parameterized by intent" goal
  without a Recipe-signature break. The pricing-card recipe family is product.
