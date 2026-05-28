# pptx-go — Contributor & Agent Normatives

> This file is **binding** for anyone — human or AI — modifying this
> repository. It is mirrored **verbatim** in `AGENTS.md` so any
> spec-compliant coding agent picks it up automatically. If the two files
> diverge, the most recent commit timestamp wins; flag the drift in your
> PR.

If a rule below conflicts with the RFC or a phase plan, the **RFC wins**,
then the **phase plan**, then this file. Update whichever artifact is
wrong; never silently ignore the conflict.

---

## Starting a new session — orientation (READ THIS FIRST)

pptx-go is a multi-phase, doc-driven build. The design surface is large on
purpose: hygiene up front is cheaper than retrofitting it. Before
substantive work, skim, in order:

1. **§1 — What pptx-go is.** The product and its four binding properties.
2. **§2 — Authoritative sources.** The priority chain: RFC > phase plans
   > this file > research briefs > code comments.
3. **§16 — Authoring a phase plan.** The binding workflow for any
   contributor touching a phase. Skipping it is the single largest source
   of design drift.

**Drift-hygiene artifacts (live references):**

- `RFC-001-pptx-go.md` — the design source of truth.
- `docs/decisions.md` — append-only log of settled decisions (`D-NNN`).
  When tempted to re-litigate something, grep here first.
- `docs/glossary.md` — pptx-go vocabulary. New terms land here in the
  same PR.
- `docs/research/INDEX.md` — subsystem → research-brief reverse index.
- `docs/plans/_template.md` — phase plan template; new phases start as a
  copy.
- `scripts/drift-audit.sh` — mechanical drift checks (`make drift-audit`).

If asked to do something that doesn't fit a phase (a one-off fix, a
question, a small doc edit), proceed without the full §16 ritual — but
mention any drift risk you spot.

---

## 1. What pptx-go is

pptx-go is a **pure-Go library for authoring and reading PowerPoint
(PPTX, Open Office XML) files**. Module path
`github.com/hurtener/pptx-go`. Apache-2.0. Pre-V1: v0.x. Public packages:
`pptx` (Layer 1, the builder) and `scene` (Layer 2, the optional scene
renderer).

The library has two public layers:

```text
Layer 2 — scene   (IR-driven scene renderer; composes the builder)
Layer 1 — pptx    (general-purpose, theme-aware PPTX builder)
            ↑
            │   (raw OOXML lives below the public surface)
            ↓
internal/  (opc, ooxml, render, ids, coveragecheck — PRIVATE)
```

A direct consumer of `pptx` writes generic Go and gets a production-grade
deck. A consumer of `scene` builds a typed `Scene` (an IR modeled on
pengui-slides v4) and gets a PPTX with no opinionated PowerPoint
boilerplate. The two consumer paths are independent: `pptx` doesn't know
the IR; `scene` doesn't reach under the builder.

**Four binding properties.** A change that weakens any of them is wrong —
reach for the RFC, not the keyboard.

1. **P1 — Two layers, one library.** `pptx` (builder) and `scene`
   (renderer) are the only two public layers. `scene` composes `pptx`;
   nothing in `scene` reaches under it. A new scene primitive adds a new
   builder call **only** when the primitive requires a genuinely new
   OOXML capability.
2. **P2 — Tokens, not literals.** All visual properties (color,
   typography, spacing, radius, elevation) flow through a `Theme` whose
   semantic tokens map to OOXML values. The token path is the documented
   default; literals are an escape hatch. A theme swap re-renders the
   same builder/scene input in the new visual language.
3. **P3 — OOXML by isolation.** Raw OOXML wire types — XML structs,
   element names, namespace URIs, schema specifics — live only in
   `internal/ooxml`. Neither `pptx` nor `scene` imports raw XML structs.
4. **P4 — No CGo, stdlib-only runtime.** The shipped artifact compiles
   `CGO_ENABLED=0`; runtime imports are stdlib only. `-race` tests are
   the one CGo exception (test binaries are not shipped). Curated
   `go:embed` assets are vendored bytes, not dependencies.

---

## 2. Authoritative sources (in priority order)

1. `RFC-001-pptx-go.md` — product intent and design decisions.
2. `docs/plans/phase-NN-*.md` — implementation specifications.
   Acceptance criteria are binding.
3. `docs/plans/README.md` — the master phase plan: cross-cutting
   conventions and the phase index.
4. This file (`AGENTS.md` / `CLAUDE.md`) — operational rules.
5. `docs/research/*.md` — phase-planning research briefs. Authoritative
   for *context*, not for design — the RFC and phase plans are where
   decisions land.
6. Code comments and godoc — last and least authoritative.

When a phase plan and the RFC drift, the RFC wins. File a follow-up to
fix the plan.

---

## 3. Repository layout

```text
.
├── RFC-001-pptx-go.md          # design RFC — source of truth
├── README.md                   # user-facing intro (no internal vocabulary — §19)
├── CHANGELOG.md                # release notes (Keep a Changelog)
├── AGENTS.md / CLAUDE.md       # this file (verbatim copies)
├── LICENSE                     # Apache-2.0
├── LICENSE.upstream            # MIT — upstream's license, preserved
├── Makefile                    # canonical build / test / lint commands
├── go.mod / go.sum / go.work
├── .github/                    # CI, PR template, codeowners, dependabot, release (Phase 20+)
├── .golangci.yml / .editorconfig / .gitignore
│
├── pptx/                       # Layer 1 — the builder (PUBLIC)
│   ├── presentation.go         # Presentation: New / Open / Save / Stream
│   ├── slide.go                # Slide builder
│   ├── shape.go                # Shape primitives + geometry
│   ├── text.go                 # RichText, Paragraph, Run, AutoFit
│   ├── table.go                # Table builder
│   ├── media.go                # Image / media insertion + dedup
│   ├── theme.go                # Theme, Token, ColorRole, TypeRole, …
│   ├── tokenresolve.go         # Token → OOXML value at write time
│   ├── units.go                # EMU / Pt / Cm / In / Px
│   ├── geom.go                 # Box, Position, Size, Anchor, Inset
│   ├── stream.go               # Streaming pass-through to internal/opc
│   └── doc.go
│
├── scene/                      # Layer 2 — the scene renderer (PUBLIC)
│   ├── scene.go                # Scene, SceneSlide, Render, Stats, RenderOption
│   ├── nodes.go                # IR catalog (leaves + containers)
│   ├── richtext.go             # Scene-side RichText
│   ├── tokens.go               # ColorRole / TextColorRole / TypeRole re-exports
│   ├── validate.go             # Stage 1 + Stage 2 validation
│   ├── policy.go               # Per-node rendering policy doc + test (per RFC §12)
│   ├── layout/                 # Layout engine (two_column, grid, card, flow, decoration)
│   ├── icons/                  # Curated icon registry (closed-name + caller extension)
│   ├── ornaments/              # Curated ornament registry
│   ├── frames/                 # Curated frame chrome registry
│   └── doc.go
│
├── internal/                   # PRIVATE — Go enforces non-importability
│   ├── opc/                    # OPC package: ZIP, content-types, rels, pack URIs
│   ├── ooxml/                  # OOXML XML structs (P3 isolation seam)
│   │   ├── namespaces.go       # canonical NS URIs
│   │   ├── presentation/       # presentation.xml, presProps, viewProps
│   │   ├── slide/              # slide, slideLayout, slideMaster
│   │   ├── theme/              # theme1.xml
│   │   ├── core/               # core.xml, app.xml
│   │   ├── drawing/            # drawingML shapes, fills, text bodies, geometries
│   │   ├── chart/              # c:chart wire types (V2 — placeholder in V1)
│   │   ├── relations/          # relationship XML structs
│   │   └── media/              # media-part typing
│   ├── render/                 # Builder-internal helpers (e.g. text body XML generation,
│   │                            #   SVG path → OOXML translator)
│   ├── ids/                    # Shape-id / rel-id atomic allocation
│   └── coveragecheck/          # Mechanical coverage band gate + coverage.json
│
├── assets/                     # Embedded curated assets (go:embed)
│   ├── icons/                  # Lucide-subset SVGs
│   ├── ornaments/              # Preset ornament Go recipes
│   └── frames/                 # Device frame Go recipes
│
├── templates/                  # Starter .pptx templates for scene ingestion
│   └── _default-theme.pptx     # Built into the binary; loaded when no theme is set
│
├── skills/                     # Agent Skills — one SKILL.md per workflow (§19, Phase 20+)
├── examples/                   # Runnable Go examples (1+ per scene node)
├── test/integration/           # Cross-layer integration tests
│
├── scripts/
│   ├── preflight.sh            # the preflight gate
│   ├── drift-audit.sh          # design-coherence checks (incl. §19 hook)
│   ├── smoke/                  # per-phase smoke scripts
│   │   └── _template.sh        # smoke skeleton
│   ├── hooks/pre-commit
│   └── install-hooks.sh
│
└── docs/
    ├── plans/                  # master plan (README.md) + phase plans + _template.md
    ├── research/               # phase-planning research briefs + INDEX.md
    ├── specifications/         # vendored OOXML / OPC spec snapshots
    ├── design/                 # token catalog, per-node rendering policy, theme spec
    ├── site/                   # published tech-docs site — VitePress (§19, Phase 20+)
    ├── screenshots/            # in-repo screenshots referenced by docs + PR bodies
    ├── release/                # in-tree release dry-run transcripts (Phase 21)
    ├── V2-BACKLOG.md           # consolidated post-V1 deferrals (Phase 21)
    ├── RELEASING.md            # release procedure for maintainers (Phase 21)
    ├── decisions.md            # append-only D-NNN log
    └── glossary.md
```

Directories are created as the phases that own them land. Anything that
doesn't have a home above is wrong — if you need a new top-level
directory, propose it in the RFC first; `§3` is the binding layout.

---

## 4. Build, test, lint, run

All targets are canonical and run by CI. Targets no-op gracefully before
the code they act on exists.

```bash
make build         # build/check the library compiles (CGo-free)
make test          # go test -race ./...
make coverage      # per-package coverage profile + the mechanical band gate
make bench         # run the Go benchmarks (on demand — not a CI gate)
make vet           # go vet ./...
make lint          # golangci-lint run
make drift-audit   # design-coherence checks (RFC/plans/briefs/mirror/forbidden names)
make check-mirror  # verify AGENTS.md == CLAUDE.md
make preflight     # build + smoke checks + drift-audit
make install-hooks # install the pre-commit hook (one-time, per clone)
```

There is no `make web` target — pptx-go has no frontend (Phase 20's docs
site is built by its own CI workflow under `.github/workflows/pages.yml`,
not the Makefile).

### 4.1 Preflight gate — non-negotiable

`make preflight` is the same gate the pre-commit hook and CI enforce: it
builds, runs every per-phase smoke script (which SKIP gracefully where
the surface isn't built yet), and runs `drift-audit`. Do not bypass the
pre-commit hook with `--no-verify` outside a documented emergency.

### 4.2 Phase implementor contract

A phase is **done** only when: (a) every acceptance criterion in its
plan passes; (b) coverage targets for touched packages are met;
(c) `scripts/smoke/phase-NN.sh` reports `OK ≥ count(criteria)` and
`FAIL = 0`; (d) prior phases' smoke scripts still pass. A new public API
on `pptx` or `scene`, or a new scene IR node ⇒ a smoke check in the
**same** PR. A new theme token / IR field ⇒ documented in the plan, the
glossary, and a smoke check. Once the agent skills and published docs
exist (Phase 20, §19), a PR that changes user-facing surface also updates
the affected skill(s) and docs in the same PR.

### 4.3 Reasonable plan deviations

Plans are specifications, not straitjackets. A reasonable deviation
discovered during implementation is fine — document it in the PR
description and update the plan file **in the same PR**. Silent
divergence from a plan or the RFC is drift.

### 4.4 Extensibility seams (project-wide policy)

Any subsystem with a plausible alternate backend lives behind an
**interface + factory + driver** pattern. V1 mandates this for: the
`Color` interface (token vs literal), the `ImageSource` interface
(file / bytes / reader), the curated registries (icons / ornaments /
frames — closed-name plus caller extension), and the scene asset
registry (caller plugs in `AssetID` resolvers).

V1 does **not** mandate this for the `Theme` itself (a single `*Theme`
type is the seam; alternate theme sources are constructors, not driver
implementations) or for any persistent store (there is no persistent
store — pptx-go is stateless, §9).

---

## 5. Code conventions (Go)

- **Toolchain.** Go 1.24, pinned in `go.mod`. **No CGo in the shipped
  artifact** — `make build` pins `CGO_ENABLED=0`; a runtime dependency
  that needs CGo is rejected. Test binaries are the one exception:
  `make test` runs with `CGO_ENABLED=1` because the `-race` detector
  requires CGo — tests are not shipped, so this does not weaken the
  CGo-free guarantee.
- **No third-party runtime deps.** The shipped artifact imports stdlib
  only. A PR that adds a non-stdlib import to `pptx`, `scene`, or
  anything under `internal/` is rejected (P4). `go:embed`'d bytes are
  not imports.
- **Style.** `gofmt -s`; `go vet` and `golangci-lint run` clean.
  Generated code (none in V1) would be marked with a `// Code generated
  … DO NOT EDIT.` header and would stay boring and readable.
- **Errors.** `errors.Is`/`errors.As`, `%w` wrapping, sentinel errors,
  `errors.Join`. Wrap with context. **Never `panic` for control flow**
  and never panic across a public API boundary.
- **Context.** `context.Context` is the first parameter of any call
  that does I/O, blocks, or can be canceled. Honor cancellation.
- **Logging.** `log/slog` only — no `log.Printf`, no `logrus`/`zap`.
  Callers inject the logger via `pptx.WithLogger` / `scene.WithLogger`;
  there's no global logger. No unredacted secrets in logs.
- **Concurrency.** Race detector mandatory on tests. A reusable artifact
  (a theme, an asset registry, a master) must be safe under concurrent
  use; prove it.
- **Tests.** Table-driven where it fits; golden tests for OOXML codec
  output and round-trip; `-race` always.
- **XML.** Stdlib `encoding/xml`. Strict defaults (no external entity
  resolution).
- **JSON.** Stdlib `encoding/json` (v1). `encoding/json/v2` is deferred.

---

## 6. The non-negotiable product rules

These enforce P1–P4 (§1). They are binding on every phase.

- **Two layers, one library (P1).** `scene` imports `pptx`; `pptx` does
  not import `scene`. `pptx` does not import any `scene/...` subpackage,
  including `scene/icons` and `scene/ornaments` — those compose with
  `pptx` via the builder API, never the reverse. New scene primitives
  add new builder capability only when a genuinely new OOXML feature is
  needed; otherwise they're composers.
- **Tokens, not literals (P2).** Builder APIs that take a color/font/
  spacing/etc. accept a token type. Literal escape hatches exist
  (`pptx.RGB`, `pptx.Pt`) but the *documented* path in godoc, examples,
  skills, and docs site is **tokens**. A new visual property added to
  the builder must have a token taxonomy entry (`docs/design/THEME.md`)
  in the same PR.
- **OOXML by isolation (P3).** Raw OOXML wire types live **only** in
  `internal/ooxml`. Handler-facing and scene-facing APIs never expose
  raw XML structs (no `*xml.Encoder`, no `xml.Name`, no XML tag types
  in `pptx`/`scene` signatures). A schema change is one PR localized to
  the affected `internal/ooxml/*` subpackage.
- **No CGo, stdlib-only runtime (P4).** Every artifact compiles
  CGo-free and cross-compiles. The race detector is the lone
  test-binary exception.
- **Per-node rendering policy, not per-deck (D-011, D-018).** A scene
  node renders as either native PPTX shapes or a `pic` shape with
  caller-supplied bytes; the choice is intrinsic to the node type
  (whether its IR carries an `asset_id` field). No deck-wide mode
  toggle, no `Disposition` enum, no per-deck "raster everything"
  switch — product behavior lives in callers (D-026).
- **Round-trip fidelity for self-authored decks (G6).** Every shape /
  text run / fill / line / table / image pptx-go emits round-trips
  losslessly through `pptx.Open`. A phase that adds builder API ships a
  round-trip golden test in the same PR.
- **Single-module distribution (D-010).** `scene` is a subpackage of
  `github.com/hurtener/pptx-go`, not a separate module.
- **Engine, not product (D-026).** pptx-go converts a typed scene IR
  into PPTX and nothing else. Product behavior — render modes,
  legibility heuristics, validation pipelines, markdown ingestion,
  comments, recipes — lives in callers. A proposed feature that
  encodes an opinion about *what* the deck should contain, *how it
  should look*, or *what audience it's for* is rejected; a proposed
  feature that exposes a *mechanism* the caller can drive (asset
  resolution, font embedding, theme tokens, slide grouping) is
  considered on its merits.

---

## 7. Security — non-negotiable rules

- No hardcoded secrets, anywhere — including generated code and tests.
- **Image bytes** are embedded verbatim. We verify content-type matches
  the declared MIME and reject obviously malformed bytes (e.g. missing
  PNG signature), but we do **not** parse pixel data. A malicious image
  is the caller's problem at *display* time.
- **Hyperlinks** are URL strings emitted verbatim into the OOXML
  relationship. We do **not** fetch or validate URLs at write time.
  Callers sanitize.
- **XML / XXE.** Parsers use `encoding/xml` with strict defaults — no
  external entity resolution, no XInclude. Any code path that toggles
  these is forbidden.
- **Zip-slip.** The streaming open path validates that every part path
  stays within the package root; absolute or `..` paths are rejected at
  parse time.
- **Memory bounds on read.** A part exceeding the caller-configurable
  per-part limit (default 100 MB) is rejected with `ErrPartTooLarge` —
  not allocated.
- **`scene.WithIconExtension` and siblings** accept caller bytes. SVG
  bytes go through the same translator constraints (single path, solid
  fill, no gradients). A non-conforming SVG fails at registration, not
  at render — registration is the safe spot to fail.

---

## 8. Observability — slog + Stats, not a protocol

- pptx-go is a **library**, not a service. There is no `obs/v1`-style
  event-stream protocol (D-016).
- The runtime emits structured events via an optional `*slog.Logger`
  injected by the caller (`pptx.WithLogger`, `scene.WithLogger`).
  No logger = no logs.
- `scene.Render` returns a `Stats` struct: per-slide render time, shape
  counts, asset counts, warnings. Callers integrate this into their own
  telemetry.
- Non-fatal layout / token / asset issues surface in
  `Stats.Warnings []LayoutWarning`. A V1.x `strict` render option turns
  warnings into errors.
- Emit paths are **non-blocking**: a slow logger never stalls the
  renderer. If `slog.Logger.Handler.Handle` blocks, it's the caller's
  problem.

---

## 9. Persistence — there is none

pptx-go is **stateless**. It reads a file (or `io.Reader`), produces a
file (or `io.Writer`), and holds no state between calls. There is no
`Store` interface, no migrations, no database.

The upstream's "streaming open" — lazy-load a `*Presentation` and write
modifications back without materializing every part — is preserved
under `pptx.OpenStream` / `pptx.SaveStream` (RFC §17.2). Streaming is an
I/O mode, not a persistence layer.

---

## 10. Forward-compatibility — the `internal/ooxml` rules

- Every OOXML spec pptx-go implements against is vendored into
  `docs/specifications/`, pinned by ISO edition + date. A spec re-read
  is a deliberate, reviewed update + a codec PR + golden re-pin in the
  same change.
- `internal/ooxml` is the only package that imports raw XML wire types.
  The builder and scene renderer consume *Go domain objects* from this
  package, never `xml.Marshaler`/`xml.Unmarshaler` instances.
- V1 codecs are **single-version**: no per-`protocolVersion` branching.
  The multi-version codec pattern is reserved for V2 (and introduced
  only when a real compatibility case forces it, e.g. PowerPoint Online
  emitting a new chart-XML shape).
- A new OOXML feature pptx-go consumes is a vendored spec snapshot +
  a codec PR — never an ad-hoc XML emission in the upper layer.

---

## 11. Testing rules

- `-race` on every test run. CI fails on a race.
- **Round-trip golden tests** are the primary correctness test from
  Phase 03 onward: write → read → assert model equality. Every shipped
  scene node, every shipped builder primitive has one.
- **Codec golden tests**: fixed input → fixed XML output (with stable
  attribute ordering). A codec change updates goldens in the same PR
  with a one-line rationale in the commit message.
- **Spec compliance** is tested against the vendored specs, not against
  live PowerPoint. PowerPoint compatibility is tested manually on
  reference decks (one per wave).
- **Integration tests** (`test/integration/`) ship per `CLAUDE.md §17`:
  whenever a phase's `Deps` name a different subsystem's shipped phase,
  closes a seam another phase opened, or introduces a public interface
  other phases will build on.
- Coverage defaults (override per phase): 85% new `pptx` packages, 80%
  new `scene` packages, 85% `internal/opc` and `internal/ooxml/*`
  codecs, 80% `internal/render`/`internal/ids`, 70% CLI/tooling (V1:
  none).
- **The coverage bands are a mechanical gate, not an aspiration.**
  `make coverage` runs the per-package coverage checker
  (`internal/coveragecheck`) against the thresholds in
  `internal/coveragecheck/coverage.json`; CI runs it and a coverage
  regression — or a new package with no configured threshold — fails
  the build. A package below its band is raised by adding tests; a
  band genuinely unreachable hermetically gets a documented override
  (class + reason) in the config and a decision entry — never a silent
  lowering.
- **Fuzz targets** (`FuzzXxx`) cover parse / decode surfaces in
  `internal/ooxml` and `internal/opc` with a seed corpus and an
  asserted invariant; the corpus runs as an ordinary CI test.
- **Benchmarks** (`BenchmarkXxx`) cover hot reusable artifacts (theme
  resolution, single-slide render, 100-slide streaming render,
  `scene.Render` worker scaling). Run on demand via `make bench` — a
  baseline, not a CI gate.

---

## 12. Commit and PR conventions

- **Commits:** imperative mood, scoped (`feat(scene): …`, `fix(pptx): …`,
  `chore: …`, `docs: …`, `feat(ooxml/drawing): …`). Small and coherent.
- **Branches:** never commit feature work directly to `main`; use
  `feat/phase-NN-slug` (or `chore/*`, `docs/*`). Once the project is
  past scaffolding, do not modify `main` directly — use a worktree or
  branch.
- **PRs:** reference the RFC section(s) and the phase. State any plan
  deviation and update the plan in the same PR. The pre-merge checklist
  (§14) gates the PR.
- **Merge:** squash unless history is meaningful. CI green is mandatory.

---

## 13. Forbidden practices

- Hardcoded secrets, including in tests.
- `panic` for control flow; panicking across a public API boundary.
- A CGo runtime dependency, or building the shipped artifact with
  `CGO_ENABLED=1` (`-race` test runs use `CGO_ENABLED=1` and are
  exempt — §5).
- A non-stdlib runtime import in `pptx`, `scene`, or `internal/...`
  (violates P4).
- `scene` reaching under `pptx` to call `internal/...` or to construct
  raw OOXML (violates P1).
- Importing raw XML wire types (`encoding/xml` element types, OOXML
  namespace URIs, XML structs from `internal/ooxml`) outside
  `internal/ooxml` (violates P3).
- Adding a builder API for a visual property without an entry in the
  theme/token taxonomy (violates P2).
- Reading runtime internals to observe instead of using the
  `WithLogger` hook or returning `Stats` data — there is no other
  observability seam (violates §8).
- Hardcoding per-PowerPoint-version branches in a codec (the
  multi-version codec pattern is V2 only — §10).
- Adding a render-mode toggle, a legibility heuristic, a doc-mode IR
  node, or any other **product-style behavior** to the library
  (violates D-026). The library is the engine; product behavior
  lives in callers. Mechanisms that callers can drive (asset
  resolution, font embedding, theme tokens, slide grouping, speaker
  notes) are fine.
- Adding a public API on `pptx` / `scene`, or a new scene IR node,
  without a smoke check in the same PR.
- Changing user-facing surface without updating the affected agent
  skill(s) and the published docs in the same PR, once those exist
  (§19).
- Editing a vendored spec snapshot in `docs/specifications/` without
  also re-pinning the affected codec and goldens.
- Bypassing the pre-commit hook with `--no-verify` outside a documented
  emergency.

---

## 14. Pre-merge checklist

- [ ] `make drift-audit` passes.
- [ ] `make check-mirror` passes (`AGENTS.md` == `CLAUDE.md`).
- [ ] `make preflight` passes.
- [ ] `go test -race ./...` and `golangci-lint run` are clean.
- [ ] All cross-references (`RFC §X.Y`, `D-NNN`, `brief NN`) resolve.
- [ ] Coverage on touched packages ≥ the phase's stated target —
      `make coverage` passes (the mechanical band gate; a new package is
      added to `internal/coveragecheck/coverage.json` in the same PR).
- [ ] A new public API on `pptx` / `scene`, a new scene IR node, or a
      new manifest/spec field has a smoke check in this PR.
- [ ] If a reusable artifact changed (a theme, an asset registry, a
      master): a concurrent-reuse test passes under `-race`.
- [ ] If a cross-subsystem seam was opened or consumed: an integration
      test exists (§17).
- [ ] If a builder API for a visual property was added: a theme/token
      taxonomy entry was added in `docs/design/THEME.md`.
- [ ] If a phase ships builder API: a **round-trip golden test** lands
      in the same PR.
- [ ] New vocabulary added to `docs/glossary.md` in this PR.
- [ ] A new architectural decision (or a departure from a brief) is
      filed in `docs/decisions.md`.
- [ ] If user-facing surface changed and the agent skills / docs site
      exist (Phase 20+): the affected skill(s) and published docs are
      updated in this PR.

---

## 15. When in doubt

The RFC wins. If the RFC is silent, the phase plan decides; if both are
silent, raise it — do not invent a decision and bury it in code. A new
settled decision is an entry in `docs/decisions.md`; a change to a
settled decision is an RFC PR plus a superseding decision entry, never
a silent edit.

---

## 16. Authoring a phase plan (workflow)

The canonical workflow for any contributor starting a phase. The
drift-audit gate enforces what it can; this workflow covers what it
can't.

1. **Read the master plan entry.** Open `docs/plans/README.md`, find the
   Phase N detail block. Note owning subsystem, RFC sections,
   dependencies, risks.
2. **Read the cited RFC sections** in `RFC-001-pptx-go.md`.
3. **Read the relevant briefs** per `docs/research/INDEX.md`. A phase
   plan that cites no informing brief is a drift signal.
4. **Read the glossary** for any term you're unsure about; pre-write
   the entry for any new term you introduce.
5. **Read the decisions log** (`docs/decisions.md`) for entries touching
   this subsystem. Settled decisions are not re-litigated silently.
6. **Copy the template:** `cp docs/plans/_template.md
   docs/plans/phase-NN-slug.md`. Fill every section. "Brief findings
   incorporated" and "Findings I'm departing from" are forcing
   functions — they make brief inheritance visible.
7. **Author the smoke skeleton:**
   `cp scripts/smoke/_template.sh scripts/smoke/phase-NN.sh`.
8. **Run `make drift-audit` and `make preflight`** before committing.
9. **Commit only when both pass.** The PR references the RFC section
   and any superseded decision.

---

## 17. End-to-end + integration testing

Per-package unit tests miss two classes of bug: **cross-package wiring
gaps** (two phases each ship their half of a seam, neither connects
them) and **cross-subsystem concurrency interactions**.

A phase ships an integration test whenever its `Deps` name a different
subsystem's shipped phase, or it closes a seam another phase opened, or
it introduces a public interface other phases will build on.
Integration tests use **real drivers** on the seam (no mocks at the
boundary — real `internal/opc` writes, real `encoding/xml` decode,
real disk I/O via temp files), prove identity/capability propagation,
cover ≥1 failure mode, and run under `-race`. They live in-package
when the package *is* the wiring boundary, otherwise in
`test/integration/`.

At wave boundaries a read-only **checkpoint audit** reviews every
shipped phase for wiring gaps, RFC drift, weak tests, and hygiene
regressions, and lands its punch list as one `chore(checkpoint)` PR.
When an integration test surfaces a bug, fix it in the same PR — even
when the root cause is in an earlier phase.

---

## 18. Mirroring

`AGENTS.md` and `CLAUDE.md` are kept **verbatim identical**. After any
edit:

```bash
diff -q AGENTS.md CLAUDE.md   # expected: no output
```

CI enforces this; the `mirror` job fails the build if they differ.

---

## 19. Agent skills & published documentation

From **Phase 20** onward, pptx-go ships two developer-experience
artifacts, and **keeping them in sync with the surface is mandatory
repo hygiene** — drift in either is a defect, the same kind of defect
as RFC drift:

- **Agent skills** (`skills/`) — a set of Agent Skills in the
  `SKILL.md` format (agentskills.io conventions) that teach an AI
  coding agent how to build presentations with pptx-go: scaffold a
  presentation, define a Theme, load a brand template, compose a
  scene, embed a chart raster, embed a code-block raster, extend the
  icon set, register an asset. A developer building with pptx-go via
  an agent should be productive from day one.
- **Published technical documentation** — a GitHub Pages site, built
  and deployed by CI from the in-repo docs (`docs/site/`).

**The rule.** Any PR that adds or changes **user-facing surface** —
a public API on `pptx` or `scene`, a scene IR node, a theme token, a
curated asset, a template field — **updates the affected skill(s) and
the docs in the same PR.** A phase plan whose work touches user-facing
surface lists the skill/doc updates in its `Files added or changed`
section. The §14 pre-merge checklist enforces it.

**User-facing vocabulary.** pptx-go's internal phase-by-phase build
methodology is **contributor vocabulary**. It lives in `docs/plans/`,
`docs/decisions.md`, `docs/research/`, the RFC, `AGENTS.md`/`CLAUDE.md`,
the glossary, the design spec, Makefile/workflow comments, and internal
code. It **must not** bleed into user-facing surfaces — the root
`README.md`, `CHANGELOG.md`, `docs/site/**/*.md`, or
`examples/*/README.md`. User-facing surfaces describe what the library
*does* and *is*, not when it was built. "Phase N", "phase-N", and
similar wording is forbidden on those paths. `D-NNN` decision-log
citations are acceptable in `docs/site/**/*.md` and `examples/*/README.md`
(they cross-link the public decisions reference page). The
`drift-audit` script's §19 hook enforces this mechanically; a future
regression fails CI before merge.

Before Phase 20 lands, `skills/` and the docs site do not yet exist
and the rule is inert; Phase 20 establishes both and turns the rule on.

---

## 20. No frontend design system

pptx-go is a Go library. The only HTML surface is the published docs
site (`docs/site/`), built statically by VitePress. There is no SPA, no
component inventory, no shared design system in the Dockyard sense.

The published site is a static reference (quickstart, API docs, scene
catalog, theme guide, examples). It composes a small set of bespoke
components inside VitePress — extending that set is not "establishing a
design system", it's editing a docs site. The §20 design-system rules
from sibling projects (Dockyard) do **not** apply to pptx-go.

**The library's design system is the Theme** (§7 of the RFC). The
canonical theme/token catalog is `docs/design/THEME.md`. When a phase
adds a visual property to the builder, a token taxonomy entry lands in
the same PR — that's the pptx-go analogue of "compose the shared UI
inventory".
