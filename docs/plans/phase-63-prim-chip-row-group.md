# Phase 63 — prim chip row group

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**RFC sections:** §11.1 (leaf node catalog), §12 (per-node rendering policy)
**Deps:** Phase 62 (Checklist row-packing pattern, D-095), Phase 61 (content-fit pill,
D-094), the existing `Chip` pill (`render_leaves.go`)
**Status:** Done

---

## 1. Goal

Add a `ChipRow` scene node — a wrapping horizontal row of content-fit chip pills with an
optional leading label — so a slide can carry a tag / category / capability strip
instead of a single inline `Chip` or a broken bullet list.

## 2. Why now

R12.5 is the first HIGH Wave-12 primitive after the two CRITICALs. The recreation
rendered chip rows as bullet lists or dropped them; the IR has only a single `Chip`.
See `docs/plans/README.md` Wave 12 and D-059 (engine side of a `both` req).

## 3. RFC sections implemented

- `RFC §11.1` — extends the leaf catalog with `ChipRow`.
- `RFC §12` — native policy (rounded-rect pills + text + optional custGeom icons), no
  `AssetID`.

## 4. Brief findings incorporated

- `docs/research/46-prim-chip-row-group.md` — *"greedy left-to-right wrap, two-pass
  packer shared by render + preferredHeight"* → `chipRowLines`.
- `46-...md` — *"leading label rides line 0"* → the `Label` consumes `labelW + gap` on
  the first line and participates in its alignment.
- `46-...md` — *"`Wrap` is the engine mechanism, zero = single line (D-026)"* → a plain
  bool; the product sets `Wrap: true`.
- `46-...md` — *"per-chip icon validated via `walkIconRefs`"* → `case ChipRow`.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-095` / `D-094` — Checklist / Button — the new-node wiring + content-fit-pill +
  pinned-metric/token-color patterns reused.
- `D-026` — engine not product — `Wrap` and the tones are mechanisms; the zero value is
  the minimal (single-line) behavior, the product drives the rest.
- New: `D-096` — prim-chip-row-group (filed in this PR).

## 7. Architecture

```text
ChipRow{Label, Chips[]{Label,Tone,Color,Icon}, Wrap, Align}
  chipWidthOf(chip) = naturalWidth(label@TypeBodySmall) + 2*padX (+ icon)
  chipRowLines(box.W): greedy-pack chips into lines (label rides line 0); Wrap=false → 1 line
  per line: startX = box.X + hAlignOffset(box.W - lineW); place label (line 0) then chips
  each chip: RadiusFull rounded-rect (ChipTone fill) + optional leading icon + centered label
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — KindChipRow + String; ChipSpec; ChipRow struct
scene/policy.go                      # CHANGED — KindChipRow policy (native, no asset)
scene/validate.go                    # CHANGED — ChipRow case (chips, tone range)
scene/render.go                      # CHANGED — dispatch + preferredHeight + isFlexible(false) + nodeUsesAssets(false) + nodeEffectiveHAlign
scene/render_card.go                 # CHANGED — walkIconRefs case ChipRow
scene/render_chiprow.go              # NEW — chip-row packer + composer
scene/render_chiprow_test.go         # NEW — white-box: wrap packing, content-fit width, label, align
scene/render_chiprow_render_test.go  # NEW — black-box: pills (not bullets), wrap, unknown icon fails, determinism
scene/scene_test.go                  # CHANGED — allNodes + catalog count 24 -> 25
scene/render_adversarial_test.go     # CHANGED — hostile wrapping labeled chip row
test/integration/roundtrip_test.go   # CHANGED — everyNodeScene + kind-range loop -> KindChipRow
scripts/smoke/phase-63.sh            # NEW — phase smoke
docs/research/46-...md + INDEX.md    # NEW / CHANGED — brief 46
docs/plans/phase-63-...md + README.md# NEW / CHANGED — this plan + Wave 12 entry
docs/glossary.md                     # CHANGED — ChipRow
docs/site/catalog/text-leaves.md     # CHANGED — ChipRow
skills/compose-a-scene/SKILL.md      # CHANGED — ChipRow node
docs/decisions.md                    # CHANGED — adds D-096
```

## 9. Public API surface

```go
// scene
type ChipSpec struct { Label string; Tone ChipTone; Color ColorRole; Icon string }

type ChipRow struct {
    Label string      // optional leading TypeCaption label; "" = none
    Chips []ChipSpec
    Wrap  bool         // wrap to new lines (zero = single line)
    Align HAlign       // per-node horizontal alignment override; 0 = inherit slide
}
func (ChipRow) NodeKind() NodeKind // KindChipRow

const KindChipRow NodeKind = ... // appended after KindChecklist
```

## 10. Risks

- **R1 — chips render as bullets again.** **Mitigation:** a black-box test asserts the
  XML carries N `roundRect` pills and no `buChar`/`buAutoNum`.
- **R2 — per-chip icon not validated.** **Mitigation:** `walkIconRefs case ChipRow`; a
  test asserts an unknown chip icon fails `Render`.
- **R3 — wrap overflow off-canvas.** **Mitigation:** greedy wrap keeps each line within
  `box.W`; the adversarial fixture (`Wrap: true`, long labels) asserts on-canvas.
- **R4 — catalog/kind drift.** **Mitigation:** count 24 → 25 and the integration loop →
  `KindChipRow`.

## 11. Acceptance criteria

1. Chips render as real pills (never bullets); a labeled row shows the label then the
   chips.
2. With `Wrap: true`, chips that exceed the width wrap onto new lines without clipping,
   with even gaps; each line stays within `box.W`.
3. `Align` offsets each line's chips left / center / right within the box.
4. An unknown `ChipSpec.Icon` fails icon validation at `Render`.
5. Identical input is byte-identical across worker counts.
6. `scene` catalog count is 25; the integration kind-range loop covers `KindChipRow`.
7. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-63.sh` greps for the new surface (KindChipRow, the struct, the
composer + packer, policy/validate/walkIconRefs entries, catalog 25) and runs the
white/black-box tests; SKIPs gracefully before the code exists. OK ≥ the
acceptance-criteria count, FAIL = 0.
