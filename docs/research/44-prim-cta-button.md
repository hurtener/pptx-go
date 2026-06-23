# Brief 44 — prim-cta-button

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**Authored:** 2026-06-23
**Motivating phase:** Phase 61 — prim-cta-button (R12.1, CRITICAL · both)

## 1. Question

A professional sales/investor deck ends on a button — a "Talk to the team →"
closing CTA, a "Start free" / "Contact sales" at the foot of every pricing card,
a "Start free →" inside a promo banner. The scene IR has no button primitive at
all (no `KindButton`), so a closing slide stops at bare prose and a pricing card
stops at the price with empty space where the action belongs. What is the minimal
first-class action/button primitive the engine should add, and how does it wire
into the catalog without disturbing byte-identity for decks that never use it?

## 2. Prior art surveyed

- **`scene/render_card.go` header pill** — `cardPillWidthOf` = `naturalWidth(label
  @ role) + 2·padX`, floored at a circular minimum, clamped to inner width; the
  drawn pill is `ShapeRoundRect` + `WithRadius(RadiusFull)` + a middle-anchored
  centered text frame, with a `fitScale` tail when the label is clamped. This is
  exactly the content-fit-pill geometry a button needs.
- **`scene/render_card.go` icon path** — `r.cfg.icons.Lookup(name)` → `ps.AddIcon(svg,
  box)` renders a native custGeom glyph (media-free); the closed-name icon registry
  is the seam, and `walkIconRefs` (consumed by `validateIconRefs`) Stage-1-validates
  every icon name in the tree. A button's leading/trailing icons reuse both.
- **`scene/render_stat.go` / `render_leaves.go renderChip`** — the leaf-renderer
  shape: `AddTextFrame(box).Anchor(AnchorMiddle)` + a centered paragraph; `Chip`
  already maps a tone enum to fill/line treatments (Solid / Outline / Tint).
- **`scene/contrast.go onCardSurface`** — variant-aware auto-contrast token (light
  on dark, nil on light = byte-identical); a filled button's label resolves against
  its own tone-resolved fill the same way the pill label resolves against
  `ColorSurfaceAlt`.
- **`scene/metrics.go fitScale`** — the deterministic shrink ratio (0 when it fits);
  reused so a label too wide for its clamped box stays one line.
- **New-node wiring precedent** — R5 Bento (D-056) and R6 Stat (D-057) document the
  full checklist (NodeKind enum + String, policy, validate, render dispatch +
  preferredHeight + isFlexible + nodeUsesAssets, walk* recursions, catalog count,
  integration kind-range loop). Button is a leaf like Stat (no body), but unlike
  Stat it carries icon names, so it must extend `walkIconRefs`.
- **D-059 scope:** R12.1 is tagged `both`. The engine side is the `Button` node +
  its native render; the product side (a soul-tokened default tone, the contract IR
  field, agent guidance to end on a button) operates on Deckard's own packages and
  is out of this repo.

## 3. Findings

- **Button is a leaf, sized content-fit, presentational only.** It carries no child
  nodes and no `AssetID` (native shapes + custGeom icons), so `nodeUsesAssets` is
  false and it composes in the parallel pool. Per the spec it is a presentational
  shape — **no hyperlink/action wiring** (the deck is static), so there is no new
  OOXML capability and no builder change (P1): it composes existing `AddShape` /
  `AddIcon` / `AddTextFrame` calls.
- **Width = content-fit, clamped to the box.** `leading-icon + label + trailing-icon
  + 2·pad`, where each present icon contributes `iconSz + gap`. Clamp to `box.W`;
  when `Align` centers/right-aligns and the fit width is narrower than the box, the
  pill is offset within the box (the body stack already offsets a `Chip` this way —
  the same nodeNaturalWidth-driven offset path). A label clamped to the box gets a
  `fitScale` tail so it never wraps.
- **Tone → ColorRole, with a ghost outline variant.** `ButtonPrimary` = `ColorAccent`
  solid / `TextInverse`; `ButtonAccentAlt` = `ColorAccentAlt` solid / `TextInverse`;
  `ButtonNeutral` = `ColorSurfaceAlt` solid / default text; `ButtonGhost` = `NoFill`
  + an accent hairline (`WithLine`) / `TextAccent`. All token-bound (P2) so a theme
  swap re-skins it. The label auto-contrasts against the resolved fill via
  `onCardSurface` for the solid tones (nil on a light surface = byte-identical).
- **Size → a pinned height + padding + icon-size scale.** SM/MD/LG are pinned layout
  metrics (a height/inset triple), not theme tokens — the same call the brief made
  for `cardPillPadX` and the list tight-indent base. Documented as a pinned metric,
  not a token (no THEME.md entry for the size scale itself; the *colors* are tokens).
- **Byte-identity is automatic.** A deck with no `Button` emits nothing new; the node
  is purely additive. The icon names flow through the existing `walkIconRefs` /
  `validateIconRefs` Stage-1 path once `case Button` is added (empty names are
  skipped by the existing `check`).
- **`RadiusFull` pill matches the reference.** `ShapeRoundRect` + `WithRadius(
  RadiusFull)` is the same capsule the header pill draws; the button is a larger,
  tone-filled instance of it with optional flanking glyphs.

## 4. Recommendation

Add `KindButton` + a `Button` leaf node `{ Label; Tone ButtonTone; Size ButtonSize;
LeadingIcon, TrailingIcon string; Align HAlign }`, render it in a new
`scene/render_button.go` composer (content-fit `RadiusFull` pill, tone-resolved fill
/ ghost hairline, middle-anchored bold `TypeBody` label flanked by native custGeom
icons, `fitScale` tail, `Align` offset), and do the full new-node wiring checklist:
policy (native, no asset), validate (non-empty label), render dispatch +
`preferredHeight` (size-driven) + `isFlexible` (false) + `nodeUsesAssets` (false),
extend `walkIconRefs` with `case Button` (both icon fields), bump the catalog count
to 23 and extend the integration kind-range loop to `KindButton`. Tone colors are
tokens (P2); the size scale is a pinned metric. Additive ⇒ byte-identical when
unused. Extend the R11.12 adversarial harness with a hostile (long-label, ghost,
dark-variant) button so the on-canvas / contrast / fit invariants cover it.
