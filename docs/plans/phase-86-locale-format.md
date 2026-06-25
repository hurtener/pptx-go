# Phase 86 — number / currency / percent / locale format

**Subsystem:** `scene` (NumberFormat + Stat)
**RFC sections:** §11.1 (Stat leaf), §10.1 (backward-compat), §7 (stdlib-only — P4)
**Deps:** D-074 (Stat shrink-to-fit); brief 69.
**Status:** Done

---

## 1. Goal

Add a deterministic `NumberFormat` + a typed numeric path on `Stat` so prices /
metrics / KPIs render with correct separators, currency, percent, and compact
notation — and stay on one line via the existing shrink-to-fit — fixing the
"$4,000+" wrap regression generally.

## 2. Why now

Wave 14 coverage classes (`docs/plans/README.md`); every pricing/financials deck
needs locale-correct figures, and big raw-string numbers wrap or mis-shrink.
Engine req R14.13 (HIGH · both per D-059). It also closes the engine atom R14.2
(brand-styled charts) was waiting on.

## 3. RFC sections implemented

- `RFC §11.1` — `Stat` gains a typed numeric value path.
- `RFC §10.1` — raw-string `Stat.Value` is byte-identical; zero `NumberFormat` is
  identity-ish.
- `RFC §7 / P4` — the formatter is stdlib-only (`strconv`), no locale library.

## 4. Brief findings incorporated

- `docs/research/69-number-locale-format.md` — *"a pure, stdlib-only formatter is
  deterministic"* → `strconv.FormatFloat` + manual grouping/affixes.
- `69` — *"a mechanism, not taste"* → no `THEME.md` token entry (not a visual
  property); the caller supplies the format.
- `69` — *"the typed path is additive"* → `Stat.Number`/`Format` + `displayValue()`;
  raw `Value` byte-identical.
- `69` — *"closes R14.2's engine atom"* → noted in D-121.

## 5. Findings I'm departing from

- none. (Native data-mark formatting and per-run prose formatting are out of
  scope — the typed path is for `Stat` values; R14.8 will reuse `NumberFormat`.)

## 6. Decisions referenced

- `D-059` — engine extension.
- `D-074` — the Stat shrink-to-fit the formatted value flows through.
- `D-026` — the format is a caller-supplied mechanism.
- `D-004` — V1 charts are caller-rasterized (why R14.2's remainder is product).
- `D-121` (new) — files `NumberFormat` + the Stat numeric path.

## 7. Architecture

`scene.NumberFormat{Decimals int; GroupSep, DecimalSep, CurrencySymbol string;
SymbolAfter, Percent, Compact bool; CompactThreshold float64; Prefix, Suffix
string}` + `FormatNumber(v, f)` (layout: Prefix · sign · [symbol-before] · body ·
[%] · [symbol-after] · Suffix; body = grouped fixed-decimal or compact K/M/B/T).
`Stat` gains `Number *float64` + `Format *NumberFormat` + `displayValue()`
(formats `Number` via `Format`-or-zero, else raw `Value`). `renderStat` and
`statValueFit` use `displayValue()`; `validate` accepts a `Number` when `Value`
is empty.

```text
Stat{Number: 4000, Format:{GroupSep:",", CurrencySymbol:"$", Suffix:"+"}, AutoFit}
  → "$4,000+" (one line, shrink-to-fit)
Stat{Value: "99.999%"}  (raw) → unchanged (byte-identical)
```

## 8. Files added or changed

```text
scene/numberformat.go            # NEW — NumberFormat + FormatNumber + groupedDecimal
scene/numberformat_test.go       # NEW — table-driven FormatNumber + Stat numeric path
scene/nodes.go                   # CHANGED — Stat.Number/Format + displayValue()
scene/render_stat.go             # CHANGED — render displayValue()
scene/validate.go                # CHANGED — Stat accepts Number when Value empty
scene/render_adversarial_test.go # CHANGED — a typed-number Stat in the strip
scripts/smoke/phase-86.sh        # NEW — phase smoke
docs/research/69-number-locale-format.md  # NEW — brief
docs/research/INDEX.md           # CHANGED — registers brief 69
docs/plans/phase-86-locale-format.md  # NEW — this plan
docs/plans/README.md             # CHANGED — Phase 86 detail
docs/glossary.md                 # CHANGED — NumberFormat term
docs/decisions.md                # CHANGED — adds D-121 (+ closes R14.2 engine atom)
docs/site/catalog/visual-leaves.md  # CHANGED — Stat numeric path
skills/compose-a-scene/SKILL.md  # CHANGED — Stat.Number/Format
```

## 9. Public API surface

```go
// scene
type NumberFormat struct { Decimals int; GroupSep, DecimalSep, CurrencySymbol string; SymbolAfter, Percent, Compact bool; CompactThreshold float64; Prefix, Suffix string }
func FormatNumber(v float64, f NumberFormat) string
// Stat gains: Number *float64; Format *NumberFormat
```

Additive; no break.

## 10. Risks

- **R1 — byte-identity.** **Mitigation:** `Number == nil` returns raw `Value`; the
  existing Stat tests + a zero-format identity case pin it.
- **R2 — determinism.** **Mitigation:** `strconv` round-half-to-even; a
  table-driven test pins every format path.

## 11. Acceptance criteria

1. 4000 + USD format → "$4,000+" on one line (no orphan "+", no wrap); 0.92 +
   percent → "92%"; 4000 + de-DE → "4.000"; deterministic.
2. A raw-string Stat is unchanged.
3. A Stat with `Number`+`Format` renders the formatted value.
4. `make lint`/`coverage`/`preflight`/`check-mirror` clean; `-race` clean.

## 12. Coverage targets

| Package | Target | Rationale |
|---|---|---|
| `scene` | 80% | formatter + Stat path |

## 13. Smoke check

`scripts/smoke/phase-86.sh`:

1. `OK:` library builds CGo-free.
2. `OK:` `NumberFormat` / `FormatNumber` / `Stat.Number` / `displayValue()` /
   `renderStat` wiring present.
3. `OK:` `FormatNumber` + Stat numeric-path tests.

## 14. Tests

- **Black-box (`scene_test`):** a table-driven `FormatNumber` (usd+suffix, plain,
  percent, de-DE, compact M/B, euro-after, decimals, negative, zero-identity); a
  Stat numeric path renders the formatted value.
- **Adversarial:** a typed-number Stat in the strip slide (one-line under AutoFit).
- **Integration / Fuzz:** no (no new node).
