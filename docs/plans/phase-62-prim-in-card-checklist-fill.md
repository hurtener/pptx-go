# Phase 62 — prim in card checklist fill

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**RFC sections:** §11.1 (leaf node catalog), §12 (per-node rendering policy)
**Deps:** Phase 61 (Button new-node wiring + icon-glyph pattern, D-094), Phase 47
(list hanging-indent, D-078), Phase 23 (VAlignFill slack math, D-052)
**Status:** Done

---

## 1. Goal

Add a `Checklist` scene node — true filled status glyphs (check/cross/dot), a
glyph-width hanging indent, 1–3 column row-major reflow, and a fill-to-box mode — so an
offer / pricing / comparison card gets a dense, self-distributing "what you get" list
instead of empty-square bullets.

## 2. Why now

R12.2 is the second CRITICAL Wave-12 primitive. The recreation's `ListChecklist`
renders empty white squares (the `Checked` bool is never read), a broken indent, and
cannot reflow or fill — the single most visible defect on every offer/pricing card.
Phase 61 just landed the icon-glyph + new-node pattern this builds on directly. See
`docs/plans/README.md` Wave 12 and D-059 (R12.2 is engine-tagged).

## 3. RFC sections implemented

- `RFC §11.1` — extends the leaf-node catalog with `Checklist`.
- `RFC §12` — native policy (custGeom glyphs + text), no `AssetID`.

## 4. Brief findings incorporated

- `docs/research/45-prim-in-card-checklist-fill.md` — *"the glyph is a curated icon, not
  a font checkbox"* → `CheckDone`/`CheckNo`/`CheckNeutral` map to the curated
  `check`/`x`/`dot` single-path SVGs rendered via `ps.AddIcon` with a token fill.
- `docs/research/45-...md` — *"glyph color: per-state default + optional `*ColorRole`
  override"* → `GlyphTone *ColorRole` (nil = per-state default), the D-054 pattern (a
  §4.3 deviation from the spec's value `ColorRole`).
- `docs/research/45-...md` — *"hanging indent = glyph width + gap; columns reflow
  row-major; Fill distributes inter-row slack like Justify"* → the composer geometry.
- `docs/research/45-...md` — *"add `Checklist` to `isFlexible` so a parent can grow it
  to fill"* → wired so `VAlignFill`/`BodyVAlign` cards can spread a short list.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-094` — Button — the icon-glyph fill + new-node wiring pattern reused verbatim.
- `D-054` — rich card visuals — the `*ColorRole` "nil = default" pattern for `GlyphTone`.
- `D-078` — list bullet indent density — the hanging-indent calibration approach.
- `D-052` — vertical fill — the slack-distribution math `Fill` mirrors per-row.
- `D-026` — engine not product — the glyphs/columns/fill are mechanisms; the engine
  renders `Text` verbatim and picks no content.
- New: `D-095` — prim-in-card-checklist-fill (filed in this PR).

## 7. Architecture

```text
Checklist{Items[]{Text,State,Icon}, Columns, GlyphTone *ColorRole, Fill}
  cols = clamp(Columns,1,3); rows = ceil(n/cols); item i -> (row=i/cols, col=i%cols)
  colW = (box.W - colGap*(cols-1)) / cols ;  textColW = colW - glyphSz - glyphGap
  rowHeights[r] = max cell wrappedLines(text@TypeBody, textColW) * lineH
  Fill: rowGap = (box.H - Σrows)/(rows-1) so the last row meets box.Bottom()
  per cell: [glyph custGeom (check/x/dot, per-state|GlyphTone token fill) | text frame]
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — KindChecklist + String; CheckState; ChecklistItem; Checklist struct
scene/policy.go                      # CHANGED — KindChecklist policy (native, no asset)
scene/validate.go                    # CHANGED — Checklist case (items, columns 0..3, state range)
scene/render.go                      # CHANGED — renderNode dispatch + preferredHeight + isFlexible(true) + nodeUsesAssets(false)
scene/render_card.go                 # CHANGED — walkIconRefs case Checklist (per-item Icon overrides)
scene/render_checklist.go            # NEW — checklist geometry + composer
scene/render_checklist_test.go       # NEW — white-box: columns reflow, glyph-per-state, hanging indent, fill
scene/render_checklist_render_test.go# NEW — black-box: filled glyph emitted (not empty box), columns, unknown icon fails
scene/scene_test.go                  # CHANGED — allNodes + catalog count 23 -> 24
scene/render_adversarial_test.go     # CHANGED — hostile 2-column filled checklist
test/integration/roundtrip_test.go   # CHANGED — everyNodeScene + kind-range loop -> KindChecklist + slide counts
scripts/smoke/phase-62.sh            # NEW — phase smoke
docs/research/45-...md               # NEW — brief; docs/research/INDEX.md — registers it
docs/plans/phase-62-...md            # NEW — this plan; docs/plans/README.md — adds Phase 62
docs/glossary.md                     # CHANGED — Checklist, check state
docs/design/THEME.md                 # CHANGED — checklist glyph tone note (tokens)
docs/site/catalog/text-leaves.md     # CHANGED — Checklist
skills/compose-a-scene/SKILL.md      # CHANGED — Checklist node
docs/decisions.md                    # CHANGED — adds D-095
```

## 9. Public API surface

```go
// scene
type CheckState int
const ( CheckDone CheckState = iota; CheckNo; CheckNeutral )

type ChecklistItem struct { Text RichText; State CheckState; Icon string }

type Checklist struct {
    Items     []ChecklistItem
    Columns   int        // 1..3 (0 = 1)
    GlyphTone *ColorRole // nil = per-state default color
    Fill      bool       // distribute rows to fill the box height
}
func (Checklist) NodeKind() NodeKind // KindChecklist

const KindChecklist NodeKind = ... // appended after KindButton
```

## 10. Risks

- **R1 — empty-square regression.** The whole point is a *filled* glyph. **Mitigation:**
  a black-box test asserts the emitted XML carries a `custGeom` glyph and the
  `check`/`x`/`dot` path is present, with no `bu␣Char`/checkbox bullet autonumber.
- **R2 — per-item icon not validated.** **Mitigation:** `walkIconRefs case Checklist`
  validates every non-empty `Item.Icon`; a test asserts an unknown one fails `Render`.
- **R3 — `isFlexible` change affects fill modes.** Adding `Checklist` to `isFlexible`
  lets a `VAlignFill` parent grow it. **Mitigation:** a non-Fill checklist in a grown
  box top-aligns its rows (no overlap, occupies the box); covered by a determinism +
  on-canvas test. Default top-anchored stacks are unaffected (`isFlexible` is only read
  under the opt-in fill modes).
- **R4 — catalog/kind drift.** **Mitigation:** count 23 → 24 and the integration loop →
  `KindChecklist`; both fail loudly if missed.

## 11. Acceptance criteria

1. Every checklist row shows a **solid** glyph (a custGeom check/cross/dot), never an
   empty box; the text never collides with the glyph (hanging indent from glyph width).
2. A 2-column checklist reflows items row-major into balanced columns within `box.W`
   and never overflows the width.
3. `Fill` makes a short list span the box height evenly (last row meets the bottom);
   `Fill=false` top-aligns at a pinned gap.
4. An unknown per-item `Icon` fails icon validation at `Render`.
5. Identical input is byte-identical across worker counts.
6. `scene` catalog count is 24; the integration kind-range loop covers `KindChecklist`.
7. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-62.sh` greps for the new surface (KindChecklist, the struct +
CheckState, the composer, the policy/validate entries, the walkIconRefs case, the
catalog count 24) and runs the white/black-box tests; SKIPs gracefully before the code
exists. OK ≥ the acceptance-criteria count, FAIL = 0.
