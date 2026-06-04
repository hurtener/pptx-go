# Phase 19 — External-deck read robustness (best-effort)

**Subsystem:** pptx (read) + internal/ooxml (parsers)
**RFC sections:** §16
**Deps:** Phase 18 (the navigable read model external decks degrade against)
**Status:** Draft

---

## 1. Goal

`pptx.NewFromBytes` / `OpenStream` load third-party (non-pptx-go) decks without
panicking and surface what could not be faithfully read via
`Presentation.ReadWarnings()`, while every package part pptx-go does not model
passes through unchanged on re-save.

## 2. Why now

Phase 19 closes Wave 6 (reading + round-trip). Phase 18 delivered lossless
round-trip of **pptx-go-authored** decks (RFC §16's binding guarantee). The
remaining Wave-6 work is the RFC's **best-effort** posture toward decks pptx-go
did *not* write (PowerPoint, Keynote export, other libraries): they must open
without crashing and report their degradation, even though fidelity is not
promised. After this phase, a Wave-6 checkpoint audit closes the wave
(`CLAUDE.md §17`).

## 3. RFC sections implemented

- `RFC §16` (best-effort half) — third-party PPTX is best-effort with **graceful
  degradation**: "an unrecognized extension element is ignored at parse time, a
  recognized one is surfaced … we do not promise round-trip fidelity. V2 invests
  in third-party robustness." Phase 18 implemented §16's *authored-deck* round-trip
  guarantee; this phase implements §16's *external-deck* graceful-degradation
  clause. Fidelity preservation of unrecognized content stays V2.

## 4. Brief findings incorporated

- `docs/research/08-roundtrip-read.md` — the read-subsystem brief.
  - **Read is reconstruct-over-the-parsed-tree** → external decks reuse the same
    `FromXML`/`repopulate*` path Phase 18 reads from; this phase hardens that
    path against shapes/parts/structures the parsers do not model, rather than
    adding a parallel external reader.
  - **Brief Q (third-party robustness) was explicitly deferred from Phase 18** →
    this phase is that deferred work, scoped to the RFC floor (warn, don't
    preserve) per D-048.

## 5. Findings I'm departing from

None from the brief. **This plan departs from the master plan's Phase 19 entry
(`docs/plans/README.md` §"Wave 6")**, which says unrecognized OOXML is "preserved
as opaque `RawShape` / `RawPart` carriers." Departing because RFC §16 (which
outranks the master plan) parks fidelity preservation in V2 ("V2 invests in
third-party robustness") and only requires that unrecognized elements be
*ignored* gracefully. Phase 19 therefore delivers **graceful degradation +
reporting** (no-panic + `ReadWarnings`), not byte-preserving shape carriers.
Unknown *parts* already round-trip via the OPC pass-through (a property this
phase verifies and tests), so "`RawPart`" is realized as that existing
pass-through, not a new carrier type. The master-plan §19 entry is updated to
match in the PR#1 commit, and the change is recorded as **D-048**.

## 6. Decisions referenced

- `D-035` — deterministic byte-identical saves — underpins the part pass-through
  test (an external deck's unmodeled parts re-emit unchanged).
- `D-047` — the navigable read model Phase 18 built; external decks degrade
  against it.
- **`D-048` (new, this phase)** — Phase 19 implements RFC §16's external-deck
  clause as **best-effort graceful degradation**: `NewFromBytes` / `OpenStream`
  never panic on a third-party deck, surface unrecognized/dropped content in
  `Presentation.ReadWarnings()`, and pass unmodeled parts through unchanged on
  save. Opaque `RawShape`/`RawPart` *preservation* of unrecognized content is
  **deferred to V2** (RFC §16), superseding the master-plan §19 "opaque carriers"
  wording.

## 7. Architecture

All read work reuses the Phase 18 path (`FromXML` → `repopulate*`); no parallel
external reader. Three pieces:

```text
1. Reporting surface (pptx)
   pptx/read_warnings.go   ReadWarning, ReadWarningKind, (*Presentation).ReadWarnings()

2. Collection (internal/ooxml/slide → pptx)
   XSpTree.UnmarshalXML records dropped child element names (today it silently
   d.Skip()s them) in an unexported `dropped []string` (xml:"-"); SlidePart
   exposes them; repopulateSlides maps them to ReadWarnings tagged with the
   slide part URI. Dangling/unreadable slide references (today skipped silently)
   also become warnings.

3. Robustness (pptx + internal/ooxml + tests)
   Defensive guards on the repopulate*/FromXML paths (no nil-deref on missing
   required children); a synthetic external-deck corpus under
   test/integration/testdata/external/; extended parse fuzz seeds. Part
   pass-through (opc emits every loaded part) is verified + tested.
```

The collection seam stays P3-clean: `internal/ooxml/slide` records bare element
*names* (plain strings), never pptx types; `pptx` owns the `ReadWarning` mapping.

## 8. Files added or changed

```text
# PR#1 — reporting surface + dropped-element collection
pptx/read_warnings.go                 # NEW — ReadWarning, ReadWarningKind, ReadWarnings()
pptx/presentation.go                  # CHANGED — collect warnings in repopulateSlides; store on Presentation
internal/ooxml/slide/slide_marshal.go # CHANGED — XSpTree.UnmarshalXML records dropped child names
internal/ooxml/slide/slide_types.go   # CHANGED — XSpTree.dropped []string `xml:"-"`; SlidePart accessor
pptx/read_warnings_test.go            # NEW — dropped-element + dangling-ref warning round-trip
scripts/smoke/phase-19.sh             # NEW — phase smoke (criteria flip across PRs)
docs/decisions.md                     # CHANGED — D-048
docs/plans/README.md                  # CHANGED — reconcile §19 entry to D-048
docs/glossary.md                      # CHANGED — read warning, external deck
docs/plans/phase-19-external-read-robustness.md  # NEW (this file)

# PR#2 — no-panic hardening + corpus + fuzz
pptx/presentation.go                  # CHANGED — defensive guards on the read path
internal/ooxml/.../*.go               # CHANGED — guards where a malformed part can nil-deref
test/integration/external_read_test.go        # NEW — synthetic external-deck corpus, no-panic + warnings
test/integration/testdata/external/*.pptx     # NEW — hand-authored external-style decks
internal/ooxml/slide/fuzz_test.go     # CHANGED — external-style seeds (or opc fuzz, where the seam is)
CHANGELOG.md                          # CHANGED — ReadWarnings under Unreleased
```

## 9. Public API surface

```go
// pptx (additive; no break to NewFromBytes / OpenStream signatures)

// ReadWarningKind classifies a non-fatal issue encountered while reading a deck.
type ReadWarningKind int
const (
    WarnDroppedElement ReadWarningKind = iota // an unrecognized element was ignored at parse time
    WarnUnreadablePart                        // a referenced part was missing or could not be parsed; skipped
)

// ReadWarning is one non-fatal degradation noted while reading a (third-party) deck.
type ReadWarning struct {
    Kind    ReadWarningKind
    Part    string // the part URI the warning relates to, e.g. "/ppt/slides/slide2.xml"
    Element string // element local-name (WarnDroppedElement); empty otherwise
    Detail  string // human-readable context
}

// ReadWarnings returns the warnings collected when the deck was opened, in a
// stable order (by part, then element). It is empty for a pptx-go-authored deck.
func (p *Presentation) ReadWarnings() []ReadWarning
```

No write-side change. `ReadWarnings()` returns nil for any deck pptx-go authored
(Phase 18 round-trips it losslessly), so the existing test suite is unaffected.

## 10. Risks

- **R1 — Warning noise on pathological decks.** A deck with hundreds of group
  shapes would emit hundreds of warnings. **Mitigation:** de-duplicate per
  `(Part, Element)` — one `WarnDroppedElement` per distinct element type per
  part, not per occurrence; documented in the godoc.
- **R2 — P3 leakage via the collection seam.** Threading warnings out of
  `encoding/xml` unmarshalers is awkward. **Mitigation:** `internal/ooxml`
  records bare element *names* (strings) on the part; `pptx` does the
  `ReadWarning` mapping — no pptx type crosses into `internal/ooxml`.
- **R3 — "No panic" is unbounded.** We cannot enumerate every malformed input.
  **Mitigation:** the acceptance is a *fixed corpus* of synthetic external-style
  decks plus the existing parse fuzzers (extended with external-style seeds) —
  not an open-ended guarantee. The RFC scope is best-effort (D-048).
- **R4 — Part pass-through regressions.** A future write-path change could drop
  unmodeled parts. **Mitigation:** an explicit round-trip test asserts an
  external deck's unmodeled parts (e.g. a stray custom-XML part) survive
  `NewFromBytes` → `WriteToBytes` byte-for-byte (D-035).

## 11. Acceptance criteria

1. Opening a deck whose slide carries an unrecognized shape-tree element (e.g.
   `<p:grpSp>`, `<mc:AlternateContent>`) succeeds and reports a
   `WarnDroppedElement` naming the part and element; a pptx-go-authored deck
   reports zero warnings (PR#1).
2. A referenced-but-missing/unparseable part yields a `WarnUnreadablePart`
   rather than a hard error or panic (PR#1).
3. An external deck's unmodeled parts survive `NewFromBytes` → `WriteToBytes`
   unchanged (part pass-through; D-035) (PR#1/PR#2).
4. Every deck in the synthetic external-deck corpus loads without panic under
   `-race`, and the parse fuzzers (with external-style seeds) find no panic
   (PR#2).
5. `make coverage` ≥ bands for touched packages; `scripts/smoke/phase-19.sh`
   `OK ≥ count`, `FAIL = 0`; prior smokes pass (PR#1/PR#2).

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` | new read-warning code exercised | `pptx` is not in the mechanical gate (`coverage.json`), as in Phase 18; the new `read_warnings.go` is covered by unit + integration tests. |
| `internal/ooxml/slide` | 85% | codec band; the dropped-element collection lands here. |

No band override needed; no decision required for coverage.

## 13. Smoke check

`scripts/smoke/phase-19.sh` (criteria flip across the PRs):

1. `OK:` library builds CGo-free.
2. `OK:` ReadWarnings surfaces a dropped element on an external-style slide;
   an authored deck reports none (PR#1).
3. `OK:` an external deck's unmodeled parts round-trip byte-for-byte (PR#1/PR#2).
4. `OK:` the synthetic external-deck corpus loads without panic (PR#2).

## 14. Tests

- **Unit:** `pptx` (ReadWarning collection/mapping/dedupe/order),
  `internal/ooxml/slide` (UnmarshalXML records dropped names).
- **Round-trip golden:** part pass-through (external unmodeled part survives
  open → save byte-for-byte).
- **Integration** (`test/integration/external_read_test.go`): the synthetic
  external-deck corpus — no-panic + expected warnings through real
  `internal/opc` + `encoding/xml`. Required: this phase consumes Phase 18's read
  path across the opc/ooxml seam.
- **Fuzz:** extend the existing `internal/ooxml`/`opc` parse fuzz seeds with
  external-style structures (group shapes, AlternateContent, extra namespaces,
  truncated parts); the asserted invariant is "no panic, error or value
  returned."
- **Benchmark:** none (read robustness is not a hot reusable artifact).

## 15. Vocabulary added

- `read warning` — a non-fatal degradation (`ReadWarning`) noted while opening a
  deck pptx-go did not author; surfaced via `Presentation.ReadWarnings()`.
- `external deck` — a PPTX pptx-go did not write (PowerPoint, Keynote export,
  another library); read support is best-effort (RFC §16, D-048).

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-19.sh` reports `OK ≥ count(criteria)` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (D-048).
- [ ] Master-plan §19 entry reconciled to D-048.
- [ ] (Phase 20+) Docs site / skills updated. (inert)
