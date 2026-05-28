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
`pptx.Open`. The scene IR's `SceneSlide.Notes` field maps directly.
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

*Append new entries below this line.*
