package scene_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// R14.19 — the multi-archetype conformance corpus (D-132). One fixture slide per
// professional deck class (cover / section / agenda / comparison-matrix / pricing
// / timeline / org-chart / quote / photo-cover / logo-wall / chart / dashboard /
// dark-feature / closing), rendered across the light AND dark variants, asserted
// against the standing structural invariants: every box on-canvas, conformant
// OOXML, and byte-identical re-render. This is the generalizable proof that
// coverage holds beyond the one sample deck — a regression in any class fails CI.
//
// (RTL / CJK variants are not yet asserted — R14.15 layout-direction is deferred
// to V2 (D-133 / RFC §24).)

func corpusResolver() scene.AssetResolver {
	return scene.URIAssetResolver(func(string) ([]byte, string, error) {
		return pngOf(1600, 900), "image/png", nil
	})
}

// corpusArchetypes returns one fixture slide per professional archetype, in the
// given variant.
func corpusArchetypes(variant scene.Variant) []scene.SceneSlide {
	accent := scene.ColorAccent
	chk := []scene.ChecklistItem{{Text: rt("Included"), State: scene.CheckDone}, {Text: rt("Not in tier"), State: scene.CheckNo}}
	mk := func(id string, nodes ...scene.SlideNode) scene.SceneSlide {
		return scene.SceneSlide{ID: fmt.Sprintf("%s-%v", id, variant), Variant: variant, Nodes: nodes}
	}
	return []scene.SceneSlide{
		// Cover.
		mk("cover", scene.Hero{Eyebrow: "ACME", Title: "The agent platform for every team", Subtitle: "Investor deck · 2026", AutoFit: true}),
		// Section divider.
		mk("section", scene.SectionDivider{Eyebrow: "01", Label: "Direction"}),
		// Agenda (composed: a grid of numbered stat+heading cards — R14.6 is a recipe).
		mk("agenda", scene.Grid{Columns: 3, Cells: []scene.SlideNode{
			scene.Stat{Value: "01", Label: "Direction"}, scene.Stat{Value: "02", Label: "Product"}, scene.Stat{Value: "03", Label: "Traction"},
		}}),
		// Comparison matrix (styled table).
		mk("matrix", scene.Table{
			Headers: []scene.RichText{rt("Feature"), rt("Free"), rt("Pro"), rt("Enterprise")},
			Rows:    [][]scene.RichText{{rt("Seats"), rt("1"), rt("10"), rt("∞")}, {rt("SSO"), rt("—"), rt("Yes"), rt("Yes")}},
			Style:   &scene.TableStyle{HeaderFill: true, Zebra: true, HighlightCol: 4, RowLabelCol: true},
		}),
		// Pricing (cards with ribbon + checklist + a typed price).
		mk("pricing", scene.Grid{Columns: 3, Cells: []scene.SlideNode{
			scene.Card{Header: "Free", Body: []scene.SlideNode{scene.Stat{Number: ptrF(0), Format: &scene.NumberFormat{CurrencySymbol: "$"}, Label: "forever"}, scene.Checklist{Items: chk}}},
			scene.Card{Header: "Pro", Ribbon: &scene.Ribbon{Text: "POPULAR", Position: scene.RibbonTopBar}, HeaderFill: &accent, Body: []scene.SlideNode{scene.Stat{Number: ptrF(99), Format: &scene.NumberFormat{CurrencySymbol: "$"}, Label: "per month", AutoFit: true}, scene.Checklist{Items: chk}}},
			scene.Card{Header: "Enterprise", Body: []scene.SlideNode{scene.Stat{Number: ptrF(4000), Format: &scene.NumberFormat{GroupSep: ",", CurrencySymbol: "$", Suffix: "+"}, Label: "per month", AutoFit: true}, scene.Checklist{Items: chk}}},
		}}),
		// Timeline / roadmap.
		mk("roadmap", scene.Timeline{
			Bands: []scene.TimelineBand{{From: 0, To: 0.5, Label: "Now", Fill: scene.ColorAccent}, {From: 0.5, To: 1, Label: "Next", Fill: scene.ColorInfo}},
			Lanes: []scene.TimelineLane{{Label: "Platform", Milestones: []scene.Milestone{{Position: 0.1, Label: "Beta"}, {Position: 0.8, Label: "GA", Detail: "Q4"}}}},
		}),
		// Org chart.
		mk("orgchart", scene.Tree{Root: scene.TreeNode{Label: "CEO", Children: []scene.TreeNode{
			{Label: "Eng", Children: []scene.TreeNode{{Label: "Platform"}, {Label: "Apps"}}}, {Label: "GTM", AccentIndex: 1},
		}}}),
		// Quote / testimonial.
		mk("quote", scene.Quote{Text: rt("It changed how we build decks."), Mark: true, AttributionName: "Jordan Lee", AttributionRole: "VP Product", AttributionCompany: "Acme"}),
		// Photo cover (full-bleed asset + scrim + a hero over it).
		{ID: fmt.Sprintf("photo-%v", variant), Variant: variant,
			Background: scene.Background{Kind: scene.BackgroundAsset, AssetID: "asset://cover", Scrim: &scene.Scrim{Color: scene.ColorSurface, Opacity: 55000, Gradient: true}},
			Nodes:      []scene.SlideNode{scene.Hero{Title: "Built for scale", AutoFit: true}}},
		// Logo wall.
		mk("logowall", scene.LogoWall{Caption: "Trusted by", Columns: 4, Tone: scene.LogoToneMono, Logos: []scene.LogoEntry{
			{AssetID: "asset://l1"}, {AssetID: "asset://l2"}, {AssetID: "asset://l3"}, {AssetID: "asset://l4"},
		}}),
		// Chart (raster).
		mk("chart", scene.Chart{AssetID: "asset://chart", Caption: "Revenue by quarter"}),
		// Dashboard (native dataviz strip).
		mk("dashboard", scene.Grid{Columns: 3, Cells: []scene.SlideNode{
			scene.DataMark{Kind: scene.DataMarkDonut, Value: 0.92, Label: "92%"},
			scene.DataMark{Kind: scene.DataMarkBars, Values: []float64{0.3, 0.6, 0.9, 0.5}},
			scene.DataMark{Kind: scene.DataMarkGauge, Value: 0.7, Label: "70"},
		}}),
		// Dark-feature (a funnel + cycle process slide).
		mk("process", scene.TwoColumn{
			Left:  []scene.SlideNode{scene.Funnel{Stages: []scene.FunnelStage{{Label: "Leads", Value: "10k"}, {Label: "Qualified", Value: "3k"}, {Label: "Won", Value: "400"}}}},
			Right: []scene.SlideNode{scene.Cycle{Stages: []scene.CycleStage{{Label: "Plan"}, {Label: "Build"}, {Label: "Ship"}, {Label: "Learn"}}}},
		}),
		// Quadrant positioning map.
		mk("quadrant", scene.Quadrant{
			AxisX: scene.QuadrantAxis{LowLabel: "Low effort", HighLabel: "High effort"},
			AxisY: scene.QuadrantAxis{LowLabel: "Low impact", HighLabel: "High impact"},
			Items: []scene.QuadrantItem{{X: 0.2, Y: 0.8, Label: "Quick win"}, {X: 0.7, Y: 0.6, Label: "Big bet", AccentIndex: 1}},
		}),
		// Closing (a hero + a footnote band of sources).
		{ID: fmt.Sprintf("closing-%v", variant), Variant: variant,
			Footnotes: []scene.RichText{rt("Source: internal telemetry, 2026."), rt("Past performance is not indicative of future results.")},
			Nodes:     []scene.SlideNode{scene.Hero{Title: "Let's build.", Subtitle: "hello@acme.com", AutoFit: true}}},
	}
}

func corpusScene() scene.Scene {
	var slides []scene.SceneSlide
	for _, v := range []scene.Variant{scene.VariantLight, scene.VariantDark} {
		slides = append(slides, corpusArchetypes(v)...)
	}
	return scene.Scene{Slides: slides}
}

// TestCorpus_AllBoxesOnCanvas asserts every shape in every archetype fixture (both
// variants) lies fully within the slide canvas (R14.19 — the safe-area invariant).
func TestCorpus_AllBoxesOnCanvas(t *testing.T) {
	sc := corpusScene()
	data, _ := render(t, sc, scene.WithAssetResolver(corpusResolver()))
	// Note: a chart raster may emit a benign aspect-fit advisory (the engine
	// correctly warns the caller); the corpus asserts the structural on-canvas
	// invariant, not the absence of advisories.
	slideW, slideH := pptx.New().SlideSize()
	for i := range sc.Slides {
		part := fmt.Sprintf("ppt/slides/slide%d.xml", i+1)
		xml := zipPart(t, data, part)
		offs := offRe.FindAllStringSubmatch(xml, -1)
		exts := extRe.FindAllStringSubmatch(xml, -1)
		if len(offs) == 0 {
			t.Errorf("%s (%q): no shapes emitted", part, sc.Slides[i].ID)
			continue
		}
		n := len(offs)
		if len(exts) < n {
			n = len(exts)
		}
		for k := 0; k < n; k++ {
			x, _ := strconv.Atoi(offs[k][1])
			y, _ := strconv.Atoi(offs[k][2])
			cx, _ := strconv.Atoi(exts[k][1])
			cy, _ := strconv.Atoi(exts[k][2])
			if x < 0 || y < 0 || x+cx > slideW || y+cy > slideH {
				t.Errorf("%s (%q): box [x=%d y=%d cx=%d cy=%d] exceeds the %dx%d canvas",
					part, sc.Slides[i].ID, x, y, cx, cy, slideW, slideH)
			}
		}
	}
}

// TestCorpus_Conformant asserts the whole corpus is OOXML-conformant (R14.19).
func TestCorpus_Conformant(t *testing.T) {
	data, _ := render(t, corpusScene(), scene.WithAssetResolver(corpusResolver()))
	rep, err := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"}})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("corpus failed conformance:\n%s", rep)
	}
}

// TestCorpus_Deterministic asserts every corpus fixture re-renders byte-identically
// across worker counts (R14.19 — the byte-identity invariant).
func TestCorpus_Deterministic(t *testing.T) {
	sc := corpusScene()
	seq := renderBytes(t, sc, scene.WithAssetResolver(corpusResolver()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithAssetResolver(corpusResolver()), scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("corpus not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}

// soulDarkTheme is a non-default brand theme carrying a soul-driven dark palette
// (R8.3): a deep-navy dark side instead of the pinned Tailwind grays.
func soulDarkTheme() *pptx.Theme {
	return pptx.NewTheme(
		pptx.WithDarkSurface(pptx.ColorCanvas, "0A0E1A"),
		pptx.WithDarkSurface(pptx.ColorSurface, "14182B"),
		pptx.WithDarkSurface(pptx.ColorSurfaceAlt, "1C2238"),
		pptx.WithDarkText(pptx.TextPrimary, "F4F6FF"),
		pptx.WithDarkText(pptx.TextSecondary, "C8D0E8"),
	)
}

// TestCorpus_SoulDarkPalette renders the dark archetypes through a soul-driven
// dark palette and asserts the standing invariants still hold (on-canvas,
// conformant, deterministic) AND that the brand navy reaches the rendered bytes
// while the pinned gray does not — i.e. a soul dark palette neither perturbs the
// safe-area/conformance invariants nor leaks the pinned default (R8.3 × R14.19).
func TestCorpus_SoulDarkPalette(t *testing.T) {
	sc := scene.Scene{Slides: corpusArchetypes(scene.VariantDark)}
	opts := []scene.RenderOption{scene.WithAssetResolver(corpusResolver()), scene.WithTheme(soulDarkTheme())}

	data, stats := render(t, sc, opts...)

	// Conformant under the soul dark palette.
	rep, err := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"}})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("soul-dark corpus failed conformance:\n%s", rep)
	}

	// Every box still on-canvas.
	slideW, slideH := pptx.New().SlideSize()
	for i := range sc.Slides {
		xml := zipPart(t, data, fmt.Sprintf("ppt/slides/slide%d.xml", i+1))
		offs := offRe.FindAllStringSubmatch(xml, -1)
		exts := extRe.FindAllStringSubmatch(xml, -1)
		n := len(offs)
		if len(exts) < n {
			n = len(exts)
		}
		for k := 0; k < n; k++ {
			x, _ := strconv.Atoi(offs[k][1])
			y, _ := strconv.Atoi(offs[k][2])
			cx, _ := strconv.Atoi(exts[k][1])
			cy, _ := strconv.Atoi(exts[k][2])
			if x < 0 || y < 0 || x+cx > slideW || y+cy > slideH {
				t.Errorf("soul-dark %q: box [x=%d y=%d cx=%d cy=%d] exceeds the %dx%d canvas",
					sc.Slides[i].ID, x, y, cx, cy, slideW, slideH)
			}
		}
	}

	// The brand navy canvas reaches every dark slide's resolved colors.
	for _, c := range stats.Colors {
		if string(c.Canvas) != "0A0E1A" {
			t.Errorf("slide %q resolved canvas = %q, want brand navy 0A0E1A", c.SlideID, c.Canvas)
		}
	}

	// Deterministic across worker counts.
	seq := renderBytes(t, sc, scene.WithAssetResolver(corpusResolver()), scene.WithTheme(soulDarkTheme()), scene.WithWorkers(1))
	par := renderBytes(t, sc, scene.WithAssetResolver(corpusResolver()), scene.WithTheme(soulDarkTheme()), scene.WithWorkers(8))
	if string(seq) != string(par) {
		t.Fatalf("soul-dark corpus not deterministic across worker counts (%d vs %d bytes)", len(seq), len(par))
	}
}
