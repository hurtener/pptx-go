# Phase 03 — Builder spine

**Subsystem:** pptx (core builder)
**RFC sections:** §5, §8.1–8.3, §8.6, §8.7, §8.8
**Deps:** Phase 02; the validation harness (D-031).
**Status:** In progress

---

## 1. Goal

Turn the inherited upstream builder into the RFC's builder spine: `pptx.New`
emits a **complete, valid, themed** deck; shapes/fills/lines/images take
tokens through a `Color` interface; sections and speaker notes are
first-class; an always-on hygiene pass keeps PowerPoint from showing the
"repaired" prompt (D-020). Every shipped primitive round-trips losslessly
and passes the validity gate (D-031).

## 2. Why now

The builder is the substrate every higher layer composes (RFC §8). Wave 2
(scene renderer) and Phase 04 (rich text) build directly on it. Phase 02
delivered the theme/token model; this phase wires it into builder calls
(closing D-030) and makes the validator (D-031) gate complete decks.

## 3. RFC sections implemented

- `§5` — toolchain/packaging baseline (already met; this phase adds no deps).
- `§8.1` — `Presentation`/`Slide` top-level API: `New(opts)`, `Open`,
  `OpenStream`, `Save`, `SaveStream`, `Write`, `Theme`/`SetTheme`,
  `AddSlide`, `Slides`, `Close`.
- `§8.2`/`§8.3` — `Box`/`Anchor` (Phase 02 shipped units/geom);
  `AddShape(geom, box)`, `ShapeGeometry` presets, `Fill` interface, `Line`.
- `§8.6` — media: `AddImage`, `ImageSource` (`ImageFile`/`ImageBytes`/
  `ImageReader`), alt text, crop, fit, dedup.
- `§8.7` — backgrounds, masters/layouts (enough to emit a complete deck;
  full template ingestion is Phase 09); **sections**.
- `§8.8` — speaker notes.
- Rich text (§8.4) is **Phase 04**, not here. `AddText`/`TextFrame` land
  there; this phase ships the shape/media/section/notes spine.

## 4. Brief findings incorporated

No informing brief — the builder API is specified directly in RFC §8 and
the decisions log (D-012, D-019, D-020, D-022, D-026, D-030). The upstream
builder is the substrate being reshaped (RFC §17.1).

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-012`/`D-030` — `Color` becomes an interface (`tokenColor`/`literalColor`)
  with `TokenColor()`/`RGB()`; this phase makes that real (it was deferred
  here).
- `D-019` — `WithFontSource` option (moved from the Phase 02 `SetFontSource`).
- `D-020` — always-on repair-prompt hygiene pass (`internal/render/hygiene.go`).
- `D-022` — speaker notes are V1.
- `D-026` — engine, not product: no render modes/heuristics added.
- `D-031` — the validity gate; this phase turns on full-deck conformance +
  schema + LibreOffice and closes the `r:id` baseline gap.
- `D-032` *(new, from the Phase 03 investigation)* — one emission path
  (`xml.Marshal` + a `RestoreNamespaces` write pass); the hand-rolled slide
  `XMLWriter` is deleted. The architectural basis for Chunk A1.

## 7. Architecture & chunking

Phase 03 is large; it ships as a sequence of coherent, individually-green
PRs under this plan (CLAUDE.md §4.3). Each keeps CI green and advances the
validity gate.

**Chunk A — rebuild the emission so decks are valid, complete, themed.**
Investigation (recorded in D-032) found the inherited emission is broken in
several ways with one root cause: no consistent namespace handling on write.
`presentation.xml` emits with **no namespaces**; slides use a hand-rolled
`XMLWriter` that writes **attributes as text** (`<p:cNvPr>1 name=…</p:cNvPr>`);
presentation→slide relationships are never wired (`sldId rid=""`); `New()`
emits no master/layout/theme; and `AddAutoShape` mistreats EMU inputs as
pixels. Chunk A, in verifiable steps:

- **A0 — harden the harness first (this PR).** Extend `internal/conformance`
  to catch what it missed (empty/missing rel-id references; root elements
  with no namespace) and make the LibreOffice proxy assert real PDF content.
  This turns the false-green red and gives the rebuild a target.
- **A1 — one emission path (D-032).** Serialize every part via `xml.Marshal`
  (bare names) + a shared `RestoreNamespaces` write pass (inverse of
  `StripNamespacePrefixes`); **delete the hand-rolled slide `XMLWriter`**.
  Fixes the missing namespaces, the attributes-as-text bug, and `rid`→`r:id`
  at the root. Golden-tested.
  - **A1a — `RestoreNamespaces`** (done): the write-side inverse, with the
    element→prefix table extracted from the writer; declares only the used
    prefixes on the root; golden-tested. Proven to fix namespaces +
    attributes when wired.
  - **A1b — complete the structs + custom container marshaling (done).** The
    inherited structs did *not* fully represent the OOXML the writer emitted:
    the heterogeneous, ordered `spTree` children were `xml:"-"` (so
    `xml.Marshal` dropped shapes) and shape geometry lived in
    `XSp.ShapePreset string` (`xml:"-"`, never serialized). A1b added a custom
    `MarshalXML`/`UnmarshalXML` pair for `spTree` (`slide_marshal.go`) and the
    missing typed fields (`prstGeom` via `XPresetGeometry`, `nvPr`,
    `graphicData` via `XGraphicData`, `solidFill` via `XSolidFill`), reordered
    the shape structs to the CT_* schema, rewired `SlidePart.ToXML` and
    `PresentationPart.ToXML` onto `xml.Marshal` + `ooxml.RestoreNamespaces`,
    and **deleted the ~700-line hand-rolled `XMLWriter`/`XMLWriterPool` and all
    `WriteXML` methods**. Verified by a codec round-trip golden
    (`slide_roundtrip_test.go`), a builder-facing structure test
    (`test/parts`), and the conformance gate (presentation + slides now carry a
    namespaced root; `cNvPr` attributes serialize as attributes, not text).
- **A2 — wire relationships + seed a complete deck (done).** `AddSlide`
  allocates a real presentation→slide rId and adds the relationship (plus a
  slide→layout rel); `New()` seeds a master + blank layout + theme with their
  rels (`pptx/scaffold.go` + `scaffold_assets.go`). `presentation.xml` is
  reordered to the CT_Presentation sequence (sldMasterIdLst, sldIdLst, sldSz,
  notesSz) with a valid `sldMasterId` id (≥ 2147483648). The full-deck
  conformance gate is on (`TestConformance_BuilderOutput` asserts
  `RequiredParts`: presentation, slide1, master, layout, theme; `rep.OK()`),
  the LibreOffice job asserts a 2-page render (poppler `pdfinfo`), and the
  manual PowerPoint check is queued for the maintainer
  (`docs/validation/POWERPOINT-CHECKS.md`).
- **A3 — EMU coordinates + options (done).** The builder path no longer scales
  inputs by 9525 — `Slide.AddTextBox`/`AddAutoShape`/`AddRectangle`/`AddPicture`/
  `AddTable` now take EMU directly (`pptx.In`/`Cm`/`Pt`/`Px`/`Box` compute them),
  fixing the off-canvas coordinates. `New(opts ...Option)` with
  `WithFormat(Slides16x9|Slides4x3)`, `WithFontSource` (Option form;
  `SetFontSource` kept as a deprecated alias) and `WithTheme`; `Theme()`/
  `SetTheme()` accessors (default `DefaultTheme`). `PxToEMU` is deprecated in
  favour of `pptx.Px`. The Box-native `AddShape(geom ShapeGeometry, box Box)`
  with fills/lines is **Chunk B**.
- **A4 — `internal/render/hygiene.go` (done).** `render.Sanitize` runs
  unconditionally on every emitted XML part across all write paths
  (`applyHygiene` in `Save`/`Write`/`WriteToBytes`); no caller-facing switch
  (D-020). V1 trigger list (`docs/design/HYGIENE.md`): H1 strip a leading
  UTF-8 BOM, H2 drop empty `lang=""`. Conservative + idempotent; golden tests
  assert it touches only triggers. `internal/render` banded at 80%.

**Chunk B — Color/Fill/Line + shapes (done; D-033).**
- The upstream concrete `Color` struct (with `ColorMap`/`ParseColor`/presets)
  is retired in favour of a **sealed `Color` interface**. The Phase-02 `RGB`
  type implements it (`pptx.RGB(...)` is the literal), `pptx.RGBA` adds alpha,
  and `pptx.TokenColor`/`TokenTextColor` are tokens that resolve against the
  active theme at apply time.
- `Fill` interface with `SolidFill(Color)` and `NoFill()`; `Line` (width/color/
  dash); `AddShape(geom ShapeGeometry, box Box, …WithFill/WithLine)` with preset
  geometry constants, returning an opaque `*Shape`. Round-trip goldens (codec)
  + theme-swap proven end-to-end (acceptance criterion 7) in `test/pptx`.
- Gradient/pattern/picture fills are deferred (picture fills land with media,
  Chunk C); the `Color`/`Fill`/`Line` interfaces are sealed so they extend
  without breaking callers.

**Chunk C — media, sections, notes, streaming (done; D-034).**
- `AddImage(src, box)` + `ImageSource` (`ImageFile`/`ImageBytes`/`ImageReader`,
  the §4.4 seam), alt text, crop, and a V1 `Fit` (`FitFill`/`FitNone` — aspect-
  aware fits need pixel dimensions, forbidden by §7); content-type sniff +
  verify with no pixel parsing; content-dedup via the upstream MediaManager.
  Closed the A2 relationship seam: the slide owns its rels (layout + image +
  notes in one `rId` namespace), `syncSlides` mirrors them onto the package
  part, `syncMedia` writes the deduped bytes. Fixed the `AddPictureFromFile`
  `io.ReadAll(nil)` stub.
- `AddSection`/`Section.Include`/`Sections()` → a `p14:sectionLst` injected into
  `presentation.xml`'s extLst (D-034), with read-back for round-trip.
- `SpeakerNotes(text)` → a `notesSlide` part + hand-authored `notesMaster1.xml`,
  wired via rels + `notesMasterIdLst`.
- `OpenStream(path)`/`SaveStream(path)` over the `internal/opc` streaming
  package, with the D-020 hygiene pass on the write path. Fixed a streaming
  reader bug that dropped the package `.rels` on open.
- Round-trip goldens: codec-level picture descr/srcRect; pptx-level media,
  sections, notes, and streaming.

```text
internal/render/hygiene.go         # NEW (A) — D-020 repair pass
internal/ooxml/**/* (rel codecs)   # CHANGED (A) — r:id namespace fix
pptx/presentation.go               # CHANGED (A,C) — New(opts), format, sections, streaming
pptx/options.go                    # NEW (A) — Option, WithFormat, WithFontSource
pptx/color.go                      # CHANGED (B) — Color interface
pptx/shape.go                      # NEW/CHANGED (B) — AddShape, Fill, Line, geometry
pptx/media.go                      # CHANGED (C) — ImageSource, AddImage
pptx/section.go, pptx/notes.go     # NEW (C)
docs/design/HYGIENE.md             # NEW (A)
```

## 8. Files added or changed

Per chunk above; each chunk lists its exact files in its PR. The §14
pre-merge checklist gates every chunk PR.

## 9. Public API surface

The RFC §8.1–8.3/8.6–8.8 surface. Breaking changes to the inherited builder
(notably `Color` struct → interface) carry deprecated aliases where a
drop-in exists; otherwise documented in `CHANGELOG.md` (pre-V1, v0.x).

## 10. Risks

- **R1 — `r:id` namespace fix scope.** Go's `encoding/xml` attribute
  namespacing is finicky. *Mitigation:* fix at the codec layer with a
  golden + the schema/LibreOffice/manual layers verifying; bounded to the
  rel-referencing structs.
- **R2 — `Color` struct → interface migration breakage.** Upstream shape
  code uses the struct. *Mitigation:* narrow footprint (2 files, Phase 00
  survey); migrate with build/vet/test as the gate; aliases where needed.
- **R3 — Complete-deck wiring (masters/layouts).** Emitting a minimal valid
  master+layout is non-trivial. *Mitigation:* reuse the upstream master
  manager + the default theme; the conformance + LibreOffice gates prove
  completeness; full template ingestion stays Phase 09.
- **R4 — Hygiene pass false edits.** A post-processor that rewrites XML
  could corrupt valid output. *Mitigation:* a conservative, documented
  trigger list (HYGIENE.md); golden tests asserting it only touches
  triggers; runs through the validity gate.

## 11. Acceptance criteria

1. `pptx.New()` (no config) emits a **complete** deck that passes the
   full-deck conformance gate, `xmllint` schema validation (once vendored),
   the LibreOffice open-proxy, and opens in real PowerPoint with no repaired
   prompt (manual, recorded in `docs/validation/`).
2. Relationship references emit as `r:id` and resolve (the `rid` gap closed).
3. A 1-slide deck with a rect + image round-trips losslessly (model equality)
   through `pptx.Open`.
4. `pptx.Slides4x3` produces a 4:3 canvas; round-trips.
5. A `Section` of slides round-trips; PowerPoint shows it in the slide sorter.
6. `SpeakerNotes()` text round-trips losslessly.
7. A shape filled with `SolidFill(TokenColor(ColorAccent))` re-renders under
   a swapped theme (token, not literal).
8. The hygiene pass runs on every write; emitted decks show no repaired
   prompt.
9. `make build`/`test`/`lint`/`coverage`/`preflight`/`check-mirror` +
   the validate workflow green.

## 12. Coverage targets

New `pptx` builder files: 85%. `internal/render`: 80%. The `pptx` package
comes fully under the coverage gate + linter as this phase consolidates it
(closing the D-029/Phase-02 staging); inherited files are rewritten or
removed here.

## 13. Smoke check

`scripts/smoke/phase-03.sh`: build; `New()` emits a deck that passes the
full conformance gate; round-trip golden passes; 4:3 format; section +
notes round-trip; hygiene pass present.

## 14. Tests

- **Round-trip golden** (primary, Phase 03+): every primitive write→Open→
  assert model equality.
- **Conformance** (D-031): full-deck gate on every emitted deck.
- **Integration**: theme-swap end-to-end; section/notes round-trip.
- **Fuzz/Bench**: deferred (parse-surface fuzz lands Phase 18–19).

## 15. Vocabulary added

- `Fill` / `Line` / `ShapeGeometry` — shape appearance primitives.
- `ImageSource` — file/bytes/reader image input.
- `Section` — named slide grouping.
- `RepairPromptHygiene` — already in glossary (D-020).

## 16. Plan deviations encountered during implementation

- **A1b — `clrMapOvr` + `cSld` correctness.** The retired writer emitted an
  invalid `<a:defRgbClrModel val="bg1"/>` inside `clrMapOvr` and wrote the
  `spTree` directly under `<p:sld>` (no `<p:cSld>` wrapper). The rebuilt
  emission emits the standard `<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>`
  and the required `<p:cSld>` envelope.
- **A1b — table/graphicFrame namespace fidelity deferred.** A graphic frame's
  transform is `<p:xfrm>` (PresentationML), but `RestoreNamespaces` keys on a
  single element→prefix table that maps `xfrm`→`a` (correct for the far more
  common shape `spPr` case). Tables are not a Phase 03 acceptance primitive and
  appear in no gated deck, so `graphicFrame`/`tbl` emission is left best-effort
  (parity with the old writer, which also emitted `a:xfrm`); full table
  namespace fidelity lands when tables are formally shipped. No regression.
- **A1b — `PxToEMU` left in place.** `AddShape`-path shapes still multiply EMU
  inputs by 9525 (off-canvas coordinates); this is A3's EMU `Box` API work and
  is out of scope for the emission rebuild. Conformance does not check
  coordinates, so the gate is unaffected.
- **A2 — static scaffold theme instead of `DefaultTheme` emission.** The seeded
  master/layout/theme are hand-authored, namespaced OOXML constants
  (`scaffold_assets.go`), the lowest-risk path to a PowerPoint-valid deck now
  (plan R3). Emitting `theme1.xml` from the `Theme` model (and fixing the theme
  codec to emit `a:`-prefixed elements — it currently emits a bare-element
  `<theme xmlns:a>` that is not in any deck) is deferred to **Chunk B**, which
  owns Color/theme-swap; B replaces `scaffoldThemeXML` with token-driven
  emission. The scaffold theme colors already mirror `DefaultTheme`.
- **A2 — `RequiredParts` covers parts, not package structures.** The
  conformance gate requires the five real parts (presentation, slide, master,
  layout, theme); `[Content_Types].xml` and `_rels/.rels` are package
  structures (not in the part collection), already validated by the
  content-type-coverage and relationship-resolution checks.
- **A3 — builder coordinates are EMU, not a new `Box`-typed signature.** The
  inherited `Slide.Add*` methods keep their `(x, y, cx, cy int, …)` shape but
  now interpret the ints as EMU (the px→EMU scaling is dropped). The Box-native
  shape API the plan envisions (`AddShape(geom, box Box)`) belongs to Chunk B,
  which reshapes the shape surface with geometry/fill/line; introducing a
  parallel Box signature in A3 would be churn B immediately reworks. The
  px-based `SlideViewport`/boundary-check helpers are a separate placement
  utility and stay px.
- **C — `AddImage(src, box)` argument order follows the RFC example** (`src`
  first), not the orientation note's `(box, src)`; this matches the RFC §8.6
  snippet and the `AddShape(geom, box)` shape convention (the "what" precedes
  the "where").
- **C — `Fit` ships as `FitFill`/`FitNone` only.** PowerPoint stores no single
  fit value and cover/contain math needs the image's pixel dimensions, which §7
  forbids parsing. `FitFill` emits `<a:stretch>`, `FitNone` omits it; richer,
  aspect-aware fits are deferred (caller-side `Box` sizing drives aspect —
  D-026). (D-034.)
- **C — speaker notes are a plain-text setter, not `*TextFrame`.** RFC §8.8
  sketches `SpeakerNotes() *TextFrame`, but `TextFrame` is the rich-text model
  (a later phase). V1 ships `SpeakerNotes(text string)`; the accessor evolves
  with rich text. (D-034.)
- **C — streaming uses the RFC §9 path-based signatures.** `OpenStream(path)` /
  `SaveStream(path)` match the RFC exactly; CLAUDE.md §5's context-first
  convention yields to the explicit RFC signature (RFC > this file, §2). A
  context-aware streaming API would be an RFC change + superseding decision.
  (D-034.)
- **C — section list emitted as an injected `p14` fragment.** `RestoreNamespaces`
  is single-table and can't emit both `p:sldId` and `p14:sldId`, so the section
  list is built as a literal fragment and injected into `presentation.xml`; the
  read path parses it back via prefix-stripped structs. (D-034.)
- **C — fixed a pre-existing `internal/opc` streaming bug.** `StreamPackage`
  dropped the package `.rels` on open (it tested `IsPackageRels` on the source
  URI, not the rels URI), which would have made `OpenStream`→`SaveStream`
  output invalid. Corrected to match `opc.Package.loadRelationships`. (D-034.)

## 17. Sign-off

Per chunk: acceptance criteria for that chunk, coverage, smoke, validity
gate, glossary/decisions updated. The phase is done when criteria 1–9 pass.
