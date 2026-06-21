# Changelog

All notable changes to pptx-go are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
While the library is pre-V1 (`v0.x`), minor versions may carry breaking
changes.

## [Unreleased]

### Added

- Build, test, and lint scaffolding: `Makefile`, `scripts/preflight.sh`,
  `scripts/drift-audit.sh`, the smoke-script template, git hooks, the
  GitHub Actions CI workflow, `.golangci.yml`, and `.editorconfig`.
- `internal/coveragecheck` — mechanical per-package coverage band gate
  driven by `coverage.json`.
- `pptx.New` now accepts options: `WithFormat(Slides16x9 | Slides4x3)`,
  `WithFontSource`, and `WithTheme`. Added `Presentation.Theme()` and
  `SetTheme`.
- Token-aware color and shapes: a `Color` interface with `RGB`/`RGBA` literals
  and `TokenColor`/`TokenTextColor` theme tokens (tokens resolve against the
  active theme, so a theme swap recolors the same input); `Fill`
  (`SolidFill`/`NoFill`), `Line`, and `Slide.AddShape(geom, box, WithFill/
  WithLine)` with preset `ShapeGeometry` constants.
- A new presentation is a complete, valid deck out of the box: `New()` emits a
  slide master, a blank slide layout, and a theme, with all relationships
  wired, so the file opens in PowerPoint without a repair prompt.
- Images: `Slide.AddImage(src, box)` with an `ImageSource` (`ImageFile`,
  `ImageBytes`, `ImageReader`); the returned handle's `SetAltText`, `SetCrop`,
  and `SetFit` adjust the picture. Image bytes are verified against a known
  signature and rejected if malformed or mismatched; identical bytes are
  embedded once.
- Slide grouping: `Presentation.AddSection(name)`, `Section.Include(slide)`, and
  `Presentation.Sections()` — sections appear in PowerPoint's slide sorter and
  round-trip.
- Speaker notes attach to a slide (emitted as a notes page with a notes master)
  and round-trip.
- Streaming I/O: `pptx.OpenStream(path)` and `Presentation.SaveStream(path)`
  read and write decks through the streaming package without buffering the
  whole file.
- Rich text: `Slide.AddTextFrame(box)` returns a `TextFrame` with
  `AddParagraph`/`AutoFit`/`Anchor`/`Margins`; `Paragraph` with
  `AddRun`/`AddBreak`/`AddHyperlink`/`Align`/`Indent`/`Bullet`; and a
  token-typed `RunStyle` (type role, color, bold/italic/underline/strike/
  baseline, inline code). Bullets support disc, numbered, and checklist styles;
  run colors resolve against the active theme; inline code renders monospace
  with a subtle themed tint; hyperlinked runs carry their URL through an
  external relationship.
- Speaker notes are now rich text: `Slide.SpeakerNotes()` returns a `TextFrame`,
  with `Slide.SetSpeakerNotes(text)` as a plain-text convenience.
- `scene.Render` now composes the text-heavy leaf nodes — `hero`, `prose`,
  `heading`, `list`, `divider`, `quote`, `callout`, `chip`, `arrow`,
  `code_block`, and `section_divider` — onto the builder with a deterministic
  top-level layout, populating `Stats`. Text sizes are rendered verbatim from
  the theme (no boosting). Other leaf nodes not yet rendered surface a
  `LayoutWarning`.
- `scene.Render` composes the `two_column` and `grid` containers: a `scene/layout`
  geometry engine subdivides each container into ratio/column slots and renders
  each child into its slot (nesting composes). A `grid` whose cell count is not a
  multiple of its column count is now a validation error.
- Tables: `Slide.AddTable(box, rows, cols)` returns a `Table` with `SetHeaderRow`,
  `SetBanding`, `SetColumnWidths`, and `Cell(row, col)`. A `Cell` has
  `TextFrame`/`SetText` (rich-text cells), `SetFill`, `SetBorders`, and
  `MergeRight`/`MergeDown`. Header rows and banding emit concrete alternating
  fills. `scene.Render` composes the `table` node (with an optional `Caption`
  above it).
- Graphic frames now emit the correct PresentationML `<p:xfrm>` transform.
- Shape corner radius: `Slide.AddShape(ShapeRoundRect, box, WithRadius(role))`
  rounds a rectangle's corners from a theme radius token (`RadiusNone`…
  `RadiusFull`). The absolute radius resolves against the active theme and is
  converted to OOXML's `roundRect` adjust at write time, so a theme swap
  re-rounds the same input (P2); `RadiusFull` yields a full capsule (pill). The
  option is ignored on non-`roundRect` geometries.
- `scene.Render` composes slides concurrently across a worker pool
  (`runtime.GOMAXPROCS(0)` by default, configurable with
  `scene.WithWorkers(n)`); output is byte-identical regardless of worker count.
  `Stats.Timings` now reports per-slide composition time (`SlideTiming`) in
  scene order.
- Template ingestion (brand kits): an opened deck exposes its theme via
  `Theme()` and its slide masters/layouts via `Masters()` (`Master`/`Layout`),
  and `pptx.New(pptx.FromTemplate(brand))` seeds a new presentation from a
  template — adopting its theme, masters, and layouts (the template's parts are
  cloned and its slides stripped). On the scene side, `scene.WithTheme(theme)`
  renders against a brand theme and `scene.WithLayoutMap(m)` maps each slide's
  `LayoutKind` to a named template layout (`scene.DefaultLayoutMap` covers
  PowerPoint's standard names); an unknown layout falls back to the blank layout
  with a `LayoutWarning`.
- Reading back authored decks: `pptx.NewFromBytes` / `OpenStream` reconstructs a
  navigable model, not just the bytes. `Slide.Shapes()` enumerates a reopened
  slide's shapes, and each `Shape` exposes read accessors — `Geometry`,
  `Rotation`, `Fill` (`Kind`/`SolidColor`/`Gradient`), `Line`, `Shadow`,
  `TextFrame`, `Table`, and `Image`. A reopened `TextFrame` yields `Paragraphs` →
  `Runs` with their resolved style, color, bullet, alignment, and hyperlink
  target, plus frame-level `AutoFitMode` / `VerticalAnchor` / `MarginInsets`; a
  reopened `Table` yields its rows/columns, header/banding, per-cell text, fill,
  and merge spans (`GridSpan` and `RowSpan`); a reopened `Image` yields its alt
  text, crop, fit, rotation, opacity, and embedded bytes. Every shape, run, fill,
  line, table, and image pptx-go emits round-trips back into the same model, and a
  self-authored deck reopens byte-identically.
- Best-effort reading of third-party decks: opening a deck pptx-go did not
  author (PowerPoint, Keynote export, another library) no longer fails or
  panics on content it cannot model. `Presentation.ReadWarnings()` reports each
  degradation — an unrecognized shape-tree element ignored at parse time
  (`WarnDroppedElement`), or a referenced part that was missing, dangling, or
  unparseable and was skipped (`WarnUnreadablePart`) — while every part pptx-go
  does not model passes through unchanged on re-save. Dropped-element warnings
  now also cover nested unmodeled content (e.g. an `<a:fld>` field inside a shape's
  or a table cell's text body), and a theme part that exists but cannot be parsed degrades to a
  `WarnUnreadablePart` (the deck keeps the default theme) rather than failing the
  open. Fidelity preservation of unrecognized content is not promised (D-048); a
  self-authored deck reports no warnings.
- Read constructors (`NewFromBytes`, `NewFromFile`, `OpenStream`) now accept
  options: `WithLogger` makes read-time degradation visible to logs (a `Warn`
  event per dropped element / skipped part, mirroring `ReadWarnings`), and
  `WithReadPartLimit(n)` overrides the per-part size ceiling (default 100 MB;
  `n <= 0` disables it). (D-049.)
- Caller-driven scene layout and composition mechanisms (each additive; a zero
  value reproduces the prior render): content-aware text height with truthful
  overflow warnings (D-051); `VAlignFill` grows the flexible nodes to fill the
  frame (D-052); opt-in slide chrome — `Scene.Chrome` (brand slot + page total)
  with `SceneSlide.Section` / `PageNumber` rendering a section eyebrow and an
  `N / total` footer outside the body region (D-053); rich `Card` visuals
  `HeaderFill` / `StatusDot` (`*ColorRole`) and `Watermark` (D-054); a
  `TwoColumn` seam element via `ColumnJoin` (`JoinBadge` / `JoinArrow`) +
  `JoinLabel` (D-055); a row-labeled `Bento` grid node (`BentoRow` / `BentoCell`
  with variable column spans) (D-056); a `Stat` leaf node — a hero number with a
  label and an optional `Delta` toned by `DeltaTone` (D-057); and resolved
  per-slide colors in `Stats.Colors` (`SlideColors`, the derived dark palette for
  a dark-variant slide) so callers can run their own contrast checks (D-058).

### Fixed

- Speaker notes now round-trip: a reopened deck's notes are reconstructed into a
  navigable `SpeakerNotes()` text frame. Previously notes were invisible after
  reopen, and inspecting `SpeakerNotes()` then saving destroyed the existing
  notes — both fixed. (D-050.)

### Security

- Opening a deck is now memory-bounded and zip-slip-safe by default (CLAUDE.md
  §7): a part whose decompressed size exceeds the per-part limit (default
  100 MB, configurable via `WithReadPartLimit`) is rejected with
  `opc.ErrPartTooLarge` rather than allocated, and a ZIP entry whose path escapes
  the package root (absolute or containing `..`) is rejected at parse time with
  `opc.ErrUnsafePartPath`. Both eager and streaming opens are covered, and the
  external-ingest parse surfaces (`opc.Open`, `presentation.FromXML`, the rels
  and content-types parsers) gained fuzz targets. (D-049.)

- Slide layouts read from a deck now carry their name and type (the layout
  parser previously discarded both), so layouts can be selected by name.
- Saving a presentation is now deterministic: the same presentation written
  twice produces byte-identical bytes. ZIP entries carry a fixed timestamp
  instead of the wall clock, and the content-types and embedded-media parts are
  emitted in a stable order (previously each save differed, breaking
  snapshot-based tests).

- `Slide.AddPictureFromFile` and `AddPictureFromBytes` now embed the image
  bytes and wire the relationship correctly (previously the file path read was
  a stub and image relationships were not emitted).
- The streaming reader now preserves the package-level relationship, so a deck
  opened with `OpenStream` re-saves into a valid file.
- Opening a deck (`Open`/`NewFromBytes`/`NewFromFile`/`OpenStream`) now rebuilds
  its slide and section models, so an opened presentation can be read, edited,
  and re-saved losslessly. Previously `Slides()` returned nothing and sections
  were dropped on re-save.
- `AddSlideAt` now inserts the slide at the requested position in the emitted
  slide list (previously it was appended, so the on-disk order didn't match).
- `RemoveSlide` now drops the slide's presentation relationship and notes part,
  so removing a slide no longer leaves a dangling relationship.
- Images added to a reopened deck no longer collide with existing media names.
- Generated decks no longer trigger PowerPoint's "repair" prompt. Two schema
  violations were emitted: a table cell used `<p:txBody>` (must be the DrawingML
  `<a:txBody>` inside the `a:`-namespaced table), and the notes-master list used
  `<p:sldMasterId>` (must be `<p:notesMasterId>`, with only `r:id`). Both are now
  correct, and the schema-validity layer (ISO/IEC 29500 XSDs) is vendored and
  active in CI so this class of bug is caught automatically. Speaker-notes text
  bodies also no longer carry redundant `xmlns` declarations.
- `New()` now seeds the presentation-level parts PowerPoint expects —
  `presProps.xml`, `viewProps.xml`, `tableStyles.xml`, and `docProps/core.xml` +
  `app.xml` — with their relationships and content types. A deck missing these
  opened but prompted to "repair" (notably `tableStyles.xml`, which a table's
  `tableStyleId` references).
- Decks with sections or speaker notes no longer prompt PowerPoint to repair.
  Section GUIDs are now well-formed, non-nil v4-shaped values — the implicit
  "Default Section" emitted the nil GUID (`{00000000-…-000000000000}`), which
  PowerPoint rejects. The notes master now references its own theme part
  (`theme2.xml`) instead of sharing the slide master's `theme1.xml`; PowerPoint
  repaired the shared-theme case by splitting off a `theme2.xml` itself.
- Embedded images now carry a shape geometry (`<a:prstGeom prst="rect">`).
  Without it the picture had no region for the blip to fill, so renderers other
  than PowerPoint (macOS Quick Look, Keynote, LibreOffice) drew nothing — the
  image bytes were embedded and wired correctly but appeared blank.

### Changed

- Module path is now `github.com/hurtener/pptx-go`.
- Project is licensed under Apache-2.0; the upstream MIT license is
  preserved at `LICENSE.upstream`.
- Slide and presentation XML is emitted with correct namespaces and
  attributes, so emitted decks pass OPC conformance and open cleanly.
- Builder shape methods (`Slide.AddTextBox`, `AddRectangle`, `AddEllipse`,
  `AddRoundRect`, `AddAutoShape`, `AddPicture`) now take coordinates and sizes in
  **EMU** rather than pixels. Compute them with `pptx.In`, `pptx.Cm`, `pptx.Pt`,
  `pptx.Px`, or a `pptx.Box`.
- `Slide.AddTable` now takes a `pptx.Box` and returns a `*Table` (was
  `(x,y,cx,cy,rows,cols int) *XGraphicFrame`); `Slide.SetTableCellText` is
  replaced by `Table.Cell(r,c).SetText(...)`.

### Deprecated

- `Presentation.SetFontSource` — pass `pptx.WithFontSource` to `pptx.New`.
- `pptx.PxToEMU` — use `pptx.Px`, which returns a typed `EMU`.

### Removed

- The inherited concrete `Color` struct and its ecosystem (`ColorMap`,
  `ParseColor`, named `Color*` presets, `RGBColor`/`SchemeColor`,
  `Slide.ValidateColor`/`ResolveColor(string)`), superseded by the `Color`
  interface (`RGB`/`RGBA`, `TokenColor`/`TokenTextColor`).

[Unreleased]: https://github.com/hurtener/pptx-go/commits/main
