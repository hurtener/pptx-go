# Brief 85 — Soul→engine roundtrip verification (R8.10)

> Informs Phase 102 (Wave 15 — theme/soul engine bits; the Wave-15 capstone).
> Engine side of the `both`-tagged requirement R8.10
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · both; D-059). The engine extends
> the `Stats.Colors` hook + proves resolved == theme; the soul-fidelity *gate*
> is Deckard's.

## 1. Motivating phase

R8.10 wants a deterministic check that what a soul declares is what the engine
renders — across light and dark variants, surfaces, text, and accents — so brand
drift is caught before a deck ships. The engine already reports a per-slide
resolved canvas/surface/primary-text via `Stats.Colors` (D-058), but only those
three roles, so a caller cannot verify the accent set or the alternate surface
reached the render per variant. Phase 102 is the theme analogue of the
multi-archetype conformance corpus (R14.19): it extends the resolved-color hook
and proves the full brand soul (light + dark palette + multi-accent + gradient)
round-trips to the rendered colors on both variants.

## 2. Subsystem / files

- `scene/scene.go` — `SlideColors` (the D-058 struct) + `Stats.Colors`.
- `scene/render.go` — `composeOne` captures `SlideColors` from `sr.theme`
  (already the derived dark theme for VariantDark).
- `scene/render_colors_test.go` — the existing per-slide-colors tests.

## 3. Findings

- **`SlideColors` must stay comparable.** `render_colors_test.go`'s determinism
  test compares `seq.Colors[i] != par.Colors[i]` with `==`, so adding a slice
  field (e.g. the multi-accent palette) would break it. Extend with **scalar
  `pptx.RGB` fields only** — `SurfaceAlt`, `Accent`, `AccentAlt`, `TextAccent` —
  which keeps `SlideColors` comparable and the existing test intact.
- **The new fields dark-resolve for free.** `composeOne` already captures from
  `sr.theme` *after* `composeSlide` swapped in the dark theme for a VariantDark
  slide, so `ResolveColor(ColorAccent)` etc. report the **dark-variant** value
  (including a soul's `DarkColors` override). No new plumbing — just more
  `ResolveColor` / `ResolveTextColor` reads in the existing capture block.
- **Additive + byte-identical.** `SlideColors` is pure metadata returned in
  `Stats`; it is never emitted into the deck. Adding fields changes no rendered
  byte. The capture is a pure map lookup per slide → deterministic across worker
  counts (the existing determinism test now covers the new fields via `==`).
- **The multi-accent *palette* is theme-level, not per-slide.** `Theme.Accents`
  is the same across slides; a fidelity check compares it directly against the
  theme, not per-slide. So `SlideColors` need not carry the palette — the
  per-slide accent *roles* (`ColorAccent`/`ColorAccentAlt`), which dark-resolve,
  are what belongs there.
- **The fidelity *comparison* is Deckard's (D-059).** "Resolved == the soul's
  declared value" requires the soul's intended tokens, which live in Deckard. The
  engine cannot know "the soul's value" — it reports what it resolved
  (`Stats.Colors`) and proves, in a round-trip test, that the resolved colors
  equal the active theme's tokens per variant. Deckard wires its soul-fidelity
  gate (intended-vs-resolved) into its own contract/capability gate using
  `Stats.Colors`.

## 4. Recommendations

- Extend `SlideColors` with `SurfaceAlt`, `Accent`, `AccentAlt`, `TextAccent`
  (`pptx.RGB`), captured in `composeOne` from `sr.theme` (so they dark-resolve).
  Update the struct doc to note the new roles + that it stays comparable.
- Ship a Wave-15 fidelity capstone test: a non-default brand soul (custom light
  surfaces/accents via `WithAccents`/`WithPaper`, a `DarkColors` dark palette
  overriding accent + text, a named brand gradient) rendered across a light and a
  dark slide; assert every `SlideColors` field equals the **variant** theme's
  resolved value (the active theme for light, `darkThemeFrom` for dark); a
  deliberate token mismatch fails the assertion; identical across worker counts.
- File **D-140** and consolidate the **Wave-15 close**: R8.1 / R8.2 / R8.8 / R8.9
  are pure-product (or product-dominant) — note the engine atoms each relies on
  are already present (the complete `Surfaces`/`Text`/`DarkColors`/`Accents`/
  `Gradients` theme + `pptx.FromTemplate` theme extraction + `Stats.Colors`), and
  the product half (bootstrap params, brand-source acquisition, establish-before-
  build flow) is Deckard's — mirroring the Wave-14 close (D-133).
- Light §19: a THEME.md / docs-site note that `Stats.Colors` reports the resolved
  surface/accent/text per slide per variant for fidelity verification.

## 5. Open questions

- Should the engine expose a public soul-fidelity helper? No — the
  intended-vs-resolved comparison needs the soul's tokens (Deckard's). The engine
  ships the resolved-color hook + the proof test; Deckard owns the gate (D-059).
- Should `SlideColors` carry the full `Accents` palette per slide? No — it is
  theme-level and a slice (would break comparability); compare it against the
  theme directly. The per-slide accent *roles* suffice.
- Wire the fidelity fixture into preflight? The engine ships it as an ordinary
  `-race` test (runs in CI); Deckard wires its soul-fidelity gate into its own
  contract gate (D-059).
