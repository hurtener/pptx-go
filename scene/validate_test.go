package scene_test

import (
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

func sceneWith(nodes ...scene.SlideNode) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: nodes}}}
}

// TestValidateScene_AcceptsValid proves Stage 1 accepts a valid instance of
// every node type.
func TestValidateScene_AcceptsValid(t *testing.T) {
	valid := []scene.SlideNode{
		scene.Hero{Title: "T"},
		scene.Prose{Paragraphs: []scene.RichText{{{Text: "x"}}}},
		scene.Heading{Text: scene.RichText{{Text: "h"}}, Level: 2},
		scene.List{Kind: scene.ListChecklist, Items: []scene.ListItem{{Checked: true}}},
		scene.Divider{},
		scene.Quote{Text: scene.RichText{{Text: "q"}}},
		scene.Callout{Kind: scene.CalloutWarning, Body: scene.RichText{{Text: "b"}}},
		scene.Image{AssetID: "asset://1", Alt: "logo"},
		scene.Chip{Label: "new", Tone: scene.ChipSolid},
		scene.Arrow{Direction: scene.ArrowRight},
		scene.CodeBlock{AssetID: "asset://2", Language: "go"},
		scene.Chart{AssetID: "asset://3"},
		scene.Table{Headers: []scene.RichText{{{Text: "A"}}, {{Text: "B"}}}, Rows: [][]scene.RichText{{{}, {}}}},
		scene.Flow{Steps: []scene.FlowStep{{Label: scene.RichText{{Text: "1"}}}}},
		scene.Decoration{Kind: scene.DecorationAsset, AssetID: "asset://4", Layer: scene.LayerBackground},
		scene.SectionDivider{Label: "Part II"},
		scene.TwoColumn{Ratio: scene.Ratio12, Left: []scene.SlideNode{scene.Prose{}}, Right: []scene.SlideNode{scene.Prose{}}},
		scene.Grid{Columns: 3, Ratio: []int{1, 1, 2}, Cells: []scene.SlideNode{scene.Prose{}, scene.Prose{}, scene.Prose{}}},
		scene.Card{Header: "C", Body: []scene.SlideNode{scene.Prose{}}},
		scene.CardSection{Header: "S", Body: []scene.SlideNode{scene.Grid{Columns: 2, Cells: []scene.SlideNode{scene.Prose{}, scene.Prose{}}}}},
	}
	if err := scene.ValidateScene(sceneWith(valid...)); err != nil {
		t.Fatalf("valid catalog rejected: %v", err)
	}
}

// TestValidateScene_RejectsNegatives proves Stage 1 catches malformed nodes.
func TestValidateScene_RejectsNegatives(t *testing.T) {
	negatives := map[string]scene.SlideNode{
		"heading level 0":       scene.Heading{Level: 0},
		"heading level 7":       scene.Heading{Level: 7},
		"empty list":            scene.List{},
		"image without asset":   scene.Image{},
		"chart without asset":   scene.Chart{},
		"code without asset":    scene.CodeBlock{},
		"preset without name":   scene.Decoration{Kind: scene.DecorationPreset},
		"asset deco no id":      scene.Decoration{Kind: scene.DecorationAsset},
		"flow without steps":    scene.Flow{},
		"table ragged row":      scene.Table{Headers: []scene.RichText{{}, {}}, Rows: [][]scene.RichText{{{}}}},
		"two_column empty side": scene.TwoColumn{Left: []scene.SlideNode{scene.Prose{}}},
		"grid 1 column":         scene.Grid{Columns: 1, Cells: []scene.SlideNode{scene.Prose{}}},
		"grid ratio mismatch":   scene.Grid{Columns: 2, Ratio: []int{1, 1, 1}, Cells: []scene.SlideNode{scene.Prose{}}},
		"card_section empty":    scene.CardSection{},
	}
	for name, n := range negatives {
		if err := scene.ValidateScene(sceneWith(n)); err == nil {
			t.Errorf("%s: expected a validation error, got nil", name)
		}
	}
}

// TestValidateScene_NilNode rejects a nil node.
func TestValidateScene_NilNode(t *testing.T) {
	if err := scene.ValidateScene(sceneWith(nil)); err == nil {
		t.Error("expected an error for a nil node")
	}
}

// TestValidateScene_RecursesContainers proves a malformed child inside a
// container is reported.
func TestValidateScene_RecursesContainers(t *testing.T) {
	bad := scene.Card{Body: []scene.SlideNode{scene.Image{}}} // image without asset
	if err := scene.ValidateScene(sceneWith(bad)); err == nil {
		t.Error("expected a nested validation error for an asset-less image in a card")
	}
}
