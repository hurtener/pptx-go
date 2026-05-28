package parts_test

import (
	"os"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
)

// ============================================================================
// Master parse tests
// ============================================================================

func TestParseMasterFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		// slideMaster1 has only a background image, no placeholders.
		{"slideMaster1", "../test-data/test/ppt/slideMasters/slideMaster1.xml"},
		// slideMaster2 has title, body, date, footer, and slide-number placeholders.
		{"slideMaster2", "../test-data/test/ppt/slideMasters/slideMaster2.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData, err := os.ReadFile(tt.filePath)
			if err != nil {
				if os.IsNotExist(err) {
					t.Skipf("fixture not present; skipping (gitignored, not committed upstream): %v", err)
				}
				t.Fatalf("reading file failed: %v", err)
			}

			master, err := slide.ParseMaster(xmlData)
			if err != nil {
				t.Fatalf("ParseMaster returned an error: %v", err)
			}
			if master == nil {
				t.Fatal("ParseMaster returned nil")
			}

			// Confirm that master data was parsed.
			t.Logf("%s: placeholder count = %d", tt.name, len(master.Placeholders()))
		})
	}
}

// ============================================================================
// Master background and placeholder assertions
// ============================================================================

func TestParseMaster_BackgroundAndPlaceholders(t *testing.T) {
	xmlData, err := os.ReadFile("../test-data/test/ppt/slideMasters/slideMaster2.xml")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("fixture not present; skipping (gitignored, not committed upstream): %v", err)
		}
		t.Fatalf("reading file failed: %v", err)
	}

	master, err := slide.ParseMaster(xmlData)
	if err != nil {
		t.Fatalf("ParseMaster returned an error: %v", err)
	}

	// Verify the background was extracted correctly.
	bg := master.Background()
	if bg == nil {
		t.Fatal("master Background is nil; expected background data")
	}

	// Verify the master contains placeholders.
	placeholders := master.Placeholders()
	if len(placeholders) == 0 {
		t.Fatal("master Placeholders length is 0; expected at least one placeholder")
	}

	// Masters typically contain date, footer, and slide-number placeholders.
	hasDatePh, hasFooterPh, hasSlideNumPh := false, false, false
	for _, ph := range placeholders {
		switch ph.Type() {
		case slide.PlaceholderTypeDateTime:
			hasDatePh = true
		case slide.PlaceholderTypeFooter:
			hasFooterPh = true
		case slide.PlaceholderTypeSlideNumber:
			hasSlideNumPh = true
		}
	}

	if !hasDatePh {
		t.Log("date placeholder (dt) not found")
	}
	if !hasFooterPh {
		t.Log("footer placeholder (ftr) not found")
	}
	if !hasSlideNumPh {
		t.Log("slide-number placeholder (sldNum) not found")
	}

	// There must be at least one master-level placeholder.
	t.Logf("master contains %d placeholder(s)", len(placeholders))
}
