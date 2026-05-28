# Phase 00 — Module rename and hygiene scaffolding

**Subsystem:** Repo / build / docs
**RFC sections:** §3.3, §3.4, §5
**Deps:** none.
**Status:** In progress

---

## 1. Goal

Re-base the upstream fork onto the canonical module path
`github.com/hurtener/pptx-go`, relicense to Apache-2.0, and stand up the
build / test / lint / CI hygiene gates so every later phase has a green,
mechanically-enforced baseline to build on.

## 2. Why now

It is Wave 0 — the foundation (`docs/plans/README.md §2`). Nothing else
can be authored against a stable import path or measured against a
coverage gate until the rename and the gates exist. This phase eats its
own dogfood: it is the first phase to follow the §16 authoring workflow
and ship its own smoke script.

## 3. RFC sections implemented

- `RFC §3.3` — repository layout and subsystem ownership (the canonical
  `§3` tree in `CLAUDE.md`); this phase creates `internal/coveragecheck`
  and the `scripts/` + `.github/` scaffolding that the layout mandates.
- `RFC §3.4` — module identity and distribution (single module, canonical
  path).
- `RFC §5` — licensing and toolchain baseline (Apache-2.0; Go 1.24;
  CGo-free shipped artifact).

## 4. Brief findings incorporated

No informing brief — this is foundational scaffolding with no prior-art
investigation needed. (`docs/research/INDEX.md` lists no Wave 0 brief;
the rename + gates are mechanical.)

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-010` — Single-module distribution. The rename keeps `scene` a
  subpackage of `github.com/hurtener/pptx-go`, not a separate module.
- `D-026` — Engine, not product. The scaffolding adds gates and tooling
  only; no product behaviour enters the library here.

This plan creates one new settled decision — see `docs/decisions.md`
`D-027` (coverage-gate strictness ramp).

## 7. Architecture

The work is entirely build/CI/tooling plus a mechanical rename. The one
piece of shipped Go is `internal/coveragecheck`: a stdlib-only library
that parses a Go coverage profile, computes per-package statement
coverage, and compares it against bands declared in `coverage.json`. A
thin `cmd/coveragecheck` wraps it for the `make coverage` target.

```text
make preflight ─► scripts/preflight.sh ─► build (CGO_ENABLED=0)
                                        ├► scripts/smoke/phase-*.sh
                                        └► scripts/drift-audit.sh ─► mirror, module-path,
                                                                     P1, P3, §19 checks
make coverage  ─► go test -coverprofile ─► cmd/coveragecheck ─► coverage.json bands
```

## 8. Files added or changed

```text
go.mod                                   # CHANGED — module github.com/hurtener/pptx-go
**/*.go                                  # CHANGED — import paths rewritten
LICENSE                                  # CHANGED — now Apache-2.0
LICENSE.upstream                         # NEW — preserved upstream MIT
README.md, docs/*.md                     # CHANGED — usage import paths rewritten
                                         #   (RFC + decisions log keep historical upstream refs)
Makefile                                 # NEW — canonical targets (§4)
scripts/preflight.sh                     # NEW — build + smoke + drift-audit gate
scripts/drift-audit.sh                   # NEW — mirror/module/P1/P3/§19 checks
scripts/smoke/_template.sh               # NEW — smoke skeleton
scripts/smoke/phase-00.sh                # NEW — this phase's smoke
scripts/hooks/pre-commit                 # NEW — runs preflight
scripts/install-hooks.sh                 # NEW — installs the hook
.github/workflows/ci.yml                 # NEW — mirror/preflight/test/lint jobs
.golangci.yml                            # NEW — lint config
.editorconfig                            # NEW
.gitignore                               # CHANGED — coverage/build artifacts
CHANGELOG.md                             # NEW — Keep a Changelog
internal/coveragecheck/coveragecheck.go  # NEW — band-gate library
internal/coveragecheck/coveragecheck_test.go  # NEW — unit tests (98% cov)
internal/coveragecheck/cmd/coveragecheck/main.go  # NEW — gate CLI
internal/coveragecheck/coverage.json     # NEW — initial bands
docs/plans/phase-00-foundation.md        # NEW — this plan
docs/decisions.md                        # CHANGED — adds D-027
docs/glossary.md                         # CHANGED — adds "Coverage band", "Drift audit"
```

No user-facing surface (`pptx`/`scene`) changes; the §19 skill/doc rule
is inert (Phase 20 turns it on).

## 9. Public API surface

No public (`pptx`/`scene`) API changes. The only new Go is the private
`internal/coveragecheck` package:

```go
// internal/coveragecheck
func LoadConfig(path string) (Config, error)
func ParseProfile(r io.Reader) (map[string]*PackageCoverage, error)
func Check(coverage map[string]*PackageCoverage, cfg Config) Report
func Format(rep Report) string
func (r Report) Failed() bool
```

The mechanical import-path rewrite changes no exported symbols — only the
module prefix. No deprecation aliases needed.

## 10. Risks

- **R1 — Pre-existing red test suite (independent of the rename).** Three
  categories of failure exist on the upstream tree before this phase and
  the rename touches none of them:
  1. `test/` and `test/parts/` import a `slide` package that an upstream
     commit renamed to `pptx` without updating the imports — those two
     packages do not compile.
  2. `test/opc` has a failing assertion `TestNormalizeURI_vs_NormalizeZipPath`
     (backslash normalization) — a real `opc` behaviour question.
  3. `test/opc/TestParseRelationshipsFromFile` reads `test/test-data/…`,
     which is `.gitignore`d and absent from a clean checkout.
  **Mitigation:** all out of scope for Phase 00 (build/CI scaffolding +
  rename). The test relocation + reorg is Phase 01 (`parts/` →
  `internal/ooxml/*`); the normalization bug and the test-data fixture are
  Phase 01 cleanup items. `make build`, `make preflight`,
  `make check-mirror`, and `make drift-audit` are green; `make test` /
  `make coverage` carry these known-red packages until Phase 01. The
  Phase 00 acceptance criteria (§11) deliberately do not assert a fully
  green suite for this reason; the master-plan Phase 00 entry's "upstream
  tests still green" assumption did not hold on the inherited tree.
- **R2 — Coverage gate too strict for the pre-reorg tree.** Enabling
  `require_configured` now would fail on every un-banded upstream package.
  **Mitigation:** ship the gate with `require_configured=false` and only
  `internal/coveragecheck` banded; ramp to `true` in Phase 01 once the
  reorg lands and bands are set (`D-027`).

## 11. Acceptance criteria

1. `make build` succeeds (library compiles CGo-free under the new module
   name with no behaviour change).
2. `make check-mirror` passes (`AGENTS.md == CLAUDE.md`).
3. `make drift-audit` passes (mirror, no stale module path, P1/P3 seams
   clean).
4. `make preflight` passes (build + smoke + drift-audit).
5. `internal/coveragecheck` is ≥ its 70% tooling band under `make coverage`.
6. No tracked source/config file outside the RFC and the decisions log
   references the upstream module path.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `internal/coveragecheck` | 70% | CLI/tooling band (`CLAUDE.md §11`) |

`require_configured` is `false` for this phase (`D-027`); other packages
are not yet banded.

## 13. Smoke check

`scripts/smoke/phase-00.sh` verifies:

1. `OK:` library builds CGo-free (`CGO_ENABLED=0 go build ./...`).
2. `OK:` `AGENTS.md == CLAUDE.md`.
3. `OK:` no stale upstream module path outside RFC/decisions log.
4. `OK:` Apache-2.0 `LICENSE` present and upstream MIT preserved at
   `LICENSE.upstream`.
5. `OK:` `internal/coveragecheck` tests pass and the gate CLI runs.
6. `OK:` `go.mod` declares the canonical module path.

## 14. Tests

- **Unit:** `internal/coveragecheck` (profile parsing, band evaluation,
  config loading, formatting; table-driven; `-race`).
- **Round-trip golden:** no — Phase 00 ships no builder API (round-trip
  golden testing begins Phase 03 per master plan §1.8).
- **Integration:** no — Phase 00 has no `Deps` on another subsystem's
  shipped phase.
- **Fuzz:** no parse surface ships here (the coverage-profile parser is
  internal tooling, not an OOXML/OPC decode surface).
- **Benchmark:** none.

## 15. Vocabulary added

- `Coverage band` — the minimum per-package statement coverage enforced
  mechanically by `internal/coveragecheck` against `coverage.json`.
- `Drift audit` — the `scripts/drift-audit.sh` design-coherence gate.

(`Preflight` and `Smoke check` already exist in the glossary.)

## 16. Plan deviations encountered during implementation

- **Coverage-gate strictness ramp.** `CLAUDE.md §11` requires a new
  package with no configured band to fail the build. The pre-reorg tree
  has many un-banded upstream packages, so enforcing that now would fail
  CI wholesale. Shipped the gate with `require_configured=false` and
  filed `D-027` to ramp it to `true` in Phase 01. Re-states no acceptance
  criterion (criterion 5 only asserts the tooling band).
- **Pre-existing broken tests left for Phase 01.** See R1. `make test`
  is not green on `test/` and `test/parts/` due to stale `slide` imports;
  the rename does not introduce this and the fix belongs to Phase 01's
  reorg. Acceptance criteria 1–6 are met without it.

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-00.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass (none prior to Phase 00).
- [ ] Glossary updated.
- [ ] Decision entries added (`D-027`).
- [ ] (Phase 20+) Docs site updated — N/A (inert).
- [ ] (Phase 20+) Affected agent skill(s) updated — N/A (inert).
