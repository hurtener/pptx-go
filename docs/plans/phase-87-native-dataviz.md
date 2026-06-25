# Phase 87 — native vector micro-charts (DataMark: bar / bars / sparkline)

**Subsystem:** `scene` (new IR node) + `pptx` (a `WithFlipV` shape option)
**RFC sections:** §11 (IR node), §12 (per-node policy), §10.1 (backward-compat), §7.1 (token colors)
**Deps:** brief 70.
**Status:** Done

---

## 1. Goal

Add a native `DataMark` node — crisp, brand-colored vector micro-charts (a
progress bar, a bar group, a sparkline) drawn from preset shapes, deterministic
and resolution-independent, with no asset pipeline.

## 2. Why now

Wave 14 coverage classes (`docs/plans/README.md`); pricing/KPI decks need simple
in-card data marks that today would raster a trivial bar. Engine req R14.8 (HIGH ·
engine per D-059). The bar family ships now; arc-based marks (donut/gauge) follow.

## 3. RFC sections implemented

- `RFC §11` — a new IR leaf node (catalog 29 → 30).
- `RFC §12` — native shapes (rects + lines), not media.
- `RFC §10.1` — additive; a deck with no DataMark is byte-identical (absent node).
- `RFC §7.1` — mark + track colors are theme tokens (P2).

## 4. Brief findings incorporated

- `docs/research/70-native-dataviz.md` — *"Bar/Bars/Sparkline are pure rect+line
  geometry"* → native, `nodeUsesAssets:false`, `HasAsset:false`.
- `70` — *"a diagonal line needs flipV"* → a new `pptx.WithFlipV` shape option.
- `70` — *"arc-based marks need a builder arc seam"* → donut/gauge → Phase 88.
- `70` — *"colors are tokens"* → `Color *ColorRole` (nil = accent), track SurfaceAlt.

## 5. Findings I'm departing from

- **Donut + gauge** are deferred to Phase 88 (a `blockArc`/`pie` adjust-guide
  builder seam — the icon translator forbids elliptical arcs, D-040, and the
  builder only plumbs the `roundRect` adjust today). The bar family is the
  high-frequency case; the `DataMarkKind` enum appends `Donut`/`Gauge` then. §4.3.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-026` — the engine draws the mark; the caller supplies values + color.
- `D-040` — the icon-translator arc constraint (why arcs are a separate seam).
- `D-122` (new) — files `DataMark` (bar family) + `WithFlipV`.

## 7. Architecture

`DataMark{Kind DataMarkKind; Value float64; Values []float64; Orientation
FlowOrientation; Color *ColorRole; Label string}`; `DataMarkKind` = `DataMarkBar`,
`DataMarkBars`, `DataMarkSparkline`. `render_datamark.go`: a bar (track rounded-rect
+ accent fill to `Value`, optional inline label; vertical fills from the bottom), a
bar group (N rects, each `Values[i]`-tall from the bottom), and a sparkline (N-1
connected `ShapeLine` segments — upward segments use `pptx.WithFlipV` on a
positive-extent box — plus an accent end dot). `dataMarkPreferredHeight` is a thin
slot for a horizontal bar, taller for bars/sparkline. Pinned geometry consts;
colors are tokens.

```text
DataMark{Bar, Value:0.92, Label:"92%"} → track + accent fill to 92% + "92%"
DataMark{Bars, Values:[…]}             → N accent rects
DataMark{Sparkline, Values:[…]}        → polyline (flipV on up-segments) + end dot
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — KindDataMark + DataMark + DataMarkKind + markColor()
scene/policy.go                      # CHANGED — KindDataMark policy {} (native)
scene/validate.go                    # CHANGED — DataMark validation (kind + values in [0,1])
scene/render.go                      # CHANGED — dispatch + preferredHeight + nodeUsesAssets
scene/render_datamark.go             # NEW — the composer
scene/render_datamark_test.go        # NEW — bar / bars+sparkline(flipV) / in-card / invalid / determinism
scene/scene_test.go                  # CHANGED — allNodes + catalog count 29 → 30
scene/render_adversarial_test.go     # CHANGED — a dataviz card in the torture fixture
pptx/shape.go                        # CHANGED — WithFlipV shape option
test/integration/roundtrip_test.go   # CHANGED — DataMarks on the existing "button" slide + kind-loop bound
scripts/smoke/phase-87.sh            # NEW — phase smoke
docs/research/70-native-dataviz.md   # NEW — brief
docs/research/INDEX.md               # CHANGED — registers brief 70
docs/plans/phase-87-native-dataviz.md  # NEW — this plan
docs/plans/README.md                 # CHANGED — Phase 87 detail
docs/design/THEME.md                 # CHANGED — DataMark color mechanism note
docs/glossary.md                     # CHANGED — DataMark term
docs/decisions.md                    # CHANGED — adds D-122
docs/site/catalog/visual-leaves.md   # CHANGED — DataMark section
skills/compose-a-scene/SKILL.md      # CHANGED — DataMark node row
```

## 9. Public API surface

```go
// scene
type DataMark struct { Kind DataMarkKind; Value float64; Values []float64; Orientation FlowOrientation; Color *ColorRole; Label string }
type DataMarkKind int // DataMarkBar | DataMarkBars | DataMarkSparkline
// pptx
func WithFlipV(flip bool) ShapeOption
```

Additive new node + a new shape option; no break.

## 10. Risks

- **R1 — off-canvas.** **Mitigation:** the R11.3 clamp + the adversarial dataviz
  card assert on-canvas.
- **R2 — determinism.** **Mitigation:** integer-EMU geometry; a 1-vs-8-worker
  test asserts byte-identity.
- **R3 — diagonal line direction.** **Mitigation:** `WithFlipV` for upward
  segments; a test asserts an upward segment emits `flipV`.

## 11. Acceptance criteria

1. A progress bar fills to its value in soul colors as native rounded rects;
   conformant, no warnings.
2. A bar group + a sparkline (with an upward segment) render natively
   (line segments + flipV), conformant.
3. A DataMark embeds in a card body without overflow.
4. An out-of-range value fails Stage-1 validation; the marks are worker-count
   deterministic.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | DataMark composer |
| `pptx` | 85% | the WithFlipV option |

## 13. Smoke check

`scripts/smoke/phase-87.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `KindDataMark` / `DataMark` / `DataMarkKind` / policy / validation /
   composer / `WithFlipV` / catalog-30 present.
3. `OK:` bar / bars+sparkline / in-card / invalid / determinism / integration tests.

## 14. Tests

- **Black-box (`scene_test`):** a bar renders native rects + label (conformant); a
  bar group + sparkline render (line segments + `flipV`, conformant); a DataMark
  embeds in a card without overflow; an invalid value fails validation; the marks
  are worker-count deterministic.
- **Adversarial:** a dataviz card (bar + bars + sparkline).
- **Integration:** DataMarks added to the all-kinds fixture (kind-loop bound bumped
  to `KindDataMark`).
