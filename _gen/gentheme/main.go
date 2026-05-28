// Command gentheme emits templates/_default-theme.pptx: a minimal package
// carrying the default theme's theme1.xml, consumed by the scaffold (RFC §7.5).
// Run via `go run ./_gen/gentheme`.
package main

import (
	"log"
	"os"

	"github.com/hurtener/pptx-go/internal/opc"
	"github.com/hurtener/pptx-go/pptx"
)

func main() {
	xml, err := pptx.DefaultTheme().ThemeXML()
	if err != nil {
		log.Fatal(err)
	}
	pkg := opc.NewPackage()
	if _, err := pkg.CreatePart(opc.NewPackURI("/ppt/theme/theme1.xml"), opc.ContentTypeTheme, xml); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("templates/_default-theme.pptx")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := pkg.Save(f); err != nil {
		log.Fatal(err)
	}
}
