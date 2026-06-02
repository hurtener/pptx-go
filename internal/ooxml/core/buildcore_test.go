package core

import (
	"strings"
	"testing"
)

// TestBuildCorePropsXML covers escaping, omit-empty, determinism, and the
// absence of timestamps.
func TestBuildCorePropsXML(t *testing.T) {
	out := string(BuildCorePropsXML("A & B <x>", "Acme", ""))
	if !strings.Contains(out, "<dc:title>A &amp; B &lt;x&gt;</dc:title>") {
		t.Errorf("title not escaped: %s", out)
	}
	if !strings.Contains(out, "<dc:creator>Acme</dc:creator>") {
		t.Errorf("creator missing: %s", out)
	}
	if strings.Contains(out, "<dc:subject>") {
		t.Errorf("empty subject should be omitted: %s", out)
	}
	if strings.Contains(out, "dcterms:created") || strings.Contains(out, "dcterms:modified") {
		t.Errorf("no timestamps expected (determinism): %s", out)
	}

	// Deterministic: identical input → identical bytes.
	a := BuildCorePropsXML("T", "C", "S")
	b := BuildCorePropsXML("T", "C", "S")
	if string(a) != string(b) {
		t.Error("BuildCorePropsXML is not deterministic")
	}

	// All-empty: a well-formed but propertyless core.xml.
	empty := string(BuildCorePropsXML("", "", ""))
	if !strings.Contains(empty, "<cp:coreProperties") || strings.Contains(empty, "<dc:") {
		t.Errorf("empty metadata should produce a bare coreProperties element: %s", empty)
	}
}
