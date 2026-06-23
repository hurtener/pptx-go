# Brief 45 — prim-in-card-checklist-fill

**Subsystem:** scene — Layer 2 renderer (new IR leaf node)
**Authored:** 2026-06-23
**Motivating phase:** Phase 62 — prim-in-card-checklist-fill (R12.2, CRITICAL · engine)

## 1. Question

The "what you get" feature list is the heart of every offer / pricing / comparison
card. The recreation renders `List{Kind: ListChecklist}` via `pptx.BulletCheckbox`,
which draws an **empty white square** (the `ListItem.Checked` bool is never consulted),
hangs on a broken bullet indent, sits cramped at native body size, and cannot reflow
into columns or fill the card's lower half. What primitive gives a dense, self-
distributing feature checklist with *filled* status glyphs, optional column reflow, and
a fill-to-card mode — and how does it reuse the engine's existing mechanisms?

## 2. Prior art surveyed

- **Curated icon registry** — `assets/icons/{check,x,dot}.svg` are single-path *filled*
  glyphs; `r.cfg.icons.Lookup(name)` → `ps.AddIcon(svg, box, WithFill(color))` renders a
  native custGeom shape (media-free). This is exactly the "true filled check/cross
  (not a font checkbox)" the spec demands — no new asset, no new builder API.
- **`scene/render_button.go` (Phase 61, D-094)** — the just-landed content-fit-pill
  composer establishes the pattern: pinned layout metrics, `r.cfg.icons` glyph fill with
  a label-color token, and the new-node wiring checklist (incl. `walkIconRefs` for an
  icon-bearing leaf).
- **`scene/render.go alignedStackIn` VAlignJustify / VAlignFill** — the inter-element
  slack-distribution math (`slack/(n-1)` for justify; proportional fill) the checklist's
  `Fill` mode mirrors *within its own box* for inter-row spacing.
- **`scene/render_leaves.go renderList`** — the bullet-list composer the checklist
  replaces for the feature-list use case; `listTightIndent` (D-078/D-090) shows the
  hanging-indent calibration approach (a marker-to-text offset proportional to body
  size). The checklist computes its hanging indent from the *glyph width* instead.
- **`scene/metrics.go wrappedLines`** — per-cell line-count estimation at the column
  text width, for per-row heights (content-aware, like every Wave-10 node).

## 3. Findings

- **The glyph is a curated icon, not a font checkbox.** `CheckDone` → `check`, `CheckNo`
  → `x`, `CheckNeutral` → `dot`, each rendered via `AddIcon` with a token fill — a real
  filled custGeom glyph, fixing the empty-square bug by construction. A per-item `Icon`
  overrides the name (Stage-1 validated through `walkIconRefs`).
- **Glyph color: per-state default, with an optional `GlyphTone` override.** `CheckDone`
  defaults to `ColorAccent`, `CheckNo`/`CheckNeutral` to a muted text token. Because
  `ColorRole`'s zero value is a real color (`ColorCanvas`), the override is a
  `*ColorRole` (nil = per-state default) — the same call D-054 made for
  `HeaderFill`/`StatusDot`. A §4.3 deviation from the spec's value `ColorRole`,
  documented in the decision.
- **Hanging indent = glyph width + gap.** Each row is `[glyph | text]`; the text frame
  is offset right by `glyphSz + glyphGap`, so wrapped lines align under the text, never
  under the glyph (no PPTX auto-bullet, no broken indent).
- **Columns reflow row-major.** `cols = clamp(Columns, 1, 3)`; item `i` → `row = i/cols`,
  `col = i%cols`; `rows = ceil(n/cols)`. Each column shares `box.W` with a pinned gap.
  Pure integer indexing — no map iteration, deterministic.
- **Fill distributes inter-row slack like VAlignJustify.** Per-row heights come from the
  tallest cell's wrapped-line count; `Fill` spreads `box.H − Σrows` across the
  `rows − 1` gaps so the last row meets the box bottom. Off (default) top-aligns the
  rows at a pinned gap. For a short list to *receive* a tall box (fill the card), the
  node must be flexible — add `Checklist` to `isFlexible` so a `VAlignFill`/`BodyVAlign`
  parent grows it; a non-Fill checklist in a grown box simply top-aligns (harmless).
- **All-pinned layout metrics, token colors.** Glyph size, glyph gap, column gap, row
  gap, and per-line height are pinned EMU (layout, not visual) — like `buttonMetrics`;
  the glyph colors are tokens (P2). Additive: a deck with no `Checklist` is
  byte-identical (the existing `List`/`BulletCheckbox` path is untouched).

## 4. Recommendation

Add `KindChecklist` + a `Checklist` leaf node `{Items []ChecklistItem{Text RichText;
State CheckState; Icon string}; Columns int; GlyphTone *ColorRole; Fill bool}` and a
`scene/render_checklist.go` composer: row-major column reflow, a per-row content-aware
height, a `[glyph | text]` layout with the hanging indent from the glyph width, the
glyph as a curated `check`/`x`/`dot` custGeom filled with the per-state (or `GlyphTone`)
token color, and a `Fill` mode that distributes inter-row slack so a short list spans
the box. Full new-node wiring (policy native, validate non-empty items + columns 0..3 +
state range, `renderNode` dispatch + `preferredHeight` + `isFlexible` true +
`nodeUsesAssets` false, `walkIconRefs case Checklist` for per-item icon overrides,
catalog 23 → 24, integration kind-range loop → `KindChecklist`). Extend the R11.12
adversarial fixture with a 2-column filled checklist under hostile content.
