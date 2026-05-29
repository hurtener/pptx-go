# Phase 05 — Scene package scaffold + IR catalog + AssetResolver

**Subsystem:** scene (types only)
**RFC sections:** §10.1, §10.4 (Stage 1), §10.6, §11, §12
**Deps:** Phase 04 (the rich-text model the IR composes; `scene.RichText`
mirrors `pptx`).
**Status:** Done

---

## 1. Goal

Stand up the `scene` package (Layer 2): the full typed IR catalog, the
`Scene`/`Render` entrypoint (a no-op stub), Stage 1 structural validation, the
`AssetResolver` seam, and the per-node rendering-policy table — types and
contracts only, no rendering yet.

## 2. Why now

`scene` is the second public layer (P1) and Wave 2's foundation; every
rendering phase (06–17) composes this IR. Landing the catalog + validation +
asset seam first gives those phases a stable surface and lets go-slides pin its
export tool against the types (D-025) before any rendering exists. It depends on
Phase 04 because the IR's text fields are `RichText` (RFC §9).

## 3. RFC sections implemented

- `RFC §10.1` — `Scene`/`SceneSlide`/`Render`/`RenderOption`/`Stats`/`Metadata`/
  `Variant`/`LayoutKind`. `Render` is a callable no-op stub here (real rendering
  is Phase 06+).
- `RFC §10.4` — **Stage 1** validation only (well-formed unions, field
  constraints). Stage 2 (token + asset resolution) lands with rendering.
- `RFC §10.6` — `AssetID`, `AssetResolver`, `URIAssetResolver`,
  `WithAssetResolver`, `ErrAssetNotFound` (D-024).
- `RFC §11` — the full node catalog: 16 leaf nodes + 4 container nodes as Go
  structs implementing a sealed `SlideNode` union.
- `RFC §12` — the per-node rendering-policy table as a documented, test-asserted
  map (D-011/D-018); intrinsic, not configurable.

The `scene/icons|ornaments|frames` registries (§10.5) and the layout engine
(§10.2) are later phases; `scene/layout/` lands as a placeholder.

## 4. Brief findings incorporated

No informing brief — `docs/research/INDEX.md` lists scene briefs (layout-engine
survey, scene-IR JSON wire form) only as *candidates*; the IR and contracts are
specified directly in RFC §10–§12 and D-011/D-018/D-024/D-026. (Same posture as
Phases 03–04.)

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-011` / `D-018` — per-node rendering policy is intrinsic to the node type
  and keyed on whether the node's IR carries an `asset_id` field; no
  `Disposition` enum, no deck-wide mode. `scene/policy.go` encodes this as a
  documented table + a compile-/test-time assertion.
- `D-024` — `AssetID` is free-form; `URIAssetResolver` is the `asset://` helper.
- `D-026` — engine, not product: the IR carries no legibility/mode opinions;
  `Render` composes only.
- `D-025` — the go-slides contract (pure data both ways); this phase fixes the
  `Scene` types that contract pins against.

## 7. Architecture

```text
scene/
  scene.go      Scene, SceneSlide, Metadata, Variant, LayoutKind,
                Render (no-op stub), RenderOption, Stats, LayoutWarning
  nodes.go      SlideNode (sealed union) + 16 leaf + 4 container structs
  richtext.go   RichText ([]TextRun), TextRun, RunStyle, TextColor (token/literal)
  tokens.go     type-alias re-exports of pptx token enums + const re-exports
  validate.go   Stage 1: ValidateScene → structural errors
  asset.go      AssetID, AssetResolver, URIAssetResolver, ErrAssetNotFound,
                WithAssetResolver
  policy.go     per-node policy table (Native | Image, asset?) + Policy(kind)
  doc.go        package doc
  layout/doc.go placeholder for the layout engine (later phases)
```

**The union.** `SlideNode` is a sealed interface — an unexported marker method
plus `NodeKind() NodeKind` — implemented by each node struct. Go's type system
discriminates; validation and (later) rendering type-switch on the concrete
type. This is the idiomatic-Go equivalent of the RFC's "discriminated `type`
field" and needs no custom JSON union decoding (go-slides passes a typed
`Scene`, D-025; the JSON wire form is a deferred research candidate).

**Asset policy.** Exactly the nodes that render as an image carry an `AssetID`
field: `Image`, `Chart`, `CodeBlock`, and asset-kind `Decoration`. `policy.go`
maps every `NodeKind` to `{Flavour, HasAsset}`; `policy_test.go` reflects over
each node struct and asserts an `AssetID` field is present **iff** the table
says `HasAsset` — the §12.1 table and the structs can't drift.

**Validation.** `ValidateScene` walks slides and nodes; Stage 1 checks
structural invariants (e.g. `two_column` left/right non-empty, `grid` cell count
vs `columns`, `table` row width vs header width, `heading.Level ∈ 1..6`, a
closed enum value is in range). Returns a joined error (`errors.Join`).

## 8. Files added or changed

```text
scene/scene.go        # NEW — entrypoint types + Render stub
scene/nodes.go        # NEW — IR catalog (sealed SlideNode union)
scene/richtext.go     # NEW — RichText/TextRun/RunStyle/TextColor
scene/tokens.go       # NEW — pptx token re-exports
scene/validate.go     # NEW — Stage 1 validation
scene/asset.go        # NEW — AssetID/AssetResolver/URIAssetResolver
scene/policy.go        # NEW — per-node policy table
scene/doc.go          # NEW — package doc
scene/layout/doc.go   # NEW — placeholder
scene/*_test.go       # NEW — catalog/validation/asset/policy tests
docs/glossary.md      # CHANGED — SlideNode, RichText (scene), NodeKind, the node names
scripts/smoke/phase-05.sh  # NEW
internal/coveragecheck/coverage.json  # CHANGED — band the scene package
```

No raster/registry/layout code (later phases). User-facing docs site/skills are
Phase 20 (inert now).

## 9. Public API surface

```go
// scene
func Render(pres *pptx.Presentation, s Scene, opts ...RenderOption) (Stats, error)
type Scene struct { Theme *pptx.Theme; Slides []SceneSlide; Meta Metadata }
type SceneSlide struct { ID string; Layout LayoutKind; Nodes []SlideNode; Notes RichText; Variant Variant }
type Stats struct { Slides int; Shapes int; Assets int; Warnings []LayoutWarning; … }
type RenderOption func(*renderConfig)
func WithAssetResolver(r AssetResolver) RenderOption
func WithLogger(*slog.Logger) RenderOption

type SlideNode interface { NodeKind() NodeKind; /* sealed */ }
// leaves: Hero, Prose, Heading, List, Divider, Quote, Callout, Image, Chip,
//   Arrow, CodeBlock, Chart, Table, Flow, Decoration, SectionDivider
// containers: TwoColumn, Grid, Card, CardSection

type AssetID string
type AssetResolver interface { Resolve(ctx context.Context, id AssetID) ([]byte, string, error) }
func URIAssetResolver(fn func(uuid string) ([]byte, string, error)) AssetResolver
var ErrAssetNotFound error

func ValidateScene(s Scene) error           // Stage 1
type Policy struct { Image bool; HasAsset bool }
func PolicyFor(k NodeKind) Policy

// richtext
type RichText []TextRun
type TextRun struct { Text string; Style RunStyle; Color TextColor }
func LiteralColor(hex string) TextColor
```

`scene`'s token enums are **type aliases** of `pptx`'s (`type ColorRole =
pptx.ColorRole`, etc.) with const re-exports, so callers use one vocabulary.

## 10. Risks

- **R1 — union modeling.** A wrong union shape churns every later phase.
  *Mitigation:* sealed interface + `NodeKind()` (idiomatic, type-switchable);
  no JSON union decoding committed (deferred). Pin with a per-node fixture.
- **R2 — policy/struct drift.** The §12.1 table and the structs could diverge.
  *Mitigation:* `policy_test.go` reflects over each struct and asserts the
  `AssetID`-field ⇔ `HasAsset` invariant — a build-failing check.
- **R3 — over-specifying node fields now.** Fields the renderer needs aren't
  fully known until Phases 06–17. *Mitigation:* fields follow RFC §11 purposes +
  the §12 policy; later phases extend a struct (additive) without reshaping the
  union. Document any field added later as a plan deviation in that phase.

## 11. Acceptance criteria

1. The full IR catalog compiles; `scene` imports `pptx` and not vice-versa (P1).
2. `ValidateScene` accepts a valid fixture per node type and rejects negatives
   (malformed union/field constraints).
3. `Render` is callable and returns a zero `Stats` (nil error) on an empty
   `Scene`.
4. `URIAssetResolver` resolves an `asset://<uuid>` id to bytes via the caller's
   callback, and returns `ErrAssetNotFound` for a miss.
5. `policy_test.go` asserts every node whose struct carries an `AssetID` field
   matches the §12.1 policy table (and vice-versa).
6. `make build`/`test`/`lint`/`coverage`/`preflight`/`check-mirror` green; prior
   smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default for a new scene package |

`scene` is added to `internal/coveragecheck/coverage.json`. Tests are
co-located (`scene/*_test.go`) so self-coverage is attributed.

## 13. Smoke check

`scripts/smoke/phase-05.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `pptx` does not import `scene` (P1) — grep guard.
3. `OK:` IR-catalog + validation tests pass.
4. `OK:` Render-stub-on-empty-scene test passes.
5. `OK:` URIAssetResolver test passes.
6. `OK:` policy/struct assertion test passes.

`SKIP` each until its surface lands; `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` — a fixture per node type (valid) + negatives for Stage 1;
  Render-stub; URIAssetResolver hit/miss; policy reflection assertion.
- **Round-trip golden:** n/a (no emission yet — Phase 06+).
- **Integration:** n/a yet (scene renders nothing); the first scene→pptx
  integration test lands with Phase 06.
- **Fuzz/Bench:** none.

## 15. Vocabulary added

- `SlideNode` — the sealed scene IR union (leaf + container nodes).
- `NodeKind` — the discriminator for a `SlideNode`.
- `Policy` — a node's intrinsic rendering policy (Native vs Image).
- The node names (`Hero`, `Prose`, `Callout`, `TwoColumn`, `CardSection`, …) —
  catalog entries; the catalog is documented in the glossary as a cluster.

(Most container/leaf concepts and `Scene`/`SceneSlide`/`AssetResolver`/`LayoutKind`/
`Variant`/`Stats`/`LayoutWarning` already have glossary entries.)

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages (scene banded at 80%, 84.6% actual).
- [x] `scripts/smoke/phase-05.sh` reports `OK ≥ 6` and `FAIL = 0` (6 OK).
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (SlideNode, NodeKind, Policy).
- [x] Decision entries added (none new — D-011/D-018/D-024 suffice).
