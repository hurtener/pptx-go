package integration

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// renderDecorationDeck renders a slide with background + foreground ornaments
// around body content; called twice to prove determinism.
func renderDecorationDeck(t *testing.T) []byte {
	t.Helper()
	sc := scene.Scene{
		Meta: scene.Metadata{Title: "Decorated"},
		Slides: []scene.SceneSlide{{
			ID: "deck",
			Nodes: []scene.SlideNode{
				scene.Decoration{Kind: scene.DecorationPreset, Preset: "radial_glow", Layer: scene.LayerBackground, Anchor: scene.AnchorTopRight, Opacity: 0.4, Bleed: true},
				scene.Decoration{Kind: scene.DecorationPreset, Preset: "grid_dots", Layer: scene.LayerBackground, Anchor: scene.AnchorBottomLeft},
				scene.Heading{Text: scene.RichText{{Text: "Quarterly Review"}}, Level: 1},
				scene.Decoration{Kind: scene.DecorationPreset, Preset: "chevron_arrow", Layer: scene.LayerForeground, Anchor: scene.AnchorCenterRight, Rotation: 90},
			},
		}},
	}
	pres := pptx.New()
	if _, err := scene.Render(pres, sc); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestDecoration_DeckConformsAndDeterministic gates the Phase-13 PR #2 seam: a
// layered-decoration deck (gradients + dots + rotated chevron) is conformant,
// re-renders byte-identically (D-035), and reopens.
func TestDecoration_DeckConformsAndDeterministic(t *testing.T) {
	data := renderDecorationDeck(t)

	rep, err := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("decoration deck failed conformance:\n%s", rep)
	}

	if again := renderDecorationDeck(t); !bytes.Equal(data, again) {
		t.Fatalf("decoration render is not byte-identical (%d vs %d bytes)", len(data), len(again))
	}

	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	for _, want := range []string{`<a:gradFill>`, `prst="ellipse"`, `prst="chevron"`, `rot="5400000"`} {
		if !strings.Contains(slide, want) {
			t.Errorf("decoration slide missing %q", want)
		}
	}

	if reopened, err := pptx.NewFromBytes(data); err != nil {
		t.Fatalf("reopen decoration deck: %v", err)
	} else if reopened.SlideCount() != 1 {
		t.Errorf("reopened slide count = %d, want 1", reopened.SlideCount())
	}
}
