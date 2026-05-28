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
	s.AddRectangle(914400, 914400, 2743200, 1371600)
	s.AddTextBox(914400, 2743200, 6858000, 914400, "pptx-go reference deck")
	pres.AddSlide()

	if err := pres.Save(out); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", out)
}
