# Brief 50 — prim-icon-label-rows

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**Authored:** 2026-06-23
**Motivating phase:** Phase 67 — prim-icon-label-rows (R12.7, MED · engine)

## 1. Question

A card that is a vertical stack of `[icon | label | optional meta]` rows — integrations
("Salesforce · Slack"), capabilities ("Chat & Q&A", "Specialized agents") — reads as
designed rows, not bullets. The recreation rendered these as plain bullet lists with the
title overlapping. What primitive pairs a leading icon with a label (and optional trailing
meta) per row and fills the card?

## 2. Prior art surveyed

- **`scene/render_checklist.go` (Phase 62)** — the row-list pattern: per-row content-aware
  heights, a `[glyph | text]` layout, a `Fill` mode distributing inter-row slack, the
  icon-registry glyph fill, `walkIconRefs` for per-row icons. IconRows is the same shape
  with an arbitrary leading icon, a rich label, an optional right-aligned meta, and a pill
  frame.
- **`scene/render_chiprow.go drawChip` / `render_card.go` pill** — the `SurfaceAlt`
  rounded-rect frame the `RowPill` tone reuses.
- **`scene/metrics.go` `naturalWidth` / `wrappedLines`** — meta width + per-row line count.
- **D-059:** R12.7 is `engine`.

## 3. Findings

- **A leaf node, single column, with a `Fill` mode.** `IconRows{Rows []IconRow{Icon string;
  Label RichText; Meta RichText; Tone RowTone}; Fill bool; GlyphColor ColorRole}`. Like
  `Checklist` it is added to `isFlexible` so a `VAlignFill` card can grow it; `Fill`
  distributes inter-row slack so the rows span the card height.
- **`GlyphColor` defaults to accent.** `ColorRole`'s zero is `ColorCanvas` (a real color);
  a canvas-colored icon is invisible, so the zero value maps to `ColorAccent` (the
  Banner/ribbon pattern). Documented.
- **Per-row layout: `[icon | label | meta]`.** The icon is vertically centered on the
  first line; the label takes the middle; an optional `Meta` is right-aligned in a
  content-fit column (clamped to a third of the row). `RowPill` draws a `RadiusMD`
  `SurfaceAlt` rounded-rect behind the row, with the content inset by a small pad.
- **Per-row icon validated.** `walkIconRefs case IconRows` validates each row's `Icon` (a
  closed-name registry icon). A row with an empty `Icon` simply omits the glyph and the
  label starts at the left.
- **Pinned metrics, token colors.** Icon size, gaps, row gap, line height, pill pad are
  pinned EMU; the glyph color and the pill surface are tokens (P2).

## 4. Recommendation

Add `KindIconRows` + an `IconRows` leaf node and a `scene/render_iconrows.go` composer
mirroring the checklist row engine: content-aware per-row heights, `[icon | label | meta]`
layout, optional `RowPill` frame, and a `Fill` slack distribution. Full new-node wiring
(policy native, validate rows + tone range, `renderNode` dispatch + `preferredHeight` +
`isFlexible` true + `nodeUsesAssets` false, `walkIconRefs case IconRows`, catalog 26 → 27,
integration kind-range loop → `KindIconRows`). Extend the R11.12 adversarial fixture with a
filled `RowPill` icon-rows list under hostile content. Additive ⇒ byte-identical when
unused.
