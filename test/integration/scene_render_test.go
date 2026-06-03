package integration

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestConformance_SceneRender gates the first scene→pptx seam (Phase 06): a
// multi-leaf scene rendered through scene.Render onto the builder produces a
// structurally sound deck (D-031). Closes the seam Phase 05 opened and Phase 06
// implements; Deps name Phases 03/04 (the builder it composes).
func TestConformance_SceneRender(t *testing.T) {
	png := append([]byte("\x89PNG\r\n\x1a\n"), []byte("shot")...)
	resolver := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		return png, "image/png", nil
	})

	sc := scene.Scene{
		Meta: scene.Metadata{Title: "Deck"},
		Slides: []scene.SceneSlide{
			{
				ID: "cover",
				Nodes: []scene.SlideNode{
					scene.Hero{Eyebrow: "2025", Title: "Annual Review", Subtitle: "Highlights"},
				},
				Notes: scene.RichText{{Text: "Open with the headline number."}},
			},
			{
				ID: "content",
				Nodes: []scene.SlideNode{
					scene.Heading{Text: scene.RichText{{Text: "Agenda"}}, Level: 2},
					scene.List{Kind: scene.ListBullet, Items: []scene.ListItem{
						{Text: scene.RichText{{Text: "Results"}}},
						{Text: scene.RichText{{Text: "Outlook"}}},
					}},
					scene.Callout{Kind: scene.CalloutTip, Body: scene.RichText{{Text: "Stay on time."}}},
					scene.CodeBlock{AssetID: "asset://snippet", Language: "go", Caption: "demo.go"},
				},
			},
		},
	}

	pres := pptx.New()
	stats, err := scene.Render(pres, sc, scene.WithAssetResolver(resolver))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if stats.Slides != 2 || stats.Assets != 1 {
		t.Fatalf("stats = %+v, want 2 slides / 1 asset", stats)
	}

	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{
			"/ppt/presentation.xml",
			"/ppt/slides/slide1.xml",
			"/ppt/slides/slide2.xml",
			"/ppt/media/image1.png",
			"/ppt/notesSlides/notesSlide1.xml",
		},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("scene-rendered deck failed conformance:\n%s", rep)
	}
}

// TestConformance_FlowIconSeam exercises the flow→icon-registry seam end-to-end
// through real internal/opc + encoding/xml (Phase 15, D-044): a flow step's
// closed-name icon resolves to a native custGeom that reaches the slide part,
// and an unknown step icon name fails Stage-1 before compose. Deps name Phase 14
// (the icon-registry wiring this consumes).
func TestConformance_FlowIconSeam(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "flow",
		Nodes: []scene.SlideNode{scene.Flow{
			Connector: scene.ConnectorCycle,
			Steps: []scene.FlowStep{
				{Label: scene.RichText{{Text: "Plan"}}, Icon: "star"},
				{Label: scene.RichText{{Text: "Ship"}}, Icon: "check"},
			},
		}},
	}}}

	pres := pptx.New()
	if _, err := scene.Render(pres, sc); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("flow-icon deck failed conformance:\n%s", rep)
	}

	// An unknown step icon name must fail Stage-1, before any compose.
	bad := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "bad",
		Nodes: []scene.SlideNode{scene.Flow{Steps: []scene.FlowStep{{Label: scene.RichText{{Text: "x"}}, Icon: "no-such-icon"}}}},
	}}}
	if _, err := scene.Render(pptx.New(), bad); err == nil {
		t.Error("unknown flow step icon accepted; expected a Stage-1 render error")
	}
}
