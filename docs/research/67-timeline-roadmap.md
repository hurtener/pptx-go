# Brief 67 — Timeline / roadmap node (R14.4)

> Informs Phase 84 (Wave 14). Engine req R14.4
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine; D-059).

## 1. Motivating phase

Strategy and investor decks routinely need a "where we're going" roadmap — a time
axis with non-uniform dated milestones, phase/horizon bands, and swimlanes. `Flow`
renders an *equal-step* linear pipeline and cannot express any of that. Phase 84
adds a new `Timeline` IR node.

## 2. Subsystem / files

- `scene/nodes.go` — the node catalog + `NodeKind`.
- `scene/policy.go`, `scene/validate.go`, `scene/render.go` — the new-node wiring.
- `scene/render_flow.go` — axis lines + markers + icon + label patterns to reuse.
- `scene/render_card.go` — `walkIconRefs` (milestone icons).
- `test/integration/roundtrip_test.go` — the all-kinds fixture + kind-loop bound.

## 3. Findings

- **It is a genuinely new node** (not a `Flow` extension): a continuous axis with
  proportional positions, bands, and lanes is a different layout from Flow's
  equal-step pipeline. The catalog grows 28 → 29; the full new-node wiring applies.
- **Markers / axis / labels compose from existing preset shapes** (`ShapeLine`,
  `ShapeEllipse`, `AddIcon`, text frames) — no new builder capability (P1) and no
  media, so the node is `nodeUsesAssets:false` (parallel-safe) and `HasAsset:false`.
- **Proportional positions keep it deterministic.** Each milestone's `Position` is
  a `0..1` fraction of the axis width; `cx = lane.X + Position*lane.W` in integer
  EMU. Non-uniform dates map to fractions at author time (the soul/caller does the
  date→fraction math; the engine places the fraction — D-026). A date type in the
  IR would push date arithmetic + formatting into the engine (R14.13 territory);
  out of scope.
- **Anti-collision by stagger.** Labels alternate above (even index) / below (odd)
  the axis, each clamped to stay within the axis box. This halves dense-label
  overlap deterministically without a packing solver — sufficient for the
  acceptance ("no label overlaps" on a 6-milestone roadmap).
- **Lanes vs single axis.** Empty `Lanes` ⇒ one implicit lane from the top-level
  `Milestones`; non-empty `Lanes` ⇒ swimlane rows (a left-gutter label + its own
  axis). Bands span the full width behind every lane.
- **AccentIndex cycles a pinned token set.** `[ColorAccent, ColorAccentAlt,
  ColorInfo, ColorSuccess, ColorWarning]` — all existing roles (P2); the soul
  drives which index a phase uses. (This anticipates the Wave-15 multi-accent
  palette without depending on it.)
- **Orientation.** Horizontal is the dominant roadmap; vertical is deferred
  (§4.3) to bound scope — the layout primitives transpose cleanly later.

## 4. Recommendations

- Node: `Timeline{Milestones []Milestone; Lanes []TimelineLane; Bands
  []TimelineBand}`; `Milestone{Position, Label, Detail, Icon, AccentIndex}`;
  `TimelineLane{Label, Milestones}`; `TimelineBand{From, To, Label, Fill}`.
- Composer `render_timeline.go`: band rects (low-alpha) + labels; per-lane axis
  line + markers (dot or icon) + staggered labels. `timelinePreferredHeight` =
  band-label strip + pinned per-lane height. Pinned geometry consts; colors are
  tokens.
- Validate: ≥1 milestone; every `Position` in `[0,1]`; every band `0<=From<=To<=1`.
- Full wiring: `KindTimeline` (appended last) + String + policy `{}` + validate +
  `renderNode` dispatch + `preferredHeight` + `nodeUsesAssets:false` + walkIconRefs
  + catalog 29 + integration kind-loop `..KindTimeline` (add to the existing
  "button" slide) + adversarial roadmap slide + smoke. THEME.md, glossary,
  compose-a-scene skill, docs/site. D-119.

## 5. Open questions

- Vertical orientation → a follow-up (transpose the lane/axis math).
- A `date` milestone type + tick labels → ties to R14.13 number/locale format;
  deferred (the caller maps dates to `0..1` today).
- Dense-label packing beyond stagger (a true collision solver) → V1.x if needed.
