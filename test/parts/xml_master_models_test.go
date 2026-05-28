package parts_test

import (
	"encoding/xml"
	"testing"

	"github.com/hurtener/pptx-go/parts"
)

// ============================================================================
// Base coordinate and size tests — XMLOffset and XMLExtents
// ============================================================================

func TestParseXMLOffset(t *testing.T) {
	tests := []struct {
		name      string
		xmlData   string
		wantX     int64
		wantY     int64
		wantError bool
	}{
		{
			name:    "happy-path-positive-coords",
			xmlData: `<a:off x="1524000" y="1143000"/>`,
			wantX:   1524000,
			wantY:   1143000,
		},
		{
			name:    "happy-path-zero-coords",
			xmlData: `<a:off x="0" y="0"/>`,
			wantX:   0,
			wantY:   0,
		},
		{
			name:    "happy-path-large-values",
			xmlData: `<a:off x="9144000" y="6858000"/>`,
			wantX:   9144000,
			wantY:   6858000,
		},
		{
			name:    "edge-missing-x-defaults-to-zero",
			xmlData: `<a:off y="1143000"/>`,
			wantX:   0,
			wantY:   1143000,
		},
		{
			name:    "edge-missing-y-defaults-to-zero",
			xmlData: `<a:off x="1524000"/>`,
			wantX:   1524000,
			wantY:   0,
		},
		{
			name:    "edge-empty-element-defaults-to-zero",
			xmlData: `<a:off/>`,
			wantX:   0,
			wantY:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var off parts.XMLOffset
			err := xml.Unmarshal([]byte(tt.xmlData), &off)

			if tt.wantError {
				if err == nil {
					t.Error("expected an error but parsing succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			if off.X != tt.wantX {
				t.Errorf("X = %d, want %d", off.X, tt.wantX)
			}
			if off.Y != tt.wantY {
				t.Errorf("Y = %d, want %d", off.Y, tt.wantY)
			}
		})
	}
}

func TestParseXMLExtents(t *testing.T) {
	tests := []struct {
		name      string
		xmlData   string
		wantCx    int64
		wantCy    int64
		wantError bool
	}{
		{
			name:    "happy-path-standard-size",
			xmlData: `<a:ext cx="6858000" cy="5143500"/>`,
			wantCx:  6858000,
			wantCy:  5143500,
		},
		{
			name:    "happy-path-zero-size",
			xmlData: `<a:ext cx="0" cy="0"/>`,
			wantCx:  0,
			wantCy:  0,
		},
		{
			name:    "happy-path-widescreen-size",
			xmlData: `<a:ext cx="12192000" cy="6858000"/>`,
			wantCx:  12192000,
			wantCy:  6858000,
		},
		{
			name:    "edge-missing-cx-defaults-to-zero",
			xmlData: `<a:ext cy="5143500"/>`,
			wantCx:  0,
			wantCy:  5143500,
		},
		{
			name:    "edge-missing-cy-defaults-to-zero",
			xmlData: `<a:ext cx="6858000"/>`,
			wantCx:  6858000,
			wantCy:  0,
		},
		{
			name:    "edge-empty-element-defaults-to-zero",
			xmlData: `<a:ext/>`,
			wantCx:  0,
			wantCy:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ext parts.XMLExtents
			err := xml.Unmarshal([]byte(tt.xmlData), &ext)

			if tt.wantError {
				if err == nil {
					t.Error("expected an error but parsing succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			if ext.Cx != tt.wantCx {
				t.Errorf("Cx = %d, want %d", ext.Cx, tt.wantCx)
			}
			if ext.Cy != tt.wantCy {
				t.Errorf("Cy = %d, want %d", ext.Cy, tt.wantCy)
			}
		})
	}
}

// ============================================================================
// Composite transform structure test — XMLTransform
// ============================================================================

// xmlSpPrWrapper is a test wrapper for XMLTransform (via p:spPr).
type xmlSpPrWrapper struct {
	Xfrm *parts.XMLTransform `xml:"xfrm"`
}

func TestParseXMLTransform(t *testing.T) {
	tests := []struct {
		name    string
		xmlData string
		wantX   int64
		wantY   int64
		wantCx  int64
		wantCy  int64
		hasOff  bool
		hasExt  bool
	}{
		{
			name:    "happy-path-full-transform",
			xmlData: `<spPr><xfrm><off x="1524000" y="1143000"/><ext cx="6858000" cy="5143500"/></xfrm></spPr>`,
			wantX:   1524000,
			wantY:   1143000,
			wantCx:  6858000,
			wantCy:  5143500,
			hasOff:  true,
			hasExt:  true,
		},
		{
			name:    "happy-path-widescreen-size",
			xmlData: `<spPr><xfrm><off x="0" y="0"/><ext cx="12192000" cy="6858000"/></xfrm></spPr>`,
			wantX:   0,
			wantY:   0,
			wantCx:  12192000,
			wantCy:  6858000,
			hasOff:  true,
			hasExt:  true,
		},
		{
			name:    "edge-offset-only-no-extent",
			xmlData: `<spPr><xfrm><off x="9144000" y="6858000"/></xfrm></spPr>`,
			wantX:   9144000,
			wantY:   6858000,
			wantCx:  0,
			wantCy:  0,
			hasOff:  true,
			hasExt:  false,
		},
		{
			name:    "edge-extent-only-no-offset",
			xmlData: `<spPr><xfrm><ext cx="4572000" cy="3429000"/></xfrm></spPr>`,
			wantX:   0,
			wantY:   0,
			wantCx:  4572000,
			wantCy:  3429000,
			hasOff:  false,
			hasExt:  true,
		},
		{
			name:    "edge-empty-transform-element",
			xmlData: `<spPr><xfrm/></spPr>`,
			wantX:   0,
			wantY:   0,
			wantCx:  0,
			wantCy:  0,
			hasOff:  false,
			hasExt:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrapper xmlSpPrWrapper
			err := xml.Unmarshal([]byte(tt.xmlData), &wrapper)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			if wrapper.Xfrm == nil {
				t.Fatal("Xfrm is nil")
			}

			xfrm := wrapper.Xfrm

			// Check whether Off is present.
			if tt.hasOff {
				if xfrm.Off == nil {
					t.Fatal("expected Off to be non-nil")
				}
				if xfrm.Off.X != tt.wantX {
					t.Errorf("Off.X = %d, want %d", xfrm.Off.X, tt.wantX)
				}
				if xfrm.Off.Y != tt.wantY {
					t.Errorf("Off.Y = %d, want %d", xfrm.Off.Y, tt.wantY)
				}
			} else {
				if xfrm.Off != nil {
					t.Errorf("expected Off to be nil, got %+v", xfrm.Off)
				}
			}

			// Check whether Ext is present.
			if tt.hasExt {
				if xfrm.Ext == nil {
					t.Fatal("expected Ext to be non-nil")
				}
				if xfrm.Ext.Cx != tt.wantCx {
					t.Errorf("Ext.Cx = %d, want %d", xfrm.Ext.Cx, tt.wantCx)
				}
				if xfrm.Ext.Cy != tt.wantCy {
					t.Errorf("Ext.Cy = %d, want %d", xfrm.Ext.Cy, tt.wantCy)
				}
			} else {
				if xfrm.Ext != nil {
					t.Errorf("expected Ext to be nil, got %+v", xfrm.Ext)
				}
			}
		})
	}
}

// ============================================================================
// Placeholder structure test — XMLPlaceholder
// ============================================================================

func TestParseXMLPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		xmlData  string
		wantType string
		wantIdx  string
	}{
		{
			name:     "standard-title-placeholder",
			xmlData:  `<ph type="title"/>`,
			wantType: "title",
			wantIdx:  "",
		},
		{
			name:     "body-placeholder-with-index",
			xmlData:  `<ph type="body" idx="1"/>`,
			wantType: "body",
			wantIdx:  "1",
		},
		{
			name:     "date-placeholder-with-extra-sz-attr",
			xmlData:  `<ph type="dt" sz="half"/>`,
			wantType: "dt",
			wantIdx:  "",
		},
		{
			name:     "slide-number-placeholder",
			xmlData:  `<ph type="sldNum"/>`,
			wantType: "sldNum",
			wantIdx:  "",
		},
		{
			name:     "footer-placeholder",
			xmlData:  `<ph type="ftr"/>`,
			wantType: "ftr",
			wantIdx:  "",
		},
		{
			name:     "center-title-placeholder",
			xmlData:  `<ph type="ctrTitle"/>`,
			wantType: "ctrTitle",
			wantIdx:  "",
		},
		{
			name:     "subtitle-placeholder",
			xmlData:  `<ph type="subTitle"/>`,
			wantType: "subTitle",
			wantIdx:  "",
		},
		{
			name:     "chart-placeholder-with-index",
			xmlData:  `<ph type="chart" idx="2"/>`,
			wantType: "chart",
			wantIdx:  "2",
		},
		{
			name:     "table-placeholder",
			xmlData:  `<ph type="tbl"/>`,
			wantType: "tbl",
			wantIdx:  "",
		},
		{
			name:     "picture-placeholder",
			xmlData:  `<ph type="pic"/>`,
			wantType: "pic",
			wantIdx:  "",
		},
		{
			name:     "idx-only-no-type",
			xmlData:  `<ph idx="0"/>`,
			wantType: "",
			wantIdx:  "0",
		},
		{
			name:     "edge-empty-placeholder-element",
			xmlData:  `<ph/>`,
			wantType: "",
			wantIdx:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ph parts.XMLPlaceholder
			err := xml.Unmarshal([]byte(tt.xmlData), &ph)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			if ph.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", ph.Type, tt.wantType)
			}
			if ph.Idx != tt.wantIdx {
				t.Errorf("Idx = %q, want %q", ph.Idx, tt.wantIdx)
			}
		})
	}
}
