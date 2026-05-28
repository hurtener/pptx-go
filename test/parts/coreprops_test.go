package parts_test

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/core"
)

// ============================================================================
// Core Properties parse tests
// ============================================================================

func TestParseCoreProps(t *testing.T) {
	tests := []struct {
		name          string
		xmlData       string
		wantTitle     string
		wantCreator   string
		wantSubject   string
		wantCreated   string
		wantModified  string
		wantKeywords  string
		wantLastModBy string
		wantRevision  string
		wantCategory  string
		wantError     bool
	}{
		{
			name: "happy-path-full-core-props",
			xmlData: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Presentation Title</dc:title>
  <dc:creator>John Doe</dc:creator>
  <dc:subject>Test Subject</dc:subject>
  <dc:description>This is a test document</dc:description>
  <cp:keywords>test,example,PPT</cp:keywords>
  <cp:lastModifiedBy>Jane Doe</cp:lastModifiedBy>
  <cp:revision>3</cp:revision>
  <cp:category>Presentations</cp:category>
  <dcterms:created xsi:type="dcterms:W3CDTF">2024-01-15T10:30:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2024-01-16T14:45:00Z</dcterms:modified>
</cp:coreProperties>`,
			wantTitle:     "Presentation Title",
			wantCreator:   "John Doe",
			wantSubject:   "Test Subject",
			wantKeywords:  "test,example,PPT",
			wantLastModBy: "Jane Doe",
			wantRevision:  "3",
			wantCategory:  "Presentations",
			wantCreated:   "2024-01-15T10:30:00Z",
			wantModified:  "2024-01-16T14:45:00Z",
		},
		{
			name: "happy-path-basic-fields-only",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <dc:title>Simple Title</dc:title>
  <dc:creator>Author Name</dc:creator>
</cp:coreProperties>`,
			wantTitle:   "Simple Title",
			wantCreator: "Author Name",
		},
		{
			name: "happy-path-with-revision-info",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>Revision Test</dc:title>
  <dc:creator>Dev Team</dc:creator>
  <cp:revision>15</cp:revision>
  <dcterms:created xsi:type="dcterms:W3CDTF">2024-03-20T08:00:00Z</dcterms:created>
</cp:coreProperties>`,
			wantTitle:    "Revision Test",
			wantCreator:  "Dev Team",
			wantRevision: "15",
			wantCreated:  "2024-03-20T08:00:00Z",
		},
		{
			name: "edge-empty-core-props-element",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"/>`,
		},
		{
			name: "edge-namespace-declarations-only",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/">
</cp:coreProperties>`,
		},
		{
			name: "edge-timestamps-only",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dcterms:created xsi:type="dcterms:W3CDTF">2024-01-01T00:00:00Z</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">2024-12-31T23:59:59Z</dcterms:modified>
</cp:coreProperties>`,
			wantCreated:  "2024-01-01T00:00:00Z",
			wantModified: "2024-12-31T23:59:59Z",
		},
		{
			name:      "error-invalid-xml",
			xmlData:   `<invalid><unclosed>`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp, err := core.ParseCoreProps([]byte(tt.xmlData))

			if tt.wantError {
				if err == nil {
					t.Error("expected an error but parsing succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			if cp == nil {
				t.Fatal("ParseCoreProps returned nil")
			}

			// Verify base fields.
			if cp.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", cp.Title, tt.wantTitle)
			}
			if cp.Creator != tt.wantCreator {
				t.Errorf("Creator = %q, want %q", cp.Creator, tt.wantCreator)
			}
			if cp.Subject != tt.wantSubject {
				t.Errorf("Subject = %q, want %q", cp.Subject, tt.wantSubject)
			}
			if cp.Keywords != tt.wantKeywords {
				t.Errorf("Keywords = %q, want %q", cp.Keywords, tt.wantKeywords)
			}
			if cp.LastModifiedBy != tt.wantLastModBy {
				t.Errorf("LastModifiedBy = %q, want %q", cp.LastModifiedBy, tt.wantLastModBy)
			}
			if cp.Revision != tt.wantRevision {
				t.Errorf("Revision = %q, want %q", cp.Revision, tt.wantRevision)
			}
			if cp.Category != tt.wantCategory {
				t.Errorf("Category = %q, want %q", cp.Category, tt.wantCategory)
			}

			// Verify timestamp fields.
			if gotCreated := cp.GetCreated(); gotCreated != tt.wantCreated {
				t.Errorf("Created = %q, want %q", gotCreated, tt.wantCreated)
			}
			if gotModified := cp.GetModified(); gotModified != tt.wantModified {
				t.Errorf("Modified = %q, want %q", gotModified, tt.wantModified)
			}
		})
	}
}

// ============================================================================
// Core Properties serialization tests
// ============================================================================

func TestCorePropsToXML(t *testing.T) {
	t.Run("serialize-full-props", func(t *testing.T) {
		cp := core.NewXMLCoreProperties()
		cp.Title = "Test Title"
		cp.Creator = "Test Author"
		cp.Subject = "Test Subject"
		cp.Keywords = "keyword1,keyword2"
		cp.LastModifiedBy = "Modifier"
		cp.Revision = "2"
		cp.SetCreated("2024-01-15T10:30:00Z")
		cp.SetModified("2024-01-16T14:45:00Z")

		data, err := cp.ToXML()
		if err != nil {
			t.Fatalf("serialization failed: %v", err)
		}

		// Verify the output can be re-parsed.
		parsed, err := core.ParseCoreProps(data)
		if err != nil {
			t.Fatalf("re-parse failed: %v", err)
		}

		if parsed.Title != cp.Title {
			t.Errorf("Title mismatch: got %q, want %q", parsed.Title, cp.Title)
		}
		if parsed.Creator != cp.Creator {
			t.Errorf("Creator mismatch: got %q, want %q", parsed.Creator, cp.Creator)
		}
		if parsed.GetCreated() != cp.GetCreated() {
			t.Errorf("Created mismatch: got %q, want %q", parsed.GetCreated(), cp.GetCreated())
		}
	})

	t.Run("serialize-empty-props", func(t *testing.T) {
		cp := core.NewXMLCoreProperties()

		data, err := cp.ToXML()
		if err != nil {
			t.Fatalf("serialization failed: %v", err)
		}

		// Verify the output can be re-parsed.
		parsed, err := core.ParseCoreProps(data)
		if err != nil {
			t.Fatalf("re-parse failed: %v", err)
		}

		if parsed.Title != "" {
			t.Errorf("empty Title should be an empty string, got %q", parsed.Title)
		}
	})
}
