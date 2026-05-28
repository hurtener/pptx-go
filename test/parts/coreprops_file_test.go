package parts_test

import (
	"os"
	"testing"

	"github.com/hurtener/pptx-go/parts"
)

// ============================================================================
// CoreProperties real-file parse tests
// ============================================================================

func TestParseCorePropsFromFile(t *testing.T) {
	// Read the real core.xml file.
	xmlData, err := os.ReadFile("../test-data/test/docProps/core.xml")
	if err != nil {
		t.Fatalf("reading file failed: %v", err)
	}

	// Parse.
	cp, err := parts.ParseCoreProps(xmlData)
	if err != nil {
		t.Fatalf("ParseCoreProps returned an error: %v", err)
	}
	if cp == nil {
		t.Fatal("ParseCoreProps returned nil")
	}

	// Assert that core fields are non-empty.
	if cp.Title == "" {
		t.Error("Title is empty")
	}
	if cp.Creator == "" {
		t.Error("Creator is empty")
	}
	if cp.LastModifiedBy == "" {
		t.Error("LastModifiedBy is empty")
	}
	if cp.Revision == "" {
		t.Error("Revision is empty")
	}
	if cp.GetCreated() == "" {
		t.Error("Created is empty")
	}
	if cp.GetModified() == "" {
		t.Error("Modified is empty")
	}

	// Verify exact values.
	if cp.Title != "PowerPoint Presentation" {
		t.Errorf("Title = %q, want %q", cp.Title, "PowerPoint Presentation")
	}
	if cp.Creator != "YouPin PPT" {
		t.Errorf("Creator = %q, want %q", cp.Creator, "YouPin PPT")
	}
	if cp.LastModifiedBy != "kan" {
		t.Errorf("LastModifiedBy = %q, want %q", cp.LastModifiedBy, "kan")
	}
	if cp.Revision != "91" {
		t.Errorf("Revision = %q, want %q", cp.Revision, "91")
	}
	if cp.GetCreated() != "2019-05-16T00:04:14Z" {
		t.Errorf("Created = %q, want %q", cp.GetCreated(), "2019-05-16T00:04:14Z")
	}
	if cp.GetModified() != "2022-05-30T10:23:18Z" {
		t.Errorf("Modified = %q, want %q", cp.GetModified(), "2022-05-30T10:23:18Z")
	}

	t.Logf("parse succeeded: Title=%q, Creator=%q, Revision=%q", cp.Title, cp.Creator, cp.Revision)
}
