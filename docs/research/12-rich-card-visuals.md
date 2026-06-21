# Brief 12 — rich-card-visuals

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-20
**Motivating phase:** Phase 25 — rich card visuals

## 1. Question

`Card` already supports fill, icon, eyebrow, header pill, border, size, and
elevation. Reference "designed" cards add three visuals it cannot express: a
**colored header band** (the top portion of the card a solid accent, the body in
surface below — distinct from a full `Fill`), a **status dot** (a small colored
dot in the top-right corner), and a **watermark** (a large, low-opacity label
drawn behind the card body, e.g. a ghosted `01`). How can the engine add these
additively, deterministically, and byte-identically when unset?

## 2. Prior art surveyed

- **`render_card.go::renderCardChrome`.** Already draws the card background
  (rounded rect), a left accent stripe, and the header row (icon / eyebrow /
  title / right-aligned pill), then returns the body region. It computes the
  header's bottom (`bodyY`) incrementally as it emits. This is the function the
  three visuals slot into; the header band needs that `bodyY` boundary.
- **`pptx` color + shape primitives.** `TokenColorAlpha(role, alpha)` already
  emits an `<a:alpha>` child, and a text run's color goes through the same
  `srgbFrom` path (`fill.go` → `text_layout.go`), so **a run can be drawn at low
  opacity** — exactly what a ghosted watermark needs. `ShapeEllipse` exists for
  the status dot. No new builder primitive is required.
- **`ColorRole` zero value.** `ColorRole` is `iota`-based with `ColorCanvas == 0`
  — a *real* color, not "unset". So an optional color field typed plainly as
  `ColorRole` cannot express "omit" via its zero value, which the acceptance
  ("each zero-value omits its element") requires.
- **D-043 (additive Card expansion).** The Card has grown additively before
  (border, size, layout, elevation), each with a zero value that reproduces the
  prior render. R4 follows the same discipline.
- **D-026 (engine, not product).** The caller supplies the band color, the dot
  color, and the watermark text; the engine renders them faithfully and picks
  only mechanical defaults (the watermark's faint opacity), never taste.

## 3. Findings

- **Optional colors need an explicit "unset" — use `*ColorRole`.** Because
  `ColorRole`'s zero value is a valid color (`ColorCanvas`), `HeaderFill` and
  `StatusDot` must be pointers (`*ColorRole`): `nil` omits the element, a
  non-nil role draws it. This is the only representation that satisfies "zero
  value omits + byte-identical when unset" without inventing a sentinel role.
  `Watermark` is a `string`, whose `""` zero value already means "omit".
- **The header band needs the header's bottom boundary.** The band runs from the
  card top to where the body begins (`bodyY`). `renderCardChrome` already
  computes `bodyY`, but only after emitting the header shapes — and the band must
  be drawn *behind* that text. The clean fix: a small pure helper that computes
  `bodyY` from the chrome inputs, used to size the band drawn right after the
  background (before the header text). Sharing the header-row height constants
  between the helper and the emit code keeps them from drifting.
- **Byte-identity falls out of conditional emission.** Each visual is emitted
  only when its field is set; with all three unset, `renderCardChrome` emits the
  exact same shapes in the same order as today. Extracting the existing header
  literals (`0.45"`, `0.26"`, `0.40"`) into named constants of identical value
  is value-preserving, so the unset path stays byte-for-byte.
- **The watermark is true low-opacity text.** Drawing the watermark inside
  `renderCardChrome` (before the body content the caller stacks afterward) puts
  it *behind* the body; rendering it with `TokenColorAlpha` at a low pinned alpha
  gives the ghosted look while staying token-bound (P2). No new token.
- **Determinism is free.** All three are deterministic native shapes (pinned EMU
  geometry, pinned alpha); they register no media, so cards stay parallel-safe.

## 4. Recommendations

1. Add `HeaderFill *ColorRole`, `StatusDot *ColorRole`, `Watermark string` to
   `Card` (and the internal `cardChrome`). All zero values inert.
2. Extract the header-row height literals into shared constants; add a pure
   `cardHeaderBottom` helper; draw the header band (a rounded rect in
   `HeaderFill`, top of card → `bodyY`) right after the background.
3. Draw the status dot as a small `ShapeEllipse` in the top-right corner (inset
   by the card padding) and the watermark as a large `TokenColorAlpha`
   `TypeDisplay` run anchored in the body region, behind the body content.
4. Apply to `Card` only — `CardSection` builds its own chrome without these
   fields, so it is untouched.

## 5. Open questions

- **Pill + dot both occupy the top-right.** The header pill and the status dot
  share the top-right corner; combining them overlaps. The reference uses one or
  the other, so this is left to the caller; a future layout could offset the dot
  when a pill is present. Deferred.
- **Watermark color/size knob.** R4 gives the watermark only a `string`; the
  engine picks a faint accent token and `TypeDisplay`. Exposing a watermark color
  or scale is a plausible later field; deferred (it adds IR surface and the
  reference look needs neither).
- **Header-band corner geometry.** The band is a rounded rect matching the
  card's radius; its bottom corners are rounded (small notches reveal the body
  fill). Per-corner radii (flat band bottom) would need a builder primitive;
  deferred as cosmetic.
- **Recursive interaction with grow-to-fit (R2).** A grown card gets a taller
  box; the band tracks the header (fixed), the body region grows below it — they
  compose without special handling.
