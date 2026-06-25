package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// TestFormatNumber is R14.13 acceptance: locale/currency/percent/compact formats
// render deterministically, fixing the "$4,000+" wrap and covering de-DE + percent.
func TestFormatNumber(t *testing.T) {
	usd := scene.NumberFormat{GroupSep: ",", CurrencySymbol: "$"}
	deDE := scene.NumberFormat{GroupSep: ".", DecimalSep: ","}
	cases := []struct {
		name string
		v    float64
		f    scene.NumberFormat
		want string
	}{
		{"usd plus suffix", 4000, scene.NumberFormat{GroupSep: ",", CurrencySymbol: "$", Suffix: "+"}, "$4,000+"},
		{"usd plain", 1200, usd, "$1,200"},
		{"percent", 0.92, scene.NumberFormat{Percent: true}, "92%"},
		{"de-DE grouping", 4000, deDE, "4.000"},
		{"compact millions", 1200000, scene.NumberFormat{Compact: true}, "1.2M"},
		{"compact billions 2dp", 3450000000, scene.NumberFormat{Compact: true, Decimals: 2}, "3.45B"},
		{"euro after", 4000, scene.NumberFormat{GroupSep: ".", CurrencySymbol: "€", SymbolAfter: true}, "4.000 €"},
		{"decimals", 3.14159, scene.NumberFormat{Decimals: 2}, "3.14"},
		{"negative usd", -1500, scene.NumberFormat{GroupSep: ",", CurrencySymbol: "$"}, "-$1,500"},
		{"zero format identity", 4000, scene.NumberFormat{}, "4000"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := scene.FormatNumber(c.v, c.f); got != c.want {
				t.Errorf("FormatNumber(%g) = %q, want %q", c.v, got, c.want)
			}
		})
	}
}

// TestStat_NumberPath verifies a Stat with Number+Format renders the formatted
// value (the "$4,000+" regression fix) and that a raw-Value Stat is unchanged.
func TestStat_NumberPath(t *testing.T) {
	n := 4000.0
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "price",
		Nodes: []scene.SlideNode{scene.Stat{
			Number:  &n,
			Format:  &scene.NumberFormat{GroupSep: ",", CurrencySymbol: "$", Suffix: "+"},
			Label:   "per month",
			AutoFit: true,
		}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("stat number path: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "$4,000+") {
		t.Errorf("stat number path did not render formatted value:\n%s", slide)
	}
}
