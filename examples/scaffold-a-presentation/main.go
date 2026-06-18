// Command scaffold-a-presentation builds a small PPTX deck from scratch with
// the pptx (Layer 1) builder, then reopens it to confirm a self-authored deck
// round-trips losslessly (G6).
//
// It exercises the core builder surface: a themed title text frame, a
// token-filled shape, a small table, an embedded PNG, speaker notes, and a
// named section. Styling flows through theme tokens (P2) — RGB / Pt are the
// escape hatch, not the default. Nothing is written to disk: the deck is
// serialized to memory with WriteToBytes and re-read with NewFromBytes.
//
// Run it:
//
//	go run ./examples/scaffold-a-presentation
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"

	"github.com/hurtener/pptx-go/pptx"
)

func main() {
	// 1. Create a 16:9 deck. With no theme it uses pptx.DefaultTheme(); every
	//    token below resolves against it at authoring time.
	pres := pptx.New(pptx.WithFormat(pptx.Slides16x9))

	// 2. First slide: a themed title + subtitle.
	title := pres.AddSlide()

	heading := title.AddTextFrame(pptx.Box{X: pptx.In(0.8), Y: pptx.In(0.7), W: pptx.In(11), H: pptx.In(1.4)})
	hp := heading.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignLeft})
	hp.AddRun("Quarterly Review", pptx.RunStyle{
		TypeRole: pptx.TypeDisplay,
		Color:    pptx.TokenTextColor(pptx.TextPrimary),
		Bold:     true,
	})
	sub := title.AddTextFrame(pptx.Box{X: pptx.In(0.8), Y: pptx.In(2.0), W: pptx.In(11), H: pptx.In(0.8)})
	sub.AddParagraph(pptx.ParagraphOpts{}).AddRun(
		"Built with the pptx-go builder",
		pptx.RunStyle{TypeRole: pptx.TypeH4, Color: pptx.TokenTextColor(pptx.TextSecondary)},
	)

	// A token-filled rounded shape as an accent band.
	title.AddShape(
		pptx.ShapeRoundRect,
		pptx.Box{X: pptx.In(0.8), Y: pptx.In(3.1), W: pptx.In(4), H: pptx.In(0.18)},
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))),
		pptx.WithRadius(pptx.RadiusFull),
		pptx.WithElevation(pptx.ElevationRaised),
	)

	// An embedded PNG generated at runtime (a valid 64x64 swatch).
	img, err := title.AddImage(
		pptx.ImageBytes(makePNG(), "image/png"),
		pptx.Box{X: pptx.In(9.6), Y: pptx.In(3.1), W: pptx.In(2), H: pptx.In(2)},
	)
	if err != nil {
		log.Fatalf("add image: %v", err)
	}
	img.SetAltText("brand swatch")

	// Speaker notes for the presenter.
	title.SetSpeakerNotes("Open with the headline number, then walk the table.")

	// 3. Second slide: a small table with a header row and banding.
	data := pres.AddSlide()
	data.AddTextFrame(pptx.Box{X: pptx.In(0.8), Y: pptx.In(0.6), W: pptx.In(11), H: pptx.In(0.9)}).
		AddParagraph(pptx.ParagraphOpts{}).
		AddRun("Results by Region", pptx.RunStyle{TypeRole: pptx.TypeH1, Color: pptx.TokenTextColor(pptx.TextPrimary)})

	tbl := data.AddTable(pptx.Box{X: pptx.In(0.8), Y: pptx.In(1.8), W: pptx.In(7), H: pptx.In(2.5)}, 3, 2)
	tbl.SetHeaderRow(true)
	tbl.SetBanding(true, false)
	tbl.Cell(0, 0).SetText("Region")
	tbl.Cell(0, 1).SetText("Revenue")
	tbl.Cell(1, 0).SetText("North")
	tbl.Cell(1, 1).SetText("$1.2M")
	tbl.Cell(2, 0).SetText("South")
	tbl.Cell(2, 1).SetText("$0.9M")

	// 4. Group both slides under a named section.
	sec := pres.AddSection("Overview")
	for _, s := range pres.Slides() {
		sec.Include(s)
	}

	// 5. Serialize to memory (no junk files on disk).
	data2, err := pres.WriteToBytes()
	if err != nil {
		log.Fatalf("write deck: %v", err)
	}

	// 6. Reopen and inspect: a self-authored deck round-trips with no warnings.
	reopened, err := pptx.NewFromBytes(data2)
	if err != nil {
		log.Fatalf("reopen deck: %v", err)
	}
	slides := reopened.Slides()
	warnings := reopened.ReadWarnings()

	fmt.Printf("OK: authored %d bytes, reopened %d slides, %d read warnings\n",
		len(data2), len(slides), len(warnings))
}

// makePNG returns a small, valid PNG (a 64x64 solid swatch) so the example is
// self-contained and never reads a fixture file.
func makePNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	swatch := color.RGBA{R: 0x25, G: 0x63, B: 0xEB, A: 0xFF} // matches ColorAccent
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, swatch)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}
