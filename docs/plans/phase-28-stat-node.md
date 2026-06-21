# Phase 28 — stat node

**Subsystem:** scene — Layer 2 renderer (`RFC §3.3`)
**RFC sections:** §11.1 (leaf nodes), §10.4 (Stage-1 validation)
**Deps:** Phase 06 (leaf composers), Phase 07 (Grid, for the strip). External:
none.
**Status:** In progress

---

## 1. Goal

Add a `Stat` leaf node — a display-scale value with a label and an optional
directional delta — so pricing/metric slides use a real hero-number node instead
of faking it with headings, and a row of stats in a `Grid` forms a metric strip.

## 2. Why now

Sixth unit of **Wave 8** (`DECKARD-PRODUCT-REQUIREMENTS.md` R6, LOW), the first
of the two remaining LOW units, picked up after R1–R5 per the
one-requirement-per-PR cadence. The engine has no hero-number node, so callers
fake stats with `Heading`s and lose the value/label/delta structure and the
directional color.

## 3. RFC sections implemented

- `RFC §11.1` — a new native text leaf (`Stat`) composed from existing builder
  text primitives (no new OOXML — P1).
- `RFC §10.4` — Stage-1 structural validation for the new node.

## 4. Brief findings incorporated

- `docs/research/15-stat-node.md` — *`Stat` is a leaf, not a container* →
  `Stat{Value, Label, Delta string, DeltaTone}`; policy `{}`; no `walk*`
  recursion.
- `docs/research/15-stat-node.md` — *a `DeltaTone` enum with a neutral zero keeps
  the delta optional and the default inert* → `DeltaNeutral`/`DeltaUp`/
  `DeltaDown`; an empty `Delta` omits the delta line.
- `docs/research/15-stat-node.md` — *the metric strip is free* → a `Grid` of
  `Stat`s is the strip; no new container.
- `docs/research/15-stat-node.md` — *validation is minimal* → Stage-1 requires
  `Value != ""`.
- `docs/research/15-stat-node.md` — *not flexible* → `Stat` is a fixed number
  block, not added to `isFlexible` (a `Grid` of stats still grows as a container).

## 5. Findings I'm departing from

None. The brief's open-questions (numeric formatting, delta arrow glyph, value
auto-fit) are explicitly deferred there.

## 6. Decisions referenced

- `D-026` — *Engine, not product.* The caller supplies value/label/delta/tone
  verbatim; the engine renders them and picks only the tone→color mapping. It
  formats no number.
- `D-012`/`D-030` — token color resolution; the delta tone maps to existing
  tokens (`ColorSuccess`/`ColorError`/`TextMuted`) — no new token (P2).
- This plan files **D-057 — Stat node** in `docs/decisions.md`.

## 7. Architecture

```text
scene/nodes.go     Stat + DeltaTone (DeltaNeutral/DeltaUp/DeltaDown) + KindStat (+ String)

scene/policy.go    policyTable[KindStat] = {}

scene/validate.go  case Stat: Value non-empty

scene/render.go    renderNode:      case Stat → renderStat
                   preferredHeight: case Stat (fixed block height)
                   nodeUsesAssets:  Stat in the leaf "return false" set
                   isFlexible:      NOT added (fixed number block)

scene/render_stat.go  renderStat: one anchored text frame — value (TypeDisplay
                      bold) + label (TypeCaption muted) + optional toned delta;
                      deltaToneColor(tone)                                    NEW
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — Stat/DeltaTone + KindStat + String
scene/policy.go                      # CHANGED — policyTable[KindStat]
scene/validate.go                    # CHANGED — case Stat
scene/render.go                      # CHANGED — renderNode, preferredHeight, nodeUsesAssets
scene/render_stat.go                 # NEW — renderStat + deltaToneColor
scene/render_stat_test.go            # NEW — value/label/delta tone, strip, validation, determinism
scene/scene_test.go                  # CHANGED — allNodes + catalog count 21 → 22
test/integration/roundtrip_test.go   # CHANGED — everyNodeScene (Grid of Stats) + kind-range loop
scripts/smoke/phase-28.sh            # NEW — phase smoke
docs/research/15-stat-node.md        # NEW — informing brief
docs/research/INDEX.md               # CHANGED — registers brief 15
docs/plans/phase-28-stat-node.md     # NEW — this plan
docs/plans/README.md                 # CHANGED — adds Phase 28 to Wave 8
docs/decisions.md                    # CHANGED — adds D-057
docs/glossary.md                     # CHANGED — adds "Stat", "Delta tone"
docs/site/catalog/text-leaves.md     # CHANGED — Stat node docs (§19)
skills/compose-a-scene/SKILL.md      # CHANGED — Stat node entry (§19)
```

## 9. Public API surface

```go
// scene (nodes.go)
type DeltaTone int
const (
    DeltaNeutral DeltaTone = iota // muted (zero value)
    DeltaUp                       // success/green
    DeltaDown                     // error/red
)

type Stat struct {
    node
    Value     string    // big number, rendered at display scale (e.g. "$2,200")
    Label     string    // caption below the value
    Delta     string    // optional delta (e.g. "+12%"); "" = no delta line
    DeltaTone DeltaTone // delta color direction
}
func (Stat) NodeKind() NodeKind // KindStat
```

New scene IR node ⇒ a smoke check and Stage-1 validation land in this PR
(§4.2/§13). No new builder API, no new theme token (P2).

## 10. Risks

- **R1 — incomplete node wiring.** **Mitigation:** the round-trip every-node
  guard (contiguous `KindHero..KindStat`) and the catalog count (22) +
  `TestPolicy_MatchesStructs` fail loudly if `Stat` is missed.
- **R2 — wrong delta colors.** **Mitigation:** a test asserts the up/down/neutral
  runs resolve to the success/error/muted token colors.
- **R3 — Stat wrongly stretched under fill.** **Mitigation:** `Stat` is omitted
  from `isFlexible`; a test asserts `isFlexible(Stat{})` is false while a `Grid`
  of stats is flexible.

## 11. Acceptance criteria

1. A `Stat` renders a large value + label, and a delta line when `Delta` is set;
   the delta run's color follows `DeltaTone` (up=success, down=error,
   neutral=muted).
2. A `Grid` of `Stat`s renders a metric/pricing strip (one stat per cell).
3. Stage-1 rejects a `Stat` with an empty `Value`.
4. `Stat` is not flexible; a bare `Stat` round-trips and the catalog has 22 kinds.
5. A stat deck renders byte-identical across 1 vs N workers.
6. `make coverage` shows `scene` ≥ its band; `make preflight` + `make lint` pass.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `scene` | 80% | default for the scene renderer package (no override) |

No new package ⇒ no `coverage.json` entry; new branches covered by
`render_stat_test.go` + the catalog/round-trip updates.

## 13. Smoke check

`scripts/smoke/phase-28.sh` verifies each acceptance criterion:

1. `OK:` library builds CGo-free.
2. `OK:` a Stat renders value + label + toned delta (criterion 1).
3. `OK:` a Grid of Stats renders a strip (criterion 2).
4. `OK:` Stage-1 rejects an empty-value Stat (criterion 3).
5. `OK:` catalog has 22 kinds / round-trip covers Stat (criterion 4).
6. `OK:` stat render is deterministic across workers (criterion 5).

`SKIP` is used for none — the surface lands entirely in this PR.

## 14. Tests

- **Unit:** `scene` — `deltaToneColor`/render assertions (value/label/delta text +
  tone color in the slide XML), `isFlexible(Stat)` false, validation.
- **Round-trip golden:** the existing `everyNodeScene` round-trip extends to
  `Stat`.
- **Integration** (`test/integration/`): the round-trip "every node" test gains
  `Stat` (inside a `Grid` strip).
- **Fuzz / Benchmark:** no.

## 15. Vocabulary added

Filed in `docs/glossary.md` in this PR (alphabetical):

- `Stat` — a scene IR leaf node: a display-scale value with a label and an
  optional directional delta; a `Grid` of them forms a metric strip.
- `Delta tone` — a `Stat.Delta`'s color direction (`DeltaTone`):
  `DeltaUp` (success), `DeltaDown` (error), `DeltaNeutral` (muted).

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for `scene`.
- [ ] `scripts/smoke/phase-28.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] `make lint` clean.
- [ ] Glossary updated.
- [ ] Decision entry D-057 added.
- [ ] Docs site updated for the Stat node (§19).
- [ ] Affected agent skill (`compose-a-scene`) updated (§19).
