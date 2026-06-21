# Brief 15 — stat-node

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-21
**Motivating phase:** Phase 28 — Stat leaf node

## 1. Question

Pricing and metric slides use big-number stats (`$2,200`, `38%`) with a label and
an optional delta (`+12%`, colored by direction). The scene IR has no hero-number
node, so callers fake them with `Heading`s — losing the value/label/delta
structure and the directional color. How can the engine add a `Stat` leaf —
display-scale value + label + optional toned delta — and let a row of them inside
a `Grid` form a metric/pricing strip?

This is `DECKARD-PRODUCT-REQUIREMENTS.md` R6 (LOW).

## 2. Prior art surveyed

- **`Hero` / `renderHero`.** A multi-line text block (eyebrow + title + subtitle)
  in one anchored text frame — the exact idiom a stat reuses: value (big) +
  label (caption) + delta (small, colored) as three paragraphs in one frame.
- **`render_leaves.go` color idioms.** Caption/muted runs (`TextMuted`) and
  token run colors (a run's `Color` is any `pptx.Color`, resolved at write time)
  are established. A delta's directional color is just a token run color.
- **`ColorRole` palette.** The theme already carries `ColorSuccess`,
  `ColorError`, and `ColorWarning` — so an up/down/neutral delta maps to existing
  tokens (`ColorSuccess` / `ColorError` / `TextMuted`); **no new token** (P2).
- **`Grid` (container).** A `Grid` of leaf cells already lays children into a
  uniform strip; a row of `Stat`s in a `Grid` *is* the metric strip — no new
  container is needed, only the leaf.
- **The new-leaf wiring checklist.** A leaf node (unlike a container) touches
  `NodeKind`+`String`, `policyTable`, `validateNode`, `renderNode`,
  `preferredHeight`, and `nodeUsesAssets` — but **no** `walk*`/`isFlexible`
  recursion (a leaf has no children, and a number block doesn't stretch).
- **D-026 (engine, not product).** The caller supplies the value, label, delta,
  and tone; the engine renders them and picks only the mechanical tone→color
  mapping. It computes no metric and judges no number.

## 3. Findings

- **`Stat` is a leaf, not a container.** `Stat{Value, Label, Delta string,
  DeltaTone}` renders as native text — one anchored frame with up to three
  paragraphs. It carries no children and no `AssetID`, so its policy is `{}` and
  it never forces serial composition.
- **A `DeltaTone` enum with a `neutral` zero keeps the delta optional and the
  default inert.** `DeltaNeutral` (zero) → muted; `DeltaUp` → success (green);
  `DeltaDown` → error (red). The `Delta` string being empty omits the delta line
  entirely, so a stat with no delta is just value + label.
- **The metric strip is free.** A `Grid{Columns: N, Cells: [Stat, Stat, …]}`
  lays the stats into a uniform strip with no new code — the grid cell box drives
  each stat's size.
- **Validation is minimal.** A `Stat` with an empty `Value` is meaningless;
  Stage-1 requires `Value != ""`. (Label and Delta are optional.)
- **Not flexible.** A `Stat` is a fixed-size number block; it stays at its
  preferred height under `VAlignFill` (stretching a number is meaningless), so it
  is **not** added to `isFlexible` — but a `Grid` of stats still grows as a
  container, taking the strip with it.

## 4. Recommendations

1. Add `Stat` + `DeltaTone` (`DeltaNeutral`/`DeltaUp`/`DeltaDown`) + `KindStat`
   (+ `String`), a `policyTable` entry (`{}`), and a `validateNode` case
   (`Value` non-empty).
2. `renderStat`: one anchored text frame — `Value` at `TypeDisplay` (bold),
   `Label` at `TypeCaption` (muted), and, when `Delta != ""`, a `TypeBodySmall`
   delta run colored by `deltaToneColor(tone)` (`ColorSuccess` / `ColorError` /
   `TextMuted`).
3. Wire the leaf through `renderNode`, `preferredHeight` (a fixed block height),
   and `nodeUsesAssets` (false) — but **not** `isFlexible` or the `walk*`
   recursions.
4. Extend the catalog (21 → 22 kinds) and the round-trip `everyNodeScene` /
   kind-range loop; exercise the metric strip via a `Grid` of `Stat`s.

## 5. Open questions

- **Numeric formatting.** The engine renders `Value`/`Delta` verbatim — it does
  not format numbers, currency, or signs (that is the caller's, D-026). A
  formatting helper is out of scope.
- **Delta arrow glyph.** The delta is text only; prepending a ▲/▼ glyph by tone
  is a plausible refinement, deferred — the caller can include it in the `Delta`
  string.
- **Value auto-fit.** A very long `Value` may overflow a narrow cell. Auto-fit /
  shrink is deferred; the cell width is the caller's layout choice (grid columns).
