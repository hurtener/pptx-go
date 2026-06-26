package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/scene"
)

func orgTree() scene.Tree {
	return scene.Tree{Root: scene.TreeNode{Label: "CEO", Children: []scene.TreeNode{
		{Label: "Engineering", Children: []scene.TreeNode{{Label: "Platform"}, {Label: "Apps"}, {Label: "Infra"}}},
		{Label: "Go-to-market", AccentIndex: 1, Children: []scene.TreeNode{{Label: "Sales"}, {Label: "Marketing"}}},
	}}}
}

// TestTree is R14.10 acceptance: a 3-level tree (1→2→3/2) renders balanced node
// cards + parent-child elbow edges inside the safe area, conformant, no warnings.
func TestTree(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "org", Nodes: []scene.SlideNode{orgTree()}}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("tree: warnings: %+v", stats.Warnings)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, `prst="roundRect"`); n < 6 {
		t.Errorf("tree: want >=6 node cards (1+2+5), got %d", n)
	}
	if !strings.Contains(slide, `prst="line"`) {
		t.Errorf("tree: missing connector edges")
	}
	if !strings.Contains(slide, "<a:t>CEO</a:t>") || !strings.Contains(slide, "<a:t>Platform</a:t>") {
		t.Errorf("tree: missing node labels")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("tree deck failed conformance:\n%s", rep)
	}
}

// TestTree_LeftRight renders a left-right tree (transposed axes).
func TestTree_LeftRight(t *testing.T) {
	tr := orgTree()
	tr.Orientation = scene.FlowHorizontal
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "lr", Nodes: []scene.SlideNode{tr}}}}
	_, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("left-right tree: warnings: %+v", stats.Warnings)
	}
}

// TestTree_Deterministic guards worker-count independence.
func TestTree_Deterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{ID: "org", Nodes: []scene.SlideNode{orgTree()}}}}
	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("tree not deterministic (%d vs %d bytes)", len(seq), len(par))
	}
}
