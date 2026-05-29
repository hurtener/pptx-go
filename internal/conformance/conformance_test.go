package conformance

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
)

func mkPart(t *testing.T, pkg *opc.Package, uri, ct string, blob []byte) *opc.Part {
	t.Helper()
	p, err := pkg.CreatePart(opc.NewPackURI(uri), ct, blob)
	if err != nil {
		t.Fatalf("CreatePart %s: %v", uri, err)
	}
	return p
}

func hasError(rep Report, substr string) bool {
	for _, i := range rep.Errors() {
		if strings.Contains(i.Message, substr) {
			return true
		}
	}
	return false
}

func TestValidCleanPackage(t *testing.T) {
	pkg := opc.NewPackage()
	mkPart(t, pkg, "/ppt/theme/theme1.xml", opc.ContentTypeTheme, []byte(`<a:theme xmlns:a=\"http://schemas.openxmlformats.org/drawingml/2006/main\"/>`))
	pres := mkPart(t, pkg, "/ppt/presentation.xml", opc.ContentTypePresentation,
		[]byte(`<p:presentation xmlns:p=\"http://schemas.openxmlformats.org/presentationml/2006/main\"><p:sldId rid="rId1"/></p:presentation>`))
	if _, err := pres.AddRelationship(opc.RelTypeOfficeDocument, "theme/theme1.xml", false); err != nil {
		t.Fatal(err)
	}
	// Re-point the rel id to rId1 to match the XML reference.
	// (CreatePart's AddRelationship allocates rId1 as the first id.)

	rep := Validate(pkg, Options{})
	if !rep.OK() {
		t.Fatalf("expected clean package, got:\n%s", rep)
	}
}

func TestMissingContentType(t *testing.T) {
	pkg := opc.NewPackage()
	_ = pkg.AddPart(opc.NewPart(opc.NewPackURI("/ppt/orphan.xml"), "", []byte("<x/>")))
	rep := Validate(pkg, Options{})
	if !hasError(rep, "no content type") {
		t.Fatalf("expected content-type error, got:\n%s", rep)
	}
}

func TestDanglingRIDReference(t *testing.T) {
	pkg := opc.NewPackage()
	// XML references rId9 but the part has no relationships at all.
	mkPart(t, pkg, "/ppt/slides/slide1.xml", opc.ContentTypeSlide,
		[]byte(`<sld><blip embed="rId9"/></sld>`))
	rep := Validate(pkg, Options{})
	if !hasError(rep, "rId9") {
		t.Fatalf("expected dangling-rId error, got:\n%s", rep)
	}
}

func TestUnresolvedRelTarget(t *testing.T) {
	pkg := opc.NewPackage()
	p := mkPart(t, pkg, "/ppt/presentation.xml", opc.ContentTypePresentation, []byte("<presentation/>"))
	// Relationship to a part that does not exist.
	if _, err := p.AddRelationship(opc.RelTypeSlide, "slides/slide99.xml", false); err != nil {
		t.Fatal(err)
	}
	rep := Validate(pkg, Options{})
	if !hasError(rep, "not in the package") {
		t.Fatalf("expected unresolved-target error, got:\n%s", rep)
	}
}

func TestRequiredPartsMissing(t *testing.T) {
	pkg := opc.NewPackage()
	mkPart(t, pkg, "/ppt/theme/theme1.xml", opc.ContentTypeTheme, []byte(`<a:theme xmlns:a=\"http://schemas.openxmlformats.org/drawingml/2006/main\"/>`))
	rep := Validate(pkg, Options{RequiredParts: []string{"/ppt/presentation.xml"}})
	if !hasError(rep, "required part is missing") {
		t.Fatalf("expected required-part error, got:\n%s", rep)
	}
	// And present when it exists.
	mkPart(t, pkg, "/ppt/presentation.xml", opc.ContentTypePresentation, []byte("<presentation/>"))
	rep = Validate(pkg, Options{RequiredParts: []string{"/ppt/presentation.xml"}})
	if hasError(rep, "required part is missing") {
		t.Fatalf("required part present but still flagged:\n%s", rep)
	}
}

func TestExternalRelationshipNotFlagged(t *testing.T) {
	pkg := opc.NewPackage()
	p := mkPart(t, pkg, "/ppt/slides/slide1.xml", opc.ContentTypeSlide, []byte(`<p:sld xmlns:p=\"http://schemas.openxmlformats.org/presentationml/2006/main\"/>`))
	// An external hyperlink target must not be reported as a missing part.
	if _, err := p.AddRelationship(opc.RelTypeHyperlink, "https://example.com", true); err != nil {
		t.Fatal(err)
	}
	rep := Validate(pkg, Options{})
	if !rep.OK() {
		t.Fatalf("external relationship should not fail conformance:\n%s", rep)
	}
}

func TestValidateBytesRoundTrip(t *testing.T) {
	pkg := opc.NewPackage()
	mkPart(t, pkg, "/ppt/theme/theme1.xml", opc.ContentTypeTheme, []byte(`<a:theme xmlns:a=\"http://schemas.openxmlformats.org/drawingml/2006/main\"/>`))
	var buf bytes.Buffer
	if err := pkg.Save(&buf); err != nil {
		t.Fatal(err)
	}
	rep, err := ValidateBytes(buf.Bytes(), Options{RequiredParts: []string{"/ppt/theme/theme1.xml"}})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("round-tripped package should be clean:\n%s", rep)
	}
	// And a corrupt byte slice errors.
	if _, err := ValidateBytes([]byte("not a zip"), Options{}); err == nil {
		t.Error("expected error opening non-zip bytes")
	}
}

func TestMissingRootNamespaceFlagged(t *testing.T) {
	pkg := opc.NewPackage()
	// A presentation part with a bare (namespace-less) root — the inherited
	// emission bug (D-032).
	mkPart(t, pkg, "/ppt/presentation.xml", opc.ContentTypePresentation, []byte("<presentation/>"))
	rep := Validate(pkg, Options{})
	if !hasError(rep, "no XML namespace") {
		t.Fatalf("expected root-namespace error, got:\n%s", rep)
	}
}

func TestEmptyRelReferenceFlagged(t *testing.T) {
	pkg := opc.NewPackage()
	// sldId with an empty rid — the orphaned-slide bug (D-032).
	mkPart(t, pkg, "/ppt/presentation.xml", opc.ContentTypePresentation,
		[]byte(`<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"><p:sldId rid=""/></p:presentation>`))
	rep := Validate(pkg, Options{})
	if !hasError(rep, "empty relationship reference") {
		t.Fatalf("expected empty-rel-reference error, got:\n%s", rep)
	}
}

func TestReportStringAndOK(t *testing.T) {
	var empty Report
	if !empty.OK() || !strings.Contains(empty.String(), "OK") {
		t.Errorf("empty report should be OK: %q", empty.String())
	}
	r := Report{Issues: []Issue{{Severity: SeverityError, Part: "/x", Message: "boom"}}}
	if r.OK() {
		t.Error("report with an error must not be OK")
	}
	if !strings.Contains(r.String(), "boom") {
		t.Errorf("String missing message: %q", r.String())
	}
}
