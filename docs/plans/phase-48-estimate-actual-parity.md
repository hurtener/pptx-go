# Phase 48 — estimate/actual parity

**Subsystem:** scene — Layer 2 renderer
**RFC sections:** §10.2 (content-aware metrics)
**Deps:** Phase 22 (content-aware `preferredHeight`), Phase 39 (R10.1 wrapped header, D-070), Phase 41 (Bento span geometry, D-072)
**Status:** Done

---

## 1. Goal

Align `preferredHeight`'s slot estimates with what the composers actually emit —
the card chrome estimate becomes wrapped-header-aware and the bento estimate uses
each cell's actual span width — so overflow detection and the fit pass are
trustworthy, while single-line / span-1 cases stay byte-identical.

## 2. Why now

R10.10 is the final Wave-10 engine unit and closes the `cardChromeEst` parity
explicitly deferred by R10.1 (D-070). With the fit pass (R10.2), weighted bento
rows (R10.3), and the other density controls in place, accurate estimates make the
overflow warning and `VAlignFit` operate on the right numbers (DECKARD R10.10 gap).

## 3. RFC sections implemented

- `RFC §10.2` — the deterministic height estimators match the composed geometry
  (wrapped headers, span widths).

## 4. Brief findings incorporated

- `docs/research/31-estimate-actual-parity.md` — *keep the baseline and add the
  wrapped-header increment (Phase-22 shape)* → `cardChromeEst + extraHeaderLines`.
- `docs/research/31-estimate-actual-parity.md` — *measure each bento cell at its
  actual span width* → `span·unitW + estGap·(span−1)`.
- `docs/research/31-estimate-actual-parity.md` — *refactor the card-header helpers
  to theme-taking free functions + method wrappers* → `cardPaddingBase` /
  `cardPaddingScaled` / `cardHeaderColumnWOf` + wrappers.

## 5. Findings I'm departing from

- The spec also asks the card body inset estimate to match `cardPadding`. This plan
  keeps `cardBodyInsetEst` pinned. **Departing because** matching it (0.20"/side →
  e.g. 0.11" for MD) would change the body wrap count → single-line output, and the
  chrome + span-width fixes already bring the representative nodes within one
  line-height. Exact inset parity is a future refinement. (§4.3.)

## 6. Decisions referenced

- `D-070` — content-aware card header height — this closes its deferred
  `cardChromeEst` parity, sharing the same row-height constants.
- `D-072` — content-weighted bento rows / `cellWidth` — the span-width geometry
  mirrored in the estimate.
- `D-076` — `cardPaddingFor` — the header column width uses the scaled padding.
- `D-026` — engine, not product — accurate estimates are a mechanism, not taste.
- **New:** `D-079` — estimate/actual parity — filed in this PR.

## 7. Architecture

```text
render_card.go: cardPaddingBase(theme,size) / cardPaddingScaled(theme,c) /
                cardHeaderColumnWOf(theme,box,c)   // free; methods delegate

preferredHeight Card/CardSection:
  headerW = cardHeaderColumnWOf(theme, {W:avail}, chromeOf(node))
  extra   = (wrappedLines(eyebrow,…,headerW)−1)·cardEyebrowRowH
          + (wrappedLines(header,…,headerW)−1)·cardTitleRowH      // 0 single-line
  return cardChromeEst + extra + nodesHeight(body, avail−2·cardBodyInsetEst) + estGap

preferredHeight Bento:
  spanW(cell) = span·unitW + estGap·(span−1)                       // span-1 == unitW
  maxCell over cells at spanW; nRows·maxCell + estGap·(nRows−1)
```

## 8. Files added or changed

```text
scene/render_card.go                 # CHANGED — cardPadding*/cardHeaderColumnW free funcs + method wrappers
scene/render.go                      # CHANGED — Card/CardSection chrome increment; Bento span-width estimate
scene/render_height_test.go          # CHANGED — wrapped-header card est grows; span-width bento est; single-line/span-1 byte-identical
scripts/smoke/phase-48.sh            # NEW — phase smoke
docs/research/31-estimate-actual-parity.md # NEW — brief 31
docs/research/INDEX.md               # CHANGED — register brief 31
docs/plans/phase-48-estimate-actual-parity.md # NEW — this plan
docs/plans/README.md                 # CHANGED — Wave 10 phase index row
docs/decisions.md                    # CHANGED — adds D-079
docs/glossary.md                     # CHANGED — estimate/actual parity note on preferredHeight
docs/site/guide/scene.md             # CHANGED — overflow-accuracy note (if user-facing)
```

No new public API — this is an internal estimator change (the only user-visible
effect is a more accurate overflow warning and a taller slot for wrapped-header
cards / a tighter slot for wide-span bento cells).

## 9. Public API surface

None. (Internal `preferredHeight` accuracy; `cardChromeEst`/`estGap` stay pinned
constants.)

## 10. Risks

- **R1 — single-line regression.** **Mitigation:** the increment is 0 for
  single-line and the bento span width equals `unitW` for span-1, so those cases
  are byte-identical; tests assert it, and the existing card/bento determinism +
  height tests pass.
- **R2 — determinism.** **Mitigation:** integer EMU; `wrappedLines` is the existing
  deterministic estimator; the existing determinism guards cover it.
- **R3 — over/under estimate.** **Mitigation:** the increment only *grows* the card
  slot (never shrinks single-line); the bento span fix only shrinks an
  over-counted wide-span cell — both safe directions. A test asserts the overflow
  warning fires for a genuinely-overflowing wrapped-header card.

## 11. Acceptance criteria

1. A multi-line-header card's `preferredHeight` exceeds the single-line card's by
   the wrapped-header increment (the slot accounts for the wrapped header).
2. A wide-span bento cell's `preferredHeight` is ≤ the same content measured at the
   unit width (no longer over-counted), and a span-1 bento is byte-identical.
3. A single-line-header card is byte-identical to the pre-change estimate.
4. The `content overflows its region` warning fires iff the composed content
   actually exceeds the region (verified on a wrapped-header card).
5. Identical inputs yield identical EMU geometry (deterministic).
6. `make coverage` keeps `scene` ≥ its band.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-48.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` wrapped-header card estimate grows
   (`TestPreferredHeight_WrappedCardGrows`).
3. `OK:` wide-span bento estimate ≤ unit-width + span-1 byte-identical
   (`TestPreferredHeight_BentoSpanWidth`).
4. `OK:` single-line card estimate byte-identical
   (`TestPreferredHeight_SingleLineCardUnchanged`).

## 14. Tests

- **Unit:** `scene` (white-box) — the card chrome increment, the bento span-width
  estimate, single-line/span-1 byte-identity, and the overflow warning on a
  genuinely-overflowing wrapped-header card.

## 15. Vocabulary added

- *(none new — refines `preferredHeight` / the `Content-aware text height`
  glossary entry.)*

## 16. Plan deviations encountered during implementation

- **Card body inset parity deferred** (per §5); chrome + span-width fixes deliver
  the within-one-line-height accuracy.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-48.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-079).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s): no user-facing API change (overflow-accuracy only).
