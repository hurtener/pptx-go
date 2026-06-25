package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// matrixScene builds a 4-column feature×plan comparison matrix with the given
// style.
func matrixScene(style *scene.TableStyle) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID: "matrix",
		Nodes: []scene.SlideNode{scene.Table{
			Headers: []scene.RichText{rt("Feature"), rt("Free"), rt("Pro"), rt("Enterprise")},
			Rows: [][]scene.RichText{
				{rt("Seats"), rt("1"), rt("10"), rt("Unlimited")},
				{rt("SSO"), rt("—"), rt("Yes"), rt("Yes")},
				{rt("SLA"), rt("—"), rt("99.9%"), rt("99.99%")},
			},
			Style: style,
		}},
	}}}
}

// TestTableStyle_Emits is R14.3 acceptance (D-118): a styled matrix renders a
// header band, zebra striping, and a highlighted column — all token-driven — with
// no warnings and a conformant deck.
func TestTableStyle_Emits(t *testing.T) {
	accent := pptx.DefaultTheme().ResolveColor(pptx.ColorAccent)
	surfaceAlt := pptx.DefaultTheme().ResolveColor(pptx.ColorSurfaceAlt)
	data, stats := render(t, matrixScene(&scene.TableStyle{
		HeaderFill: true, Zebra: true, HighlightCol: 3, RowLabelCol: true,
	}))
	if len(stats.Warnings) != 0 {
		t.Errorf("styled table: unexpected warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<a:tbl>") {
		t.Fatalf("styled table did not emit a tbl:\n%s", slide)
	}
	if !strings.Contains(slide, string(accent)) {
		t.Errorf("styled table missing accent fill %s (header band / highlight col)", accent)
	}
	if !strings.Contains(slide, string(surfaceAlt)) {
		t.Errorf("styled table missing surfaceAlt fill %s (zebra / row labels)", surfaceAlt)
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("styled table deck failed conformance:\n%s", rep)
	}
}

// TestTableStyle_HeaderGroups verifies a grouped header row merges its spans.
func TestTableStyle_HeaderGroups(t *testing.T) {
	data, stats := render(t, matrixScene(&scene.TableStyle{
		HeaderFill: true,
		HeaderGroups: []scene.HeaderGroup{
			{Label: "Plan", Span: 1},
			{Label: "Paid tiers", Span: 3},
		},
	}))
	if len(stats.Warnings) != 0 {
		t.Errorf("grouped header: unexpected warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slide, "<a:t>Paid tiers</a:t>") {
		t.Errorf("grouped header label missing:\n%s", slide)
	}
	if !strings.Contains(slide, `gridSpan="3"`) {
		t.Errorf("grouped header did not merge a 3-column span:\n%s", slide)
	}
}

// TestTableStyle_NilByteIdentical verifies a nil Style is byte-identical to a
// table with no Style field (R14.3: zero renders the plain banded table).
func TestTableStyle_NilByteIdentical(t *testing.T) {
	withNil, _ := render(t, matrixScene(nil))
	plain, _ := render(t, scene.Scene{Slides: []scene.SceneSlide{{
		ID: "matrix",
		Nodes: []scene.SlideNode{scene.Table{
			Headers: []scene.RichText{rt("Feature"), rt("Free"), rt("Pro"), rt("Enterprise")},
			Rows: [][]scene.RichText{
				{rt("Seats"), rt("1"), rt("10"), rt("Unlimited")},
				{rt("SSO"), rt("—"), rt("Yes"), rt("Yes")},
				{rt("SLA"), rt("—"), rt("99.9%"), rt("99.99%")},
			},
		}},
	}}})
	if !bytes.Equal(withNil, plain) {
		t.Errorf("nil Style not byte-identical to absent field (%d vs %d bytes)", len(withNil), len(plain))
	}
}

// TestTableStyle_Deterministic guards that the styled table is worker-count
// independent.
func TestTableStyle_Deterministic(t *testing.T) {
	sc := matrixScene(&scene.TableStyle{HeaderFill: true, Zebra: true, HighlightCol: 2, RowLabelCol: true,
		HeaderGroups: []scene.HeaderGroup{{Label: "A", Span: 2}, {Label: "B", Span: 2}}})
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("styled table not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
