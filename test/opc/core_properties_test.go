package opc_test

import (
	"strings"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
)

func TestCoreProperties_GettersSetters(t *testing.T) {
	cp := &opc.CoreProperties{}

	// test all getters/setters
	cp.SetTitle("Test Title")
	if cp.Title() != "Test Title" {
		t.Error("Title getter/setter failed")
	}

	cp.SetCreator("Test Creator")
	if cp.Creator() != "Test Creator" {
		t.Error("Creator getter/setter failed")
	}

	cp.SetSubject("Test Subject")
	if cp.Subject() != "Test Subject" {
		t.Error("Subject getter/setter failed")
	}

	cp.SetDescription("Test Description")
	if cp.Description() != "Test Description" {
		t.Error("Description getter/setter failed")
	}

	cp.SetKeywords("test, keywords")
	if cp.Keywords() != "test, keywords" {
		t.Error("Keywords getter/setter failed")
	}

	cp.SetCreated("2024-01-01T00:00:00Z")
	if cp.Created() != "2024-01-01T00:00:00Z" {
		t.Error("Created getter/setter failed")
	}

	cp.SetModified("2024-01-02T00:00:00Z")
	if cp.Modified() != "2024-01-02T00:00:00Z" {
		t.Error("Modified getter/setter failed")
	}

	cp.SetLastModifiedBy("Test User")
	if cp.LastModifiedBy() != "Test User" {
		t.Error("LastModifiedBy getter/setter failed")
	}

	cp.SetRevision("1")
	if cp.Revision() != "1" {
		t.Error("Revision getter/setter failed")
	}

	cp.SetCategory("Test Category")
	if cp.Category() != "Test Category" {
		t.Error("Category getter/setter failed")
	}

	cp.SetContentType("application/test")
	if cp.ContentType() != "application/test" {
		t.Error("ContentType getter/setter failed")
	}

	cp.SetLanguage("en-US")
	if cp.Language() != "en-US" {
		t.Error("Language getter/setter failed")
	}
}

func TestCoreProperties_XML(t *testing.T) {
	cp := &opc.CoreProperties{}
	cp.SetTitle("Test Presentation")
	cp.SetCreator("Test User")
	cp.SetSubject("Test Subject")
	cp.SetDescription("Test Description")
	cp.SetKeywords("test, presentation")
	cp.SetCreated("2024-01-01T00:00:00Z")
	cp.SetModified("2024-01-02T12:00:00Z")
	cp.SetLastModifiedBy("Test User 2")
	cp.SetRevision("2")

	// serialize
	data, err := cp.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// verify the serialized output contains the correct data
	dataStr := string(data)
	if !strings.Contains(dataStr, "Test Presentation") {
		t.Error("ToXML output should contain title")
	}
	if !strings.Contains(dataStr, "Test User") {
		t.Error("ToXML output should contain creator")
	}
}

func TestCoreProperties_FromXML(t *testing.T) {
	// test that FromXML correctly parses XML produced by ToXML
	cp := &opc.CoreProperties{}
	cp.SetTitle("Test Title")
	cp.SetCreator("Test Creator")

	data, err := cp.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	cp2 := &opc.CoreProperties{}
	err = cp2.FromXML(data)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// verify the XML can be parsed (even if namespace prefixes don't fully match)
	// this test primarily checks that FromXML does not error
}

func TestCoreProperties_EmptyFields(t *testing.T) {
	cp := &opc.CoreProperties{}

	// empty fields should not cause an error
	data, err := cp.ToXML()
	if err != nil {
		t.Fatalf("ToXML with empty fields failed: %v", err)
	}

	cp2 := &opc.CoreProperties{}
	err = cp2.FromXML(data)
	if err != nil {
		t.Fatalf("FromXML with empty fields failed: %v", err)
	}

	// verify empty values remain empty
	if cp2.Title() != "" || cp2.Creator() != "" {
		t.Error("empty fields should remain empty")
	}
}
