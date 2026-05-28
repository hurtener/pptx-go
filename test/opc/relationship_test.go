package opc_test

import (
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/hurtener/pptx-go/internal/opc"
)

func TestRelationship_New(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rel := opc.NewRelationship("rId1", opc.RelTypeSlide, "/ppt/slides/slide1.xml", false, source)

	if rel.RID() != "rId1" {
		t.Errorf("RID() = %q, want %q", rel.RID(), "rId1")
	}
	if rel.Type() != opc.RelTypeSlide {
		t.Errorf("Type() = %q, want %q", rel.Type(), opc.RelTypeSlide)
	}
	if rel.IsExternal() {
		t.Error("internal relationship should not be external")
	}
	if rel.TargetMode() != "Internal" {
		t.Errorf("TargetMode() = %q, want %q", rel.TargetMode(), "Internal")
	}
}

func TestRelationship_External(t *testing.T) {
	source := opc.NewPackURI("/ppt/slides/slide1.xml")
	rel := opc.NewRelationship("rId2", opc.RelTypeHyperlink, "http://example.com", true, source)

	if !rel.IsExternal() {
		t.Error("external relationship should be external")
	}
	if rel.TargetMode() != "External" {
		t.Errorf("TargetMode() = %q, want %q", rel.TargetMode(), "External")
	}
}

func TestRelationship_TargetURI(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rel := opc.NewRelationship("rId1", opc.RelTypeSlide, "/ppt/slides/slide1.xml", false, source)

	target := rel.TargetURI()
	if target.URI() != "/ppt/slides/slide1.xml" {
		t.Errorf("TargetURI() = %q, want %q", target.URI(), "/ppt/slides/slide1.xml")
	}
}

func TestRelationship_TargetRef(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rel := opc.NewRelationship("rId1", opc.RelTypeSlide, "/ppt/slides/slide1.xml", false, source)

	ref := rel.TargetRef()
	if ref == "" {
		t.Error("TargetRef should not be empty")
	}

	// external relationship
	externalRel := opc.NewRelationship("rId2", opc.RelTypeHyperlink, "http://example.com", true, source)
	extRef := externalRel.TargetRef()
	if extRef != "http://example.com" {
		t.Errorf("external TargetRef = %q, want %q", extRef, "http://example.com")
	}
}

func TestRelationship_Equals(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rel1 := opc.NewRelationship("rId1", opc.RelTypeSlide, "/ppt/slides/slide1.xml", false, source)
	rel2 := opc.NewRelationship("rId1", opc.RelTypeSlide, "/ppt/slides/slide1.xml", false, source)
	rel3 := opc.NewRelationship("rId2", opc.RelTypeSlide, "/ppt/slides/slide1.xml", false, source)

	if !rel1.Equals(rel2) {
		t.Error("identical relationships should be equal")
	}
	if rel1.Equals(rel3) {
		t.Error("different relationships should not be equal")
	}
	if rel1.Equals(nil) {
		t.Error("relationship should not equal nil")
	}
}

func TestRelationships_New(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)

	if rels == nil {
		t.Fatal("NewRelationships returned nil")
	}
	if rels.Count() != 0 {
		t.Error("new relationships should be empty")
	}
}

func TestRelationships_Add(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rel := opc.NewRelationship("rId1", opc.RelTypeSlide, "/ppt/slides/slide1.xml", false, source)

	err := rels.Add(rel)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if rels.Count() != 1 {
		t.Errorf("Count() = %d, want 1", rels.Count())
	}

	// adding a duplicate rID should fail
	rel2 := opc.NewRelationship("rId1", opc.RelTypeSlide, "/ppt/slides/slide2.xml", false, source)
	err = rels.Add(rel2)
	if err == nil {
		t.Error("adding duplicate rID should fail")
	}
}

func TestRelationships_AddNew(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)

	rel, err := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)
	if err != nil {
		t.Fatalf("AddNew failed: %v", err)
	}
	if rel.RID() != "rId1" {
		t.Errorf("first rID = %q, want %q", rel.RID(), "rId1")
	}

	rel2, err := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide2.xml", false)
	if err != nil {
		t.Fatalf("AddNew failed: %v", err)
	}
	if rel2.RID() != "rId2" {
		t.Errorf("second rID = %q, want %q", rel2.RID(), "rId2")
	}
}

func TestRelationships_Get(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rel, _ := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)

	got := rels.Get("rId1")
	if got == nil {
		t.Fatal("Get returned nil")
	}
	if got.RID() != rel.RID() {
		t.Error("Get returned wrong relationship")
	}

	// get a non-existent rID
	if rels.Get("rId999") != nil {
		t.Error("Get for non-existent rID should return nil")
	}
}

func TestRelationships_GetByType(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide2.xml", false)
	rels.AddNew(opc.RelTypeTheme, "/ppt/theme/theme1.xml", false)

	slides := rels.GetByType(opc.RelTypeSlide)
	if len(slides) != 2 {
		t.Errorf("GetByType(slide) returned %d, want 2", len(slides))
	}

	themes := rels.GetByType(opc.RelTypeTheme)
	if len(themes) != 1 {
		t.Errorf("GetByType(theme) returned %d, want 1", len(themes))
	}

	images := rels.GetByType(opc.RelTypeImage)
	if len(images) != 0 {
		t.Error("GetByType for non-existent type should return empty")
	}
}

func TestRelationships_GetByTarget(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)

	rel := rels.GetByTarget(opc.NewPackURI("/ppt/slides/slide1.xml"))
	if rel == nil {
		t.Fatal("GetByTarget returned nil")
	}

	// get a non-existent target
	if rels.GetByTarget(opc.NewPackURI("/ppt/slides/slide999.xml")) != nil {
		t.Error("GetByTarget for non-existent target should return nil")
	}
}

func TestRelationships_Remove(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)

	err := rels.Remove("rId1")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if rels.Count() != 0 {
		t.Error("relationship should be removed")
	}

	// removing a non-existent rID should fail
	err = rels.Remove("rId999")
	if err == nil {
		t.Error("removing non-existent rID should fail")
	}
}

func TestRelationships_Contains(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)

	if !rels.Contains("rId1") {
		t.Error("should contain rId1")
	}
	if rels.Contains("rId999") {
		t.Error("should not contain rId999")
	}
}

func TestRelationships_All(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide2.xml", false)

	all := rels.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d, want 2", len(all))
	}
}

func TestRelationships_NextRID(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)

	// empty set should return rId1
	if rels.NextRID() != "rId1" {
		t.Error("first NextRID should be rId1")
	}

	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)
	if rels.NextRID() != "rId2" {
		t.Error("second NextRID should be rId2")
	}
}

func TestRelationships_Clone(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)

	clone := rels.Clone()
	if clone.Count() != rels.Count() {
		t.Error("clone should have same count")
	}

	// modifying the clone should not affect the original
	clone.AddNew(opc.RelTypeSlide, "/ppt/slides/slide2.xml", false)
	if rels.Count() == clone.Count() {
		t.Error("modifying clone should not affect original")
	}
}

func TestRelationships_XML(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)

	// serialize
	data, err := rels.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// deserialize
	rels2 := opc.NewRelationships(source)
	err = rels2.FromXML(data)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	if rels2.Count() != 1 {
		t.Errorf("Count after round-trip = %d, want 1", rels2.Count())
	}

	rel := rels2.Get("rId1")
	if rel == nil {
		t.Fatal("rId1 not found after round-trip")
	}
	if rel.Type() != opc.RelTypeSlide {
		t.Errorf("Type after round-trip = %q, want %q", rel.Type(), opc.RelTypeSlide)
	}
}

func TestRelationships_FromXML(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide2.xml" TargetMode="External"/>
</Relationships>`

	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)
	err := rels.FromXML([]byte(xmlData))
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	if rels.Count() != 2 {
		t.Errorf("Count = %d, want 2", rels.Count())
	}

	rel1 := rels.Get("rId1")
	if rel1 == nil || rel1.IsExternal() {
		t.Error("rId1 should be internal")
	}

	rel2 := rels.Get("rId2")
	if rel2 == nil || !rel2.IsExternal() {
		t.Error("rId2 should be external")
	}
}

func TestParseRelationshipsFromXML(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
</Relationships>`

	source := opc.NewPackURI("/ppt/presentation.xml")
	rels, err := opc.ParseRelationshipsFromXML([]byte(xmlData), source)
	if err != nil {
		t.Fatalf("ParseRelationshipsFromXML failed: %v", err)
	}

	if rels.Count() != 1 {
		t.Errorf("Count = %d, want 1", rels.Count())
	}
}

// ===== Concurrent ID allocation tests =====

func TestRelationships_ConcurrentIDAllocation(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)

	// add relationships concurrently
	const goroutines = 10
	const relationsPerGoroutine = 100

	var wg sync.WaitGroup
	rIDs := make([][]string, goroutines)

	for i := 0; i < goroutines; i++ {
		rIDs[i] = make([]string, 0, relationsPerGoroutine)
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < relationsPerGoroutine; j++ {
				rel, err := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide.xml", false)
				if err != nil {
					t.Errorf("goroutine %d: AddNew failed: %v", idx, err)
					return
				}
				rIDs[idx] = append(rIDs[idx], rel.RID())
			}
		}(i)
	}

	wg.Wait()

	// verify total relationship count
	if rels.Count() != goroutines*relationsPerGoroutine {
		t.Errorf("Count = %d, want %d", rels.Count(), goroutines*relationsPerGoroutine)
	}

	// verify all IDs are unique
	idSet := make(map[string]bool)
	for _, ids := range rIDs {
		for _, id := range ids {
			if idSet[id] {
				t.Errorf("duplicate rID found: %s", id)
			}
			idSet[id] = true
		}
	}

	// verify ID format is correct
	for id := range idSet {
		if !strings.HasPrefix(id, "rId") {
			t.Errorf("invalid rID format: %s", id)
		}
	}
}

func TestRelationships_InitRIDCounter(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")

	// load a set of relationships with existing IDs from XML
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide1.xml"/>
  <Relationship Id="rId5" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide2.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide3.xml"/>
</Relationships>`

	rels := opc.NewRelationships(source)
	err := rels.FromXML([]byte(xmlData))
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// FromXML automatically initializes the counter to the max value 5.
	// NextRID previews the next value (load + 1) without consuming the counter.
	nextRID := rels.NextRID()
	if nextRID != "rId6" {
		t.Errorf("NextRID after InitRIDCounter = %s, want rId6", nextRID)
	}

	// AddNew allocates the next ID (uses Add, returns the new value).
	// The first AddNew should get rId6.
	rel, err := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide4.xml", false)
	if err != nil {
		t.Fatalf("AddNew failed: %v", err)
	}
	if rel.RID() != "rId6" {
		t.Errorf("new relationship RID = %s, want rId6", rel.RID())
	}

	// The second AddNew should get rId7.
	rel2, err := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide5.xml", false)
	if err != nil {
		t.Fatalf("AddNew failed: %v", err)
	}
	if rel2.RID() != "rId7" {
		t.Errorf("second new relationship RID = %s, want rId7", rel2.RID())
	}
}

func TestRelationships_ClonePreservesCounter(t *testing.T) {
	source := opc.NewPackURI("/ppt/presentation.xml")
	rels := opc.NewRelationships(source)

	// add some relationships
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)
	rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide2.xml", false)

	// clone
	cloned := rels.Clone()

	// NextRID of the clone should match the original
	if rels.NextRID() != cloned.NextRID() {
		t.Errorf("cloned NextRID = %s, want %s", cloned.NextRID(), rels.NextRID())
	}

	// adding a relationship to the clone should not affect the original
	cloned.AddNew(opc.RelTypeSlide, "/ppt/slides/slide3.xml", false)
	if rels.Count() != 2 {
		t.Errorf("original count changed after clone modification")
	}
}

// TestParseRelationshipsFromFile tests parsing media relationships from a real .rels file.
func TestParseRelationshipsFromFile(t *testing.T) {
	// read a real .rels file
	data, err := os.ReadFile("../test-data/test/ppt/slides/_rels/slide4.xml.rels")
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("fixture not present; skipping (gitignored, not committed upstream): %v", err)
		}
		t.Fatalf("failed to read .rels file: %v", err)
	}

	// deserialize into a Relationships struct
	source := opc.NewPackURI("/ppt/slides/slide4.xml")
	rels, err := opc.ParseRelationshipsFromXML(data, source)

	// assert no error and non-nil result
	if err != nil {
		t.Fatalf("ParseRelationshipsFromXML failed: %v", err)
	}
	if rels == nil {
		t.Fatal("ParseRelationshipsFromXML returned nil")
	}

	// assert the parsed Relationship slice has at least one element
	allRels := rels.All()
	if len(allRels) == 0 {
		t.Fatal("parsed Relationship slice is empty")
	}
	t.Logf("parsed %d relationships", len(allRels))

	// iterate the slice and find nodes whose Type contains "image" or "media"
	var foundMediaRel bool
	for _, rel := range allRels {
		relType := rel.Type()
		// check whether the type contains "image" or "media"
		if strings.Contains(relType, "image") || strings.Contains(relType, "media") {
			foundMediaRel = true

			// assert Id (rId) is non-empty
			rID := rel.RID()
			if rID == "" {
				t.Error("media relationship Id (rId) is empty")
			} else {
				t.Logf("found media relationship: Id=%s", rID)
			}

			// assert Target is non-empty
			target := rel.TargetURI()
			if target == nil || target.URI() == "" {
				t.Error("media relationship Target is empty")
			} else {
				t.Logf("  Target=%s", target.URI())
			}

			// assert Type is non-empty
			if relType == "" {
				t.Error("media relationship Type is empty")
			} else {
				t.Logf("  Type=%s", relType)
			}
		}
	}

	// ensure at least one media relationship was found
	if !foundMediaRel {
		t.Error("no relationship with Type containing 'image' or 'media' was found")
	}
}
