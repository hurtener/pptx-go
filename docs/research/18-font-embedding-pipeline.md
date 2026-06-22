# Brief 18 — font-embedding-pipeline

**Subsystem:** pptx — Layer 1 builder (font embedding)
**Authored:** 2026-06-22
**Motivating phase:** Phase 35 — automatic font-embedding pass

## 1. Question

A brand deck's identity is its type. When a theme names a non-system display
face (e.g. a serif like Playfair Display) the runs are emitted carrying that
`a:latin` typeface, but unless the *bytes* are embedded in the package
PowerPoint and any rasterizer substitute a host sans — the brand face never
appears. The engine already has the embedding **mechanism**
(`pptx/fonts.go`: the `FontSource` interface + `Presentation.EmbedFont` +
`WithFontSource`), but nothing walks the deck and embeds the faces it actually
uses; the caller must enumerate and `EmbedFont` every `(family, style, weight)`
by hand. How can the engine, opt-in, collect the distinct faces a deck uses and
embed them — additively, deterministically, byte-identical when off, warn-don't-
fail on a missing face, and idempotent against manual `EmbedFont` calls?

This is `DECKARD-PRODUCT-REQUIREMENTS.md` R9.1 (`font-embedding-pipeline`,
CRITICAL · both). Per D-059 pptx-go implements the **engine half**: collect the
used faces + call `Presentation.EmbedFont`; the `FontProvider`/soul side
(embed:// providers, bundled brand bytes) is Deckard's.

## 2. Prior art surveyed

- **`pptx/fonts.go::EmbedFont(name, style, weight)`.** Already resolves bytes via
  the registered `FontSource`, writes a `/ppt/fonts/fontN.fntdata` part, relates
  it to presentation.xml, and records the face in `embeddedFontLst` via
  `presentationPart.AddEmbeddedFont(name, StyleFor(weight, italic), rid)`. It
  takes `p.mu.Lock()` — so it cannot be called from inside `prepareForWrite`
  (which already holds the lock). A lock-free inner body is needed.
- **`prepareForWrite`** (`pptx/presentation.go`) is the shared body of every
  write path (Save / Write / WriteToBytes / SaveStream). The in-memory slide
  builder structs (`s.part.SpTree()`) already hold every emitted run before sync,
  so the walk can run there, before `syncPresentationPart` serializes the
  `embeddedFontLst`.
- **`internal/ooxml/slide`** run model: `XSpTree.Children` are `*XSp` /
  `*XPicture` / `*XGraphicFrame`; text lives in `XSp.TextBody` and in
  `XGraphicFrame` table cells (`XTableCell.TextBody`) — exactly the set
  `XSpTree.DroppedDescendants` already walks. A run's resolved face is
  `XTextRun.TextProperties.Latin.Typeface` + `.Bold == "1"` + `.Italic == "1"`.
- **The OOXML weight gotcha.** An emitted `a:rPr` carries only `b`/`i`
  (bold/italic), **not** a numeric weight. So from the runs the distinct faces
  are `(family, bold-bucket, italic)` — exactly the four OOXML
  `embeddedFont` slots (`regular`/`bold`/`italic`/`boldItalic`). True per-numeric-
  weight embedding (R9.8) needs the weight tracked at `AddRun` time and is a
  later phase.
- **`internal/ooxml/embeddings.StyleFor(weight, italic)`** already buckets a
  `(weight, italic)` into the four slots (`weight >= 600` is bold), and
  `presentation.AddEmbeddedFont` groups by typeface — so the dedup key for "is
  this face already embedded" is `(typeface, StyleFor(...))`.
- **`register-an-asset` warn-don't-fail precedent.** A missing asset warns via
  the logger and degrades, never fails the render. A missing font face must
  follow the same contract.

## 3. Findings

- **Opt-in + source-gated.** A new `WithFontEmbedding()` option sets a flag; the
  pass runs only when the flag is set **and** a `FontSource` is registered. Off,
  or with no source, it makes zero `EmbedFont` calls → byte-identical to today.
- **Walk the in-memory runs, collect the distinct `(family, bold, italic)`.** A
  `(s *SlidePart) UsedFontFaces() []FontFace` in `internal/ooxml/slide` keeps the
  XML walk inside the codec package (it already owns `DroppedDescendants`'s
  identical traversal). A run with no explicit `Latin` typeface inherits the
  theme major/minor fonts (carried by `theme1.xml`, not a per-run face) and is
  **not** embedded by this pass — the pass embeds the explicitly-set per-run
  faces, which is exactly where a brand display/heading face lands (D-063 routes
  the family through `a:latin`).
- **Deterministic order.** Merge faces across slides into a set, then sort by
  `(family, bold, italic)` and embed in that order, so the `fontN.fntdata` part
  numbering and relationship ids are byte-identical regardless of worker count.
  The pass itself runs single-threaded inside `prepareForWrite`.
- **Idempotent vs manual `EmbedFont`.** Before embedding a face, skip it when
  `presentationPart` already records `(typeface, StyleFor(weight, italic))` — so
  a caller that hand-embedded Playfair bold won't get a duplicate part. A
  `HasEmbeddedFace(typeface, style)` accessor on the presentation part supplies
  the check.
- **Warn-don't-fail.** Map `bold → weight 700` / `regular → 400` and
  `italic → style "italic"`, call the lock-free `embedFontLocked`; on
  `ErrFontNotFound` (or any resolve error) log a Warn and continue — the Save
  succeeds with the faces that resolved, the rest fall through to the host's
  substitution / fallback chain (R9.6).
- **Lock discipline.** Split `EmbedFont` into the public lock-taking wrapper and
  a `embedFontLocked` inner body (caller holds `p.mu`), matching the existing
  `ensurePresentationOPCPart` "caller holds p.mu" convention; `autoEmbedFonts`
  runs under the write-path lock and calls the inner form.

## 4. Recommendations

1. `pptx/options.go`: `WithFontEmbedding()` Option → `p.fontEmbedding = true`.
2. `pptx/presentation.go`: `fontEmbedding bool` field; call `autoEmbedFonts()` in
   `prepareForWrite` just before `syncPresentationPart`.
3. `internal/ooxml/slide/fonts.go`: `FontFace{Typeface string; Bold, Italic bool}`
   + `(s *SlidePart) UsedFontFaces() []FontFace` (walks `XSp` + table-cell text
   bodies, deduped per slide).
4. `internal/ooxml/presentation/embeddedfont.go`:
   `HasEmbeddedFace(typeface, style string) bool`.
5. `pptx/fonts.go`: refactor `EmbedFont` into a lock-free `embedFontLocked`; add
   `autoEmbedFonts()` — collect, sort, dedup, embed, warn-don't-fail.
6. White-box + black-box tests: faces embedded for themed runs; off /
   no-source byte-identical; determinism (stable part order); idempotent vs
   manual `EmbedFont`; missing face warns and the Save still succeeds. Smoke
   `scripts/smoke/phase-35.sh`.
7. `docs/glossary.md` term; `D-065`; `docs/site` + the `scaffold-a-presentation`
   skill (the existing font-embedding home).

## 5. Open questions

- **Theme major/minor faces.** Runs that inherit the theme font (no per-run
  `a:latin`) are not embedded here. Embedding the theme scheme faces would mean
  resolving the active theme's `HeadingFont`/`BodyFont` and is a possible
  follow-on; for the brand-display goal the per-run faces are what matter.
- **Subsetting & license bits (R9.12).** This pass embeds full faces. Subsetting
  to used glyphs and honoring OS/2 `fsType` is deferred (LOW; pure-Go TTF work →
  V2).
- **Weight buckets (R9.8).** Only the bold/regular bucket is available from the
  emitted `rPr`; true per-numeric-weight embedding needs weight tracked at
  `AddRun` and is a later phase.
