# Phase 12 — curated icons

**Subsystem:** assets/icons + scene/icons + internal/render (SVG→OOXML) +
internal/ooxml/slide (custom geometry) + pptx (builder `AddIcon`)
**RFC sections:** §14.1, §14.4
**Deps:** Phase 09 (template/theme), Phase 10 (the curated-asset extension-seam
pattern this mirrors).
**Status:** Done

> **Note (drift):** the master plan's "preset/path geom translator" is
> implemented as a `custGeom` (path) translator; the ≈60-icon target (D-005) is
> delivered as a ~16-icon engine-complete starter set with the full set deferred
> to a content follow-up (D-040). Icon *placement* is Phases 14–15 (no icon IR
> node exists; D-005).

---

## 1. Goal

A curated icon SVG renders as a native PPTX path shape filled with the accent
token, and a caller can register additional icons by name via
`scene.WithIconExtension`, with an SVG that violates the translator constraints
failing at registration.

## 2. Why now

The master plan (Wave 4) opens with curated icons. Icons are the first curated
set that needs **custom path geometry** (frames and ornaments use preset
shapes), so Phase 12 builds the engine the visual system depends on: the
`custGeom` wire types, the SVG→OOXML translator, the builder `AddIcon` API, and
the icon registry mirroring the Phase-10 frames seam. Later phases consume it —
icons are referenced by name from `card` (Phase 14), `flow` steps (Phase 15),
and `header_pill` (D-005, glossary `Icon`); there is **no standalone icon IR
node**.

## 3. RFC sections implemented

- `RFC §14.1` — **Icons.** Curated lucide-*style* SVGs rendered as native
  `custGeom` path shapes via the SVG→OOXML translator (single path, solid fill,
  no gradients). Ships the **engine + a ~16-icon starter set**; the full ≈60 is
  a tracked content follow-up (D-040).
- `RFC §14.4` — **Extensibility.** `scene.WithIconExtension(name, svg)`; the
  closed curated set is a Stage-1 validation error when a name is missing, and a
  caller SVG that violates the translator fails at registration.

## 4. Brief findings incorporated

- `docs/research/03-svg-path-to-ooxml-translator.md` —
  - **F1** (lucide raw data unusable) → the curated set is hand-authored filled
    single paths, not lucide stroke data.
  - **F2** (`custGeom` is the general target) → every icon → one `custGeom` /
    one `path`; no preset-shape mapping.
  - **F3** (viewBox→path coords, no y-flip) → path `w/h` = viewBox ×100,
    integer points, shared top-left/y-down origin.
  - **F4** (subset `M L H V C Q Z` + relative + `S`/`T`; no arcs) → the
    translator expands `S`/`T` to `C`/`Q` and rejects `A`; curves are Béziers.
  - **F5** (reject at registration) → curated SVGs validated by a build-time
    test; caller SVGs validated when `WithIconExtension` is applied (Stage-1,
    before compose).
  - **F6** (fill is the token) → the SVG color is discarded; the path fills with
    the accent token (default) or a caller `WithFill`.
  - **F7** (layering) → wire types in `internal/ooxml/slide`, translator in
    `internal/render`, `pptx.AddIcon(svg []byte, …)` (P3-safe), registry in
    `scene/icons`; `scene` never imports `internal/...` — it validates via
    `pptx.ValidateIcon` (P1).
  - **F8** (per-command `XMLName`) → ordered heterogeneous path commands marshal
    via a per-command `xml.Name`, with new element locals added to `elementNS`.

## 5. Findings I'm departing from

None from the brief. The one deliberate scope call (starter set vs. the full
≈60) is recorded as **D-040**, not a brief departure (the brief itself
recommends the starter-set-now approach in §5).

## 6. Decisions referenced

- `D-005` — *Curated assets via go:embed; single path / solid fill / no
  gradients; closed set; fail at registration.* The governing decision; Phase 12
  implements it.
- **D-040 (NEW)** — *Phase 12 ships the icon engine + a ~16 starter set (≈60
  deferred); translator excludes elliptical arcs; `AddIcon` takes SVG bytes.*
  Refines D-005's "≈60" to a starter set now (engine is the hard part; each
  further icon is one validated SVG), documents the no-arc translator
  constraint, and fixes the P3-safe builder signature.
- `D-038` — *Curated-asset extension seam (frames).* Icons mirror the
  `Curated`/`With`/`Lookup`/`Names` registry + per-render overlay, with the
  difference that an icon extension is **validated at registration** (D-005),
  not merely name-checked at render.
- `D-035` — *Byte-identical idempotency.* The translator is pure integer math;
  the per-render icon registry is read-only during compose.
- `D-032` — *Self-authored slides round-trip.* The new `custGeom` wire types
  ship a write→read round-trip test.

## 7. Architecture

```text
assets/icons/*.svg + icons.go (//go:embed)      ← first go:embed in the repo (D-005)
        │  name → SVG bytes
        ▼
scene/icons/registry.go  (Registry: Curated/With/Lookup/Names)   ← mirrors scene/frames
        │  scene.WithIconExtension(name, svg) → validated via pptx.ValidateIcon (P1)
        ▼
pptx.ValidateIcon(svg) / Slide.AddIcon(svg, box, opts)  ← public, SVG-in, *Shape-out (P3-safe)
        │
        ▼
internal/render/svgpath.go: Translate(svg) (*slide.XCustomGeometry, error)   ← parse + validate + emit
        │
        ▼
internal/ooxml/slide: XCustomGeometry / XPath / XPathCommand / XPoint   ← new custGeom wire types
        + restorenamespaces.go: pathLst/path/moveTo/lnTo/cubicBezTo/quadBezTo/close/pt → "a"
```

Import graph (no cycle, P1/P3 intact): `internal/ooxml` ← `internal/render` ←
`pptx` ← {`scene`, `scene/icons`}; `scene/icons` ← `assets/icons`; `scene` and
`scene/icons` never import `internal/...` (they validate through
`pptx.ValidateIcon`).

## 8. Files added or changed

```text
internal/ooxml/slide/geometry.go        # NEW — XCustomGeometry, XPath, XPathCommand, XPoint (+ marshal)
internal/ooxml/slide/slide_types.go     # CHANGED — XShapeProperties gains CustomGeom *XCustomGeometry
internal/ooxml/restorenamespaces.go     # CHANGED — pathLst/path/moveTo/lnTo/cubicBezTo/quadBezTo/close/pt → "a"
internal/ooxml/slide/geometry_test.go   # NEW — custGeom marshal + round-trip
internal/render/svgpath.go              # NEW — SVG single-path → XCustomGeometry translator + validation
internal/render/svgpath_test.go         # NEW — subset coverage, S/T expansion, rejection cases, determinism
pptx/icon.go                            # NEW — Slide.AddIcon, ValidateIcon
pptx/icon_test.go                       # NEW — AddIcon emits custGeom + accent fill; round-trip golden
pptx/slide_builder.go                   # CHANGED — AddCustomShape (custGeom auto-shape)
assets/icons/*.svg                      # NEW — ~16 curated single-path filled icons
assets/icons/icons.go                   # NEW — //go:embed FS + Names()/Read(name)
assets/icons/icons_test.go              # NEW — every embedded icon translates (build-time validity)
scene/icons/registry.go                 # NEW — Registry: Curated/With/Lookup/Names
scene/icons/registry_test.go            # NEW — curated lookup, overlay immutability, nil safety
scene/scene.go                          # CHANGED — WithIconExtension, ValidateIcon, registration-time validation
scene/icons_validate_test.go            # NEW — WithIconExtension valid/invalid, ValidateIcon
internal/coveragecheck/coverage.json    # CHANGED — assets/icons, scene/icons bands
scripts/drift-audit.sh                  # CHANGED — P3 allowlist adds internal/render (SVG input parse, D-040)
scripts/smoke/phase-12.sh               # NEW — phase smoke
docs/research/03-svg-path-to-ooxml-translator.md  # NEW — informing brief
docs/research/INDEX.md                   # CHANGED — lists brief 03
docs/decisions.md                        # CHANGED — adds D-040
docs/glossary.md                         # CHANGED — Icon (updated), custGeom, SVG translator
docs/plans/phase-12-curated-icons.md     # NEW — this plan
```

No `docs/design/THEME.md` change (icons reuse the accent color token — no new
token). No user-facing doc-site / skills (inert pre-Phase 20).

## 9. Public API surface

```go
// pptx
//
// AddIcon adds a single-path SVG glyph as a native custom-geometry shape,
// positioned by box and filled with the accent token by default (override with
// WithFill). It errors if the SVG violates the translator constraints (single
// path, solid fill, no gradients, no elliptical arcs).
func (s *Slide) AddIcon(svg []byte, box Box, opts ...ShapeOption) (*Shape, error)

// ValidateIcon reports whether svg satisfies the icon translator constraints,
// without drawing it — the registration-time check (D-005).
func ValidateIcon(svg []byte) error

// scene
//
// WithIconExtension registers a caller icon under name for this render (RFC
// §14.4). The SVG is validated when the option is applied; an SVG that violates
// the translator constraints fails the render with a Stage-1 error (not at
// compose). Registering a curated name overrides it for this render only.
func WithIconExtension(name string, svg []byte) RenderOption

// ValidateIcon re-exports pptx.ValidateIcon so callers validate at their own
// registration point (P1 — scene never reaches under pptx).
func ValidateIcon(svg []byte) error

// scene/icons
type Registry struct{ /* name → SVG bytes */ }
func Curated() *Registry                                  // the embedded starter set
func (r *Registry) With(name string, svg []byte) *Registry
func (r *Registry) Lookup(name string) ([]byte, bool)
func (r *Registry) Names() []string

// assets/icons
func Names() []string                 // sorted curated icon names
func Read(name string) ([]byte, bool) // embedded SVG bytes
```

No prior public surface breaks (all additive).

## 10. Risks

- **R1 — PowerPoint "repair" on `custGeom`.** A malformed path list makes
  PowerPoint reject the deck. **Mitigation:** the wire types ship a round-trip
  test; `AddIcon` output goes through the conformance gate + `validate-schema.sh`
  (preflight); and the user runs an icon deck through PowerPoint / Quick Look
  (the §10.3 eyeball — only real oracle for "does it render").
- **R2 — Translator gaps.** An unsupported `d` command would silently mis-render.
  **Mitigation:** the translator **rejects** anything outside the documented
  subset (fails, never approximates silently); the curated build-time test
  proves every shipped icon translates; the rejection cases are unit-tested.
- **R3 — Determinism.** Float coordinate math could vary. **Mitigation:**
  coordinates round to integers deterministically; the translator is pure (no
  map iteration, no wall-clock); an icon deck re-renders byte-identically
  (asserted).
- **R4 — Scope (≈60 vs ~16).** Shipping fewer icons than D-005's target.
  **Mitigation:** D-040 records the deliberate split; the engine is complete and
  each further icon is one validated SVG; the gap is logged (no silent
  truncation), `Names()` reports exactly what ships.

## 11. Acceptance criteria

1. Each curated icon renders via `Slide.AddIcon` as a native `<a:custGeom>` path
   shape (not a `pic`), filled with the accent token.
2. A `custGeom` shape round-trips losslessly through `pptx.NewFromBytes`
   (write→read model equality at the wire level).
3. `scene.WithIconExtension(name, validSVG)` registers and renders the icon;
   `scene.ValidateIcon` accepts a valid icon.
4. An SVG that violates the translator constraints (multi-path, gradient/`none`
   fill, an elliptical-arc command, unparseable `d`) fails at **registration** —
   `pptx.ValidateIcon` errors, and a `WithIconExtension` of it makes `Render`
   return a Stage-1 error (no slide composes).
5. Every embedded curated SVG translates successfully (build-time asset
   validity).
6. An icon deck re-renders **byte-identically** (D-035); `make test -race` and
   `make coverage` pass for touched packages.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `internal/render` | 80% | internal/render band (unchanged) |
| `internal/ooxml/slide` | per existing band | new custGeom code covered by geometry_test |
| `pptx` | per existing band | AddIcon/ValidateIcon covered by icon_test |
| `scene` | 80% | scene band (unchanged) |
| `scene/icons` | 80% | new scene package |
| `assets/icons` | 80% | new curated-asset package (embed + Names/Read covered) |

New packages (`scene/icons`, `assets/icons`) get `coverage.json` entries in this
PR.

## 13. Smoke check

`scripts/smoke/phase-12.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` `AddIcon` emits a `custGeom` path shape with accent fill (`pptx` test).
3. `OK:` `custGeom` shape round-trips (`pptx`/`internal/ooxml/slide` test).
4. `OK:` `WithIconExtension` registers + renders a valid icon (`scene` test).
5. `OK:` an invalid SVG fails at registration (`pptx`/`scene` test).
6. `OK:` every embedded curated icon translates (`assets/icons` test).
7. `OK:` icon deck round-trip + byte-identical (`pptx`/integration test).

## 14. Tests

- **Unit:** translator subset + rejection + determinism (`internal/render`);
  custGeom marshal + round-trip (`internal/ooxml/slide`); `AddIcon`/`ValidateIcon`
  (`pptx`); registry overlay/immutability/nil (`scene/icons`); `WithIconExtension`
  valid/invalid (`scene`); every embedded icon translates (`assets/icons`).
- **Round-trip golden:** yes — an `AddIcon` shape round-trips (new builder API).
- **Integration:** the builder seam is exercised end-to-end by
  `test/pptx/icon_test.go` with real drivers (real `internal/opc` write,
  `encoding/xml` decode, the conformance gate, byte-identical round-trip) — the
  cross-package path `pptx → internal/render → internal/ooxml`. A scene→icon
  *placement* integration arrives with the first consuming node (Phase 14); no
  IR node places an icon in Phase 12.
- **Fuzz:** a `FuzzTranslate` over the SVG parser (a new parse/decode surface in
  `internal/render`) — seed corpus of valid + malformed `d` strings, invariant:
  never panics, returns an error or a well-formed geometry.
- **Benchmark:** none required.

## 15. Vocabulary added

Filed in `docs/glossary.md`:

- `custGeom` — OOXML custom path geometry (`a:custGeom`), the wire form an icon
  renders to.
- `SVG translator` — `internal/render`'s SVG-single-path → `custGeom` converter
  with the documented constraint subset.
- `Icon` (updated) — note the starter set + the translator path.

## 16. Plan deviations encountered during implementation

- **`internal/render` imports `encoding/xml` → P3 drift-audit allowlist
  extended.** The translator parses the SVG *input* with `encoding/xml`; the
  P3 check confined that import to `{ooxml, opc, conformance}`. Since
  `internal/render` defines/exposes no OOXML wire types (it produces
  `internal/ooxml` structs and nothing above the wall touches XML), the check's
  allowlist now adds `render`, with the rationale in `drift-audit.sh` and D-040.
  Intent-preserving, not a P3 weakening.
- **The assembled per-render icon registry is not stored on `renderConfig`
  yet.** Phase 12 has no IR node that *places* an icon (placement is Phases
  14–15), so storing `cfg.icons` would be a write-only field. Phase 12 instead
  validates icon extensions at registration (the testable deliverable) and ships
  the `scene/icons` registry package (unit-tested, `Curated/With/Lookup/Names`)
  for the consuming nodes to assemble. The registry-into-`renderConfig` wiring
  lands with its first consumer (Phase 14).
- **Icon builder tests live in `test/pptx/icon_test.go`** (alongside the image
  tests + zip helpers), not `pptx/icon_test.go` as sketched in §8 — consistent
  with where `media_test.go` lives.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages (render 86.1%, scene/icons
      100%, assets/icons 91.7%; gate 0 failing).
- [x] `scripts/smoke/phase-12.sh` reports `OK ≥ 7` and `FAIL = 0` (7 OK, 0 FAIL).
- [x] Prior phases' smoke scripts still pass (preflight PASS).
- [x] Glossary updated (Icon, custGeom, SVG translator).
- [x] Decision entries added (D-040).
- [x] (Phase 20+) Docs site / skills — N/A (inert pre-Phase 20).
