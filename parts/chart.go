package parts

// ============================================================================
// ChartPart - chart part
// ============================================================================
//
// Corresponds to /ppt/charts/chartN.xml
//
// Design strategy: string template + placeholders
// - Chart XML is extremely complex (hundreds of element combinations) and
//   does not map cleanly to a Go struct hierarchy.
// - Predefined templates with placeholder tokens are used; callers inject
//   data by replacing those tokens.
// - Common chart types each have a predefined template constant.
//
// ============================================================================

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hurtener/pptx-go/opc"
)

// ChartPart holds a chart part for a presentation slide.
type ChartPart struct {
	uri         *opc.PackURI
	template    string // XML template with placeholder tokens
	externalRID string // external data relationship ID (e.g. embedded Excel)
	mu          sync.RWMutex
}

// NewChartPart creates a new chart part with the given numeric ID.
func NewChartPart(id int) *ChartPart {
	return &ChartPart{
		uri:      opc.NewPackURI(fmt.Sprintf("/ppt/charts/chart%d.xml", id)),
		template: ChartTemplateBar, // default to bar chart
	}
}

// NewChartPartWithURI creates a chart part using the specified URI.
func NewChartPartWithURI(uri *opc.PackURI) *ChartPart {
	return &ChartPart{
		uri:      uri,
		template: ChartTemplateBar,
	}
}

// NewChartPartWithType creates a chart part of the specified type.
func NewChartPartWithType(id int, chartType ChartType) *ChartPart {
	return &ChartPart{
		uri:      opc.NewPackURI(fmt.Sprintf("/ppt/charts/chart%d.xml", id)),
		template: GetChartTemplate(chartType),
	}
}

// PartURI returns the part URI.
func (c *ChartPart) PartURI() *opc.PackURI {
	return c.uri
}

// ============================================================================
// Template methods
// ============================================================================

// SetTemplate sets the chart XML template.
func (c *ChartPart) SetTemplate(tmpl string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.template = tmpl
}

// SetRawXML sets the chart content from raw XML bytes (equivalent to SetTemplate).
func (c *ChartPart) SetRawXML(xml []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.template = string(xml)
}

// Template returns the current XML template.
func (c *ChartPart) Template() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.template
}

// ReplacePlaceholder replaces a single placeholder token in the template.
// placeholder is the token name without the surrounding {{ }}.
func (c *ChartPart) ReplacePlaceholder(placeholder, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.template = strings.ReplaceAll(c.template, "{{"+placeholder+"}}", value)
}

// ReplacePlaceholders replaces multiple placeholder tokens in one call.
func (c *ChartPart) ReplacePlaceholders(replacements map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range replacements {
		c.template = strings.ReplaceAll(c.template, "{{"+k+"}}", v)
	}
}

// ============================================================================
// External data reference
// ============================================================================

// SetExternalDataRID sets the external data relationship ID.
func (c *ChartPart) SetExternalDataRID(rid string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.externalRID = rid
}

// GetExternalDataRID returns the external data relationship ID.
func (c *ChartPart) GetExternalDataRID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.externalRID
}

// HasExternalData reports whether an external data reference is set.
func (c *ChartPart) HasExternalData() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.externalRID != ""
}

// ============================================================================
// XML serialization
// ============================================================================

// ToXML serializes the chart to XML, injecting the external data tag when set.
func (c *ChartPart) ToXML() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// inject externalData element before </c:chartSpace> when a reference is set
	xml := c.template
	if c.externalRID != "" {
		externalDataTag := fmt.Sprintf(`<c:externalData xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" r:id="%s"/>`, c.externalRID)
		xml = strings.Replace(xml, "</c:chartSpace>", externalDataTag+"</c:chartSpace>", 1)
	}

	return []byte(xml), nil
}

// FromXML loads chart XML, storing it directly as the template.
func (c *ChartPart) FromXML(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.template = string(data)
	return nil
}

// ============================================================================
// Chart type enumeration
// ============================================================================

// ChartType identifies the chart kind.
type ChartType int

const (
	ChartTypeBar      ChartType = iota // bar chart
	ChartTypePie                       // pie chart
	ChartTypeLine                      // line chart
	ChartTypeArea                      // area chart
	ChartTypeScatter                   // scatter chart
	ChartTypeDoughnut                  // doughnut chart
)

// GetChartTemplate returns the XML template for the given chart type.
func GetChartTemplate(chartType ChartType) string {
	switch chartType {
	case ChartTypeBar:
		return ChartTemplateBar
	case ChartTypePie:
		return ChartTemplatePie
	case ChartTypeLine:
		return ChartTemplateLine
	case ChartTypeArea:
		return ChartTemplateArea
	case ChartTypeScatter:
		return ChartTemplateScatter
	case ChartTypeDoughnut:
		return ChartTemplateDoughnut
	default:
		return ChartTemplateBar
	}
}

// ============================================================================
// Data structures (helpers for template population)
// ============================================================================

// ChartSeriesData holds data for a single chart series.
type ChartSeriesData struct {
	Name   string   // series name
	Values []string // data values
}

// ChartCategoryData holds categories and series for a chart.
type ChartCategoryData struct {
	Categories []string          // category labels
	Series     []ChartSeriesData // series data
}
