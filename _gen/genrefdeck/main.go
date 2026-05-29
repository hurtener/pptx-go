// Command genrefdeck emits a reference .pptx used by the validity layers
// (LibreOffice headless open-proxy in CI, and the manual per-wave PowerPoint
// check). It writes to the path given as the first argument, default
// test-output/reference.pptx. Run via `go run ./_gen/genrefdeck [path]`.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hurtener/pptx-go/pptx"
)

func main() {
	out := "test-output/reference.pptx"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		log.Fatal(err)
	}

	pres := pptx.New()
	s := pres.AddSlide()
	// A themed, filled & outlined shape (token fill resolves against the theme)
	// plus a translucent accent — exercises the Color/Fill/Line surface.
	s.AddShape(
		pptx.ShapeRoundRect,
		pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(3), H: pptx.In(1.5)},
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))),
		pptx.WithLine(pptx.Line{Width: pptx.Pt(1.5), Color: pptx.RGB("111827")}),
	)
	s.AddShape(
		pptx.ShapeEllipse,
		pptx.Box{X: pptx.In(4.5), Y: pptx.In(1), W: pptx.In(2), H: pptx.In(2)},
		pptx.WithFill(pptx.SolidFill(pptx.RGBA("2563EB", 40000))),
	)
	s.AddTextBox(int(pptx.In(1)), int(pptx.In(3)), int(pptx.In(7.5)), int(pptx.In(1)), "pptx-go reference deck")
	pres.AddSlide()

	if err := pres.Save(out); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", out)
}
