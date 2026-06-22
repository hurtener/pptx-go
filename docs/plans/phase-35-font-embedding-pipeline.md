# Phase 35 — font-embedding pipeline

**Subsystem:** pptx — Layer 1 builder (font embedding)
**RFC sections:** §7.6
**Deps:** Phase 33 (display-face role, D-063 — routes a brand face through the
run `a:latin`); the existing `EmbedFont`/`FontSource` mechanism (D-019).
**Status:** Done

---

## 1. Goal

An opt-in save-time pass (`pptx.WithFontEmbedding()`) that walks every slide's
runs, collects the distinct font faces the deck actually uses, and embeds each
via the registered `FontSource` — so a deck themed with a brand display/heading
face ships those faces and renders with them on any machine, with no host
install.

## 2. Why now

R9.1 (`font-embedding-pipeline`, CRITICAL) is the gating unit of the Wave 9
font cluster: R9.6 (fallback stack), R9.7 (emphasis-as-italic-display), and
R9.8 (weight-aware embedding) all build on a real embedding pass. The engine
already has the per-face mechanism (`Presentation.EmbedFont`) but nothing
collects and embeds the *used* set; D-059 puts that engine half in pptx-go (the
`FontProvider`/soul half is Deckard's). The five "designed type" foundation
phases (30–34) are done; embedding is what makes a brand serif actually appear.

## 3. RFC sections implemented

- `RFC §7.6` — font embedding. This phase adds the automatic *collection* layer
  on top of the existing per-face `EmbedFont` primitive; the wire format
  (`/ppt/fonts/fontN.fntdata` parts + `embeddedFontLst`) is unchanged.

## 4. Brief findings incorporated

- `docs/research/18-font-embedding-pipeline.md` — *opt-in + source-gated* → the
  pass runs only when `WithFontEmbedding()` is set **and** a `FontSource` is
  registered; off / no-source is byte-identical.
- `18-font-embedding-pipeline.md` — *walk the in-memory runs in the codec
  package* → `(s *SlidePart) UsedFontFaces()` in `internal/ooxml/slide`, reusing
  the `DroppedDescendants` traversal (XSp + table-cell text bodies).
- `18-font-embedding-pipeline.md` — *deterministic order* → faces merged into a
  set then sorted by `(family, bold, italic)` before embedding, so part
  numbering and rel ids are byte-identical across worker counts.
- `18-font-embedding-pipeline.md` — *idempotent vs manual `EmbedFont`* →
  `presentationPart.HasEmbeddedFace(typeface, style)` skips an already-embedded
  bucket.
- `18-font-embedding-pipeline.md` — *warn-don't-fail* → a missing face logs a
  Warn and the Save still succeeds (same contract as `register-an-asset`).
- `18-font-embedding-pipeline.md` — *lock discipline* → `EmbedFont` split into a
  lock-taking wrapper and a lock-free `embedFontLocked`; `autoEmbedFonts` runs
  under the write-path lock.

## 5. Findings I'm departing from

None. The brief's recommendations are implemented as written.

## 6. Decisions referenced

- `D-019` — font embedding mechanism (`FontSource` + `EmbedFont`). This phase
  adds an automatic collection pass on top; the per-face primitive is unchanged.
- `D-059` — Wave 2 scope: pptx-go implements the engine half of `both`
  requirements. R9.1's engine half is "collect used faces + call `EmbedFont`";
  the provider/soul half is Deckard's.
- `D-063` — display-face role routes a brand face through the run `a:latin`, so
  the walk over `Latin.Typeface` picks up the themed display/heading faces.
- New: `D-065` — the automatic font-embedding pass (this phase).

## 7. Architecture

```text
Save / Write / WriteToBytes / SaveStream
  └─ prepareForWrite()                       (holds p.mu)
       ├─ syncNotes / syncSlides / syncMedia / syncSections
       ├─ autoEmbedFonts()   ← NEW, gated on fontEmbedding && fontSource!=nil
       │     for each slide: slide.UsedFontFaces()      (XSp + table cells)
       │     merge into a set of (family, bold, italic)
       │     sort by (family, bold, italic)             (determinism)
       │     for each face: skip if HasEmbeddedFace; else embedFontLocked()
       │                    (warn-don't-fail on a missing face)
       └─ syncPresentationPart()  ← serializes embeddedFontLst + font rels
```

`embedFontLocked` is the lock-free body of `EmbedFont` (caller holds `p.mu`),
matching the `ensurePresentationOPCPart` convention. The walk lives in
`internal/ooxml/slide` so the XML traversal stays in the codec package.

## 8. Files added or changed

```text
internal/ooxml/slide/fonts.go              # NEW — FontFace, SlidePart.UsedFontFaces
internal/ooxml/presentation/embeddedfont.go# CHANGED — HasEmbeddedFace accessor
pptx/options.go                            # CHANGED — WithFontEmbedding option
pptx/presentation.go                       # CHANGED — fontEmbedding field; autoEmbedFonts in prepareForWrite
pptx/fonts.go                              # CHANGED — embedFontLocked split; autoEmbedFonts pass
pptx/fonts_test.go                         # CHANGED — auto-embed tests
internal/ooxml/slide/fonts_test.go         # NEW — UsedFontFaces tests
internal/ooxml/presentation/embeddedfont_test.go # CHANGED — HasEmbeddedFace test
scripts/smoke/phase-35.sh                  # NEW — phase smoke
docs/research/18-font-embedding-pipeline.md# NEW — brief
docs/research/INDEX.md                      # CHANGED — register brief 18
docs/plans/README.md                        # CHANGED — Wave 9 phase index
docs/decisions.md                           # CHANGED — adds D-065
docs/glossary.md                            # CHANGED — font-embedding pass term
docs/site/reference/pptx.md                 # CHANGED — WithFontEmbedding
docs/site/guide/builder.md                  # CHANGED — auto font embedding note
skills/scaffold-a-presentation/SKILL.md     # CHANGED — auto font embedding
```

## 9. Public API surface

```go
// pptx
func WithFontEmbedding() Option // opt-in: at save, embed every used face via the
                                // registered FontSource (no-op without a source)
```

No break to existing surface. `EmbedFont` / `WithFontSource` / `SetFontSource`
are unchanged; the pass composes them.

## 10. Risks

- **R1 — Deadlock on the write-path lock.** `EmbedFont` takes `p.mu.Lock()`;
  `prepareForWrite` already holds `p.mu`. **Mitigation:** split off a lock-free
  `embedFontLocked` that the pass calls; `EmbedFont` stays the public
  lock-taking wrapper.
- **R2 — Non-determinism from set iteration.** A Go map iterates in random
  order. **Mitigation:** collect into a set, then sort by `(family, bold,
  italic)` before embedding; the pass is single-threaded.
- **R3 — Double-embedding when the caller also called `EmbedFont`.**
  **Mitigation:** `HasEmbeddedFace(typeface, StyleFor(...))` dedup skips an
  already-recorded bucket.
- **R4 — A missing face failing the Save.** **Mitigation:** warn-don't-fail —
  log a Warn, continue; the deck saves with the faces that resolved.

## 11. Acceptance criteria

1. A deck with `WithFontSource` + `WithFontEmbedding`, themed with a non-default
   family on a run, saves a `.pptx` whose `embeddedFontLst` covers every
   `(family, bold, italic)` used on a slide (`EmbeddedFontCount > 0`,
   `/ppt/fonts/font1.fntdata` present).
2. The same deck saved twice is byte-identical (deterministic part order / rel
   ids regardless of worker count).
3. A deck **without** `WithFontEmbedding` (or with no `FontSource`) is
   byte-identical to the pre-change output (zero `EmbedFont` calls).
4. A face already embedded via a manual `EmbedFont` is not embedded twice by the
   pass (`EmbeddedFontCount` counts it once).
5. A `FontSource` that can't resolve a used face logs a Warn and the Save still
   succeeds, embedding the faces that did resolve.
6. `make coverage` shows touched packages ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | band per `coverage.json` | low-by-design band; `0 failing` is the gate |
| `internal/ooxml/slide` | 85% | codec band (new `UsedFontFaces`) |
| `internal/ooxml/presentation` | 85% | codec band (new `HasEmbeddedFace`) |

## 13. Smoke check

`scripts/smoke/phase-35.sh`:

1. `OK:` Build a deck with a fake `FontSource` + `WithFontEmbedding`, a themed
   run; assert `EmbeddedFontCount > 0` and a `font1.fntdata` part exists.
2. `OK:` Save twice; assert byte-identical output.
3. `OK:` Same deck without `WithFontEmbedding`; assert `EmbeddedFontCount == 0`.
4. `OK:` Manual `EmbedFont` then the pass; assert no duplicate.

## 14. Tests

- **Unit:** `internal/ooxml/slide` (`UsedFontFaces`),
  `internal/ooxml/presentation` (`HasEmbeddedFace`), `pptx` (auto-embed: on/off
  byte-identity, determinism, idempotency, warn-don't-fail).
- **Round-trip golden:** the embedded faces survive in `embeddedFontLst`
  (`EmbeddedFontCount` after reopen) — covered by the existing embed round-trip
  plus the new on/off assertions; no new builder *shape field* is added.
- **Integration:** no — this is a single-subsystem pass over the builder's own
  in-memory model; the seam (`FontSource`) is exercised by the unit tests with a
  real fake source.
- **Fuzz:** none (no new parse surface).
- **Benchmark:** none required.

## 15. Vocabulary added

- `font-embedding pass` — the opt-in save-time pass (`WithFontEmbedding`) that
  collects a deck's distinct used faces and embeds each via the registered
  `FontSource`.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-35.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-065).
- [x] Docs site updated for user-facing surface changes.
- [x] Affected agent skill(s) updated.
