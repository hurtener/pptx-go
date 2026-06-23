# Brief 52 — prim-attribution-lockup

**Subsystem:** scene — Layer 2 renderer (new IR leaf node, asset-bearing)
**Authored:** 2026-06-23
**Motivating phase:** Phase 69 — prim-attribution-lockup (R12.9, LOW · engine)

## 1. Question

A branded deck places a "POWERED BY [logo] CLEAR TECH" / "in partnership with" lockup — a
small caption paired with a partner logo as one inline, centerable unit — on its cover and
closing. The recreation dropped the logo and rendered the caption as plain text. `Image` and
`Chip` exist separately but nothing composes a small logo with a caption as a single lockup.

## 2. Prior art surveyed

- **`scene/render_image.go renderImage`** — the asset path: `r.resolve(AssetID)` →
  `ps.AddImage(box)`, `stats.Assets++`. An unresolved asset degrades to a warning
  (warn-don't-fail). The lockup's asset side reuses this exactly.
- **`scene/render_button.go` / icon glyphs** — the media-free `r.cfg.icons.Lookup` →
  `ps.AddIcon` path for the icon variant.
- **`scene/render.go nodeUsesAssets`** — `Image`/`Chart`/`CodeBlock` return true so they
  render in the serial pass for deterministic media part numbering (RFC §10.1). A lockup
  returns true **iff** it carries an `AssetID`; the icon variant is media-free → parallel.
- **`Decoration` policy** — `{HasAsset: true}` with `Image:false`: an asset-kind node whose
  asset renders as a pic at render time while a non-asset variant is native. The lockup
  mirrors this (asset → pic, icon → custGeom).
- **§7** — no pixel-dimension parsing; aspect is the caller's problem at display. The logo
  box is height-bounded; without dimensions a square box is the honest default.

## 3. Findings

- **A leaf node with an asset OR an icon.** `Lockup{Caption string; AssetID AssetID; Icon
  string; AssetSide AssetSide; MaxHeight EMU; Align HAlign}`. Validation requires exactly
  one of `AssetID` / `Icon` (a mark), the `AssetSide` in range, and `MaxHeight >= 0`. The
  icon name flows through `walkIconRefs`; the asset resolves at render (warn-don't-fail).
- **`nodeUsesAssets` is `AssetID != ""`.** So an asset lockup renders serially (deterministic
  part numbering) and an icon lockup stays parallel-safe. `policy = {HasAsset: true}` (it
  carries an `AssetID` field — the policy_test invariant) with `Image:false` (conditional,
  like `Decoration`).
- **Layout: a centered inline group.** `[caption | gap | logo]` (`LeadCaption`) or `[logo |
  gap | caption]` (`TrailCaption`), the whole group centered/aligned within the box and
  vertically centered. The caption is `TypeCaption` muted; the logo box is height-bounded by
  `MaxHeight` (a pinned default when 0), square (no pixel aspect — §7), filled via `AddImage`
  (asset) or `AddIcon` (icon).
- **Pinned metrics; tokens for the caption color.** The gap, default height, and pad are
  pinned EMU; the caption uses `TextMuted`.

## 4. Recommendation

Add `KindLockup` + a `Lockup` leaf node and a `scene/render_lockup.go` composer: measure the
caption + logo widths, compose a centered horizontal group with a small gap, height-bound the
logo, and render the logo via `AddImage` (asset) or `AddIcon` (icon). Full new-node wiring
(policy `{HasAsset:true}`, validate the asset-XOR-icon + side + height, `renderNode` dispatch
+ `preferredHeight` + `isFlexible` false + `nodeUsesAssets`(`AssetID != ""`), `walkIconRefs`
for the icon, catalog 27 → 28, integration kind-range loop → `KindLockup`). Cover both the
icon (media-free) and asset (serial part-numbering) paths in tests. Additive ⇒ byte-identical
when unused.
