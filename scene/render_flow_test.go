package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func flowSteps(labels ...string) []scene.FlowStep {
	steps := make([]scene.FlowStep, len(labels))
	for i, l := range labels {
		steps[i] = scene.FlowStep{Label: rt(l)}
	}
	return steps
}

// TestRenderFlow is criterion 1: a 4-step horizontal arrow flow renders 4 pills
// + 3 arrow glyphs as native shapes, conforms, and emits no pic.
func TestRenderFlow(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "flow",
		Nodes: []scene.SlideNode{scene.Flow{Steps: flowSteps("Plan", "Build", "Ship", "Learn")}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("unexpected warnings: %+v", stats.Warnings)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("flow deck failed conformance:\n%s", rep)
	}
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(xml, `prst="roundRect"`); n < 4 {
		t.Errorf("flow rendered %d pills, want >= 4:\n%s", n, xml)
	}
	if n := strings.Count(xml, `prst="rightArrow"`); n != 3 {
		t.Errorf("flow rendered %d arrows, want 3 (between 4 pills)", n)
	}
	for _, lbl := range []string{"Plan", "Build", "Ship", "Learn"} {
		if !strings.Contains(xml, "<a:t>"+lbl+"</a:t>") {
			t.Errorf("flow missing step label %q", lbl)
		}
	}
	if strings.Contains(xml, "<p:pic>") {
		t.Errorf("flow unexpectedly contains a pic shape")
	}
}

// TestFlowConnectorKinds is criteria 2–5: cycle appends a return arrow; vertical
// rotates connectors (down arrows); arrow_dashed = dashed line + chevron; plus =
// mathPlus glyph.
func TestFlowConnectorKinds(t *testing.T) {
	cases := []struct {
		name      string
		connector scene.ConnectorKind
		orient    scene.FlowOrientation
		want      []string
	}{
		{"cycle", scene.ConnectorCycle, scene.FlowHorizontal, []string{`prst="rightArrow"`, `prst="leftArrow"`}},
		{"vertical", scene.ConnectorArrow, scene.FlowVertical, []string{`prst="downArrow"`}},
		{"dashed", scene.ConnectorArrowDashed, scene.FlowHorizontal, []string{`prst="line"`, `val="dash"`, `prst="chevron"`}},
		{"plus", scene.ConnectorPlus, scene.FlowHorizontal, []string{`prst="mathPlus"`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sc := scene.Scene{Slides: []scene.SceneSlide{{
				ID: tc.name,
				Nodes: []scene.SlideNode{scene.Flow{
					Orientation: tc.orient,
					Connector:   tc.connector,
					Steps:       flowSteps("A", "B", "C"),
				}},
			}}}
			data, _ := render(t, sc)
			xml := zipPart(t, data, "ppt/slides/slide1.xml")
			for _, want := range tc.want {
				if !strings.Contains(xml, want) {
					t.Errorf("%s flow missing %q in:\n%s", tc.name, want, xml)
				}
			}
		})
	}
}

// TestFlowIcon is criterion 6: a flow step icon places a custGeom; an unknown
// name fails Stage-1 before compose.
func TestFlowIcon(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "fi",
		Nodes: []scene.SlideNode{scene.Flow{Steps: []scene.FlowStep{
			{Label: rt("Start"), Icon: "star"},
			{Label: rt("End"), Icon: "check"},
		}}},
	}}}
	data, _ := render(t, sc)
	xml := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(xml, "<a:custGeom>") {
		t.Errorf("flow step icon did not render a custGeom shape:\n%s", xml)
	}

	bad := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bad",
		Nodes: []scene.SlideNode{scene.Flow{Steps: []scene.FlowStep{{Label: rt("x"), Icon: "nope-not-real"}}}},
	}}}
	if _, err := scene.Render(pptx.New(), bad); err == nil {
		t.Error("unknown flow step icon accepted; expected a Stage-1 render error")
	}
}

// TestFlowParallel is criterion 7: a flow scene renders byte-identically at
// workers=1 and workers=N (D-035/D-015).
func TestFlowParallel(t *testing.T) {
	mk := func() scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{
			{ID: "a", Nodes: []scene.SlideNode{scene.Flow{Connector: scene.ConnectorCycle, Steps: []scene.FlowStep{{Label: rt("One"), Icon: "star"}, {Label: rt("Two")}}}}},
			{ID: "b", Nodes: []scene.SlideNode{scene.Flow{Orientation: scene.FlowVertical, Connector: scene.ConnectorArrowDashed, Steps: flowSteps("x", "y", "z")}}},
		}}
	}
	seq, _ := render(t, mk(), scene.WithWorkers(1))
	par, _ := render(t, mk(), scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Error("flow scene render differs between workers=1 and workers=4 (idempotency broken)")
	}
}
