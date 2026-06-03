# Brief 08 — round-trip read of self-authored decks

**Subsystem:** pptx (read) + internal/ooxml (parsers)
**Authored:** 2026-06-03
**Motivating phase:** Phase 18 — Round-trip read of self-authored decks (RFC §16)

## 1. Question

RFC §16 guarantees that a pptx-go-authored deck reopens into **the same Go
model it was written from** — `pres.Slides()[0].Shapes()[0]` is the Shape model
we wrote, and "every shape, text run, fill, line, table, image, master, layout,
and theme pptx-go emits is parsable back into the same Go model." What does
`pptx.Open` reconstruct today, what's missing, and how should the read model be
shaped to satisfy the guarantee?

## 2. Prior art surveyed

(From a read-only sweep of the repo.)

- **`pptx.NewFromBytes` / `OpenStream`** (`pptx/presentation.go:181`,
  `pptx/stream.go`): parse OPC + `presentation.xml`, then `repopulateSlides`,
  `repopulateSections`, `loadThemeFromPackage`, `buildMasterRegistry`.
  **Verdict: hybrid** — high-level structure (Presentation, Slides, Theme,
  Masters/Layouts, Sections) is reconstructed into Go objects; **slide shape
  content is preserved as opaque OOXML** in `SlidePart.spTree` (`*XSpTree` of
  `*XSp` / `*XPicture` / `*XGraphicFrame`).
- **Codec round-trip (G6) already holds** at the wire level:
  `internal/ooxml/slide/slide_roundtrip_test.go` proves `ToXML → FromXML`
  losslessly preserves shapes, fills (+alpha), lines (+dash), text
  (runs/styles/bullets/breaks), tables (merge/banding), pictures (alt/crop).
  `test/integration/reorg_roundtrip_test.go` proves write → `NewFromBytes` →
  re-save identity. So **byte/codec fidelity is done**; the gap is the *public
  navigable model*.
- **Read-side public API today:** `Presentation.Slides/GetSlide/SlideCount/
  Theme/Sections`; `Slide.Index/Layout/SlideSize`. **No** `Slide.Shapes()`, no
  shape/fill/line/text/table/image read accessors. The builder `Shape` type
  (`pptx/shape.go`) wraps `*XSp` and exposes only `Box()`.
- **Parsers that exist:** master/layout (`master_parser.go` → domain models),
  presentation, theme (`themecodec.go`), core props (partial), slide
  (`FromXML` → opaque `XSpTree`). **Missing:** XSp→Shape-props, fill→Fill,
  line→Line, text-body→TextFrame, graphicFrame→Table, blipFill→Image readers
  that yield the *public* model.

## 3. Findings

- **F1 — The guarantee is reconstruct, not preserve.** RFC §16 shows
  `Slides()[0].Shapes()[0]` returning "the same Shape model we wrote." Byte
  round-trip (already working) is necessary but not sufficient; Phase 18 must
  expose a navigable read model. (Confirmed with the maintainer.)
- **F2 — Extend the builder types (one model).** "The same Shape model" ⇒ the
  read side reuses the builder's `Shape`/`Fill`/`Line`/`TextFrame`/`Table`
  types, adding read accessors + a `Slide.Shapes()` enumerator, rather than a
  parallel `Read*` hierarchy. (Maintainer choice.) Implications:
  - `Shape` gains `Geometry()`, `Rotation()`, `Fill()`, `Line()`, `Shadow()`,
    `TextFrame()` (and keeps `Box()`); `Slide.Shapes() []*Shape` enumerates the
    `spTree` children, wrapping each `*XSp` (and picture/graphicFrame variants).
  - `Fill` is currently a write-only sealed interface (`applyFill`). Reading
    needs either concrete accessors (e.g. a `SolidFill` with `Color()`,
    `LinearGradient` with `Stops()`) or a discriminator. The plan picks the
    read shape of `Fill`/`Line` so a read value compares equal to the written
    one for the golden tests.
- **F3 — Codec already parses everything; reconstruction is a thin mapping.**
  Because `FromXML` populates the full `XSp`/`XGraphicFrame`/`XPicture` struct
  tree, the read accessors are mostly pure functions `X*` → public type — no new
  XML parsing, just domain mapping in `pptx` (P3 stays intact: `pptx` reads the
  `internal/ooxml` Go structs it already consumes, never raw XML).
- **F4 — Split by primitive group.** The surface (shapes, text, tables, images)
  is large; deliver as PR#1 shapes+props, PR#2 text, PR#3 tables+images, then a
  comprehensive `test/integration/roundtrip_test.go`. (Maintainer choice.)
- **F5 — Theme/master/layout/section already reconstruct** — no Phase-18 work
  beyond confirming the round-trip test covers them.
- **F6 — Picture/image read** needs the `blipFill` rId → media bytes resolution
  (the media manager already holds parts on open) plus crop/alt/fit mapping.

## 4. Recommendations

- **R1 — `Slide.Shapes() []*Shape`** enumerating reconstructed shapes; add
  `Shape` read accessors (`Geometry`, `Rotation`, `Fill`, `Line`, `Shadow`,
  `TextFrame`). Pure mapping over the already-parsed `XSp` tree.
- **R2 — Make `Fill`/`Line` readable**: expose concrete read accessors so a
  reopened fill/line compares equal to the authored one (the golden assertion).
- **R3 — Text read**: `TextFrame.Paragraphs()` → `Paragraph.Runs()` → run
  style/color/link/bullet accessors, mapping `XTextBody`.
- **R4 — Table read**: `Slide.Shapes()` surfaces a table-bearing graphicFrame as
  a readable `Table` (rows/cols/cells/merge), mapping `XGraphic.…Table`.
- **R5 — Image read**: a picture shape exposes its bytes (via rId→media), alt,
  crop, fit.
- **R6 — Comprehensive `test/integration/roundtrip_test.go`** walking every
  shipped builder primitive + scene IR node: author → save → `Open` → assert the
  reconstructed model equals the authored model, plus a fixture-deck
  byte-identity check (modulo documented reorderings).

## 5. Open questions

- **Q1 — `Fill`/`Line` read shape** (concrete typed accessors vs a tagged-union
  reader). A plan-level design detail; pick the form that makes golden equality
  cleanest.
- **Q2 — Permissible reorderings** for the byte-identity fixture check (e.g.
  rel-id ordering, attribute order). Document the allowed set in the plan; the
  deterministic-save work (D-035) already minimizes these.
- **Q3 — Scene-layer read?** RFC §16 is about the *builder* model; the scene IR
  is one-way (scene → builder). Phase 18 reconstructs the builder model only;
  reading back a `Scene` is out of scope (not promised by the RFC).
