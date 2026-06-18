package pptx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestTextFrame_BodyPropsRoundTrip closes a round-trip read gap: a text frame's
// auto-fit, vertical anchor, and internal margins are authored, then verified
// through the read accessors after a write → Open cycle (G6).
func TestTextFrame_BodyPropsRoundTrip(t *testing.T) {
	tf := firstTextFrame(t, func(s *pptx.Slide) {
		f := s.AddTextFrame(fxBox)
		f.AutoFit(pptx.AutoFitNormal)
		f.Anchor(pptx.AnchorMiddle)
		f.Margins(pptx.EMU(45720), pptx.EMU(91440), pptx.EMU(45720), pptx.EMU(91440))
		f.AddParagraph(pptx.ParagraphOpts{}).AddRun("body", pptx.RunStyle{TypeRole: pptx.TypeBody})
	})

	if got := tf.AutoFitMode(); got != pptx.AutoFitNormal {
		t.Errorf("AutoFitMode() = %v, want AutoFitNormal", got)
	}
	if got := tf.VerticalAnchor(); got != pptx.AnchorMiddle {
		t.Errorf("VerticalAnchor() = %v, want AnchorMiddle", got)
	}
	top, right, bottom, left := tf.MarginInsets()
	if top != 45720 || right != 91440 || bottom != 45720 || left != 91440 {
		t.Errorf("MarginInsets() = %d,%d,%d,%d, want 45720,91440,45720,91440", top, right, bottom, left)
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
