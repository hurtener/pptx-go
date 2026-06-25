# Phase 85 — quote / testimonial enrichment

**Subsystem:** `scene` (Quote node + renderer)
**RFC sections:** §11.1 (leaf node), §11 (asset path), §10.1 (backward-compat), §10.2 (degrade), §7.1 (tokens)
**Deps:** D-114 (rounded image); brief 68.
**Status:** Done

---

## 1. Goal

Extend the `Quote` node additively into a designed testimonial — an oversized
quotation mark, a rounded author avatar, structured name/role/company
attribution, and a customer logo — laid out as one balanced unit, byte-identical
when unused.

## 2. Why now

Wave 14 coverage classes (`docs/plans/README.md`); social-proof testimonial
slides are core to sales/investor decks and the minimal `Quote` reads as plain
text. Engine req R14.5 (HIGH · engine per D-059).

## 3. RFC sections implemented

- `RFC §11.1` — the `Quote` leaf gains testimonial treatment.
- `RFC §11` — avatar/logo compose the existing `AssetResolver` + media path.
- `RFC §10.1` — a Text+Attribution quote is byte-identical.
- `RFC §10.2` — a missing avatar/logo warns and is omitted; no panic.
- `RFC §7.1` — the quote mark color is a theme token (P2).

## 4. Brief findings incorporated

- `docs/research/68-quote-testimonial.md` — *"additive fields, a branched
  renderer"* → `enriched()` predicate; plain path verbatim.
- `68` — *"avatar/logo make the Quote asset-bearing"* → `nodeUsesAssets(Quote)`.
- `68` — *"policy stays HasAsset:false"* → fields named `AvatarAssetID`/
  `LogoAssetID`, not `AssetID`; `KindQuote` unchanged; catalog stays 29.
- `68` — *"avatar reuses SetCornerRadius(RadiusFull)"*; *"the mark is a font glyph,
  not a checkbox"* (low-alpha, behind the text).
- `68` — *"preferredHeight reserves the strip"*.

## 5. Findings I'm departing from

- **Logo cover-crop** via `WithImageFill` is deferred (logos are usually
  pre-trimmed; the height-bounded stretch is adequate). §4.3.
- **A side avatar-left layout** is deferred; the bottom attribution strip
  satisfies the acceptance (one balanced unit, 1-line and multi-line quotes).

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-114` — the rounded-image method (avatar).
- `D-026` — the engine renders the testimonial; the soul supplies the assets/text.
- `D-120` (new) — files the Quote testimonial enrichment.

## 7. Architecture

`Quote` gains `Mark bool`, `AvatarAssetID AssetID`, `AttributionName/Role/Company
string`, `LogoAssetID AssetID`, and an `enriched()` predicate. `renderQuote` runs
the existing plain path when `!enriched()` (byte-identical) and `renderTestimonial`
otherwise: an optional oversized `“` (TypeDisplay, low-alpha accent, drawn first),
the quote text (TypeH3, anchored over the lower half of the mark), then a bottom
attribution strip `[rounded avatar | Name(bold) / Role · Company(muted) | logo]`.
Avatar/logo resolve via `r.resolve` (warn + omit on miss); the avatar is
`SetCornerRadius(RadiusFull)`. `nodeUsesAssets(Quote)` is true when an avatar/logo
is set; `preferredHeight` adds the strip (+ half the mark) to the slot estimate.

```text
Quote{Text, Mark, AvatarAssetID, AttributionName/Role/Company, LogoAssetID}
  → “ (behind) + quote text + [avatar | name/role·company | logo]
Quote{Text, Attribution}  (no enrichment) → plain centered text (byte-identical)
```

## 8. Files added or changed

```text
scene/nodes.go                       # CHANGED — Quote enrichment fields + enriched()
scene/render_leaves.go               # CHANGED — renderQuote branch + renderTestimonial(+Strip)
scene/render.go                      # CHANGED — Quote preferredHeight + nodeUsesAssets + dispatch slideID
scene/render_testimonial_test.go     # NEW — testimonial, plain byte-identical, missing-warns, determinism
scene/render_adversarial_test.go     # CHANGED — an enriched-quote slide in the torture fixture
scripts/smoke/phase-85.sh            # NEW — phase smoke
docs/research/68-quote-testimonial.md  # NEW — brief
docs/research/INDEX.md               # CHANGED — registers brief 68
docs/plans/phase-85-quote-testimonial.md  # NEW — this plan
docs/plans/README.md                 # CHANGED — Phase 85 detail
docs/design/THEME.md                 # CHANGED — quote-mark color mechanism note
docs/glossary.md                     # CHANGED — testimonial term
docs/decisions.md                    # CHANGED — adds D-120
docs/site/catalog/text-leaves.md     # CHANGED — Quote testimonial fields
skills/compose-a-scene/SKILL.md      # CHANGED — Quote enrichment
```

## 9. Public API surface

```go
// scene — Quote gains:
//   Mark bool; AvatarAssetID AssetID; AttributionName, AttributionRole, AttributionCompany string; LogoAssetID AssetID
```

Additive; no break.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** `!enriched()` runs the unchanged plain
  path; a test asserts a plain quote emits no pic.
- **R2 — determinism.** **Mitigation:** an avatar/logo quote is serial; a
  1-vs-8-worker test asserts byte-identity.
- **R3 — off-canvas.** **Mitigation:** the strip is reserved in `preferredHeight`;
  the adversarial enriched-quote slide asserts on-canvas.

## 11. Acceptance criteria

1. A testimonial with avatar + name/role/company + logo + mark renders as one unit
   (2 pics, rounded avatar, role · company), conformant, no warnings.
2. A Quote with only Text+Attribution is byte-identical (no pic, no mark).
3. An unresolvable avatar/logo warns and is omitted (the quote still renders).
4. The testimonial is worker-count deterministic.
5. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | testimonial composer |

## 13. Smoke check

`scripts/smoke/phase-85.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `Quote.Mark` / `AvatarAssetID` / structured attribution / `enriched()` /
   `renderTestimonial` / the serial-asset rule present.
3. `OK:` testimonial / plain-byte-identical / missing-warns / determinism tests.

## 14. Tests

- **Black-box (`scene_test`):** a full testimonial renders (2 pics, rounded
  avatar, role·company, conformant, warning-free); a plain quote is byte-identical
  and emits no pic; a missing avatar warns; the testimonial is worker-count
  deterministic.
- **Adversarial:** an enriched quote (mark + structured attribution, long text).
- **Integration / Fuzz:** no (no new node; `Quote` is already integration-covered).
