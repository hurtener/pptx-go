# Phase 37 — italic-aware font fallback (emphasis-as-italic-display)

**Subsystem:** pptx — Layer 1 builder (theme/typography + font embedding)
**RFC sections:** §7.6
**Deps:** Phase 33 (display family, D-063), Phase 35 (embedding, D-065),
Phase 36 (fallback, D-066).
**Status:** Done

---

## 1. Goal

Guarantee that an italic emphasis run renders in (and embeds) a real italic cut:
the role's own when present, else — when the family ships no italic cut — a
declared fallback that does, instead of a faux-italic sans.

## 2. Why now

R9.7 (`emphasis-as-italic-display`, MED) follows the font cluster (Phases
33/35/36). Its display-italic guarantee is already delivered by D-063 (display
family on the run's `a:latin`) + D-065 (the embedding pass collects and embeds the
italic cut); the remaining engine work is making the D-066 fallback italic-aware,
plus a latent `<p:font>` prefix bug surfaced while testing.

## 3. RFC sections implemented

- `RFC §7.6` — font handling. Extends the write-time fallback realization to be
  italic-cut-aware; fixes the `embeddedFontLst` element prefix.

## 4. Brief findings incorporated

- `docs/research/20-emphasis-italic-display.md` — *the display-italic guarantee is
  already delivered* → covered by a verification test, no new code.
- `20-emphasis-italic-display.md` — *the incremental work is italic-aware
  fallback* → `resolveFontFallbacks` resolves per `(family, italic)`.
- `20-emphasis-italic-display.md` — *generalize the codec rewrite to a resolver* →
  `RewriteFontFaces(func(typeface string, bold, italic bool) string)`.
- `20-emphasis-italic-display.md` — *additive + deterministic* → both cuts
  resolving = no substitution; no chain / no source = byte-identical.

## 5. Findings I'm departing from

None.

## 6. Decisions referenced

- `D-063` — display family on the run's `a:latin` (the guarantee's family half).
- `D-065` — embedding collects/embeds the italic cut (the guarantee's bytes half).
- `D-066` — the fallback pass this phase makes italic-aware.
- `D-059` — engine half of `both`; the emphasis-treatment switch is Deckard's.
- New: `D-067` — italic-aware fallback + the `<p:font>` prefix fix (this phase).

## 7. Architecture

```text
resolveFontFallbacks()  (prepareForWrite, before syncSlides/autoEmbedFonts)
  for each role with Fallback, for italic in {false, true}:
     primary cut resolvable?  → keep (primary wins)
     else first [Family]+Fallback whose <italic?> cut resolves → resolved
  mapping[(family, italic)] = resolved
  RewriteFontFaces(resolve)  ← resolver keys on the run's italic flag

RestoreNamespaces: "font" → p:  (embeddedFontLst <p:font> was emitted bare)
```

## 8. Files added or changed

```text
internal/ooxml/slide/fonts.go              # CHANGED — RewriteFontFaces → resolver callback
pptx/fonts.go                              # CHANGED — resolveFontFallbacks italic-aware
internal/ooxml/restorenamespaces.go        # CHANGED — "font": "p" (embeddedFontLst child)
pptx/fonts_test.go                         # CHANGED — emphasis tests + p:font byte assert
internal/ooxml/slide/fonts_test.go         # CHANGED — resolver sig + italic-aware test
internal/ooxml/restorenamespaces_test.go   # CHANGED — embeddedFontLst prefix test
scripts/smoke/phase-37.sh                  # NEW — phase smoke
docs/research/20-emphasis-italic-display.md# NEW — brief
docs/research/INDEX.md                      # CHANGED — register brief 20
docs/plans/phase-37-emphasis-italic-display.md # NEW — this plan
docs/plans/README.md                        # CHANGED — Wave 9 phase index
docs/decisions.md                           # CHANGED — adds D-067
docs/glossary.md                            # CHANGED — fallback chain italic-aware note
docs/site/guide/theme.md                    # CHANGED — italic-aware fallback note
skills/define-a-theme/SKILL.md              # CHANGED — italic-aware fallback note
```

## 9. Public API surface

No new public symbol. `FontSpec.Fallback` (D-066) is now italic-aware in behavior.
The codec `SlidePart.RewriteFontFaces` signature changed (internal only) from
`map[string]string` to a resolver callback.

## 10. Risks

- **R1 — Behavior change vs Phase 36 for italic runs.** Now an italic run can be
  substituted while its upright siblings are not. **Mitigation:** this is exactly
  R9.7's target ("regular present, italic absent"); both-cuts-resolving and
  no-chain/no-source remain byte-identical (tested).
- **R2 — `"font": "p"` over-prefixing a DrawingML `<a:font>`.** **Mitigation:**
  the theme font collection is a literal string (not processed by
  `RestoreNamespaces`) and slides emit `a:latin`, never `font`; `font` occurs only
  in presentation's `embeddedFontLst` — verified at the byte level.

## 11. Acceptance criteria

1. An italic run of a family that ships a regular but no italic cut falls back to
   the first chain family whose italic cut resolves; the family's upright runs
   keep the primary.
2. An italic run at the display role embeds the display face's italic cut.
3. The emitted `embeddedFontLst` `<p:font>` element carries the `p:` prefix.
4. A deck with no `Fallback` / no `FontSource` is byte-identical to baseline;
   two saves are byte-identical.
5. `make coverage` shows touched packages ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | band per `coverage.json` | low-by-design; `0 failing` is the gate |
| `internal/ooxml` | 85% | codec band (prefix map) |
| `internal/ooxml/slide` | 85% | codec band (resolver rewrite) |

## 13. Smoke check

`scripts/smoke/phase-37.sh`: italic run falls back / upright keeps primary;
italic display run embeds the italic cut; codec rewrite is italic-aware; embedded
`<p:font>` is `p:`-prefixed (codec + deck level).

## 14. Tests

- **Unit:** `pptx` (italic fallback, display-italic embed, byte-level `p:font`),
  `internal/ooxml/slide` (italic-aware resolver), `internal/ooxml`
  (`embeddedFontLst` prefix).
- **Round-trip golden:** the resolved `a:latin` round-trips via `Run.Font()`; no
  prior golden emitted an `embeddedFontLst`, so the prefix fix re-pins nothing.
- **Integration / Fuzz / Benchmark:** none.

## 15. Vocabulary added

- (none new) — the `Font fallback chain` glossary entry gains an italic-aware note.

## 16. Plan deviations encountered during implementation

- A latent bug surfaced: `<p:font>` in `embeddedFontLst` was emitted bare
  (`<font>`), invalid OOXML that broke embedding since D-019. Fixed in this PR
  (§17) by adding `font` to the `RestoreNamespaces` prefix map + a byte-level
  test. Folded into D-067.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-37.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-067).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated.
