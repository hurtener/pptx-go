package parts_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/chart"
	"github.com/hurtener/pptx-go/internal/opc"
)

// routeCChartXML is a minimal chart XML with no Excel dependency (Route C).
// It contains only strCache and numCache — no <c:externalData> tag.
const routeCChartXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<c:chartSpace xmlns:c="http://schemas.openxmlformats.org/drawingml/2006/chart" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <c:chart>
    <c:plotArea>
      <c:barChart>
        <c:ser>
          <c:cat>
            <c:strRef>
              <c:strCache>
                <c:ptCount val="2"/>
                <c:pt idx="0"><c:v>Q1</c:v></c:pt>
                <c:pt idx="1"><c:v>Q2</c:v></c:pt>
              </c:strCache>
            </c:strRef>
          </c:cat>
          <c:val>
            <c:numRef>
              <c:numCache>
                <c:ptCount val="2"/>
                <c:pt idx="0"><c:v>150</c:v></c:pt>
                <c:pt idx="1"><c:v>200</c:v></c:pt>
              </c:numCache>
            </c:numRef>
          </c:val>
        </c:ser>
      </c:barChart>
    </c:plotArea>
  </c:chart>
</c:chartSpace>`

// ============================================================================
// Test 1: ChartPart raw-XML carrier test
// Goal: prove that ChartPart can preserve template-generated XML verbatim.
// ============================================================================
func TestChartPart_RawXMLCarrier(t *testing.T) {
	// Load raw XML via SetRawXML.
	chartPart := chart.NewChartPart(1)
	chartPart.SetRawXML([]byte(routeCChartXML))

	// Simulate serialization.
	outputBytes, err := chartPart.ToXML()
	if err != nil {
		t.Fatalf("chart serialization failed: %v", err)
	}
	outputXML := string(outputBytes)

	// Core check 1: cached data must be present.
	if !strings.Contains(outputXML, "<c:strCache>") || !strings.Contains(outputXML, "Q1") {
		t.Error("string cache data (strCache) was lost")
	}
	if !strings.Contains(outputXML, "<c:numCache>") || !strings.Contains(outputXML, "200") {
		t.Error("numeric cache data (numCache) was lost")
	}

	// Core check 2: Route C charts must never contain externalData.
	if strings.Contains(outputXML, "<c:externalData") {
		t.Fatal("critical: Route C chart contains an external Excel reference tag")
	}

	t.Log("ChartPart raw-data carrier test passed")
}

// ============================================================================
// Test 2: OPC relationship purity test (no Excel attachment)
// Goal: prove that no relationship pointing to an embedding folder is generated.
// ============================================================================
func TestChartPart_NoExcelRelationship(t *testing.T) {
	pkg := opc.NewPackage()

	// 1. Create and write the chart part.
	chartURI := opc.NewPackURI("/ppt/charts/chart1.xml")
	chartPartOp, err := pkg.CreatePart(chartURI, "application/vnd.openxmlformats-officedocument.drawingml.chart+xml", []byte(routeCChartXML))
	if err != nil {
		t.Fatalf("creating Chart Part failed: %v", err)
	}

	// 2. Inspect the relationships on this chartPartOp.
	// In Route C the chart itself must have no relationships (no Excel dependency).
	rels := chartPartOp.Relationships()
	if rels != nil && rels.Count() > 0 {
		// Scan for any Excel-type relationship.
		for _, rel := range rels.All() {
			if strings.Contains(rel.Type(), "officeDocument/2006/relationships/package") {
				t.Fatalf("architecture violation: chart has an Excel relationship -> %s", rel.TargetURI())
			}
		}
	}

	t.Log("Chart OPC relationship purity test passed (no Excel dependency)")
}

// ============================================================================
// Test 3: Placeholder replacement test
// Goal: prove that the template-placeholder strategy injects data correctly.
// ============================================================================
func TestChartPart_PlaceholderReplacement(t *testing.T) {
	chartPart := chart.NewChartPartWithType(1, chart.ChartTypeBar)

	// Replace placeholders.
	chartPart.ReplacePlaceholder("CHART_TITLE", "Sales Report")
	chartPart.ReplacePlaceholder("SERIES_NAME", "FY2024")
	chartPart.ReplacePlaceholder("CAT_COUNT", "3")
	chartPart.ReplacePlaceholder("CAT_COUNT_PLUS_1", "4")

	// Serialize.
	outputBytes, err := chartPart.ToXML()
	if err != nil {
		t.Fatalf("chart serialization failed: %v", err)
	}
	outputXML := string(outputBytes)

	// Verify replacements.
	if !strings.Contains(outputXML, "Sales Report") {
		t.Error("CHART_TITLE placeholder replacement failed")
	}
	if !strings.Contains(outputXML, "FY2024") {
		t.Error("SERIES_NAME placeholder replacement failed")
	}
	if !strings.Contains(outputXML, `<c:ptCount val="3"/>`) {
		t.Error("CAT_COUNT placeholder replacement failed")
	}

	t.Log("ChartPart placeholder replacement test passed")
}

// ============================================================================
// Test 4: External data reference test
// Goal: prove that an external Excel data reference can be set correctly.
// ============================================================================
func TestChartPart_ExternalDataReference(t *testing.T) {
	chartPart := chart.NewChartPart(1)

	// A newly created chart must not have external data.
	if chartPart.HasExternalData() {
		t.Error("newly created chart should not have external data")
	}

	// Set external data reference.
	chartPart.SetExternalDataRID("rId1")

	// Verify.
	if !chartPart.HasExternalData() {
		t.Error("external data should be present after setting it")
	}
	if chartPart.GetExternalDataRID() != "rId1" {
		t.Errorf("expected rId1, got %s", chartPart.GetExternalDataRID())
	}

	// Serialized output must contain the externalData tag.
	outputBytes, err := chartPart.ToXML()
	if err != nil {
		t.Fatalf("chart serialization failed: %v", err)
	}
	outputXML := string(outputBytes)

	if !strings.Contains(outputXML, `<c:externalData`) {
		t.Error("serialized output should contain an externalData tag")
	}
	if !strings.Contains(outputXML, `r:id="rId1"`) {
		t.Error("externalData tag should contain the correct r:id")
	}

	t.Log("ChartPart external data reference test passed")
}
