# Brief 10 — grow-to-fit

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-20
**Motivating phase:** Phase 23 — vertical fill / grow-to-fit

## 1. Question

A heading followed by one block leaves the bottom half of the slide empty and
reads thin. The body-stack alignment options added in Phase 13 — `VAlignCenter`,
`VAlignJustify` — *float* the stack (they move it, or spread the inter-node
gaps), but the reference "designed" look is the **heading pinned at the top and
the content growing to consume the remaining frame** (tall cards, full-bleed
grids). How can the engine distribute leftover body height to the nodes that can
absorb it — deterministically, additively, and without baking in any opinion
about *which* nodes a caller wants to grow?

## 2. Prior art surveyed

- **`scene/render.go::alignedStackIn` (Phase 13).** The body stack already
  computes each node's preferred height and the total stack height, derives a
  vertical start `Y` from the `VAlign`, and (for `VAlignJustify`) redistributes
  slack into the gaps. This is the exact seam where a "grow the nodes" mode
  belongs — it already owns the slack arithmetic.
- **The scene geometry engine (`scene/layout`).** `layout.Grid` divides
  `parent.H` into equal rows (`availH / rows`); `layout.Columns` gives each
  column the full `parent.H`. Both are **already height-driven**: hand them a
  taller box and they produce taller cells/columns with no change.
- **`render_card.go::renderCardChrome`.** The card background rect uses the
  whole `box`, the accent stripe scales with `box.H`, and the body region runs
  to `box.Bottom() - pad`. A taller card box yields a taller card with no
  change.
- **`render_image.go` / `render_chart.go` / `render_table.go`.** Image fills its
  box (`AddImage(interior)`), Chart contains-to-fit within its slot, Table is
  created at `box.H`. All consume the box height they're handed.
- **CSS flexbox `flex-grow`.** The canonical model: fixed-basis items keep their
  size; flexible items share the leftover main-axis space in proportion to a
  grow factor. The engine analogue is "fixed leaves keep preferred height;
  flexible nodes share the slack".
- **D-026 (engine, not product).** *Which* nodes should grow, and whether a
  sparse slide should be filled at all, is taste — the caller's. The engine must
  only provide the *mechanism* (a fill mode + a deterministic distribution),
  never decide a slide "looks thin" on its own.

## 3. Findings

- **The renderers already honor a taller box.** The DECKARD requirement warns
  that container renderers must distribute extra height to their rows/cells/
  bodies "rather than rendering at a fixed height in a taller box." Auditing the
  current code, they already do: every flexible renderer is geometry-driven off
  the slot box (`layout.Grid` rows, `layout.Columns` height, card chrome to
  `box.Bottom()`, image/chart/table to `box.H`). So the *only* new mechanism
  needed is growing the flexible node's **slot** in the body stack; the taller
  box then propagates correctly — including one level of nesting, since a grown
  Grid passes its taller cell box straight to the child's renderer.
- **A single new `VAlign` value is the cleanest surface.** `VAlignFill` (a new
  enum constant, so the zero value and every existing value are untouched) is
  top-pinned like `VAlignTop` and, after fixed nodes take their preferred
  height, grows the flexible nodes to consume the slack. No new field on
  `SceneSlide`, no per-node flag — the existing `Content.Vertical` carries it.
- **"Flexible" is a fixed, intrinsic node set.** Per the requirement: containers
  (`Grid`, `TwoColumn`, `Card`, `CardSection`, `Table`) and the two stretchable
  visuals (`Chart`, `Image`). Text leaves and atoms (`Heading`, `Prose`, `List`,
  `Quote`, `Callout`, `Hero`, `Chip`, `Arrow`, `Divider`, `SectionDivider`,
  `Flow`) stay at preferred height — stretching text is meaningless. `CodeBlock`
  is a raster of monospaced code; growing it would distort the listing, so it is
  **not** flexible.
- **Proportional distribution is deterministic and natural.** Share the slack
  among flexible nodes in proportion to their preferred height (the bigger node
  grows more, relative proportions preserved), with the rounding remainder
  assigned to the last flexible node. Pure integer EMU math → worker-count
  independent. When the flexible preferred heights sum to zero, fall back to an
  equal split.
- **Backward-compatibility is total.** `VAlignFill` is opt-in; every existing
  scene (which uses `VAlignTop`/`Center`/`Bottom`/`Justify`) renders
  byte-for-byte unchanged. A `VAlignFill` slide with no flexible node degrades
  to top-alignment (nothing to grow), with no warning — the absence of a
  growable node is a caller choice, not an error.

## 4. Recommendations

1. Add `VAlignFill` to the `VAlign` enum (after `VAlignJustify`) + its `String`
   case. Top-pinned start, standard gaps.
2. In `alignedStackIn`, after the overflow check, when `Vertical == VAlignFill`
   and `slack = box.H − totalH > 0`, distribute `slack` across the flexible
   nodes (proportional to preferred height; remainder to the last; equal split
   if their heights sum to zero) by adding to the per-node height before
   placement.
3. Add an unexported `isFlexible(SlideNode) bool` listing the seven flexible
   kinds; keep it the single source of truth for the set.
4. Touch **no** container renderer — they already consume the box height. Record
   this audit finding in the plan so the absence of renderer changes is
   intentional, not an oversight.

## 5. Open questions

- **Recursive fill inside a container.** `VAlignFill` grows a container's slot
  and the geometry propagates one level (grid cell → child renderer). Making a
  container *also* redistribute slack among its own stacked children (so a tall
  card's body content spreads, not just the card chrome) is a deeper refinement;
  the acceptance only requires the container's cells to grow, so it is deferred.
- **Min/max grow bounds & per-node grow weights.** flexbox has `flex-basis` /
  `flex-grow` / `max-height`. A future unit could expose a per-node grow weight;
  not needed for the reference look and deferred (it would add IR surface).
- **Interaction with content-aware height (Phase 22).** Fill consumes *positive*
  slack only; when content already overflows (`slack ≤ 0`) nothing grows and the
  Phase-22 overflow warning still fires. The two compose cleanly.
