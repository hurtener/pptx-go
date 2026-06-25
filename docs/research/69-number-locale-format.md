# Brief 69 — Number / currency / percent / locale format (R14.13)

> Informs Phase 86 (Wave 14). Engine req R14.13
> (`DECKARD-PRODUCT-REQUIREMENTS.md`, HIGH · both — engine half; D-059). Also
> closes the engine atom R14.2 was waiting on.

## 1. Motivating phase

Pricing / metrics / KPI decks need locale-correct figures: thousands separators,
currency symbol + placement, percent, and compact notation. The engine has no
number-format concept, so big numbers are raw strings that wrap or mis-shrink
(the reference's "$4,000+" wraps to a stray "+" on its own line). Phase 86 adds a
deterministic `NumberFormat` + a typed numeric path on `Stat`.

## 2. Subsystem / files

- `scene/numberformat.go` (new) — the format type + `FormatNumber`.
- `scene/nodes.go` — `Stat` gains `Number`/`Format` + a `displayValue()` helper.
- `scene/render_stat.go` — renders `displayValue()`; `statValueFit` shrinks it.
- `scene/validate.go` — `Stat` accepts a `Number` when `Value` is empty.

## 3. Findings

- **A pure, stdlib-only formatter is deterministic.** `strconv.FormatFloat(v,
  'f', decimals, 64)` rounds half-to-even (byte-stable); manual integer-part
  grouping + a chosen decimal separator + currency/percent/compact affixes need no
  locale library (P4). Generalizes to any locale via `GroupSep`/`DecimalSep` (","
  / "." / " ") and any currency symbol/placement.
- **It is a mechanism, not taste (D-026).** The caller (soul) supplies the
  `NumberFormat`; the engine only applies it. So it is **not** a visual color/
  spacing token — no `THEME.md` token entry is required (the P2 token rule is for
  visual properties). It lives in `scene` (consumed by `Stat`; reusable by the
  future native data-marks of R14.8).
- **The typed path is additive.** `Stat.Number *float64` + `Stat.Format
  *NumberFormat`; when `Number` is non-nil, `displayValue()` returns
  `FormatNumber(*Number, *Format-or-zero)`; otherwise the raw `Value` string
  (byte-identical). `statValueFit` then keeps it on one line (the existing D-074
  shrink-to-fit), fixing the wrap regression.
- **Zero `NumberFormat` is an identity-ish format** (no grouping, no decimals, no
  affixes): `4000 → "4000"`. The en-US currency token sets `GroupSep:","` +
  `CurrencySymbol:"$"`; de-DE sets `GroupSep:"."` + `DecimalSep:","`.
- **Closes R14.2's engine atom.** R14.2 (brand-styled charts) named a
  `numberFormat` as part of its ChartStyle bundle. With `NumberFormat` shipped and
  the theme palette already exposed, R14.2's engine side is complete; the
  ChartStyle bundle + the rasterizer remain product (D-004 — V1 charts are
  caller-rasterized; there is no in-repo rasterizer to consume a style).

## 4. Recommendations

- `scene.NumberFormat{Decimals int; GroupSep, DecimalSep, CurrencySymbol string;
  SymbolAfter, Percent, Compact bool; CompactThreshold float64; Prefix, Suffix
  string}` + `FormatNumber(v float64, f NumberFormat) string`.
- `Stat.Number *float64` + `Stat.Format *NumberFormat` + `displayValue()`;
  `renderStat`/`statValueFit`/validate use it.
- Tests: a table-driven `FormatNumber` (usd+suffix, plain, percent, de-DE,
  compact M/B, euro-after, decimals, negative, zero-identity); a Stat numeric path
  renders the formatted value; an adversarial typed-number Stat. Glossary,
  compose-a-scene skill, docs/site text-leaves (Stat). D-121.

## 5. Open questions

- Native data-mark formatting (R14.8) → will consume the same `NumberFormat`.
- Per-run / prose number formatting → out of scope (the typed path is for `Stat`
  values; prose stays raw `RichText`).
