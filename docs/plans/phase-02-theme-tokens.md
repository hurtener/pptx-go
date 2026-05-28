# Phase 02 — Theme & token model + font embedding

**Subsystem:** pptx (theme + fonts)
**RFC sections:** §7 (§7.1–§7.6)
**Deps:** Phase 01.
**Status:** In progress

---

## 1. Goal

Make the `Theme` the single source of visual truth (P2): a Go-constructible,
template-loadable, OOXML-round-trippable theme whose semantic tokens
(color/text-color/type/space/radius/elevation) resolve deterministically to
OOXML values, plus the caller-driven font-embedding mechanism (D-019).

## 2. Why now

Wave 1 (`docs/plans/README.md §2`) builds the theme + builder spine. The
theme/token model is the prerequisite for the builder spine (Phase 03),
which takes tokens, and for scene rendering (Wave 2). It sits directly on
the `internal/ooxml/theme` codec relocated in Phase 01.

## 3. RFC sections implemented

- `RFC §7.1` — token taxonomy (the six role enums) and the `Theme` struct
  (`ColorPalette`, `Typography`, `Spacing`, `Radii`, `Elevations`).
- `RFC §7.3` — theme sources: inline Go (`NewTheme`) and `.pptx` template
  (`LoadTheme`). JSON/YAML theme files are V1.1+ (not this phase).
- `RFC §7.4` — token resolution semantics: `Theme.Resolve*` is
  deterministic; the lazy write-time binding into the builder lands with
  the builder spine (Phase 03 — see D-030).
- `RFC §7.5` — the default theme + `templates/_default-theme.pptx`.
- `RFC §7.6` — font embedding mechanism: `FontSource`, `WithFontSource`,
  `EmbedFont`, no auto-embed (D-019).
- Also lands `pptx/units.go` (Pt/Cm/In/Px, EMU) and `pptx/geom.go`
  (Box/Position/Size/Anchor/Inset) per the §3 layout — the units/geometry
  the theme spacing scale and the builder both need.

## 4. Brief findings incorporated

No informing brief — the theme/token taxonomy and font-embedding mechanism
are specified directly in RFC §7 (and D-003, D-012, D-019). The pengui-slides
v4 "design soul" model that motivated the taxonomy is captured in D-003.

## 5. Findings I'm departing from

none

## 6. Decisions referenced

- `D-003` — Theme first-class, tokens the default authoring path.
- `D-012` — Theme resolution is lazy at write time; `Color` is an interface
  (`tokenColor`/`literalColor`). **Honored, but the interface + builder
  constructors land in Phase 03 — see D-030.**
- `D-019` — Font-embedding mechanism, no auto-embed.
- `D-030` *(new, this PR)* — the `pptx.Color` interface and the
  `pptx.TokenColor()` / `pptx.RGB()` builder constructors are introduced in
  Phase 03 (the builder spine), not here. See §7.

## 7. Architecture

### D-030 — Color interface + token builder API land with the builder spine

RFC §7.2/D-012 make `pptx.Color` an interface with `tokenColor` (lazy,
write-time) and `literalColor` implementations, surfaced via
`pptx.TokenColor(role)` / `pptx.RGB(...)`. The inherited `pptx` package
already defines a concrete `Color` struct used by the upstream shape
builder. Replacing it with the interface is part of migrating the builder
to take tokens — i.e. the Phase 03 builder spine, which "migrates the
upstream pptx incrementally; new files supersede old ones; old API kept as
deprecated aliases." Doing it here would pull that migration forward
without the builder context that motivates its exact shape.

**Decision:** Phase 02 delivers the Theme model and a **deterministic
resolver** that returns concrete OOXML values
(`Theme.ResolveColor(ColorRole) string`, `ResolveSpace(SpaceRole) EMU`,
`ResolveType(TypeRole) FontSpec`, …). The lazy write-time `Color`
interface and `TokenColor`/`RGB` constructors land in Phase 03 alongside
the builder API that consumes them. The theme-swap guarantee is proven at
the resolver level here (one token, two themes → two values) and end-to-end
through the builder in Phase 03.

### Layout

```text
pptx/units.go         # EMU type; Pt/Cm/In/Px constructors; slide-size consts
pptx/geom.go          # Box, Position, Size, Anchor, Inset
pptx/theme.go         # role enums; Theme + ColorPalette/Typography/Spacing/
                       #   Radii/Elevations; NewTheme, DefaultTheme; LoadTheme
pptx/tokenresolve.go  # Theme.Resolve* (deterministic, concrete values)
pptx/fonts.go         # FontSource, WithFontSource, EmbedFont, errors, OS source
internal/ooxml/embeddings/   # font-embedding part writer (*.fntdata + refs)
internal/ooxml/theme/        # extend: pptx.Theme <-> XTheme mapping
templates/_default-theme.pptx# emitted default theme (scaffold; wired in P03)
docs/design/THEME.md         # token catalog (P2 taxonomy entry)
```

`pptx.Theme` is a pure Go domain object (no `encoding/xml`); the
`internal/ooxml/theme` codec (XTheme/ParseTheme/DefaultTheme, relocated in
Phase 01) does the wire mapping. `pptx` consumes domain objects only (P3).

## 8. Files added or changed

```text
pptx/units.go                 # NEW
pptx/geom.go                  # NEW
pptx/theme.go                 # NEW
pptx/tokenresolve.go          # NEW
pptx/fonts.go                 # NEW
pptx/*_test.go                # NEW — unit + round-trip + theme-swap tests
internal/ooxml/embeddings/embeddings.go        # NEW — font part writer
internal/ooxml/theme/mapping.go                # NEW — XTheme <-> domain map
templates/_default-theme.pptx # NEW — emitted default theme
docs/design/THEME.md          # NEW — token taxonomy catalog
scripts/smoke/phase-02.sh     # NEW
docs/decisions.md             # CHANGED — adds D-030
docs/glossary.md              # CHANGED — Typography/Spacing/Radii/Elevations/FontSource
internal/coveragecheck/coverage.json  # CHANGED — band new pptx files
```

## 9. Public API surface

```go
// pptx (units / geom)
type EMU int64
func Pt(float64) EMU; func Cm(float64) EMU; func In(float64) EMU; func Px(float64) EMU
type Box struct { X, Y, W, H EMU }
type Position struct { X, Y EMU }; type Size struct { W, H EMU }
type Anchor int; type Inset struct { Top, Right, Bottom, Left EMU }

// pptx (theme + tokens)
type ColorRole int;  type TextColorRole int; type TypeRole int
type SpaceRole int;  type RadiusRole int;    type ElevationRole int
type Theme struct { /* ColorPalette, Typography, Spacing, Radii, Elevations */ }
type FontSpec struct { Family string; Size float64; Weight int; Italic bool }
func NewTheme(opts ...ThemeOption) *Theme
func DefaultTheme() *Theme
func LoadTheme(path string) (*Theme, error)
func LoadThemeFromBytes(b []byte) (*Theme, error)
func (t *Theme) ResolveColor(ColorRole) string          // 6-hex RGB
func (t *Theme) ResolveTextColor(TextColorRole) string
func (t *Theme) ResolveType(TypeRole) FontSpec
func (t *Theme) ResolveSpace(SpaceRole) EMU
func (t *Theme) ResolveRadius(RadiusRole) EMU
func (t *Theme) ResolveElevation(ElevationRole) Elevation

// pptx (fonts — D-019)
type FontSource interface { Resolve(name, style string, weight int) ([]byte, error) }
func WithFontSource(FontSource) Option
func (p *Presentation) EmbedFont(name, style string, weight int) error
var ErrFontNotFound, ErrNoFontSource error
```

Deferred to Phase 03 (D-030): `type Color interface`, `TokenColor`, `RGB`.

## 10. Risks

- **R1 — `Color` name clash with the upstream struct.** *Mitigation:*
  Phase 02 does not declare `type Color`; the resolver returns concrete
  values. The interface lands in Phase 03 (D-030).
- **R2 — theme1.xml round-trip fidelity.** PowerPoint's theme XML is
  positional (clr1..12, major/minor fonts). *Mitigation:* build on the
  relocated `internal/ooxml/theme` codec (`ParseTheme`/`XTheme`) which
  already models it; map the role taxonomy onto the positional scheme; a
  golden round-trip test guards it.
- **R3 — Font-embedding part shape.** OOXML font embedding uses
  `fntdata` parts + presentation-level `embeddedFontLst` refs. *Mitigation:*
  vendor the spec shape into `internal/ooxml/embeddings`; round-trip test
  that the bytes survive `Open`.
- **R4 — EMU type duplication with `utils/emu.go`.** *Mitigation:* `pptx`
  owns its public `EMU` type + constructors; `utils` stays an internal
  helper. No public dependence on `utils`.

## 11. Acceptance criteria

1. A `Theme` constructed in Go round-trips through OOXML (theme → XTheme →
   bytes → XTheme → theme; role values preserved).
2. `LoadTheme` on a PowerPoint-emitted template yields a `Theme` whose
   `ResolveColor(ColorAccent)` returns the template's accent.
3. A token resolves to the same OOXML value across two resolutions with the
   same theme (determinism).
4. Theme swap: the same token resolved against two themes yields two
   different OOXML colors.
5. A presentation with a registered `FontSource` + an explicit `EmbedFont`
   ships the font bytes; a round-trip preserves the embed.
6. A presentation with no `EmbedFont` call ships no embedded fonts.
7. `make build`/`test`/`lint`/`coverage`/`preflight`/`check-mirror` green.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `pptx` (new theme/units/geom/fonts files) | 85% | new pptx builder code |
| `internal/ooxml/embeddings` | 85% | codec band |
| `internal/ooxml/theme` (new mapping) | 85% | codec band |

**Deviation (discovered):** the coverage gate is per-*package*, not
per-file, and `pptx` and `internal/ooxml/*` are inherited packages excluded
from the gate (D-029 staging) and from lint until the Phase 03 rewrite. So
the new theme/units/geom/fonts files cannot be banded independently of the
low-coverage inherited `pptx` files. They ship with thorough in-package
tests (resolution, swap, round-trip, embed) and are gofmt/vet-clean; they
come under the coverage gate + linter when Phase 03 consolidates `pptx`.
`require_configured` stays `false`. No `coverage.json` change this phase.

## 13. Smoke check

`scripts/smoke/phase-02.sh`:
1. `OK:` library builds CGo-free.
2. `OK:` `pptx.DefaultTheme()` resolves every role without panic (a tiny
   example binary or `go test -run` smoke).
3. `OK:` theme round-trip test passes.
4. `OK:` font-embed round-trip test passes.
5. `OK:` `templates/_default-theme.pptx` exists and opens via `internal/opc`.

## 14. Tests

- **Unit:** units conversions, geom, token resolution determinism + swap.
- **Round-trip golden:** Theme ↔ theme1.xml; font-embed bytes survive Open.
- **Integration** (`test/integration/`): theme load from a template +
  resolve (consumes the Phase 01 opc/ooxml seam).
- **Fuzz/Bench:** none new (theme resolve bench optional, deferred).

## 15. Vocabulary added

- `Typography` / `Spacing` / `Radii` / `Elevations` — the Theme's per-role
  value maps.
- `FontSpec` — resolved typography (family, size, weight, italic).
- `FontSource` — caller-injected font-bytes resolver (D-019).

## 16. Plan deviations encountered during implementation

- **Font-scheme codec bug fixed.** The inherited `XFontCollection` modelled
  Latin/EastAsia/Complex as a malformed `latin typeface,attr` tag, emitting
  invalid OOXML (`<majorFont latin:typeface="…">`) that PowerPoint ignored.
  Remodelled as nested `<latin typeface="…" panose="…"/>` faces (`XFontFace`)
  so theme fonts emit validly and round-trip. In-scope for the theme phase.
- **Per-package coverage gate** can't band the new pptx files independently
  of the inherited ones — see §12.
- **`SetFontSource` instead of `WithFontSource`.** `pptx.New()` has no option
  plumbing yet (it arrives with `pptx.WithFormat(...)` in Phase 03), so the
  V1 registration path is the `pres.SetFontSource(src)` setter. (D-030.)
- **`internal/ooxml/embeddings`** holds the font-part wire constants/naming;
  the `<p:embeddedFontLst>` XML types live in `internal/ooxml/presentation`
  (its own part family) so no cross-family import is introduced.
- **`templates/_default-theme.pptx`** is a minimal theme-only package
  (theme1.xml), regenerated by `go run ./_gen/gentheme`; the scaffold reads
  its theme1.xml. A full starter deck is not needed for the Phase 02 scope.

## 17. Sign-off

- [ ] All acceptance criteria pass.
- [ ] `make coverage` clean for touched packages.
- [ ] `scripts/smoke/phase-02.sh` reports `OK ≥ 5` and `FAIL = 0`.
- [ ] Prior phases' smoke scripts still pass.
- [ ] Glossary updated; `docs/design/THEME.md` token entry added.
- [ ] Decision entries added (`D-030`).
