# Brief 25 — card-body-vertical-distribution

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-22
**Motivating phase:** Phase 42 — card body vertical distribution (R10.4, HIGH · engine)

## 1. Question

`renderCard` always lays a card's body out top-anchored via `stackIn`, so
secondary content floats in the upper card with large dead space below (the
recreation's Vision/Mission lists, the path cards on slide 8, the pricing cards on
slide 9 leave ~40–60% of the card blank). How can a card opt its body into
vertical distribution — center / bottom / justify / fill — within the card body
region, **without** changing the default top-anchored output?

## 2. Prior art surveyed

- `scene/render_card.go` `renderCard`: computes the body region via
  `renderCardChrome` (which returns the `bodyBox` below the wrapped header,
  inside the padding), then lays the vertical body out with
  `r.stackIn(body, v.Body, slideID)` (top-anchored, `SpaceMD` gap). A separate
  `BodyHorizontal` path uses `layout.Columns`.
- `scene/render.go` `alignedStackIn(box, nodes, slideID, align)`: the body-stack
  layout with full vertical distribution — `VAlignCenter` offsets the start Y,
  `VAlignBottom` pins to the box bottom, `VAlignJustify` expands inter-node gaps,
  `VAlignFill` grows flexible nodes (`distributeFill`), and (R10.2) `VAlignFit`
  compresses an over-full stack. Its godoc already states: *with the zero
  Alignment {VAlignTop, HAlignLeft} … produces placements byte-identical to
  `stackIn`.*
- `scene/align.go` — the `VAlign` enum (Top/Center/Bottom/Justify/Fill/Fit) and
  the `Alignment` struct.
- Phase 13 (alignment, D-?) deliberately scoped `alignedStackIn` to the top-level
  body stack and kept containers on `stackIn`. R10.4 reverses that *for the card
  body specifically*, on an opt-in field.
- DECKARD R10.4 spec: add an additive `Card.BodyVAlign` field (zero = Top =
  today); route the body through the same vertical-distribution logic as
  `alignedStackIn` on the card's `bodyBox`; reuse `distributeFill`/`effectiveGap`;
  zero value byte-identical; deterministic integer EMU.

## 3. Findings

- The mechanism the spec asks for **already exists** — `alignedStackIn` is the
  vertical-distribution engine. The card body just needs to call it instead of
  `stackIn`, parameterized by a new `Card.BodyVAlign VAlign`.
- **Byte-identity is provable and already documented.** For `{VAlignTop,
  HAlignLeft}`: `alignedStackIn` sets `startY = box.Y`, `effectiveGap = gap`,
  emits `{X:box.X, Y:y, W:box.W, H:preferredHeight}` per node, and warns when
  `totalH > box.H`. `stackIn` does exactly the same, and its warn condition
  `last-bottom > box.Bottom()` is algebraically identical to `totalH > box.H`
  (last bottom = `box.Y + totalH`, `box.Bottom() = box.Y + box.H`). So replacing
  the card's vertical `stackIn` with `alignedStackIn(..., {Vertical: BodyVAlign})`
  is byte-identical when `BodyVAlign == VAlignTop`.
- **Horizontal stays left.** Passing `Horizontal: HAlignLeft` makes
  `nodeEffectiveHAlign` return left for every body node (no per-node override in a
  card body), so the Chip-narrowing branch is skipped and text leaves keep their
  full-width box — identical to the `stackIn` path.
- **Free composition.** Because `BodyVAlign` is a `VAlign`, a card body can also
  use `VAlignFit` (R10.2 compression) or `VAlignFill` (grow flexible body nodes)
  for free — the same engine. The spec's named set (center/bottom/justify/fill) is
  a subset.
- **Scope: Card only.** `renderCardSection` also stacks via `stackIn`, but the
  R10.4 gap and acceptance are about `Card` (price/list bodies). `CardSection`
  bodies are containers (grids/columns) that already grow under slide-level fill;
  adding a `CardSection.BodyVAlign` is a separable, lower-value follow-up.
- **`BodyHorizontal` is unaffected** — `BodyVAlign` is a vertical-distribution
  control; the horizontal column path keeps its current behavior.

## 4. Recommendations

1. Add an additive `Card.BodyVAlign VAlign` (zero = `VAlignTop` = today).
2. In `renderCard`, replace the vertical `r.stackIn(body, v.Body, slideID)` with
   `r.alignedStackIn(body, v.Body, slideID, Alignment{Vertical: v.BodyVAlign})`.
   Leave the `BodyHorizontal` branch unchanged.
3. Update `alignedStackIn`'s godoc: it is no longer the *sole* alignment caller —
   a card body opts in via `BodyVAlign`.
4. Tests: `BodyVAlign=Bottom` pins the last body node's bottom to the card body
   bottom; `BodyVAlign=Top` (zero) is byte-identical to the current card output;
   a determinism guard; smoke `phase-42.sh`.

## 5. Open questions

- **`CardSection.BodyVAlign`** — deferred (container bodies; separable).
- **Per-body-node `Align` overrides inside a card** — not needed by R10.4;
  `alignedStackIn` already supports them if a future req wants per-node control.
