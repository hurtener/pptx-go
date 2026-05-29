package pptx_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// TestNew_DefaultFormat confirms a no-option deck is 16:9 and themed.
func TestNew_DefaultFormat(t *testing.T) {
	pres := pptx.New()
	cx, cy := pres.SlideSize()
	if cx != int(pptx.Slide16x9Width) || cy != int(pptx.Slide16x9Height) {
		t.Errorf("default size = (%d, %d), want 16:9 (%d, %d)", cx, cy, int(pptx.Slide16x9Width), int(pptx.Slide16x9Height))
	}
	if pres.Theme() == nil {
		t.Fatal("Theme() is nil; expected DefaultTheme")
	}
}

// TestWithFormat_4x3 confirms WithFormat sizes the canvas and that it reaches
// the emitted presentation.xml.
func TestWithFormat_4x3(t *testing.T) {
	pres := pptx.New(pptx.WithFormat(pptx.Slides4x3))
	pres.AddSlide()

	cx, cy := pres.SlideSize()
	if cx != int(pptx.Slide4x3Width) || cy != int(pptx.Slide4x3Height) {
		t.Errorf("size = (%d, %d), want 4:3 (%d, %d)", cx, cy, int(pptx.Slide4x3Width), int(pptx.Slide4x3Height))
	}

	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	pres1 := readZipPart(t, data, "ppt/presentation.xml")
	if !strings.Contains(pres1, `<p:sldSz cx="9144000" cy="6858000"/>`) {
		t.Errorf("presentation.xml missing 4:3 sldSz in:\n%s", pres1)
	}
}

// TestWithTheme_and_SetTheme confirms the theme option and setter take effect.
func TestWithTheme_and_SetTheme(t *testing.T) {
	custom := pptx.DefaultTheme()
	custom.Name = "brand"

	pres := pptx.New(pptx.WithTheme(custom))
	if pres.Theme().Name != "brand" {
		t.Errorf("WithTheme: Theme().Name = %q, want brand", pres.Theme().Name)
	}

	another := pptx.DefaultTheme()
	another.Name = "brand2"
	pres.SetTheme(another)
	if pres.Theme().Name != "brand2" {
		t.Errorf("SetTheme: Theme().Name = %q, want brand2", pres.Theme().Name)
	}

	// A nil theme is ignored, not adopted.
	pres.SetTheme(nil)
	if pres.Theme().Name != "brand2" {
		t.Errorf("SetTheme(nil) overwrote the theme: %q", pres.Theme().Name)
	}
}

// fakeFontSource is a no-op FontSource for option wiring.
type fakeFontSource struct{}

func (fakeFontSource) Resolve(string, string, int) ([]byte, error) { return []byte{0x01}, nil }

// TestWithFontSource confirms the option registers a source EmbedFont can use.
func TestWithFontSource(t *testing.T) {
	pres := pptx.New(pptx.WithFontSource(fakeFontSource{}))
	// With a source registered, EmbedFont must not return ErrNoFontSource.
	if err := pres.EmbedFont("Arial", "regular", 400); err != nil {
		t.Errorf("EmbedFont with WithFontSource: unexpected error %v", err)
	}
}
