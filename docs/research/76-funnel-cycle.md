# Brief 76 — Funnel + cycle non-linear process diagrams (R14.11)

> Informs Phase 93 (Wave 14). Engine req R14.11
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059).

## 1. Motivating phase
`Flow` is a linear pipeline; it can't express a funnel (tapering stages), a closed
cycle (ring of stages with arrows), or a branch (1→M). Marketing-funnel,
lifecycle-loop, and decision slides are common and uncovered.

## 2. Findings
- **Funnel = a stepped stack of rects.** A vertical stack of centered rounded-rect
  bands of linearly decreasing width (widest at top, floored at `funnelMinFrac`),
  each with a centered contrast label + optional value. No custom geometry — clean
  + deterministic. New `Funnel{Stages []FunnelStage{Label, Value, AccentIndex}}`.
- **Cycle = stage cards on a ring + connectors.** N stage cards placed clockwise
  from the top at evenly-spaced angles (trig → `math.Round` to integer EMU, then
  deterministic); consecutive stages (incl. last→first) connected by a straight
  line (flip-aware for upward runs, via `WithFlipV`) + a chevron arrowhead rotated
  to the chord (`math.Atan2`). New `Cycle{Stages []CycleStage{Label, Icon,
  AccentIndex}}`.
- **Branch = the Tree node.** A "step splitting into M paths" is a 2-level tree;
  the R14.10 `Tree` node covers it by composition, so no separate Flow.Branch is
  added (documented in D-128).
- Both native (`nodeUsesAssets:false`); cycle stage icons recurse in walkIconRefs.

## 3. Recommendations
- Two new nodes + composers + full wiring. Tests: a 4-stage funnel (≥4 bands +
  values, conformant), a 5-stage cycle (≥5 cards + chevrons, conformant), empty
  fails, determinism (incl. trig); an adversarial funnel+cycle. Glossary, THEME,
  visual-leaves, skill. D-128.

## 4. Open questions
- A true trapezoid funnel (custGeom) + curved ring arrows → V1.x; the stepped
  funnel + straight ring connectors satisfy the acceptance.
