# Phase 36 — font fallback chain

**Subsystem:** pptx — Layer 1 builder (theme/typography + font embedding)
**RFC sections:** §7.6
**Deps:** Phase 35 (font-embedding pass, D-065); the `FontSource` mechanism
(D-019).
**Status:** Done

---

## 1. Goal

A type role can declare an ordered fallback chain (`FontSpec.Fallback`) that the
engine realizes at write time: when a registered `FontSource` cannot resolve the
role's primary family, the run's single-valued `a:latin` typeface is rewritten to
the first family in the chain the source can resolve — a controlled near-match
instead of an arbitrary host default.

## 2. Why now

R9.6 follows R9.1 (Phase 35): with a real embedding pass and `FontSource`
machinery in place, the fallback pass shares them and runs just before embedding
so the embedded bytes match the emitted face. It also unblocks R9.7's "no italic
cut → fall back rather than faux-italicize a sans". MED, engine half (D-059).

## 3. RFC sections implemented

- `RFC §7.6` — font handling. Adds a write-time realization of a per-role
  fallback chain on top of the run font emission; no new wire format.

## 4. Brief findings incorporated

- `docs/research/19-font-fallback-stack.md` — *the chain lives on `FontSpec`* →
  `FontSpec.Fallback []string`.
- `19-font-fallback-stack.md` — *a `FontSource` is the availability oracle* →
  the pass treats a family the source resolves as available; with no source it is
  a no-op.
- `19-font-fallback-stack.md` — *deterministic write-time substitution* →
  `resolveFontFallbacks` builds a `primary → resolved` map from the theme in a
  fixed role order, then rewrites runs via `slide.SlidePart.RewriteFontFaces`.
- `19-font-fallback-stack.md` — *idempotent + byte-identical across saves* → the
  map is keyed on primaries, so a substituted run (now carrying the fallback) is
  not rewritten again.
- `19-font-fallback-stack.md` — *independent of embedding, ordered before it* →
  runs in `prepareForWrite` before `syncSlides`/`autoEmbedFonts`; gated only on a
  `FontSource` + a declared chain.

## 5. Findings I'm departing from

None.

## 6. Decisions referenced

- `D-019` — `FontSource`/`EmbedFont` mechanism (the availability oracle + the
  embed path the resolved face flows into).
- `D-059` — engine half of `both` requirements.
- `D-064` — precedent: a per-role `FontSpec` field that is a theme-time config
  documented in THEME.md (here, `Fallback`).
- `D-065` — the embedding pass `resolveFontFallbacks` runs just before.
- New: `D-066` — the font fallback chain (this phase).

## 7. Architecture

```text
prepareForWrite()                          (holds p.mu)
  ├─ syncNotes
  ├─ resolveFontFallbacks()  ← NEW, self-gated on fontSource!=nil + declared chain
  │     build primary→resolved map from p.theme (roles TypeDisplay…TypeCode):
  │        primary resolvable?  → keep (primary wins)
  │        else first [Family]+Fallback the source resolves → resolved
  │     for each slide: SlidePart.RewriteFontFaces(map)   (a:latin rewrite)
  ├─ syncSlides            ← serializes the now-resolved runs
  ├─ syncMedia / syncSections
  ├─ autoEmbedFonts()      ← embeds the resolved faces (D-065)
  └─ syncPresentationPart
```

## 8. Files added or changed

```text
pptx/theme.go                              # CHANGED — FontSpec.Fallback field
pptx/fonts.go                              # CHANGED — resolveFontFallbacks pass
pptx/presentation.go                       # CHANGED — call it in prepareForWrite
internal/ooxml/slide/fonts.go              # CHANGED — SlidePart.RewriteFontFaces
pptx/fonts_test.go                         # CHANGED — fallback tests
pptx/theme_test.go                         # CHANGED — DeepEqual (FontSpec now non-comparable)
internal/ooxml/slide/fonts_test.go         # CHANGED — RewriteFontFaces test
scripts/smoke/phase-36.sh                  # NEW — phase smoke
docs/research/19-font-fallback-stack.md    # NEW — brief
docs/research/INDEX.md                      # CHANGED — register brief 19
docs/plans/phase-36-font-fallback-stack.md # NEW — this plan
docs/plans/README.md                        # CHANGED — Wave 9 phase index
docs/decisions.md                           # CHANGED — adds D-066
docs/glossary.md                            # CHANGED — Font fallback chain term
docs/design/THEME.md                        # CHANGED — fallback entry + embedding refresh
docs/site/guide/theme.md                    # CHANGED — fallback note
skills/define-a-theme/SKILL.md              # CHANGED — Fallback field
```

## 9. Public API surface

```go
// pptx — FontSpec gains:
Fallback []string // ordered substitute families realized at write time (D-066)
```

No break (additive field). Adding a slice makes `FontSpec` non-comparable with
`==`; the only in-repo comparison (a `ResolveType` determinism test) moved to
`reflect.DeepEqual`.

## 10. Risks

- **R1 — Non-determinism from theme map iteration.** **Mitigation:** roles are
  iterated in a fixed slice order; the `primary → resolved` map is first-seen.
- **R2 — Cross-save mutation breaking byte-identity.** Substitution mutates the
  in-memory runs. **Mitigation:** the map is keyed on primaries, so a second save
  finds only fallback families (not keys) and rewrites nothing — two saves are
  byte-identical (tested).
- **R3 — Surprising behavior from merely registering a `FontSource`.**
  **Mitigation:** the pass is also gated on a *declared* chain, so a deck with no
  `Fallback` is byte-identical even with a source registered (tested).
- **R4 — Availability probe ambiguity.** Availability is a family question;
  the probe resolves the regular cut. **Mitigation:** documented; a family with
  only a non-regular cut is listed explicitly by the soul.

## 11. Acceptance criteria

1. A deck whose primary face the `FontSource` cannot resolve renders the run in
   the first resolvable fallback family (`a:latin` rewritten), not the primary.
2. The primary wins (no substitution) when the source resolves it.
3. A deck with an empty chain, or with no `FontSource`, is byte-identical to the
   pre-change output.
4. Two saves of the same deck are byte-identical (deterministic + idempotent).
5. With `WithFontEmbedding` on, the resolved fallback face is what gets embedded
   (the primary is not).
6. `make coverage` shows touched packages ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | band per `coverage.json` | low-by-design; `0 failing` is the gate |
| `internal/ooxml/slide` | 85% | codec band (new `RewriteFontFaces`) |

## 13. Smoke check

`scripts/smoke/phase-36.sh`: substitutes an unavailable primary; primary wins
when available; byte-identical when unused; deterministic + idempotent; embeds
the resolved fallback face; codec rewrites matching faces.

## 14. Tests

- **Unit:** `pptx` (substitution, primary-wins, byte-identity, determinism,
  embed-resolved), `internal/ooxml/slide` (`RewriteFontFaces`).
- **Round-trip golden:** the *resolved* `a:latin` round-trips via `Run.Font()`;
  `Fallback` is theme config, not a persisted field — no new round-trip surface.
- **Integration:** no (single-subsystem pass; the `FontSource` seam is exercised
  with a real fake in unit tests).
- **Fuzz / Benchmark:** none.

## 15. Vocabulary added

- `Font fallback chain` — `FontSpec.Fallback`, realized at write time against the
  registered `FontSource`.

## 16. Plan deviations encountered during implementation

- Adding a slice to `FontSpec` made it non-comparable; an existing
  `ResolveType` determinism test (`!=`) moved to `reflect.DeepEqual`. No
  acceptance-criterion change.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-36.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-066).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated.
