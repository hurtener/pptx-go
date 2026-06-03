# Phase 17 — Chart (image-shape V1)

**Subsystem:** scene (chart) + pptx (chart placeholder helper)
**RFC sections:** §15.1, §11.1 (chart), §12 (D-004)
**Deps:** Phase 11 (image/pic path), Phase 16 (code_block raster precedent)
**Status:** Draft

---

## 1. Goal

Render the `chart` scene node as a caller-rasterized image that fits its slot
preserving aspect, with an optional caption and an aspect-mismatch warning; add
a `pptx.ChartPlaceholder` builder helper that draws a labeled chart slot.

## 2. Why now

Phase 17 opens Wave 5 (charts) in the master plan. V1 charts are image-shape
(D-004): the caller renders the chart and pptx-go embeds the bytes — the same
raster discipline as code_block (Phase 16), so the path is well-trodden. It adds
the first chart surface ahead of V2's native `c:chart`.

## 3. RFC sections implemented

- `RFC §15.1` — V1 image-shape chart: `pic` from caller bytes, sized from the
  slot, caption below, **aspect-divergence `LayoutWarning`**.
- `RFC §11.1` — the `chart` leaf node.
- `RFC §12` (D-004) — per-node policy "`chart` → Image (`pic`), `asset_id` (V1)".

## 4. Brief findings incorporated

- `docs/research/07-chart-image-shape.md`:
  - **F1 (chart ≈ code_block raster)** → resolve → `pic` → caption, reusing the
    Phase-16 path shape (without the language badge).
  - **F2 / Q1 (aspect source)** → read header dimensions via stdlib
    `image.DecodeConfig` (no pixel decode); §7 forbids pixel *data*, not the
    dimension *header* — recorded in D-046.
  - **F3 (read dims)** → automatic aspect from the caller's bytes; degrade
    silently (no warning) if dimensions can't be read.
  - **F4 (fit-within + thresholded warning)** → place the chart to **contain**
    within the slot preserving aspect (centered, letterboxed); fire one
    `LayoutWarning` when divergence exceeds the threshold (15%); round the
    reported percent for deterministic text (D-035).
  - **F5 / Q2 (ChartPlaceholder)** → a builder helper drawing a labeled bordered
    slot, returning `*Shape`; the scene composer reuses it for an unresolved
    asset (a labeled slot instead of a blank gap).

## 5. Findings I'm departing from

None. (Brief Q4 — *shipping* aspect-aware fit for the Image node — stays out of
scope; Phase 17 fixes the Image `Fit` comment to remove the inaccurate "§7
forbids" claim but does not add an Image fit mode.)

## 6. Decisions referenced

- `D-004` — charts are image-shape in V1 (native `c:chart` is V2).
- `D-014` / `D-045` — the code_block raster + caption path this mirrors.
- `D-036` — asset-resolution failure degrades to a warning (here: render the
  placeholder + warn, not an error).
- `D-026` — engine, not product — fit/warn are mechanisms; pptx-go renders no
  chart and makes no chart-content opinion.
- **`D-046` (new, this PR)** — reading image dimension headers
  (`image.DecodeConfig`) is permitted (decoding pixel *data* is not, §7);
  chart contains-to-fit with a thresholded aspect warning; `ChartPlaceholder`
  draws a labeled slot.

## 7. Architecture

```text
scene/render_chart.go   # NEW — renderChart: resolve → contain-fit pic → caption
                        #       + aspect warning; ChartPlaceholder for unresolved
pptx/chart.go           # NEW — (*Slide).ChartPlaceholder(box, opts) *Shape
pptx/media.go           # CHANGED — fix the Fit comment (§7 wording: pixel DATA)
scene/render.go         # CHANGED — dispatch Chart (Chart already asset-bearing)
```

`renderChart`: resolve `AssetID`; on success read dims (`image.DecodeConfig`),
compute the contained box (largest slot-fitting box at the image's aspect,
centered), `AddImage` there, and warn if |slotAR−imgAR|/imgAR > 0.15; on failure
draw `ChartPlaceholder` + warn (D-036). Caption renders below the slot.

`ChartPlaceholder` (builder): a `roundRect` with a dashed/tinted border and a
centered "Chart" label, theme-tokened; returns the `*Shape`.

## 8. Files added or changed

```text
pptx/chart.go              # NEW — ChartPlaceholder
pptx/chart_test.go         # NEW — round-trip golden + unit
pptx/media.go              # CHANGED — Fit comment fix (no behavior change)
scene/render_chart.go      # NEW — renderChart + aspect read/warn
scene/render.go            # CHANGED — dispatch Chart
scene/render_chart_test.go # NEW — pic, caption, aspect warning, placeholder, parallel
scripts/smoke/phase-17.sh  # NEW
docs/decisions.md          # CHANGED — D-046
docs/glossary.md           # CHANGED — chart placeholder, aspect warning
docs/plans/phase-17-chart.md            # NEW (this file)
docs/research/07-chart-image-shape.md   # NEW (committed with plan)
```

## 9. Public API surface

```go
// pptx
// ChartPlaceholder draws a labeled chart slot (a bordered roundRect + "Chart"
// label) at box without committing chart bytes, and returns the shape. The
// scene chart composer reuses it when an asset is unresolved (RFC §15.1).
func (s *Slide) ChartPlaceholder(box Box, opts ...ShapeOption) *Shape
```

No scene IR change (the existing `Chart{AssetID, Caption}` is rendered). No new
theme token (placeholder reuses surface/text/line tokens).

## 10. Risks

- **R1 — Reading dims is mistaken for §7 violation.** **Mitigation:** D-046
  states the boundary explicitly (header read ≠ pixel-data parse); only
  `image.DecodeConfig` is used (stdlib, no pixel decode); drift-audit's P4 check
  stays green (stdlib import).
- **R2 — Non-deterministic warning text** (float percent). **Mitigation:** round
  to an integer percent; a parallel workers=1-vs-N test guards byte-identity.
- **R3 — DecodeConfig fails on an exotic-but-valid image.** **Mitigation:** on
  any decode-config error, skip the aspect warning (render the pic fit to the
  full slot); never error.
- **R4 — New builder API without round-trip coverage.** **Mitigation:**
  `pptx/chart_test.go` ships a round-trip golden (G6) in this PR.

## 11. Acceptance criteria

1. A `chart` node with a PNG raster renders a `pic` contained within its slot
   (aspect preserved) plus a caption when present; the deck conforms.
2. An aspect-ratio mismatch (image AR vs slot AR beyond the threshold) surfaces
   a `LayoutWarning`; a near-match does not.
3. An unresolved chart asset renders a `ChartPlaceholder` (labeled slot) and a
   warning, not an error (D-036).
4. `pptx.ChartPlaceholder` emits a labeled bordered slot that round-trips
   through `pptx.Open` (G6).
5. `scene.Render` is byte-identical for a chart scene at `workers=1` and
   `workers=N` (deterministic dims + warning text).
6. `make coverage` shows touched packages ≥ their bands.
7. `scripts/smoke/phase-17.sh` reports `OK ≥ count(criteria)`, `FAIL = 0`;
   prior smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | default; ChartPlaceholder adds to the existing package |
| `scene` | 80% | default; chart composer adds to the existing package |

## 13. Smoke check

`scripts/smoke/phase-17.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` a chart with a raster renders a contained pic + caption.
3. `OK:` an aspect mismatch warns; a near-match does not.
4. `OK:` an unresolved chart renders a ChartPlaceholder + warning.
5. `OK:` ChartPlaceholder round-trips.
6. `OK:` chart render is byte-identical workers=1 vs N.

## 14. Tests

- **Unit:** `pptx` (ChartPlaceholder shape), `scene` (contain-fit math, aspect
  warning threshold both sides, unresolved → placeholder, parallel equivalence).
- **Round-trip golden:** yes — `ChartPlaceholder` (new public builder API, G6).
- **Integration:** extend the Phase-06 integration deck with a `chart` node
  (real opc/xml) asserting the pic reaches the slide.
- **Fuzz / Benchmark:** none.

## 15. Vocabulary added

- `chart placeholder` — a labeled bordered chart slot drawn by
  `pptx.ChartPlaceholder` (and by the chart composer for an unresolved asset).
- `aspect warning` — the `LayoutWarning` raised when a chart image's aspect
  ratio diverges from its slot beyond the threshold.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-17.sh` reports `OK ≥ count(criteria)` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (D-046).
- [ ] Round-trip golden for the new builder API lands in this PR.
- [ ] (Phase 20+) Docs site / skills updated. (inert)
