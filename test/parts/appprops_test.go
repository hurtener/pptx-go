package parts_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/core"
)

// validAppXML is a minimal app.xml as produced by PowerPoint natively.
const validAppXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <TotalTime>0</TotalTime>
  <Words>150</Words>
  <Application>Microsoft Office PowerPoint</Application>
  <PresentationFormat>Widescreen</PresentationFormat>
  <Paragraphs>20</Paragraphs>
  <Slides>5</Slides>
  <Notes>0</Notes>
  <HiddenSlides>0</HiddenSlides>
  <MMClips>0</MMClips>
  <ScaleCrop>false</ScaleCrop>
  <HeadingPairs>
    <vt:vector size="4" baseType="variant">
      <vt:variant><vt:lpstr>Theme</vt:lpstr></vt:variant>
      <vt:variant><vt:i4>1</vt:i4></vt:variant>
      <vt:variant><vt:lpstr>Slide Titles</vt:lpstr></vt:variant>
      <vt:variant><vt:i4>5</vt:i4></vt:variant>
    </vt:vector>
  </HeadingPairs>
  <TitlesOfParts>
    <vt:vector size="6" baseType="lpstr">
      <vt:lpstr>PowerPoint Presentation</vt:lpstr>
      <vt:lpstr>Slide 1</vt:lpstr>
      <vt:lpstr>Slide 2</vt:lpstr>
      <vt:lpstr>Slide 3</vt:lpstr>
      <vt:lpstr>Slide 4</vt:lpstr>
      <vt:lpstr>Slide 5</vt:lpstr>
    </vt:vector>
  </TitlesOfParts>
  <Company>Microsoft Corporation</Company>
  <LinksUpToDate>false</LinksUpToDate>
  <SharedDoc>false</SharedDoc>
  <HyperlinksChanged>false</HyperlinksChanged>
  <AppVersion>16.0000</AppVersion>
</Properties>`

// 1. Round-trip and mutation test.
func TestAppProperties_RoundTripAndMutate(t *testing.T) {
	appProps := &core.XMLAppProps{}
	err := xml.Unmarshal([]byte(validAppXML), appProps)
	if err != nil {
		t.Fatalf("parsing valid app.xml failed: %v", err)
	}

	// Verify parsing succeeded.
	if appProps.Application != "Microsoft Office PowerPoint" {
		t.Errorf("expected Application 'Microsoft Office PowerPoint', got '%s'", appProps.Application)
	}
	if *appProps.Slides != 5 {
		t.Errorf("expected Slides 5, got %d", *appProps.Slides)
	}

	// Simulate lower-level mutation.
	appProps.Application = "Go-pptx Engine"
	appProps.Company = "My AI Company"
	*appProps.Slides = 99

	// Re-serialize.
	outputBytes, err := xml.Marshal(appProps)
	if err != nil {
		t.Fatalf("serializing app.xml failed: %v", err)
	}
	outputXML := string(outputBytes)

	// Verify the mutated values were written correctly.
	if !strings.Contains(outputXML, "<Application>Go-pptx Engine</Application>") {
		t.Error("Application field was not modified and serialized correctly")
	}
	if !strings.Contains(outputXML, "<Company>My AI Company</Company>") {
		t.Error("Company field was not modified and serialized correctly")
	}
	if !strings.Contains(outputXML, "<Slides>99</Slides>") {
		t.Error("Slides field was not modified and serialized correctly")
	}
	// Verify nested structures were preserved (HeadingPairs and vt:vector).
	if !strings.Contains(outputXML, "<HeadingPairs>") || !strings.Contains(outputXML, "vt:vector") {
		t.Error("nested elements (HeadingPairs/vt:vector) were lost during round-trip serialization")
	}

	t.Log("App properties round-trip and mutation test passed")
}

// 2. Namespace and root element test.
func TestAppProperties_Namespaces(t *testing.T) {
	appProps := &core.XMLAppProps{
		Application: "Go-pptx Engine",
		XmlnsProp:   core.NamespaceExtendedProperties,
		XmlnsVt:     core.NamespaceDocPropsVTypes,
	}

	outputBytes, err := xml.Marshal(appProps)
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}
	outputXML := string(outputBytes)

	// OOXML requires both of these namespaces.
	expectedNS1 := `xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"`
	expectedNS2 := `xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"`

	if !strings.Contains(outputXML, expectedNS1) {
		t.Errorf("missing extended-properties namespace declaration: %s", expectedNS1)
	}
	if !strings.Contains(outputXML, expectedNS2) {
		t.Errorf("missing VT namespace declaration: %s", expectedNS2)
	}

	t.Log("App namespace test passed")
}

// 3. Omitempty tag suppression test.
func TestAppProperties_Omitempty(t *testing.T) {
	// Struct with only Application set; all other fields remain zero.
	appProps := &core.XMLAppProps{
		Application: "Go-pptx Engine",
	}

	outputBytes, err := xml.Marshal(appProps)
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}
	outputXML := string(outputBytes)

	// Unset fields must not produce empty tags such as <Company></Company>.
	if strings.Contains(outputXML, "<Company>") {
		t.Error("unset Company field must not appear in XML; check struct tag for missing omitempty")
	}
	if strings.Contains(outputXML, "<Manager>") {
		t.Error("unset Manager field must not appear in XML")
	}

	t.Log("App omitempty tag suppression test passed")
}
