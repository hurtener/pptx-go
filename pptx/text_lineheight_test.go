package pptx_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func lineHeightDeck(t *testing.T, lh float64) []byte {
	t.Helper()
	p := pptx.New()
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 914400, W: 5000000, H: 1500000})
	tf.AddParagraph(pptx.ParagraphOpts{LineHeight: lh}).AddRun("Line one", pptx.RunStyle{TypeRole: pptx.TypeBody})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestLineHeight_EmitsLnSpc is acceptance criterion 1: a non-single line-height
// emits a:pPr/a:lnSpc/a:spcPct in 1/1000 percent.
func TestLineHeight_EmitsLnSpc(t *testing.T) {
	xml := trackSlideXML(t, lineHeightDeck(t, 102))
	if !strings.Contains(xml, "<a:lnSpc>") || !strings.Contains(xml, `<a:spcPct val="102000"`) {
		t.Errorf("line-height 102 should emit <a:lnSpc><a:spcPct val=\"102000\"/>; not found")
	}
}

// TestLineHeight_ZeroAndSingleByteIdentical is acceptance criterion 2: 0 and 100
// emit nothing and are byte-identical (single is the default).
func TestLineHeight_ZeroAndSingleByteIdentical(t *testing.T) {
	bare := lineHeightDeck(t, 0)
	single := lineHeightDeck(t, 100)
	if !bytes.Equal(bare, single) {
		t.Errorf("LineHeight 100 (single) should be byte-identical to 0 (unset)")
	}
	if strings.Contains(trackSlideXML(t, bare), "lnSpc") {
		t.Error("unset line-height should emit no lnSpc")
	}
}

// TestLineHeight_RoundTrips is acceptance criterion 3 (G6): a paragraph's
// line-height reopens via Paragraph.LineHeight().
func TestLineHeight_RoundTrips(t *testing.T) {
	re, err := pptx.NewFromBytes(lineHeightDeck(t, 130))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	tf, ok := re.Slides()[0].Shapes()[0].TextFrame()
	if !ok {
		t.Fatal("reopened shape has no text frame")
	}
	if got := tf.Paragraphs()[0].LineHeight(); got != 130 {
		t.Errorf("reopened LineHeight() = %v, want 130", got)
	}
}
