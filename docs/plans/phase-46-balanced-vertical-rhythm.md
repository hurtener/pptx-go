# Phase 46 — balanced vertical rhythm

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §10 (layout engine)
**Deps:** Phase 13 (alignment), Phase 44 (VAlignFillCapped even-rhythm primitive, D-075)
**Status:** Done

---

## 1. Goal

Add an opt-in `VAlignBalanced` body-stack mode that distributes a sparse stack's
slack as an even rhythm — a top margin and widened inter-node gaps — optically
centered, so a sparse cover or closing reads balanced instead of clustered with a
large void.

## 2. Why now

R10.8 is the next Wave-10 unit. It fixes the recreation's sparse cover (elements
clustered in the middle, description dropped far below) and closing (empty lower
frame) — DECKARD R10.8 gap. It reuses the even-`(n+1)`-unit residual primitive
introduced in R10.6 (D-075).

## 3. RFC sections implemented

- `RFC §10` — a new vertical-distribution mode for sparse stacks; the existing
  modes are unchanged.

## 4. Brief findings incorporated

- `docs/research/29-balanced-vertical-rhythm.md` — *the even-rhythm primitive
  already exists in `VAlignFillCapped`'s residual (`unit = slack/(n+1)`)* →
  `VAlignBalanced` reuses it: top margin + widened gaps.
- `docs/research/29-balanced-vertical-rhythm.md` — *distinct from Justify (no
  margins) and Center (fixed gaps)* → `VAlignBalanced` distributes into both
  margins and gaps.
- `docs/research/29-balanced-vertical-rhythm.md` — *optical center = bias up by
  shrinking the top margin (`× 0.85`)* → `startY = box.Y + unit ×
  balancedOpticalBP/10000`.
- `docs/research/29-balanced-vertical-rhythm.md` — *group-aware weighting is the
  caller's (D-026)* → the engine ships even rhythm + optical center only.

## 5. Findings I'm departing from

- The spec mentions distributing slack into gaps *weighted by a pinned ratio*
  (larger gap before a description block). This plan ships the **optical-center
  bias** as the engine's pinned ratio and leaves per-node gap weighting to the
  caller. **Departing because** knowing which node is a "description block" is
  content taste (D-026); a caller orders nodes or inserts a spacer. (§4.3.)

## 6. Decisions referenced

- `D-075` — `VAlignFillCapped` — the `(n+1)`-even-unit residual primitive reused.
- `D-026` — engine, not product — even rhythm is a mechanism; node grouping is the
  caller's.
- **New:** `D-077` — balanced vertical rhythm — filed in this PR.

## 7. Architecture

```text
alignedStackIn (VAlignBalanced):
  slack = box.H − totalH
  if slack > 0:
      unit = slack/(n+1)
      startY = box.Y + unit × balancedOpticalBP/10000   // optical top margin (< even)
      effectiveGap = gap + unit                          // widened internal gaps
```

`balancedOpticalBP = 8500` (top margin = 85 % of an even unit, so the stack sits a
touch above geometric center). Integer / basis-point math. `VAlignTop`/`Center`/
`Justify` are untouched.

## 8. Files added or changed

```text
scene/align.go                       # CHANGED — VAlignBalanced const + String()
scene/render.go                      # CHANGED — VAlignBalanced branch; balancedOpticalBP
scene/render_balanced_test.go        # NEW — even rhythm + optical bias + Top/Center byte-identical
scene/render_parallel_test.go        # CHANGED — TestRenderDeterministic_VAlignBalanced guard
scripts/smoke/phase-46.sh            # NEW — phase smoke
docs/research/29-balanced-vertical-rhythm.md # NEW — brief 29
docs/research/INDEX.md               # CHANGED — register brief 29
docs/plans/phase-46-balanced-vertical-rhythm.md # NEW — this plan
docs/plans/README.md                 # CHANGED — Wave 10 phase index row
docs/decisions.md                    # CHANGED — adds D-077
docs/glossary.md                     # CHANGED — VAlignBalanced term
docs/site/guide/scene.md             # CHANGED — document VAlignBalanced
skills/compose-a-scene/SKILL.md      # CHANGED — VAlignBalanced in the alignment notes
```

## 9. Public API surface

```go
// scene
const VAlignBalanced VAlign = … // distribute a sparse stack's slack as an even
                                // rhythm (top margin + widened gaps), optically
                                // centered. Byte-identical-when-no-slack.
func (v VAlign) String() string // adds the "balanced" case
```

Additive enum value appended after `VAlignFillCapped`; existing values unchanged.

## 10. Risks

- **R1 — regression of Top/Center/Justify.** **Mitigation:** `VAlignBalanced` is
  the only new branch; a test asserts `VAlignTop`/`VAlignCenter` placements are
  unchanged.
- **R2 — determinism.** **Mitigation:** integer / basis-point math; a guard renders
  a `VAlignBalanced` deck at 1 and 8 workers.
- **R3 — overflow.** **Mitigation:** only positive slack is distributed and
  `unit = slack/(n+1)` floors, so the placed stack stays within the box; a test
  asserts the last node bottom ≤ `box.Bottom()`.

## 11. Acceptance criteria

1. A 3-node sparse stack under `VAlignBalanced` has a non-zero top margin and
   widened inter-node gaps (the slack distributed as even rhythm, no single large
   residual void), and the stack stays within the box.
2. The optical bias places the top margin below an even share (stack sits above
   geometric center).
3. `VAlignTop` and `VAlignCenter` remain byte-identical.
4. Identical inputs yield identical EMU geometry (deterministic at any worker
   count).
5. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-46.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` balanced distributes even rhythm within the box
   (`TestBalanced_EvenRhythmWithinBox`).
3. `OK:` Top/Center byte-identical (`TestBalanced_TopCenterByteIdentical`).
4. `OK:` balanced render stays deterministic
   (`TestRenderDeterministic_VAlignBalanced`).

## 14. Tests

- **Unit:** `scene` (white-box) — balanced top margin + widened gaps sum the
  slack; optical bias; last node within box; Top/Center placements unchanged.
- **Round-trip golden / Integration / Fuzz / Bench:** none.

## 15. Vocabulary added

- `VAlignBalanced` — the opt-in body-stack mode that distributes a sparse stack's
  slack as an optically-centered even rhythm.

## 16. Plan deviations encountered during implementation

- **Per-node gap weighting left to the caller** (per §5); the engine ships even
  rhythm + optical center.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-46.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-077).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
