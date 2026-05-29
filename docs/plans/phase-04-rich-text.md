# Phase 04 — Rich text model

**Subsystem:** pptx (text)
**RFC sections:** §8.4, §8.8 (notes ← TextFrame), §9
**Deps:** Phase 03 (builder spine — shapes, media, notes, the one-emission
path D-032, and the token `Color` model D-033).
**Status:** In progress

---

## 1. Goal

A shared rich-text model — `TextFrame` → `Paragraph` → `Run` with token-typed
styling, bullets, hyperlinks, inline code, and auto-fit — that every text-
bearing shape (and speaker notes) builds on and that round-trips losslessly.

## 2. Why now

The builder spine (Phase 03) emits shapes, media, sections, and a plain-text
notes part, but text is still a single unstyled run (`Slide.AddTextBox`). Rich
text is the substrate the scene renderer composes (`scene` maps its `RichText`
`[]TextRun` directly onto a builder `Paragraph`, RFC §9), so it must exist
before Wave 2 (Phase 05 depends on Phase 04). It also closes the speaker-notes
seam D-034 left open: D-022/RFC §8.8 specify `slide.SpeakerNotes() *TextFrame`,
deferred in Phase 03 to a plain-text setter because `TextFrame` did not yet
exist — this phase delivers it.

## 3. RFC sections implemented

- `RFC §8.4` — the `TextFrame`/`Paragraph`/`Run`/`RunStyle` API: paragraphs,
  styled runs, breaks, hyperlinks, bullets, indent, alignment, anchor,
  margins, and the `AutoFit` modes (`AutoFitNone`/`AutoFitNormal`/
  `AutoFitShape`).
- `RFC §9` — the shared rich-text model: token-based run colors honor theme
  swaps; `scene` will re-export and map `[]TextRun` onto a `Paragraph`. (This
  phase ships the `pptx` half; the `scene` re-export lands with the scene
  scaffold, Phase 05.)
- `RFC §8.8` — speaker notes become a `TextFrame` (the `pptx` notes part now
  carries rich text). Closes the D-034 deferral.

## 4. Brief findings incorporated

No informing brief — the rich-text API is specified directly in RFC §8.4/§9
and the decisions log (D-012/D-013/D-022/D-033). `docs/research/INDEX.md`
lists "rich-text auto-fit modes in OOXML" only as a *candidate* brief; the RFC
already fixes the three modes, so no prior-art investigation is needed (same
posture as Phase 03, RFC §17.1).

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-012`/`D-033` — `Color` is the sealed token/literal interface; run colors
  reuse it. Run colors resolve against the **active theme at apply time** (the
  same call-time resolution Chunk B uses for shape fills), so a theme set
  before authoring re-colors the text. (A fully write-time-lazy pass over text
  is a possible future refinement, shared with the fills caveat — D-033.)
- `D-013` — inline code is a run-level flag (`RunStyle.Code = true`), not a
  separate node: monospace typeface + a subtle background tint. This phase
  implements it.
- `D-022` — speaker notes are V1 and round-trip; this phase upgrades them to a
  `TextFrame`.
- `D-034` — Chunk C shipped notes as a plain-text setter with the `*TextFrame`
  accessor "deferred to the rich-text phase" — this is that phase.

## 7. Architecture

The public model is three nested builders over the existing `internal/ooxml/
slide` text wire types (`XTextBody`/`XTextParagraph`/`XTextRun`/`XTextProperties`/
`XBodyPr`), which this phase extends to carry the missing OOXML the model
needs. `RestoreNamespaces` already maps the text element set
(`p`/`pPr`/`r`/`rPr`/`t`/`br`/`hlinkClick`/`buChar`/`buAutoNum`/`buNone`/`latin`
→ `a:`), so the codec change is struct fields, not namespace-table work.

```text
TextFrame  (shape-level: body props, autofit, anchor, margins)
  └─ Paragraph  (alignment, indent level, bullet)
       └─ Run   (text, RunStyle: type role, color, bold/italic/underline/
                 strike/baseline, code, hyperlink)

pptx.TextFrame ──builds──▶ *slide.XTextBody   (P3: wire types stay internal)
```

A `TextFrame` is attached to a shape's `txBody`. `Slide.AddTextFrame(box)`
creates a text-box shape and returns its `*TextFrame`; `Slide.SpeakerNotes()`
returns the notes slide's `*TextFrame`. Run colors and the inline-code tint
resolve against `Slide`'s active theme when the run is added (the Chunk B
pattern). Hyperlink runs add an external relationship to the slide's
relationship set (mirrored to the package part by `syncSlides`, the Phase 03 C
seam); the URL is emitted verbatim (§7 — no fetch, no validation).

### Chunking

- **A — core model + styles + frame properties.** `TextFrame`/`Paragraph`/
  `Run`/`RunStyle`; bold/italic/underline/strike/baseline/size/typeface/color
  (token + literal); `AddBreak`; paragraph `Align`/`Indent`; frame `AutoFit`/
  `Anchor`/`Margins`. `Slide.AddTextFrame`. Wire-type extensions + round-trip
  goldens. `AddTextBox` keeps working (a one-run convenience).
- **B — bullets, inline code, hyperlinks, notes-as-TextFrame.** `BulletKind`
  (`none`/`disc`/`number`/`checkbox`); `RunStyle.Code` (D-013); `AddHyperlink`
  with relationship wiring; `Slide.SpeakerNotes() *TextFrame` replacing the
  Chunk C text setter (a pre-V1 break, CHANGELOG-noted).

## 8. Files added or changed

```text
pptx/text.go                       # NEW (A) — TextFrame, Paragraph, Run, RunStyle, enums
pptx/text_layout.go                # NEW (A) — autofit/anchor/margins/align/indent helpers
pptx/text_hyperlink.go             # NEW (B) — AddHyperlink + relationship wiring
pptx/slide.go                      # CHANGED (A) — AddTextFrame; AddTextBox composes TextFrame
pptx/notes.go                      # CHANGED (B) — SpeakerNotes() *TextFrame
internal/ooxml/slide/slide_types.go# CHANGED (A,B) — rPr fill/typeface/u/strike/baseline,
                                   #   hlinkClick; pPr marL/indent/algn/lvl + bullets; br; xml:space
docs/design/THEME.md               # CHANGED (A,B) — type role + inline-code tint token entries
docs/glossary.md                   # CHANGED — TextFrame/Paragraph/Run/RunStyle/BulletKind/…
docs/decisions.md                  # CHANGED (B) — notes-TextFrame supersede note (if needed)
scripts/smoke/phase-04.sh          # NEW — phase smoke
```

## 9. Public API surface

```go
// pptx (text.go)
type TextFrame struct{ ... }
func (s *Slide) AddTextFrame(box Box) *TextFrame
func (tf *TextFrame) AddParagraph(opts ParagraphOpts) *Paragraph
func (tf *TextFrame) AutoFit(mode AutoFitMode) *TextFrame
func (tf *TextFrame) Anchor(v TextAnchor) *TextFrame
func (tf *TextFrame) Margins(top, right, bottom, left EMU) *TextFrame

type Paragraph struct{ ... }
func (p *Paragraph) AddRun(text string, style RunStyle) *Run
func (p *Paragraph) AddBreak()
func (p *Paragraph) AddHyperlink(text, target string, style RunStyle) *Run
func (p *Paragraph) Bullet(kind BulletKind) *Paragraph
func (p *Paragraph) Indent(level int) *Paragraph
func (p *Paragraph) Align(a Alignment) *Paragraph

type Run struct{ ... }
type RunStyle struct {
    TypeRole    TypeRole
    Color       Color
    Bold, Italic bool
    Underline   Underline
    Strike      Strike
    BaselineRel BaselineShift
    Code        bool
}
type (AutoFitMode, TextAnchor, Alignment, BulletKind, Underline, Strike, BaselineShift int)
type ParagraphOpts struct { Align Alignment; Level int; Bullet BulletKind }

// pptx (notes.go) — supersedes Phase 03's SpeakerNotes(text string)
func (s *Slide) SpeakerNotes() *TextFrame
```

`SpeakerNotes`'s signature changes from `(text string)` to `() *TextFrame`
(pre-V1, v0.x). A `SetSpeakerNotes(text string)` one-line convenience is kept
for the plain-text path. The break is documented in `CHANGELOG.md`.

## 10. Risks

- **R1 — run-color resolution timing.** Tokens must reflect the theme but the
  codec stores a resolved sRGB. *Mitigation:* resolve at `AddRun` against the
  slide's active theme (the Chunk B fill pattern); a swap before authoring
  re-colors. Documented; consistent with shapes.
- **R2 — inline-code styling has no single OOXML element.** *Mitigation:* map
  `Code=true` to a monospace `latin` typeface + a subtle `highlight` tint
  sourced from the theme; pin it with a golden and a THEME.md token.
- **R3 — whitespace fidelity.** PowerPoint collapses runs of spaces without
  `xml:space="preserve"`. *Mitigation:* emit `<a:t xml:space="preserve">` and
  assert it in a golden (the D-014 rationale for rastering full code blocks
  still stands — this is inline text only).
- **R4 — SpeakerNotes signature break.** *Mitigation:* pre-V1; keep a
  `SetSpeakerNotes(text)` convenience; CHANGELOG + smoke updated in the same
  PR.

## 11. Acceptance criteria

1. A `TextFrame` with multiple paragraphs and styled runs (bold/italic/
   underline/strike/colored/sized) round-trips losslessly through `pptx.Open`.
2. An inline-code run (`RunStyle.Code = true`) renders with a monospace
   typeface + a subtle background tint sourced from the default theme.
3. A hyperlinked run carries its URL through the relationships layer
   (`hlinkClick r:id` → an external relationship), and the deck stays
   conformant.
4. Bulleted, numbered, and checklist paragraphs emit the correct bullet
   (`buChar`/`buAutoNum`/`buChar` with a check glyph) and round-trip.
5. A run color expressed as a token re-colors when the theme is swapped before
   authoring (token, not literal).
6. `slide.SpeakerNotes()` returns a `TextFrame`; notes authored through it
   round-trip and the deck passes the conformance gate.
7. `make build`/`test`/`lint`/`coverage`/`preflight`/`check-mirror` green;
   prior phases' smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` (text files) | 85% | default for new pptx surface |
| `internal/ooxml/slide` (text wire types) | 85% | codec band (round-trip goldens) |

The `pptx` package coverage is measured by the external `test/pptx` suite
(D-029); new text behavior is covered there + codec goldens in
`internal/ooxml/slide`.

## 13. Smoke check

`scripts/smoke/phase-04.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` text round-trip golden (styled runs) passes.
3. `OK:` inline-code styling test passes.
4. `OK:` hyperlink run + relationship test passes; conformance gate green.
5. `OK:` bullet/number/checklist test passes.
6. `OK:` token run-color theme-swap test passes.
7. `OK:` `SpeakerNotes() *TextFrame` round-trip + conformance passes.

`SKIP` each criterion until its chunk lands; `FAIL = 0` always.

## 14. Tests

- **Unit:** `pptx` (text builders) and `internal/ooxml/slide` (codec).
- **Round-trip golden:** yes — every styled run / paragraph / bullet /
  hyperlink / notes frame writes → reads → asserts model equality.
- **Integration** (`test/integration/`): the conformance gate on a deck with a
  hyperlinked, bulleted, notes-bearing text frame (closes the text→rels seam
  with the Phase 03 C relationship mirroring).
- **Fuzz:** none (no new parse entry point; slide parse is exercised by
  round-trip).
- **Benchmark:** optional — a styled-paragraph build/emit micro-bench.

## 15. Vocabulary added

- `TextFrame` — shape-level rich-text container.
- `Paragraph` — a line block within a `TextFrame`.
- `Run` — a styled text span within a `Paragraph`.
- `RunStyle` — token-typed run styling (type role, color, weight, …).
- `BulletKind` — `none`/`disc`/`number`/`checkbox`.
- `Alignment` / `TextAnchor` / `AutoFitMode` — paragraph alignment, frame
  vertical anchor, frame auto-fit behavior.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-04.sh` reports `OK ≥ 7` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (if any).
