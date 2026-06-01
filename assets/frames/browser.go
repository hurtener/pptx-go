package frames

import "github.com/hurtener/pptx-go/pptx"

// Browser draws a browser-window bezel: a rounded window (the chrome) with a
// top toolbar strip carrying three traffic-light dots. The interior is the
// content viewport below the toolbar, inset on the other three sides. The
// window body is a single rounded rect so its rounded corners are never
// undercut by a square toolbar rect — the toolbar is simply the strip of the
// window above the interior.
func Browser(sl *pptx.Slide, region pptx.Box) (interior pptx.Box, shapes int) {
	// Window body / chrome (rounded). Its fill is the toolbar color; the image
	// covers the content area, leaving the top strip reading as a toolbar.
	sl.AddShape(pptx.ShapeRoundRect, region, fill(pptx.ColorSurface), pptx.WithRadius(pptx.RadiusMD))
	shapes++

	// Toolbar strip height: ~12% of the region, with a sane minimum.
	barH := maxEMU(region.H*12/100, pptx.In(0.3))
	if barH > region.H/2 {
		barH = region.H / 2
	}

	// Three traffic-light dots, left-aligned in the strip.
	dia := barH * 36 / 100
	gap := dia / 2
	pad := dia
	dy := region.Y + (barH-dia)/2
	for i, role := range []pptx.ColorRole{pptx.ColorError, pptx.ColorWarning, pptx.ColorSuccess} {
		dot := pptx.Box{
			X: region.X + pad + pptx.EMU(i)*(dia+gap),
			Y: dy,
			W: dia,
			H: dia,
		}
		sl.AddShape(pptx.ShapeEllipse, dot, fill(role))
		shapes++
	}

	inset := region.W * 2 / 100
	interior = pptx.Box{
		X: region.X + inset,
		Y: region.Y + barH,
		W: region.W - 2*inset,
		H: region.H - barH - inset,
	}
	return interior, shapes
}
