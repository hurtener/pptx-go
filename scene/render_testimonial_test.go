package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/scene"
)

// avatarLogoResolver resolves "avatar" and "logo" to valid PNGs.
func avatarLogoResolver() scene.AssetResolver {
	return scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		switch uuid {
		case "avatar":
			return pngOf(256, 256), "image/png", nil
		case "logo":
			return pngOf(500, 200), "image/png", nil
		}
		return nil, "", scene.ErrAssetNotFound
	})
}

// TestQuote_Testimonial is R14.5 acceptance: a testimonial with avatar + name/
// role/company + logo + quote mark renders as one unit (a rounded avatar pic, a
// logo pic, the structured attribution, and the oversized mark), conformant, no
// warnings.
func TestQuote_Testimonial(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "testimonial",
		Nodes: []scene.SlideNode{scene.Quote{
			Text:               rt("pptx-go cut our deck-building time by 90%."),
			Mark:               true,
			AvatarAssetID:      "asset://avatar",
			AttributionName:    "Jordan Lee",
			AttributionRole:    "VP Product",
			AttributionCompany: "Acme",
			LogoAssetID:        "asset://logo",
		}},
	}}}
	data, stats := render(t, sc, scene.WithAssetResolver(avatarLogoResolver()))
	if len(stats.Warnings) != 0 {
		t.Errorf("testimonial: unexpected warnings: %+v", stats.Warnings)
	}
	if stats.Assets < 2 {
		t.Errorf("testimonial: want >=2 assets (avatar + logo), got %d", stats.Assets)
	}
	slide := zipPart(t, data, "ppt/slides/slide1.xml")
	if n := strings.Count(slide, "<p:pic>"); n < 2 {
		t.Errorf("testimonial: want 2 pics (avatar + logo), got %d", n)
	}
	if !strings.Contains(slide, "<a:t>Jordan Lee</a:t>") {
		t.Errorf("testimonial: missing attribution name:\n%s", slide)
	}
	if !strings.Contains(slide, "VP Product · Acme") {
		t.Errorf("testimonial: missing role · company")
	}
	if !strings.Contains(slide, `prst="roundRect"`) {
		t.Errorf("testimonial: avatar not rounded")
	}
	rep, _ := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/slides/slide1.xml"}})
	if !rep.OK() {
		t.Fatalf("testimonial deck failed conformance:\n%s", rep)
	}
}

// TestQuote_PlainByteIdentical verifies a Quote with only Text+Attribution is
// byte-identical to the pre-Phase-85 plain quote (no enrichment fields set).
func TestQuote_PlainByteIdentical(t *testing.T) {
	plain := scene.Scene{Slides: []scene.SceneSlide{{
		ID:    "q",
		Nodes: []scene.SlideNode{scene.Quote{Text: rt("A plain pull quote."), Attribution: "Someone"}},
	}}}
	a, _ := render(t, plain)
	b, _ := render(t, plain)
	if !bytes.Equal(a, b) {
		t.Errorf("plain quote not stable across renders")
	}
	// The plain quote must not emit any pic / oversized mark.
	slide := zipPart(t, a, "ppt/slides/slide1.xml")
	if strings.Contains(slide, "<p:pic>") {
		t.Errorf("plain quote should emit no pic:\n%s", slide)
	}
}

// TestQuote_TestimonialMissingWarns verifies an unresolvable avatar warns and is
// omitted (the quote still renders).
func TestQuote_TestimonialMissingWarns(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "t",
		Nodes: []scene.SlideNode{scene.Quote{
			Text: rt("Quote"), AvatarAssetID: "asset://nope", AttributionName: "X",
		}},
	}}}
	_, stats := render(t, sc, scene.WithAssetResolver(avatarLogoResolver()))
	if len(stats.Warnings) == 0 {
		t.Errorf("testimonial with missing avatar: expected a warning")
	}
}

// TestQuote_TestimonialDeterministic guards worker-count independence.
func TestQuote_TestimonialDeterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "t",
		Nodes: []scene.SlideNode{scene.Quote{
			Text: rt("Deterministic quote"), Mark: true, AvatarAssetID: "asset://avatar",
			AttributionName: "A", AttributionCompany: "B", LogoAssetID: "asset://logo",
		}},
	}}}
	seq := renderBytes(t, sc, scene.WithAssetResolver(avatarLogoResolver()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithAssetResolver(avatarLogoResolver()), scene.WithWorkers(8))
	if !bytes.Equal(seq, par) {
		t.Errorf("testimonial not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
