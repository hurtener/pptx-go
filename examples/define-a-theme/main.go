// Command define-a-theme shows the pptx-go token system (P2) end to end:
// how to construct a Theme two ways, how to fill shapes and type text runs
// with semantic tokens, and how the SAME builder input re-skins when it is
// rendered under a different theme.
//
// Run it with:
//
//	go run ./examples/define-a-theme
//
// Tokens resolve at apply time — when you call AddShape / AddRun — against the
// presentation's active theme (D-033). The theme must therefore be set before
// you author content. A "theme swap" is just running the same builder calls
// under a different active theme: that is exactly what buildDeck does below.
package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/hurtener/pptx-go/pptx"
)

// buildDeck authors one identical slide — a token-filled card plus a
// token-typed heading — under whichever theme it is handed, then serializes it.
// The builder calls never mention a literal color or font: every visual
// property flows through a semantic token, so the bytes this returns reflect
// the supplied theme's palette and typography.
func buildDeck(theme *pptx.Theme) ([]byte, error) {
	pres := pptx.New(pptx.WithTheme(theme))
	slide := pres.AddSlide()

	// A rounded card filled with the ACCENT surface token and rounded with the
	// LG radius token. Swap the theme and the same call paints a new color.
	card := pptx.Box{X: pptx.In(1), Y: pptx.In(1), W: pptx.In(8), H: pptx.In(2.5)}
	slide.AddShape(
		pptx.ShapeRoundRect,
		card,
		pptx.WithFill(pptx.SolidFill(pptx.TokenColor(pptx.ColorAccent))),
		pptx.WithRadius(pptx.RadiusLG),
		pptx.WithElevation(pptx.ElevationRaised),
	)

	// A heading typed with the H1 typography token and colored with the
	// INVERSE text token (legible on the accent card).
	tf := slide.AddTextFrame(card.Inset(pptx.Inset{
		Top: pptx.In(0.4), Right: pptx.In(0.4), Bottom: pptx.In(0.4), Left: pptx.In(0.4),
	}))
	para := tf.AddParagraph(pptx.ParagraphOpts{Align: pptx.AlignCenter})
	para.AddRun("Same input, themed surface", pptx.RunStyle{
		TypeRole: pptx.TypeH1,
		Color:    pptx.TokenTextColor(pptx.TextInverse),
	})

	return pres.WriteToBytes()
}

func main() {
	// Theme A — built with NewTheme + functional options. NewTheme starts from
	// DefaultTheme() and applies each option, so unset roles keep their
	// documented defaults (ColorAccent default is 2563EB).
	themeA := pptx.NewTheme(
		pptx.WithName("Brand A — Indigo"),
		pptx.WithAccent(pptx.RGB("2563EB")),
		pptx.WithFonts("Calibri Light", "Calibri"),
	)

	// Theme B — built by cloning the default theme and mutating the maps
	// directly. Clone() is a deep copy, so these writes never touch the
	// original. This is the full-control path when the option helpers are not
	// enough (here: a custom accent AND a resized H1).
	themeB := pptx.DefaultTheme().Clone()
	themeB.Name = "Brand B — Magenta"
	themeB.Colors.Surfaces[pptx.ColorAccent] = pptx.RGB("DB2777")
	themeB.Colors.Text[pptx.TextInverse] = pptx.RGB("FFFFFF")
	h1 := themeB.Typography[pptx.TypeH1]
	h1.Size = 40
	h1.Family = "Georgia"
	themeB.Typography[pptx.TypeH1] = h1

	// Render the SAME builder input under each theme.
	deckA, err := buildDeck(themeA)
	if err != nil {
		log.Fatalf("build deck A: %v", err)
	}
	deckB, err := buildDeck(themeB)
	if err != nil {
		log.Fatalf("build deck B: %v", err)
	}

	if bytes.Equal(deckA, deckB) {
		log.Fatalf("expected the two themed decks to differ, but the bytes were identical")
	}

	fmt.Printf("OK: theme swap re-skinned the same input — %q produced %d bytes, %q produced %d bytes (content differs).\n",
		themeA.Name, len(deckA), themeB.Name, len(deckB))
}
