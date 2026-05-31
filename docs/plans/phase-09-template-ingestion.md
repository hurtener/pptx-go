# Phase 09 — Template ingestion (Theme + Masters)

**Subsystem:** pptx (template) + internal/ooxml
**RFC sections:** §13 (§13.1–§13.4), §7.3 (theme source #2), §8.7 (masters/layouts)
**Deps:** Phase 02 (theme + tokens), Phase 03 (builder spine, scaffold), Phase 05 (scene RenderOption surface)
**Status:** Done

---

## 1. Goal

Consume a PowerPoint-emitted `.pptx` brand kit: extract its `Theme` and seed a
new presentation's masters + layouts + theme from it (`pptx.FromTemplate`), and
let a scene render against that brand via `scene.WithTheme` and a
`LayoutKind → layout` `scene.WithLayoutMap`.

## 2. Why now

Wave 3 ("Templates, masters, frames") opens here per `docs/plans/README.md` §3.
Phases 10 (frame chrome) and 12 (icons) both list Phase 09 as a dep, and the
brand-kit story is the canonical theme source (RFC §7.3): everything visual
downstream resolves against a `Theme`, so making that `Theme` come from a real
template is the gating capability. The theme *reader* already shipped early (see
§5); this phase closes the seam between it and the builder/scene.

## 3. RFC sections implemented

- `RFC §13.1` — template ingestion: open a `.pptx`, expose its theme + masters;
  use the theme alone or as the new presentation's starting point.
- `RFC §13.2` — masters & layouts: `Master` exposes layouts; the scene renderer
  maps `LayoutKind` to named layouts via a caller-supplied `LayoutMap`.
- `RFC §13.3` — theme propagation: the theme attached to a `*Presentation` is
  the active theme for subsequent builder calls (per-slide overrides remain V2).
- `RFC §13.4` — brand kits: consume any valid PowerPoint-emitted template;
  emitting a hand-editable template is explicitly **not** in scope (V1.x).
- `RFC §7.3` (source #2) — theme loaded from a `.pptx`. The `LoadTheme` half
  shipped in Phase 02; this phase consumes it (see §5).
- `RFC §8.7` — master/layout management exposed for ingestion.

## 4. Brief findings incorporated

- `docs/research/01-master-layout-theme-ingestion.md` — **F1 (12-slot scheme
  behind `clrMap`)** → the plan tests against a real PowerPoint template with
  `sysClr` slots and `clrMap` indirection, not only a self-authored theme.
- `01-...` — **F3 (inheritance is a relationship chain)** → `FromTemplate`
  **copies the template's theme/master/layout parts wholesale** and rewires
  relationships, rather than reconstructing them from the parsed `Theme`
  (preserves placeholder geometry, list styles, backgrounds).
- `01-...` — **F4 (major/minor fonts)** → the extracted `Theme` carries the
  template's heading/body fonts via the existing `themeFromPart`; font
  *embedding* stays caller-driven (D-019) and out of scope.
- `01-...` — **F5 (`LayoutKind` → named/typed layouts)** → `LayoutMap` plus a
  default mapping onto PowerPoint standard layout types; unmapped/absent layouts
  fall back to blank, never error.
- `01-...` — **F6 (permissive reader)** → the master/layout read path skips
  unrecognized placeholders/extensions and degrades to "theme + standard
  layouts" instead of failing (first foreign-XML read; consistent with Phase 18).

## 5. Findings I'm departing from

- `docs/research/01-master-layout-theme-ingestion.md` — **F2: `LoadTheme`
  already shipped and is tested.** The master plan (`docs/plans/README.md` §3,
  Phase 09 "What lands") lists `pptx.LoadTheme(path)` as Phase 09 scope, but it
  landed with Phase 02's theme work (`pptx/themecodec.go`; `themecodec_test.go`
  already asserts `Resolve(ColorAccent)` from a loaded theme). **Departing
  because** re-introducing it would duplicate shipped code. This plan *consumes*
  `LoadTheme`, adds the brand-kit seeding + scene wiring that did **not** ship,
  and files a follow-up to correct the master-plan entry in the same PR (a
  documented plan deviation, `CLAUDE.md §4.3`). Acceptance criterion 1 is
  re-stated to assert it against a *real PowerPoint template fixture* (stronger
  than the existing self-authored test).

## 6. Decisions referenced

- `D-026` — engine, not product. `FromTemplate` exposes the *mechanism* (adopt a
  template's chrome + theme); choosing a brand kit is the caller's policy. No
  legibility/auto-styling opinion is added.
- `D-035` — byte-identical saves. Copying template parts must keep output
  deterministic: parts are copied in a stable order and relationship ids are
  allocated deterministically; the round-trip golden asserts byte-identity.
- `D-030 / D-033 / D-012` — `Color` interface + token resolution: the extracted
  `Theme` plugs into the same token-resolution path, so a scene authored against
  `ColorAccent` re-renders in the brand accent (P2).
- **New this phase (proposed, number assigned when filed):** template ingestion
  copies the template's theme/master/layout parts wholesale (not reconstructed
  from the parsed `Theme`); `FromTemplate` replaces the `New()` scaffold. Filed
  in `docs/decisions.md` in the implementation PR.

## 7. Architecture

`FromTemplate` is a `New()` option. It opens the template package, swaps the
seeded scaffold parts for the template's theme/master/layout parts, extracts the
`Theme`, and populates `masterCache`. The scene renderer gains two render
options that ride the existing `renderConfig` surface (Phase 05).

```text
pptx.New(pptx.FromTemplate(src))
        │
        ├─ open src (.pptx)  ──▶ copy theme1.xml + slideMaster* + slideLayout* (+rels)
        │                          replacing the New() scaffold; allocate rels deterministically
        ├─ themeFromPart(theme1.xml) ──▶ *Theme  (existing Phase 02 path) ──▶ pres.SetTheme
        └─ MasterManager.LoadFromZip ──▶ masterCache (Master/Layout)

scene.Render(pres, sc,
    scene.WithTheme(brandTheme),      // applies a *pptx.Theme at render time
    scene.WithLayoutMap(myMap))       // LayoutKind → layout name; default map otherwise
        │
        └─ AddSlide(layoutName) ──▶ resolves in masterCache ──▶ emits slide→layout rel
                                     (closes the pptx/presentation.go TODO)
```

`TemplateSource` is an interface (file path / bytes / reader) behind a factory,
mirroring the `§4.4` `ImageSource` seam, so alternate template inputs are
constructors, not new code paths.

## 8. Files added or changed

```text
pptx/master.go                 # NEW — Master, Layout, LayoutMap public wrappers (P3-safe)
pptx/template_ingest.go        # NEW — FromTemplate option + TemplateSource (file/bytes/reader)
pptx/presentation.go           # CHANGED — wire the AddSlide slide→layout relationship (TODO at ~:427, ~:486)
pptx/options.go                # CHANGED — register FromTemplate as a New() Option
scene/scene.go                 # CHANGED — WithTheme, WithLayoutMap RenderOptions on renderConfig
scene/render.go                # CHANGED — honor WithTheme; map LayoutKind→layout via LayoutMap on AddSlide
scene/layoutmap.go             # NEW — default LayoutKind → standard-layout mapping
internal/ooxml/slide/...       # CHANGED (if needed) — permissive master/layout parse hardening (F6)
testdata/brand-template.pptx   # NEW — a genuine PowerPoint-emitted brand kit fixture
scripts/smoke/phase-09.sh      # NEW — phase smoke
docs/research/01-master-layout-theme-ingestion.md  # NEW (this PR's sibling) — brief
docs/research/INDEX.md         # CHANGED — register brief 01
docs/decisions.md              # CHANGED — adds D-037 (wholesale part copy)
docs/design/THEME.md           # CHANGED (if a token/role surfaces) — taxonomy note
docs/glossary.md               # CHANGED — Master, Layout, LayoutMap, TemplateSource, FromTemplate
docs/plans/README.md           # CHANGED — correct Phase 09 entry (LoadTheme shipped in Phase 02)
```

No user-facing skill/doc-site updates: Phase 20 hasn't established them yet
(`CLAUDE.md §19` is inert pre-Phase 20).

## 9. Public API surface

As shipped (supersedes the `TemplateSource` draft — see §16, D-037):

```go
// pptx
func FromTemplate(brand *Presentation) Option   // New() option: adopt brand theme + masters + layouts

type Master struct{ /* … */ }                   // read wrapper over internal master data (P3)
func (m *Master) Name() string
func (m *Master) Layouts() []*Layout
func (m *Master) Layout(name string) (*Layout, bool)

type Layout struct{ /* … */ }
func (l *Layout) Name() string

func (p *Presentation) Masters() []*Master       // RFC §13.1 brand.Masters()
func (p *Presentation) HasLayout(name string) bool

// scene
type LayoutMap map[LayoutKind]string             // LayoutKind → template layout name
func WithTheme(t *pptx.Theme) RenderOption        // apply a brand theme at render time (RFC §13.1/§13.3)
func WithLayoutMap(m LayoutMap) RenderOption       // map scene layout intents to template layouts
func DefaultLayoutMap() LayoutMap                  // LayoutKind → PowerPoint standard layout
```

`pptx.LoadTheme(path)` / `LoadThemeFromBytes` are **already public** (Phase 02);
this phase adds no alias and does not change their signatures.

## 10. Risks

- **R1 — Foreign master/layout XML breaks the parser.** First read of
  non-pptx-go-authored master/layout parts. **Mitigation:** permissive parse
  (skip unknown placeholders/exts, F6); a malformed template degrades to
  "theme extracted, no custom layouts" with a returned error only when the
  package itself is unreadable. Fuzz the master/layout parse path.
- **R2 — Copied parts break the deck's relationship graph.** Rewiring
  theme/master/layout rels into a presentation that already seeded a scaffold can
  orphan or double-wire parts (the PR #13 / repair-prompt class of bug).
  **Mitigation:** `FromTemplate` *replaces* the scaffold parts atomically before
  any slide is added; conformance + a real-PowerPoint open (manual) gate it; the
  round-trip golden asserts no dangling `rId`.
- **R3 — Non-determinism from part copy.** Copying parts via map iteration would
  reintroduce the D-035 class of bug. **Mitigation:** copy in sorted part-URI
  order, allocate rels deterministically; the round-trip golden asserts
  byte-identity across two renders.
- **R4 — `LayoutKind` ↔ template layout mismatch.** A template may lack a layout
  the default map names. **Mitigation:** fall back to the blank/first layout and
  emit a `LayoutWarning`; never error (D-026 — the caller owns layout policy).

## 11. Acceptance criteria

1. Loading a **genuine PowerPoint-emitted** template's theme produces a `Theme`
   whose `ResolveColor(ColorAccent)` equals the template's `accent1` (resolved
   through the master `clrMap`), and whose heading/body fonts match its font
   scheme. (Strengthens the existing self-authored `LoadTheme` test.)
2. `pptx.New(pptx.FromTemplate(TemplateFile("testdata/brand-template.pptx")))`
   yields a presentation whose `Theme()` is the brand theme and whose
   `Masters()` expose the template's layouts by name.
3. A scene rendered with `scene.WithTheme(brandTheme)` emits the brand's accent
   color for an accent-token shape (assert the resolved `srgbClr` in the slide
   XML).
4. `scene.WithLayoutMap` causes `AddSlide` to emit a slide→layout relationship
   pointing at the mapped layout; an unmapped `LayoutKind` falls back to blank
   and records a `LayoutWarning` (no error).
5. A deck built with `FromTemplate` round-trips losslessly through `pptx.Open`
   (G6) and is byte-identical across two renders (D-035); conformance passes.
6. `make coverage` shows the touched/new packages ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `pptx` | 85% | default for the builder package (new `master.go`, `template_ingest.go`) |
| `scene` | 80% | default for the scene package (new options + layoutmap) |
| `internal/ooxml/slide` | 85% | codec band — any master/layout parse hardening |

No band overrides proposed.

## 13. Smoke check

`scripts/smoke/phase-09.sh` (skeleton committed with this plan; SKIPs until the
surface lands) verifies each criterion:

1. `OK:` library builds CGo-free.
2. `OK:` `LoadTheme` on the brand fixture resolves the template accent + fonts (criterion 1).
3. `OK:` `FromTemplate` seeds theme + masters; `Masters()` lists the template layouts (criterion 2).
4. `OK:` `scene.WithTheme` renders the brand accent into slide XML (criterion 3).
5. `OK:` `scene.WithLayoutMap` emits the slide→layout rel; unmapped → blank + warning (criterion 4).
6. `OK:` `FromTemplate` deck round-trips + is byte-identical + conformant (criterion 5).

## 14. Tests

- **Unit:** `pptx` (FromTemplate, Master/Layout wrappers, AddSlide layout rel),
  `scene` (WithTheme/WithLayoutMap, DefaultLayoutMap fallback).
- **Round-trip golden:** yes — a `FromTemplate` deck written → read → model
  equality, plus byte-identity across two renders (D-035).
- **Integration** (`test/integration/`): yes — closes the template→builder→scene
  seam (Deps name Phases 02/03/05); a real-fixture end-to-end:
  `FromTemplate` → `scene.Render(WithTheme, WithLayoutMap)` → reopen → assert
  brand accent + mapped layout on the slide.
- **Fuzz** (`FuzzParseMaster` / `FuzzParseLayout`): yes — first foreign-XML read
  surface (R1); seed corpus = the brand fixture's master/layout parts + a
  truncated/garbage variant; invariant = no panic, returns either data or error.
- **Benchmark:** optional — `BenchmarkFromTemplate` (part-copy cost).

## 15. Vocabulary added

- `Master` — read wrapper over a template's slide master; exposes its layouts.
- `Layout` — read wrapper over a slide layout; exposes its name.
- `LayoutMap` — `scene.LayoutMap`, a `LayoutKind → layout name` mapping.
- `TemplateSource` — sealed builder input for `FromTemplate` (file/bytes/reader).
- `FromTemplate` — the `New()` option that seeds a presentation from a brand kit.
- `Brand kit` — a `.pptx` template with a populated theme + ≥1 master with
  layouts; pptx-go consumes, does not author (RFC §13.4).

## 16. Plan deviations encountered during implementation

- **`FromTemplate` takes a `*Presentation`, not a `TemplateSource`.** RFC §13.1
  shows `pptx.New(pptx.FromTemplate(brand))` where `brand` is an opened deck, and
  `New` returns no error — so adopting an already-valid in-memory presentation
  is cleaner than a path/bytes source that would force `New` to surface an open
  error. The drafted `TemplateSource`/`TemplateFile`/`Bytes`/`Reader` were not
  added (RFC > plan, §2). Filed as D-037. §9's surface is superseded accordingly.
- **Ingestion clones the template package and strips slides** (D-037) rather than
  grafting parts into the scaffold — robust by construction, dissolving risk R2.
- **`LayoutMap` lives in `scene`, not `pptx/master.go`.** It keys on
  `scene.LayoutKind`, and `pptx` must not import `scene` (P1); `scene.LayoutMap`
  + `scene.DefaultLayoutMap` are the home. `pptx/master.go` holds `Master`/
  `Layout` only.
- **Opening any deck now extracts its theme + master registry**
  (`loadPresentationPart`), not just templates — required so `brand.Theme()` /
  `brand.Masters()` work on an opened kit (RFC §13.1). Best-effort (brief 01 F6).
- **Internal parser fixed to capture the layout name + type.** `ParseLayout`
  discarded `cSld@name` (always `""`) and the `sldLayout@type`; both are now
  extracted, which is what makes name-based layout selection possible. (A pre-
  existing gap the survey missed.)
- **Test fixture is hermetic, not a committed PowerPoint binary.** Criterion 1 is
  asserted against a brand `theme1.xml` derived from the conformant scaffold
  theme — which is PowerPoint-shaped (`sysClr` `dk1`/`lt1`, brief 01 F1) — with a
  custom accent + fonts patched in, plus an injected named layout. This exercises
  the `sysClr` fallback and a non-default named layout without depending on an
  opaque binary or the bare-name `ThemeXML()` writer (whose standalone output
  isn't namespace-conformant; a separate, out-of-scope issue).
- **No new theme token**, so `docs/design/THEME.md` is unchanged.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-09.sh` reports `OK ≥ 6` and `FAIL = 0` (6 OK, 0 FAIL).
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (Brand kit, FromTemplate, Master, Layout, LayoutMap).
- [x] Decision entry added (D-037 — clone + strip ingestion).
- [x] Master-plan Phase 09 entry corrected (LoadTheme shipped in Phase 02).
- [x] (Phase 20+) Docs site updated for user-facing surface changes — inert.
- [x] (Phase 20+) Affected agent skill(s) updated — inert.
