package pptx_test

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestTextFrame_BodyPropsRoundTrip closes a round-trip read gap: a text frame's
// auto-fit, vertical anchor, and internal margins are authored, then verified
// through the read accessors after a write → Open cycle (G6). Margins use four
// distinct values so a top↔bottom or left↔right transposition is caught.
func TestTextFrame_BodyPropsRoundTrip(t *testing.T) {
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		f.AutoFit(pptx.AutoFitNormal)
		f.Anchor(pptx.AnchorMiddle)
		f.Margins(pptx.EMU(12700), pptx.EMU(25400), pptx.EMU(38100), pptx.EMU(50800))
		f.AddParagraph(pptx.ParagraphOpts{}).AddRun("body", pptx.RunStyle{TypeRole: pptx.TypeBody})
	})

	if got := tf.AutoFitMode(); got != pptx.AutoFitNormal {
		t.Errorf("AutoFitMode() = %v, want AutoFitNormal", got)
	}
	if got := tf.VerticalAnchor(); got != pptx.AnchorMiddle {
		t.Errorf("VerticalAnchor() = %v, want AnchorMiddle", got)
	}
	top, right, bottom, left := tf.MarginInsets()
	if top != 12700 || right != 25400 || bottom != 38100 || left != 50800 {
		t.Errorf("MarginInsets() = %d,%d,%d,%d, want 12700,25400,38100,50800 (top,right,bottom,left)",
			top, right, bottom, left)
	}
}

// TestTextFrame_AutoFitAndAnchorVariants round-trips every AutoFitMode and
// VerticalAnchor value through the read accessors, so the AutoFitShape and
// AnchorBottom branches are exercised (not just the Normal/Middle pair).
func TestTextFrame_AutoFitAndAnchorVariants(t *testing.T) {
	fits := []pptx.AutoFitMode{pptx.AutoFitNone, pptx.AutoFitNormal, pptx.AutoFitShape}
	for _, fit := range fits {
		fit := fit
		t.Run("autofit", func(t *testing.T) {
			tf := firstTextFrame(t, func(s *pptx.Slide) {
				f := s.AddTextFrame(fxBox)
				f.AutoFit(fit)
				f.AddParagraph(pptx.ParagraphOpts{}).AddRun("x", pptx.RunStyle{TypeRole: pptx.TypeBody})
			})
			if got := tf.AutoFitMode(); got != fit {
				t.Errorf("AutoFitMode() = %v, want %v", got, fit)
			}
		})
	}
	anchors := []pptx.TextAnchor{pptx.AnchorTop, pptx.AnchorMiddle, pptx.AnchorBottom}
	for _, a := range anchors {
		a := a
		t.Run("anchor", func(t *testing.T) {
			tf := firstTextFrame(t, func(s *pptx.Slide) {
				f := s.AddTextFrame(fxBox)
				f.Anchor(a)
				f.AddParagraph(pptx.ParagraphOpts{}).AddRun("x", pptx.RunStyle{TypeRole: pptx.TypeBody})
			})
			if got := tf.VerticalAnchor(); got != a {
				t.Errorf("VerticalAnchor() = %v, want %v", got, a)
			}
		})
	}
}

// TestTextFrame_BodyPropsDefaults confirms the read accessors report the OOXML
// defaults for a frame that set no body properties.
func TestTextFrame_BodyPropsDefaults(t *testing.T) {
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		f.AddParagraph(pptx.ParagraphOpts{}).AddRun("plain", pptx.RunStyle{TypeRole: pptx.TypeBody})
	})
	if got := tf.AutoFitMode(); got != pptx.AutoFitNone {
		t.Errorf("AutoFitMode() = %v, want AutoFitNone", got)
	}
	if got := tf.VerticalAnchor(); got != pptx.AnchorTop {
		t.Errorf("VerticalAnchor() = %v, want AnchorTop", got)
	}
	if top, r, b, l := tf.MarginInsets(); top|r|b|l != 0 {
		t.Errorf("MarginInsets() = %d,%d,%d,%d, want all 0", top, r, b, l)
	}
}

// TestTableRead_RowSpanRoundTrip closes a round-trip read gap: a genuine vertical
// merge (RowSpan > 1) survives the write → Open cycle. The existing table test
// merges from the last row, which clamps to span 1; this authors a tall-enough
// merge so the span is real.
func TestTableRead_RowSpanRoundTrip(t *testing.T) {
	shapes := reopenShapes(t, func(s *pptx.Slide) {
		tbl := s.AddTable(fxBox, 3, 2)
		tbl.Cell(0, 0).SetText("tall").MergeDown(2)
	})
	tbl, ok := shapes[0].Table()
	if !ok {
		t.Fatal("first shape Table() ok = false, want true")
	}
	if span := tbl.Cell(0, 0).RowSpan(); span != 2 {
		t.Errorf("cell(0,0) RowSpan() = %d, want 2", span)
	}
	if !tbl.Cell(1, 0).Covered() {
		t.Error("cell(1,0) Covered() = false, want true (covered by the vertical merge)")
	}
	// A cell below the span is its own anchor again.
	if span := tbl.Cell(2, 0).RowSpan(); span != 1 {
		t.Errorf("cell(2,0) RowSpan() = %d, want 1", span)
	}
}

// TestOpenStream_Parity asserts the streaming and in-memory read entry points
// reconstruct the same model — same slide count, same speaker notes, same shape
// count, and the same (empty) read warnings for a self-authored deck. Their
// parity was previously untested even though both flow through loadPresentationPart.
func TestOpenStream_Parity(t *testing.T) {
	p := pptx.New()
	s := p.AddSlide()
	s.AddRectangle(914400, 914400, 2743200, 1371600)
	s.SetSpeakerNotes("streaming parity notes")

	path := filepath.Join(t.TempDir(), "deck.pptx")
	if err := p.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	mem, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	str, err := pptx.OpenStream(path)
	if err != nil {
		t.Fatalf("OpenStream: %v", err)
	}

	if len(mem.Slides()) != len(str.Slides()) || len(str.Slides()) != 1 {
		t.Fatalf("slide counts differ: mem=%d stream=%d", len(mem.Slides()), len(str.Slides()))
	}
	if a, b := len(mem.Slides()[0].Shapes()), len(str.Slides()[0].Shapes()); a != b {
		t.Errorf("shape counts differ: mem=%d stream=%d", a, b)
	}
	memNotes := mem.Slides()[0].SpeakerNotes().Paragraphs()[0].Runs()[0].Text()
	strNotes := str.Slides()[0].SpeakerNotes().Paragraphs()[0].Runs()[0].Text()
	if memNotes != strNotes || strNotes != "streaming parity notes" {
		t.Errorf("notes differ: mem=%q stream=%q", memNotes, strNotes)
	}
	if len(mem.ReadWarnings()) != 0 || len(str.ReadWarnings()) != 0 {
		t.Errorf("authored deck should have no warnings: mem=%d stream=%d",
			len(mem.ReadWarnings()), len(str.ReadWarnings()))
	}
}

// TestOpenStream_WarningParity asserts the two read entry points surface the same
// ReadWarnings on a degraded deck — not just that both are empty on a clean one.
// A regression where OpenStream swallowed warnings would otherwise pass.
func TestOpenStream_WarningParity(t *testing.T) {
	p := pptx.New()
	p.AddSlide().AddRectangle(914400, 914400, 2743200, 1371600)
	clean, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	// Inject an unrecognized shape-tree element so both paths must warn.
	degraded := mutateZipPart(t, clean, "ppt/slides/slide1.xml", func(s string) string {
		return strings.Replace(s, "</p:spTree>", `<p:grpSp/></p:spTree>`, 1)
	})

	path := filepath.Join(t.TempDir(), "degraded.pptx")
	if err := os.WriteFile(path, degraded, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	mem, err := pptx.NewFromBytes(degraded)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	str, err := pptx.OpenStream(path)
	if err != nil {
		t.Fatalf("OpenStream: %v", err)
	}

	mw, sw := fmtWarnings(mem.ReadWarnings()), fmtWarnings(str.ReadWarnings())
	if len(mw) == 0 {
		t.Fatal("expected at least one warning on the degraded deck")
	}
	if strings.Join(mw, "|") != strings.Join(sw, "|") {
		t.Errorf("warning parity mismatch:\n NewFromBytes: %v\n OpenStream:   %v", mw, sw)
	}
}

// fmtWarnings renders warnings as sorted "kind|part|element" keys for comparison.
func fmtWarnings(ws []pptx.ReadWarning) []string {
	out := make([]string, 0, len(ws))
	for _, w := range ws {
		out = append(out, w.Kind.String()+"|"+w.Part+"|"+w.Element)
	}
	sort.Strings(out)
	return out
}

// mutateZipPart rewrites one entry of a ZIP (by name) through transform and
// returns the re-zipped bytes.
func mutateZipPart(t *testing.T, data []byte, name string, transform func(string) string) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		b, _ := io.ReadAll(rc)
		_ = rc.Close()
		if f.Name == name {
			b = []byte(transform(string(b)))
		}
		w, err := zw.Create(f.Name)
		if err != nil {
			t.Fatalf("create %s: %v", f.Name, err)
		}
		if _, err := w.Write(b); err != nil {
			t.Fatalf("write %s: %v", f.Name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}
