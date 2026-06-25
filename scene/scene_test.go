package scene_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// allNodes is one instance of every catalog node, used by the catalog and
// policy assertions.
func allNodes() []scene.SlideNode {
	return []scene.SlideNode{
		scene.Hero{}, scene.Prose{}, scene.Heading{Level: 1}, scene.List{Items: []scene.ListItem{{}}},
		scene.Divider{}, scene.Quote{}, scene.Callout{}, scene.Image{AssetID: "a"}, scene.Chip{},
		scene.Arrow{}, scene.CodeBlock{AssetID: "a"}, scene.Chart{AssetID: "a"},
		scene.Table{Headers: []scene.RichText{{}}}, scene.Flow{Steps: []scene.FlowStep{{}}},
		scene.Decoration{Kind: scene.DecorationPreset, Preset: "p"}, scene.SectionDivider{},
		scene.TwoColumn{Left: []scene.SlideNode{scene.Prose{}}, Right: []scene.SlideNode{scene.Prose{}}},
		scene.Grid{Columns: 2, Cells: []scene.SlideNode{scene.Prose{}}},
		scene.Card{}, scene.CardSection{Body: []scene.SlideNode{scene.Prose{}}},
		scene.Bento{Columns: 2, Rows: []scene.BentoRow{{Label: "L", Cells: []scene.BentoCell{{Span: 1, Node: scene.Prose{}}}}}},
		scene.Stat{Value: "42"},
		scene.Button{Label: "Go"},
		scene.Checklist{Items: []scene.ChecklistItem{{Text: scene.RichText{{Text: "done"}}}}},
		scene.ChipRow{Chips: []scene.ChipSpec{{Label: "tag"}}},
		scene.Banner{Lead: scene.RichText{{Text: "Big takeaway"}}},
		scene.IconRows{Rows: []scene.IconRow{{Label: scene.RichText{{Text: "row"}}}}},
		scene.Lockup{Caption: "POWERED BY", AssetID: "a"},
		scene.Timeline{Milestones: []scene.Milestone{{Position: 0.5, Label: "M1"}}},
	}
}

// TestCatalog_KindsDistinct proves every node implements SlideNode and reports a
// distinct, named kind (the full catalog compiles and is discriminable).
func TestCatalog_KindsDistinct(t *testing.T) {
	seen := map[scene.NodeKind]bool{}
	for _, n := range allNodes() {
		k := n.NodeKind()
		if seen[k] {
			t.Errorf("duplicate NodeKind %s", k)
		}
		seen[k] = true
		if k.String() == "unknown" {
			t.Errorf("node %T has no String()", n)
		}
	}
	if len(seen) != 29 {
		t.Errorf("catalog has %d distinct kinds, want 29", len(seen))
	}
}

// TestPolicy_MatchesStructs is acceptance criterion 5: a node renders as an
// image iff its struct carries an AssetID field, matching the §12.1 table.
func TestPolicy_MatchesStructs(t *testing.T) {
	for _, n := range allNodes() {
		k := n.NodeKind()
		_, hasField := reflect.TypeOf(n).FieldByName("AssetID")
		want := scene.PolicyFor(k).HasAsset
		if hasField != want {
			t.Errorf("%s: AssetID field present=%v, policy HasAsset=%v", k, hasField, want)
		}
	}
}

// TestRenderStub_EmptyScene is acceptance criterion 3: Render is callable and
// returns a zero Stats with no error on an empty scene.
func TestRenderStub_EmptyScene(t *testing.T) {
	pres := pptx.New()
	stats, err := scene.Render(pres, scene.Scene{})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if stats.Slides != 0 || stats.Shapes != 0 || stats.Assets != 0 || len(stats.Warnings) != 0 {
		t.Errorf("Render on empty scene = %+v, want zero Stats", stats)
	}
}

// TestRenderStub_ValidatesScene proves the stub surfaces Stage 1 errors.
func TestRenderStub_ValidatesScene(t *testing.T) {
	pres := pptx.New()
	bad := scene.Scene{Slides: []scene.SceneSlide{{Nodes: []scene.SlideNode{scene.Image{}}}}}
	if _, err := scene.Render(pres, bad); err == nil {
		t.Error("Render accepted a scene with an asset-less image")
	}
}

// TestURIAssetResolver is acceptance criterion 4: asset:// ids resolve via the
// caller callback; a miss returns ErrAssetNotFound.
func TestURIAssetResolver(t *testing.T) {
	want := []byte("PNGDATA")
	r := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		if uuid == "0fa6" {
			return want, "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})

	data, ct, err := r.Resolve(context.Background(), "asset://0fa6")
	if err != nil || ct != "image/png" || string(data) != string(want) {
		t.Fatalf("resolve hit = (%q, %q, %v), want (%q, image/png, nil)", data, ct, err, want)
	}
	if _, _, err := r.Resolve(context.Background(), "asset://missing"); !errors.Is(err, scene.ErrAssetNotFound) {
		t.Errorf("resolve miss err = %v, want ErrAssetNotFound", err)
	}
}
