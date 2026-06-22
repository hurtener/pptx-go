---
name: compose-a-scene
description: Describe a deck declaratively as a typed scene IR (the scene Layer 2) and render it to a PPTX. Use when you want to build slides from a tree of typed nodes — Hero, Heading, Prose, List, Callout, Table, cards, grids, columns — instead of imperatively placing shapes, and let the renderer handle layout, theming, and validation. For asset-backed nodes (Image/Chart/CodeBlock) see register-an-asset and the raster skills.
---

# Compose a scene

## Overview

`scene` is pptx-go's Layer 2: an IR-driven renderer that composes the Layer 1
`pptx` builder (P1 — `scene` imports `pptx`, never the reverse, and never
reaches under it). Instead of placing shapes by hand, you describe a deck as a
typed `Scene` — a theme plus ordered slides, each a tree of typed nodes — and
call `scene.Render`. The renderer validates the tree, lays each node out, and
drives the builder to emit native PPTX shapes (or, for asset nodes, a picture).

Everything visual flows through theme tokens (P2): colors, type, spacing,
elevation are semantic roles, so a theme swap re-renders the same IR in a new
visual language. The token enums are re-exported from `pptx` under the same
names (`scene.ColorAccent`, `scene.TypeH2`, …).

## The Scene type

```go
type Scene struct {
    Theme  *pptx.Theme   // optional; nil = the builder's default theme
    Slides []SceneSlide
    Meta   Metadata
}

type SceneSlide struct {
    ID      string        // your label; used in warnings/timings
    Layout  LayoutKind    // structural intent → a master layout
    Nodes   []SlideNode   // top-level node list
    Notes   RichText      // speaker notes (optional)
    Variant Variant       // theme variant (see below)
}

type Metadata struct {
    Title   string
    Author  string
    Subject string
}
```

`Metadata` (when non-empty) is written to `docProps/core.xml`.

**`LayoutKind`** — `LayoutCover`, `LayoutTitleContent`, `LayoutTwoColumn`,
`LayoutCardGrid`, `LayoutFullBleed`, `LayoutBlank`. A structural intent; it maps
to a named master layout via `WithLayoutMap` (a kind with no mapped layout, or a
name the template lacks, falls back to the blank layout and records a warning).

**`Variant`** — `VariantLight` (default), `VariantDark`, `VariantPrint`.
Variant selection is **not yet implemented**: a non-`VariantLight` slide renders
with the active theme and surfaces a `LayoutWarning` (it is not silently
dropped).

## The node catalog — exhaustive

`SlideNode` is a sealed union (the marker is unexported, so the set is closed to
the package). Construct the concrete struct values directly. Exactly the nodes
that render as a **picture** carry an `AssetID` field — `Image`, `Chart`,
`CodeBlock`, and asset-kind `Decoration` (D-011/D-018); **every other node
renders as native PPTX shapes**.

### Leaf nodes

| Node | Fields | Render |
|------|--------|--------|
| `Hero` | `Eyebrow, Title, Subtitle string; AutoFit bool` | native; `AutoFit` shrinks the title to fit one line when it would overflow |
| `Prose` | `Paragraphs []RichText` | native |
| `Heading` | `Text RichText; Level int (1..6); AutoFit bool` | native; `AutoFit` shrinks the heading to fit one line when it would overflow |
| `List` | `Kind ListKind; Items []ListItem` | native |
| `Divider` | `Spacing SpaceRole` | native |
| `Quote` | `Text RichText; Attribution string` | native |
| `Callout` | `Kind CalloutKind; Title string; Body RichText` | native |
| `Chip` | `Label string; Tone ChipTone; Color ColorRole` | native |
| `Arrow` | `Direction ArrowDirection; Label string` | native |
| `Stat` | `Value, Label, Delta string; DeltaTone; AutoFit bool` | native; hero big-number (value at display scale + label + optional delta). `DeltaTone` = `DeltaUp` (success), `DeltaDown` (error), `DeltaNeutral` (muted). A `Grid` of `Stat`s is a metric strip. `AutoFit` (opt-in) shrinks a long value/price to fit its column on one line, down to 60% of the display size; off / fitting values are byte-identical |
| `SectionDivider` | `Eyebrow, Label string` | native (full-bleed) |
| `Table` | `Headers []RichText; Rows [][]RichText; Caption string` | native |
| `Flow` | `Orientation FlowOrientation; Steps []FlowStep; Connector ConnectorKind` | native |
| `Image` | `AssetID AssetID; Alt string; Frame FrameKind; FrameName string; Crop Crop; Fit Fit` | **picture** |
| `CodeBlock` | `AssetID AssetID; Language, Caption string` | **picture** |
| `Chart` | `AssetID AssetID; Caption string` | **picture** |
| `Decoration` | `Kind DecorationKind; Preset string; AssetID AssetID; Layer Layer; Anchor Anchor; Offset Position; Size Size; Bleed bool; Opacity float64; Rotation float64` | native (preset) / **picture** (asset) |

Supporting enums and structs:

- `ListItem{Text RichText; Level int; Checked bool}`; `ListKind` =
  `ListBullet`, `ListNumber`, `ListChecklist`.
- `CalloutKind` = `CalloutNote`, `CalloutWarning`, `CalloutTip`,
  `CalloutImportant`.
- `ChipTone` = `ChipTint`, `ChipSolid`, `ChipOutline`.
- `ArrowDirection` = `ArrowRight`, `ArrowLeft`, `ArrowUp`, `ArrowDown`.
- `FlowOrientation` = `FlowHorizontal`, `FlowVertical`; `ConnectorKind` =
  `ConnectorArrow` (default), `ConnectorArrowDashed`, `ConnectorCycle`,
  `ConnectorPlus`; `FlowStep{Label RichText; Detail RichText; Icon string}`
  (`Icon` is a closed-name curated/extension icon, Stage-1 validated).
- `FrameKind` = `FrameNone`, `FrameBrowser`, `FramePhone`, `FrameDesktop`,
  `FrameLaptop`; `FrameName` (when set) overrides `Frame` and selects a
  `WithFrameExtension` frame. `Crop`/`Fit` are re-exported builder types
  (`FitFill` is the default).
- `DecorationKind` = `DecorationPreset` (curated ornament, native) /
  `DecorationAsset` (caller bytes, picture); `Layer` = `LayerBackground`,
  `LayerForeground`. `Opacity` is `0..1` (0 = opaque).

### Container nodes

| Node | Fields | Notes |
|------|--------|-------|
| `TwoColumn` | `Ratio ColumnRatio; Left, Right []SlideNode; Join ColumnJoin; JoinLabel string` | both sides must be non-empty. `Join` draws a centered seam element: `JoinBadge` (a "VS"-style `JoinLabel` badge) or `JoinArrow` (a connector arrow); `JoinNone` (default) draws nothing |
| `Grid` | `Columns int (2..4); Ratio []int; Gap SpaceRole; Cells []SlideNode` | cell count must be a multiple of `Columns`; `Ratio` empty or len == `Columns` |
| `Bento` | `Columns int (≥1); Rows []BentoRow` (`{Label string; Cells []BentoCell}`, `BentoCell{Span int; Node SlideNode}`); `WeightedRows bool` | row-labeled grid: rows with an optional left label and cells of variable column span on a shared grid (a span-S cell = S of `Columns` units). A row's spans sum to ≤ `Columns`; the gutter is reserved only when some row has a `Label`. `WeightedRows` (opt-in) sizes each row to its content's preferred height — dense rows grow, sparse rows shrink, clamped so the bento always fits its region; default equal rows are byte-identical |
| `Card` | `Header, Eyebrow, Icon, HeaderPill string; Body []SlideNode; BodyLayout BodyLayout; Fill ColorRole; Outline bool; BorderStyle BorderStyle; Size CardSize; Layout CardLayout; Elevation ElevationRole; HeaderFill, StatusDot *ColorRole; Watermark string; BodyVAlign VAlign` | accent card; all fields beyond `Header/Body/BodyLayout/Fill/Outline/Elevation` are additive (zero values reproduce the prior render). Rich visuals: `HeaderFill` (colored header band, body keeps `Fill`), `StatusDot` (top-right dot), `Watermark` (large faint label behind the body). `HeaderFill`/`StatusDot` are `*ColorRole` — take a role's address; `nil` omits. `BodyVAlign` (opt-in) distributes the vertical body within the card — `VAlignBottom` pins secondary content to the card bottom, `Justify`/`Fill`/`Fit` spread/grow/compress it; zero `VAlignTop` is byte-identical |
| `CardSection` | `Header string; Body []SlideNode` | top-level card accepting grids / two-columns / nested cards; `Body` must be non-empty |

Supporting enums: `ColumnRatio` = `Ratio11`, `Ratio12`, `Ratio21`;
`BodyLayout` = `BodyVertical`, `BodyHorizontal`; `BorderStyle` =
`BorderDefault` (defer to `Outline`), `BorderNone`, `BorderSolid`,
`BorderAccent`; `CardSize` = `CardSizeMD`, `CardSizeSM`, `CardSizeLG`;
`CardLayout` = `CardLayoutDefault`, `CardLayoutIconTop`.

## RichText

Text content is a `RichText` — an ordered list of styled runs:

```go
type RichText []TextRun

type TextRun struct {
    Text  string
    Style RunStyle
    Color TextColor
}

type RunStyle struct {
    TypeRole  TypeRole // typography scale (TypeBody, TypeH2, TypeCode, …)
    Bold      bool
    Italic    bool
    Underline bool
    Strike    bool
    Code      bool   // inline code (mono + tint)
    Link      bool   // marks a hyperlink; Href is the target
    Href      string
}
```

Run color is a `TextColor`:

- `scene.TokenTextColor(role)` — a theme-bound `TextColorRole` (the documented
  default path). The zero `TextColor` is the token `TextPrimary`.
- `scene.LiteralColor("RRGGBB")` — a literal RGB escape hatch, bypassing the
  theme.

`TextColorRole` values: `TextPrimary`, `TextSecondary`, `TextTertiary`,
`TextInverse`, `TextMuted`, `TextAccent`, `TextAccentAlt`, `TextSuccess`,
`TextWarning`, `TextError`.

## Rendering

```go
func Render(pres *pptx.Presentation, s Scene, opts ...RenderOption) (Stats, error)
```

`Render` applies the theme, validates (Stage 1), then lays out and composes each
slide. It is deterministic: the same scene + theme produces byte-identical
output regardless of worker count. Slides are created in scene order, then
media-free slides fan out across a worker pool while media-bearing slides render
sequentially for stable media numbering.

`RenderOption`s:

- `WithTheme(*pptx.Theme)` — active theme; takes precedence over `Scene.Theme`.
- `WithWorkers(int)` — slides composed concurrently (`<=0` = GOMAXPROCS;
  `1` = sequential).
- `WithLogger(*slog.Logger)` — structured render diagnostics (no logger = no
  logs); emits a Warn per `LayoutWarning`.
- `WithContext(context.Context)` — passed to the resolver; honors cancellation
  between slides.
- `WithLayoutMap(LayoutMap)` — map each `LayoutKind` to a named master layout
  (`scene.DefaultLayoutMap()` covers stock English template names).
- `WithAssetResolver(AssetResolver)` — resolves an `AssetID` to bytes for
  Image/Chart/CodeBlock/asset-Decoration (see register-an-asset).
- `WithIconExtension(name, svg)` — register a caller icon SVG for this render
  (validated at registration against the translator subset).
- `WithFrameExtension(name, recipe)` — register a caller device frame.
- `WithOrnamentExtension(name, recipe)` — register a caller ornament preset.

## Validation

`scene.ValidateScene(s) error` runs Stage 1 (structural) validation; `Render`
calls it automatically and returns its error before composing anything. It
returns a joined error so you see every problem at once. Per-node rules:

- `Heading`: `Level` in `1..6`.
- `List`: valid `Kind`; at least one item.
- `Callout`: valid `Kind`.
- `Image`/`Chart`/`CodeBlock`: non-empty `AssetID` (`Image` also validates
  `Crop`: each edge in `[0,1]`, `Left+Right < 1`, `Top+Bottom < 1`).
- `Decoration`: preset kind needs a `Preset`; asset kind needs an `AssetID`;
  `Opacity` in `[0,1]`.
- `Flow`: at least one step.
- `Stat`: non-empty `Value` (label and delta are optional).
- `Table`: at least one header column; every row width equals the header width.
- `TwoColumn`: non-empty `Left` and `Right` (children validated recursively).
- `Grid`: `Columns` in `2..4`; `Ratio` empty or length == `Columns`; non-empty
  `Cells`; cell count a multiple of `Columns`.
- `Card`/`CardSection`: children validated recursively (`CardSection.Body`
  non-empty).
- `Bento`: `Columns >= 1`; non-empty `Rows`; each row non-empty; each cell
  `Span >= 1` with a non-nil node; a row's spans sum to `<= Columns` (children
  validated recursively).

Registry-aware checks run inside `Render`: an Image's resolved frame name, a
card/flow `Icon`, and a preset `Decoration`'s name must all resolve to a curated
or registered entry, else the render fails with a Stage-1 error.

## Stats and warnings

```go
type Stats struct {
    Slides   int
    Shapes   int
    Assets   int
    Warnings []LayoutWarning
    Timings  []SlideTiming  // per-slide wall-clock, scene order
    Colors   []SlideColors  // per-slide resolved Canvas/Surface/PrimaryText (RGB), scene order
}

type LayoutWarning struct { SlideID, Node, Message string }
```

Non-fatal issues (content overflow, a reserved variant, an unmapped layout
falling back to blank, an unresolved asset) surface as `Warnings` — pptx-go
**warns, it does not fail** on layout problems, and has no strict mode. A caller
that wants warnings to be fatal inspects `Stats.Warnings` itself.

`Stats.Colors` reports, per slide (scene order), the resolved `Canvas`/`Surface`/
`PrimaryText` RGBs the engine rendered with — the **derived dark palette** for a
`VariantDark` slide — so a caller can compute its own text/surface contrast
against the real background. The engine does no contrast logic; it only reports
what resolved.

Slot heights are **content-aware**: a text node's height grows with the number
of lines its text wraps to, so a long paragraph doesn't overlap the node beneath
it, and a slide whose text genuinely exceeds the body region records a
`content overflows its region` warning (the signal to flag a slide as too full).
The estimate is deterministic; short single-line content is allotted the same
compact height and never falsely warns. A `Card`/`CardSection` header (and
eyebrow) is likewise wrapped-aware: a header that wraps to several lines in a
narrow card pushes the body down below the wrapped header instead of colliding
with it (single-line headers are unchanged).

`SceneSlide.Content.Vertical` aligns the body stack: `VAlignTop` (default),
`VAlignCenter`, `VAlignBottom`, `VAlignJustify`, `VAlignFill`, `VAlignFillCapped`,
and `VAlignFit`. `VAlignFill` pins fixed leaves at the top and **grows the flexible
nodes** — the containers (`Grid`, `TwoColumn`, `Card`, `CardSection`, `Bento`,
`Table`) plus `Image`/`Chart` — to consume the remaining height, so a sparse slide
fills its frame. The leftover height is shared proportionally and
deterministically; text leaves keep their size, and a slide with no flexible node
just top-aligns. `VAlignFillCapped` is the same but caps each flexible node's
growth at a pinned factor of its preferred height (at most double), turning the
leftover slack into even spacing — so a near-empty card can't balloon while a dense
one starves.

`VAlignFit` is the compression inverse, for the **over-full** slide: when the
body stack is taller than its region, it shrinks the stack to fit instead of
letting content spill off-slide. The engine first tightens inter-node gaps toward
a pinned floor, then — only if still overflowing — scales every node's slot
height toward a pinned ratio floor (60% of preferred), so the last node lands
inside the frame. A stack that already fits is byte-identical to `VAlignTop`; an
overflow too large for the pinned floors still raises the `content overflows its
region` warning. Both modes are opt-in and deterministic — the engine never
decides on its own that a slide is too thin or too full.

Set `Scene.Chrome` (`Enabled`, `Brand`/`BrandAsset`, `Total`) for opt-in
per-slide **chrome** drawn outside the body region (the body shrinks to fit): a
bottom footer with the brand slot + an `N / total` page number on every slide,
and a top section eyebrow + hairline on slides that set `SceneSlide.Section`.
Page total and per-slide number auto-derive (slide count; 1-based position) and
are overridable. Chrome uses theme tokens; the zero value (disabled) is
byte-identical.

## Complete example

See `examples/compose-a-scene/main.go` for a runnable program: a cover slide
(`Hero`) plus a content slide using `Heading` + `Prose` + `List` + `Callout` +
`TwoColumn` of `Card`s, with token colors and speaker notes, rendered with
`scene.Render` and saved via `pres.WriteToBytes()`.

```go
sc := scene.Scene{
    Meta: scene.Metadata{Title: "Quarterly Review", Author: "Platform Team"},
    Slides: []scene.SceneSlide{
        {ID: "cover", Layout: scene.LayoutCover, Nodes: []scene.SlideNode{
            scene.Hero{Eyebrow: "Q2 2026", Title: "Quarterly Review", Subtitle: "What shipped, what's next"},
        }},
        {ID: "highlights", Layout: scene.LayoutTitleContent, Nodes: []scene.SlideNode{
            scene.Heading{Level: 2, Text: scene.RichText{{Text: "Highlights"}}},
            scene.List{Kind: scene.ListChecklist, Items: []scene.ListItem{
                {Text: scene.RichText{{Text: "Cut p99 latency by 38%"}}, Checked: true},
            }},
        }},
    },
}
pres := pptx.New()
stats, err := scene.Render(pres, sc)
```

## What the caller owns (D-026)

pptx-go is an **engine, not a product**. The scene IR converts a typed
description into PPTX and nothing else. The caller owns everything opinionated:
what slides to emit and in what order, what text to write, which tokens to pick,
how to rasterize charts/code/images and resolve their `AssetID`s, and whether a
`LayoutWarning` should be treated as an error. There is no render-mode toggle,
no legibility heuristic, no markdown ingestion, no "raster everything" switch —
those are caller concerns.

## See also

- **register-an-asset** — supply bytes for `Image`/`Chart`/`CodeBlock` /
  asset-`Decoration` via an `AssetResolver`.
- **extend-the-icon-set** — register caller icons for `Card.Icon` /
  `FlowStep.Icon` with `WithIconExtension`.
- **embed-a-chart-raster** — rasterize a chart and place it as a `Chart` node.
- **embed-a-code-block-raster** — rasterize code and place it as a `CodeBlock`.
- **define-a-theme** — build the `*pptx.Theme` whose tokens this IR resolves
  against.
