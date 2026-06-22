package pptx_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Per-run FontScale (Phase 43, R10.5, D-074): a RunStyle.FontScale multiplier
// shrinks the resolved type-role size, emitted as a:rPr/@sz and round-tripping
// via Run.FontSize. Zero/unset is byte-identical.

func scaleSlideXML(t *testing.T, data []byte) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rc, _ := f.Open()
			defer func() { _ = rc.Close() }()
			b, _ := io.ReadAll(rc)
			return string(b)
		}
	}
	t.Fatal("slide1.xml not found")
	return ""
}

func scaleDeck(t *testing.T, rs pptx.RunStyle) []byte {
	t.Helper()
	p := pptx.New()
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 914400, W: 5000000, H: 1000000})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Scaled", rs)
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestRunFontScale_EmitsReducedSz: a 0.6 FontScale on a display run emits the
// reduced a:rPr/@sz (60% of the resolved role size, in 1/100 pt).
func TestRunFontScale_EmitsReducedSz(t *testing.T) {
	base := pptx.DefaultTheme().ResolveType(pptx.TypeDisplay).Size // points
	wantSz := int(base * 0.6 * 100)
	xml := scaleSlideXML(t, scaleDeck(t, pptx.RunStyle{TypeRole: pptx.TypeDisplay, FontScale: 0.6}))
	if !strings.Contains(xml, fmt.Sprintf(`sz="%d"`, wantSz)) {
		t.Errorf("FontScale 0.6 on TypeDisplay should emit sz=%q; xml:\n%s", fmt.Sprintf(`sz="%d"`, wantSz), xml)
	}
	// And it must be smaller than the unscaled size.
	full := int(base * 100)
	if wantSz >= full {
		t.Fatalf("scaled sz %d should be < full %d", wantSz, full)
	}
}

// TestRunFontScale_ZeroByteIdentical: FontScale 0 (and 1) leaves the role size
// unchanged, byte-identical to a run with no FontScale.
func TestRunFontScale_ZeroByteIdentical(t *testing.T) {
	bare := scaleDeck(t, pptx.RunStyle{TypeRole: pptx.TypeDisplay})
	zero := scaleDeck(t, pptx.RunStyle{TypeRole: pptx.TypeDisplay, FontScale: 0})
	if !bytes.Equal(bare, zero) {
		t.Errorf("FontScale 0 is not byte-identical to no FontScale (%d vs %d)", len(zero), len(bare))
	}
	one := scaleDeck(t, pptx.RunStyle{TypeRole: pptx.TypeDisplay, FontScale: 1})
	if !bytes.Equal(bare, one) {
		t.Errorf("FontScale 1 is not byte-identical to no FontScale (%d vs %d)", len(one), len(bare))
	}
}

// TestRunFontScale_RoundTrip (G6): a scaled run reopens with the reduced
// Run.FontSize.
func TestRunFontScale_RoundTrip(t *testing.T) {
	base := pptx.DefaultTheme().ResolveType(pptx.TypeDisplay).Size
	want := float64(int(base*0.6*100)) / 100.0
	data := scaleDeck(t, pptx.RunStyle{TypeRole: pptx.TypeDisplay, FontScale: 0.6})
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	tf, ok := re.Slides()[0].Shapes()[0].TextFrame()
	if !ok {
		t.Fatal("reopened shape has no text frame")
	}
	runs := tf.Paragraphs()[0].Runs()
	if len(runs) == 0 {
		t.Fatal("no runs reopened")
	}
	if got := runs[0].FontSize(); got != want {
		t.Errorf("reopened FontSize() = %v, want %v (0.6 × %v)", got, want, base)
	}
}
