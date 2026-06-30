# Brief 83 — Dark-variant accents & extensions (R8.7)

> Informs Phase 100 (Wave 15 — theme/soul engine bits). Engine side of the
> `both`-tagged requirement R8.7 (`DECKARD-PRODUCT-REQUIREMENTS.md`,
> MED · both; D-059). This is a **verify-and-close** of the engine half —
> the mechanism shipped in Phase 97 (D-135); the extension tokens are Deckard's.

## 1. Motivating phase

R8.7 wants a VariantDark slide to re-resolve accent surfaces and the
border/borderStrong/accentSoft extension tokens to dark-appropriate values
rather than inheriting light-theme values (the reference's dark cards use
dark-tuned hairlines, not the soul's warm-cream `#E0D5CA` border). Phase 100
confirms the engine already provides the per-variant override mechanism and that
the engine's own borders dark-resolve correctly, and records the split: the
extension *tokens* + the derived-dark-hairline *default* are Deckard's product
half (the engine has no such roles).

## 2. Subsystem / files

- `scene/background.go` — `darkThemeFrom(base)` (the VariantDark derivation).
- `pptx/theme.go` — `Theme.DarkColors *DarkPalette` (D-135) — the override seam.
- `scene/render_card.go` — card border / divider drawing (the borders R8.7 names).
- `scene/contrast.go` — accent-text legibility (the *default* dark accent-text
  re-derivation is the **sibling** requirement R8.6 / Phase 101, kept separate).

## 3. Findings

- **`darkThemeFrom`'s `DarkColors` overlay is already general** (D-135, Phase 97):
  it overlays `base.DarkColors.Surfaces` / `.Text` for **any** role key, not just
  the six the pinned default touches. The Phase-97 white-box test
  (`TestDarkThemeFrom_Overlay`) already overrides `ColorAccent` for the dark
  variant and asserts it wins. So "accent surfaces overridable per variant via
  the dark palette" — the explicit R8.7 capability — is **already shipped**; a
  soul re-resolves any accent/semantic/text role for dark with
  `WithDarkSurface(ColorAccent, …)` / `WithDarkText(TextAccent, …)`.
- **The engine's own borders already dark-resolve.** `render_card.go` draws the
  default card border and divider with `pptx.TokenColor(pptx.ColorSurfaceAlt)`,
  and `darkThemeFrom` overrides `ColorSurfaceAlt` to the dark surface — so a dark
  card's default hairline is the **dark** SurfaceAlt, never a light value. There
  is no cream-border bug in the engine; Deckard's cream border was its soul's
  `border` *extension* token (`#E0D5CA`), which the engine does not have.
- **The accent border (`BorderAccent`) uses `ColorAccent`**, preserved-from-light
  by default (brand identity survives the swap) and re-resolvable per variant via
  `DarkColors.Surfaces[ColorAccent]`. Default-preserve-light is correct engine
  behavior (D-026 — the engine can't invent a "dark-appropriate accent" without
  taste); the soul overrides it.
- **The extension tokens (`border` / `borderStrong` / `accentSoft`) and the
  derived-dark-hairline default are not engine roles.** The engine's `ColorRole`
  set is Canvas/Surface/SurfaceAlt/Accent/AccentAlt/AccentWarm/Success/Warning/
  Error/Info/Paper — no `border`/`accentSoft`. Those live in Deckard's
  `internal/soul` Extensions map, with `ResolveExtension(token, variant)` and the
  low-alpha dark-hairline derivation. That is the product half (D-059).
- **Conclusion: verify-and-close.** The engine mechanism (per-variant accent +
  any-role override) shipped in Phase 97; the engine's default borders already
  dark-resolve. Phase 100 adds the R8.7 acceptance goldens and the decision
  recording the engine/product split — no production change (the R11.1 / Phase-49
  precedent).

## 4. Recommendations

- Ship acceptance goldens proving: (a) a dark card's default border resolves to
  the **dark** SurfaceAlt (distinct from the light SurfaceAlt); (b) a soul's
  `DarkColors.Surfaces[ColorAccent]` re-tints a dark accent border distinct from
  the light accent; (c) a soul's `DarkColors.Text[TextAccent]` re-tints dark
  accent text; (d) byte-identity holds when no `DarkColors` is set.
- File **D-138** recording: R8.7 engine atom = the Phase-97 `DarkColors` overlay
  (already covers accent/semantic/text per variant) + the engine's
  SurfaceAlt-based borders already dark-resolve; the `border`/`borderStrong`/
  `accentSoft` extension tokens + the derived-dark-hairline default are Deckard's
  product half.
- Light §19 touch: clarify the THEME.md dark-palette entry + the define-a-theme
  skill that accent/semantic roles (not just canvas/surface/text) are
  dark-overridable — the capability already exists; the note makes it explicit.
  No new public API.

## 5. Open questions

- Should the engine derive a default dark accent (darken/lighten the light
  accent) when no `DarkColors` is set? No — that is taste (D-026); the engine
  preserves the brand accent and lets the soul override. The *accent-text*
  legibility default per variant is the sibling R8.6 (Phase 101), handled there.
- Should the engine gain `border`/`accentSoft` roles? No — they are Deckard soul
  extension tokens; the engine's SurfaceAlt-based borders already dark-resolve.
  V2 backlog only if a real engine border-token case appears.
