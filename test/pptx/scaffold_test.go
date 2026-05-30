package pptx_test

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// readZipPart returns the bytes of a single entry from a .pptx byte slice.
func readZipPart(t *testing.T, data []byte, name string) string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("open %s: %v", name, err)
			}
			defer func() { _ = rc.Close() }()
			b, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			return string(b)
		}
	}
	t.Fatalf("part %s not found in package", name)
	return ""
}

// TestScaffold_CompleteDeck proves New() seeds a complete, wired deck (Phase 03
// A2): the master, blank layout and theme parts exist with their relationship
// files, and presentation.xml references the master and each slide by a
// resolvable r:id.
func TestScaffold_CompleteDeck(t *testing.T) {
	pres := pptx.New()
	pres.AddSlide()
	pres.AddSlide()

	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}

	// Required scaffold + slide parts and their relationship files.
	for _, name := range []string{
		"ppt/presentation.xml",
		"ppt/_rels/presentation.xml.rels",
		"ppt/theme/theme1.xml",
		"ppt/slideMasters/slideMaster1.xml",
		"ppt/slideMasters/_rels/slideMaster1.xml.rels",
		"ppt/slideLayouts/slideLayout1.xml",
		"ppt/slideLayouts/_rels/slideLayout1.xml.rels",
		"ppt/slides/slide1.xml",
		"ppt/slides/_rels/slide1.xml.rels",
		"ppt/slides/slide2.xml",
		// Auxiliary parts PowerPoint expects (a deck without them repairs).
		"ppt/presProps.xml",
		"ppt/viewProps.xml",
		"ppt/tableStyles.xml",
		"docProps/core.xml",
		"docProps/app.xml",
	} {
		if readZipPart(t, data, name) == "" {
			t.Errorf("missing or empty part: %s", name)
		}
	}

	// presentation.xml wires the master and both slides; no empty r:id.
	pres1 := readZipPart(t, data, "ppt/presentation.xml")
	for _, want := range []string{
		"<p:sldMasterIdLst>",
		`<p:sldMasterId id="2147483648" r:id="rId1"/>`,
		// Each slide is wired to a non-empty r:id (exact value depends on how
		// many auxiliary relationships precede it).
		`<p:sldId id="257" r:id="rId`,
		`<p:sldId id="258" r:id="rId`,
		"<p:notesSz",
	} {
		if !strings.Contains(pres1, want) {
			t.Errorf("presentation.xml missing %q in:\n%s", want, pres1)
		}
	}
	if strings.Contains(pres1, `r:id=""`) {
		t.Errorf("presentation.xml has an unwired (empty) r:id:\n%s", pres1)
	}

	// The master references the blank layout via its own relationship.
	master := readZipPart(t, data, "ppt/slideMasters/slideMaster1.xml")
	if !strings.Contains(master, "<p:sldLayoutIdLst>") || strings.Contains(master, "%LAYOUT_RID%") {
		t.Errorf("master sldLayoutIdLst not wired:\n%s", master)
	}

	// Slides relate to the layout.
	slideRels := readZipPart(t, data, "ppt/slides/_rels/slide1.xml.rels")
	if !strings.Contains(slideRels, "slideLayout") || !strings.Contains(slideRels, "../slideLayouts/slideLayout1.xml") {
		t.Errorf("slide1 is not related to the layout:\n%s", slideRels)
	}
}
