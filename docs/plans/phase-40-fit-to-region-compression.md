# Phase 40 — fit-to-region compression

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §10 (layout engine), §10.2 (content-aware height)
**Deps:** Phase 13 (alignment / `alignedStackIn`), Phase 23 (`VAlignFill` /
`distributeFill`, D-052), Phase 39 (R10.1 wrapped-header geometry, D-070)
**Status:** Done

---

## 1. Goal

Add an opt-in `VAlignFit` body-stack mode that deterministically compresses an
over-full slide — shrinking inter-node gaps then slot heights toward pinned
floors — so its content lands inside the region instead of clipping off-slide.

## 2. Why now

R10.2 is the second CRITICAL of Wave 10 (R10 content fit & density) and the
direct fix for the recreation's off-canvas bottom row (DECKARD R10.2 gap). It
builds on the wrapped-header geometry landed in Phase 39 (D-070) and the
`VAlignFill` grow-to-fit primitive (D-052) — `VAlignFit` is its inverse. The
master plan (`docs/plans/README.md` Wave 10) sequences it right after R10.1.

## 3. RFC sections implemented

- `RFC §10` — the layout engine's body-stack placement gains a compression
  pass; partially — container-internal fit (bento rows, card body) is owned by
  sibling phases R10.3 / R10.4.
- `RFC §10.2` — content-aware placement: the fit pass consumes the same
  deterministic `preferredHeight` estimates the overflow warning uses.

## 4. Brief findings incorporated

- `docs/research/23-fit-to-region-compression.md` — *the off-slide clip is a
  placement (not content) problem; the two levers inside `alignedStackIn` are
  gaps and slot heights* → the fit pass floors the inter-node gap toward
  `gapMin`, then proportionally scales slot heights toward `sMin`.
- `docs/research/23-fit-to-region-compression.md` — *a `VAlignFit` enum value is
  the minimal additive surface; it is the inverse of `VAlignFill`* → added to
  the `VAlign` enum after `VAlignFill`, routed through the same switch, no
  `Alignment` struct change.
- `docs/research/23-fit-to-region-compression.md` — *pinned floors
  `gapMin=SpaceXS`, `sMin=0.60`; basis-point integer math is worker-count
  independent* → `fitCompress` uses `ResolveSpace(SpaceXS)` and `h*sBP/10000`.
- `docs/research/23-fit-to-region-compression.md` — *keep the warn mode-aware:
  non-Fit unchanged; Fit warns iff content still overflows after compression* →
  the overflow warn recomputes the achieved height for `VAlignFit`.
- `docs/research/09-text-height-metrics.md` — *deterministic `preferredHeight`
  feeds overflow detection* → the fit pass operates on those same heights, so it
  is exactly as trustworthy (R10.10 closes remaining estimator gaps).

## 5. Findings I'm departing from

- `docs/research/23-fit-to-region-compression.md` recommends keeping
  `fitCompress` reusable so the **container stackers** adopt it later. This plan
  scopes the actual *wiring* of fit into containers to R10.3 (bento) / R10.4
  (card body) — `stackIn` has no alignment parameter and threading a mode
  through every caller belongs with those reqs. **Departing because** the
  CRITICAL off-slide clip is the top-level body stack; `fitCompress` ships as a
  reusable theme-aware primitive but only `alignedStackIn` calls it this phase.
  Documented as a §4.3 deviation (see §16).
- The R10.2 spec lists a card-padding sub-step (→ `padMin`) and a font
  type-scale sub-step inside the fit pass. **Departing because** those live in
  `renderCard` (R10.7 `density-aware-card-padding`) and the display-text shrink
  (R10.5 `display-text-shrink-to-fit`); the spec itself cross-references them.
  Phase 40 delivers the gap + slot-height steps, which satisfy the ≤25%-overflow
  acceptance on their own; R10.7 / R10.5 plug into the same pass.

## 6. Decisions referenced

- `D-052` — `VAlignFill` grow-to-fit — `VAlignFit` is the compression inverse,
  reusing the same integer-EMU / last-node-remainder determinism discipline.
- `D-070` — content-aware card header height — the prior Wave-10 content-fit unit
  this builds on; estimator parity continues to be R10.10's job.
- `D-026` — engine, not product — compressing slot geometry to fit a frame is a
  mechanism the caller opts into (`VAlignFit`), not a taste decision.
- **New:** `D-071` — fit-to-region compression — filed in this PR.

## 7. Architecture

`VAlignFit` is a new `VAlign` value (after `VAlignFill`). The body-stack layout
`alignedStackIn` gains one branch: when the mode is `VAlignFit` and the stack
overflows (`totalH > box.H`), it calls the new renderer method `fitCompress`,
which mutates the per-node `heights` slice in place and returns the compressed
inter-node gap. Placement is top-pinned (`startY = box.Y`).

```text
alignedStackIn(box, nodes, align)
  heights[i] = preferredHeight(nodes[i], box.W)
  totalH     = Σheights + gap*(n-1)
  if align.Vertical == VAlignFit && totalH > box.H:
      effGap = fitCompress(heights, bodyH, gap, box)   // mutates heights
          step 1: effGap = clamp((box.H - bodyH)/(n-1), gapMin, gap)   // gaps → SpaceXS
          step 2: if bodyH + effGap*(n-1) > box.H:                     // still over?
                      sBP  = clamp((box.H - effGap*(n-1))*10000/bodyH, 6000, 10000)
                      heights[i] = heights[i] * sBP / 10000            // slot heights → ×0.60 floor
  warn iff (non-Fit: totalH>box.H) or (Fit: Σheights + effGap*(n-1) > box.H)
  place top-pinned with effGap
```

All math is integer EMU / basis points — a pure function of the heights, the
gap, and `box.H`, so the output is identical regardless of worker count.

## 8. Files added or changed

```text
scene/align.go                          # CHANGED — VAlignFit const + String() case
scene/render.go                         # CHANGED — fitCompress helper; VAlignFit branch + mode-aware warn in alignedStackIn
scene/render_fit_test.go                # NEW — white-box fitCompress + placement acceptance + byte-identical-when-fits
scene/render_parallel_test.go           # CHANGED — TestRenderDeterministic_VAlignFit guard
scripts/smoke/phase-40.sh               # NEW — phase smoke
docs/research/23-fit-to-region-compression.md   # NEW — brief 23
docs/research/INDEX.md                  # CHANGED — register brief 23
docs/plans/phase-40-fit-to-region-compression.md # NEW — this plan
docs/plans/README.md                    # CHANGED — Wave 10 phase index row
docs/decisions.md                       # CHANGED — adds D-071
docs/glossary.md                        # CHANGED — VAlignFit, fit-to-region compression
docs/site/guide/scene-layout.md         # CHANGED — document VAlignFit (user-facing)
skills/compose-a-scene/SKILL.md         # CHANGED — VAlignFit in the alignment section
```

## 9. Public API surface

```go
// scene
const VAlignFit VAlign = … // opt-in: when the body stack overflows its region,
                            // deterministically compress gaps then slot heights
                            // toward pinned floors so content fits the frame.
                            // Inverse of VAlignFill; byte-identical when content
                            // already fits. Set via SceneSlide.Content.Vertical.
func (v VAlign) String() string // adds the "fit" case
```

No breaking change: `VAlignFit` is appended to the enum; all existing iota
values and the zero value (`VAlignTop`) are unchanged.

## 10. Risks

- **R1 — byte-identical regression for existing modes.** A change to
  `alignedStackIn` could perturb Top/Center/Bottom/Justify/Fill output.
  **Mitigation:** the `VAlignFit` branch is the only new code path and is gated
  on `align.Vertical == VAlignFit`; the warn for non-Fit modes is the unchanged
  `totalH > box.H`. A test asserts `VAlignFit` == `VAlignTop` placement when
  content fits.
- **R2 — determinism under parallel render.** **Mitigation:** integer/basis-
  point math only; `TestRenderDeterministic_VAlignFit` renders the same deck at
  1 and 8 workers and asserts byte equality.
- **R3 — over-compression hides genuine overflow.** A deck that overflows beyond
  the pinned floors must still warn. **Mitigation:** the mode-aware warn fires
  when the post-compression height still exceeds the region; the `sMin=0.60`
  floor bounds compression so an extreme overflow surfaces honestly.

## 11. Acceptance criteria

1. For a body stack whose preferred height exceeds the region by up to ~25%,
   `VAlignFit` produces a layout whose last node's bottom is ≤ `box.Bottom()`
   (no off-slide content), using only the pinned compression steps.
2. A stack that already fits renders byte-identically with `VAlignFit` vs
   `VAlignTop` (flag on or off, same output).
3. Identical inputs always yield identical EMU geometry; a deck rendered at 1
   and 8 workers is byte-identical (`TestRenderDeterministic_VAlignFit`).
4. When compression cannot fully fit the content within the pinned floors, the
   `content overflows its region` warning still fires; when it fits, the warning
   is suppressed.
5. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package (no override) |

## 13. Smoke check

`scripts/smoke/phase-40.sh` verifies the criteria:

1. `OK:` library builds CGo-free.
2. `OK:` `VAlignFit` fits an over-full stack inside the region
   (`TestFitCompress_FitsOverfullStack`).
3. `OK:` `VAlignFit` is byte-identical to `VAlignTop` when content fits
   (`TestFitCompress_ByteIdenticalWhenFits`).
4. `OK:` fit render stays deterministic
   (`TestRenderDeterministic_VAlignFit`).

## 14. Tests

- **Unit:** `scene` (white-box `package scene`) — `fitCompress` gap-only fit,
  gap+height fit, floor-capped residual, no-op when fitting; placement-level
  acceptance via `r.layout`; byte-identical `VAlignFit` vs `VAlignTop`.
- **Round-trip golden:** n/a — no new builder primitive or OOXML; this is a
  scene-layout placement change.
- **Integration:** no — no cross-subsystem seam opened (scene-internal).
- **Fuzz:** no parse/decode surface.
- **Benchmark:** no new hot reusable artifact.

## 15. Vocabulary added

- `VAlignFit` — opt-in body-stack mode that compresses an over-full stack
  (gaps then slot heights, toward pinned floors) so content fits its region.
- `fit-to-region compression` — the deterministic engine pass behind
  `VAlignFit`; the inverse of `VAlignFill`'s grow-to-fit.

## 16. Plan deviations encountered during implementation

- **Container-internal fit deferred to R10.3 / R10.4.** `fitCompress` ships as a
  reusable theme-aware primitive but only `alignedStackIn` (the top-level body
  stack — the CRITICAL off-slide-clip site) calls it this phase. `stackIn` has
  no alignment parameter; wiring fit into bento rows and card bodies is the
  explicit scope of R10.3 / R10.4. Restates AC1 to the body stack. (§4.3.)
- **Card-padding / type-scale sub-steps deferred to R10.7 / R10.5,** per the
  R10.2 spec's own cross-references; the gap + slot-height steps satisfy the
  ≤25%-overflow acceptance on their own.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-40.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-071).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
