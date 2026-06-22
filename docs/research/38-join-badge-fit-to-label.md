# Brief 38 — join-badge-fit-to-label

**Subsystem:** scene — Layer 2 renderer (TwoColumn join)
**Authored:** 2026-06-22
**Motivating phase:** Phase 55 — join-badge-fit-to-label (R11.7, HIGH · engine)

## 1. Question

The TwoColumn inter-column join badge is drawn at a fixed `joinBadgeSz = In(0.62)`
ellipse; a label like "One agent" breaks mid-word into "One / age / nt" inside it
(recreation slide 8). How can the badge contain any short connector label intact and
centered, byte-identically for labels that already fit ("vs")?

## 2. Prior art surveyed

- **`scene/render_container.go renderColumnJoin`** — the `JoinBadge` case draws an
  `joinBadgeSz` ellipse centered on the column seam with a centered `TypeBodySmall`
  bold label.
- **`scene/metrics.go naturalWidth` / `fitScale`** + `pptx.RunStyle.FontScale` — the
  reusable one-line measure + shrink primitives (R10.5/R11.5).
- DECKARD R11.7 spec: compute the needed diameter from `naturalWidth(JoinLabel) +
  padding`; `joinBadgeSz = max(In(0.62), neededDiameter)` clamped to the inter-column
  gap, OR keep a max diameter and reduce the label size until one line; keep the
  ellipse + centered label; byte-identical for labels that already fit.

## 3. Findings

- **Grow to fit, cap, then shrink.** `needed = naturalWidth(label @ TypeBodySmall) +
  2·joinBadgePadX`. `badgeSz = clamp(needed, joinBadgeSz, joinBadgeMaxSz)`. If the
  label still does not fit at the cap, `labelScale = fitScale(natW, badgeSz −
  2·joinBadgePadX)` shrinks it to one line. The "clamp to the inter-column gap" the
  spec mentions is the wrong axis here — the badge deliberately *overlaps* both
  columns (the inter-column gap is only ~`SpaceMD`), so the cap is a pinned
  `joinBadgeMaxSz = In(1.5)` instead, keeping the badge a reasonable size while
  letting `fitScale` handle the overflow tail.
- **Byte-identical for "vs".** A 2-char label measures `~0.2"`; `needed ≈ 0.44" <
  In(0.62)`, so `badgeSz` stays at the base, and `fitScale(0.2, 0.62 − 0.24)` returns
  0 (no scale). The badge and the run are unchanged → byte-identical. The existing
  column-join goldens (which use "vs") pass unchanged.
- **Circle containment.** A single centered line spans the badge's horizontal
  diameter at its vertical center; `naturalWidth + 2·padX` with `padX = In(0.12)`
  leaves margin so the label stays inside the ellipse, and one `TypeBodySmall` line
  (~0.17") fits any diameter ≥ `In(0.62)` vertically.
- **Reuses the established primitives** (`naturalWidth`, `fitScale`, `FontScale`) — no
  new mechanism, consistent with R11.5's pill fit.

## 4. Recommendation

In the `JoinBadge` case, grow `badgeSz` to `clamp(naturalWidth + 2·padX, joinBadgeSz,
joinBadgeMaxSz)` and apply `fitScale` to the label run. Tests: a render-level check
that a multi-word label grows the ellipse diameter beyond the base while "vs" keeps
the base (byte-identical), and that a pathological label caps at `joinBadgeMaxSz`.
D-087 records the fit-to-label badge.

## 5. Open questions

- The arrow connector (`JoinArrow`) carries no label, so it is unaffected. Only
  `JoinBadge` fits.
