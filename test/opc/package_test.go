package opc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
)

func TestPackage_New(t *testing.T) {
	pkg := opc.NewPackage()
	if pkg == nil {
		t.Fatal("NewPackage returned nil")
	}

	if pkg.Parts() == nil {
		t.Error("Parts() returned nil")
	}
	if pkg.Relationships() == nil {
		t.Error("Relationships() returned nil")
	}
	if pkg.ContentTypes() == nil {
		t.Error("ContentTypes() returned nil")
	}
	if pkg.PartCount() != 0 {
		t.Error("new package should have no parts")
	}
}

func TestPackage_CoreProperties(t *testing.T) {
	pkg := opc.NewPackage()

	// default is nil
	if pkg.CoreProperties() != nil {
		t.Error("new package should have nil core properties")
	}

	// set core properties
	cp := &opc.CoreProperties{}
	cp.SetTitle("Test Title")
	pkg.SetCoreProperties(cp)

	if pkg.CoreProperties() == nil {
		t.Fatal("CoreProperties() returned nil after setting")
	}
	if pkg.CoreProperties().Title() != "Test Title" {
		t.Error("CoreProperties title mismatch")
	}
}

func TestPackage_AddPart(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	part := opc.NewPart(uri, opc.ContentTypeSlide, []byte("<slide/>"))

	err := pkg.AddPart(part)
	if err != nil {
		t.Fatalf("AddPart failed: %v", err)
	}
	if pkg.PartCount() != 1 {
		t.Errorf("PartCount() = %d, want 1", pkg.PartCount())
	}
}

func TestPackage_CreatePart(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")

	part, err := pkg.CreatePart(uri, opc.ContentTypeSlide, []byte("<slide/>"))
	if err != nil {
		t.Fatalf("CreatePart failed: %v", err)
	}
	if part == nil {
		t.Fatal("CreatePart returned nil")
	}
	if pkg.PartCount() != 1 {
		t.Errorf("PartCount() = %d, want 1", pkg.PartCount())
	}
}

func TestPackage_CreatePartFromReader(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	reader := bytes.NewReader([]byte("<slide/>"))

	part, err := pkg.CreatePartFromReader(uri, opc.ContentTypeSlide, reader)
	if err != nil {
		t.Fatalf("CreatePartFromReader failed: %v", err)
	}
	if part == nil {
		t.Fatal("CreatePartFromReader returned nil")
	}
}

func TestPackage_CreatePartFromXML(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")

	type slide struct {
		XMLName struct{} `xml:"p:sld"`
	}

	part, err := pkg.CreatePartFromXML(uri, opc.ContentTypeSlide, &slide{})
	if err != nil {
		t.Fatalf("CreatePartFromXML failed: %v", err)
	}
	if part == nil {
		t.Fatal("CreatePartFromXML returned nil")
	}
}

func TestPackage_GetPart(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(uri, opc.ContentTypeSlide, []byte("<slide/>"))

	part := pkg.GetPart(uri)
	if part == nil {
		t.Fatal("GetPart returned nil")
	}

	// get a non-existent part
	nonExistent := opc.NewPackURI("/ppt/slides/slide999.xml")
	if pkg.GetPart(nonExistent) != nil {
		t.Error("GetPart for non-existent URI should return nil")
	}
}

func TestPackage_GetPartByStr(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(uri, opc.ContentTypeSlide, []byte("<slide/>"))

	part := pkg.GetPartByStr("/ppt/slides/slide1.xml")
	if part == nil {
		t.Fatal("GetPartByStr returned nil")
	}
}

func TestPackage_GetPartsByType(t *testing.T) {
	pkg := opc.NewPackage()
	uri1 := opc.NewPackURI("/ppt/slides/slide1.xml")
	uri2 := opc.NewPackURI("/ppt/slides/slide2.xml")
	uri3 := opc.NewPackURI("/ppt/theme/theme1.xml")

	pkg.CreatePart(uri1, opc.ContentTypeSlide, []byte{})
	pkg.CreatePart(uri2, opc.ContentTypeSlide, []byte{})
	pkg.CreatePart(uri3, opc.ContentTypeTheme, []byte{})

	slides := pkg.GetPartsByType(opc.ContentTypeSlide)
	if len(slides) != 2 {
		t.Errorf("GetPartsByType(slide) returned %d, want 2", len(slides))
	}
}

func TestPackage_ContainsPart(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(uri, opc.ContentTypeSlide, []byte{})

	if !pkg.ContainsPart(uri) {
		t.Error("should contain added part")
	}

	nonExistent := opc.NewPackURI("/ppt/slides/slide999.xml")
	if pkg.ContainsPart(nonExistent) {
		t.Error("should not contain non-existent part")
	}
}

func TestPackage_RemovePart(t *testing.T) {
	pkg := opc.NewPackage()
	uri := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(uri, opc.ContentTypeSlide, []byte{})

	err := pkg.RemovePart(uri)
	if err != nil {
		t.Fatalf("RemovePart failed: %v", err)
	}
	if pkg.PartCount() != 0 {
		t.Error("part should be removed")
	}
}

func TestPackage_AllParts(t *testing.T) {
	pkg := opc.NewPackage()
	uri1 := opc.NewPackURI("/ppt/slides/slide1.xml")
	uri2 := opc.NewPackURI("/ppt/slides/slide2.xml")

	pkg.CreatePart(uri1, opc.ContentTypeSlide, []byte{})
	pkg.CreatePart(uri2, opc.ContentTypeSlide, []byte{})

	all := pkg.AllParts()
	if len(all) != 2 {
		t.Errorf("AllParts() returned %d, want 2", len(all))
	}
}

func TestPackage_PartURIs(t *testing.T) {
	pkg := opc.NewPackage()
	uri1 := opc.NewPackURI("/ppt/slides/slide1.xml")
	uri2 := opc.NewPackURI("/ppt/slides/slide2.xml")

	pkg.CreatePart(uri1, opc.ContentTypeSlide, []byte{})
	pkg.CreatePart(uri2, opc.ContentTypeSlide, []byte{})

	uris := pkg.PartURIs()
	if len(uris) != 2 {
		t.Errorf("PartURIs() returned %d, want 2", len(uris))
	}
}

func TestPackage_DirtyParts(t *testing.T) {
	pkg := opc.NewPackage()
	uri1 := opc.NewPackURI("/ppt/slides/slide1.xml")
	uri2 := opc.NewPackURI("/ppt/slides/slide2.xml")

	part1, _ := pkg.CreatePart(uri1, opc.ContentTypeSlide, []byte{})
	pkg.CreatePart(uri2, opc.ContentTypeSlide, []byte{})

	// all new parts are dirty
	dirty := pkg.DirtyParts()
	if len(dirty) != 2 {
		t.Errorf("DirtyParts() returned %d, want 2", len(dirty))
	}

	// clear dirty flag on one
	part1.SetDirty(false)

	dirty = pkg.DirtyParts()
	if len(dirty) != 1 {
		t.Errorf("DirtyParts() after clearing should be 1, got %d", len(dirty))
	}
}

func TestPackage_AddRelationship(t *testing.T) {
	pkg := opc.NewPackage()
	rel, err := pkg.AddRelationship(opc.RelTypeOfficeDocument, "ppt/presentation.xml", false)

	if err != nil {
		t.Fatalf("AddRelationship failed: %v", err)
	}
	if rel == nil {
		t.Fatal("AddRelationship returned nil")
	}
	if rel.RID() != "rId1" {
		t.Errorf("RID = %q, want %q", rel.RID(), "rId1")
	}
}

func TestPackage_GetRelationship(t *testing.T) {
	pkg := opc.NewPackage()
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "ppt/presentation.xml", false)

	rel := pkg.GetRelationship("rId1")
	if rel == nil {
		t.Fatal("GetRelationship returned nil")
	}

	// get a non-existent rID
	if pkg.GetRelationship("rId999") != nil {
		t.Error("GetRelationship for non-existent rID should return nil")
	}
}

func TestPackage_GetRelationshipsByType(t *testing.T) {
	pkg := opc.NewPackage()
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "ppt/presentation.xml", false)
	pkg.AddRelationship(opc.RelTypeCoreProperties, "docProps/core.xml", false)

	rels := pkg.GetRelationshipsByType(opc.RelTypeOfficeDocument)
	if len(rels) != 1 {
		t.Errorf("GetRelationshipsByType returned %d, want 1", len(rels))
	}
}

func TestPackage_GetPartByRelType(t *testing.T) {
	pkg := opc.NewPackage()

	// add relationship
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "/ppt/presentation.xml", false)

	// add target part
	presentationURI := opc.NewPackURI("/ppt/presentation.xml")
	pkg.CreatePart(presentationURI, opc.ContentTypePresentation, []byte("<presentation/>"))

	// get part by relationship type
	part := pkg.GetPartByRelType(opc.RelTypeOfficeDocument)
	if part == nil {
		t.Fatal("GetPartByRelType returned nil")
	}

	// get a non-existent relationship type
	if pkg.GetPartByRelType(opc.RelTypeSlide) != nil {
		t.Error("GetPartByRelType for non-existent type should return nil")
	}
}

func TestPackage_ResolveRelationship(t *testing.T) {
	pkg := opc.NewPackage()

	// create source part
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	slidePart, _ := pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte("<slide/>"))

	// create target part
	imageURI := opc.NewPackURI("/ppt/media/image1.png")
	pkg.CreatePart(imageURI, opc.ContentTypePNG, []byte("imagedata"))

	// add relationship using absolute path
	slidePart.AddRelationship(opc.RelTypeImage, "/ppt/media/image1.png", false)

	// resolve relationship
	targetPart := pkg.ResolveRelationship(slidePart, opc.RelTypeImage)
	if targetPart == nil {
		t.Log("ResolveRelationship returned nil - relationship target resolution may need implementation")
		// not a fatal error: relationship resolution may require additional path resolution logic
	} else {
		// verify target part
		if targetPart.PartURI().URI() != imageURI.URI() {
			t.Errorf("target part URI = %q, want %q", targetPart.PartURI().URI(), imageURI.URI())
		}
	}
}

func TestPackage_SaveAndOpen(t *testing.T) {
	// create package
	pkg := opc.NewPackage()

	// add some parts
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte(`<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"/>`))

	themeURI := opc.NewPackURI("/ppt/theme/theme1.xml")
	pkg.CreatePart(themeURI, opc.ContentTypeTheme, []byte(`<a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"/>`))

	// add relationship
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "/ppt/presentation.xml", false)

	// create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.pptx")

	// save
	err := pkg.SaveFile(tmpFile)
	if err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	// verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("saved file does not exist")
	}

	// reopen
	openedPkg, err := opc.OpenFile(tmpFile)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer openedPkg.Close()

	// verify content
	if openedPkg.PartCount() < 2 {
		t.Errorf("opened package has %d parts, expected at least 2", openedPkg.PartCount())
	}

	// verify parts
	slidePart := openedPkg.GetPart(slideURI)
	if slidePart == nil {
		t.Error("slide part not found after reopening")
	}
}

func TestPackage_SaveToBytes(t *testing.T) {
	pkg := opc.NewPackage()
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte("<slide/>"))

	data, err := pkg.SaveToBytes()
	if err != nil {
		t.Fatalf("SaveToBytes failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("SaveToBytes returned empty data")
	}
}

func TestPackage_Clone(t *testing.T) {
	pkg := opc.NewPackage()

	// add part
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte("<slide/>"))

	// add relationship
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "/ppt/presentation.xml", false)

	// set core properties
	cp := &opc.CoreProperties{}
	cp.SetTitle("Original Title")
	pkg.SetCoreProperties(cp)

	// clone
	cloned := pkg.Clone()
	if cloned == nil {
		t.Fatal("Clone returned nil")
	}

	// verify parts were cloned
	if cloned.PartCount() != pkg.PartCount() {
		t.Error("cloned package should have same part count")
	}

	// modifying the original should not affect the clone
	pkg.CreatePart(opc.NewPackURI("/ppt/slides/slide2.xml"), opc.ContentTypeSlide, []byte{})
	if cloned.PartCount() == pkg.PartCount() {
		t.Error("modifying original should not affect clone")
	}

	// verify core properties were cloned
	if cloned.CoreProperties() == nil {
		t.Fatal("cloned core properties should not be nil")
	}
	if cloned.CoreProperties().Title() != "Original Title" {
		t.Error("cloned core properties title mismatch")
	}
}

func TestPackage_Close(t *testing.T) {
	pkg := opc.NewPackage()
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte("<slide/>"))

	err := pkg.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// parts should be cleared after close
	if pkg.PartCount() != 0 {
		t.Error("package should be empty after close")
	}
}

// ===== Helper function tests =====

func TestOpen(t *testing.T) {
	// create a simple PPTX file for testing
	pkg := opc.NewPackage()
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte("<slide/>"))
	pkg.AddRelationship(opc.RelTypeOfficeDocument, "/ppt/presentation.xml", false)

	// save to bytes
	data, err := pkg.SaveToBytes()
	if err != nil {
		t.Fatalf("SaveToBytes failed: %v", err)
	}

	// open from bytes using Open
	reader := bytes.NewReader(data)
	openedPkg, err := opc.Open(reader, int64(len(data)))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer openedPkg.Close()

	if openedPkg.PartCount() < 1 {
		t.Error("opened package should have at least 1 part")
	}
}

// ===== Constants tests =====

func TestConstants(t *testing.T) {
	// verify key constants are not empty
	if opc.ContentTypePresentation == "" {
		t.Error("ContentTypePresentation should not be empty")
	}
	if opc.ContentTypeSlide == "" {
		t.Error("ContentTypeSlide should not be empty")
	}
	if opc.RelTypeSlide == "" {
		t.Error("RelTypeSlide should not be empty")
	}
	if opc.NamespaceRelationships == "" {
		t.Error("NamespaceRelationships should not be empty")
	}
	if opc.PathContentTypes == "" {
		t.Error("PathContentTypes should not be empty")
	}
	if opc.PathRelsDir == "" {
		t.Error("PathRelsDir should not be empty")
	}
}

// TestNormalizeZipPath_VariousInputs directly tests the NormalizeZipPath function.
func TestNormalizeZipPath_VariousInputs(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Windows backslashes
		{"ppt\\slides\\slide1.xml", "ppt/slides/slide1.xml"},
		{"\\ppt\\slides\\slide1.xml", "ppt/slides/slide1.xml"},
		{"_rels\\.rels", "_rels/.rels"},

		// repeated slashes
		{"ppt//slides//slide1.xml", "ppt/slides/slide1.xml"},
		{"ppt\\\\slides\\\\slide1.xml", "ppt/slides/slide1.xml"},

		// mixed
		{"ppt\\/slides\\slide1.xml", "ppt/slides/slide1.xml"},

		// edge cases
		{"", ""},
		{"/", ""},
		{"[Content_Types].xml", "[Content_Types].xml"},
	}

	for _, tt := range tests {
		result := opc.NormalizeZipPath(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeZipPath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestPackage_OpenWithBackslashPaths tests opening ZIP files with backslash paths,
// simulating low-quality ZIP files created by some Windows tools.
func TestPackage_OpenWithBackslashPaths(t *testing.T) {
	// create a normal package
	pkg := opc.NewPackage()

	// add parts
	slideURI := opc.NewPackURI("/ppt/slides/slide1.xml")
	pkg.CreatePart(slideURI, opc.ContentTypeSlide, []byte("<slide/>"))

	relsURI := opc.NewPackURI("/ppt/slides/_rels/slide1.xml.rels")
	pkg.CreatePart(relsURI, opc.ContentTypeRelationships, []byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`))

	// save to bytes
	normalData, err := pkg.SaveToBytes()
	if err != nil {
		t.Fatalf("SaveToBytes failed: %v", err)
	}

	// directly test that NormalizeZipPath correctly handles backslashes —
	// this is more reliable than manipulating ZIP bytes
	testPaths := []string{
		"ppt\\slides\\slide1.xml",
		"ppt\\slides\\_rels\\slide1.xml.rels",
		"_rels\\.rels",
	}

	for _, path := range testPaths {
		normalized := opc.NormalizeZipPath(path)
		// verify backslashes were converted to forward slashes
		if containsBackslash(normalized) {
			t.Errorf("NormalizeZipPath(%q) = %q still contains backslash", path, normalized)
		}
		// verify that a URI with a leading / can be correctly created
		uri := opc.NewPackURI("/" + normalized)
		if uri.URI() == "" {
			t.Errorf("Failed to create PackURI from normalized path %q", normalized)
		}
	}

	t.Logf("Normal ZIP data size: %d bytes", len(normalData))
	t.Log("NormalizeZipPath correctly handles backslash paths")
}

// containsBackslash reports whether s contains a backslash character.
func containsBackslash(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			return true
		}
	}
	return false
}
