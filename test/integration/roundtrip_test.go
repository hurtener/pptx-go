// roundtrip_test is the Phase 18 PR#4 capstone: it walks every shipped builder
// primitive and every scene IR node through author → save → Open, asserting the
// navigable model the read accessors reconstruct equals what was authored (RFC
// §16), and that a self-authored fixture reopens byte-identically (D-035).
//
// Scene-level read is out of scope (D-047): a rendered scene is walked at the
// builder level — every node kind renders into a navigable, byte-stable deck.
package integration

import (
	"bytes"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

// fixtureBox is a reusable placement box.
var fixtureBox = pptx.Box{X: 914400, Y: 914400, W: 1828800, H: 914400}

// buildFixture authors a deck exercising every shipped builder primitive group:
// shapes (geometry / solid + gradient fill / line / shadow / rotation), rich
// text (paragraphs / styled & hyperlinked runs / bullets / alignment / level),
// a table (structure / header / banding / merge / per-cell text), and an image
// (alt / crop / fit / rotation / opacity / embedded bytes). It is the fixture
// behind both the comprehensive walk and the byte-identity check.
func buildFixture(t *testing.T) (*pptx.Presentation, []byte) {
	t.Helper()
	p := pptx.New()
	s := p.AddSlide()

	// Shapes: solid + gradient fill, line, shadow, rotation, geometry.
	s.AddShape(pptx.ShapeRoundRect, fixtureBox,
		pptx.WithFill(pptx.SolidFill(pptx.RGB("2563EB"))),
		pptx.WithLine(pptx.Line{Width: pptx.Pt(2), Color: pptx.RGB("FF0000"), Dash: "dash"}),
		pptx.WithRotation(30),
		pptx.WithShadow(pptx.Elevation{Blur: pptx.Pt(12), OffsetY: pptx.Pt(4), Color: "000000", Alpha: 35000}))
	s.AddShape(pptx.ShapeEllipse, pptx.Box{X: 3000000, Y: 914400, W: 1828800, H: 914400},
		pptx.WithFill(pptx.LinearGradient(90,
			pptx.GradientStop{Pos: 0, Color: pptx.RGB("FF0000")},
			pptx.GradientStop{Pos: 1, Color: pptx.RGB("0000FF")})))

	// Rich text: bullet, alignment, level; a styled run + a hyperlink run.
	tf := s.AddTextFrame(pptx.Box{X: 914400, Y: 2200000, W: 5000000, H: 1200000})
	par := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter, Level: 1, Bullet: pptx.BulletDisc})
	par.AddRun("styled ", pptx.RunStyle{TypeRole: pptx.TypeBody, Bold: true, Italic: true, Color: pptx.RGB("123456")})
	par.AddHyperlink("link", "https://example.com/x?a=1&b=2", pptx.RunStyle{TypeRole: pptx.TypeBody})

	// Table: header + banding, a merged cell, per-cell text.
	tbl := s.AddTable(pptx.Box{X: 914400, Y: 3600000, W: 4572000, H: 1200000}, 2, 2)
	tbl.SetColumnWidths(pptx.Cm(4), pptx.Cm(6))
	tbl.SetHeaderRow(true)
	tbl.SetBanding(true, false)
	tbl.Cell(0, 0).SetText("H1")
	tbl.Cell(0, 1).SetText("H2")
	tbl.Cell(1, 0).SetText("span").MergeRight(2)

	// Image: alt, crop, fit, rotation, opacity, embedded bytes.
	img, err := s.AddImage(pptx.ImageBytes(fixturePNG, "image/png"),
		pptx.Box{X: 6000000, Y: 914400, W: 1828800, H: 1828800})
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	img.SetAltText("a logo").SetCrop(pptx.Crop{Left: 0.1, Right: 0.05}).
		SetFit(pptx.FitNone).SetRotation(90).SetOpacity(40000)

	data, err := p.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return p, data
}

var fixturePNG = append([]byte("\x89PNG\r\n\x1a\n"), []byte("fixture-image-payload")...)

// TestRoundTrip_BuilderPrimitives is PR#4 acceptance criterion 4 (builder side):
// every shipped builder primitive reopens into the navigable model the read
// accessors reconstruct, equal to what was authored.
func TestRoundTrip_BuilderPrimitives(t *testing.T) {
	_, data := buildFixture(t)
	re, err := pptx.NewFromBytes(data)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	slides := re.Slides()
	if len(slides) != 1 {
		t.Fatalf("Slides() = %d, want 1", len(slides))
	}
	shapes := slides[0].Shapes()

	var roundRect, ellipse, text, table, image *pptx.Shape
	for _, sh := range shapes {
		switch {
		case sh.Geometry() == pptx.ShapeRoundRect:
			roundRect = sh
		case sh.Geometry() == pptx.ShapeEllipse:
			ellipse = sh
		}
		if _, ok := sh.TextFrame(); ok && text == nil && sh.Geometry() == pptx.ShapeRect {
			text = sh
		}
		if _, ok := sh.Table(); ok {
			table = sh
		}
		if _, ok := sh.Image(); ok {
			image = sh
		}
	}

	// --- Shape geometry / fill / line / shadow / rotation. ---
	if roundRect == nil {
		t.Fatal("round-rect shape not found on reopen")
	}
	if got := roundRect.Rotation(); got != 30 {
		t.Errorf("rotation = %v, want 30", got)
	}
	if f := roundRect.Fill(); f == nil || f.Kind() != pptx.FillSolid {
		t.Errorf("fill = %#v, want solid", f)
	} else if c, _ := f.SolidColor(); c != pptx.RGB("2563EB") {
		t.Errorf("fill color = %#v, want 2563EB", c)
	}
	if ln := roundRect.Line(); ln.Width != pptx.Pt(2) || ln.Color != pptx.RGB("FF0000") || ln.Dash != "dash" {
		t.Errorf("line = %+v, want Pt(2)/FF0000/dash", ln)
	}
	if _, ok := roundRect.Shadow(); !ok {
		t.Error("shadow not reconstructed")
	}

	// --- Gradient fill. ---
	if ellipse == nil {
		t.Fatal("ellipse shape not found on reopen")
	}
	if g, ok := ellipse.Fill().Gradient(); !ok || g.Radial || g.Angle != 90 || len(g.Stops) != 2 {
		t.Errorf("gradient = %+v, ok=%v, want linear 90° / 2 stops", g, ok)
	}

	// --- Rich text: run style, color, hyperlink, bullet, alignment, level. ---
	if text == nil {
		t.Fatal("text frame shape not found on reopen")
	}
	frame, _ := text.TextFrame()
	paras := frame.Paragraphs()
	if len(paras) != 1 {
		t.Fatalf("paragraphs = %d, want 1", len(paras))
	}
	if paras[0].Alignment() != pptx.AlignCenter || paras[0].Level() != 1 || paras[0].BulletStyle() != pptx.BulletDisc {
		t.Errorf("paragraph props = align %v / level %d / bullet %v",
			paras[0].Alignment(), paras[0].Level(), paras[0].BulletStyle())
	}
	runs := paras[0].Runs()
	if len(runs) != 2 {
		t.Fatalf("runs = %d, want 2", len(runs))
	}
	if !runs[0].Bold() || !runs[0].Italic() {
		t.Error("first run lost bold/italic")
	}
	if c, ok := runs[0].Color(); !ok || c != pptx.RGB("123456") {
		t.Errorf("first run color = %#v, ok=%v, want 123456", c, ok)
	}
	if url, ok := runs[1].Hyperlink(); !ok || url != "https://example.com/x?a=1&b=2" {
		t.Errorf("hyperlink = %q, ok=%v", url, ok)
	}

	// --- Table: structure, header/banding, merge, cell text. ---
	if table == nil {
		t.Fatal("table shape not found on reopen")
	}
	tbl, _ := table.Table()
	if tbl.RowCount() != 2 || tbl.ColCount() != 2 || !tbl.HeaderRow() || !tbl.RowBanding() {
		t.Errorf("table = %dx%d header=%v band=%v", tbl.RowCount(), tbl.ColCount(), tbl.HeaderRow(), tbl.RowBanding())
	}
	if tbl.Cell(1, 0).GridSpan() != 2 || !tbl.Cell(1, 1).Covered() {
		t.Error("merge not reconstructed")
	}
	if got := tbl.Cell(0, 0).TextFrame().Paragraphs()[0].Runs()[0].Text(); got != "H1" {
		t.Errorf("cell(0,0) text = %q, want H1", got)
	}

	// --- Image: alt / crop / fit / rotation / opacity / bytes. ---
	if image == nil {
		t.Fatal("image shape not found on reopen")
	}
	im, _ := image.Image()
	if im.AltText() != "a logo" || im.Fit() != pptx.FitNone || im.Rotation() != 90 || im.Opacity() != 40000 {
		t.Errorf("image props = alt %q / fit %v / rot %v / opacity %d",
			im.AltText(), im.Fit(), im.Rotation(), im.Opacity())
	}
	if got := im.Crop(); got.Left != 0.1 || got.Right != 0.05 {
		t.Errorf("crop = %+v, want Left 0.1 / Right 0.05", got)
	}
	gotBytes, err := im.Bytes()
	if err != nil {
		t.Fatalf("image Bytes(): %v", err)
	}
	if !bytes.Equal(gotBytes, fixturePNG) {
		t.Errorf("image bytes = %d, want the %d authored bytes", len(gotBytes), len(fixturePNG))
	}
}

// TestRoundTrip_FixtureByteIdentical is PR#4 acceptance criterion 5: a
// self-authored fixture deck reopens byte-identically — save → Open → save
// yields the same bytes, with no permissible-reordering caveat for the builder
// surface (deterministic saves, D-035).
func TestRoundTrip_FixtureByteIdentical(t *testing.T) {
	_, data1 := buildFixture(t)
	re, err := pptx.NewFromBytes(data1)
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	data2, err := re.WriteToBytes()
	if err != nil {
		t.Fatalf("re-save: %v", err)
	}
	if !bytes.Equal(data1, data2) {
		t.Errorf("fixture not byte-identical on reopen: first %d bytes, re-saved %d bytes", len(data1), len(data2))
	}
}

// collectKinds records the NodeKind of every node in nodes, recursing into the
// container kinds (their children render per their own policy).
func collectKinds(nodes []scene.SlideNode, set map[scene.NodeKind]bool) {
	for _, n := range nodes {
		set[n.NodeKind()] = true
		switch c := n.(type) {
		case scene.TwoColumn:
			collectKinds(c.Left, set)
			collectKinds(c.Right, set)
		case scene.Grid:
			collectKinds(c.Cells, set)
		case scene.Card:
			collectKinds(c.Body, set)
		case scene.CardSection:
			collectKinds(c.Body, set)
		case scene.Bento:
			for _, row := range c.Rows {
				for _, cell := range row.Cells {
					collectKinds([]scene.SlideNode{cell.Node}, set)
				}
			}
		case scene.Banner:
			collectKinds(c.Trailing, set)
		}
	}
}

// everyNodeScene is a scene exercising all 35 shipped scene IR node kinds (the
// scene/policy.go policyTable set): the 23 leaf kinds and the 5 container kinds.
// Asset-bearing kinds (Image, CodeBlock, Chart, Decoration-asset) resolve through
// the stub resolver.
func everyNodeScene() scene.Scene {
	rt := func(s string) scene.RichText { return scene.RichText{{Text: s}} }
	return scene.Scene{
		Meta: scene.Metadata{Title: "Every node"},
		Slides: []scene.SceneSlide{
			{
				ID: "leaves-a",
				Nodes: []scene.SlideNode{
					scene.Hero{Eyebrow: "2025", Title: "All nodes", Subtitle: "round-trip"},
					scene.Prose{Paragraphs: []scene.RichText{rt("A body paragraph.")}},
					scene.Heading{Text: rt("Heading"), Level: 2},
					scene.List{Kind: scene.ListBullet, Items: []scene.ListItem{{Text: rt("one")}, {Text: rt("two")}}},
					scene.Divider{},
					scene.Quote{Text: rt("A quotation."), Attribution: "Someone"},
					scene.Callout{Kind: scene.CalloutTip, Title: "Tip", Body: rt("Mind the gap.")},
				},
			},
			{
				ID: "leaves-b",
				Nodes: []scene.SlideNode{
					scene.Chip{Label: "new", Tone: scene.ChipSolid, Color: scene.ColorAccent},
					scene.Arrow{Direction: scene.ArrowRight, Label: "next"},
					scene.Image{AssetID: "asset://img", Alt: "shot"},
					scene.CodeBlock{AssetID: "asset://code", Language: "go", Caption: "demo.go"},
					scene.Chart{AssetID: "asset://chart", Caption: "Q1"},
					scene.Table{
						Headers: []scene.RichText{rt("A"), rt("B")},
						Rows:    [][]scene.RichText{{rt("1"), rt("2")}},
						Caption: "data",
					},
					scene.Flow{Orientation: scene.FlowHorizontal, Steps: []scene.FlowStep{
						{Label: rt("Plan")}, {Label: rt("Build")}, {Label: rt("Ship")},
					}},
					scene.Decoration{
						Kind: scene.DecorationAsset, AssetID: "asset://deco",
						Anchor: scene.AnchorCenter, Size: scene.Size{W: 914400, H: 914400},
					},
				},
			},
			{
				ID:    "section",
				Nodes: []scene.SlideNode{scene.SectionDivider{Eyebrow: "Part II", Label: "Containers"}},
			},
			{
				ID: "containers",
				Nodes: []scene.SlideNode{
					scene.TwoColumn{
						Ratio: scene.Ratio11,
						Left:  []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("left")}}},
						Right: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("right")}}},
					},
				},
			},
			{
				ID: "cards",
				Nodes: []scene.SlideNode{
					scene.CardSection{
						Header: "Section",
						Body: []scene.SlideNode{
							scene.Grid{Columns: 2, Cells: []scene.SlideNode{
								scene.Card{Header: "One", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("a")}}}},
								scene.Card{Header: "Two", Body: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{rt("b")}}}},
							}},
						},
					},
				},
			},
			{
				ID: "stats",
				Nodes: []scene.SlideNode{
					scene.Grid{Columns: 3, Cells: []scene.SlideNode{
						scene.Stat{Value: "$2,200", Label: "ARR", Delta: "+12%", DeltaTone: scene.DeltaUp},
						scene.Stat{Value: "38%", Label: "Margin", Delta: "-3%", DeltaTone: scene.DeltaDown},
						scene.Stat{Value: "4.8", Label: "NPS"},
					}},
				},
			},
			{
				ID: "bento",
				Nodes: []scene.SlideNode{
					scene.Bento{Columns: 3, Rows: []scene.BentoRow{
						{Label: "Row A", Cells: []scene.BentoCell{
							{Span: 2, Node: scene.Prose{Paragraphs: []scene.RichText{rt("wide")}}},
							{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt("narrow")}}},
						}},
						{Label: "Row B", Cells: []scene.BentoCell{
							{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt("a")}}},
							{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt("b")}}},
							{Span: 1, Node: scene.Prose{Paragraphs: []scene.RichText{rt("c")}}},
						}},
					}},
				},
			},
			{
				ID: "button",
				Nodes: []scene.SlideNode{
					scene.Button{Label: "Talk to the team", Tone: scene.ButtonPrimary, Size: scene.ButtonLG, TrailingIcon: "arrow-right"},
					scene.Checklist{Columns: 2, Items: []scene.ChecklistItem{
						{Text: rt("Understands your data"), State: scene.CheckDone},
						{Text: rt("Follows your rules"), State: scene.CheckDone},
						{Text: rt("No training on prompts"), State: scene.CheckNo},
					}},
					scene.ChipRow{Label: "COMMON BUILDS", Wrap: true, Chips: []scene.ChipSpec{
						{Label: "Finance"}, {Label: "HR"}, {Label: "Sales", Tone: scene.ChipSolid, Color: scene.ColorAccent},
					}},
					scene.Banner{
						Lead: rt("Run it internally, sell it externally"), Body: rt("the power of an agentic platform"),
						Icon: "star", Fill: scene.ColorAccent,
						Trailing: []scene.SlideNode{
							scene.Button{Label: "Start free", TrailingIcon: "arrow-right"},
							// A Lockup inside Banner.Trailing (Wave-12 checkpoint composite):
							// exercises walkIconRefs recursion through Banner.Trailing into a leaf.
							scene.Lockup{Caption: "by ACME", Icon: "star"},
						},
					},
					scene.IconRows{Rows: []scene.IconRow{
						{Icon: "star", Label: rt("Chat & Q&A"), Meta: rt("core"), Tone: scene.RowPill},
						{Icon: "check", Label: rt("Specialized agents")},
					}},
					scene.Lockup{Caption: "POWERED BY CLEAR TECH", Icon: "star", AssetSide: scene.TrailCaption},
					scene.Timeline{
						Bands: []scene.TimelineBand{{From: 0, To: 0.5, Label: "Now", Fill: scene.ColorAccent}, {From: 0.5, To: 1, Label: "Next", Fill: scene.ColorInfo}},
						Lanes: []scene.TimelineLane{
							{Label: "Platform", Milestones: []scene.Milestone{{Position: 0.1, Label: "Beta", Icon: "star"}, {Position: 0.8, Label: "GA", Detail: "Q4"}}},
							{Label: "Go-to-market", Milestones: []scene.Milestone{{Position: 0.4, Label: "Pilot", AccentIndex: 1}}},
						},
					},
					scene.DataMark{Kind: scene.DataMarkBar, Value: 0.92, Label: "92%"},
					scene.DataMark{Kind: scene.DataMarkBars, Values: []float64{0.3, 0.6, 0.9, 0.5}},
					scene.DataMark{Kind: scene.DataMarkSparkline, Values: []float64{0.2, 0.5, 0.4, 0.8, 0.6, 1.0}},
					scene.DataMark{Kind: scene.DataMarkDonut, Value: 0.92, Label: "92%"},
					scene.DataMark{Kind: scene.DataMarkGauge, Value: 0.5, Label: "50"},
					scene.Quadrant{
						AxisX: scene.QuadrantAxis{LowLabel: "Low effort", HighLabel: "High effort"},
						AxisY: scene.QuadrantAxis{LowLabel: "Low impact", HighLabel: "High impact"},
						Items: []scene.QuadrantItem{{X: 0.2, Y: 0.8, Label: "Quick win"}, {X: 0.7, Y: 0.6, Label: "Big bet", AccentIndex: 1}},
					},
					scene.LogoWall{Caption: "Trusted by", Columns: 3, Tone: scene.LogoToneMono, Logos: []scene.LogoEntry{
						{AssetID: "asset://l1", Alt: "L1"}, {AssetID: "asset://l2"}, {AssetID: "asset://l3"}, {AssetID: "asset://l4"},
					}},
					scene.Tree{Root: scene.TreeNode{Label: "CEO", Children: []scene.TreeNode{
						{Label: "Eng", Icon: "star", Children: []scene.TreeNode{{Label: "Platform"}, {Label: "Apps"}}},
						{Label: "GTM", AccentIndex: 1, Children: []scene.TreeNode{{Label: "Sales"}}},
					}}},
					scene.Funnel{Stages: []scene.FunnelStage{{Label: "Leads", Value: "10k"}, {Label: "Qualified", Value: "3k"}, {Label: "Won", Value: "400"}}},
					scene.Cycle{Stages: []scene.CycleStage{{Label: "Plan", Icon: "star"}, {Label: "Build"}, {Label: "Ship"}, {Label: "Learn"}}},
				},
			},
		},
	}
}

// TestRoundTrip_SceneNodes is PR#4 acceptance criterion 4 (scene side): a scene
// exercising every shipped IR node kind renders, and the rendered deck reopens
// into a navigable, byte-stable model. Scene-level read is out of scope (D-047),
// so the assertion is at the builder level: every slide reopens with navigable
// shapes and the deck re-saves byte-identically.
func TestRoundTrip_SceneNodes(t *testing.T) {
	sc := everyNodeScene()

	// Mechanically assert the fixture covers every shipped node kind, so adding a
	// node without extending this walk fails loudly (the kinds are contiguous,
	// KindHero..KindTimeline).
	kinds := map[scene.NodeKind]bool{}
	for _, sl := range sc.Slides {
		collectKinds(sl.Nodes, kinds)
	}
	for k := scene.KindHero; k <= scene.KindCycle; k++ {
		if !kinds[k] {
			t.Errorf("scene fixture does not exercise node kind %v", k)
		}
	}

	resolver := scene.URIAssetResolver(func(string) ([]byte, string, error) {
		return append([]byte("\x89PNG\r\n\x1a\n"), []byte("asset")...), "image/png", nil
	})

	pres := pptx.New()
	stats, err := scene.Render(pres, sc, scene.WithAssetResolver(resolver))
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if stats.Slides != 8 {
		t.Fatalf("stats.Slides = %d, want 8", stats.Slides)
	}

	data1, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	re, err := pptx.NewFromBytes(data1)
	if err != nil {
		t.Fatalf("NewFromBytes on scene deck: %v", err)
	}
	slides := re.Slides()
	if len(slides) != 8 {
		t.Fatalf("reopened slides = %d, want 8", len(slides))
	}
	for i, s := range slides {
		if len(s.Shapes()) == 0 {
			t.Errorf("reopened slide %d has no navigable shapes", i)
		}
	}

	// The scene-rendered deck re-saves byte-identically through Open (D-035).
	data2, err := re.WriteToBytes()
	if err != nil {
		t.Fatalf("re-save scene deck: %v", err)
	}
	if !bytes.Equal(data1, data2) {
		t.Errorf("scene deck not byte-identical on reopen: first %d, re-saved %d", len(data1), len(data2))
	}
}
