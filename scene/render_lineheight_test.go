package scene_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestSceneLineHeight_RoleDriven is the scene side of R9.4 (D-061): a theme whose
// type role declares a line-height tightens that role's paragraphs; the default
// theme (no line-height) is byte-identical (emits no lnSpc).
func TestSceneLineHeight_RoleDriven(t *testing.T) {
	th := pptx.DefaultTheme().Clone()
	body := th.ResolveType(pptx.TypeBody)
	body.LineHeight = 120
	th.Typography[pptx.TypeBody] = body

	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "s1",
		Nodes: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("a body line")}}},
	}}}

	themed, _ := render(t, sc, scene.WithTheme(th))
	if x := zipPart(t, themed, "ppt/slides/slide1.xml"); !strings.Contains(x, `<a:spcPct val="120000"`) {
		t.Error("a TypeBody LineHeight=120 theme should emit spcPct val=\"120000\" on the prose paragraph")
	}

	plain, _ := render(t, sc)
	if x := zipPart(t, plain, "ppt/slides/slide1.xml"); strings.Contains(x, "lnSpc") {
		t.Error("default theme (no line-height) should emit no lnSpc")
	}
}
