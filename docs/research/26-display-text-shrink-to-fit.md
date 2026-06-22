# Brief 26 — display-text-shrink-to-fit

**Subsystem:** scene — Layer 2 renderer (+ a pptx builder run override)
**Authored:** 2026-06-22
**Motivating phase:** Phase 43 — display text shrink-to-fit (R10.5, HIGH · engine)

## 1. Question

A display-class run (Hero title, Stat value, big price, Heading) renders at a
fixed `TypeRole` size and wraps or clips when it is wider than its box (the
recreation's "$4,000+" wrapped to two lines in a narrow pricing column; titles on
slides 4/5 wrapped where the reference keeps one line). How can the engine,
**opt-in and deterministically**, downscale such a run's font size to fit its box
width on one line — within a pinned minimum ratio — without any measurement and
byte-identical when off?

## 2. Prior art surveyed

- `scene/metrics.go` `naturalWidth(rt, theme)` — the pinned, deterministic
  text-width estimator (`Σ len(text) × floor(size_pt × factor × emuPerPoint)`).
  Already feeds `wrappedLines` / `preferredHeight`. A pure function, no
  measurement.
- `scene/render_leaves.go` `renderHero` (Title at `TypeDisplay` via `plainPara`),
  `renderHeading` (Text at `headingRole(level)` via `addRichText`);
  `scene/render_stat.go` `renderStat` (Value at `TypeDisplay`, Bold).
- `pptx/text_layout.go` `RunStyle.toProps`: the run's emitted font size comes
  from `spec.Size` (`t.ResolveType(rs.TypeRole).Size`) → `a:rPr/@sz` (1/100 pt).
  `RunStyle` already carries per-run *overrides* (`Tracking`, `Case`) that win
  over the role — the precedent for a per-run size modifier.
- `pptx/text.go` `Run.FontSize()` — the read accessor returning the emitted size
  (so a scaled run round-trips through `pptx.Open`).
- Phase 40 (R10.2, D-071) / Phase 41 (R10.3, D-072) — the pinned-floor +
  quantized basis-point pattern this brief reuses for the scale.
- DECKARD R10.5 spec: opt-in AutoFit; `scale = clamp(boxW/naturalWidth, ratioMin,
  1)`, applied as a reduced effective point size *quantized to a fixed step* when
  `naturalWidth > boxW`; never upscale; default OFF byte-identical; a pure
  function of `(text, role, boxW, theme)` — no measurement.

## 3. Findings

- The width estimator already exists (`naturalWidth`); the only missing pieces are
  (a) a **builder seam** to emit a scaled size, and (b) a **pure scale function**
  the renderer computes from `naturalWidth` vs the box width.
- **Builder seam: a per-run `FontScale` multiplier, not an absolute size.** A
  multiplier on the resolved role size keeps P2 intact — the `TypeRole` size token
  stays the source of truth and a theme swap still re-skins the base; AutoFit only
  scales it. An absolute per-run point size would be a literal that ignores the
  theme. `FontScale` mirrors `Tracking`/`Case` (per-run overrides) and is
  byte-identical when 0/unset (`size = spec.Size` exactly). Its effect round-trips
  via the existing `Run.FontSize()` (the emitted `@sz`).
- **Quantized, integer scale.** `fitScale(natW, boxW)` returns 0 (no scaling) when
  the text fits, else `q = floor(boxW·10000/natW)` quantized **down** to a fixed
  step (0.025) and floored at the pinned ratio (`0.60`). Flooring down guarantees
  `natW · q ≤ boxW` (fits); quantizing to a fixed step keeps the float
  deterministic; never returns ≥ 1 (never upscale). Pure integer EMU / basis
  point — no measurement, identical inputs → identical scale.
- **Opt-in per display node.** An additive `AutoFit bool` on `Hero`, `Stat`, and
  `Heading` (the display class the spec names). Off (zero) → no scale → the run
  renders at the role size, byte-identical. Only the *display* run scales (Hero
  Title, Stat Value, the whole Heading text) — eyebrows/labels/subtitles keep
  their role.
- **Heading is multi-run.** `addRichText` is factored into `addRichTextScaled`
  (taking a uniform `FontScale`); `renderHeading` computes one scale for the whole
  text and applies it to every run, so a styled heading downscales as a unit.
- **No estimator feedback loop.** AutoFit does not change `preferredHeight` (slot
  height) — it only shrinks the emitted font so the run fits horizontally. Vertical
  fit is R10.2/R10.10's concern; keeping them separate avoids coupling.

## 4. Recommendations

1. **pptx:** add `RunStyle.FontScale float64` (0/unset = role size). In `toProps`,
   `size := spec.Size; if rs.FontScale > 0 { size = spec.Size × rs.FontScale }`;
   emit `int(size×100)`. Round-trip test: a scaled run emits the expected `@sz`
   and `Run.FontSize()` reads it back.
2. **scene:** add `fitScale(natW, boxW) float64` (pure; quantized, pinned floor).
3. **scene:** add `AutoFit bool` to `Hero`, `Stat`, `Heading`; wire the display
   run through `FontScale = fitScale(naturalWidth(display text at its role),
   box.W)`. Factor `addRichTextScaled` for the multi-run Heading.
4. Tests: `fitScale` (fits→0, overflow→fitting quantized scale, floor cap);
   per-node shrink-when-overflow + byte-identical-when-off + emitted `@sz`;
   determinism guard; smoke `phase-43.sh`.

## 5. Open questions

- **AutoFit on `Chip`/`Arrow` labels / table cells** — out of scope; the spec
  names the display class. Reusable (`fitScale` + `FontScale`) if a later req
  wants them.
- **Multi-line shrink-to-fit-height** (shrink until the text fits a fixed box
  height, not just width) — a different mechanism; not R10.5.
