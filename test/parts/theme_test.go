package parts_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/parts"
)

func TestThemePart_FromXML(t *testing.T) {
	// Use actual theme XML data.
	themeXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme">
	<a:themeElements>
		<a:clrScheme name="Office">
			<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
			<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
			<a:dk2><a:srgbClr val="44546A"/></a:dk2>
			<a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
			<a:accent1><a:srgbClr val="5B9BD5"/></a:accent1>
			<a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
			<a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
			<a:accent4><a:srgbClr val="FFC000"/></a:accent4>
			<a:accent5><a:srgbClr val="4472C4"/></a:accent5>
			<a:accent6><a:srgbClr val="70AD47"/></a:accent6>
			<a:hlink><a:srgbClr val="0563C1"/></a:hlink>
			<a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
		</a:clrScheme>
	</a:themeElements>
</a:theme>`

	themePart := parts.NewThemePart(1)
	err := themePart.FromXML([]byte(themeXML))
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Verify the theme name.
	theme := themePart.Theme()
	if theme == nil {
		t.Fatal("Theme is nil")
	}
	if theme.Name != "Office Theme" {
		t.Errorf("Expected theme name 'Office Theme', got '%s'", theme.Name)
	}

	// Verify the color scheme.
	colorScheme := themePart.ColorScheme()
	if colorScheme == nil {
		t.Fatal("ColorScheme is nil")
	}
	if colorScheme.Name != "Office" {
		t.Errorf("Expected color scheme name 'Office', got '%s'", colorScheme.Name)
	}
}

func TestThemePart_GetColor(t *testing.T) {
	themeXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Test Theme">
	<a:themeElements>
		<a:clrScheme name="Test">
			<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
			<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
			<a:dk2><a:srgbClr val="44546A"/></a:dk2>
			<a:lt2><a:srgbClr val="E7E6E6"/></a:lt2>
			<a:accent1><a:srgbClr val="5B9BD5"/></a:accent1>
			<a:accent2><a:srgbClr val="ED7D31"/></a:accent2>
			<a:accent3><a:srgbClr val="A5A5A5"/></a:accent3>
			<a:accent4><a:srgbClr val="FFC000"/></a:accent4>
			<a:accent5><a:srgbClr val="4472C4"/></a:accent5>
			<a:accent6><a:srgbClr val="70AD47"/></a:accent6>
			<a:hlink><a:srgbClr val="0563C1"/></a:hlink>
			<a:folHlink><a:srgbClr val="954F72"/></a:folHlink>
		</a:clrScheme>
	</a:themeElements>
</a:theme>`

	themePart := parts.NewThemePart(1)
	if err := themePart.FromXML([]byte(themeXML)); err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	tests := []struct {
		role     parts.ColorRole
		expected string
		isSystem bool
	}{
		{parts.ColorRoleDark1, "000000", true},   // system color
		{parts.ColorRoleLight1, "FFFFFF", true},  // system color
		{parts.ColorRoleDark2, "44546A", false},  // RGB color
		{parts.ColorRoleLight2, "E7E6E6", false}, // RGB color
		{parts.ColorRoleAccent1, "5B9BD5", false},
		{parts.ColorRoleAccent2, "ED7D31", false},
		{parts.ColorRoleAccent3, "A5A5A5", false},
		{parts.ColorRoleAccent4, "FFC000", false},
		{parts.ColorRoleAccent5, "4472C4", false},
		{parts.ColorRoleAccent6, "70AD47", false},
		{parts.ColorRoleHyperlink, "0563C1", false},
		{parts.ColorRoleFollowedHyperlink, "954F72", false},
	}

	for _, tt := range tests {
		t.Run(tt.role.String(), func(t *testing.T) {
			rgb := themePart.GetThemeColorRGB(tt.role)
			if rgb != tt.expected {
				t.Errorf("GetThemeColorRGB(%v) = '%s', want '%s'", tt.role, rgb, tt.expected)
			}

			colorType := themePart.GetThemeColorType(tt.role)
			if tt.isSystem && colorType != parts.ColorTypeSystem {
				t.Errorf("Expected ColorTypeSystem for %v, got %v", tt.role, colorType)
			}
			if !tt.isSystem && colorType != parts.ColorTypeRGB {
				t.Errorf("Expected ColorTypeRGB for %v, got %v", tt.role, colorType)
			}
		})
	}
}

func TestThemePart_New(t *testing.T) {
	themePart := parts.NewThemePart(1)
	if themePart == nil {
		t.Fatal("NewThemePart returned nil")
	}

	// Verify URI.
	uri := themePart.PartURI()
	if uri == nil {
		t.Fatal("PartURI is nil")
	}
	if uri.URI() != "/ppt/theme/theme1.xml" {
		t.Errorf("Expected URI '/ppt/theme/theme1.xml', got '%s'", uri.URI())
	}
}

// ============================================================================
// Test 1: XML round-trip losslessness test
// Goal: prove the Go struct captures all required XML tags without data loss.
// ============================================================================
func TestTheme_RoundTrip(t *testing.T) {
	// Use the full default theme XML.
	theme := parts.DefaultTheme()
	if theme == nil {
		t.Fatal("DefaultTheme returned nil")
	}

	// Serialize.
	outputBytes, err := xml.Marshal(theme)
	if err != nil {
		t.Fatalf("serializing Theme failed: %v", err)
	}
	outputXML := string(outputBytes)

	// Core check: the three main pillars must all be present.
	requiredTags := []string{
		"<clrScheme",
		"<fontScheme",
		"<fmtScheme",
		"<accent1",
		"<fillStyleLst",
		"<lnStyleLst",
		"<effectStyleLst",
		"<bgFillStyleLst",
	}

	for _, tag := range requiredTags {
		if !strings.Contains(outputXML, tag) {
			t.Errorf("round-trip failed: key tag %s was lost after serialization", tag)
		}
	}
	t.Log("Theme round-trip losslessness test passed")
}

// ============================================================================
// Test 2: Concurrent clone safety test
// Goal: prove that a CloneTheme copy can be mutated without polluting the global template.
// ============================================================================
func TestTheme_CloneSafety(t *testing.T) {
	// Obtain two independent clones.
	themeA := parts.CloneTheme()
	themeB := parts.CloneTheme()

	if themeA == nil || themeB == nil {
		t.Fatal("CloneTheme returned nil")
	}

	// Create a ThemePart and assign themeA.
	partA := parts.NewThemePart(1)
	partA.SetThemeData(themeA)

	// Mutate themeA Accent1 to red.
	partA.SetThemeColorRGB(parts.ColorRoleAccent1, "FF0000")

	// Verify themeB was not polluted (it should retain the original value 5B9BD5).
	partB := parts.NewThemePart(2)
	partB.SetThemeData(themeB)
	colorB := partB.GetThemeColorRGB(parts.ColorRoleAccent1)

	if colorB == "FF0000" {
		t.Fatal("clone safety test failed: shallow copy caused memory aliasing — mutating A affected B")
	}
	if colorB != "5B9BD5" {
		t.Errorf("expected themeB Accent1 '5B9BD5', got '%s'", colorB)
	}

	t.Log("Theme clone isolation test passed")
}

// ============================================================================
// Test 3: Low-level mutator behavior test
// Goal: prove that SetThemeColorRGB correctly modifies the XML node tree.
// ============================================================================
func TestTheme_SetThemeColorRGB(t *testing.T) {
	theme := parts.CloneTheme()
	part := parts.NewThemePart(1)
	part.SetThemeData(theme)

	// Change Accent1 to pure green.
	targetColor := "00FF00"
	part.SetThemeColorRGB(parts.ColorRoleAccent1, targetColor)

	// Verify the retrieved color reflects the new value.
	gotColor := part.GetThemeColorRGB(parts.ColorRoleAccent1)
	if gotColor != targetColor {
		t.Errorf("mutator test failed: expected '%s', got '%s'", targetColor, gotColor)
	}

	// After serialization the RGB value must be present in the XML string.
	themeData := part.Theme()
	outputBytes, err := xml.Marshal(themeData)
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}
	outputXML := string(outputBytes)

	expectedNode := `<srgbClr val="00FF00"`
	if !strings.Contains(outputXML, expectedNode) {
		t.Errorf("mutator test failed: modified color node %s not found in XML tree", expectedNode)
	}

	t.Log("Theme low-level mutator test passed")
}
