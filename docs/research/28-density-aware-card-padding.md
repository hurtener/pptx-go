# Brief 28 — density-aware-card-padding

**Subsystem:** scene — Layer 2 renderer
**Authored:** 2026-06-22
**Motivating phase:** Phase 45 — density-aware card padding (R10.7, MED · engine)

## 1. Question

`cardPadding` maps the 3-value `CardSize` enum to fixed `SpaceSM/MD/XL`, so a
dense card carries the same generous interior inset as a sparse one — wasting
interior space and pushing content out (the recreation's dense cards use generous
fixed padding where the reference packs tighter). How can a caller tune a card's
interior padding on a finer scale — or auto-tighten it under overflow — **without
literals** and byte-identical by default?

## 2. Prior art surveyed

- `scene/render_card.go` `cardPadding(size) → ResolveSpace(SpaceSM|MD|XL)`; called
  at three sites (`cardHeaderColumnW`, `cardHeaderBottom`, `renderCardChrome`), all
  from a `cardChrome` value carrying `size`.
- `scene/render_card.go` `cardChrome` — the internal struct the three padding sites
  read; `renderCard` builds it from the `Card` fields.
- Phase 43 (R10.5, D-074) `FontScale` and Phase 41 (R10.3, D-072) — the
  basis-point multiplier + pinned-floor pattern this brief reuses.
- `pptx/theme.go` spacing tokens `SpaceXS=Pt(2)` … `SpaceXL=Pt(24)`;
  `ResolveSpace` maps role → EMU.
- DECKARD R10.7 spec: add `Card.PaddingScale` (basis-point multiplier on the
  resolved size padding, default 10000 = unchanged) **or** an auto-tighten step
  inside the fit pass toward a pinned `padMin`; resolve through theme spacing
  tokens (no literals); zero/default byte-identical; deterministic.

## 3. Findings

- The cleanest mechanism is the spec's first option — **`Card.PaddingScale`**, a
  basis-point multiplier on the size-resolved padding. It is caller-driven (no
  coupling to the fit pass), resolves through the existing `ResolveSpace` token
  (so P2 holds — a theme swap still re-skins the base padding, then this scales
  it), and is byte-identical at the default.
- **Pinned `padMin` floor.** A tightened card must not collapse its inset to zero;
  floor the scaled padding at `ResolveSpace(SpaceXS)` (the smallest spacing token,
  Pt 2) so content never touches the card edge. Token-resolved — no literal.
- **Threading.** Add `paddingScale int` to `cardChrome` and a `cardPaddingFor(c)`
  method that scales `cardPadding(c.size)` and floors it; route the three sites
  through it. `cardPadding(size)` stays as the base resolver (and keeps its
  existing white-box test). `CardSection` builds a bare `cardChrome` (scale 0 →
  unchanged), so it is unaffected.
- **Byte-identical default.** `PaddingScale` zero (and 10000) returns the base
  padding unchanged; the existing card golden/determinism tests pass through the
  new method.
- **Auto-tighten-in-fit deferred.** The spec's alternative — reducing card padding
  inside the fit-to-region pass — operates at a different layer (the fit pass is
  stack-level, R10.2; it does not reach inside a card's body). `PaddingScale`
  satisfies the requirement as a caller-driven mechanism; an auto-tighten hook can
  layer on later (it would call the same `cardPaddingFor` seam).
- **Effect.** A smaller `PaddingScale` reduces the interior inset on all four
  sides, so the card body box (computed below the header, inside the padding)
  grows — measurably more body height, as the acceptance requires.

## 4. Recommendations

1. Add `Card.PaddingScale int` (basis points; 0 or 10000 = unchanged).
2. Add `paddingScale int` to `cardChrome`; populate from `v.PaddingScale` in
   `renderCard`.
3. Add `cardPaddingFor(c cardChrome) → cardPadding(c.size)` scaled by
   `paddingScale` (when > 0 and ≠ 10000), floored at `ResolveSpace(SpaceXS)`; route
   the three padding sites through it.
4. Tests: a tighter scale reduces the inset and increases body height; the padMin
   floor caps an extreme scale; default (0) is byte-identical; determinism guard;
   smoke `phase-45.sh`.

## 5. Open questions

- **Auto-tighten under overflow** — deferred (the fit pass is stack-level; would
  reuse `cardPaddingFor`).
- **`CardSection.PaddingScale`** — out of scope; `CardSection` bodies are
  containers, and the gap is about `Card` density.
