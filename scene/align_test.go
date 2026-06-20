package scene

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Internal alignment tests (package scene). These tests verify the placement
// logic introduced in Phase 13: vertical body-stack alignment and per-node
// horizontal alignment. They also guard backward compatibility.

// newTestRenderer returns a renderer initialized against a fresh default
// presentation and theme, suitable for calling layout/alignedStackIn.
func newTestRenderer(t *testing.T) *renderer {
	t.Helper()
	pres := pptx.New()
	return &renderer{pres: pres, theme: pres.Theme()}
}

// TestAlignCenter_Vertical_SingleHero checks that VAlignCenter places a Hero
// node's Y at exactly box.Y + (box.H - heroH)/2, i.e., vertically centered
// in the body region.
func TestAlignCenter_Vertical_SingleHero(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()
	heroH := preferredHeight(Hero{})

	slack := body.H - heroH
	if slack <= 0 {
		t.Skip("no vertical slack for centering (unusual slide size)")
	}
	wantY := body.Y + slack/2

	nodes := []SlideNode{Hero{Title: "Centered"}}
	placements := r.layout(nodes, "test", Alignment{Vertical: VAlignCenter})

	var got *placement
	for i := range placements {
		if _, ok := placements[i].node.(Hero); ok {
			got = &placements[i]
			break
		}
	}
	if got == nil {
		t.Fatal("Hero placement not found")
	}
	if got.box.Y != wantY {
		t.Errorf("Hero Y = %d, want %d (vertically centered)", got.box.Y, wantY)
	}
	// Must be strictly below the body top edge.
	if got.box.Y <= body.Y {
		t.Errorf("centered hero Y (%d) should be > body.Y (%d)", got.box.Y, body.Y)
	}
}

// TestAlignBottom_Vertical checks VAlignBottom places the body stack flush
// with the body region's bottom edge.
func TestAlignBottom_Vertical(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()
	heroH := preferredHeight(Hero{})

	nodes := []SlideNode{Hero{Title: "Bottom"}}
	placements := r.layout(nodes, "test", Alignment{Vertical: VAlignBottom})

	var got *placement
	for i := range placements {
		if _, ok := placements[i].node.(Hero); ok {
			got = &placements[i]
			break
		}
	}
	if got == nil {
		t.Fatal("Hero placement not found")
	}
	wantY := body.Bottom() - heroH
	if wantY < body.Y {
		wantY = body.Y
	}
	if got.box.Y != wantY {
		t.Errorf("Hero Y = %d, want %d (bottom-aligned)", got.box.Y, wantY)
	}
}

// TestAlignCenter_Horizontal_Heading checks that HAlignCenter keeps a Heading's
// box at full body width (paragraph alignment handles centering within the frame)
// and stores HAlignCenter on the placement so the renderer can set algn="ctr".
func TestAlignCenter_Horizontal_Heading(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()

	h := Heading{Text: RichText{{Text: "Short"}}, Level: 2}
	nodes := []SlideNode{h}
	placements := r.layout(nodes, "test", Alignment{Horizontal: HAlignCenter})

	var got *placement
	for i := range placements {
		if _, ok := placements[i].node.(Heading); ok {
			got = &placements[i]
			break
		}
	}
	if got == nil {
		t.Fatal("Heading placement not found")
	}

	// Text nodes keep the full body-width box; paragraph alignment centers within it.
	if got.box.X != body.X {
		t.Errorf("centered Heading X = %d, want body.X %d (full-width box)", got.box.X, body.X)
	}
	if got.box.W != body.W {
		t.Errorf("centered Heading W = %d, want body.W %d (full-width box)", got.box.W, body.W)
	}
	// The effective alignment is carried on the placement for the renderer.
	if got.hAlign != HAlignCenter {
		t.Errorf("placement.hAlign = %v, want HAlignCenter", got.hAlign)
	}
}

// TestAlignRight_Horizontal_Heading checks that HAlignRight keeps a Heading's
// box at full body width and stores HAlignRight on the placement so the renderer
// can set algn="r" on every paragraph.
func TestAlignRight_Horizontal_Heading(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()

	h := Heading{Text: RichText{{Text: "Short"}}, Level: 2}
	nodes := []SlideNode{h}
	placements := r.layout(nodes, "test", Alignment{Horizontal: HAlignRight})

	var got *placement
	for i := range placements {
		if _, ok := placements[i].node.(Heading); ok {
			got = &placements[i]
			break
		}
	}
	if got == nil {
		t.Fatal("Heading placement not found")
	}

	// Text nodes keep the full body-width box; paragraph alignment right-aligns within it.
	if got.box.X != body.X {
		t.Errorf("right-aligned Heading X = %d, want body.X %d (full-width box)", got.box.X, body.X)
	}
	if got.box.W != body.W {
		t.Errorf("right-aligned Heading W = %d, want body.W %d (full-width box)", got.box.W, body.W)
	}
	if got.hAlign != HAlignRight {
		t.Errorf("placement.hAlign = %v, want HAlignRight", got.hAlign)
	}
}

// TestAlignPerNode_Override_HeadingRight checks that a per-node HAlignRight
// overrides the slide-level HAlignLeft default and that the placement carries
// HAlignRight so the renderer can set algn="r" on its paragraphs.
func TestAlignPerNode_Override_HeadingRight(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()

	h := Heading{Text: RichText{{Text: "Right-node"}}, Level: 1, Align: HAlignRight}

	// Slide default = left; node overrides to right.
	nodes := []SlideNode{h}
	placements := r.layout(nodes, "test", Alignment{Horizontal: HAlignLeft})

	var got *placement
	for i := range placements {
		if _, ok := placements[i].node.(Heading); ok {
			got = &placements[i]
			break
		}
	}
	if got == nil {
		t.Fatal("Heading placement not found")
	}

	// Full-width box: the text frame spans body.W; paragraph right-aligns within it.
	if got.box.X != body.X {
		t.Errorf("per-node right Heading X = %d, want body.X %d", got.box.X, body.X)
	}
	if got.box.W != body.W {
		t.Errorf("per-node right Heading W = %d, want body.W %d", got.box.W, body.W)
	}
	if got.hAlign != HAlignRight {
		t.Errorf("placement.hAlign = %v, want HAlignRight", got.hAlign)
	}
}

// TestAlignPerNode_Override_NodeCenterSlideLeft checks that a per-node
// HAlignCenter overrides a slide-level HAlignLeft, keeps the full-width box,
// and records HAlignCenter on the placement for the renderer.
func TestAlignPerNode_Override_NodeCenterSlideLeft(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()

	h := Heading{Text: RichText{{Text: "Ctr"}}, Level: 2, Align: HAlignCenter}
	nodes := []SlideNode{h}
	placements := r.layout(nodes, "test", Alignment{Horizontal: HAlignLeft})

	var got *placement
	for i := range placements {
		if _, ok := placements[i].node.(Heading); ok {
			got = &placements[i]
			break
		}
	}
	if got == nil {
		t.Fatal("Heading placement not found")
	}

	// Full-width box: no physical centering of the frame; paragraph centering does the work.
	if got.box.X != body.X {
		t.Errorf("per-node center: X = %d, want body.X %d", got.box.X, body.X)
	}
	if got.box.W != body.W {
		t.Errorf("per-node center: W = %d, want body.W %d", got.box.W, body.W)
	}
	if got.hAlign != HAlignCenter {
		t.Errorf("placement.hAlign = %v, want HAlignCenter", got.hAlign)
	}
}

// TestAlignZeroValue_BackwardCompat proves that zero-value Alignment produces
// placements byte-identical to the pre-alignment stackIn output. This is the
// backward-compatibility regression guard.
func TestAlignZeroValue_BackwardCompat(t *testing.T) {
	r := newTestRenderer(t)
	nodes := []SlideNode{
		Heading{Text: RichText{{Text: "Section"}}, Level: 1},
		Prose{Paragraphs: []RichText{{{Text: "Body text."}}}},
		List{Items: []ListItem{
			{Text: RichText{{Text: "a"}}},
			{Text: RichText{{Text: "b"}}},
		}},
	}

	// Legacy path: stackIn directly.
	legacy := r.stackIn(r.bodyRegion(), nodes, "test")
	// New path: alignedStackIn with zero alignment (must match exactly).
	aligned := r.alignedStackIn(r.bodyRegion(), nodes, "test", Alignment{})

	if len(legacy) != len(aligned) {
		t.Fatalf("placement count: legacy=%d aligned=%d", len(legacy), len(aligned))
	}
	for i := range legacy {
		if legacy[i].box != aligned[i].box {
			t.Errorf("placement[%d] box differs: legacy=%+v aligned=%+v", i, legacy[i].box, aligned[i].box)
		}
	}
}

// TestAlignContainers_AlwaysFullWidth proves that container nodes (Grid,
// TwoColumn, Callout, Divider) keep their full body-region box even when the
// slide has HAlignCenter set — alignment within containers is their own
// concern (Phase 13 spec: OUT of scope).
func TestAlignContainers_AlwaysFullWidth(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()

	nodes := []SlideNode{
		Divider{},
		Callout{Kind: CalloutNote, Body: RichText{{Text: "note"}}},
	}
	placements := r.layout(nodes, "test", Alignment{Horizontal: HAlignCenter})

	for _, pl := range placements {
		switch pl.node.(type) {
		case Divider, Callout:
			if pl.box.X != body.X {
				t.Errorf("%T: X = %d, want body.X %d (containers keep full width)", pl.node, pl.box.X, body.X)
			}
			if pl.box.W != body.W {
				t.Errorf("%T: W = %d, want body.W %d (containers keep full width)", pl.node, pl.box.W, body.W)
			}
		}
	}
}

// TestAlignJustify_Vertical checks that VAlignJustify distributes the slack
// into the inter-node gaps, making the last node's bottom coincide (or nearly
// so) with the body region's bottom.
func TestAlignJustify_Vertical(t *testing.T) {
	r := newTestRenderer(t)
	body := r.bodyRegion()

	nodes := []SlideNode{
		Heading{Text: RichText{{Text: "H"}}, Level: 1},
		Prose{Paragraphs: []RichText{{{Text: "P"}}}},
		Divider{},
	}

	placements := r.layout(nodes, "test", Alignment{Vertical: VAlignJustify})

	// Collect only the stacked (non-decoration, non-section) placements.
	var stacked []placement
	for _, pl := range placements {
		switch pl.node.(type) {
		case Decoration, SectionDivider:
			// skip
		default:
			stacked = append(stacked, pl)
		}
	}
	if len(stacked) != 3 {
		t.Fatalf("expected 3 stacked placements, got %d", len(stacked))
	}
	// The last node's bottom should equal the body bottom (when totalH ≤ body.H).
	last := stacked[len(stacked)-1]
	totalH := preferredHeight(nodes[0]) + preferredHeight(nodes[1]) + preferredHeight(nodes[2])
	if totalH <= body.H {
		wantBottom := body.Bottom()
		if last.box.Bottom() != wantBottom {
			t.Errorf("VAlignJustify: last node bottom = %d, want body.Bottom() = %d", last.box.Bottom(), wantBottom)
		}
	}
}

// TestAlignDeterminism_ByteIdentical is the Phase-13 determinism guard for
// aligned scenes: rendering with 1 worker vs 4 workers must produce byte-
// identical PPTX output (RFC §10.1 + Phase 13 requirement).
func TestAlignDeterminism_ByteIdentical(t *testing.T) {
	sc := Scene{
		Slides: []SceneSlide{
			{
				ID:      "s1",
				Content: Alignment{Vertical: VAlignCenter, Horizontal: HAlignCenter},
				Nodes: []SlideNode{
					Heading{Text: RichText{{Text: "Centered Title"}}, Level: 1},
					Prose{Paragraphs: []RichText{{{Text: "Body text here"}}}},
					List{Items: []ListItem{{Text: RichText{{Text: "item one"}}}, {Text: RichText{{Text: "item two"}}}}},
				},
			},
			{
				ID:      "s2",
				Content: Alignment{Vertical: VAlignBottom, Horizontal: HAlignRight},
				Nodes: []SlideNode{
					Heading{Text: RichText{{Text: "Right bottom"}}, Level: 2},
					Prose{Paragraphs: []RichText{{{Text: "detail"}}}},
				},
			},
			{
				ID:      "s3",
				Content: Alignment{Vertical: VAlignJustify},
				Nodes: []SlideNode{
					Heading{Text: RichText{{Text: "Justified"}}, Level: 3},
					Divider{},
					Prose{Paragraphs: []RichText{{{Text: "last"}}}},
				},
			},
		},
	}

	doRender := func(workers int) []byte {
		pres := pptx.New()
		if _, err := Render(pres, sc, WithWorkers(workers)); err != nil {
			t.Fatalf("Render(workers=%d): %v", workers, err)
		}
		data, err := pres.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}

	seq := doRender(1)
	par := doRender(4)
	if !bytes.Equal(seq, par) {
		t.Fatalf("aligned scene: parallel render (%d bytes) differs from sequential (%d bytes)", len(par), len(seq))
	}

	// Run twice at default workers to confirm stability.
	a := doRender(0)
	b := doRender(0)
	if !bytes.Equal(a, b) {
		t.Fatal("aligned scene: two default-worker renders are not byte-identical")
	}
	if !bytes.Equal(a, seq) {
		t.Fatal("aligned scene: default-worker render differs from sequential render")
	}
}

// alignZipPart extracts a named part from a PPTX zip archive.
func alignZipPart(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, _ := f.Open()
			defer func() { _ = rc.Close() }()
			b, _ := io.ReadAll(rc)
			return string(b)
		}
	}
	t.Fatalf("part %s not found", name)
	return ""
}

// TestAlignParagraph_XMLAttributes verifies that the rendered slide XML contains
// the correct OOXML algn attributes for centered and right-aligned headings, and
// that a left-aligned (default) heading emits no algn attribute.
//
// This is the key correctness guard: it proves that paragraph alignment (not just
// placement box arithmetic) reaches the PPTX bytes.
func TestAlignParagraph_XMLAttributes(t *testing.T) {
	sc := Scene{
		Slides: []SceneSlide{
			{
				ID:      "center-slide",
				Content: Alignment{Horizontal: HAlignCenter},
				Nodes: []SlideNode{
					Heading{Text: RichText{{Text: "Centered"}}, Level: 1},
				},
			},
			{
				ID:      "right-slide",
				Content: Alignment{Horizontal: HAlignRight},
				Nodes: []SlideNode{
					Heading{Text: RichText{{Text: "Right"}}, Level: 2},
				},
			},
			{
				ID:      "left-slide",
				Content: Alignment{Horizontal: HAlignLeft},
				Nodes: []SlideNode{
					Heading{Text: RichText{{Text: "Left"}}, Level: 3},
				},
			},
		},
	}

	pres := pptx.New()
	if _, err := Render(pres, sc); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	// Slide 1: HAlignCenter → paragraph must carry algn="ctr".
	slide1 := alignZipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide1, `algn="ctr"`) {
		t.Errorf("slide1 (center heading): expected algn=\"ctr\" in XML\n%.400s", slide1)
	}

	// Slide 2: HAlignRight → paragraph must carry algn="r".
	slide2 := alignZipPart(t, data, "ppt/slides/slide2.xml")
	if !strings.Contains(slide2, `algn="r"`) {
		t.Errorf("slide2 (right heading): expected algn=\"r\" in XML\n%.400s", slide2)
	}

	// Slide 3: HAlignLeft → no algn attribute (OOXML default = left).
	slide3 := alignZipPart(t, data, "ppt/slides/slide3.xml")
	if strings.Contains(slide3, `algn=`) {
		t.Errorf("slide3 (left heading): unexpected algn attr in XML (should default to left)\n%.400s", slide3)
	}
}
