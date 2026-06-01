package icons_test

import (
	"testing"

	"github.com/hurtener/pptx-go/assets/icons"
	"github.com/hurtener/pptx-go/pptx"
)

// TestCuratedIcons_AllTranslate is the build-time asset-validity gate (RFC
// §14.1, D-005): every embedded curated icon must satisfy the SVG translator
// constraints. A new icon that violates them fails here, not at render.
func TestCuratedIcons_AllTranslate(t *testing.T) {
	names := icons.Names()
	if len(names) < 16 {
		t.Fatalf("curated set has %d icons, want the starter set (>= 16)", len(names))
	}
	for _, name := range names {
		svg, ok := icons.Read(name)
		if !ok {
			t.Errorf("Names() listed %q but Read returned not-found", name)
			continue
		}
		if err := pptx.ValidateIcon(svg); err != nil {
			t.Errorf("curated icon %q does not translate: %v", name, err)
		}
	}
}

// TestCuratedIcons_ReadMiss confirms an unknown name reports not-found.
func TestCuratedIcons_ReadMiss(t *testing.T) {
	if _, ok := icons.Read("no-such-icon"); ok {
		t.Error("Read(no-such-icon) reported found")
	}
}
