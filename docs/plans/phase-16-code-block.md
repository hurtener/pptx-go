# Phase 16 — CodeBlock (raster path)

**Subsystem:** scene (code_block)
**RFC sections:** §11.1 (code_block), §12 (D-014)
**Deps:** Phase 11 (image/pic path), Phase 14 (pill chrome precedent)
**Status:** Draft

---

## 1. Goal

Finalize the `code_block` raster path: a caller-rasterized code image with an
optional caption and an optional **language badge**, wiring the previously
set-but-unused `CodeBlock.Language` field.

## 2. Why now

Phase 16 is the last of Wave 4 (`docs/plans/README.md`). The raster path
(image + caption) already shipped (Phase 06 stub, Phase 11 image path), but
`CodeBlock.Language` is carried in the IR and never rendered — a set-but-unused
field (a wiring gap). Finalizing it closes Wave 4's leaf set before Wave 5
(charts).

## 3. RFC sections implemented

- `RFC §11.1` — the `code_block` leaf (completes it: image + caption + badge).
- `RFC §12` (D-014) — the per-node policy "`code_block` → Image (`pic`),
  `asset_id`": unchanged; the badge is a native overlay shape on top of the pic.

## 4. Brief findings incorporated

No informing brief — this finalizes an existing raster path whose design is
already settled by **D-014** (caller-side raster; caption below). The only new
surface is the language badge, a small native pill reusing the chip/card-pill
precedent. (Per `CLAUDE.md §16`, a brief is not authored when the phase adds no
new prior-art-requiring design — recorded here explicitly.)

## 5. Findings I'm departing from

None.

## 6. Decisions referenced

- `D-014` — `code_block` renders as a caller-side raster (`pic`) with a caption
  below. This phase implements its final form.
- `D-036` — asset-resolution failures degrade to a warning (not an error); the
  code_block image path already follows this and is unchanged.
- `D-043` — the card header-pill is the chrome precedent the language badge
  reuses (rounded rect + caption text).
- **`D-045` (new, this PR)** — the language badge is a native overlay pill in
  the image's top-right; `CodeBlock.Language` drives it; relocate the renderer
  to `scene/render_code_block.go`.

## 7. Architecture

A small, single-PR change in `scene`. The existing `renderCodeBlock` moves from
`render_leaves.go` to its own `render_code_block.go` (parity with
`render_card.go` / `render_flow.go`), and gains the badge overlay.

```text
┌──────────────────────────────┐
│                       [ go ] │  ← language badge (top-right overlay pill)
│                              │
│      code raster (pic)       │
│                              │
└──────────────────────────────┘
            main.go                ← caption (existing, below the image)
```

The badge renders after the `pic` so it overlays (OOXML shape-tree order = z).
It is omitted when `Language` is empty (byte-identical to today).

## 8. Files added or changed

```text
scene/render_code_block.go   # NEW — renderCodeBlock relocated here + language badge
scene/render_leaves.go       # CHANGED — renderCodeBlock removed (moved)
scene/render_code_block_test.go # NEW — badge present/absent, caption, raster, parallel
scripts/smoke/phase-16.sh    # NEW — phase smoke
docs/decisions.md            # CHANGED — adds D-045
docs/plans/phase-16-code-block.md # NEW (this file)
docs/glossary.md             # CHANGED — language badge
```

No new public API (the badge is internal compose behavior over the existing
`CodeBlock.Language` field); no theme token added (badge reuses surface/text
tokens).

## 9. Public API surface

None added. `CodeBlock.Language` (already public) becomes rendered. No breaks.

## 10. Risks

- **R1 — Badge overlaps code content.** **Mitigation:** the badge is a small
  fixed-size pill inset into the top-right corner; callers who don't want it
  leave `Language` empty. A render test asserts the badge box stays within the
  image box.
- **R2 — Relocation changes output.** **Mitigation:** the move is mechanical
  (same code); a no-`Language` code_block must stay byte-identical — a
  workers=1 vs N + a golden-ish string check guard it.

## 11. Acceptance criteria

1. A `code_block` with a registered raster renders a `pic` (D-014) plus, when
   `Language != ""`, a native badge pill containing the language text in the
   image's top-right.
2. A `code_block` with `Caption != ""` renders the caption below the image
   (unchanged).
3. A `code_block` with empty `Language` renders **no** badge (byte-identical to
   the pre-Phase-16 output).
4. An unresolved asset degrades to a `LayoutWarning`, not an error (D-036).
5. `scene.Render` is byte-identical for a code_block scene at `workers=1` and
   `workers=N`.
6. `make coverage` shows `scene` ≥ its band.
7. `scripts/smoke/phase-16.sh` reports `OK ≥ count(criteria)`, `FAIL = 0`;
   prior smokes still pass.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; small addition to an existing package |

## 13. Smoke check

`scripts/smoke/phase-16.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` a code_block with a raster + language renders pic + badge.
3. `OK:` a code_block with empty language renders no badge.
4. `OK:` caption renders below the image.
5. `OK:` code_block render is byte-identical workers=1 vs N.

## 14. Tests

- **Unit:** `scene` (badge present/absent, caption, unresolved-asset warning,
  parallel equivalence).
- **Round-trip golden:** none (no new builder primitive; the pic + native pill
  are already round-trip-tested shapes).
- **Integration:** the code_block asset path is already covered by the Phase-06
  integration deck (`test/integration/scene_render_test.go` includes a
  code_block); extend it to carry a `Language` so the badge path is exercised
  end-to-end.
- **Fuzz / Benchmark:** none.

## 15. Vocabulary added

- `language badge` — the small overlay pill on a `code_block` image showing the
  source language (`CodeBlock.Language`).

## 16. Plan deviations encountered during implementation

- *(empty until implementation)*

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-16.sh` reports `OK ≥ count(criteria)` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated.
- [ ] Decision entries added (D-045).
- [ ] (Phase 20+) Docs site updated for user-facing surface changes. (inert)
- [ ] (Phase 20+) Affected agent skill(s) updated. (inert)
