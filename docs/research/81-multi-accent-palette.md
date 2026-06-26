# Brief 81 — Multi-accent brand palette (R8.4)

> Informs Phase 98 (Wave 15 — theme/soul engine bits). Engine side of the
> `both`-tagged requirement R8.4 (`DECKARD-PRODUCT-REQUIREMENTS.md`,
> HIGH · both; D-059 puts the ordered-palette *mechanism* + index cycle here,
> the soul *populating* the palette from a brand source in Deckard).

## 1. Motivating phase

Phase 98 lets a theme define an **ordered brand-accent palette** (4+ coordinated
hues) that the scene renderer's per-element accent cycle reads, instead of the
pinned five-role cycle. A pro deck rotates jade / orange / violet / lime across
cards, columns, timeline phases, quadrant points, tree nodes; today the engine's
`timelineAccent` cycles a fixed `[ColorAccent, ColorAccentAlt, ColorInfo,
ColorSuccess, ColorWarning]`, so a brand with four *accent* hues collapses to two
plus three semantic colors. The gap: there is no theme-level accent palette for
the cycle to read.

## 2. Subsystem / files

- `pptx/theme.go` — `Theme` struct, `Clone()`, the `WithAccent`/`WithDarkSurface`
  option pattern.
- `scene/render_timeline.go` — `timelineAccent(idx) ColorRole`, the pinned cycle.
- `scene/render_funnel_cycle.go`, `scene/render_image.go`,
  `scene/render_quadrant.go`, `scene/render_tree.go` — the five `timelineAccent`
  call sites (fill/stroke + contrast text).
- `scene/contrast.go` / `scene/render_table.go` — `onCardSurface(role)` /
  `cellTextOn(role)`: the accent index also feeds auto-contrast text, so the
  resolver must expose a resolved RGB, not just a role.

## 3. Findings

- **`timelineAccent` is a free function returning a `ColorRole`; all five callers
  wrap it in `pptx.TokenColor(...)` (fill/stroke) and two also pass it to
  `r.cellTextOn(role)` (contrast text).** Generalizing it to an arbitrary-length
  brand palette means the cycle can no longer be expressed purely as a role — a
  brand's 4th/5th hue has no spare `ColorRole`/OOXML accent slot (accent1..6 are
  already claimed by Accent/AccentAlt/AccentWarm/Success/Warning/Error). So the
  palette is best modeled as an ordered list of **literal RGB** hues that the
  cycle resolves by index (`index → hex`, a pure lookup — exactly the
  requirement's spec).
- **Byte-identity pivots on "palette empty".** When the theme defines no brand
  palette, the cycle must return *exactly* `TokenColor(timelineAccent(idx))` and
  the contrast path must resolve *exactly* `ResolveColor(timelineAccent(idx))` —
  i.e. the pinned five-role behavior, unchanged. The new path is taken only when
  `len(Theme.Accents) > 0`.
- **The contrast path needs a resolved-RGB core.** `onCardSurface` /
  `cellTextOn` take a `ColorRole` and resolve it internally. A literal-RGB accent
  has no role, so factor a `bg RGB`-keyed core (`onSurfaceRGB`, `cellTextOnColor`)
  out of them; the existing role-keyed wrappers route through it
  (`ResolveColor(role)` → core). Behavior-preserving and byte-identical — it just
  lets the accent contrast be computed from a literal hue too.
- **No new scene field / IR node.** The accent-bearing nodes already carry an
  `AccentIndex` (Milestone, funnel/cycle stage, image highlight/pin, quadrant
  item, tree node). Generalizing the cycle they all call satisfies the mechanism;
  the catalog stays 35. Extending `AccentIndex` to `Card`/`Stat` headerFill is
  additional scope (those already take an explicit `*ColorRole`) — defer.
- **`pptx.RGB` is already a `pptx.Color`** (implements `resolve`) and
  `relLuminance(pptx.RGB)` consumes it directly — so one palette entry serves both
  the fill path (as a `Color`) and the contrast path (as an `RGB`) with no
  conversion.
- **Determinism / concurrency.** Index→entry is a pure modulo lookup over an
  immutable slice; `Clone()` must deep-copy the slice (a `Theme` is a reusable
  artifact, §5) and a `-race` concurrent-reuse test proves it.

## 4. Recommendations

- Add `Theme.Accents []RGB` (ordered brand-accent palette) + a `WithAccents(...RGB)`
  option. Empty (zero) = the pinned role cycle (byte-identical). `Clone()`
  deep-copies the slice.
- Add two renderer resolvers: `accentColorAt(idx) pptx.Color` (fill/stroke) and
  `accentRGBAt(idx) pptx.RGB` (contrast) — both cycle `Theme.Accents` when
  non-empty, else fall back to `TokenColor`/`ResolveColor(timelineAccent(idx))`.
  Keep `timelineAccent` as the pinned fallback cycle.
- Refactor the five call sites to the resolvers; refactor `onCardSurface` /
  `cellTextOn` to route through RGB-keyed cores so accent contrast works from a
  literal hue. Byte-identical when `Accents` is empty.
- Document in THEME.md (the accent palette is a visual-color property, P2) +
  glossary + the define-a-theme skill + docs/site. No theme1.xml slot (an
  arbitrary-length palette has no fixed slots; like `DarkColors`/`ColorPaper` the
  resolved accent RGB round-trips, the field does not).
- This unblocks the deferred R14.2 ChartStyle palette and R14.14 per-section
  accent — both want an ordered accent palette to index into.

## 5. Open questions

- Should accents be `[]ColorRole` (reorder existing roles) instead of `[]RGB`?
  No — a brand's 4–6 hues exceed the available roles/slots; the requirement asks
  for arbitrary brand hexes indexed by position (`index → hex`).
- Should `Card`/`Stat`/`Column` gain an `AccentIndex` that resolves through the
  palette? Useful but additional scope (they already accept an explicit
  `*ColorRole`); defer to a follow-up. R8.4's mechanism is the cycle reading the
  palette.
- Auto-deriving on-light/on-dark accent *text* colors per palette entry is R8.6
  (contrast-aware accent text) — the palette here carries surface hues; R8.6 adds
  the per-variant legible text derivation. Kept separate.
