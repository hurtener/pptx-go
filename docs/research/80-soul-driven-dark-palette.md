# Brief 80 — Soul-driven dark palette (R8.3)

> Informs Phase 97 (Wave 15 — theme/soul engine bits). Engine side of the
> `both`-tagged requirement R8.3 (`DECKARD-PRODUCT-REQUIREMENTS.md`,
> CRITICAL · both; D-059 puts the engine mechanism here, the soul
> dark-palette *derivation* in Deckard).

## 1. Motivating phase

Phase 97 makes the scene renderer's VariantDark theme derivation
**soul-drivable**. Today `scene/background.go`'s `darkThemeFrom` hard-codes
the dark variant to a pinned Tailwind gray scale (`darkCanvas=#111827`,
`darkSurface=#1F2937`, `darkSurfaceAlt=#374151`, text `F9FAFB/E5E7EB/9CA3AF`)
and ignores the active theme's brand identity entirely. A brand whose dark
side is a rich deep navy (not neutral gray) cannot be expressed — the only
escape hatch is aiming a `Background` gradient at the accent role, which
yields a flat wash with none of the brand's depth. The gap: a theme carries
no dark-variant overrides, so `darkThemeFrom` has nothing brand-specific to
consume.

## 2. Subsystem / files

- `pptx/theme.go` — `Theme` struct, `Clone()`, the `ThemeOption` builders
  (`WithAccent`/`WithPaper` pattern), `ColorPalette` (`Surfaces`/`Text` maps).
- `scene/background.go` — `darkThemeFrom(base)` + the pinned `dark*` consts.
- `scene/render.go` — `composeOne` captures `stats.Colors` from the resolved
  (possibly dark) per-slide theme via `ResolveColor`/`ResolveTextColor`; no
  change needed (it already reports whatever `darkThemeFrom` produced).

## 3. Findings

- **`darkThemeFrom` already clones the base theme and overlays a fixed set of
  six roles.** Adding soul-driven overrides is a pure overlay *after* the
  pinned defaults: write the pinned grays first (unchanged), then, when the
  theme supplies a dark palette, overwrite role-by-role. An empty/nil dark
  palette leaves the pinned grays untouched → **byte-identical** to today.
- **`Colors.Surfaces`/`Colors.Text` are dynamic `map`s** (grep confirms no
  fixed-size `[N]ColorRole` arrays). A dark palette modeled as the same two
  map types (`map[ColorRole]RGB`, `map[TextColorRole]RGB`) overlays cleanly
  and lets a soul override **any** role — not just the six the pinned default
  touches. This also pre-builds the seam R8.7 (dark accent/extension
  overrides) needs: those are just additional keys in the same maps.
- **A dark palette is *not* an OOXML token and needs no theme1.xml slot.** It
  is consumed only by the scene renderer to derive a per-slide VariantDark
  theme; it is never serialized. Like `ColorPaper`/`TextMuted` (D-104) it is a
  field-without-a-slot. **G6 is about emitted output:** a VariantDark slide's
  resolved dark canvas/surface/text emit as literal `srgbClr` fills/runs that
  round-trip through `pptx.Open` like any color; the *field* is not recovered
  from a reopened deck. The read accessor is the exported `Theme.DarkColors`
  field itself, asserted via `Clone()` deep-copy + the `stats.Colors` hook.
- **`stats.Colors` (D-058) already verifies the resolved dark palette.**
  `composeOne` reads `sr.theme.ResolveColor(ColorCanvas/ColorSurface)` and
  `ResolveTextColor(TextPrimary)` *after* `composeSlide` swapped in the dark
  theme, so a VariantDark slide built on a soul dark palette reports the
  soul's hexes with no new plumbing. R8.3's acceptance is a `stats.Colors`
  assertion; R8.10 generalizes it into a full fidelity gate.
- **Determinism / concurrency.** The overlay is a pure map iteration with
  fixed assignment (no ordering dependence — each role written once). A
  `Theme` is a reusable artifact (§5): the new `DarkColors` field must be
  deep-cloned by `Clone()` and proven race-safe under `-race` concurrent
  reuse, exactly like the existing `Surfaces`/`Text` maps.

## 4. Recommendations

- Add an optional `Theme.DarkColors *DarkPalette` field (nil = pinned-gray
  fallback, byte-identical). `DarkPalette{Surfaces map[ColorRole]RGB; Text
  map[TextColorRole]RGB}` mirrors `ColorPalette` so R8.7 can grow it.
- Add granular, composable options `WithDarkSurface(role, c)` /
  `WithDarkText(role, c)` that lazily allocate `DarkColors` (mirrors the
  `WithAccent`/`WithPaper` ergonomics; a soul that builds a whole palette can
  also set the field directly).
- `darkThemeFrom`: keep the pinned-gray assignments verbatim, then overlay
  `base.DarkColors.Surfaces`/`.Text` when non-nil. Byte-identical when nil.
- `Clone()`: deep-copy `DarkColors` (both maps) when non-nil; leave nil when
  unset.
- Document in THEME.md (the dark-palette mechanism is a visual-color property,
  P2) + glossary + the define-a-theme skill + docs/site theme reference. No
  theme1.xml slot, no codec element, no `restorenamespaces` change.

## 5. Open questions

- Should the dark palette derive sensible defaults from the light palette
  (e.g. darken canvas toward the brand's deepest hue) when unset? No — that is
  the soul's *derivation* job (D-026/D-059, Deckard's half). The engine
  exposes the override mechanism only and falls back to the pinned grays.
- Should `DarkColors` round-trip through theme1.xml? No — VariantDark is a
  scene concept with no OOXML slot; the resolved bytes round-trip, the field
  does not (consistent with `ColorPaper`, D-104). Note in V2 backlog only if a
  real read-back case appears.
- Dark accent/extension overrides (border, accentSoft, accent surfaces) are
  R8.7 — the same maps carry them; Phase 97 ships the seam, Phase 100 wires
  the additional roles + a derived dark-hairline default.
