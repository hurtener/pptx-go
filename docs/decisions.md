# Decisions

> Append-only log of settled architectural decisions. Each entry has an
> immutable ID (`D-NNN`) and a date. A new decision is a new entry; a
> changed decision is a **new** entry that supersedes the prior one.
>
> When tempted to re-litigate something, grep here first.
>
> **Format.** Each entry has a context (why this decision matters), the
> decision itself, and consequences (what this entry binds and what it
> doesn't). Citations: backticks for RFC sections (`RFC §7`), bracketed
> superseding IDs (`Supersedes: D-NNN`).

---

## D-001 — Binding properties P1–P4

**Date:** 2026-05-27
**Status:** Settled
**Context:** Without explicit binding properties, design drift is invisible.
A change that weakens a foundational property gets debated case-by-case
forever. The Dockyard project's P1–P4 model is the proven pattern.
**Decision:** pptx-go's binding properties are:

- **P1 — Two layers, one library.** `pptx` (builder) and `scene`
  (renderer) are the only two public layers. `scene` composes `pptx`;
  nothing in `scene` reaches under it.
- **P2 — Tokens, not literals.** Theme tokens are the documented authoring
  path for visual properties; literals are an escape hatch.
- **P3 — OOXML by isolation.** Raw OOXML wire types live only in
  `internal/ooxml`.
- **P4 — No CGo, stdlib-only runtime.** The shipped artifact is pure Go
  with no third-party runtime deps.

Restated verbatim in `CLAUDE.md §1` and `RFC §4`.
**Consequences:** A PR that weakens any of P1–P4 is rejected; an exception
needs an RFC PR and a superseding D-NNN. P1–P4 are not negotiable in
phase plans.

---

## D-002 — Two-layer public API: builder + scene renderer

**Date:** 2026-05-27
**Status:** Settled
**Context:** The library has two consumer shapes: a generic Go author who
wants to build slides without learning anyone's IR, and an IR-driven
consumer (pengui-slides v2) who wants to hand over a typed scene model.
A single combined API forces one consumer to learn the other's vocabulary.
**Decision:** Public API is split into:

- `pptx` — Layer 1. The builder. Theme-aware, token-typed. Generic
  consumers stop here.
- `scene` — Layer 2. A scene IR + a `Render` entrypoint that composes
  `pptx`. IR-driven consumers use this.

`scene` imports `pptx`; `pptx` does not import `scene`. (RFC §3.)
**Consequences:** `scene` is optional. Generic users can `go get
github.com/hurtener/pptx-go` and never see `scene`. The IR catalog
(`scene/nodes.go`) is `scene`'s identity; growing the IR doesn't reshape
the builder.

---

## D-003 — Theme is first-class; tokens are the default authoring path

**Date:** 2026-05-27
**Status:** Settled
**Context:** PowerPoint's own theme model is positional (color1..color12)
and has no semantic intent. pengui-slides has a "design soul" theming
model with semantic roles (canvas/surface/accent/…). The user wants
pengui-slides v2 to shed its renderer and call pptx-go directly — so the
semantic-role contract has to live in pptx-go, not above it.
**Decision:** pptx-go ships a first-class `Theme` whose tokens are
semantic roles (`ColorAccent`, `TextPrimary`, `TypeH1`, `SpaceLG`,
`RadiusMD`, `ElevationRaised`). Builder calls accept tokens. Literal
escape hatches exist (`pptx.RGB(...)`) but are not the documented default.
Theme resolution is lazy at write time, so theme swaps re-render the same
input.
**Consequences:** Generic users using the literal API get a degraded but
working experience. Token consumers get theme-portability for free. Brand
kits are real: load a `.pptx` template, get a `Theme`, render any scene
in the brand.

---

## D-004 — Charts in V1: image-shape; V2: native c:chart

**Date:** 2026-05-27
**Status:** Settled
**Context:** Native OOXML chart XML (`c:chart`) is wide and Excel-coupled.
pengui-slides today renders charts as images. The V1 goal is "match
pengui-slides quickly" — investing a wave in native charts conflicts with
that goal and isn't visually differentiating (native charts win only when
the recipient wants to edit data in PowerPoint).
**Decision:**

- V1 — chart node renders as a native PPTX `pic` shape (image). Caller
  provides bytes via `AssetID`. (`RFC §15.1`.)
- V2 — `internal/ooxml/chart` codec + native chart node disposition. The
  chart node's IR shape is stable across V1/V2; V2 changes disposition.

V1.0.0 ships with image-shape only. Native is a V2 wave with its own
RFC supplement.
**Consequences:** Chart-quality is the caller's responsibility in V1
(matplotlib / ECharts / chartjs / custom). pptx-go provides positioning,
sizing, caption, and disposition warnings — no chart rendering.

---

## D-005 — Curated assets: icons, ornaments, frames embedded via go:embed

**Date:** 2026-05-27
**Status:** Settled
**Context:** pengui-slides v4 leans on a lucide-style icon allowlist
(~36 icons), six preset ornaments, and four device frames. Re-implementing
these in every consumer is wasteful; locking them behind a private API is
inflexible. The right answer is a curated default set with caller-supplied
extensions.
**Decision:** V1 ships:

- `assets/icons/<name>.svg` — a curated lucide-subset (≈ 60 icons at
  V1.0.0). Rendered as native PPTX shape paths via an SVG-to-OOXML
  translator (constraints: single path, solid fill, no gradients).
- `assets/ornaments/<name>.go` — six preset ornament shape recipes
  (`glow_ring`, `radial_glow`, `grid_dots`, `corner_bracket`,
  `chevron_arrow`, `noise_overlay`).
- `assets/frames/<name>.go` — four device frame shape recipes
  (`browser`, `phone`, `desktop`, `laptop`).

All embedded via `go:embed`. Extensible at render time:
`scene.WithIconExtension(name, svg)`, `scene.WithOrnamentExtension(...)`,
`scene.WithFrameExtension(...)`. (`RFC §14`.)
**Consequences:** The set is intentionally closed by default (a name
not in the registry is a Stage-1 validation error). Caller-extension is
first-class. Icons go through the translator's constraints; assets that
exceed them fail at registration time.

---

## D-006 — Module rename to github.com/hurtener/pptx-go

**Date:** 2026-05-27
**Status:** Settled
**Context:** The upstream module is `github.com/Muprprpr/Go-pptx`, an
unmaintained-feeling base that we're forking aggressively. Distribution
under the user's personal account (hurtener) is the V1 plan.
**Decision:** Rename module to `github.com/hurtener/pptx-go`. Wave 0
Phase 00 performs the rename and updates every import. The upstream's
license (MIT) is compatible with the new license (Apache-2.0; see D-007);
the upstream's LICENSE is preserved alongside the new LICENSE per the
MIT terms.
**Consequences:** No backwards-compatibility shims for the old import
path. v0.x signals API-instability while we reshape.

---

## D-007 — License Apache-2.0; starting version v0.1.0

**Date:** 2026-05-27
**Status:** Settled
**Context:** Upstream is MIT. Apache-2.0 gives an explicit patent grant,
which is preferred for libraries that may be brought into larger
ecosystems (a Dockyard app composing pptx-go, for example). The MIT
upstream is compatible: Apache-2.0 is a strict superset for permissive
purposes when paired with attribution.
**Decision:** License: Apache-2.0. Starting version: v0.1.0. Pre-V1
versions may evolve breaking-ly between minor releases; CHANGELOG.md
notes breaking changes.
**Consequences:** Contributors agree (by contribution) to Apache-2.0
terms. Files preserved from upstream keep their MIT headers; new files
get Apache-2.0 headers (or no header — Apache-2.0 doesn't require
per-file headers).

---

## D-008 — Incremental refactor of the upstream, not a greenfield rewrite

**Date:** 2026-05-27
**Status:** Settled
**Context:** The upstream's hard-won correctness — namespace handling,
content-types ordering, relationship ID atomic allocation, streaming save
— is invisible until something breaks. Greenfield re-discovers all of it.
**Decision:** Wave 0 renames the module and reorganizes the upstream
under `internal/opc` and `internal/ooxml` (split). The `pptx/` builder is
rewritten incrementally across Wave 1 with the upstream API preserved as
deprecation aliases where the new API isn't a drop-in. Each phase
preserves the upstream's test coverage on the surface it touches.
**Consequences:** A phase plan that proposes to drop upstream code
without preserving its tested behavior needs a documented rationale and
a passing round-trip test in the same PR.

---

## D-009 — Doc-driven build (Dockyard methodology)

**Date:** 2026-05-27
**Status:** Settled
**Context:** Multi-month builds drift unless the design surface is
managed explicitly. The Dockyard project's full methodology (RFC + phase
plans + decisions log + glossary + drift-audit + preflight + agent skills
+ published docs) is the proven pattern.
**Decision:** pptx-go adopts the Dockyard methodology in full:

- `RFC-001-pptx-go.md` — design source of truth.
- `docs/plans/` — master plan + per-phase plans + `_template.md`.
- `docs/decisions.md` — this file, append-only.
- `docs/glossary.md` — controlled vocabulary.
- `docs/research/INDEX.md` — subsystem→brief reverse index.
- `CLAUDE.md` / `AGENTS.md` — operational rules, mirrored verbatim.
- `Makefile` — canonical build / test / lint / coverage / preflight /
  drift-audit / check-mirror.
- `scripts/` — preflight, drift-audit, smoke per-phase, pre-commit
  hook installer.
- `skills/` — agent skills (lands in Phase 20+).
- `docs/site/` — published docs (lands in Phase 20+).

(`RFC §3.3`, `CLAUDE.md §3` + §§16, 19.)
**Consequences:** Phase 20 establishes `skills/` and `docs/site/`; the §19
"keep user-facing docs in sync" rule activates then. Until Phase 20, the
rule is inert.

---

## D-010 — `scene` is part of pptx-go, not a separate module

**Date:** 2026-05-27
**Status:** Settled
**Context:** A separate module (`github.com/hurtener/pptx-go-scene`) was
considered briefly. It would cleanly express the optional nature of the
scene layer but adds release-orchestration cost and forces a stable
internal interface earlier than warranted.
**Decision:** `scene` is a subpackage of `github.com/hurtener/pptx-go`.
A consumer that uses only `pptx` doesn't compile in `scene`'s code (Go's
import analysis handles this); a consumer that imports `scene` gets both.
**Consequences:** Single-module releases. No cross-module versioning
juggling. The optional nature of `scene` is enforced by import discipline,
not module separation.

---

## D-011 — Per-node rendering policy, not per-deck

**Date:** 2026-05-27
**Status:** Settled
**Context:** A deck-wide rendering mode would be coarse and either
over-fit (everything raster, losing edit-ability) or under-fit
(everything native, losing fidelity on tricky content like `code_block`).
**Decision:** The per-node rendering policy (`RFC §12`) is a per-node
decision. A `code_block` renders as a `pic` shape; a `card` containing
a `code_block` renders the card chrome natively and the code area as a
`pic` inside. Mixed-policy decks are the norm, not the exception.
**Consequences:** The scene renderer composes mixed policies
transparently. Callers don't get a "raster everything" switch in V1;
they don't need one because the per-node default is right.

---

## D-012 — `Theme` resolution is lazy at write time

**Date:** 2026-05-27
**Status:** Settled
**Context:** Eager token resolution (at builder-call time) is simpler to
reason about but defeats theme portability — the same builder state can't
re-render in a different theme. The theme-swap guarantee is the chief
distinguishing feature of P2.
**Decision:** Tokens carry their role through the builder; resolution
happens at write time against the `*Presentation`'s active theme. Callers
needing an early-bound value can `theme.Resolve(token)`.
**Consequences:** `Color` is an interface, not a concrete RGB. Two color
implementations: `tokenColor` (resolves at write time) and `literalColor`
(carries an RGB directly). Both satisfy the interface; APIs that take a
color are not duplicated.

---

## D-013 — Inline code styling: `Run.Code = true`, no separate node

**Date:** 2026-05-27
**Status:** Settled
**Context:** pengui-slides v4 distinguishes `code_block` (block-level
code with whitespace preservation) from inline code (`like.this`
mid-sentence). The block-level case is a leaf node (`code_block`); the
inline case is a run-level styling flag.
**Decision:** `pptx.RunStyle.Code bool` is the inline-code styling toggle:
mono font + subtle background tint. Mapped from pengui-slides v4's
inline `code: true` text-run flag.
**Consequences:** No `inline_code` IR node. Inline code styling is a
property of an existing `TextRun`, matching the upstream IR shape.

---

## D-014 — `code_block` renders as a caller-side raster in V1

**Date:** 2026-05-27
**Status:** Settled
**Context:** Native PPTX text frames at small monospace sizes don't
preserve whitespace fidelity (PowerPoint silently collapses multiple
spaces inside text runs unless wrapped in `<a:t xml:space="preserve">`
*and* the renderer uses tab-stops, which doesn't extend to deep code
indentation). Splitting one line per shape is the alternative; it
fragments selection in PowerPoint and inflates shape counts.
**Decision:** V1 `code_block` renders as a `pic` shape: the caller
renders the code (typically via a syntax-highlighter into a PNG) and
pptx-go embeds it. A caption text shape renders below the image if
`caption` is present.
**Consequences:** Caller is responsible for code-snippet rendering.
pengui-slides v4 already does this; the V1 behavior preserves the
existing flow. V2 may revisit if the OOXML mono-text problem becomes
addressable.

---

## D-015 — `scene.Render` is internally parallel; slides render concurrently

**Date:** 2026-05-27
**Status:** Settled
**Context:** Slide rendering is embarrassingly parallel — each slide's
OOXML is independent of every other slide (relationship IDs are
slide-local; the only shared state is the package's content-types and
media dedup pool). Single-threaded rendering wastes cores on multi-slide
decks.
**Decision:** `scene.Render` spawns a worker pool sized to
`runtime.GOMAXPROCS(0)` by default. Slides are rendered concurrently;
shared state (content-types, media dedup) is protected by the upstream's
existing concurrency primitives. Number of workers is configurable via
`scene.WithWorkers(n)`.
**Consequences:** A V1.0.0 100-slide deck renders meaningfully faster
than the upstream's single-threaded baseline. The `Stats` struct returns
per-slide times so callers can detect imbalance.

---

## D-016 — No `obs/v1`-style observability protocol; slog hook + Stats struct

**Date:** 2026-05-27
**Status:** Settled
**Context:** Dockyard's `obs/v1` event-stream protocol exists because
Dockyard is a runtime (a server you run). pptx-go is a library — there's
no in-process server, no canonical event log. Imposing a streaming
protocol on a library API is hostile.
**Decision:** Observability surface is:

- `pptx.WithLogger(*slog.Logger)` / `scene.WithLogger(*slog.Logger)`
  hooks. Library emits structured events for phase boundaries, slow
  paths, asset failures, layout overflows. No logger = no logs.
- `scene.Render` returns a `Stats` struct: per-slide render time, shape
  counts, asset counts, warnings list.

(`RFC §18`.)
**Consequences:** Callers integrate pptx-go into their own telemetry.
No global state, no pluggable emitters, no SSE — pptx-go does not own
the caller's observability stack.

---

## D-017 — Vendored OOXML specs; single-version codecs in V1

**Date:** 2026-05-27
**Status:** Settled
**Context:** OOXML has multiple ISO editions (1–5) plus Microsoft's
transitional/strict profile distinction. pptx-go targets transitional
(what PowerPoint emits). Spec drift over time is real but slow.
**Decision:**

- Every spec pptx-go implements against is vendored under
  `docs/specifications/<part>-<edition>-<date>.txt` (or .pdf excerpt).
- V1 codecs are single-version. The multi-version codec pattern (codec
  keyed on a discriminator) is introduced only if a real compat case
  forces it.
- A spec re-read is a vendored update + a codec PR + golden re-pin in
  the same change.

(`RFC §20`.)
**Consequences:** A spec bump is one PR, localized to `internal/ooxml`.
Golden tests surface every wire-format change. The multi-version codec
machinery is V2 if and only if needed.

---

## D-018 — Per-node rendering policy is intrinsic to the IR schema; no `Disposition` enum

**Date:** 2026-05-27
**Status:** Settled
**Context:** Every scene IR node has a fixed rendering policy: either it
renders as native PPTX shapes built from typed fields, or it renders as
a `pic` shape with caller-supplied bytes. Codifying this as a runtime
`Disposition` enum (Native / Image) adds an indirection without payoff
— the answer is the same for every instance of a given node type and
is determined by whether the node's IR schema carries an `asset_id`
field.
**Decision:** No `scene.Disposition` enum. The per-node rendering policy
is documented (`RFC §12`) and intrinsic:

- Nodes whose IR schema carries an `asset_id` field render as a `pic`
  shape with caller-supplied bytes (resolved via `AssetResolver`):
  `image`, `chart`, `decoration` of `asset_ref` kind, `code_block`.
- All other nodes render as native PPTX shapes.

No `WithDispositionOverride` option, no deck-level mode toggle.
**Consequences:** `scene/policy.go` is a doc/test file that asserts the
policy per node type — not a runtime decision table. Adding a new node
type that needs rasterization adds an `asset_id` field; the rendering
policy follows from the schema.

---

## D-019 — Font embedding mechanism (no auto-embed default)

**Date:** 2026-05-27
**Status:** Settled
**Context:** PowerPoint renders a font only if the viewer machine has it
installed OR the font is embedded in the PPTX. The upstream
`Muprprpr/Go-pptx` doesn't embed fonts. pengui-slides v4 embeds soul-
referenced fonts via `resolveFontsForEmbedding`. Whether to embed (and
which fonts) is a distribution decision (file size vs portability), not
a library opinion.
**Decision:** pptx-go provides the **mechanism**:

- `pptx.FontSource` interface — caller-injected, resolves
  `(name, style, weight) → bytes`.
- `pres.EmbedFont(name, style, weight) error` — explicit per-font
  embedding using the registered `FontSource`.

There is **no auto-embedding** of theme-referenced fonts. The caller
decides whether and which fonts to embed. A common idiom: iterate the
theme's typography and call `EmbedFont` for each unique
name+style+weight. Subsetting (embed only used glyphs) is V1.x.
(`RFC §7.6`.)
**Consequences:** Callers who want every theme font embedded write one
small loop. Callers who want no embedding write nothing. The library
doesn't decide. go-slides registers a `FontSource` backed by its asset
store and embeds every soul-referenced font; other consumers may not.

---

## D-020 — PowerPoint repair-prompt hygiene (always-on)

**Date:** 2026-05-27
**Status:** Settled
**Context:** PowerPoint shows a "this file has been repaired" prompt on
certain OOXML quirks (empty `lang=""` attributes; namespace order
issues; malformed but technically-valid XML). The prompt is harmless
mechanically but spooks recipients. pengui-slides v4.23 added a
post-processor that scans emitted XML and strips known triggers.
Emitting OOXML that doesn't trigger PowerPoint's repair prompt is
**correctness** (the alternative is "valid OOXML that nonetheless looks
broken"), not preference.
**Decision:** A repair-prompt hygiene pass runs unconditionally in
`internal/render/hygiene.go` as part of every write path. **No
caller-facing option to disable it.** The trigger list is documented in
`docs/design/HYGIENE.md` and grows as new triggers surface.
(`RFC §6` product rules.)
**Consequences:** No `WithRepairPromptHygiene` option. Less API
surface. A future trigger discovered in the wild gets a documented fix
+ a hygiene-list entry in a single PR — never a silent post-processor
change.

---

## D-021 — PPTX sections (slide grouping) are V1

**Date:** 2026-05-27
**Status:** Settled
**Context:** PowerPoint groups slides into named **sections** via the
OOXML `sectionLst` element on the presentation. Sections show up in
the slide-sorter, in the deck outline, and survive edit-save round-
trips. pengui-slides v4's IR doesn't model sections today, but
go-slides will likely want them as a deck-organization primitive (a
strategy deck with "Setup", "Findings", "Recommendations" sections is
a common ask). Adding it to pptx-go costs little; deferring it would
force go-slides to manage the OOXML directly.
**Decision:** V1 ships `*pptx.Presentation.AddSection(name)` returning
a `*Section` with an `Include(slide)` method. Slides are assigned to
sections explicitly. Section ordering follows insertion order. Slides
not assigned to a section live in the implicit default section. The
distinct concept "scene IR `section_divider`" (a slide whose content
is a full-bleed break — RFC §11.1) is **unrelated** to PPTX sections;
the IR node is a slide, and it can or can't be inside a PPTX section.
(`RFC §8.7`.)
**Consequences:** `pres.Sections()` iterates and `pres.Open` parses
existing sections.

---

## D-022 — Speaker notes are V1

**Date:** 2026-05-27
**Status:** Settled
**Context:** pengui-slides v4 carries speaker notes on `Slide.metadata.
speakerNotes` as a first-class authoring field. The OOXML surface (a
`notesSlide` part per slide) is small. Punting them out of V1 would
force go-slides to plumb notes around pptx-go (the slide goes in, the
notes go to a separate channel) — untenable.
**Decision:** V1 ships speaker notes as a property of every `*pptx.Slide`
via `slide.SpeakerNotes() *TextFrame`. The text frame is RichText,
themed (notes inherit the theme's body type), and round-trips through
`pptx.NewFromBytes` / `OpenStream`. The scene IR's `SceneSlide.Notes` field maps directly.
(`RFC §8.8`.)
**Consequences:** Phase 03's scope includes ~1 small file
(`pptx/notes.go`). Round-trip tests cover notes in Phase 03.

---

## D-023 — Slide formats: V1 ships Slides16x9 and Slides4x3

**Date:** 2026-05-27
**Status:** Settled
**Context:** A library that hard-defaults to 16:9 is fine for the modern
case but leaves 4:3 (legacy presentations, education contexts, projectors
that haven't been upgraded in a decade) inaccessible. The cost of adding
4:3 is one constant and the dimensions in the master; the upstream
library supports it but doesn't expose it as a first-class option.
**Decision:** V1 ships two slide formats: `pptx.Slides16x9` (default,
`9144000 × 6858000` EMU) and `pptx.Slides4x3` (`9144000 × 6858000`
adjusted — see master). A future slide-format addition is a constant +
a master template; never per-format branching in user code. Print
formats (`PrintA4Portrait`, `PrintLetterPortrait`) are out of pptx-go's
scope entirely (document concerns; see D-026). (`RFC §5`.)
**Consequences:** `pptx.WithFormat(...)` is a constructor option.
Theme defaults vary slightly by format (font sizes scale).

---

## D-024 — AssetResolver: free-form IDs, with an `asset://`-URI helper

**Date:** 2026-05-27
**Status:** Settled
**Context:** pengui-slides v4 emits `asset://<UUID>` URI markers in
the compiled HTML; an asset resolver replaces them with data URIs at
render time. The scene IR carries `asset_id: string` on the relevant
nodes. pptx-go needs a callback to resolve those ids to bytes. The
design question: does pptx-go impose the `asset://` URI scheme, or is
the resolver free-form?
**Decision:** `AssetID` is `type AssetID string` — free-form. The
caller's `AssetResolver` interprets the string however it likes. A
helper `scene.URIAssetResolver(func(uuid string) ([]byte, string,
error))` accepts `asset://`-prefixed URIs and dispatches to the
caller's UUID-keyed lookup. The scheme is a convention the helper
applies; the core interface stays scheme-agnostic. (`RFC §10.6`.)
**Consequences:** Other callers (not go-slides) aren't forced into
URI semantics. go-slides uses the helper and stays compatible with
its existing UUID asset store.

---

## D-025 — go-slides integration contract: pure data both ways

**Date:** 2026-05-27
**Status:** Settled
**Context:** With pengui-slides as the primary V1 consumer, it's
worth pinning the integration contract explicitly in the RFC. Without
it, "what go-slides passes pptx-go" is implicit and drifts.
**Decision:** The contract is recorded in `RFC §21.7`:

- go-slides passes pptx-go: `*Theme`, `Scene`, `AssetResolver`,
  optionally a `FontSource` + explicit `EmbedFont` calls, optionally a
  `*slog.Logger`.
- go-slides keeps in-house: compiled HTML, Playwright preview pool,
  validators, markdown compiler, comments, recipes, editor state.
- pptx-go returns: a serialized PPTX (file or `io.Writer`) + `Stats`.
- Both directions are pure data; no callbacks back into pptx-go's
  caller beyond the `AssetResolver` and `FontSource` interfaces.

**Consequences:** A change to this contract is a new D-NNN entry, not
a silent API change. The contract is what go-slides will pin its
PPTX-export tool against.

---

## D-026 — pptx-go is the engine; product behavior lives in callers

**Date:** 2026-05-28
**Status:** Settled
**Context:** A library that grows render-mode toggles, legibility
heuristics, validation pipelines, or other product behavior becomes a
small product itself. The intended consumer (go-slides) already has
the natural home for those decisions; pushing them into pptx-go means
every other consumer inherits go-slides' opinions whether they wanted
them or not. The cleanest engine-or-product line is drawn here.
**Decision:** pptx-go is the **engine** for converting a typed scene IR
into PPTX. It does not decide:

- *what* should be in the deck — that's the caller's IR.
- *how the deck should look* — that's the caller's Theme.
- *what content fidelity to aim for* — that's the caller's rendering
  choices (per-node `asset_id` for caller-rasterized content).
- *which fonts to embed* — that's the caller's distribution choice
  (D-019).
- *what render policies apply per use case* — that's the caller's
  product layer (image vs editable hybrid, legibility boosts,
  validation pipelines).

The library decides only:

- *how to faithfully encode the IR as valid OOXML* (P3, internal
  correctness incl. repair-prompt hygiene — D-020).
- *which mechanisms to expose* (theme tokens, asset resolver, font
  source, section grouping, speaker notes, slide formats, etc.).

Product behavior belongs in go-slides (or any other caller).
**Consequences:** A smaller API surface. A clearer mental model: "what
would an engine need to do?" answers most "should we add this?"
questions. Document-mode concepts (TOCs, bibliographies, page breaks,
print formats) are doc-mode product concerns and don't enter pptx-go
at any version.

---

## D-027 — Coverage-gate strictness ramps from Phase 01

**Date:** 2026-05-28
**Status:** Settled
**Context:** `CLAUDE.md §11` mandates that a new package with no
configured coverage band fails the build (`require_configured`). At
Phase 00 the tree is still the pre-reorg upstream: many packages
(`opc`, `parts`, `pptx`, `utils`) carry no band and would fail wholesale
the moment the gate turns strict. The reorg that gives each package its
permanent home and band is Phase 01.
**Decision:** `internal/coveragecheck/coverage.json` ships in Phase 00
with `require_configured: false` and a single banded package
(`internal/coveragecheck` at 70%). The gate enforces *configured* bands
immediately (a configured package below its band fails) but does not yet
fail on un-banded packages. Phase 01 sets `require_configured: true` and
bands every surviving package as part of the `internal/` reorg.
**Consequences:** The coverage gate is live and meaningful from Phase 00
(it gates the tooling it ships), without blocking on the pre-reorg
tree's untracked coverage. The strict-mode flip is an explicit Phase 01
acceptance item rather than a silent default.

---

## D-028 — drawingML types stay in `internal/ooxml/slide` for Phase 01

**Date:** 2026-05-28
**Status:** Settled
**Context:** RFC §6.2 lists `internal/ooxml/drawing` as its own
subpackage. In the inherited `parts/` package, every drawingML type
(`XSp`, `XShapeProperties`, `XFillProperties`, `XBlip`, `XTransform2D`,
`XTextBody`, `XTable*`, …) and the `XMLWriter`/`XMLWriterPool`
serialization base their `WriteXML` methods depend on are referenced
**only** by the slide family (`slide.go` / `slide_types.go`). Extracting
`drawing/` during Phase 01 would force `XMLWriter` into a shared `common/`
package and split `slide_types.go` before any cross-family consumer
exists — avoidable surgery during a "move, don't rewrite" phase.
**Decision:** Phase 01 keeps the drawingML types and `XMLWriter` inside
`internal/ooxml/slide` and ships `internal/ooxml/drawing` as a documented
placeholder (the same treatment RFC §6.2 gives `chart/`). The types
migrate to `drawing/` — with `XMLWriter` moving to a shared helper — when
the builder (Phase 03+) or the SVG→OOXML translator (Phase 12) first needs
them outside the slide family.
**Consequences:** Phase 01 stays a relocation with no cross-family
coupling introduced, honoring §6.2's independence rule. The `drawing/`
extraction is deferred to the phase that first has a real consumer, where
it can be done with that consumer's requirements in view. RFC §6.2's
literal layout is reached incrementally, not in one move.

---

## D-029 — Coverage-gate strict flip + test co-location deferred past Phase 01

**Date:** 2026-05-28
**Status:** Settled (refines D-027)
**Context:** D-027 committed to flipping `require_configured: true` and
banding every relocated package in Phase 01. In practice the upstream
test suite lives in external packages under `test/` (`parts_test`,
`pptx_test`, …) and is heavily fixture-dependent (the `*FromFile` tests
skip without `test/test-data`, which was never committed). With the
standard `make coverage` (no `-coverpkg`), per-package **self**-coverage
of `internal/opc` and `internal/ooxml/*` is therefore 0% — their tests
are measured against the external `test/` packages. Switching to
`-coverpkg=./...` would attribute cross-package coverage but emits
duplicate coverage blocks across test binaries that the
`internal/coveragecheck` parser sums rather than de-duplicates,
producing wrong numbers. Banding the relocated packages at 0% would be
noise, not signal.
**Decision:** Phase 01 keeps `require_configured: false` and bands only
`internal/coveragecheck`. The relocated upstream tests are **preserved
and passing** but stay under `test/` for this phase; co-locating them
into the new package directories (so self-coverage is attributed) and
flipping `require_configured: true` with meaningful bands is deferred to
the phase that next hardens each package's tests and resolves the
fixture story (Phase 02+ as the builder is built on these packages).
`internal/coveragecheck` will gain block de-duplication when/if
`-coverpkg` is adopted.
**Consequences:** The coverage gate stays correct and green (no
double-counting, no fabricated 0% bands). Phase 01 stays a structural
move. The strict flip and test co-location become an explicit acceptance
item of the phase that earns it with real tests, rather than a
box-checking exercise on inherited code. Supersedes the Phase-01 timing
in D-027; the strict-mode intent stands.

---

## D-030 — Color interface + token builder constructors land in Phase 03

**Date:** 2026-05-28
**Status:** Settled (sequences D-012)
**Context:** RFC §7.2 / D-012 make `pptx.Color` an interface with a
write-time-lazy `tokenColor` and a `literalColor`, surfaced via
`pptx.TokenColor(role)` and `pptx.RGB(...)`. The inherited `pptx` package
already defines a concrete `Color` struct used by the upstream shape
builder. Turning `Color` into the interface is part of migrating the
builder to take tokens — the Phase 03 builder spine, which explicitly
"migrates the upstream pptx incrementally; new files supersede old ones;
old API kept as deprecated aliases." Doing it in Phase 02 would pull that
migration forward without the builder context that fixes its exact shape.
**Decision:** Phase 02 ships the `Theme` model and a **deterministic
resolver** returning concrete OOXML values (`ResolveColor → RGB`,
`ResolveSpace → EMU`, `ResolveType → FontSpec`, …). The lazy `Color`
interface and the `TokenColor`/`RGB` constructors land in Phase 03 with
the builder API that consumes them. D-012's lazy-resolution intent is
preserved — only its surfacing point moves.
**Consequences:** The theme-swap guarantee is proven at the resolver
level in Phase 02 (one token, two themes → two values) and end-to-end
through the builder in Phase 03. Phase 02 introduces no `type Color`, so
it does not disturb the upstream struct ahead of the builder rewrite.

---

## D-031 — PPTX validity is verified in four layers; harness lands before Phase 03

**Date:** 2026-05-28
**Status:** Settled
**Context:** Round-trip tests (write → our own Open → assert) prove we read
back what we wrote, but not that the output is *valid* — a malformed writer
and a matching reader pass round-trip while PowerPoint rejects the file.
`CLAUDE.md §11` already mandates: spec compliance against vendored specs
(not live PowerPoint), and PowerPoint compatibility tested manually on
reference decks, one per wave. This decision operationalizes that.
**Decision:** Validity is checked in four layers, cheapest/most-deterministic
first:
1. **OPC integrity** — `internal/conformance`, pure-Go, gates every emitted
   deck in tests: content-type coverage, relationship-target resolution,
   dangling `rId` references, pack-URI validity, required-parts.
2. **Schema conformance** — vendored ISO/IEC 29500 *transitional* XSDs in
   `docs/specifications/`, validated via `xmllint --schema` in CI. Known
   PowerPoint-isms get annotated exceptions, not a chase for 100%.
3. **Office-app open proxy** — a CI job runs LibreOffice headless
   (`soffice --headless --convert-to`) over reference decks; a failed
   convert = invalid. The closest automatable proxy to "a real app opens
   it without the repaired prompt."
4. **Manual PowerPoint check** — one reference deck per wave opened in real
   PowerPoint (the maintainer's Mac); result recorded in `docs/validation/`.
   The ground truth the first three layers approximate.
All automated tooling (xmllint, LibreOffice, python-pptx) is **test/CI-only**
and never enters the shipped artifact (P4 holds). The harness is built
**before** the Phase 03 builder spine so the new builder is developed
against a working validator; Phase 03's acceptance turns on the full-deck
conformance gate (it is the first phase to emit a complete deck + the D-020
hygiene pass).
**Consequences:** A malformed-output regression fails CI at layer 1–3 long
before a human opens PowerPoint. The validator applied to *current* output
establishes a baseline of known gaps (e.g. relationship attributes emitted
as `rid=` rather than `r:id=`) that Phase 03 must close. Vendoring the full
ISO schemas requires obtaining the schema bundle; until present, the xmllint
layer SKIPs rather than failing.

---

## D-032 — One OOXML emission path: xml.Marshal + a RestoreNamespaces write pass

**Date:** 2026-05-28
**Status:** Settled
**Context:** Investigating the builder spine (Phase 03) revealed the inherited
emission is broken in several ways with one architectural root cause: Go's
`encoding/xml` cannot cleanly emit namespace-prefixed names, so the upstream
took two divergent, separately-broken approaches. `presentation.xml`/theme/
core serialize via `xml.Marshal` and come out **without any namespaces**
(`<presentation>` not `<p:presentation xmlns:p=…>`); slides use a
**hand-rolled `XMLWriter`** that writes element attributes as **text**
(`<p:cNvPr>1 name="Layout"</p:cNvPr>` instead of `<p:cNvPr id="1"
name="Layout"/>`). The read path already normalizes the inverse with
`StripNamespacePrefixes`.
**Decision:** Unify all OOXML emission on a single path: serialize every part
with stdlib `xml.Marshal` using **bare** element/attribute names (which
serializes attributes correctly), then run one shared **`RestoreNamespaces`**
write pass that adds the canonical `p:`/`a:`/`r:` prefixes and `xmlns`
declarations per part — the exact inverse of `StripNamespacePrefixes`. The
hand-rolled slide `XMLWriter` is **deleted**. Read continues to strip;
write restores. Both directions live in `internal/ooxml` as two symmetric,
golden-tested functions.
**Consequences:** One correct emission path replaces two broken ones; the
`<p:cNvPr>`-attributes-as-text bug and the missing-presentation-namespaces
bug are both fixed at the root, as is `rid`→`r:id`. The `RestoreNamespaces`
pass needs a per-element prefix map (bounded; OOXML's prefix conventions are
fixed) and is verified by goldens plus the D-031 conformance/schema/
LibreOffice gates. Relationship wiring (presentation→slide rIds) and
seeding a complete master/layout/theme are separate builder fixes; the EMU
`Box` API supersedes the upstream pixels-via-`PxToEMU` coordinate handling.

---

## D-033 — Color is a sealed interface; the RGB type is the literal

**Date:** 2026-05-29
**Status:** Settled
**Context:** D-012/D-030 deferred turning `Color` into an interface (token vs
literal) to Phase 03. Phase 02 had already shipped the token model — an `RGB`
string type, `ColorRole`/`TextColorRole`, and `Theme.ResolveColor`. The
inherited `pptx.Color` *struct* (plus `ColorMap`, `ParseColor`, named presets)
was a separate, parallel color system wired to nothing in the builder or theme.
A naïve `pptx.RGB(...)` literal constructor would have collided with the
existing `RGB` type.
**Decision:** `Color` is a **sealed interface** (`resolve(*Theme) resolvedColor`,
unexported). The existing `RGB` string type **implements** it, so
`pptx.RGB("2563EB")` is both a value and a literal fill color — no naming
collision, no second constructor. `pptx.RGBA` adds alpha; `pptx.TokenColor`
(surface) and `pptx.TokenTextColor` (text) are tokens that resolve against the
**active theme at apply time**, which is the theme-swap mechanism (P2). The
inherited concrete `Color` struct, `ColorMap`, `ParseColor`, presets, and the
`Slide.ValidateColor`/`ResolveColor(string)` helpers are **retired**. `Fill`
(`SolidFill`/`NoFill`) and `Line` are likewise sealed and theme-resolving.
**Consequences:** One color model, owned by the theme. Tokens are honoured at
write time, not baked, so a theme swap re-colors the same builder input. Sealed
interfaces keep callers from supplying a color/fill the codec can't emit.
Gradient/pattern/picture fills are not yet implemented (picture fills arrive
with media). The token→`theme1.xml` *emission* (replacing the static scaffold
theme, D-032/A2) remains a follow-up; resolution-to-`srgbClr` covers V1.

---

## D-034 — Section list is an injected p14 fragment; the slide owns its rels

**Date:** 2026-05-29
**Status:** Settled
**Context:** Phase 03 Chunk C adds media, sections, speaker notes, and
streaming on top of the D-032 emission path (`xml.Marshal` bare names +
`RestoreNamespaces`). Two structural problems surfaced. (1) PowerPoint stores
slide sections under a `p14:sectionLst` whose `<p14:sldId>` shares the local
name `sldId` with the top-level `<p:sldIdLst><p:sldId>`. `RestoreNamespaces`
keys on a single element→prefix table (one local name → one prefix), so it
cannot emit `p:sldId` and `p14:sldId` from the same name — and it declares
namespaces only on the root, with no `p14`. (2) A2 left the slide's
relationships split: `slide.SlidePart` carried image/media rels in its own
`opc.Relationships`, while the package's `opc.Part` for the slide carried the
layout rel separately — the two used independent `rId` namespaces (both
starting at `rId1`) and the builder's image rels were never emitted.
**Decision:**

- **Sections** are emitted as a **literal `p14` XML fragment** injected as the
  last child of `<p:presentation>` after `RestoreNamespaces` runs
  (`injectSectionLst`), carrying its own `xmlns:p14` on `<p14:sectionLst>`.
  The codec marshals no section structs (the `XExtLst`/`XSection` types exist
  for the **read** path only; `StripNamespacePrefixes` makes them plain on
  parse). Section GUIDs are derived deterministically from a counter so decks
  stay byte-stable. Unassigned slides fall into a leading implicit "Default
  Section" so the list spans every slide (PowerPoint requires it).
- **Slide relationships** live canonically on `slide.SlidePart`'s relationship
  set (a single `rId` namespace: layout `rId1`, then images/notes); the slide
  layout rel moves there, and `syncSlides` **mirrors** that set onto the
  package `opc.Part` (preserving `rId`s) so they are emitted. Media bytes are
  written as package parts by `syncMedia`.
- **Speaker notes** ship in V1 as a plain-text setter, `Slide.SpeakerNotes(text
  string)`, not the `*TextFrame` accessor RFC §8.8 sketches — `TextFrame` is
  the rich-text model (a later phase). The setter emits a `notesSlide` part
  with a hand-authored `notesMaster1.xml` (the A2 scaffold pattern).
- **Streaming** follows the RFC §9 path-based signatures `OpenStream(path)` /
  `SaveStream(path)`. CLAUDE.md §5's context-first convention yields to the
  explicit RFC signature here (RFC > CLAUDE.md, §2); a context-aware streaming
  API would be an RFC change plus a superseding decision.

**Consequences:** Sections round-trip (write injects, read parses into the
presentation part). The relationship seam is closed: builder-added images and
notes are emitted and resolve under the conformance gate. A pre-existing
`internal/opc` streaming bug — package `.rels` dropped on open because
`StreamPackage.loadRelationships` tested `IsPackageRels` on the source URI
rather than the rels URI — is fixed so `OpenStream`→`SaveStream` stays valid.
Notes-as-text and the streaming-signature choice are recorded as Phase 03
plan deviations.

---

## D-035 — Byte-identical saves: fixed ZIP epoch + stable map serialization

**Date:** 2026-05-30
**Status:** Settled
**Context:** RFC §10.1 makes byte-identical render output a hard requirement
(pengui-slides snapshot-tests on the bytes). The Wave 2 checkpoint, while
landing D-015's parallel renderer, found that requirement was already violated
on `main` — independent of concurrency. Three `internal/opc` save paths
stamped every ZIP entry with `time.Now()`; `ContentTypes.ToXML` ranged its
`defaults`/`overrides` maps; and `MediaManager.AllGlobalMedia` ranged a
`sync.Map`, so `syncMedia` added media parts (and thus ZIP entries) in random
order. Saving the same presentation twice produced different bytes.
**Decision:** Saves are deterministic.

- Every ZIP entry is stamped with a **fixed timestamp**, the 1980-01-01
  MS-DOS/ZIP epoch (`opc.fixedZipModTime`), not the wall clock. 1980 is the
  earliest the MS-DOS format represents, so it stays a *valid* time — keeping
  the Windows Explorer workaround the previous `time.Now()` stamp intended.
- `ContentTypes.ToXML` emits defaults and overrides **sorted** (by extension /
  part name).
- `MediaManager.AllGlobalMedia` returns resources in **media-file-number
  order** (`image1`, `image2`, …), so part materialization is stable.

**Consequences:** `scene.Render` (and any builder save) is byte-identical
across runs, satisfying RFC §10.1 and making D-015's parallel renderer safely
testable for idempotency. The OPC layer owns determinism, so it holds for every
caller, not just the scene renderer. The fixed media-numbering relies on global
media being created in deterministic order — see D-036.

---

## D-036 — V1 degrades every asset-resolution failure to a warning

**Date:** 2026-05-30
**Status:** Settled
**Context:** RFC §10.6 distinguishes optional asset failures (surface a
`LayoutWarning`) from *required* assets (the render fails). Phase 06 shipped
all asset failures as warnings and documented the deviation, but it was never
reconciled with the RFC. The V1 IR carries no "required" designation on any
asset-bearing node (`code_block` is the only shipped one), so the required-
failure branch is unreachable today; pptx-go is an engine, and "this asset is
mandatory" is a caller policy (D-026).
**Decision:** In V1, every unresolved asset degrades to a `LayoutWarning` and
the node is skipped; there is no render-fatal asset path and no `strict` render
mode. A future "required asset" mechanism (an IR field plus the fatal branch)
is the trigger to revisit RFC §10.6 — at which point it lands with a superseding
decision. The byte-identical guarantee (D-035) further requires that any slide
which *may* register global media composes in deterministic scene order; the
renderer enforces this by scheduling asset-bearing slides sequentially while
media-free slides fan out across the D-015 worker pool.
**Consequences:** RFC §10.6's required-asset sentence is documented as
not-yet-exercised rather than silently contradicted. Callers that need a missing
asset to be fatal inspect `Stats.Warnings` themselves (RFC §10.2). When the
image node lands (Phase 11) it inherits this rule and the sequential-scheduling
constraint.

---

## D-037 — Template ingestion clones the template package and strips slides

**Date:** 2026-05-31
**Status:** Settled
**Context:** Phase 09 seeds a presentation from a brand kit (RFC §13.1).
Brief 01 (F3) noted that a template's identity is a relationship chain
(slide→layout→master→theme) plus placeholder geometry and backgrounds the
semantic `Theme` does not capture, and recommended copying the template's parts
wholesale rather than reconstructing them. Hand-grafting those parts into a
freshly scaffolded package risks the PowerPoint "repair" class of bug (orphaned
or double-wired relationships — the PR #13 lesson).
**Decision:** `pptx.FromTemplate(brand *Presentation)` is a `New` option that
adopts the brand kit by **cloning its OPC package and stripping any slides**,
rather than grafting parts into the scaffold. Cloning preserves the template's
already-valid relationship graph, theme, masters, layouts, and auxiliary parts
verbatim; `clearTemplateSlides` then removes slide parts + their
presentation→slide relationships + `sldIdLst` entries so the new deck starts
empty. The brand's theme (extracted on its open) and its read-only master/layout
registry are adopted. Adoption falls back to the default scaffold on any failure,
so `New` never yields a broken deck.

Two supporting changes land with it:
- **Opening a deck extracts its theme + masters.** `loadPresentationPart` now
  sets the presentation's theme from `theme1.xml` and builds a `Master`/`Layout`
  registry, so an opened deck can act as a brand kit (`brand.Theme()`,
  `brand.Masters()`). Both are best-effort: a missing theme keeps `DefaultTheme`,
  an unparseable master contributes nothing (brief 01 F6 — permissive reader).
- **`FromTemplate` takes a `*Presentation`, per RFC §13.1**, not the
  `TemplateSource` the phase plan drafted. The caller opens the kit
  (`OpenStream`/`NewFromFile`) — which can return an error — then `New` adopts
  the in-memory value, so `New` needs no error return. (Plan deviation, recorded
  in the Phase 09 plan §16.)

**Consequences:** Ingestion is robust by construction — no manual rewiring — and
deterministic (the clone + our fixed-epoch save, D-035, keep `FromTemplate`
output byte-identical). A slide added to a seeded deck relates to the template's
named layout (the `slideLayout1.xml` default still exists in the clone). Multi-
master rel-precise layout grouping is approximated (unclaimed layouts attach to
the first master); rich per-placeholder targeting from the IR is deferred.

---

## D-038 — Frame reference: FrameKind enum alias + named registry

**Date:** 2026-06-01
**Status:** Settled
**Context:** Phase 10 ships the four curated device frames (RFC §14.3) and the
§14.4 extension seam `scene.WithFrameExtension(name, recipe)`, which references
frames **by string name**. The `Image` IR node, however, already shipped (Phase
05) with a closed `Frame FrameKind` enum (`FrameNone` + `FrameBrowser`/
`FramePhone`/`FrameDesktop`/`FrameLaptop`). An enum cannot name a
caller-registered frame, so the two reference mechanisms must be reconciled
without breaking the shipped enum. Brief 02 (F5) surveyed the IR's own
precedent: `Decoration` already carries both a `DecorationKind` enum **and** a
free-form `Preset string` curated-name — the identical shape of problem.
**Decision:** The frame **registry is keyed by name**. The four curated
`FrameKind` values map to the four reserved curated names (`browser`, `phone`,
`desktop`, `laptop`). `Image` gains an additive optional field
`FrameName string`:

- `FrameName != ""` → selects that name (curated **or** caller-registered) and
  **takes precedence** over the enum.
- `FrameName == ""` → the `FrameKind` enum selects: `FrameNone` ⇒ no frame,
  otherwise the curated name for that kind.

The enum stays as the zero-import ergonomic path for the curated four; the
string is the §14.4 extension seam. An `Image` whose **resolved** frame name is
absent from the render's registry (curated ∪ `WithFrameExtension` set) fails
**Stage-1 validation** (closed-name semantics, §14.4) — checked in `Render`
after the option-free `ValidateScene`, because the registry is derived from
render options. A curated `FrameKind` always resolves. Extensions are
**per-render** (folded over a copy of the curated registry, read-only during
the parallel compose) — not process-global state — preserving D-035 byte-
identical determinism and concurrency safety.

**Consequences:** No break to the shipped `Image.Frame` surface (`FrameName` is
additive; its zero value preserves prior behavior). The enum-plus-name pattern
is now consistent across `Image` (frames) and `Decoration` (ornaments), and is
the template Phases 12 (icons) and 13 (ornaments) follow for their own curated
sets. The registry being name-keyed lets a caller override a curated frame for
one render by registering its name. A true OOXML group-shape for the bezel is
**not** part of this decision (the builder has no group primitive in V1; a
framed image is a cluster of sibling native shapes — deferred post-V1).

---

## D-039 — Phase 11: media work already delivered; scene Image gains crop/fit; no media-manager relocation

**Date:** 2026-06-01
**Status:** Settled
**Context:** The master plan scopes Phase 11 as "Image node + media manager
refactor", naming a `pptx/media.go` refactor (dedup pool moved to `internal/opc`
or a new `internal/media`, alt-text first class, MIME detection) plus scene-side
"full image node composition (asset resolution, alt text, crop, fit, frame)".
Inspection at the start of Phase 11 found the builder half **already delivered**:
the foundation builder phase shipped `pptx/media_manager.go` (content-hash MD5
dedup, global media pool, deterministic ordering), `SetAltText`, `SetCrop`,
`SetFit` (`FitFill`/`FitNone`), and `sniffImage` MIME detection — all tested
(`test/parts/media_manager_dedup_test.go`, `test/pptx/media_test.go`). Phase 10
wired scene asset resolution, alt text, and frame composition. The remaining gap
was scene-side only: the `Image` IR node could not express **crop** or **fit**.
The master plan thus drifted from reality (it assumed an upstream media-manager
refactor still pending). Resolving the drift requires a settled call on three
points.
**Decision:**
1. **Phase 11 adds no builder media code.** The media manager, dedup pool, alt
   text, crop, fit, and MIME detection are delivered and tested; re-doing them
   would be redundant. Phase 11's acceptance criteria that name those
   capabilities (dedup writes one part; alt text round-trips) are satisfied by
   the existing builder and verified by new **scene-seam** tests.
2. **The `internal/media` relocation is declined.** The dedup pool's wire type
   (`MediaResource`) already lives in `internal/ooxml/media` (the P3-isolated
   seam); the *orchestrator* (`MediaManager`) coordinates `Slide`/`Presentation`
   state and is correctly placed in package `pptx`. The master plan offered the
   move as optional ("`internal/opc` **or** a new `internal/media`"); relocating
   working, tested code for nominal purity is churn with no functional gain
   (`CLAUDE.md §4.3` — a reasonable deviation, documented here).
3. **The scene `Image` IR gains `Crop` and `Fit`** as the genuine Phase 11
   deliverable — mechanism exposure of the builder's existing crop/fit (engine,
   not product — D-026). `Crop`/`Fit` are re-exported builder types (type
   aliases, like the design tokens in `scene/tokens.go`); both fields are
   additive and their zero values (`Crop{}`, `FitFill`) reproduce Phase-10
   behavior byte-for-byte. `Fit` is limited to `FitFill`/`FitNone`: aspect-aware
   cover/contain would require reading pixel dimensions, forbidden by §7 (the
   RFC §8.6 example's `FitCover` is therefore **not** in V1). An out-of-range or
   over-crop fails Stage-1 validation rather than being silently clamped.

**Consequences:** Phase 11 is a focused scene-IR phase (two `Image` fields +
their wiring + crop-range validation + consolidation tests), not a builder
refactor. The media manager stays in `pptx`. A future need for aspect-aware
fitting (cover/contain) is a separate decision gated on a pixel-dimension source
the caller supplies (not a pptx-go read), preserving §7. The drift is recorded
so the master-plan Phase 11 block is understood as superseded by this entry.

---

## D-040 — Phase 12: icon engine + ~16 starter set (≈60 deferred); no arcs; AddIcon takes SVG bytes

**Date:** 2026-06-01
**Status:** Settled
**Context:** D-005 commits V1 to a curated lucide-subset of ≈60 icons rendered as
native PPTX shape paths via an SVG→OOXML translator (single path, solid fill, no
gradients). Phase 12 must build the engine that capability needs — none of it
existed: custom path geometry (`a:custGeom`) wire types, the SVG translator, a
builder API to place a path glyph, and the icon registry. Two implementation
questions D-005 left open needed settling: how many icons ship in this phase
(hand-authoring ≈60 quality single-path glyphs is a large, error-prone content
task — and lucide's real icons are stroke-based multi-element, so they cannot be
copied; the curated set is lucide-*style*, authored as filled single paths), and
how far the SVG subset extends.
**Decision:**
1. **Ship the full icon engine + a ~16-icon starter set; defer the ≈60 target.**
   The engine (wire types + translator + `AddIcon` + registry + extension seam +
   registration-time validation) lands complete and at quality. The curated set
   is a hand-authored starter set (~18 single-path filled glyphs: arrows,
   chevrons, check, x, plus, minus, square, circle, dot, triangle, diamond,
   star). Growing toward D-005's ≈60 is a **content follow-up** — each addition
   is one validated `.svg` file, no code change; `icons.Names()` reports exactly
   what ships (no silent truncation). (Confirmed with the maintainer.)
2. **The translator subset excludes elliptical arcs (`A`/`a`).** SVG→`a:arcTo`
   conversion is lossy and wide; curved glyphs are authored with cubic/quadratic
   Béziers (a circle is four cubics). Supported: `M L H V C S Q T Z` (absolute +
   relative); `S`/`T` expand to `C`/`Q` by reflecting the previous control point.
   An arc — or any element/fill outside the subset — fails translation, i.e. at
   **registration**, never silently at render.
3. **The builder API is `Slide.AddIcon(svg []byte, box, opts…) (*Shape, error)`
   plus `pptx.ValidateIcon(svg) error`** — SVG bytes in, an opaque `*Shape` out;
   the `custGeom` OOXML wire types stay in `internal/ooxml` (P3). `scene` never
   reaches under `pptx`: `scene.WithIconExtension` / `scene.ValidateIcon`
   delegate to `pptx.ValidateIcon` (P1). The default fill is the accent token
   (P2); a caller `WithFill` overrides the color.

**Consequences:** pptx-go gains custom path geometry — reusable beyond icons
(future vector shapes). The icon registry mirrors the frames seam (D-038) with
one difference: an icon extension is **validated at registration** (its SVG is
translated up front), not merely name-checked at render, per D-005. The starter
set is usable immediately; the ≈60 target is tracked content work, not an engine
gap. Icon *placement* by IR nodes (`card`, `flow`, `header_pill`) arrives with
those nodes (Phases 14–15) — Phase 12 ships the engine + registry they consume.

One seam note: `internal/render` now imports `encoding/xml` to parse the SVG
*input* (an XML dialect). This does not weaken P3 — `internal/render` defines and
exposes **no** OOXML wire types (those stay in `internal/ooxml`, which it imports
and produces); nothing above the internal wall (`pptx`, `scene`) touches
`encoding/xml`. The `drift-audit.sh` P3 allowlist is extended from
`{ooxml, opc, conformance}` to add `render`, with this rationale in the script.

---

## D-041 — V1 ships gradient fills (linear + radial); rotation + token-alpha land with them

**Date:** 2026-06-02
**Status:** Settled
**Context:** Phase 13's ornaments include `radial_glow` and `glow_ring`, which
RFC §14.2 describes as gradient/glow effects. The builder shipped only `SolidFill`
+ `NoFill` (alpha on literal colors only) and **no gradient fill** — `pptx/fill.go`
even notes "Gradient, pattern and picture fills are tracked separately." Rendering
a glow as alpha-layered concentric solids bands visibly. The Phase-02 builder
block listed `GradientFill` as in-scope but it was never built (a drift).
**Decision:** Build **gradient fills in V1**. Add `XGradientFill` (a `gsLst` of
stops plus either `lin` for linear or `path="circle"` + `fillToRect` for radial)
to `internal/ooxml/slide`, a `GradientFill` field on `XShapeProperties`, and a
public `pptx.LinearGradient(angle, stops…)` / `pptx.RadialGradient(stops…)` Fill.
A radial glow is a 2-stop `path="circle"` gradient: accent (opaque, centre) →
accent at `alpha=0` (edge). Land two adjacent builder primitives in the same
change: `pptx.WithRotation(deg)` (the `XTransform2D.Rotation` wire field already
exists; `chevron_arrow` and rotated asset decorations need it) and
`pptx.TokenColorAlpha(role, alpha)` (alpha on a *token* color, so a recipe can
honor a decoration's opacity while staying token-based — P2). **Group-shape unit
rotation is NOT built** (the builder has no group transform — V2); a multi-shape
ornament rotates per-shape, which suits the rotationally-symmetric glows/grid and
is the documented V1 behavior.
**Consequences:** The builder gains a general gradient primitive (reusable beyond
ornaments) and shape rotation. `pattFill` (pattern/hatch) stays unbuilt (no V1
ornament needs it). Glows render as true radial gradients. Gradients are
deterministic (fixed stops, integer angles) so D-035 holds.

## D-042 — Phase 13 absorbs two carried-forward builder fixes and splits delivery

**Date:** 2026-06-02
**Status:** Settled
**Context:** The Phase-12 wiring audit surfaced two builder-layer gaps left
unfixed at the time (they were feature-sized, not broken wires): `Scene.Meta`
(Title/Author/Subject) was silently dropped because `core.xml` is a static empty
part with no setter, and `pptx.WithLogger` is promised by RFC §18 / `CLAUDE.md`
§8 but never existed (only the scene logger, fixed in the audit). The maintainer
asked to fold both into Phase 13. Phase 13 is already large (gradient primitives
+ six ornaments + Decoration IR + z-order).
**Decision:** Land the two carried fixes **with the builder foundations in
PR #1**, separate from the ornaments/Decoration work in **PR #2** (one Phase-13
plan covers both). PR #1: gradient fills, `WithRotation`, token-alpha,
`Presentation.SetMetadata` (regenerates `docProps/core.xml` deterministically —
XML-escaped caller strings, **no created/modified timestamps**, preserving
byte-identical output), `pptx.WithLogger` (a builder logger emitting a save-time
event, RFC §18), and `scene.Render` writing `Scene.Meta` through `SetMetadata`.
PR #2 builds on PR #1's primitives. Rationale: the fixes and primitives are
builder-layer and orthogonal to ornaments; shipping them first keeps each review
focused and clears the audit debt before the ornament content lands.
**Consequences:** Deck metadata round-trips (the Phase-11/12 decks' `Scene.Meta`
titles now reach `core.xml`). `pptx.WithLogger` closes the doc-vs-code drift and
gives builder/scene observability parity. The split means two PRs for one phase;
the plan's acceptance criteria are grouped per PR.

## D-043 — Phase 14 builds the `outerShdw` elevation primitive; splits delivery; Card IR grows additively

**Date:** 2026-06-02
**Status:** Settled
**Context:** Phase 14 (`card`/`card_section`) renders elevation as a real card
drop shadow, but the builder has **no shadow primitive** — every shape emits an
empty `<a:effectLst/>` and the theme-resolved `Elevation` (blur/offset/color/
alpha) is dropped. This is the same situation gradients were for Phase 13
ornaments (D-041): a real visual property the theme already tokenizes but the
builder cannot emit. The shipped `Card` also lacks the v4 knobs RFC §16 maps
1:1 (`Eyebrow`, `Icon`, `HeaderPill`, `BorderStyle`, `Size`, `Layout`), and the
icon registry built in Phase 12 (D-040) was never *consumed* — `cfg.icons` was
deliberately not stored to avoid a write-only field. The shipped `Card` already
carries `Outline bool`; the v4 knob is a richer `border_style`.
**Decision:** **Build** the `<a:outerShdw>` shadow primitive in V1 (not
approximate, not defer) — the D-041 precedent. Add `XEffectList`/`XOuterShadow`
to `internal/ooxml/slide`, an `EffectList` field on `XShapeProperties` (after
`Line`, per `CT_ShapeProperties` order), `effectLst`/`outerShdw` →`a:` entries
in `restorenamespaces`, and public `pptx.WithElevation(role)` (token path, P2)
+ `pptx.WithShadow(Elevation)` (literal escape hatch) `ShapeOption`s. The theme
`Elevation`'s cartesian `OffsetX/OffsetY` convert to `outerShdw`'s polar
`dist`/`dir` with integer rounding (D-035 holds); a flat elevation emits **no**
`effectLst` (byte-identical to today). **Deliver as a split** (the D-042
pattern): PR#1 = the builder shadow primitive (self-contained, round-trip
golden); PR#2 = `card`/`card_section` + icon-registry wiring (store `cfg.icons`
= curated ∪ extensions, `validateIconRefs` closed-name Stage-1 check mirroring
`validateOrnamentRefs`, name→bytes→`AddIcon` at compose). **Grow the Card IR
additively**: new fields' zero values reproduce current output byte-for-byte;
new enums (`BorderStyle`/`CardSize`/`CardLayout`) re-exported into `scene`.
**Keep `Outline` and `BorderStyle` both** — `BorderDefault` (zero) defers to
`Outline` (false→no border, true→neutral solid); a non-default `BorderStyle`
wins. Preservation over folding, so every existing `Card{…, Outline:…}` stays
byte-identical (no field removal).
**Consequences:** The builder gains a general, reusable drop-shadow primitive
(any node, not just cards). Elevation tokens finally drive output. The icon
registry's consumption half lands, closing the Phase-12 deferral; an unknown
icon name fails before compose. Cards render with the full v4 knob set. A
plain (text+icon) card stays media-free and parallel (`AddIcon` is `custGeom`,
not a `pic`); only an image/code-bearing card body renders sequentially —
`nodeUsesAssets` recurses `Card.Body`/`CardSection.Body`. Elevation is a
**mechanism over the existing token** (no new theme token — a THEME.md note,
per D-041), reusing the `Elevation` role.

## D-044 — Flow renders by composition (no new builder API); flow-level connector kind; arrow_dashed = dashed line + chevron

**Date:** 2026-06-03
**Status:** Settled
**Context:** Phase 15 (`flow`) renders a sequential step pipeline — pills joined
by connectors (`arrow`, `arrow_dashed`, `cycle`, `plus`), horizontal or vertical.
The RFC §8.1 sketches `Slide.AddConnector(kind, from, to Anchor)` (an anchored
`cxnSp`) but it was **never built**. Flow connectors are decorative glyphs in the
gaps between pills, not routed between anchors, so they do not need it — they
compose existing preset shapes (the `Arrow` leaf already renders `rightArrow`
etc. via `AddShape`). Two wrinkles: block-arrow presets are filled (can't be
"dashed"), and `pptx.Line` has no line-end arrowhead, so `arrow_dashed` has no
one-shape rendering; and flow steps commonly carry an icon.
**Decision:** Render flow by **pure composition — no new builder API** (do not
build `AddConnector` in V1; defer it to if/when a node needs true routed
connectors). Connectors are a **flow-level** `Flow.Connector ConnectorKind`
applied between every adjacent pair (per-pair `[]ConnectorKind` is a future
additive field). `cycle` = inter-pair arrows plus one trailing return arrow (a
`circularArrow` preset glyph). `plus` = a `mathPlus` glyph per gap.
**`arrow_dashed` = a thin dashed line (`ShapeLine` + `Line.Dash`) plus a small
solid chevron head** (two shapes per connector); real OOXML `lnEnd` arrowheads on
`pptx.Line` are **deferred** (a future builder addition if a node needs
arrow-terminated lines). Grow the IR **additively**: `Flow.Connector` (zero =
`ConnectorArrow`, preserving the prior default) and `FlowStep.Icon` (optional,
resolved through the Phase-14 icon registry; `validateIconRefs` extended to walk
flow steps); `ConnectorKind` is a re-exported scene enum. The step pill is a
dedicated lighter `renderFlowStep` (roundRect + centered label + optional detail
+ optional icon), **not** the heavier card chrome.
**Consequences:** Phase 15 is a single PR (no builder change, so no split).
Flow is media-free (native shapes + custGeom icons) → parallel-safe, classified
not-asset-bearing in `nodeUsesAssets`. No new theme token (pills/connectors reuse
color/radius/space tokens), so no THEME.md change. The unbuilt RFC `AddConnector`
and `pptx.Line` arrowheads remain V1.x candidates, documented here so the gap is
explicit rather than silent.

## D-045 — `code_block` language badge; renderer relocated

**Date:** 2026-06-03
**Status:** Settled
**Context:** The `code_block` raster path (a caller-rendered code image + an
optional caption below) shipped in Phases 06/11 per D-014, but the
`CodeBlock.Language` field has been carried in the IR and never rendered — a
set-but-unused field surfaced by the Phase-16 finalize.
**Decision:** Render `CodeBlock.Language` (when non-empty) as a small **native
overlay badge** — a rounded-rect pill with the language text — inset into the
**top-right** corner of the code image, drawn after the `pic` so it overlays
(shape-tree order = z-order). The badge reuses the card header-pill chrome
precedent (D-043): surface-tone fill, caption text, no new theme token. An empty
`Language` emits no badge (byte-identical to prior output). Relocate
`renderCodeBlock` from `render_leaves.go` to its own `scene/render_code_block.go`
for parity with `render_card.go` / `render_flow.go`. No public API is added (the
badge is compose behavior over the existing field); the per-node policy (D-014:
`pic` + `asset_id`) is unchanged — the badge is a native overlay on top.
**Consequences:** The language field finally drives output. Code blocks read as
labeled snippets. The raster/caption behavior and D-014/D-036 contracts are
unchanged. Pure composition — no builder change, no new token.

---

## D-046 — Reading image dimension headers is permitted; chart contains-to-fit with an aspect warning; ChartPlaceholder

**Date:** 2026-06-03
**Status:** Settled
**Context:** RFC §15.1 requires the V1 `chart` (image-shape, D-004) to warn when
the chart image's aspect ratio diverges from its assigned slot — which needs the
image's dimensions. §7 says the library "does not parse **pixel data**"; the
Phase-11 Image `Fit` comment over-read this as "pixel dimensions … forbidden by
§7" and deferred aspect-aware fit. Image dimensions live in the format **header**
(PNG `IHDR`, JPEG `SOFn`), not in the pixel data; Go's stdlib
`image.DecodeConfig` returns them without decoding pixels.
**Decision:** Reading image **dimension headers** via `image.DecodeConfig`
(stdlib, CGo-free) is **permitted**; decoding **pixel data** remains forbidden
(§7 unchanged in intent — the boundary is now explicit). The `chart` composer
reads the caller bytes' dimensions, places the `pic` to **contain** within its
slot preserving aspect (centered/letterboxed), and raises one `LayoutWarning`
when `|slotAR − imgAR| / imgAR` exceeds **0.15**, with the divergence rounded to
an integer percent for deterministic text (D-035). If `DecodeConfig` fails, the
chart fills the slot and no aspect warning is raised (degrade, never error).
Add `pptx.ChartPlaceholder(box, opts…) *Shape` — a builder helper that draws a
labeled bordered slot ("Chart") without bytes; the chart composer reuses it when
the asset is unresolved (a labeled slot instead of a blank gap, D-036). Fix the
Image `Fit` comment to state the §7 boundary correctly (no behavior change; it
does **not** ship Image aspect-aware fit — out of scope, now unblocked for V1.x).
**Consequences:** Charts size correctly and warn on mismatch. The §7 boundary is
stated, not implied, so future header reads (e.g. aspect-aware image fit) have a
clear precedent. No native chart rendering enters the library (D-004 holds);
`ChartPlaceholder` is the only new public API. Stdlib-only (P4 intact).

## D-047 — Round-trip read reconstructs the navigable model by extending the builder types; 4-PR split

**Date:** 2026-06-03
**Status:** Settled
**Context:** RFC §16 guarantees a pptx-go-authored deck reopens into "the same
Shape model we wrote" — `pres.Slides()[0].Shapes()[0]` is navigable. Today
`pptx.NewFromBytes` / `OpenStream` reconstructs high-level structure (presentation, slides, theme,
masters, sections) but **preserves slide shapes as opaque OOXML** in the
`spTree`; byte/codec round-trip already holds (G6 `ToXML→FromXML` goldens), but
there is no public read API to inspect shapes/fills/lines/text/tables/images.
This is the read-vs-preserve distinction Phase 18 must close.
**Decision:** Phase 18 **reconstructs the navigable model** (preserve alone is
insufficient — RFC §16 outranks the byte-identity acceptance line). The read
model **extends the existing builder types** — add read accessors to
`Shape`/`Fill`/`Line`/`TextFrame`/`Table`/`Image` plus a `Slide.Shapes()`
enumerator — rather than a parallel `Read*` hierarchy ("the same Shape model").
Read accessors are **pure mappings** from the `internal/ooxml` Go structs that
`Open` already populates (via `FromXML`) to the public types — no new XML
parsing, so P3 holds (`pptx` consumes `internal/ooxml` domain structs, never raw
XML). Deliver as a **4-PR split** (one plan): PR#1 shapes + geometry / rotation /
fill / line / shadow + `Slide.Shapes()`; PR#2 text (paragraphs / runs / styles /
links / bullets); PR#3 tables + images; PR#4 a comprehensive
`test/integration/roundtrip_test.go` walking every shipped primitive + IR node
plus a fixture byte-identity check (modulo documented reorderings, D-035).
Reading back a scene `Scene` is **out of scope** (RFC §16 is the builder model;
scene is one-way). The write-only `Fill`/`Line` shapes gain read accessors in
PR#1 so reopened values compare field-equal to authored ones (the golden
assertion).
**Consequences:** pptx-go gains a real read/inspection API — the "R" in the
name. No write-side breaks (accessors are additive). Theme/master/layout/section
already reconstruct, so the round-trip test mostly confirms them. Third-party
read robustness stays Phase 19 (best-effort). Stdlib-only; P1/P3 intact.

---

## D-048 — External-deck read is best-effort graceful degradation (warn, don't preserve); opaque-carrier preservation deferred to V2

**Date:** 2026-06-04
**Status:** Settled
**Context:** RFC §16 commits to lossless round-trip of **pptx-go-authored** decks
(delivered in Phase 18, D-047) and to **best-effort** reading of third-party
decks: "an unrecognized extension element is ignored at parse time, a recognized
one is surfaced … we do not promise round-trip fidelity. V2 invests in
third-party robustness." The master plan's Phase 19 entry (`docs/plans/README.md`,
Wave 6) overstated this as preserving unrecognized OOXML "as opaque `RawShape` /
`RawPart` carriers." Today `XSpTree.UnmarshalXML` silently `d.Skip()`s
unrecognized shape-tree children (data loss with no signal), while the OPC layer
already re-emits every loaded part on save (unmodeled parts round-trip for free).
**Decision:** Phase 19 implements RFC §16's external-deck clause as **best-effort
graceful degradation**, not byte-preserving carriers (the RFC outranks the master
plan, and parks fidelity preservation in V2). Concretely: `NewFromBytes` /
`OpenStream` **never panic** on a third-party deck; unrecognized/dropped content
is surfaced in `Presentation.ReadWarnings()` (a `[]ReadWarning`, de-duplicated per
part+element); and parts pptx-go does not model **pass through unchanged** on
re-save (verified + tested, D-035) — so "`RawPart`" is realized as the existing
OPC pass-through, not a new type. Opaque **`RawShape` preservation** of
unrecognized shape-tree children (re-emitting their raw XML) is **deferred to
V2**. The collection seam stays P3-clean: `internal/ooxml/slide` records bare
element *names* on the part; `pptx` owns the `ReadWarning` mapping. The
master-plan §19 entry is updated to match in the same PR.
**Consequences:** pptx-go opens third-party decks without crashing and reports
its degradation, satisfying the RFC's best-effort posture and the master plan's
no-panic + `ReadWarnings` acceptance — without the risk-heavy raw-XML capture
through the bare-name/RestoreNamespaces codec. External decks lose unrecognized
*shapes* on re-save (warned), but keep unrecognized *parts*. Additive public API
(no write-side break); V1 round-trip fidelity of authored decks is unchanged.

---

## D-049 — Read-path security bounds (§7) are enforced in internal/opc with a caller-configurable limit; read constructors accept Options and log degradation (§8)

**Date:** 2026-06-18
**Status:** Settled
**Context:** The Wave 6 checkpoint audit found two CLAUDE.md §7 invariants
unimplemented on the read path that Phase 18/19 exercise: there was no per-part
memory bound (`io.ReadAll` was unbounded — a malicious external part could OOM
the process, and `ErrPartTooLarge`/the documented 100 MB default did not exist),
and no zip-slip guard (`NormalizeZipPath` did not reject `..` or absolute
entries). Separately (§8), the read path recorded degradations only in
`ReadWarnings()`; no logger could even be injected, since the read constructors
took no options.
**Decision:** `internal/opc` enforces a per-part decompressed-size ceiling at
open (default `DefaultMaxPartBytes` = 100 MB) on both the eager (`Open`/`OpenFile`)
and streaming (`OpenStream`/`OpenStreamFromReader`) paths, returning
`ErrPartTooLarge`; it rejects entries whose normalized path escapes the package
root with `ErrUnsafePartPath` (`safePartPath`). The bound is caller-configurable
through `opc.WithMaxPartBytes` (variadic `OpenOption`, so existing internal
callers compile unchanged and inherit the default). The `pptx` read constructors
(`NewFromBytes`, `NewFromFile`, `OpenStream`) now take `...Option`:
`WithReadPartLimit(n)` maps to the opc bound (n ≤ 0 = unlimited), and
`WithLogger` now applies on read — `addReadWarning` emits a `Warn` event per
distinct degradation when a logger is present, so degradation is visible to logs,
not just the `ReadWarnings` slice. The streaming path validates the declared size
and entry path at open (the body is still read lazily).
**Consequences:** Opening an untrusted deck is memory-bounded and zip-slip-safe
by default, satisfying §7. Read-time observability matches §8 (no global logger;
zero-cost when absent). The opc `OpenOption` seam is additive; no caller breaks.
Build-time options passed to a read constructor (e.g. `WithFormat`) are harmless
no-ops. A lying-header zip bomb on the lazy streaming read remains a smaller
follow-up (the eager path is fully guarded).

---

## D-050 — Speaker notes are reconstructed on open (round-trip), closing a G6 gap and a read-then-save data-loss footgun

**Date:** 2026-06-18
**Status:** Settled
**Context:** `Slide.SpeakerNotes()` (D-022) shipped its write half without a read
half: a reopened deck's notes were invisible (`HasSpeakerNotes()` returned false,
`SpeakerNotes()` returned a fresh empty frame), and merely calling
`SpeakerNotes()` to inspect a reopened deck and then `Save()` overwrote the
existing `notesSlide` part with empty content — silent data loss, a G6
round-trip-fidelity violation for a shipped builder API.
**Decision:** `repopulateSlides` reconstructs each slide's notes from its
`notesSlide` part on open (`slide.ParseNotesBody` extracts the body placeholder's
`<p:txBody>`; the slide→notesSlide relationship locates the part). The
reconstructed `*TextFrame` is the same type the builder writes, so notes are
navigable via `SpeakerNotes()` and re-emit on save. A referenced-but-unreadable
notes part degrades to a `WarnUnreadablePart` rather than failing the open
(best-effort, D-048); external decks with a non-pptx-go notes layout are
best-effort (the first text-bearing shape wins).
**Consequences:** Notes now round-trip losslessly for self-authored decks (G6),
and the inspect-then-save data loss is fixed. Additive; no write-side change.

---

## D-051 — Content-aware `preferredHeight`: a node's slot grows with its wrapped text, and overflow is reported truthfully

**Date:** 2026-06-20
**Status:** Settled
**Context:** The scene layout engine's `preferredHeight` (`scene/render.go`)
allotted a **fixed** slot per node regardless of text length — a `Prose` got
`0.4"` per paragraph, a `List` `0.32"` per item, a `Heading` `0.6"`, etc. — so
a paragraph that wrapped to several lines was given the space of one and its
text frame overran the next stacked node. The same under-count meant the total
stack height was under-reported, so the `RFC §10.2` overflow `LayoutWarning`
never fired when real content ran off the slide. `RFC §10.2` already mandates a
*content-driven* preferred bbox; the fixed-height shortcut was the gap. This is
the first phase of **Wave 8** (post-V1 engine extensions requested by the
product built on pptx-go, `DECKARD-PRODUCT-REQUIREMENTS.md` R1).
**Decision:** `preferredHeight` (and its helper `nodesHeight`) take the
available width and the active theme and become content-aware for the
text-bearing nodes (`Prose`, `List`, `Heading`, `Quote`, `Callout`, `Table`).
A deterministic `wrappedLines` estimate (`scene/metrics.go`) —
`ceil(naturalWidth / availableWidth)`, floored at 1, reusing the Phase-13 pinned
char-width model — feeds `height = lines × line-height`. Each node's prior fixed
constant is reused as its line height, so **single-line content reduces to
exactly the prior height (byte-identical)**, and the `avail ≤ 0` / nil-theme
path returns the fixed height unchanged. The existing `totalH > box.H` overflow
check now consumes content-aware heights, so it fires on real wrapped overflow
with no new warning plumbing. The estimate stays a placement *mechanism*, not a
content opinion — no render mode, no "too full" judgment, no text resizing
(D-026). This is the **one intentional layout change** of Wave 8: multi-line
text now occupies more vertical space (less overlap, truthful overflow);
single-line content is unaffected.
**Consequences:** Stacked nodes stop overlapping and overflow is reported when
content genuinely exceeds the body region, giving callers the truthful
`Stats.Warnings` signal the product needs. No public `pptx`/`scene` API changes
and no new scene IR node — the change is internal to `scene` layout. Determinism
holds (pure integer math; a multi-line fixture renders byte-identically across
worker counts). There are no byte-golden snapshots to regenerate; the
parallel≡sequential determinism guard and the single-line reduction tests are
the regression guards. Grow-to-fit (distributing slack to flexible nodes,
Deckard R2) is the inverse direction and is a separate Wave 8 phase.

---

## D-052 — `VAlignFill` grow-to-fit: flexible nodes consume leftover body height

**Date:** 2026-06-20
**Status:** Settled
**Context:** A heading followed by one block leaves the bottom of the slide empty
and reads thin. The Phase-13 alignments `VAlignCenter`/`VAlignJustify` only
*float* the body stack (move it, or spread the inter-node gaps); the product's
reference "designed" look is the heading pinned at the top with the content
**growing** to fill the frame (tall cards, full-bleed grids). The engine had no
way to express that. Second unit of Wave 8 (`DECKARD-PRODUCT-REQUIREMENTS.md`
R2), built on D-051's content-aware heights (the basis for the slack).
**Decision:** Add `VAlignFill`, a new opt-in value of the `VAlign` enum on
`SceneSlide.Content.Vertical`. It is top-pinned (like `VAlignTop`) with the
standard inter-node gap; after the fixed leaves take their preferred height,
`alignedStackIn` distributes the positive leftover height (`slack = box.H −
totalH`) to the **flexible** nodes — the containers (`Grid`, `TwoColumn`, `Card`,
`CardSection`, `Table`) and the stretchable visuals (`Image`, `Chart`) — so they
grow to consume it. Text leaves and atoms stay at preferred height (stretching
text is meaningless); `CodeBlock` is excluded (growing a monospaced-code raster
distorts the listing). The share is proportional to each flexible node's
preferred height, with the rounding remainder assigned to the last flexible node
(equal split when the flexible heights sum to zero) — pure integer EMU math, so
the result is worker-count independent. **No container renderer changed:** the
scene geometry engine already honors a taller box (`layout.Grid` scales rows to
`parent.H`, `layout.Columns` give full-height columns, card chrome runs to
`box.Bottom()`, image/chart/table consume `box.H`), and the grown slot box
propagates one nesting level (a grown `Grid` hands its taller cell box to the
child renderer). Fill is a *mechanism*, not a judgment (D-026): the engine never
decides a slide "looks thin"; the caller opts a slide into fill.
**Consequences:** Sparse slides can fill their frame on demand, matching the
reference look, with no new public type/function/field beyond the one enum
constant. Additive and fully backward-compatible: every existing `VAlign` value
and the zero value are unchanged, so every existing scene renders byte-identical.
Fill composes with D-051 — it consumes only positive slack, so when content
already overflows (`slack ≤ 0`) nothing grows and the overflow `LayoutWarning`
still fires. Determinism holds (a `VAlignFill` deck renders byte-identically
across worker counts). Deferred: recursive fill *inside* a container (spreading a
tall card's own body children) and per-node grow weights — both noted in
`docs/research/10-grow-to-fit.md`.

---

## D-053 — Opt-in slide chrome: section eyebrow + footer page number outside a shrunk body region

**Date:** 2026-06-20
**Status:** Settled
**Context:** Reference decks read "designed" partly because every content slide
carries a section eyebrow + hairline rule at the top (`01 — DIRECTION`) and a
footer with a brand mark and an `N / total` page number. The engine had no
concept of chrome — recurring per-slide furniture drawn outside the content — so
a caller could not produce it without hand-placing shapes on every slide. Third
unit of Wave 8 (`DECKARD-PRODUCT-REQUIREMENTS.md` R3).
**Decision:** Add opt-in chrome driven by new fields: a `Chrome` struct on
`Scene` (`Enabled`, `Brand`, `BrandAsset AssetID`, `Total`) and `Section` +
`PageNumber` on `SceneSlide`. When `Chrome.Enabled`, `bodyRegion` shrinks by a
fixed eyebrow-band height (top) and footer-band height (bottom), and
`renderChrome` draws — in the reclaimed margin — a top section eyebrow + hairline
rule (only when the slide sets `Section`) and a bottom footer with a brand slot
(left) and an `N / total` page number (right). Shrinking the body makes overlap
structurally impossible. Page `Total` defaults to `len(Slides)` and the per-slide
number defaults to the 1-based scene position; both are overridable. The brand
slot is text-or-asset: brand *text* is a native run (renders in parallel); a
brand *image* (`BrandAsset`, resolved via the existing `AssetResolver`) is the
only global-media touch, so it forces sequential composition deck-wide for stable
part numbering, and an unresolved brand asset degrades to a `LayoutWarning`
(warn-don't-fail). Chrome colors resolve through existing tokens (`TextMuted`,
`ColorSurfaceAlt`) — **no new builder API, no new token** (P2), so no `THEME.md`
entry. Chrome is drawn after the body so the footer stays visible over a
full-bleed background. It is a *mechanism*, not a judgment (D-026): the engine
draws the bands it is handed and composes the page-number string, but invents no
brand and no section names, and never decides a deck "needs" chrome.
**Consequences:** A caller can turn on consistent, theme-aware chrome with a few
fields; the eyebrow is per-slide, the footer consistent. Additive and fully
backward-compatible: every new field's zero value is inert, so a chrome-free deck
is byte-identical to one authored before the fields existed. Determinism holds
across worker counts for both brand text (parallel) and brand asset (serial).
New public scene surface (the `Chrome` type + fields) ⇒ a smoke check lands in
the same PR (§4.2). Deferred: per-slide chrome opt-out, a custom page-number
format, extra footer slots (date/confidential), and authoring chrome as true
master placeholders — all noted in `docs/research/11-slide-chrome.md`.

---

*Append new entries below this line.*
