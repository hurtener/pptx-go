# Brief 21 — weight-aware-embedding

**Subsystem:** pptx — Layer 1 builder (font embedding)
**Authored:** 2026-06-22
**Motivating phase:** Phase 38 — weight-aware font embedding

## 1. Question

Brands use a weight ladder (e.g. light/regular display, medium labels, regular
body — 300/400/500/700). The embedding pass (D-065) keys the used-face set on
`(family, bold, italic)` taken from each run's `rPr`, which carries only `b`/`i`
— no numeric weight. So a `500` "medium" run collapses to the regular bucket and,
worse, the pass embeds a *synthetic* weight (`700` if bold else `400`): a medium
weight ships as the 400 file, not the 500 the soul asked for. How can the engine
embed the **correct physical weight file** for each weight a deck actually uses —
additively, deterministically, and byte-identical when unused? This is
`DECKARD-PRODUCT-REQUIREMENTS.md` R9.8 (`weight-aware-embedding`, MED · both;
engine half — D-059).

## 2. Prior art surveyed

- **`pptx/text_layout.go::toProps`.** Resolves the role `FontSpec` (which carries
  `Weight`) but emits only `b="1"` when `Weight >= 600` (or a per-run bold
  override); the numeric weight is discarded after the bold bit.
- **`internal/ooxml/slide` run model.** `XTextProperties` is the in-memory run
  property struct walked by `UsedFontFaces` (D-065). A non-serialized field
  (`xml:"-"`) on it survives from `toProps` to the embedding walk without
  touching the wire format.
- **D-065 embedding pass.** `autoEmbedFonts` collects `slide.FontFace{Typeface,
  Bold, Italic}` and calls `EmbedFont(family, style, weight)` with a synthetic
  `weight` (700/400). `EmbedFont` resolves the bytes via the `FontSource` and
  records the OOXML `embeddedFont` bucket via `embeddings.StyleFor(weight,
  italic)` — one of four slots (regular/bold/italic/boldItalic) **per typeface**.
- **OOXML `<p:embeddedFontLst>`.** A `<p:embeddedFont>` has exactly four face
  slots; `buildEmbeddedFontList` groups entries by typeface and fills one ref per
  slot (last write wins). PowerPoint's renderer is limited to these four cuts —
  it cannot select between two same-bucket weights of one family.

## 3. Findings

- **Track the resolved numeric weight per run.** Add `XTextProperties.Weight int`
  with `xml:"-"` (in-memory only, never serialized/parsed → byte-identical and
  round-trip-neutral). `toProps` sets it to the effective weight (the role's
  `FontSpec.Weight`, bumped to ≥700 when a per-run bold override is set). The
  collector reads it; when it is `0` (a parsed/round-tripped deck), it infers
  `700/400` from the bold bit as before.
- **Key the used-face set on weight.** `slide.FontFace` gains `Weight int` and
  `UsedFontFaces` populates it, so a `400` and a `500` of the same family are
  distinct used faces.
- **Embed the actual weighted file, bucketed deterministically.**
  `autoEmbedFonts` collects the distinct `(family, weight, italic)` set, sorts it,
  and groups by OOXML bucket `(family, weight≥600, italic)`. For each bucket it
  picks the **nearest-nominal** weight (regular nominal `400`, bold nominal
  `700`; ties → the lower weight) and calls `EmbedFont(family, style, weight)`
  with that *actual* weight — so the provider returns the correct physical file
  (a 500-only regular bucket ships the medium file, not 400). When a bucket has
  more than one used weight, the winner is deterministic and the coalescing is
  logged.
- **Honor the 4-bucket limit; don't ship orphan parts.** Because `<p:embeddedFont>`
  has four slots and PowerPoint uses only those, the engine embeds **one file per
  bucket**. Embedding additional same-bucket weight files purely so an external
  rasterizer can pick a finer weight would create `embeddedFontLst`-unreferenced
  font parts — risking the no-repair-prompt guarantee (D-020) for zero PowerPoint
  benefit. That is a product/rasterizer concern (D-026): a caller whose rasterizer
  needs extra cuts can call `EmbedFont` for them explicitly. Deferred from the
  engine.
- **Additive + deterministic.** With no `FontSource` / `WithFontEmbedding` off the
  pass still does nothing → byte-identical. The weight is metadata that never
  flips `toProps`'s emit flag, so unstyled runs still return `nil` rPr
  (byte-identical). Nearest-nominal selection is a pure integer function.

## 4. Recommendations

1. `internal/ooxml/slide`: `XTextProperties.Weight int \`xml:"-"\``;
   `FontFace.Weight int`; `UsedFontFaces` populates weight (infer from bold when 0).
2. `pptx/text_layout.go`: `toProps` sets `p.Weight` to the effective weight.
3. `pptx/fonts.go`: `autoEmbedFonts` collects `(family, weight, italic)`, buckets
   by `(family, bold, italic)`, picks the nearest-nominal weight per bucket, and
   embeds that actual weight; logs coalescing.
4. Tests: a medium-weight role embeds the medium file (provider asked for the
   resolved weight); a single-weight deck embeds one file; two weights on one
   bucket coalesce to the nearest-nominal winner (+ warn); byte-identical when off;
   deterministic.
5. Docs: THEME.md embedding note + glossary; D-068 records the 4-bucket deviation.

## 5. Open questions

- **Multi-file-per-bucket for external rasterizers** (the literal "embeds three
  distinct files" acceptance) — deferred to the product/V2 as above; the resolver
  callback and weight-keyed collector are already shaped to support it if the
  no-repair-prompt question is later resolved for unreferenced font parts.
- **Bold-cut fallback** (R9.7 follow-on) and **subsetting** (R9.12, → V2) remain
  separate.
