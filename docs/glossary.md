# Glossary

> pptx-go vocabulary. New terms land here in the same PR that introduces
> them. Cross-referenced from `RFC-001-pptx-go.md`, `CLAUDE.md`, and the
> phase plans.
>
> **Ordering policy.** Alphabetical, case-insensitive. Cross-references use
> backticks for the term being defined: `Builder`, `Theme`.

---

## Anchor

A point on a shape (top-left, top-center, …, bleed-top-right, …) used to
position a `Decoration` or attach a `Connector` endpoint. The IR refers to
anchors by name; the `scene` renderer translates anchor + offset into EMU
coordinates at render time. Distinct from a `LayoutSlot`, which is the
region a layout engine assigns to a content node.

## Asset

Any non-OOXML byte payload referenced by content: image bytes, chart PNG,
icon SVG, ornament recipe input. Assets are passed by `AssetID` through
the scene IR; the scene's asset registry resolves the ID to bytes at render
time.

## AssetID

`scene.AssetID` — a free-form string identifier the IR uses to refer to
asset bytes (`image.asset_id`, `chart.asset_id`, etc.). Resolved to
bytes at render time by an `AssetResolver`. pengui-slides uses
`asset://<UUID>` URIs; pptx-go is scheme-agnostic.

## AssetResolver

`scene.AssetResolver` — the caller-injected interface that maps an
`AssetID` to bytes + content type. Called lazily during render. A
missing asset surfaces as a `LayoutWarning` (or render-fatal if the
asset was required). See `RFC §10.6`.

## Bleed

A `Decoration` placement that extends past the slide canvas edge.
Implemented via negative `<a:off>` coordinates in OOXML (PowerPoint
accepts these for partial-shape placement). Bleed `Anchor` names start
with `bleed_*`.

## Block (Block node)

A scene IR node that's not a container — a `Leaf`. The catalog of leaves
is documented in `RFC-001-pptx-go.md §11.1`.

## Alignment

A paragraph's horizontal alignment on the builder (`pptx.AlignLeft`,
`AlignCenter`, `AlignRight`, `AlignJustify`). Set via `Paragraph.Align` or
`ParagraphOpts.Align`; maps to the OOXML `algn` attribute. (RFC §8.4.)

## AutoFitMode

A `TextFrame`'s text-fit behavior (`pptx.AutoFitNone`, `AutoFitNormal`
[shrink font], `AutoFitShape` [grow shape]). Set via `TextFrame.AutoFit`.
(RFC §8.4.)

## Box

`pptx.Box{X, Y, W, H int64}` — a rectangle in EMU coordinates. The
builder's universal placement primitive. Distinct from `LayoutSlot`,
which is the higher-level abstract region assigned by the scene's layout
engine.

## Builder

The `pptx` package's public API. Layer 1 of the library. Theme-aware,
token-typed, OOXML-free. Composes `internal/opc` and `internal/ooxml`. The
single substrate for all higher-level content authoring. See
`RFC-001-pptx-go.md §8`.

## Card

A scene IR container node — accent strip + optional icon / eyebrow /
header-pill, with leaf children. Renders as a native PPTX shape group.
See `RFC-001-pptx-go.md §11.2`.

## CardSection

A scene IR top-level container with card chrome that can accept *non-leaf*
children (grid, two_column, nested cards). Distinct from `Card` to avoid
recursive type cycles; `Card` body is leaf-only.

## Chart node

A scene IR leaf representing a data chart. V1 disposition: image-shape
(caller-rendered bytes). V2 disposition: native `c:chart` parts. See
`RFC-001-pptx-go.md §15`.

## Color (interface)

A sealed builder interface for a write-time-resolved color: either a literal
RGB (`pptx.RGB`, `pptx.RGBA`) or a theme token (`pptx.TokenColor`,
`pptx.TokenTextColor`). A token resolves against the active `Theme` when
applied, so a theme swap re-renders the same input in the new palette (P2;
D-012, D-030, D-033). The interface is sealed — callers cannot supply a color
type the codec can't emit.

## ColorRole

A semantic color role (e.g. `canvas`, `surface`, `accent`, `accent_warm`,
`success`). Page-level surfaces resolve via the active `Theme`'s
`ColorPalette` to an OOXML color value. See `RFC-001-pptx-go.md §7.1`.

## Container node

A scene IR node that contains child nodes (`two_column`, `grid`, `card`,
`card_section`). See `RFC-001-pptx-go.md §11.2`.

## CGo

Go's C-interop facility. The shipped pptx-go artifact compiles with
`CGO_ENABLED=0` (P4). `-race` tests are the one exception (the race
detector requires CGo).

## Coverage band

The minimum per-package statement-coverage percentage enforced
mechanically by `internal/coveragecheck` against
`internal/coveragecheck/coverage.json`. A package below its band fails
`make coverage`. Class defaults live in `CLAUDE.md §11`; a band override
records a class + reason. See also `Drift audit`, `Preflight`.

## Conformance (validity layers)

How pptx-go proves emitted decks are *valid*, not merely round-trippable
(D-031): **(1)** `internal/conformance` — pure-Go OPC integrity (content
types, resolved relationships, no dangling `rId`, pack URIs); **(2)**
`xmllint` schema validation against vendored ISO 29500 transitional XSDs;
**(3)** a LibreOffice headless **open-proxy** CI job; **(4)** a manual
per-wave PowerPoint check. Layers 1–3 are automated/CI; layer 4 is the
ground truth they approximate.

## Decoration

A scene IR leaf that places a curated `Ornament` or asset image at a slide
position (in-canvas or bleed). Pure visual; carries no text. Layer is
`background` (renders behind body) or `foreground` (renders on top).

## Drift audit

The `scripts/drift-audit.sh` design-coherence gate (`make drift-audit`),
run as part of `Preflight`. It mechanically checks file mirroring
(`AGENTS.md == CLAUDE.md`), the canonical module path, the P1/P3 layering
seams, and (from Phase 20) the §19 user-facing-vocabulary rule. Checks
grow as phases land.

## EMU

English Metric Unit. OOXML's canonical length unit. `1 inch = 914,400
EMU`. Integer, no floating-point. `pptx/units.go` provides conversion
helpers (`pptx.Pt`, `pptx.Cm`, `pptx.In`, `pptx.Px`).

## Eyebrow

A small label rendered above a heading or card body (e.g. "01 · TRAZABILIDAD"
in pengui-slides reference decks). Plain `RichText`. Not a distinct IR node;
a field on `Hero`, `Card`, `CardSection`.

## Fill

A shape's interior fill (builder). V1 ships `pptx.SolidFill(Color)` and
`pptx.NoFill()`; a fill resolves its `Color` against the active `Theme` when
applied. Gradient, pattern and picture (blip) fills are tracked for later.

## Flow node

A scene IR leaf for a sequential step pipeline (horizontal/vertical) with
a connector glyph between adjacent steps. See `RFC-001-pptx-go.md §11.1`.
Renders as native PPTX shape group.

## FontSource

`pptx.FontSource` — the caller-injected interface that resolves a
font name + style + weight to bytes. Registered via
`pptx.WithFontSource(...)`. The presentation's `EmbedFont(name, style,
weight)` method uses it to write font-embedding parts. No auto-embed
default: the caller invokes `EmbedFont` explicitly for each font to
embed. (D-019, `RFC §7.6`.) Registered in V1 via
`pres.SetFontSource(...)`; the functional `pptx.WithFontSource(...)`
option arrives with the builder spine (D-030).

## FontSpec

A resolved typography value: font `Family`, `Size` (points), `Weight`
(100–900; ≥600 is bold), and `Italic`. `Theme.ResolveType(TypeRole)`
returns one.

## Format

`pptx.Format` — a first-class enum on `*Presentation` that selects
slide dimensions, master defaults, and theme-default font sizes. V1
ships `Slides16x9` (default) and `Slides4x3`; print formats
(`PrintA4Portrait`, `PrintLetterPortrait`) are out of pptx-go's scope
(D-026 — document concerns). See `RFC §5`.

## Frame chrome

A curated device/browser bezel (browser / phone / desktop / laptop)
rendered as native PPTX shapes wrapping an inner image. Applied via the
`Image.Frame` field. See `RFC-001-pptx-go.md §14.3`.

## Icon

A curated lucide-style glyph in the `assets/icons/` registry. Rendered as
native PPTX shape paths (NOT as raster). Referenced by name from IR nodes
that accept an icon (`card`, `flow` steps, `header_pill`). Caller-supplied
icons go through the same translator (`scene.WithIconExtension`).

## ImageSource

`pptx.ImageSource` — sealed interface for image input to `Slide.AddImage`,
constructed with `pptx.ImageFile(path)`, `pptx.ImageBytes(data, mime)`, or
`pptx.ImageReader(r, mime)` (the §4.4 interface + factory + driver seam). The
bytes are verified against a known image signature (PNG/JPEG/GIF/BMP/WebP) —
mismatched or malformed input is rejected — but pixels are never parsed (§7).
Identical bytes across a deck are written once (dedup). (D-022 sibling; `RFC
§8.6`.)

## IR

Intermediate representation. In pptx-go, "the scene IR" is the typed Go
struct catalog under `scene/nodes.go` consumed by `scene.Render`. The IR is
a strict superset of pengui-slides v4's IR (`RFC-001-pptx-go.md §21`).

## Layer (in Decoration)

`background` (renders behind body content) or `foreground` (renders on
top). The `scene` layout engine processes background decorations before
body nodes and foreground decorations after.

## LayoutKind

A scene-level enum naming the slide's structural intent (`cover`,
`title-content`, `two-column`, `card-grid`, `full-bleed`, …). Maps to the
template's named master layouts via a `LayoutMap`.

## LayoutSlot

The abstract region assigned by the scene's layout engine to a content
node. Resolved to a `pptx.Box` at composition time. Containers introduce
sub-slots.

## LayoutWarning

A non-fatal layout issue surfaced in `Stats.Warnings` (e.g. content
overflow, aspect-ratio mismatch for a chart, missing optional asset).
A V1.x `strict` mode upgrades warnings to errors.

## Leaf

A scene IR node that doesn't contain children. Listed in
`RFC-001-pptx-go.md §11.1`. Opposite of `Container`.

## Line

A shape's outline (builder): width (EMU), `Color`, and optional preset dash.
Like `Fill`, its color resolves against the active `Theme`.

## Mirror (file mirroring)

`AGENTS.md` and `CLAUDE.md` are kept verbatim identical. CI's `make
check-mirror` enforces this; the §18 rule in `CLAUDE.md` is the
specification.

## OOXML

Office Open XML, ISO/IEC 29500. The XML schemas above OPC that describe
PPTX/DOCX/XLSX content. pptx-go targets the **transitional** profile.

## OPC

Open Packaging Convention, ISO/IEC 29500-2. The ZIP-container layer with
`[Content_Types].xml`, `_rels/`, pack-URI semantics. Shared by PPTX, DOCX,
XLSX. Lives under `internal/opc`.

## Ornament

A curated preset decoration shape recipe in the `assets/ornaments/`
registry: `glow_ring`, `radial_glow`, `grid_dots`, `corner_bracket`,
`chevron_arrow`, `noise_overlay`. Rendered as native PPTX shapes in the
active accent token.

## Part

An OOXML logical content unit inside the OPC package: a slide, a master,
the theme, an image, the core props XML. Identified by a pack URI.
`internal/opc.Part` is the Go model.

## Part family

An OOXML part grouping that owns one `internal/ooxml` subpackage:
`presentation`, `slide`, `theme`, `core`, `chart`, `relations`, `media`,
`drawing`. Families stay independent — a subpackage imports another only
via shared helpers in the `internal/ooxml` root package (namespace URIs,
`StripNamespacePrefixes`), so a spec bump in one family is localized
(RFC §6.2).

## NodeKind

The discriminator for a scene `SlideNode` (`scene.NodeKind`): a typed
constant (`KindHero`, `KindCard`, …) with an IR name via `String()`.
Used by validation and rendering to switch on a node's concrete type.

## Per-node rendering policy

The decision per scene IR node type about whether the node renders as
native PPTX shapes (most nodes) or as a `pic` shape with caller-
supplied bytes (nodes whose IR carries an `asset_id` field: `image`,
`chart`, `decoration` of `asset_ref` kind, `code_block`). Intrinsic
to the node type — not a runtime enum, not a per-deck toggle. Encoded
as `scene.Policy` / `scene.PolicyFor(kind)` and asserted against the
node structs by `policy_test.go`. See `RFC §12`, D-018.

## Phase plan

A `docs/plans/phase-NN-*.md` file that specifies a chunk of implementation
work. Authored from `docs/plans/_template.md`. Acceptance criteria are
binding. See `CLAUDE.md §16`.

## Preflight

The local gate (`make preflight`) that runs build + per-phase smoke +
drift-audit. The pre-commit hook and CI enforce it. Non-negotiable.

## BulletKind

A paragraph bullet style on the builder (`pptx.BulletNone`, `BulletDisc`,
`BulletNumber`, `BulletCheckbox`). Set via `Paragraph.Bullet`; emits the
OOXML `buChar`/`buAutoNum`/`buNone` with a hanging indent. (RFC §8.4.)

## Paragraph

A line block within a `TextFrame` (`pptx.Paragraph`): alignment, indent
level, an optional bullet, and an ordered sequence of `Run`s and breaks.
Created via `TextFrame.AddParagraph`. (RFC §8.4.)

## Presentation

`pptx.Presentation` — the top-level builder type. Owns slides, sections,
masters, the theme, and the OPC package. Created via `pptx.New` or
`pptx.Open[Stream]`.

## PptxSection

`pptx.Section` — slide grouping primitive in PowerPoint (the
`sectionLst` element on the presentation). A `*Section` has a name and
a list of slides; created via `pres.AddSection(name)`. **Distinct from
the scene IR's `section_divider` node**, which is a slide whose content
is a full-bleed chapter break. A scene `section_divider` may or may not
be inside a `pptx.Section`. (D-021, `RFC §8.7`.)

## Preset (ornament name)

A name in the closed `PresetOrnamentName` enum identifying one of the
curated ornaments. Distinct from the open-name caller-supplied asset path
under a `Decoration` source `asset_ref`.

## RepairPromptHygiene

An always-on XML post-processor that strips known PowerPoint
"repair-prompt" triggers from emitted OOXML (empty `lang=""`,
malformed namespace declarations, …). Lives in
`internal/render/hygiene.go`. **Not configurable** — emitting OOXML
that PowerPoint accepts cleanly is correctness, not preference.
Trigger list documented in `docs/design/HYGIENE.md`. (D-020.)

## RFC

`RFC-001-pptx-go.md`. The design source of truth. Authoritative source
when conflicts arise (`CLAUDE.md §2`).

## RichText

`[]TextRun` — pptx-go's text model. Each run carries plain text + an
inline style + an optional `TextColorRole` token. Same model in `pptx`
(builder) and `scene` (renderer).

## Run

A styled text span within a `Paragraph` (`pptx.Run`), created via
`Paragraph.AddRun(text, RunStyle)`. The builder analogue of a scene
`TextRun`. (RFC §8.4/§9.)

## RunStyle

The token-typed styling of a `Run` (`pptx.RunStyle`): a `TypeRole`
(typography), a `Color` (token or literal), and bold/italic/underline/
strike/baseline/code flags. Tokens resolve against the active theme when
the run is added. (RFC §8.4; D-013 for `Code`.)

## Round-trip fidelity

The V1 guarantee that PPTX files pptx-go authors can be parsed back into
the same Go model losslessly. Third-party PPTX is best-effort, not a
contract. See `RFC-001-pptx-go.md §16`.

## Scene

A `scene.Scene` value: a theme + an ordered list of `SceneSlide`s + scene
metadata. The input to `scene.Render`.

## Scene renderer

The `scene` package's public API. Layer 2 of the library. Consumes a
typed `Scene`; emits via the `Builder`. Never reaches under the builder
(P1).

## SceneSlide

A single slide in a `Scene`: a layout kind, a list of top-level
`SlideNode`s, optional notes, an optional theme variant.

## section_divider

A scene IR leaf node representing a slide whose content is a single
full-bleed chapter break (label + optional ornament). **Distinct from
`PptxSection`** (the OOXML slide-grouping primitive). A
`section_divider` slide can be inside a `PptxSection` or not.

## ShapeGeometry

A preset shape outline on the builder (`pptx.ShapeRect`, `ShapeEllipse`,
`ShapeRoundRect`, …), carrying the OOXML preset-geometry (`prst`) name.
Passed to `Slide.AddShape(geom, box, …)` with a `Box` (EMU) and optional
`Fill`/`Line`.

## SlideNode

The sealed scene IR union (`scene.SlideNode`): every leaf and container
node implements it. Closed to the `scene` package (an unexported marker)
and discriminated by `NodeKind`. The catalog is `RFC §11.1` (leaves) and
`§11.2` (containers): `Hero`, `Prose`, `Heading`, `List`, `Divider`,
`Quote`, `Callout`, `Image`, `Chip`, `Arrow`, `CodeBlock`, `Chart`,
`Table`, `Flow`, `Decoration`, `SectionDivider`, `TwoColumn`, `Grid`,
`Card`, `CardSection`.

## SlideDocument

A pengui-slides v4 term for the compiled "shape inventory" representation
of a slide (post-IR-compilation, pre-export). In pptx-go's vocabulary
the analogous concept is **`Scene`** + the per-slide rendering state
inside `scene.Render`. Listed here only to disambiguate: a pengui-slides
"SlideDocument" maps to pptx-go's "Scene + Stats + emitted shapes",
not to a single pptx-go type.

## slog

`log/slog` — Go's structured logging stdlib package. pptx-go accepts an
optional `*slog.Logger` via `pptx.WithLogger` / `scene.WithLogger`. No
logger = no logs (zero cost).

## Smoke check

`scripts/smoke/phase-NN.sh` — a per-phase shell script that verifies the
phase's acceptance criteria mechanically. SKIPs gracefully when its
surface isn't built yet. The §4.2 contract requires `OK ≥ count(criteria)`
and `FAIL = 0` for a phase to be "done".

## SpeakerNotes

A per-slide RichText frame emitted into the slide's `notesSlide` part.
Accessed via `slide.SpeakerNotes() *TextFrame`. V1 (D-022). The scene
IR's `SceneSlide.Notes` field maps directly.

## Stage 1 / Stage 2 validation

Scene-side validation phases. Stage 1: structural correctness (well-formed
IR, field constraints). Stage 2: mode constraint + token resolution +
asset resolution. See `RFC-001-pptx-go.md §10.4`.

## Stats

The struct returned by `scene.Render` carrying per-slide render time,
shape counts, asset counts, and warnings. The library's observability
surface (we have no `obs/v1` protocol; pptx-go is a library, not a
service).

## Streaming

`pptx.OpenStream` and `pptx.SaveStream` — lazy-load and streaming-save
modes for large decks. The upstream's streaming primitives (in
`internal/opc/`) are first-class V1.

## Subsystem

An RFC-level grouping of related code (OPC, OOXML, Builder, Scene
renderer, Theme, Assets, Charts). Phase plans name their owning subsystem.

## Theme

The token-to-OOXML mapping the builder consults at write time. Tokens are
resolved lazily; theme swaps re-render the same builder/scene input in the
new visual language. A `Theme` holds five per-role value maps —
**ColorPalette** (surface + text colors), **Typography** (`TypeRole` →
`FontSpec`), **Spacing** (`SpaceRole` → `EMU`), **Radii** (`RadiusRole` →
`EMU`), and **Elevations** (`ElevationRole` → `Elevation` shadow spec) —
plus the major/minor font faces. See `RFC-001-pptx-go.md §7`,
`docs/design/THEME.md`.

## Table (builder)

`pptx.Table` / `pptx.Cell` — the builder's native table API
(`Slide.AddTable(box, rows, cols)`). Supports header rows, banding, merged
cells (`Cell.MergeRight`/`MergeDown`), per-cell fills/borders, and rich-text
cells (`Cell.TextFrame`). Renders as an OOXML `tbl` in a graphic frame.
**Distinct from** the scene IR's `Table` node, which the scene renderer
composes onto this builder (with an optional caption above). (RFC §8.5.)

## TextColorRole

A semantic text color role (`primary`, `secondary`, `tertiary`, `inverse`,
`muted`, `accent`, `accent_alt`, `success`, `warning`, `error`). Applied
to inline `TextRun`s.

## TextFrame

A shape-level rich-text container on the builder (`pptx.TextFrame`):
auto-fit, vertical anchor, margins, and an ordered list of `Paragraph`s.
Created via `Slide.AddTextFrame`; also backs `Slide.SpeakerNotes`. The
`pptx` half of the shared rich-text model (RFC §8.4/§9).

## TextRun

A single run of text in a paragraph: plain text + inline style +
`TextColorRole`. The atomic unit of `RichText`.

## Token

A semantic role (color, type, space, radius, elevation) consumed by
builder calls. Resolves to a concrete OOXML value via the active `Theme`.
The canonical authoring path for visual properties (P2).

## TwoColumn

A scene IR container with two cells (`left`, `right`) and a ratio
(`1:1` / `1:2` / `2:1`). Renders as a layout-only container with no
shape of its own; cells host leaf children.

## V1 / V2

Major version milestones. V1.0.0 is the first stable release; V2 is the
next major. V1 promises Layer 1 + Layer 2 + image-shape charts +
round-trip self-authored. V2 adds native `c:chart`, third-party PPTX
read fidelity, animations, transitions, broader OOXML coverage.

## Variant (theme variant)

A named theme variant (`light`, `dark`, `print`, brand-specific). The
scene's `Variant` field selects between variants the caller registered.
Per-slide theme overrides are V2; per-scene variant is V1.

## Wave

A grouping of phases sharing a milestone. See `docs/plans/README.md` for
the V1 wave structure.

---

*Add new terms in alphabetical order. Keep entries terse — link to the
RFC section that owns the concept, don't re-explain.*
