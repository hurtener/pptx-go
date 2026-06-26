# Brief 74 — Footnotes / sources / disclaimers (R14.12)

> Informs Phase 91 (Wave 14). Engine req R14.12
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · both — engine half; D-059).

## 1. Motivating phase
Investor/financial/research decks carry source lines, numbered footnote markers,
and legal disclaimer bands. There was no footnote/source primitive and no
superscript-marker mechanism, so citations got jammed into body copy or omitted.

## 2. Findings
- **A slide-level field, not a node.** `SceneSlide.Footnotes []RichText` renders
  in a reserved band pinned to the bottom (above the chrome footer) in the muted
  caption role. The body region must shrink to reserve the band so footnotes never
  overlap the body or the page-number footer — done via a per-slide renderer field
  (`footnoteH`) set in `composeSlide` before layout (each slide gets a fresh
  renderer in `composeOne`, so the field is not shared). Empty = byte-identical.
- **Superscript is already in the builder.** `pptx.BaselineShift`/`Superscript` +
  `RunStyle.BaselineRel` exist; the scene only needed a `RunStyle.Superscript bool`
  mapped through `addRichTextScaled`, so a marker run on any figure/stat renders
  raised + reduced.
- **Region cap.** Lines past `footnoteMaxLns` (3) are dropped + warned
  (ellipsized) rather than overflowing.

## 3. Recommendations
- `SceneSlide.Footnotes` + `footnoteBandHeight`/`renderFootnotes`; `bodyRegion`
  reserves the band; `scene.RunStyle.Superscript` → `pptx.Superscript`. Tests:
  sources render muted + a superscript marker emits a baseline shift (no
  warnings); empty byte-identical; >cap warns; determinism; an adversarial
  footnote band. Glossary, THEME (muted = mechanism), skill. D-126.

## 4. Open questions
- Auto-numbered marker↔footnote linking → product (the caller numbers them).
