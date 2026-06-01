# Brief 03 — SVG-path → OOXML translator constraints

**Subsystem:** Curated assets (icons, ornaments, frames)
**Authored:** 2026-06-01
**Motivating phase:** Phase 12 — Curated icons (RFC §14.1, D-005)

## 1. Question

How should pptx-go render a curated icon SVG as **native PPTX shapes** (custom
path geometry, no raster), such that:

- the path reads as the intended glyph at slide scale,
- the fill is the active **accent token** (P2),
- the geometry is **deterministic** (byte-identical re-render — D-035),
- the translator handles a **documented SVG subset** and **rejects** anything
  outside it **at registration** (D-005), and
- the whole path stays **in `internal/ooxml` / `internal/render`** (P3) — the
  builder and scene expose SVG bytes and a `*Shape`, never raw OOXML.

## 2. Prior art surveyed

- **SVG path grammar** (W3C SVG 1.1 §8.3): the `d` attribute is a sequence of
  commands `M/m L/l H/h V/v C/c S/s Q/q T/t A/a Z/z` over a user coordinate
  space (the element's `viewBox`). Absolute (upper) and relative (lower) forms;
  `S`/`T` are "smooth" shorthands that reflect the previous control point;
  `A` is an elliptical arc parameterized by radii + flags.
- **OOXML custom geometry** (`a:custGeom`, ECMA-376 §20.1.9.8): a shape
  geometry expressed as `<a:pathLst>` of `<a:path w=".." h="..">`, each holding
  an ordered list of `moveTo / lnTo / cubicBezTo / quadBezTo / arcTo / close`,
  every point an `<a:pt x=".." y=".."/>` in the path's own `w×h` space, scaled
  by PowerPoint to the shape's `spPr` extent. **DrawingML and SVG share the
  same origin convention** — top-left, y-down — so no axis flip is needed.
  `a:arcTo` is parameterized by `wR/hR/stAng/swAng` (sweep angle), **not** by
  SVG's endpoint+flags form.
- **lucide** (the named set pengui-slides v4 references): the icons are
  **stroke-based, multi-element** — `fill="none" stroke="currentColor"
  stroke-width="2"`, with several `<path>`/`<circle>`/`<line>`/`<polyline>`
  children per glyph. They are **not** single filled paths.
- **The frames registry** (`scene/frames/registry.go`, D-038): the curated-asset
  extension-seam pattern (`Curated`/`With`/`Lookup`/`Names` + a per-render
  overlay) icons will mirror.
- **The slide marshaling mechanism** (`internal/ooxml/slide/slide_marshal.go`,
  `internal/ooxml/restorenamespaces.go`): wire structs marshal **bare** (no
  namespace prefix); `RestoreNamespaces` re-applies `a:`/`p:` from the
  `elementNS` map; an ordered **heterogeneous** child list uses a custom
  `MarshalXML` (as `XSpTree` does). `custGeom`/`avLst`/`rect` are already in
  the namespace map; `pathLst`/`path`/`moveTo`/`lnTo`/`cubicBezTo`/`quadBezTo`/
  `close`/`pt` are not yet.

## 3. Findings

- **F1 — lucide raw data is unusable as-is; the set is lucide-*style*.** Because
  lucide glyphs are stroke-based multi-element, they violate the "single path,
  solid fill" constraint by construction. The curated set must be **authored as
  filled single paths** (same names/concepts as lucide where they overlap), not
  copied from lucide's `d` data. This is forced by the RFC §14.1 / D-005
  constraint, not a choice.
- **F2 — `custGeom` is the general target; no preset-shape mapping.** Every icon
  becomes one `<a:custGeom>` with a single `<a:path w h>`. Mapping "simple"
  glyphs onto preset geometries (circle→ellipse) is a fragile optimization that
  buys nothing — `custGeom` covers all of them uniformly. The master plan's
  "preset/path" is read as "path".
- **F3 — coordinate space: `w/h` from the viewBox, integer points.** Set the
  path `w`/`h` to the viewBox dimensions scaled by a fixed factor (×100) and
  round each point to an integer in `[0,w]×[0,h]`. PowerPoint scales the path
  grid to the shape extent, so a square 24×24 viewBox renders correctly in any
  square box; the caller sizes the box. No y-flip (shared origin, F-prior-art).
- **F4 — supported command subset: `M L H V C Q Z` (+ relative + `S`/`T`).**
  `S`/`T` expand to `C`/`Q` by reflecting the previous control point. **`A`
  (elliptical arc) is rejected** in V1: SVG→`arcTo` conversion is lossy and
  wide; curved glyphs are authored with cubic/quadratic Béziers instead (a
  circle is 4 cubics via the kappa≈0.5523 approximation). Rejecting `A` keeps
  the translator small and exact and is a documented constraint.
- **F5 — reject at registration, fail-fast (D-005).** The translator returns an
  error for: more than one `<path>`; any non-path drawable element
  (`circle/line/rect/polyline/polygon` — author them as a path); `fill="none"`
  or a gradient/pattern fill (`url(#…)`); an `A` command; an unparseable `d`.
  Curated assets are validated by a **build-time test** that every embedded SVG
  translates; caller SVGs are validated when `WithIconExtension` is applied
  (before any slide composes), surfacing a Stage-1 error — not a render-time
  warning. (This differs from frames, which validate only name existence at
  render.)
- **F6 — fill is the token, not the SVG's color.** The translator discards the
  SVG's fill *color* (it only checks the fill is solid, not a gradient/none);
  the icon renders filled with the active accent token at `AddIcon` time (P2),
  exactly like a shape's `WithFill(SolidFill(TokenColor(ColorAccent)))`.
- **F7 — layering keeps P1/P3 intact.** `internal/ooxml/slide` gets the
  `custGeom` wire types (+ the `elementNS` additions); `internal/render/
  svgpath.go` parses SVG → those wire types with validation (internal→internal,
  allowed); `pptx` gains `Slide.AddIcon(svg []byte, box, opts…) (*Shape, error)`
  — SVG bytes in, `*Shape` out, raw OOXML never in the signature; `scene/icons`
  holds the name→SVG registry + `ValidateIcon`, and `scene.WithIconExtension`
  wires it. The builder knows how to draw an SVG path but **not** icon *names*
  (P1: names are a scene/curated concept).
- **F8 — ordered heterogeneous path commands need per-command `XMLName`.** Like
  the shape-tree children, a `[]command` where each is `moveTo`/`lnTo`/… can't
  be expressed with static struct tags; each command carries its own
  `xml.Name`, and the new element locals are added to `elementNS` (all `a:`).

## 4. Recommendations

1. **Wire types** (`internal/ooxml/slide`): `XCustomGeometry{ PathList
   XPathList }`, `XPathList{ Paths []XPath }`, `XPath{ W, H int64; Commands
   []XPathCommand }`, `XPathCommand{ XMLName xml.Name; Pts []XPoint }`,
   `XPoint{ X, Y int64 }`. Add `CustomGeom *XCustomGeometry` to
   `XShapeProperties` (sibling of `PresetGeom`). Add `pathLst/path/moveTo/lnTo/
   cubicBezTo/quadBezTo/close/pt` to `elementNS` (`a:`). A round-trip test
   proves a `custGeom` shape survives write→read.
2. **Translator** (`internal/render/svgpath.go`): `Translate(svg []byte)
   (*slide.XCustomGeometry, error)` — parse the single path, validate F5, expand
   `S`/`T`, scale coords (F3), emit commands. Pure + deterministic.
3. **Builder** (`pptx`): `Slide.AddIcon(svg []byte, box Box, opts …ShapeOption)
   (*Shape, error)` — translate, build an auto-shape carrying `custGeom`, apply
   the fill (default accent token if no `WithFill`), return the `*Shape`. New
   public API ⇒ a round-trip golden + smoke in this PR.
4. **Registry** (`scene/icons`): mirror frames — `Registry{ name→[]byte }`,
   `Curated()` (from `go:embed`), `With`, `Lookup`, `Names`; plus
   `scene.ValidateIcon(svg) error` and `scene.WithIconExtension(name, svg)`
   that validates at apply time.
5. **Assets** (`assets/icons`): a hand-authored **starter set (~16)** of
   single-path filled glyphs (arrows, chevrons, check, x, plus, minus, square,
   circle, triangle, star, diamond, dot), `go:embed`'d. Grow toward ~60 in a
   tracked follow-up.

## 5. Open questions

- **Elliptical arcs (`A`).** Deferred (F4). If a future curated glyph genuinely
  needs an arc that a Bézier can't approximate cleanly, add SVG-arc→`arcTo`
  conversion then — a separate, well-scoped change with its own test corpus.
- **The full ≈60-icon set.** This phase ships the engine + a ~16 starter set;
  growing to the D-005/RFC §14.1 target of ~60 is a content follow-up (tracked
  in the plan and `docs/V2-BACKLOG.md` when it lands), not an engine change —
  each new icon is one validated SVG.
- **Visual fidelity.** Automated checks prove the path is conformant and
  deterministic but cannot judge whether a glyph *reads* as its name; that needs
  the PowerPoint / Quick Look eyeball pass the wave already budgets.
