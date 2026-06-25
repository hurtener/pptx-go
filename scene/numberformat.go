package scene

import (
	"strconv"
	"strings"
)

// NumberFormat is a deterministic number / currency / percent / locale format
// (R14.13, D-121). It is a caller-supplied mechanism (the soul's number token):
// the engine formats a numeric value with it, but never decides the format
// itself (D-026). FormatNumber applies it; a Stat carries an optional Number +
// NumberFormat that render through it (raw-string Stat.Value is unaffected).
//
// The zero value formats a number with no grouping, no decimals, and no affixes
// (e.g. 4000 → "4000") — an identity-ish format. A en-US currency sets
// GroupSep "," and CurrencySymbol "$"; a de-DE locale sets GroupSep "." and
// DecimalSep ",".
type NumberFormat struct {
	// Decimals is the fixed number of decimal places (0 = integer). For compact
	// notation, 0 is treated as 1 (so 1_200_000 → "1.2M", not "1M").
	Decimals int
	// GroupSep is the thousands separator ("," / "." / " "); "" = no grouping.
	GroupSep string
	// DecimalSep is the decimal point; "" defaults to ".".
	DecimalSep string
	// CurrencySymbol is prepended (or appended, see SymbolAfter); "" = none.
	CurrencySymbol string
	// SymbolAfter places the currency symbol after the number (e.g. "4.000 €").
	SymbolAfter bool
	// Percent multiplies the value by 100 and appends "%".
	Percent bool
	// Compact renders large magnitudes as K / M / B / T (e.g. 1_200_000 → "1.2M").
	Compact bool
	// CompactThreshold is the magnitude at/above which Compact applies; 0 = 1000.
	CompactThreshold float64
	// Prefix / Suffix are arbitrary affixes (e.g. a "+" suffix for "$4,000+").
	Prefix string
	Suffix string
}

var compactUnits = []struct {
	v float64
	s string
}{{1e12, "T"}, {1e9, "B"}, {1e6, "M"}, {1e3, "K"}}

// FormatNumber renders v per f, deterministically (stdlib only; rounding is
// strconv's round-half-to-even, so output is byte-stable). The layout is
// Prefix · sign · [symbol if !SymbolAfter] · body · [%] · [symbol if
// SymbolAfter] · Suffix.
func FormatNumber(v float64, f NumberFormat) string {
	neg := v < 0
	mag := v
	if neg {
		mag = -mag
	}
	if f.Percent {
		mag *= 100
	}

	decSep := f.DecimalSep
	if decSep == "" {
		decSep = "."
	}

	var body, compactLetter string
	if f.Compact {
		thresh := f.CompactThreshold
		if thresh == 0 {
			thresh = 1000
		}
		for _, u := range compactUnits {
			if mag >= u.v && mag >= thresh {
				mag /= u.v
				compactLetter = u.s
				break
			}
		}
	}
	decimals := f.Decimals
	if compactLetter != "" && decimals == 0 {
		decimals = 1
	}
	body = groupedDecimal(mag, decimals, f.GroupSep, decSep) + compactLetter

	var b strings.Builder
	b.WriteString(f.Prefix)
	if neg {
		b.WriteString("-")
	}
	if f.CurrencySymbol != "" && !f.SymbolAfter {
		b.WriteString(f.CurrencySymbol)
	}
	b.WriteString(body)
	if f.Percent {
		b.WriteString("%")
	}
	if f.CurrencySymbol != "" && f.SymbolAfter {
		b.WriteString(" " + f.CurrencySymbol)
	}
	b.WriteString(f.Suffix)
	return b.String()
}

// groupedDecimal formats mag (non-negative) with `decimals` fixed decimal places,
// grouping the integer part by `groupSep` (every 3 digits) and joining the
// fraction with `decSep`.
func groupedDecimal(mag float64, decimals int, groupSep, decSep string) string {
	s := strconv.FormatFloat(mag, 'f', decimals, 64)
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i+1:]
	}
	if groupSep != "" && len(intPart) > 3 {
		var g strings.Builder
		off := len(intPart) % 3
		if off > 0 {
			g.WriteString(intPart[:off])
		}
		for i := off; i < len(intPart); i += 3 {
			if g.Len() > 0 {
				g.WriteString(groupSep)
			}
			g.WriteString(intPart[i : i+3])
		}
		intPart = g.String()
	}
	if frac != "" {
		return intPart + decSep + frac
	}
	return intPart
}
