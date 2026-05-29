package scene_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func rt(text string) scene.RichText { return scene.RichText{{Text: text}} }

func zipPart(t *testing.T, data []byte, name string) string {
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

func render(t *testing.T, sc scene.Scene, opts ...scene.RenderOption) ([]byte, scene.Stats) {
	t.Helper()
	pres := pptx.New()
	stats, err := scene.Render(pres, sc, opts...)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data, stats
}

// TestRenderLeaves_EachLeaf is acceptance criterion 1: one of each in-scope text
// leaf renders to a conformant PPTX with native shapes.
func TestRenderLeaves_EachLeaf(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{ID: "hero", Nodes: []scene.SlideNode{scene.Hero{Eyebrow: "Q3", Title: "Results", Subtitle: "FY25"}}},
		{ID: "prose", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("Body text.")}}}},
		{ID: "heading", Nodes: []scene.SlideNode{scene.Heading{Text: rt("Section"), Level: 2}}},
		{ID: "list", Nodes: []scene.SlideNode{scene.List{Kind: scene.ListChecklist, Items: []scene.ListItem{{Text: rt("one")}, {Text: rt("two")}}}}},
		{ID: "divider", Nodes: []scene.SlideNode{scene.Divider{}}},
		{ID: "quote", Nodes: []scene.SlideNode{scene.Quote{Text: rt("To be."), Attribution: "Hamlet"}}},
		{ID: "callout", Nodes: []scene.SlideNode{scene.Callout{Kind: scene.CalloutWarning, Title: "Heads up", Body: rt("Careful.")}}},
		{ID: "chip", Nodes: []scene.SlideNode{scene.Chip{Label: "New", Tone: scene.ChipSolid, Color: scene.ColorAccent}}},
		{ID: "arrow", Nodes: []scene.SlideNode{scene.Arrow{Direction: scene.ArrowRight, Label: "next"}}},
		{ID: "sectiondiv", Nodes: []scene.SlideNode{scene.SectionDivider{Eyebrow: "Part", Label: "Two"}}},
	}}

	data, stats := render(t, sc)
	if stats.Slides != 10 {
		t.Errorf("Stats.Slides = %d, want 10", stats.Slides)
	}
	if stats.Shapes == 0 {
		t.Errorf("Stats.Shapes = 0, want > 0")
	}
	if len(stats.Warnings) != 0 {
		t.Errorf("unexpected warnings: %+v", stats.Warnings)
	}

	rep, err := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"}})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("rendered deck failed conformance:\n%s", rep)
	}

	// The hero slide carries native text shapes (no pic).
	hero := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(hero, "<p:sp>") || !strings.Contains(hero, "<a:t>Results</a:t>") {
		t.Errorf("hero slide missing native title text:\n%s", hero)
	}
	if strings.Contains(hero, "<p:pic>") {
		t.Errorf("hero slide unexpectedly contains a pic shape")
	}
}

// TestRenderCodeBlock is acceptance criterion 2: a code_block with a registered
// asset renders the image and the caption (the image flavor, D-014).
func TestRenderCodeBlock(t *testing.T) {
	png := append([]byte("\x89PNG\r\n\x1a\n"), []byte("code-shot")...)
	resolver := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "code1" {
			return png, "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})

	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "code",
		Nodes: []scene.SlideNode{scene.CodeBlock{AssetID: "asset://code1", Language: "go", Caption: "main.go"}},
	}}}

	data, stats := render(t, sc, scene.WithAssetResolver(resolver))
	if stats.Assets != 1 {
		t.Errorf("Stats.Assets = %d, want 1", stats.Assets)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<p:pic>") || !strings.Contains(slide, "r:embed=") {
		t.Errorf("code_block slide missing pic/embed:\n%s", slide)
	}
	if !strings.Contains(slide, "<a:t>main.go</a:t>") {
		t.Errorf("code_block caption missing:\n%s", slide)
	}
	if zipPart(t, data, "ppt/media/image1.png") != string(png) {
		t.Errorf("code_block image bytes not embedded verbatim")
	}
}

// TestRenderCodeBlock_NoResolver warns and skips when the asset can't resolve.
func TestRenderCodeBlock_NoResolver(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "code", Nodes: []scene.SlideNode{scene.CodeBlock{AssetID: "asset://x", Caption: "c"}},
	}}}
	_, stats := render(t, sc) // no resolver
	if len(stats.Warnings) == 0 {
		t.Error("expected a LayoutWarning for an unresolved code_block asset")
	}
	if stats.Assets != 0 {
		t.Errorf("Stats.Assets = %d, want 0", stats.Assets)
	}
}

// TestRender_NoBoost_NinePt is acceptance criterion 3: a 9pt body role renders
// at 9pt — the library does not boost text sizes (D-026).
func TestRender_NoBoost_NinePt(t *testing.T) {
	th := pptx.DefaultTheme()
	th.Typography[pptx.TypeBody] = pptx.FontSpec{Family: "Arial", Size: 9, Weight: 400}

	sc := scene.Scene{
		Theme:  th,
		Slides: []scene.SceneSlide{{ID: "s", Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("tiny")}}}}},
	}
	data, _ := render(t, sc)
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, `sz="900"`) {
		t.Errorf("9pt body not rendered at sz=900 (no boosting):\n%s", slide)
	}
}

// TestRenderRoundTrip is acceptance criterion 5: scene → PPTX → re-read the
// shape model via pptx.Open.
func TestRenderRoundTrip(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "s", Nodes: []scene.SlideNode{
			scene.Heading{Text: rt("Title"), Level: 1},
			scene.Prose{Paragraphs: []scene.RichText{rt("Body")}},
		},
	}}}
	data, _ := render(t, sc)

	reopened, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if n := len(reopened.Slides()); n != 1 {
		t.Fatalf("reopened slides = %d, want 1", n)
	}
	// The reopened slide carries the two text shapes.
	shapes := 0
	for _, c := range reopened.Slides()[0].Part().SpTree().Children {
		_ = c
		shapes++
	}
	if shapes < 2 {
		t.Errorf("reopened slide has %d shapes, want >= 2", shapes)
	}
}

// TestRender_UnimplementedNodeWarns proves a not-yet-rendered node (e.g. a
// container) warns and is skipped rather than failing the render.
func TestRender_UnimplementedNodeWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "s", Nodes: []scene.SlideNode{scene.Grid{Columns: 2, Cells: []scene.SlideNode{scene.Prose{}}}},
	}}}
	pres := pptx.New()
	stats, err := scene.Render(pres, sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if len(stats.Warnings) == 0 {
		t.Error("expected a LayoutWarning for an unimplemented container node")
	}
}
