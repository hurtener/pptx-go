# Phase 01 — OPC + OOXML reorg

**Subsystem:** internal/opc, internal/ooxml
**RFC sections:** §6 (§6.1, §6.2)
**Deps:** Phase 00.
**Status:** In progress

---

## 1. Goal

Move the upstream `opc/` and `parts/` packages under `internal/` so raw
OPC/OOXML wire types are non-importable from outside the module (P3), with
`parts/` reorganized into per-part-family subpackages under
`internal/ooxml/` — a relocation + repackaging, not a rewrite.

## 2. Why now

Wave 1 builds the theme and builder spine (`docs/plans/README.md §2`) on
top of `internal/opc` and `internal/ooxml`. Those packages must exist at
their final import paths, behind the `internal/` wall, before Phase 02–04
compose against them. It is the first structural move after the Phase 00
scaffolding and unblocks every later phase.

## 3. RFC sections implemented

- `RFC §6.1` — `internal/opc`: relocate the upstream OPC package verbatim
  (Package, StreamPackage, Part, ContentTypes, Relationships, PackURI,
  collectors, ResourceDedupPool). Behavior preserved; the move is a rename.
- `RFC §6.2` — `internal/ooxml`: reorganize the monolithic `parts/` package
  into per-family subpackages, keeping families independent (no cross-family
  XML-type imports except documented shared helpers). `namespaces.go` lands
  with canonical NS URI constants.

## 4. Brief findings incorporated

No informing brief — this is a structural relocation with no prior-art
investigation needed (`docs/research/INDEX.md` lists no Wave 1 brief for
this move). The dependency analysis below stands in for a brief.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-010` — Single-module distribution: subpackages live under the one
  module; `internal/` enforces non-importability by other modules.
- `D-027` — Coverage-gate strictness ramp: this phase flips
  `require_configured: true` and bands each relocated package.
- `D-028` *(new, this PR)* — drawingML types and the `XMLWriter`
  serialization base stay within `internal/ooxml/slide` for Phase 01;
  `internal/ooxml/drawing` ships as a documented placeholder. See §7.
- `D-029` *(new, this PR)* — the `require_configured: true` flip and
  co-locating the upstream tests into the new package directories are
  deferred past Phase 01 (refines D-027). Self-coverage of the relocated
  packages is 0% because their tests are external and fixture-dependent;
  `-coverpkg` double-counts blocks in the gate. See §12.

## 7. Architecture

### Dependency analysis (why the split is safe)

The `parts/` package is monolithic but its type families form an **acyclic**
graph:

- `XMLWriter` / `XMLWriterPool` and every drawingML type (`XSp`,
  `XShapeProperties`, `XFillProperties`, `XBlip`, `XTransform2D`,
  `XTextBody`, `XTable*`, …) are referenced **only** by `slide.go` /
  `slide_types.go` — they are slide-family-internal today.
- `master_*` and `xml_master_models.go` reference only their own `XML*`
  types and the slide family — they belong in `slide/`.
- `theme`, `core`, `chart`, `relations`, `media` reference no other
  family's types.
- The one cross-family edge is `presentation.AddSlide(rId, *SlidePart)` →
  `presentation` imports `slide`. No file in the slide family references
  `SlideSize` or `PresentationPart`, so there is **no cycle**.

### Target layout

```text
internal/opc/                  # MOVED verbatim from opc/ (package opc)
internal/ooxml/
├── namespaces.go              # NEW — canonical NS URI constants (package ooxml)
├── presentation/              # ← presentation.go            (imports slide)
├── slide/                     # ← slide.go, slide_types.go, master_types.go,
│                              #   master_parser.go, xml_master_models.go,
│                              #   xmlutils.go  (incl. drawingML types + XMLWriter)
├── theme/                     # ← theme.go, theme_types.go, theme_default.go
├── core/                      # ← coreprops.go, appprops.go, appprops_types.go
├── chart/                     # ← chart.go, chart_types.go
├── relations/                # ← relationship.go
├── media/                     # ← media.go, embedding.go
└── drawing/                   # NEW — placeholder doc.go (D-028)
```

### D-028 — drawing types stay in `slide/` for Phase 01

RFC §6.2 lists `drawing/` as its own subpackage. Today every drawingML type
is slide-internal, and the `XMLWriter` base that their `WriteXML` methods
use also lives in the slide family. Extracting `drawing/` now would force
`XMLWriter` into a shared `common/` package and split `slide_types.go`
before any **cross-family** consumer exists. Decision: keep drawingML types
and `XMLWriter` in `internal/ooxml/slide` for Phase 01 and ship `drawing/`
as a documented placeholder (mirroring how RFC §6.2 treats `chart/`). The
types migrate to `drawing/` — with `XMLWriter` moving to a shared helper —
when the builder (Phase 03+) or the SVG translator (Phase 12) first needs
them outside the slide family. This honors the §6.2 independence rule (no
new cross-family coupling) while deferring avoidable surgery.

### Downstream

`pptx/` (~150 `parts.X` references) and the `test/` packages are
re-qualified from `parts` to the owning subpackage. This is mechanical
(symbol → subpackage map) and verified by `go build` / `go test`.

## 8. Files added or changed

```text
opc/* -> internal/opc/*                  # MOVED (git mv; package opc unchanged)
parts/<family>.go -> internal/ooxml/<family>/  # MOVED + repackaged
internal/ooxml/namespaces.go             # NEW — NS URI constants
internal/ooxml/drawing/doc.go            # NEW — placeholder (D-028)
pptx/*.go                                # CHANGED — import paths + qualifiers
main.go                                  # CHANGED — opc import path
test/**/*.go                             # CHANGED — import paths + qualifiers;
                                         #   relocated alongside their packages where apt
.golangci.yml                            # CHANGED — lint exclusions move to internal/opc, internal/ooxml
scripts/drift-audit.sh                   # CHANGED — old-path + P3 isolation checks
docs/decisions.md                        # CHANGED — adds D-028
docs/glossary.md                         # CHANGED — OOXML subpackage vocabulary if needed
docs/plans/phase-01-opc-ooxml-reorg.md   # NEW — this plan
scripts/smoke/phase-01.sh                # NEW — phase smoke
test/integration/reorg_roundtrip_test.go # NEW — spot-check round-trip (§14)
```

## 9. Public API surface

No `pptx`/`scene` public API change in intent — `pptx`'s exported surface
is unchanged; only its (internal) import paths move. `opc` and `parts`
types become non-importable from outside the module (the point of the
move). No deprecation aliases: nothing outside the module consumed them
(pre-V1, no released tags).

## 10. Risks

- **R1 — Import cycle during the split.** *Mitigation:* the dependency
  analysis (§7) shows an acyclic graph; the one edge is `presentation →
  slide`. Build after each family moves; a cycle fails fast.
- **R2 — Missed re-qualification in `pptx/`.** ~150 `parts.X` references
  re-qualify to subpackages. *Mitigation:* a scripted symbol→subpackage
  map, then `go build ./...` + `go vet` catch every miss; commit only when
  green.
- **R3 — Behavior drift from "just a move".** *Mitigation:* no logic edits
  during the move; the relocated upstream tests plus a round-trip
  spot-check (open a built deck, save, reopen, assert equivalence) guard
  against drift.
- **R4 — Coverage gate flip breaks CI.** Flipping `require_configured=true`
  would fail the relocated packages, whose self-coverage is 0% (external,
  fixture-dependent tests). *Mitigation:* keep `require_configured=false`
  and defer the flip + bands to when tests co-locate (D-029); `make
  coverage` stays green.

## 11. Acceptance criteria

1. `make build` and `make test` pass after the reorg (`-race`).
2. `opc/` and `parts/` no longer exist; their code lives under
   `internal/opc/` and `internal/ooxml/*`; each `internal/ooxml`
   subpackage compiles independently.
3. A round-trip spot-check passes: author a deck via the builder, save,
   reopen, assert model/structural equivalence.
4. `make drift-audit` passes, including a check that no package outside
   `internal/...` is broken by the move and that `encoding/xml` / raw
   OOXML structs remain isolated to `internal/ooxml` (P3).
5. `make coverage` passes; `internal/coveragecheck` stays at its 70%
   band. The `require_configured: true` flip and per-package bands for the
   relocated code are deferred (D-029) — see §12.
6. `make preflight`, `make check-mirror`, and `golangci-lint run` pass;
   prior smoke (`phase-00.sh`) still passes.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `internal/coveragecheck` | 70% | unchanged tooling band |
| `internal/opc`, `internal/ooxml/*` | deferred (D-029) | bands land when tests co-locate |

The relocated packages' tests are external (`test/…`) and
fixture-dependent, so per-package self-coverage is 0% and `-coverpkg`
double-counts in the gate. Per **D-029**, `require_configured` stays
`false` and these packages are banded once their tests are co-located and
hardened (Phase 02+). This is a deliberate refinement of D-027's Phase-01
timing, not a silent lowering.

## 13. Smoke check

`scripts/smoke/phase-01.sh` verifies:

1. `OK:` `opc/` and `parts/` directories are gone; `internal/opc` and
   `internal/ooxml/<families>` exist.
2. `OK:` library builds CGo-free.
3. `OK:` no package imports the old `pptx-go/opc` or `pptx-go/parts` paths.
4. `OK:` `internal/ooxml` subpackages build independently.
5. `OK:` round-trip spot-check test passes.

## 14. Tests

- **Unit:** all upstream tests preserved and passing; they remain under
  `test/` for this phase (re-qualified to the new subpackages). Physically
  co-locating them into the package directories is deferred (D-029).
- **Round-trip golden:** a spot-check in `test/integration/` (open → save →
  reopen → assert) — first integration test (Deps name Phase 00; this phase
  closes the opc/ooxml seam the builder will consume).
- **Integration:** yes — `test/integration/reorg_roundtrip_test.go`.
- **Fuzz:** none new (parse-surface fuzzing lands when the parsers are
  hardened in Phase 18–19).
- **Benchmark:** none new.

## 15. Vocabulary added

- `Part family` — an OOXML part grouping that owns one `internal/ooxml`
  subpackage (presentation, slide, theme, core, chart, relations, media,
  drawing).

## 16. Plan deviations encountered during implementation

- **`StripNamespacePrefixes` is a cross-family helper.** It (and
  `XMLDeclaration`) is used by presentation/theme/slide/core/relations, so
  it landed in the `internal/ooxml` root package (`package ooxml`) rather
  than a family subpackage — the shared-helper home RFC §6.2 anticipates.
  `xmlutils.go` moved to `internal/ooxml/` instead of `slide/`.
- **`RelType*` URIs.** Used by `slide`; they stay owned by the `relations`
  subpackage and `slide` references `relations.RelType*` (one-way edge).
  `namespaces.go` holds the namespace URIs only, to avoid two sources of
  truth.
- **Test-file/package-name collisions.** `slide` and `theme` are used as
  local variable names in four test files that also import the new
  packages of those names; the colliding imports were aliased
  (`slidex`/`themex`) and only true package references re-qualified.
- **Coverage flip deferred (D-029).** See §12 / R4.

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-01.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (`D-028`).
- [ ] (Phase 20+) Docs/skills — N/A (inert).
