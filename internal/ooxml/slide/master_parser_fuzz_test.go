package slide

import "testing"

// FuzzParseLayout exercises the slide-layout parser — the first foreign-XML read
// surface template ingestion depends on (Phase 09, brief 01 F6). Invariant: the
// parser never panics; it returns either parsed data or an error.
func FuzzParseLayout(f *testing.F) {
	f.Add([]byte(`<p:sldLayout xmlns:p="x" type="title"><p:cSld name="Title Slide"><p:spTree/></p:cSld></p:sldLayout>`))
	f.Add([]byte(`<p:sldLayout><p:cSld><p:spTree/></p:cSld></p:sldLayout>`))
	f.Add([]byte(``))
	f.Add([]byte(`not xml`))
	f.Add([]byte(`<p:sldLayout`))

	f.Fuzz(func(t *testing.T, data []byte) {
		layout, err := ParseLayout(data)
		if err == nil && layout == nil {
			t.Fatal("ParseLayout returned nil data and nil error")
		}
	})
}

// FuzzParseMaster exercises the slide-master parser under the same invariant.
func FuzzParseMaster(f *testing.F) {
	f.Add([]byte(`<p:sldMaster xmlns:p="x"><p:cSld name="Office Theme"><p:spTree/></p:cSld></p:sldMaster>`))
	f.Add([]byte(`<p:sldMaster><p:cSld><p:spTree/></p:cSld></p:sldMaster>`))
	f.Add([]byte(``))
	f.Add([]byte(`garbage`))

	f.Fuzz(func(t *testing.T, data []byte) {
		master, err := ParseMaster(data)
		if err == nil && master == nil {
			t.Fatal("ParseMaster returned nil data and nil error")
		}
	})
}
