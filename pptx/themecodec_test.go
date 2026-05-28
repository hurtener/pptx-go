package pptx

import (
	"bytes"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/theme"
	"github.com/hurtener/pptx-go/internal/opc"
)

// TestThemeRoundTripOOXML proves a theme survives the trip through theme1.xml
// for the roles that have OOXML scheme slots (acceptance §11.1).
func TestThemeRoundTripOOXML(t *testing.T) {
	src := NewTheme(WithName("Brand"), WithAccent("AB12CD"), WithFonts("Inter", "Inter"))

	xml, err := src.ThemeXML()
	if err != nil {
		t.Fatalf("ThemeXML: %v", err)
	}

	tp := theme.NewThemePart(1)
	if err := tp.FromXML(xml); err != nil {
		t.Fatalf("FromXML: %v", err)
	}
	got := themeFromPart(tp)

	if got.ResolveColor(ColorAccent) != "AB12CD" {
		t.Errorf("accent: got %q want AB12CD", got.ResolveColor(ColorAccent))
	}
	if got.ResolveTextColor(TextPrimary) != src.ResolveTextColor(TextPrimary) {
		t.Errorf("text-primary: got %q want %q", got.ResolveTextColor(TextPrimary), src.ResolveTextColor(TextPrimary))
	}
	if got.HeadingFont != "Inter" || got.BodyFont != "Inter" {
		t.Errorf("fonts: heading=%q body=%q", got.HeadingFont, got.BodyFont)
	}
	if got.ResolveType(TypeH1).Family != "Inter" {
		t.Errorf("H1 family after round-trip: %q", got.ResolveType(TypeH1).Family)
	}
}

// TestLoadThemeFromBytes proves LoadThemeFromBytes pulls the accent out of a
// package's theme1.xml (acceptance §11.2).
func TestLoadThemeFromBytes(t *testing.T) {
	src := NewTheme(WithAccent("123456"))
	xml, err := src.ThemeXML()
	if err != nil {
		t.Fatal(err)
	}

	pkg := opc.NewPackage()
	if _, err := pkg.CreatePart(opc.NewPackURI("/ppt/theme/theme1.xml"), opc.ContentTypeTheme, xml); err != nil {
		t.Fatalf("CreatePart: %v", err)
	}
	var buf bytes.Buffer
	if err := pkg.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadThemeFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("LoadThemeFromBytes: %v", err)
	}
	if loaded.ResolveColor(ColorAccent) != "123456" {
		t.Errorf("loaded accent: got %q want 123456", loaded.ResolveColor(ColorAccent))
	}
}

func TestLoadThemeNoThemePart(t *testing.T) {
	pkg := opc.NewPackage()
	var buf bytes.Buffer
	if err := pkg.Save(&buf); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadThemeFromBytes(buf.Bytes()); err == nil {
		t.Fatal("expected error for package with no theme part")
	}
}
