# RFC-001 — pptx-go

> **Status:** Draft (v0). This is the source of truth for pptx-go's design.
> Conflicts with `CLAUDE.md` / `AGENTS.md` resolve in favour of this file
> (per `CLAUDE.md §2`). When the RFC and a phase plan disagree, the RFC wins;
> file a follow-up to fix the plan.

---

## 1. Executive summary

**pptx-go** is a pure-Go library for authoring and reading PowerPoint (PPTX,
Open Office XML) files. It is published under
`github.com/hurtener/pptx-go` and licensed Apache-2.0. The library has two
public layers:

```text
Layer 2 — the scene renderer (`scene`)
            ▲
            │   pure composition; never reaches under the builder API
            │
Layer 1 — the builder (`pptx`)
            │
            │   token-aware slide/shape/text/media/table primitives
            ▼
Layer 0 — the OOXML & OPC foundation (`internal/opc`, `internal/ooxml`)
            (private; raw XML wire types live here and nowhere else)
```

Layer 1 — `pptx` — is a clean, theme-aware **builder**: presentations, slides,
shapes, rich text, tables, media, and a `Theme` whose semantic tokens drive
every visual property. A direct consumer of `pptx` writes generic Go and gets
a production-grade deck.

Layer 2 — `scene` — is an optional **renderer** for a high-level scene IR
modeled on pengui-slides' v4 IR (hero/prose/card/grid/flow/decoration/…). A
consumer of `scene` builds a typed `Scene` value, calls `Render`, and gets a
PPTX. The renderer is a thin composer of `pptx` calls; nothing in `scene`
reaches under the builder. This decoupling means pengui-slides v2 (running on
Dockyard) can shed its own renderer and call `scene.Render(...)` directly,
while a generic consumer can use `pptx` and stay clear of any opinion about
slide content.

**The library is the engine. The IR is a consumer.** When the IR evolves, the
scene renderer changes — the builder doesn't.

The starting feature set is the slide-rendering subset of pengui-slides
v4's IR (§21 — *Compatibility with pengui-slides' IR*). Most nodes
render as native PPTX shapes built from the IR's typed fields; a
handful (`image`, `chart`, `decoration` of `asset_ref` kind,
`code_block`) carry an `asset_id` and render as `pic` shapes whose
bytes the caller provides through an `AssetResolver` (§10.6). Native
`c:chart` parts and broader OOXML-feature coverage are explicit V2
follow-ups (§24).

pptx-go also exposes the mechanisms to embed fonts (§7.6), carry
speaker notes (§8.8), group slides into named PPTX sections (§8.7),
and ship in 16:9 and 4:3 slide formats (§5). Internal correctness
hardening — emitting OOXML that doesn't trigger PowerPoint's "this
file has been repaired" prompt — runs unconditionally as part of the
write path (§6).

**Engine, not opinion.** pptx-go is the IR-to-PPTX engine. It does not
decide *what* should be in the deck (that's the caller's IR) or *how*
the deck should look (that's the caller's Theme); it decides only how
to render valid OOXML faithfully. Product behaviors — image-only
deck modes, projection-legibility text boosts, validation pipelines,
HTML preview, markdown ingestion — live in callers (go-slides), not
here (D-026).

---

## 2. Goals and non-goals

### 2.1 Goals (V1)

- **G1 — Match-and-surpass pengui-slides' IR.** Every node type pengui-slides
  v4 emits today renders through `scene` to a high-quality PPTX (§21). Quality
  ≥ pengui-slides' current output; ideally better.
- **G2 — Theme-driven rendering.** Visual properties flow through semantic
  tokens (colors by role, type scale, spacing scale, radius, elevation). A
  scene authored against tokens renders unchanged when the theme swaps (§7).
- **G3 — Two clean public layers.** A direct consumer of `pptx` (Layer 1)
  does not need to learn the scene model; a consumer of `scene` (Layer 2)
  does not need to think in OOXML. The seam between them is API surface, not
  conditional logic (§3).
- **G4 — Pure Go, zero runtime dependencies.** The shipped artifact compiles
  CGo-free; runtime imports are stdlib only. Embedded curated assets (icons,
  ornaments, frames) ship via `go:embed` and are not dependencies (§17).
- **G5 — OOXML-by-isolation.** Raw OOXML wire types live only in
  `internal/ooxml`. A schema bump is a localized, golden-tested update;
  consumers above the seam never re-author against XML.
- **G6 — Round-trip fidelity for pptx-go-authored decks.** Every PPTX
  `pptx-go` emits, `pptx-go` can read back into the same model without
  loss (§16). Parsing arbitrary third-party PPTX is a stretch goal, not a
  V1 contract.
- **G7 — Streaming-friendly.** The upstream's streaming/concurrent-save
  primitives are preserved, hardened, and made first-class for the builder
  (§17). The scene renderer can render slides in parallel.
- **G8 — Production-grade engineering hygiene.** Doc-driven build with an
  RFC, phase plans, decisions log, glossary, drift-audit, preflight gate,
  agent skills, and published docs site (`CLAUDE.md §§ 1, 16, 19`).

### 2.2 Non-goals (V1)

- **N1 — Native `c:chart` parts.** V1 ships chart support via image shapes
  (caller renders PNG/SVG bytes; library inserts and positions). Native
  c:chart parts are a V2 wave (§15).
- **N2 — General third-party PPTX read.** V1 reads back its own output
  losslessly. Reading arbitrary external decks (PowerPoint files saved from
  Excel, Keynote exports, third-party libraries) is best-effort, not a
  contract.
- **N3 — Animations, transitions, speaker notes.** Out of V1 scope. Speaker
  notes are a likely V1.x add (well-bounded, broadly useful); animations and
  transitions are a V2+ topic.
- **N4 — SmartArt-equivalents as native PPTX SmartArt.** Process pipelines
  (the scene `flow` node), grids, and cards render as composed PPTX shapes,
  not as native SmartArt diagrams (which would couple us to PowerPoint's
  SmartArt XML dialect).
- **N5 — A live editor / inspector / GUI.** This is a library, not a service.
  No HTTP surface, no daemon, no inspector. Diagnostics are returned as Go
  values (§18).
- **N6 — Document-authoring concerns.** pptx-go renders slides to PPTX
  and nothing else. Documents (PDF, HTML, print formats) and the
  document-authoring concepts that go with them are outside its scope
  by design (D-026).

---

## 3. Architecture overview

### 3.1 The two-layer rule

```text
                                ┌────────────────────────────┐
   ▲                            │       scene  (Layer 2)     │
   │                            │   IR types, validators,    │
   │     optional, opinionated  │   layout policy, asset     │
   │     "scene IR → PPTX"      │   registry, disposition    │
   │                            │   matrix, render loop      │
   │                            └─────────────┬──────────────┘
   │                                          │  composes
   │                                          ▼
   │                            ┌────────────────────────────┐
   │                            │        pptx (Layer 1)      │
   │     mandatory,             │   Presentation, Slide,     │
   │     general-purpose        │   Shape, RichText, Table,  │
   │     "PPTX builder"         │   Media, Theme, Token res. │
   │                            └─────────────┬──────────────┘
   │                                          │  uses
   │                                          ▼
   │                            ┌────────────────────────────┐
   │     PRIVATE                │   internal/opc + ooxml     │
   │     "raw OOXML wire"       │   XML types, namespaces,   │
   ▼                            │   parts, rels, codecs      │
                                └────────────────────────────┘
```

**The rule.** Code in `scene` calls code in `pptx`. Code in `pptx` calls code
in `internal/...`. Calls in the reverse direction are forbidden, as is any
import of `internal/...` from outside the module. The §3 layout makes this
mechanically enforceable: Go's `internal/` convention does it for the lower
seam; a drift-audit lint does it for the upper seam (`scripts/drift-audit.sh`
greps for `internal/` imports from `scene/`).

### 3.2 Why two layers, not one

Two layers are the smallest architecture that satisfies both consumers
without bleeding either's concerns into the other.

- A generic consumer (someone authoring a deck programmatically without an
  IR) needs a low-level builder that *doesn't* assume their content model.
  That's `pptx`.
- An IR-driven consumer (pengui-slides v2, or anyone else with a typed scene
  model) needs a high-level renderer that *does* assume one. That's `scene`.

A single combined layer would either force the generic consumer to learn the
scene IR (and tolerate its opinions), or force the IR consumer to write
their own renderer (which is what pengui-slides does today, and the cost we
are removing). The two-layer split keeps the surface honest: changes to the
scene IR don't reshape the builder, and changes to the builder don't reshape
the IR consumer's call sites.

### 3.3 Module layout

```text
github.com/hurtener/pptx-go
├── pptx/                    # Layer 1 — the builder (PUBLIC)
│   ├── presentation.go      # Presentation: New, Open, Save, StreamSave
│   ├── slide.go             # Slide: builder + introspection
│   ├── shape.go             # Shape primitives (text, geometric, group, conn)
│   ├── text.go              # RichText, Paragraph, Run, paragraph layout
│   ├── table.go             # Table builder
│   ├── media.go             # Image / media insertion + dedup
│   ├── theme.go             # Theme, Token, ColorRole, TypeScale, etc.
│   ├── tokenresolve.go      # Token → OOXML value resolution
│   ├── units.go             # EMU / Pt / Px / Cm conversions
│   ├── geom.go              # Position, Size, Box, Anchor, Inset
│   ├── stream.go            # Streaming open / save passthrough
│   └── doc.go               # Package-level godoc overview
│
├── scene/                   # Layer 2 — the scene renderer (PUBLIC)
│   ├── scene.go             # Scene, Slide, Render entrypoint
│   ├── nodes.go             # IR node types (leaf + container, discriminated)
│   ├── richtext.go          # Scene-side RichText (tokens, semantic colors)
│   ├── tokens.go            # ColorRole, TextColorRole, etc.
│   ├── layout/              # Layout engine: two_column, grid, card, flow
│   ├── icons/               # Curated icon registry (lucide subset)
│   ├── ornaments/           # Curated preset ornaments
│   ├── frames/              # Curated device frames (browser/phone/desktop/laptop)
│   ├── disposition.go       # Per-node native vs raster decision
│   ├── validate.go          # Stage-1 scene-side validation
│   └── doc.go
│
├── internal/                # PRIVATE
│   ├── opc/                 # Was upstream `opc/` — package/part/rels/zip
│   ├── ooxml/               # Was upstream `parts/` — OOXML XML structs
│   │   ├── presentation/    # presentation.xml, presProps, viewProps
│   │   ├── slide/           # slide, slideLayout, slideMaster
│   │   ├── theme/           # theme1.xml
│   │   ├── core/            # core.xml, app.xml, custom.xml
│   │   ├── drawing/         # drawingML shapes, fills, text bodies
│   │   ├── relations/       # relationship XML structs
│   │   └── namespaces.go    # canonical namespace URIs
│   ├── render/              # Internal shape composition primitives used by both
│   │                        # `pptx` (high-level shape helpers) and `scene` is
│   │                        # NOT here — `scene` only uses `pptx`. This package
│   │                        # is for builder-internal helpers (e.g. text body
│   │                        # XML generation).
│   ├── ids/                 # Shape-id / rel-id allocation, atomic counters
│   └── coveragecheck/       # Mechanical coverage band gate
│
├── assets/                  # Embedded curated assets (go:embed)
│   ├── icons/               # Lucide-subset SVGs
│   ├── ornaments/           # Preset ornament SVGs
│   └── frames/              # Device frame SVGs + shape recipes
│
├── docs/                    # Author / contributor docs
│   ├── plans/               # Master plan + phase plans + _template.md
│   ├── research/            # Phase-planning research briefs + INDEX.md
│   ├── specifications/      # Vendored OOXML / OPC spec snapshots
│   ├── design/              # Theme/token catalog, disposition matrix
│   ├── site/                # Published tech-docs site (Phase 20+)
│   ├── glossary.md
│   └── decisions.md
│
├── examples/                # Runnable Go examples (1 per node type at minimum)
├── test/integration/        # Cross-layer integration tests
├── scripts/                 # preflight.sh, drift-audit.sh, smoke/, hooks/
├── skills/                  # Agent skills (Phase 20+)
├── templates/               # Starter .pptx templates for `scene` ingestion
│
├── RFC-001-pptx-go.md       # This file — the design RFC
├── README.md                # User-facing intro (no internal vocabulary — §19)
├── CHANGELOG.md             # Keep a Changelog (Phase 0)
├── CLAUDE.md / AGENTS.md    # Operational rules — mirrored verbatim
├── LICENSE                  # Apache-2.0
├── Makefile
├── go.mod
└── go.sum
```

A new top-level directory not listed above is wrong; propose it in this RFC
first.

### 3.4 What changes from the upstream

The upstream (`github.com/Muprprpr/Go-pptx`) is the substrate. Wave 0
renames the module to `github.com/hurtener/pptx-go` and reorganizes the
existing code as follows:

| Upstream | New location | Change |
|----------|--------------|--------|
| `opc/` | `internal/opc/` | Move into `internal/`; stays public API to lower layers; no rename of types. |
| `parts/` | `internal/ooxml/` (with subpackages) | Move into `internal/`; the XML structs were the right idea, the package boundary wasn't (no subpackages, mixed concerns). Phase 01 reorganizes by OOXML part family. |
| `pptx/` | `pptx/` | Stays at the top-level. Phase 03–04 rewrite this as the new builder; the upstream surface is preserved as deprecated aliases where it makes sense, removed where the new API supersedes it cleanly. |
| `utils/` | folded into `pptx/units.go` and `pptx/geom.go` | Small, no reason to keep a `utils` package. |
| `test/` | `test/integration/` for cross-layer tests; per-package tests live next to their package | Match Go convention. |
| `docs/` | `docs/` | Restructured per §3.3; existing per-part docs land under `docs/site/` once the published-docs phase happens. |
| `main.go` | deleted | Library has no binary; the upstream `main.go` is dev scratch. |

Greenfield is rejected (D-008): the upstream's correctness on namespace
handling, relationship ID allocation, content-types ordering, and streaming
save is hard-won and worth keeping. Incremental refactor preserves it.

---

## 4. Binding properties

A change that weakens any of P1–P4 is **wrong**. Reach for this RFC, not the
keyboard. P1–P4 are restated verbatim in `CLAUDE.md §1`.

### P1 — Two layers, one library

The public surface is `pptx` (Layer 1, the builder) and `scene` (Layer 2,
the renderer). `scene` composes `pptx`; nothing in `scene` reaches under
the builder. A new scene primitive adds a new builder call **only** when the
primitive requires a genuinely new OOXML capability. Otherwise the new
primitive composes existing builder calls.

**The seam test.** Given a new scene node, ask: "could a user of `pptx` write
this themselves with current builder calls?" If yes, the node lives entirely
in `scene` (a composer). If no, the new builder capability is added first;
the scene node lands in the same or a follow-up PR. A node that smuggles raw
OOXML through `scene` is wrong.

### P2 — Tokens, not literals

All visual properties — colors, typography, spacing, radius, elevation —
flow through a `Theme` whose semantic tokens map to OOXML values. The default
authoring path on the builder is **token-typed**: `slide.SetFill(theme.Color
(pptx.ColorAccent))`, not `slide.SetFillHex("#3366FF")`.

Literal/escape-hatch APIs exist for power users (`pptx.RGB(0x33, 0x66, 0xFF)`
returns a `Color` that the builder accepts), but the *idiomatic* and
*documented* path is tokens. The library's "make it look right by default"
behavior comes from theme inheritance, not from hardcoded defaults.

This matches pengui-slides' design-soul guarantee: change the theme, the same
scene renders in the new visual language. We push the same guarantee into the
builder so generic consumers benefit too.

### P3 — OOXML by isolation

Raw OOXML wire types — XML structs, element names, namespace URIs, schema
specifics — live only in `internal/ooxml`. Neither `pptx` nor `scene` imports
raw XML structs; user-facing types are pure Go data, and the codec layer
translates. A spec/schema bump is a deliberate update with a vendored spec
snapshot in `docs/specifications/`, a golden-test diff, and a single
codec change in `internal/ooxml`.

**The codec rule.** `internal/ooxml` exposes typed Go domain objects to
`pptx`; it never exposes its own `xml.Marshaler`/`xml.Unmarshaler` types.
The reverse direction — a `pptx` change that requires `internal/ooxml` to
emit a new XML element — adds the element to the codec, never inlines XML
generation in the upper layer.

### P4 — No CGo, stdlib-only runtime

The shipped artifact is pure Go with no CGo. Runtime imports are the
standard library only — no third-party Go modules. `-race` tests run with
`CGO_ENABLED=1` (the race detector requires CGo); shipped binaries do not.

Embedded curated assets (`assets/icons/`, `assets/ornaments/`,
`assets/frames/`) ship as vendored bytes via `go:embed`. Bytes are not
dependencies; the rule is about *Go modules*, not file content.

A future feature that would require a non-stdlib runtime dep (e.g. native
PNG → WEBP conversion, a third-party SVG raster) is **not** added; the
caller is responsible for pre-processed assets.

---

## 5. Module identity

| | |
|---|---|
| Module | `github.com/hurtener/pptx-go` |
| License | Apache-2.0 |
| Starting version | v0.1.0 |
| Go version (pinned) | 1.24 |
| Public packages | `pptx`, `scene` (and `scene/...` subpackages) |
| Private packages | everything under `internal/` |
| Runtime deps | none beyond the standard library |
| Slide formats (V1) | `pptx.Slides16x9` (default), `pptx.Slides4x3` |
| Print/document formats | V2 (out of V1 — no PDF rendering in pptx-go) |

`pptx.Format` is a first-class enum on `*Presentation`:

```go
pres := pptx.New(pptx.WithFormat(pptx.Slides16x9))
// or:
pres := pptx.New(pptx.WithFormat(pptx.Slides4x3))
```

The format selects slide dimensions, master defaults, and theme-default
font sizes. Print formats (A4 / Letter portrait) are out of pptx-go's
scope — they're document-rendering concerns, not slide concerns. A
future slide-format addition is a codec change behind a new constant;
no per-format branching in user code.

API-stability promise: V1.0.0 (when shipped) commits to backwards-compatible
evolution of the `pptx` and `scene` public APIs. Pre-V1 (v0.x), public APIs
may evolve breaking-ly between minor versions; breaking changes are noted in
CHANGELOG.md.

---

## 6. The OPC + OOXML foundation (`internal/opc`, `internal/ooxml`)

The OPC (Open Packaging Convention) layer is the ZIP container plus the
`[Content_Types].xml`, relationship-XML, and pack-URI plumbing every OOXML
file shares (PPTX, DOCX, XLSX). The OOXML layer is the OOXML-specific XML
schemas: presentation, slide, master, layout, theme, drawingML shapes, etc.

### 6.1 `internal/opc`

The upstream `opc/` package is solid and ships these capabilities, which we
preserve:

- `Package` — eager-load model (small files).
- `StreamPackage` — lazy-load + streaming-save model (large files).
- `Part`, `PartCollection` — addressable container content.
- `ContentTypes` — `[Content_Types].xml` model.
- `Relationships` — `.rels` model with thread-safe atomic ID allocation.
- `PackURI` — pack-URI normalization and resolution.
- `ConcurrentZipCollector` / `ConcurrentStreamSave` — goroutine-based
  parallel ZIP write.
- `ResourceDedupPool` — `sync.Map`-based media deduplication.

These move to `internal/opc/` in Wave 0. The public package retains all of
the upstream's behavior; the move is a rename, not a rewrite.

### 6.2 `internal/ooxml`

The upstream `parts/` package is the right intent (OOXML XML structs) but
the wrong shape: a single package mixes presentation, slide, master, theme,
chart, media, embedding, and namespace utilities. Phase 01 reorganizes it as
subpackages, one per OOXML part family:

```text
internal/ooxml/
├── namespaces.go      # canonical NS URIs as constants
├── presentation/      # presentation.xml, presProps, viewProps
├── slide/             # slide, slideLayout, slideMaster
├── theme/             # theme1.xml: colors, fonts, formatScheme
├── core/              # core.xml, app.xml
├── drawing/           # drawingML: shapes, fills, geometries, text bodies
├── chart/             # c:chart wire types (V2 — placeholder package now)
├── relations/         # relationship XML structs
└── media/             # media-part typing (image/png, audio/mpeg, …)
```

Each subpackage is small and single-purpose; its types are Go structs that
marshal cleanly to OOXML with `encoding/xml`. **No subpackage imports another
subpackage's XML types except via documented shared helpers in
`namespaces.go` and a future `internal/ooxml/common`** — the rule is to keep
the part families independent so a spec bump in one family is localized.

### 6.3 Codec versioning and vendored specs

OOXML is evolving (ISO/IEC 29500 Editions 1–5; Microsoft's transitional and
strict profiles). pptx-go targets the **transitional** profile by default,
which is what PowerPoint emits and reads cleanly. Spec snapshots used in
implementation land in `docs/specifications/` pinned by edition + date. A
codec change motivated by a spec re-read updates the vendored snapshot in
the same PR (`CLAUDE.md §10`).

Per-part codecs accept a `protocolVersion` style discriminator only when the
spec genuinely has multiple wire shapes for the same concept (chart XML is
the most likely site). For V1, every codec is single-version; the
multi-version codec pattern is reserved for V2.

---

## 7. The Theme & Token model

The Theme is the single source of visual truth at render time (P2). It maps
**semantic tokens** to OOXML values. Builder calls take tokens; the resolver
materializes the OOXML value at write time.

### 7.1 Token taxonomy

```go
// Semantic color roles — page-level surfaces.
type ColorRole int
const (
    ColorCanvas ColorRole = iota
    ColorSurface
    ColorSurfaceAlt
    ColorAccent
    ColorAccentAlt
    ColorAccentWarm
    ColorSuccess
    ColorWarning
    ColorError
    ColorInfo
)

// Semantic text colors — for inline runs.
type TextColorRole int
const (
    TextPrimary TextColorRole = iota
    TextSecondary
    TextTertiary
    TextInverse
    TextMuted
    TextAccent
    TextAccentAlt
    TextSuccess
    TextWarning
    TextError
)

// Typography scale.
type TypeRole int
const (
    TypeDisplay TypeRole = iota
    TypeH1
    TypeH2
    TypeH3
    TypeH4
    TypeH5
    TypeBody
    TypeBodySmall
    TypeCaption
    TypeMono
    TypeCode
)

// Spacing scale (returns EMUs at resolve time).
type SpaceRole int
const (
    SpaceXS SpaceRole = iota
    SpaceSM
    SpaceMD
    SpaceLG
    SpaceXL
    Space2XL
)

// Radius / corner scale.
type RadiusRole int
const (
    RadiusNone RadiusRole = iota
    RadiusSM
    RadiusMD
    RadiusLG
    RadiusFull
)

// Elevation / shadow scale.
type ElevationRole int
const (
    ElevationFlat ElevationRole = iota
    ElevationRaised
    ElevationElevated
)
```

A `Theme` is a struct holding a `ColorPalette`, a `Typography`, a `Spacing`,
a `Radii`, and an `Elevations`; each is a map from role to the concrete
value (RGB triple, font face + size + weight, EMU, etc.).

### 7.2 Builder API for tokens

```go
sl := pres.AddSlide()
sl.SetFill(pptx.TokenColor(pptx.ColorCanvas))           // token path
sl.SetFill(pptx.RGB(0xF7, 0xF7, 0xF7))                  // literal escape hatch

t := sl.AddText(pptx.Box{...})
t.Paragraph(pptx.AlignLeft).
  Run("Hello, ", pptx.RichStyle{Type: pptx.TypeH1, Color: pptx.TextPrimary}).
  Run("world", pptx.RichStyle{Type: pptx.TypeH1, Color: pptx.TextAccent})
```

`pptx.TokenColor(role)` returns a `Color` whose representation defers to the
active theme at write time; `pptx.RGB(...)` returns a literal `Color`. Both
satisfy the same interface (`pptx.Color`), so APIs that take a color are
not duplicated.

### 7.3 Theme sources

A Theme can be:

1. **Defined inline in Go** — `pptx.NewTheme()` with field-by-field setters
   or a struct literal.
2. **Loaded from a `.pptx` template** — `pptx.LoadTheme("brand.pptx")`
   reads `theme1.xml` and the master/layout color map, plus the first
   master's font scheme, and exposes them under the role taxonomy.
3. **Loaded from a JSON/YAML theme file** — `pptx.LoadThemeFile(path)`
   is V1.1+; V1 ships the two above.

Theme ingestion from a PPTX template is the canonical brand-kit story
(`scene.Render(brandedScene, scene.WithTheme(brandTheme))`).

### 7.4 Token resolution timing

Tokens are resolved lazily at write time, not at builder-call time. This is
why theme swaps work: a scene authored against `ColorAccent` re-renders in
the new accent when the theme swaps. Callers can pre-resolve a token to a
literal via `theme.Resolve(token)` when they need an early-bound value.

### 7.5 The default theme

V1 ships one default theme (light surface, neutral palette, system font
stack) sufficient to render every IR node legibly without any theme
configuration. The default is documented in `docs/design/THEME.md` and is
emitted into `templates/_default-theme.pptx` (the scaffold consumes this
when a caller didn't provide a theme).

### 7.6 Font embedding (mechanism, no default policy)

PowerPoint renders a font only if it's installed on the viewer's machine
**or** embedded in the PPTX file. A theme that names "Inter" or
"JetBrains Mono" produces a broken-looking deck on a machine without
those fonts unless the deck embeds them. pptx-go provides the
**mechanism** to embed fonts; whether to embed (and which fonts) is the
caller's choice — that's a product/distribution decision (file size vs
portability), not a library opinion.

**API.** A `Theme`'s typography references font *names*. A
`FontSource` resolves a name + style + weight to bytes; the
presentation embeds them on demand:

```go
type FontSource interface {
    // Resolve returns the bytes for the named font + style + weight.
    // A missing font returns (nil, ErrFontNotFound).
    Resolve(name, style string, weight int) ([]byte, error)
}

// On the presentation:
func (p *Presentation) EmbedFont(name, style string, weight int) error
```

`EmbedFont` reads bytes via the registered `FontSource` (set with
`pptx.WithFontSource(src)`) and writes them as an OOXML font-embedding
part. With no `FontSource` registered, `EmbedFont` returns
`ErrNoFontSource`.

**No auto-embedding.** pptx-go does **not** automatically embed fonts
referenced by the active Theme. The caller invokes `EmbedFont` for
each font it wants embedded (a common idiom: iterate the theme's
typography and call EmbedFont for each unique name+style+weight).

**Subsetting.** V1 embeds the full font face. Font subsetting (embed
only the glyphs the deck actually uses) is V1.x — the V1 contract is
faithful embedding; the V1.x optimization reduces file size.

**Why this is a mechanism, not a default.** A library doesn't know
whether a deck will be distributed to machines with the named fonts
installed or to machines without. Embedding adds file size; not
embedding risks substitution. That trade-off is the caller's. go-slides
will register a `FontSource` backed by its asset store and embed every
soul-referenced font; a different consumer that ships only to
in-corporate machines might skip embedding entirely.

---

## 8. The builder API (`pptx`) — Layer 1

The builder is the lowest layer of public API. It is the *substrate* of the
library: every higher-level surface (the scene renderer, hand-written
examples, third-party packages that wrap pptx-go) composes builder calls.

### 8.1 Top-level types

```go
type Presentation struct { ... }

func New(opts ...Option) *Presentation
func Open(path string) (*Presentation, error)
func OpenStream(path string) (*Presentation, error)
func (p *Presentation) Save(path string) error
func (p *Presentation) SaveStream(path string) error
func (p *Presentation) Write(w io.Writer) error
func (p *Presentation) Theme() *Theme
func (p *Presentation) SetTheme(*Theme)
func (p *Presentation) AddSlide(layout ...LayoutID) *Slide
func (p *Presentation) Slides() []*Slide
func (p *Presentation) AddMaster(master *Master) MasterID
func (p *Presentation) Close() error

type Slide struct { ... }

func (s *Slide) AddText(box Box) *TextFrame
func (s *Slide) AddShape(geom ShapeGeometry, box Box) *Shape
func (s *Slide) AddImage(src ImageSource, box Box) *Image
func (s *Slide) AddTable(box Box, rows, cols int) *Table
func (s *Slide) AddGroup(box Box) *Group
func (s *Slide) AddConnector(kind ConnectorKind, from, to Anchor) *Connector
func (s *Slide) Background() *Background
```

### 8.2 Boxes, positions, anchors, units

- `Box{X, Y int64; W, H int64}` — EMU coordinates. EMUs are the OOXML
  canonical unit and are exact; floating point is not used for layout.
- `Anchor{ShapeID; Side}` — Anchor for connector endpoints.
- Conversions in `pptx/units.go`:
  `pptx.Pt(12)`, `pptx.Cm(2.54)`, `pptx.In(1)`, `pptx.Px(96, dpi)`.

Slide dimensions follow the master; defaults to 16:9 at PowerPoint's
standard `9144000 x 6858000` EMU.

### 8.3 Shape, geometry, fill, line

`ShapeGeometry` is an enum over OOXML preset shapes (`rect`, `roundRect`,
`ellipse`, `triangle`, `arrow*`, `flowChart*`, …). The full preset list is
in `internal/ooxml/drawing/presets.go` and exposed as typed Go constants in
`pptx/shape.go`. Custom geometries (path commands) are V1.x — a `CustomGeom`
struct is reserved.

`Fill` is an interface implemented by `SolidFill`, `GradientFill`,
`PatternFill`, `BlipFill` (image fill), and `NoFill`. Theme tokens compose:
`pptx.SolidFill(pptx.TokenColor(pptx.ColorSurface))`.

`Line` covers stroke color, width (EMU), dash style, cap, join, and head/
tail arrow. `Line{}` (zero value) means "no line".

### 8.4 Rich text model

The rich text model is shared by `pptx` and `scene` (via re-export). Three
levels:

```go
type TextFrame struct { ... }    // shape-level container

func (tf *TextFrame) AddParagraph(opts ParagraphOpts) *Paragraph
func (tf *TextFrame) AutoFit(mode AutoFitMode)
func (tf *TextFrame) Anchor(v TextAnchor)
func (tf *TextFrame) Margins(top, right, bottom, left int64)

type Paragraph struct { ... }

func (p *Paragraph) AddRun(text string, style RunStyle) *Run
func (p *Paragraph) AddBreak()
func (p *Paragraph) AddHyperlink(text string, target string, style RunStyle) *Run
func (p *Paragraph) Bullet(kind BulletKind)              // none/disc/number/checkbox
func (p *Paragraph) Indent(level int)
func (p *Paragraph) Align(a Alignment)

type Run struct { ... }
type RunStyle struct {
    TypeRole     pptx.TypeRole      // token: typography scale
    Color        pptx.Color         // token or literal
    Bold         bool
    Italic       bool
    Underline    Underline
    Strike       Strike
    BaselineRel  BaselineShift
    Code         bool                // mono + subtle background (D-013)
}
```

Whitespace and line-wrap behavior is exact OOXML behavior: spaces are
preserved, line breaks are explicit (`AddBreak`), and word-wrap is the
text frame's responsibility. `AutoFit` modes are `AutoFitNone`,
`AutoFitNormal` (font-size shrink), `AutoFitShape` (shape grows to text).

### 8.5 Tables

```go
tbl := slide.AddTable(box, rows, cols)
tbl.SetHeaderRow(true)
tbl.SetBanding(true, false)
cell := tbl.Cell(row, col)
cell.MergeRight(2); cell.MergeDown(1)
cell.SetFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorSurfaceAlt)))
tf := cell.TextFrame()
tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Header", style)
```

Banding, header rows, alternating fills, and merged cells are all
first-class. Cell-level borders compose `pptx.Line`. A table caption is a
separate `TextFrame` shape positioned above the table by the caller (the
scene renderer composes this for `scene.TableNode`).

### 8.6 Media, dedup, alt text

```go
img := slide.AddImage(pptx.ImageFile("logo.png"), box)
img.SetAltText("Acme logo")
img.SetCrop(pptx.Crop{...})
img.SetFit(pptx.FitCover)
```

The image source is `pptx.ImageFile(path)`, `pptx.ImageBytes(b, mime)`, or
`pptx.ImageReader(r, mime)`. Dedup happens transparently: identical bytes
are written once to the package. The upstream's `ResourceDedupPool` is
preserved.

### 8.7 Backgrounds, layouts, masters

Slide backgrounds compose `Fill`. Master/layout management is exposed for
template ingestion (§13); a typical authoring flow doesn't touch masters
directly.

### 8.7 Sections (slide grouping)

PowerPoint supports grouping slides into named **sections** ("Intro",
"Body", "Conclusion") via the OOXML `sectionLst` element on the
presentation. Sections show up in the slide-sorter view and are
preserved through edit-save round-trips. pptx-go exposes sections as a
first-class builder API:

```go
intro := pres.AddSection("Introduction")
slide := pres.AddSlide()
intro.Include(slide)

body := pres.AddSection("Body")
body.Include(pres.AddSlide())
body.Include(pres.AddSlide())
```

A slide may belong to exactly one section. Sections are ordered;
`pres.AddSection` appends. A slide not assigned to a section is in the
implicit "default" section (the section header isn't emitted). This is
a low-cost V1 deliverable that pengui-slides today can't express via
its IR (yet) but go-slides will likely want as a deck-organization
primitive.

### 8.8 Speaker notes

Every slide carries an optional speaker-notes text frame, emitted into
the slide's `notesSlide` part:

```go
notes := slide.SpeakerNotes()
notes.AddParagraph(pptx.ParagraphOpts{}).
    AddRun("Talking point 1: …", pptx.RunStyle{TypeRole: pptx.TypeBody})
```

Speaker notes are RichText, themed, and round-trip through `pptx.Open`.
pengui-slides v4 carries notes on `Slide.metadata.speakerNotes`; the
scene IR's `SceneSlide.Notes` field maps directly. (D-022.)

### 8.9 Validation

The builder validates **structural** errors (a missing relationship target,
a negative dimension, a text run with no parent paragraph) and refuses to
write a malformed presentation. **Semantic** errors (legibility, contrast,
overflow) are not the builder's responsibility — the scene renderer's
validation layer (§10.4) and the caller's own checks own those.

---

## 9. The Rich Text model (shared)

The rich text model is shared in detail between `pptx` and `scene` so that
the IR can describe text uniformly with the builder's actual model.

The scene IR's `RichText` is `[]TextRun`; each `TextRun` has `text` (plain
string), an inline style (font weight, italic, underline, code, link), and
a `color` (a `TextColorRole` token). The scene renderer maps `TextRun[]`
directly to a builder `Paragraph` with one `Run` per `TextRun`.

The token-based color model means inline run colors honor theme swaps the
same way page colors do. Power users can still emit literals via
`scene.LiteralColor(...)` when they need an unbound color.

---

## 10. The scene renderer (`scene`) — Layer 2

The scene renderer takes a typed `Scene` value and renders to a
`*pptx.Presentation`. It is a pure composer of `pptx` calls — its own
private layer is a layout engine + an asset registry + a per-node
"disposition" decision; everything else is `pptx`.

### 10.1 Entrypoint

```go
type Scene struct {
    Theme  *pptx.Theme           // optional; default theme if nil
    Slides []SceneSlide
    Meta   Metadata              // title, author, subject, etc.
}

type SceneSlide struct {
    ID       string
    Layout   LayoutKind           // matches pengui-slides layout intents
    Nodes    []SlideNode          // the IR node list, top-level
    Notes    RichText             // speaker notes — V1.x
    Variant  Variant              // light/dark/etc. — selects theme variant
}

func Render(pres *pptx.Presentation, scene Scene, opts ...RenderOption) (Stats, error)
```

`scene.Render` is **idempotent given the same scene + theme**: re-rendering
the same scene produces a byte-identical PPTX. This is a hard requirement
because pengui-slides snapshot-tests on render output.

### 10.2 Layout policy

The renderer reserves slide regions for `decoration` (foreground/background
layer), `chart` overlays, `bleed`-anchored elements, then computes a body
layout for the remaining top-level nodes. Body layout follows the scene's
top-level node order; container nodes (`two_column`, `grid`, `card_section`)
introduce sub-regions.

The layout engine is **not** a full constraint solver. It's a deterministic,
priority-ordered placement algorithm:

1. Decorations (`background` layer) — fill / corner / bleed regions.
2. Hero (if present, always first content-region item).
3. Body nodes in IR order.
4. Decorations (`foreground` layer) — overlaid last.
5. `section_divider` slide nodes — override body to full-bleed.

Each container has its own internal layout (two_column splits by ratio;
grid uses `columns` + `ratio` + `gap`; card stacks children top-to-bottom
or left-to-right per `body_layout`).

Layout is **content-bbox-driven**: a node reports its preferred bbox; the
engine assigns it a slot; the node fits to the slot. Overflow is reported
as a `LayoutWarning` (not an error). A caller that wants overflows to be
errors handles them in its own code path (treat warnings as errors when
inspecting `Stats.Warnings`) — pptx-go doesn't expose a `strict` mode.

### 10.3 Rendering policy per node

See §12 for the per-node rendering policy table. Summary: every leaf
and container node has a fixed policy — either Native PPTX shapes (the
default for most nodes) or a `pic` shape built from caller-supplied
bytes (for nodes that carry an `asset_id` field). The policy is
intrinsic to the node type, not configurable.

### 10.4 Validation

Scene validation runs in two stages:

- **Stage 1** — structural validation: the discriminated union is well-formed,
  field-level constraints hold (row-vs-header length for tables, chart-data
  shape per chart type, two_column has non-empty left/right, grid cell count
  matches `columns × ⌈cells/columns⌉`, etc).
- **Stage 2** — token + asset resolution: every token referenced by the
  scene exists in the active theme; assets referenced by `image` /
  `chart` / `decoration` (asset) / `code_block` resolve through the
  `AssetResolver`.

Stage 1 returns errors; Stage 2 returns errors for unresolved tokens/assets
and warnings for everything else (e.g. legibility hints surfaced by the
caller's logger hook).

### 10.5 Asset registry

The scene renderer's curated assets live under `scene/icons/`,
`scene/ornaments/`, `scene/frames/`. Each is a registry of named primitives
the IR references by name. The registry is **closed** (a name not in the
registry is a validation error in stage 1) but **extensible**: a caller can
register additional icons/ornaments/frames at render time via
`scene.WithIconExtension(name, svg)`.

### 10.6 AssetResolver

Assets enter the IR by reference, not by value. The IR's `image`,
`chart`, asset-kind `decoration`, and (in V1) `code_block` nodes carry
an `asset_id` field. An `AssetResolver` maps that id to bytes at render
time:

```go
type AssetResolver interface {
    // Resolve returns bytes for the asset and a content-type hint
    // (`image/png`, `image/jpeg`, `image/svg+xml`, ...).
    // A missing asset returns (nil, "", ErrAssetNotFound).
    Resolve(ctx context.Context, id AssetID) ([]byte, string, error)
}
```

The `AssetID` is a free-form string. pengui-slides uses a UUID-style
form via the `asset://` URI scheme (`asset://0fa6d3...`); pptx-go's
`AssetResolver` doesn't impose a scheme — the caller chooses. A V1
helper `scene.URIAssetResolver` accepts `asset://`-prefixed URIs and
delegates to a `func(uuid string) ([]byte, string, error)` callback
the caller supplies.

```go
stats, err := scene.Render(pres, myScene,
    scene.WithAssetResolver(myResolver))
```

Asset resolution is **lazy**: the resolver is called for each asset on
first reference per render. Resolution failures surface as
`LayoutWarning`s in `Stats.Warnings` (unless the asset was required,
in which case the render fails).

---

## 11. The scene node catalog (V1)

The V1 node catalog is the union of pengui-slides v4's IR plus a small set
of extensions. Every node has: a Go struct in `scene/nodes.go`, a validator,
a layout policy, a disposition decision (§12), and a builder composition.

### 11.1 Leaf nodes

| Node | Purpose |
|---|---|
| `hero` | Title + optional subtitle + optional eyebrow. Top of cover slides. |
| `prose` | Body paragraph(s). |
| `heading` | Section heading (`level: 1–6`). |
| `list` | Bullet / numbered / checklist. |
| `divider` | Horizontal rule with spacing. |
| `quote` | Pull quote with optional attribution. |
| `callout` | Note / warning / tip / important — colored side-bar. |
| `image` | Asset image with optional frame chrome. |
| `chip` | Inline pill (tint / solid / outline tone). |
| `arrow` | Inline directional connector with optional label. |
| `code_block` | Block-level code with optional language hint + caption. |
| `chart` | Image-shape chart (V1) / native c:chart (V2). |
| `table` | Headered tabular data with cells of RichText. |
| `flow` | Sequential step pipeline (horizontal/vertical). |
| `decoration` | Anchored ornament (asset ref or curated preset). |
| `section_divider` | Full-bleed slide break (a slide whose content is a single chapter-break shape). |

### 11.2 Container nodes

| Node | Purpose |
|---|---|
| `two_column` | 1:1 / 1:2 / 2:1 split with leaf children. |
| `grid` | 2/3/4 columns with weighted ratios, leaf children per cell. |
| `card` | Leaf-bodied accent card with fill / outline / pill header / elevation. |
| `card_section` | Top-level card that accepts grid / two_column / nested cards. |

### 11.3 pptx-go extensions (beyond pengui-slides v4)

These are reserved for the V1 catalog but land in specific later phases.
Each is a candidate "and more" capability:

- `timeline` — horizontal timeline with anchored events.
- `kpi_cards` — a small primitive for KPI grids (number + label + trend).
- `quote_card` — a pull-quote variant with avatar + attribution + accent.
- `comparison` — left/right comparison with a centered divider.
- `frame` (standalone) — a frame chrome without an inner image (e.g. an
  empty browser chrome with a label).

These are **not promised for V1.0.0**; they're documented here so callers
know the intended growth area. Specific phase plans add specific
extensions; D-NNN entries gate "in V1 vs not" decisions.

---

## 12. Per-node rendering policy

For every scene node, the library has a fixed rendering policy that's
intrinsic to the node type. There are two flavours:

- **Native** — the node renders as native PPTX shapes (text frames,
  geometries, tables, etc.) built from the IR's typed fields. The
  result is editable in PowerPoint.
- **Image (`pic` shape)** — the node renders as one `pic` shape
  positioned at the node's layout slot. The bytes come from the
  caller via an `AssetResolver` (§10.6); the IR carries an `asset_id`
  field naming the bytes. The result is not editable in PowerPoint
  (it's a raster image).

The flavour is **not configurable** — it's part of the node type's
contract. A scene IR node either has an `asset_id` field (in which
case it renders as an image) or it doesn't (in which case it renders
natively). No deck-level "make everything an image" mode and no
per-node override: if the caller wants a slide to be a single PNG,
they emit one `image` node per slide and the library does the same
thing it always does for `image` nodes.

### 12.1 The per-node policy table

| Node | Renders as | Asset field? |
|---|---|---|
| `hero` | Native (title + subtitle text shapes) | — |
| `prose` | Native (text shape) | — |
| `heading` | Native (text shape) | — |
| `list` | Native (text shape with paragraph bullet props) | — |
| `divider` | Native (`cxnSp` or thin rect) | — |
| `quote` | Native (body + attribution text shapes) | — |
| `callout` | Native (rect + icon + text shapes) | — |
| `image` | Image (`pic` shape) | `asset_id` |
| `chip` | Native (rounded rect + text shape) | — |
| `arrow` | Native (preset-arrow geometry + optional text shape) | — |
| `chart` | Image V1 / Native V2 (`c:chart`) | `asset_id` (V1) |
| `code_block` | Image — D-014 | `asset_id` |
| `table` | Native (`tbl` element) | — |
| `flow` | Native (step pills + connectors) | — |
| `decoration` (preset) | Native (preset SVG → OOXML preset/path geom) | — |
| `decoration` (asset) | Image (`pic` with bleed-aware offsets) | `asset_id` |
| `two_column` / `grid` | Layout container — renders nothing itself; children render per their own policy | — |
| `card` / `card_section` | Native (rounded rect + accent stripe + body shapes; children render per their own policy) | — |
| `frame` chrome (on image) | Native group around an `image` node's `pic` | (via inner) |
| `section_divider` | Native (full-bleed text shape with background fill) | — |

### 12.2 Mixed-policy slides are the norm

A `card_section` containing a `code_block`: the card chrome renders
natively, the inner code renders as a `pic` shape with the caller-
pre-rasterized code bytes. The slide carries native shapes for the
card chrome and one `pic` shape for the code area. Both editable and
raster content coexist freely.

### 12.3 The caller's rasterization contract

For every node with an `asset_id` field, the caller pre-rasterizes the
content and registers bytes via an `AssetID`. The `AssetResolver`
(§10.6) yields bytes; pptx-go emits the `pic` shape. **pptx-go never
invokes a rasterizer itself** — there is no HTML→PNG pipeline, no
chart renderer, no syntax-highlight-to-PNG step inside the library.
The rasterization burden lives squarely with the caller (go-slides
already has the Playwright pool, the chart renderer, and the
code-highlighter in place; pptx-go gives those outputs a home).

---

## 13. Templates, masters, and theme ingestion

### 13.1 Template ingestion

```go
brand, err := pptx.OpenStream("brand-template.pptx")
if err != nil { return err }
defer brand.Close()
theme := brand.Theme()      // extracted from theme1.xml + master color map
masters := brand.Masters()  // *Master; iterable
```

`pptx.OpenStream` opens a `.pptx` template and exposes its theme and masters
without loading slide content. The caller can use the theme alone
(`scene.Render(pres, scene, scene.WithTheme(theme))`) or use the template as
the new presentation's starting point (`pres := pptx.New(pptx.FromTemplate
(brand))`).

### 13.2 Masters & layouts

`Master` exposes layouts; layouts can be referenced by ID when adding a
slide (`pres.AddSlide(masters[0].Layout("title-content"))`). The scene
renderer maps its `LayoutKind` enum to the template's named layouts via a
caller-supplied `LayoutMap` (defaults map to PowerPoint's standard layouts).

### 13.3 Theme propagation

The theme attached to a `*Presentation` is the active theme for all
subsequent builder calls. Slide-level theme overrides are V2 (per-slide
variants are V1 — the scene `Variant` selects between named theme variants
the caller registered).

### 13.4 Brand kits

A "brand kit" is a `.pptx` template with: a populated theme1.xml, ≥1
master with layouts, optionally pre-set fonts. pptx-go can consume any
PowerPoint-emitted template that's valid OOXML. The reverse — emitting a
template that a brand designer can hand-edit in PowerPoint — is V1.x: the
RFC promises *consumption*, not authoring, of templates.

---

## 14. Curated assets (icons, ornaments, frames)

### 14.1 Icons

V1 ships a curated lucide-style allowlist. The starter set is the union of
pengui-slides v4.17's `IconNameSchema` plus a modest expansion (≈ 60 icons
total at V1.0.0). Icons live as inline SVGs under `assets/icons/<name>.svg`
and are loaded via `go:embed`.

The renderer materializes an icon as **native PPTX shapes** (path geometry
+ fill in the active accent token). This is the chief reason for keeping
the set curated: every icon goes through an SVG-to-OOXML path translator,
and that translator only handles a documented subset of SVG (single path,
solid fill, no gradients). An icon that doesn't fit the subset is a build-
time failure of the asset.

Caller-supplied icons go through `scene.WithIconExtension(name, svg)`,
which uses the same translator — same constraints apply.

### 14.2 Ornaments

V1 ships these preset ornaments (matching pengui-slides v4.16):

- `glow_ring` — halo around a focal point.
- `radial_glow` — soft gradient backdrop.
- `grid_dots` — dotted texture.
- `corner_bracket` — L-shaped bracket frame.
- `chevron_arrow` — directional accent.
- `noise_overlay` — subtle grain overlay.

Each preset is a shape recipe (PPTX shapes composed at render time) under
`assets/ornaments/<name>.go`. They render in the active accent token by
default; bleeds use negative shape offsets per OOXML.

### 14.3 Frame chrome

V1 ships these frames:

- `browser` — title bar + URL bar + traffic-light dots.
- `phone` — rounded-corner device + status bar + home indicator.
- `desktop` — monitor + stand.
- `laptop` — laptop bezel.

Each frame is a `Frame` recipe: a shape group whose interior region accepts
an image. The recipe is positioned by the renderer; the image is inserted
into the interior region. Native shapes — no rasters.

### 14.4 Extensibility

Each curated set is extensible by name registration: `scene.With...Extension
(name, recipe)`. The closed-name validation in Stage 1 picks up the extended
set. Extensions live on the `Scene`/`RenderOption`, not on global state.

---

## 15. Charts strategy

### 15.1 V1 — image-shape charts

The scene `chart` node in V1 renders as a native PPTX `pic` shape: the
caller renders the chart externally (matplotlib, ECharts headless,
chartjs-to-image, custom Go SVG, …) and provides the bytes through an
`AssetID`. The scene renderer:

- Sizes the chart region from the IR's layout (container slot or top-level
  region).
- Inserts a `pic` shape with `BlipFill` referencing the caller-provided
  bytes.
- Renders a caption text shape below the chart if `caption` is present.
- Surfaces a `LayoutWarning` if the caller-provided aspect ratio diverges
  significantly from the assigned slot (the chart fits within the slot;
  the warning lets the caller know).

This matches pengui-slides v4 today (the existing render path produces a
chart PNG and emits an image shape).

### 15.2 V2 — native `c:chart` parts

V2 adds native chart support. Phasing:

1. Wire-format work: `internal/ooxml/chart` codec for `c:chart`, `c:plotArea`,
   per-chart-type series shapes, axis definitions, data point styling.
2. Builder API: `slide.AddNativeChart(spec ChartSpec, box Box)`.
3. Scene wiring: `chart` node's disposition flips from "image" to "native"
   for chart types the codec supports; unsupported types fall back to
   image-shape.
4. Edit-in-PowerPoint validation: a native-chart deck round-trips through
   PowerPoint without losing chart editability.

V2 chart support is **not** promised at V1.0.0. The chart node's IR shape
is stable across the V1/V2 boundary; V2 changes the disposition, not the
contract.

### 15.3 Why not native in V1

Native chart XML is wide and Excel-coupled. Doing it justice is a wave-
sized investment, and it conflicts with the V1 goal "match pengui-slides
quickly". Image-shape charts are visually indistinguishable from native
charts in print/export; native charts win only when the recipient wants to
edit data in PowerPoint, which is not pengui-slides' common case.

---

## 16. Reading & round-trip

V1 commits to **round-trip of pptx-go-authored decks**:

```go
pres, _ := pptx.OpenStream("authored-by-pptx-go.pptx")
// pres.Slides()[0].Shapes()[0] is the same Shape model we wrote.
```

Every shape, text run, fill, line, table, image, master, layout, and theme
pptx-go emits is parsable back into the same Go model. This is the guarantee
the test suite enforces via per-phase round-trip golden tests.

Third-party PPTX (PowerPoint output, Keynote export, third-party libraries):
V1 is best-effort. We aim for graceful degradation — an unrecognized
extension element is ignored at parse time, a recognized one is surfaced —
but we do not promise round-trip fidelity. V2 invests in third-party
robustness.

---

## 17. Concurrency & performance

### 17.1 Concurrency model

- **Builder**: a `*Presentation` is a single-writer object; mutations must
  be serialized. Reads are concurrent-safe. The upstream's atomic-counter
  pattern for shape and rel IDs (`internal/ids`) preserves correctness if
  the caller violates this rule, but the documented contract is
  single-writer.
- **Scene renderer**: `Render` is internally parallel — slides render
  concurrently (each slide is independent in OOXML). The number of
  workers is `runtime.GOMAXPROCS(0)` by default and configurable.
- **Reusable artifacts** — themes, asset registries, master objects — are
  read-only after construction and safe for concurrent use.

### 17.2 Streaming

`OpenStream` / `SaveStream` are first-class; large decks (>50MB, hundreds of
slides) work without loading everything into memory. The upstream's
`StreamPackage`, `ConcurrentZipCollector`, and `ConcurrentStreamSave` move
under `internal/opc/` and stay first-class.

### 17.3 Benchmarks

Hot reusable artifacts carry `BenchmarkXxx` benchmarks under
`internal/...` and `pptx/` — single-slide rendering, 100-slide rendering,
1000-slide streaming render, theme resolution cost. Benchmarks are not a
CI gate (`make bench` runs on demand) but serve as the baseline against
which V1 vs V2 regressions are judged.

---

## 18. Observability hooks

The library is a library, not a service. There's no `obs/v1` protocol.
What we offer:

- **slog hook**: the builder and scene renderer accept an optional
  `*slog.Logger` via `pptx.WithLogger(l)` / `scene.WithLogger(l)`. When
  set, the library emits structured events for phase boundaries, slow
  paths (rendering taking >X ms, asset load failures, layout overflows).
  No logger = no logs (zero-cost).
- **Stats**: `scene.Render` returns a `Stats` struct: per-slide render
  time, shape counts, asset count, warnings list. Callers integrate this
  into their own telemetry.
- **Warnings**: any non-fatal issue (layout overflow, unresolved optional
  asset) is surfaced as a `LayoutWarning` in `Stats.Warnings`. The caller
  decides whether to treat them as errors.

---

## 19. Security

The library is *content-creating*. There's no network, no eval, no MCP
surface. Concerns:

- **Image bytes from disk / readers**: the library does not parse pixel
  data; it embeds bytes verbatim. A malicious image is the caller's problem
  at *display* time, not at embed time. We do verify content-type matches
  declared MIME and reject obviously malformed bytes (e.g. PNG signature
  missing).
- **Hyperlinks**: a hyperlink target is a URL string emitted verbatim into
  the OOXML relationship. We do **not** fetch or validate URLs. Callers
  are responsible for URL sanitization.
- **XML external entity / XXE**: parsers use `encoding/xml` with strict
  defaults; we do not resolve external entities.
- **Zip-slip**: PPTX is a ZIP. The streaming open path validates that
  every part path stays within the package root; absolute or `..` paths
  are rejected.
- **Memory bounds on read**: third-party PPTX with a maliciously oversized
  part fails open with a documented error (`ErrPartTooLarge`) above a
  caller-configurable limit (default 100 MB per part).

No hardcoded secrets, anywhere — including generated code and tests.

---

## 20. Forward-compatibility strategy

OOXML evolves slowly but does evolve. pptx-go's strategy:

1. **Vendor specs.** Every OOXML spec we implement against is vendored in
   `docs/specifications/<part>-<edition>-<date>.txt` (or `.pdf` excerpt
   when the public form is a PDF). A spec change is a vendored update +
   golden re-pin in the same PR.
2. **Single isolation seam.** `internal/ooxml` is the only package with
   raw XML structs. A schema bump is one PR localized to the affected
   subpackage.
3. **Single-version codecs in V1.** No multi-version branching. V2
   introduces the `protocolVersion`-keyed codec pattern *only if* a real
   compat scenario forces it (e.g. PowerPoint Online emitting a new chart
   shape).
4. **Deprecation policy.** When the upstream's API has obsolete surface,
   we ship a deprecation comment + the new API in the same V1.x. Removal
   waits until V2.

---

## 21. Compatibility with pengui-slides' IR

The scene IR is a **strict superset** of pengui-slides v4's IR. The mapping
is one-to-one for every v4 node; the type field maps to the same string.

### 21.1 Node-for-node mapping

| pengui-slides v4 IR (`SlideNode`) | scene IR (`SlideNode`) | Notes |
|---|---|---|
| `hero` | `Hero` | 1:1 |
| `prose` | `Prose` | 1:1 |
| `image` | `Image` | + optional `frame` chrome |
| `callout` | `Callout` | 1:1 |
| `heading` | `Heading` | 1:1 |
| `list` | `List` | 1:1 |
| `divider` | `Divider` | 1:1 |
| `quote` | `Quote` | 1:1 |
| `table` | `Table` | 1:1 |
| `chart` | `Chart` | V1: image; V2: native c:chart |
| `card` | `Card` | 1:1 (incl. `header_pill`, `body_layout`, `layout`, `fill`, `border_style`, `size`, `elevation`) |
| `card_section` | `CardSection` | 1:1 |
| `chip` | `Chip` | 1:1 |
| `arrow` | `Arrow` | 1:1 |
| `code_block` | `CodeBlock` | Caller-rasterized (D-014) |
| `two_column` | `TwoColumn` | 1:1 |
| `grid` | `Grid` | 1:1 (incl. weighted `ratio`) |
| `section_divider` | `SectionDivider` | 1:1 |
| `decoration` | `Decoration` | + preset names + anchor + offset + bleed anchors |
| `flow` | `Flow` | 1:1 (incl. connector kinds) |

pengui-slides v4's IR additionally carries `toc`, `bibliography`, and
`page_break` nodes for its document-authoring mode. These have no PPTX
semantics and are not part of pptx-go's scene IR (D-026). If go-slides
wants a TOC *slide*, it composes a slide with `heading` + `list` nodes.

### 21.2 Rich text + token mapping

pengui-slides v4 `TextRun` → scene `TextRun`: same shape (text + inline
style + semantic color). pengui-slides v4 `ColorRole` and `TextColorRole`
enums become `scene.ColorRole` and `scene.TextColorRole` enums (re-
exported from `pptx`).

### 21.3 "And more"

The scene IR includes a documented growth area (§11.3): `timeline`,
`kpi_cards`, `quote_card`, `comparison`, `frame`. These extensions land
in specific phases beyond the V1.0.0 minimum; go-slides can adopt them
as they become available.

### 21.4 Migration helper (out-of-tree)

pengui-slides v2 holds the IR-to-scene conversion (the v4 IR is TypeScript;
the scene IR is Go). A small JSON wire form (matching the v4 IR schema's
shape) can be defined as a stable interchange — the V1.x deliverable
`scene.UnmarshalJSON` would consume a v4-JSON scene and produce a scene
value. The wire form lives in pengui-slides' schema, not in pptx-go.

### 21.5 What does NOT cross the pptx-go boundary

pengui-slides v4's surface that pptx-go has no concept of, by
construction (D-026):

- **Soul motion / tone / narrative voice / do-don't rules** — these
  govern the LLM's authoring behavior, not the file's rendered
  output. The Theme model carries only the visual layers (`RFC §7.1`).
- **Doc-mode IR nodes** (`toc`, `bibliography`, `page_break`) — no
  PPTX semantics.
- **Layout recipes** — go-slides instantiates a recipe into IR before
  rendering; pptx-go receives the resulting IR, not the recipe.
- **Markdown source** — go-slides ingests markdown and produces IR.
  pptx-go consumes IR.
- **Validators** (Stage 1 / Stage 2 lint, diagram-legibility,
  chart-shape, density) — these are pre-render policies; pptx-go
  assumes the IR it receives is already valid.
- **HTML preview / Playwright pool** — pptx-go has no HTML path.
- **Comments, sections in the collaboration sense, editor state** —
  authoring metadata, not render input.
- **Render-mode toggles** — if a caller wants "every slide is one
  PNG", it composes one `image` node per slide. pptx-go has no mode
  switch.
- **Render-time legibility heuristics** — opinions about the output,
  not corrections to the input. Callers preprocess the IR to apply
  such heuristics.

When go-slides hands pptx-go a Theme, the motion/tone layers are
silently dropped. When go-slides has a recipe, it expands it first.
When go-slides wants image-mode output, it preprocesses the IR to a
one-image-per-slide form before calling `scene.Render`.

### 21.6 Soul→Theme mapping table

| Soul layer | pptx-go `Theme` mapping |
|---|---|
| Colors (neutrals + text + semantic) | `ColorPalette` (10 `ColorRole`s) + text colors |
| Typography (3 families × weights × sizes) | `Typography` (`TypeRole` scale); embedded via `EmbedFont` (caller-controlled) |
| Spacing (base + 8 steps) | `Spacing` (6 `SpaceRole`s; soul's 8 steps collapse where redundant) |
| Shape (radius tokens) | `Radii` (5 `RadiusRole`s) |
| Depth (shadow presets + borders) | `Elevations` (3 `ElevationRole`s) + line styling |
| Components (card/button/input/badge defaults) | Per-component defaults attached to the relevant scene-node renderer |
| Motion / tone / voice / do-don't | **Not mapped.** Authoring-only. |

A soul that names "Inter Display" and a 12-step spacing scale produces
a pptx-go Theme that lists "Inter Display" (go-slides registers a
`FontSource` and calls `EmbedFont`) and collapses the 12-step scale
into 6 `SpaceRole`s using a documented mapping (the soul's
`xxs..xxxl` → pptx-go's `xs..2xl`, dropping the extremes when they're
indistinguishable for slides). The mapping table lives in
`docs/design/SOUL-MAPPING.md`.

### 21.7 The go-slides integration contract

This subsection enumerates the boundary between go-slides (the MCP
server that holds the source-of-truth IR, the assets, the validators,
the recipes, the comments, the HTML preview) and pptx-go (the
rendering engine that emits PPTX). The boundary is the
`scene.Render(...)` call.

**go-slides passes pptx-go:**

1. A `*pptx.Theme` (constructed from the active soul, with motion/tone
   dropped).
2. A `scene.Scene` value (built from the IR, with token references
   intact and asset IDs unresolved).
3. An `AssetResolver` whose backend is go-slides' asset store.
4. Optionally, a `FontSource` whose backend is go-slides' font
   registry, plus explicit `pres.EmbedFont(...)` calls for each font
   to embed.
5. Optionally, a `*slog.Logger` for emit-side telemetry.

**go-slides keeps in-house and does NOT pass:**

- The compiled HTML representation (pptx-go has no HTML pipeline).
- The Playwright preview pool (PPTX export is text + native shapes
  + raster bytes; HTML preview is browser-rendered separately).
- The Stage-1 / Stage-2 validator suite (validation happens before
  `scene.Render` is called; pptx-go assumes the input is valid).
- The markdown-to-IR compiler.
- The comments/sections/recipes/editor state.

**pptx-go returns:**

- A serialized `.pptx` file (via `pres.Save(path)` or
  `pres.Write(w io.Writer)`).
- A `scene.Stats` struct: per-slide render time, shape counts, asset
  counts, warnings.

**Both directions are pure data.** No callbacks back into go-slides
beyond the `AssetResolver` and `FontSource` interfaces; both are read-
only in pptx-go's calling pattern (pptx-go asks for bytes by id; the
caller's implementation does whatever it does on its side).

---

## 22. Stack decisions

| Concern | Choice | Notes |
|---|---|---|
| Go version | 1.24 (pinned) | Generics, structured iteration, `slices`/`maps` stdlib utility. |
| Runtime deps | none beyond stdlib | P4 — preserve upstream's zero-dep promise. |
| Test deps | `testing`, stdlib only | Goldens via `golden.txt` files + `bytes.Equal`. |
| Build deps | `go` toolchain only | `Makefile` orchestrates; no `mage`/`task`. |
| XML | `encoding/xml` | stdlib v1. |
| JSON | `encoding/json` | stdlib v1; `encoding/json/v2` deferred. |
| Logging | `log/slog` | injected via option; no global logger. |
| Errors | `errors.Is/As`, `%w`, `errors.Join` | No panic across public boundaries. |
| Concurrency | stdlib `sync`, `sync/atomic`, `context` | No goroutine pool library. |
| Embedding | `go:embed` | Curated assets. |
| Codegen | none in V1 | If V2 needs codegen (e.g. SVG → OOXML), it'll be a `go run`-style build step, not a runtime dep. |
| Frontend | none | Library only; the published docs site (Phase 20+) is the only Web surface, and it is Markdown → static HTML. |

---

## 23. Resolved questions

These are settled and will not re-litigate without a superseding D-NNN.

- **Q1: One library or two (pptx-go and pengui-go)?** One. The scene
  renderer is a subpackage, not a separate module. (D-002)
- **Q2: Tokens or hex first?** Tokens. The literal path is an escape
  hatch, not the documented default. (D-003)
- **Q3: Native or image charts in V1?** Image. Native is V2. (D-004)
- **Q4: Ship curated icons/ornaments/frames?** Yes, embedded. Caller-
  supplied additions are first-class. (D-005)
- **Q5: Rename to hurtener?** Yes. v0.1.0 starts the new module life.
  (D-006)
- **Q6: License?** Apache-2.0 (D-007); upstream MIT is compatible.
- **Q7: Incremental refactor or greenfield?** Incremental. (D-008)
- **Q8: Doc-driven (Dockyard-style) methodology?** Yes. (D-009)
- **Q9: Code-block native or raster?** Caller-rasterized (image
  `pic` shape with bytes via `AssetResolver`). (D-014)
- **Q10: Per-node rendering policy enum?** No enum. Policy is
  intrinsic to the node type — whether the IR carries an `asset_id`
  field. (D-018)
- **Q11: Font embedding?** Mechanism provided (`FontSource` +
  `pres.EmbedFont`). No auto-embed default; caller decides. (D-019)
- **Q12: PowerPoint repair-prompt hygiene?** Always-on. Internal
  correctness pass in `internal/render/hygiene.go`. Not configurable
  — emitting OOXML that doesn't trigger the prompt is correctness,
  not preference. (D-020)
- **Q13: PPTX sections (slide grouping)?** V1. (D-021)
- **Q14: Speaker notes in V1?** Yes. (D-022)
- **Q15: Slide formats in V1?** `Slides16x9` and `Slides4x3`. Print
  formats are out of pptx-go's scope entirely (document concerns).
  (D-023)
- **Q16: AssetResolver scheme?** Free-form `AssetID` strings; helper
  for `asset://`-URI scheme. (D-024)
- **Q17: Engine vs product?** Engine. pptx-go is the IR-to-PPTX
  conversion mechanism. Product behavior (render modes, legibility
  heuristics, doc-mode constructs, validation pipelines) lives in
  callers. (D-026)

---

## 24. Out of scope

This section has two halves: things explicitly **never** in pptx-go
(product concerns, document concerns), and things **deferred to V2**
(work that pptx-go could legitimately do, but won't in V1).

### 24.1 Never in pptx-go (any version)

Concerns that don't belong in the engine, by design (D-026):

- **Render modes** (image / editable_hybrid toggles). If a caller
  wants "every slide is one PNG", it composes one `image` node per
  slide.
- **Legibility heuristics** (readable-text boosts, contrast
  remediation, density warnings). Callers preprocess the IR if they
  want them.
- **Doc-mode rendering** (TOCs, bibliographies, page breaks, print
  formats). pptx-go is slide-rendering only.
- **HTML→PNG rendering inside the library.** The library never
  invokes a rasterizer. Image-shape bytes come from the caller via
  `AssetResolver`.
- **Markdown ingestion, validation pipelines, comments, recipes,
  editor state.** All authoring-side concerns; live in callers.
- **Animations and slide-transition effects.** Visual product
  decisions; not the library's call. PowerPoint supports them and
  the library could emit them, but no V1 consumer needs them — when
  one does, an RFC supplement opens the discussion.

### 24.2 V2 backlog

Filed for explicit consideration in V2 planning. None of these is
rejected; they are simply not V1:

- Native `c:chart` parts (the V2 chart wave).
- General third-party PPTX read with round-trip fidelity.
- SmartArt-equivalent native diagrams (we ship composed shapes
  instead).
- Per-slide theme overrides.
- Multi-version codec pattern (only added when a real compat case
  forces it).
- Authoring of brand-kit templates (we consume them in V1).
- Font subsetting (V1 embeds full fonts; V1.x subsets — D-019).
- `scene.UnmarshalJSON` (consume pengui-slides v4 IR JSON directly).
  Useful enough to land in V1.x; not required for V1.0.0.

---

## Appendix A — Phase ↔ RFC section cross-reference

This is filled in as phase plans land; see `docs/plans/README.md` for the
authoritative phase index.

| Phase | Subsystem | RFC sections |
|---|---|---|
| 00 | Foundation, module rename, hygiene scaffolding | §3.3, §3.4, §5 |
| 01 | OPC + OOXML reorg | §6 |
| 02 | Theme & token model + font embedding | §7, §7.6 |
| 03 | Builder spine (Presentation, Slide, Shape, Media, Sections, Notes, Format) | §8, §8.7, §8.8 |
| 04 | Rich text model | §8.4, §9 |
| 05 | Scene package scaffold + IR catalog + AssetResolver | §10, §10.6, §11 |
| 06 | Leaf-node rendering | §11.1, §12 |
| 07 | Container nodes (two_column, grid) | §11.2 |
| 08 | Table | §8.5, §11.1 |
| 09 | Template ingestion (Theme + Masters) | §13 |
| 10 | Frame chrome | §14.3 |
| 11 | Image node + media manager refactor | §8.6, §11.1 |
| 12 | Curated icons | §14.1 |
| 13 | Curated ornaments + Decoration node | §14.2, §11.1 |
| 14 | Card + CardSection | §11.2 |
| 15 | Flow | §11.1 |
| 16 | CodeBlock (caller-rasterized) | §12 |
| 17 | Chart (caller-rasterized V1) | §15.1 |
| 18 | Round-trip read of self-authored decks | §16 |
| 19 | External-deck read robustness (best-effort) | §16 |
| 20 | Agent skills + published docs site | `CLAUDE.md §19` |
| 21 | v0.1.0 release prep | §5, §24 |

---

## Appendix B — Glossary seed

See `docs/glossary.md` for the canonical glossary. Seed entries:

- **Builder** — the `pptx` package's public API (Layer 1).
- **Scene renderer** — the `scene` package's public API (Layer 2).
- **Theme** — the active token-to-OOXML mapping at write time.
- **Token** — a semantic role (e.g. `ColorAccent`) that resolves to an
  OOXML value via the Theme.
- **Block / Leaf / Container** — scene IR node taxonomy.
- **Per-node rendering policy** — the per-node decision (native shapes
  vs caller-rasterized `pic` shape). Intrinsic to the node type
  (whether its IR carries an `asset_id`), not configurable.
- **OPC** — Open Packaging Convention; the ZIP+content-types+rels layer.
- **OOXML** — Office Open XML; the per-part XML schemas above OPC.
- **EMU** — English Metric Unit; OOXML's canonical length unit (1 inch =
  914,400 EMU).
- **Round-trip fidelity** — pptx-go-authored PPTX is parsable back into
  the same model losslessly.
- **Frame chrome** — a curated device/browser bezel rendered as native
  shapes around an image.
- **Ornament** — a curated preset decoration (glow_ring, grid_dots, …).
- **Decoration** — an IR node that places an ornament or asset image at
  an anchor (in-canvas or bleed).

---

*End of RFC-001-pptx-go.*
