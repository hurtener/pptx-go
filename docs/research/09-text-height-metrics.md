# Brief 09 — text-height-metrics

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-20
**Motivating phase:** Phase 22 — content-aware text height

## 1. Question

The scene layout engine allots a **fixed** slot height per node regardless of
how much text the node carries: `preferredHeight` (`scene/render.go`) gives a
`Prose` `0.4"` per paragraph, a `List` `0.32"` per item, a `Heading` `0.6"`, a
`Quote` `1.1"`, a `Callout` `1.0"`, a `Table` `0.4"` per row — independent of
wrapping. A paragraph that wraps to three lines is therefore given the space of
one, so its text frame overruns the next stacked node, and the total stack
height is under-counted, so the overflow `LayoutWarning` never fires when real
content runs off the slide.

How can the engine estimate a node's wrapped height **deterministically**
(no live text measurement, no platform fonts) so that (a) stacked nodes stop
overlapping and (b) overflow is reported truthfully — while keeping single-line
content byte-identical to today's output?

## 2. Prior art surveyed

- **`scene/metrics.go` (Phase 13).** `naturalWidth(rt, theme)` already
  estimates the horizontal span of a `RichText` with a pinned char-width model:
  per run, `len(text) × floor(fontSize_pt × avgCharWidthFactor × emuPerPoint)`,
  `avgCharWidthFactor = 0.5`. It is pure, allocation-free, and deterministic,
  and is already trusted in production for horizontal alignment. It is the
  natural building block for a line-count estimate.
- **CSS / browser line-box model.** Real wrapping depends on glyph metrics,
  kerning, hyphenation, and the box model. pptx-go cannot reproduce that without
  a font engine (forbidden at runtime by P4 / §7 — no pixel or glyph parsing)
  and must not (determinism, RFC §10.1 byte-identical output regardless of
  worker count).
- **OOXML autofit (`normAutofit`, `spAutoFit`).** PowerPoint itself reflows and
  shrinks text at *display* time. That is the renderer's authority, not ours:
  the engine's job is to allot a *slot* whose height does not under-count the
  content, not to predict PowerPoint's exact reflow.
- **RFC §10.2.** "Layout is content-bbox-driven: a node reports its preferred
  bbox; the engine assigns it a slot; the node fits to the slot. Overflow is
  reported as a `LayoutWarning` (not an error)." The preferred bbox is supposed
  to be content-driven — the fixed-height shortcut is the gap this brief closes.
- **D-026 (engine, not product).** Height estimation is a *placement mechanism*,
  not a content opinion. Estimating wrapped lines is mechanism; deciding a slide
  is "too full" and rewording it is the caller's product behavior.

## 3. Findings

- **A char-budget line model is sufficient and deterministic.** Wrapped line
  count ≈ `ceil(naturalWidth(paragraph) / availableWidth)`, floored at 1. It
  reuses the existing pinned constants, needs no new calibration, and is a pure
  integer computation once `naturalWidth` returns. It over- or under-counts a
  given real paragraph by a bounded amount, but it is *monotonic* in text length
  and *never* claims a long paragraph is one line — which is the only property
  the overlap/overflow guarantees need.
- **Width must be threaded into `preferredHeight`.** The current signature
  (`preferredHeight(n)`) has neither the available width nor the theme, so it
  *cannot* be content-aware. Both the body region width (top-level stack) and
  the column/cell width (containers) must reach the estimator. The change is
  mechanical — every caller already knows its box width.
- **Single-line byte-identity falls out for free.** If every fixed per-line
  constant is reused as the *line height* (`Prose` stays `0.4"`/line, `List`
  `0.32"`/item, `Heading` `0.6"`, etc.), then a node whose text fits on one line
  yields exactly today's height (`lines = 1`). For the fixed-chrome nodes
  (`Quote`, `Callout`), the pattern `oldFixedHeight + (lines − 1) × lineHeight`
  preserves the single-line value exactly and only grows on wrap. This satisfies
  the additive / backward-compatible invariant for all single-line content.
- **Overflow becomes truthful with no separate code path.** The existing
  `totalH > box.H` check in `alignedStackIn`/`stackIn` already emits the
  overflow warning; once `totalH` is computed from content-aware heights it
  fires exactly when real wrapped content exceeds the body region. No new
  warning plumbing is needed.
- **Containers thread width but keep their shape.** `Grid` cells get
  `avail / cols`, `TwoColumn` sides get ~`avail / 2`, `Card`/`CardSection`
  bodies get `avail − padding`, `Table` cells get `avail / cols`. For
  single-line nested content these reduce to today's estimates; multi-line
  nested content grows the container's slot, which is a strict improvement
  (less clipping), not a regression.

## 4. Recommendations

1. Add `wrappedLines(rt, baseRole, avail, theme) int` to `scene/metrics.go`:
   `ceil(naturalWidthAt(rt, baseRole, avail) / avail)`, floored at 1, returning
   1 when `avail ≤ 0` or `theme == nil` (the byte-identical fallback path).
2. Change `preferredHeight` (and its helper `nodesHeight`) to take
   `(n, avail, theme)`. Make `Prose`, `List`, `Heading`, `Quote`, `Callout`,
   and `Table` content-aware via the line model; keep `Hero`, `Divider`,
   `Chip`, `Arrow`, `Image`, `Chart`, `CodeBlock` fixed (visuals/atoms do not
   text-wrap).
3. Keep all line-height / line-count constants pinned compile-time `EMU`
   literals (mirroring `cardChromeEst` / `estGap`) so output stays deterministic
   and worker-count-independent.
4. Re-baseline any existing assertion that hard-codes a multi-line node's slot
   geometry; single-line fixtures need no change.

## 5. Open questions

- **Calibrating the char-width factor per type role.** `avgCharWidthFactor`
  is a single global `0.5`. A future unit could resolve a per-family/per-weight
  factor for a tighter estimate; not needed for the overlap/overflow guarantees
  and deferred (it would touch determinism goldens). Pick up when a real deck
  shows systematic over/under-allotment.
- **Bullet/numbering indent in `List` width.** The list estimate uses the full
  column width, ignoring the marker indent (a slight under-count of wrap). A
  later refinement could subtract the resolved indent; deferred as it changes
  single-line-adjacent output only marginally.
- **Grow-to-fit (R2).** This brief sizes a node to its content; the inverse —
  distributing *slack* to flexible nodes so they grow to fill the frame — is the
  motivating product's R2 and gets its own phase/brief.
