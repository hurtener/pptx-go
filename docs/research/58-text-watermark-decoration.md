# Brief 58 — Text/number watermark decoration (R13.9)

> Informs Phase 75 (Wave 13). Engine req R13.9
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059).

## 1. Motivating phase

The reference uses oversized translucent index numbers as a structural device
(big faint "01/02/03" behind way-cards, chapter numbers). Deckard's `Decoration`
node is preset-or-asset only — no text mode — so a slide-level ghost number must
be faked via a card watermark (which overflowed badly in the recreation). Phase
75 adds a slide-level text watermark decoration: one large, low-opacity glyph
behind the body.

## 2. Subsystem / files

- `scene/nodes.go` — `DecorationKind` + the `Decoration` struct.
- `scene/render_decoration.go` — `renderDecoration` switch.
- `scene/validate.go` — decoration Stage-1 validation.
- `scene/render_card.go` — the `Card.Watermark` pattern to reuse:
  `ps.AddTextFrame(box).Anchor(...)` → `AddRun(text, RunStyle{TypeRole,
  Color: TokenColorAlpha(role, alpha)})`.

## 3. Findings

- **The watermark text pattern already exists.** `Card.Watermark` (D-054) draws
  a `TypeDisplay` run at `TokenColorAlpha(role, ~13%)` in a text frame — exactly
  the slide-level mechanism, lifted to a `Decoration`.
- **A new `DecorationKind` is the clean shape.** Add `DecorationText` after
  `DecorationAsset` (last → existing kinds byte-identical). It carries a `Text`
  string and an optional `FontSize` (points; 0 = a box-derived "fill the box"
  default). It reuses `Decoration.Color` (D-107, nil = `ColorAccent`) and
  `Opacity` (→ alpha) — a faint colored glyph.
- **Size via `RunStyle.FontScale`.** `RunStyle` has no absolute font size, but
  `FontScale` multiplies the role size (and >1 grows — `text_layout.go` applies
  `spec.Size * FontScale` for any `FontScale > 0`). So map the target points to
  `FontScale = pt / ResolveType(TypeDisplay).Size` on a `TypeDisplay` run. The
  size round-trips via `Run.FontSize`. Default `pt` = box height in points
  (`box.H / Pt(1)`) — a deterministic "fill the box" auto-size.
- **Wiring is minimal.** `DecorationText` is native (not `DecorationAsset`), so
  `nodeUsesAssets` stays false (parallel-safe); it is a `Decoration`, so it
  inherits the renderNode safe-area exemption (decorations bleed by design) and
  the layer z-order split (default `LayerBackground` → behind body). The
  ornament-name validation already early-returns for non-`DecorationPreset`.
- **No overflow logic.** It is decorative — one centered run, no wrap handling
  needed (a short number/word); the text frame clips. Deterministic integer math
  for the scale (round to a fixed precision via the existing `@sz` 1/100-pt path).
- **No OOXML / `restorenamespaces` change.** Same text-run XML the card
  watermark already emits.

## 4. Recommendations

- Add `DecorationText` kind + `Decoration.Text string` + `Decoration.FontSize
  float64` (points; 0 = box-height default).
- `renderDecoration` `case DecorationText`: a text frame at the decoration box,
  one centered `TypeDisplay` run, `Color = TokenColorAlpha(role, opacityAlpha)`,
  `FontScale = targetPt / displaySize`.
- `validate.go`: `DecorationText` requires `Text != ""`.
- Tests: a text watermark emits a run with the text + a low `<a:alpha>`; unused
  decorations byte-identical; determinism. THEME.md note, glossary,
  compose-a-scene skill, docs/site. D-109.

## 5. Open questions

- True "fill the box" font fitting (measure glyph vs box) is out of scope —
  the box-height heuristic is deterministic and good enough for a decorative
  glyph; a precise fit can follow if needed.
- Rotation of a watermark glyph (diagonal "DRAFT") → the builder has no
  rotated-text primitive (same limit as the ribbon diagonal, D-098); deferred.
