# Phase 44 — fill cap (no over-grow)

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §10 (layout engine)
**Deps:** Phase 23 (`VAlignFill`/`distributeFill`, D-052), Phase 40 (VAlignFit, D-071)
**Status:** Done

---

## 1. Goal

Add an opt-in `VAlignFillCapped` body-stack mode that grows flexible nodes only up
to a pinned ceiling, turning the leftover slack into balanced whitespace instead
of one ballooned node.

## 2. Why now

R10.6 is the next Wave-10 unit. It fixes the recreation's "Canvas" card that grew
to an enormous height holding one sentence under `VAlignFill` (DECKARD R10.6 gap).
It composes the `distributeFill` primitive (D-052) and mirrors the opt-in
`VAlignFit` shape (D-071).

## 3. RFC sections implemented

- `RFC §10` — the body-stack fill gains a bounded variant; uncapped `VAlignFill`
  is unchanged.

## 4. Brief findings incorporated

- `docs/research/27-fill-cap-no-overgrow.md` — *the cap is a bounded
  `distributeFill`; the residual becomes balanced spacing* → `distributeFillCapped`
  returns `used`; `alignedStackIn` distributes `slack − used` as even spacing.
- `docs/research/27-fill-cap-no-overgrow.md` — *opt-in as a new `VAlign` value
  mirroring `VAlignFit`; a pinned `growthMax` beats per-node `MaxGrow`* →
  `VAlignFillCapped` + `fillGrowthMaxBP = 10000` (≤ 1× preferred added).
- `docs/research/27-fill-cap-no-overgrow.md` — *even spacing = `residual/(n+1)`
  into the top margin and inter-node gaps; reuse the Justify/Fit offset mechanism*
  → `startY += space`, `effectiveGap += space`.

## 5. Findings I'm departing from

- The spec offers a per-node `MaxGrow` as an alternative to a pinned `growthMax`.
  This plan uses the **pinned `growthMax`** only. **Departing because** a per-node
  cap would touch every flexible node type for marginal benefit; the
  `distributeFillCapped` seam can take a per-node cap later. (§4.3.)

## 6. Decisions referenced

- `D-052` — `VAlignFill`/`distributeFill` — the primitive this bounds.
- `D-071` — `VAlignFit` — the opt-in-new-`VAlign` shape mirrored here.
- `D-026` — engine, not product — capping growth is an opt-in mechanism.
- **New:** `D-075` — fill cap (no over-grow) — filed in this PR.

## 7. Architecture

```text
distributeFillCapped(nodes, heights, slack) → used:
  for each flexible idx:
      share = slack·heights[idx]/flexH            // proportional, floored
      cap   = heights[idx]·fillGrowthMaxBP/10000  // ≤ 1× preferred added
      add   = min(share, cap); heights[idx] += add; used += add
  return used                                     // ≤ slack

alignedStackIn (VAlignFillCapped):
  slack = box.H − totalH
  used  = distributeFillCapped(...)
  residual = slack − used
  space = residual/(n+1); startY += space; effectiveGap += space   // balanced
```

`VAlignFill` (uncapped) keeps calling the unchanged `distributeFill`, so it is
byte-identical.

## 8. Files added or changed

```text
scene/align.go                       # CHANGED — VAlignFillCapped const + String()
scene/render.go                      # CHANGED — distributeFillCapped; VAlignFillCapped branch
scene/render_fill_test.go            # CHANGED — capped growth + residual spacing white-box tests
scene/render_parallel_test.go        # CHANGED — TestRenderDeterministic_VAlignFillCapped guard
scripts/smoke/phase-44.sh            # NEW — phase smoke
docs/research/27-fill-cap-no-overgrow.md # NEW — brief 27
docs/research/INDEX.md               # CHANGED — register brief 27
docs/plans/phase-44-fill-cap-no-overgrow.md # NEW — this plan
docs/plans/README.md                 # CHANGED — Wave 10 phase index row
docs/decisions.md                    # CHANGED — adds D-075
docs/glossary.md                     # CHANGED — VAlignFillCapped term
docs/site/guide/scene.md             # CHANGED — document VAlignFillCapped
skills/compose-a-scene/SKILL.md      # CHANGED — VAlignFillCapped in the alignment notes
```

## 9. Public API surface

```go
// scene
const VAlignFillCapped VAlign = … // like VAlignFill but each flexible node grows
                                  // by at most a pinned factor of its preferred
                                  // height; leftover slack becomes balanced
                                  // spacing. Byte-identical-when-no-slack.
func (v VAlign) String() string   // adds the "fill-capped" case
```

Additive enum value appended after `VAlignFit`; existing values unchanged.

## 10. Risks

- **R1 — uncapped regression.** **Mitigation:** `VAlignFill` still calls the
  unchanged `distributeFill`; only `VAlignFillCapped` is new. A test asserts a
  `VAlignFill` deck is byte-identical to before.
- **R2 — determinism.** **Mitigation:** integer / basis-point math; a guard
  renders a `VAlignFillCapped` deck at 1 and 8 workers.
- **R3 — residual mis-distribution overflows the box.** **Mitigation:** `used ≤
  slack` and `space = residual/(n+1)` floors, so the placed stack never exceeds
  `box.H`; a test asserts the last node bottom ≤ `box.Bottom()`.

## 11. Acceptance criteria

1. In a `VAlignFillCapped` stack with a 1-line (sparse) flexible node and a dense
   flexible node, the sparse node grows by no more than its cap
   (`≤ fillGrowthMaxBP × preferred`), and the leftover slack appears as even
   spacing (top margin + widened gaps), not as one ballooned node.
2. Uncapped `VAlignFill` is byte-identical to the current output.
3. Identical inputs yield identical EMU geometry (deterministic at any worker
   count), and the stack stays within the box.
4. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-44.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` capped growth bounds a sparse node + residual becomes spacing
   (`TestDistributeFillCapped_BoundsAndResidual`).
3. `OK:` `VAlignFillCapped` placement stays in the box + spaces evenly
   (`TestFillCapped_EvenSpacingWithinBox`).
4. `OK:` capped fill render stays deterministic
   (`TestRenderDeterministic_VAlignFillCapped`).

## 14. Tests

- **Unit:** `scene` (white-box) — `distributeFillCapped` caps growth and returns
  `used < slack` for a sparse+dense stack; the `alignedStackIn` capped branch
  distributes even spacing and stays in the box.
- **Round-trip golden:** n/a (scene layout change).
- **Integration / Fuzz / Bench:** none.

## 15. Vocabulary added

- `VAlignFillCapped` — the opt-in capped-fill body-stack mode that bounds each
  flexible node's growth and turns leftover slack into balanced spacing.

## 16. Plan deviations encountered during implementation

- **Pinned `growthMax`, no per-node `MaxGrow`** (per §5).

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-44.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-075).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
