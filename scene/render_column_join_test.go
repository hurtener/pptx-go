package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/scene"
)

// TwoColumn column join (Phase 26, R5 a+b): a centered "VS"-style badge or a
// connector arrow on the column seam, opt-in and byte-identical when JoinNone.

func twoColScene(tc scene.TwoColumn) scene.Scene {
	return scene.Scene{Slides: []scene.SceneSlide{{ID: "s1", Nodes: []scene.SlideNode{tc}}}}
}

func twoColCols() ([]scene.SlideNode, []scene.SlideNode) {
	left := []scene.SlideNode{scene.Card{Header: "Plan A"}}
	right := []scene.SlideNode{scene.Card{Header: "Plan B"}}
	return left, right
}

// TestColumnJoin_Badge is acceptance criterion 1: a JoinBadge renders an ellipse
// and the label text on the seam.
func TestColumnJoin_Badge(t *testing.T) {
	l, r := twoColCols()
	data, _ := render(t, twoColScene(scene.TwoColumn{Left: l, Right: r, Join: scene.JoinBadge, JoinLabel: "VS"}))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, `prst="ellipse"`) {
		t.Error("JoinBadge: expected an ellipse badge shape")
	}
	if !strings.Contains(xml, "<a:t>VS</a:t>") {
		t.Error("JoinBadge: expected the 'VS' label text")
	}
}

// TestColumnJoin_Arrow is acceptance criterion 2: a JoinArrow renders a
// right-arrow connector.
func TestColumnJoin_Arrow(t *testing.T) {
	l, r := twoColCols()
	data, _ := render(t, twoColScene(scene.TwoColumn{Left: l, Right: r, Join: scene.JoinArrow}))
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, `prst="rightArrow"`) {
		t.Error("JoinArrow: expected a right-arrow connector shape")
	}
}

// TestColumnJoin_NoneByteIdentical is acceptance criterion 3: JoinNone (zero)
// renders byte-identical to a TwoColumn with the join fields absent, and emits
// no join shapes.
func TestColumnJoin_NoneByteIdentical(t *testing.T) {
	l, r := twoColCols()
	bare, _ := render(t, twoColScene(scene.TwoColumn{Left: l, Right: r}))
	none, _ := render(t, twoColScene(scene.TwoColumn{Left: l, Right: r, Join: scene.JoinNone, JoinLabel: "ignored"}))
	if !bytes.Equal(bare, none) {
		t.Errorf("JoinNone is not byte-identical to an absent join (%d vs %d bytes)", len(none), len(bare))
	}
	xml := zipPart(t, bare, "ppt/slides/slide1.xml")
	if strings.Contains(xml, `prst="rightArrow"`) {
		t.Error("no join: unexpected connector arrow")
	}
}

// TestColumnJoin_Deterministic is acceptance criterion 4: a deck of join columns
// renders byte-identical across worker counts.
func TestColumnJoin_Deterministic(t *testing.T) {
	sc := scene.Scene{}
	for i := 0; i < 12; i++ {
		l, r := twoColCols()
		join := scene.JoinBadge
		if i%2 == 1 {
			join = scene.JoinArrow
		}
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID:    string(rune('A' + i)),
			Nodes: []scene.SlideNode{scene.TwoColumn{Left: l, Right: r, Join: join, JoinLabel: "VS"}},
		})
	}
	seq, _ := render(t, sc, scene.WithWorkers(1))
	par, _ := render(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("join deck: parallel render differs from sequential (%d vs %d bytes)", len(par), len(seq))
	}
}
