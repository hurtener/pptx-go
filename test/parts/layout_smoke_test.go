package parts_test

import (
	"os"
	"testing"

	"github.com/hurtener/pptx-go/parts"
)

// ============================================================================
// Layout smoke tests
// ============================================================================

func TestParseLayoutFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{"slideLayout5", "../test-data/test/ppt/slideLayouts/slideLayout5.xml"},
		{"slideLayout7", "../test-data/test/ppt/slideLayouts/slideLayout7.xml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData, err := os.ReadFile(tt.filePath)
			if err != nil {
				t.Fatalf("reading file failed: %v", err)
			}

			layout, err := parts.ParseLayout(xmlData)
			if err != nil {
				t.Fatalf("ParseLayout returned an error: %v", err)
			}
			if layout == nil {
				t.Fatal("ParseLayout returned nil")
			}

			placeholders := layout.Placeholders()
			if len(placeholders) == 0 {
				t.Error("Placeholders length is 0; expected at least 1 placeholder")
			}
		})
	}
}

// ============================================================================
// Layout precise coordinate and type assertions
// ============================================================================

func TestParseLayout_PlaceholderDetails(t *testing.T) {
	xmlData, err := os.ReadFile("../test-data/test/ppt/slideLayouts/slideLayout5.xml")
	if err != nil {
		t.Fatalf("reading file failed: %v", err)
	}

	layout, err := parts.ParseLayout(xmlData)
	if err != nil {
		t.Fatalf("ParseLayout returned an error: %v", err)
	}

	// Retrieve the title placeholder.
	titlePh := layout.PlaceholderByType(parts.PlaceholderTypeTitle)
	if titlePh == nil {
		t.Fatal("title-type placeholder not found")
	}

	// Title placeholder coordinates and size must all be positive.
	if titlePh.X() <= 0 {
		t.Errorf("title placeholder X = %d, want > 0", titlePh.X())
	}
	if titlePh.Y() <= 0 {
		t.Errorf("title placeholder Y = %d, want > 0", titlePh.Y())
	}
	if titlePh.Cx() <= 0 {
		t.Errorf("title placeholder Cx (width) = %d, want > 0", titlePh.Cx())
	}
	if titlePh.Cy() <= 0 {
		t.Errorf("title placeholder Cy (height) = %d, want > 0", titlePh.Cy())
	}

	// Attempt to retrieve the body placeholder (may not exist in this layout).
	bodyPh := layout.PlaceholderByType(parts.PlaceholderTypeBody)
	if bodyPh != nil {
		// Body placeholder coordinates and size must all be positive.
		if bodyPh.X() <= 0 {
			t.Errorf("body placeholder X = %d, want > 0", bodyPh.X())
		}
		if bodyPh.Y() <= 0 {
			t.Errorf("body placeholder Y = %d, want > 0", bodyPh.Y())
		}
		if bodyPh.Cx() <= 0 {
			t.Errorf("body placeholder Cx (width) = %d, want > 0", bodyPh.Cx())
		}
		if bodyPh.Cy() <= 0 {
			t.Errorf("body placeholder Cy (height) = %d, want > 0", bodyPh.Cy())
		}
	}
}
