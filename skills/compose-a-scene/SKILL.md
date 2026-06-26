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
    ID         string       // your label; used in warnings/timings
    Layout     LayoutKind   // structural intent → a master layout
    Nodes      []SlideNode  // top-level node list
    Notes      RichText     // speaker notes (optional)
    Variant    Variant      // theme variant (see below)
    Background  Background   // full-bleed slide background (optional; zero = none)
}

type Metadata struct {
    Title   string
    Author  string
    Subject string
}
```

`Metadata` (when non-empty) is written to `docProps/core.xml`.

### Slide background

Each `SceneSlide` may also carry `Footnotes []RichText` — source/citation/disclaimer lines pinned to a reserved muted band at the bottom (above the chrome footer; the body shrinks to reserve it, so no overlap); mark a reference on a figure with `RunStyle{Superscript: true}`. Lines past a 3-line cap are dropped + warned.

`SceneSlide.Background` paints a full-bleed fill behind all content (the lowest
z-layer). A zero `Background` (`BackgroundNone`) draws nothing. The kinds:

```go
type Background struct {
    Kind     BackgroundKind    // BackgroundNone | BackgroundColor | BackgroundGradient | BackgroundAsset | BackgroundRadial | BackgroundMesh
    Color    pptx.ColorRole    // BackgroundColor — solid fill (e.g. pptx.ColorPaper for a tinted paper canvas); BackgroundMesh — the base canvas under the glows
    Gradient [2]pptx.ColorRole // BackgroundGradient — legacy 2-role linear gradient (used when Stops is empty)
    Stops    []scene.GradientStop // BackgroundGradient — multi-stop wash: 2..8 ascending stops in [0,1]
    Angle    int               // linear gradient angle (degrees clockwise from +x; 0 = left→right, 90 = top→bottom)
    AssetID  scene.AssetID     // BackgroundAsset — full-bleed picture (needs an AssetResolver)
    Mesh     []scene.MeshGlow  // BackgroundMesh — N pooled radial glows over the base canvas
    Scrim    *scene.Scrim      // legibility overlay over any drawn fill; nil = none
    Duotone  *scene.Duotone    // two-tone recolor of a photo background (BackgroundAsset); nil = none
}

type MeshGlow struct { Anchor Anchor; Color pptx.ColorRole; Radius pptx.EMU; Alpha int } // a soft pooled glow

type GradientStop struct { Pos float64; Color pptx.ColorRole } // Pos in [0,1]

type Scrim struct { Color pptx.ColorRole; Opacity int; Gradient bool; GradientAngle int } // darkening/tinting overlay
type Duotone struct { Shadow, Highlight pptx.ColorRole } // photo shadows → Shadow, highlights → Highlight
```

For a **photographic** slide — a full-bleed photo with legible overlay text —
set `Kind: BackgroundAsset` with an `AssetID`, tint it on-brand with `Duotone`,
and guarantee text legibility with a gradient `Scrim`:

```go
Background: scene.Background{
    Kind:    scene.BackgroundAsset,
    AssetID: "asset://cover-photo",
    Duotone: &scene.Duotone{Shadow: pptx.ColorAccent, Highlight: pptx.ColorCanvas},
    Scrim:   &scene.Scrim{Color: pptx.ColorSurface, Opacity: 55000, Gradient: true},
},
```

The `Scrim` works over any background kind (it darkens whatever fill was drawn);
`Duotone` applies only to a `BackgroundAsset` photo. Both use theme tokens, so a
theme swap re-tints. A nil `Scrim`/`Duotone` renders byte-identically.

For a multi-hue hero wash, set `Stops` (it supersedes `Gradient`):

```go
Background: scene.Background{
    Kind: scene.BackgroundGradient,
    Stops: []scene.GradientStop{
        {Pos: 0, Color: pptx.ColorAccent},
        {Pos: 0.5, Color: pptx.ColorAccentAlt},
        {Pos: 1, Color: pptx.ColorCanvas},
    },
    Angle: 45,
},
```

`Stops` must be 2–8 stops, each `Pos` in `[0,1]`, strictly ascending — otherwise
the renderer records one warning and skips the fill (it never panics). An empty
`Stops` falls back to the two-role `Gradient` pair (byte-identical to the
2-stop form).

For a center-out spotlight/vignette (good behind a dark hero or closing slide),
use `BackgroundRadial` — it draws the same `Stops` (or the 2-role `Gradient`) as
a radial fill with a centered focal:

```go
Background: scene.Background{
    Kind: scene.BackgroundRadial,
    Stops: []scene.GradientStop{
        {Pos: 0, Color: pptx.ColorSurface}, // brighter center
        {Pos: 1, Color: pptx.ColorCanvas},  // darker edges
    },
},
```

For a soft cover "mesh" wash — diffuse colored light pooling at one or two
corners over the paper — use `BackgroundMesh`: a base canvas (`Color`) plus N
low-alpha radial glows pooled at caller anchors. Keep each `Alpha` low so the
pools whisper:

```go
Background: scene.Background{
    Kind:  scene.BackgroundMesh,
    Color: pptx.ColorPaper, // the base canvas under the glows
    Mesh: []scene.MeshGlow{
        {Anchor: scene.AnchorTopLeft, Color: pptx.ColorAccent, Radius: pptx.In(4), Alpha: 12000},
        {Anchor: scene.AnchorBottomRight, Color: pptx.ColorAccentAlt, Radius: pptx.In(5), Alpha: 10000},
    },
},
```

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
| `List` | `Kind ListKind; Items []ListItem; Indent ListIndent` | native; `Indent` = `IndentNormal` (default) / `IndentTight` — tight halves the bullet marker-to-text gap (to `In(0.25)`) for dense lists; default byte-identical |
| `Divider` | `Spacing SpaceRole` | native |
| `Quote` | `Text RichText; Attribution string; Mark bool; AvatarAssetID AssetID; AttributionName, AttributionRole, AttributionCompany string; LogoAssetID AssetID` | native — a plain pull-quote, or (when any enrichment field is set) a **testimonial**: `Mark` draws an oversized low-emphasis quotation glyph behind the text; `AvatarAssetID` is a rounded author avatar (via the AssetResolver); `AttributionName`/`Role`/`Company` is structured attribution (name bold, role · company muted); `LogoAssetID` is a customer logo. Laid out as one unit (mark + quote + a bottom `[avatar | name/role·company | logo]` strip). A Quote with only `Text`+`Attribution` is byte-identical to the plain quote; a missing avatar/logo warns and is omitted |
| `Callout` | `Kind CalloutKind; Title string; Body RichText` | native |
| `Chip` | `Label string; Tone ChipTone; Color ColorRole` | native |
| `Arrow` | `Direction ArrowDirection; Label string` | native |
| `Stat` | `Value, Label, Delta string; DeltaTone; AutoFit bool; Number *float64; Format *NumberFormat` | native; hero big-number (value at display scale + label + optional delta). `DeltaTone` = `DeltaUp` (success), `DeltaDown` (error), `DeltaNeutral` (muted). A `Grid` of `Stat`s is a metric strip. `AutoFit` (opt-in) shrinks a long value/price to fit its column on one line, down to 60% of the display size; off / fitting values are byte-identical. For locale-correct prices/metrics set `Number` + `Format` (a `*NumberFormat` = `{Decimals; GroupSep, DecimalSep, CurrencySymbol; SymbolAfter, Percent, Compact; CompactThreshold; Prefix, Suffix}`) instead of `Value`: `Number:4000` + `{GroupSep:",", CurrencySymbol:"$", Suffix:"+"}` → "$4,000+" on one line; `{Percent:true}` on 0.92 → "92%"; de-DE `{GroupSep:"."}` → "4.000". `scene.FormatNumber(v,f)` formats directly. A raw `Value` (nil `Number`) is byte-identical |
| `Button` | `Label string; Tone ButtonTone; Size ButtonSize; LeadingIcon, TrailingIcon string; Align HAlign` | native; a CTA / action pill (a closing slide, a pricing-card footer). `Tone` = `ButtonPrimary` (accent solid) / `ButtonAccentAlt` / `ButtonGhost` (outline) / `ButtonNeutral` (surface). `Size` = `ButtonMD` (default) / `ButtonSM` / `ButtonLG`. Width is content-fit to the label + icons, clamped to its box; `Align` centers/right-aligns it. `LeadingIcon`/`TrailingIcon` are closed-name registry icons (e.g. `arrow-right`); `""` = none. Presentational only — no hyperlink (the deck is static) |
| `Checklist` | `Items []ChecklistItem` (`{Text RichText; State CheckState; Icon string}`); `Columns int (1..3)`; `GlyphTone *ColorRole`; `Fill bool` | native; a dense feature / "what you get" list with **true filled** glyphs (a real check/cross/dot custGeom, never an empty font checkbox). `State` = `CheckDone` (check, accent) / `CheckNo` (cross, muted) / `CheckNeutral` (dot, muted); a per-item `Icon` overrides the glyph. `Columns` reflows items row-major into balanced columns; text hangs indented from the glyph width. `GlyphTone` (a `*ColorRole`, `nil` = per-state default) re-skins all glyphs. `Fill` distributes rows to span the box height — place it in a `VAlignFill` card body to fill the card. Use this, not `List{Kind: ListChecklist}`, for feature/offer lists |
| `ChipRow` | `Label string; Chips []ChipSpec` (`{Label string; Tone ChipTone; Color ColorRole; Icon string}`); `Wrap bool`; `Align HAlign` | native; a horizontal row of content-fit chip pills (a tag / category / capability strip) with an optional leading `Label`. Each `ChipSpec` mirrors `Chip` (`ChipTint`/`ChipSolid`/`ChipOutline` + an optional leading `Icon`). `Wrap` reflows chips onto new lines within the width (zero = one line; set it true for a long strip); `Align` offsets each line. Use this, not a bullet `List`, for a tag/category row |
| `Banner` | `Lead RichText; Body RichText; Icon string; Fill ColorRole; TextColor TextColorRole; Trailing []SlideNode` | native; a full-width filled "big takeaway / promo / CTA" strip with a leading icon + bold `Lead` + `Body` on the left and optional right-aligned `Trailing` children (a `Stat`/`Button`). `Fill` defaults to accent (its zero value); the text auto-contrasts against the fill unless `TextColor` is set. Distinct from the small side-bar `Callout` — use `Banner` for a wide promo/closing band. Embed a `Button` in `Trailing` for the action |
| `IconRows` | `Rows []IconRow` (`{Icon string; Label RichText; Meta RichText; Tone RowTone}`); `Fill bool`; `GlyphColor ColorRole` | native; a vertical stack of `[icon | label | optional right-aligned meta]` rows — the "integrations / capabilities / sources" list. `RowTone` = `RowPlain` / `RowPill` (a `SurfaceAlt` frame). `Fill` distributes rows to span the box (place it in a `VAlignFill` card); `GlyphColor` tints the icons (zero = accent). Use this, not a bullet `List`, for an icon-label row list |
| `Lockup` | `Caption string; AssetID AssetID; Icon string; AssetSide AssetSide; MaxHeight pptx.EMU; Align HAlign` | a "powered by / in partnership with" mark — a caption + a small partner logo as one centered inline unit. Set **exactly one** of `AssetID` (a logo resolved via the AssetResolver → a pic) or `Icon` (a curated glyph, media-free). `AssetSide` = `LeadCaption` (caption then logo) / `TrailCaption` (logo then caption). `MaxHeight` bounds the square logo (0 = default; the engine does not parse pixel aspect — your bytes drive it). Use on a cover/closing slide |
| `SectionDivider` | `Eyebrow, Label string` | native (full-bleed) |
| `Table` | `Headers []RichText; Rows [][]RichText; Caption string; Style *TableStyle` | native — a nil `Style` is a plain banded table; a non-nil `Style` (`{HeaderFill, Zebra bool; HighlightCol int; RowLabelCol bool; HeaderGroups []HeaderGroup{Label string; Span int}}`) renders a **comparison matrix**: an accent header band (auto-contrast text), zebra body striping, a highlighted 1-based column (accent tint + heavier border), an emphasized first column (row labels), and a grouped header row (each group merges `Span` columns). All token-driven; nil = byte-identical plain table. For **check/cross/dot/bar value cells**, compose a `Bento` of `Checklist`/`IconRows` cells (a native table cell holds only text, so glyph cells live there, not in `Table`) |
| `Flow` | `Orientation FlowOrientation; Steps []FlowStep; Connector ConnectorKind` | native |
| `DataMark` | `Kind DataMarkKind; Value float64; Values []float64; Orientation FlowOrientation; Color *ColorRole; Label string` | native vector micro-chart (no raster). `DataMarkBar` = a progress/capacity bar (track + accent fill to `Value` 0..1, optional inline `Label`; horizontal or vertical); `DataMarkBars` = a small bar group (one bar per `Values` entry); `DataMarkSparkline` = a trend polyline through `Values` + end dot; `DataMarkDonut` = a single-value ring + centered `Label` (e.g. a 331° arc at 0.92); `DataMarkGauge` = a 270° speedometer (donut/gauge are native `blockArc` ring sectors). Values are 0..1; `Color` (nil = accent) over a SurfaceAlt track. Embeds in a `Card`/`Bento` cell. Use for in-card KPIs/capacity instead of rastering a trivial chart via `Chart` |
| `LogoWall` | `Logos []LogoEntry; Columns int; Tone LogoToneKind; Caption string` | native asset grid (customers/investors/integrations). `LogoEntry` = `{AssetID, Alt}`. Each logo is contained (not cropped) + centered at a common size; `Tone` = `LogoToneNone` / `LogoToneMono` (brand-neutral) / `LogoToneBrand` (accent) recolors the set uniformly via duotone so mixed logos cohere; `Caption` = an optional heading. Needs an `AssetResolver`; a missing logo warns + is skipped |
| `Funnel` | `Stages []FunnelStage` | native tapering funnel — `FunnelStage` = `{Label, Value string; AccentIndex int}`; a vertical stack of bands narrowing top→bottom with optional per-stage values (conversion/marketing funnel) |
| `Cycle` | `Stages []CycleStage` | native loop — `CycleStage` = `{Label, Icon string; AccentIndex int}`; stage cards on a ring (clockwise from top) with directional arrows (lifecycle/loop). For a **branch** (1→M) use `Tree` |
| `Tree` | `Root TreeNode; Orientation FlowOrientation` | native hierarchy / org chart / taxonomy. `TreeNode` = `{Label, Detail, Icon string; Children []TreeNode; AccentIndex int}`. A balanced tidy tree (leaves spread evenly, internal nodes centered over descendants) with parent→child elbow edges + rounded-rect node cards (accent border). `FlowVertical` (top-down, default) / `FlowHorizontal` (left-right). Use for org/structure/decomposition slides; depth/breadth past the region clamps + warns |
| `Quadrant` | `AxisX, AxisY QuadrantAxis; Quadrants [4]QuadrantCell; Items []QuadrantItem` | native 2x2 positioning map (effort/impact, BCG, landscape). `QuadrantAxis` = `{LowLabel, HighLabel}`; `QuadrantCell` = `{Title; Fill *ColorRole}` (0=TL,1=TR,2=BL,3=BR); `QuadrantItem` = `{X, Y float64 (0..1, origin bottom-left); Label; AccentIndex}`. Labeled axes + per-quadrant tints + plotted dots, all native + deterministic. Use for positioning/landscape slides |
| `Timeline` | `Milestones []Milestone; Lanes []TimelineLane; Bands []TimelineBand` | native — a roadmap. `Milestone` = `{Position float64 (0..1 along the axis); Label, Detail, Icon string; AccentIndex int}`; `TimelineLane` = `{Label string; Milestones []Milestone}`; `TimelineBand` = `{From, To float64; Label string; Fill ColorRole}`. A horizontal axis with markers (accent dot or curated `Icon`) at proportional positions; labels stagger above/below to avoid collision. Empty `Lanes` = one implicit lane from `Milestones`; non-empty `Lanes` = swimlane rows. `Bands` are phase regions (Now/Next/Later or Q1/Q2…) behind the axis. `AccentIndex` cycles `[Accent, AccentAlt, Info, Success, Warning]`. Map dates to `0..1` yourself. Use for "where we're going" / roadmap slides, not `Flow` (an equal-step pipeline) |
| `Image` | `AssetID AssetID; Alt string; Frame FrameKind; FrameName string; Crop Crop; Fit Fit; CornerRadius RadiusRole; Elevation ElevationRole; Annotations *ImageAnnotations` | **picture**. `CornerRadius` (a `RadiusRole`) rounds the picture's corners and `Elevation` (an `ElevationRole`) casts a soft drop shadow — both from theme tokens, matching the card finish; zero values leave the picture rectangular/shadowless (byte-identical). `Annotations` (a `*ImageAnnotations` = `{Pins []ImagePin{X,Y 0..1; Label, Caption; AccentIndex}; Highlights []ImageHighlight{X,Y,W,H; AccentIndex}}`) overlays numbered callout pins (+ leader captions) and highlight boxes at fractional coords for annotated screenshots/diagrams; nil = none |
| `CodeBlock` | `AssetID AssetID; Language, Caption string` | **picture** |
| `Chart` | `AssetID AssetID; Caption string` | **picture** |
| `Decoration` | `Kind DecorationKind; Preset string; AssetID AssetID; Text string; FontSize float64; Layer Layer; Anchor Anchor; Offset Position; Size Size; Bleed bool; Opacity float64; Rotation float64; Color *pptx.ColorRole` | native (preset) / **picture** (asset) / **text watermark**. `Kind` = `DecorationPreset` / `DecorationAsset` / `DecorationText`. `Color` (a `*pptx.ColorRole`, `nil` = `ColorAccent`) re-colors a preset ornament — set it for a neutral-grey paper grain, an inverse-white starfield (or the scatter family: scatter_dot/_star/_plus/_ring), or any brand-role texture/glow. `DecorationText` draws an oversized, low-opacity ghost number/word (`Text`, sized by `FontSize` in points or a box-height default) behind the body — colored by `Color` at the `Opacity` alpha; set a low `Opacity` (e.g. 0.08) for a faint structural number |

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
  `DecorationAsset` (caller bytes, picture) / `DecorationText` (text watermark);
  `Layer` = `LayerBackground`, `LayerForeground`. `Opacity` is `0..1` (0 = opaque).
- Curated ornament `Preset` names: `grid_dots`, `noise_overlay`, `starfield`
  (an organic scatter of size/alpha-varied dots — a dark-slide starfield; more
  dots over a bigger `Size`/`Bleed` box), `radial_glow`, `glow_ring`,
  `corner_bracket`, `chevron_arrow`. Set `Color` to recolor any of them.
- `Decoration.Pitch` (EMU) sets the lattice spacing for the pattern ornaments
  (`grid_dots` / `noise_overlay` / `starfield`): their dot count derives from the
  box at that pitch, so a full-bleed texture keeps a fine, consistent density
  (e.g. `Pitch: pptx.In(0.4)` over a full-slide box). 0 = the legacy fixed count.
  Capped for file size (a too-fine pitch warns and stops at the cap).

### Container nodes

| Node | Fields | Notes |
|------|--------|-------|
| `TwoColumn` | `Ratio ColumnRatio; Left, Right []SlideNode; Join ColumnJoin; JoinLabel string; JoinPosition JoinPosition` | both sides must be non-empty. `Join` draws a centered seam element: `JoinBadge` (a "VS"-style `JoinLabel` badge) or `JoinArrow` (a connector arrow); `JoinNone` (default) draws nothing. `JoinPosition` = `JoinSeam` (centered, default) / `JoinTopBridge` / `JoinBottomBridge` — a horizontal accent bracket + content-fit label pill spanning both column tops/bottoms (the "one X, two ways" header); a bridge reserves a band so it sits above/below the content |
| `Grid` | `Columns int (2..4); Ratio []int; Gap SpaceRole; Cells []SlideNode; Connectors []GridConnector` | cell count must be a multiple of `Columns`; `Ratio` empty or len == `Columns`. `Connectors` (`{Between [2]int; Kind ConnectorKind; Label string}`) draw a glyph in the gutter between two adjacent columns (`Between` = `{c, c+1}`) — reuse the flow `ConnectorArrow`/`ConnectorBiArrow`/… set — so an architecture/pipeline grid reads as flow; empty = none, byte-identical |
| `Bento` | `Columns int (≥1); Rows []BentoRow` (`{Label string; Cells []BentoCell}`, `BentoCell{Span int; Node SlideNode}`); `WeightedRows bool` | row-labeled grid: rows with an optional left label and cells of variable column span on a shared grid (a span-S cell = S of `Columns` units). A row's spans sum to ≤ `Columns`; the gutter is reserved only when some row has a `Label`. `WeightedRows` (opt-in) sizes each row to its content's preferred height — dense rows grow, sparse rows shrink, clamped so the bento always fits its region; default equal rows are byte-identical |
| `Card` | `Header, Eyebrow, Icon, HeaderPill string; Body []SlideNode; BodyLayout BodyLayout; Fill ColorRole; Outline bool; BorderStyle BorderStyle; Size CardSize; Layout CardLayout; Elevation ElevationRole; HeaderFill, StatusDot *ColorRole; Watermark string; BodyVAlign VAlign; PaddingScale int; Ribbon *Ribbon; FillGradient *GradientFill; Backdrop *Decoration; ImageFill AssetID` | accent card; all fields beyond `Header/Body/BodyLayout/Fill/Outline/Elevation` are additive (zero values reproduce the prior render). `FillGradient` (a `*GradientFill` = `{From, To pptx.ColorRole; Angle int}`, nil = solid `Fill`) replaces the flat surface with a 2-stop linear gradient for a top-to-bottom depth shift. `Backdrop` (a `*Decoration`, nil = none) draws a focal glow/halo behind the card's box before its fill — set a center-anchored, bleeding `radial_glow` with a low `Opacity` to spotlight the card. Rich visuals: `HeaderFill` (colored header band, body keeps `Fill`), `StatusDot` (top-right dot), `Watermark` (large faint label behind the body). `HeaderPill` sizes to its label on one line (any length renders intact, not wrapped inside a fixed chip). `HeaderFill`/`StatusDot` are `*ColorRole` — take a role's address; `nil` omits. `BodyVAlign` (opt-in) distributes the vertical body within the card via any of the 8 `VAlign` modes — `VAlignBottom` pins secondary content to the card bottom, `Justify`/`Fill`/`FillCapped`/`Balanced`/`Fit` spread/grow/cap/balance/compress it; zero `VAlignTop` is byte-identical. `PaddingScale` (basis points; 0/10000 = unchanged) tightens a dense card's interior inset (floored at a minimum) so it reclaims body space. `Ribbon *Ribbon` (`{Text; Position RibbonPos(TopBar/CornerTL/CornerTR/CornerStar); Color *ColorRole; TextColor}`, nil = none) pins an emphasis badge outside the header flow — distinct from `HeaderPill`; `RibbonTopBar` reserves a band so the body shifts down, `RibbonCornerStar` is a star glyph, corner TL/TR are corner text tabs — use it to flag the "most popular" tier. `ImageFill AssetID` (resolved via the `AssetResolver`, "" = the solid/gradient `Fill`) fills the card surface with a **cover-fit photo** instead of a color — the photo is center-cropped to cover and the rounded corners still clip it; an unresolvable ID warns and falls back to `Fill`. Pair it with a dark `HeaderFill` or a `Backdrop` so header text stays legible over the photo |
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
- `Button`: non-empty `Label`; any `LeadingIcon`/`TrailingIcon` must resolve in
  the icon registry (curated ∪ extensions).
- `Checklist`: non-empty `Items`; `Columns` in `0..3` (0 = 1); each item's
  `State` valid; any per-item `Icon` must resolve in the icon registry.
- `ChipRow`: non-empty `Chips`; each chip's `Tone` valid; any chip `Icon` must
  resolve in the icon registry.
- `Banner`: `Trailing` children validated recursively; the banner `Icon` and any
  `Trailing` child icon must resolve in the icon registry.
- `IconRows`: non-empty `Rows`; each row's `Tone` valid; any row `Icon` must
  resolve in the icon registry.
- `Lockup`: exactly one of `AssetID` / `Icon` set; `AssetSide` valid;
  `MaxHeight >= 0`; an `Icon` must resolve in the icon registry.
- `Table`: at least one header column; every row width equals the header width.
- `TwoColumn`: non-empty `Left` and `Right` (children validated recursively).
- `Grid`: `Columns` in `2..4`; `Ratio` empty or length == `Columns`; non-empty
  `Cells`; cell count a multiple of `Columns`; each `Connectors` entry between
  adjacent in-range columns (`Between[1] == Between[0]+1`) with a valid kind.
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
against the real background.

Card and container **chrome text auto-contrasts** against the surface behind it:
a card header, eyebrow, header pill, the TwoColumn join-badge label, and a `Stat`
value pick a light text color on a dark fill / dark-variant slide and the normal
dark default on a light surface, so a header is never black-on-dark and a same-hue
eyebrow never goes invisible. This is a deterministic, byte-identical-on-light
**mechanism**, not a policy — supply an explicit run `Color` to override it, and a
light-surface deck renders exactly as before. (The engine still encodes no taste:
it picks the contrast-correct token by a fixed luminance rule and reports the
resolved colors via `Stats.Colors`.)

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
`VAlignBalanced`, and `VAlignFit`. `VAlignFill` pins fixed leaves at the top and **grows the flexible
nodes** — the containers (`Grid`, `TwoColumn`, `Card`, `CardSection`, `Bento`,
`Table`) plus `Image`/`Chart` — to consume the remaining height, so a sparse slide
fills its frame. The leftover height is shared proportionally and
deterministically; text leaves keep their size, and a slide with no flexible node
just top-aligns. `VAlignFillCapped` is the same but caps each flexible node's
growth at a pinned factor of its preferred height (at most double), turning the
leftover slack into even spacing — so a near-empty card can't balloon while a dense
one starves. `VAlignBalanced` distributes a sparse stack's slack as an even rhythm
(top margin + widened gaps), optically centered — for a sparse cover/closing that
would otherwise cluster with a large void (distinct from `VAlignJustify`, which
puts all slack into the gaps with no margins).

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
