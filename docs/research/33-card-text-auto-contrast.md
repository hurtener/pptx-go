# Brief 33 — card-text-auto-contrast

**Subsystem:** scene — Layer 2 renderer (card / container chrome)
**Authored:** 2026-06-22
**Motivating phase:** Phase 50 — card-text auto-contrast (R11.2, CRITICAL · engine)

## 1. Question

A card/container chrome text run whose color is left unset (the card header
`RunStyle{TypeRole: TypeH3}` with no `Color`) renders as the slide's default near-
black, regardless of the surface behind it. On a dark card fill or a dark-variant
slide, that header is black-on-dark and barely legible (recreation slides 3, 7); a
`TextAccent` eyebrow on a same-hue header band is invisible (slide 2). How can the
engine make every chrome run legible against whatever surface it lands on —
**without** becoming opinionated about color (D-026, D-058: the engine has *no*
contrast logic and exposes only resolved colors) and **without** changing the
byte output of the common light-surface card?

## 2. Prior art surveyed

- **`scene/render_card.go renderCardChrome`** — the header title (`TypeH3`, no
  `Color` → inherits the placeholder default, i.e. black), the eyebrow
  (`Color: TokenTextColor(TextAccent)`), and the pill label (`TypeCaption`, no
  `Color`). The chrome *knows* its surface here: the card `c.fill`, the
  `c.headerFill` band, and the pill's own `ColorSurfaceAlt` fill.
- **`scene/render_container.go renderColumnJoin`** — the join-badge label already
  uses `Color: TokenTextColor(TextInverse)` (white) on the accent ellipse; correct
  for a dark accent, wrong for a light-accent theme.
- **`scene/render_stat.go`** — the Stat value (`TypeDisplay`, no `Color` → black);
  the label (`TextMuted`) and delta (semantic tone) are explicit.
- **`scene/render_bento.go`** — row labels use `TextMuted` (a deliberate mid-gray).
- **`scene/background.go darkThemeFrom`** — the VariantDark theme: `ColorCanvas
  111827`, `ColorSurface 1F2937`, `ColorSurfaceAlt 374151` (all dark);
  `TextPrimary F9FAFB`, `TextInverse FFFFFF` (preserved white). So for a dark slide
  `r.theme.ResolveColor(role)` already returns the dark value — the same
  `onColor(surface)` computed against `r.theme` handles the variant for free.
- **D-058 (Phase 29)** — `Stats.Colors` exposes resolved per-slide colors; the
  engine deliberately ships *no* contrast logic, leaving the decision to the
  caller.
- DECKARD R11.2 spec: a deterministic `onColor(bg) → TextColor` via sRGB relative
  luminance (pinned integer coefficients + threshold); `renderCardChrome` passes
  an explicit `Color` computed from the surface a run overlaps; eyebrow keeps the
  accent tint only when it clears a min-contrast check; byte-identical when the
  resolved color equals the prior light-surface default; a mechanism the caller can
  still override (D-026).

## 3. Findings

- **Reconciling D-058 / D-026 with R11.2.** R11.2 does not make the engine
  *opinionated*: it adds a deterministic, pinned **mechanism** (`onColor`) that a
  caller can always override by supplying an explicit `Color`, and which is
  **byte-identical** to today on the light-surface default. The engine still
  expresses no taste — it picks the contrast-correct token by a fixed luminance
  rule, the same way `deltaToneColor` maps a tone to a token. This is the
  "mechanism the caller drives" side of D-026, not a legibility *policy* baked into
  product behavior. Recorded as the resolution of the D-058 tension.
- **The byte-identical lever is "nil on a light surface".** An unset `Color` emits
  no `a:solidFill` (`pptx/text_layout.go:190`), so the run inherits the dark
  default. Therefore `onColor` must return **nil** (leave `Color` unset) when the
  surface is light — reproducing today's bytes exactly — and a **light token** only
  when the surface is dark enough that the inherited dark default would be
  illegible. Light-surface cards are byte-identical by construction; only dark
  surfaces (the bug) change.

      func (r *renderer) onCardSurface(bg ColorRole) pptx.Color {
          if relLuminance(r.theme.ResolveColor(bg)) < darkSurfaceLumaMax {
              return pptx.TokenTextColor(pptx.TextInverse) // light text on a dark surface
          }
          return nil // light surface: inherit the dark default (byte-identical)
      }

- **sRGB relative luminance, pinned, integer per call.** WCAG relative luminance
  needs gamma expansion (a linear-luma proxy under-weights blue and would flip a
  legible blue-on-white accent — breaking byte-identical). Precompute the sRGB
  gamma curve once into a 256-entry integer table (`srgbLinear[i] = round(lin ×
  100000)`), then `L = (2126·lin[r] + 7152·lin[g] + 722·lin[b]) / 10000` in
  `[0, 100000]`. The table build uses `math.Pow` at `init` (deterministic); every
  per-call lookup is pure integer → worker-count independent.
- **The dark-surface threshold is the black/white crossover.** Black-vs-white
  contrast is equal at relative luminance `L* = √(1.05·0.05) − 0.05 ≈ 0.179`; above
  it black wins, below it white wins. Pin `darkSurfaceLumaMax = 17912` (`0.17912 ×
  100000`). A saturated teal band (`L ≈ 0.23`) is *above* it → black text (correct:
  black on teal out-contrasts white on teal); a dark navy card (`L ≈ 0.02`) is
  below → white text.
- **Eyebrow keeps its accent only when it clears the surface.** Compute the WCAG
  contrast ratio of `TextAccent` vs the surface; keep `TextAccent` when `ratio ≥
  4.5`, else fall back to `onColor`. On the default white card the default accent
  (`2563EB`) gives `5.17:1` → kept → **byte-identical**; on a same-hue header band
  it fails → falls back to legible. Integer compare: `(max+5000)·10 ≥ 45·(min+5000)`
  with the `0.05` offset scaled to `5000`.
- **Where the surface is known vs not.** `renderCardChrome` knows the surface
  (`c.fill` / `c.headerFill` / pill `ColorSurfaceAlt`); `renderColumnJoin` knows the
  badge `ColorAccent`. These get `onCardSurface` directly and are byte-identical on
  the default theme (header on light fill → nil; join white on dark accent →
  `TextInverse`). Leaf nodes (Stat, Bento label) do **not** receive their container
  surface through `renderNode`; threading it is a broad dispatch change deferred to
  a follow-up. The Stat **value** (uncolored → black) gets `onCardSurface` against
  the slide variant surface (`ColorCanvas`), fixing the common dark-variant case;
  byte-identical on light. The `TextMuted` bento/stat labels are a deliberate
  mid-gray that already clears `4.5:1`-ish on both light and dark, so they are left
  unchanged (a real semantic choice, D-026) — not the reported bug.

## 4. Recommendation

Add `scene/contrast.go` — `srgbLinear` table, `relLuminance(RGB) int`,
`onCardSurface(bg) pptx.Color`, `accentLegible(bg) bool`, pinned
`darkSurfaceLumaMax` / `accentMinContrastT10`. Wire `renderCardChrome` (header
title, pill label → `onCardSurface`; eyebrow → accent-then-`onCardSurface`),
`renderColumnJoin` (join label → `onCardSurface(ColorAccent)`), and the Stat value
(→ `onCardSurface(ColorCanvas)`). Byte-identical on the default light theme; fixes
dark cards / dark variant / same-hue bands. D-082 records the mechanism and the
D-058 reconciliation. Leaf surface-plumbing for Stat-in-colored-card and Bento
labels is a documented follow-up.

## 5. Open questions

- Should `Stat` gain an explicit color override field so a caller can drive its
  contrast inside a strongly-colored card? Deferred — the slide-surface estimate
  covers the reported cases; a `Stat.ValueColor` is additive future work.
