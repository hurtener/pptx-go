# Brief 42 — decoration-watermark-anticollision (R11.11 verify-and-close)

**Subsystem:** scene — Layer 2 renderer (card watermark / decorations)
**Authored:** 2026-06-22
**Motivating phase:** Phase 59 — decoration-watermark-anticollision (R11.11, LOW · engine)

## 1. Question

The D-054 card watermark is anchored bottom-right inside the body box at low opacity;
under dense content a bottom-anchored watermark drawn after the body would sit over
real text. R11.11 asks that watermarks/decorations never reduce body legibility. Is
the current z-order + opacity sufficient, or is a residual-region restriction needed?

## 2. Prior art surveyed

- **`scene/render_card.go renderCardChrome`** — the watermark is step 6 (the **last**
  chrome shape), emitted *before* `renderCardChrome` returns the body box; `renderCard`
  then renders the body content. So the watermark is emitted **before** the body in
  document order → PowerPoint paints it **behind** (z-order first). Its color is
  `TokenColorAlpha(ColorAccent, cardWatermarkAlpha)` with `cardWatermarkAlpha = 13000`
  (~13% opacity).
- **`scene/render.go layout()`** — background decorations (`LayerBackground`) are
  emitted before the stacked body; foreground decorations (`LayerForeground`) after
  (the caller's explicit on-top choice).
- DECKARD R11.11 spec: ensure the watermark is emitted before body content in z-order
  (it is); keep the alpha low; **optionally** restrict the watermark to the residual
  region below the measured body content. Acceptance (an **OR**): the watermark either
  occupies only the residual empty region **or** is drawn behind content at a legible
  alpha; "a test … asserts the watermark box is below the measured content bottom
  **or** that z-order places it first".

## 3. Findings

- **The acceptance's z-order branch is already satisfied (D-054).** The watermark is
  emitted before the body content (z-order first) at ~13% opacity, and background
  decorations are likewise emitted before the body. The R11.11 acceptance is an
  explicit **OR** — "occupies the residual region OR is drawn behind content at a low
  alpha" — and the engine already takes the second branch. So R11.11 is a verify-and-
  close, not new mechanism.
- **The residual-region restriction is the *optional* alternative, not required.**
  Restricting the watermark to the measured residual region below the body would
  couple the chrome to the body's wrapped-line estimate for a LOW-priority cosmetic
  gain, and would change the watermark position (not byte-identical) for the common
  case where it currently sits behind sparse content harmlessly. The z-order + low
  alpha already guarantees legibility (the body text paints opaque on top of a 13%
  ghost), which is the binding requirement.
- **The close is the acceptance test.** Assert (1) the watermark text frame is emitted
  before the body content in the slide XML (z-order behind); (2) the watermark run
  carries the low ~13% alpha; (3) a card without a watermark emits no alpha run
  (inert / byte-identical).

## 4. Recommendation

Close R11.11 as already implemented by D-054 (watermark z-order-behind + low alpha;
background decorations z-order-behind), with the acceptance test above. D-091 records
the closure and that the optional residual-region restriction is intentionally not
adopted (the z-order + alpha branch of the OR-acceptance is the simpler, byte-
identical guarantee).

## 5. Open questions

- A caller wanting the watermark confined to empty space can place it as a foreground
  decoration in a chosen region instead; the engine's default (behind, low alpha) is
  the safe, legible one.
