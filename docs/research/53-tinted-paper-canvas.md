# Brief 53 — Tinted-paper canvas token (R13.1)

> Informs Phase 70 (Wave 13 — backgrounds & finish). Engine side of the
> `both`-tagged requirement R13.1 (`DECKARD-PRODUCT-REQUIREMENTS.md`,
> HIGH · both; D-059 puts the engine half here, the soul/composer half in
> Deckard).

## 1. Motivating phase

Phase 70 adds a `ColorPaper` surface token so a deck can paint a faintly
tinted off-white "paper" canvas (≈ `#FAFAF8`) distinct from pure white. Pro
reference decks never use flat `#FFFFFF` for content slides — the warm/cool
paper tone is part of what reads "designed". The engine already has
`BackgroundColor` + a surface-role palette; the gap is that there is no
dedicated canvas-tint token *separate from* pure white, so a caller cannot
ask for "the paper tone" by role.

## 2. Subsystem / files

- `pptx/theme.go` — `ColorRole` iota + `DefaultTheme().Colors.Surfaces`.
- `pptx/tokenresolve.go` — `ResolveColor` (map lookup, dynamic).
- `pptx/themecodec.go` — `writeSlots` / `themeFromPart` (theme1.xml mapping).
- `scene/background.go` / `scene/render.go` — `BackgroundColor` already
  resolves any `ColorRole` via `pptx.TokenColor(bg.Color)`; no change needed.

## 3. Findings

- **`Colors.Surfaces` is a `map[ColorRole]RGB`, not a fixed-size array.**
  Appending a `ColorRole` value (after `ColorInfo`) cannot break any
  role-indexed array — `grep` confirms no `[N]ColorRole` arrays or
  range-over-fixed-roles loops exist; `Clone()` copies the whole map
  dynamically; `ResolveColor` is a plain map lookup with a black fallback.
- **theme1.xml has exactly 12 OOXML slots** (`dk1/lt1/dk2/lt2/accent1..6/
  hlink/folHlink`). They are already fully claimed by the 10 surface + 2 text
  roles in `writeSlots`. There is **no spare slot** for `ColorPaper`, and
  `themecodec.go`'s own header already documents the convention: *"Roles
  without a slot (e.g. `TextMuted`) keep their default."* `ColorPaper` joins
  that set — it is **not** serialized to theme1.xml and reads back as its
  in-memory default. This is correct and consistent: the soul/caller sets the
  paper tint on the `*Theme` it constructs at write time; it is not recovered
  from a re-opened deck's theme1.xml (the resolved RGB *is* recovered — see
  next finding).
- **G6 round-trip is about emitted output, not the token.** A
  `BackgroundColor{Color: ColorPaper}` slide resolves `ColorPaper` to a
  literal `srgbClr` inside the full-slide rect's `solidFill` at write time.
  That fill round-trips losslessly through `pptx.Open` like any other solid
  color — the round-trip test asserts the reopened rect carries the off-white
  RGB, not that the *token* survived.
- **Byte-identity.** `ColorPaper` defaults to `FFFFFF` (= `ColorCanvas`), so
  a `BackgroundColor{Color: ColorPaper}` slide on the default theme is
  byte-identical to a `ColorCanvas` one today, and any untouched deck (no
  `ColorPaper` reference) is wholly unaffected. The off-white only appears
  when a caller/soul overrides the token on its theme.

## 4. Recommendations

- Append `ColorPaper` to the `ColorRole` iota **after `ColorInfo`** (last
  position → existing values unchanged).
- Add `ColorPaper: "FFFFFF"` to `DefaultTheme().Colors.Surfaces` (= canvas,
  byte-identical default; settable to off-white).
- Add a `WithPaper(c RGB)` theme option for ergonomics (mirrors `WithAccent`).
- Document in THEME.md: `ColorPaper` is a surface role with **no theme1.xml
  slot** (keeps its default on read-back, like `TextMuted`); the resolved
  background RGB still round-trips. Glossary term + define-a-theme skill +
  docs/site theme reference.
- No new `BackgroundKind`, no codec element, no `restorenamespaces` change —
  this is a pure semantic-token addition.

## 5. Open questions

- Should `ColorPaper` claim an OOXML slot so it survives a deck round-trip?
  No — all 12 slots are claimed and the soul owns the paper tint at author
  time (D-026: the engine renders what the soul supplies). Deferred unless a
  real read-back-the-paper-token case appears (note in V2 backlog if so).
- Soul/composer auto-applying `Background = {BackgroundColor, ColorPaper}` on
  every light slide is **Deckard's product half** (D-059) — out of scope here.
