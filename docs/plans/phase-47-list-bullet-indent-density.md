# Phase 47 — list bullet indent density

**Subsystem:** scene — Layer 2 renderer (+ a `pptx` paragraph option)
**RFC sections:** §8.4 (rich text / paragraphs), §11.1 (List)
**Deps:** Phase 03 (text/paragraph builder), Phase 11 (List node)
**Status:** Done

---

## 1. Goal

Add an opt-in list density control — `List.Indent` (`IndentNormal`/`IndentTight`),
plumbed through a new `ParagraphOpts.BulletIndent` — so a caller can tighten the
bullet hanging indent for dense lists, byte-identical by default.

## 2. Why now

R10.9 is the next Wave-10 unit. It fixes the recreation's loose lists with a very
wide marker-to-text gap (DECKARD R10.9 gap). It reuses the per-paragraph layout
metric pattern from `LineHeight` (D-061).

## 3. RFC sections implemented

- `RFC §8.4` — a per-paragraph bullet-indent override on the builder.
- `RFC §11.1` — a list density preset on the `List` node.

## 4. Brief findings incorporated

- `docs/research/30-list-bullet-indent-density.md` — *a per-paragraph
  `BulletIndent` override mirrors `LineHeight`* → added to `ParagraphOpts`;
  `AddParagraph` overrides `MarL`/`Indent` when set.
- `docs/research/30-list-bullet-indent-density.md` — *scene presets, not arbitrary
  values; `IndentTight → In(0.25)`* → `ListIndent` enum + `bulletIndent` mapping.
- `docs/research/30-list-bullet-indent-density.md` — *no token; pinned presets;
  byte-identical default* → `IndentNormal` passes 0, builder keeps its default.

## 5. Findings I'm departing from

- None. (Per-level indent nesting and a continuous scene value are noted as
  out-of-scope future work in the brief; the `ParagraphOpts.BulletIndent` seam
  already accepts any EMU for a direct `pptx` caller.)

## 6. Decisions referenced

- `D-061` — `LineHeight` per-paragraph metric — the byte-identical-at-zero,
  emits-an-`a:pPr`-attribute pattern mirrored here.
- `D-026` — engine, not product — a list density preset is an opt-in mechanism.
- **New:** `D-078` — list bullet indent density — filed in this PR.

## 7. Architecture

```text
pptx:  ParagraphOpts.BulletIndent EMU   // 0 = default 0.5" hanging indent
       AddParagraph: if Bullet set && BulletIndent>0 → MarL=BulletIndent, Indent=-BulletIndent
scene: ListIndent {IndentNormal(0), IndentTight}; List.Indent
       listTightIndent = In(0.25); bulletIndent(IndentTight)=listTightIndent, else 0
       renderList: ParagraphOpts{..., BulletIndent: bulletIndent(v.Indent)}
```

`BulletIndent` 0 (and `IndentNormal`) keeps the builder's default `457200`
hanging indent — byte-identical.

## 8. Files added or changed

```text
pptx/text.go                                 # CHANGED — ParagraphOpts.BulletIndent + AddParagraph override
pptx/text_bullet_indent_test.go              # NEW — tight marL/indent emit + default byte-identical + round-trip
scene/nodes.go                               # CHANGED — ListIndent enum + List.Indent
scene/render_leaves.go                       # CHANGED — listTightIndent, bulletIndent, renderList wiring
scene/render_list_indent_test.go             # NEW — tight list smaller offset + default byte-identical + determinism
scripts/smoke/phase-47.sh                    # NEW — phase smoke
docs/research/30-list-bullet-indent-density.md # NEW — brief 30
docs/research/INDEX.md                       # CHANGED — register brief 30
docs/plans/phase-47-list-bullet-indent-density.md # NEW — this plan
docs/plans/README.md                         # CHANGED — Wave 10 phase index row
docs/decisions.md                            # CHANGED — adds D-078
docs/glossary.md                             # CHANGED — List indent / BulletIndent terms
docs/design/THEME.md                         # CHANGED — list indent pinned-preset note
docs/site/catalog/text-leaves.md             # CHANGED — document List.Indent
skills/compose-a-scene/SKILL.md              # CHANGED — List.Indent in the List section
```

## 9. Public API surface

```go
// pptx
type ParagraphOpts struct {
    // …
    BulletIndent EMU // the bullet hanging indent (marker-to-text offset); 0 = the
                    // default 0.5". Applies only when a bullet is set.
}

// scene
type ListIndent int
const ( IndentNormal ListIndent = iota; IndentTight )
type List struct { /* … */ Indent ListIndent }
```

Additive: `BulletIndent` 0 and `IndentNormal` reproduce the current output.

## 10. Risks

- **R1 — byte-identical regression.** **Mitigation:** `BulletIndent` 0 skips the
  override; a test asserts a default list emits the unchanged `marL="457200"`.
- **R2 — determinism.** **Mitigation:** pinned integer values; a determinism guard
  renders a tight-list deck at 1 and 8 workers.

## 11. Acceptance criteria

1. A list rendered with `IndentTight` shows a smaller, consistent marker-to-text
   offset (`In(0.25)`) across all items and levels.
2. The `IndentNormal`/default preset is byte-identical to the current output.
3. The emitted `marL`/`indent` round-trip through the slide XML.
4. Identical inputs yield identical EMU geometry (deterministic).
5. `make coverage` keeps `pptx` and `scene` ≥ their bands.

## 12. Coverage targets

| Package | Target | Rationale (if override) |
|---|---|---|
| `pptx` | 85% | small AddParagraph branch + test |
| `scene` | 80% | default for the scene package |

## 13. Smoke check

`scripts/smoke/phase-47.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` tight bullet indent emits the smaller marL/indent
   (`TestBulletIndent_TightEmits`).
3. `OK:` default bullet indent byte-identical (`TestBulletIndent_DefaultByteIdentical`).
4. `OK:` a tight List renders a smaller offset (`TestListIndent_TightSmallerOffset`).

## 14. Tests

- **Unit / codec:** `pptx` — the tight `marL`/`indent` emit, default byte-identical,
  round-trip.
- **Unit:** `scene` — a tight `List` emits the smaller offset; default
  byte-identical; determinism.

## 15. Vocabulary added

- `List indent (density)` — the `List.Indent` preset (`IndentNormal`/`IndentTight`)
  controlling the bullet hanging indent.
- `BulletIndent` — the `pptx.ParagraphOpts.BulletIndent` per-paragraph bullet
  hanging-indent override.

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-47.sh` reports `OK ≥ 4` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-078).
- [x] Docs site + THEME.md updated for user-facing surface changes.
- [x] Affected agent skill(s) updated (compose-a-scene).
