# Phase 67 — prim icon label rows

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**RFC sections:** §11.1, §12
**Deps:** Phase 62 (Checklist row engine, D-095), the icon registry; brief 50.
**Status:** Done

---

## 1. Goal

Add an `IconRows` node — a vertical stack of `[icon | label | optional meta]` rows with an
optional pill frame and a fill-to-card mode — so an integrations / capabilities list reads
as designed rows rather than bullets.

## 2. Why now

R12.7 is a MED Wave-12 primitive built directly on the Phase-62 checklist row engine. The
recreation rendered these rows as bullet lists with the title overlapping. See
`docs/plans/README.md` Wave 12 and D-059 (engine-tagged).

## 3. RFC sections implemented

- `RFC §11.1` — extends the leaf catalog with `IconRows`.
- `RFC §12` — native policy (icon custGeom + text + optional pill); no asset.

## 4. Brief findings incorporated

- `docs/research/50-prim-icon-label-rows.md` — *"mirror the checklist row engine; a leaf
  with a Fill mode; GlyphColor defaults to accent; per-row icon validated"* → the composer
  and wiring.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-095` — Checklist — the row-height + Fill slack + icon-glyph pattern reused.
- `D-026` — engine not product. New: `D-100` — prim-icon-label-rows (this PR).

## 7. Architecture

```text
IconRows{Rows []IconRow{Icon, Label RichText, Meta RichText, Tone RowTone}, Fill, GlyphColor}
  per-row height = max wrappedLines(Label @ TypeBody, labelW) * lineH (>= iconSz)
  Fill: distribute inter-row slack (last row meets the bottom), like Checklist
  per row: [RowPill SurfaceAlt frame] [icon | label | right-aligned meta]
```

## 8. Files added or changed

```text
scene/nodes.go              # CHANGED — KindIconRows + String; RowTone; IconRow; IconRows struct
scene/policy.go             # CHANGED — KindIconRows (native, no asset)
scene/validate.go           # CHANGED — IconRows case (rows, tone range)
scene/render.go             # CHANGED — dispatch + preferredHeight + isFlexible(true) + nodeUsesAssets(false)
scene/render_card.go        # CHANGED — walkIconRefs case IconRows (per-row Icon)
scene/render_iconrows.go    # NEW — icon-rows composer
scene/render_iconrows_test.go ; render_iconrows_render_test.go # NEW — white/black-box
scene/scene_test.go         # CHANGED — allNodes + catalog 26 -> 27
scene/render_adversarial_test.go ; test/integration/roundtrip_test.go # CHANGED
scripts/smoke/phase-67.sh   # NEW
docs/research/50-...md + INDEX.md ; docs/plans/phase-67-...md + README.md
docs/glossary.md ; docs/design/THEME.md ; docs/site/catalog/text-leaves.md ; skills/compose-a-scene/SKILL.md ; docs/decisions.md (D-100)
```

## 9. Public API surface

```go
// scene
type RowTone int
const ( RowPlain RowTone = iota; RowPill )
type IconRow struct { Icon string; Label RichText; Meta RichText; Tone RowTone }
type IconRows struct { Rows []IconRow; Fill bool; GlyphColor ColorRole }
func (IconRows) NodeKind() NodeKind // KindIconRows
const KindIconRows NodeKind = ... // appended after KindBanner
```

## 10. Risks

- **R1 — per-row icon not validated.** **Mitigation:** `walkIconRefs case IconRows`; a
  test asserts an unknown row icon fails `Render`.
- **R2 — invisible default glyph.** **Mitigation:** `GlyphColor` zero (`ColorCanvas`) maps
  to `ColorAccent`; a test asserts the resolved color.
- **R3 — catalog/kind drift.** **Mitigation:** count 26 → 27 and the integration loop →
  `KindIconRows`.

## 11. Acceptance criteria

1. Each row shows its leading icon aligned to its label with no overlap onto the card
   title; an optional `Meta` right-aligns; `Fill` spreads rows to the card bottom.
2. `RowPill` frames each row in a `SurfaceAlt` rounded-rect.
3. An unknown row `Icon` fails Stage-1 validation.
4. Identical input is byte-identical across worker counts.
5. `scene` catalog count is 27; the integration kind-range loop covers `KindIconRows`.
6. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-67.sh` greps the new surface (KindIconRows, the struct, the composer,
policy/validate/walkIconRefs entries, catalog 27) and runs the white/black-box tests;
SKIPs gracefully before the code exists. OK ≥ the acceptance-criteria count, FAIL = 0.
