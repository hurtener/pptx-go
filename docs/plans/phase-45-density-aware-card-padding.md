# Phase 45 — density-aware card padding

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §11.2 (Card), §7 (theme tokens)
**Deps:** Phase 14 (Card chrome), Phase 39 (cardHeaderBottom, D-070)
**Status:** Done

---

## 1. Goal

Add an additive `Card.PaddingScale` so a caller can tighten (or loosen) a card's
interior inset on a finer scale than the SM/MD/LG enum — letting dense cards
reclaim interior space — floored at a pinned minimum, byte-identical by default.

## 2. Why now

R10.7 is the next Wave-10 unit. It fixes the recreation's dense cards that waste
interior space with generous fixed padding (DECKARD R10.7 gap). It reuses the
`cardPadding` token resolver and the basis-point-multiplier pattern from R10.5.

## 3. RFC sections implemented

- `RFC §11.2` — the Card gains a density-aware padding control.
- `RFC §7` — the padding resolves through theme spacing tokens (no literals); the
  scale and floor are token-bound.

## 4. Brief findings incorporated

- `docs/research/28-density-aware-card-padding.md` — *use `Card.PaddingScale`, a
  basis-point multiplier on the size-resolved padding* → added; resolves through
  `cardPadding(size)` then scales.
- `docs/research/28-density-aware-card-padding.md` — *pinned `padMin =
  ResolveSpace(SpaceXS)`; token-resolved, no literal* → `cardPaddingFor` floors at
  `SpaceXS`.
- `docs/research/28-density-aware-card-padding.md` — *thread via `cardChrome`;
  byte-identical default* → `paddingScale` on `cardChrome`, the three padding sites
  route through `cardPaddingFor`.

## 5. Findings I'm departing from

- The spec offers an auto-tighten-inside-the-fit-pass alternative. This plan ships
  **`Card.PaddingScale`** only. **Departing because** the fit pass (R10.2) is
  stack-level and does not reach inside a card body; `PaddingScale` satisfies the
  requirement as a caller-driven mechanism, and an auto-tighten hook can later
  reuse the `cardPaddingFor` seam. (§4.3.)

## 6. Decisions referenced

- `D-043` — Card additive fields — `PaddingScale` is another additive Card field.
- `D-070` — `cardHeaderBottom` — one of the three padding sites this routes.
- `D-074` — `FontScale` — the basis-point-multiplier + pinned-floor pattern
  mirrored here.
- `D-026` — engine, not product — padding density is an opt-in mechanism.
- **New:** `D-076` — density-aware card padding — filed in this PR.

## 7. Architecture

```text
Card.PaddingScale int                       // basis points; 0/10000 = unchanged
cardChrome.paddingScale int                 // populated from v.PaddingScale
cardPaddingFor(c) = clampMin(cardPadding(c.size) × paddingScale / 10000, SpaceXS)
  // the three sites (cardHeaderColumnW, cardHeaderBottom, renderCardChrome) route through it
```

`cardPadding(size)` stays the base resolver. `PaddingScale` 0 (and 10000) returns
the base unchanged → byte-identical.

## 8. Files added or changed

```text
scene/nodes.go                              # CHANGED — Card.PaddingScale int
scene/render_card.go                        # CHANGED — cardChrome.paddingScale; cardPaddingFor; route 3 sites; renderCard populates it
scene/render_card_padding_test.go           # NEW — tighter scale shrinks inset/grows body; padMin floor; default byte-identical
scripts/smoke/phase-45.sh                   # NEW — phase smoke
docs/research/28-density-aware-card-padding.md # NEW — brief 28
docs/research/INDEX.md                      # CHANGED — register brief 28
docs/plans/phase-45-density-aware-card-padding.md # NEW — this plan
docs/plans/README.md                        # CHANGED — Wave 10 phase index row
docs/decisions.md                           # CHANGED — adds D-076
docs/glossary.md                            # CHANGED — Card PaddingScale term
docs/design/THEME.md                        # CHANGED — card padding scale note
docs/site/catalog/containers.md             # CHANGED — document PaddingScale
skills/compose-a-scene/SKILL.md             # CHANGED — PaddingScale in the Card section
```

## 9. Public API surface

```go
// scene
type Card struct {
    // …
    PaddingScale int // basis-point multiplier on the size-resolved interior
                    // padding (0 or 10000 = unchanged; <10000 tightens, floored
                    // at a pinned minimum; >10000 loosens).
}
```

Additive: zero `PaddingScale` reproduces the current SM/MD/LG output exactly.

## 10. Risks

- **R1 — byte-identical regression.** **Mitigation:** `PaddingScale` 0/10000
  returns the base; the existing card golden/determinism tests pass through
  `cardPaddingFor`; a test asserts default == the SM/MD/LG output.
- **R2 — collapse to zero inset.** **Mitigation:** the `SpaceXS` `padMin` floor; a
  test asserts an extreme scale floors there.
- **R3 — determinism.** **Mitigation:** integer / basis-point math; the existing
  card determinism guard covers it (deck includes scaled cards).

## 11. Acceptance criteria

1. A tighter `PaddingScale` (e.g. 5000) measurably reduces the card's interior
   inset and increases the body height, staying ≥ `padMin`.
2. Default `PaddingScale` (0) is byte-identical to the current SM/MD/LG output.
3. An extreme scale floors the inset at `padMin` (does not collapse to zero).
4. Identical inputs yield identical EMU geometry (deterministic).
5. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-45.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` a tighter scale reduces the inset + grows the body
   (`TestCardPaddingScale_TighterReducesInset`).
3. `OK:` default padding byte-identical (`TestCardPaddingScale_DefaultByteIdentical`).
4. `OK:` an extreme scale floors at padMin (`TestCardPaddingScale_FloorsAtMin`).

## 14. Tests

- **Unit:** `scene` (white-box) — `cardPaddingFor` scale + floor; the body box
  grows with a tighter scale.
- **Render byte-identical:** default `PaddingScale` card vs the unset card.
- **Round-trip golden / Integration / Fuzz / Bench:** none.

## 15. Vocabulary added

- `Card PaddingScale` — the additive `Card.PaddingScale` basis-point multiplier on
  the card's size-resolved interior padding.

## 16. Plan deviations encountered during implementation

- **Auto-tighten-in-fit deferred** (per §5); `Card.PaddingScale` only.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-45.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-076).
- [x] Docs site + THEME.md updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
