# Phase 52 — content-region reserves chrome (R11.4 verify-and-close)

**Subsystem:** scene — Layer 2 renderer (body region / chrome)
**RFC sections:** §10, §13.3
**Deps:** Phase 24 (D-053 slide chrome / `bodyRegion` shrink), Phase 51 (D-083
safe-area clamp), brief 35.
**Status:** Done

---

## 1. Goal

Verify that the body content region reserves the section-eyebrow and footer bands
when chrome is enabled — so no body content occupies the chrome rows — and close
R11.4 with its named acceptance: a content-maximal chromed slide's body boxes never
intersect the reserved bands, chrome-off slides byte-identical.

## 2. Why now

R11.4 is the HIGH unit after the three Wave-11 CRITICALs. Its mechanism (the
`bodyRegion` shrink) already landed with D-053, and the overflow hole that defeated
it in the recreation is now closed by R11.3's clamp (D-083). Per `CLAUDE.md §17`,
the correct close for an already-implemented requirement is the acceptance test +
a decision, not a reimplementation.

## 3. RFC sections implemented

- `RFC §13.3` — slide chrome (eyebrow + footer); the body region reserves their bands.
- `RFC §10` — layout inside the (reduced) body region.

## 4. Brief findings incorporated

- `docs/research/35-content-region-reserves-chrome.md`:
  - "R11.4 is already implemented by D-053" → `bodyRegion()` reserves
    `chromeEyebrowH + chromeBandGap` (top) and `chromeFooterH + chromeBandGap`
    (bottom), using the chrome composer's own constants; the body stack lays out
    inside it.
  - "the overflow hole is closed by R11.3" → `clampToSafeArea` (safe area =
    `bodyRegion`) caps containers to the reserved region under hostile content.
  - "the open gap is the acceptance test" → this plan ships the disjointness +
    chrome-off byte-identical assertions.

## 5. Findings I'm departing from

*"none"*

## 6. Decisions referenced

- `D-053` — slide chrome + the `bodyRegion` reservation (the mechanism).
- `D-083` — the safe-area clamp that makes the reservation hold under overflow.
- `D-084` — **new** — records R11.4 closed by D-053 + D-083 with the acceptance.

## 7. Architecture

No production change. The reserved region:

```text
bodyRegion() (chrome on):
  top    = bodyMargin + chromeEyebrowH + chromeBandGap   (below the eyebrow band + rule)
  bottom = bodyMargin + chromeFooterH  + chromeBandGap   (above the footer band)
  → disjoint from the bands chrome.go draws (chromeBandGap > chromeRuleH)
clampToSafeArea (R11.3) caps containers to this region under overflow.
```

## 8. Files added or changed

```text
scene/render_chrome_region_test.go   # NEW — R11.4 disjointness + chrome-off byte-identical + clamp-above-footer
scripts/smoke/phase-52.sh            # NEW — phase smoke
docs/research/35-content-region-reserves-chrome.md  # NEW — brief 35
docs/research/INDEX.md               # CHANGED — registers brief 35
docs/plans/phase-52-content-region-reserves-chrome.md  # NEW — this plan
docs/plans/README.md                 # CHANGED — Wave 11 Phase 52
docs/decisions.md                    # CHANGED — adds D-084
```

No public API change, no new token, no user-facing surface change → no skill /
docs-site update required.

## 9. Public API surface

None.

## 10. Risks

- **R1 — a vacuous disjointness assertion.** **Mitigation:** the bands are recomputed
  from the chrome constants (`chromeEyebrowH`/`chromeFooterH`/`chromeRuleH`), not
  from `bodyRegion`, so the test would fail if the reservation drifted below a band.

## 11. Acceptance criteria

1. On a chromed slide, `bodyRegion().Y >=` the eyebrow band bottom and
   `bodyRegion().Bottom() <=` the footer band top (no intersection).
2. With chrome off, `bodyRegion()` is the plain margin box (byte-identical).
3. A container handed an overflowing box on a chromed slide is clamped above the
   footer band (R11.4 × R11.3).
4. `go test -race ./scene/...` passes; `make coverage` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | unchanged; test-only addition |

## 13. Smoke check

`scripts/smoke/phase-52.sh` runs the three acceptance tests. All `OK`, `FAIL = 0`.

## 14. Tests

- **Unit:** `scene` white-box (`render_chrome_region_test.go`).
- **Round-trip golden / Integration / Fuzz / Benchmark:** n/a.

## 15. Vocabulary added

- *(none — "safe area" / "section eyebrow" / "slide chrome" already defined.)*

## 16. Plan deviations encountered during implementation

- *(none)*

## 17. Sign-off

- [x] All acceptance criteria pass.
- [x] `make coverage` clean for touched packages.
- [x] `scripts/smoke/phase-52.sh` reports `OK ≥ 3` and `FAIL = 0`.
- [x] Prior phases' smoke scripts still pass.
- [x] Glossary updated (n/a — no new term).
- [x] Decision entries added (D-084).
- [x] Docs site updated (n/a — no surface change).
- [x] Affected agent skill(s) updated (n/a).
