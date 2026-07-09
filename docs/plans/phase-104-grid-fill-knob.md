# Phase 104 — grid and bento fill knob

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §10.2 (layout policy)
**Deps:** Phase 13 (alignment), Phase 41 (content-weighted bento rows, D-072), Phase 44 (fill-capped, D-075), Phase 103 (estimator gap parity, D-142)
**Status:** Done

---

## 1. Goal

Add an opt-in `Fill bool` on `Grid` and `Bento` so a container can claim the body stack's leftover height under `VAlignTop` without turning on slide-wide `VAlignFill`.

## 2. Why now

Deckard R15.4's engine half is the remaining agreed Wave-16 engine unit after Phase 103. The renderer already knows how to subdivide a taller grid/bento box into equal-height rows; the missing mechanism is a per-node opt-in independent of slide-level `VAlignFill`.

## 3. RFC sections implemented

- `RFC §10.2` — container nodes can opt into a taller assigned slot while keeping deterministic body-stack ordering and overflow semantics.

## 4. Brief findings incorporated

- `docs/research/31-estimate-actual-parity.md` — deterministic layout parity matters when a box grows; `Grid`/`Bento` already honor a taller box, so this phase opens the node-side seam.
- `docs/research/22-content-aware-height.md` — integer-EMU only; the fill route reuses the existing `distributeFill` math.

## 5. Findings I'm departing from

- R15.4's phrasing suggests a broad "equal-height fill" gap. **Departing because** equal-height cells and slide-level `VAlignFill` already ship; this phase adds only the missing per-node opt-in.
- The broadest possible interpretation would let `Fill` run under `VAlignCenter` / `VAlignBottom` / `VAlignJustify` / `VAlignBalanced`. **Departing because** those modes compute `startY` / gap distribution from pre-fill `totalH`; a Fill-grown stack would overshoot or fight the gap math. V1 restricts the route to `VAlignTop` only; the field comment documents that scope.

## 6. Decisions referenced

- `D-026` — engine, not product — this is a mechanism, not a taste heuristic.
- `D-072` — weighted bento rows — a taller bento box already subdivides correctly.
- `D-075` — fill-capped — reuses `distributeFill` semantics and keeps slide-level `VAlignFillCapped` unchanged.
- `D-092` — safe-area clamp — a Fill-grown grid/bento still clamps through `renderNode` before subdivision.
- `D-142` — estimator inter-node gap parity — the body stack's slot math is already truthful before this phase routes slack into Fill nodes.
- **New:** `D-143` — per-Grid/Bento Fill opt-in region-fill under `VAlignTop`.

## 7. Architecture

```text
scene/nodes.go
  Grid  { ..., Fill bool }
  Bento { ..., WeightedRows bool, Fill bool }

scene/render.go
  isFillWant(SlideNode) bool          // Grid/Bento with Fill=true
  alignedStackIn(..., VAlignTop):
    if slack > 0 and any node wants Fill:
      partition those nodes only
      distributeFill(fillNodes, fillHeights, slack)
      write back heights[idx]

scene/render_container.go / render_bento.go
  unchanged geometry; the grown box already produces taller rows/cells
```

`Fill` is additive and local: when unset, the slide body stack is byte-identical. Under `VAlignTop`, slack routes only to `Fill=true` grids/bentos; other flexible nodes remain at preferred height.

## 8. Files added or changed

```text
scene/nodes.go                              # CHANGED — Grid.Fill and Bento.Fill
scene/render.go                             # CHANGED — isFillWant + VAlignTop Fill routing in alignedStackIn
scene/render_fillnode_test.go               # NEW — white-box Fill routing tests
scene/render_parallel_test.go               # CHANGED — determinism case for Grid/Bento Fill
scene/render_adversarial_test.go            # CHANGED — Fill fixture stays on-canvas and fills
scripts/smoke/phase-104.sh                  # NEW — phase smoke
docs/plans/phase-104-grid-fill-knob.md      # NEW — this plan
docs/decisions.md                           # CHANGED — adds D-143
docs/glossary.md                            # CHANGED — `Fill (Grid/Bento)`
docs/site/guide/scene.md                    # CHANGED — document per-node Fill on Grid/Bento under VAlignTop
skills/compose-a-scene/SKILL.md             # CHANGED — Grid/Bento field tables mention Fill
```

## 9. Public API surface

```go
// scene
type Grid struct {
    Columns    int
    Ratio      []int
    Gap        SpaceRole
    Cells      []SlideNode
    Connectors []GridConnector
    Fill       bool
}

type Bento struct {
    Columns      int
    Rows         []BentoRow
    WeightedRows bool
    Fill         bool
}
```

Both zero values are backward-compatible.

## 10. Risks

- **R1 — fill/gap interaction drift.** `VAlignTop` only. **Mitigation:** skip the route under every other vertical mode.
- **R2 — accidental broad growth.** Other flexible nodes must not grow when only one grid/bento asks for Fill. **Mitigation:** route a partitioned subset through `distributeFill`.
- **R3 — public-surface drift.** New scene fields require docs/skill updates. **Mitigation:** land glossary + docs site + compose-a-scene updates in the same PR.

## 11. Acceptance criteria

1. `Grid{Fill:true}` under `VAlignTop` grows to consume the body stack's leftover height; the last row's bottom reaches the body-region bottom.
2. `Bento{Fill:true}` under `VAlignTop` grows likewise; default equal-mode and weighted rows both honor the taller box.
3. `Fill=false` on Grid/Bento is byte-identical to the pre-phase render.
4. Slide-level `VAlignFill`, `VAlignFillCapped`, `VAlignCenter`, `VAlignBottom`, `VAlignJustify`, `VAlignBalanced`, and `VAlignFit` are unchanged by the new field (no routing there).
5. Worker-count determinism holds.
6. `render_adversarial_test.go` invariants remain green.
7. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-104.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` Grid Fill under `VAlignTop` fills the body region.
3. `OK:` Bento Fill under `VAlignTop` fills the body region.
4. `OK:` Fill=false is byte-identical.
5. `OK:` worker-count determinism stays green.
6. `OK:` docs/skill updated for the new public field.

## 14. Tests

- **Unit:** `scene` white-box fill-routing tests (`alignedStackIn`, grid/bento bottom parity, byte-identity when Fill=false).
- **Round-trip golden:** yes — a Fill grid/bento fixture renders byte-stably and Fill=false fixtures are unchanged.
- **Integration:** no new subsystem seam.
- **Fuzz:** none.
- **Benchmark:** none.

## 15. Vocabulary added

- `Fill (Grid/Bento)` — a per-container opt-in that claims leftover body height under `VAlignTop`.

## 16. Plan deviations encountered during implementation

- **Per-node Fill restricted to `VAlignTop` only.** This was planned and held. The
  field stays inert under `VAlignCenter` / `VAlignBottom` / `VAlignJustify` /
  `VAlignBalanced` / `VAlignFit` because those modes compute `startY` / gaps from
  pre-fill `totalH`; routing Fill there would overshoot or fight their own slack
  distribution.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-104.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-143).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated.
