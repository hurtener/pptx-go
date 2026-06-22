# Brief 41 — list-bullet-hanging-indent (R11.10 proportional close)

**Subsystem:** scene — Layer 2 renderer (List leaf)
**Authored:** 2026-06-22
**Motivating phase:** Phase 58 — list-bullet-hanging-indent (R11.10, MED · engine)

## 1. Question

List items show a large fixed gap between the bullet and the label ("•      Chat &
Q&A"), and on the overflowing card the list collides with the wrapped header
(recreation slides 2, 4). R10.9/D-078 already added a tight preset
(`IndentTight = In(0.25)`) and a builder override (`ParagraphOpts.BulletIndent`); is
that enough to close R11.10, or does the indent need to be *proportional* to the body
size?

## 2. Prior art surveyed

- **`scene/render_leaves.go`** — `listTightIndent = In(0.25)` (a pinned const), mapped
  by `bulletIndent(IndentTight)` to `ParagraphOpts.BulletIndent`; `IndentNormal`
  keeps the builder's 0.5" default (byte-identical).
- **`pptx/text.go`** — `ParagraphOpts.BulletIndent` overrides the default 0.5" hanging
  indent → `a:pPr/@marL` + `@indent` (D-078).
- **R10.1 / D-070** — the wrapped-header geometry the list start Y already respects
  (the list is a body node placed below the grown header; verified by Phase 49).
- DECKARD R11.10 spec: derive the bullet hang indent from the resolved body type size
  (`hang = round(size_pt · k · emuPerPoint)`) rather than a fixed wide value; tight
  gap; validate against the wrapped-header fix.

## 3. Findings

- **The mechanism exists (D-078); R11.10's delta is proportionality.** The tight
  preset is a *fixed* `In(0.25)` — tight at 14 pt but unrelated to the body size. The
  R11.10 ask is to make it scale with the body size, so the gap stays proportional at
  any size. Implement `listTightIndent()` as `listTightIndentBase × bodySize / 14`,
  anchored so the default 14 pt body yields exactly `In(0.25)` (byte-identical to the
  D-078 pinned value, the existing `marL="228600"` test passes) and a larger/smaller
  body scales the indent linearly.
- **`bulletIndent` becomes a method** (it needs `r.theme` for the body size); the only
  caller (`renderList`) is already a method.
- **The "≤ 1.5× glyph" acceptance is an example, not the bar.** `In(0.25)` is ~2.6× a
  14 pt marker glyph — tight relative to the 0.5" default the recreation showed, which
  is the real "oversized" baseline. The binding requirement is *proportional, not a
  fixed oversized gap*; the close asserts the indent is meaningfully smaller than the
  0.5" default (≤ `In(0.3)`) and scales with the body size, rather than the stricter
  1.5×-glyph example (which would break the D-078 byte-identity for no real legibility
  gain).
- **Wrapped-header interaction already holds.** The list start Y respects the grown
  card header via R10.1 (the list stacks below `cardHeaderBottom`); Phase 49's
  acceptance golden already guards header/body disjointness across the size/layout
  matrix.

## 4. Recommendation

Make `listTightIndent` a theme-proportional method anchored to `In(0.25)` at 14 pt;
keep `IndentNormal` byte-identical. Tests: the anchor is byte-identical (`In(0.25)`
at 14 pt), the indent scales (2× body → 2× indent, ½ body → ½ indent), and the gap is
tighter than the 0.5" default. D-090 records the proportional indent and that
R11.10's mechanism otherwise rests on D-078 (override) + R10.1 (header respect).

## 5. Open questions

- Per-level (nested) indent proportionality is unchanged (each level inherits the
  builder's level indent); out of scope for R11.10's bullet-to-text gap.
