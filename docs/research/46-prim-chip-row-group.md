# Brief 46 — prim-chip-row-group

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**Authored:** 2026-06-23
**Motivating phase:** Phase 63 — prim-chip-row-group (R12.5, HIGH · both)

## 1. Question

A horizontal row of tag/category chips — "Understand · Operate · Execute", a labeled
bottom strip "COMMON BUILDS · Finance · HR · Sales · …" — is a universal slide element.
The IR has only a single inline `Chip`; the recreation rendered chip rows as broken
bullet lists (hanging dots, no pills) or dropped them. What primitive lays out a
*sequence* of chips with wrapping and an optional leading label?

## 2. Prior art surveyed

- **`scene/render_leaves.go renderChip`** — the single-chip pill: a `ShapeRoundRect` +
  a tone-driven fill (tint = `ColorSurfaceAlt`, solid = `Color`, outline = `Color`
  hairline) + a centered `TypeBodySmall` label. The chip-row reuses this exact pill
  treatment per chip.
- **`scene/render_button.go buttonWidthOf` (Phase 61) / `cardPillWidthOf`** — the
  content-fit-pill width pattern (`naturalWidth(label) + 2·pad`, clamp): the chip-row
  sizes each chip the same way.
- **`scene/render.go alignedStackIn`** — the body stack offsets a `Chip` per `HAlign`;
  the chip-row applies the same `HAlign` offset per *wrapped line*.
- **`scene/render_checklist.go` (Phase 62)** — the row-packing + per-row-height +
  pinned-metric-and-token-color pattern; the chip-row's greedy wrap is the horizontal
  analogue.
- **D-059:** R12.5 is `both`. Engine side = the `ChipRow` node + layout; the product
  side (emit a tag strip instead of a bullet list, a soul tone) is Deckard's.

## 3. Findings

- **Greedy left-to-right wrap, deterministic.** Each chip's width is content-fit
  (`naturalWidth(label@TypeBodySmall) + 2·pad`, plus an optional leading icon). Pack
  chips onto a line until the next would exceed `box.W`, then break — pure integer
  arithmetic over an ordered slice, no map iteration. A two-pass design (pack into
  lines, then place) lets the per-line `HAlign` offset and the slot-height estimate
  share one packer.
- **The leading label rides line 0.** A non-empty `Label` renders as a `TypeCaption`
  run before the first chip, consuming `labelW + gap` on the first line; it participates
  in line 0's width for alignment.
- **`Wrap` is the engine mechanism, zero = single line (D-026).** A plain Go bool can't
  encode "default true"; the engine zero value is the minimal behavior (one line), and
  the *product/soul* sets `Wrap: true` for a strip that should reflow. Wrapping is the
  safe path (keeps chips on-canvas); a `Wrap: false` row that overflows is the caller's
  explicit choice (the `Decoration.Bleed` posture). The adversarial fixture uses
  `Wrap: true` so the on-canvas invariant covers the reflow.
- **Per-chip icon needs validation.** `ChipSpec.Icon` (a closed-name registry icon)
  flows through `walkIconRefs case ChipRow` exactly like the button/checklist icons.
- **Pinned metrics, token colors.** Chip height, padding, icon size, and the inter-chip
  / inter-line gaps are pinned EMU; the tone fill colors are the existing `ChipTone` →
  token mapping (P2). Reuses `ChipTone`/`ColorRole` — no new token.

## 4. Recommendation

Add `KindChipRow` + a `ChipRow` leaf node `{Label string; Chips []ChipSpec{Label string;
Tone ChipTone; Color ColorRole; Icon string}; Wrap bool; Align HAlign}` and a
`scene/render_chiprow.go` composer: a shared greedy packer (`chipRowLines`) feeding both
`preferredHeight` (line count × chip height) and the renderer (per-line `HAlign` offset,
each chip a content-fit `RadiusFull`-ish rounded-rect pill with the `ChipTone` fill and
an optional leading custGeom icon, the leading `Label` on line 0). Full new-node wiring
(policy native, validate non-empty chips + tone range, `renderNode` dispatch +
`preferredHeight` + `isFlexible` false + `nodeUsesAssets` false + `nodeEffectiveHAlign`,
`walkIconRefs case ChipRow`, catalog 24 → 25, integration kind-range loop →
`KindChipRow`). Extend the R11.12 adversarial fixture with a hostile wrapping labeled
chip row. Additive ⇒ byte-identical when unused.
