# Brief 13 — column-join

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-20
**Motivating phase:** Phase 26 — TwoColumn column join (center badge + connector)

## 1. Question

Reference decks compare two cards with a centered **"VS" badge** sitting on the
seam between them, and sometimes a **connector arrow** linking one column to the
next. `TwoColumn` has no concept of an element between its columns. How can the
engine add a centered inter-column element — a text badge or a connector arrow —
additively, deterministically, and byte-identically when unset?

This is sub-units (a) + (b) of `DECKARD-PRODUCT-REQUIREMENTS.md` R5
("composition primitives"). Sub-unit (c) — the row-labeled bento grid — is a
distinct new IR node and lands separately (Phase 27), as R5 explicitly permits
("these can land as separate sub-units").

## 2. Prior art surveyed

- **`render_container.go::renderTwoColumn`.** Splits the box into two columns via
  `layout.Columns` (which leaves a `SpaceMD` gap) and stacks each side's children.
  The gap between the two column boxes is exactly where a center element belongs;
  its midpoint is `(left.X+left.W + right.X) / 2`.
- **Flow connector glyphs (D-044).** `Flow` already renders connector glyphs
  (`ConnectorArrow`, …) between steps by composing preset shapes — precedent that
  an inter-element connector is native shapes, not an anchored `AddConnector`.
- **Chip / pill / badge idioms.** A small filled shape + a centered caption run
  (`AnchorMiddle`, `AlignCenter`, `TextInverse`) is the established badge idiom
  (`renderChip`, the card header pill). The VS badge reuses it.
- **`ColumnRatio` zero value.** `Ratio11` is the `TwoColumn` zero — a real split —
  so the split field already follows the "zero = sensible default" pattern. A new
  *optional* element needs its own zero-is-absent value.
- **D-026 (engine, not product).** The caller supplies the badge label and asks
  for a badge or an arrow; the engine draws it on the seam. It invents no label
  and makes no decision that two columns "should" be compared.

## 3. Findings

- **One enum with a `None` zero cleanly covers both (a) and (b).** A
  `ColumnJoin` enum — `JoinNone` (zero), `JoinBadge`, `JoinArrow` — placed on
  `TwoColumn` lets the same field express "nothing" (default, byte-identical), a
  text badge (a), or a connector arrow (b). No pointer, no companion bool: the
  zero value is naturally "absent". A `JoinLabel string` carries the badge text.
- **The element sits on the seam, overlapping both columns.** Centering a
  ~0.6" badge on the column boundary (not inside the gap, which is only `SpaceMD`
  wide) reproduces the reference look — the badge straddles both cards. Drawing
  it *after* the column content puts it on top.
- **Byte-identity is automatic.** With `Join == JoinNone` (the zero value),
  `renderTwoColumn` draws exactly as today; the join shapes emit only for
  `JoinBadge` / `JoinArrow`.
- **Determinism is free.** Fixed integer-EMU geometry, native shapes
  (`ShapeEllipse` for the badge, `ShapeRightArrow` for the connector), no media —
  two-column slides stay parallel-safe.
- **Two columns only.** A `TwoColumn` has exactly two columns, so the connector
  is the 2-column case (A → B). The general N-column architecture-diagram
  connector (the "arrows between 3 columns" in the product need) is a
  multi-column-container layout feature, deferred with sub-unit (c)'s successor —
  it is not in R5's acceptance.

## 4. Recommendations

1. Add `ColumnJoin` (`JoinNone`/`JoinBadge`/`JoinArrow`) and fields
   `Join ColumnJoin` + `JoinLabel string` to `TwoColumn`.
2. After the two column stacks render, draw the join element centered on the
   column seam: an accent ellipse + centered inverse label for `JoinBadge`, an
   accent right-arrow for `JoinArrow`. Pinned EMU sizes.
3. No Stage-1 validation needed (optional visual); an empty `JoinLabel` with
   `JoinBadge` draws the badge shape without text.
4. Land the row-labeled bento (R5 c) as its own phase — it is a new IR node with
   its own layout, validation, and rendering.

## 5. Open questions

- **N-column connectors (architecture diagram).** Arrows between 3+ columns need
  a multi-column container with connector routing; deferred (not in R5's
  acceptance, and larger than a LOW unit). Revisit if a real deck needs it.
- **Badge shape / size knobs.** The badge is a fixed-size accent ellipse; a
  caller-controlled size or shape is a plausible later field, deferred.
- **Vertical placement.** The element centers vertically in the `TwoColumn` box.
  A caller-chosen vertical position is deferred; center matches the reference.
