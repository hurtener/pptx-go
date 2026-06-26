# Phase 91 — footnotes / sources / disclaimers

**Subsystem:** `scene` (SceneSlide field + footer band + superscript run style)
**RFC sections:** §11.1, §10.1, chrome reservation (D-053)
**Deps:** D-053 (chrome bands); brief 74.
**Status:** Done

## 1. Goal
Attach slide-level footnotes/sources/disclaimers in a reserved bottom band (muted),
never overlapping the body or page-number footer; add a superscript marker style.

## 2. Why now
Wave 14 coverage; regulated/investor decks need citations. R14.12 (MED · both,
engine half; D-059).

## 3-6. RFC/brief/decisions
RFC §11.1 (Footnotes field), §10.1 (empty byte-identical), D-053 chrome footer
reservation. Brief 74. Decisions D-059, D-026, D-126 (new).

## 7. Architecture
`SceneSlide.Footnotes []RichText`. `composeSlide` sets `r.footnoteH =
footnoteBandHeight(Footnotes)` before layout; `bodyRegion` subtracts the band (+
gap); `renderFootnotes` draws muted caption lines in the band above the chrome
footer (cap 3, drop+warn past it). `scene.RunStyle.Superscript` →
`pptx.Superscript` (BaselineRel) in `addRichTextScaled`.

## 8. Files
scene/scene.go (Footnotes), scene/richtext.go (RunStyle.Superscript), scene/render.go
(footnoteH field + bodyRegion reserve + composeSlide + renderFootnotes +
addRichTextMuted + superscript map), scene/render_footnotes_test.go (NEW),
scene/render_adversarial_test.go (footnotes on strip slide), scripts/smoke/phase-91.sh,
docs/research/74 + INDEX, this plan, README, glossary, THEME, decisions D-126,
skills/compose-a-scene.

## 9. Public API
`SceneSlide.Footnotes []RichText`; `scene.RunStyle.Superscript bool`. Additive.

## 10-11. Risks/acceptance
Overlap (mitigated: bodyRegion reserves the band; adversarial footnote slide);
byte-identity (empty self-gates). Accept: 2 sources render muted + a superscript
marker (no warnings, no overlap); empty byte-identical; >cap warns; deterministic.

## 12-14. Coverage/smoke/tests
scene 80%. scripts/smoke/phase-91.sh. Black-box: footnotes render + superscript,
empty byte-identical, cap warns, determinism; adversarial footnote band.
