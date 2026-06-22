# Phase 42 — card body vertical distribution

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §11.2 (Card), §10 (layout engine)
**Deps:** Phase 13 (`alignedStackIn`), Phase 14 (Card chrome), Phase 40 (VAlignFit)
**Status:** Done

---

## 1. Goal

Add an opt-in `Card.BodyVAlign` so a card's body can center, bottom-pin, justify,
or fill within the card body region instead of always floating top-anchored —
eliminating the dead space below a card's content.

## 2. Why now

R10.4 is the next Wave-10 unit. It is the fix for the recreation's
Vision/Mission, path, and pricing cards that leave ~40–60% of the card blank
(DECKARD R10.4 gap). The vertical-distribution engine (`alignedStackIn`) already
exists from Phase 13 and now also carries VAlignFit (D-071); R10.4 simply lets the
card body opt into it.

## 3. RFC sections implemented

- `RFC §11.2` — the Card body gains opt-in vertical distribution.
- `RFC §10` — reuses the body-stack alignment engine on a container's body box (a
  partial; CardSection is deferred).

## 4. Brief findings incorporated

- `docs/research/25-card-body-vertical-distribution.md` — *the mechanism already
  exists (`alignedStackIn`); the card body just needs to call it* → `renderCard`'s
  vertical body routes through `alignedStackIn` with `{Vertical: BodyVAlign}`.
- `docs/research/25-card-body-vertical-distribution.md` — *byte-identity is
  provable for {VAlignTop, HAlignLeft} and already documented* → `BodyVAlign=Top`
  (zero) reproduces the `stackIn` output exactly.
- `docs/research/25-card-body-vertical-distribution.md` — *horizontal stays left;
  free composition with VAlignFit/Fill* → `Horizontal: HAlignLeft`; any `VAlign`
  is accepted for `BodyVAlign`.

## 5. Findings I'm departing from

- `docs/research/25-card-body-vertical-distribution.md` notes `CardSection` also
  stacks via `stackIn`. This plan scopes `BodyVAlign` to **Card** only.
  **Departing because** the R10.4 gap and acceptance are about `Card` price/list
  bodies; `CardSection` bodies are containers that already grow under slide-level
  fill, and a `CardSection.BodyVAlign` is a separable, lower-value follow-up.
  (§4.3, see §16.)

## 6. Decisions referenced

- `D-043` — Card chrome / additive fields — `BodyVAlign` is another additive Card
  field (zero value byte-identical).
- `D-071` — VAlignFit — composes for free as a `BodyVAlign` value.
- `D-026` — engine, not product — vertical distribution is an opt-in mechanism.
- **New:** `D-073` — card body vertical distribution — filed in this PR.

## 7. Architecture

`renderCard` already computes the card body box (below the wrapped header, inside
the padding) and, for the vertical layout, stacks it via `r.stackIn`. This phase
swaps that single call for `r.alignedStackIn(body, v.Body, slideID, Alignment{
Vertical: v.BodyVAlign})`. The `BodyHorizontal` branch is untouched.

```text
renderCard:
  body = renderCardChrome(...)            // card body box
  if BodyLayout == BodyHorizontal: columns(...)   // unchanged
  else:
    for pl in alignedStackIn(body, Body, {Vertical: BodyVAlign}):   // was stackIn
        renderNode(pl.box, pl.node)
```

Byte-identity (zero `BodyVAlign` == `VAlignTop`): `alignedStackIn` with
`{VAlignTop, HAlignLeft}` emits the same per-node boxes as `stackIn` and warns on
the algebraically identical condition (`totalH > box.H` ⟺ last-bottom >
box.Bottom()).

## 8. Files added or changed

```text
scene/nodes.go                                  # CHANGED — Card.BodyVAlign VAlign
scene/render_card.go                            # CHANGED — vertical body via alignedStackIn
scene/render.go                                 # CHANGED — alignedStackIn godoc (no longer sole caller)
scene/render_card_body_test.go                  # NEW — bottom pins last node; top byte-identical; determinism
scripts/smoke/phase-42.sh                       # NEW — phase smoke
docs/research/25-card-body-vertical-distribution.md # NEW — brief 25
docs/research/INDEX.md                          # CHANGED — register brief 25
docs/plans/phase-42-card-body-vertical-distribution.md # NEW — this plan
docs/plans/README.md                            # CHANGED — Wave 10 phase index row
docs/decisions.md                               # CHANGED — adds D-073
docs/glossary.md                                # CHANGED — Card BodyVAlign term
docs/site/catalog/containers.md                 # CHANGED — document BodyVAlign
skills/compose-a-scene/SKILL.md                 # CHANGED — BodyVAlign in the Card section
```

## 9. Public API surface

```go
// scene
type Card struct {
    // …
    BodyVAlign VAlign // vertical distribution of the card body within the card
                      // body region (Top/Center/Bottom/Justify/Fill/Fit).
                      // Zero (VAlignTop) = top-anchored, byte-identical to today.
}
```

Additive field; the zero value reproduces the current card layout exactly.

## 10. Risks

- **R1 — byte-identical regression for Top.** **Mitigation:** the swap is to the
  documented byte-identical zero-Alignment path; a test asserts a `BodyVAlign=Top`
  card is byte-identical to the same card rendered before (golden bytes), and the
  existing card determinism/render tests still pass.
- **R2 — determinism.** **Mitigation:** `alignedStackIn` is integer-EMU; a
  determinism guard renders a `BodyVAlign` card deck at 1 and 8 workers.

## 11. Acceptance criteria

1. A card with `BodyVAlign=VAlignBottom` places its last body node's bottom at the
   card body bottom (within padding).
2. `BodyVAlign=VAlignJustify` spreads the inter-item gaps to fill the body region.
3. `BodyVAlign=VAlignTop` (zero) is byte-identical to the current card output.
4. Identical inputs yield identical EMU geometry (deterministic at any worker
   count).
5. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package (no override) |

## 13. Smoke check

`scripts/smoke/phase-42.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `BodyVAlign=Bottom` pins the last body node
   (`TestCardBodyVAlign_BottomPinsLastNode`).
3. `OK:` `BodyVAlign=Top` byte-identical
   (`TestCardBodyVAlign_TopByteIdentical`).
4. `OK:` card body valign render stays deterministic
   (`TestCardBodyVAlign_Deterministic`).

## 14. Tests

- **Unit:** `scene` (white-box) — bottom pins last body node; justify expands
  gaps; top byte-identical to `stackIn` placements.
- **Render byte-identical:** a `BodyVAlign=Top` card vs the pre-change render.
- **Round-trip golden:** n/a (scene layout change).
- **Integration / Fuzz / Bench:** none.

## 15. Vocabulary added

- `Card BodyVAlign` — the opt-in `Card.BodyVAlign` field selecting vertical
  distribution of the card body within the card body region.

## 16. Plan deviations encountered during implementation

- **CardSection deferred.** `BodyVAlign` is added to `Card` only; the analogous
  `CardSection` body keeps top-anchored `stackIn`. (§4.3.)

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-42.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-073).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
