# Brief 07 — chart image-shape & aspect-ratio detection

**Subsystem:** scene (chart) + pptx (chart placeholder helper)
**Authored:** 2026-06-03
**Motivating phase:** Phase 17 — Chart (image-shape V1) (RFC §15.1, §11.1, D-004)

## 1. Question

How should pptx-go render the `chart` scene node in V1 — a caller-rasterized
chart image with an optional caption — and satisfy RFC §15.1's requirement to
**warn when the chart's aspect ratio diverges from its assigned slot**? What
does the `pptx.ChartPlaceholder(box)` builder helper do? And how is the
aspect-ratio requirement reconciled with §7's "do not parse pixel data" rule and
the Phase-11 Image node's deferral of aspect-aware fit?

## 2. Prior art surveyed

- **The code_block raster path** (`scene/render_code_block.go`, D-014/D-045):
  the exact shape a chart needs — resolve `AssetID` → `pic` via
  `AddImage(ImageBytes(data, ct), box)`, caption text below, asset-failure
  degrades to a warning (D-036). A chart is a code_block without the language
  badge, plus the aspect warning.
- **The Image node** (`scene/render_image.go`, `pptx/media.go`, D-039): ships
  `FitFill` / `FitNone`; its `Fit` comment says "aspect-aware cover/contain
  would require reading **pixel dimensions** (forbidden by §7)" and defers them.
- **§7 (security)**: "the library does not parse **pixel data**; it embeds bytes
  verbatim … We do verify content-type matches declared MIME and reject
  obviously malformed bytes (e.g. PNG signature missing)." The rule names *pixel
  data*, not the dimension header.
- **`image.DecodeConfig`** (Go stdlib `image`, with `image/png`,`image/jpeg`,
  `image/gif` registered): returns an `image.Config{Width, Height, ColorModel}`
  by reading only the format header (PNG `IHDR`, JPEG `SOFn`, GIF screen
  descriptor) — it does **not** decode pixels. CGo-free, stdlib-only (P4-safe).
- **RFC §15.1**: chart → `pic` from caller bytes; size from the layout slot;
  caption below; "Surfaces a `LayoutWarning` if the caller-provided aspect ratio
  diverges significantly from the assigned slot (the chart fits within the slot;
  the warning lets the caller know)."
- **pengui-slides v4**: already renders charts to PNG and emits an image shape —
  the V1 behavior mirrors this.

## 3. Findings

- **F1 — A chart is the code_block raster path minus the badge, plus a warning.**
  `pic` + caption is identical; the new surface is (a) the aspect-ratio warning
  and (b) the `ChartPlaceholder` builder helper. A shared internal helper for
  "resolve → pic → caption" could back both code_block and chart, but the two
  composers are tiny; duplication or a small shared helper are both fine.
- **F2 — §7 forbids parsing *pixel data*, not the *dimension header*.** Reading
  `Width`/`Height` via `image.DecodeConfig` reads the header only (no pixel
  decode), so it is consistent with §7 as written. The Phase-11 Image comment
  conflated "pixel dimensions" (in the header) with "pixel data" (the pixels).
  RFC §15.1 *requires* aspect detection, and the RFC outranks a code comment, so
  V1 must either read the dimensions or take them from the caller. This needs an
  explicit decision (a D-NNN) so the §7 boundary is stated, not implied.
- **F3 — Two ways to get the chart aspect:**
  - **(A) Read header dims** with `image.DecodeConfig`. Automatic; "caller-
    provided aspect ratio" = the aspect of the caller's bytes, read by the
    renderer. Reconciles §7 (header ≠ pixel data) and would later *unblock*
    aspect-aware image fit (a side benefit, not in scope here). Cost: the
    renderer inspects bytes (header only); a malformed/odd header yields no
    dimensions → no warning (degrade gracefully, never error).
  - **(B) Caller supplies aspect** via an IR field (e.g. `Chart.AspectW/AspectH`
    or `Chart.Aspect float64`). No byte inspection; pushes the work to the
    caller; more IR surface; the warning is only as good as the caller's hint.
- **F4 — The warning is fit-aware, not fatal (D-026, RFC §15.1).** The chart is
  placed to **fit within** the slot preserving aspect (letterboxed), and a
  `LayoutWarning` fires only when the divergence crosses a threshold (e.g.
  |slotAR − imgAR| / imgAR > ~15%). The renderer does not crop or distort to
  force-fit; it informs. A threshold avoids noise on near-matches.
- **F5 — `ChartPlaceholder(box)` is a builder slot helper.** "Sizes and
  positions a chart slot without committing bytes" → it draws a visible
  placeholder (a bordered/à-tint rounded rect with a centered "Chart" label) at
  `box`, so a deck authored before the chart raster exists shows a labeled slot.
  It returns the `*Shape` (or the slot `Box`) for the caller to position around.
  It is a builder primitive (`pptx`), reused by the scene chart composer when an
  asset is unresolved (instead of a blank gap).
- **F6 — Determinism & parallelism.** Chart is native `pic` + text; it registers
  media (the pic), so a chart-bearing slide is asset-bearing → already handled
  by `nodeUsesAssets` (Chart returns true via the default). Aspect read is a
  pure function of the bytes (deterministic). The warning text must be
  deterministic (no float formatting drift — round to a stable integer percent).

## 4. Recommendations

- **R1 — Read header dims via `image.DecodeConfig` (option A).** Automatic
  aspect detection, §7-consistent (header, not pixels), stdlib/CGo-free. Record
  the §7 boundary in a D-NNN ("reading image dimension headers is permitted;
  decoding pixel data is not"). Fix the Phase-11 Image comment to match.
- **R2 — Fit-within + thresholded warning.** Place the chart to fit the slot
  preserving aspect; fire one `LayoutWarning` when the aspect divergence exceeds
  ~15%. Round the reported delta to an integer percent for deterministic text.
  Degrade silently (no warning) if dimensions can't be read.
- **R3 — `ChartPlaceholder(box, opts…)` draws a labeled slot** (bordered rounded
  rect + centered "Chart" caption), returns the `*Shape`. The scene chart
  composer calls it when the asset is unresolved, so a missing chart shows a
  labeled placeholder instead of a blank region.
- **R4 — Keep the composer code_block-shaped**; a small shared "resolve → pic →
  caption" helper is optional, not required.

## 5. Open questions

- **Q1 — Aspect source: header read (A) vs caller field (B).** A fork to the
  maintainer; it sets the §7-boundary precedent and affects the IR. Recommend A.
- **Q2 — `ChartPlaceholder` visual: a labeled bordered slot vs geometry-only**
  (return a Box, draw nothing). Recommend the labeled slot (more useful, matches
  "sizes and positions a … slot"). A fork.
- **Q3 — Divergence threshold** (~10% / 15% / 20%). A tuning value; default 15%,
  not worth a fork unless the maintainer has a preference.
- **Q4 — Should fixing the Image `Fit` comment also *ship* aspect-aware fit?**
  Out of scope for Phase 17 (chart only); note it as a now-unblocked V1.x
  candidate.
