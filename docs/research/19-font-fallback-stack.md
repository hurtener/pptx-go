# Brief 19 — font-fallback-stack

**Subsystem:** pptx — Layer 1 builder (theme/typography + font embedding)
**Authored:** 2026-06-22
**Motivating phase:** Phase 36 — declared fallback chain per type role

## 1. Question

`pptx.FontSpec.Family` is a single string with no fallback. When a brand face is
neither installed nor embedded, PowerPoint and any rasterizer silently
substitute an arbitrary host font — so a deck that names "Playfair Display" but
ships without it looks like generic Arial rather than a controlled near-serif. A
soul has no way to say "if my display face is unavailable, fall back to *this*
specific serif" instead of the host default. How can a type role declare an
ordered fallback chain that the engine realizes deterministically at write time —
additively, byte-identical when unused, and without the engine knowing what is
installed on the consumer's machine?

This is `DECKARD-PRODUCT-REQUIREMENTS.md` R9.6 (`font-fallback-stack`, MED ·
both). Per D-059 pptx-go implements the engine half (carry the chain + resolve
the emitted face at write time); the default-soul fallback families are Deckard's.

## 2. Prior art surveyed

- **`pptx.FontSpec`** already carries the resolved type-scale fields (Family,
  Size, Weight, Italic, Tracking, LineHeight, Case, AvgCharWidth). A fallback
  chain is one more per-role field, set by the soul — the same shape as the other
  R9 type-detail tokens.
- **OOXML run properties.** The run `a:latin` typeface is **single-valued**;
  DrawingML has no multi-font fallback at the run level (unlike CSS
  `font-family`). So a fallback cannot be *emitted* as a list — it must be
  *realized* by choosing one face at write time and recording it as the run's
  typeface.
- **Phase 35 / D-065 (font-embedding pass).** Already runs a save-time pass in
  `prepareForWrite` that consults the registered `FontSource` and walks every
  slide's runs. A `FontSource` is the natural authority on which faces "exist"
  for a deck: a face the source resolves is available (and embeddable); one it
  cannot resolve is unavailable. The fallback pass shares that machinery and runs
  just before embedding.
- **D-064 (AvgCharWidth).** Precedent for a per-role `FontSpec` field that is a
  theme-time constant and is documented in `docs/design/THEME.md` as a
  type-scale input rather than a visual token. Fallback follows the same model.

## 3. Findings

- **The chain lives on `FontSpec` (per role).** `FontSpec.Fallback []string` —
  ordered substitute families (e.g. `["Playfair Display", …]` declares the role's
  primary in `Family` and substitutes in `Fallback`). Empty (the default) means
  no fallback — byte-identical.
- **A `FontSource` is the availability oracle.** The engine cannot know what is
  installed on the consumer's machine, but it can ask the registered
  `FontSource`: a family the source resolves is treated as available; one it
  cannot is unavailable. With **no** `FontSource`, the engine makes no
  availability judgement and emits the primary as-is (byte-identical).
- **Realization is a deterministic write-time substitution.** Before the slides
  serialize, build a resolution map from the active theme: for each role with a
  non-empty `Fallback`, if the source resolves `Family` keep it; else pick the
  first `Fallback` entry the source resolves; if none resolve, keep the primary.
  Then rewrite every run whose `a:latin` typeface equals a substituted primary to
  the chosen face. Iterate roles in a fixed order (`TypeDisplay…TypeCode`) so the
  map is deterministic; the substitution is a pure function of (theme, source).
- **Idempotent and byte-identical across saves.** The map is keyed on primary
  families; after the first save a substituted run carries the *fallback* family,
  which is not a primary key, so a second save is a no-op — two saves are
  byte-identical, and two identical decks resolve identically.
- **Independent of embedding (but ordered before it).** Fallback applies whenever
  a `FontSource` is registered and a role declares a chain — even with
  `WithFontEmbedding` off (the R9.6 accept: "with embedding disabled, render the
  declared fallback"). Running it before `autoEmbedFonts` means embedding (when
  on) picks up the *resolved* faces, so the embedded bytes match the emitted
  typeface, and ties R9.7 (no italic cut → fall back rather than faux-italic).
- **Mechanism, not taste.** The engine carries and resolves the chain; the chain
  *contents* (which serif, which sans) are the soul's choice (D-026).

## 4. Recommendations

1. `pptx/theme.go`: `FontSpec.Fallback []string` (ordered; empty = none).
2. `pptx/fonts.go`: `resolveFontFallbacks()` — build the theme→source resolution
   map (sorted role order), rewrite runs' `a:latin` to the resolved face; run in
   `prepareForWrite` before `syncSlides`/`autoEmbedFonts`. Gated on
   `fontSource != nil` + a declared chain (byte-identical otherwise).
3. A read path: `FontSpec.Fallback` is an exported field; the *resolved* emitted
   face round-trips via the existing `Run.Font()` (no new OOXML field).
4. `docs/design/THEME.md`: document the fallback chain on the type scale; refresh
   the font-embedding section for `WithFontEmbedding` (D-065).
5. Tests: substitution to the first resolvable fallback; primary wins when the
   source resolves it; no `FontSource` / empty chain byte-identical; determinism
   + idempotency across two saves.

## 5. Open questions

- **Availability probe style/weight.** Availability is a *family* question; the
  probe resolves the regular cut (`style ""`, weight 400). A family that ships
  only a non-regular cut is an edge case the soul can address by listing it
  explicitly; documented.
- **Same family, different chains across roles.** If two roles share a `Family`
  but declare different `Fallback` chains, the fixed role-iteration order makes
  the first-seen mapping win deterministically. Rare; documented.
- **Theme-scheme (major/minor) fallback.** Runs that inherit the theme font (no
  per-run `a:latin`) are not rewritten; the theme1.xml font scheme already
  carries the host's substitution. The per-run faces are where a brand face
  lands.
