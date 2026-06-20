# Phase 22 — content aware height

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §10.2 (content-bbox-driven layout, overflow as `LayoutWarning`)
**Deps:** Phase 13 (alignment + `scene/metrics.go`), Phase 05–08 (scene spine,
leaves, containers, table). External: none.
**Status:** In progress

---

## 1. Goal

Make a scene node's allotted slot height reflect how much its text actually
wraps, so stacked nodes stop overlapping and the overflow `LayoutWarning` fires
truthfully — while single-line content renders byte-identically to today.

## 2. Why now

This phase opens **Wave 8 — Post-V1 engine extensions**, the first wave of
caller-driven engine mechanisms requested by the product built on pptx-go
(`DECKARD-PRODUCT-REQUIREMENTS.md`, requirement R1, HIGH). R1 is the highest
priority of that backlog: agent-authored slides overlap and clip because
`preferredHeight` allots a fixed height per paragraph / list item regardless of
text length, and the same under-count means overflow is silently under-reported.
Content-aware height also unblocks truthful overflow, a prerequisite for the
product's "this slide is too full" signal. It honors `RFC §10.2`, which already
mandates a *content-driven* preferred bbox — the fixed-height shortcut is the
gap this phase closes.

## 3. RFC sections implemented

- `RFC §10.2` — "a node reports its preferred bbox; the engine assigns it a
  slot; … Overflow is reported as a `LayoutWarning` (not an error)." This phase
  makes the *preferred bbox height* content-driven (the section's stated model)
  instead of a fixed per-node constant, and makes the overflow warning fire on
  real wrapped overflow. The companion *grow-to-fit* direction (distributing
  slack to flexible nodes) is **out of scope** here and is a sibling Wave 8
  phase (Deckard R2).

## 4. Brief findings incorporated

- `docs/research/09-text-height-metrics.md` — *a char-budget line model
  (`ceil(naturalWidth / availableWidth)`) is deterministic and sufficient* →
  this plan adds `wrappedLines` to `scene/metrics.go` built on the existing
  pinned `naturalWidth` model, no new calibration.
- `docs/research/09-text-height-metrics.md` — *width must be threaded into
  `preferredHeight`* → the signature becomes `preferredHeight(n, avail, theme)`
  and every caller passes its box width.
- `docs/research/09-text-height-metrics.md` — *single-line byte-identity falls
  out by reusing each fixed per-line constant as the line height* → `Prose`
  stays `0.4"`/line, `List` `0.32"`/item, `Heading` `0.6"`; fixed-chrome nodes
  use `oldFixed + (lines−1)×lineHeight`, so `lines = 1` reproduces today's value
  exactly.
- `docs/research/09-text-height-metrics.md` — *overflow becomes truthful with no
  separate code path* → the existing `totalH > box.H` check now consumes
  content-aware heights; no new warning plumbing.

## 5. Findings I'm departing from

None. This plan implements brief 09's recommendations as written. (The brief's
two refinement open-questions — per-role char-width calibration and list-marker
indent — are explicitly deferred there, not departed from.)

## 6. Decisions referenced

- `D-026` — *Engine, not product.* Height estimation is a placement mechanism,
  not a content opinion; this phase adds no render mode, legibility heuristic,
  or "too full" judgment — it only sizes a slot to its content and reports
  overflow as data.
- This plan files a new decision **D-051 — content-aware `preferredHeight`** in
  `docs/decisions.md` (the one intentional layout change for multi-line text,
  and the deterministic char-budget model behind it).

## 7. Architecture

The estimator and the slot-sizing function both already live in `scene/`; this
phase deepens them, adding no package and no exported scene/pptx symbol.

```text
scene/metrics.go      naturalWidth(rt, theme)              (existing)
                      └─ wrappedLines(rt, base, avail, theme) int   NEW
                         = ceil(naturalWidthAt(rt, base, avail) / avail), ≥1
                           (1 when avail ≤ 0 or theme == nil → byte-identical)

scene/render.go       preferredHeight(n)            →  preferredHeight(n, avail, theme)
                      nodesHeight(nodes)            →  nodesHeight(nodes, avail, theme)
                      stackIn / alignedStackIn pass box.W + r.theme
                      container cases thread a sub-width (avail/cols, avail/2,
                      avail − padding) into the recursion.

Overflow:  alignedStackIn / stackIn keep their `totalH > box.H` check; with
           content-aware `totalH` it now fires on real wrapped overflow.
```

Single source of truth: `preferredHeight` stays *the* slot-height function
(no parallel "content height" function), so the body stack, containers, and the
alignment math can never disagree about a node's height.

## 8. Files added or changed

```text
scene/metrics.go                         # CHANGED — adds wrappedLines; header/comment refresh
scene/render.go                          # CHANGED — preferredHeight/nodesHeight take (avail, theme);
                                         #           Prose/List/Heading/Quote/Callout/Table content-aware;
                                         #           container cases thread sub-width; new pinned EMU consts
scene/render_height_test.go              # NEW — no-overlap, overflow-fires, single-line byte-identity,
                                         #       determinism for a multi-line fixture
scene/align_test.go                      # CHANGED — 3 preferredHeight call sites take the new args
scripts/smoke/phase-22.sh                # NEW — phase smoke
docs/research/09-text-height-metrics.md  # NEW — informing brief
docs/research/INDEX.md                   # CHANGED — registers brief 09 under scene
docs/plans/phase-22-content-aware-height.md  # NEW — this plan
docs/plans/README.md                     # CHANGED — opens Wave 8, adds Phase 22 to the index
docs/decisions.md                        # CHANGED — adds D-051
docs/glossary.md                         # CHANGED — adds "content-aware height", "wrapped-line estimate"
skills/compose-a-scene/SKILL.md          # CHANGED — overflow/height behavior note (§19)
docs/site/...                            # CHANGED — scene layout/overflow doc note (§19)
```

No new exported API ⇒ no skill *surface* change beyond a behavior clarification;
the §19 update is a documentation note that overflow is now content-aware.

## 9. Public API surface

**None.** This phase changes no exported symbol on `pptx` or `scene`. The
affected functions (`preferredHeight`, `nodesHeight`, `wrappedLines`) are all
unexported `scene` internals. `Stats.Warnings` already carries the overflow
`LayoutWarning`; its shape is unchanged — only *when* it fires changes.

Because no public API and no new scene IR node is added, the §13/§4.2 "new
public API ⇒ smoke check" rule is satisfied by a behavior smoke (overflow now
fires) rather than an API smoke.

## 10. Risks

- **R1 — silent geometry shift for existing multi-line fixtures.** Any test or
  golden that hard-codes a multi-line node's slot Y/height changes.
  **Mitigation:** there are no byte-golden snapshots in the repo (determinism is
  proven by parallel≡sequential equality, which is preserved); the full scene
  suite is run and any assertion keyed to a multi-line node is re-baselined to
  the new (correct, non-overlapping) value, recorded in §16.
- **R2 — non-determinism via float width math.** `naturalWidth` multiplies a
  float font size. **Mitigation:** the float result is truncated to integer EMU
  exactly as today (already production-trusted for alignment), and `wrappedLines`
  is pure integer ceil division thereafter; the determinism test renders a
  multi-line fixture under 1 vs N workers and asserts byte-identity.
- **R3 — single-line regression.** A formula change could perturb single-line
  output. **Mitigation:** every content-aware case reduces to its prior constant
  at `lines = 1`; an explicit test renders the same single-line scene before/
  after-shaped fixtures and asserts byte-identity against the fixed-height path
  (`avail ≤ 0` fallback returns the old value).

## 11. Acceptance criteria

1. A `Prose` paragraph long enough to wrap to N lines is allotted ≥ N
   line-heights: its slot height ≥ `N × 0.4"`.
2. In a multi-line fixture (long `Prose` followed by a `Heading`), the second
   node's box `Y` is ≥ the first node's box `Bottom()` — no overlap.
3. A slide whose content-aware stack height exceeds the body region emits the
   "content overflows its region" `LayoutWarning` (it does not today).
4. A single-line scene renders **byte-identical** to the pre-phase fixed-height
   output (additive / backward-compatible invariant).
5. A multi-line scene renders **byte-identical** across 1 vs N workers
   (determinism holds).
6. `wrappedLines` is monotonic in text length and returns 1 for empty text,
   `avail ≤ 0`, or `theme == nil`.
7. `make coverage` shows `scene` ≥ its band; `make preflight` passes.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry added; the change keeps `scene` at or
above its existing band (new branches are covered by `render_height_test.go`).

## 13. Smoke check

`scripts/smoke/phase-22.sh` verifies each acceptance criterion mechanically:

1. `OK:` library builds CGo-free.
2. `OK:` no-overlap test passes (criterion 2).
3. `OK:` overflow-fires test passes (criterion 3).
4. `OK:` single-line byte-identity test passes (criterion 4).
5. `OK:` multi-line determinism test passes (criterion 5).
6. `OK:` `wrappedLines` unit test passes (criteria 1, 6).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — `wrappedLines` (monotonic, ceil, empty/zero/nil fallback);
  `preferredHeight` content-aware cases (Prose/List/Heading/Quote/Callout/Table)
  via white-box package tests.
- **Round-trip golden:** N/A — no builder primitive or scene node is added; this
  is a layout-sizing change, not a new emitted shape (the existing round-trip
  coverage of the affected nodes is unchanged).
- **Integration** (`test/integration/`): no — no cross-subsystem seam opened or
  consumed; the change is internal to `scene` layout.
- **Fuzz:** no — no parse/decode surface added.
- **Benchmark:** optional — `preferredHeight` is on the single-slide render hot
  path; a micro-benchmark may be added but is not a gate.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `content-aware height` — a node's slot height derived from its wrapped text
  (line count × line height), not a fixed per-node constant.
- `wrapped-line estimate` — the deterministic `ceil(naturalWidth /
  availableWidth)` line count the layout engine uses to size a text slot.

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-22.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entry D-051 added.
- [ ] Docs site updated for the overflow/height behavior change (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
