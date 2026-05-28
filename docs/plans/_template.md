# Phase NN — {{ slug-with-spaces }}

> Replace `NN` and the slug. Fill every section. "Brief findings
> incorporated" and "Findings I'm departing from" are forcing functions —
> they make brief inheritance visible. Keep this template's structure;
> use the headings as written.

**Subsystem:** {{ owning subsystem from `RFC §3.3` }}
**RFC sections:** {{ comma-separated, e.g. `§7, §8.4` }}
**Deps:** {{ prior phases or external prereqs; `none` if foundational }}
**Status:** Draft / In progress / Done

---

## 1. Goal

One sentence. What this phase delivers, in user-visible terms. If the
sentence needs hedging, the phase is two phases.

## 2. Why now

Why this phase is the next phase. Reference the master plan
(`docs/plans/README.md`) and any decision (`D-NNN`) that motivates the
order.

## 3. RFC sections implemented

List the RFC sections this phase materializes. If a section is partially
implemented (split across phases), say so and name the sibling phase.

- `RFC §X.Y` — { what part of §X.Y this phase implements }
- `RFC §X.Z` — { … }

## 4. Brief findings incorporated

For each research brief cited:

- `docs/research/NN-slug.md` — `{ finding }` → `{ how this plan
  honours it }`.

A phase that cites **no** brief is a drift signal. Either name a brief or
say explicitly *"no informing brief — this is foundational scaffolding
with no prior-art investigation needed"*.

## 5. Findings I'm departing from

Any brief finding this plan rejects, weakens, or supersedes. Include the
rationale; don't bury a departure in code.

- `docs/research/NN-slug.md` — `{ finding }`. **Departing because** …

If none, write *"none"*.

## 6. Decisions referenced

Decisions in `docs/decisions.md` this plan relies on or extends:

- `D-NNN` — `{ title }` — `{ how it applies }`

If this plan creates a new settled decision, file the entry in
`docs/decisions.md` in the same PR and reference it here.

## 7. Architecture

A short description of the shape of the work. Module layout, exported
types, key interfaces, the seam this phase opens or closes. If a diagram
helps, ASCII it inline.

```text
{ optional ASCII sketch }
```

## 8. Files added or changed

```text
{ file tree of new and changed paths, with one-line annotations }

pptx/foo.go                   # NEW — Foo, NewFoo, FooOption
pptx/bar.go                   # CHANGED — adds BarMode
internal/ooxml/baz/baz.go     # NEW — Baz codec
scripts/smoke/phase-NN.sh     # NEW — phase smoke
docs/decisions.md             # CHANGED — adds D-NNN
```

Anything user-facing? Doc site (Phase 21+) and skills (Phase 21+) updates
land **in this PR** per `CLAUDE.md §19`.

## 9. Public API surface

Every new or changed exported symbol, with a one-line description. This
is the API the rest of the project will compose against.

```go
// pptx
type Foo struct { ... }
func NewFoo(opts ...FooOption) *Foo
func (f *Foo) Bar(x int) (*Bar, error)
```

If a symbol breaks a prior public surface, call it out and either
preserve a deprecation alias (preferred for v0.x) or document the break
in `CHANGELOG.md`.

## 10. Risks

Known unknowns and how the plan handles them.

- **R1 — { name }** — { description }. **Mitigation:** { … }.
- **R2 — { name }** — …

If a risk's mitigation is "we'll find out during implementation", the
phase isn't ready to land — refine the plan first.

## 11. Acceptance criteria

The binding list. The smoke script verifies these mechanically. Write
them as **observable, testable assertions**, not as work-items.

1. { e.g. "A `Foo` with a `BarMode` round-trips losslessly through
   `pptx.Open`." }
2. { e.g. "`make coverage` shows the new packages ≥ their bands." }
3. …

## 12. Coverage targets

Per-package, overriding `CLAUDE.md §11` defaults if needed.

| Package | Target | Rationale (if override) |
|---|---|---|
| `pptx/foo` | 85% | default for new pptx package |
| `internal/ooxml/baz` | 85% | codec band |
| `scene/bar` | 80% | default for new scene package |

If lowering a default, file a decision (`D-NNN`) explaining why.

## 13. Smoke check

`scripts/smoke/phase-NN.sh` verifies each acceptance criterion. Outline
the checks here (the script implements them):

1. `OK:` Build the binary / a sample.
2. `OK:` Run the example for criterion 1.
3. `OK:` Run the round-trip test for criterion 2.
4. …

Use `SKIP:` for criteria whose surface isn't built yet (the smoke script
must run cleanly before this phase lands — `SKIP` is fine, `FAIL` blocks
the merge).

## 14. Tests

What kinds of tests this phase ships:

- **Unit:** { which packages }
- **Round-trip golden:** { yes/no — Phase 03+ is yes by default }
- **Integration** (`test/integration/`): { yes/no — required if `Deps`
  names a different subsystem's shipped phase, or this phase closes a
  seam another phase opened }
- **Fuzz** (`FuzzXxx`): { for any parse/decode surface — required if
  this phase ships one }
- **Benchmark** (`BenchmarkXxx`): { for any hot reusable artifact —
  optional but encouraged }

## 15. Vocabulary added

New terms — file each in `docs/glossary.md` in this PR, alphabetical
order. List them here:

- `Foo` — { one-line description }
- `Bar` — { … }

## 16. Plan deviations encountered during implementation

Filled in **as** implementation happens — don't pre-populate. Each
deviation: what changed, why, and which acceptance criterion is now
re-stated. Silent divergence is drift (`CLAUDE.md §4.3`).

- *(empty until implementation)*

## 17. Sign-off

When the phase is done, fill the section below. The PR cannot merge
without this section completed.

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-NN.sh` reports `OK ≥ {count(criteria)}` and
      `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (if any).
- [ ] (Phase 21+) Docs site updated for user-facing surface changes.
- [ ] (Phase 21+) Affected agent skill(s) updated.

---

*Replace this footer with the actual phase plan. Keep the section
headings.*
