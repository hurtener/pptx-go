# Brief 30 — list-bullet-indent-density

**Subsystem:** scene — Layer 2 renderer (+ a pptx paragraph option)
**Authored:** 2026-06-22
**Motivating phase:** Phase 47 — list bullet indent density (R10.9, MED · engine)

## 1. Question

`Paragraph.Bullet` hard-codes a `0.5"` hanging indent (`MarL = 457200`,
`Indent = −457200`), so every list marker sits a wide, fixed gap from its text —
lists read loose and sparse (the recreation's "Understand/Operate/Execute" and the
slide-3 checklist show a very wide marker-to-text gap). How can a caller opt a
list into a tighter, consistent marker-to-text offset — across all items and
levels — opt-in and byte-identical by default?

## 2. Prior art surveyed

- `pptx/text.go` `Paragraph.Bullet(kind)` — sets the bullet glyph and, for any
  non-none bullet, the fixed `MarL = 457200` / `Indent = −457200` hanging indent.
- `pptx/text.go` `AddParagraph(opts)` — applies `opts.Level`, `opts.Bullet`,
  `opts.LineHeight`. `LineHeight` (D-061) is the precedent for a per-paragraph
  layout metric that emits an `a:pPr` attribute and is byte-identical at its zero
  value.
- `scene/render_leaves.go` `renderList` — one paragraph per item with
  `ParagraphOpts{Bullet, Level, LineHeight}`.
- `scene/nodes.go` `List` / `ListItem` — the IR; `ListKind` + per-item `Level`.
- DECKARD R10.9 spec: expose a list density / indent control on the `List` node (or
  a tighter default hanging-indent + marker gap) plumbed through `renderList`'s
  `ParagraphOpts` to the builder's bullet indent metrics; a small set of
  deterministic presets (tight/normal); default preserves current output; no
  measurement; deterministic.

## 3. Findings

- The builder seam is a **per-paragraph bullet-indent override** on
  `ParagraphOpts` — `BulletIndent EMU` (0 = the current default). When set,
  `AddParagraph` overrides the `MarL`/`Indent` the `Bullet` method assigned, so the
  marker-to-text offset becomes `BulletIndent` uniformly. Mirrors `LineHeight`:
  emits an `a:pPr` value, byte-identical at zero.
- **Scene presets, not arbitrary values.** A `List.Indent ListIndent` enum
  (`IndentNormal` = 0 / default, `IndentTight`) maps to a pinned `BulletIndent`:
  `IndentTight → In(0.25)` (`228600`), about half the `0.5"` default. Pinned and
  deterministic — the spec's "small set of presets". `IndentNormal` passes 0, so
  the builder keeps its default → byte-identical.
- **Consistent across items and levels.** The override is applied per paragraph
  with the same value for every item, so the marker-to-text offset is uniform
  regardless of `Level` (the existing code already uses a level-independent `MarL`,
  so this preserves that and just tightens it).
- **No token.** The indent is a layout metric, not a theme color/spacing token; the
  presets are pinned (like the existing `0.5"` literal). P2's token-default rule is
  about visual *style* properties (color/typography/spacing/radius/elevation); a
  bullet hanging indent preset is a layout mechanism — documented in `THEME.md` as
  a pinned preset (not a token) for traceability.
- **Round-trip.** The emitted `MarL`/`Indent` attributes are preserved in the slide
  XML (like `LineHeight`'s `spcPct`); a test asserts the tight values emit and the
  default is unchanged.

## 4. Recommendations

1. **pptx:** add `ParagraphOpts.BulletIndent EMU`; in `AddParagraph`, when a bullet
   is set and `BulletIndent > 0`, set `MarL = BulletIndent`, `Indent =
   −BulletIndent`.
2. **scene:** add `ListIndent` (`IndentNormal`, `IndentTight`) + `List.Indent`;
   `renderList` passes `BulletIndent = bulletIndent(v.Indent)` (0 for normal,
   `listTightIndent` for tight).
3. Tests: the builder emits the tight `marL`/`indent` and the default is
   byte-identical; a tight `List` renders a smaller, consistent offset than a
   normal one; determinism guard; smoke `phase-47.sh`.

## 5. Open questions

- **Per-level indent nesting** (deeper levels indent further) — out of scope; the
  current renderer uses a level-independent `MarL` and R10.9 only tightens the
  hanging gap. A future req could add per-level steps.
- **A continuous indent value** (vs presets) — the `ParagraphOpts.BulletIndent`
  seam already accepts any EMU; the scene exposes presets per the spec, and a
  caller using `pptx` directly can pass any value.
