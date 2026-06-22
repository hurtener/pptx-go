# Phase 51 — container slide-bounds clamp (R11.3)

**Subsystem:** scene — Layer 2 renderer (container layout)
**RFC sections:** §10, §10.1
**Deps:** Phase 24 (chrome bands / `bodyRegion`), Phase 40 (R10.2 `VAlignFit`),
Phase 27 (Bento), brief 34.
**Status:** Done

---

## 1. Goal

Guarantee that no Bento / Grid / Card ever draws below the slide's printable area:
clamp a container's box to the per-slide safe area so an over-full stack can never
push cells off the bottom edge onto the chrome footer — deterministically, and
byte-identically when the content already fits.

## 2. Why now

R11.3 is the third CRITICAL of Wave 11 (recreation slides 6, 7 clip the bottom
bento row and overlap the footer). It pairs with the opt-in `VAlignFit` (R10.2):
`VAlignFit` reflows an over-full stack when asked; this clamp guarantees the
off-canvas *invariant* even for the default top-anchored stack that opts out.

## 3. RFC sections implemented

- `RFC §10` — layout: containers subdivide a bounded region.
- `RFC §10.1` — deterministic, worker-count-independent geometry (the clamp is a
  pure integer cap).

## 4. Brief findings incorporated

- `docs/research/34-container-slide-bounds-clamp.md`:
  - "`safeArea` already exists as `bodyRegion()`" → `safeArea()` is a named alias;
    the clamp compares against `r.safeArea().Bottom()` (single source of truth with
    the chrome bands).
  - "a single entry-point clamp covers all three containers" → `clampToSafeArea`
    called at the top of `renderBento`, `renderGrid`, `renderCard`.
  - "only fires when `box.Bottom() > safeArea.Bottom()`" → fitting boxes (and
    `VAlignFill`, and a sole container filling the region) are byte-identical.
  - "nesting does not double-warn" → an outer clamp shrinks the box so inner
    containers never individually overflow; one warning at the outermost container.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-071` — `VAlignFit` (R10.2). The clamp is the complementary *invariant*: cap the
  drawn height rather than reflow content; the two compose.
- `D-053` — slide chrome (the eyebrow/footer bands `bodyRegion` reserves).
- `D-083` — **new** — the safe-area clamp.

## 7. Architecture

```text
safeArea() = bodyRegion()                         slide − margins − chrome bands
clampToSafeArea(box, slideID):
    if box.Bottom() > safeArea().Bottom():        cap H to safeArea bottom + warn once
        box.H = safeArea().Bottom() - box.Y       (else unchanged → byte-identical)

renderBento / renderGrid / renderCard  → box = clampToSafeArea(box, slideID) at entry
```

## 8. Files added or changed

```text
scene/render.go                       # CHANGED — safeArea(), clampToSafeArea()
scene/render_bento.go                 # CHANGED — clamp at renderBento entry
scene/render_container.go             # CHANGED — clamp at renderGrid entry
scene/render_card.go                  # CHANGED — clamp at renderCard entry
scene/render_bounds_test.go           # NEW — white-box clamp + bento-within-safe-area
scene/render_bounds_render_test.go    # NEW — black-box overflow warns / fits no-warn
scene/render_parallel_test.go         # CHANGED — TestRenderDeterministic_BoundsClamp
scripts/smoke/phase-51.sh             # NEW — phase smoke
docs/research/34-container-slide-bounds-clamp.md  # NEW — brief 34
docs/research/INDEX.md                # CHANGED — registers brief 34
docs/plans/phase-51-container-slide-bounds-clamp.md  # NEW — this plan
docs/plans/README.md                  # CHANGED — Wave 11 Phase 51
docs/decisions.md                     # CHANGED — adds D-083
docs/glossary.md                      # CHANGED — adds "safe area"
docs/site/guide/scene.md              # CHANGED — overflow note: safe-area clamp
```

No new exported symbol.

## 9. Public API surface

None. `safeArea` / `clampToSafeArea` are unexported. The only user-observable change
is an additional `LayoutWarning` ("container overflow … clamped") on overflow,
surfaced through the existing `Stats.Warnings`.

## 10. Risks

- **R1 — clamping a legitimately-region-filling container.** A sole container handed
  the full body region has `Bottom() == safeArea.Bottom()`. **Mitigation:** the
  clamp fires only on strict `>`, so `==` is unchanged; existing goldens
  (`VAlignFill` to the region bottom included) pass unchanged.
- **R2 — double-warning on nested containers.** **Mitigation:** the outer clamp
  shrinks the box; inner containers get sub-boxes that fit → no inner clamp. A
  test renders a Bento-of-Cards and asserts a single overflow signal.

## 11. Acceptance criteria

1. A box whose bottom exceeds the safe area is shrunk so `Bottom() ==
   safeArea.Bottom()` and a warning is logged (`TestClampToSafeArea_ShrinksOverflow`).
2. A box within (or exactly at) the safe area is returned unchanged with no warning
   (`TestClampToSafeArea_FitsByteIdentical`).
3. After clamping, every emitted bento cell box sits inside the safe area
   (`TestBentoBoxesWithinSafeArea`).
4. An over-tall container warns at render; a fitting one does not
   (`TestContainerOverflow_Warns`, `TestContainerFits_NoWarn`).
5. The clamp is byte-identical across worker counts
   (`TestRenderDeterministic_BoundsClamp`); the full existing golden suite passes.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | default; clamp unit- and render-covered |

## 13. Smoke check

`scripts/smoke/phase-51.sh` runs the six acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` white-box (`render_bounds_test.go`).
- **Round-trip golden:** n/a (no builder API).
- **Integration:** no — same subsystem.
- **Fuzz / Benchmark:** no.

## 15. Vocabulary added

- `safe area` — the slide's printable region (slide minus content margins minus the
  reserved chrome bands); no container may draw below its bottom. Equal to the body
  region; the bound the R11.3 clamp enforces.

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-51.sh` reports `OK ≥ 6` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated.
- [x] Decision entries added (D-083).
- [x] Docs site updated for the new overflow warning.
- [x] Affected agent skill(s) updated (n/a — no surface change).
