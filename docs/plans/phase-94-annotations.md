# Phase 94 — image / diagram annotations

**Subsystem:** `scene` (Image field-ext)
**RFC sections:** §14.3 (image), §10.1, §7.1
**Deps:** brief 77.
**Status:** Done

## 1. Goal
Overlay numbered pins (+ leader captions) and highlight boxes on an `Image` at
fractional coordinates.

## 2. Why now
Wave 14 coverage; annotated screenshots/diagrams are uncovered. R14.17 (LOW ·
engine; D-059).

## 3-6. RFC/brief/decisions
RFC §14.3 (Image), §10.1 (nil byte-identical), §7.1 (token colors). Brief 77.
Decisions D-059, D-026, D-130 (new).

## 7. Architecture
`Image.Annotations *ImageAnnotations{Pins []ImagePin{X,Y,Label,Caption,
AccentIndex}; Highlights []ImageHighlight{X,Y,W,H,AccentIndex}}`. After the pic,
`renderImageAnnotations` draws highlight rect outlines, then accent pin discs with
a number + an optional leader line to a clamped caption. Fractions of the image
interior box. Native (reuse hvLine/timelineAccent/cellTextOn). nil = byte-identical.

## 8. Files
nodes.go (Image.Annotations + ImageAnnotations/ImagePin/ImageHighlight), validate.go
(coords in [0,1]), render_image.go (renderImageAnnotations + geom consts),
render_annotations_test.go (NEW), scripts/smoke/phase-94.sh, docs/research/77 +
INDEX, this plan, README, THEME, glossary, decisions D-130,
docs/site/catalog/asset-leaves.md, skills/compose-a-scene.

## 9. Public API
`Image.Annotations *ImageAnnotations` + `ImagePin` + `ImageHighlight`. Additive.

## 10-11. Risks/acceptance
Off-image (clamp + edge-flip captions); byte-identity (nil self-gates).
Accept: 3 pins + leader captions + a highlight render each at its coordinate
(conformant); an out-of-range pin fails Stage-1; nil byte-identical; deterministic.

## 12-14. Coverage/smoke/tests
scene 80%. scripts/smoke/phase-94.sh. Black-box: pins+caption+highlight,
invalid-coord, nil byte-identical, determinism.
