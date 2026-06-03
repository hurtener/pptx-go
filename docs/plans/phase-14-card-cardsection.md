# Phase 14 — Card + CardSection

**Subsystem:** scene (composites) + pptx (builder shadow primitive)
**RFC sections:** §11.2 (card, card_section), §12 (per-node policy), §16 (v4 mapping)
**Deps:** Phase 07 (two_column/grid layout), Phase 12 (icons), Phase 13 (chrome/ornaments + gradient/rotation primitives)
**Status:** Draft

---

## 1. Goal

Render the `card` and `card_section` scene nodes as native, theme-aware PPTX
chrome — rounded-rect background, accent stripe, optional icon / eyebrow /
header / header-pill, and a body laid out by `body_layout` — across all v4
card knobs, with real (editable) drop-shadow elevation; `card_section`
composes grids / two_columns / nested cards.

## 2. Why now

Phase 14 is the next phase in the master plan (`docs/plans/README.md`, Wave 4),
and its deps have all shipped: containers (Phase 07), the icon engine
(Phase 12, D-040), and the chrome/ornament + gradient/rotation builder
primitives (Phase 13, D-041). Phase 15 (Flow) depends on Phase 14's card-like
step pill, so cards unblock the rest of Wave 4. Phase 14 also closes a
documented Phase-12 deferral — the icon registry was built but never
*consumed*; the card's optional icon is the first node to place one (T3).

## 3. RFC sections implemented

- `RFC §11.2` — the `card` and `card_section` container nodes (this phase
  completes the §11.2 container set begun by `two_column`/`grid` in Phase 07).
- `RFC §12` — the per-node policy row "`card` / `card_section` → Native
  (rounded rect + accent stripe + body shapes; children render per their own
  policy)" (§12.1 table) and the mixed-policy example in §12.2 (a
  `card_section` containing a `code_block`: native chrome + one `pic`).
- `RFC §16` — the v4 → IR mapping: `card → Card` (1:1 incl. `header_pill`,
  `body_layout`, `layout`, `fill`, `border_style`, `size`, `elevation`) and
  `card_section → CardSection` (1:1).
- `RFC §8` — extends the builder shape surface with a shadow effect primitive
  (the only genuinely new OOXML capability this phase needs — P1).

## 4. Brief findings incorporated

- `docs/research/05-card-chrome-and-shadow-primitive.md`:
  - **F1 (only the shadow is new OOXML)** → the card composes existing builder
    calls; the sole new *builder capability* is the shadow primitive (PR#1),
    keeping P1 intact.
  - **F2 (shadow is a real builder primitive)** → `pptx.WithShadow(Elevation)`
    / `pptx.WithElevation(role)` `ShapeOption`s emit `<a:effectLst><a:outerShdw>`;
    wire types in `internal/ooxml/slide`, `restorenamespaces` entries, a
    round-trip golden — the D-041 pattern.
  - **F3 (dir/dist conversion)** → the theme `Elevation`'s cartesian
    `OffsetX/OffsetY` map to polar `dist`/`dir` with integer rounding, keeping
    D-035 byte-identical determinism.
  - **F4 (icon consumption closes the Phase-12 deferral)** → store
    `cfg.icons` (curated ∪ `iconExt`) in `renderConfig`, add `validateIconRefs`
    to Stage-1, resolve name→bytes→`AddIcon` at compose.
  - **F5 (icons are native → plain card is media-free)** → only a card whose
    *body* holds an `Image`/`CodeBlock` is media-bearing; extend
    `nodeUsesAssets` to recurse `Card.Body`/`CardSection.Body`.
  - **F6 (additive IR)** → new fields default to zero = current output; new
    enums re-exported into `scene`.
  - **F7 (card_section = card with container body)** → shared chrome helper,
    two body strategies (leaves stacked by `body_layout` vs containers
    recursed through the layout engine).
  - **F8 (card internal geometry)** → background rect → accent stripe →
    header row (icon + eyebrow/title + right-aligned pill) → body region,
    padding scaling with `Size`, all integer EMU.

## 5. Findings I'm departing from

- `docs/research/05-…` **R4 / F6 — "reconcile `Outline` vs `BorderStyle`,
  the cleaner option folds `Outline` into `BorderStyle`."** The brief lists
  folding as cleaner but flags it breaks the existing field. **Departing
  toward preservation:** keep the shipped `Outline bool` field and add
  `BorderStyle BorderStyle` alongside it, with `BorderDefault` (zero) meaning
  "defer to `Outline`" (Outline=false → no border, Outline=true → neutral
  solid border). This preserves byte-identical output for every existing
  `Card{…, Outline: …}` (D-035) — additive growth, no field removal — and is
  recorded in D-043. A future major version may fold them; V1 (v0.x, but the
  IR is becoming load-bearing) keeps both.

## 6. Decisions referenced

- `D-041` — gradient/rotation primitives built in V1 — the precedent this
  phase follows for **building** (not approximating) the `outerShdw` shadow
  primitive; the card's accent stripe / fills reuse its token-color
  mechanisms.
- `D-040` — icon engine + `AddIcon(svg, box, opts)` — the card consumes this
  to place its optional icon.
- `D-038` — curated-asset registry pattern (Curated ∪ extension overlay,
  closed-name Stage-1 ref check) — `validateIconRefs` mirrors
  `validateFrameRefs`/`validateOrnamentRefs`.
- `D-035` — byte-identical saves — the card layout, the shadow
  dir/dist rounding, and the additive IR all preserve determinism.
- `D-015` — parallel render; media-bearing slides sequential — extends
  `nodeUsesAssets` to classify image-bearing cards.
- `D-026` — engine, not product — the card renders the IR's typed knobs with
  no legibility heuristics, no auto-sizing opinions, no doc-mode.
- **`D-043` (new, this PR)** — Phase 14 builds the `outerShdw` shadow
  primitive and splits delivery (PR#1 primitive, PR#2 cards + icon wiring);
  the Card IR grows additively, `Outline` and `BorderStyle` coexist. Filed in
  `docs/decisions.md` in PR#1.

## 7. Architecture

Two PRs, one plan (the D-042 split pattern).

**PR#1 — builder shadow primitive** (`pptx` + `internal/ooxml/slide`):

```text
internal/ooxml/slide/effect.go     # NEW — XEffectList, XOuterShadow wire types
internal/ooxml/slide/slide_types.go# CHANGED — XShapeProperties gains EffectList (after Line)
internal/ooxml/restorenamespaces.go# CHANGED — "effectLst","outerShdw" → a:
pptx/shape.go                      # CHANGED — WithShadow / WithElevation ShapeOptions
```

`<a:effectLst>` follows `<a:ln>` in `CT_ShapeProperties`'s content model, so
`EffectList` appends after `Line` in `XShapeProperties`. `WithElevation(role)`
resolves the theme `Elevation` at `AddShape` time; `WithShadow(Elevation)` is
the literal escape hatch (P2 — token path is `WithElevation`). A flat
elevation (`IsFlat()`) emits no `effectLst` (byte-identical to today).

```text
ShapeOption: WithElevation(ElevationRaised)
  → theme.ResolveElevation → Elevation{Blur,OffsetX,OffsetY,Color,Alpha}
  → <a:effectLst><a:outerShdw blurRad=Blur dist=hypot(dx,dy)
       dir=atan2(dy,dx) rotWithShape="0">
       <a:srgbClr val=Color><a:alpha val=Alpha/></a:srgbClr>
     </a:outerShdw></a:effectLst>
```

**PR#2 — Card + CardSection + icon wiring** (`scene`):

```text
scene/nodes.go            # CHANGED — Card gains Eyebrow/Icon/HeaderPill/BorderStyle/Size/Layout
scene/render_card.go      # NEW — chrome helper + renderCard
scene/render_card_section.go # NEW — renderCardSection (container body)
scene/render.go           # CHANGED — dispatch Card/CardSection; nodeUsesAssets recursion; preferredHeight
scene/scene.go            # CHANGED — cfg.icons built in Render; validateIconRefs
scene/validate.go         # CHANGED — validateIconRefs (closed-name Stage-1)
scene/tokens.go (or nodes.go) # CHANGED — re-export BorderStyle/CardSize/CardLayout consts
```

A card's slot decomposes deterministically:

```text
┌─────────────────────────────┐  ← background roundRect (Fill, BorderStyle, WithElevation)
│▌ [icon] Eyebrow      (pill) │  ← accent stripe (left edge) + header row (padding inset)
│▌       Header               │
│▌                            │
│▌  body (BodyLayout)         │  ← body region: leaves stacked V/H (card)
│▌                            │     or containers recursed (card_section)
└─────────────────────────────┘
```

`renderCardChrome` (shared) emits the background + stripe + header; `renderCard`
stacks leaf body via `stackIn` (vertical) or a `layout.Columns` split
(horizontal); `renderCardSection` recurses each container child through the
normal `renderNode` dispatch (like `renderGrid`). Nesting composes because the
body region is just another `Box`.

## 8. Files added or changed

```text
# PR#1 — shadow primitive
internal/ooxml/slide/effect.go        # NEW — XEffectList, XOuterShadow
internal/ooxml/slide/slide_types.go   # CHANGED — XShapeProperties.EffectList
internal/ooxml/restorenamespaces.go   # CHANGED — effectLst/outerShdw → a:
pptx/shape.go                         # CHANGED — WithShadow, WithElevation
pptx/shape_shadow_test.go             # NEW — round-trip golden + unit
docs/decisions.md                     # CHANGED — adds D-043
docs/design/THEME.md                  # CHANGED — elevation/shadow mechanism note
scripts/smoke/phase-14.sh             # NEW — smoke (PR#1 criteria OK, card criteria SKIP)
docs/glossary.md                      # CHANGED — outer shadow / elevation primitive
docs/research/05-card-chrome-and-shadow-primitive.md # NEW (committed with plan)
docs/plans/phase-14-card-cardsection.md              # NEW (this file)

# PR#2 — Card + CardSection + icon wiring
scene/nodes.go                        # CHANGED — Card IR expansion; BorderStyle/CardSize/CardLayout
scene/render_card.go                  # NEW — chrome helper + renderCard
scene/render_card_section.go          # NEW — renderCardSection
scene/render.go                       # CHANGED — dispatch + nodeUsesAssets + preferredHeight
scene/scene.go                        # CHANGED — cfg.icons; build registry; validateIconRefs call
scene/validate.go                     # CHANGED — validateIconRefs
scene/render_card_test.go             # NEW — render + round-trip goldens, idempotency
scene/icons_consume_test.go           # NEW — icon name→AddIcon; unknown name fails Stage-1
scripts/smoke/phase-14.sh             # CHANGED — flip card SKIPs to OK
docs/glossary.md                      # CHANGED — card knob vocabulary
```

No user-facing skills/docs site yet (pre-Phase 20 — §19 inert).

## 9. Public API surface

```go
// pptx (PR#1)
// WithElevation renders a drop shadow from the active theme's Elevation token
// for the given role (the token path — P2). Flat elevation emits no effect.
func WithElevation(role ElevationRole) ShapeOption
// WithShadow renders a drop shadow from a literal Elevation (the escape hatch).
func WithShadow(e Elevation) ShapeOption

// scene (PR#2) — additive Card fields + new enums
type BorderStyle int
const ( BorderDefault BorderStyle = iota; BorderNone; BorderSolid; BorderAccent )
type CardSize int
const ( CardSizeMD CardSize = iota; CardSizeSM; CardSizeLG )   // MD = zero = current
type CardLayout int
const ( CardLayoutDefault CardLayout = iota; CardLayoutIconLeft; CardLayoutIconTop )

type Card struct {
    node
    Header     string
    Eyebrow    string        // NEW — kicker label above the header
    Icon       string        // NEW — curated/extension icon name (closed-name)
    HeaderPill string        // NEW — pill badge text, right of the header row
    Body       []SlideNode
    BodyLayout BodyLayout
    Fill       ColorRole
    Outline    bool          // preserved (D-043): zero-state border shorthand
    BorderStyle BorderStyle  // NEW — explicit border; BorderDefault defers to Outline
    Size       CardSize      // NEW — padding/min-height scale; CardSizeMD = current
    Layout     CardLayout    // NEW — header arrangement
    Elevation  ElevationRole
}
// CardSection unchanged in shape; gains the same chrome via the shared helper.
```

No prior public surface breaks: every added field/enum has a zero value
reproducing current behavior.

## 10. Risks

- **R1 — Shadow breaks byte-identical output for existing shapes.**
  **Mitigation:** `WithElevation`/`WithShadow` are opt-in `ShapeOption`s; a
  shape with no shadow option (every shape today) emits no `effectLst` —
  identical bytes. `IsFlat()` elevations also emit nothing. A golden over the
  existing scaffold proves no diff.
- **R2 — dir/dist float rounding drifts across runs.** **Mitigation:** the
  conversion is `int(math.Round(...))` over integer EMU inputs; the serialized
  value is a single rounded integer (1/60000°), deterministic by construction
  (D-035). A round-trip golden pins it.
- **R3 — Icon wiring makes a plain card render sequentially (loses
  parallelism).** **Mitigation:** `AddIcon` emits `custGeom`, not media;
  `nodeUsesAssets` classifies a card by its *body* only (Image/CodeBlock),
  not by its icon — a text+icon card stays parallel. A parallel-equivalence
  test (workers=1 vs N) over an icon card proves byte-identity.
- **R4 — `Outline`/`BorderStyle` ambiguity.** **Mitigation:** explicit
  precedence in D-043 and code: `BorderDefault` → defer to `Outline`; any
  non-default `BorderStyle` wins. A table-driven test covers all combinations.
- **R5 — `card_section` body accepts a leaf (mis-use).** **Mitigation:**
  Stage-2 validation already constrains container bodies; `card_section`
  recurses whatever it's given through `renderNode` (a leaf renders in the
  body region, harmlessly). No panic; a structural warning if a leaf appears
  where a container is expected (non-fatal — D-026).

## 11. Acceptance criteria

1. A shape created with `WithElevation(ElevationRaised)` round-trips losslessly
   through `pptx.Open` (the `outerShdw` blurRad/dist/dir/color/alpha survive).
2. A shape with no shadow option, and a shape with `WithElevation(ElevationFlat)`,
   emit **no** `<a:effectLst>` (byte-identical to pre-Phase-14 output).
3. A `Card` with `{Header, Body}` and all new fields at zero renders
   byte-identically to the pre-Phase-14 Card render (additive-IR invariant).
4. Each card knob renders: `Fill` (surface role), `BorderStyle`
   (none/solid/accent), `Size` (sm/md/lg padding), `Elevation`
   (flat/raised/elevated shadow), `BodyLayout` (vertical/horizontal),
   `Layout` (icon-left/icon-top), `HeaderPill`, `Eyebrow`, `Icon`.
5. A `Card` with `Icon: "<curated-name>"` places that icon (a `custGeom`
   shape) in the header; an unknown icon name fails Stage-1 validation with a
   render error **before** compose.
6. A `card_section` containing a `grid` of `card`s (card-of-cards) renders the
   section chrome plus each nested card's chrome and body.
7. A `card_section` containing a `code_block` renders native chrome + one
   `pic` shape (mixed-policy, RFC §12.2).
8. `scene.Render` is byte-identical for an icon/card scene at `workers=1` and
   `workers=N` (idempotency under parallelism, D-035/D-015).
9. `make coverage` shows the touched packages ≥ their bands.
10. `scripts/smoke/phase-14.sh` reports `OK ≥ count(criteria)`, `FAIL = 0`;
    prior phases' smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `pptx` | 85% | default; shadow primitive adds to the existing package |
| `internal/ooxml/slide` | 85% | codec band — new effect wire types |
| `scene` | 80% | default; card composers add to the existing package |

No new packages → no new `coverage.json` band entry (cards/shadow extend
existing packages). If a new package is introduced during implementation, add
its band in the same PR.

## 13. Smoke check

`scripts/smoke/phase-14.sh` (PR#1 lands the script with card criteria as
`SKIP`; PR#2 flips them to `OK`):

1. `OK:` library builds CGo-free.
2. `OK:` `WithElevation(ElevationRaised)` emits `<a:outerShdw>` that round-trips.
3. `OK:` a no-shadow shape emits no `<a:effectLst>` (byte-identical guard).
4. `OK:` (PR#2) a Card renders chrome (background + stripe + header).
5. `OK:` (PR#2) each card knob renders (fill/border/size/elevation/body_layout/layout/pill/eyebrow).
6. `OK:` (PR#2) a Card with a curated icon name places the icon; unknown name fails Stage-1.
7. `OK:` (PR#2) a card_section of cards renders (card-of-cards).
8. `OK:` (PR#2) a card_section with a code_block renders native chrome + a pic.
9. `OK:` (PR#2) render is byte-identical at workers=1 vs N for a card scene.

## 14. Tests

- **Unit:** `pptx` (shadow option → wire shape; dir/dist conversion;
  flat-elevation no-op), `scene` (chrome geometry, knob coverage, border
  precedence table, icon name resolution, nodeUsesAssets classification).
- **Round-trip golden:** yes — `outerShdw` (PR#1) and a representative Card
  (PR#2) write → `Open` → assert model equality (G6).
- **Integration** (`test/integration/`): yes — Phase 14 `Deps` name Phase 12
  (icons, a different subsystem) and Phase 14 *consumes* the icon-registry
  seam Phase 12 opened. An integration test renders a real deck with an icon
  card through real `internal/opc` + `encoding/xml`, asserting the icon
  `custGeom` reaches the slide part and an unknown name errors.
- **Fuzz:** none (no new parse/decode surface; `outerShdw` read path is
  covered by the round-trip golden).
- **Benchmark:** optional — a single-slide card-grid render (not a gate).

## 15. Vocabulary added

File in `docs/glossary.md` (alphabetical) in the respective PR:

- `accent stripe` — the thin themed bar along a card's edge (PR#2).
- `card chrome` — a card's non-content shapes: background rect + accent
  stripe + header row (PR#2).
- `elevation primitive` — the builder `outerShdw` drop-shadow effect that
  realizes the `Elevation` token (`WithElevation`/`WithShadow`) (PR#1).
- `eyebrow` — a small kicker label above a card's header (PR#2).
- `header pill` — a pill-shaped badge in a card's header row (PR#2).

## 16. Plan deviations encountered during implementation

- **CardLayout ships two values, not three.** §9 listed
  `CardLayoutDefault / CardLayoutIconLeft / CardLayoutIconTop`. Implemented
  `CardLayoutDefault` (icon left of the eyebrow/header stack) and
  `CardLayoutIconTop` (icon stacked above) — `IconLeft` folded into `Default`
  to avoid a redundant synonym (Default *is* icon-left). Honors brief Q1
  (ship the variants the reference decks use; defer the rest — RFC §11.3). No
  acceptance criterion changes.
- **Acceptance criterion 3 restated.** As written it asserts a zero-field Card
  "renders byte-identically to the pre-Phase-14 Card render." Pre-Phase-14 the
  renderer had no Card case — a Card produced a `LayoutWarning` ("not yet
  implemented; node skipped") and *no shapes*, so there is no prior render to
  match. The meaningful invariant — render determinism for the additive IR
  (D-035) — is tested instead by `TestCardParallel` (byte-identical at
  workers=1 vs N) plus the unchanged catalog/policy tests (`scene_test.go`).
  The shadow primitive's own byte-identical guard (`TestShadowOmittedWhenFlat`,
  PR#1) covers the no-perturbation property at the builder layer.
- **CardSection IR kept minimal.** It carries `Header` + `Body` only (no
  Fill/Icon/BorderStyle/etc.); it shares `renderCardChrome` with chrome
  defaults. Expanding its knob set is deferred — additive when a reference deck
  needs it.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-14.sh` reports `OK ≥ 9` and `FAIL = 0` (9 OK, 0 FAIL).
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-043).
- [ ] (Phase 20+) Docs site updated for user-facing surface changes. (inert)
- [ ] (Phase 20+) Affected agent skill(s) updated. (inert)
