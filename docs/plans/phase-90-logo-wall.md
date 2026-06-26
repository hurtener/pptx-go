# Phase 90 — logo wall / customer grid

**Subsystem:** `scene` (new IR node)
**RFC sections:** §11, §12, §11 (asset path), §10.1, §10.2, §7
**Deps:** D-116 (duotone), D-114; brief 73.
**Status:** Done

## 1. Goal
Add a `LogoWall` scene node — an N-up grid of logo assets normalized to a common
optical size, optionally recolored to a uniform tone — for customer/partner walls.

## 2. Why now
Wave 14 coverage classes; social-proof logo walls appear in nearly every B2B deck
and the single-logo lockup doesn't cover a grid. Engine req R14.7 (MED · engine).

## 3–6. RFC / brief / decisions
RFC §11/§12 (new node), §11 (AssetResolver pic path), §10.1/§10.2 (absent / degrade),
§7 (header-dim read only). Brief 73. Decisions: D-059, D-116 (duotone seam reused
for tone), D-026, D-125 (new).

## 7. Architecture
`LogoWall{Logos []LogoEntry{AssetID, Alt}; Columns int; Tone LogoToneKind;
Caption string}`. `renderLogoWall`: an optional caption strip + a Columns×rows
grid; each logo is contained (not cropped) + centered via `containBox` (reads the
format-header dims through `imageDims`); `Tone` recolors via `SetDuotone`
(`LogoToneMono` = TextPrimary→Canvas, `LogoToneBrand` = Accent→Canvas). Asset-
bearing (`nodeUsesAssets:true`); a missing logo warns + skips. Policy
`HasAsset:false` (no single `AssetID` field).

## 8. Files
nodes.go (KindLogoWall + LogoWall/LogoEntry/LogoToneKind), policy.go, validate.go,
render.go (dispatch + preferredHeight + nodeUsesAssets), render_logo_wall.go (NEW),
scene_test.go (catalog 32), render_logo_wall_test.go (NEW), render_adversarial_test.go,
test/integration (kind-loop ..KindLogoWall + LogoWall on the button slide),
scripts/smoke/phase-90.sh, docs/research/73 + INDEX, this plan, README, THEME,
glossary, docs/site/catalog/visual-leaves.md, skills/compose-a-scene, D-125.

## 9. Public API
`type LogoWall struct {...}` + `LogoEntry` + `LogoToneKind` (None/Mono/Brand).
Additive new node; no break.

## 10–11. Risks / acceptance
Off-canvas (grid cells + contained logos within the box; adversarial wall);
determinism (serial asset render; 1-vs-8-worker test). Accept: a 12-logo mono wall
renders 12 contained pics + duotone (conformant); LogoToneNone = no recolor; a
missing logo warns + skips; worker-count deterministic; an empty wall fails Stage-1.

## 12–14. Coverage / smoke / tests
scene 80%. `scripts/smoke/phase-90.sh`. Black-box: 12-logo mono wall, none-tone,
missing-warns, determinism, empty-fails; adversarial wall; integration all-kinds.
