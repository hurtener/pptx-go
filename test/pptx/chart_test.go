package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
)

// TestChartPlaceholder_EmitsAndRoundTrips is Phase 17 acceptance criterion 4:
// ChartPlaceholder emits a labeled bordered slot, the deck conforms, and the
// slot survives Open → re-save (G6).
func TestChartPlaceholder_EmitsAndRoundTrips(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	s.ChartPlaceholder(fxBox)

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("chart-placeholder deck failed conformance:\n%s", rep)
	}
	slide := readZipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{`prst="roundRect"`, `<a:prstDash val="dash"`, "<a:t>Chart</a:t>"} {
		if !strings.Contains(slide, want) {
			t.Errorf("chart placeholder missing %q in:\n%s", want, slide)
		}
	}
	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	resaved, err := reopened.WriteToBytes()
	if err != nil {
		t.Fatalf("re-save: %v", err)
	}
	if rs := readZipPart(t, resaved, "ppt/slides/slide1.xml"); !strings.Contains(rs, "<a:t>Chart</a:t>") || !strings.Contains(rs, `prst="roundRect"`) {
		t.Errorf("chart placeholder did not survive round-trip:\n%s", rs)
	}
}
