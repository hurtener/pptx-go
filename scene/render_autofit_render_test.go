package scene_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// Display text shrink-to-fit render tests (Phase 43, R10.5). Black-box: AutoFit
// emits a reduced sz for an over-wide display run, fitting/off content is
// byte-identical, and the render is deterministic.

// TestAutoFit_Stat_EmitsReducedSz: an over-wide Stat value with AutoFit emits a
// smaller a:rPr/@sz than the same value without AutoFit (which stays at the full
// TypeDisplay size).
func TestAutoFit_Stat_EmitsReducedSz(t *testing.T) {
	long := strings.Repeat("8", 80) // overflow even at full body width
	mk := func(af bool) []byte {
		sc := scene.Scene{Slides: []scene.SceneSlide{{
			ID:    "s",
			Nodes: []scene.SlideNode{scene.Stat{Value: long, Label: "metric", AutoFit: af}},
		}}}
		data, _ := render(t, sc)
		return data
	}
	full := int(pptx.DefaultTheme().ResolveType(pptx.TypeDisplay).Size * 100)
	offXML := zipPart(t, mk(false), "ppt/slides/slide1.xml")
	onXML := zipPart(t, mk(true), "ppt/slides/slide1.xml")
	if !strings.Contains(offXML, fmt.Sprintf(`sz="%d"`, full)) {
		t.Errorf("AutoFit-off Stat should emit the full display sz=%q", fmt.Sprintf(`sz="%d"`, full))
	}
	if strings.Contains(onXML, fmt.Sprintf(`sz="%d"`, full)) {
		t.Errorf("AutoFit-on over-wide Stat should NOT emit the full display sz=%d (it should shrink)", full)
	}
}

// TestAutoFit_OffByteIdentical: a fitting display value renders byte-identically
// with AutoFit on or off (the scale is 0 when it already fits).
func TestAutoFit_OffByteIdentical(t *testing.T) {
	mk := func(af bool) []byte {
		sc := scene.Scene{Slides: []scene.SceneSlide{{
			ID:    "s",
			Nodes: []scene.SlideNode{scene.Stat{Value: "42", Label: "ok", AutoFit: af}},
		}}}
		data, _ := render(t, sc)
		return data
	}
	if !bytes.Equal(mk(false), mk(true)) {
		t.Error("a fitting Stat is not byte-identical with AutoFit on vs off")
	}
}

// TestAutoFit_Deterministic: an AutoFit deck (Hero + Heading + Stat) renders
// byte-identically across worker counts.
func TestAutoFit_Deterministic(t *testing.T) {
	long := strings.Repeat("7", 50)
	sc := scene.Scene{}
	for i := 0; i < 12; i++ {
		sc.Slides = append(sc.Slides, scene.SceneSlide{
			ID: string(rune('A' + i)),
			Nodes: []scene.SlideNode{
				scene.Hero{Title: long, AutoFit: true},
				scene.Heading{Text: rt(long), Level: 1, AutoFit: true},
				scene.Grid{Columns: 4, Cells: []scene.SlideNode{
					scene.Stat{Value: long, AutoFit: true},
					scene.Stat{Value: long, AutoFit: true},
					scene.Stat{Value: long, AutoFit: true},
					scene.Stat{Value: long, AutoFit: true},
				}},
			},
		})
	}
	seq, _ := render(t, sc, scene.WithWorkers(1))
	par, _ := render(t, sc, scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("AutoFit deck: parallel render differs from sequential (%d vs %d bytes)", len(par), len(seq))
	}
}
