# Phase 84 — timeline / roadmap node

**Subsystem:** `scene` (new IR node + renderer)
**RFC sections:** §11 (IR nodes), §12 (per-node policy), §10.1 (backward-compat), §7.1 (token colors)
**Deps:** brief 67.
**Status:** Done

---

## 1. Goal

Add a `Timeline` scene node — a roadmap axis with milestones at caller-specified
proportional positions, optional phase bands, and optional swimlanes — laid out
deterministically and fit to its region.

## 2. Why now

Wave 14 coverage classes (`docs/plans/README.md`); the timeline/roadmap is a core
"where we're going" slide that `Flow` (equal-step pipeline) can't express. Engine
req R14.4 (HIGH · engine per D-059).

## 3. RFC sections implemented

- `RFC §11` — a new IR leaf node (the catalog grows 28 → 29).
- `RFC §12` — native shapes (axis lines, marker dots/icons, labels), not media.
- `RFC §10.1` — additive; a deck with no Timeline is byte-identical (absent node).
- `RFC §7.1` — band fills + marker colors resolve from theme tokens (P2).

## 4. Brief findings incorporated

- `docs/research/67-timeline-roadmap.md` — *"genuinely a new node, not a Flow
  extension"* → full new-node wiring.
- `67` — *"markers/axis/labels compose from existing preset shapes"* → no new
  builder capability (P1); `nodeUsesAssets:false`, `HasAsset:false`.
- `67` — *"proportional positions keep it deterministic"* → `Position 0..1`; the
  caller maps dates → fractions (a date type is R14.13 territory, out of scope).
- `67` — *"anti-collision by stagger"* → labels alternate above/below, clamped.
- `67` — *"AccentIndex cycles a pinned token set"* → `[Accent, AccentAlt, Info,
  Success, Warning]`.

## 5. Findings I'm departing from

- **Vertical orientation** is deferred (§4.3): horizontal is the dominant roadmap;
  the lane/axis math transposes cleanly when demand warrants.
- **A `date` milestone type + tick labels** is deferred (ties to R14.13
  number/locale format); the caller maps dates to `0..1` positions today.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-026` — the engine places proportional positions; the caller/soul decides
  positions, phases, and accents.
- `D-119` (new) — files the `Timeline` node.

## 7. Architecture

`Timeline{Milestones []Milestone; Lanes []TimelineLane; Bands []TimelineBand}`;
`Milestone{Position float64; Label, Detail, Icon string; AccentIndex int}`;
`TimelineLane{Label string; Milestones []Milestone}`; `TimelineBand{From, To
float64; Label string; Fill ColorRole}`. `renderTimeline` reserves a band-label
strip (top) and a lane-label gutter (left), draws bands (low-alpha rects +
labels), then one lane per row: an axis line + per-milestone marker (an accent dot
or a curated icon) + a staggered label (above even, below odd). Empty `Lanes` ⇒
one implicit lane from `Milestones`. `timelinePreferredHeight` = band-label strip
+ pinned per-lane height × lane count. Pinned geometry consts; colors are tokens.

```text
Timeline{Bands:[{0,.5,Now},{.5,1,Next}], Lanes:[{Platform,[…]},{GTM,[…]}]}
  → band rects + labels; lane 1 axis + markers/labels; lane 2 axis + markers/labels
Timeline{Milestones:[…]}  (no Lanes) → one implicit lane
```

## 8. Files added or changed

```text
scene/nodes.go                          # CHANGED — KindTimeline + Timeline/Milestone/TimelineLane/TimelineBand
scene/policy.go                         # CHANGED — KindTimeline policy {} (native, no asset)
scene/validate.go                       # CHANGED — Timeline validation (>=1 milestone, positions, band spans)
scene/render.go                         # CHANGED — renderNode dispatch + preferredHeight + nodeUsesAssets
scene/render_card.go                    # CHANGED — walkIconRefs visits milestone icons
scene/render_timeline.go                # NEW — the composer + timelineAccent / timelinePreferredHeight
scene/render_timeline_test.go           # NEW — white-box: accent cycle, preferred height
scene/render_timeline_render_test.go    # NEW — black-box: roadmap, single-lane, invalid, determinism
scene/scene_test.go                     # CHANGED — allNodes + catalog count 28 → 29
scene/render_adversarial_test.go        # CHANGED — a dense roadmap slide in the torture fixture
test/integration/roundtrip_test.go      # CHANGED — Timeline on the existing "button" slide + kind-loop bound
scripts/smoke/phase-84.sh               # NEW — phase smoke
docs/research/67-timeline-roadmap.md    # NEW — brief
docs/research/INDEX.md                  # CHANGED — registers brief 67
docs/plans/phase-84-timeline.md         # NEW — this plan
docs/plans/README.md                    # CHANGED — Phase 84 detail
docs/design/THEME.md                    # CHANGED — timeline color mechanism note
docs/glossary.md                        # CHANGED — Timeline / roadmap term
docs/decisions.md                       # CHANGED — adds D-119
docs/site/reference/scene.md            # CHANGED — KindTimeline in the catalog
skills/compose-a-scene/SKILL.md         # CHANGED — Timeline node row
```

## 9. Public API surface

```go
// scene
type Timeline struct { Milestones []Milestone; Lanes []TimelineLane; Bands []TimelineBand }
type Milestone struct { Position float64; Label, Detail, Icon string; AccentIndex int }
type TimelineLane struct { Label string; Milestones []Milestone }
type TimelineBand struct { From, To float64; Label string; Fill ColorRole }
```

Additive new node; no break.

## 10. Risks

- **R1 — off-canvas labels.** **Mitigation:** labels clamp to the axis box and
  stagger; the adversarial roadmap slide asserts on-canvas under hostile content.
- **R2 — determinism.** **Mitigation:** integer-EMU proportional layout; a
  1-vs-8-worker test asserts byte-identity.
- **R3 — new-node wiring gaps.** **Mitigation:** the integration kind-loop bound
  is bumped to `KindTimeline` and the catalog count to 29 (both fail loudly if a
  touchpoint is missed).

## 11. Acceptance criteria

1. A 6-milestone / 3-band / 2-lane roadmap renders within the safe area (axis
   lines + markers + band fills + labels), conformant, no warnings.
2. The implicit single-lane path renders an axis with markers.
3. An out-of-range milestone position fails Stage-1 validation.
4. The timeline is worker-count deterministic.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | timeline composer + helpers |

## 13. Smoke check

`scripts/smoke/phase-84.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `KindTimeline` / `Timeline` / `Milestone` / policy / validation /
   dispatch / composer / walkIconRefs / catalog-29 present.
3. `OK:` roadmap / single-lane / invalid / accent-cycle / preferred-height /
   determinism / integration tests.

## 14. Tests

- **White-box (`scene`):** the accent cycle (wrap + clamp) and preferred-height
  (per-lane + band-label strip).
- **Black-box (`scene_test`):** a roadmap renders (axis lines, marker dots, band +
  lane labels, conformant), a single-lane timeline, an invalid position fails
  validation, and the timeline is worker-count deterministic.
- **Adversarial:** a dense roadmap with long labels in the torture fixture.
- **Integration:** `Timeline` added to the all-kinds fixture (kind-loop bound
  bumped to `KindTimeline`).
