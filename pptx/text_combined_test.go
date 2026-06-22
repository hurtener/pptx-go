package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestRun_CombinedTypeAttributes proves several Wave-9 run-level type attributes
// (Tracking D-060, Case D-062, Bold, Color) coexist on a single a:rPr and round
// trip together — guarding against a regression that drops one when another is
// present (they share toProps's emit path and `set` flag).
func TestRun_CombinedTypeAttributes(t *testing.T) {
	data := trackDeck(t, pptx.RunStyle{
		TypeRole: pptx.TypeBody,
		Bold:     true,
		Tracking: ptr(1.5),
		Case:     ptrCase(pptx.CaseUpper),
		Color:    pptx.RGB("FF0000"),
	})

	xml := trackSlideXML(t, data)
	for _, want := range []string{`spc="150"`, `cap="all"`, `b="1"`, `val="FF0000"`} {
		if !strings.Contains(xml, want) {
			t.Errorf("emitted run missing %q (combined attributes dropped one):\n%s", want, xml)
		}
	}

	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	tf, ok := re.Slides()[0].Shapes()[0].TextFrame()
	if !ok {
		t.Fatal("reopened shape has no text frame")
	}
	r := tf.Paragraphs()[0].Runs()[0]
	if got := r.Tracking(); got != 1.5 {
		t.Errorf("round-trip Tracking() = %v, want 1.5", got)
	}
	if got := r.Case(); got != pptx.CaseUpper {
		t.Errorf("round-trip Case() = %v, want CaseUpper", got)
	}
	if !r.Bold() {
		t.Error("round-trip Bold() = false, want true")
	}
	if _, ok := r.Color(); !ok {
		t.Error("round-trip Color() missing")
	}
}

// TestRun_WeightNotEmittedAsAttribute guards XTextProperties.Weight's xml:"-"
// tag (D-068): the in-memory weight must never leak onto the emitted a:rPr (a
// stray weight= attribute would be invalid and the reader would silently swallow
// it — the same local-name-masking class that hid the bare <p:font> bug).
func TestRun_WeightNotEmittedAsAttribute(t *testing.T) {
	theme := pptx.NewTheme()
	spec := theme.Typography[pptx.TypeBody]
	spec.Weight = 500
	theme.Typography[pptx.TypeBody] = spec

	p := pptx.New(pptx.WithTheme(theme))
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("x", pptx.RunStyle{TypeRole: pptx.TypeBody})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if xml := trackSlideXML(t, data); strings.Contains(xml, "weight=") {
		t.Errorf("weight= leaked onto emitted run XML (xml:\"-\" regressed):\n%s", xml)
	}
}

// TestRun_RunOverrideBeatsRole proves a run-level Tracking/Case override (the
// *float64 / *TextCase nil=inherit design) wins over the role's value.
func TestRun_RunOverrideBeatsRole(t *testing.T) {
	// Role sets tracking 5.0 / upper; the run overrides to 1.0 / none.
	theme := pptx.NewTheme()
	spec := theme.Typography[pptx.TypeBody]
	spec.Tracking, spec.Case = 5.0, pptx.CaseUpper
	theme.Typography[pptx.TypeBody] = spec

	p := pptx.New(pptx.WithTheme(theme))
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("x", pptx.RunStyle{
		TypeRole: pptx.TypeBody, Tracking: ptr(1.0), Case: ptrCase(pptx.CaseNone),
	})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	xml := trackSlideXML(t, data)
	if !strings.Contains(xml, `spc="100"`) {
		t.Errorf("run-level tracking override (1.0 → spc=100) did not win over role 5.0:\n%s", xml)
	}
	if strings.Contains(xml, "cap=") {
		t.Errorf("run-level CaseNone override did not suppress the role's upper:\n%s", xml)
	}
}

// TestRun_DisplayFaceWithTrackingAndCase proves the Phase-33 display face
// (D-063) coexists with role-level tracking + case on one run — the canonical
// tracked-caps display eyebrow that motivated the Wave-9 cluster.
func TestRun_DisplayFaceWithTrackingAndCase(t *testing.T) {
	theme := pptx.NewTheme(pptx.WithDisplayFont("Playfair Display"))
	spec := theme.Typography[pptx.TypeDisplay]
	spec.Tracking, spec.Case = 2.0, pptx.CaseUpper
	theme.Typography[pptx.TypeDisplay] = spec

	p := pptx.New(pptx.WithTheme(theme))
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 0, Y: 0, W: pptx.Slide16x9Width, H: pptx.Slide16x9Height})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("vision", pptx.RunStyle{TypeRole: pptx.TypeDisplay})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	xml := trackSlideXML(t, data)
	for _, want := range []string{`typeface="Playfair Display"`, `spc="200"`, `cap="all"`} {
		if !strings.Contains(xml, want) {
			t.Errorf("display+tracking+case missing %q:\n%s", want, xml)
		}
	}
}
