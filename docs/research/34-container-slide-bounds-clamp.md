# Brief 34 — container-slide-bounds-clamp

**Subsystem:** scene — Layer 2 renderer (container layout)
**Authored:** 2026-06-22
**Motivating phase:** Phase 51 — container-slide-bounds-clamp (R11.3, CRITICAL · engine)

## 1. Question

A Bento / Grid / Card handed a box whose bottom exceeds the slide's printable area
(because an over-full body stack pushed it down, or a tall fixed-height container
was requested) divides that box as given and draws cells off the bottom edge — they
clip, and the slide chrome footer is overlapped (recreation slides 6, 7). How can a
container be guaranteed never to draw below the slide safe area, deterministically
and byte-identically when its content already fits?

## 2. Prior art surveyed

- **`scene/render.go bodyRegion()`** — already computes the per-slide body box as
  `slide − 2·bodyMargin`, and (when `r.chrome.Enabled`) subtracts
  `chromeEyebrowH + chromeBandGap` from the top and `chromeFooterH + chromeBandGap`
  from the bottom. This is exactly the "slide minus margins minus reserved chrome
  bands" the R11.3 spec calls `safeArea`. So the safe area already exists — it is
  `bodyRegion()`.
- **`scene/render.go layout()`** — assigns each top-level node a slot via the
  body-stack; an over-full stack (Σ preferredHeight > region) places the last
  node(s) with their bottom **below** `bodyRegion.Bottom()`. `VAlignFit` (R10.2)
  compresses such a stack, but it is opt-in; a default `VAlignTop` stack still
  overflows.
- **`scene/render_bento.go bentoGeometry`** — `rowH = (box.H − gaps)/nRows` from
  whatever box it is handed; no clamp to the printable area.
- **`scene/render_container.go renderGrid`** — `layout.Grid(box, …)` from the handed
  box; `scene/layout` is a pure geometry package (no theme/renderer), so the clamp
  belongs in the renderer, before the `layout.Grid` call.
- **`scene/render_card.go renderCard`** — likewise consumes its box directly.
- **`pptx.Box.Bottom() = Y + H`** — the primitive the clamp compares against.
- DECKARD R11.3 spec: compute a per-slide `safeArea`; pass safeArea-derived boxes
  into `renderBento`/`renderGrid`/`renderCard`; clamp `box.H` so `box.Bottom() <=
  safeArea.Bottom()`, shrinking `rowH` proportionally and emitting `r.warn` on
  clamp; determinism unaffected; byte-identical when content fits.

## 3. Findings

- **`safeArea` already exists as `bodyRegion()`.** The R11.3 safe area is identical
  to the chrome-aware body region the renderer already computes. So the clamp does
  not need a new geometry derivation — it clamps against `r.bodyRegion().Bottom()`.
  Exposing it as a named `safeArea()` (a thin alias) keeps the intent legible and
  gives R11.4 a shared name.
- **A single entry-point clamp covers all three containers.** Rather than threading
  a new parameter, clamp the incoming box at the top of `renderBento`, `renderGrid`,
  and `renderCard`:

      func (r *renderer) clampToSafeArea(box pptx.Box, slideID string) pptx.Box {
          if sb := r.safeArea().Bottom(); box.Bottom() > sb {
              h := sb - box.Y
              if h < 0 { h = 0 }
              if h < box.H {
                  box.H = h
                  r.warn(slideID, "container overflow: content exceeds the slide safe area, clamped")
              }
          }
          return box
      }

  The clamp only fires when `box.Bottom() > safeArea.Bottom()` — a container whose
  bottom already sits at or above the safe-area bottom is returned unchanged, so
  every fitting layout (and `VAlignFill`, which grows *to* the region bottom, and a
  sole container handed the full body region whose `Bottom() == safeArea.Bottom()`)
  is **byte-identical**. Pure integer clamp → deterministic.
- **Nesting does not double-warn.** A clamp on an outer Grid shrinks its box, so the
  Cards/Bentos it lays out get sub-boxes inside the clamped region and never
  individually exceed the safe area → their own clamp is a no-op. The warning fires
  once, at the outermost overflowing container.
- **The clamp is defense-in-depth, complementary to `VAlignFit`.** `VAlignFit`
  (opt-in) compresses an over-full stack *before* placement; the clamp guarantees
  the *invariant* (nothing below the safe area) even for the default top-anchored
  stack that opts out — it caps the drawn height rather than reflowing content.
  Together: `VAlignFit` makes content fit legibly when asked; the clamp guarantees
  it never draws off-canvas regardless.
- **Top edge.** The body stack always starts at `bodyRegion().Y` (≥ the top margin
  / chrome band), so `box.Top() >= topMargin` already holds; the clamp need only
  guard the bottom. A property test still asserts both bounds for completeness.

## 4. Recommendation

Add `safeArea()` (alias of `bodyRegion()`) + `clampToSafeArea(box, slideID)` to
`scene/render.go`; call the clamp at the entry of `renderBento`, `renderGrid`, and
`renderCard`. Warn once on clamp. Byte-identical when content fits; a property test
feeds an over-tall Bento and Grid and asserts no emitted box exceeds the safe area
and a warning logged, plus a fitting-layout byte-identical check and a parallel
determinism guard. D-083 records the clamp as a defense-in-depth bound
complementary to the opt-in `VAlignFit`.

## 5. Open questions

- Should the clamp also compress inner content (like `VAlignFit`) rather than cap
  the box? No — capping is the deterministic *invariant*; reflowing is the opt-in
  `VAlignFit` job. Mixing them would make the default path non-byte-identical.
