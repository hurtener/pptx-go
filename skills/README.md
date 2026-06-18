# pptx-go agent skills

These are [Agent Skills](https://agentskills.io) (one `SKILL.md` per directory)
that teach an AI coding agent how to build PowerPoint decks with **pptx-go** — a
pure-Go, CGo-free PPTX authoring and reading library
(`github.com/hurtener/pptx-go`). Each skill states the relevant binding
properties, gives the exact public API, and links a runnable program under
[`examples/`](../examples) that the Phase 20 *skill smoke* compiles and runs, so
a skill cannot silently drift from the API.

## The two layers

- **`pptx`** — Layer 1, the general-purpose builder. Author slides, shapes,
  rich text, tables, images, speaker notes, sections; read decks back.
- **`scene`** — Layer 2, the IR renderer. Describe a deck as a typed `Scene`
  (nodes like `Hero`, `Card`, `Grid`, `Table`, `Flow`) and `Render` it onto a
  builder. `scene` composes `pptx`; nothing in `scene` reaches under it (P1).

Use `pptx` directly when you want imperative control; use `scene` when you want
to describe a deck declaratively.

## Binding properties every skill assumes

- **P1 — two layers, one library.** `scene` imports `pptx`, never the reverse.
- **P2 — tokens, not literals.** Color/typography/spacing/radius/elevation flow
  through a `Theme` via semantic tokens; `RGB`/`Pt` literals are an escape hatch.
  A theme swap re-renders the same input in a new visual language.
- **P3 — OOXML by isolation.** You never see raw XML in `pptx`/`scene`
  signatures; you work with Go domain types.
- **P4 — no CGo, stdlib-only runtime.** `go get` pulls zero third-party runtime
  dependencies; the artifact cross-compiles `CGO_ENABLED=0`.
- **Engine, not product (D-026).** pptx-go turns a typed scene/builder input into
  PPTX and nothing else — content choices, chart/code rasterization, font
  embedding decisions, and render policy live in *your* code. The library exposes
  mechanisms (tokens, asset resolver, font source, sections, notes); you supply
  policy.

## Skills

| Skill | Use it when you need to… |
|---|---|
| [scaffold-a-presentation](scaffold-a-presentation/SKILL.md) | Build a deck from scratch with the `pptx` builder — slides, shapes, text, tables, images, notes, sections — and read a deck back. |
| [define-a-theme](define-a-theme/SKILL.md) | Set brand colors, fonts, and spacing through the token system, and understand how a theme swap re-skins the same input. |
| [load-a-brand-template](load-a-brand-template/SKILL.md) | Start from an existing `.pptx` brand kit so new decks inherit its theme and named layouts. |
| [compose-a-scene](compose-a-scene/SKILL.md) | Describe a deck declaratively as a typed scene IR (the full node catalog) and render it. |
| [embed-a-chart-raster](embed-a-chart-raster/SKILL.md) | Put a chart in a deck by supplying a pre-rendered image (charts are image-shapes in V1). |
| [embed-a-code-block-raster](embed-a-code-block-raster/SKILL.md) | Show source code in a deck by supplying a pre-rendered image, with a language badge and caption. |
| [extend-the-icon-set](extend-the-icon-set/SKILL.md) | Register a custom SVG icon (beyond the curated set) for cards and flows, within the translator constraints. |
| [register-an-asset](register-an-asset/SKILL.md) | Supply the bytes that asset-bearing scene nodes (`Image`, `Chart`, `CodeBlock`) reference by `AssetID`. |

## Conventions

- Every example writes its output to `os.TempDir()` (or uses `WriteToBytes`) and
  prints one success line — running an example never litters the working tree.
- Skills cross-link each other in their **See also** sections; start with
  *scaffold-a-presentation* (builder) or *compose-a-scene* (IR).
