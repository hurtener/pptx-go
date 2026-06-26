# Brief 77 — Image / diagram annotations (R14.17)

> Informs Phase 94 (Wave 14). Engine req R14.17
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, LOW · engine; D-059).

## 1. Motivating phase
Product/demo decks annotate a screenshot with numbered callout pins, leader lines
to a caption, and highlight boxes. The `Image` node places an asset but can't pin
a marker to a coordinate.

## 2. Findings
- **An additive field on Image.** `Image.Annotations *ImageAnnotations{Pins
  []ImagePin{X,Y,Label,Caption,AccentIndex}; Highlights []ImageHighlight{X,Y,W,H,
  AccentIndex}}`. Coordinates are fractions of the image interior box (node-
  relative), so they track the pic across any placement. Pins = accent discs with
  a number + an optional leader line to a clamped caption; highlights = outlined
  rects. Native shapes (reuse `hvLine`, `timelineAccent`, `cellTextOn`,
  `clampUnit01`). Drawn after the pic. nil = byte-identical. No new node (a field).
- Validate rejects out-of-`[0,1]` pin/highlight coords.

## 3. Recommendations
- Field + `renderImageAnnotations` + validation. Tests: pins + leader caption +
  highlight render (conformant), invalid coord fails, nil byte-identical,
  determinism. Glossary, THEME, skill, docs/site. D-130.

## 4. Open questions
- Anti-collision of dense pins / curved leaders → V1.x; the clamp + edge-flip
  caption satisfies the acceptance.
