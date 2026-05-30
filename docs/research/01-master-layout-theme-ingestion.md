# Brief 01 — master/layout inheritance & theme ingestion

**Subsystem:** pptx (template) + internal/ooxml
**Authored:** 2026-05-30
**Motivating phase:** Phase 09 — Template ingestion (Theme + Masters)

## 1. Question

When pptx-go consumes a PowerPoint-emitted `.pptx` "brand kit", how is its
visual identity actually encoded in OOXML, and what is the minimal, robust
way to (a) extract a `Theme` from it and (b) seed a new presentation's
masters + layouts + theme from it — without `internal/ooxml` wire types
leaking past the `pptx` seam (P3) and without re-synthesizing parts the
template already contains?

## 2. Prior art surveyed

- ISO/IEC 29500-1 (Transitional) — `a:theme`, `p:sldMaster`, `p:sldLayout`,
  `p:clrMap`/`p:clrMapOvr`, the 12-slot `a:clrScheme`, and `a:fontScheme`
  (`a:majorFont`/`a:minorFont`). Vendored under `docs/specifications/`.
- The existing pptx-go surface (read of the code, not a spec): `LoadTheme` /
  `LoadThemeFromBytes` + `themeFromPart`/`toThemePart` (`pptx/themecodec.go`);
  `MasterCache`, `MasterManager.LoadFromZip`, `slide.ParseMaster`/`ParseLayout`
  and the `SlideMasterData`/`SlideLayoutData`/`Placeholder` read models; the
  scaffold seeded by `New()` (`pptx/scaffold.go`).
- PowerPoint's own emitted decks (reference decks under `_gen/`), which carry
  one `theme1.xml`, one `slideMaster1.xml`, and ~11 `slideLayoutN.xml` with
  `type` attributes (`title`, `obj`, `secHead`, `twoObj`, `blank`, …).

## 3. Findings

- **F1 — Color lives in a 12-slot scheme behind a name map.** `theme1.xml`'s
  `a:clrScheme` holds `dk1 lt1 dk2 lt2 accent1..6 hlink folHlink`. `dk1`/`lt1`
  are frequently `a:sysClr` (`windowText`/`window`) with an `lastClr` fallback,
  not `a:srgbClr`. The slide master's `p:clrMap` indirects the *semantic* names
  (`bg1 tx1 bg2 tx2 accent1…`) onto those slots; layouts and slides inherit the
  master's map unless they carry a `p:clrMapOvr`. So "the brand accent" is
  conventionally `accent1`, resolved through the master's `clrMap`.

- **F2 — The role→slot mapping is already implemented and tested.**
  `pptx.LoadTheme(path)` → `themeFromPart` maps `accent1 → ColorAccent`,
  `dk1 → TextPrimary`, etc. `pptx/themecodec_test.go` already asserts
  `LoadThemeFromBytes(...).ResolveColor(ColorAccent)` returns the embedded
  accent. **Phase 09's first acceptance criterion is therefore already met by
  earlier work** (the master plan lists `LoadTheme` as Phase 09 scope — drift;
  see the phase plan §5). Phase 09's value-add is the *brand-kit seeding* and
  *scene-side wiring*, not the theme reader.

- **F3 — Inheritance is a relationship chain, not embedding.** slide →(`slideLayout` rel)→ layout →(`slideMaster` rel)→ master →(`theme` rel)→ theme.
  Placeholders inherit geometry and text style down the chain by `type`/`idx`.
  The cleanest, most faithful ingestion is therefore to **copy the template's
  theme + master + layout parts wholesale** into the new package and rewire the
  relationships, rather than reconstruct them from a parsed `Theme` — copying
  preserves placeholder geometry, list styles, and background fills the
  semantic `Theme` does not capture.

- **F4 — Fonts are a major/minor pair.** `a:fontScheme` carries `a:majorFont`
  (headings, `+mj-lt`) and `a:minorFont` (body, `+mn-lt`). `themeFromPart`
  already maps these to `Theme.HeadingFont`/`BodyFont`. Embedding the bytes is
  a separate, caller-driven concern (D-019, RFC §7.6) and out of Phase 09.

- **F5 — `LayoutKind` must map to *named/typed* layouts.** The scene renderer's
  `LayoutKind` (cover, title-content, two-column, card-grid, full-bleed, blank)
  is an intent enum; a template exposes layouts by name and by `type`. A
  caller-supplied `LayoutMap` (`LayoutKind → layout name`) bridges them, with a
  default map onto PowerPoint's standard layout types. Lookups must degrade: an
  unmapped or absent layout falls back to the blank layout, never an error.

- **F6 — Readers must be permissive (first external XML).** This is the first
  phase that reads *foreign* master/layout XML. PowerPoint emits placeholder
  types, extension lists, and layout variants pptx-go doesn't model. The parse
  path must skip the unrecognized and keep what it understands; a malformed or
  exotic template degrades to "theme extracted, standard layouts only", not a
  hard failure. (Consistent with Phase 18's best-effort read goals.)

## 4. Recommendations

1. **`pptx.FromTemplate(src)` copies parts wholesale.** Open the template
   package, copy its `theme1.xml`, `slideMaster*`, and `slideLayout*` parts (and
   their rels) into the new presentation, replacing the seeded scaffold; then
   extract the `Theme` via the existing `themeFromPart` path so token resolution
   matches the copied parts. Populate `masterCache` from the copied layouts.
2. **Thin public `Master`/`Layout`/`LayoutMap` wrappers** over the internal
   `SlideMasterData`/`SlideLayoutData` (P3: expose names/ids/placeholder bounds,
   never the XML structs). `LayoutMap` is `map[scene.LayoutKind]string`.
3. **`scene.WithTheme(*pptx.Theme)` and `scene.WithLayoutMap(LayoutMap)`** as
   render options; `WithTheme` applies at render time (the brand-kit flow), and
   a nil/absent map uses the default `LayoutKind → standard layout` mapping.
4. **Wire the `AddSlide` layout relationship** (the `pptx/presentation.go`
   `TODO`): when a layout name resolves in `masterCache`, emit the slide→layout
   relationship so the slide actually inherits the template layout.
5. **Test against a genuine PowerPoint-emitted template fixture**, not only a
   self-authored theme, to exercise `sysClr` slots, `clrMap` indirection, and a
   full layout set (F1, F6).

## 5. Open questions

- **Per-slide layout selection from the scene IR.** V1 maps `LayoutKind` →
  layout; richer per-node placeholder targeting (dropping a `hero` into the
  template's title placeholder) is deferred — flagged for the scene composition
  phases / a future RFC note.
- **Emitting a hand-editable template** (the reverse of ingestion) is V1.x per
  RFC §13.4; not in scope here.
- **`clrMapOvr` on layouts/slides.** V1 honors the master's `clrMap`; per-layout
  color-map overrides are rare in brand kits — revisit if a real template needs
  it (would land with a decision entry).
