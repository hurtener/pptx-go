// Command genshowcase emits a single .pptx exercising the full pptx-go surface
// available today (builder: shapes/fills/lines, rich text, images, sections,
// notes, tables; scene: every rendered leaf + containers + table). It is a
// quality/eyeball artifact, not a test. Run: go run ./_gen/genshowcase [path]
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/pptx"
	"github.com/hurtener/pptx-go/scene"
)

func main() {
	out := "test-output/showcase.pptx"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		log.Fatal(err)
	}

	// A lightly branded theme (proves theming flows through both layers).
	th := pptx.DefaultTheme()
	th.Colors.Surfaces[pptx.ColorAccent] = "5B21B6"    // violet
	th.Colors.Surfaces[pptx.ColorAccentAlt] = "0891B2" // teal

	pres := pptx.New(pptx.WithTheme(th), pptx.WithFormat(pptx.Slides16x9))

	// ---- Builder-authored slides ------------------------------------------
	intro := pres.AddSection("Builder")
	intro.Include(coverSlide(pres))
	intro.Include(shapesSlide(pres))
	intro.Include(richTextSlide(pres))
	intro.Include(imageSlide(pres))
	intro.Include(tableSlide(pres))

	// ---- Scene-rendered slides --------------------------------------------
	logo := codeShot()
	resolver := scene.URIAssetResolver(func(uuid string) ([]byte, string, error) {
		return logo, "image/png", nil
	})
	stats, err := scene.Render(pres, showcaseScene(th), scene.WithAssetResolver(resolver))
	if err != nil {
		log.Fatalf("scene.Render: %v", err)
	}

	// Group the scene slides into a section too.
	sceneSec := pres.AddSection("Scene")
	all := pres.Slides()
	for _, s := range all[5:] { // the scene slides follow the 5 builder slides
		sceneSec.Include(s)
	}

	data, err := pres.WriteToBytes()
	if err != nil {
		log.Fatalf("WriteToBytes: %v", err)
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		log.Fatal(err)
	}

	rep, _ := conformance.ValidateBytes(data, conformance.Options{
		RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"},
	})
	abs, _ := filepath.Abs(out)
	log.Printf("wrote %s (%d slides, %d KB)", abs, pres.SlideCount(), len(data)/1024)
	log.Printf("scene stats: %d slides, %d shapes, %d assets, %d warnings", stats.Slides, stats.Shapes, stats.Assets, len(stats.Warnings))
	log.Printf("conformance OK: %v", rep.OK())
	if !rep.OK() {
		log.Printf("%s", rep)
	}
}

// ---- builder slides -------------------------------------------------------

func coverSlide(p *pptx.Presentation) *pptx.Slide {
	s := p.AddSlide()
	// Full-bleed accent panel on the left third.
	s.AddShape(pptx.ShapeRect,
		pptx.Box{X: 0, Y: 0, W: pptx.In(4.4), H: pptx.In(7.5)},
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))))
	// Translucent accent circle (alpha).
	s.AddShape(pptx.ShapeEllipse,
		pptx.Box{X: pptx.In(3.0), Y: pptx.In(4.5), W: pptx.In(3), H: pptx.In(3)},
		pptx.WithFill(pptx.SolidFill(pptx.RGBA("0891B2", 35000))))

	tf := s.AddTextFrame(pptx.Box{X: pptx.In(4.9), Y: pptx.In(2.2), W: pptx.In(8), H: pptx.In(3)})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("pptx-go", pptx.RunStyle{TypeRole: pptx.TypeCaption, Color: pptx.TokenTextColor(pptx.TextAccent)})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Showcase deck", pptx.RunStyle{TypeRole: pptx.TypeDisplay, Bold: true})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun("Everything the builder and the scene renderer can do today.", pptx.RunStyle{TypeRole: pptx.TypeBody, Color: pptx.TokenTextColor(pptx.TextSecondary)})

	s.SetSpeakerNotes("Open on the headline: this deck is generated entirely by pptx-go, no PowerPoint.")
	return s
}

func shapesSlide(p *pptx.Presentation) *pptx.Slide {
	s := p.AddSlide()
	title(s, "Shapes · fills · lines")
	geoms := []pptx.ShapeGeometry{pptx.ShapeRoundRect, pptx.ShapeEllipse, pptx.ShapeTriangle, pptx.ShapeDiamond, pptx.ShapeHexagon, pptx.ShapeChevron}
	x := pptx.In(0.7)
	for i, g := range geoms {
		col := []pptx.Color{pptx.TokenColor(pptx.ColorAccent), pptx.TokenColor(pptx.ColorAccentAlt), pptx.TokenColor(pptx.ColorSuccess), pptx.TokenColor(pptx.ColorWarning), pptx.TokenColor(pptx.ColorError), pptx.TokenColor(pptx.ColorInfo)}[i]
		s.AddShape(g, pptx.Box{X: x, Y: pptx.In(2.0), W: pptx.In(1.7), H: pptx.In(1.7)},
			pptx.WithFill(pptx.SolidFill(col)),
			pptx.WithLine(pptx.Line{Width: pptx.Pt(1.25), Color: pptx.RGB("111827")}))
		x += pptx.In(2.0)
	}
	// An outlined + dashed shape row.
	s.AddShape(pptx.ShapeRoundRect, pptx.Box{X: pptx.In(0.7), Y: pptx.In(4.2), W: pptx.In(11.3), H: pptx.In(1.4)},
		pptx.WithFill(pptx.NoFill()),
		pptx.WithLine(pptx.Line{Width: pptx.Pt(2), Color: pptx.TokenColor(pptx.ColorAccent), Dash: "dash"}))
	return s
}

func richTextSlide(p *pptx.Presentation) *pptx.Slide {
	s := p.AddSlide()
	title(s, "Rich text · runs · bullets · links")
	tf := s.AddTextFrame(pptx.Box{X: pptx.In(0.7), Y: pptx.In(1.7), W: pptx.In(11.6), H: pptx.In(5.2)})

	p1 := tf.AddParagraph(pptx.ParagraphOpts{})
	p1.AddRun("Inline styles: ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	p1.AddRun("bold", pptx.RunStyle{TypeRole: pptx.TypeBody, Bold: true})
	p1.AddRun(", ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	p1.AddRun("italic", pptx.RunStyle{TypeRole: pptx.TypeBody, Italic: true})
	p1.AddRun(", ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	p1.AddRun("underline", pptx.RunStyle{TypeRole: pptx.TypeBody, Underline: pptx.UnderlineSingle})
	p1.AddRun(", ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	p1.AddRun("strike", pptx.RunStyle{TypeRole: pptx.TypeBody, Strike: pptx.StrikeSingle})
	p1.AddRun(", ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	p1.AddRun("colored", pptx.RunStyle{TypeRole: pptx.TypeBody, Color: pptx.TokenTextColor(pptx.TextAccent)})
	p1.AddRun(", and ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	p1.AddRun("go test ./...", pptx.RunStyle{TypeRole: pptx.TypeBody, Code: true})
	p1.AddRun(".", pptx.RunStyle{TypeRole: pptx.TypeBody})

	p2 := tf.AddParagraph(pptx.ParagraphOpts{})
	p2.AddRun("A link: ", pptx.RunStyle{TypeRole: pptx.TypeBody})
	p2.AddHyperlink("github.com/hurtener/pptx-go", "https://github.com/hurtener/pptx-go", pptx.RunStyle{TypeRole: pptx.TypeBody, Color: pptx.TokenTextColor(pptx.TextAccent), Underline: pptx.UnderlineSingle})

	tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletDisc}).AddRun("Disc bullet item", pptx.RunStyle{TypeRole: pptx.TypeBody})
	tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletDisc, Level: 1}).AddRun("Nested bullet", pptx.RunStyle{TypeRole: pptx.TypeBody})
	tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletNumber}).AddRun("Numbered item", pptx.RunStyle{TypeRole: pptx.TypeBody})
	tf.AddParagraph(pptx.ParagraphOpts{Bullet: pptx.BulletCheckbox}).AddRun("Checklist item", pptx.RunStyle{TypeRole: pptx.TypeBody})
	return s
}

func imageSlide(p *pptx.Presentation) *pptx.Slide {
	s := p.AddSlide()
	title(s, "Images · dedup · alt text")
	logo := gradientPNG(640, 360)
	box := pptx.Box{X: pptx.In(0.7), Y: pptx.In(1.8), W: pptx.In(5.6), H: pptx.In(3.15)}
	if img, err := s.AddImage(pptx.ImageBytes(logo, "image/png"), box); err == nil {
		img.SetAltText("A generated gradient")
	}
	// Same bytes again → deduped to one media part, second placement.
	box2 := pptx.Box{X: pptx.In(6.7), Y: pptx.In(1.8), W: pptx.In(5.6), H: pptx.In(3.15)}
	if _, err := s.AddImage(pptx.ImageBytes(logo, "image/png"), box2); err != nil {
		log.Printf("AddImage 2: %v", err)
	}
	return s
}

func tableSlide(p *pptx.Presentation) *pptx.Slide {
	s := p.AddSlide()
	title(s, "Tables · header · banding · merge")
	t := s.AddTable(pptx.Box{X: pptx.In(0.7), Y: pptx.In(1.8), W: pptx.In(11.6), H: pptx.In(4.5)}, 5, 4)
	headers := []string{"Region", "Q1", "Q2", "Q3"}
	for c, h := range headers {
		t.Cell(0, c).SetText(h)
	}
	rows := [][]string{
		{"North", "120", "138", "151"},
		{"South", "98", "104", "119"},
		{"EMEA", "210", "225", "240"},
		{"APAC", "", "", ""},
	}
	for r, row := range rows {
		for c, v := range row {
			t.Cell(r+1, c).SetText(v)
		}
	}
	t.Cell(4, 0).SetText("APAC — launching Q4").MergeRight(4)
	t.SetHeaderRow(true).SetBanding(true, false)
	return s
}

func title(s *pptx.Slide, text string) {
	tf := s.AddTextFrame(pptx.Box{X: pptx.In(0.7), Y: pptx.In(0.5), W: pptx.In(11.6), H: pptx.In(0.9)})
	tf.AddParagraph(pptx.ParagraphOpts{}).AddRun(text, pptx.RunStyle{TypeRole: pptx.TypeH1, Bold: true, Color: pptx.TokenTextColor(pptx.TextPrimary)})
}

// ---- scene deck -----------------------------------------------------------

func showcaseScene(th *pptx.Theme) scene.Scene {
	body := func(s string) scene.RichText {
		return scene.RichText{{Text: s, Style: scene.RunStyle{TypeRole: scene.TypeBody}}}
	}
	return scene.Scene{
		Theme: th,
		Meta:  scene.Metadata{Title: "pptx-go showcase", Author: "pptx-go"},
		Slides: []scene.SceneSlide{
			{ID: "scene-divider", Nodes: []scene.SlideNode{scene.SectionDivider{Eyebrow: "Layer 2", Label: "Scene renderer"}}},
			{ID: "hero", Layout: scene.LayoutCover, Nodes: []scene.SlideNode{
				scene.Hero{Eyebrow: "IR-driven", Title: "Typed scenes → PPTX", Subtitle: "Author a Scene; the renderer composes the builder."},
			}, Notes: body("The scene layer maps a typed IR onto the builder.")},
			{ID: "leaves", Nodes: []scene.SlideNode{
				scene.Heading{Text: scene.RichText{{Text: "Leaf nodes", Style: scene.RunStyle{TypeRole: scene.TypeH2}}}, Level: 2},
				scene.Prose{Paragraphs: []scene.RichText{body("Prose, headings, lists, quotes, callouts, chips and arrows all render natively.")}},
				scene.List{Kind: scene.ListChecklist, Items: []scene.ListItem{
					{Text: body("hero / prose / heading"), Checked: true},
					{Text: body("list / divider / quote"), Checked: true},
					{Text: body("callout / chip / arrow"), Checked: true},
				}},
				scene.Divider{},
				scene.Quote{Text: body("Engine, not product — the library encodes the IR; the caller decides the content."), Attribution: "RFC §D-026"},
			}},
			{ID: "callouts", Nodes: []scene.SlideNode{
				scene.Callout{Kind: scene.CalloutTip, Title: "Tip", Body: body("Tokens resolve against the active theme at render time.")},
				scene.Callout{Kind: scene.CalloutWarning, Title: "Heads up", Body: body("Overflow is a warning, never a silent crop.")},
			}},
			{ID: "twocol", Nodes: []scene.SlideNode{
				scene.TwoColumn{Ratio: scene.Ratio12,
					Left:  []scene.SlideNode{scene.Heading{Text: body("1:2 split"), Level: 3}, scene.Prose{Paragraphs: []scene.RichText{body("The left column.")}}},
					Right: []scene.SlideNode{scene.Prose{Paragraphs: []scene.RichText{body("A wider right column with more room for body content and a code sample below.")}}, scene.CodeBlock{AssetID: "asset://code", Language: "go", Caption: "render.go"}},
				},
			}},
			{ID: "grid", Nodes: []scene.SlideNode{
				scene.Heading{Text: body("Grid of cards"), Level: 2},
				scene.Grid{Columns: 3, Cells: []scene.SlideNode{
					scene.Callout{Kind: scene.CalloutNote, Title: "One", Body: body("First")},
					scene.Callout{Kind: scene.CalloutTip, Title: "Two", Body: body("Second")},
					scene.Callout{Kind: scene.CalloutImportant, Title: "Three", Body: body("Third")},
					scene.Chip{Label: "alpha", Tone: scene.ChipSolid, Color: scene.ColorAccent},
					scene.Chip{Label: "beta", Tone: scene.ChipTint, Color: scene.ColorAccentAlt},
					scene.Chip{Label: "ga", Tone: scene.ChipOutline, Color: scene.ColorSuccess},
				}},
			}},
			{ID: "table", Nodes: []scene.SlideNode{
				scene.Table{
					Caption: "Scene-composed table with a caption",
					Headers: []scene.RichText{body("Feature"), body("Layer"), body("Status")},
					Rows: [][]scene.RichText{
						{body("Shapes"), body("builder"), body("done")},
						{body("Rich text"), body("builder"), body("done")},
						{body("Containers"), body("scene"), body("done")},
						{body("Table"), body("both"), body("done")},
					},
				},
			}},
		},
	}
}

// ---- generated images -----------------------------------------------------

func gradientPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(91 + 120*x/w),
				G: uint8(33 + 80*y/h),
				B: uint8(182 - 60*x/w),
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func codeShot() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 720, 360))
	bg := color.RGBA{R: 17, G: 24, B: 39, A: 255}
	for y := 0; y < 360; y++ {
		for x := 0; x < 720; x++ {
			img.Set(x, y, bg)
		}
	}
	// A few "code line" bars.
	bar := color.RGBA{R: 0x34, G: 0xd3, B: 0x99, A: 255}
	for i := 0; i < 6; i++ {
		yy := 40 + i*40
		for x := 40; x < 40+(120+i*70); x++ {
			for y := yy; y < yy+12; y++ {
				img.Set(x, y, bar)
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
