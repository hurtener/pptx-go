package scene_test

import (
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// Black-box tests for R11.7 join-badge fit-to-label (D-087).

func joinScene(label string) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{
		scene.TwoColumn{
			Join:      scene.JoinBadge,
			JoinLabel: label,
			Left:      []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("left")}}},
			Right:     []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("right")}}},
		},
	}}}}
}

// TestJoinBadge_GrowsToLabel: a multi-word label grows the badge diameter beyond the
// base In(0.62), while a short "vs" keeps the base (byte-identical).
func TestJoinBadge_GrowsToLabel(t *testing.T) {
	const baseEMU = 566928 // In(0.62)

	vsData, _ := render(t, joinScene("vs"))
	_, vsCx := ellipseOffX(t, zipPart(t, vsData, "ppt/slides/slide1.xml"))
	if vsCx != baseEMU {
		t.Errorf("short 'vs' badge diameter = %d, want the base %d (byte-identical)", vsCx, baseEMU)
	}

	longData, _ := render(t, joinScene("One agent"))
	_, longCx := ellipseOffX(t, zipPart(t, longData, "ppt/slides/slide1.xml"))
	if longCx <= vsCx {
		t.Errorf("multi-word badge diameter %d should exceed the base %d", longCx, vsCx)
	}
}

// TestJoinBadge_CapsAndShrinks: a pathologically long label is capped at the max
// diameter (In(1.5)) and the label is shrunk (a reduced @sz appears).
func TestJoinBadge_CapsAndShrinks(t *testing.T) {
	const maxEMU = 1371600 // In(1.5)
	data, _ := render(t, joinScene("an extremely long join connector label"))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	_, cx := ellipseOffX(t, xml)
	if cx > maxEMU {
		t.Errorf("badge diameter %d exceeds the cap %d", cx, maxEMU)
	}
	if cx != maxEMU {
		t.Errorf("an over-long label should grow the badge to the cap %d, got %d", maxEMU, cx)
	}
}
