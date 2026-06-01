# Phase 10 — frame chrome

**Subsystem:** assets/frames + scene
**RFC sections:** §14.3, §14.4
**Deps:** Phase 09 (template ingestion — the scene render path, theme, and
layout map this phase composes against).
**Status:** Done

---

## 1. Goal

A scene `Image` node can wrap its picture in one of four curated native-shape
device frames — `browser`, `phone`, `desktop`, `laptop` — and a caller can
register additional frames by name via `scene.WithFrameExtension`.

## 2. Why now

The master plan (`docs/plans/README.md`, Wave 3) sequences frame chrome as
Phase 10, immediately after template ingestion (Phase 09) and immediately
before the full Image node + media-manager refactor (Phase 11). Frames are the
project's **first curated-asset extension seam** — the interface + factory +
driver pattern (`CLAUDE.md §4.4`) that icons (Phase 12) and ornaments (Phase
13) will each repeat. Landing it here, on the simplest of the three curated
sets (four shape recipes, no SVG translator), establishes the registry +
per-render-overlay + Stage-1-closed-name-validation shape the later curated
phases copy. Phase 10 also makes the `Image` node render for the first time —
today it falls through `renderNode`'s default and only emits a
"not yet implemented" warning.

## 3. RFC sections implemented

- `RFC §14.3` — **Frame chrome.** Fully implemented: the four curated frames
  as native shape recipes whose interior region accepts an image; the recipe
  is positioned by the renderer and the image inserted into the interior.
- `RFC §14.4` — **Extensibility.** Implemented for frames:
  `scene.WithFrameExtension(name, recipe)`; closed-name Stage-1 validation
  picks up the extended set; extensions live on the `RenderOption`, not global
  state. (Icons/ornaments repeat this pattern in Phases 12–13.)
- `RFC §14.3` (image insertion) — **partially**, shared with **Phase 11**.
  Phase 10 renders the `Image` node only as far as framing requires: resolve
  the asset, place the picture (into the frame interior when framed), set alt
  text. Crop/fit, dedup-pool refactor, MIME-detection hardening, and
  aspect-aware fitting are **Phase 11** (`render_image.go` is the seam Phase 11
  extends).

## 4. Brief findings incorporated

- `docs/research/02-device-frame-shape-geometry.md` —
  - **F1** (rounded rects + ellipses suffice) → recipes use only
    `ShapeRect`/`ShapeRoundRect`/`ShapeEllipse`; no custom path, so Phase 10
    stays independent of the Phase 12 SVG translator.
  - **F2** (recipe returns interior, renderer inserts image) → `Recipe`
    signature is `func(*pptx.Slide, pptx.Box) (interior pptx.Box, shapes int)`;
    `render_image.go` calls the recipe, then `AddImage` into the interior.
  - **F3** (proportional geometry, no theme param) → recipes compute device
    proportions as integer-EMU ratios of the region; no `SpaceRole`
    resolution, no `*Theme` handle.
  - **F4** (existing tokens, no new token) → bezel color flows through
    `ColorSurfaceAlt`/`ColorSurface` and the traffic lights through
    `ColorError`/`ColorWarning`/`ColorSuccess`; **no `docs/design/THEME.md`
    entry needed** (frames are token composers, P2 by reuse).
  - **F5** (reference by name; enum is the ergonomic alias) → add
    `Image.FrameName string`; it wins when non-empty, else `FrameKind`
    selects a curated name. Recorded as **D-038**.
  - **F6** (per-render extension overlay) → `WithFrameExtension` folds over a
    copy of the curated registry per `Render`; the resulting registry is
    read-only during the parallel compose (D-035 determinism preserved).
  - **F7 / Open-Q "true group shape"** → Phase 10 emits the bezel as
    individually-positioned native shapes bounded by the region; a real OOXML
    group-shape builder primitive is deferred (noted in `docs/V2-BACKLOG.md`
    when that file lands; tracked here in §10 R3).

## 5. Findings I'm departing from

None. The plan adopts brief 02's recommendations as written. The two brief
**open questions** (aspect-ratio handling; true group shape) are resolved by
*deferral*, not departure: aspect-aware fit → Phase 11; group-shape primitive
→ post-V1 (§10 R3). The phone-notch approximation (status strip, not a
subtractive cutout) is adopted from F-Open-Q pending the Phase 12 path
translator.

## 6. Decisions referenced

- **D-038 (NEW, filed in this PR)** — *Frame reference: enum alias + named
  registry.* `Image` keeps its shipped `Frame FrameKind` enum and gains an
  optional `FrameName string`. The frame registry is keyed by name; the four
  curated `FrameKind` values map to the four reserved curated names
  (`browser`/`phone`/`desktop`/`laptop`); `FrameName` selects a name (curated
  **or** caller-registered) and takes precedence over the enum when set;
  `FrameNone` + empty `FrameName` ⇒ no frame. Mirrors the `Decoration`
  enum-plus-`Preset`-string precedent. An unknown resolved frame name fails
  Stage-1 validation (closed-name semantics, §14.4).
- `D-024` — *Assets by reference (AssetID + AssetResolver).* The framed image's
  bytes arrive through the resolver, exactly as `CodeBlock` already does.
- `D-026` — *Engine, not product.* A frame is a mechanism (a caller opts into a
  named bezel); there is no deck-wide "always frame images" mode and no
  legibility heuristic. An unresolved asset or (defensively) an unknown frame
  at render degrades to a `LayoutWarning`, never a panic.
- `D-035` — *Byte-identical idempotency.* Recipes are pure integer-EMU
  geometry (no map iteration, no wall-clock); the per-render registry is
  read-only during compose. The existing scene determinism test covers a
  framed deck.
- `D-015` — *Parallel render with sequential media slides.* An `Image` slide
  registers global media, so it must be classified asset-bearing
  (`nodeUsesAssets`) and render sequentially in scene order.

## 7. Architecture

A frame **recipe** is pure geometry that composes the public `pptx` builder
(P1): given a slide and a region, it emits the bezel as native shapes and
returns the **interior** box plus its bezel shape count. The renderer inserts
the image into the interior.

```text
scene.Render(opts: …, WithFrameExtension("retro", recipe))
   │  build per-render registry = curated ∪ extensions   (F6, read-only after)
   │  Stage-1: ValidateScene(s) + validateFrameRefs(s, registry)   (D-038)
   ▼
renderNode → case Image → renderImage(ps, box, v, slideID)
   │
   │  name := resolveFrameName(v)         // FrameName ?: kindName(Frame)
   │  if name != "" {
   │     recipe := registry.lookup(name)
   │     interior, n := recipe(ps, box)   // bezel emitted; n shapes
   │     stats.Shapes += n
   │  } else { interior = box }
   │
   ▼
ps.AddImage(ImageBytes(assetBytes, ct), interior).SetAltText(v.Alt)   // Phase 11 extends

assets/frames  (package frames)         scene/frames  (package frames)
  Browser/Phone/Desktop/Laptop            type Recipe func(*pptx.Slide, pptx.Box) (pptx.Box, int)
  func(sl, region) (interior, shapes)     Curated() *Registry        // wires the four
  imports: pptx only                      (*Registry) With/Lookup/Names
                                          imports: pptx, assets/frames
scene (package scene)
  type FrameRecipe = frames.Recipe        // alias → public extension API
  WithFrameExtension(name, FrameRecipe) RenderOption
  render_image.go: renderImage + frame-name resolution + validateFrameRefs
  imports: pptx, scene/frames
```

Import graph (no cycle, P1 intact): `pptx` ← `assets/frames` ← `scene/frames`
← `scene`. `pptx` imports none of them; `scene` never reaches under `pptx`.

## 8. Files added or changed

```text
assets/frames/frames.go             # NEW — Recipe type doc + shared geometry helpers
assets/frames/browser.go            # NEW — Browser recipe (window + toolbar + traffic lights)
assets/frames/phone.go              # NEW — Phone recipe (slab + status strip + home indicator)
assets/frames/desktop.go            # NEW — Desktop recipe (monitor + stand + foot)
assets/frames/laptop.go             # NEW — Laptop recipe (screen + base deck)
assets/frames/frames_test.go        # NEW — interior-within-region + shape-count + determinism
scene/frames/registry.go            # NEW — Recipe type, Registry, Curated(), Lookup, Names, curated-name consts
scene/frames/registry_test.go       # NEW — curated lookup, extension overlay, unknown-name miss
scene/render_image.go               # NEW — renderImage, frame-name resolution, validateFrameRefs
scene/render_image_test.go          # NEW — framed/unframed render, alt text, unknown-frame validation error
scene/nodes.go                      # CHANGED — Image gains FrameName; frameKindName helper
scene/render.go                     # CHANGED — dispatch Image; nodeUsesAssets(Image)=true; registry plumbed into renderer
scene/scene.go                      # CHANGED — FrameRecipe alias, WithFrameExtension, renderConfig.frames, registry build + validateFrameRefs in Render
test/integration/frame_image_test.go# NEW — framed-image deck round-trips + byte-identical + conformant
scripts/smoke/phase-10.sh           # NEW — phase smoke (acceptance criteria)
docs/research/02-device-frame-shape-geometry.md   # NEW — informing brief
docs/research/INDEX.md              # CHANGED — lists brief 02
docs/decisions.md                   # CHANGED — adds D-038
docs/glossary.md                    # CHANGED — Frame, FrameRecipe, Frame registry
docs/plans/phase-10-frame-chrome.md # NEW — this plan
```

No user-facing doc-site / skills updates: those artifacts do not exist until
Phase 20 (`CLAUDE.md §19` is inert pre-Phase 20). No `docs/design/THEME.md`
change — frames introduce no new visual property/token (brief F4).

## 9. Public API surface

```go
// scene
//
// FrameRecipe draws a device frame's bezel into region and returns the
// interior Box an image is placed into, plus the number of bezel shapes
// emitted. It composes the public pptx builder only (P1).
type FrameRecipe = frames.Recipe

// WithFrameExtension registers a caller frame under name for this render.
// The name joins the closed curated set {browser, phone, desktop, laptop};
// a duplicate name overrides the curated recipe for this render only. The
// extension is per-render (not global). An Image referencing name then
// resolves and renders like a curated frame.
func WithFrameExtension(name string, recipe FrameRecipe) RenderOption

// scene — Image gains an optional named-frame selector (D-038).
type Image struct {
    node
    AssetID   AssetID
    Alt       string
    Frame     FrameKind // curated enum (ergonomic alias)
    FrameName string    // NEW — named frame; wins over Frame when non-empty
}

// scene/frames
type Recipe func(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int)

type Registry struct{ /* unexported */ }
func Curated() *Registry                                 // the four curated frames
func (r *Registry) With(name string, rec Recipe) *Registry  // returns a copy + overlay
func (r *Registry) Lookup(name string) (Recipe, bool)
func (r *Registry) Names() []string                      // sorted, for validation messages

const ( NameBrowser = "browser"; NamePhone = "phone"
        NameDesktop = "desktop"; NameLaptop = "laptop" )

// assets/frames
func Browser(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int)
func Phone(sl *pptx.Slide, region pptx.Box)   (interior pptx.Box, shapes int)
func Desktop(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int)
func Laptop(sl *pptx.Slide, region pptx.Box)  (interior pptx.Box, shapes int)
```

No prior public surface breaks: `Image.FrameName` is additive (the existing
`Frame FrameKind` field and its zero value `FrameNone` are unchanged).

## 10. Risks

- **R1 — PowerPoint "repair" / blank-render on the new bezel geometry.** Frames
  add overlapping rounded-rect + ellipse clusters. **Mitigation:** every shape
  is a builder primitive that already round-trips and passes
  `scripts/validate-schema.sh` (preflight); the integration test asserts the
  framed deck is conformant + byte-identical; and the user runs the framed-image
  deck through PowerPoint and macOS Quick Look/Keynote (the only oracles for
  "does it render", per the carry-forward note).
- **R2 — Frame interior aspect ≠ image aspect.** A device interior implies an
  aspect; the caller image may differ. **Mitigation:** Phase 10 stretches to
  fill (builder default `FitFill`) and documents that aspect-aware fit is Phase
  11. No silent cap — the behavior is stated in the plan and godoc.
- **R3 — "Shape group" fidelity.** RFC §14.3 says "shape group"; the builder has
  no public group primitive in V1, so a framed image is a cluster of sibling
  shapes (it won't move as one object in PowerPoint). **Mitigation:** visually
  and round-trip-equivalent for V1; a builder `AddGroup` primitive is recorded
  as a post-V1 backlog item (brief 02 Open-Q), not silently dropped.
- **R4 — Extension-registry concurrency.** `WithFrameExtension` mutating shared
  state under the parallel renderer would race. **Mitigation (F6):** the
  registry is built once before compose and is read-only during it; `With`
  returns a copy rather than mutating in place; the determinism + `-race`
  tests cover a framed multi-slide deck.

## 11. Acceptance criteria

1. Each curated frame (`browser`, `phone`, `desktop`, `laptop`) renders its
   inner image **inside the bezel interior** — the image box is strictly within
   the node region and does not equal it (the bezel occupies the margin).
2. A caller-extended frame registered via `scene.WithFrameExtension("name",
   recipe)` resolves and renders an image through that recipe's interior.
3. An `Image` with `Frame == FrameNone` and empty `FrameName` renders the
   picture at the node box with **no bezel shapes** (back-compat).
4. An `Image` whose resolved `FrameName` is not in the render's registry fails
   **Stage-1 validation** with a clear error (closed-name, §14.4); a curated
   `FrameKind` always resolves.
5. A framed-image deck **round-trips** through `pptx.Open` (the picture and its
   alt text survive) and re-renders **byte-identically** (D-035), and the
   package is OOXML-conformant.
6. `Image` is classified asset-bearing (`nodeUsesAssets`) so framed slides
   render sequentially in scene order; `make test -race` and `make coverage`
   pass for touched packages.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for scene packages |
| `scene/frames` | 80% | default for new scene package |
| `assets/frames` | 80% | new curated-recipe package; treated as a scene-adjacent composer (recipes are exercised by `frames_test.go` + the scene render tests). Added to `internal/coveragecheck/coverage.json` in this PR. |

No band lowered; `assets/frames` is a new package and gets an 80% entry in
`coverage.json` (a new package with no configured threshold fails the gate —
`CLAUDE.md §11`).

## 13. Smoke check

`scripts/smoke/phase-10.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` each curated frame places the image inside the interior
   (`scene` test: framed image box ⊂ node box, bezel shapes > 0).
3. `OK:` `scene.WithFrameExtension` renders a caller frame
   (`scene` test).
4. `OK:` `FrameNone` renders no bezel (back-compat) (`scene` test).
5. `OK:` unknown `FrameName` is a Stage-1 validation error (`scene` test).
6. `OK:` framed-image deck round-trips + byte-identical + conformant
   (`test/integration`).

## 14. Tests

- **Unit:** `assets/frames` (interior ⊂ region, shape count, determinism of the
  geometry math); `scene/frames` (curated lookup, extension overlay copy,
  unknown-name miss); `scene` (frame-name resolution precedence, validation,
  framed vs unframed render, alt text, asset-bearing classification).
- **Round-trip golden:** yes — a framed `Image` round-trips through `pptx.Open`
  (picture + alt text), and the deck re-renders byte-identically.
- **Integration** (`test/integration/`): yes — `Deps` names Phase 09 (a
  different subsystem's shipped phase) and this phase opens the curated-asset
  extension seam other phases (12/13) build on. Real `internal/opc` write +
  `encoding/xml` decode + temp-file round-trip, `-race`, ≥1 failure mode
  (unresolved asset → warning; unknown frame → validation error).
- **Fuzz:** none (no new parse/decode surface — frames are emit-only).
- **Benchmark:** none required (a frame is a handful of `AddShape` calls; not a
  hot reusable artifact).

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Frame` — device-bezel chrome (browser/phone/desktop/laptop) drawn as
  native shapes around a scene `Image`; selected by `FrameKind` or
  `FrameName`.
- `Frame recipe` (`FrameRecipe` / `scene/frames.Recipe`) — the function that
  emits a frame's bezel into a region and returns the image interior.
- `Frame registry` — the per-render closed-name set of frame recipes
  (curated ∪ `WithFrameExtension`), consulted at render and validated in
  Stage 1.

## 16. Plan deviations encountered during implementation

Filled in **as** implementation happens.

- **Browser toolbar drawn as the window's top strip, not a separate bar
  rect.** §7 sketched the browser as a rounded window plus a toolbar
  rectangle. A square-cornered toolbar rect starting at the window's top
  would undercut the window's rounded top corners (the rect corners poke
  past the rounded body). The recipe instead fills the whole rounded window
  with the chrome color (`ColorSurface`) and treats the strip above the
  interior as the toolbar — the image covers the content area, leaving the
  top strip reading as the toolbar, with the three traffic-light dots on it.
  Same visual, no corner artifact, one fewer shape. Acceptance criterion 1
  is unchanged (image still strictly inside the interior; bezel shapes > 0).
- **`Recipe` type lives in `scene/frames`, not `assets/frames`.** Defining it
  once in `scene/frames` and matching the curated functions structurally
  (Go assigns an unnamed func type to the named `Recipe`) avoids an
  `assets/frames` → `scene/frames` import that would cycle. `assets/frames`
  stays a pure `pptx`-only leaf; `scene/frames` owns the type + registry.
  (Consistent with §7's import graph; the §8 annotation on `frames.go`
  refers to its package-doc description of the recipe shape.)

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages (assets/frames 98.3%,
      scene/frames 100%, scene 90.4%).
- [x] `scripts/smoke/phase-10.sh` reports `OK ≥ 6` and `FAIL = 0` (7 OK, 0
      FAIL).
- [x] Prior phases' smoke scripts still pass (preflight PASS).
- [x] Glossary updated (Frame chrome / Frame recipe / Frame registry).
- [x] Decision entries added (D-038).
- [x] (Phase 20+) Docs site updated — N/A (inert pre-Phase 20).
- [x] (Phase 20+) Affected agent skill(s) updated — N/A (inert pre-Phase 20).
