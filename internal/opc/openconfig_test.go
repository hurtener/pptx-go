package opc

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"
)

// minimalContentTypes is a valid, empty [Content_Types].xml so Open gets past
// loadContentTypes and reaches the part/zip-slip/size checks.
const minimalContentTypes = `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`

// buildZip writes name→content entries into a ZIP and returns the bytes.
func buildZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func openBytes(t *testing.T, data []byte, opts ...OpenOption) (*Package, error) {
	t.Helper()
	return Open(bytes.NewReader(data), int64(len(data)), opts...)
}

func TestSafePartPath(t *testing.T) {
	cases := []struct {
		name string
		path string
		ok   bool
	}{
		{"normal", "ppt/slides/slide1.xml", true},
		{"nested", "ppt/media/image1.png", true},
		{"empty", "", true},
		{"parent escape", "../evil.xml", false},
		{"nested parent escape", "ppt/../../evil.xml", false},
		{"rels parent escape", "_rels/../evil.rels", false},
		{"absolute", "/etc/passwd", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := safePartPath(c.path)
			if c.ok && err != nil {
				t.Errorf("safePartPath(%q) = %v, want nil", c.path, err)
			}
			if !c.ok {
				if err == nil {
					t.Errorf("safePartPath(%q) = nil, want ErrUnsafePartPath", c.path)
				} else if !errors.Is(err, ErrUnsafePartPath) {
					t.Errorf("safePartPath(%q) = %v, want ErrUnsafePartPath", c.path, err)
				}
			}
		})
	}
}

// TestOpen_RejectsZipSlip is §7: a ZIP entry escaping the package root is
// rejected at parse time rather than admitted as a part.
func TestOpen_RejectsZipSlip(t *testing.T) {
	data := buildZip(t, map[string]string{
		PathContentTypes:        minimalContentTypes,
		"ppt/slides/slide1.xml": "<sld/>",
		"../evil.xml":           "pwned",
	})
	_, err := openBytes(t, data)
	if !errors.Is(err, ErrUnsafePartPath) {
		t.Fatalf("Open with zip-slip entry = %v, want ErrUnsafePartPath", err)
	}
}

// TestOpen_RejectsZipSlipInRels covers the .rels load path's zip-slip guard.
func TestOpen_RejectsZipSlipInRels(t *testing.T) {
	data := buildZip(t, map[string]string{
		PathContentTypes:   minimalContentTypes,
		"_rels/../evil.rels": `<Relationships/>`,
	})
	_, err := openBytes(t, data)
	if !errors.Is(err, ErrUnsafePartPath) {
		t.Fatalf("Open with zip-slip .rels = %v, want ErrUnsafePartPath", err)
	}
}

// TestOpen_RejectsOversizedPart is §7: a part exceeding the per-part limit is
// rejected with ErrPartTooLarge rather than allocated.
func TestOpen_RejectsOversizedPart(t *testing.T) {
	big := strings.Repeat("A", 4096)
	data := buildZip(t, map[string]string{
		PathContentTypes:        minimalContentTypes,
		"ppt/slides/slide1.xml": big,
	})
	_, err := openBytes(t, data, WithMaxPartBytes(1024))
	if !errors.Is(err, ErrPartTooLarge) {
		t.Fatalf("Open with oversized part = %v, want ErrPartTooLarge", err)
	}
}

// TestOpen_WithinLimitAndUnlimited confirms a part under the bound opens, and
// that WithMaxPartBytes(0) disables the bound.
func TestOpen_WithinLimitAndUnlimited(t *testing.T) {
	big := strings.Repeat("A", 4096)
	data := buildZip(t, map[string]string{
		PathContentTypes:        minimalContentTypes,
		"ppt/slides/slide1.xml": big,
	})
	if _, err := openBytes(t, data, WithMaxPartBytes(8192)); err != nil {
		t.Errorf("Open within limit: %v", err)
	}
	if _, err := openBytes(t, data, WithMaxPartBytes(0)); err != nil {
		t.Errorf("Open unlimited: %v", err)
	}
}

// TestOpen_DefaultBoundAppliesAndAllowsNormalDecks confirms the default bound is
// active without any option (a normal small deck opens; the 100 MB ceiling is the
// implicit default).
func TestOpen_DefaultBoundActive(t *testing.T) {
	if DefaultMaxPartBytes != 100<<20 {
		t.Fatalf("DefaultMaxPartBytes = %d, want 100 MiB", DefaultMaxPartBytes)
	}
	data := buildZip(t, map[string]string{
		PathContentTypes:        minimalContentTypes,
		"ppt/slides/slide1.xml": "<sld/>",
	})
	if _, err := openBytes(t, data); err != nil {
		t.Fatalf("Open normal deck with default bound: %v", err)
	}
}
