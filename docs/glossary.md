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

## Auto-contrast

The engine mechanism (`onCardSurface`) that picks a card/container chrome run's
text color from the [[relative luminance]] of the surface behind it: a light text
token on a dark surface, the inherited dark default (no explicit color) on a light
one. Pinned sRGB luminance + a black/white-crossover threshold, integer per call →
deterministic. A *mechanism*, not a policy (`D-026`): a caller's explicit `Color`
always wins, and a light-surface card is byte-identical to the pre-mechanism output
(`D-082`). Reconciles the `D-058` "engine ships no contrast logic" stance — it is a
fixed token picker, the color analog of `deltaToneColor`, not opinionated taste.

## Average char width

`FontSpec.AvgCharWidth` — a role face's average glyph advance as a fraction of
font size, used only by the deterministic wrap/overflow estimator (`naturalWidth`);
it never renders. A soul sets a measured factor for a serif/display face; `0`
uses the built-in `~0.5` sans fallback (byte-identical). See `docs/design/THEME.md`,
`D-064`.

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

## AutoFit (shrink-to-fit)

The opt-in `AutoFit bool` on the display nodes (`Hero`, `Stat`, `Heading`, D-074):
when the display run's estimated `naturalWidth` exceeds its box, the engine
downscales its font (via `FontScale`) so it fits one line, quantized to a fixed
step and floored at a 0.60 ratio. Never upscales; the zero value (off) and
already-fitting text are byte-identical. The engine never shrinks on its own —
the caller opts a node in. See `Fit-to-region compression` (the vertical analog)
and `RFC-001-pptx-go.md §10.2`.

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

## Bento

A scene IR container node (`scene.Bento`): rows that each carry an optional left
label (`BentoRow.Label`) and a left-to-right sequence of cells of variable
`Column span` (`BentoCell.Span`) measured against `Bento.Columns` shared column
units, so columns align across rows. Distinct from `Grid` (uniform columns, one
child per cell). A native container — cells render per their own policy (D-056).
Rows are equal-height by default; the opt-in `Bento weighted rows` mode sizes
them to content. The left row-label **gutter sizes to its widest label**
(`naturalWidth(label) + padding`, clamped to a min/max), used by both the layout and
the slot estimate, so a label like "Control plane" no longer wraps in a fixed gutter
(R11.9, `D-089`). See `RFC-001-pptx-go.md §11.2`.

## Bento weighted rows

The opt-in `Bento.WeightedRows` mode (D-072): instead of equal-height rows
(`(box.H − gaps)/nRows`), each bento row sizes to its content's preferred height —
the tallest cell's `preferredHeight` at that cell's span width — so a dense row
no longer shares an equal band with a sparse one. When the preferred rows would
overflow, a single deterministic basis-point scale clamps them so `Σ rows + gaps
≤ box.H` (no off-slide row); when they fit, rows keep their preferred height
(top-aligned, slack as bottom whitespace). The zero value (`false`) is
byte-identical to the equal-row layout. See `RFC-001-pptx-go.md §11.2`.

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

## Button (node)

A scene IR leaf node: a presentational CTA / action affordance — a content-fit
`RadiusFull` pill with a bold label and optional leading/trailing icons, droppable
standalone (a closing slide), inside a card body (a pricing card), or inside a banner.
Its `ButtonTone` selects a token fill (Primary/AccentAlt solid, Neutral surface, Ghost
= an accent outline); its `ButtonSize` (MD/SM/LG) scales the pinned height/padding/icon
geometry. It is a shape only — no hyperlink/action wiring (the deck is static). Renders
as native shapes + custGeom icons (no media). Absent ⇒ byte-identical. See `D-094`.

## Checklist (node)

A scene IR leaf node: a dense feature / "what you get" list — rows of a **true filled**
status glyph (a curated `check` / `x` / `dot` custGeom, selected by `CheckState`, never
an empty font checkbox) before rich text, with the text hanging-indented from the glyph
width. `Columns` (1–3) reflows items row-major into balanced columns; `GlyphTone`
(`*ColorRole`, nil = per-state default) re-skins the glyphs; `Fill` distributes
inter-row spacing so a short list spans its box (and lets a `VAlignFill` card grow it).
Renders as native shapes + custGeom glyphs (no media). Absent ⇒ byte-identical. See
`D-095`.

## ChipRow (node)

A scene IR leaf node: a horizontal, wrap-to-next-line row of content-fit chip pills with
an optional leading `TypeCaption` label — a tag / category / capability strip. Each chip
sizes to its label (plus an optional leading icon); chips lay left-to-right and, when
`Wrap` is set, reflow onto new lines within the box width. `Wrap` is the engine
mechanism (zero = single line; the product sets it true). `Align` offsets each line.
Renders as native rounded-rect pills + text + optional custGeom icons (no media). Absent
⇒ byte-identical. See `D-096`.

## Banner (node)

A scene IR node: a full-width filled "big takeaway / promo / CTA" strip — a leading icon
+ a bold lead phrase + a supporting body on the left, with optional right-aligned
`Trailing` children (a `Stat` and/or `Button`). Distinct from the side-bar `Callout` (the
banner is a wide, full-fill `RadiusLG` band). `Fill` defaults to accent (its zero value);
the lead/body auto-contrast against the fill unless an explicit `TextColor` is set.
Renders native (filled rect + text + custGeom icon; children per their own policy). A
node with children — recurses like a container in every walk. Absent ⇒ byte-identical.
See `D-097`.

## Lockup (node)

A scene IR leaf node: a compact "powered by / in partnership with" attribution mark — a
caption paired with a small partner logo composed as one centered inline unit. The mark is
either an `AssetID` (a logo resolved via the `AssetResolver`, rendered as a pic) or an
`Icon` (a curated glyph, media-free); exactly one is set. `AssetSide` orders the
caption/logo pair; `MaxHeight` bounds the (square — no pixel aspect, §7) logo; `Align`
positions the whole group. An asset lockup composes serially for deterministic media part
numbering. Absent ⇒ byte-identical. See `D-102`.

## IconRows (node)

A scene IR leaf node: a vertical stack of `[icon | label | optional right-aligned meta]`
rows — the "integrations / capabilities / sources" list that reads as designed rows rather
than bullets. Each `IconRow` pairs a leading icon with a rich label and an optional meta;
`RowPill` frames a row in a `SurfaceAlt` rounded-rect. `Fill` distributes inter-row spacing
so the rows span the box (a `VAlignFill` card grows it); `GlyphColor` tints the icons (its
zero value defaults to accent). Renders native (icon custGeom + text + optional pill).
Absent ⇒ byte-identical. See `D-100`.

## Grid connector

An inter-column connector glyph drawn in the gutter between two adjacent columns of a
`Grid` (`Grid.Connectors []GridConnector{Between [2]int; Kind ConnectorKind; Label}`), so
an architecture / pipeline grid reads as data flow rather than mere adjacency. The gutter
box is derived from the cell layout; the glyph reuses the Flow connector set plus
`ConnectorBiArrow` (a bidirectional arrow). An empty `Connectors` slice ⇒ byte-identical.
See `D-099`.

## Ribbon (Card field)

A pinned emphasis badge on a `Card` (`Card.Ribbon`) — a "MOST POPULAR" / "RECOMMENDED" /
"NEW" marker that singles one card out of a row, sitting OUTSIDE the header text flow
(distinct from `HeaderPill`, an in-row pill). `RibbonTopBar` is a full-width tab that
reserves a band so the card body shifts down below it; `RibbonCornerStar` is a star glyph;
`RibbonCornerTL`/`RibbonCornerTR` are content-fit corner text tabs. `Color` (a
`*ColorRole`, nil = accent) and `TextColor` (auto-contrast by default) are tokens. `nil`
⇒ no ribbon, byte-identical. See `D-098`.

## Card

A scene IR container node — accent strip + optional icon / eyebrow /
header-pill, with leaf children. Renders as a native PPTX shape group.
The body is top-anchored by default; the opt-in `Card BodyVAlign` distributes it
vertically. The opt-in `FillGradient` (a `*GradientFill`, D-108) replaces the
solid `Fill` with a 2-stop linear surface gradient (depth shift); nil = solid.
See `RFC-001-pptx-go.md §11.2`.

## Card BodyVAlign

The opt-in `Card.BodyVAlign VAlign` field (D-073): the vertical distribution of a
card's body within the card body region — any of the eight `VAlign` modes:
`VAlignTop` (default, top-anchored), `VAlignCenter`, `VAlignBottom` (pin to the
body bottom), `VAlignJustify` (spread the inter-item gaps), `VAlignFill` (grow
flexible body nodes), `VAlignFillCapped` (capped grow + even spacing),
`VAlignBalanced` (even rhythm, optical-center), or `VAlignFit` (compress an
over-full body). The card body routes through the same
`alignedStackIn` engine as the slide body stack; the zero value reproduces the
top-anchored layout byte-for-byte. Applies to the vertical body only
(`BodyLayout != BodyHorizontal`). See `RFC-001-pptx-go.md §11.2`.

## Card PaddingScale

The additive `Card.PaddingScale int` (D-076): a basis-point multiplier on the
card's size-resolved interior padding (`CardSize` → `SpaceSM/MD/XL`). The zero
value and 10000 leave it unchanged (byte-identical); below 10000 tightens a dense
card (floored at a pinned `SpaceXS` minimum so the inset never collapses), above
10000 loosens it. The base and the floor both resolve through theme spacing tokens
— no literals (P2). A tighter scale shrinks the inset and grows the card body. See
`RFC-001-pptx-go.md §11.2`.

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
`success`, `paper`). Page-level surfaces resolve via the active `Theme`'s
`ColorPalette` to an OOXML color value. See `RFC-001-pptx-go.md §7.1`.

## ColorPaper (paper canvas)

The `ColorPaper` surface role (D-104): a faintly tinted off-white "paper"
canvas distinct from pure white, for a designed background tone on content
slides. Defaults to `ColorCanvas`'s value (white) — byte-identical until a theme
sets a tint via `pptx.WithPaper(RGB("FAFAF8"))` — and is pointed at by a
`Background{Kind: BackgroundColor, Color: ColorPaper}`. Like `TextMuted` it has
no theme1.xml slot, so a re-opened deck's theme reads it back at its default; its
resolved background RGB still round-trips on the slide. See `RFC-001-pptx-go.md §7.1`.

## Column join

The optional element a `TwoColumn` draws centered on its seam
(`TwoColumn.Join ColumnJoin`): `JoinNone` (nothing — the byte-identical default),
`JoinBadge` (a `VS badge` — a circular `JoinLabel` straddling the seam), or
`JoinArrow` (a right-arrow connector between the two columns). Native shapes
reusing accent/inverse tokens (D-055). The `JoinBadge` **sizes to its label** —
the diameter grows to contain the `JoinLabel` (up to a cap, then the label is shrunk
to one line) so a multi-word label renders intact instead of breaking mid-word; a
short label like "vs" keeps the base diameter (byte-identical) (R11.7, `D-087`). The
`TwoColumn.JoinPosition` selects where it sits: `JoinSeam` (centered, the default) or
`JoinTopBridge`/`JoinBottomBridge` — a horizontal accent bracket (a spanning line + two end
stubs + a content-fit centered label pill, no mid-word wrap) across the top/bottom of both
columns, the "one X, two ways" header (`D-101`). See `RFC-001-pptx-go.md §11.2`.

## Column span

A `BentoCell`'s width in shared column units (`Span`): a span-S cell occupies S
of the bento's `Columns` units, so a span-2 cell is twice a span-1 cell (plus the
inter-unit gap) and columns align across rows. See `Bento`,
`RFC-001-pptx-go.md §11.2`.

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
content keeps its prior fixed height (byte-identical). The Card/CardSection slot
estimate is wrapped-header-aware (the `cardChromeEst` baseline plus the extra
eyebrow/title lines) and the Bento estimate measures each cell at its actual span
width, so the estimators match the composed geometry — overflow detection and the
`Fit-to-region compression` pass are trustworthy (D-079). See
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
(0..1) and `Rotation` (degrees) style it. `Color *pptx.ColorRole` (D-107)
overrides the ornament's color role (nil = `ColorAccent`, byte-identical), so a
texture/glow can be neutral grey, inverse-white, or any brand role. The
`DecorationText` kind (D-109) instead draws an oversized, low-opacity ghost
number/word (`Text` + optional `FontSize`) behind the body — colored by `Color`
at the `Opacity` alpha. Otherwise pure visual; carries no text. `Layer` is
`background` (renders behind body) or `foreground` (renders on top) — the
renderer imposes that z-order (RFC §10.2).

## Delta tone

A `Stat.Delta`'s color direction (`DeltaTone`): `DeltaUp` (success/green),
`DeltaDown` (error/red), `DeltaNeutral` (muted, the zero value). Maps to existing
theme tokens — no new token (D-057). See `Stat`, `RFC-001-pptx-go.md §11.1`.

## Display font

The optional third font-scheme face on a `Theme` (`Theme.DisplayFont`, set via
`WithDisplayFont`) used by the `TypeDisplay` role — the big editorial face,
independent of `HeadingFont`. Empty = `TypeDisplay` inherits `HeadingFont`
(byte-identical). Lets a brand pair a serif display with a separate sans for
headings. See `docs/design/THEME.md`, `D-063`.

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

## Fit-to-region compression

The deterministic engine pass behind `VAlignFit`: when a body stack's preferred
height exceeds its region, it shrinks the inter-node gaps toward a pinned floor
(`SpaceXS`), then — if still overflowing — proportionally scales every node's
slot height toward a pinned ratio floor (0.60), so the last node lands inside the
region instead of clipping off-slide. The compression inverse of `Grow-to-fit`;
integer-EMU / basis-point, worker-count independent; byte-identical when the
content already fits. The card-padding and display-type-scale sub-steps are
layered in by later engine units. See `D-071` and `RFC-001-pptx-go.md §10`.

## Fill

A shape's interior fill (builder). V1 ships `pptx.SolidFill(Color)` and
`pptx.NoFill()`; a fill resolves its `Color` against the active `Theme` when
applied. Gradient, pattern and picture (blip) fills are tracked for later.

## Flexible node

A scene node whose slot grows under `VAlignFill`: the containers (`Grid`,
`TwoColumn`, `Card`, `CardSection`, `Bento`, `Table`) plus the stretchable visuals
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

## Font-embedding pass

The opt-in save-time pass enabled by `pptx.WithFontEmbedding()` that walks
every slide's runs, collects the distinct used faces — `(family, weight,
italic)` — in a stable sorted order, and `EmbedFont`s each via the
registered `FontSource`, so a deck themed with a brand display/heading face
ships those faces. It is **weight-aware** (D-068): it embeds the actual
resolved weight file per OOXML bucket (the four regular/bold/italic/boldItalic
cuts), choosing the weight nearest the bucket nominal when several collide, so
a soul's medium (500) regular role ships the medium file rather than a
synthetic 400. It is a no-op without a `FontSource`, warns (does not fail) on
a face the source cannot resolve, is idempotent against manual `EmbedFont`
calls, and is byte-identical to the prior output when off. (D-065, D-068,
R9.1/R9.8, `RFC §7.6`.)

## Font fallback chain

`FontSpec.Fallback []string` — an ordered list of substitute families for a
type role. When a `FontSource` is registered and it cannot resolve the
role's primary `Family`, the write-time fallback pass rewrites the run's
single-valued `a:latin` typeface to the first family in `[Family]` +
`Fallback` the source can resolve, so output degrades to a controlled
near-match instead of an arbitrary host default. Resolution is
**italic-aware** (D-067): it is keyed per `(family, italic)` — the italic cut
is probed for italic runs, the regular cut for upright ones — so an italic
emphasis run whose family lacks an italic cut falls back to an italic-capable
face (not a faux-italic), while upright runs keep the primary. Empty (the zero
value) and "no `FontSource`" are byte-identical; resolution is deterministic and
idempotent across saves. The chain *contents* are the soul's choice; the
engine provides the carry-and-resolve mechanism (D-066, D-067, R9.6/R9.7,
`RFC §7.6`).

## FontSource

`pptx.FontSource` — the caller-injected interface that resolves a
font name + style + weight to bytes. Registered via
`pptx.WithFontSource(...)`. The presentation's `EmbedFont(name, style,
weight)` method uses it to write font-embedding parts. By default the
caller invokes `EmbedFont` explicitly for each face; the opt-in
`Font-embedding pass` (`pptx.WithFontEmbedding()`) automates this by
embedding every face a deck actually uses. (D-019, D-065, `RFC §7.6`.)
Registered in V1 via `pres.SetFontSource(...)`; the functional
`pptx.WithFontSource(...)` option arrives with the builder spine (D-030).

## FontScale

`pptx.RunStyle.FontScale` — a per-run multiplier on the resolved type-role size
(D-074). The role's `Size` token stays the source of truth; `FontScale` only
scales it, so a theme swap still re-skins the base. The zero value (and 1) leaves
the size unchanged (byte-identical); a value in (0,1) emits the reduced
`a:rPr/@sz`, round-tripping via `Run.FontSize`. The scene `AutoFit` (shrink-to-fit)
path computes it deterministically; there is no per-role `FontScale` token.

## FontSpec

A resolved typography value: font `Family`, `Size` (points), `Weight`
(100–900; ≥600 is bold), and `Italic`, plus the Wave-9 type-detail fields —
`Tracking` (letter-spacing, D-060), `LineHeight` (leading, applied by the
scene layer, D-061), `Case` (transform, D-062), `AvgCharWidth` (estimator
metric, D-064), and `Fallback` (substitute chain, D-066). `Theme.ResolveType(
TypeRole)` returns one. (Slices make it non-comparable — compare with
`reflect.DeepEqual`.)

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

## Gradient stops (background)

The scene `Background.Stops []GradientStop` field (D-105): a 2–8-stop,
ascending-in-`[0,1]` multi-hue background wash, each stop a `{Pos float64;
Color pptx.ColorRole}`. When non-empty it supersedes the legacy two-role
`Background.Gradient` pair (which stays byte-identical when `Stops` is empty).
Invalid stops (`<2`, `>8`, out of `[0,1]`, or not strictly ascending) record a
`LayoutWarning` and skip the fill (D-026 — no panic). The slice makes
`Background` non-comparable. See `DECKARD-PRODUCT-REQUIREMENTS.md` R13.3.

## Radial background

The scene `BackgroundRadial` kind (D-106): a center-out radial slide fill (a
spotlight/vignette for dark hero/section/closing slides) drawn via
`pptx.RadialGradient` with a centered 50%-inset circular focal. It consumes the
same `Background.Stops` (or the legacy 2-role `Background.Gradient`) as the
linear `BackgroundGradient`. The focal point is centered; a focal-offset knob is
deferred (center-only in V1). See `DECKARD-PRODUCT-REQUIREMENTS.md` R13.2.

## Mesh background

The scene `BackgroundMesh` kind (D-112): a soft cover "mesh" wash — a base canvas
fill plus the N low-alpha radial glows in `Background.Mesh` (each a `MeshGlow` =
`{Anchor; Color pptx.ColorRole; Radius EMU; Alpha int}`), pooled at caller
anchors over the canvas, drawn via `pptx.RadialGradient` in slice order
(deterministic). An empty `Mesh` draws nothing. See
`DECKARD-PRODUCT-REQUIREMENTS.md` R13.4.

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
`Fill`. `nil` omits it. Part of the rich card visuals (D-054). Its height is the
wrapped header height (`cardHeaderBottom − box.Y`), so for an eyebrow/title that
wraps to N lines the band grows to contain every header glyph and the body still
starts below it (R10.1/D-070; verified across all `CardSize × CardLayout` combos
by D-081). See `RFC-001-pptx-go.md §11.2`.

## Header column width

The inner text column at which a card's eyebrow and title wrap
(`cardHeaderColumnWOf`): the card inner width minus the left-icon shift (icon-left
layouts) and the reserved header-pill width. The wrapped-line count measured at
this width drives both the header band height and the body-region top, so the two
never drift (R10.1/D-070).

## Header pill

A small pill-shaped badge (rounded-rect + short label) rendered in a card's
header row, typically right-aligned (e.g. a status tag like "NEW" or "BETA").
A field on `Card` / `CardSection`; part of the card chrome. The pill **sizes to its
label** on a single line (`cardPillWidthOf` = `naturalWidth(label) + padding`,
floored at a circular minimum, clamped to the card inner width; a label too long for
the inner width is shrunk to one line via `FontScale`), and the header text column
reserves exactly that width — so any label renders intact instead of wrapping inside
a fixed chip (R11.5, `D-085`). See `RFC-001-pptx-go.md §11.2`.

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

## Case (type)

A type role's case transform on a `FontSpec` (`FontSpec.Case`: `CaseNone` /
`CaseUpper` / `CaseSmallCaps`) — rendered via OOXML `a:rPr/@cap` (`all`/`small`),
so the run text stays original-case (round-trips) while the display is cased. A
`RunStyle.Case` overrides per run; `CaseNone` = none (byte-identical). Pairs with
`Tracking` for the tracked-caps eyebrow. See `docs/design/THEME.md`, `D-062`.

## Line height

Paragraph leading on a `FontSpec` (`FontSpec.LineHeight`, percent of single;
100 = single) — part of the resolved type scale. The scene renderer applies a
node's role line-height to its paragraphs, emitted as OOXML
`a:pPr/a:lnSpc/a:spcPct` (1/1000 percent); `ParagraphOpts.LineHeight` is the
builder-level override. 0/100 = none (byte-identical). See `Tracking`,
`docs/design/THEME.md`, and `D-061`.

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
`chevron_arrow`, `noise_overlay`, `starfield`. Rendered as native PPTX shapes in
the decoration's color role (default accent — D-107), at a caller opacity and
rotation (the glows use radial `Gradient fill`s; `noise_overlay` is a
deterministic sparse-dot grain approximation — D-041; `starfield` (D-110) is an
organic, box-derived scatter of dots with per-dot size and alpha variance,
hash-perturbed and deterministic). The three pattern recipes (`grid_dots`,
`noise_overlay`, `starfield`) derive their dot count from the box at a
`Decoration.Pitch` (D-111; 0 = a legacy fixed count), capped for file size.
`chevron_arrow` and `corner_bracket` honor `Rotation`
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

## BulletIndent

`pptx.ParagraphOpts.BulletIndent` — a per-paragraph override of a bulleted
paragraph's hanging indent (the marker-to-text offset), in EMU (D-078). The zero
value keeps the default 0.5" hanging indent (byte-identical); a positive value
sets a tighter (or wider) marker gap, emitted as `a:pPr/@marL` + `@indent`. The
scene `List indent` presets drive it. Applies only when a bullet is set.

## BulletKind

A paragraph bullet style on the builder (`pptx.BulletNone`, `BulletDisc`,
`BulletNumber`, `BulletCheckbox`). Set via `Paragraph.Bullet`; emits the
OOXML `buChar`/`buAutoNum`/`buNone` with a hanging indent. (RFC §8.4.)

## List indent (density)

The scene `List.Indent` preset (`scene.IndentNormal` / `scene.IndentTight`, D-078)
controlling a list's bullet hanging-indent density. `IndentNormal` (zero) is
byte-identical to the default; `IndentTight` tightens the marker-to-text offset so
dense lists sit tight to their markers, consistently across items and levels. The
tight indent is **proportional to the body type size** — anchored to `In(0.25)` at
the default 14 pt body and scaling with it — so the bullet-to-text gap stays tight at
any size (R11.10, `D-090`). Plumbed through `renderList` to the builder's
`BulletIndent`. Pinned calibration, not a theme token. See `RFC-001-pptx-go.md §11.1`.

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

## Relative luminance

The WCAG sRGB perceptual brightness of a color, in `[0, 1]` (here scaled to
`[0, 100000]`): channels are gamma-expanded then weighted `0.2126·R + 0.7152·G +
0.0722·B`. The basis for [[auto-contrast]] — a surface below the black/white
crossover (`≈ 0.179`) gets light text, above it the dark default. Computed via a
256-entry integer table built once at init, so the decision is pure integer and
worker-count independent (`D-082`).

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
`Paragraph.AddRun(text, RunStyle)`. The builder analog of a scene
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

## Safe area

The slide's printable region: the slide box minus the content margins minus the
reserved chrome bands (the section eyebrow at top, the footer/page-number at
bottom). Equal to the body region (`bodyRegion`). **Every content node** — every
container *and* every leaf — is clamped to it so an over-full stack or card body can
never push content off the slide onto the footer (R11.3, generalized in R11.12 from
containers to all nodes, `D-083`/`D-092`). The full-slide overlays (`Decoration`,
which may bleed off-canvas by design, and `SectionDivider`) are exempt. Complementary
to the opt-in `Fit-to-region compression` (`VAlignFit`), which reflows; the clamp
caps.

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

## SlideColors

An entry in `Stats.Colors`: the `SlideID` plus the resolved `Canvas`, `Surface`,
and `PrimaryText` RGBs the engine rendered that slide with — the derived dark
palette for a `VariantDark` slide. Lets a caller compute its own text/surface
contrast against the real background; the engine performs no contrast logic
(D-058, D-026). Never serialized into the PPTX. See `Stats`,
`RFC-001-pptx-go.md §10.1`.

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

## Stat

A scene IR leaf node (`scene.Stat`): a hero big-number metric — a display-scale
`Value` with a `Label` and an optional directional `Delta` (toned by `Delta
tone`). A `Grid` of `Stat`s forms a metric/pricing strip. The engine renders the
value/delta verbatim — it formats no numbers (D-057). With `AutoFit`, the `Value`
stays on **one line** via a pinned role ladder (`TypeDisplay → H1 → H2`, then a font
shrink at the floor), so a wide value like "$4,000+" never wraps and crowds the
caption (R11.8, `D-088`); `AutoFit`-off keeps the full display size (byte-identical).
Distinct from `Stats` (the render-result struct). See `RFC-001-pptx-go.md §11.1`.

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

## Tracking

Letter-spacing on a `FontSpec` (`FontSpec.Tracking`, points, signed) — part of
the resolved type scale, emitted as OOXML `a:rPr/@spc` (1/100 pt). Positive opens
glyphs apart (wide-tracked eyebrows), negative tightens (display headlines); a
`RunStyle.Tracking` overrides it per run. Zero = none (byte-identical). See
`docs/design/THEME.md` and `D-060`.

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

## VAlignFit

The body-stack vertical alignment (`scene.VAlignFit`, on
`SceneSlide.Content.Vertical`) that, when the stack overflows its region, applies
the deterministic `Fit-to-region compression` pass — the compression inverse of
`VAlignFill`. Opt-in; byte-identical to `VAlignTop` when the content already
fits. See `D-071` and `RFC-001-pptx-go.md §10`.

## VAlignBalanced

The body-stack vertical alignment (`scene.VAlignBalanced`, on
`SceneSlide.Content.Vertical`) that distributes a sparse stack's slack as an even
rhythm (D-077): the slack splits across the `n+1` spaces of the stack into a top
margin and widened inter-node gaps, with an optical-center bias (top margin = 85%
of an even unit) that seats the stack slightly above geometric center. Unlike
`VAlignJustify` (all slack into gaps, no margins) and `VAlignCenter` (equal
margins, fixed gaps), it spreads whitespace across both, so a sparse cover or
closing reads balanced rather than clustered with a large void. With no slack it
is `VAlignTop`. Per-node gap weighting is the caller's (D-026). See `D-077` and
`RFC-001-pptx-go.md §10`.

## VAlignFillCapped

The body-stack vertical alignment (`scene.VAlignFillCapped`, on
`SceneSlide.Content.Vertical`) that is `VAlignFill` with a ceiling (D-075): each
`Flexible node` grows by at most a pinned factor of its preferred height
(`fillGrowthMaxBP`, +1.0×), so a near-empty node cannot balloon. The leftover
slack beyond the caps becomes balanced spacing — an even top margin and widened
inter-node gaps (`residual/(n+1)`) — instead of inflating one node. With no
flexible node, or no slack, it is equivalent to `VAlignTop`; uncapped
`VAlignFill` is unchanged. See `D-075` and `RFC-001-pptx-go.md §10`.

## Variant (theme variant)

A named theme variant (`light`, `dark`, `print`, brand-specific). The
scene's `Variant` field selects between variants the caller registered.
Per-slide theme overrides are V2; per-scene variant is V1.

## VS badge

A `Column join` text badge (`TwoColumn.Join = JoinBadge` + `JoinLabel`): a
circular accent label straddling the seam between two compared columns (e.g.
"VS"). See `RFC-001-pptx-go.md §11.2`.

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
