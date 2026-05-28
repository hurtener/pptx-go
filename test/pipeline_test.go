package pptx_test

import (
	"archive/zip"
	"os"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/opc"
	"github.com/hurtener/pptx-go/parts"
	"github.com/hurtener/pptx-go/pptx"
)

// ============================================================================
// Full-capability pipeline test for pptx-go
// ============================================================================
//
// Goal: prove the complete data flow from parsing to creation across the
// entire library stack.
//
// Stages:
//   Stage 1: OPC layer — parsing and unpacking
//   Stage 2: Parts layer — deserialization
//   Stage 3: Parts layer — creation and serialization
//   Stage 4: OPC layer — writing and routing (critical)
//   Stage 5: Parts layer — data update
//   Stage 6: OPC layer — safe repackaging
//
// Dependency: test/test.pptx (a real PPTX file)
// ============================================================================

const testPPTXPath = "test.pptx"

// requireTestPPTX skips the test when the test.pptx fixture is absent. The
// fixture is gitignored and was never committed upstream, so it is not present
// in a clean checkout or in CI. (Phase 01 owns providing/relocating fixtures.)
func requireTestPPTX(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(testPPTXPath); err != nil {
		t.Skipf("fixture %s not present; skipping (gitignored, not committed upstream)", testPPTXPath)
	}
}

// ============================================================================
// Stage 1: OPC layer — parsing and unpacking
// ============================================================================

// TestOPC_ParseAndUnpack proves that the OPC layer can correctly parse and
// unpack a PPTX file.
func TestOPC_ParseAndUnpack(t *testing.T) {
	requireTestPPTX(t)

	// 1.2 Open the OPC package.
	pkg, err := opc.OpenFile(testPPTXPath)
	if err != nil {
		t.Fatalf("opc.OpenFile failed: %v", err)
	}
	defer pkg.Close()

	// 1.3 Verify package-level relationships exist.
	rootRels := pkg.Relationships()
	if rootRels == nil {
		t.Fatal("package-level relationships are nil")
	}
	if rootRels.Count() == 0 {
		t.Fatal("package-level relationship count is 0")
	}

	// 1.4 Verify ContentTypes loaded correctly.
	ct := pkg.ContentTypes()
	if ct == nil {
		t.Fatal("ContentTypes is nil")
	}

	// 1.5 Verify part count (test.pptx contains several parts).
	partCount := pkg.PartCount()
	if partCount == 0 {
		t.Fatal("part count is 0")
	}
	t.Logf("Stage 1: OPC unpack succeeded, part count = %d", partCount)

	// 1.6 Verify required parts are present.
	requiredParts := []string{
		"/ppt/presentation.xml",
		"/ppt/slides/slide1.xml",
		"/ppt/slideMasters/slideMaster1.xml",
	}
	for _, uri := range requiredParts {
		if !pkg.ContainsPart(opc.NewPackURI(uri)) {
			t.Errorf("missing required part: %s", uri)
		}
	}

	// 1.7 Verify ZIP structural integrity.
	file, _ := os.Open(testPPTXPath)
	defer file.Close()
	stat, _ := file.Stat()
	zipReader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		t.Fatalf("ZIP read failed: %v", err)
	}

	files := make(map[string]bool)
	for _, f := range zipReader.File {
		files[f.Name] = true
	}

	// Verify no ZIP entries have a leading "/" (Windows ZIP spec violation).
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "/") {
			t.Errorf("ZIP path violation (leading /): %s", f.Name)
		}
	}

	t.Logf("Stage 1: OPC parse and unpack passed (files=%d)", len(files))
	_ = files // suppress unused-variable warning
}

// ============================================================================
// Stage 2: Parts layer — deserialization
// ============================================================================

// TestParts_Deserialize proves that the Parts layer can correctly deserialize
// XML from a PPTX file.
func TestParts_Deserialize(t *testing.T) {
	requireTestPPTX(t)
	// 2.1 Open OPC package.
	pkg, err := opc.OpenFile(testPPTXPath)
	if err != nil {
		t.Fatalf("opc.OpenFile failed: %v", err)
	}
	defer pkg.Close()

	// 2.2 Get PresentationPart and deserialize.
	presPart := pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml"))
	if presPart == nil {
		t.Fatal("missing presentation.xml")
	}

	pres := parts.NewPresentationPart()
	if err := pres.FromXML(presPart.Blob()); err != nil {
		t.Fatalf("PresentationPart.FromXML failed: %v", err)
	}

	// 2.3 Verify PresentationPart data.
	slideCount := pres.SlideCount()
	if slideCount == 0 {
		t.Error("slide count is 0")
	}
	t.Logf("Stage 2: Presentation deserialized, slide count = %d", slideCount)

	slideSize := pres.SlideSize()
	if slideSize.Cx == 0 || slideSize.Cy == 0 {
		t.Error("invalid slide size")
	}
	t.Logf("Stage 2: slide size = %dx%d EMU", slideSize.Cx, slideSize.Cy)

	// 2.4 Get SlidePart and deserialize.
	slidePart := pkg.GetPart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	if slidePart == nil {
		t.Fatal("missing slide1.xml")
	}

	slide := parts.NewSlidePart(1)
	if err := slide.FromXML(slidePart.Blob()); err != nil {
		t.Fatalf("SlidePart.FromXML failed: %v", err)
	}

	// 2.5 Verify SlidePart data.
	shapeCount := slide.ShapeIDCount()
	t.Logf("Stage 2: Slide deserialized, shape count = %d", shapeCount)

	t.Log("Stage 2: Parts layer deserialization passed")
}

// ============================================================================
// Stage 3: Parts layer — creation and serialization
// ============================================================================

// TestParts_CreateAndSerialize proves that the Parts layer can create new
// parts and serialize them.
func TestParts_CreateAndSerialize(t *testing.T) {
	// 3.1 Create a new PresentationPart.
	pres := parts.NewPresentationPart()

	// 3.2 Create a new SlidePart and add text via the builder.
	slidePart := parts.NewSlidePart(1)
	builder := pptx.NewSlideBuilder(slidePart)
	builder.AddTextBox(914400, 457200, 4572000, 457200, "Hello from Go Engine!")

	// 3.3 Serialize PresentationPart.
	presXML, err := pres.ToXML()
	if err != nil {
		t.Fatalf("PresentationPart.ToXML failed: %v", err)
	}
	if len(presXML) == 0 {
		t.Fatal("PresentationPart.ToXML returned empty data")
	}
	if !strings.HasPrefix(string(presXML), "<?xml") {
		t.Error("XML is missing the declaration header")
	}

	// 3.4 Serialize SlidePart.
	slideXML, err := slidePart.ToXML()
	if err != nil {
		t.Fatalf("SlidePart.ToXML failed: %v", err)
	}
	if len(slideXML) == 0 {
		t.Fatal("SlidePart.ToXML returned empty data")
	}
	if !strings.Contains(string(slideXML), "Hello from Go Engine!") {
		t.Error("slide XML does not contain expected text")
	}

	t.Logf("Stage 3: Presentation XML size = %d bytes", len(presXML))
	t.Logf("Stage 3: Slide XML size = %d bytes", len(slideXML))
	t.Log("Stage 3: Parts layer creation and serialization passed")
}

// ============================================================================
// Stage 4: OPC layer — writing and routing (critical)
// ============================================================================

// TestOPC_WriteAndRoute proves that the OPC layer can correctly write parts
// and route relationships.
func TestOPC_WriteAndRoute(t *testing.T) {
	// 4.1 Create a new package.
	pkg := opc.NewPackage()

	// 4.2 Create PresentationPart.
	pres := parts.NewPresentationPart()
	presXML, _ := pres.ToXML()
	presPart, err := pkg.CreatePart(
		opc.NewPackURI("/ppt/presentation.xml"),
		"application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml",
		presXML,
	)
	if err != nil {
		t.Fatalf("CreatePart presentation.xml failed: %v", err)
	}

	// 4.3 Create SlidePart.
	slidePart4 := parts.NewSlidePart(1)
	slideBuilder4 := pptx.NewSlideBuilder(slidePart4)
	slideBuilder4.AddTextBox(914400, 457200, 4572000, 457200, "Hello from OPC!")
	slideXML, _ := slidePart4.ToXML()
	slidePart, err := pkg.CreatePart(
		opc.NewPackURI("/ppt/slides/slide1.xml"),
		"application/vnd.openxmlformats-officedocument.presentationml.slide+xml",
		slideXML,
	)
	if err != nil {
		t.Fatalf("CreatePart slide1.xml failed: %v", err)
	}

	// 4.4 Establish relationships.
	// Package-level rel: points to presentation.xml.
	pkg.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument",
		"/ppt/presentation.xml",
		false,
	)

	// presentation.xml -> slide1.xml
	presPart.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide",
		"slides/slide1.xml",
		false,
	)

	// slide1.xml -> slideLayout1.xml (layout relationship)
	slidePart.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout",
		"../slideLayouts/slideLayout1.xml",
		false,
	)

	// 4.5 Verify relationship routing.
	// Resolve the slide via its relationship from the presentation.
	slideViaRel := pkg.ResolveRelationship(presPart,
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide")
	if slideViaRel == nil {
		t.Fatal("failed to resolve slide via relationship")
	}
	if slideViaRel.PartURI().URI() != "/ppt/slides/slide1.xml" {
		t.Errorf("resolved URI is incorrect: %s", slideViaRel.PartURI().URI())
	}

	// 4.6 Save to file.
	outputPath := "test_opc_output.pptx"
	if err := pkg.SaveFile(outputPath); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}
	defer os.Remove(outputPath)

	// 4.7 Verify the output ZIP structure.
	file, _ := os.Open(outputPath)
	defer file.Close()
	stat, _ := file.Stat()
	zipReader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		t.Fatalf("ZIP read failed: %v", err)
	}

	outputFiles := make(map[string]bool)
	for _, f := range zipReader.File {
		outputFiles[f.Name] = true
		// Verify no leading "/" in paths.
		if strings.HasPrefix(f.Name, "/") {
			t.Errorf("ZIP path violation: %s", f.Name)
		}
	}

	// Verify all required OPC files are present.
	requiredFiles := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"ppt/presentation.xml",
		"ppt/_rels/presentation.xml.rels",
		"ppt/slides/slide1.xml",
		"ppt/slides/_rels/slide1.xml.rels",
	}
	for _, name := range requiredFiles {
		if !outputFiles[name] {
			t.Errorf("output is missing: %s", name)
		}
	}

	t.Logf("Stage 4: OPC write succeeded, output size = %d bytes", stat.Size())
	t.Log("Stage 4: OPC writing and routing passed")
}

// ============================================================================
// Stage 5: Parts layer — data update
// ============================================================================

// TestParts_Update proves that the Parts layer can update existing data.
func TestParts_Update(t *testing.T) {
	requireTestPPTX(t)
	// 5.1 Open OPC package.
	pkg, err := opc.OpenFile(testPPTXPath)
	if err != nil {
		t.Fatalf("opc.OpenFile failed: %v", err)
	}
	defer pkg.Close()

	// 5.2 Get and deserialize SlidePart.
	slidePart := pkg.GetPart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	if slidePart == nil {
		t.Fatal("missing slide1.xml")
	}

	slidePart5 := parts.NewSlidePart(1)
	if err := slidePart5.FromXML(slidePart.Blob()); err != nil {
		t.Fatalf("SlidePart.FromXML failed: %v", err)
	}

	originalShapeCount := slidePart5.ShapeIDCount()
	t.Logf("Stage 5: original shape count = %d", originalShapeCount)

	// 5.3 Add a new text box.
	slideBuilder5 := pptx.NewSlideBuilder(slidePart5)
	slideBuilder5.AddTextBox(1000000, 1000000, 2000000, 500000, "Updated Text!")

	newShapeCount := slidePart5.ShapeIDCount()
	if newShapeCount <= originalShapeCount {
		t.Error("shape count did not increase after adding a shape")
	}
	t.Logf("Stage 5: updated shape count = %d", newShapeCount)

	// 5.4 Re-serialize.
	newXML, err := slidePart5.ToXML()
	if err != nil {
		t.Fatalf("SlidePart.ToXML failed: %v", err)
	}
	if !strings.Contains(string(newXML), "Updated Text!") {
		t.Error("updated XML does not contain the new text")
	}

	// 5.5 Update the part in the OPC package.
	newSlidePart := opc.NewPart(
		opc.NewPackURI("/ppt/slides/slide1.xml"),
		slidePart.ContentType(),
		newXML,
	)
	if err := pkg.AddPart(newSlidePart); err != nil {
		// Part already exists; remove it first, then re-add.
		pkg.RemovePart(opc.NewPackURI("/ppt/slides/slide1.xml"))
		pkg.AddPart(newSlidePart)
	}

	// 5.6 Verify the update.
	updatedPart := pkg.GetPart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	if string(updatedPart.Blob()) != string(newXML) {
		t.Error("part update failed")
	}

	t.Log("Stage 5: Parts layer data update passed")
}

// ============================================================================
// Stage 6: OPC layer — safe repackaging
// ============================================================================

// TestOPC_SecurePackaging proves that the OPC layer can safely repackage a
// modified presentation.
func TestOPC_SecurePackaging(t *testing.T) {
	requireTestPPTX(t)
	// 6.1 Open the original file.
	originalPkg, err := opc.OpenFile(testPPTXPath)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}

	// 6.2 Record original part count and content.
	originalPartCount := originalPkg.PartCount()
	originalSlidePart := originalPkg.GetPart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	_ = string(originalSlidePart.Blob()) // capture for later verification

	// 6.3 Replace SlidePart content.
	slide6 := parts.NewSlidePart(1)
	slideBuilder6 := pptx.NewSlideBuilder(slide6)
	slideBuilder6.AddTextBox(1000000, 1000000, 2000000, 500000, "Modified Content!")
	newSlideXML, _ := slide6.ToXML()

	newSlidePart := opc.NewPart(
		opc.NewPackURI("/ppt/slides/slide1.xml"),
		originalSlidePart.ContentType(),
		newSlideXML,
	)
	originalPkg.RemovePart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	originalPkg.AddPart(newSlidePart)

	// 6.4 Save the modified package.
	outputPath := "test_secure_output.pptx"
	if err := originalPkg.SaveFile(outputPath); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}
	originalPkg.Close()
	defer os.Remove(outputPath)

	// 6.5 Reopen and verify.
	reopenedPkg, err := opc.OpenFile(outputPath)
	if err != nil {
		t.Fatalf("failed to reopen output: %v", err)
	}
	defer reopenedPkg.Close()

	// 6.6 Verify data integrity.
	reopenedPartCount := reopenedPkg.PartCount()
	if reopenedPartCount != originalPartCount {
		t.Errorf("part count mismatch: got %d, want %d", reopenedPartCount, originalPartCount)
	}

	reopenedSlidePart := reopenedPkg.GetPart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	if string(reopenedSlidePart.Blob()) != string(newSlideXML) {
		t.Error("content mismatch after reopening")
	}

	// 6.7 Verify ContentTypes are present.
	ct := reopenedPkg.ContentTypes()
	if ct == nil {
		t.Error("ContentTypes is nil")
	}

	// 6.8 Verify ZIP spec compliance (no leading slashes).
	file, _ := os.Open(outputPath)
	defer file.Close()
	stat, _ := file.Stat()
	zipReader, _ := zip.NewReader(file, stat.Size())
	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "/") {
			t.Errorf("ZIP path violation: %s", f.Name)
		}
	}

	t.Logf("Stage 6: safe repackaging succeeded, file size = %d bytes", stat.Size())
	t.Log("Stage 6: OPC safe repackaging passed")
}

// ============================================================================
// Full pipeline integration test
// ============================================================================

// TestPipeline_FullIntegration runs the complete pipeline:
// parse -> deserialize -> modify -> repackage.
func TestPipeline_FullIntegration(t *testing.T) {
	requireTestPPTX(t)
	t.Log("========== starting full pipeline test ==========")

	// Stage 1: parse
	t.Log("----- Stage 1: OPC parse and unpack -----")
	pkg, err := opc.OpenFile(testPPTXPath)
	if err != nil {
		t.Fatalf("Stage 1 failed: %v", err)
	}
	t.Logf("Stage 1 passed: unpacked %d parts", pkg.PartCount())

	// Stage 2: deserialize
	t.Log("----- Stage 2: Parts deserialization -----")
	presPart := pkg.GetPart(opc.NewPackURI("/ppt/presentation.xml"))
	pres := parts.NewPresentationPart()
	if err := pres.FromXML(presPart.Blob()); err != nil {
		t.Fatalf("Stage 2 failed: %v", err)
	}
	t.Logf("Stage 2 passed: deserialized %d slides", pres.SlideCount())

	// Stage 3: create
	t.Log("----- Stage 3: Parts creation and serialization -----")
	newSlidePart3 := parts.NewSlidePart(1)
	newSlidePart3Builder := pptx.NewSlideBuilder(newSlidePart3)
	newSlidePart3Builder.AddTextBox(914400, 457200, 4572000, 457200, "Integration Test!")
	newSlideXML, err := newSlidePart3.ToXML()
	if err != nil {
		t.Fatalf("Stage 3 serialization failed: %v", err)
	}
	t.Logf("Stage 3 passed: created new slide XML (%d bytes)", len(newSlideXML))

	// Stage 4: write
	t.Log("----- Stage 4: OPC write and route -----")
	newPkg := opc.NewPackage()
	newPkg.CreatePart(
		opc.NewPackURI("/ppt/presentation.xml"),
		"application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml",
		presPart.Blob(),
	)
	newPkg.CreatePart(
		opc.NewPackURI("/ppt/slides/slide1.xml"),
		"application/vnd.openxmlformats-officedocument.presentationml.slide+xml",
		newSlideXML,
	)
	newPkg.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument",
		"/ppt/presentation.xml",
		false,
	)
	outputPath := "test_integration_output.pptx"
	if err := newPkg.SaveFile(outputPath); err != nil {
		t.Fatalf("Stage 4 failed: %v", err)
	}
	pkg.Close()
	defer os.Remove(outputPath)
	t.Logf("Stage 4 passed: wrote %d bytes", func() int {
		info, _ := os.Stat(outputPath)
		return int(info.Size())
	}())

	// Stage 5: update
	t.Log("----- Stage 5: Parts data update -----")
	updatedPkg, _ := opc.OpenFile(outputPath)
	updatedSlide := parts.NewSlidePart(1)
	updatedSlideBuilder := pptx.NewSlideBuilder(updatedSlide)
	updatedSlideBuilder.AddTextBox(500000, 500000, 3000000, 300000, "Updated via Stage 5!")
	updatedXML, _ := updatedSlide.ToXML()
	updatedPkg.RemovePart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	updatedPkg.CreatePart(
		opc.NewPackURI("/ppt/slides/slide1.xml"),
		"application/vnd.openxmlformats-officedocument.presentationml.slide+xml",
		updatedXML,
	)
	updatedPath := "test_updated_output.pptx"
	updatedPkg.SaveFile(updatedPath)
	updatedPkg.Close()
	defer os.Remove(updatedPath)

	// Verify update.
	verifyPkg, _ := opc.OpenFile(updatedPath)
	verifySlidePart := verifyPkg.GetPart(opc.NewPackURI("/ppt/slides/slide1.xml"))
	if !strings.Contains(string(verifySlidePart.Blob()), "Updated via Stage 5!") {
		t.Error("Stage 5 verification failed: updated content not found")
	}
	verifyPkg.Close()
	t.Log("Stage 5 passed: data update successful")

	// Stage 6: safe repackage
	t.Log("----- Stage 6: OPC safe repackaging -----")
	finalPkg, _ := opc.OpenFile(updatedPath)
	finalPath := "test_final_output.pptx"
	finalPkg.SaveFile(finalPath)
	finalPkg.Close()
	defer os.Remove(finalPath)

	// Verify final file.
	finalVerifyPkg, _ := opc.OpenFile(finalPath)
	finalPartCount := finalPkg.PartCount()
	finalVerifyPkg.Close()
	t.Logf("Stage 6 passed: final file contains %d parts", finalPartCount)

	t.Log("========== full pipeline test passed ==========")

	// Clean up any remaining temp files.
	cleanupFiles := []string{
		"test_opc_output.pptx",
		"test_secure_output.pptx",
		"test_integration_output.pptx",
		"test_updated_output.pptx",
		"test_final_output.pptx",
	}
	for _, f := range cleanupFiles {
		os.Remove(f)
	}
}
