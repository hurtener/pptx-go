package opc

import (
	"archive/zip"
	"bytes"
	"testing"
)

// FuzzOpen exercises the top-level package-open path — the orchestration that
// ingests untrusted external bytes (ZIP walk, [Content_Types].xml parse, part
// load with the size/zip-slip guards, relationship resolution). Invariant: Open
// never panics; it returns a package or an error (CLAUDE.md §11, §7).
func FuzzOpen(f *testing.F) {
	// A minimal valid package.
	f.Add(buildSeedZip(map[string]string{
		PathContentTypes:        minimalContentTypes,
		"ppt/slides/slide1.xml": "<sld/>",
	}))
	// A package with a relationship part.
	f.Add(buildSeedZip(map[string]string{
		PathContentTypes:           minimalContentTypes,
		"_rels/.rels":              `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="t" Target="ppt/presentation.xml"/></Relationships>`,
		"ppt/presentation.xml":     "<p:presentation/>",
	}))
	// Degenerate inputs.
	f.Add([]byte(""))
	f.Add([]byte("PK\x03\x04 not really a zip"))
	f.Add([]byte("garbage"))

	f.Fuzz(func(t *testing.T, data []byte) {
		pkg, err := Open(bytes.NewReader(data), int64(len(data)))
		if err == nil && pkg == nil {
			t.Fatal("Open returned nil package and nil error")
		}
	})
}

// FuzzContentTypesFromXML fuzzes the [Content_Types].xml parser. Invariant: no
// panic; parsed-or-error.
func FuzzContentTypesFromXML(f *testing.F) {
	f.Add([]byte(minimalContentTypes))
	f.Add([]byte(`<Types><Default Extension="xml" ContentType="application/xml"/><Override PartName="/a.xml" ContentType="x"/></Types>`))
	f.Add([]byte(""))
	f.Add([]byte("<Types"))
	f.Add([]byte("not xml"))

	f.Fuzz(func(t *testing.T, data []byte) {
		ct := NewContentTypes()
		_ = ct.FromXML(data)
	})
}

// FuzzRelationshipsFromXML fuzzes the .rels parser. Invariant: no panic;
// parsed-or-error.
func FuzzRelationshipsFromXML(f *testing.F) {
	f.Add([]byte(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="t" Target="x.xml"/></Relationships>`))
	f.Add([]byte(`<Relationships><Relationship Id="rId1" Type="t" Target="../escape" TargetMode="External"/></Relationships>`))
	f.Add([]byte(""))
	f.Add([]byte("<Relationships"))
	f.Add([]byte("garbage"))

	src := NewPackURI("/ppt/_rels/presentation.xml.rels")
	f.Fuzz(func(t *testing.T, data []byte) {
		rs := NewRelationships(src.SourceURI())
		_ = rs.FromXML(data)
	})
}

// buildSeedZip builds a ZIP from name→content entries for fuzz seeding. It mirrors
// the helper in openconfig_test.go but panics on error (seed construction must
// not fail); the fuzz harness only ever calls it with valid inputs.
func buildSeedZip(entries map[string]string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			panic(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			panic(err)
		}
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
