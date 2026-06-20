# Phase 23 — grow to fit

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §10.2 (content-bbox-driven layout, slot assignment)
**Deps:** Phase 13 (alignment + `alignedStackIn`), Phase 22 (content-aware
height). External: none.
**Status:** In progress

---

## 1. Goal

Add an opt-in vertical fill mode (`VAlignFill`) that pins fixed leaves at the top
and grows the flexible nodes (containers + Image/Chart) to consume the remaining
body height, so a sparse slide fills its frame instead of reading thin.

## 2. Why now

This is the second unit of **Wave 8 — post-V1 engine extensions**
(`DECKARD-PRODUCT-REQUIREMENTS.md` R2, HIGH), picked up immediately after R1
(Phase 22) per the agreed one-requirement-per-PR cadence. The product's
reference "designed" look is heading-at-top + content-grows-to-fill; today's
`VAlignCenter`/`VAlignJustify` only *float* the stack, they don't grow it. R2
builds directly on Phase 22's content-aware `preferredHeight` (the basis for the
slack arithmetic) and honors `RFC §10.2` (the engine assigns slots; this phase
lets a flexible node's slot absorb leftover height).

## 3. RFC sections implemented

- `RFC §10.2` — "a node reports its preferred bbox; the engine assigns it a
  slot; the node fits to the slot." This phase lets the slot assignment grow a
  flexible node beyond its preferred bbox to consume otherwise-empty body
  height. Scoped to the top-level body stack only (consistent with the Phase-13
  alignment scoping); container-internal layout is unchanged.

## 4. Brief findings incorporated

- `docs/research/10-grow-to-fit.md` — *a single new `VAlign` value is the
  cleanest surface* → adds `VAlignFill` to the enum; no new `SceneSlide` field,
  no per-node flag.
- `docs/research/10-grow-to-fit.md` — *the renderers already honor a taller box*
  → **no container renderer is changed**; `alignedStackIn` grows the flexible
  node's slot and `layout.Grid`/`layout.Columns`/card chrome/image/chart/table
  fill it.
- `docs/research/10-grow-to-fit.md` — *flexible is a fixed intrinsic set* →
  `isFlexible` = {`Grid`, `TwoColumn`, `Card`, `CardSection`, `Table`, `Chart`,
  `Image`}; `CodeBlock` is excluded (growing a code raster distorts it).
- `docs/research/10-grow-to-fit.md` — *proportional distribution is
  deterministic* → slack is shared in proportion to preferred height, remainder
  to the last flexible node, equal split when flexible heights sum to zero.

## 5. Findings I'm departing from

None. The brief's two refinement open-questions (recursive fill inside a
container; per-node grow weights) are explicitly deferred there, not departed
from.

## 6. Decisions referenced

- `D-026` — *Engine, not product.* Fill is a caller-driven mechanism: the engine
  provides `VAlignFill` and a deterministic distribution; it never decides a
  slide "looks thin" or fills one on its own.
- `D-051` — *Content-aware `preferredHeight`.* Fill consumes the slack computed
  from content-aware heights; the two compose (fill grows positive slack, the
  Phase-22 overflow warning fires when slack ≤ 0).
- This plan files **D-052 — `VAlignFill` grow-to-fit** in `docs/decisions.md`.

## 7. Architecture

```text
scene/align.go    VAlign enum: + VAlignFill   (new constant after VAlignJustify)

scene/render.go   alignedStackIn:
                    compute heights[] (content-aware, Phase 22) + totalH
                    overflow warning (unchanged: totalH > box.H)
                    if Vertical == VAlignFill && slack > 0:
                        distributeFill(nodes, heights, slack)   NEW
                    place top-pinned with standard gaps
                  isFlexible(SlideNode) bool                    NEW (the flex set)
                  distributeFill(nodes, heights, slack)         NEW (proportional)

Renderers:        UNCHANGED — layout.Grid scales rows to parent.H, layout.Columns
                  gives full-height columns, card chrome runs to box.Bottom(),
                  Image fills its box, Chart contains-to-fit its slot, Table is
                  built at box.H. The grown slot box propagates one nesting level
                  (a grown Grid hands its taller cell box to the child renderer).
```

`VAlignFill` is top-pinned (start `Y = box.Y`) and uses the standard inter-node
gap; it differs from `VAlignTop` only in that flexible nodes' slot heights are
enlarged to consume the slack, so the last node's bottom reaches the body
region's bottom margin.

## 8. Files added or changed

```text
scene/align.go                       # CHANGED — adds VAlignFill + String case
scene/render.go                      # CHANGED — isFlexible, distributeFill, VAlignFill in alignedStackIn
scene/render_fill_test.go            # NEW — fills to margin, proportional growth, zero-fill identity,
                                     #       no-flex degenerate, determinism
scripts/smoke/phase-23.sh            # NEW — phase smoke
docs/research/10-grow-to-fit.md      # NEW — informing brief
docs/research/INDEX.md               # CHANGED — registers brief 10
docs/plans/phase-23-grow-to-fit.md   # NEW — this plan
docs/plans/README.md                 # CHANGED — adds Phase 23 to Wave 8
docs/decisions.md                    # CHANGED — adds D-052
docs/glossary.md                     # CHANGED — adds "VAlignFill", "Flexible node", "Grow-to-fit"
docs/site/guide/scene.md             # CHANGED — alignment section documents VAlignFill (§19)
skills/compose-a-scene/SKILL.md      # CHANGED — VAlign options incl. fill (§19)
```

## 9. Public API surface

```go
// scene (align.go) — one added enum constant; no signature changes.
const (
    VAlignTop VAlign = iota
    VAlignCenter
    VAlignBottom
    VAlignJustify
    VAlignFill // NEW — top-pinned; flexible nodes grow to consume leftover body height
)
```

No new type, function, or field on the public surface beyond the enum constant.
`SceneSlide.Content.Vertical` carries it. Additive: every existing value keeps
its meaning and the zero value (`VAlignTop`) is unchanged.

## 10. Risks

- **R1 — determinism.** Distribution must be worker-count independent.
  **Mitigation:** pure integer EMU proportional split with a fixed remainder
  rule (last flexible node); a determinism test renders a `VAlignFill` deck
  under 1 vs N workers and asserts byte-identity.
- **R2 — backward-compat regression.** A change to `alignedStackIn` could
  perturb non-fill output. **Mitigation:** the fill branch is gated on
  `Vertical == VAlignFill`; all other paths are untouched. A test renders the
  same scene under every non-fill `VAlign` before/after and the full existing
  scene + align suites must pass unchanged.
- **R3 — degenerate fill (no flexible node).** **Mitigation:** `distributeFill`
  returns early when no flexible node is present, so the slide top-aligns (no
  growth, no warning) — covered by an explicit test.

## 11. Acceptance criteria

1. A heading (fixed) + grid (flexible) slide under `VAlignFill` renders the
   heading at the top and the grid's slot growing so its bottom reaches the body
   region's bottom margin.
2. Given a taller slot, `layout.Grid` produces proportionally taller cells (a
   grown Grid/Card/TwoColumn fills the extra height).
3. With two flexible nodes, the slack is shared in proportion to their preferred
   height (the larger grows more); the last flexible node absorbs the remainder.
4. A `VAlignFill` slide with no flexible node top-aligns (byte-identical to
   `VAlignTop`).
5. Zero fill (every non-`VAlignFill` mode) renders byte-identical to today.
6. A `VAlignFill` deck renders byte-identical across 1 vs N workers
   (determinism holds).
7. `make coverage` shows `scene` ≥ its band; `make preflight` passes.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry; new branches are covered by
`render_fill_test.go`.

## 13. Smoke check

`scripts/smoke/phase-23.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` fill grows the last flexible node to the bottom margin (criteria 1, 2).
3. `OK:` proportional split between two flexible nodes (criterion 3).
4. `OK:` no-flex `VAlignFill` top-aligns (criterion 4).
5. `OK:` non-fill modes byte-identical (criterion 5).
6. `OK:` `VAlignFill` render is deterministic across workers (criterion 6).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — `isFlexible` set; `distributeFill` proportional + remainder
  + equal-split fallback; `alignedStackIn` under `VAlignFill` (white-box).
- **Round-trip golden:** N/A — no builder primitive or new scene node; layout
  sizing only.
- **Integration** (`test/integration/`): no — internal to `scene` layout, no
  cross-subsystem seam.
- **Fuzz:** no — no parse/decode surface.
- **Benchmark:** optional — `alignedStackIn` is on the render hot path; not a
  gate.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Flexible node` — a scene node whose slot grows under `VAlignFill` (the
  containers plus `Image`/`Chart`).
- `Grow-to-fit` — the layout mode that distributes leftover body height to the
  flexible nodes so they consume the frame.
- `VAlignFill` — the body-stack vertical alignment that pins fixed leaves at the
  top and grows the flexible nodes to fill the remaining height.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-23.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entry D-052 added.
- [ ] Docs site updated for the `VAlignFill` surface (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
