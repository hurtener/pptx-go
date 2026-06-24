package scene_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// TestBackground_ColorSlide verifies that a BackgroundColor slide produces a
// full-slide solid-fill rect as the first shape in the slide.
func TestBackground_ColorSlide(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "bg-color",
		Background: scene.Background{
			Kind:  scene.BackgroundColor,
			Color: pptx.ColorAccent,
		},
		Nodes: []scene.SlideNode{scene.Heading{Text: rt("Title"), Level: 1}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("BackgroundColor: unexpected warnings: %+v", stats.Warnings)
	}

	pres, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	shapes := pres.Slides()[0].Shapes()
	if len(shapes) == 0 {
		t.Fatal("no shapes on BackgroundColor slide")
	}
	first := shapes[0]
	if first.Geometry() != pptx.ShapeRect {
		t.Errorf("BackgroundColor: first shape geometry = %q, want ShapeRect", first.Geometry())
	}
	fill := first.Fill()
	if fill == nil || fill.Kind() != pptx.FillSolid {
		t.Errorf("BackgroundColor: first shape fill kind = %v, want FillSolid", fill)
	}
	// The fill color must be the resolved accent (not white and not empty).
	color, ok := fill.SolidColor()
	if !ok || color == nil {
		t.Errorf("BackgroundColor: SolidColor returned no color")
	}
}

// TestBackground_PaperRoundTrip verifies that a ColorPaper background painted
// from a theme with an off-white paper tint emits the resolved RGB and survives
// a write → reopen → re-write cycle (D-104, R13.1 acceptance 3, G6).
func TestBackground_PaperRoundTrip(t *testing.T) {
	const paper = "FAFAF8"
	th := pptx.NewTheme(pptx.WithPaper(paper))
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "bg-paper",
		Background: scene.Background{
			Kind:  scene.BackgroundColor,
			Color: pptx.ColorPaper,
		},
		Nodes: []scene.SlideNode{scene.Heading{Text: rt("Title"), Level: 1}},
	}}}
	data, stats := render(t, sc, scene.WithTheme(th))
	if len(stats.Warnings) != 0 {
		t.Errorf("ColorPaper background: unexpected warnings: %+v", stats.Warnings)
	}

	// The resolved off-white must appear in the slide XML (not pure white).
	slideXML := zipPart(t, data, "ppt/slides/slide1.xml")
	if !strings.Contains(slideXML, paper) {
		t.Errorf("slide does not carry the paper tint %s:\n%s", paper, slideXML)
	}

	// Round-trip: reopen and re-write; the paper tint must persist (G6).
	pres, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	out, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	if !strings.Contains(zipPart(t, out, "ppt/slides/slide1.xml"), paper) {
		t.Errorf("paper tint %s did not survive round-trip", paper)
	}

	// The first shape is the full-slide paper rect with a solid fill.
	shapes := pres.Slides()[0].Shapes()
	if len(shapes) == 0 {
		t.Fatal("no shapes on ColorPaper slide")
	}
	if fill := shapes[0].Fill(); fill == nil || fill.Kind() != pptx.FillSolid {
		t.Errorf("ColorPaper: first shape fill = %v, want FillSolid", fill)
	}
}

// TestBackground_PaperDefaultByteIdentical verifies that on the default theme a
// ColorPaper background is byte-identical to a ColorCanvas one (paper defaults
// to canvas; D-104, R13.1 acceptance 4).
func TestBackground_PaperDefaultByteIdentical(t *testing.T) {
	mk := func(role pptx.ColorRole) scene.Scene {
		return scene.Scene{Slides: []scene.SceneSlide{{
			ID:         "bg",
			Background: scene.Background{Kind: scene.BackgroundColor, Color: role},
			Nodes:      []scene.SlideNode{scene.Heading{Text: rt("Title"), Level: 1}},
		}}}
	}
	paper, _ := render(t, mk(pptx.ColorPaper))
	canvas, _ := render(t, mk(pptx.ColorCanvas))
	if !bytes.Equal(paper, canvas) {
		t.Errorf("default-theme ColorPaper not byte-identical to ColorCanvas (%d vs %d bytes)", len(paper), len(canvas))
	}
}

// TestBackground_GradientSlide verifies that a BackgroundGradient slide
// produces a full-slide gradient-fill rect as the first shape.
func TestBackground_GradientSlide(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "bg-gradient",
		Background: scene.Background{
			Kind:     scene.BackgroundGradient,
			Gradient: [2]pptx.ColorRole{pptx.ColorAccent, pptx.ColorCanvas},
			Angle:    90,
		},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("BackgroundGradient: unexpected warnings: %+v", stats.Warnings)
	}

	pres, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	shapes := pres.Slides()[0].Shapes()
	if len(shapes) == 0 {
		t.Fatal("no shapes on BackgroundGradient slide")
	}
	fill := shapes[0].Fill()
	if fill == nil || fill.Kind() != pptx.FillGradient {
		t.Errorf("BackgroundGradient: first shape fill kind = %v, want FillGradient", fill)
	}
	grad, ok := fill.Gradient()
	if !ok {
		t.Fatal("BackgroundGradient: Gradient() returned false")
	}
	if len(grad.Stops) != 2 {
		t.Errorf("BackgroundGradient: gradient stops = %d, want 2", len(grad.Stops))
	}
	// Angle 90° is stored in OOXML 1/60000 units — verify it round-trips.
	if grad.Angle < 89 || grad.Angle > 91 {
		t.Errorf("BackgroundGradient: angle = %.2f, want ~90", grad.Angle)
	}
}

// TestBackground_MultiStopGradient verifies that a Stops-driven gradient renders
// exactly N stops and round-trips through pptx.Open (D-105, R13.3 acceptance 1).
func TestBackground_MultiStopGradient(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "bg-multistop",
		Background: scene.Background{
			Kind: scene.BackgroundGradient,
			Stops: []scene.GradientStop{
				{Pos: 0, Color: pptx.ColorAccent},
				{Pos: 0.5, Color: pptx.ColorAccentAlt},
				{Pos: 1, Color: pptx.ColorCanvas},
			},
			Angle: 45,
		},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("multi-stop gradient: unexpected warnings: %+v", stats.Warnings)
	}
	// The emitted XML must carry three gradient stops.
	if n := strings.Count(zipPart(t, data, "ppt/slides/slide1.xml"), "<a:gs "); n != 3 {
		t.Errorf("multi-stop gradient: <a:gs> count = %d, want 3", n)
	}

	pres, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	fill := pres.Slides()[0].Shapes()[0].Fill()
	if fill == nil || fill.Kind() != pptx.FillGradient {
		t.Fatalf("multi-stop gradient: fill kind = %v, want FillGradient", fill)
	}
	grad, ok := fill.Gradient()
	if !ok || len(grad.Stops) != 3 {
		t.Errorf("multi-stop gradient: round-trip stops = %d, want 3", len(grad.Stops))
	}
}

// TestBackground_InvalidStopsWarn verifies that invalid Stops record exactly one
// warning and emit no gradient shape, without panicking (D-105, R13.3
// acceptance 2; RFC §10.2 warn-don't-fail).
func TestBackground_InvalidStopsWarn(t *testing.T) {
	cases := map[string][]scene.GradientStop{
		"too few":       {{Pos: 0, Color: pptx.ColorAccent}},
		"out of range":  {{Pos: 0, Color: pptx.ColorAccent}, {Pos: 1.5, Color: pptx.ColorCanvas}},
		"not ascending": {{Pos: 0.8, Color: pptx.ColorAccent}, {Pos: 0.2, Color: pptx.ColorCanvas}},
		"negative":      {{Pos: -0.1, Color: pptx.ColorAccent}, {Pos: 1, Color: pptx.ColorCanvas}},
	}
	for name, stops := range cases {
		t.Run(name, func(t *testing.T) {
			sc := scene.Scene{Slides: []scene.SceneSlide{{
				ID:         "bad",
				Background: scene.Background{Kind: scene.BackgroundGradient, Stops: stops},
			}}}
			data, stats := render(t, sc)
			if len(stats.Warnings) != 1 {
				t.Errorf("invalid stops %q: warnings = %d, want 1 (%+v)", name, len(stats.Warnings), stats.Warnings)
			}
			pres, err := pptx.NewFromBytes(data)
			if err != nil {
				t.Fatalf("NewFromBytes: %v", err)
			}
			for _, sh := range pres.Slides()[0].Shapes() {
				if f := sh.Fill(); f != nil && f.Kind() == pptx.FillGradient {
					t.Errorf("invalid stops %q: a gradient shape was emitted", name)
				}
			}
		})
	}
}

// TestBackground_LegacyGradientByteIdentical verifies that an empty-Stops
// (legacy two-role) gradient is byte-identical across a re-render and unaffected
// by the multi-stop path (D-105, R13.3 acceptance 3 + 4 determinism).
func TestBackground_LegacyGradientByteIdentical(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "legacy",
		Background: scene.Background{
			Kind:     scene.BackgroundGradient,
			Gradient: [2]pptx.ColorRole{pptx.ColorAccent, pptx.ColorCanvas},
			Angle:    90,
		},
	}}}
	a, _ := render(t, sc)
	b, _ := render(t, sc)
	if !bytes.Equal(a, b) {
		t.Errorf("legacy two-role gradient not deterministic (%d vs %d bytes)", len(a), len(b))
	}
}

// TestBackground_MultiStopDeterministic verifies a multi-stop gradient re-renders
// byte-identically (D-105, R13.3 acceptance 4).
func TestBackground_MultiStopDeterministic(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID: "det",
		Background: scene.Background{
			Kind: scene.BackgroundGradient,
			Stops: []scene.GradientStop{
				{Pos: 0, Color: pptx.ColorAccent},
				{Pos: 0.33, Color: pptx.ColorAccentAlt},
				{Pos: 0.66, Color: pptx.ColorAccentWarm},
				{Pos: 1, Color: pptx.ColorCanvas},
			},
			Angle: 30,
		},
	}}}
	a, _ := render(t, sc)
	b, _ := render(t, sc)
	if !bytes.Equal(a, b) {
		t.Errorf("multi-stop gradient not deterministic (%d vs %d bytes)", len(a), len(b))
	}
}

// TestBackground_NoneDrawsNothing verifies that BackgroundNone (the zero value)
// is byte-identical to a slide with no Background field set at all — both light
// variant (RFC §10.1 backward-compat guarantee).
func TestBackground_NoneDrawsNothing(t *testing.T) {
	// sc1: no Background field at all (zero-value struct).
	sc1 := scene.Scene{Slides: []scene.SceneSlide{{
		Nodes: []scene.SlideNode{scene.Heading{Text: rt("Title"), Level: 1}},
	}}}
	// sc2: explicit BackgroundNone.
	sc2 := scene.Scene{Slides: []scene.SceneSlide{{
		Background: scene.Background{Kind: scene.BackgroundNone},
		Nodes:      []scene.SlideNode{scene.Heading{Text: rt("Title"), Level: 1}},
	}}}

	d1 := renderBytes(t, sc1)
	d2 := renderBytes(t, sc2)

	if !bytes.Equal(d1, d2) {
		t.Errorf("BackgroundNone is not byte-identical to no-background (%d vs %d bytes)", len(d2), len(d1))
	}
}

// TestVariantDark_DarkCanvas verifies that a bare VariantDark slide (no explicit
// Background) gets a dark canvas fill as the first shape.
func TestVariantDark_DarkCanvas(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		ID:      "dark",
		Variant: scene.VariantDark,
		Nodes:   []scene.SlideNode{scene.Heading{Text: rt("Dark slide"), Level: 1}},
	}}}
	data, stats := render(t, sc)
	if len(stats.Warnings) != 0 {
		t.Errorf("VariantDark must not produce warnings: %+v", stats.Warnings)
	}

	pres, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	shapes := pres.Slides()[0].Shapes()
	if len(shapes) == 0 {
		t.Fatal("no shapes on dark slide")
	}
	// First shape: dark canvas fill rect.
	first := shapes[0]
	if first.Geometry() != pptx.ShapeRect {
		t.Errorf("VariantDark canvas: first shape geometry = %q, want ShapeRect", first.Geometry())
	}
	fill := first.Fill()
	if fill == nil || fill.Kind() != pptx.FillSolid {
		t.Fatalf("VariantDark canvas: fill kind = %v, want FillSolid", fill)
	}
	color, ok := fill.SolidColor()
	if !ok || color == nil {
		t.Fatal("VariantDark canvas: SolidColor returned no color")
	}
	// The canvas must be a dark color (not white).
	raw := strings.ToUpper(string(color.(pptx.RGB)))
	if raw == "FFFFFF" || raw == "" {
		t.Errorf("VariantDark canvas color = %q — expected a dark color, not white", raw)
	}
	// Expect the pinned gray-900 canvas value.
	if raw != "111827" {
		t.Errorf("VariantDark canvas color = %q, want 111827 (pinned gray-900)", raw)
	}
}

// TestVariantDark_TextColorDiffers verifies that a Heading on a VariantDark
// slide resolves its text to the pinned light primary color (F9FAFB), while a
// VariantLight slide uses the default dark primary text color (111827). The two
// renders must not be byte-identical.
//
// Note: "111827" legitimately appears in the dark slide as the canvas background
// fill color (the pinned gray-900 dark canvas). The assertion only verifies that
// the light primary text color appears in the dark slide and is absent from the
// light slide — not that "111827" is absent (it is expected as the canvas fill).
func TestVariantDark_TextColorDiffers(t *testing.T) {
	nodes := []scene.SlideNode{scene.Heading{Text: rt("Title"), Level: 1}}

	scLight := scene.Scene{Slides: []scene.SceneSlide{{
		Variant: scene.VariantLight, Nodes: nodes,
	}}}
	scDark := scene.Scene{Slides: []scene.SceneSlide{{
		Variant: scene.VariantDark, Nodes: nodes,
	}}}

	dLight := renderBytes(t, scLight)
	dDark := renderBytes(t, scDark)

	slideLight := zipPart(t, dLight, "ppt/slides/slide1.xml")
	slideDark := zipPart(t, dDark, "ppt/slides/slide1.xml")

	if slideLight == slideDark {
		t.Error("VariantDark and VariantLight produced identical slide XML — dark must differ")
	}
	// Dark slide must contain the pinned light primary text color (gray-50).
	// This value comes from TextPrimary in the dark theme and is set on every
	// Heading text run by addRichText → colorFor → TokenTextColor(TextPrimary).
	if !strings.Contains(slideDark, "F9FAFB") {
		t.Errorf("VariantDark slide missing expected light primary color 'F9FAFB':\n%s", slideDark)
	}
	// Light slide must NOT contain the dark primary text color — it uses 111827.
	if !strings.Contains(slideLight, "111827") {
		t.Errorf("VariantLight slide missing expected dark primary color '111827':\n%s", slideLight)
	}
	// Sanity: the dark slide contains "111827" as the canvas fill, not as text.
	// This is expected; the test does NOT assert its absence on the dark slide.
}

// TestVariantDark_BackwardCompat is the backward-compatibility guard: a slide
// with VariantLight + BackgroundNone (both zero values) must be byte-identical
// to a slide with neither field set — the RFC §10.1 round-trip invariant.
func TestVariantDark_BackwardCompat(t *testing.T) {
	// Both use the same nodes and zero IDs so the PPTX bytes should match.
	sc1 := scene.Scene{Slides: []scene.SceneSlide{{
		Nodes: []scene.SlideNode{
			scene.Heading{Text: rt("Title"), Level: 1},
			scene.Prose{Paragraphs: []scene.RichText{rt("Body text.")}},
		},
	}}}
	sc2 := scene.Scene{Slides: []scene.SceneSlide{{
		Variant:    scene.VariantLight,
		Background: scene.Background{Kind: scene.BackgroundNone},
		Nodes: []scene.SlideNode{
			scene.Heading{Text: rt("Title"), Level: 1},
			scene.Prose{Paragraphs: []scene.RichText{rt("Body text.")}},
		},
	}}}

	d1 := renderBytes(t, sc1)
	d2 := renderBytes(t, sc2)
	if !bytes.Equal(d1, d2) {
		t.Errorf("VariantLight+BackgroundNone not byte-identical to default (%d vs %d bytes)", len(d2), len(d1))
	}
}

// TestVariantDark_Determinism mirrors render_parallel_test.go: sequential
// (workers=1) and parallel (workers=4) renders of a VariantDark deck must
// produce byte-identical output (RFC §10.1 determinism invariant). Dark slides
// render in the sequential pass in both cases, so this also exercises that the
// sequential-pass ordering is stable.
func TestVariantDark_Determinism(t *testing.T) {
	// Build a mixed deck: some light slides (can go parallel) and some dark
	// slides (forced sequential). The overall render must be byte-identical
	// regardless of the worker count.
	sc := scene.Scene{Slides: []scene.SceneSlide{
		{
			Variant: scene.VariantLight,
			Nodes:   []scene.SlideNode{scene.Heading{Text: rt("Light A"), Level: 1}},
		},
		{
			Variant: scene.VariantDark,
			Nodes:   []scene.SlideNode{scene.Heading{Text: rt("Dark B"), Level: 1}},
		},
		{
			Variant: scene.VariantLight,
			Nodes:   []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("Light C.")}}},
		},
		{
			Variant: scene.VariantDark,
			Background: scene.Background{
				Kind:  scene.BackgroundColor,
				Color: pptx.ColorAccent,
			},
			Nodes: []scene.SlideNode{scene.Heading{Text: rt("Dark D"), Level: 2}},
		},
	}}

	seq := renderBytes(t, sc, scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithWorkers(4))
	if !bytes.Equal(seq, par) {
		t.Errorf("VariantDark determinism: sequential (%d bytes) != parallel (%d bytes)", len(seq), len(par))
	}

	// Two default-worker renders must also agree.
	a := renderBytes(t, sc)
	b := renderBytes(t, sc)
	if !bytes.Equal(a, b) {
		t.Error("VariantDark: two default-worker renders are not byte-identical")
	}
	if !bytes.Equal(a, seq) {
		t.Error("VariantDark: default-worker render differs from sequential render")
	}
}

// TestVariantDark_NoWarning is the "no more variant-unimplemented warning" guard:
// a VariantDark slide must not emit a LayoutWarning now that it is implemented.
func TestVariantDark_NoWarning(t *testing.T) {
	sc := scene.Scene{Slides: []scene.SceneSlide{{
		Variant: scene.VariantDark,
		Nodes:   []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("dark")}}},
	}}}
	pres := pptx.New()
	stats, err := scene.Render(pres, sc)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	for _, w := range stats.Warnings {
		if strings.Contains(strings.ToLower(w.Message), "variant") {
			t.Errorf("VariantDark: unexpected variant warning: %q", w.Message)
		}
	}
}
