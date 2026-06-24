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

## D-054 — Rich card visuals: header band, status dot, watermark (optional colors are `*ColorRole`)

**Date:** 2026-06-20
**Status:** Settled
**Context:** `Card` already supported fill/icon/eyebrow/header-pill/border/size/
elevation, but reference "designed" cards add three visuals it could not express:
a colored header band (the top of the card a solid accent, the body in surface
below — distinct from a full `Fill`), a small status dot in the top-right corner,
and a large low-opacity watermark behind the body (a ghosted `01`). Fourth unit
of Wave 8 (`DECKARD-PRODUCT-REQUIREMENTS.md` R4), the next additive Card growth
after D-043.
**Decision:** Add three additive `Card` fields rendered in `render_card.go`:
`HeaderFill *ColorRole` (a banded header region from the card top to the header
bottom, body keeps `Fill`), `StatusDot *ColorRole` (a small `ShapeEllipse` in the
top-right, inset by the card padding), and `Watermark string` (a large
`TypeDisplay` run drawn in the body region behind the body content, faint via
`TokenColorAlpha` at a pinned ~13% alpha). The header band is sized by a pure
`cardHeaderBottom` helper that shares the header-row height constants with the
emit code, so it ends exactly where the body begins. All three are token-bound
(P2) and compose from existing builder primitives — **no new builder API and no
new token**. Each is opt-in; with all three unset, `renderCardChrome` emits the
same shapes in the same order as before (the header-row literals were extracted
to value-identical constants), so an unset card is byte-for-byte unchanged.
Applies to `Card` only; `CardSection` builds its chrome without the new fields.
**Deviation from the requirement (§4.3).** R4 specifies `HeaderFill ColorRole`
and `StatusDot ColorRole` (value types), but `ColorRole`'s zero value is
`ColorCanvas` — a real color, not "unset" — so a value-typed field cannot satisfy
the same requirement's acceptance "each zero-value omits its element." The fields
therefore ship as `*ColorRole` (nil = omit), the only representation that honors
the binding acceptance without inventing a sentinel role. `Watermark` is a
`string` whose `""` already means omit.
**Consequences:** A card can match the reference look (banded header, status dot,
ghosted watermark) with three opt-in fields; the caller supplies the colors and
label, the engine renders them and picks only the watermark's mechanical faint
opacity (D-026). Additive and backward-compatible (byte-identical when unset);
deterministic native shapes (cards stay parallel-safe). New public scene surface
(the Card fields) ⇒ a smoke check lands in the same PR (§4.2). Deferred: a
pill+dot offset when both occupy the top-right, a watermark color/size knob, and
flat-bottomed band corners — all noted in `docs/research/12-rich-card-visuals.md`.

---

## D-055 — TwoColumn column join: a centered seam badge or connector

**Date:** 2026-06-20
**Status:** Settled
**Context:** Reference decks compare two cards with a centered "VS" badge sitting
on the seam between them, and sometimes link a column to the next with a
connector arrow. `TwoColumn` had no concept of an element between its columns.
Fifth unit of Wave 8 (`DECKARD-PRODUCT-REQUIREMENTS.md` R5), sub-units (a) center
badge and (b) inter-column connector. R5's third sub-unit (c), the row-labeled
bento grid, is a distinct new IR node and lands as its own phase — R5 explicitly
permits separate sub-units.
**Decision:** Add a `ColumnJoin` enum — `JoinNone` (zero), `JoinBadge`,
`JoinArrow` — and fields `Join ColumnJoin` + `JoinLabel string` on `TwoColumn`.
After the two column stacks render, `renderColumnJoin` draws the element centered
on the column seam (`(left.X+left.W + right.X)/2`, vertically centered),
overlapping both columns: for `JoinBadge`, an accent `ShapeEllipse` plus a
centered inverse `JoinLabel` run (the "VS" badge); for `JoinArrow`, an accent
`ShapeRightArrow` connector. Drawn after the column content so it sits on top.
All native shapes reusing existing tokens (`ColorAccent`, `TextInverse`) — no new
builder API, no new token. The single enum with a `JoinNone` zero cleanly
expresses "absent" (no pointer, no companion bool), so an existing `TwoColumn` is
byte-for-byte unchanged. No Stage-1 validation (optional visual); an empty
`JoinLabel` with `JoinBadge` draws the badge shape without text.
**Consequences:** A two-column compare can carry a VS badge or a connector arrow
with two opt-in fields; the caller supplies the label and picks badge-or-arrow
(D-026). Additive and backward-compatible (byte-identical when `JoinNone`),
deterministic native shapes (two-column slides stay parallel-safe). New public
scene surface (the enum + fields) ⇒ a smoke check lands in the same PR (§4.2).
Scope: two columns only — the N-column architecture-diagram connector (arrows
between 3+ columns) is a multi-column-container layout feature, deferred (not in
R5's acceptance), noted in `docs/research/13-column-join.md`.

---

## D-056 — Bento node: a row-labeled grid with variable column spans

**Date:** 2026-06-20
**Status:** Settled
**Context:** Reference decks use a row-labeled bento grid — rows that each carry a
left label and cells of variable column span on a shared column grid (a wide cell
beside two narrow ones, the next row split differently, columns aligned). The
existing `Grid` is uniform (N equal columns, one child per cell, no labels, no
spans) and its `Ratio` is per-column, so it cannot express a bento. Completes
Wave 8 unit R5 (`DECKARD-PRODUCT-REQUIREMENTS.md`), sub-unit (c); sub-units (a)+(b)
(the TwoColumn column join) shipped in D-055.
**Decision:** Add a new container node `Bento{Columns, Rows}`, with
`BentoRow{Label, Cells}` and `BentoCell{Span, Node}`, rather than overload `Grid`
(the requirement's "extend Grid.Ratio" premise doesn't hold — `Ratio` is
per-column). `renderBento` reserves a fixed left-label gutter only when at least
one row is labeled, splits the box into equal-height rows, and lays each row's
cells left-to-right by **absolute** span on a shared unit width (`unitW` from
`Columns`); a span-S cell is `S·unitW + (S−1)·gap`, so a span-1 cell is always
one unit and columns align across rows. The geometry is a pure `bentoGeometry`
helper (unit-tested). Stage-1 validation enforces `Columns ≥ 1`, non-empty
rows/cells, `Span ≥ 1`, non-nil cell nodes, and row spans ≤ `Columns`. The node
is wired through every node switch — `policyTable` (`{}`, native container),
`validateNode`, `renderNode`, `preferredHeight`, `isFlexible` (a bento grows
under `VAlignFill`), `nodeUsesAssets`, and the `walkIconRefs`/`walkImages`/
`walkDecorations` recursions — via a `cellNodes()` helper that flattens cells.
Labels use existing tokens (`TextMuted`) — no new builder API, no new token (P2).
**Consequences:** A deck can lay out a row-labeled bento; `Grid`/`TwoColumn` are
untouched. The catalog grows to 21 kinds and the round-trip "every node" guard
(contiguous `KindHero..KindBento`) covers `Bento`, so a future node that forgets a
switch fails loudly. Deterministic integer-EMU geometry; an asset-bearing cell
still forces serial composition (cells recurse through `nodeUsesAssets`). New
scene IR node ⇒ a smoke check + Stage-1 validation land in the same PR (§4.2).
Deferred: content-height rows, rowspan / per-cell vertical alignment, and
alternate gutter placement — noted in `docs/research/14-bento-grid.md`.

---

## D-057 — Stat node: a hero big-number leaf with a toned delta

**Date:** 2026-06-21
**Status:** Settled
**Context:** Pricing and metric slides use big-number stats (`$2,200`, `38%`)
with a label and an optional delta (`+12%`, colored by direction). The scene IR
had no hero-number node, so callers faked them with `Heading`s, losing the
value/label/delta structure and the directional color. Sixth unit of Wave 8
(`DECKARD-PRODUCT-REQUIREMENTS.md` R6, LOW).
**Decision:** Add a `Stat` **leaf** node — `Stat{Value, Label, Delta string,
DeltaTone}` — rendered as native text (`render_stat.go`): one anchored frame with
the `Value` at `TypeDisplay` (bold), the `Label` at `TypeCaption` (muted), and,
when `Delta != ""`, a `TypeBodySmall` delta run colored by `deltaToneColor(tone)`.
`DeltaTone` (`DeltaNeutral` zero / `DeltaUp` / `DeltaDown`) maps to existing
tokens — `ColorSuccess` / `ColorError` / `TextMuted` — so a theme swap re-skins it
and **no new token** is introduced (P2). The engine renders `Value`/`Delta`
verbatim — it formats no numbers (D-026). A `Grid` of `Stat`s is a metric/pricing
strip with no new container. Stage-1 requires `Value != ""` (label and delta are
optional). As a leaf, `Stat` carries no children and no `AssetID`: its
`policyTable` entry is `{}`, it is in the `nodeUsesAssets` "false" set (never
forces serial), and it is **not** added to `isFlexible` (a number block does not
stretch under `VAlignFill`) nor to any `walk*` recursion. The catalog grows to 22
kinds, and the round-trip every-node guard (contiguous `KindHero..KindStat`)
covers it.
**Consequences:** Callers get a real hero-number node with directional delta
color; a row of them in a `Grid` is a strip. Additive (a new node; existing
scenes unchanged), deterministic native text, parallel-safe. New scene IR node ⇒
a smoke check + Stage-1 validation land in the same PR (§4.2). Deferred: numeric/
currency formatting, a tone-driven ▲/▼ glyph, and value auto-fit — noted in
`docs/research/15-stat-node.md`.

---

## D-058 — Expose resolved per-slide colors in Stats (no contrast logic)

**Date:** 2026-06-21
**Status:** Settled
**Context:** A caller (the product built on pptx-go) validates text/surface
contrast but cannot see the colors the engine actually *resolved* per slide —
especially for `VariantDark`, where the engine swaps to a derived dark palette
(`darkThemeFrom`). The caller's validator checks against the light theme and
false-flags white-on-dark. Final unit of Wave 8 (`DECKARD-PRODUCT-REQUIREMENTS.md`
R7, LOW).
**Decision:** Add a `SlideColors{SlideID string, Canvas, Surface, PrimaryText
pptx.RGB}` type and a `Stats.Colors []SlideColors` field. In `composeOne`, after
`composeSlide` returns, capture the three resolved RGBs from `sr.theme` —
`ResolveColor(ColorCanvas)`, `ResolveColor(ColorSurface)`,
`ResolveTextColor(TextPrimary)`. Because `composeSlide` leaves `sr.theme` as the
derived dark theme for `VariantDark` (and the active theme otherwise), the
captured values are exactly what the codec emitted with — the dark palette
included. The per-slide results merge into `Stats.Colors` in `Render`'s
scene-order loop (like `Timings`). The engine performs **no** contrast logic
(D-026): it returns RGBs only; WCAG ratios and large-text thresholds stay the
caller's. A `Stats` field (not a separate query API) is chosen — the lighter,
already-established observability shape (D-016).
**Consequences:** A caller can compute true text/surface contrast against the
real rendered background, dark variant included, closing the false-flag gap.
Additive and **output-invariant**: `Stats.Colors` is pure metadata, never emitted
into the `.pptx`, so rendered bytes are byte-identical; existing callers ignore
the field. Deterministic and scene-ordered. New observability field ⇒ a smoke
check lands in the same PR (§4.2). **This completes Wave 8 (R1–R7).** Deferred:
exposing more roles (accent, secondary text, surface-alt), per-node resolved
colors, and a post-`Render` query API — noted in `docs/research/16-resolved-colors.md`.

---

## D-059 — Wave 2 (R8–R14) scope: pptx-go implements the engine mechanisms; product requirements are Deckard's

**Date:** 2026-06-21
**Status:** Settled
**Context:** `DECKARD-PRODUCT-REQUIREMENTS.md` was expanded with R8–R14 ("Deckard
Wave 2", 88 professional-bar sub-requirements derived from a reference-vs-
recreation gap analysis). Each sub-requirement is tagged `engine`, `product`, or
`both`. Many `product` (and the product side of `both`) requirements operate on
Deckard's own codebase — `internal/soul/` (the soul/bootstrap/refine system),
`contracts/`, `exportstore/`, `render/` — which does not exist in this repo;
pptx-go is the engine (`pptx`, `scene`, `internal/…`), and D-026 keeps product
behavior (palettes, bootstrap, the autofit loop, brand acquisition) in the
caller.
**Decision:** pptx-go implements the **engine** mechanisms and the **engine side
of `both`** requirements; the `product`-tagged requirements are out of scope for
this repo and are Deckard's to build against the new engine surface. The work is
sequenced into waves by requirement family: **Wave 9** = R9 typography & type
system; **Wave 10** = R10 content fit & density; **Wave 11** = R11 rendering
robustness; **Wave 12** = R12 component primitives; **Wave 13** = R13 backgrounds
& finish; **Wave 14** = R14 coverage classes; **Wave 15** = the R8 theme/soul
*engine* bits (soul-driven dark palette, multi-accent palette, named brand
gradients, dark-variant extensions). Foundational type/theme primitives (Wave 9)
land first because R10–R14 build on them. Each engine sub-requirement (or a
tightly-coupled cluster sharing one code change) is one phase / PR, following the
§16 ritual; every change stays additive, deterministic, and byte-identical when
its new field is unused. Each wave closes with the §17 adversarial checkpoint
audit.
**Consequences:** A clear, durable in-scope filter prevents drift into Deckard-
only work and makes the engine surface the product will consume explicit. The
requirements doc remains the source of per-requirement detail (gap / capability /
spec / accept); the master plan's wave map and the per-phase plans track the
in-scope subset. Requirements whose engine portion is already satisfied (e.g.
R8.2's `pptx.FromTemplate` theme read, R8.10's `Stats.Colors` hook from D-058)
are noted as such in their phase plan rather than re-implemented. Cross-cutting
`both` requirements (e.g. R9.1 font embedding) implement only the engine half
(collect used faces + call `Presentation.EmbedFont`), leaving the provider/
bootstrap half to Deckard.

---

## D-060 — Letter-spacing (tracking) token on FontSpec, emitted as a:rPr/@spc

**Date:** 2026-06-21
**Status:** Settled
**Context:** Pro decks open up eyebrows/labels with wide letter-spacing and
tighten display headlines slightly — the single biggest "designed vs default"
tell — but the engine had no tracking: `FontSpec` was `{Family, Size, Weight,
Italic}`, `RunStyle` carried none, and `toProps` never emitted `a:rPr/@spc`. First
engine unit of Wave 9 (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.3, HIGH · engine;
D-059 scope).
**Decision:** Add `FontSpec.Tracking float64` (letter-spacing in points, signed;
the primary per-role mechanism a soul sets) and an optional `RunStyle.Tracking
*float64` override (nil = inherit the role; non-nil including 0 = win). `toProps`
emits the effective tracking as `a:rPr/@spc = round(pt × 100)` (OOXML 1/100 pt)
when non-zero; zero emits nothing. The OOXML struct gains `XTextProperties.Spc
int` (`spc` attr, omitempty) so the attribute survives parse + re-marshal, and a
`*Run.Tracking()` read accessor returns `Spc / 100` (G6 round-trip). Points are
chosen over em-relative units for a direct OOXML mapping and round-trip clarity.
**Consequences:** A soul can set per-role tracking (and a caller per-run);
tracked-caps eyebrows and tight display are expressible and round-trip losslessly.
Additive and deterministic: zero tracking is byte-identical (the `spc` attr is
omitted), and `round(pt × 100)` is a pure function. New builder visual property ⇒
a `docs/design/THEME.md` taxonomy entry (P2) lands in the same PR. Siblings
line-height (R9.4, paragraph-level `a:lnSpc`) and case (R9.11, run-text
transform) are separate type-detail tokens in their own phases; a scene-side
per-run tracking override is deferred (the role-level mechanism is what the soul
drives).

---

## D-061 — Line-height (leading) token: FontSpec.LineHeight → a:pPr/a:lnSpc/a:spcPct

**Date:** 2026-06-21
**Status:** Settled
**Context:** Pro decks set multi-line display headlines tight (~100–105%) and body
readable (~120–135%); the engine had no leading control (`FontSpec` lacked it,
`XParaProps` carried no line spacing, no `a:lnSpc` was ever emitted). Wave 9 unit
(`DECKARD-PRODUCT-REQUIREMENTS.md` R9.4, HIGH · engine; D-059).
**Decision:** Add `FontSpec.LineHeight float64` (line spacing as a percent of
single; the per-role token a soul sets) and `pptx.ParagraphOpts.LineHeight`
(builder-level per-paragraph value); `AddParagraph` emits `a:pPr/a:lnSpc/a:spcPct`
= `round(pct × 1000)` (OOXML 1/1000 percent) when the value is non-zero and not
100. The scene leaf renderers apply a node's base-role `FontSpec.LineHeight` to
its paragraphs (a `lineH(role)` helper + `plainPara` routing). A
`Paragraph.LineHeight()` read accessor returns the value (G6). New OOXML structs
`XLnSpc`/`XSpcPct` are placed as `pPr`'s first child (schema order), and
`lnSpc`/`spcPct` are registered in `RestoreNamespaces` so they emit with the `a:`
prefix (a write-path correctness fix — bare `<lnSpc>` is invalid OOXML).
**Consequences:** A soul can set per-role leading and the scene tightens/loosens
paragraphs accordingly; additive and deterministic (0/100 emit nothing →
byte-identical; the default theme sets no line-height). Round-trips losslessly.
**Deferred:** feeding leading into the `preferredHeight`/`wrappedLines` estimator
(R9.4's "smaller preferredHeight" acceptance) — the per-line height model is a
fixed constant, not leading-derived, so making it leading-aware is a model rework
folded with R9.5 (per-face metrics) / R10.10 (estimate-actual parity). This phase
delivers the visual leading; the estimator-accuracy refinement follows.

---

## D-062 — Case-transform token: FontSpec.Case → a:rPr/@cap (text preserved)

**Date:** 2026-06-21
**Status:** Settled
**Context:** Reference eyebrows/section labels are uppercase as a *style* decision
(paired with wide tracking); today a caller must pre-uppercase the literal string,
so casing is inconsistent and not theme-controlled. There was no case token in
`FontSpec`/`RunStyle`. Wave 9 unit (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.11, LOW ·
engine; D-059).
**Decision:** Add `FontSpec.Case TextCase` (`CaseNone`/`CaseUpper`/`CaseSmallCaps`)
and an optional `RunStyle.Case *TextCase` override (nil = inherit role). `toProps`
emits the OOXML `a:rPr/@cap` attribute (`all`/`small`) — chosen over rewriting the
run text so the **run text stays original-case** (the transform is a display
attribute that PowerPoint/the rasterizer applies), which round-trips both the text
and the case. A `*Run.Case()` read accessor returns the inverse. `cap` is an
attribute on the already-prefixed `rPr` element, so (unlike line-height's new
`lnSpc` element, D-061) no `RestoreNamespaces` registration is needed. Like
tracking (D-060) it is run-level and flows automatically through `toProps` from
the resolved role — no scene-renderer change.
**Consequences:** A soul can set per-role case (e.g. `TypeCaption = CaseUpper`)
and eyebrow text is cased by the theme without the caller pre-uppercasing;
combined with tracking it reproduces the tracked-caps eyebrow. Additive and
deterministic: `CaseNone` emits nothing (byte-identical), and the **engine does
not** uppercase the default caption role — making the default caption uppercase is
the soul's choice (D-026), not the engine default, so `pptx.DefaultTheme()` is
unchanged. Round-trips losslessly (text + cap). New builder visual property ⇒
`docs/design/THEME.md` entry (P2). Lower/title-case (not OOXML `cap` values) would
need a text rewrite and are deferred.

---

## D-063 — Display font: a first-class TypeDisplay face distinct from HeadingFont

**Date:** 2026-06-21
**Status:** Settled
**Context:** Pro decks pair a serif **display** face for hero/section titles with
a separate sans for headings/labels, but `pptx.Theme` had only `HeadingFont` +
`BodyFont`, so a brand could not cleanly say "serif display, different sans
heading"; `TypeDisplay` always inherited `HeadingFont`. Wave 9 unit
(`DECKARD-PRODUCT-REQUIREMENTS.md` R9.2, HIGH; engine half — D-059).
**Decision:** Add `Theme.DisplayFont string` and a `WithDisplayFont(family)`
`ThemeOption`. When `DisplayFont` is non-empty, the `TypeDisplay` role's family is
it; when empty, `TypeDisplay` inherits `HeadingFont` (byte-identical to a 2-font
theme). `WithFonts(heading, body)` keeps its signature (no break) and is made
`DisplayFont`-aware (it sets `TypeDisplay` from `DisplayFont` when present), so
`WithDisplayFont` and `WithFonts` are order-independent. Only `TypeDisplay` maps
to the display face (H1–H5 stay on `HeadingFont`), matching the reference's
serif-display / separate-heading split. The family flows through the existing run
`a:latin` emit, so a display run renders (and round-trips) with the display
typeface — no OOXML or scene change.
**Consequences:** A soul can assign three independent font tiers (display /
heading-body / mono) and refine one without disturbing the others. Additive and
deterministic: `DefaultTheme().DisplayFont` is empty → `TypeDisplay` stays on
`HeadingFont`, byte-identical. New theme font-scheme field ⇒ a `docs/design/
THEME.md` entry (P2). The product (Deckard) wires `DisplayFont` into bootstrap/
the soul (the `product` side of R9.2); pptx-go provides the theme mechanism. The
per-role `Typography[role].Family` override remains the escape hatch.

---

## D-064 — Per-face width metric on FontSpec for the wrap estimator (not a global registry)

**Date:** 2026-06-21
**Status:** Settled
**Context:** `scene/metrics.go` pinned a single `avgCharWidthFactor = 0.5`
calibrated for the default sans; a serif/display face (or any non-sans) makes
`naturalWidth`/`wrappedLines` mis-estimate advance widths, contributing to
title/body overlaps once a soul uses a non-sans face. Wave 9 unit
(`DECKARD-PRODUCT-REQUIREMENTS.md` R9.5, HIGH · engine; D-059).
**Decision:** Add `FontSpec.AvgCharWidth float64` (a role face's average glyph
advance as a fraction of font size) and have `naturalWidth` use it when set, else
the built-in `~0.5` sans fallback. This is consulted per run via the resolved
`FontSpec` (the run's role's family). **Deviation from the spec's "per-family
table keyed by font family" (§4.3):** implemented as a per-role `FontSpec` field
the soul sets, NOT a mutable global family-keyed registry — it is deterministic
(no shared mutable state, no concurrency hazard), theme-scoped, requires no
invented/unmeasured factor table in the engine, and matches how the other R9
type-detail tokens (tracking/line-height/case) live on `FontSpec`. Roles sharing
a family each carry the factor (minor redundancy; the soul sets the type scale as
a unit anyway). All values are soul-time constants — no runtime measurement, no
DOM — so the estimator stays deterministic.
**Consequences:** A soul's serif/display roles get accurate wrap/overflow
estimates; the field never renders (it tunes the layout estimate only). Additive
and deterministic: `AvgCharWidth == 0` uses the 0.5 fallback, so the default sans
and every existing theme produce byte-identical estimates and output. Documented
in `docs/design/THEME.md` as an estimator input (not a visual token). A
package-level built-in table of well-known faces could later seed the field from
a family name, but is not needed for the mechanism and is deferred.

---

## D-065 — Automatic font-embedding pass: WithFontEmbedding collects used faces and EmbedFonts each

**Date:** 2026-06-22
**Status:** Settled
**Context:** A brand deck's identity is its type, but the engine only had the
per-face embedding *mechanism* (`pptx/fonts.go`: the `FontSource` interface +
`Presentation.EmbedFont` + `WithFontSource`, D-019) — nothing walked the deck and
embedded the faces it actually uses, so a caller had to enumerate and `EmbedFont`
every `(family, style, weight)` by hand. A theme that names a non-system display
face (D-063 routes it through the run `a:latin`) emitted the typeface name but not
the bytes, so PowerPoint and any rasterizer substituted a host sans. Gating unit
of the Wave 9 font cluster (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.1, CRITICAL ·
both; D-059 puts the engine half — collect used faces + call `EmbedFont` — in
pptx-go, leaving the `FontProvider`/soul half to Deckard).
**Decision:** Add an opt-in `pptx.WithFontEmbedding()` Option and an
`autoEmbedFonts` save-time pass run inside `prepareForWrite` (before
`syncPresentationPart`). The pass is gated on `fontEmbedding && fontSource != nil`
— off, or with no source, it makes zero `EmbedFont` calls and output is
**byte-identical**. It walks every slide's runs via a new
`slide.SlidePart.UsedFontFaces()` (the codec-side traversal mirroring
`DroppedDescendants`: shape + table-cell text bodies), collecting the distinct
`(family, bold, italic)` faces — the **only** information an emitted `a:rPr`
carries (it has `b`/`i`, not a numeric weight), i.e. the four OOXML
`embeddedFont` slots. Faces are merged into a set and **sorted by `(family,
bold, italic)`** so the `fontN.fntdata` parts and relationship ids are
byte-identical regardless of render order or worker count. Each face maps to
`weight 700/400` + `style "italic"/""` and is embedded via a new lock-free
`embedFontLocked` (the body of `EmbedFont`; caller holds `p.mu`, matching
`ensurePresentationOPCPart`). A face already recorded (a manual `EmbedFont`) is
skipped via `presentation.PresentationPart.HasEmbeddedFace(typeface, slot)`
(idempotent). A face the source cannot resolve is **warned, not fatal** (same
contract as `register-an-asset`); the Save succeeds with the faces that resolved
and the rest fall through to the host's substitution / fallback chain (R9.6).
**Consequences:** A caller flips one option and a brand-themed deck ships and
renders with its faces on any machine. Additive and deterministic: off / no
source is byte-identical, the sort pins part order, and the pass never fails a
Save. Runs that inherit the theme major/minor fonts (no per-run `a:latin`) are
**not** embedded by this pass — embedding the theme-scheme faces would require
resolving the active theme and is a possible follow-on; the per-run faces are
where a brand display/heading face lands, which is the R9.1 goal. **Deferred:**
subsetting + OS/2 `fsType` license bits (R9.12, LOW → V2) and true per-numeric-
weight embedding (R9.8, needs weight tracked at `AddRun`). New public API ⇒ a
smoke check (`scripts/smoke/phase-35.sh`) and docs/skill updates land in the same
PR (§14/§19).

---

## D-066 — Font fallback chain: FontSpec.Fallback resolved to the emitted face at write time

**Date:** 2026-06-22
**Status:** Settled
**Context:** `pptx.FontSpec.Family` is a single string with no fallback. When a
brand face is neither installed nor embedded, PowerPoint and any rasterizer
silently substitute an arbitrary host font — so a deck that names "Playfair
Display" but ships without it looks like generic Arial rather than a controlled
near-serif. A soul had no way to say "if my display face is unavailable, fall back
to *this* specific serif." OOXML run fonts (`a:latin`) are single-valued, so a
fallback cannot be emitted as a list — it must be *realized* at write time. Wave 9
unit (`DECKARD-PRODUCT-REQUIREMENTS.md` R9.6, MED · both; engine half — D-059).
**Decision:** Add `FontSpec.Fallback []string` (ordered substitute families; empty
= none) and a save-time `resolveFontFallbacks` pass (in `prepareForWrite`, before
`syncSlides`/`autoEmbedFonts`). A registered `FontSource` is the **availability
oracle**: a family whose regular cut the source resolves is treated as available,
one it cannot as unavailable. The pass builds a deterministic `primary → resolved`
map from the active theme (roles iterated in a fixed order `TypeDisplay…TypeCode`,
first-seen wins): for each role with a declared `Fallback`, the effective chain is
`[Family] + Fallback` and the resolved face is the first entry the source
resolves; the primary wins when available. Every run whose `a:latin` typeface
equals a substituted primary is rewritten to the resolved face
(`slide.SlidePart.RewriteFontFaces`). The pass is **self-gated**: a no-op (and
zero `FontSource` calls) when no `FontSource` is registered or no role declares a
`Fallback`, so output is byte-identical when unused.
**Consequences:** A soul declares per-role fallback families and a deck whose
primary face is unavailable renders in the declared substitute (identical across
machines) instead of the host default; with embedding on the resolved face is what
gets embedded (the pass runs before `autoEmbedFonts`), and it ties R9.7 (no italic
cut → fall back rather than faux-italic). Additive and deterministic: empty chain
/ no source is byte-identical, the role order pins the map, and the substitution
is idempotent (after the first save a run carries the *fallback* family, which is
not a primary key, so a second save is a no-op — two saves are byte-identical).
Adding a slice to `FontSpec` makes it non-comparable (`==`); a `ResolveType`
determinism test switched to `reflect.DeepEqual`. Mechanism, not taste: the engine
carries and resolves the chain; the chain contents are the soul's (D-026). The
*resolved* face round-trips via the existing `Run.Font()`; `Fallback` itself is a
theme-time config (like `AvgCharWidth`, D-064), documented in `docs/design/
THEME.md`, not a persisted OOXML field. New visual-affecting type-scale field ⇒ a
THEME.md entry + a smoke check land in the same PR (§14/§19). **Deferred:**
theme-scheme (major/minor) fallback for runs that inherit the theme font.

---

## D-067 — Italic-aware font fallback + embedded <p:font> prefix fix

**Date:** 2026-06-22
**Status:** Settled
**Context:** R9.7 (`emphasis-as-italic-display`, MED · both) wants an italic
emphasis run inside a display/heading role to render in that face's **italic cut**
(a serif italic), not a faux-italic sans. Most of the engine half is already
satisfied: D-063 routes the display family onto the run's `a:latin`, and D-065's
embedding pass already collects `(family, bold, italic)` from each run's `rPr` and
embeds the italic bucket. The gap is the fallback path (D-066): it probed and
substituted at the regular-cut/family level only, so an italic run of a family
that ships *regular but not italic* was left on the primary (the regular cut
resolves → "primary wins"), and the consumer faux-italicized the upright. While
testing this, a latent bug surfaced: the `<p:embeddedFontLst>`'s `<p:font
typeface=…>` child was emitted **bare** (`<font…>`) because `font` was missing
from the `RestoreNamespaces` element-prefix map — invalid OOXML PowerPoint cannot
bind, silently breaking *all* embedding since D-019.
**Decision:** (1) Make the fallback pass **italic-aware**: resolve per
`(family, italic)` rather than per family. For each role with a `Fallback`, probe
the italic cut (`Resolve(family, "italic", 400)`) for the italic key and the
regular cut for the upright key, and substitute each independently — so an italic
run whose family's italic cut is unavailable falls back to the first chain family
whose italic cut resolves, while the upright runs keep the primary. The codec
rewrite generalizes from `RewriteFontFaces(map[string]string)` to a resolver
callback `RewriteFontFaces(func(typeface string, bold, italic bool) string)`
(shaped to also carry weight for R9.8). (2) Add `"font": "p"` to the
`RestoreNamespaces` element map so `<p:font>` is correctly prefixed; `font` only
occurs in presentation's `embeddedFontLst` (the theme's `<a:font>` collection is a
literal string, not processed by this pass; slides emit `a:latin`, never `font`),
so the global mapping is unambiguous.
**Consequences:** An italic emphasis run renders in (and embeds) a real italic cut
— the role's own when present, else a controlled fallback — never a faux-italic
sans. The display-italic guarantee itself (R9.7 accept #1) was already delivered
by D-063+D-065 and is now covered by a verification test. Additive and
deterministic: both cuts of the primary resolving → no substitution (identical to
D-066); no `Fallback`/no `FontSource` → byte-identical to the true baseline; the
only new behavior is the "regular present, italic absent" case (R9.7's target).
The `<p:font>` fix makes embedded fonts actually bind in PowerPoint — verified at
the byte level (the reader matches by local name and hid the bare-element bug);
single-version codec, goldens unaffected (no prior golden emitted an
`embeddedFontLst`). The emphasis-*treatment* choice (italic vs bold vs
accent-color) stays in the soul (D-026, D-059). **Deferred:** bold-cut–specific
fallback (a bold run uses the regular-cut resolution); the resolver callback can
extend to it and to numeric weight (R9.8).

---

## D-068 — Weight-aware font embedding: embed the actual weight file per OOXML bucket

**Date:** 2026-06-22
**Status:** Settled
**Context:** Brands use a weight ladder (300/400/500/700). The embedding pass
(D-065) keyed the used-face set on `(family, bold, italic)` from each run's `rPr`
— which carries only `b`/`i`, no numeric weight — and embedded a *synthetic*
weight (`700` if bold else `400`). So a soul's `500` "medium" role collapsed to the
regular bucket and shipped the 400 file, not the medium it asked for. Wave 9 unit
(`DECKARD-PRODUCT-REQUIREMENTS.md` R9.8, MED · both; engine half — D-059), the
last R9 engine requirement (R9.12 subsetting deferred to V2).
**Decision:** Track the resolved numeric weight per run: add
`slide.XTextProperties.Weight int` with `xml:"-"` (in-memory only — OOXML run
props have no numeric weight; never serialized or parsed, so byte-identical and
round-trip-neutral). `toProps` sets it to the effective weight (the role's
`FontSpec.Weight`, bumped to ≥700 for a per-run bold override). `slide.FontFace`
gains `Weight` and `UsedFontFaces` populates it (inferring `700/400` from the bold
bit when `Weight==0`, e.g. a parsed deck). `autoEmbedFonts` collects the distinct
`(family, weight, italic)` set, groups by OOXML bucket
`(family, weight≥600, italic)`, and per bucket embeds the **actual** weight
nearest the bucket's nominal (`400` regular, `700` bold; ties → the lower weight)
via `EmbedFont` — so the provider returns the correct physical file. When several
used weights collide on one bucket the extra ones are coalesced (logged at Debug).
Deterministic bucket ordering preserves byte-identical part/rel ids.
**Consequences:** A soul's medium/light weights ship the right file within each of
PowerPoint's four cuts per family; additive and deterministic (no `FontSource` /
embedding off → byte-identical; the weight never flips `toProps`'s emit flag, so
unstyled runs still emit no `rPr`). **Deviation (§4.3) from R9.8's literal "embeds
three distinct files" acceptance:** the engine embeds **one file per OOXML bucket**
(the four `embeddedFont` slots), not one per numeric weight. Embedding additional
same-bucket weight files purely so an *external rasterizer* can pick a finer weight
would create `embeddedFontLst`-unreferenced font parts, risking the no-repair-
prompt guarantee (D-020) for zero PowerPoint benefit (PowerPoint exposes only four
cuts). Per D-026 that rasterizer concern is the caller's — a caller can call
`EmbedFont` for extra cuts explicitly. The weight-keyed collector and the resolver
callback (D-067) are shaped to support multi-file-per-bucket later if the
unreferenced-part question is resolved. **Deferred:** subsetting + OS/2 `fsType`
(R9.12 → V2). **Limitation (Wave-9 checkpoint):** the numeric weight is
`xml:"-"` — carried in memory for the authoring session only. A deck written then
re-opened (`NewFromBytes`) loses the numeric weight, so a subsequent embedding
pass infers `700/400` from the bold bit; weight-aware embedding is an in-memory-
authoring property, not a round-trip one. Serializing numeric weight is a
Wave-10+ decision.

---

## D-069 — Wave 9 checkpoint: exclusive save lock + font-pipeline hygiene

**Date:** 2026-06-22
**Status:** Settled
**Context:** The §17 adversarial checkpoint of Wave 9 (Phases 30–38, typography +
font cluster) found one genuine correctness cluster plus documentation/test gaps.
The save methods (`Save`/`Write`/`WriteToBytes`/`SaveStream`) held `p.mu.RLock`,
yet `prepareForWrite` mutates shared builder state — and the Wave-9 font work made
that mutation heavier: `autoEmbedFonts` did a non-atomic `p.fontCounter++` (every
other counter is atomic) and `resolveFontFallbacks` rewrites run structs in place.
Under concurrent saves of the *same* `*Presentation` (an advertised path —
`Write`'s godoc said "high-concurrency streaming output") this is a `-race`-
detectable data race.
**Decision:** (1) The four save entry points take `p.mu.Lock` (exclusive) instead
of `RLock`: `prepareForWrite` mutates, so concurrent saves of one presentation
serialize (saves of distinct presentations are independent; no callee re-acquires
`p.mu`, so no deadlock). Godoc now says to `WriteToBytes` once and share the bytes
(or `Clone`) for high-concurrency fan-out of a single deck. (2) `p.fontCounter`
increments via `atomic.AddInt32`, matching `slideCounter`/`relCounter`/
`chartCounter`. (3) The font-fallback run rewrite is documented as an intentional
in-memory mutation (after a substituting save, `Run.Font` reports the resolved
face; this keeps the pass idempotent). (4) Documentation truth-ups:
`FontSource.Resolve` bytes are embedded **verbatim** — no size cap, no
signature/format validation (the caller's responsibility, parallel to image/SVG
bytes under §7); a `FontSource` may be called from the save path and must be safe
for concurrent use when shared across presentations; role-level
`FontSpec.LineHeight` is applied by the **scene** layer (a direct `pptx` user sets
`ParagraphOpts.LineHeight`); plus the weight-aware embedding re-open limitation
(D-068). (5) Test hardening landed in the same `chore(checkpoint)` PR: a `-race`
concurrent-save guard; combined single-run attributes (tracking+case+bold+color)
and run-overrides-role; the exhausted-fallback-chain branch; a four-cut family
embed-order determinism check; and byte-level/negative OOXML assertions
(`a:lnSpc`/`a:spcPct`, `p:bold`/`p:italic` prefixing, and that `weight=` never
leaks onto `a:rPr`).
**Consequences:** Concurrent same-instance saves are correct (serialized) and the
`-race` suite proves it; the font pipeline's counter and mutation are sound. No
public API change (the lock is internal; the godoc clarifies concurrency). Wave 9
is closed as healthy: the typography token additions remain additive,
byte-identical when unused, and OOXML-valid. Phase plans 30–34 flipped to `Done`.
Deferred to V1.x/V2: weight serialization, an embedded-font public read accessor +
read-side round-trip, and font subsetting.

---

## D-070 — Content-aware card header height (wrapped-title-aware)

**Date:** 2026-06-22
**Status:** Settled
**Context:** A card's header row advanced by a fixed `cardEyebrowRowH`/
`cardTitleRowH` in both `cardHeaderBottom` (body-Y) and `renderCardChrome` (the
emitted header frame + the D-054 band). A header that wraps to two lines in a
narrow card left the body starting at the single-line bottom — the second header
line overlapped the body. First CRITICAL unit of Wave 10
(`DECKARD-PRODUCT-REQUIREMENTS.md` R10.1, engine; opens R10 content-fit).
**Decision:** Add two shared helpers in `scene/render_card.go`:
`cardHeaderColumnW(box,c)` (the true header text column = innerW − icon-left shift
− header-pill reservation) and `cardHeaderRowHeights(box,c)` (eyebrow/title
height = the per-row constant × `wrappedLines(text, role, headerW, theme)`).
`cardHeaderBottom` and `renderCardChrome` both route through them, so the body Y,
the D-054 header band (sized off `cardHeaderBottom`), and the emitted eyebrow/
title text frames all agree and a wrapped header never collides with the body.
`wrappedLines` (brief 09) is the existing deterministic `ceil(naturalWidth/avail)`
estimator; a plain header string wraps as `RichText{{Text: s}}`.
**Consequences:** Long card headers lay out below their wrapped height (no
overlap); the header band sizes to the wrapped height. Additive and
deterministic: `wrappedLines` returns 1 for fitting text, so a single-line header
is byte-identical (same boxes, band, body Y) — only wrapping headers change (the
fix). The per-line advance stays the fixed `cardTitleRowH`/`cardEyebrowRowH`
constant (not D-061 leading-derived), preserving the R9.4 estimator deferral.
**Deferred to R10.10 (estimate-actual-parity):** making `preferredHeight`'s
`cardChromeEst` (a fixed ~1.2") wrapped-header-aware — this phase fixes the
*composed* geometry (the CRITICAL overlap); the slot-size estimate parity follows
(same visual-first / estimator-later split as D-061).

---

## D-071 — Fit-to-region compression (VAlignFit)

**Date:** 2026-06-22
**Status:** Settled
**Context:** When a slide's body stack was taller than its region, the scene
renderer placed the overflowing nodes off-box and only recorded a
`content overflows its region` warning — the content clipped below the slide
edge (the recreation drew its bottom bento row partially off-canvas yet shipped).
`alignedStackIn` had no mechanism to make an over-full stack fit. Second CRITICAL
of Wave 10 (`DECKARD-PRODUCT-REQUIREMENTS.md` R10.2, engine).
**Decision:** Add an opt-in `VAlignFit` value to the `VAlign` enum (the
compression inverse of `VAlignFill`), set via `SceneSlide.Content.Vertical`. When
the mode is `VAlignFit` and the stack overflows (`totalH > box.H`),
`alignedStackIn` calls a new renderer method `fitCompress(heights, bodyH, gap,
box)` that runs two pinned steps in priority order: (1) shrink the inter-node gap
toward `gapMin = ResolveSpace(SpaceXS)` — `effGap = clamp((box.H - bodyH)/(n-1),
gapMin, gap)`; (2) if still overflowing at the gap floor, scale every slot height
by a single basis-point factor `sBP = clamp(avail*10000/bodyH, 6000, 10000)`
(where `avail = box.H - effGap*(n-1)`) toward the pinned ratio floor `sMin=0.60`.
It mutates `heights` in place and returns the compressed gap; placement is
top-pinned. The overflow warning is recomputed against the post-compression
geometry for `VAlignFit` (a successful fit suppresses it; an overflow the floors
cannot absorb still surfaces), while every non-Fit mode keeps the verbatim
`totalH > box.H` warn. All math is integer EMU / basis-point — a pure function of
the heights, the gap, and `box.H` — so output is identical regardless of worker
count.
**Consequences:** An over-full `VAlignFit` slide (up to ~25% overflow) lands its
last node bottom ≤ region bottom using only the pinned steps; a stack that
already fits is byte-identical to `VAlignTop` (the compression branch is skipped),
and with the flag off no path changes. The enum value is appended, so all
existing `VAlign` iota values and the zero value (`VAlignTop`) are unchanged.
`fitCompress` ships as a reusable theme-aware primitive but only `alignedStackIn`
(the top-level body stack — the CRITICAL off-slide-clip site) calls it this
phase. **Deviation (§4.3) from R10.2's literal spec:** the card-interior-padding
sub-step (→ `padMin`) and the explicit display-type-scale step are deferred to
R10.7 (`density-aware-card-padding`, an auto-tighten step inside the same pass)
and R10.5 (`display-text-shrink-to-fit`) respectively, per the R10.2 spec's own
cross-references; container-internal fit (bento rows, card body) is deferred to
R10.3 / R10.4, which will reuse the `fitCompress` primitive and pinned floors.
The gap + slot-height steps satisfy the ≤25%-overflow acceptance on their own.
Extends `D-052` (`VAlignFill`); builds on `D-070` (content-aware card header).

---

## D-072 — Content-weighted bento rows (Bento.WeightedRows)

**Date:** 2026-06-22
**Status:** Settled
**Context:** `bentoGeometry` forced every bento row to the same height
`(box.H − gaps)/nRows` regardless of content. A sparse one-line row and a dense
four-line row got identical bands, so the sparse row wasted a huge area while the
dense row starved and overflowed off-slide (the recreation's slide-6 "Canvas"
bento). R10.3 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine), the Wave-10 unit
after the two CRITICALs.
**Decision:** Add an additive `Bento.WeightedRows bool`. When set, each row's
height is the preferred height of its tallest cell at that cell's span width
(`r.bentoWeightedRowHeights`); if the rows plus gaps would overflow the region, a
single basis-point scale `sBP = avail·10000/Σpref` (flooring, so `Σ ≤ avail`)
clamps every row so all rows fit inside the frame — unlike the R10.2 `fitCompress`
ratio floor, the bento clamp has no floor because the rows must always fit. When
they fit, rows keep their preferred height (top-aligned; leftover slack is bottom
whitespace). `bentoGeometry` is refactored: the shared horizontal math moves to
`bentoColumns` (gutter, content X, unit width) and `cellWidth(span)`, and it now
returns per-row `rowYs`/`rowHs` slices plus the cell boxes and accepts an optional
`rowHs []EMU` (nil ⇒ equal mode). `renderBento` threads the weighted heights in
when `WeightedRows` and uses the per-row Y/H for both gutter labels and cells, so
labels anchor-middle within their actual row height. All math is integer / basis
point — deterministic regardless of worker count.
**Consequences:** A weighted bento gives a dense row more height than a sparse one
and never overflows the region. The zero value (`WeightedRows=false`) reproduces
the equal-row layout byte-for-byte (the per-row-array refactor fills every row
with the same `rowH` and accumulates `rowY` identically). An additive bool field —
`Bento` stays comparable; no round-trip codec (scene IR is rendered, not
serialized). **Deviation (§4.3) from R10.3's literal spec:** (1) the **Grid analog**
is deferred — the acceptance is bento-specific, `Grid` cells are single children
laid out by the pure theme-free `layout.Grid`, and content-weighting Grid is a
separable change (folds into R10.10 / a follow-up); (2) the leftover-slack
distribution defaults to **top-align** rather than proportional fill (the spec
lists top-align as an option, and it already prevents a sparse row stealing space
from a dense one). The **slot estimate** (`preferredHeight` for the bento) is left
unchanged this phase; estimator/actual parity (wide-span cells, weighted rows) is
R10.10's explicit job. Extends `D-056` (Bento); mirrors `D-071`'s scale-to-fit
shape.

---

## D-073 — Card body vertical distribution (Card.BodyVAlign)

**Date:** 2026-06-22
**Status:** Settled
**Context:** `renderCard` always laid a card's vertical body out top-anchored via
`stackIn`, so secondary content floated in the upper card with large dead space
below (the recreation's Vision/Mission lists, the slide-8 path cards, and the
slide-9 pricing cards left ~40–60% of the card blank). R10.4
(`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine).
**Decision:** Add an additive `Card.BodyVAlign VAlign`. `renderCard`'s vertical
body now routes through the existing body-stack alignment engine —
`r.alignedStackIn(body, v.Body, slideID, Alignment{Vertical: v.BodyVAlign})` —
instead of `r.stackIn`. This gives the card body the full `VAlign` set on its own
body box: `VAlignCenter` offsets the start Y, `VAlignBottom` pins the last node to
the body bottom, `VAlignJustify` widens the inter-node gaps, `VAlignFill` grows
flexible body nodes (`distributeFill`), and `VAlignFit` (D-071) compresses an
over-full body. The `BodyHorizontal` column path is unchanged.
**Consequences:** A card can pin secondary content to its bottom or spread it to
fill, eliminating dead space. The zero value (`VAlignTop`) is byte-identical to
the prior top-anchored layout: `alignedStackIn` with `{VAlignTop, HAlignLeft}`
emits the same per-node boxes as `stackIn` and warns on the algebraically
identical condition (`totalH > box.H` ⟺ last-bottom > `box.Bottom()`), as its
godoc already documents. An additive field — `Card` keeps its shape; the existing
card golden/determinism tests pass unchanged through the new path. **Deviation
(§4.3):** `BodyVAlign` is added to `Card` only; `CardSection` (whose body is
containers, not leaves) keeps top-anchored `stackIn` — a `CardSection.BodyVAlign`
is a separable, lower-value follow-up. Extends `D-043` (Card additive fields);
reuses the Phase-13 alignment engine and `D-071` (VAlignFit).

---

## D-074 — Display text shrink-to-fit (AutoFit + RunStyle.FontScale)

**Date:** 2026-06-22
**Status:** Settled
**Context:** A display-class run (Hero title, Stat value, big price, Heading) was
rendered at a fixed `TypeRole` size and wrapped or clipped when wider than its box
(the recreation's "$4,000+" wrapped to two lines in a narrow pricing column;
titles on slides 4/5 wrapped where the reference keeps one line). R10.5
(`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine).
**Decision:** Two additive pieces. (1) **Builder:** a per-run
`RunStyle.FontScale float64` multiplier on the resolved type-role size. In
`toProps`, `size := spec.Size; if rs.FontScale > 0 { size = spec.Size ×
rs.FontScale }`; the result is emitted as `a:rPr/@sz` and round-trips via
`Run.FontSize`. The role's size token stays the source of truth (a theme swap
re-skins the base, then this scales it), so P2 is intact; there is no per-role
`FontScale` token. (2) **Scene:** an opt-in `AutoFit bool` on `Hero`, `Stat`, and
`Heading`, plus a pure `fitScale(natW, boxW)` that returns 0 when the text already
fits, else `floor(boxW·10000/natW)` quantized **down** to a 0.025 step and floored
at a 0.60 ratio — expressed as a fraction in `[0.60, 1)`. The renderer sets the
display run's `FontScale = fitScale(naturalWidth(display text at its role),
box.W)`; `addRichText` is factored into `addRichTextScaled` so a multi-run Heading
scales as a unit. All integer / quantized basis-point math — no measurement, pure
function of `(text, role, boxW, theme)`.
**Consequences:** An over-wide display run downscales to fit its box on one line
at a font no smaller than 60 % of the role size (an extreme overflow caps at the
floor and accepts residual). `FontScale` never upscales. The zero `FontScale` (and
`AutoFit=false`, and already-fitting text where `fitScale` returns 0) is
byte-identical: `size = spec.Size` exactly and no scale is applied — the existing
Hero/Stat/Heading render tests pass unchanged through the new path. **Scope
(§5/§4.3):** `AutoFit` is limited to the display class (`Hero`/`Stat`/`Heading`)
the spec names; `Chip`/`Arrow`/table-cell labels are out of scope but reuse the
same `fitScale` + `FontScale` mechanism if a later req wants them. AutoFit changes
only the emitted font size (horizontal fit) — it does **not** alter
`preferredHeight`; vertical fit stays R10.2/R10.10's concern. Reuses `D-064`
(`naturalWidth`); mirrors `D-071`/`D-072`'s quantized-scale-to-a-pinned-floor
shape.

---

## D-075 — Fill cap / no over-grow (VAlignFillCapped)

**Date:** 2026-06-22
**Status:** Settled
**Context:** `VAlignFill` grows flexible nodes proportionally to their preferred
height with no ceiling, so a near-empty node balloons (the recreation's "Canvas"
card grew to an enormous height holding one sentence while the dense rows
overflowed). R10.6 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine).
**Decision:** Add an opt-in `VAlignFillCapped` value to the `VAlign` enum (after
`VAlignFit`). It calls a new `distributeFillCapped(nodes, heights, slack)` that
grows each flexible node by its proportional share *capped* at `fillGrowthMaxBP ×
its preferred height` (`fillGrowthMaxBP = 10000` → at most +1.0×, i.e. a node can
at most double), and returns the total growth used (`≤ slack`). `alignedStackIn`
then distributes the residual (`slack − used`) as balanced spacing — `space =
residual/(n+1)` added to the top margin (`startY`) and to each inter-node gap
(`effectiveGap`) — reusing the Justify/Fit offset-and-gap mechanism, so the
leftover reads as even whitespace rather than one oversized node. All integer /
basis-point math — deterministic regardless of worker count.
**Consequences:** A sparse node in a capped-fill stack grows by no more than its
cap and the surplus becomes even spacing; the placed stack stays within the box
(`used ≤ slack`, floored spacing). Uncapped `VAlignFill` is untouched (still calls
the unchanged `distributeFill`) — byte-identical; the enum value is appended, so
all existing `VAlign` values and the zero value are unchanged. **Deviation
(§4.3):** a pinned `growthMax` is used, not the spec's alternative per-node
`MaxGrow` — a per-node cap would touch every flexible node type for marginal
benefit; `distributeFillCapped` can take a per-node cap later. Orthogonal to
`D-071` (VAlignFit compresses an over-full stack; this bounds an under-full one).
Extends `D-052` (`VAlignFill`/`distributeFill`); mirrors `D-071`'s opt-in-new-mode
shape.

---

## D-076 — Density-aware card padding (Card.PaddingScale)

**Date:** 2026-06-22
**Status:** Settled
**Context:** `cardPadding` mapped the 3-value `CardSize` enum to fixed
`SpaceSM/MD/XL`, so a dense card carried the same generous interior inset as a
sparse one, wasting interior space (the recreation's dense cards use generous
fixed padding where the reference packs tighter). R10.7
(`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine).
**Decision:** Add an additive `Card.PaddingScale int` — a basis-point multiplier
on the size-resolved padding. A new `paddingScale` field on the internal
`cardChrome` carries it, and a new `cardPaddingFor(c)` method returns
`cardPadding(c.size)` scaled by `paddingScale` (when > 0 and ≠ 10000), floored at a
pinned `padMin = ResolveSpace(SpaceXS)` so a tightened card never collapses its
inset. The three padding sites (`cardHeaderColumnW`, `cardHeaderBottom`,
`renderCardChrome`) route through `cardPaddingFor`; `cardPadding(size)` stays the
base resolver. Both the base and the floor resolve through theme spacing tokens —
no literals (P2). Deterministic integer math.
**Consequences:** A tighter `PaddingScale` (e.g. 5000) shrinks the inset on all
sides and grows the card body (the body box is computed below the header inside
the padding), letting a dense card reclaim interior space; an extreme scale floors
at `SpaceXS`. The zero value (and 10000) returns the base unchanged — byte-
identical to the prior SM/MD/LG output; the existing card golden/determinism tests
pass through `cardPaddingFor`. `CardSection` builds a bare `cardChrome`
(`paddingScale` 0), so it is unaffected. **Deviation (§4.3):** ships
`Card.PaddingScale` only, not the spec's auto-tighten-inside-the-fit-pass
alternative — the fit pass (D-071) is stack-level and does not reach inside a card
body; an auto-tighten hook can later reuse the `cardPaddingFor` seam. Extends
`D-043` (Card additive fields); mirrors `D-074`'s basis-point-multiplier +
pinned-floor pattern.

---

## D-077 — Balanced vertical rhythm (VAlignBalanced)

**Date:** 2026-06-22
**Status:** Settled
**Context:** On sparse slides the body stack used fixed inter-node gaps, so the
elements clustered with a large void — the recreation cover clusters
eyebrow/title/subtitle in the middle then drops the description far below; the
closing leaves the lower frame empty. `VAlignJustify` spreads slack into the gaps
but pins the stack edge-to-edge (no margins); `VAlignCenter` adds margins but keeps
fixed gaps. R10.8 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine).
**Decision:** Add an opt-in `VAlignBalanced` value to the `VAlign` enum (after
`VAlignFillCapped`). When the stack has positive slack, `alignedStackIn` computes
`unit = slack/(n+1)` and distributes it across the stack's `n+1` spaces — a top
margin plus widened inter-node gaps (`effectiveGap = gap + unit`) — with an
optical-center upward bias: the top margin is `balancedOpticalBP = 8500` (85 %) of
an even unit, so the stack seats slightly above geometric center (the freed space
falls to the bottom margin). This reuses the `(n+1)`-even-unit residual primitive
from `VAlignFillCapped` (D-075). All integer / basis-point math — deterministic
regardless of worker count.
**Consequences:** A sparse cover/closing reads as an even rhythm (margins +
widened gaps, no single large void) optically centered, instead of clustered. It
is distinct from `VAlignJustify` (all slack into gaps) and `VAlignCenter` (all
slack into equal margins). `VAlignTop`/`VAlignCenter`/`VAlignJustify` are untouched
(byte-identical); the enum value is appended, so existing `VAlign` values and the
zero value are unchanged; with no slack `VAlignBalanced` is `VAlignTop`.
**Deviation (§4.3):** the spec's "gaps weighted by a pinned ratio (larger gap
before a description block)" is left to the caller — knowing which node is a
"description block" is content taste (D-026); the engine ships the even rhythm +
the optical-center bias (its pinned ratio), and a caller that wants a larger
pre-description gap orders its nodes or inserts a spacer. Reuses `D-075`'s
`(n+1)`-unit primitive.

---

## D-078 — List bullet indent density (List.Indent + ParagraphOpts.BulletIndent)

**Date:** 2026-06-22
**Status:** Settled
**Context:** `Paragraph.Bullet` hard-coded a 0.5" hanging indent (`MarL = 457200`,
`Indent = −457200`), so every list marker sat a wide fixed gap from its text —
lists read loose and sparse (the recreation's "Understand/Operate/Execute" and the
slide-3 checklist show a very wide marker-to-text gap). R10.9
(`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine).
**Decision:** Two additive pieces. (1) **Builder:** a per-paragraph
`ParagraphOpts.BulletIndent EMU`. In `AddParagraph`, when a bullet is set and
`BulletIndent > 0`, override the `MarL`/`Indent` that `Bullet` assigned with
`MarL = BulletIndent`, `Indent = −BulletIndent`, so the marker-to-text offset
becomes `BulletIndent` uniformly (emitted as `a:pPr/@marL` + `@indent`). (2)
**Scene:** a `ListIndent` enum (`IndentNormal` = 0 / default, `IndentTight`) +
`List.Indent`; `renderList` passes `BulletIndent = bulletIndent(v.Indent)` —
`0` for normal, the pinned `listTightIndent = In(0.25)` for tight (about half the
0.5" default). The presets are pinned and deterministic; there is no per-role
token (a bullet hanging indent is a layout mechanism, not a visual style token —
P2's token-default rule is for color/typography/spacing/radius/elevation).
**Consequences:** A tight list shows a smaller, consistent marker-to-text offset
across all items and levels (the override is level-independent, matching the prior
behavior); the emitted `marL`/`indent` round-trip through the slide XML. The zero
`BulletIndent` (and `IndentNormal`) keeps the builder's default 0.5" hanging
indent — byte-identical; the existing list render tests pass unchanged. The
`ParagraphOpts.BulletIndent` seam accepts any EMU, so a direct `pptx` caller can
set a continuous value; the scene exposes presets per the spec. Mirrors `D-061`'s
(`LineHeight`) per-paragraph-metric pattern (byte-identical at zero, emits an
`a:pPr` attribute).

---

## D-079 — Estimate/actual parity (wrapped card chrome + bento span width)

**Date:** 2026-06-22
**Status:** Settled
**Context:** `preferredHeight`'s slot estimates diverged from the composed
geometry, so overflow detection and the fit pass operated on wrong numbers:
`cardChromeEst` was a fixed ~1.2" regardless of a wrapped multi-line header (the
parity R10.1/D-070 explicitly deferred), and the bento estimate measured every
cell at the unit column width even though a wide-span cell renders wider and wraps
less. R10.10 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine).
**Decision:** (1) The Card/CardSection `preferredHeight` becomes
`cardChromeEst + cardHeaderExtraHeight(theme, avail, c) + body + estGap`, where
`cardHeaderExtraHeight` is the extra eyebrow/title lines beyond the first — each at
its per-row constant (`cardEyebrowRowH`/`cardTitleRowH`), measured at the card's
true header column width via the shared `cardHeaderColumnWOf`. A single-line (or
empty) header gives an increment of 0, so the estimate is **byte-identical**; a
multi-line header grows the slot to account for the wrapped header (the R10.1
deferral, sharing the same constants). (2) The Bento `preferredHeight` measures
each cell at its actual span width `span·unitW + estGap·(span−1)` instead of the
unit width; a span-1 cell yields `unitW` (byte-identical), a wide-span cell wraps
less so its over-counted estimate shrinks to the accurate value. The card-header
helpers (`cardPadding`/`cardPaddingFor`/`cardHeaderColumnW`) are refactored to
theme-taking free functions (`cardPaddingBase`/`cardPaddingScaled`/
`cardHeaderColumnWOf`) with thin `*renderer` method wrappers, so the estimators
share the composers' logic without changing any composer or test call site.
**Consequences:** The overflow warning and `VAlignFit` now operate on estimates
that match what the composers emit (within one line-height for the representative
wrapped-header card and wide-span bento cell). Single-line headers and span-1
bento cells are byte-identical (the increment is 0 / span width equals unit
width); the existing card/bento determinism + height tests pass unchanged. No new
public API — `cardChromeEst`/`estGap` stay pinned constants and the change is an
internal accuracy improvement (the only user-visible effect is a more accurate
overflow warning, a taller slot for wrapped-header cards, and a tighter slot for
wide-span bento cells). **Deviation (§4.3):** the spec also asks the card body
inset estimate to match `cardPadding`; `cardBodyInsetEst` is left pinned because
matching it would change the body wrap count → single-line output, and the chrome
+ span-width fixes already deliver the within-one-line-height accuracy — exact
inset parity is a future refinement. Closes the `cardChromeEst` parity deferred by
`D-070`; mirrors `D-072`'s span-width (`cellWidth`) geometry.

---

## D-080 — Wave 10 checkpoint: doc/accessor hygiene + three documented intentional behaviors

**Date:** 2026-06-22
**Status:** Settled
**Context:** The §17 adversarial checkpoint of Wave 10 (Phases 39–48, the
content-fit & density layout cluster, D-070..D-079) ran 40 agents over 8
dimensions with 2 skeptics per finding. It found **no broken runtime invariant** —
the additive byte-identity, integer-EMU/basis-point determinism, division/negative
guards, and the no-new-OOXML-element constraint all held against the source. The 13
confirmed findings were documentation drift and white-box test-coverage gaps.
**Decision:** Land the punch list as one `chore(checkpoint)` PR. (1) **Doc
hygiene:** the `docs/site/reference/pptx.md` struct snapshots are refreshed to
include the Wave-9/10 additions (`RunStyle.Tracking`/`Case`/`FontScale`,
`ParagraphOpts.LineHeight`/`BulletIndent`) and the new read accessors; the
`Card.BodyVAlign` enumeration is completed to all 8 `VAlign` modes (it fed
`alignedStackIn` directly, so `VAlignFillCapped`/`VAlignBalanced` were always
reachable) in the glossary, the catalog, and the skill; the Phase-48 plan file
listing and a card-body call-site comment are corrected. (2) **Accessor:**
`Paragraph.BulletIndent()` is added to restore the read-inverse pattern every
sibling field has (`Run.FontSize`, `Paragraph.LineHeight`, `Run.Tracking/Case`);
the round-trip test now asserts the Go-model accessor, not only raw XML. (3)
**Test hardening** in the same PR: the wrapped-header-card overflow warning
(Phase-48 criterion 4), `fitCompress` n==1 / ~25%-band / extreme-overflow-still-
warns, `distributeFillCapped` zero-preferred-flex, `VAlignBalanced` single-node,
`cardHeaderExtraHeight` eyebrow wrapping, `FontScale`+bold and a dirty-quantum
(0.65) round-trip, parallel-determinism guards for the float `AutoFit` path and
the weighted bento, three cross-feature interaction tests, and a card-body
per-node-`Align` guard.

**Three intentional behaviors documented (not defects):**
- **Bento `WeightedRows` slot estimate stays a uniform over-estimate.** The bento
  `preferredHeight` (`render.go`) uses `nRows × global-max-cell` and does not
  branch on `WeightedRows`, while the composer sizes rows per-row. The estimate is
  always ≥ the composed height (over-estimate → taller slot → no clip), so it is
  safe; reconciling it exactly would couple the estimator to the weighted-row pass
  for no overflow-safety gain. Deferred.
- **Bento clamps silently (no overflow warning).** `bentoWeightedRowHeights` /
  `bentoGeometry` floor their scale and never call `r.warn` (unlike
  `alignedStackIn`); the `content overflows its region` warning is a slide-stack
  signal. Pre-existing for equal rows; the R10.3 clamp inherits it intentionally.
- **`FontScale` emits a truncated `@sz`.** `toProps` computes `int(size ×
  FontScale × 100)` (truncate), matching the pre-existing unscaled `int(spec.Size
  × 100)` sz emission — consistent within the size attribute (the `math.Round`
  used for `@spc`/tracking is a different attribute). Deterministic; documented via
  a dirty-quantum (0.65) round-trip test.
**Consequences:** Wave 10 is closed as healthy — the density/fit cluster's
additivity, determinism, and bounds hold, and the reference docs, the
`BulletIndent` read inverse, and the new-branch test coverage are restored. No
public API change beyond the additive `Paragraph.BulletIndent()` accessor. The
three documented behaviors are settled as intentional. Audit dimensions and the
full punch list are preserved in the workflow transcript.

---

## D-081 — R11.1 closed by D-070/D-079 (card-header content-aware height verify-and-close)

**Date:** 2026-06-22
**Status:** Settled
**Context:** R11.1 (`DECKARD-PRODUCT-REQUIREMENTS.md`, CRITICAL · engine) requires a
card's header band height and body-start Y to grow with the *actual* wrapped line
count of the eyebrow + title — at any title length, any inner width, any
size/layout — so a long header never spills past its band onto the body. R11.1's
spec text describes verbatim the mechanism already shipped in R10.1/D-070:
`cardHeaderColumnWOf` (the true header column width, icon/pill aware) +
`cardHeaderRowHeights` (`per-row-const × wrappedLines`), consumed *identically* by
both `cardHeaderBottom` (body-region top) and `renderCardChrome` (the D-054 header
band + emitted text frames), with single-line headers byte-identical. The estimator
side closed in R10.10/D-079 (`cardHeaderExtraHeight` feeds `preferredHeight`). The
open gap was test coverage: R11.1's acceptance names "a golden test … across all
`CardSize/CardLayout` combinations", but the existing guards covered one combo.
**Decision:** Close R11.1 as **already implemented by D-070** (mechanism) **and
D-079** (estimator), with no renderer change. The close is the named acceptance
golden — `TestCardBodyBelowWrappedHeader_AllCombos` — which sweeps a deliberately
long, wrapping header across `{CardSizeMD, SM, LG} × {CardLayoutDefault,
CardLayoutIconTop}` (6 combos) and asserts, per combo: (1) the composed body top
**equals** `cardHeaderBottom` (so the D-054 band, drawn at height
`cardHeaderBottom − box.Y`, meets the body exactly — the wrapped header is fully
inside the band, no overlap and no drift); (2) the header actually wraps to ≥ 2
lines (the test is not vacuous); (3) a single-line header yields
`titleH == cardTitleRowH` (byte-identical to the legacy fixed advance). Per
`CLAUDE.md §17`, proving the invariant under the full combinatorial content the
requirement names — then recording the closure — is the correct shape for a
verify-and-close; reimplementing already-correct geometry would be churn.
**Consequences:** R11.1 is closed; the card-header overlap class is regression-
guarded across the size/layout matrix, not a single combo. No public API change, no
new token, no OOXML change. Opens Wave 11 (R11 component-rendering robustness). The
next units add new mechanism: R11.2 (card-text auto-contrast), R11.3
(container-slide-bounds clamp).

---

## D-082 — Card-text auto-contrast (onCardSurface mechanism)

**Date:** 2026-06-22
**Status:** Settled
**Context:** A card/container chrome run left uncolored (the card header
`RunStyle{TypeRole: TypeH3}` with no `Color`) emits no `a:solidFill` and inherits
the slide's near-black placeholder default, regardless of the surface behind it. On
a dark card fill or a VariantDark slide the header is black-on-dark and barely
legible (recreation slides 3, 7); a `TextAccent` eyebrow on a same-hue header band
is invisible (slide 2). R11.2 (`DECKARD-PRODUCT-REQUIREMENTS.md`, CRITICAL · engine)
requires every chrome run to derive its color from the surface it sits on. This
tensions with **D-058**, which deliberately gave the engine *no* contrast logic
(only resolved-color exposure via `Stats.Colors`), leaving the decision to the
caller.
**Decision:** Add a deterministic, pinned auto-contrast **mechanism**
(`scene/contrast.go`) and wire it into the chrome whose surface is known at
emission. `onCardSurface(bg ColorRole) pptx.Color` resolves `bg` against the active
(possibly dark-variant) theme, computes WCAG sRGB **relative luminance** (a
256-entry `srgbLinear` table built once at init via `math.Pow`, then integer
weighted-sum per call — worker-count independent), and returns the light
`TextInverse` token when the surface is below the black/white crossover
`darkSurfaceLumaMax = 17912` (`L ≈ 0.179`), or **nil** otherwise. nil leaves the
run's `Color` unset → the inherited dark default → **byte-identical** on a light
surface. The threshold being the exact crossover guarantees both branches clear
~4.58:1. Wired sites: the card header title and pill (`onCardSurface`), the eyebrow
(keeps `TextAccent` when `accentLegible(surface)` clears 4.5:1 — true for the
default accent on a white card, so the common eyebrow is byte-identical — else
falls back to `onCardSurface`), the TwoColumn join-badge label
(`onCardSurface(ColorAccent)`, nil → `TextPrimary`; byte-identical to the prior
hardcoded `TextInverse` on the dark default accent), and the Stat value
(`onCardSurface(ColorCanvas)`).
**Reconciliation with D-058 / D-026:** the engine remains unopinionated — this is a
*mechanism the caller drives*, not a legibility *policy*. The luminance rule is
fixed and pinned (the color analog of `deltaToneColor`'s tone→token map); a
caller's explicit `Color` always wins; and the light-surface default is unchanged
byte-for-byte. That is the "mechanism, not taste" side of D-026, so D-058's
no-contrast-logic stance is refined (the engine still encodes no *opinion*), not
reversed.
**Consequences:** Black-on-dark headers, invisible same-hue eyebrows, and
black-on-dark pills/join-labels/stat-values are fixed for any fill and any variant;
the full existing golden suite passes unchanged (light cards byte-identical); a
parallel determinism guard asserts byte-identical output across worker counts. **A
documented limitation / follow-up:** leaf nodes (Stat, Bento label) do not receive
their *container* surface through `renderNode`, so a Stat inside a strongly-colored
card contrasts against the slide surface, not the card fill; threading the
container surface (or a `Stat.ValueColor` override) is additive future work. The
`TextMuted` bento/stat labels are left unchanged — a deliberate mid-gray legible on
both light and dark, not the reported bug. No public API change.

---

## D-083 — Container slide-bounds clamp (safe-area invariant)

**Date:** 2026-06-22
**Status:** Settled
**Context:** A Bento / Grid / Card handed a box whose bottom exceeds the slide's
printable area — because an over-full body stack (`VAlignTop`) placed its slot
low/tall, or a tall fixed-height container was requested — divides that box as
given and draws cells off the bottom edge, clipping them and overlapping the chrome
footer (recreation slides 6, 7). `bentoGeometry`/`layout.Grid` compute
`rowH = box.H/nRows` from whatever box they are handed, with no clamp to the
printable region. `VAlignFit` (R10.2/D-071) compresses such a stack, but it is
opt-in; the default top-anchored stack still overflows. R11.3
(`DECKARD-PRODUCT-REQUIREMENTS.md`, CRITICAL · engine).
**Decision:** Add `safeArea()` — a named alias of the chrome-aware `bodyRegion()`
(slide − content margins − the eyebrow/footer chrome bands) — and a
`clampToSafeArea(box, slideID)` guard called at the entry of `renderBento`,
`renderGrid`, and `renderCard`. When `box.Bottom() > safeArea().Bottom()` the box's
`H` is capped to `safeArea().Bottom() − box.Y` and a single warning
(`container overflow: content exceeds the slide safe area, clamped`) is logged; when
the box already fits (or its bottom is exactly the safe-area bottom) it is returned
unchanged. Pure integer cap → deterministic at any worker count.
**Why a cap, not a reflow:** the clamp is the deterministic *invariant* (nothing
draws below the safe area); reflowing an over-full stack to fit legibly is the
opt-in `VAlignFit` job. The two compose — `VAlignFit` makes content fit when asked,
the clamp guarantees it never draws off-canvas regardless — and keeping the clamp a
pure cap is what makes the default path byte-identical. Firing only on strict `>`
means fitting layouts, `VAlignFill` (which grows *to* the region bottom), and a
sole container handed the full body region (`Bottom() == safeArea.Bottom()`) are
all unchanged. Nesting does not double-warn: an outer clamp shrinks the box so the
containers it lays out get sub-boxes inside the clamped region and never
individually overflow → one warning, at the outermost container.
**Consequences:** Off-slide / footer-overlapping container content is fixed for any
dense bento or tall card grid; the full existing golden suite passes unchanged
(fitting content byte-identical); a parallel determinism guard asserts byte-
identical output across worker counts. The only user-observable change is an
additional `LayoutWarning` on overflow, surfaced through the existing
`Stats.Warnings`. No public API change.

---

## D-084 — R11.4 closed by D-053 + D-083 (content-region reserves chrome)

**Date:** 2026-06-22
**Status:** Settled
**Context:** R11.4 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine) requires the
body content region to reserve the section-eyebrow band (top) and the footer/
page-number band (bottom) when chrome is enabled, so no body content occupies the
chrome rows (recreation slides 6, 7 drew content over the footer).
**Decision:** Close R11.4 as **already implemented by D-053** (the `bodyRegion()`
shrink) **and made overflow-proof by D-083** (the safe-area clamp), with no renderer
change. `bodyRegion()` already adds `chromeEyebrowH + chromeBandGap` to the top inset
and `chromeFooterH + chromeBandGap` to the bottom inset — the chrome composer's own
constants (single source of truth with `chrome.go`) — and the body stack is laid out
inside it (`layout()`), so the region is disjoint from the bands by construction
(`chromeBandGap` exceeds the eyebrow's `chromeRuleH`). The recreation's footer
overlap was not a missing reservation but an *over-full stack* placing nodes below
the reserved region; D-083's `clampToSafeArea` (safe area = `bodyRegion()`) now caps
containers to the reserved region, and `VAlignFit` reflows over-full stacks, so the
reservation is honored under hostile content. The close is the named acceptance —
`render_chrome_region_test.go` — asserting (1) the chrome-on body region is disjoint
from both bands (recomputed from the chrome constants, not from `bodyRegion`, so a
drift would fail it); (2) chrome-off `bodyRegion` is the plain margin box
(byte-identical); (3) a container handed an overflowing box on a chromed slide is
clamped above the footer band (R11.4 × R11.3).
**Consequences:** R11.4 is closed; the chrome-overlap class is regression-guarded.
No public API change, no new token, no OOXML change. Pre-existing minor behavior
(noted, unchanged): `bodyRegion` reserves the top eyebrow band whenever chrome is
enabled even on a slide with no `Section` — a conservative over-reservation from
D-053.

---

## D-085 — Header-pill fit-to-label

**Date:** 2026-06-22
**Status:** Settled
**Context:** A card header pill is drawn at a fixed `pillW = In(1.0)`, and
`cardHeaderColumnWOf` reserves the same fixed width from the header text column. A
label wider than 1.0" (e.g. "CUSTOMIZABLE") wraps to two lines inside the rounded
chip and overflows it (recreation slide 5). R11.5
(`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine).
**Decision:** Extract `cardPillWidthOf(theme, pill, innerW) =
clamp(naturalWidth(pill @ TypeCaption) + 2·cardPillPadX, cardPillMinW, innerW)` and
call it from **both** `cardHeaderColumnWOf` (the header-width reservation) and
`renderCardChrome` (the drawn pill), so the reserved and drawn widths never drift.
`cardPillPadX = In(0.10)` per side (a pinned layout metric, not a token — it absorbs
the text frame's default inset so the measured label fits); `cardPillMinW = In(0.30)
= cardPillH` keeps a one-character pill a proper circular chip. The pill run gets
`FontScale = fitScale(naturalWidth, pillW − 2·cardPillPadX)` (the R10.5 primitive,
D-074): 0 — no shrink — when the pill is not clamped, a shrink-to-one-line when a
long label clamped the pill to `innerW`.
**Not byte-identical, by design:** every pill resizes from the fixed `In(1.0)` to its
fitted width (and the reserved header column shrinks with it). R11.5 explicitly does
not require byte-identity; determinism still holds (pure integer `naturalWidth`), and
the existing pill tests assert presence / shape counts, not the fixed width, so they
pass unchanged.
**Consequences:** Any caller-supplied pill label renders intact on one line inside
its chip, for any card width; the header title/eyebrow column reserves exactly the
fitted width so the wrapped-header geometry (R10.1) stays consistent. No public API
change (the `Card.HeaderPill` field's behavior improves). A parallel determinism
guard asserts byte-identical pill output across worker counts.

---

## D-086 — Card chrome anti-collision (status dot × header pill)

**Date:** 2026-06-22
**Status:** Settled
**Context:** A card's header pill (right-aligned, top of the header row) and its
status dot (top-right corner) are positioned independently — both right edges resolve
to `box.X + box.W − pad` and both sit at the top — so when both are set they overlap
(recreation slide 9: a "POPULAR" pill with the dot on its right edge). R11.6
(`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine).
**Decision:** In the status-dot block of `renderCardChrome`, when both `c.pill != ""`
and `c.statusDot != nil`, place the dot to the left of the pill:
`dotX = max(innerX, pillX − gapSM − cardStatusDotSz)` where
`pillX = innerX + innerW − cardPillWidthOf(theme, pill, innerW)`. The dot's right
edge is then `pillX − gapSM`, a gap to the left of the pill's left edge, so the two
boxes are disjoint by construction for any pill length (the dot derives from the same
`cardPillWidthOf` the pill is drawn with). When only one of the two is set the dot
keeps its corner placement (`box.X + box.W − pad − cardStatusDotSz`) → **byte-
identical** (the existing rich-visuals goldens set the dot without a pill, so they are
unchanged). The `innerX` floor keeps the dot on-card for a pathologically wide pill.
**Scope:** the pill × dot top-right pair only; the watermark (bottom-anchored, behind
the body) is a separate z-order/region concern for R11.11.
**Consequences:** A card with both a header pill and a status dot renders them
side-by-side without overlap, for any pill label. No public API change, no new token.

---

## D-087 — Join-badge fit-to-label

**Date:** 2026-06-22
**Status:** Settled
**Context:** The TwoColumn join badge is drawn at a fixed `joinBadgeSz = In(0.62)`
ellipse with a centered label; a label like "One agent" breaks mid-word into "One /
age / nt" inside it (recreation slide 8). R11.7 (`DECKARD-PRODUCT-REQUIREMENTS.md`,
HIGH · engine).
**Decision:** In the `JoinBadge` case of `renderColumnJoin`, grow the badge to
contain its label: `badgeSz = clamp(naturalWidth(label @ TypeBodySmall) +
2·joinBadgePadX, joinBadgeSz, joinBadgeMaxSz)` with `joinBadgePadX = In(0.12)` and a
pinned cap `joinBadgeMaxSz = In(1.5)`; a label that still does not fit at the cap is
shrunk to one line via `FontScale = fitScale(natW, badgeSz − 2·joinBadgePadX)` (the
R10.5/D-074 primitive). Keep the ellipse + centered label.
**Deviation from the spec's "clamp to the inter-column gap":** the badge deliberately
*overlaps* both columns (it straddles the seam; the actual inter-column gap is only
~`SpaceMD`), so clamping the diameter to that gap would collapse the badge. A pinned
`joinBadgeMaxSz` cap is used instead, with `fitScale` handling the overflow tail.
**Byte-identical for short labels:** "vs" measures `~0.2"`, so `needed < joinBadgeSz`
→ the base diameter is kept and `fitScale` returns 0 (no scale); the badge and run are
unchanged, and the existing column-join goldens (which use "vs") pass unchanged.
**Consequences:** Any short connector label renders intact and centered inside the
badge for any length; `JoinArrow` (no label) is unaffected. No public API change, no
new token; a parallel determinism guard asserts byte-identical badge output across
worker counts.

---

## D-088 — Stat-value overflow guard (role ladder)

**Date:** 2026-06-22
**Status:** Settled
**Context:** A Stat renders its value at a fixed `TypeDisplay` with no width
awareness; a wide value like "$4,000+" wraps to "$4,000 / +" and the stray line
pushes down and crowds the caption beneath it (recreation slide 9). R11.8
(`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine).
**Decision:** Add `statValueFit(autofit, value, boxW) (TypeRole, float64)`: a pinned
role ladder `[TypeDisplay, TypeH1, TypeH2]` — return the first role whose value fits
one line (`wrappedLines == 1`), else the `TypeH2` floor plus `fitScale(naturalWidthAt
(value, TypeH2), boxW)` to shrink it to one line. `renderStat` emits the value at the
returned role and `FontScale`. **Gated on `AutoFit`** (the R10.5/D-074 opt-in): when
AutoFit is off, or the value already fits at `TypeDisplay`, it returns
`(TypeDisplay, 0)` — byte-identical to the pre-R11.8 render. This refines R10.5's Stat
shrink: instead of scaling the font *within* `TypeDisplay`, step through real type
roles (40 → 32 → 28 pt) first, then sub-role scale only at the floor.
**Why gate on AutoFit:** an always-on ladder would change an AutoFit-off wide value,
but D-074 made shrink opt-in and a test pins "AutoFit-off Stat emits the full display
sz". Gating keeps AutoFit-off and AutoFit-on-fitting byte-identical, and applies the
ladder only to AutoFit-on wide values — the cases the caller asked to fit (the
product drives AutoFit per D-026). The existing `TestAutoFit_Stat_EmitsReducedSz` /
`_OffByteIdentical` / `_Deterministic` stay green.
**Deferred:** the optional value+label+delta stack-height clamp (the spec's "also
clamp") needs a `slideID` plumbed into `renderStat` to warn; the one-line value
guarantee already removes the reported caption-crowding (the wrapped value was the
cause). The `fitScale` 0.60 floor is a legibility bound — a sub-~0.85" box may still
overflow at the floor (accepted per the spec's "or a floor is reached").
**Consequences:** A wide Stat value renders on one line for any box width (above the
floor) when AutoFit is set, without crowding the caption. No public API change (the
existing `Stat.AutoFit` field now drives a role ladder for the value).

---

## D-089 — Bento row-label gutter fit-to-label

**Date:** 2026-06-22
**Status:** Settled
**Context:** The bento row-label gutter is a fixed `bentoGutterW = In(1.2)`;
"Control plane" wraps awkwardly to two lines in it and "The core" sits near the
footer (recreation slide 6) — the gutter width is unrelated to the actual label
widths. R11.9 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine).
**Decision:** Add `bentoGutterWidthOf(theme, v) = clamp(max over rows of
naturalWidth(label @ TypeCaption) + 2·bentoGutterPadX, bentoGutterMinW,
bentoGutterMaxW)` (0 when no row is labeled) with `bentoGutterMinW = In(0.8)`,
`bentoGutterMaxW = In(1.6)`, `bentoGutterPadX = In(0.1)`. Call it from **both**
`bentoColumns` (the drawn gutter) and the `preferredHeight` Bento case (the slot
estimate), so the layout and the estimate use the same gutter — closing the parity
the fixed constant left. `theme` is threaded into the `bentoColumns` / `bentoGeometry`
free functions (their white-box tests already build a `pptx.DefaultTheme()`; the
renderer passes `r.theme`).
**Not byte-identical, by design:** the gutter resizes for most label sets (a 1-char
label → `In(0.8)`, "Control plane" → its fitted width) and the unit column width
changes with it. R11.9 does not require byte-identity; determinism holds (pure integer
`naturalWidth`), and the existing bento tests assert gutter-presence, span ratios, and
equal-row heights — all gutter-width-independent — so they pass unchanged.
**Consequences:** A bento row label sizes its gutter to fit, for any label set; a
short label gets a tight gutter, a long one caps rather than starving the cells. No
public API change, no new token; the existing labeled-bento determinism tests cover
the rendered path.

---

## D-090 — Proportional list bullet hanging indent

**Date:** 2026-06-22
**Status:** Settled
**Context:** List items show a large fixed gap between the bullet and the label
("•      Chat & Q&A") (recreation slides 2, 4). R10.9/D-078 added the override
mechanism (`ParagraphOpts.BulletIndent`) and a tight preset (`IndentTight =
In(0.25)`), but the tight value is a *fixed* constant unrelated to the body size.
R11.10 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine) asks the hang indent to be
derived from the resolved body type size rather than a fixed wide value.
**Decision:** Make `listTightIndent` a theme-proportional renderer method:
`round(listTightIndentBase × bodySize / listTightIndentAnchorPt)` with
`listTightIndentBase = In(0.25)` and `listTightIndentAnchorPt = 14`. At the default
14 pt body it is exactly `In(0.25)` — **byte-identical** to the D-078 pinned value (the
existing `marL="228600"` test passes) — and a larger/smaller body scales the indent
linearly. `bulletIndent` becomes a method (it needs `r.theme` for the body size); its
only caller (`renderList`) is already a method. `IndentNormal` is unchanged (the
builder's 0.5" default, byte-identical).
**Acceptance relaxation:** R11.10's "≤ 1.5× the bullet glyph width" is an *example*
target; `In(0.25)` is ~2.6× a 14 pt glyph but tight relative to the 0.5" default the
recreation showed (the real "oversized" baseline). The close asserts the binding
requirement — proportional, and meaningfully tighter than the 0.5" default (≤ In(0.3))
— rather than the stricter 1.5×-glyph bar, which would break the D-078 byte-identity
for no real legibility gain. The wrapped-header interaction (the list start Y respects
the grown card header) already holds via R10.1/D-070 (Phase 49's golden).
**Consequences:** `List.Indent = IndentTight` scales with the deck's body size; the
default-theme output is unchanged. Deterministic (a pure function of the theme; the
result is an integer EMU). No public API change, no new token.

---

## D-091 — R11.11 closed by D-054 (watermark/decoration z-order-behind + low alpha)

**Date:** 2026-06-22
**Status:** Settled
**Context:** The D-054 card watermark is anchored inside the body box at low opacity;
R11.11 (`DECKARD-PRODUCT-REQUIREMENTS.md`, LOW · engine) asks that watermarks and
decorations never reduce body legibility under dense content.
**Decision:** Close R11.11 as **already implemented by D-054**, with no renderer
change. The watermark is emitted as the last chrome shape in `renderCardChrome`,
*before* the body content `renderCard` draws — so it is **behind** the body in z-order
— at `cardWatermarkAlpha = 13000` (~13% opacity); background decorations
(`LayerBackground`) are likewise emitted before the body in `layout()`. R11.11's
acceptance is an explicit **OR** — the watermark "occupies only the residual empty
region **OR** is drawn behind content at a legible alpha" — and the engine already
takes the second branch (opaque body text paints on top of a 13% ghost). The
**optional** residual-region restriction is intentionally **not** adopted: it would
couple the chrome to the body's wrapped-line estimate and change the watermark
position for a LOW cosmetic gain, where z-order + low alpha already guarantees
legibility. The close is the acceptance test (`render_watermark_zorder_test.go`): the
watermark text is emitted before the body text (z-order behind), carries the low ~13%
alpha, and is inert (no `<a:alpha>` run) when unset.
**Consequences:** R11.11 is closed; the watermark/decoration legibility invariant is
regression-guarded. A caller wanting the watermark confined to empty space can place
it as a foreground decoration instead; the engine's default (behind, low alpha) is the
safe one. No public API change, no new token.

---

## D-092 — Adversarial content-fit harness + leaf safe-area clamp

**Date:** 2026-06-22
**Status:** Settled
**Context:** The recreation's overlaps reproduced only under real, long, or dark
content; the existing tests pass because their fixtures use short, light, single-line
content. There was no torture suite proving each component renders correctly under
hostile content. R11.12 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · both — engine side
per D-059) asks for a reusable acceptance harness asserting the structural invariants.
**Decision:** Ship an adversarial harness (`render_adversarial_test.go` +
`render_adversarial_invariants_test.go`): an `adversarialScene()` exercising every
component × {light, dark} under hostile content, asserting (1) header band ≤ body top
(`cardHeaderBottom` ≤ `renderCardChrome` body Y), (2) **every emitted box on the
canvas** — parsed directly from each slide's `<a:off>`/`<a:ext>` pairs, no test-only
recorder needed (so it cannot perturb byte output), (3) fit-required chrome text on
one line (`cardPillWidthOf`/`statValueFit`/`fitScale`), (4) chrome contrast ≥ 4.5:1
(`onCardSurface`/`contrastRatioT10`), plus a worker-count determinism guard.
**The suite surfaced a real bug, fixed in this PR (§17).** A `Grid` of cards whose
bodies held a tall hostile `List` placed a body *leaf* below the card and off the slide
canvas — the R11.3 clamp clamps the *container* box but not a leaf an over-full card
body pushes past it. **Fix:** generalize the R11.3 safe-area clamp from the three
container composers to the single `renderNode` dispatch point, clamping **every
content node's** box, while exempting the full-slide overlays `Decoration` (which may
bleed off-canvas by design) and `SectionDivider`. This subsumes the three per-container
clamps (removed — one clamp point, one warning) and additionally caps an over-full card
body / stack leaf. Byte-identical when the box already fits (the clamp is a no-op, so
the full existing golden suite passes); pure integer → deterministic; the Phase-51
clamp tests still pass.
**Density vs bounds (D-026).** The clamp *caps the box* (the off-canvas invariant) and
*warns* (the overflow `LayoutWarning`); legibly *compressing* an over-full card body
remains the opt-in `VAlignFit` / `Card.BodyVAlign` path — the product drives density.
**Consequences:** Every component is regression-guarded against the fixed-size class
under hostile content, in CI (`make preflight`); a leaf can no longer draw off-slide.
No public API change, no new token.

---

## D-093 — Wave 11 §17 checkpoint: header/pill reservation fix + doc/test backfill

**Date:** 2026-06-22
**Status:** Settled
**Context:** The Wave 11 §17 adversarial checkpoint (a 38-agent workflow: 8 dimension
finders → 2 skeptics/finding → completeness critic → synthesis over the R11
component-robustness cluster, Phases 49–60 / D-081..D-092) returned a **clean bill of
health** on the binding invariants — pure-integer determinism (the `srgbLinear` table
is built once at init, every per-call use is an integer lookup), byte-identity on the
default-theme path (every auto-contrast / fit-to-label helper is a no-op there), and
the D-092 bounds clamp generalized correctly with the right overlay exemptions. It
surfaced **one real correctness defect** plus doc/test-hygiene gaps.
**Decision (fixes landed in this `chore(checkpoint)` PR):**
- **H1 — header/pill reservation (real defect, fixed).** `cardHeaderColumnWOf` and
  `renderCardChrome` reserved the pill width *conditionally*
  (`if headerW > pillW+gapSM`), so when a pill label clamped to the whole inner width
  (`pillW == innerW`) the header column stayed at full width and the title overlapped
  the pill. Made the reservation **unconditional** (`headerW -= pillW + gapSM`, floored
  at 0) in both — byte-identical on every non-degenerate deck (where
  `headerW > pillW+gapSM`), collapsing the header column to 0 only in the pathological
  full-width-pill case. Guarded by `TestCardHeaderColumn_PillReservation`.
- **M2 — stale safe-area docs (§19).** `docs/glossary.md` and `docs/site/guide/scene.md`
  said only "a container (Bento/Grid/Card)" is clamped; updated to "every content
  node … (full-slide overlays exempt)" per D-092.
- **M3 — `fitScale` docstring** amended to state the 0.60 floor is a legibility bound
  that does not guarantee fit (extreme over-wide text may still overflow; see D-088).
- **M4 / L1 / L2 / L3 — test backfill** of shipped behavior the per-phase tests missed:
  pill and join-badge auto-contrast colours, the narrow-card status-dot clamp, a
  nested-container overflow warn + on-canvas check, and CardSection header
  auto-contrast.
- Removed an untracked `test_check.go` debris file (a stray `package scene_test` file at
  the repo root that broke root package detection).
**Documented as intentional (no code change):**
- **H2** (status dot shares the pill's left edge) shares H1's root — a single pill
  label wider than the whole card — a pathological input; the degenerate boundary is
  noted, not contorted around.
- **H3** (the "Bento `preferredHeight` uses `estGap` → D-089 violation") is a **false
  positive**: D-089's parity is about the *row-label gutter*, which `preferredHeight`
  already threads through `bentoGutterWidthOf`; the flagged `estGap` is the
  *inter-column* gap, a conservative estimator constant (shared by the Grid/TwoColumn/
  Card cases) that only ever over-estimates height (safe).
- **M1** (a light `ColorAccent` join-badge label on a dark-variant theme can be
  low-contrast) is the **documented asymmetry** of the contrast *mechanism*
  (D-058/D-026: `onCardSurface` flips dark→light only, assuming the theme default text
  is dark): pre-existing (the prior hardcoded `TextInverse` was equally affected), and
  a caller-`Color`-override case. The default theme (dark accent) is byte-identical and
  correct.
**Consequences:** Wave 11 is closed as healthy. One geometry defect fixed
byte-identically; docs and tests reconciled with the shipped behavior. No public API
change, no new token.

---

## D-094 — prim-cta-button: a Button leaf node (R12.1)

**Date:** 2026-06-23
**Status:** Settled
**Context:** A professional sales/investor deck ends on a button (a "Talk to the team
→" closing CTA) and prices against one (a "Start free" at the foot of every pricing
card). The scene IR had no button primitive at all — a closing slide stopped at bare
prose, a pricing card at the price. R12.1 (CRITICAL · both; D-059 engine side) is the
first Wave-12 component primitive.
**Decision:** Add `KindButton` + a `Button` leaf node `{Label string; Tone ButtonTone
(ButtonPrimary/ButtonAccentAlt/ButtonGhost/ButtonNeutral); Size ButtonSize
(ButtonMD/SM/LG); LeadingIcon, TrailingIcon string; Align HAlign}`. It renders native
chrome (`scene/render_button.go`):
- A `RadiusFull` `ShapeRoundRect` pill, **content-fit** width
  `naturalWidth(label@TypeBody) + each icon(iconSz+gap) + 2·padX`, floored at the pill
  height (a circular minimum) and clamped to its box, with a `fitScale` tail so a label
  clamped to the box stays one line.
- A **tone→token** fill (P2): Primary = `ColorAccent` solid / `TextInverse` label;
  AccentAlt = `ColorAccentAlt` / `TextInverse`; Neutral = `ColorSurfaceAlt` /
  `TextPrimary`; Ghost = `NoFill` + an accent hairline (`WithLine`) / `TextAccent`.
- A middle-anchored centered bold `TypeBody` label flanked by optional native custGeom
  icons (the closed-name registry via `ps.AddIcon`, filled with the label color).
- `Align` offsets the pill within its box (the same center/right offset path the body
  stack already uses); inside a container cell it inherits `HAlignLeft`.

The `ButtonSize` SM/MD/LG height/padding/icon scale is a **pinned layout metric**
(`buttonMetrics`), not a theme token — it sizes geometry, not a visual property; the
*tone colors* are tokens (THEME.md). It is **presentational only**: no hyperlink/action
wiring (the deck is static), so it adds no builder capability (P1) and registers no
media (`nodeUsesAssets` false → parallel-safe). Full new-node wiring: policy (native),
validate (non-empty label), `renderNode` dispatch + `preferredHeight` (size height) +
`isFlexible` false, `walkIconRefs` `case Button` (both icon fields Stage-1 validated),
catalog count 22 → 23, integration kind-range loop → `KindButton`.
**Consequences:** A deck that uses no `Button` is byte-identical (the node is purely
additive). A rendered button is **not** byte-identical to any prior output — there was
none; the byte-identity obligation is only the unused/default path (consistent with the
fit-to-label note in D-089/D-093). No `Disposition`/mode toggle, no default-tone opinion
(the soul-tokened default tone is Deckard's product side, D-026). Brief 44.

---

## D-095 — prim-in-card-checklist-fill: a Checklist leaf node (R12.2)

**Date:** 2026-06-23
**Status:** Settled
**Context:** The "what you get" feature list is the heart of every offer/pricing card.
The recreation rendered `List{Kind: ListChecklist}` via `pptx.BulletCheckbox` — an
**empty white square** (the `ListItem.Checked` bool is never read), a broken bullet
indent, native body size, no column reflow, and no fill-to-card. R12.2 (CRITICAL ·
engine) adds a dedicated `Checklist` node rather than mutating `List`, so the existing
list path stays byte-identical.
**Decision:** Add `KindChecklist` + a `Checklist` leaf node `{Items []ChecklistItem{Text
RichText; State CheckState (CheckDone/CheckNo/CheckNeutral); Icon string}; Columns int;
GlyphTone *ColorRole; Fill bool}`, rendered in `scene/render_checklist.go`:
- **A true filled glyph, not a font checkbox.** `CheckDone` → the curated `check` SVG,
  `CheckNo` → `x`, `CheckNeutral` → `dot`, each via `ps.AddIcon` as a native custGeom
  shape with a token fill (a per-item `Icon` overrides the name). Fixes the empty-square
  bug by construction; reuses the Phase-61 icon-glyph mechanism.
- **Glyph color: per-state default + `*ColorRole` override.** `CheckDone` defaults to
  `ColorAccent`, the rest to `TextMuted`; `GlyphTone` (non-nil) overrides all. It is a
  `*ColorRole` because `ColorRole`'s zero value is a real color (`ColorCanvas`) — the
  D-054 pattern, a §4.3 deviation from the spec's value `ColorRole`.
- **Hanging indent = glyph width + gap.** Each row is `[glyph | text]`; the text frame
  is offset by `glyphSz + glyphGap`, so wrapped lines align under the text (no PPTX
  auto-bullet).
- **Row-major column reflow.** `cols = clamp(Columns,1,3)`; item `i` → `(row=i/cols,
  col=i%cols)`; columns share `box.W` with a pinned gap. Per-row heights are
  content-aware (`wrappedLines` × per-line height).
- **Fill distributes inter-row slack** across the `rows−1` gaps so the last row meets
  the box bottom (the VAlignJustify primitive, per-row); off (default) top-aligns at a
  pinned gap. `Checklist` is added to `isFlexible` so a `VAlignFill`/`BodyVAlign` parent
  can grow it to fill a card.

Layout metrics (glyph size/gap, column gap, row gap, line height) are pinned EMU (not
tokens); the glyph colors are tokens (THEME.md). Full new-node wiring: policy (native),
validate (items + columns 0..3 + state range), `renderNode` dispatch +
`preferredHeight` (content-aware) + `isFlexible` true + `nodeUsesAssets` false,
`walkIconRefs case Checklist` (per-item icon overrides Stage-1 validated), catalog 23 →
24, integration kind-range loop → `KindChecklist`.
**Consequences:** A deck with no `Checklist` is byte-identical (additive; the
`List`/`BulletCheckbox` path is untouched). A rendered checklist is not byte-identical
to anything — there was no equivalent. No mode toggle; the engine renders `Text`
verbatim and picks no content (D-026). Brief 45.

---

## D-096 — prim-chip-row-group: a ChipRow leaf node (R12.5)

**Date:** 2026-06-23
**Status:** Settled
**Context:** A horizontal row of tag/category chips (a labeled "COMMON BUILDS · Finance ·
HR · …" strip, a card-footer capability row) is a universal slide element. The IR had
only a single inline `Chip`; the recreation rendered chip rows as broken bullet lists or
dropped them. R12.5 (HIGH · both, engine side per D-059) adds the row primitive.
**Decision:** Add `KindChipRow` + a `ChipRow` leaf node `{Label string; Chips
[]ChipSpec{Label string; Tone ChipTone; Color ColorRole; Icon string}; Wrap bool; Align
HAlign}`, rendered in `scene/render_chiprow.go`:
- **Greedy left-to-right wrap.** Each chip is content-fit (`chipWidthOf =
  naturalWidth(label@TypeBodySmall) + 2·pad`, + a leading icon). A shared two-pass packer
  (`chipRowLines`) feeds both the renderer and `preferredHeight`; chips pack onto a line
  until the next exceeds `box.W`, then break — pure integer arithmetic, deterministic.
- **The leading label rides line 0** as a `TypeCaption` run, consuming its width before
  the first chip and participating in line 0's `HAlign` offset.
- **`Wrap` is the engine mechanism (zero = single line).** A plain Go bool can't encode
  "default true"; the engine zero is the minimal behavior, the product sets `Wrap: true`
  for a reflowing strip (D-026). A `Wrap: false` row that overflows is the caller's
  explicit choice (the `Decoration.Bleed` posture).
- **Per-line `HAlign` offset** (left / center / right within `box.W`); each chip is a
  `RadiusFull` rounded-rect with the `ChipTone` fill (reusing the single-`Chip`
  treatment), an optional leading custGeom icon, and a centered `TypeBodySmall` label
  auto-contrasted on a solid fill (`onCardSurface`).

Layout metrics (chip height, padding, icon size, gaps) are pinned EMU; the tone colors
reuse the existing `ChipTone` → token mapping (no new token). Full new-node wiring:
policy (native), validate (non-empty chips + tone range), `renderNode` dispatch +
`preferredHeight` (line count × chip height) + `isFlexible` false + `nodeUsesAssets`
false + `nodeEffectiveHAlign`, `walkIconRefs case ChipRow` (per-chip icons Stage-1
validated), catalog 24 → 25, integration kind-range loop → `KindChipRow`.
**Consequences:** A deck with no `ChipRow` is byte-identical (additive). The chip pill
uses `RadiusFull` (a fuller capsule than the single `Chip`'s default-radius rounded
rect) — a deliberate, isolated visual for the new node. No mode toggle; the engine picks
no content (D-026). Brief 46.

---

## D-097 — prim-callout-banner: a Banner node (R12.6)

**Date:** 2026-06-23
**Status:** Settled
**Context:** A full-width filled "big takeaway / promo / CTA" strip (a lime closing
banner, a dark "$0 · Start free →" promo) is a staple slide element. The IR's `Callout`
is only a small left-bar note; the recreation rendered the banner as plain overlapping
text with no fill. R12.6 (HIGH · engine) adds the wide banner primitive.
**Decision:** Add `KindBanner` + a `Banner` node `{Lead RichText; Body RichText; Icon
string; Fill ColorRole; TextColor TextColorRole; Trailing []SlideNode}`, rendered in
`scene/render_banner.go`: a full-width `RadiusLG` filled strip, a left region with the
leading custGeom icon + bold lead + body, and an optional right region stacking the
`Trailing` children (the card-body `stackIn` + `renderNode` mechanism).
- **Fill defaults to accent.** `Fill`'s zero value (`ColorCanvas`) is treated as
  `ColorAccent` — a banner is always a filled strip; a canvas one would be invisible (a
  §4.3 deviation from the spec's value field, documented).
- **Text auto-contrasts by default.** When `TextColor` is the zero value (`TextPrimary`)
  the lead/body resolve via `onCardSurface(fill)` (inverse on a dark fill, the default on
  light); an explicit non-default `TextColor` is honored verbatim. The banner runs force
  this color (and bold on the lead) so the strip is legible out of the box.
- **A node with children.** `Trailing` makes `Banner` recurse like a container in
  `validateChildren` / `walkIconRefs` / `walkImages` / `nodeUsesAssets` and the
  integration `collectKinds`; it carries no `AssetID` itself (`nodeUsesAssets` defers to
  `Trailing`, so a typical Stat/Button banner stays parallel-safe).

Layout metrics (padding, icon size, gaps, the trailing-region width band) are pinned EMU;
the fill and text colors are tokens (THEME.md). Full new-node wiring: policy (native),
validate (recurse `Trailing`), `renderNode` dispatch + `preferredHeight` (max of text and
trailing stacks + padding, floored) + `isFlexible` false + `nodeUsesAssets`(Trailing),
catalog 25 → 26, integration kind-range loop → `KindBanner`. Distinct from the side-bar
`Callout`.
**Consequences:** A deck with no `Banner` is byte-identical (additive). No mode toggle;
the engine picks no content (D-026). Brief 47.

---

## D-098 — prim-ribbon-corner-badge: a Card.Ribbon field (R12.3)

**Date:** 2026-06-23
**Status:** Settled
**Context:** To single one card out of a row ("MOST POPULAR" across a tier's top, a star
on a highlighted feature), a deck needs a pinned badge OUTSIDE the header text flow. The
only tool was `Card.HeaderPill`, an in-row pill that jammed into the title row in the
recreation. R12.3 (HIGH · engine) adds a ribbon distinct from the pill.
**Decision:** Add `Card.Ribbon *Ribbon{Text string; Position RibbonPos
(RibbonTopBar/RibbonCornerTL/RibbonCornerTR/RibbonCornerStar); Color *ColorRole; TextColor
TextColorRole}` — a **field extension, not a new node** (no catalog/kind change). Drawn
last in `renderCardChrome` (on top):
- **RibbonTopBar reserves a band; the body shifts down.** `ribbonReserveOf(c)` returns the
  band height for a top bar (0 for corner positions), threaded through `cardHeaderBottom`,
  `renderCardChrome` (the header `top := box.Y + ribbonReserveOf(c)`), and
  `cardHeaderExtraHeight` (the slot estimate) — so the reserved band, the header text, the
  D-054 band, the body Y, and `preferredHeight` all agree. A top bar therefore never
  overlaps the eyebrow/title.
- **RibbonCornerStar** renders the curated `star` custGeom in the top-right corner.
- **Color** is `Color *ColorRole` (nil = `ColorAccent`, the D-054 pointer pattern);
  `TextColor` auto-contrasts against the fill by default (`onCardSurface`), explicit
  values honored. Band/tab/star metrics are pinned EMU.
**Deviation (§4.3):** the spec's *diagonal rotated-rect corner ribbon with a label* is
not expressible in V1 — the builder has no rotated-text primitive. `RibbonCornerTL/TR`
render a **horizontal content-fit corner text tab** instead (the label is the point); the
rotated-diagonal-band visual waits on a builder text-rotation enhancement. All positions
pass the acceptance (distinct, in-corner / top, no header overlap).
**Consequences:** A card with no `Ribbon` is byte-identical (nil → `ribbonReserveOf` 0 →
the pre-feature geometry; the existing card goldens pass unchanged). No mode toggle; the
soul decides when to flag a card (D-026). Brief 48.

---

## D-099 — prim-inter-column-connectors: Grid.Connectors (R12.4)

**Date:** 2026-06-23
**Status:** Settled
**Context:** A 3-column architecture grid reads as data flow only when connector arrows
join the columns in the gutters. The recreation's plain `Grid` left the cards floating
disconnected; `TwoColumn.Join` (D-055) only places a single centered seam element. R12.4
(HIGH · engine) adds an N-column gutter connector layer.
**Decision:** Add `Grid.Connectors []GridConnector{Between [2]int; Kind ConnectorKind;
Label string}` — a **field extension, not a new node** (no catalog/kind change) — and a
new `ConnectorBiArrow` glyph. `renderGrid` calls `renderGridConnectors`, which for each
connector `{c, c+1}` derives the gutter box from the deterministic `layout.Grid` cell
boxes (`{X: cells[c].Right(), W: cells[c+1].X − cells[c].Right(), Y: box.Y, H: box.H}`)
and calls the existing `render_flow` `renderConnector` (reuse, D-044). `ConnectorBiArrow`
emits `leftRightArrow` (horizontal) / `upDownArrow` (vertical) preset geometry — a `prst`
attribute on the already-registered `prstGeom` element, so no `restorenamespaces` change.
An optional `Label` sits in the lower third of the gutter (a muted `TypeCaption`).
Stage-1 validates adjacency (`Between[1] == Between[0]+1`), range (`0..Columns−1`), and
kind.
**Consequences:** An empty `Connectors` slice is byte-identical (the helper returns
immediately; the existing grid tests pass unchanged). Gutters are narrow, so connector
labels are best kept short. The complementary single-seam case stays `TwoColumn.Join`
(R12.8 extends that for a spanning bridge). Brief 49.

---

## D-100 — prim-icon-label-rows + read-side entity-escaping fix (R12.7)

**Date:** 2026-06-23
**Status:** Settled
**Context:** A card that is a vertical stack of `[icon | label | optional meta]` rows
(integrations, capabilities) reads as designed rows, not bullets. The recreation rendered
these as bullet lists with the title overlapping. R12.7 (MED · engine) adds the row
primitive. Implementing its integration round-trip fixture (a label with `&`) **surfaced a
pre-existing codec bug** fixed here per `CLAUDE.md §17`.
**Decision (the node):** Add `KindIconRows` + an `IconRows` leaf node `{Rows []IconRow{Icon
string; Label RichText; Meta RichText; Tone RowTone}; Fill bool; GlyphColor ColorRole}`,
rendered in `scene/render_iconrows.go` mirroring the Phase-62 checklist row engine:
content-aware per-row heights, a `[icon | label | right-aligned meta]` layout, an optional
`RowPill` `RadiusMD` `SurfaceAlt` frame, and a `Fill` mode distributing inter-row slack
(added to `isFlexible` so a `VAlignFill` card grows it). `GlyphColor`'s zero value
(`ColorCanvas`) maps to `ColorAccent` (a canvas-colored glyph is invisible). Pinned layout
metrics; glyph color + pill surface are tokens. Per-row icon validated via `walkIconRefs`.
Catalog 26 → 27. Additive ⇒ byte-identical when unused.
**Decision (the codec fix):** `internal/ooxml.StripNamespacePrefixes` (the read-side
prefix-stripper that rebuilds slide XML token-by-token before `xml.Unmarshal`) re-emitted
decoded `CharData` and attribute values **raw** (`buf.Write(v)` / `buf.WriteString(attr.
Value)`), so any run text or attribute containing `&` / `<` / `>` (a label "A & B", a URL
`a=1&b=2`) became a bare entity and the reopen failed with "invalid character entity & (no
semicolon)" — silently dropping that slide on read (a G6 round-trip defect, latent because
no prior fixture used those characters in slide-body text). Fixed by re-escaping both with
`xml.EscapeText`. Hyperlink URLs were unaffected (they live in the part rels, not the slide
XML). Guarded by `TestStripNamespacePrefixes_EscapesEntities` and the everyNodeScene
round-trip (which now carries an `&` label).
**Consequences:** Any self-authored deck with `&`/`<`/`>` in run text now round-trips
losslessly (previously such a slide was dropped on reopen). The write path was already
correct (`xml.Marshal` escaped); only the read-side rebuild was wrong. Brief 50.

---

## D-101 — prim-spanning-column-bridge: TwoColumn.JoinPosition (R12.8)

**Date:** 2026-06-23
**Status:** Settled
**Context:** An option / path slide ("One agent, purpose-built — two ways to get it") wants
a labeled connector spanning the *tops* of two columns. `TwoColumn.Join` (D-055) only
places an element on the vertical seam; the recreation collapsed the bridge into a tiny
seam circle with "One age nt" wrapped mid-word. R12.8 (MED · engine) adds a spanning bridge.
**Decision:** Add `TwoColumn.JoinPosition JoinPosition` (`JoinSeam`/`JoinTopBridge`/
`JoinBottomBridge`) — a **field extension, not a new node**. `JoinSeam` (zero) keeps the
D-055 centered-seam element (byte-identical). A bridge reserves a `bridgeBandH` band at the
top (or bottom) edge — the ribbon band-reserve pattern (D-098) — so the columns lay out in
the inset region and the bracket spans above (below) them; `preferredHeight` adds the band.
`renderColumnBridge` draws an accent spanning line from the left column's left edge to the
right column's right edge, two short end stubs reaching toward the columns, and the
`JoinLabel` as a content-fit `RadiusFull` accent pill centered on the line — sized to the
label so it never wraps mid-word (a `fitScale` tail shrinks it only if it would exceed the
span). Accent colors are tokens; the band height, stub length, stroke, and pill pad are
pinned EMU. Stage-1 validates the position range.
**Consequences:** A `JoinSeam` (default) `TwoColumn` is byte-identical (the existing
column-join tests pass unchanged). The N-column gutter-connector case stays `Grid.Connectors`
(D-099); this is the 2-column spanning-bridge case. Brief 51.

---

## D-102 — prim-attribution-lockup: a Lockup node (R12.9)

**Date:** 2026-06-23
**Status:** Settled
**Context:** A branded deck places a "POWERED BY [logo] CLEAR TECH" / "in partnership with"
lockup — a caption paired with a small partner logo as one inline, centerable unit — on its
cover and closing. The recreation dropped the logo and rendered the caption as plain text.
`Image` and `Chip` exist separately but nothing composes a small logo with a caption.
**Decision:** Add `KindLockup` + a `Lockup` leaf node `{Caption string; AssetID AssetID;
Icon string; AssetSide AssetSide; MaxHeight pptx.EMU; Align HAlign}`, rendered in
`scene/render_lockup.go` as a centered inline group: `[caption | gap | logo]` (`LeadCaption`,
default) or `[logo | gap | caption]` (`TrailCaption`), the whole group aligned within the box
and vertically centered. The mark is **exactly one** of `AssetID` (a partner logo resolved
via the AssetResolver → a pic, warn-don't-fail) or `Icon` (a curated glyph, media-free);
validation rejects neither/both. The logo box is height-bounded by `MaxHeight` (a pinned
default when 0) and **square** — without pixel dimensions the engine cannot know the aspect
(§7), so a square box is the honest default (the caller's logo bytes drive the display
aspect). `nodeUsesAssets` is `AssetID != ""` so an asset lockup composes serially
(deterministic media part numbering, RFC §10.1) and an icon lockup stays parallel-safe.
`policy = {HasAsset: true}` (it carries an `AssetID` field — the policy_test invariant) with
`Image:false`, the conditional-asset shape of `Decoration`. The caption is `TypeCaption`
muted; the gap/default-height/pad are pinned EMU; the icon flows through `walkIconRefs`.
Catalog 27 → 28.
**Consequences:** A deck with no `Lockup` is byte-identical (additive). This is the last
Wave-12 component primitive; with the atoms (Button, Checklist, ChipRow, Banner, Ribbon,
Grid connectors, IconRows, column bridge, Lockup) in place, **R12.10** (the pricing-offer
card recipe) is a product-layer composition on Deckard's contract/skill (D-059) — no further
engine node. Brief 52.

---

## D-103 — Wave 12 §17 checkpoint: THEME.md backfill + documented-intentional codec/IR rationales

**Date:** 2026-06-23
**Status:** Settled
**Context:** The Wave 12 §17 adversarial checkpoint (a 22-agent workflow: 8 dimension finders
→ 2 skeptics/finding → completeness critic → synthesis over the R12 component-primitive
cluster, Phases 61–69 / D-094..D-102) returned a **clean bill of health** on the binding
invariants — every new node is wired through the full checklist, output is token-driven and
deterministic, byte-identity holds at the unused/zero-value path, and the Phase-67 codec fix
is correct and minimal. It surfaced **one genuine §19 documentation-sync defect** plus two
refuted findings to record as by-design.
**Decision (fixes landed in this `chore(checkpoint)` PR):**
- **§19 / P2 — three Wave-12 token-resolving nodes lacked a `docs/design/THEME.md` mechanism
  section.** `THEME.md` documented Button (D-094), Checklist (D-095), Banner (D-097), Ribbon
  (D-098), IconRows (D-100), and Lockup (D-102), but **ChipRow (D-096)**, **Grid.Connectors
  (D-099)**, and **TwoColumn.JoinPosition / column-bridge (D-101)** — all of which resolve
  colors through tokens — had no section, despite §19/§20 requiring a taxonomy entry in the
  same PR as a visual property. (The finders flagged two; verification surfaced the third,
  ChipRow.) Added three "(mechanism, no new token — D-0NN)" sections naming the exact token
  roles (`ChipTone`→`TokenColor`/`ColorSurfaceAlt`/`TextMuted`; connectors→`ColorAccent`/
  `TextMuted`; bridge→`ColorAccent` + `onCardSurface`). No code change.
- **Banner docstring** reworded from "(Stat/Button)" to "(e.g. Stat/Button/Lockup)" to remove
  the implied type constraint (see below).
- **Composite round-trip coverage:** added a `Lockup` inside `Banner.Trailing` to the
  `everyNodeScene` integration fixture, exercising the `walkIconRefs` recursion through
  `Banner.Trailing` into a leaf.
**Documented as intentional (no code change):**
- **Comments are written raw, correctly.** A finding claimed the `xml.Comment` cases in
  `StripNamespacePrefixes` / `RestoreNamespaces` should mirror the Phase-67 `xml.EscapeText`
  calls. **They must not:** Go's `xml.Decoder` entity-decodes `CharData` and attribute values
  but hands back comment bytes **verbatim**, so escaping them would double-escape
  (`a&amp;b` → `a&amp;amp;b`) and corrupt round-trip fidelity. The Phase-67 scope (CharData +
  attributes only) is complete and correct; this asymmetry is intentional.
- **`Banner.Trailing` accepts any `SlideNode` by design** (validated via the generic
  `validateChildren`). Restricting trailing node types would encode a layout opinion (D-026);
  the docstring's "Stat/Button" is illustrative, not normative.
**Consequences:** Wave 12 is closed as **healthy**. One §19 doc defect fixed; two rationales
recorded so a future reviewer does not "fix" the correct comment-raw-write or add a spurious
`Banner.Trailing` type restriction. No production code change was required for correctness.
The R12 component primitives (28-kind catalog) are complete.

---

## D-104 — `ColorPaper` tinted-paper canvas token (Wave 13 / Phase 70, R13.1 engine half)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.1 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · both) wants an
off-white "paper" canvas (≈ `#FAFAF8`) distinct from pure white — pro reference
decks never use flat `#FFFFFF` for content. The engine already resolves any
`ColorRole` for a `BackgroundColor`; the gap is that there is no dedicated
canvas-tint token separate from white. D-059 puts the engine half here; the soul
emitting a paper tint and the composer auto-applying it on light slides is
Deckard's product half.

**Decision:** Add a `ColorPaper` surface role, appended **last** to the
`ColorRole` iota (after `ColorInfo`) so every prior value is unchanged, plus a
`WithPaper(RGB)` theme option. It defaults to `FFFFFF` (= `ColorCanvas`) in
`DefaultTheme`, so a `Background{BackgroundColor, ColorPaper}` slide is
byte-identical to a `ColorCanvas` one until a theme overrides the tint. No new
`BackgroundKind`, no OOXML element, no `restorenamespaces` change — a pure
semantic-token addition; `BackgroundColor` already paints it.

**theme1.xml round-trip:** PowerPoint's theme has exactly 12 OOXML slots
(`dk1/lt1/dk2/lt2/accent1..6/hlink/folHlink`), all already claimed by the 10
surface + 2 text roles in `themecodec.go`'s `writeSlots`. There is no spare slot,
so `ColorPaper` is a **role without a slot** — like `TextMuted` it keeps its
in-memory default on read-back. This is correct: the soul/caller owns the paper
tint at author time (D-026); it is not recovered from a re-opened deck's
theme1.xml. The G6 guarantee is about emitted output, and it holds — a
`ColorPaper` background resolves to a literal `srgbClr` in the full-slide rect's
`solidFill`, which round-trips losslessly through `pptx.Open`.

**Consequences:** Foundational for Wave 13; later background work (multi-stop /
radial / mesh gradients, surface fills) composes the same surface palette.
`grep` confirms no `[N]ColorRole` arrays or range-over-fixed-roles loops, so the
appended role breaks nothing. Tested: default == canvas, `WithPaper` + `Clone`
carry the tint, an off-white background's RGB survives write → reopen → re-write,
and the default-theme case is byte-identical to `ColorCanvas`.

---

## D-105 — Multi-stop background gradient (`Background.Stops`) (Wave 13 / Phase 71, R13.3)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.3 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine) wants a
background gradient with an arbitrary number of color stops at caller-chosen
positions (a 3–4-hue hero wash). The scene `Background` only exposed a fixed
`Gradient [2]pptx.ColorRole` two-stop linear fill, even though
`pptx.LinearGradient` already accepts variadic `GradientStop`s — only the scene
struct capped it at two.

**Decision:** Add a scene `GradientStop{Pos float64; Color pptx.ColorRole}` type
and a `Background.Stops []GradientStop` field. When `Stops` is non-empty it
supersedes the legacy `Gradient` pair; `renderBackground` validates it via
`backgroundGradientStops` (2..8 stops, each `Pos ∈ [0,1]`, strictly ascending) and
feeds it to `pptx.LinearGradient(angle, stops...)`. When `Stops` is empty the
legacy two-`TokenColor` path runs unchanged — **byte-identical** to pre-D-105
output. Invalid stops record exactly one `LayoutWarning` and skip the fill (RFC
§10.2, D-026 — no panic), consistent with the existing background-asset warning;
backgrounds are not validated in Stage 1/2, so this is a render-time check.

**Consequences:** A scene-side field extension only — no builder change (P1, no
new OOXML capability), no new `BackgroundKind`, no OOXML element, no
`restorenamespaces` change (the legacy path already emits `<a:gradFill>`/`<a:gs>`).
The slice field makes `Background` (and thus `SceneSlide`) **non-comparable** with
`==`; tests use byte-comparison / `reflect.DeepEqual` — `grep` confirms no `==`
on `Background` in non-test code. The mapping is pure integer-EMU/alpha through
the existing fill path, so output is deterministic regardless of worker count.
Foundation for R13.2 radial backgrounds (Phase 72), which reuse the same `Stops`
list with a new `BackgroundRadial` kind feeding `pptx.RadialGradient`. Tested:
3-stop emits 3 `<a:gs>` + round-trips; `<2`/`>8`/out-of-range/descending warn +
skip; legacy 2-role byte-identical; multi-stop deterministic.

---

## D-106 — Radial-vignette background (`BackgroundRadial`) (Wave 13 / Phase 72, R13.2)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.2 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine) wants a
center-out radial background (spotlight/vignette) so a dark hero/section/closing
slide gets depth instead of a flat fill. `pptx.RadialGradient` already exists
(centered 50%-inset circular focal, behind the glow ornaments, D-041); the scene
`Background` only exposed a 2-stop *linear* fill, so radial was unreachable.

**Decision:** Add `BackgroundRadial` to `BackgroundKind` (appended **last**, after
`BackgroundAsset`, so existing values are unchanged → byte-identical).
`renderBackground` resolves a background's stops via a new shared
`backgroundGradientStopsFor(bg)` — the multi-stop `Stops` (validated 2..8
ascending, D-105) when present, else the legacy 2-role `Gradient` pair at Pos 0/1
— and feeds them to `pptx.RadialGradient(stops...)`. The existing
`BackgroundGradient` (linear) case is refactored through the same resolver; its
2-role mapping is identical, so a legacy gradient background stays byte-identical
(pinned by the existing tests). Invalid explicit stops warn and skip (RFC §10.2,
D-026 — no panic).

**Center-only focal (deferred offset).** `pptx.RadialGradient` hard-codes a
centered focal; biasing it needs a focal-rect knob on the builder. R13.2
explicitly allows *"otherwise document center-only"*, so V1 ships the centered
spotlight (the common vignette case) and **defers** the focal-offset knob — no
new `Background` field this phase. This mirrors the Phase-65 ribbon
diagonal-rotation deferral (D-098): a documented §4.3 deviation, revisited only
if a real off-center-spotlight case appears (then a builder focal-rect parameter
+ a `Background` focal field).

**Consequences:** Scene-side kind + render case only — no builder change (P1), no
new OOXML element, no `restorenamespaces` change (`<a:gradFill>`/`<a:path
path="circle">`/`<a:gs>` already emit via the glow ornaments). Deterministic
(pure integer-EMU through the existing fill path). Tested: radial (multi-stop and
legacy 2-role) emits the circular focal and round-trips with
`GradientRead.Radial == true`; invalid stops warn + skip; deterministic;
`String() == "radial"`; the refactored linear/legacy paths stay byte-identical.

---

## D-107 — Decoration color role (`Decoration.Color`) + `Recipe` signature break (Wave 13 / Phase 73, R13.5)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.5 (`DECKARD-PRODUCT-REQUIREMENTS.md`, CRITICAL · engine) — every
curated ornament hard-codes `pptx.ColorAccent`, so neutral-grey paper grain, a
white starfield, or multi-color confetti (the reference's decoration palette) are
unreachable and Deckard omits decoration. The engine must let the caller pick the
decoration color.

**Decision:** Add `Decoration.Color *pptx.ColorRole` (nil = `ColorAccent`) and
thread a trailing `role pptx.ColorRole` parameter through the `ornaments.Recipe`
signature and all six curated recipes; the `accent(alpha)` helper becomes
`roleFill(role, alpha)`, and the glow recipes use the role in their gradient
stops. `render_decoration` resolves `role := ColorAccent; if v.Color != nil {
role = *v.Color }` and passes it. A nil `Color` → `ColorAccent` → every recipe
fills exactly as before → **byte-identical**.

**`Color` is a pointer (§4.3 deviation).** R13.5's spec text says
`Decoration.Color pptx.ColorRole` (zero = ColorAccent). But `ColorRole`'s zero is
`ColorCanvas` (a real color), so a value field could not mean "accent by default"
without remapping zero — which would make `ColorCanvas` decoration impossible.
Use `*pptx.ColorRole` (nil = accent), the same resolution D-054 reached for
`Card.HeaderFill`/`StatusDot` and D-098 for `Ribbon.Color`.

**Public `Recipe` signature break.** `ornaments.Recipe` is exported (aliased as
`scene.OrnamentRecipe`, registered via `scene.WithOrnamentExtension`). Adding the
`role` parameter is a v0.x breaking change — caller ornament extensions must add
the (possibly-ignored) `role` arg. The spec explicitly calls for changing the
signature; the break is documented in `CHANGELOG.md` under Changed.

**Consequences:** No new theme token (a mechanism over existing surface roles, P2),
no OOXML/`restorenamespaces` change (same `<a:solidFill>`/`<a:gradFill>` shapes,
different resolved `srgbClr`), deterministic. `Decoration.Palette []pptx.ColorRole`
for multi-hue scatter is **deferred** to R13.6 (starfield), which will cycle the
single-color mechanism this phase ships. Tested: a `Color`-set decoration emits a
different `srgbClr` than the accent default; a nil-`Color` decoration is
byte-identical per curated preset; the role threads through all six recipes
(solid + glow); the caller-extension path compiles and renders with the role.

---

## D-108 — Surface fill gradient (`Card.FillGradient`) (Wave 13 / Phase 74, R13.8)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.8 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine) — reference
cards use gradient fills for depth (a vertical fade from a saturated top to a
lighter base); Deckard's flat single-color cards read as solid swatches. The
card surface fill needs an optional gradient.

**Decision:** Add a scene `GradientFill{From, To pptx.ColorRole; Angle int}` type
and `Card.FillGradient *GradientFill`. `renderCardChrome` picks the surface fill:
when `FillGradient` is non-nil, `pptx.LinearGradient(Angle, {0,From},{1,To})`;
else the unchanged `pptx.SolidFill(TokenColor(c.fill))`. nil = solid →
**byte-identical**. A pointer makes "unset" unambiguous (`ColorRole`'s zero is a
real color); `Card` is already non-comparable (`Body []SlideNode`), so the
pointer field adds no constraint. A distinct 2-role `{From,To,Angle}` type (vs the
N-stop `Background.Stops`) is the natural surface-depth API.

**No engine auto-tint.** R13.8 mentions a convenience where `To` defaults to a
slightly-darker role if only `From` is given. Which direction and how much is a
*taste* decision → the soul's (D-026), not the engine's. Both stops are explicit;
the auto-tint is documented as the caller's job.

**Scope: Card only.** `cardChrome` is shared with CardSection, but only `Card`
gets the field this phase; Bento-cell and Container fills are separate paths.
Card is the dominant case and satisfies the acceptance criterion; the
`GradientFill` type is reusable when CardSection/Bento/Container follow (§4.3
deviation).

**Consequences:** No new theme token (mechanism over surface roles, P2), no OOXML
/ `restorenamespaces` change (`<a:gradFill>`/`<a:gs>` already emit via the
background gradient path), deterministic. The header band / ribbon / status dot
draw over the gradient surface exactly as over a solid fill. Tested: a gradient
card emits `<a:gradFill>` with the `From`/`To` role colors; a solid card is
byte-identical and emits no surface gradient; both deterministic.

---

## D-109 — Text/number watermark decoration (`DecorationText`) (Wave 13 / Phase 75, R13.9)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.9 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine) — the
reference uses oversized translucent index numbers (big faint "01/02/03") as a
structural device. Deckard's `Decoration` is preset-or-asset only (no text mode),
so a slide-level ghost number had to be faked via a `Card.Watermark`, which
overflowed. The engine needs a slide-level text watermark.

**Decision:** Add a `DecorationText` kind (appended **last** after
`DecorationAsset`, so existing kinds are byte-identical) plus `Decoration.Text
string` and `Decoration.FontSize float64` (points; 0 = a box-height "fill the
box" default). `renderDecoration` draws one centered `TypeDisplay` run in a text
frame at the decoration box, colored `TokenColorAlpha(role, opacityAlpha)` (role =
`Decoration.Color`, nil = `ColorAccent` — D-107), sized via `RunStyle.FontScale =
targetPt / ResolveType(TypeDisplay).Size` (FontScale > 1 grows — D-074). It
reuses the `Card.Watermark` text-alpha pattern (D-054). Validation requires a
non-empty `Text`.

**Mechanism, not opinion (D-026).** The engine draws the glyph at whatever
`Opacity`/`Color` the caller supplies; it does not impose a default faintness —
the subtle-alpha band is the soul's (R13.13). Decorative: one centered run, no
overflow/wrap logic (the frame clips a short number/word). Glyph rotation
(diagonal "DRAFT") is deferred — the builder has no rotated-text primitive (same
limit as the ribbon diagonal, D-098).

**Consequences:** Minimal wiring — `DecorationText` is native (not
`DecorationAsset`), so `nodeUsesAssets` stays false (parallel-safe); it is a
`Decoration`, so it inherits the renderNode safe-area exemption (decorations
bleed by design), the layer z-order split (default `LayerBackground` → behind
body), and the ornament-name validation early-return. No new theme token
(mechanism over color/type tokens, P2), no OOXML/`restorenamespaces` change (the
same text-run XML the card watermark emits). Deterministic (integer-EMU box
height → points → `@sz` 1/100-pt). Tested: a watermark emits the text + a low
`<a:alpha>`; empty `Text` fails Stage-1; curated decorations byte-identical;
re-render deterministic.

---

## D-110 — Starfield scatter ornament (`starfield`) (Wave 13 / Phase 76, R13.6)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.6 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine) — the
reference's dark slides carry an irregular, sparse starfield (dots of varying
size and opacity) for depth. Deckard's only scatter is `noise_overlay`, a uniform
lattice of identical dots — it cannot produce the organic look.

**Decision:** Add a curated `starfield` ornament
(`assets/ornaments.Starfield`, registered `NameStarfield = "starfield"`): a
box-derived lattice (`cols = box.W/pitch`, `rows = box.H/pitch`) where each cell's
index is run through a fixed integer hash to perturb the dot position, pick its
size from `{1,2,3}pt`, pick its alpha from `{35,60,100}%` of the caller alpha, and
sieve ~20% of cells empty for sparseness. The dot color is the recipe's `role`
(Decoration.Color, default accent — D-107). No `math/rand`, no clock (D-035) — two
renders are byte-identical. The total is capped at 2000 dots to protect part size.

**Density via the box, not a caller param.** The `Recipe` signature
(`func(sl, box, alpha, rotationDeg, role) int`) has no pitch/density parameter,
and changing it again would be a third break this wave. So the dot count derives
from the box size at a fixed internal pitch: a full-bleed box (with `Bleed`) gets
a dense field, a small box a sparse one — the caller controls density by sizing
the decoration. An explicit caller pitch/density (and the past-cap
`LayoutWarning`) is R13.7's scope; the curated recipe has no `r.warn` hook, so the
cap is silent here. Multi-hue `Decoration.Palette` scatter (deferred from R13.5)
stays deferred — a `[]ColorRole` cannot flow through the fixed `Recipe` signature.

**Consequences:** Additive — the closed curated-name set grows to seven
(`TestCurated_HasSixOrnaments` updated), existing ornaments unchanged. No new
theme token (the dot color is a role, P2), no OOXML/`restorenamespaces` change
(the same `a:prstGeom` ellipses + `a:solidFill`/`a:alpha` the other patterns
emit). Tested: ≥2 distinct dot sizes and ≥2 distinct alphas; a bigger box yields
more dots than a smaller one; two renders byte-identical; the role colors the
dots.

---

## D-111 — Pattern density / pitch (`Decoration.Pitch`) + 2nd `Recipe` break (Wave 13 / Phase 77, R13.7)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.7 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine) — `grid_dots`
is a fixed 6×4 lattice and `noise_overlay` a fixed 12×8 regardless of box size,
so a full-bleed texture is 24 dots smeared across 13 inches, not the fine,
consistent dot grid the reference uses. The pattern repeat count must derive from
a caller pitch.

**Decision:** Extend the `ornaments.Recipe` signature with a trailing `pitch
pptx.EMU` (a second v0.x break this wave, after D-107's `role`) and add
`Decoration.Pitch pptx.EMU`. The three pattern recipes (`GridDots`,
`NoiseOverlay`, `Starfield`) compute `cols = box.W/pitch`, `rows = box.H/pitch`
when `pitch > 0` (via the shared `patternDims`), else their legacy fixed counts
(6×4 / 12×8 / `starfieldPitch`). Each caps at `patternMaxDots` (2000).
`render_decoration` passes `v.Pitch` and emits one `LayoutWarning` when `Pitch >
0`, the preset is a pattern (`grid_dots`/`noise_overlay`/`starfield`), and the
projected lattice exceeds the cap. The four non-pattern recipes take and ignore
the param.

**Why a 2nd positional break, not a struct refactor.** A params-struct `Recipe`
(absorbing alpha/rotation/role/pitch) would future-proof against further
additions, but it is a much larger churn across all seven recipes + caller
extensions. The minimal trailing-positional add mirrors D-107 and is mechanical;
the struct refactor is recorded as a V2 cleanup. Both breaks are documented in
`CHANGELOG.md`.

**Cap in the recipe, warning in render_decoration.** A curated recipe has no
`r.warn`/`slideID`, so the file-size cap lives in the recipe (silent) and the
operator-facing warning lives in `render_decoration` (where `slideID` is in
scope). The warning uses an upper-bound projection (`cols*rows`, before the
sieve), gated to the pattern presets so a glow with a stray `Pitch` does not
false-warn. `ornamentPatternCap` mirrors `assets/ornaments.patternMaxDots` (kept
in sync by the cap test).

**Consequences:** `pitch == 0` (every existing deck — the field is new) keeps the
legacy fixed counts → byte-identical; the starfield's `pitch == 0` path uses its
`starfieldPitch` default, identical to Phase 76. No new theme token (P2), no
OOXML/`restorenamespaces` change, deterministic (integer division, no RNG —
D-035). Tested: `grid_dots` at 0.4in over a 13in box yields ≫ 24 dots; a smaller
box proportionally fewer; a legacy pattern byte-identical per preset; a tiny
pitch warns once and caps at ≤ 2000; deterministic.

---

## D-112 — Gradient-mesh background (`BackgroundMesh` / `MeshGlow`) (Wave 13 / Phase 78, R13.4)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.4 (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · engine) — the
reference cover and light content slides carry a soft "mesh" glow: diffuse
colored light pooling at one or two corners over the paper, not a single straight
gradient. The single linear/radial full-bleed fill cannot produce the multi-pool
look.

**Decision:** Add a `BackgroundMesh` kind (appended **last** to `BackgroundKind`
→ existing values byte-identical) plus `MeshGlow{Anchor; Color pptx.ColorRole;
Radius pptx.EMU; Alpha int}` and `Background.Mesh []MeshGlow`. `renderBackground`
draws a base `SolidFill(TokenColor(bg.Color))` (zero = `ColorCanvas` = the
paper/dark canvas) then, for each glow with `Radius > 0`, a radial-gradient
ellipse centered on `Anchor.Point(full)` fading `TokenColorAlpha(Color, Alpha)`
(center) → alpha 0 (edge), in slice order (deterministic). An empty `Mesh` follows
the `BackgroundNone` path (nothing on a light slide; the dark canvas on a dark
variant) — so "absent config → no shapes".

**Background kind, not a curated ornament.** R13.4's spec offered either a curated
`gradient_mesh` ornament or a `BackgroundMesh` kind. The Background kind is the
cleaner full-slide convenience: it composes the base canvas + the pooled glows in
one self-contained spec at the lowest z-layer, and reuses the existing radial fill
(D-106) and role/alpha tokens (D-107). Callers who want foreground pools still
have the role-colored glow *ornaments* (`radial_glow`/`glow_ring`, Phase 73).

**Consequences:** No new theme token (a mechanism over surface roles + the alpha
token, P2; the soul keeps the alpha subtle, R13.13), no OOXML/`restorenamespaces`
change (the same `<a:gradFill>`/`<a:path>` radial ellipses the glow ornaments
emit), deterministic (fixed slice order, integer-EMU). Not asset-bearing, so
`slideUsesAssets` stays false (parallel-safe). `Background` is already
non-comparable (the `Stops` slice), so the `Mesh` slice adds no constraint.
Tested: a 2-glow mesh emits a base rect + 2 distinct-anchor radial ellipses; an
empty mesh is byte-identical to no background; deterministic; `String() ==
"mesh"`.

---

## D-113 — Focal glow behind a card (`Card.Backdrop`) (Wave 13 / Phase 79, R13.10)

**Status:** Accepted. **Date:** 2026-06-24.

**Context:** R13.10 (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine) — the
reference puts a soft glow precisely behind a focal element (a card sits in a
faint halo). Deckard decorations anchor only to the slide region, not a computed
node box, so a glow behind the middle card across any layout is unreachable.

**Decision:** Add `Card.Backdrop *Decoration`. `renderCard` draws it via
`r.renderDecoration(ps, box, *v.Backdrop, slideID)` **before** `renderCardChrome`,
passing the card's computed, safe-area-clamped box as the decoration's region. So
a center-anchored, larger-than-card, bleeding `radial_glow` becomes a halo
centered on the card that spills beyond it and sits behind the card's fill
(z-order: backdrop first, chrome on top). nil = nothing (byte-identical). This is
the simplest additive form the req names (option (a), a `Card`/`Container` field);
a general node-relative anchor (option (b)) is broader and deferred — the
`Card.Backdrop` form covers the motivating focal-card case.

**Consequences:** No renderer change beyond the one call — `renderDecoration`
already places a decoration within an arbitrary region. No new theme token (the
glow color is a surface role + the decoration's `Opacity` alpha, P2; the soul
keeps it subtle, R13.13), no OOXML/`restorenamespaces` change. `Card` is already
non-comparable (`Body []SlideNode`), so the pointer adds no constraint. The caller
sets `Bleed: true` for the halo (the existing off-region warning is suppressed,
as for slide-region decorations); the R13.7 pitch-cap warning only fires for
pattern presets, so a glow backdrop never trips it. Tested: a `radial_glow`
backdrop's radial ellipse precedes the card's `roundRect` fill in the slide XML;
a nil-backdrop card is byte-identical and emits no glow; deterministic.

---

*Append new entries below this line.*
