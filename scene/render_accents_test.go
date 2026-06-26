package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// fourPhaseTimeline is a Timeline whose four milestones each select a distinct
// accent index (0..3) — exercising a 4-hue brand palette cycle.
func fourPhaseTimeline() scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{
		ID: "roadmap",
		Nodes: []scene.SlideNode{scene.Timeline{Milestones: []scene.Milestone{
			{Position: 0.1, Label: "Discover", AccentIndex: 0},
			{Position: 0.4, Label: "Design", AccentIndex: 1},
			{Position: 0.7, Label: "Build", AccentIndex: 2},
			{Position: 0.95, Label: "Ship", AccentIndex: 3},
		}}},
	}}}
}

// TestMultiAccent_BrandHuesRendered is the R8.4 acceptance: a four-hue brand
// palette renders each timeline phase marker in its own brand hue (all four
// appear in the slide XML), beyond the engine's three accent roles.
func TestMultiAccent_BrandHuesRendered(t *testing.T) {
	const (
		jade   = "5EEAD4"
		orange = "F97316"
		violet = "8B5CF6"
		lime   = "A3E635" // the 4th hue — unreachable via the 3 accent roles
	)
	th := pptx.NewTheme(pptx.WithAccents(jade, orange, violet, lime))
	data, stats := render(t, fourPhaseTimeline(), scene.WithTheme(th))
	if len(stats.Warnings) != 0 {
		t.Errorf("multi-accent timeline: unexpected warnings: %+v", stats.Warnings)
	}
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, hue := range []string{jade, orange, violet, lime} {
		if !strings.Contains(slideXML, hue) {
			t.Errorf("slide missing brand accent hue %s:\n%s", hue, slideXML)
		}
	}
}

// TestMultiAccent_NilByteIdentical is the byte-identity guard: a deck on a theme
// with no brand palette is byte-for-byte identical to the same deck on the
// default theme — the pinned five-role cycle is unchanged (R8.4).
func TestMultiAccent_NilByteIdentical(t *testing.T) {
	sc := fourPhaseTimeline()
	dDefault := renderBytes(t, sc)
	dExplicit := renderBytes(t, sc, scene.WithTheme(pptx.NewTheme())) // no WithAccents
	if !bytes.Equal(dDefault, dExplicit) {
		t.Errorf("no-palette theme not byte-identical to default (%d vs %d bytes)", len(dExplicit), len(dDefault))
	}
}

// TestMultiAccent_Determinism proves a brand-palette deck renders byte-identically
// across worker counts (index→hue is a pure lookup — RFC §10.1).
func TestMultiAccent_Determinism(t *testing.T) {
	th := pptx.NewTheme(pptx.WithAccents("5EEAD4", "F97316", "8B5CF6", "A3E635"))
	// Mix a dark slide in so the sequential pass is also exercised.
	sc := scene.Scene{Slides: []scene.SceneSlide{
		fourPhaseTimeline().Slides[0],
		{ID: "dark", Variant: scene.VariantDark, Nodes: fourPhaseTimeline().Slides[0].Nodes},
	}}
	seq := renderBytes(t, sc, scene.WithTheme(th.Clone()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithTheme(th.Clone()), scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Errorf("multi-accent deck not deterministic across workers (%d vs %d bytes)", len(seq), len(par))
	}
}

// TestMultiAccent_ContrastFromHue verifies a funnel band filled with a dark brand
// accent hue gets light (inverse) contrast text — the auto-contrast path works
// from a literal palette hue, not just a ColorRole (R8.4 × D-082).
func TestMultiAccent_ContrastFromHue(t *testing.T) {
	// A very dark brand accent → its band label must use the inverse (light) text.
	th := pptx.NewTheme(pptx.WithAccents("0A0A0A"))
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "funnel",
		Nodes: []scene.SlideNode{scene.Funnel{Stages: []scene.FunnelStage{
			{Label: "Leads", Value: "100", AccentIndex: 0},
			{Label: "Qualified", Value: "40", AccentIndex: 0},
		}}},
	}}}
	data, _ := render(t, sc, scene.WithTheme(th))
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	// The default inverse text token is white (FFFFFF) — it must appear as the
	// band label color on the near-black accent fill.
	inverse := string(th.ResolveTextColor(pptx.TextInverse))
	if !strings.Contains(slideXML, inverse) {
		t.Errorf("dark brand accent band missing inverse contrast text %s:\n%s", inverse, slideXML)
	}
}
