package pptx

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
)

// TestRead_LoggerCapturesDegradation is the §8 read-path observability fix: a
// logger injected via WithLogger on a read constructor receives a Warn event for
// each degradation, not just the silent ReadWarnings slice.
func TestRead_LoggerCapturesDegradation(t *testing.T) {
	// An external-style deck with an unrecognized shape-tree element.
	parts := unzipParts(t, authoredDeck(t))
	const slide = "ppt/slides/slide1.xml"
	const grpSp = `<p:grpSp><p:nvGrpSpPr><p:cNvPr id="42" name="grp"/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:grpSp>`
	parts[slide] = []byte(strings.Replace(string(parts[slide]), "</p:spTree>", grpSp+"</p:spTree>", 1))
	data := rezipParts(t, parts)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	re, err := NewFromBytes(data, WithLogger(logger))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if len(re.ReadWarnings()) == 0 {
		t.Fatal("expected a read warning for the dropped element")
	}
	logged := buf.String()
	if !strings.Contains(logged, "read degradation") || !strings.Contains(logged, "dropped-element") {
		t.Errorf("logger did not capture the degradation; got:\n%s", logged)
	}
	if !strings.Contains(logged, "grpSp") {
		t.Errorf("logged degradation missing the element name; got:\n%s", logged)
	}
}

// TestRead_NoLoggerIsQuiet confirms degradation is silent (only via ReadWarnings)
// when no logger is injected — no global logger, zero-cost default.
func TestRead_NoLoggerIsQuiet(t *testing.T) {
	parts := unzipParts(t, authoredDeck(t))
	const slide = "ppt/slides/slide1.xml"
	parts[slide] = []byte(strings.Replace(string(parts[slide]), "</p:spTree>",
		`<p:grpSp/></p:spTree>`, 1))
	re, err := NewFromBytes(rezipParts(t, parts))
	if err != nil {
		t.Fatalf("NewFromBytes: %v", err)
	}
	if len(re.ReadWarnings()) == 0 {
		t.Error("expected the warning to still be collected without a logger")
	}
}

// TestRead_PartLimitOption is the caller-configurable half of §7: WithReadPartLimit
// rejects an oversized part through the public read constructor with an error
// wrapping opc.ErrPartTooLarge.
func TestRead_PartLimitOption(t *testing.T) {
	parts := unzipParts(t, authoredDeck(t))
	const custom = "customXml/big.xml"
	parts[custom] = bytes.Repeat([]byte("A"), 64*1024)
	ct := string(parts["[Content_Types].xml"])
	parts["[Content_Types].xml"] = []byte(strings.Replace(ct, "</Types>",
		`<Override PartName="/customXml/big.xml" ContentType="application/xml"/></Types>`, 1))
	data := rezipParts(t, parts)

	_, err := NewFromBytes(data, WithReadPartLimit(8*1024))
	if !errors.Is(err, opc.ErrPartTooLarge) {
		t.Fatalf("NewFromBytes with small read limit = %v, want ErrPartTooLarge", err)
	}

	// The same deck opens when the bound is disabled.
	if _, err := NewFromBytes(data, WithReadPartLimit(0)); err != nil {
		t.Errorf("NewFromBytes with unlimited read limit: %v", err)
	}
}
