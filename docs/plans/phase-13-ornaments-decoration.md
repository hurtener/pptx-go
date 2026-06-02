# Phase 13 — curated ornaments + Decoration node

**Subsystem:** assets/ornaments + scene (decoration) + pptx (builder
fills/transform) + internal/ooxml/slide (gradient wire types)
**RFC sections:** §14.2, §11.1 (decoration), §10.2 (layout order), §12 (policy)
**Deps:** Phase 12 (the curated-asset extension-seam pattern; the icon
custom-geometry precedent).
**Status:** Done

> **Delivery (split, confirmed with maintainer).** This one plan covers two
> PRs:
> - **PR #1 — builder foundations + carried-forward fixes.** Gradient fills
>   (linear + radial), `WithRotation`, token-alpha; plus two fixes carried from
>   the Phase-12 wiring audit: `Scene.Meta` → `core.xml`, and `pptx.WithLogger`.
>   These are general builder primitives + orthogonal fixes; they ship first.
> - **PR #2 — ornaments + Decoration node.** The six curated ornament recipes,
>   `scene/ornaments` registry, the `Decoration` IR expansion, and
>   `render_decoration.go` with layer z-order. Builds on PR #1's primitives.

---

## 1. Goal

A scene `Decoration` node places a curated ornament (or a caller asset) at an
anchored, optionally-bled position with caller opacity/rotation/size, layered
behind or above the body — and the builder gains the gradient/rotation/alpha
primitives the ornaments (and callers) need.

## 2. Why now

The master plan (Wave 4) sequences ornaments after icons. Ornaments are the last
curated set and the first to need **gradient fills** and **shape rotation** —
builder primitives the Phase-02 block named (`GradientFill`) but never built.
The Phase-12 wiring audit also surfaced two builder gaps (deck metadata silently
dropped; `pptx.WithLogger` promised by RFC §18 but absent); the maintainer asked
to fold them into Phase 13. Landing the builder foundations + fixes first (PR #1)
unblocks the ornament recipes (PR #2) and clears the audit debt.

## 3. RFC sections implemented

- `RFC §14.2` — **Ornaments.** The six presets as native shape recipes in the
  accent token; bleeds via negative offsets. (PR #2; glows use PR #1 gradients.)
- `RFC §11.1` (decoration) — the `decoration` node's full composition. (PR #2.)
- `RFC §10.2` — **layout order**: background decorations → body → foreground
  decorations → section_divider. (PR #2.)
- `RFC §12` — **policy**: preset decoration = native shapes; asset decoration =
  `pic` with bleed-aware offsets. (PR #2.)
- `RFC §8.2/§8.3` (fills, transform) — gradient fills + rotation as public
  builder API. (PR #1.)
- `RFC §18` — **observability**: `pptx.WithLogger` (PR #1; the scene half landed
  in the Phase-12 audit).
- `RFC §8.x` (core metadata) — `Scene.Meta` → `docProps/core.xml`. (PR #1.)

## 4. Brief findings incorporated

- `docs/research/04-preset-ornament-recipes.md` —
  - **F1** (real gradients) → add `XGradientFill` + `pptx.LinearGradient`/
    `RadialGradient`; glows use a `path="circle"` accent→accent@alpha0 gradient.
  - **F2** (rotation) → `WithRotation(deg)`; multi-shape unit rotation deferred
    (no group transform — V2), per-shape rotation is V1.
  - **F3** (token-alpha) → `TokenColorAlpha(role, alpha)` (P2-preserving).
  - **F4** (six recipes) → rect/ellipse/roundRect/chevron + the new primitives;
    `noise_overlay` is a deterministic sparse dot scatter (documented).
  - **F5** (Decoration IR) → add Offset, Size, Bleed, Opacity, Rotation.
  - **F6** (layer z-order) → the renderer splits decorations out of the body
    stack by `Layer`, mirroring `SectionDivider`'s special-casing.
  - **F7** (preset vs asset) → native recipe vs `pic` (Phase-11 image path).

## 5. Findings I'm departing from

None. The two brief open questions (group-shape rotation; true grain) are
resolved by *deferral*, not departure (documented in §10 and D-041).

## 6. Decisions referenced

- **D-041 (NEW)** — *V1 ships gradient fills (linear + radial); ornament glows
  use them; rotation + token-alpha land alongside.* Reverses the
  "gradients deferred" note in `pptx/fill.go`; closes the Phase-02 `GradientFill`
  gap. Multi-shape group rotation stays V2 (no builder group transform).
- **D-042 (NEW)** — *Phase 13 absorbs two carried-forward builder fixes and
  splits delivery.* `Scene.Meta` → `core.xml` (a deterministic
  `Presentation.SetMetadata`) and `pptx.WithLogger` (builder event emission,
  RFC §18) land in PR #1 with the builder primitives; ornaments + Decoration land
  in PR #2. Rationale: the fixes are builder-layer and orthogonal to ornaments;
  shipping them first keeps each review focused (the Phase-12 audit flagged both).
- `D-005` — curated assets via go:embed/recipes; closed set + extension; fail at
  registration. Ornaments mirror frames (Go recipes, not SVG).
- `D-038` — curated-asset extension seam (frames) → `scene/ornaments` registry.
- `D-035` — byte-identical idempotency: gradients, dot scatters, and core.xml are
  deterministic (no timestamps, no map iteration, no wall-clock).
- `D-026` — engine, not product: ornaments are mechanisms; `noise_overlay` is an
  approximation, not a pixel-perfect grain.

## 7. Architecture

```text
PR #1 — builder foundations + carried fixes
  internal/ooxml/slide/gradient.go  XGradientFill{ GsLst, Lin|Path+FillToRect }  (+ XShapeProperties.GradientFill)
  pptx/fill.go                      LinearGradient(angle, stops…) / RadialGradient(stops…) / GradientStop
  pptx/color.go                     TokenColorAlpha(role, alpha)  (token + alpha; P2)
  pptx/shape.go                     WithRotation(deg)  → XTransform2D.Rotation (deg×60000)
  pptx/metadata.go                  Presentation.SetMetadata(Metadata) → regen docProps/core.xml (deterministic)
  pptx/options.go + presentation.go pptx.WithLogger(l) ; emit at save (RFC §18)
  scene/scene.go                    Render writes s.Meta via pres.SetMetadata

PR #2 — ornaments + Decoration
  assets/ornaments/*.go             glow_ring/radial_glow/grid_dots/corner_bracket/chevron_arrow/noise_overlay
  scene/ornaments/registry.go       Recipe + Curated/With/Lookup/Names (mirrors scene/frames)
  scene/nodes.go                    Decoration += Offset, Size, Bleed, Opacity, Rotation
  scene/render_decoration.go        anchor+offset+size+bleed → box; opacity→token alpha; preset→recipe / asset→pic
  scene/render.go                   layout() splits decorations by Layer (bg before body, fg after — RFC §10.2)
  scene/scene.go                    WithOrnamentExtension(name, recipe)
```

Import graph: `internal/ooxml` ← `pptx` ← {`scene`, `scene/ornaments`};
`scene/ornaments` ← `assets/ornaments` ← `pptx`. No cycle; P1/P3 intact.

## 8. Files added or changed

```text
# PR #1
internal/ooxml/slide/gradient.go        # NEW — XGradientFill, XGradientStop, XLinearGradient, XPathGradient, XFillToRect
internal/ooxml/slide/slide_types.go     # CHANGED — XShapeProperties.GradientFill
internal/ooxml/restorenamespaces.go     # CHANGED — fillToRect → a
internal/ooxml/slide/gradient_test.go   # NEW — gradient marshal + round-trip
pptx/fill.go                            # CHANGED — LinearGradient/RadialGradient/GradientStop
pptx/color.go                           # CHANGED — TokenColorAlpha
pptx/shape.go                           # CHANGED — WithRotation
pptx/metadata.go                        # NEW — Presentation.SetMetadata + deterministic core.xml
pptx/options.go                         # CHANGED — WithLogger Option
pptx/presentation.go                    # CHANGED — logger field; emit at WriteToBytes/Save
pptx/*_test.go (test/pptx)              # NEW — gradient/rotation/alpha round-trip; SetMetadata round-trip; logger emits
scene/scene.go                          # CHANGED — Render calls pres.SetMetadata(s.Meta)
scene/observability_test.go             # CHANGED — Meta-in-core.xml end-to-end
docs/design/THEME.md                    # CHANGED — gradient/rotation are builder mechanisms (note; no new token role)
# PR #2
assets/ornaments/*.go + ornaments.go    # NEW — six recipes
scene/ornaments/registry.go (+test)     # NEW — registry
scene/nodes.go                          # CHANGED — Decoration fields
scene/render_decoration.go (+test)      # NEW — Decoration composition + z-order
scene/render.go                         # CHANGED — layer ordering; dispatch Decoration
scene/scene.go                          # CHANGED — WithOrnamentExtension
# both
internal/coveragecheck/coverage.json    # CHANGED — assets/ornaments, scene/ornaments bands
scripts/smoke/phase-13.sh               # NEW — phase smoke (grows across the two PRs)
docs/research/04-…, INDEX.md            # NEW/CHANGED — brief
docs/decisions.md                       # CHANGED — D-041, D-042
docs/glossary.md                        # CHANGED — Ornament, Bleed, Gradient fill, Rotation
docs/plans/phase-13-ornaments-decoration.md  # NEW — this plan
```

## 9. Public API surface (highlights)

```go
// pptx — fills & transform (PR #1)
type GradientStop struct{ Pos float64; Color Color } // Pos 0..1
func LinearGradient(angleDeg float64, stops ...GradientStop) Fill
func RadialGradient(stops ...GradientStop) Fill       // path="circle"
func TokenColorAlpha(role ColorRole, alpha int) Color // token + 0..100000 alpha
func WithRotation(deg float64) ShapeOption            // applies to AddShape

// pptx — core metadata (PR #1)
func (p *Presentation) SetMetadata(m Metadata)        // deterministic core.xml
type Metadata struct{ Title, Author, Subject string }

// pptx — observability (PR #1)
func WithLogger(l *slog.Logger) Option                // emits at save (RFC §18)

// scene — ornaments + decoration (PR #2)
type OrnamentRecipe = ornaments.Recipe
func WithOrnamentExtension(name string, recipe OrnamentRecipe) RenderOption
type Decoration struct {
    node
    Kind     DecorationKind
    Preset   string
    AssetID  AssetID
    Layer    Layer
    Anchor   Anchor
    Offset   Position // NEW — EMU from the anchor
    Size     Size     // NEW — ornament box (zero = default)
    Bleed    bool     // NEW — allow extending past the slide edge
    Opacity  float64  // NEW — 0..1 (0 = default opaque)
    Rotation float64  // NEW — degrees
}
```

All additive; no prior public surface breaks. `pptx.Metadata` mirrors
`scene.Metadata` (the scene re-exports or maps it).

## 10. Risks

- **R1 — Gradient/rotation "repair" in PowerPoint.** New wire shapes.
  **Mitigation:** round-trip + conformance + vendored-schema (preflight); the
  eyeball pass on a gradient/rotated deck.
- **R2 — Determinism.** Gradients, dot scatters, core.xml must be byte-identical.
  **Mitigation:** integer math, fixed dot grids, no timestamps in core.xml;
  determinism tests.
- **R3 — Multi-shape rotation.** No group transform; per-shape rotation only.
  **Mitigation:** documented (D-041); symmetric ornaments unaffected; a warning
  if a caller rotates a decoration whose ornament can't honor it cleanly is
  considered.
- **R4 — Meta determinism / XML-injection.** `SetMetadata` interpolates caller
  strings into XML. **Mitigation:** `encoding/xml`-escape the values; no created/
  modified timestamps (D-035).
- **R5 — Logger noise.** **Mitigation:** save event at Debug/Info, nil-guarded;
  no logger = no logs.

## 11. Acceptance criteria

**PR #1:**
1. A shape with `LinearGradient`/`RadialGradient` round-trips and is conformant;
   a `RadialGradient` emits `<a:gradFill>…<a:path path="circle">`.
2. `WithRotation(d)` sets `<a:xfrm rot="d×60000">` and round-trips.
3. `TokenColorAlpha(role, a)` emits the token's color with `<a:alpha val="a">`.
4. `pres.SetMetadata({Title,Author,Subject})` writes them into `docProps/core.xml`
   (XML-escaped), round-trips, and re-renders **byte-identically**;
   `scene.Render` writes `Scene.Meta` through it.
5. `pptx.WithLogger(l)` emits a save event; no logger = silent.

**PR #2:**
6. Each curated ornament renders at its anchored box in the accent token; a
   bleed decoration uses negative offsets (extends past the edge).
7. `Layer` z-order is honored: a `background` decoration's shapes precede the
   body's in the shape tree; a `foreground` decoration's follow.
8. A caller ornament via `WithOrnamentExtension` renders; opacity/rotation/size
   apply.
9. `make test -race` and `make coverage` pass for touched packages.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` / `internal/ooxml/slide` | existing bands | gradient/rotation/metadata covered by new tests |
| `scene` | 80% | decoration composition |
| `scene/ornaments` | 80% | new scene package |
| `assets/ornaments` | 80% | new curated-asset package |

## 13. Smoke check

`scripts/smoke/phase-13.sh` (grows across both PRs): gradient/rotation/alpha
round-trips; SetMetadata core.xml; WithLogger emits; each ornament renders;
bleed offsets; layer z-order; WithOrnamentExtension. SKIPs until each surface
lands.

## 14. Tests

- **Round-trip golden:** gradient shape, rotated shape, alpha-token shape,
  SetMetadata core.xml (PR #1); a decoration deck (PR #2).
- **Unit:** gradient marshal (ooxml); fills/rotation/alpha/metadata/logger
  (pptx); ornament recipes (deterministic, shape count); registry; decoration
  composition (anchor/offset/bleed/opacity/rotation, layer order).
- **Integration:** a gradient + rotated + metadata deck is conformant +
  byte-identical (PR #1); a layered-decoration deck (PR #2).
- **Fuzz:** none new (no parse surface added).

## 15. Vocabulary added

`Ornament`, `Bleed`, `Gradient fill`, `Rotation` (builder), `Decoration` (updated)
— filed in `docs/glossary.md`.

## 16. Plan deviations encountered during implementation

- **PR #1 (builder foundations + carried fixes) — landed.** Deviations:
  - `core.xml` is built by a new `internal/ooxml/core.BuildCorePropsXML` (so
    `encoding/xml` escaping stays below the P3 wall — pptx must not import
    `encoding/xml`); `pptx.SetMetadata` mutates the existing core part via
    `Part.SetBlob`. Deterministic (no timestamps).
  - The builder logger emits one Debug event in `prepareForWrite` (the shared
    write-path body) rather than per write method — one emit covers
    Save/Write/WriteToBytes/SaveStream.
  - No `coverage.json` change in PR #1 (gradient/rotation/metadata/logger live in
    existing packages); `internal/ooxml/core` gains a direct `BuildCorePropsXML`
    test.
- **PR #2 (ornaments + Decoration) — landed.** Deviations:
  - Ornament recipe contract is `func(sl, box, alpha int, rotationDeg float64)
    int` (accent-locked per RFC §14.2 default; opacity passed as an OOXML alpha
    the recipe applies, rather than a pre-built `Color`). This keeps
    assets/ornaments importing only `pptx` (no scene/ornaments → cycle).
  - Rotation is honored only by single-shape ornaments (`chevron_arrow`) and
    asset decorations are not rotated (no public image-rotation API; the multi-
    shape ornaments can't rotate as a unit without a group transform — D-041).
    Documented; symmetric ornaments are unaffected.
  - `Bleed` is surfaced via a warning when `false` and the box is off-canvas
    (suppressed when `true`), rather than clamping the box — simpler and keeps
    the caller's geometry intact.
  - Scene re-exports the `Anchor` constants + `Position`/`Size` aliases so the IR
    reads `scene.AnchorCenter` etc.
- **Post-PR #2 wiring audit (into PR #25).** A depth pass over the phase fixed:
  - **`decorationBox` anchor alignment** — it centred the box on *every* anchor,
    so a corner anchor (e.g. `AnchorTopLeft`) landed half off-canvas. Now the
    box's anchor-corresponding point (top-left / centre / bottom-right …) aligns
    to the anchor point; this also removes the spurious off-canvas warnings
    corner anchors used to trigger.
  - **Asset-decoration `Rotation` + `Opacity` were silently dropped** (only the
    preset path used them). Added `pptx.Image.SetRotation` (picture `xfrm rot`)
    and `pptx.Image.SetOpacity` (blip `<a:alphaModFix>`, a new wire field) and
    wired both on the asset path — every `Decoration` field is now honored
    (multi-shape preset rotation remains the one documented V1 limitation,
    D-041). `surfaceToken` gaining an alpha field was verified to introduce no
    transparency regression (both constructors set it).

## 17. Sign-off

- [x] PR #1 criteria (1–5) pass; PR #2 criteria (6–9) pass.
- [x] `make coverage` clean for touched packages (assets/ornaments 94.3%,
      scene/ornaments 100%, scene 90%).
- [x] `scripts/smoke/phase-13.sh` `OK ≥ count`, `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary + decisions (D-041, D-042) updated.
- [x] (Phase 20+) Docs site / skills — N/A (inert pre-Phase 20).
