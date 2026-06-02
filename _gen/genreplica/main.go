// Command genreplica reproduces the "Project Management Top Concerns" pitch
// deck (a real, hand-built Clear Tech deck) using only the pptx builder — a
// fidelity exercise for the engine. The original is a 20×11.25in (16:9) canvas;
// the builder's public API offers the standard 16:9 (13.333×7.5in), so every
// coordinate and the type scale are authored at 2/3 scale, preserving the
// proportions exactly.
//
// Run: go run ./_gen/genreplica test-output/replica.pptx
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hurtener/pptx-go/pptx"
)

// scale maps the original 20in canvas onto the builder's 13.333in 16:9 canvas.
const scale = 2.0 / 3.0

// sx scales an "original inches" measurement onto the target canvas.
func sx(v float64) pptx.EMU { return pptx.In(v * scale) }

// box builds a scaled Box from original-deck inch coordinates.
func box(x, y, w, h float64) pptx.Box {
	return pptx.Box{X: sx(x), Y: sx(y), W: sx(w), H: sx(h)}
}

// Palette — lifted verbatim from the source deck.
var (
	navy     = pptx.RGB("0A1A3A") // page background
	white    = pptx.RGB("FFFFFF")
	teal     = pptx.RGB("82D9BD") // accent highlight word / hero numbers
	peri     = pptx.RGB("7887EC") // section eyebrow / accent bar
	lime     = pptx.RGB("B5F431") // "critical risk" category accent
	muted    = pptx.RGB("9AA3BD") // small labels, page numbers
	bodyGrey = pptx.RGB("C5CBDC") // body paragraphs
	darkBlue = pptx.RGB("092D68") // text on the periwinkle bar
	hairline = pptx.RGB("24304F") // faint dividers
)

// assetDir holds the three images extracted from the source deck.
const assetDir = "_gen/genreplica/assets"

func main() {
	out := "test-output/replica.pptx"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	theme := pptx.NewTheme(pptx.WithName("Clear Tech"), pptx.WithFonts("DIN", "Roboto"))
	// A type scale (scaled ~2/3) covering the deck's hierarchy. Heading roles are
	// DIN; body roles are Roboto. Per-run Bold/Color refine each use.
	theme.Typography[pptx.TypeDisplay] = pptx.FontSpec{Family: "DIN", Size: 56, Weight: 700} // heroes
	theme.Typography[pptx.TypeH1] = pptx.FontSpec{Family: "DIN", Size: 35, Weight: 700}      // slide titles
	theme.Typography[pptx.TypeH2] = pptx.FontSpec{Family: "DIN", Size: 23, Weight: 700}      // lead / name
	theme.Typography[pptx.TypeH3] = pptx.FontSpec{Family: "DIN", Size: 14, Weight: 700}      // card titles
	theme.Typography[pptx.TypeH4] = pptx.FontSpec{Family: "DIN", Size: 11, Weight: 700}      // sub-headings
	theme.Typography[pptx.TypeH5] = pptx.FontSpec{Family: "DIN", Size: 8, Weight: 700}       // eyebrow labels
	theme.Typography[pptx.TypeBody] = pptx.FontSpec{Family: "Roboto", Size: 13, Weight: 400} // paragraphs
	theme.Typography[pptx.TypeBodySmall] = pptx.FontSpec{Family: "Roboto", Size: 11, Weight: 400}
	theme.Typography[pptx.TypeCaption] = pptx.FontSpec{Family: "Roboto", Size: 7, Weight: 400}

	p := pptx.New(pptx.WithTheme(theme))

	slideCover(p)
	slideLandscape(p)
	slideTakeaway(p)
	slideContact(p)

	if err := p.Save(out); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Printf("wrote %s (%d slides)", out, len(p.Slides()))
}

// --- shared chrome ------------------------------------------------------------

// background paints the full-bleed navy page.
func background(s *pptx.Slide) {
	s.AddShape(pptx.ShapeRect, box(0, 0, 20, 11.25), pptx.WithFill(pptx.SolidFill(navy)))
}

// logo places the Clear Tech wordmark top-left (original image1.png).
func logo(s *pptx.Slide, w, h float64) {
	if _, err := s.AddImage(pptx.ImageFile(filepath.Join(assetDir, "image1.png")), box(0.83, 0.5, w, h)); err != nil {
		log.Fatalf("logo: %v", err)
	}
}

// label adds a single-run uppercase eyebrow/label frame. AutoFitNormal lets
// PowerPoint shrink text that would otherwise overflow its (scaled) box.
func label(s *pptx.Slide, b pptx.Box, text string, c pptx.Color) {
	tf := s.AddTextFrame(b)
	tf.AutoFit(pptx.AutoFitNormal)
	tf.AddParagraph(pptx.ParagraphOpts{}).
		AddRun(text, pptx.RunStyle{TypeRole: pptx.TypeH5, Color: c, Bold: true})
}

// para adds a single-run paragraph at an arbitrary role.
func para(s *pptx.Slide, b pptx.Box, text string, role pptx.TypeRole, c pptx.Color, bold bool) {
	tf := s.AddTextFrame(b)
	tf.AutoFit(pptx.AutoFitNormal)
	tf.AddParagraph(pptx.ParagraphOpts{}).
		AddRun(text, pptx.RunStyle{TypeRole: role, Color: c, Bold: bold})
}

// rule draws a thin horizontal divider.
func rule(s *pptx.Slide, x, y, w float64, c pptx.Color) {
	s.AddShape(pptx.ShapeRect, box(x, y, w, 0.02), pptx.WithFill(pptx.SolidFill(c)))
}

// footer is the shared bottom strip: brand caption, hairline, page number.
func footer(s *pptx.Slide, page string) {
	label(s, box(0.83, 10.55, 6, 0.24), "CLEAR TECH · YOUR STRATEGIC ALLY", muted)
	rule(s, 6.6, 10.66, 11.7, hairline)
	para(s, box(18.4, 10.55, 1.0, 0.24), page, pptx.TypeH5, muted, true)
}

// --- slide 1: cover -----------------------------------------------------------

func slideCover(p *pptx.Presentation) {
	s := p.AddSlide()
	background(s)
	// Decorative portrait graphic bleeding off the right edge (image2.png).
	if _, err := s.AddImage(pptx.ImageFile(filepath.Join(assetDir, "image2.png")), box(12.71, -0.47, 8.12, 12.19)); err != nil {
		log.Fatalf("cover art: %v", err)
	}
	logo(s, 1.35, 0.42)
	label(s, box(17.03, 0.5, 2.4, 0.24), "01 / 04 · APRIL 2026", muted)

	// Periwinkle eyebrow pill + badge text. WithRadius(RadiusFull) makes a true
	// capsule (the corner radius is a theme token).
	s.AddShape(pptx.ShapeRoundRect, box(1.04, 1.46, 4.0, 0.44), pptx.WithFill(pptx.SolidFill(peri)), pptx.WithRadius(pptx.RadiusFull))
	para(s, box(1.27, 1.56, 8.63, 0.3), "YOUR STRATEGIC ALLY", pptx.TypeH3, darkBlue, true)

	// Title, with "management" highlighted in teal.
	title := s.AddTextFrame(box(1.04, 2.31, 9.5, 3.36))
	title.AutoFit(pptx.AutoFitNormal)
	tp := title.AddParagraph(pptx.ParagraphOpts{})
	tp.AddRun("Project ", pptx.RunStyle{TypeRole: pptx.TypeDisplay, Color: white, Bold: true})
	tp.AddRun("management ", pptx.RunStyle{TypeRole: pptx.TypeDisplay, Color: teal, Bold: true})
	tp.AddRun("top concerns", pptx.RunStyle{TypeRole: pptx.TypeDisplay, Color: white, Bold: true})

	para(s, box(1.04, 5.92, 8.0, 1.27),
		"Nine recurring risks that decide whether data & digital transformation programs succeed — and how to get ahead of them.",
		pptx.TypeBody, bodyGrey, false)

	rule(s, 1.04, 8.96, 8.83, white)

	// "PREPARED BY" column.
	label(s, box(1.04, 9.3, 3.05, 0.22), "PREPARED BY", muted)
	para(s, box(1.04, 9.58, 3.05, 0.39), "Santiago Scarafia", pptx.TypeH2, white, true)
	para(s, box(1.04, 9.93, 3.05, 0.32), "Account Manager, Clear Tech", pptx.TypeBodySmall, bodyGrey, false)
	// "DATE" column.
	label(s, box(4.59, 9.3, 1.39, 0.22), "DATE", muted)
	para(s, box(4.59, 9.58, 2.0, 0.39), "April 2026", pptx.TypeH2, white, true)
}

// --- slide 2: the landscape (header + rasterized 3×3 grid + footer) -----------

func slideLandscape(p *pptx.Presentation) {
	s := p.AddSlide()
	background(s)
	logo(s, 1.22, 0.38)
	label(s, box(13.52, 0.57, 5.82, 0.27), "KEY CONCERNS · DATA & DIGITAL TRANSFORMATION", muted)

	label(s, box(0.83, 1.42, 18.88, 0.24), "THE LANDSCAPE", peri)
	para(s, box(0.83, 1.81, 16.09, 0.9), "Nine concerns across three risk categories",
		pptx.TypeH1, white, true)

	// The nine-card matrix was rendered to a single raster in the source deck
	// (image3.png); reproduce it as the engine embeds it.
	if _, err := s.AddImage(pptx.ImageFile(filepath.Join(assetDir, "image3.png")), box(0.83, 2.95, 14.02, 7.54)); err != nil {
		log.Fatalf("grid: %v", err)
	}
	footer(s, "02 / 04")
}

// --- slide 3: the takeaway ----------------------------------------------------

func slideTakeaway(p *pptx.Presentation) {
	s := p.AddSlide()
	background(s)
	logo(s, 1.22, 0.38)
	label(s, box(15.09, 0.57, 4.2, 0.27), "OBSERVATION · THE PATTERN WE SEE", muted)

	label(s, box(0.83, 1.42, 18.88, 0.24), "THE TAKEAWAY", peri)
	para(s, box(0.83, 1.81, 16.09, 0.9), "Where most projects derail", pptx.TypeH1, white, true)
	para(s, box(0.83, 3.51, 10.35, 0.7), "The technology rarely fails. Adoption does.",
		pptx.TypeH2, white, true)
	para(s, box(0.83, 4.5, 8.37, 1.11),
		"Data quality and integration complexity are the most common root causes of project failure.",
		pptx.TypeBody, bodyGrey, false)
	para(s, box(0.83, 5.8, 8.37, 0.9),
		"Change management is consistently underestimated. In the AI era, governance and compliance now sit at board level.",
		pptx.TypeBody, bodyGrey, false)

	// Three stat cards down the right.
	cards := []struct {
		eyebrow, value, desc string
	}{
		{"ROOT CAUSE", "#1 & #2", "Data quality and integration complexity lead the failure modes."},
		{"MOST UNDERESTIMATED", "Adoption", "Change management is where momentum dies, not the tech stack."},
		{"NEW BOARD-LEVEL TOPIC", "Governance", "Compliance & AI governance now sit alongside financial risk."},
	}
	y := 3.05
	for _, c := range cards {
		s.AddShape(pptx.ShapeRoundRect, box(11.72, y, 7.45, 2.04),
			pptx.WithFill(pptx.SolidFill(pptx.RGB("0F2147"))),
			pptx.WithLine(pptx.Line{Color: hairline, Width: pptx.Pt(1)}),
			pptx.WithRadius(pptx.RadiusLG))
		label(s, box(12.06, y+0.3, 6.97, 0.23), c.eyebrow, muted)
		para(s, box(12.06, y+0.63, 6.97, 0.79), c.value, pptx.TypeH2, teal, true)
		para(s, box(12.06, y+1.48, 6.97, 0.4), c.desc, pptx.TypeBodySmall, bodyGrey, false)
		y += 2.25
	}
	footer(s, "03 / 04")
}

// --- slide 4: contact ---------------------------------------------------------

func slideContact(p *pptx.Presentation) {
	s := p.AddSlide()
	background(s)
	logo(s, 1.22, 0.38)
	label(s, box(16.73, 0.57, 2.52, 0.27), "CONTACT · LET'S TALK", muted)

	s.AddShape(pptx.ShapeRoundRect, box(0.83, 2.9, 3.19, 0.38), pptx.WithFill(pptx.SolidFill(peri)), pptx.WithRadius(pptx.RadiusFull))
	para(s, box(1.1, 3.0, 9.22, 0.3), "YOUR STRATEGIC ALLY", pptx.TypeH3, darkBlue, true)
	para(s, box(0.83, 3.52, 9.22, 2.0), "Let's turn concerns into a plan.",
		pptx.TypeDisplay, white, true)
	para(s, box(0.83, 7.57, 6.65, 0.83),
		"We help enterprise teams design, build, and adopt data & AI programs that survive contact with reality.",
		pptx.TypeBody, bodyGrey, false)

	// Contact card.
	s.AddShape(pptx.ShapeRoundRect, box(11.03, 3.74, 8.13, 4.16),
		pptx.WithFill(pptx.SolidFill(pptx.RGB("0F2147"))),
		pptx.WithLine(pptx.Line{Color: hairline, Width: pptx.Pt(1)}),
		pptx.WithRadius(pptx.RadiusLG))
	para(s, box(11.62, 4.33, 7.16, 0.57), "Santiago Scarafia", pptx.TypeH2, white, true)
	para(s, box(11.62, 4.92, 7.16, 0.32), "Account Manager · Clear Tech", pptx.TypeBodySmall, bodyGrey, false)

	contacts := []struct{ k, v string }{
		{"EMAIL", "santiago.scarafia@clear-tech.com"},
		{"AR", "+54 11 6285 0925"},
		{"US", "+1 888 253 2720"},
	}
	y := 5.57
	for _, c := range contacts {
		label(s, box(11.62, y+0.1, 1.33, 0.2), c.k, muted)
		para(s, box(13.12, y, 5.62, 0.35), c.v, pptx.TypeH3, white, false)
		rule(s, 11.62, y+0.5, 6.95, hairline)
		y += 0.74
	}
	footer(s, "04 / 04")
}
