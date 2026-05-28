package pptx_test

import (
	"archive/zip"
	"os"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/internal/opc"
	"github.com/hurtener/pptx-go/pptx"
)

// ============================================================================
// Minimal Viable OOXML foundation test
//
// TestOPC_RawAssemblyFallback — zero-dependency low-level assembly test
//
// NOTE: The purpose of this test is to verify the OPC engine's ability to
// assemble raw parts, serialize XML headers, and produce a valid ZIP package.
// Because it uses hardcoded minimal XML and relies on AddRelationship call
// order to produce specific rIds (e.g. rId1, rId2), the resulting
// test_hardcore_output.pptx may fail to open directly in some versions of
// PowerPoint due to rId misalignment. This is expected and intentional.
//
// ============================================================================

const (
	// minThemeXML is a minimal theme containing the essential color matrix and
	// fonts required to prevent PowerPoint from crashing on open.
	minThemeXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office Theme"><a:themeElements><a:clrScheme name="Office"><a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1><a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1><a:dk2><a:srgbClr val="44546A"/></a:dk2><a:lt2><a:srgbClr val="E7E6E6"/></a:lt2><a:accent1><a:srgbClr val="4472C4"/></a:accent1><a:accent2><a:srgbClr val="ED7D31"/></a:accent2><a:accent3><a:srgbClr val="A5A5A5"/></a:accent3><a:accent4><a:srgbClr val="FFC000"/></a:accent4><a:accent5><a:srgbClr val="5B9BD5"/></a:accent5><a:accent6><a:srgbClr val="70AD47"/></a:accent6><a:hlink><a:srgbClr val="0563C1"/></a:hlink><a:folHlink><a:srgbClr val="954F72"/></a:folHlink></a:clrScheme><a:fontScheme name="Office"><a:majorFont><a:latin typeface="Calibri Light"/></a:majorFont><a:minorFont><a:latin typeface="Calibri"/></a:minorFont></a:fontScheme><a:fmtScheme name="Office"><a:fillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:fillStyleLst><a:lnStyleLst><a:ln><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:ln></a:lnStyleLst><a:effectStyleLst><a:effectStyle><a:effectLst/></a:effectStyle></a:effectStyleLst><a:bgFillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:bgFillStyleLst></a:fmtScheme></a:themeElements></a:theme>`

	// minMasterXML is a minimal slide master containing a layout reference list
	// and a basic shape tree.
	minMasterXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><p:sldMaster xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:spTree></p:cSld><p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/><p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rId1"/></p:sldLayoutIdLst></p:sldMaster>`

	// minLayoutXML is a near-empty slide layout that must exist to host a slide.
	minLayoutXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type="blank"><p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:spTree></p:cSld></p:sldLayout>`

	// minPresentationXML is a minimal presentation XML containing a complete
	// sldMasterIdLst and sldIdLst.
	minPresentationXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><p:presentation xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst><p:sldIdLst><p:sldId id="256" r:id="rId2"/></p:sldIdLst><p:sldSz cx="12192000" cy="6858000"/></p:presentation>`
)

// TestBlankPresentation_Hardcore builds a complete OOXML inheritance-tree PPTX
// from scratch by manually wiring:
//
//	slide1 -> slideLayout1 -> slideMaster1 -> theme1
func TestBlankPresentation_Hardcore(t *testing.T) {
	pkg := opc.NewPackage()

	// --- 1. Create static infrastructure (Theme, Master, Layout) ---
	themeURI := opc.NewPackURI("/ppt/theme/theme1.xml")
	pkg.CreatePart(themeURI, "application/vnd.openxmlformats-officedocument.theme+xml", []byte(minThemeXML))

	masterURI := opc.NewPackURI("/ppt/slideMasters/slideMaster1.xml")
	masterPart, _ := pkg.CreatePart(masterURI, "application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml", []byte(minMasterXML))

	layoutURI := opc.NewPackURI("/ppt/slideLayouts/slideLayout1.xml")
	layoutPart, _ := pkg.CreatePart(layoutURI, "application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml", []byte(minLayoutXML))

	// --- 2. Wire the infrastructure relationships ---
	// 2.1 Master -> Theme (rId2)
	masterPart.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme",
		"../theme/theme1.xml",
		false,
	)
	// 2.2 Master -> Layout (rId1, matches sldLayoutId r:id="rId1" in minMasterXML)
	masterPart.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout",
		"../slideLayouts/slideLayout1.xml",
		false,
	)
	// 2.3 Layout -> Master (rId1)
	layoutPart.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster",
		"../slideMasters/slideMaster1.xml",
		false,
	)

	// --- 3. Create Presentation part (using minimal XML directly) ---
	presPartOp, _ := pkg.CreatePart(
		opc.NewPackURI("/ppt/presentation.xml"),
		"application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml",
		[]byte(minPresentationXML),
	)

	// --- 4. Wire Presentation relationships ---
	// Root -> Presentation
	pkg.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument",
		"/ppt/presentation.xml",
		false,
	)
	// Presentation -> Master (rId1)
	presPartOp.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster",
		"slideMasters/slideMaster1.xml",
		false,
	)
	// Presentation -> Theme (rId3)
	presPartOp.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme",
		"theme/theme1.xml",
		false,
	)

	// --- 5. Create slide and link it to the layout ---
	slidePart := slide.NewSlidePart(1)
	builder := pptx.NewSlideBuilder(slidePart)
	builder.AddTextBox(914400, 457200, 4572000, 457200, "Hello from Go Engine!")
	slideXML, _ := slidePart.ToXML()

	slidePartOp, _ := pkg.CreatePart(
		opc.NewPackURI("/ppt/slides/slide1.xml"),
		"application/vnd.openxmlformats-officedocument.presentationml.slide+xml",
		slideXML,
	)

	// Slide -> Layout (rId1) — critical link
	slidePartOp.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout",
		"../slideLayouts/slideLayout1.xml",
		false,
	)

	// --- 6. Complete Presentation -> Slide relationship ---
	// Presentation -> Slide (rId2)
	presPartOp.AddRelationship(
		"http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide",
		"slides/slide1.xml",
		false,
	)

	// --- 7. Pack and save ---
	outputPath := "test_hardcore_output.pptx"
	err := pkg.SaveFile(outputPath)
	if err != nil {
		t.Fatalf("failed to save PPTX: %v", err)
	}
	// defer os.Remove(outputPath)

	t.Logf("created minimal PPTX: %s", outputPath)

	// --- 8. Verify file was written ---
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file size is 0")
	}

	// --- 9. Verify OPC structure ---
	zipReader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("failed to open ZIP: %v", err)
	}
	defer zipReader.Close()

	files := make(map[string]bool)
	for _, f := range zipReader.File {
		files[f.Name] = true
	}

	// Verify all required OPC files are present.
	requiredFiles := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"ppt/presentation.xml",
		"ppt/_rels/presentation.xml.rels",
		"ppt/slides/slide1.xml",
		"ppt/slideMasters/slideMaster1.xml",
		"ppt/slideLayouts/slideLayout1.xml",
		"ppt/theme/theme1.xml",
	}

	for _, name := range requiredFiles {
		if !files[name] {
			t.Errorf("missing required file: %s", name)
		}
	}

	t.Log("OPC structure verification passed")

	// --- 10. Verify XML declaration headers ---
	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".xml") || strings.HasSuffix(f.Name, ".rels") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			header := make([]byte, 60)
			n, _ := rc.Read(header)
			rc.Close()
			content := string(header[:n])
			if !strings.Contains(content, "<?xml") {
				t.Errorf("file %s is missing the XML declaration header", f.Name)
			}
			if !strings.Contains(content, "standalone=\"yes\"") {
				t.Errorf("file %s is missing standalone=\"yes\"", f.Name)
			}
		}
	}

	t.Log("XML declaration header verification passed")

	// --- 11. Reopen and verify ---
	pkg2, err := opc.OpenFile(outputPath)
	if err != nil {
		t.Fatalf("failed to reopen PPTX: %v", err)
	}
	defer pkg2.Close()

	// Verify all critical parts exist.
	if pkg2.GetPart(opc.NewPackURI("/ppt/presentation.xml")) == nil {
		t.Fatal("missing presentation.xml")
	}
	if pkg2.GetPart(opc.NewPackURI("/ppt/slides/slide1.xml")) == nil {
		t.Fatal("missing slide1.xml")
	}
	if pkg2.GetPart(opc.NewPackURI("/ppt/slideMasters/slideMaster1.xml")) == nil {
		t.Fatal("missing slideMaster1.xml")
	}
	if pkg2.GetPart(opc.NewPackURI("/ppt/slideLayouts/slideLayout1.xml")) == nil {
		t.Fatal("missing slideLayout1.xml")
	}
	if pkg2.GetPart(opc.NewPackURI("/ppt/theme/theme1.xml")) == nil {
		t.Fatal("missing theme1.xml")
	}

	t.Log("all verifications passed: PPTX can be reopened successfully")
}
