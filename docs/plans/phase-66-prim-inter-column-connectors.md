# Phase 66 — prim inter column connectors

**Subsystem:** scene — Layer 2 renderer (Grid field extension)
**RFC sections:** §11.2 (Grid), §12 (policy)
**Deps:** Phase 06 (Flow `renderConnector`, D-044), the `layout.Grid` cell layout; brief 49.
**Status:** Done

---

## 1. Goal

Add `Grid.Connectors` — glyphs drawn in the gutters between adjacent columns — so a
3+-column architecture / pipeline grid reads as data flow, not just adjacency.

## 2. Why now

R12.4 is a HIGH Wave-12 primitive. The recreation's architecture grid left cards floating
disconnected; `TwoColumn.Join` only handles a single seam. See `docs/plans/README.md`
Wave 12 and D-059 (engine-tagged).

## 3. RFC sections implemented

- `RFC §11.2` — extends `Grid` with the additive `Connectors` field + `ConnectorBiArrow`.
- `RFC §12` — native policy (preset-geometry glyphs); no new asset.

## 4. Brief findings incorporated

- `docs/research/49-prim-inter-column-connectors.md` — *"the gutter box is derived from the
  cell boxes; reuse `renderConnector`"* → `renderGridConnectors`.
- `49-...md` — *"`ConnectorBiArrow` = `leftRightArrow`/`upDownArrow`, a prst attr (no
  namespace change)"* → the one new glyph case in `renderConnector`.
- `49-...md` — *"adjacency validated at Stage-1; empty ⇒ byte-identical"*.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-044` — flow connectors — the `renderConnector` glyph set reused.
- `D-055` — column join — the single-seam sibling (R12.8 extends that; R12.4 is the
  N-column gutter case). `D-026` — engine not product. New: `D-099` (this PR).

## 7. Architecture

```text
Grid.Connectors []GridConnector{Between [2]int, Kind, Label}
  renderGrid: layout.Grid -> cell boxes; renderGridConnectors(box, v, cells)
    per connector {c, c+1}: gutter = {X: cells[c].Right(), W: cells[c+1].X-..., Y: box.Y, H: box.H}
    Label -> caption in the lower third; renderConnector(glyphGutter, Kind, vertical=false)
  ConnectorBiArrow -> leftRightArrow (horizontal) / upDownArrow (vertical)
```

## 8. Files added or changed

```text
scene/nodes.go              # CHANGED — ConnectorBiArrow; GridConnector; Grid.Connectors
scene/validate.go           # CHANGED — Grid connector adjacency/range/kind validation
scene/render_flow.go        # CHANGED — ConnectorBiArrow case in renderConnector
scene/render_container.go   # CHANGED — renderGrid calls renderGridConnectors
scene/render_gridconn_render_test.go # NEW — black-box: glyphs, bi-arrow, additive, validation, determinism
scene/render_adversarial_test.go     # CHANGED — connectors on the cards grid
scripts/smoke/phase-66.sh   # NEW — phase smoke
docs/research/49-...md + INDEX.md ; docs/plans/phase-66-...md + README.md
docs/glossary.md ; docs/site/catalog/containers.md ; skills/compose-a-scene/SKILL.md ; docs/decisions.md (D-099)
```

## 9. Public API surface

```go
// scene
type GridConnector struct { Between [2]int; Kind ConnectorKind; Label string }
// Grid gains: Connectors []GridConnector
const ConnectorBiArrow ConnectorKind = ... // appended after ConnectorPlus
```

## 10. Risks

- **R1 — gutter off-canvas / overlap.** **Mitigation:** the gutter is derived from the
  deterministic cell boxes and lies strictly between them; the adversarial fixture's
  on-canvas invariant covers it.
- **R2 — bad indices.** **Mitigation:** Stage-1 validation rejects non-adjacent /
  out-of-range / bad-kind connectors; the renderer also guards defensively.
- **R3 — byte-identity.** **Mitigation:** empty `Connectors` calls into the helper which
  returns immediately; the existing grid tests pass unchanged.

## 11. Acceptance criteria

1. A connector renders centered in the correct gutter between the two named columns and
   scales with the gutter width; a `ConnectorBiArrow` is a bidirectional arrow.
2. A grid with no `Connectors` is byte-identical to before; output is byte-identical
   across worker counts.
3. A non-adjacent or out-of-range connector fails Stage-1 validation.
4. `make coverage` shows `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-66.sh` greps the new surface (ConnectorBiArrow, GridConnector,
`Grid.Connectors`, `renderGridConnectors`) and runs the black-box tests; SKIPs gracefully
before the code exists. OK ≥ the acceptance-criteria count, FAIL = 0.
