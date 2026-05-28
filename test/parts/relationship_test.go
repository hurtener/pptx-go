package parts_test

import (
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/relations"
)

// ============================================================================
// Relationships parse tests
// ============================================================================

func TestParseRelationships(t *testing.T) {
	tests := []struct {
		name           string
		xmlData        string
		wantCount      int
		wantRID1Type   string
		wantRID1Target string
		wantError      bool
	}{
		{
			name: "happy-path-image-and-video",
			xmlData: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image1.png"/>
    <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/video" Target="../media/media1.mp4"/>
</Relationships>`,
			wantCount:      2,
			wantRID1Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image",
			wantRID1Target: "../media/image1.png",
		},
		{
			name: "happy-path-external-hyperlink",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="https://example.com" TargetMode="External"/>
</Relationships>`,
			wantCount:      1,
			wantRID1Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink",
			wantRID1Target: "https://example.com",
		},
		{
			name: "edge-empty-relationships",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
</Relationships>`,
			wantCount: 0,
		},
		{
			name:      "error-invalid-xml",
			xmlData:   `<invalid><unclosed>`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs, err := relations.ParseRelationships([]byte(tt.xmlData))

			if tt.wantError {
				if err == nil {
					t.Error("expected an error but parsing succeeded")
				}
				return
			}

			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			if rs == nil {
				t.Fatal("ParseRelationships returned nil")
			}

			if rs.Count() != tt.wantCount {
				t.Errorf("Count = %d, want %d", rs.Count(), tt.wantCount)
			}

			if tt.wantCount > 0 && tt.wantRID1Type != "" {
				rel := rs.GetByID("rId1")
				if rel == nil {
					t.Fatal("GetByID(\"rId1\") returned nil")
				}
				if rel.Type != tt.wantRID1Type {
					t.Errorf("Type = %q, want %q", rel.Type, tt.wantRID1Type)
				}
				if rel.Target != tt.wantRID1Target {
					t.Errorf("Target = %q, want %q", rel.Target, tt.wantRID1Target)
				}
			}
		})
	}
}

// ============================================================================
// Relationship helper method tests
// ============================================================================

func TestXMLRelationshipsMethods(t *testing.T) {
	t.Run("GetByType", func(t *testing.T) {
		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image1.png"/>
    <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image2.png"/>
    <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/video" Target="../media/media1.mp4"/>
</Relationships>`

		rs, err := relations.ParseRelationships([]byte(xmlData))
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}

		images := rs.GetByType(relations.RelTypeImage)
		if len(images) != 2 {
			t.Errorf("GetByType(image) returned %d, want 2", len(images))
		}

		videos := rs.GetByType(relations.RelTypeMedia)
		if len(videos) != 1 {
			t.Errorf("GetByType(video) returned %d, want 1", len(videos))
		}
	})

	t.Run("GetByTarget", func(t *testing.T) {
		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image1.png"/>
</Relationships>`

		rs, err := relations.ParseRelationships([]byte(xmlData))
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}

		rel := rs.GetByTarget("../media/image1.png")
		if rel == nil {
			t.Fatal("GetByTarget returned nil")
		}
		if rel.ID != "rId1" {
			t.Errorf("ID = %q, want %q", rel.ID, "rId1")
		}
	})

	t.Run("IsExternal", func(t *testing.T) {
		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="https://example.com" TargetMode="External"/>
    <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image1.png"/>
</Relationships>`

		rs, err := relations.ParseRelationships([]byte(xmlData))
		if err != nil {
			t.Fatalf("parse failed: %v", err)
		}

		external := rs.GetByID("rId1")
		if !external.IsExternal() {
			t.Error("rId1 should be an external link")
		}

		internal := rs.GetByID("rId2")
		if internal.IsExternal() {
			t.Error("rId2 should be an internal link")
		}
	})
}
