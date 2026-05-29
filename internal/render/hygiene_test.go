package render

import (
	"bytes"
	"testing"
)

func TestSanitize_StripsBOM(t *testing.T) {
	in := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`<?xml version="1.0"?><a:t>hi</a:t>`)...)
	got := Sanitize(in)
	if bytes.HasPrefix(got, []byte{0xEF, 0xBB, 0xBF}) {
		t.Errorf("BOM survived: % x", got[:3])
	}
	if !bytes.HasPrefix(got, []byte(`<?xml`)) {
		t.Errorf("expected XML declaration at start, got: %q", got)
	}
}

func TestSanitize_RemovesEmptyLang(t *testing.T) {
	cases := []struct{ in, want string }{
		{`<a:rPr lang="" sz="1800"/>`, `<a:rPr sz="1800"/>`},
		{`<a:rPr lang=''/>`, `<a:rPr/>`},
		{`<a:rPr sz="1800" lang=""/>`, `<a:rPr sz="1800"/>`},
	}
	for _, c := range cases {
		if got := string(Sanitize([]byte(c.in))); got != c.want {
			t.Errorf("Sanitize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestSanitize_PreservesNonEmptyLang proves the pass only removes EMPTY lang,
// never a populated one.
func TestSanitize_PreservesNonEmptyLang(t *testing.T) {
	in := `<a:rPr lang="en-US" sz="1800"/>`
	if got := string(Sanitize([]byte(in))); got != in {
		t.Errorf("Sanitize altered a non-empty lang: %q", got)
	}
}

// TestSanitize_NoTriggerUnchanged proves trigger-free XML is returned
// byte-for-byte (the conservative-pass guarantee).
func TestSanitize_NoTriggerUnchanged(t *testing.T) {
	in := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<p:sld xmlns:p="x"><p:cSld><p:spTree/></p:cSld></p:sld>`)
	got := Sanitize(in)
	if !bytes.Equal(got, in) {
		t.Errorf("trigger-free XML was modified:\n in: %s\nout: %s", in, got)
	}
}

// TestSanitize_Idempotent proves a second pass is a no-op.
func TestSanitize_Idempotent(t *testing.T) {
	in := []byte("\xEF\xBB\xBF<a:rPr lang=\"\" sz=\"1800\"/>")
	once := Sanitize(in)
	twice := Sanitize(once)
	if !bytes.Equal(once, twice) {
		t.Errorf("not idempotent:\nonce:  %q\ntwice: %q", once, twice)
	}
}
