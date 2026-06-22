# Phase 38 — weight-aware font embedding

**Subsystem:** pptx — Layer 1 builder (font embedding)
**RFC sections:** §7.6
**Deps:** Phase 35 (font-embedding pass, D-065).
**Status:** Done

---

## 1. Goal

The embedding pass ships the **correct physical weight file** for each weight a
deck uses: a soul's medium (500) regular role embeds the medium file, not a
synthetic 400.

## 2. Why now

R9.8 is the final R9 engine unit (R9.12 subsetting → V2). With the embedding pass
(D-065) and italic-aware fallback (D-067) in place, weight-accuracy closes the
font cluster before the Wave 9 §17 checkpoint.

## 3. RFC sections implemented

- `RFC §7.6` — font embedding. Makes the used-face collection weight-aware; no new
  wire format.

## 4. Brief findings incorporated

- `docs/research/21-weight-aware-embedding.md` — *track the resolved weight per
  run* → `XTextProperties.Weight` (`xml:"-"`), set in `toProps`.
- `21-weight-aware-embedding.md` — *key the used-face set on weight* →
  `FontFace.Weight`, populated by `UsedFontFaces` (inferred from bold when 0).
- `21-weight-aware-embedding.md` — *embed the actual weighted file, bucketed
  deterministically* → `autoEmbedFonts` groups by OOXML bucket and embeds the
  nearest-nominal weight.
- `21-weight-aware-embedding.md` — *honor the 4-bucket limit, no orphan parts* →
  one file per bucket; multi-file-per-bucket deferred (D-068).

## 5. Findings I'm departing from

None from the brief. (The brief itself defers R9.8's literal "embeds three
distinct files" acceptance — see §6/D-068.)

## 6. Decisions referenced

- `D-065` — the embedding pass this phase makes weight-aware.
- `D-067` — the resolver-callback rewrite shape (extensible to weight).
- `D-059` — engine half of `both`.
- `D-020` — the no-repair-prompt guarantee (why orphan weight parts are deferred).
- `D-026` — product (rasterizer) behavior is the caller's.
- New: `D-068` — weight-aware embedding + the one-file-per-bucket deviation.

## 7. Architecture

```text
toProps: p.Weight = effective role weight (xml:"-", never serialized)
UsedFontFaces: FontFace{Typeface, Bold, Italic, Weight}  (infer 700/400 when 0)
autoEmbedFonts:
  collect distinct (family, weight, italic)
  group by bucket (family, weight>=600, italic)
  per bucket: pick weight nearest nominal (400/700; tie→lower) → EmbedFont(actual)
              (coalesced weights logged; one file per bucket)
```

## 8. Files added or changed

```text
internal/ooxml/slide/slide_types.go        # CHANGED — XTextProperties.Weight (xml:"-")
internal/ooxml/slide/fonts.go              # CHANGED — FontFace.Weight; UsedFontFaces weight
pptx/text_layout.go                        # CHANGED — toProps sets p.Weight
pptx/fonts.go                              # CHANGED — autoEmbedFonts weight-aware buckets
pptx/fonts_test.go                         # CHANGED — weight-aware tests
internal/ooxml/slide/fonts_test.go         # CHANGED — weight in UsedFontFaces tests
scripts/smoke/phase-38.sh                  # NEW — phase smoke
docs/research/21-weight-aware-embedding.md # NEW — brief
docs/research/INDEX.md                      # CHANGED — register brief 21
docs/plans/phase-38-weight-aware-embedding.md # NEW — this plan
docs/plans/README.md                        # CHANGED — Wave 9 phase index
docs/decisions.md                           # CHANGED — adds D-068
docs/glossary.md                            # CHANGED — embedding pass weight-aware
docs/design/THEME.md                        # CHANGED — embedding note weight-aware
docs/site/guide/builder.md                  # CHANGED — embedding note weight-aware
skills/scaffold-a-presentation/SKILL.md     # CHANGED — embedding note weight-aware
```

## 9. Public API surface

No new public symbol. `slide.XTextProperties.Weight` and `slide.FontFace.Weight`
are internal. Behavior of `WithFontEmbedding` is now weight-aware.

## 10. Risks

- **R1 — `XTextProperties.Weight` leaking into the wire / round-trip.**
  **Mitigation:** `xml:"-"` — never serialized or parsed; byte-identical and a
  parsed deck infers 700/400 from the bold bit.
- **R2 — Non-determinism from bucket map iteration.** **Mitigation:** buckets are
  sorted by `(family, bold, italic)`; nearest-nominal selection is a pure integer
  function.
- **R3 — Orphan unreferenced font parts.** **Mitigation:** one file per OOXML
  bucket — no parts outside `embeddedFontLst` (D-068); the multi-file rasterizer
  case is the caller's.

## 11. Acceptance criteria

1. A medium (500) role embeds the medium file (the provider is asked for the
   resolved weight, not a synthetic 400).
2. A single-weight deck embeds one file.
3. Two weights colliding on one bucket coalesce to the nearest-nominal winner
   (one file), logged.
4. A deck with embedding off / no `FontSource` is byte-identical; deterministic.
5. `make coverage` shows touched packages ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | band per `coverage.json` | low-by-design; `0 failing` is the gate |
| `internal/ooxml/slide` | 85% | codec band (FontFace weight) |

## 13. Smoke check

`scripts/smoke/phase-38.sh`: embeds the actual resolved weight file; single-weight
one file; colliding weights coalesce per bucket; codec carries per-run weight.

## 14. Tests

- **Unit:** `pptx` (medium-file embed, single-weight, coalesce + log),
  `internal/ooxml/slide` (`UsedFontFaces` weight + inference).
- **Round-trip golden:** `Weight` is `xml:"-"` — no wire change, goldens
  unaffected; the resolved family/cut round-trips as before.
- **Integration / Fuzz / Benchmark:** none.

## 15. Vocabulary added

- (none new) — the `Font-embedding pass` glossary entry gains the weight-aware
  note.

## 16. Plan deviations encountered during implementation

- Per D-068, the engine embeds **one file per OOXML bucket**, not one per numeric
  weight (R9.8's literal "three distinct files" acceptance) — multi-file-per-bucket
  would create unreferenced font parts, risking the no-repair-prompt guarantee for
  zero PowerPoint benefit (4-cut limit); that rasterizer case is the caller's
  (D-026).

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-38.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-068).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated.
