# Brief 70 — Native vector micro-charts (R14.8)

> Informs Phase 87 (Wave 14). Engine req R14.8
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine; D-059).

## 1. Motivating phase

Pro decks are full of simple in-card data marks — a progress/capacity bar, a tiny
bar set, a sparkline, a single-value donut, a gauge — that should be crisp native
vector shapes, never a raster. None exist (all data viz routes through the raster
`Chart` node). Phase 87 adds a native `DataMark` node for the rect/line-based
family; the arc-based marks follow (Phase 88).

## 2. Subsystem / files

- `scene/nodes.go` — the node catalog.
- `scene/policy.go` / `scene/validate.go` / `scene/render.go` — new-node wiring.
- `pptx/shape.go` — `AddShape` options (a new `WithFlipV` for the sparkline).
- `scene/render_flow.go` — the `ShapeLine` / rect patterns to reuse.

## 3. Findings

- **Bar / Bars / Sparkline are pure rect + line geometry.** A progress bar = a
  track rounded-rect + a fill rounded-rect to the value fraction; a bar group = N
  rects; a sparkline = N-1 connected line segments + an end dot. All native,
  deterministic integer EMU, no rasterizer, no `AssetResolver` →
  `nodeUsesAssets:false`, policy `HasAsset:false`.
- **A diagonal line needs `flipV`.** A `ShapeLine` draws from the top-left to the
  bottom-right of its (positive-extent) box; an *upward* segment can't be drawn
  with a negative extent. The transform already has a `flipV` field but no builder
  API — add a small `pptx.WithFlipV(bool)` shape option (P1: a genuinely new need).
  An upward segment then draws as a top-anchored positive box + `WithFlipV` (BL→TR).
  Go marshals the bool as `flipV="true"` (valid `xsd:boolean`).
- **Arc-based marks (donut, gauge) need a builder arc seam.** A 92% donut is a
  `blockArc`/`pie` `prstGeom` with start/end-angle (+ inner-radius) adjust guides;
  the builder only plumbs the `roundRect` adjust today, and the icon translator
  forbids elliptical arcs (D-040). So the donut/gauge are a follow-up (Phase 88)
  that adds an adjust-guide builder seam — a clean §4.3 split (the bar family is
  the high-frequency case and ships now). The `DataMarkKind` enum appends `Donut`/
  `Gauge` then.
- **Colors are tokens (P2).** The mark color is `Color *ColorRole` (nil =
  `ColorAccent`, the D-054 pointer pattern); the bar track is `ColorSurfaceAlt`.
- **Embeds in a card/cell.** `preferredHeight` is a thin slot for a horizontal bar
  and a taller slot for bars/sparkline; the R11.3 clamp keeps it on-canvas.

## 4. Recommendations

- Node: `DataMark{Kind DataMarkKind; Value float64; Values []float64; Orientation
  FlowOrientation; Color *ColorRole; Label string}`; `DataMarkKind` = `Bar`,
  `Bars`, `Sparkline` (this phase).
- `render_datamark.go`: bar (track + fill, optional inline label), bars (N rects),
  sparkline (segments via `WithFlipV` for upward runs + an end dot).
- Builder: `pptx.WithFlipV(bool)`.
- Validate: `Bar` value in `[0,1]`; `Bars`/`Sparkline` ≥1 value, each in `[0,1]`.
- Full wiring: `KindDataMark` (appended last) + String + policy `{}` + validate +
  dispatch + `preferredHeight` + `nodeUsesAssets:false`; catalog 29 → 30;
  integration kind-loop `..KindDataMark` (on the existing "button" slide) +
  adversarial dataviz slide + smoke. Glossary, THEME (mark colors), compose-a-scene
  skill, docs/site visual-leaves. D-122.

## 5. Open questions

- Donut + gauge (arc seam) → Phase 88 (appends `DataMarkKind` values).
- A multi-series sparkline / axis labels → V1.x.
