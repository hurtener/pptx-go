package pptx_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func ptrCase(c pptx.TextCase) *pptx.TextCase { return &c }

// TestCase_EmitsCap is acceptance criterion 1: a case transform emits a:rPr/@cap
// (all / small).
func TestCase_EmitsCap(t *testing.T) {
	if xml := trackSlideXML(t, trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeCaption, Case: ptrCase(pptx.CaseUpper)})); !strings.Contains(xml, `cap="all"`) {
		t.Error("CaseUpper should emit cap=\"all\"")
	}
	if xml := trackSlideXML(t, trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody, Case: ptrCase(pptx.CaseSmallCaps)})); !strings.Contains(xml, `cap="small"`) {
		t.Error("CaseSmallCaps should emit cap=\"small\"")
	}
}

// TestCase_NoneByteIdentical is acceptance criterion 2: no case transform emits
// no cap and is byte-identical.
func TestCase_NoneByteIdentical(t *testing.T) {
	bare := trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody})
	nilOverride := trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody, Case: nil})
	if !bytes.Equal(bare, nilOverride) {
		t.Error("nil case is not byte-identical to no case field")
	}
	if strings.Contains(trackSlideXML(t, bare), "cap=") {
		t.Error("no case transform should emit no cap attribute")
	}
	if strings.Contains(trackSlideXML(t, trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody, Case: ptrCase(pptx.CaseNone)})), "cap=") {
		t.Error("explicit CaseNone should emit no cap attribute")
	}
}

// TestCase_RoundTripsAndPreservesText is acceptance criterion 3 (G6): the case
// transform round-trips via Run.Case() and the run TEXT stays original-case (the
// cap attribute is a display transform, not a text rewrite).
func TestCase_RoundTripsAndPreservesText(t *testing.T) {
	re, err := pptx.NewFromBytes(trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeCaption, Case: ptrCase(pptx.CaseUpper)}))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	tf, ok := re.Slides()[0].Shapes()[0].TextFrame()
	if !ok {
		t.Fatal("reopened shape has no text frame")
	}
	run := tf.Paragraphs()[0].Runs()[0]
	if run.Case() != pptx.CaseUpper {
		t.Errorf("reopened Case() = %v, want CaseUpper", run.Case())
	}
	if run.Text() != "Tracked" { // trackDeck authors the literal "Tracked"
		t.Errorf("run text = %q, want original-case \"Tracked\" (cap is a display attr)", run.Text())
	}
}

// TestCase_RoleLevel: a theme whose role sets Case emits cap on an override-free
// run (the role-level FontSpec.Case path).
func TestCase_RoleLevel(t *testing.T) {
	th := pptx.DefaultTheme().Clone()
	capRole := th.ResolveType(pptx.TypeCaption)
	capRole.Case = pptx.CaseUpper
	th.Typography[pptx.TypeCaption] = capRole

	p := pptx.New(pptx.WithTheme(th))
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 914400, W: 5000000, H: 1000000})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Eyebrow", pptx.RunStyle{TypeRole: pptx.TypeCaption})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if !strings.Contains(trackSlideXML(t, data), `cap="all"`) {
		t.Error("role-level FontSpec.Case=CaseUpper should emit cap=\"all\" on an override-free run")
	}
}
