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
  the theme (no boosting). Container and other leaf nodes are not yet rendered
  and surface a `LayoutWarning`.

### Fixed

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

### Changed

- Module path is now `github.com/hurtener/pptx-go`.
- Project is licensed under Apache-2.0; the upstream MIT license is
  preserved at `LICENSE.upstream`.
- Slide and presentation XML is emitted with correct namespaces and
  attributes, so emitted decks pass OPC conformance and open cleanly.
- Builder shape methods (`Slide.AddTextBox`, `AddRectangle`, `AddEllipse`,
  `AddRoundRect`, `AddAutoShape`, `AddPicture`, `AddTable`) now take
  coordinates and sizes in **EMU** rather than pixels. Compute them with
  `pptx.In`, `pptx.Cm`, `pptx.Pt`, `pptx.Px`, or a `pptx.Box`.

### Deprecated

- `Presentation.SetFontSource` — pass `pptx.WithFontSource` to `pptx.New`.
- `pptx.PxToEMU` — use `pptx.Px`, which returns a typed `EMU`.

### Removed

- The inherited concrete `Color` struct and its ecosystem (`ColorMap`,
  `ParseColor`, named `Color*` presets, `RGBColor`/`SchemeColor`,
  `Slide.ValidateColor`/`ResolveColor(string)`), superseded by the `Color`
  interface (`RGB`/`RGBA`, `TokenColor`/`TokenTextColor`).

[Unreleased]: https://github.com/hurtener/pptx-go/commits/main
