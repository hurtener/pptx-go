# Phase 18 — Round-trip read of self-authored decks

**Subsystem:** pptx (read) + internal/ooxml (parsers)
**RFC sections:** §16
**Deps:** Phases 03–17 (every shipped builder API / IR node has a parser counterpart)
**Status:** Complete (PR#1–4 shipped)

---

## 1. Goal

Make `pptx.Open` / `OpenStream` reconstruct the **navigable builder model** from
a pptx-go-authored deck — `pres.Slides()[0].Shapes()[0]` is the same `Shape`
model that was written — for every shipped shape, text run, fill, line, shadow,
table, and image, verified by a comprehensive round-trip test.

## 2. Why now

Phase 18 opens Wave 6 (reading + round-trip). Every builder primitive (Phases
03–17) already round-trips losslessly at the **codec** level (G6 golden tests:
`ToXML → FromXML`), and `Open` already reconstructs high-level structure
(presentation, slides, theme, masters, sections). The remaining gap is the
**public navigable read model** RFC §16 guarantees — the "R" in pptx-go.

## 3. RFC sections implemented

- `RFC §16` — round-trip of pptx-go-authored decks: every emitted shape / text
  run / fill / line / table / image / master / layout / theme parses back into
  the same Go model. (Third-party robustness is Phase 19, best-effort.)

## 4. Brief findings incorporated

- `docs/research/08-roundtrip-read.md`:
  - **F1 (reconstruct, not preserve)** → byte/codec fidelity already holds; this
    phase exposes the navigable model.
  - **F2 (extend the builder types)** → add read accessors to the existing
    `Shape`/`Fill`/`Line`/`TextFrame`/`Table` types + `Slide.Shapes()`; no
    parallel `Read*` hierarchy ("the same Shape model", RFC §16).
  - **F3 (thin mapping)** → `Open` already populates the full `XSp` tree via
    `FromXML`; read accessors are pure `X*` → public mappings inside `pptx` (P3
    intact — `pptx` consumes `internal/ooxml` Go structs, never raw XML).
  - **F4 (split by primitive group)** → PR#1 shapes+props, PR#2 text, PR#3
    tables+images, PR#4 comprehensive test.
  - **F5/F6** → theme/master/layout/section already reconstruct; image read
    resolves `blipFill` rId → media bytes.

## 5. Findings I'm departing from

None. (Brief Q3 — reading back a scene `Scene` — stays out of scope; RFC §16 is
the builder model only.)

## 6. Decisions referenced

- `D-035` — deterministic byte-identical saves — underpins the fixture
  byte-identity check (modulo documented reorderings).
- `D-032` — the shape-tree custom marshal/unmarshal that already preserves
  order on read; the read model maps over its output.
- `D-030` — the `Color` interface; read fills/lines surface colors as `Color`
  values comparable to the authored ones.
- **`D-047` (new, this PR)** — round-trip read reconstructs the navigable model
  by extending the builder types (one read+write model) with read accessors +
  `Slide.Shapes()`; delivered as a 4-PR split; pure `X*`→public mapping (no new
  XML parsing); scene-level read is out of scope.

## 7. Architecture

One plan, four PRs (the D-042/D-043 split pattern). All read accessors are pure
mappings over the `internal/ooxml` structs `Open` already populates.

```text
PR#1  shapes + props
  pptx/slide.go        Slide.Shapes() []*Shape  (enumerate spTree children)
  pptx/shape.go        Shape.Geometry/Rotation/Fill/Line/Shadow read accessors
  pptx/fill.go         readable Fill (concrete accessors)
  pptx/...             readable Line / Elevation
PR#2  text
  pptx/text.go         TextFrame.Paragraphs / Paragraph.Runs / run style+link+bullet
  Shape.TextFrame()
PR#3  tables + images
  pptx/table.go        Table read (rows/cols/cells/merge) off a graphicFrame shape
  pptx/media.go        Image read (bytes via rId, alt, crop, fit) off a picture shape
PR#4  comprehensive round-trip test
  test/integration/roundtrip_test.go   walk every shipped primitive + IR node
```

`Slide.Shapes()` wraps each `spTree` child: `*XSp` → `*Shape`, `*XGraphicFrame`
(table) → a table-bearing `*Shape`, `*XPicture` → an image-bearing `*Shape`. The
existing builder `Shape{sp *XSp}` is generalized to also wrap picture/graphic
frames (or sibling read types behind the same `Shapes()` return), decided in
PR#1.

## 8. Files added or changed

```text
# PR#1
pptx/slide.go         # CHANGED — Slide.Shapes()
pptx/shape.go         # CHANGED — read accessors (Geometry/Rotation/Fill/Line/Shadow)
pptx/fill.go          # CHANGED — readable Fill/Line
pptx/shape_read_test.go            # NEW — per-primitive read round-trip
# PR#2
pptx/text.go          # CHANGED — TextFrame/Paragraph/Run read accessors
pptx/text_read_test.go             # NEW
# PR#3
pptx/table.go         # CHANGED — Table read
pptx/media.go         # CHANGED — Image read
pptx/table_image_read_test.go      # NEW
# PR#4
test/integration/roundtrip_test.go # NEW — comprehensive walk + fixture byte-identity
scripts/smoke/phase-18.sh          # NEW (lands PR#1, criteria flip across PRs)
docs/decisions.md     # CHANGED — D-047 (PR#1)
docs/glossary.md      # CHANGED — read model / Shapes()
docs/plans/phase-18-roundtrip-read.md       # NEW (this file)
docs/research/08-roundtrip-read.md          # NEW (with plan)
```

## 9. Public API surface

```go
// pptx (read side; types are the existing builder types, gaining accessors)
func (s *Slide) Shapes() []*Shape

func (sh *Shape) Geometry() ShapeGeometry
func (sh *Shape) Rotation() float64      // degrees; 0 if unset
func (sh *Shape) Fill() Fill             // solid/gradient/noFill, readable
func (sh *Shape) Line() Line
func (sh *Shape) Shadow() (Elevation, bool)
func (sh *Shape) TextFrame() (*TextFrame, bool)
func (sh *Shape) Table() (*Table, bool)
func (sh *Shape) Image() (*Image, bool)

// read accessors on Fill/Line/TextFrame/Paragraph/Run/Table/Image as needed for
// field-equality assertions (exact set finalized per PR).
```

No write-side breaks: accessors are additive; `AddShape`/`AddText`/… unchanged.

## 10. Risks

- **R1 — Read value ≠ written value (golden equality).** **Mitigation:** read
  accessors map the same `X*` fields the writer set; tests assert authored vs
  reopened field-equality per primitive; mismatches are codec bugs fixed in the
  owning PR.
- **R2 — `Fill`/`Line` are write-oriented (sealed interface).** **Mitigation:**
  PR#1 fixes the read shape (concrete accessors / discriminator) before building
  on it; documented in D-047.
- **R3 — Scope creep across layers.** **Mitigation:** the 4-PR split bounds each
  review; pure mapping (no new parsing) keeps each PR small.
- **R4 — Image bytes on read.** **Mitigation:** resolve `blipFill` rId through
  the media manager already populated on `Open`; a missing part is a read error,
  not a panic.

## 11. Acceptance criteria

1. `Slide.Shapes()` enumerates every shape on a reopened pptx-go deck; a shape's
   `Geometry/Rotation/Fill/Line/Shadow` equal what was authored (PR#1).
2. A reopened text shape's `TextFrame` yields paragraphs → runs with the
   authored style / color / hyperlink / bullet (PR#2).
3. A reopened table shape yields the authored rows/cols/cells/merge; a reopened
   image shape yields the authored bytes/alt/crop/fit (PR#3).
4. `test/integration/roundtrip_test.go` walks **every** shipped builder
   primitive and scene IR node, asserting authored == reopened (PR#4).
5. A V1 fixture deck reopens byte-identically modulo the documented permissible
   reorderings (PR#4).
6. `make coverage` ≥ bands; `scripts/smoke/phase-18.sh` `OK ≥ count`, `FAIL=0`;
   prior smokes pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | 85% | read accessors add to the existing package |
| `internal/ooxml/slide` | 85% | codec band (mapping helpers if any land here) |

## 13. Smoke check

`scripts/smoke/phase-18.sh` (criteria flip across the PRs):

1. `OK:` library builds CGo-free.
2. `OK:` Slide.Shapes() round-trips shape geometry/fill/line/shadow (PR#1).
3. `OK:` text read round-trips runs/styles/links (PR#2).
4. `OK:` table + image read round-trip (PR#3).
5. `OK:` comprehensive roundtrip test passes; fixture reopens byte-identically (PR#4).

## 14. Tests

- **Unit / round-trip golden:** per-primitive author→save→Open→assert-model
  (PR#1–3). These are the G6 goldens elevated from codec-level to public-model.
- **Integration:** `test/integration/roundtrip_test.go` — the comprehensive
  walk + fixture byte-identity (PR#4).
- **Fuzz:** the existing `internal/ooxml`/`opc` parse fuzzers cover the decode
  surface; extend a seed if a new mapping exposes one.

## 15. Vocabulary added

- `read model` — the navigable builder model `pptx.Open` reconstructs (the same
  `Shape`/`Fill`/`Line`/`TextFrame`/`Table`/`Image` types the builder writes).
- `Shapes()` — `Slide.Shapes()`, the read-side enumerator of a reopened slide.

## 16. Plan deviations encountered during implementation

**PR#1 (shapes + props).** No deviations from the RFC or D-047. Decisions the
plan left open to PR#1, now settled:

- **Fill read shape (D-047 Q1).** `Shape.Fill()` returns the reconstructed
  **sealed `Fill`** (matches §9, "one read+write model" — no parallel read
  type). Inspection is via three reader methods added to the sealed interface:
  `Kind() FillKind` (`FillSolid` / `FillNone` / `FillGradient`), `SolidColor()
  (Color, bool)`, and `Gradient() (GradientRead, bool)`. A shape with no fill
  option returns a nil `Fill` (inherits its style fill). In-package round-trip
  tests assert authored == reopened via `reflect.DeepEqual` on the sealed value;
  literal-authored fills compare equal directly.
- **`Shape` generalized to wrap pic / graphic-frame (plan §7).** `Shape` now
  carries one of `sp` / `pic` / `gf`; `props()` and `xfrm()` resolve the common
  `<spPr>` / transform across kinds. `Box()` reads through `xfrm()`. This sets
  up PR#2 (text) / PR#3 (table, image) without reworking the handle. The builder
  still constructs only `sp` shapes.
- **Reopened colors surface as resolved literals (D-030).** Theme tokens resolve
  to sRGB at write time, so the slide carries no token to reconstruct;
  `colorFromSrgb` maps a reopened `<a:srgbClr>` back to a bare `RGB` (opaque) or
  a `literalColor` (with alpha), mirroring `SolidFill(RGB(...))` /
  `SolidFill(RGBA(...))` so a reopened fill compares field-equal.
- **`Shadow()` reconstructs cartesian offsets from polar `dist`/`dir`.** An
  axis-aligned shadow round-trips exactly; an oblique one to within sub-EMU
  rounding (D-035). `Geometry()` returns `""` for a custom-geometry (icon) or
  non-shape child — preset-geometry round-trip is exact.

**Audit findings carried to later PRs (not regressions; scoped out of PR#1 by
D-047's accessor list):**

- **Corner radius (`WithRadius`) has no read accessor yet.** The roundRect
  adjust value survives at the codec level (G6 — `avLst` is preserved), but no
  navigable accessor exposes it. A token→fraction read is inherently lossy (like
  colors/shadows), so the comprehensive **PR#4** walk asserts the adjust value
  survives; a `Shape.CornerRadius()` accessor is added there if field-level read
  is wanted.
- **Custom geometry (`custGeom`, icon glyphs) is enumerated but not decoded.**
  `Shapes()` returns the shape and `Geometry()` returns `""`; the path detail is
  not surfaced (icons are a render-only primitive, not a navigable authored
  shape). PR#4 confirms icon decks reopen and re-save byte-identically.

**PR#2 (text).** No deviations from the RFC or D-047. Decisions settled in PR#2:

- **`Run` exposes resolved character properties via getter methods**, not a
  returned `RunStyle`. A run's `TypeRole` and token color resolve to a concrete
  family / size / sRGB at write time (D-033), so they cannot be recovered; the
  read accessors (`Text`, `Font`, `FontSize`, `Bold`, `Italic`, `Underline`,
  `Strike`, `Baseline`, `Color`, `Code`) report the resolved values — the same
  resolved-literal stance as PR#1's `Fill`.
- **Getter names avoid the chained-setter collision.** `Align`/`Indent`/`Bullet`
  are builder setters returning `*Paragraph`; the read inverses are
  `Alignment()`, `Level()`, `BulletStyle()`.
- **Back-references added for hyperlink resolution.** `Shape` gained `s *Slide`
  and `Run` gained `tf *TextFrame` (set in both the authoring and read paths);
  `Run.Hyperlink()` resolves the run's `<a:hlinkClick r:id>` through the slide's
  reopened relationships to the verbatim URL (`TargetRef`).
- **`Code()` is detected by the inline-code highlight tint** pptx-go applies for
  code and nothing else (D-013). Line breaks (`AddBreak`) carry no text and are
  not returned by `Runs()`. Frame body props (anchor / autofit / margins) read
  is deferred to PR#4 — the codec already preserves them (G6).

**PR#3 (tables + images).** No deviations from the RFC or D-047. Decisions
settled in PR#3:

- **`Shape.Table()` / `Shape.Image()` reconstruct the existing `Table` / `Image`
  handles** off the reopened graphic-frame / picture child, deriving the
  unexported state (`rows`/`cols` from the grid/rows, `headerOn`/`bandRowOn` from
  the `tblPr` flags). Table read accessors: `RowCount`, `ColCount`,
  `ColumnWidths`, `HeaderRow`, `RowBanding`; cell accessors: `GridSpan`,
  `RowSpan`, `Covered`, `Fill`. Cell **text** reuses the existing
  `Cell.TextFrame()` + the PR#2 read model — one model, no new text path.
- **Image bytes resolve through the package (R4).** `Image.Bytes()` follows the
  picture's `<a:blip r:embed>` relationship to its media part in the reopened
  `opc.Package` (`Image` gained an owning-slide back-ref); a missing
  relationship/part returns `ErrImagePartMissing`, never a panic. Bytes are
  returned verbatim — no pixel decode (§7). Other image reads (`AltText`, `Crop`,
  `Fit`, `Rotation`, `Opacity`) are pure field maps; `Fit` reports `FitFill` when
  a stretch fill is present (the builder default), `FitNone` otherwise.

**PR#4 (comprehensive round-trip test).** No deviations. Decisions settled:

- **The fixture reopens byte-identically with no reordering caveat.** A deck
  exercising the full builder surface (shapes + fill/line/shadow/rotation, rich
  text + bullets/links, a merged table, a cropped/rotated image) re-saves
  byte-for-byte through `Open` (D-035) — so criterion 5 asserts exact
  `bytes.Equal`, and the "modulo documented permissible reorderings" allowance
  was not needed. A scene-rendered deck (all 20 node kinds) is likewise
  byte-identical on reopen.
- **The scene side is walked at the builder level (D-047).** Scene read is out of
  scope, so `TestRoundTrip_SceneNodes` renders a scene covering every node kind
  (a mechanical `KindHero..KindCardSection` coverage assertion guards against a
  future node slipping the walk), reopens it, and asserts every slide exposes
  navigable shapes and the deck re-saves byte-identically — confirming the
  builder primitives each node composes round-trip.
- **Radius / custGeom field-level read** (flagged in PR#1) stays a V1.x follow-up:
  the integration walk confirms a roundRect and an icon (custGeom) deck reopen
  and re-save byte-identically, so the codec preserves them (G6); a navigable
  `CornerRadius()` accessor is not added in V1.

## 17. Sign-off

- [x] All acceptance criteria pass. (PR#1 shapes, PR#2 text, PR#3 tables+images,
      PR#4 comprehensive walk + byte-identity.)
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-18.sh` reports `OK ≥ count(criteria)` and `FAIL = 0`
      (5 OK, 0 FAIL, 0 SKIP).
- [x] Prior phases' smoke scripts still pass (`make preflight` PASS).
- [x] Glossary updated (read model / `Shapes()`, in the plan commit).
- [x] Decision entries added (D-047).
- [x] Round-trip goldens for the new read accessors land with each PR
      (`test/pptx/shape_read_test.go`, `text_read_test.go`,
      `table_image_read_test.go`, `test/integration/roundtrip_test.go`).
- [ ] (Phase 20+) Docs site / skills updated. (inert)
