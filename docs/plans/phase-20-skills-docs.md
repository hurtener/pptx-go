# Phase 20 — Agent skills + published docs site

**Subsystem:** skills, docs/site (+ examples/, a `scene/doc.go` fix)
**RFC sections:** `CLAUDE.md §19` (the binding DX spec), `RFC §5` (distribution)
**Deps:** Wave 6 complete (Phases 18–19 merged; the read model + best-effort
external read are the last user-facing surface the skills must describe).
**Status:** In progress

---

## 1. Goal

Ship the two developer-experience artifacts `CLAUDE.md §19` mandates — a set of
**agent skills** (`skills/`) that teach an AI coding agent the *complete*
pptx-go surface, and a **published docs site** (`docs/site/`) built and deployed
by CI — and turn on the §19 drift hook that keeps both in lockstep with the code.

## 2. Why now

Phase 20 opens Wave 7 (docs, skills, release). It depends on Wave 6 because the
skills/docs must describe the finished user-facing surface, including the read
model (D-047), best-effort external read + `ReadWarnings` (D-048), the read-path
security bounds and options (D-049), and notes round-trip (D-050) — all of which
landed in Wave 6. Phase 21 (release) depends on this phase: a v0.1.0 release
without docs/skills is not shippable. `CLAUDE.md §19` states the rule is **inert
until Phase 20 lands**; this phase establishes both artifacts and turns the rule
on.

This phase is also the first where the engine's intended role — **the backbone
of a downstream product** — directly shapes scope: the skills are the contract a
product-building agent composes against, so they must be *exhaustively accurate*
on engine characteristics, not a happy-path subset.

## 3. RFC sections implemented

- `CLAUDE.md §19` — both halves: the `skills/` SKILL.md set and the published
  `docs/site/` GitHub Pages site, plus activation of the §19 drift-audit hook
  and the §14 pre-merge enforcement.
- `RFC §5` (Apache-2.0 / distribution) — the docs site states the license,
  module path, and import surface; no new runtime dependency is added (P4 — the
  docs toolchain is dev-only CI, never a library import).

## 4. Brief findings incorporated

No informing research brief — there is no DX/onboarding brief under
`docs/research/`. **For a docs/skills phase the informing source is
`CLAUDE.md §19` itself** (the binding spec for skills and the published site),
together with the `docs/design/THEME.md` token catalog and `docs/decisions.md`
(D-001…D-050), which the skills and site reference. This is the documented
"no informing brief — the spec is the source" case, not a drift omission.

## 5. Findings I'm departing from

**Departing from the master-plan Phase 20 entry's skill list.**
`docs/plans/README.md` (Wave 7) lists **six** skills (scaffold, load a brand
theme, compose a scene, extend the icon set, rasterize+embed a code block,
rasterize+embed a chart). `CLAUDE.md §19` lists **eight** workflows — the same
six plus **"define a Theme"** and **"register an asset"** (and names "load a
brand template" explicitly). `CLAUDE.md` is the more complete and authoritative
DX spec, so this phase ships the **eight-skill** set and reconciles the
master-plan entry to match in this PR. No RFC conflict (the RFC defers DX detail
to `CLAUDE.md §19`).

## 6. Decisions referenced

- `D-001` (P1–P4 binding properties) — the skills frame every workflow inside
  the two-layer model and the tokens-not-literals default.
- `D-011`/`D-018` (per-node render policy) — the scene + asset skills explain
  native-vs-`pic` policy as intrinsic to node type.
- `D-024`/`D-036` (asset resolution, warn-don't-fail) — the *register-an-asset*,
  *embed-a-chart*, *embed-a-code-block* skills.
- `D-026` (engine, not product) — every skill states what the caller owns
  (content, rasters, fonts, policy) vs what the engine provides (mechanisms).
- `D-037` (brand-kit ingestion) — *load-a-brand-template*.
- `D-047`/`D-048`/`D-049`/`D-050` (read model, best-effort external read,
  read security bounds, notes round-trip) — the *scaffold* skill's read section.
- No **new** decision is required; the master-plan reconciliation (§5) is a plan
  fix, not a settled-decision change.

## 7. Architecture

Two artifacts plus a small fix, delivered across four PRs so the skills (the
priority) land and can be reviewed before the heavier site work.

```text
skills/                              # Agent Skills — agentskills.io SKILL.md format
  scaffold-a-presentation/SKILL.md   # pptx builder quickstart → valid deck
  define-a-theme/SKILL.md            # Theme + token taxonomy (P2)
  load-a-brand-template/SKILL.md     # FromTemplate brand-kit ingestion (D-037)
  compose-a-scene/SKILL.md           # scene IR catalog + Render + Stats
  embed-a-chart-raster/SKILL.md      # Chart node + AssetResolver (D-004)
  embed-a-code-block-raster/SKILL.md # CodeBlock node + AssetResolver (D-014)
  extend-the-icon-set/SKILL.md       # WithIconExtension + SVG constraints (D-040)
  register-an-asset/SKILL.md         # AssetResolver / AssetID / URIAssetResolver (D-024)
  README.md                          # index: which skill for which task

examples/                            # runnable Go programs each skill links to
  <one per skill>/main.go            # compiles + runs in the skill smoke

docs/site/                           # VitePress static site (dev-only toolchain)
  index / quickstart / api-reference / scene-catalog (page per node) /
  theme-guide / examples / decisions (links docs/decisions.md)

.github/workflows/pages.yml          # build + deploy the site to GitHub Pages
scripts/drift-audit.sh               # §19 hook flips from SKIP to ACTIVE
scene/doc.go                         # FIX — stale "Render is a no-op stub" claim
```

**Skill format.** Each `skills/<name>/SKILL.md` follows the agentskills.io
convention: YAML frontmatter (`name`, `description` — the *when to use*), then a
body that (a) states the relevant binding properties, (b) gives the exact public
API with signatures, (c) shows a complete, compiling example, (d) lists the
gotchas and the caller-owns-this boundary (D-026), and (e) cross-links sibling
skills. The body is verified against source — every identifier must exist.

**Site composition.** VitePress composes the in-repo Markdown; the scene catalog
has one page per IR node (acceptance criterion) with a runnable example. The
toolchain is a CI-only `node`/VitePress build — it never enters the Go module
(P4 preserved).

## 8. Files added or changed

```text
# PR#1 — plan + reconcile + smoke skeleton (this PR)
docs/plans/phase-20-skills-docs.md   # NEW (this file)
docs/plans/README.md                 # CHANGED — reconcile Phase 20 skill list to 8 (§5)
scripts/smoke/phase-20.sh            # NEW — phase smoke (criteria SKIP until built)

# PR#2 — the eight skills + runnable examples (PRIORITY)
skills/<eight>/SKILL.md              # NEW
skills/README.md                     # NEW — skill index
examples/<per-skill>/main.go         # NEW — runnable, compiled by the smoke
scene/doc.go                         # CHANGED — fix stale Render-stub doc
docs/glossary.md                     # CHANGED — "agent skill", "skill smoke"

# PR#3 — docs site + CI
docs/site/**                         # NEW — VitePress site
.github/workflows/pages.yml          # NEW — build + deploy on push to main

# PR#4 — activate the §19 drift hook
scripts/drift-audit.sh               # CHANGED — §19 hook ACTIVE (skills/docs exist)
AGENTS.md / CLAUDE.md                # CHANGED only if §19 wording needs a tweak (mirror)
```

No new exported Go symbol — skills/docs **describe** the existing surface; the
only code change is the `scene/doc.go` doc-comment fix.

## 9. Public API surface

None added. This phase documents the surface shipped through Phase 19. The
`scene/doc.go` change is a package doc-comment correction (the stale claim that
`Render` is a no-op stub), not an API change.

## 10. Risks

- **R1 — Skill drift / inaccuracy.** A skill that names a non-existent symbol or
  omits capability misleads a product-building agent (the explicit failure mode
  for this phase). **Mitigation:** every skill's API and example is verified
  against current source; each skill links a runnable `examples/` program that
  the phase-20 smoke **compiles and runs** (`go build`/`go run`), so an outdated
  identifier fails CI. The §19 drift hook (PR#4) keeps them in sync thereafter.
- **R2 — Docs toolchain leaks into the module (P4).** **Mitigation:** the
  VitePress build lives only in `pages.yml` + `docs/site/`; no Go file imports
  it; `make build` stays `CGO_ENABLED=0` and stdlib-only. `drift-audit` already
  asserts no non-stdlib runtime import.
- **R3 — §19 vocabulary bleed.** The site/skills must not use contributor-only
  "Phase N" vocabulary (§19). **Mitigation:** the drift-audit §19 hook (PR#4)
  enforces this mechanically on `docs/site/**`, `skills/**`, `examples/*/README`.
- **R4 — Turning the §19 hook on breaks unrelated PRs.** **Mitigation:** the
  hook activates only after skills/docs exist (PR#4), and scopes enforcement to
  PRs that touch `pptx/`/`scene/` user-facing surface — a doc-only or
  internal-only PR is unaffected.

## 11. Acceptance criteria

1. Eight skills exist under `skills/` in valid SKILL.md format (parseable
   frontmatter with `name` + `description`); a `skills/README.md` indexes them.
2. Each skill links a runnable example under `examples/`; every example
   `go build`s and `go run`s clean (the skill smoke).
3. Every shipped scene IR node has a docs-site catalog page with a runnable
   example.
4. The docs site builds cleanly (`vitepress build`) in CI and deploys to GitHub
   Pages via `.github/workflows/pages.yml`.
5. The §19 drift-audit hook is ACTIVE: a PR that changes `pptx/`/`scene/`
   user-facing surface without a matching `skills/`/`docs/site/` update fails
   `make drift-audit`; "Phase N" vocabulary on user-facing paths fails it.
6. `make preflight` passes; `scripts/smoke/phase-20.sh` reports `OK ≥ count` and
   `FAIL = 0`; prior phases' smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| (none) | — | No new Go package. `examples/` are `package main` programs exercised by the smoke (build+run), not unit-coverage-gated; `internal/coveragecheck` is unchanged. The `scene/doc.go` change is comment-only. |

## 13. Smoke check

`scripts/smoke/phase-20.sh` (criteria flip to OK across PRs):

1. `OK:` library builds CGo-free (regression guard).
2. `OK:` eight `skills/<name>/SKILL.md` exist with `name:`/`description:`
   frontmatter (PR#2).
3. `OK:` every `examples/*/main.go` `go build`s and `go run`s clean (PR#2).
4. `SKIP→OK:` `docs/site/` builds (where the node toolchain is available; SKIP
   without it, as the codec schema smoke does) (PR#3).
5. `SKIP→OK:` the §19 drift hook is active in `drift-audit.sh` (PR#4).

## 14. Tests

- **Unit:** none (no new Go package).
- **Round-trip golden:** n/a (no builder API added).
- **Integration:** n/a.
- **Fuzz:** n/a.
- **Smoke (build+run):** the examples are the primary mechanical check — each
  `examples/*/main.go` is compiled and run by `scripts/smoke/phase-20.sh`, so a
  skill that drifts from the API fails CI.
- **Docs build:** `vitepress build` in CI (PR#3) is the site's correctness gate.

## 15. Vocabulary added

- `agent skill` — a `skills/<name>/SKILL.md` (agentskills.io format) teaching an
  AI agent one pptx-go workflow; binding repo hygiene to keep in sync (§19).
- `skill smoke` — the phase-20 check that compiles and runs each skill's linked
  `examples/` program, so a skill cannot silently drift from the API.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages (n/a — no new package).
- [ ] `scripts/smoke/phase-20.sh` reports `OK ≥ count(criteria)` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (none — master-plan reconciliation only).
- [ ] Docs site updated for user-facing surface changes (this phase ships it).
- [ ] Affected agent skill(s) updated (this phase ships them).
