# Phase 58 — list bullet hanging indent (R11.10)

**Subsystem:** scene — Layer 2 renderer (List leaf)
**RFC sections:** §11.1, §8.4
**Deps:** Phase 47 (R10.9 / D-078 `BulletIndent`/`IndentTight`), Phase 39 (R10.1
header geometry), brief 41.
**Status:** Done

---

## 1. Goal

Make the `IndentTight` bullet hanging indent proportional to the body type size
(anchored to `In(0.25)` at 14 pt) instead of a fixed value, so the bullet-to-text
gap stays tight and scales at any body size.

## 2. Why now

R11.10 is a Wave-11 MED unit (recreation slides 2, 4). Its mechanism (the
`BulletIndent` override + the `IndentTight` preset) already landed with R10.9/D-078;
R11.10's delta is proportionality and the wrapped-header interaction (R10.1).

## 3. RFC sections implemented

- `RFC §11.1` — the List leaf.
- `RFC §8.4` — bullet paragraph indent (`marL`/`indent`).

## 4. Brief findings incorporated

- `docs/research/41-list-bullet-hanging-indent.md`:
  - "the mechanism exists; R11.10's delta is proportionality" → `listTightIndent()`
    = `listTightIndentBase × bodySize / 14`, anchored byte-identical at 14 pt.
  - "the ≤1.5×-glyph acceptance is an example" → assert the indent is tighter than the
    0.5" default (≤ `In(0.3)`) and scales, not the stricter 1.5×-glyph bar (which
    would break the D-078 byte-identity).
  - "wrapped-header interaction already holds" → R10.1 (Phase 49 golden) guards the
    list start Y below the grown header.

## 5. Findings I'm departing from

*"none"* — the `≤ 1.5× glyph` example target is relaxed to the proportional + tighter-
than-default bar (documented).

## 6. Decisions referenced

- `D-078` — `BulletIndent` / `IndentTight` (the mechanism).
- `D-070` — wrapped-header geometry (the list start Y respects it).
- `D-090` — **new** — the proportional tight indent.

## 7. Architecture

```text
listTightIndent() = round(listTightIndentBase × bodySize / 14)   # In(0.25) at 14pt
bulletIndent(IndentTight) → listTightIndent();  IndentNormal → 0 (builder 0.5")
renderList → ParagraphOpts.BulletIndent = r.bulletIndent(v.Indent)
```

## 8. Files added or changed

```text
scene/render_leaves.go                       # CHANGED — proportional listTightIndent + bulletIndent method
scene/render_list_indent_proportional_test.go # NEW — anchor byte-identical / scales / gap tight
scripts/smoke/phase-58.sh                    # NEW — phase smoke
docs/research/41-list-bullet-hanging-indent.md  # NEW — brief 41
docs/research/INDEX.md                       # CHANGED — registers brief 41
docs/plans/phase-58-list-bullet-hanging-indent.md  # NEW — this plan
docs/plans/README.md                         # CHANGED — Wave 11 Phase 58
docs/decisions.md                            # CHANGED — adds D-090
docs/glossary.md                             # CHANGED — List indent proportional note
```

No new exported symbol.

## 9. Public API surface

None. `listTightIndent` / `bulletIndent` are unexported. The user-visible effect is
that `List.Indent = IndentTight` scales with the deck's body size.

## 10. Risks

- **R1 — breaking the D-078 byte-identity.** **Mitigation:** the formula yields exactly
  `In(0.25)` at the default 14 pt body; the existing `marL="228600"` test passes.
- **R2 — float non-determinism.** **Mitigation:** a pure function of the theme; result
  is an integer EMU, identical across worker counts.

## 11. Acceptance criteria

1. At the default 14 pt body, `listTightIndent()` is exactly `In(0.25)`
   (`TestListTightIndent_AnchorByteIdentical`); the R10.9 `marL="228600"` guard passes.
2. The indent scales with the body size (2× → 2×, ½ → ½)
   (`TestListTightIndent_Proportional`).
3. The indent is tighter than the 0.5" default (≤ `In(0.3)`)
   (`TestListTightIndent_GapTight`).
4. `go test -race ./...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; indent unit-covered |

## 13. Smoke check

`scripts/smoke/phase-58.sh` runs the four acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` white-box (`render_list_indent_proportional_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "List indent" already defined; its entry gains a proportional note.)*

## 16. Plan deviations encountered during implementation

- The `≤ 1.5× glyph` acceptance example is relaxed to "proportional + tighter than the
  0.5" default" to preserve the D-078 byte-identity (`In(0.25)` is ~2.6× a glyph but
  tight vs the default). Documented in brief 41 and D-090.

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-58.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-090).
- [x] Docs site / skill updated (n/a — no surface change; glossary note suffices).
