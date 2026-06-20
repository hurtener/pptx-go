# pptx-go — product requirements from Deckard

> **Audience:** an agent (or contributor) working in *this* repo who wants to pick up the next
> engine work. This file states what **Deckard** — the agent-first slides product built on
> pptx-go — needs from the engine, and why, framed so you can implement a requirement without
> knowing Deckard's internals. Each requirement is a self-contained unit with engine-facing
> acceptance criteria.

## The relationship (read first)

pptx-go is an **engine, not a product** (D-026): it turns a typed `scene.Scene` into a
deterministic `.pptx` and nothing else. Deckard is the caller — it decides *what* slides to
emit, *what* text to write, *which* tokens to pick, and *whether* a layout warning matters.
**So these requirements ask for new engine MECHANISMS, never product opinions.** Deckard owns
the taste (palettes, when to use chrome, copy); the engine owns the faithful, deterministic
render of whatever typed scene it's handed.

The product's north star is an **agency-grade investor deck**: centered covers, dark full-bleed
section slides, cards with colored header bands + icon chips + status dots + watermark numbers, a
"VS" badge between compared cards, row-labeled bento grids, big-number stat/pricing strips, and
consistent chrome (a section eyebrow + rule, a footer logo + `02 / 11` page number). The engine
must be *capable* of all of that; Deckard drives it.

## Invariants every change MUST hold (the engine's contract)

1. **Deterministic.** Byte-identical output regardless of worker count. Pure functions, pinned
   integer-EMU constants, no `Date.now`/`rand`/map-iteration-order dependence. Add a worker-count
   determinism test (mirror `scene/render_parallel_test.go`).
2. **Additive + backward-compatible.** New capabilities are new optional fields whose **zero value
   reproduces today's render byte-for-byte**; the full existing scene suite must pass unchanged.
   (The ONE exception is R1 — see its note.)
3. **Unopinionated.** Provide the mechanism; do not bake in a palette, a default chrome, or a
   "make it pretty" heuristic. The caller supplies the values.
4. **Gate per unit:** `gofmt -l scene/` clean · `go build ./...` · `go vet ./...` ·
   `go test -race ./...` green, plus a golden/determinism test for the new behavior.

## What the engine ALREADY provides (do NOT rebuild)

- **Alignment** — `SceneSlide.Content = Alignment{Vertical VAlign; Horizontal HAlign}` + per-node
  `Align HAlign`; vertical top/center/bottom/justify of the body stack; horizontal via paragraph
  alignment for text (`scene/align.go`, `scene/render_leaves.go`).
- **Dark variant + full-bleed backgrounds** — `SceneSlide.Variant` (light/dark, dark derives a
  legible dark theme) and `SceneSlide.Background` (color/gradient/asset, full-bleed)
  (`scene/background.go`).
- **Deterministic text-width metrics** — `scene/metrics.go` (`naturalWidth`).
- **Rich card FIELDS** — `Card` already has `Eyebrow`, `Icon`, `HeaderPill`, `Fill`, `BorderStyle`,
  `Size`, `Layout` (icon-top), `Elevation`, and they render. (R4 only adds what's still missing.)

---

## Requirements (priority order)

### R1 — Content-aware text height  ·  HIGH  ·  ⚠ regenerates goldens

**Product need.** Agent-authored slides overlap and clip. `preferredHeight` (`scene/render.go`)
allots a fixed height per paragraph / per list item regardless of text length, so a paragraph
that wraps to 3 lines is given ~1 line of space and its text frame overruns into the next stacked
node. The same under-count means overflow is under-reported — content runs off the slide and the
renderer never warns, so the product can't tell the user "this slide is too full."

**Engine spec.** Make `preferredHeight` content-aware: estimate the wrapped line count from the
text runs, the resolved font metrics for the node's type role, and the available width
(reuse/extend `scene/metrics.go`), then `height = lines × line-height` (+ existing padding).
Apply to Prose (per paragraph), List (per item), Quote, Callout body, Table cells, Heading.
Keep it a pure, deterministic function (pinned chars-per-line / line-height constants).

**Acceptance.** A paragraph long enough to wrap to N lines is allotted ≥ N line-heights; the next
stacked node's Y is below the wrapped text's bottom (a golden asserts no overlap for a multi-line
fixture). A slide whose real wrapped content exceeds the body region emits the overflow
`LayoutWarning` (it currently doesn't). Determinism holds.

**Backward-compat note.** This is the ONE requirement that intentionally CHANGES layout for any
multi-line text, so existing golden snapshots must be **regenerated and eyeballed** (the new
output is the correct one — less overlap, truthful overflow). Single-line content is unaffected.

### R2 — Vertical fill / grow-to-fit  ·  HIGH

**Product need.** A heading + one block leaves the bottom half of the slide empty and reads thin.
Alignment's `VAlignCenter`/`VAlignJustify` *float* the block, but the reference look is the
**heading pinned at the top and the content GROWING to fill the frame** (tall cards, full bleed).

**Engine spec.** A fill mode (e.g. `VAlignFill`, or a per-node "flex" marker) where, after fixed
nodes (headings, prose) take their preferred height, the remaining body height is distributed to
the **flexible** nodes (containers: `Grid`, `TwoColumn`, `Card`, `CardSection`, `Table`; visuals:
`Chart`, `Image`) so they grow to consume it. Critically, the **container renderers must honor a
taller box** — `render_container.go` / `render_card.go` must distribute extra height to their
rows/cells/card bodies rather than rendering at a fixed height in a taller box.

**Acceptance.** A heading + grid slide renders the heading at top and the grid filling down to the
bottom margin; given a taller box, Grid/Card/TwoColumn produce proportionally taller cells. Zero
fill (today's mode) = byte-identical. Determinism holds.

### R3 — Slide chrome (section eyebrow + footer)  ·  MEDIUM

**Product need.** Reference decks read "designed" partly because every slide carries a section
eyebrow + hairline rule at the top (`01 — DIRECTION`) and a footer with a brand mark + a
`N / total` page number. The engine has no chrome concept, so Deckard can't produce it.

**Engine spec.** Optional, opt-in chrome rendered **outside** the body region (so it never
overlaps content): a top section-eyebrow band (eyebrow text + a hairline rule) and a bottom
footer band (a left brand slot — text or an asset id — and a right `N / total` page number).
Driven by fields on `SceneSlide`/`Scene` (e.g. `Scene` carries the brand slot + total; each
slide its section label + index). The body region shrinks to make room when chrome is enabled.

**Acceptance.** A deck with chrome enabled renders a consistent footer page number + section
eyebrow on each slide, outside the body box; chrome disabled = today's bare slide, byte-identical.

### R4 — Rich card visuals: header band, status dot, watermark  ·  MEDIUM

**Product need.** `Card` already supports fill/icon/eyebrow/headerPill/elevation. Reference cards
add three things it can't express: a **colored header band** (the top portion of the card a solid
accent color with the body in surface below — distinct from a full `Fill`), a **status dot** (a
small colored dot, top-right), and a **watermark** (a large faint number/label behind the card
content, e.g. a ghosted `01`).

**Engine spec.** Additive `Card` fields: `HeaderFill ColorRole` (a banded header region in that
color, body keeps `Fill`), `StatusDot ColorRole` (a small filled dot in the top-right corner),
`Watermark string` (a large, low-opacity label drawn behind the card body). Render in
`render_card.go`.

**Acceptance.** A Card with `HeaderFill` + `StatusDot` + `Watermark` renders all three (matches a
banded reference card); each zero-value omits its element; byte-identical when all unset.

### R5 — Composition primitives: center badge, inter-column connectors, row-labeled bento  ·  LOW

**Product need.** The reference compares two cards with a centered **"VS" badge** between them,
draws **connector arrows between 3 columns** (an architecture diagram), and uses a **row-labeled
bento grid** (rows with a left label and variable column spans).

**Engine spec.** (a) A `TwoColumn` option for a centered connector badge between the columns
(badge text). (b) Inter-column connectors for a multi-column container. (c) A grid/bento variant
with per-row left labels and variable column spans (`Ratio` per row already partly exists on
`Grid` — extend to row labels + spans). These can land as separate sub-units.

**Acceptance.** The two-card-with-center-badge slide and a row-labeled grid render; existing Grid/
TwoColumn unaffected when the new options are unset.

### R6 — Stat node (big number)  ·  LOW

**Product need.** Pricing and metric slides use big-number stats (`$2,200`, `38%`) with a label
and an optional delta. There's no engine node for a hero number, so they're faked with headings.

**Engine spec.** A `Stat` leaf node: `Value string` (rendered at display scale), `Label string`,
optional `Delta string` + `DeltaTone` (up/down/neutral color). A row of `Stat`s inside a `Grid`
forms a metric/pricing strip.

**Acceptance.** A `Stat` renders a large value + label (+ delta when set); a `Grid` of `Stat`s
renders a strip; validates in Stage-1.

### R7 — Expose resolved per-slide colors  ·  LOW (enables product-side contrast checks)

**Product need.** Deckard validates text/surface contrast, but it can't see the colors the engine
actually resolved per slide — especially for `VariantDark`, where the engine swaps to a derived
dark palette. The product validator currently checks contrast against the light theme and false-
flags white-on-dark.

**Engine spec.** Expose, per slide (e.g. in `Stats` or a query API), the **resolved** canvas /
surface / primary-text colors the engine rendered with — so a caller can compute true contrast
against the actual background, and apply large-text thresholds itself. No contrast *logic* in the
engine (that stays the caller's); just expose what was resolved.

**Acceptance.** After `Render`, a caller can read the resolved background + text colors for each
slide (including the dark-variant palette); documented; deterministic.

---

## How to pick up a unit

Pick one requirement, implement it under `scene/` (read `scene/render.go` + the relevant
`render_*.go`), add its golden + determinism test, run the gate, open a PR. Prefer the priority
order (R1 → R2 unblock the most product value and R1 unblocks truthful overflow). Keep every
change additive and deterministic per the invariants above. The caller (Deckard) will pick up each
capability by bumping its pptx-go dependency and wiring the new fields into its own contract +
guidance — you don't need to touch that side.
