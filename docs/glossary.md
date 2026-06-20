# Glossary

> pptx-go vocabulary. New terms land here in the same PR that introduces
> them. Cross-referenced from `RFC-001-pptx-go.md`, `CLAUDE.md`, and the
> phase plans.
>
> **Ordering policy.** Alphabetical, case-insensitive. Cross-references use
> backticks for the term being defined: `Builder`, `Theme`.

---

## Accent stripe

The thin themed bar a `Card` / `CardSection` draws along one edge (the accent
token color) to mark the card visually. Part of the card chrome. Distinct
from a `border` (which outlines the whole card). See `RFC-001-pptx-go.md §11.2`.

## Agent skill

A `skills/<name>/SKILL.md` file (agentskills.io format: YAML frontmatter +
instructions) that teaches an AI coding agent one pptx-go workflow — e.g.
scaffold a presentation, define a Theme, compose a scene. Keeping skills in
sync with the user-facing surface is binding repo hygiene (`CLAUDE.md §19`);
drift is a defect. Shipped in Phase 20.

## Anchor

A point on a shape (top-left, top-center, …, bleed-top-right, …) used to
position a `Decoration` or attach a `Connector` endpoint. The IR refers to
anchors by name; the `scene` renderer translates anchor + offset into EMU
coordinates at render time. Distinct from a `LayoutSlot`, which is the
region a layout engine assigns to a content node.

## Aspect warning

The `LayoutWarning` a `chart` raises when the chart image's aspect ratio
diverges from its assigned slot beyond the threshold (15%). The image's
dimensions are read from its header (`image.DecodeConfig`, not pixel data —
D-046); the chart is contained-to-fit and the warning informs the caller. See
`RFC-001-pptx-go.md §15.1`.

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

A `Decoration` placement that extends past the slide canvas edge,
selected by the `Decoration.Bleed` field (D-041). It relaxes the
on-canvas clamp so the anchor + offset may place the ornament box partly
off-slide, implemented via negative `<a:off>` coordinates in OOXML
(PowerPoint accepts these for partial-shape placement).

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

## Brand kit

A `.pptx` template carrying a populated theme + ≥1 master with layouts, used
to seed a presentation's visual identity (RFC §13.4). pptx-go *consumes* brand
kits (opens one, adopts its theme + masters via `FromTemplate`); authoring a
hand-editable template is V1.x. See `Master`, `FromTemplate`.

## Builder

The `pptx` package's public API. Layer 1 of the library. Theme-aware,
token-typed, OOXML-free. Composes `internal/opc` and `internal/ooxml`. The
single substrate for all higher-level content authoring. See
`RFC-001-pptx-go.md §8`.

## Card

A scene IR container node — accent strip + optional icon / eyebrow /
header-pill, with leaf children. Renders as a native PPTX shape group.
See `RFC-001-pptx-go.md §11.2`.

## Card chrome

A card's non-content shapes: the background rounded-rect (with `fill`,
`border_style`, and `elevation` shadow), the accent stripe, and the header row
(optional icon + eyebrow + header + header-pill). Shared by `Card` and
`CardSection`; the body region renders inside it. See `RFC-001-pptx-go.md §11.2`.

## CardSection

A scene IR top-level container with card chrome that can accept *non-leaf*
children (grid, two_column, nested cards). Distinct from `Card` to avoid
recursive type cycles; `Card` body is leaf-only.

## Chart node

A scene IR leaf representing a data chart. V1 disposition: image-shape
(caller-rendered bytes). V2 disposition: native `c:chart` parts. See
`RFC-001-pptx-go.md §15`.

## Chart placeholder

A labeled bordered chart slot drawn by `pptx.ChartPlaceholder(box)` — a
rounded rect + "Chart" label, no bytes committed. The chart composer reuses it
when an asset is unresolved, so a missing chart shows a labeled slot rather than
a blank gap (D-046). See `RFC-001-pptx-go.md §15.1`.

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

## Connector kind

A `Flow`'s inter-step glyph (`ConnectorKind`): `arrow` (solid, the default),
`arrow_dashed` (a dashed line + chevron head), `cycle` (arrows plus a trailing
return arrow), or `plus` (a `mathPlus` glyph). A flow-level choice applied
between every adjacent step pair. Composes preset shapes — pptx-go does not
build the RFC's anchored `AddConnector` in V1 (D-044). See `RFC §11.1`.

## Container node

A scene IR node that contains child nodes (`two_column`, `grid`, `card`,
`card_section`). See `RFC-001-pptx-go.md §11.2`.

## Content-aware height

A text-bearing node's slot height derived from how many lines its text wraps to
(line count × the node's line height), rather than a fixed per-node constant.
The scene renderer computes it deterministically via the `Wrapped-line estimate`
so stacked nodes don't overlap and overflow is reported truthfully; single-line
content keeps its prior fixed height (byte-identical). See
`RFC-001-pptx-go.md §10.2`.

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

## Crop

A per-edge fractional trim (0..1 from each edge) applied to an image's
source rectangle. `pptx.Crop` on the builder (`Image.SetCrop`); re-exported
as `scene.Crop` and carried on the scene `image` node (D-039). Drives the
OOXML `a:srcRect`. Pure mechanism — no pixel inspection (§7).

## custGeom

OOXML custom path geometry (`a:custGeom`): a shape outline expressed as a
`pathLst` of `path` elements, each an ordered sequence of `moveTo` / `lnTo` /
`cubicBezTo` / `quadBezTo` / `close` commands over the path's own `w×h`
coordinate space (scaled to the shape extent). The wire form an `Icon`
renders to, emitted by `Slide.AddIcon` and produced by the `SVG translator`.
Distinct from preset geometry (`a:prstGeom`, a named `ShapeGeometry`).

## Decoration

A scene IR leaf that places a curated `Ornament` (native) or an asset image
(`pic`) at an anchored slide position. The box is centered on `Anchor` shifted
by `Offset`, sized by `Size`; `Bleed` permits it off the slide edge; `Opacity`
(0..1) and `Rotation` (degrees) style it. Pure visual; carries no text. `Layer`
is `background` (renders behind body) or `foreground` (renders on top) — the
renderer imposes that z-order (RFC §10.2).

## Drift audit

The `scripts/drift-audit.sh` design-coherence gate (`make drift-audit`),
run as part of `Preflight`. It mechanically checks file mirroring
(`AGENTS.md == CLAUDE.md`), the canonical module path, the P1/P3 layering
seams, and (from Phase 20) the §19 user-facing-vocabulary rule. Checks
grow as phases land.

## Elevation primitive

The builder drop-shadow mechanism (`pptx.WithElevation(role)` /
`pptx.WithShadow(e)`) that realizes the `Elevation` token as an OOXML
`<a:effectLst><a:outerShdw>` effect. `WithElevation` resolves the role against
the active theme (P2 token path); `WithShadow` takes a literal `Elevation`
(escape hatch). A flat elevation emits no effect. Not a new token — a mechanism
over the existing `Elevation` role (D-043). See `docs/design/THEME.md`.

## EMU

English Metric Unit. OOXML's canonical length unit. `1 inch = 914,400
EMU`. Integer, no floating-point. `pptx/units.go` provides conversion
helpers (`pptx.Pt`, `pptx.Cm`, `pptx.In`, `pptx.Px`).

## External deck

A PPTX pptx-go did **not** author (PowerPoint, a Keynote export, another
library). Read support for external decks is **best-effort** (RFC §16, D-048):
they open without panicking and report degradation via [Read warning](#read-warning)s,
but round-trip fidelity is not promised — unrecognized shapes are dropped (and
warned), though unrecognized parts pass through unchanged. Contrast a
self-authored deck, which round-trips losslessly (D-047).

## Eyebrow

A small label rendered above a heading or card body (e.g. "01 · TRAZABILIDAD"
in pengui-slides reference decks). Plain `RichText`. Not a distinct IR node;
a field on `Hero`, `Card`, `CardSection`.

## Fit

An image's fill mode: `FitFill` (the default — stretches to fill its box, via
the OOXML `a:stretch`) or `FitNone` (no stretch fill). `pptx.Fit` on the
builder (`Image.SetFit`); re-exported as `scene.Fit` on the scene `image`
node (D-039). Aspect-aware cover/contain are **not** in V1 — they need pixel
dimensions, forbidden by §7.

## Fill

A shape's interior fill (builder). V1 ships `pptx.SolidFill(Color)` and
`pptx.NoFill()`; a fill resolves its `Color` against the active `Theme` when
applied. Gradient, pattern and picture (blip) fills are tracked for later.

## Flexible node

A scene node whose slot grows under `VAlignFill`: the containers (`Grid`,
`TwoColumn`, `Card`, `CardSection`, `Table`) plus the stretchable visuals
(`Image`, `Chart`). Text leaves and atoms are *fixed* (preferred height);
`CodeBlock` is fixed too (growing a code raster distorts it). See `Grow-to-fit`
and `RFC-001-pptx-go.md §10.2`.

## Flow node

A scene IR leaf for a sequential step pipeline (horizontal/vertical) with
a connector glyph between adjacent steps. See `RFC-001-pptx-go.md §11.1`.
Renders as native PPTX shape group.

## Flow step

One pill in a `Flow` (`FlowStep`): a label + optional detail line + optional
icon (a closed-name icon resolved through the curated registry, like a card's).
Rendered as a lighter rounded pill — not the full card chrome (D-044).

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

## Footer page number

The `N / total` page indicator `Slide chrome` draws bottom-right on every
chrome-enabled slide. `N` defaults to the slide's 1-based scene position
(overridable via `SceneSlide.PageNumber`); `total` defaults to the slide count
(overridable via `Scene.Chrome.Total`). See `RFC-001-pptx-go.md §10.2`.

## Format

`pptx.Format` — a first-class enum on `*Presentation` that selects
slide dimensions, master defaults, and theme-default font sizes. V1
ships `Slides16x9` (default) and `Slides4x3`; print formats
(`PrintA4Portrait`, `PrintLetterPortrait`) are out of pptx-go's scope
(D-026 — document concerns). See `RFC §5`.

## Frame chrome

A curated device/browser bezel (browser / phone / desktop / laptop)
rendered as native PPTX shapes wrapping an inner image. Selected on an
`Image` node by the `Frame` enum (`FrameKind`) or, for a caller-extended
frame, the `FrameName` string (D-038). See `RFC-001-pptx-go.md §14.3` and
`Frame recipe`, `Frame registry`.

## Frame recipe

`scene.FrameRecipe` (alias of `scene/frames.Recipe`) — a function
`func(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int)`
that draws a frame's bezel as native shapes into `region` and returns the
`interior` box the renderer inserts the image into, plus the bezel shape
count. It composes the public `Builder` only (P1) and is pure integer-EMU
geometry (deterministic). Curated recipes live in `assets/frames`; callers
register their own via `scene.WithFrameExtension`. See `RFC §14.3`.

## Frame registry

The per-render, closed-name set of `Frame recipe`s a `Render` consults:
the four curated names (`browser`, `phone`, `desktop`, `laptop`) overlaid
with any `scene.WithFrameExtension(name, recipe)` entries (`RFC §14.4`).
Built once before composition and read-only during it (concurrency-safe,
byte-identical — D-035). An `Image` whose resolved frame name is absent
from the registry fails Stage-1 validation (D-038).

## FromTemplate

`pptx.FromTemplate(brand *Presentation)` — a `New` option that seeds a
presentation from a `Brand kit`: it clones the brand's package (theme,
masters, layouts, auxiliary parts) and strips slides, so the new deck inherits
the brand's identity and starts empty (RFC §13.1, D-037).

## Gradient fill

A multi-stop shape fill (`<a:gradFill>`): a list of color stops plus a
direction — `<a:lin>` (linear, by angle) or `<a:path path="circle">`
(radial, by `pptx.RadialGradient`). Added in V1 (D-041) to render the glow
ornaments (`radial_glow`, `glow_ring`) as true gradients rather than banded
solids. Public API: `pptx.LinearGradient` / `pptx.RadialGradient` with
`GradientStop`s. Joins `SolidFill` / `NoFill` as a `Fill`.

## Grow-to-fit

The body-stack layout mode (`VAlignFill`) that, after the fixed leaves take their
preferred height, distributes the leftover body height to the `Flexible node`s so
they grow to consume the frame. The share is proportional to preferred height and
deterministic. A mechanism, not a judgment (D-026, D-052): the caller opts a
slide into fill; the engine never decides a slide looks thin. See
`RFC-001-pptx-go.md §10.2`.

## Header band

A `Card`'s optional colored top region (`Card.HeaderFill *ColorRole`): the
header in an accent color with the body in `Fill` below — distinct from a full
`Fill`. `nil` omits it. Part of the rich card visuals (D-054). See
`RFC-001-pptx-go.md §11.2`.

## Header pill

A small pill-shaped badge (rounded-rect + short label) rendered in a card's
header row, typically right-aligned (e.g. a status tag like "NEW" or "BETA").
A field on `Card` / `CardSection`; part of the card chrome. See
`RFC-001-pptx-go.md §11.2`.

## Icon

A curated lucide-*style* glyph in the `assets/icons/` registry — authored as a
single-path, solid-fill SVG (lucide's own icons are stroke-based multi-element,
so the set is lucide-style, not lucide data; D-040). Rendered as native PPTX
`custGeom` shape paths (NOT as raster) via the `SVG translator`, filled with the
accent token. Referenced by name from IR nodes that accept an icon (`card`,
`flow` steps, `header_pill`). Caller-supplied icons go through the same
translator (`scene.WithIconExtension`); an SVG outside the subset fails at
registration. V1 ships a starter set (~16), growing toward ≈60 (D-005, D-040).

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

## Language badge

The small overlay pill on a `code_block` image showing its source language
(`CodeBlock.Language`), inset top-right over the raster. A native rounded-rect +
caption text, reusing the card header-pill chrome (D-045). Empty `Language` =
no badge. See `RFC-001-pptx-go.md §11.1`.

## Layout

`pptx.Layout` — a read-only view of one slide layout in a template (its name
and the master it belongs to). Built when a deck is opened or seeded with
`FromTemplate`; exposed via `Master.Layouts()`. OOXML-free (P3). See `Master`,
`RFC-001-pptx-go.md §13.2`.

## LayoutKind

A scene-level enum naming the slide's structural intent (`cover`,
`title-content`, `two-column`, `card-grid`, `full-bleed`, …). Maps to the
template's named master layouts via a `LayoutMap`.

## LayoutMap

`scene.LayoutMap` — a `map[LayoutKind]string` that resolves each slide's
`LayoutKind` to a named layout in the active template (RFC §13.2). Passed via
`scene.WithLayoutMap`; `scene.DefaultLayoutMap` maps to PowerPoint's standard
layout names. A name the template lacks falls back to the blank layout and
records a `LayoutWarning` (never an error — D-026).

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

## Master

`pptx.Master` — a read-only view of one slide master and the `Layout`s it
owns, surfaced by `Presentation.Masters()` for a deck opened from a file or
seeded with `FromTemplate`. OOXML-free (P3); the XML wire types stay in
`internal/ooxml`. See `RFC-001-pptx-go.md §13.2`.

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
active accent token, at a caller opacity and rotation (the glows use radial
`Gradient fill`s; `noise_overlay` is a deterministic sparse-dot grain
approximation — D-041). `chevron_arrow` and `corner_bracket` honor `Rotation`
to orient (the bracket snaps to 0/90/180/270 = top-left/top-right/bottom-right/
bottom-left, paired with the matching corner `Anchor`). Selected by a
`Decoration`'s `Preset` name; callers add more via
`scene.WithOrnamentExtension` (mirrors the frame/icon seams).

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
masters, the theme, and the OPC package. Created via `pptx.New`, or read
from an existing deck via `pptx.NewFromBytes` / `NewFromFile` / `OpenStream`.

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

## Read model

The navigable builder model `pptx.NewFromBytes` / `OpenStream` reconstructs from
a pptx-go-authored deck — the **same** `Shape` / `Fill` / `Line` / `TextFrame` / `Table` / `Image`
types the builder writes, enumerated via `Slide.Shapes()` (RFC §16, D-047).
Reading maps the already-parsed `internal/ooxml` structs to public types; it is
not a parallel read hierarchy. Distinct from byte/codec round-trip (which the
G6 goldens already guarantee).

## Read warning

A non-fatal degradation (`pptx.ReadWarning`) noted while opening a deck pptx-go
did not author, surfaced via `Presentation.ReadWarnings()`. V1 reports
unrecognized shape-tree elements that were ignored (`WarnDroppedElement`) and
referenced parts that could not be read (`WarnUnreadablePart`), de-duplicated per
part + element. Empty for a self-authored deck. The mechanism behind the
[external deck](#external-deck) best-effort posture (RFC §16, D-048).

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

## Rotation

A shape's clockwise rotation about its centre, set via `pptx.WithRotation(deg)`
and stored as the OOXML `<a:xfrm rot="deg×60000">` attribute (D-041). Used by
the `chevron_arrow` ornament and the `Decoration.Rotation` field. Unit rotation
of a multi-shape ornament needs a group transform (not in V1); a multi-shape
ornament rotates per-shape.

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

## Section eyebrow

The top band of `Slide chrome`: a per-slide section label + a hairline rule,
drawn only when the slide sets `SceneSlide.Section`. Distinct from
`section_divider` (a full-bleed chapter-break *node*) — the eyebrow is chrome in
the top margin, above the body region. See `RFC-001-pptx-go.md §10.2`.

## section_divider

A scene IR leaf node representing a slide whose content is a single
full-bleed chapter break (label + optional ornament). **Distinct from
`PptxSection`** (the OOXML slide-grouping primitive). A
`section_divider` slide can be inside a `PptxSection` or not.

## Shapes (read enumerator)

`Slide.Shapes() []*Shape` — the read-side enumerator of a reopened slide; each
`*Shape` exposes the authored geometry / rotation / fill / line / shadow / text
/ table / image via read accessors (the read model, RFC §16, D-047).

## ShapeGeometry

A preset shape outline on the builder (`pptx.ShapeRect`, `ShapeEllipse`,
`ShapeRoundRect`, …), carrying the OOXML preset-geometry (`prst`) name.
Passed to `Slide.AddShape(geom, box, …)` with a `Box` (EMU) and optional
`Fill`/`Line`.

## Slide chrome

Opt-in recurring per-slide furniture drawn outside a shrunk body region: a
`Section eyebrow` at the top and a footer (a brand slot — text or image asset —
plus a `Footer page number`) at the bottom. Driven by `Scene.Chrome` (brand slot
+ page total + `Enabled`) and `SceneSlide.Section` / `.PageNumber`. Native shapes
reusing existing tokens; the zero value renders nothing (byte-identical). A
mechanism, not a judgment (D-053, D-026). See `RFC-001-pptx-go.md §10.2`.

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

## SlideTiming

An entry in `Stats.Timings`: the `SlideID` plus the wall-clock `Duration`
spent composing that slide, in scene order. Lets callers spot render
imbalance across a deck (D-015). Never serialized into the PPTX, so it
does not affect render idempotency.

## slog

`log/slog` — Go's structured logging stdlib package. pptx-go accepts an
optional `*slog.Logger` via `pptx.WithLogger` / `scene.WithLogger`. No
logger = no logs (zero cost).

## Skill smoke

The Phase 20 check that compiles and runs each agent skill's linked
`examples/` program (`go build` + `go run`), so a skill cannot silently
drift from the public API — an outdated identifier fails CI. Part of
`scripts/smoke/phase-20.sh`. See [Agent skill](#agent-skill).

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

## Status dot

A `Card`'s optional small filled dot (`Card.StatusDot *ColorRole`) in the
top-right corner — a colored status indicator. `nil` omits it. Part of the rich
card visuals (D-054). See `RFC-001-pptx-go.md §11.2`.

## Streaming

`pptx.OpenStream` and `pptx.SaveStream` — lazy-load and streaming-save
modes for large decks. The upstream's streaming primitives (in
`internal/opc/`) are first-class V1.

## Subsystem

An RFC-level grouping of related code (OPC, OOXML, Builder, Scene
renderer, Theme, Assets, Charts). Phase plans name their owning subsystem.

## SVG translator

`internal/render`'s converter from a single-path SVG to OOXML `custGeom`
(`render.Translate`). It enforces the documented subset — exactly one filled
`path`, commands `M L H V C S Q T Z` (absolute/relative; `S`/`T` reflect the
previous control point), no gradients, **no elliptical arcs** — and rejects
anything outside it at registration (D-005, D-040). Drives `Slide.AddIcon` /
`pptx.ValidateIcon`. The SVG's fill color is discarded; the icon fills with a
theme token.

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

## VAlignFill

The body-stack vertical alignment (`scene.VAlignFill`, on
`SceneSlide.Content.Vertical`) that pins fixed leaves at the top and grows the
`Flexible node`s to fill the remaining body height — the engine surface for
`Grow-to-fit`. Opt-in; the zero value `VAlignTop` is unchanged. See `D-052` and
`RFC-001-pptx-go.md §10.2`.

## Variant (theme variant)

A named theme variant (`light`, `dark`, `print`, brand-specific). The
scene's `Variant` field selects between variants the caller registered.
Per-slide theme overrides are V2; per-scene variant is V1.

## Watermark (card)

A `Card`'s optional large, low-opacity label (`Card.Watermark string`) drawn
behind the body content — e.g. a ghosted `01`. Rendered as a `TokenColorAlpha`
display run; `""` omits it. Part of the rich card visuals (D-054). See
`RFC-001-pptx-go.md §11.2`.

## Wave

A grouping of phases sharing a milestone. See `docs/plans/README.md` for
the V1 wave structure.

## WithWorkers

The `scene.RenderOption` that sets how many slides compose concurrently
(D-015). Default `runtime.GOMAXPROCS(0)`; `1` forces sequential. Output
stays byte-identical regardless: slides are created in scene order and any
slide that may register media composes sequentially (D-035, D-036).

## Wrapped-line estimate

The deterministic line count the scene layout engine uses to size a text slot:
`ceil(naturalWidth(text) / availableWidth)`, floored at 1, using the same pinned
char-width model as horizontal alignment. It drives `Content-aware height`; it
is an allotment estimate, not a prediction of PowerPoint's exact display reflow.
See `RFC-001-pptx-go.md §10.2`.

---

*Add new terms in alphabetical order. Keep entries terse — link to the
RFC section that owns the concept, don't re-explain.*
