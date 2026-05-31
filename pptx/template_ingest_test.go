package pptx

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/conformance"
	"github.com/hurtener/pptx-go/internal/opc"
)

// minimalLayoutXML is a small but schema-shaped slideLayout carrying a name and
// type — enough to exercise the name→layout registry and slide→layout targeting.
func minimalLayoutXML(name, typ string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main" type=%q preserve="1">
<p:cSld name=%q>
<p:spTree>
<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
<p:grpSpPr/>
</p:spTree>
</p:cSld>
<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sldLayout>`, typ, name)
}

// brandTemplateBytes builds an in-memory brand kit: a valid deck whose
// theme1.xml carries a custom accent + fonts, plus an extra named layout
// ("Title Slide") wired to the master. Returns the .pptx bytes.
func brandTemplateBytes(t *testing.T) []byte {
	t.Helper()
	base := New()

	// Brand theme1.xml: the conformant scaffold theme (PowerPoint-shaped — note
	// the sysClr dk1/lt1 slots, brief 01 F1) with a custom accent + fonts patched
	// in. This mirrors a real brand kit's theme part (a hand-edited theme1.xml),
	// without relying on the bare-name ThemeXML() writer that targets the
	// namespace-restore hygiene pass.
	brandThemeXML := strings.NewReplacer(
		`<a:accent1><a:srgbClr val="2563EB"/></a:accent1>`, `<a:accent1><a:srgbClr val="AB12CD"/></a:accent1>`,
		"<a:majorFont>\n<a:latin typeface=\"Arial\"/>", "<a:majorFont>\n<a:latin typeface=\"Georgia\"/>",
		"<a:minorFont>\n<a:latin typeface=\"Arial\"/>", "<a:minorFont>\n<a:latin typeface=\"Verdana\"/>",
	).Replace(scaffoldThemeXML)
	tp := base.Package().GetPart(opc.NewPackURI(themeURI))
	if tp == nil {
		t.Fatal("scaffold theme part missing")
	}
	tp.SetBlob([]byte(brandThemeXML))

	// Inject a second, named layout ("Title Slide") and wire it both ways.
	l2 := opc.NewPart(opc.NewPackURI("/ppt/slideLayouts/slideLayout2.xml"), opc.ContentTypeSlideLayout, []byte(minimalLayoutXML("Title Slide", "title")))
	_, _ = l2.AddRelationship(opc.RelTypeSlideMaster, "../slideMasters/slideMaster1.xml", false)
	_ = base.Package().AddPart(l2)
	if mp := base.Package().GetPart(opc.NewPackURI(slideMasterURI)); mp != nil {
		_, _ = mp.AddRelationship(opc.RelTypeSlideLayout, "../slideLayouts/slideLayout2.xml", false)
	}

	data, err := base.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	return data
}

// TestOpen_ExtractsThemeAndLayouts covers the open path (RFC §13.1): opening a
// brand kit yields its theme via Theme() and its layouts via Masters().
func TestOpen_ExtractsThemeAndLayouts(t *testing.T) {
	brand, err := NewFromBytes(brandTemplateBytes(t))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}

	if got := brand.Theme().ResolveColor(ColorAccent); got != "AB12CD" {
		t.Errorf("brand accent = %q, want AB12CD", got)
	}
	if got := brand.Theme().HeadingFont; got != "Georgia" {
		t.Errorf("brand heading font = %q, want Georgia", got)
	}
	// The dk1 slot is a sysClr (windowText); its lastClr fallback must resolve
	// (brief 01 F1 — PowerPoint encodes dk1/lt1 as system colors).
	if got := brand.Theme().ResolveTextColor(TextPrimary); got != "000000" {
		t.Errorf("text-primary (from sysClr dk1) = %q, want 000000", got)
	}

	masters := brand.Masters()
	if len(masters) == 0 {
		t.Fatal("Masters() empty for a deck with a master")
	}
	names := map[string]bool{}
	for _, m := range masters {
		for _, l := range m.Layouts() {
			names[l.Name()] = true
		}
	}
	if !names["Blank"] || !names["Title Slide"] {
		t.Errorf("layout names = %v, want Blank + Title Slide", names)
	}
	if !brand.HasLayout("Title Slide") {
		t.Error("HasLayout(Title Slide) = false")
	}
}

// TestFromTemplate_AdoptsThemeAndMasters is acceptance criteria 1 & 2: seeding a
// presentation from a brand kit adopts its theme and exposes its layouts.
func TestFromTemplate_AdoptsThemeAndMasters(t *testing.T) {
	brand, err := NewFromBytes(brandTemplateBytes(t))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}

	pres := New(FromTemplate(brand))
	if got := pres.Theme().ResolveColor(ColorAccent); got != "AB12CD" {
		t.Errorf("seeded theme accent = %q, want AB12CD", got)
	}
	if pres.SlideCount() != 0 {
		t.Errorf("seeded deck has %d slides, want 0", pres.SlideCount())
	}
	if !pres.HasLayout("Title Slide") {
		t.Error("seeded deck missing the template's Title Slide layout")
	}

	// The seeded deck is usable and conformant once a slide is added.
	pres.AddSlide("Title Slide")
	data, err := pres.WriteToBytes()
	if err != nil {
		t.Fatalf("WriteToBytes: %v", err)
	}
	rep, err := conformance.ValidateBytes(data, conformance.Options{RequiredParts: []string{"/ppt/presentation.xml", "/ppt/slides/slide1.xml"}})
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("seeded deck failed conformance:\n%s", rep)
	}
}

// TestAddSlide_TargetsNamedLayout proves a slide added with a named layout
// relates to that layout's part, and an unknown name falls back to blank.
func TestAddSlide_TargetsNamedLayout(t *testing.T) {
	brand, err := NewFromBytes(brandTemplateBytes(t))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	pres := New(FromTemplate(brand))

	named := pres.AddSlide("Title Slide")
	if got := layoutRelTarget(t, named); got != "../slideLayouts/slideLayout2.xml" {
		t.Errorf("named slide layout rel = %q, want ../slideLayouts/slideLayout2.xml", got)
	}

	blank := pres.AddSlide("Does Not Exist")
	if got := layoutRelTarget(t, blank); got != "../slideLayouts/slideLayout1.xml" {
		t.Errorf("fallback slide layout rel = %q, want ../slideLayouts/slideLayout1.xml", got)
	}
}

// layoutRelTarget returns the slide's slide→layout relationship target.
func layoutRelTarget(t *testing.T, s *Slide) string {
	t.Helper()
	for _, rel := range s.Part().Relationships().GetByType(opc.RelTypeSlideLayout) {
		if rel.TargetURI() != nil {
			// Compare on the trailing relative form recorded on the slide rel.
			return rel.TargetRef()
		}
	}
	t.Fatal("slide has no slideLayout relationship")
	return ""
}

// TestFromTemplate_RoundTripDeterministic is acceptance criterion 5: a deck
// seeded from a template round-trips and renders byte-identically (D-035).
func TestFromTemplate_RoundTripDeterministic(t *testing.T) {
	tmpl := brandTemplateBytes(t)

	build := func() []byte {
		brand, err := NewFromBytes(tmpl)
		if err != nil {
			t.Fatalf("NewFromBytes: %v", err)
		}
		pres := New(FromTemplate(brand))
		pres.AddSlide("Title Slide")
		pres.AddSlide()
		data, err := pres.WriteToBytes()
		if err != nil {
			t.Fatalf("WriteToBytes: %v", err)
		}
		return data
	}

	a, b := build(), build()
	if !bytes.Equal(a, b) {
		t.Fatal("two FromTemplate builds are not byte-identical")
	}

	// Reopen the seeded deck: theme + layouts survive the round-trip.
	reopened, err := NewFromBytes(a)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if reopened.SlideCount() != 2 {
		t.Errorf("reopened slide count = %d, want 2", reopened.SlideCount())
	}
	if got := reopened.Theme().ResolveColor(ColorAccent); got != "AB12CD" {
		t.Errorf("reopened accent = %q, want AB12CD", got)
	}
}
