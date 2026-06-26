# Brief 73 — Logo wall / customer & partner grid (R14.7)

> Informs Phase 90 (Wave 14). Engine req R14.7
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, MED · engine; D-059).

## 1. Motivating phase
Social-proof logo walls (customers, investors, integrations) appear in nearly
every B2B deck. The attribution lockup covers ONE logo+caption; a GRID of many
evenly-sized logos, optionally desaturated to a uniform tone, is uncovered.

## 2. Findings
- **A new asset-bearing node.** `LogoWall{Logos []LogoEntry{AssetID, Alt};
  Columns; Tone; Caption}`. Each logo is **contained** (not cropped) + centered in
  its cell — `containBox` reads the format-header dims (`imageDims`, §7/D-046) and
  fits the largest box of the logo's aspect, so mixed-aspect logos read at a common
  optical size without distortion. Asset-bearing → `nodeUsesAssets:true` (serial
  determinism); a missing logo warns + is skipped (RFC §10.2).
- **Uniform tone reuses the duotone seam.** `LogoToneMono` = `SetDuotone(TextPrimary,
  Canvas)` (a brand-neutral two-tone), `LogoToneBrand` = `SetDuotone(Accent,
  Canvas)`; `LogoToneNone` = plain. A true `<a:grayscl>` is deferred — the mono
  duotone satisfies the "uniform tone so mixed logos cohere" intent (D-116
  duotone). Policy stays `HasAsset:false` (the node has no single `AssetID` field;
  it renders a pic per entry, not as one pic node).

## 3. Recommendations
- Node + composer (caption strip + Columns×rows grid of contained, tone-recolored
  logos). Full new-node wiring; `validate` requires ≥1 logo + a tone in range.
  Tests: a 12-logo mono wall (12 pics + duotone, conformant), none-tone (no
  duotone), missing-warns, determinism, empty-fails; an adversarial wall. D-125.

## 4. Open questions
- True grayscale (`<a:grayscl>`) + baseline-grid optical normalization beyond
  contain → V1.x (the mono duotone + contain satisfy the acceptance).
