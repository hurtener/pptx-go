package pptx_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

func trackSlideXML(t *testing.T, data []byte) string {
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

func trackDeck(t *testing.T, rs pptx.RunStyle) []byte {
	t.Helper()
	p := pptx.New()
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 914400, W: 5000000, H: 1000000})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Tracked", rs)
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

func ptr(f float64) *float64 { return &f }

// TestTracking_EmitsSpc is acceptance criteria 1 & 3: a per-run tracking override
// emits a:rPr/@spc in signed 1/100 pt (positive and negative).
func TestTracking_EmitsSpc(t *testing.T) {
	if xml := trackSlideXML(t, trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody, Tracking: ptr(1.5)})); !strings.Contains(xml, `spc="150"`) {
		t.Errorf("tracking 1.5pt should emit spc=\"150\"; xml had no such attr")
	}
	if xml := trackSlideXML(t, trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeDisplay, Tracking: ptr(-0.5)})); !strings.Contains(xml, `spc="-50"`) {
		t.Errorf("tracking -0.5pt should emit spc=\"-50\"; xml had no such attr")
	}
}

// TestTracking_ZeroByteIdentical is acceptance criterion 2: no tracking emits no
// spc and reproduces the untracked deck byte-for-byte.
func TestTracking_ZeroByteIdentical(t *testing.T) {
	bare := trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody})
	nilOverride := trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody, Tracking: nil})
	if !bytes.Equal(bare, nilOverride) {
		t.Errorf("nil tracking is not byte-identical to no tracking field (%d vs %d)", len(nilOverride), len(bare))
	}
	if strings.Contains(trackSlideXML(t, bare), "spc=") {
		t.Error("untracked run should emit no spc attribute")
	}
	// An explicit zero override also emits nothing.
	if strings.Contains(trackSlideXML(t, trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody, Tracking: ptr(0)})), "spc=") {
		t.Error("explicit zero tracking should emit no spc attribute")
	}
}

// TestTracking_RoundTrips is acceptance criterion 4 (G6): a tracked run reopens
// with the same Tracking().
func TestTracking_RoundTrips(t *testing.T) {
	data := trackDeck(t, pptx.RunStyle{TypeRole: pptx.TypeBody, Tracking: ptr(2.0)})
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
	if got := runs[0].Tracking(); got != 2.0 {
		t.Errorf("reopened Tracking() = %v, want 2.0", got)
	}
}

// TestTracking_RoleLevel verifies the role-level FontSpec.Tracking path: a theme
// whose body role carries tracking emits spc on a plain (override-free) run.
func TestTracking_RoleLevel(t *testing.T) {
	th := pptx.DefaultTheme().Clone()
	body := th.ResolveType(pptx.TypeBody)
	body.Tracking = 1.0
	th.Typography[pptx.TypeBody] = body

	p := pptx.New(pptx.WithTheme(th))
	s := p.AddSlide()
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 914400, W: 5000000, H: 1000000})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Role", pptx.RunStyle{TypeRole: pptx.TypeBody})
	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if !strings.Contains(trackSlideXML(t, data), `spc="100"`) {
		t.Error("role-level FontSpec.Tracking=1.0 should emit spc=\"100\" on an override-free run")
	}
}
