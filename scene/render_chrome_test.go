package scene

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// Slide chrome (Phase 24, R3). Tests cover the page-number total derivation, the
// body-region shrink, the eyebrow/footer rendering and page numbering, the
// brand text/asset slot, byte-identity when disabled, and determinism.

func TestChromeTotalFor(t *testing.T) {
	s := Scene{Slides: make([]SceneSlide, 5)}
	if got := chromeTotalFor(s); got != 5 {
		t.Errorf("default total: want len(slides)=5, got %d", got)
	}
	s.Chrome.Total = 11
	if got := chromeTotalFor(s); got != 11 {
		t.Errorf("explicit total: want 11, got %d", got)
	}
}

func TestChromeNeedsSerial(t *testing.T) {
	cases := []struct {
		name string
		c    Chrome
		want bool
	}{
		{"disabled", Chrome{BrandAsset: "logo"}, false},
		{"enabled text only", Chrome{Enabled: true, Brand: "ACME"}, false},
		{"enabled brand asset", Chrome{Enabled: true, BrandAsset: "logo"}, true},
	}
	for _, c := range cases {
		if got := chromeNeedsSerial(Scene{Chrome: c.c}); got != c.want {
			t.Errorf("%s: want %v, got %v", c.name, c.want, got)
		}
	}
}

// TestBodyRegion_ChromeShrinks is acceptance criterion 2: chrome shrinks the
// body region (top down, bottom up) but leaves the horizontal extent unchanged.
func TestBodyRegion_ChromeShrinks(t *testing.T) {
	r := newTestRenderer(t)
	full := r.bodyRegion() // chrome disabled (zero value)
	r.chrome = Chrome{Enabled: true}
	shrunk := r.bodyRegion()

	if shrunk.Y <= full.Y {
		t.Errorf("top did not shrink down: full.Y=%d shrunk.Y=%d", full.Y, shrunk.Y)
	}
	if shrunk.Bottom() >= full.Bottom() {
		t.Errorf("bottom did not shrink up: full.Bottom=%d shrunk.Bottom=%d", full.Bottom(), shrunk.Bottom())
	}
	if shrunk.X != full.X || shrunk.W != full.W {
		t.Errorf("horizontal extent changed: full(X=%d,W=%d) shrunk(X=%d,W=%d)", full.X, full.W, shrunk.X, shrunk.W)
	}
}

// TestChrome_FooterAndEyebrow is acceptance criteria 1 & 3: each slide gets an
// "N / total" page number; a slide with a Section gets the eyebrow label, one
// without does not; the brand text appears.
func TestChrome_FooterAndEyebrow(t *testing.T) {
	sc := Scene{
		Chrome: Chrome{Enabled: true, Brand: "ACME"},
		Slides: []SceneSlide{
			{ID: "s1", Section: "DIRECTION", Nodes: []SlideNode{Heading{Text: RichText{{Text: "One"}}, Level: 1}}},
			{ID: "s2", Nodes: []SlideNode{Heading{Text: RichText{{Text: "Two"}}, Level: 1}}},
			{ID: "s3", Section: "CLOSE", Nodes: []SlideNode{Heading{Text: RichText{{Text: "Three"}}, Level: 1}}},
		},
	}
	pres := pptx.New()
	if _, err := Render(pres, sc); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	s1 := alignZipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(s1, "1 / 3") {
		t.Errorf("slide1: missing page number '1 / 3'")
	}
	if !strings.Contains(s1, "DIRECTION") {
		t.Errorf("slide1: missing section eyebrow 'DIRECTION'")
	}
	if !strings.Contains(s1, "ACME") {
		t.Errorf("slide1: missing brand 'ACME'")
	}

	s2 := alignZipPart(t, data, "ppt/slides/slide2.xml")
	if !strings.Contains(s2, "2 / 3") {
		t.Errorf("slide2: missing page number '2 / 3'")
	}
	if strings.Contains(s2, "DIRECTION") || strings.Contains(s2, "CLOSE") {
		t.Errorf("slide2 has no Section, should carry no eyebrow label")
	}

	s3 := alignZipPart(t, data, "ppt/slides/slide3.xml")
	if !strings.Contains(s3, "3 / 3") {
		t.Errorf("slide3: missing page number '3 / 3'")
	}
}

// TestChrome_PageNumberOverrideAndTotal is acceptance criterion 3: a per-slide
// PageNumber overrides the default position, and an explicit Total overrides the
// slide count.
func TestChrome_PageNumberOverrideAndTotal(t *testing.T) {
	sc := Scene{
		Chrome: Chrome{Enabled: true, Total: 12},
		Slides: []SceneSlide{
			{ID: "cover", PageNumber: 0, Nodes: []SlideNode{Heading{Text: RichText{{Text: "Cover"}}, Level: 1}}},
			{ID: "body", PageNumber: 7, Nodes: []SlideNode{Heading{Text: RichText{{Text: "Body"}}, Level: 1}}},
		},
	}
	pres := pptx.New()
	if _, err := Render(pres, sc); err != nil {
		t.Fatalf("Render: %v", err)
	}
	data, _ := pres.WriteToBytes()
	if s1 := alignZipPart(t, data, "ppt/slides/slide1.xml"); !strings.Contains(s1, "1 / 12") {
		t.Errorf("slide1: default page number should be 1, total 12 ('1 / 12')")
	}
	if s2 := alignZipPart(t, data, "ppt/slides/slide2.xml"); !strings.Contains(s2, "7 / 12") {
		t.Errorf("slide2: overridden page number should be 7 ('7 / 12')")
	}
}

// TestChrome_BrandAssetResolvesAndWarns is acceptance criterion 4: a brand asset
// renders as an image (asset counted); an unresolved brand asset warns and does
// not fail the render.
func TestChrome_BrandAssetResolvesAndWarns(t *testing.T) {
	// Resolver supplies a tiny valid PNG for "logo", nothing else.
	sc := Scene{
		Chrome: Chrome{Enabled: true, BrandAsset: "logo"},
		Slides: []SceneSlide{{ID: "s1", Nodes: []SlideNode{Heading{Text: RichText{{Text: "H"}}, Level: 1}}}},
	}
	stats, err := Render(pptx.New(), sc, WithAssetResolver(pngResolver{}))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if stats.Assets == 0 {
		t.Errorf("brand asset: want Assets > 0, got %d", stats.Assets)
	}

	// Missing asset → warn, not error.
	scMissing := Scene{
		Chrome: Chrome{Enabled: true, BrandAsset: "absent"},
		Slides: []SceneSlide{{ID: "s1", Nodes: []SlideNode{Heading{Text: RichText{{Text: "H"}}, Level: 1}}}},
	}
	statsM, err := Render(pptx.New(), scMissing, WithAssetResolver(pngResolver{}))
	if err != nil {
		t.Fatalf("Render (missing brand asset) should not error: %v", err)
	}
	if !hasWarning(statsM, "s1", "brand asset") {
		t.Errorf("unresolved brand asset should warn; got %v", statsM.Warnings)
	}
}

// TestChrome_DisabledByteIdentical is acceptance criterion 5: a chrome-disabled
// scene renders byte-identical to the same scene with no chrome fields set.
func TestChrome_DisabledByteIdentical(t *testing.T) {
	mk := func(chrome Chrome) []byte {
		sc := Scene{
			Chrome: chrome,
			Slides: []SceneSlide{{ID: "s1", Nodes: []SlideNode{Heading{Text: RichText{{Text: "Plain"}}, Level: 1}}}},
		}
		pres := pptx.New()
		if _, err := Render(pres, sc); err != nil {
			t.Fatalf("Render: %v", err)
		}
		b, err := pres.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return b
	}
	bare := mk(Chrome{})                                            // no chrome
	disabled := mk(Chrome{Enabled: false, Brand: "ACME", Total: 9}) // fields set but disabled
	if !bytes.Equal(bare, disabled) {
		t.Errorf("chrome disabled is not byte-identical to no chrome (%d vs %d bytes)", len(disabled), len(bare))
	}
}

// TestChrome_Deterministic is acceptance criterion 6: a chrome deck renders
// byte-identically across worker counts, for both brand text and brand asset
// (the latter forces sequential composition).
func TestChrome_Deterministic(t *testing.T) {
	for _, tc := range []struct {
		name   string
		chrome Chrome
		opts   []RenderOption
	}{
		{"brand text", Chrome{Enabled: true, Brand: "ACME"}, nil},
		{"brand asset", Chrome{Enabled: true, BrandAsset: "logo"}, []RenderOption{WithAssetResolver(pngResolver{})}},
	} {
		sc := Scene{Chrome: tc.chrome}
		for i := 0; i < 16; i++ {
			sc.Slides = append(sc.Slides, SceneSlide{
				ID:      string(rune('A' + i)),
				Section: "SEC",
				Nodes:   []SlideNode{Heading{Text: RichText{{Text: "H"}}, Level: 1}},
			})
		}
		seq := chromeRenderBytes(t, sc, append(tc.opts, WithWorkers(1))...)
		par := chromeRenderBytes(t, sc, append(tc.opts, WithWorkers(8))...)
		if !bytes.Equal(seq, par) {
			t.Errorf("%s: parallel render differs from sequential (%d vs %d bytes)", tc.name, len(par), len(seq))
		}
	}
}

func chromeRenderBytes(t *testing.T, sc Scene, opts ...RenderOption) []byte {
	t.Helper()
	pres := pptx.New()
	if _, err := Render(pres, sc, opts...); err != nil {
		t.Fatalf("Render: %v", err)
	}
	b, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return b
}

func hasWarning(s Stats, slideID, substr string) bool {
	for _, w := range s.Warnings {
		if w.SlideID == slideID && strings.Contains(w.Message, substr) {
			return true
		}
	}
	return false
}

// pngResolver resolves "logo" to a minimal 1×1 PNG and everything else to a
// not-found error.
type pngResolver struct{}

func (pngResolver) Resolve(_ context.Context, id AssetID) ([]byte, string, error) {
	if id == "logo" {
		return onePixelPNG(), "image/png", nil
	}
	return nil, "", ErrAssetNotFound
}

// onePixelPNG returns the bytes of a valid 1×1 transparent PNG.
func onePixelPNG() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00,
		0x0d, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
}
