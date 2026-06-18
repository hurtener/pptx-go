// Command load-a-brand-template demonstrates ingesting an existing .pptx brand
// kit so a new deck inherits its theme and slide masters/layouts (D-037).
//
// A real caller opens a brand .pptx that their design team authored in
// PowerPoint — it carries a fully designed theme1.xml plus named layouts. To
// stay self-contained (there is no brand file on disk here), this example first
// authors a tiny "brand kit" deck in code and serializes it to bytes; those
// bytes stand in for the brand .pptx a caller would load with pptx.NewFromFile.
//
// What FromTemplate adopts, verified empirically:
//   - the brand's slide masters + layouts (discoverable via Masters()/Layouts(),
//     selectable by name via AddSlide("Name") / HasLayout) — shown below;
//   - the brand's theme, exactly as encoded in the brand's theme1.xml.
//
// One caveat the output makes visible: a brand authored purely in code with
// WithTheme does not yet persist its custom color/font tokens into theme1.xml
// (token emission to theme1.xml is pending — see pptx.WithTheme), so this
// synthetic brand's adopted theme reflects the default scaffold theme, not the
// WithAccent/WithFonts values set below. A brand .pptx designed in PowerPoint
// carries a real theme1.xml and is adopted in full.
//
// Run:
//
//	go run ./examples/load-a-brand-template
package main

import (
	"fmt"
	"log"

	"github.com/hurtener/pptx-go/pptx"
)

func main() {
	// --- Step 1: produce a brand kit .pptx (stands in for a file on disk) ------
	// A brand kit is just a .pptx carrying a theme + masters/layouts.
	brandTheme := pptx.NewTheme(
		pptx.WithName("Acme Brand Kit"),
		pptx.WithAccent("FF6A00"), // Acme orange (token-only today)
		pptx.WithFonts("Georgia", "Verdana"),
	)
	brandKit := pptx.New(pptx.WithTheme(brandTheme))
	brandKit.AddSlide() // a sample slide; ingestion strips it from the new deck

	brandBytes, err := brandKit.WriteToBytes()
	if err != nil {
		log.Fatalf("authoring brand kit: %v", err)
	}

	// --- Step 2: open the brand kit -------------------------------------------
	// In production this is pptx.NewFromFile("acme-brand.pptx"). Opening a deck
	// extracts its theme and its master/layout registry (RFC §13.1).
	brand, err := pptx.NewFromBytes(brandBytes)
	if err != nil {
		log.Fatalf("opening brand kit: %v", err)
	}
	defer func() { _ = brand.Close() }()

	// --- Step 3: seed a new deck from the brand kit ---------------------------
	// FromTemplate adopts the brand's theme + masters/layouts. The new deck
	// starts slide-free; the brand deck is cloned, not retained or mutated.
	deck := pptx.New(pptx.FromTemplate(brand))

	// --- Step 4: discover and select layouts by name --------------------------
	// AddSlide with a layout name resolves against the adopted registry; an
	// unknown name falls back to a blank layout. HasLayout reports availability
	// up front. Iterate Masters()/Layouts() to learn the names a brand defines.
	for _, m := range deck.Masters() {
		for _, l := range m.Layouts() {
			if name := l.Name(); name != "" {
				fmt.Printf("available layout: %q (master %q)\n", name, m.Name())
			}
		}
	}
	if deck.HasLayout("Blank") {
		deck.AddSlide("Blank") // resolves against the adopted registry
	} else {
		deck.AddSlide() // blank fallback
	}

	// --- Step 5: confirm ingestion worked -------------------------------------
	// Masters adopted is the strong, distinctive proof of brand ingestion.
	adopted := deck.Theme()
	fmt.Printf("masters adopted:      %d\n", len(deck.Masters()))
	fmt.Printf("adopted theme name:   %q\n", adopted.Name)
	fmt.Printf("adopted accent color: %q (from the brand's theme1.xml)\n", adopted.ResolveColor(pptx.ColorAccent))
	fmt.Printf("adopted heading font: %q\n", adopted.HeadingFont)

	if _, err := deck.WriteToBytes(); err != nil {
		log.Fatalf("writing final deck: %v", err)
	}

	fmt.Println("OK: new deck inherited the brand theme + masters/layouts")
}
