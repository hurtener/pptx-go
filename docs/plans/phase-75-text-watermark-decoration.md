# Phase 75 — text/number watermark decoration

**Subsystem:** `scene` (Layer 2 — decoration)
**RFC sections:** §14.2 (decoration), §8.4 (rich text), §10.1 (backward-compat)
**Deps:** Phase 73 (Decoration.Color, D-107); brief 58.
**Status:** Done

---

## 1. Goal

Add a `DecorationText` kind so a slide can carry an oversized, low-opacity ghost
number/word behind the body (the reference's structural index numbers),
byte-identical when unused.

## 2. Why now

Wave 13 finish work (`docs/plans/README.md`); the reference's big faint
"01/02/03" device has no engine path today (a card watermark overflows). Engine
req R13.9 (MED · engine; D-059).

## 3. RFC sections implemented

- `RFC §14.2` — a text mode for the decoration node.
- `RFC §8.4` — reuses the rich-text run placement.
- `RFC §10.1` — a new kind appended last → existing decorations byte-identical.

## 4. Brief findings incorporated

- `docs/research/58-text-watermark-decoration.md` — *"the `Card.Watermark`
  pattern is the mechanism"* → a `TypeDisplay` run at `TokenColorAlpha(role,
  alpha)` in a text frame, lifted to a decoration.
- `58` — *"a new `DecorationKind` appended last"* → `DecorationText` after
  `DecorationAsset`.
- `58` — *"size via `RunStyle.FontScale` (>1 grows)"* → `FontScale = targetPt /
  displaySize`; default `targetPt` = box height in points (fill-the-box).
- `58` — *"reuse `Decoration.Color` + `Opacity`; wiring minimal"* → native (not
  asset), `nodeUsesAssets` stays false, safe-area-exempt, default
  `LayerBackground`.

## 5. Findings I'm departing from

none.

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-054` — the `Card.Watermark` text-alpha pattern.
- `D-107` — `Decoration.Color` (the watermark color role).
- `D-074` — `RunStyle.FontScale` (size multiplier; >1 grows).
- `D-109` (new) — files the `DecorationText` watermark.

## 7. Architecture

`DecorationText` joins `DecorationKind` (last). `Decoration` gains `Text string`
and `FontSize float64` (points). `renderDecoration` `case DecorationText` draws
one centered `TypeDisplay` run in a text frame at the decoration box, colored
`TokenColorAlpha(role, opacityAlpha)` (role = `Decoration.Color`, nil =
`ColorAccent`), sized via `FontScale = targetPt / displaySize` (targetPt =
`FontSize` or, when 0, the box height in points). Native, deterministic, no new
OOXML.

```text
Decoration{Kind: DecorationText, Text: "03", Color: &grey, Opacity: 0.08,
           Anchor: AnchorTopRight, Size: {6in,6in}}
  → text frame at box; run "03" @TypeDisplay, FontScale=boxH_pt/displaySize,
    Color=TokenColorAlpha(grey, ~8%) → one big faint glyph behind the body
```

## 8. Files added or changed

```text
scene/nodes.go                  # CHANGED — DecorationText kind + Decoration.Text / FontSize
scene/render_decoration.go      # CHANGED — renderDecoration text-watermark case
scene/validate.go               # CHANGED — DecorationText requires Text
scene/render_decoration_test.go # CHANGED — text watermark run + alpha; byte-identical unused; determinism
scripts/smoke/phase-75.sh       # NEW — phase smoke
docs/research/58-text-watermark-decoration.md  # NEW — brief
docs/research/INDEX.md          # CHANGED — registers brief 58
docs/plans/phase-75-text-watermark-decoration.md  # NEW — this plan
docs/plans/README.md            # CHANGED — Phase 75 detail
docs/design/THEME.md            # CHANGED — text watermark mechanism note
docs/glossary.md                # CHANGED — text watermark decoration term
docs/decisions.md               # CHANGED — adds D-109
docs/site/reference/scene.md    # CHANGED — DecorationText kind + fields
skills/compose-a-scene/SKILL.md # CHANGED — Decoration text watermark
```

## 9. Public API surface

```go
// scene
const DecorationText DecorationKind = … // appended after DecorationAsset; a text watermark
// Decoration gains: Text string; FontSize float64 // points; 0 = box-height default
```

No prior surface breaks (append-only kind + additive fields).

## 10. Risks

- **R1 — byte-identity drift.** **Mitigation:** existing `DecorationPreset`/
  `DecorationAsset` paths untouched; a new kind only adds a branch; covered by
  the curated-ornament byte-identity test.
- **R2 — non-deterministic scale.** **Mitigation:** integer-EMU box height →
  points → `FontScale` truncated through the `@sz` 1/100-pt path; a determinism
  guard pins it.

## 11. Acceptance criteria

1. A `DecorationText` watermark renders one text run carrying the text and a low `<a:alpha>` (from `Opacity`) at the decoration box behind the body.
2. A `DecorationText` with empty `Text` fails Stage-1 validation.
3. Decorations not using the text kind are byte-identical to the pre-Phase-75 build; a text watermark re-renders deterministically.
4. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default |

## 13. Smoke check

`scripts/smoke/phase-75.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `DecorationText` kind + `Decoration.Text`/`FontSize` fields.
3. `OK:` `renderDecoration` text case.
4. `OK:` text-watermark render test passes.
5. `OK:` empty-text validation + determinism tests pass.

## 14. Tests

- **Black-box (`scene_test`):** a text watermark emits the text + a low alpha;
  empty `Text` fails validation; curated decorations byte-identical; determinism.
- **Integration / Fuzz:** no.
