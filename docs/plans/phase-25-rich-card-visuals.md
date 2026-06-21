# Phase 25 — rich card visuals

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §11.2 (Card chrome)
**Deps:** Phase 14 (Card / CardSection), Phase 13 (token-alpha builder). External:
none.
**Status:** In progress

---

## 1. Goal

Add three additive `Card` visuals — a colored header band, a top-right status
dot, and a ghosted watermark label behind the body — so a card can match the
reference "designed" look, each opt-in and byte-identical when unset.

## 2. Why now

Fourth unit of **Wave 8 — post-V1 engine extensions**
(`DECKARD-PRODUCT-REQUIREMENTS.md` R4, MEDIUM), the last MEDIUM before the LOW
tail (R5–R7), picked up after R1–R3 per the one-requirement-per-PR cadence.
`Card` already carries fill/icon/eyebrow/pill/border/size/elevation; the
reference cards add a banded header, a status dot, and a watermark the IR cannot
express. R4 adds those three fields; the caller drives the colors and label.

## 3. RFC sections implemented

- `RFC §11.2` — extends the native Card chrome (rounded rect + accent stripe +
  header row) with a header band, a status dot, and a watermark, all composed
  from the existing builder primitives (no new OOXML capability — P1).

## 4. Brief findings incorporated

- `docs/research/12-rich-card-visuals.md` — *optional colors need an explicit
  unset → `*ColorRole`* → `HeaderFill` / `StatusDot` are `*ColorRole` (nil =
  omit); `Watermark` is a `string` ("" = omit).
- `docs/research/12-rich-card-visuals.md` — *the header band needs the header
  bottom* → a pure `cardHeaderBottom` helper sizes the band; the header-row
  height literals are extracted to shared constants so helper and emit code stay
  in sync.
- `docs/research/12-rich-card-visuals.md` — *the watermark is true low-opacity
  text* → drawn with `TokenColorAlpha` at a pinned alpha, inside the chrome so it
  sits behind the body content.
- `docs/research/12-rich-card-visuals.md` — *byte-identity falls out of
  conditional emission* → each visual emits only when set; const extraction is
  value-preserving, so the all-unset path is byte-for-byte.
- `docs/research/12-rich-card-visuals.md` — *apply to Card only* → `CardSection`
  builds its own chrome without the new fields and is untouched.

## 5. Findings I'm departing from

None from the brief. One **deviation from the requirement's field types** (§4.3):
`DECKARD-PRODUCT-REQUIREMENTS.md` R4 writes `HeaderFill ColorRole` and
`StatusDot ColorRole` (value types). Because `ColorRole`'s zero value is
`ColorCanvas` (a real color, not "unset"), a value-typed field cannot satisfy the
same requirement's acceptance "each zero-value omits its element". The fields
therefore ship as **`*ColorRole`** (nil = omit), which honors the binding
acceptance. Documented here and in D-054.

## 6. Decisions referenced

- `D-043` — *Additive Card expansion.* R4 is the next additive Card growth; every
  new field's zero value reproduces the prior render.
- `D-026` — *Engine, not product.* The caller supplies the band/dot colors and
  the watermark label; the engine renders them and picks only the watermark's
  mechanical faint opacity.
- `D-012` / `D-030` — token color resolution; the watermark stays token-bound via
  `TokenColorAlpha` (P2).
- This plan files **D-054 — rich card visuals** in `docs/decisions.md`.

## 7. Architecture

```text
scene/nodes.go      Card: + HeaderFill *ColorRole, + StatusDot *ColorRole, + Watermark string

scene/render_card.go
  cardChrome:        + headerFill *ColorRole, + statusDot *ColorRole, + watermark string
  consts:            cardIconSz / cardEyebrowRowH / cardTitleRowH (extracted, value-identical)
                     + cardStatusDotSz, cardWatermarkAlpha
  cardHeaderBottom:  pure helper → bodyY (mirrors the header vertical advance)   NEW
  renderCardChrome:  after background, draw header band (rounded rect, HeaderFill,
                     top → bodyY) when set; near the end, draw the status dot
                     (ellipse, top-right) and the watermark (large TokenColorAlpha
                     run anchored in the body region) when set.
  renderCard:        pass v.HeaderFill / v.StatusDot / v.Watermark into cardChrome.
```

Z-order within the card: background → header band → accent stripe → header
content → status dot → watermark → (body content, drawn by renderCard after).
The watermark is the last chrome shape, so it sits behind the body content the
caller stacks next.

## 8. Files added or changed

```text
scene/nodes.go                        # CHANGED — three additive Card fields
scene/render_card.go                  # CHANGED — cardChrome fields, consts, cardHeaderBottom, band/dot/watermark
scene/render_card_test.go             # CHANGED — render-all-three, omit-when-unset, byte-identical, determinism
scripts/smoke/phase-25.sh             # NEW — phase smoke
docs/research/12-rich-card-visuals.md # NEW — informing brief
docs/research/INDEX.md                # CHANGED — registers brief 12
docs/plans/phase-25-rich-card-visuals.md # NEW — this plan
docs/plans/README.md                  # CHANGED — adds Phase 25 to Wave 8
docs/decisions.md                     # CHANGED — adds D-054
docs/glossary.md                      # CHANGED — adds "Header band", "Status dot", "Watermark (card)"
docs/site/catalog/containers.md       # CHANGED — Card field docs (§19)
skills/compose-a-scene/SKILL.md       # CHANGED — Card field list (§19)
```

## 9. Public API surface

```go
// scene (nodes.go) — three additive Card fields; no signature changes.
type Card struct {
    // ... existing ...
    HeaderFill *ColorRole // banded header region color (body keeps Fill); nil = no band
    StatusDot  *ColorRole // small status dot, top-right corner; nil = no dot
    Watermark  string     // large, low-opacity label behind the body; "" = none
}
```

New public scene surface (Card fields) ⇒ a smoke check lands in this PR
(§4.2/§13). No new builder API, no new scene IR node, no new theme token
(P2 — the visuals reuse `ColorRole` tokens and `TokenColorAlpha`).

## 10. Risks

- **R1 — byte-identity regression for existing cards.** **Mitigation:** the new
  visuals emit only when their field is set; the const extraction is
  value-identical. The existing `TestRenderCard` / `TestCardParallel` plus a new
  all-unset byte-identity test guard it.
- **R2 — header band misaligned with the body.** The band height comes from a
  helper that duplicates the header advance. **Mitigation:** the helper and the
  emit code share the extracted height constants; a test asserts the band's
  bottom is at/above the first body shape (no overlap of band over body).
- **R3 — watermark not actually faint / not behind body.** **Mitigation:** it
  uses `TokenColorAlpha` (verified to emit `<a:alpha>`) and is drawn before the
  body content; a test asserts the emitted `<a:alpha>` and that the run text is
  present.

## 11. Acceptance criteria

1. A `Card` with `HeaderFill` + `StatusDot` + `Watermark` renders all three: a
   header-band rounded rect in the header-fill color, an ellipse in the top-right,
   and a large low-opacity (`<a:alpha>`) run carrying the watermark text.
2. Each visual is omitted when its field is unset (no band rect / no ellipse / no
   watermark run).
3. A `Card` with none of the three set renders **byte-identical** to today.
4. A deck of rich cards renders byte-identical across 1 vs N workers
   (determinism holds).
5. `make coverage` shows `scene` ≥ its band; `make preflight` passes.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry; new branches covered by the card
tests.

## 13. Smoke check

`scripts/smoke/phase-25.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` header band + status dot + watermark all render (criterion 1).
3. `OK:` each visual omitted when unset (criterion 2).
4. `OK:` a bare card is byte-identical (criterion 3).
5. `OK:` rich-card render is deterministic across workers (criterion 4).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — `cardHeaderBottom`; black-box render assertions on the
  emitted slide XML (band rect, ellipse `prst="ellipse"`, watermark `<a:alpha>` +
  text, and their absence when unset); byte-identity + parallel determinism.
- **Round-trip golden:** N/A — no builder primitive / scene node added.
- **Integration** (`test/integration/`): no — internal to `scene` Card chrome.
- **Fuzz:** no — no parse/decode surface.
- **Benchmark:** no.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Header band` — a `Card`'s optional colored top region (`HeaderFill`), the
  header in an accent color with the body in `Fill` below.
- `Status dot` — a `Card`'s optional small colored dot (`StatusDot`) in the
  top-right corner.
- `Watermark (card)` — a `Card`'s optional large, low-opacity label
  (`Watermark`) drawn behind the body content.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-25.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entry D-054 added.
- [ ] Docs site updated for the Card fields (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
