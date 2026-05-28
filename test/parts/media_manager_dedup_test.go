package parts_test

import (
	"testing"

	"github.com/hurtener/pptx-go/slide"
)

// ============================================================================
// MediaManager deduplication tests
// ============================================================================

func TestMediaManager_Deduplication(t *testing.T) {
	mgr := slide.NewMediaManager()

	// First add.
	rID1, res1 := mgr.AddMediaAuto("company_logo.png", []byte("logo_data"))
	if rID1 != "rId1" {
		t.Errorf("first add rID = %q, want rId1", rID1)
	}

	// Second add with the same filename and the same data.
	rID2, res2 := mgr.AddMediaAuto("company_logo.png", []byte("logo_data"))

	// The second add must return the same rID.
	if rID2 != rID1 {
		t.Errorf("duplicate add rID = %q, want %q (should reuse existing resource)", rID2, rID1)
	}

	// The returned resource object must be identical.
	if res2 != res1 {
		t.Error("duplicate add should return the same resource object")
	}

	// The resource pool must still contain exactly one entry.
	if mgr.Count() != 1 {
		t.Errorf("Count = %d, want 1 (dedup failed)", mgr.Count())
	}
}

func TestMediaManager_Deduplication_Multiple(t *testing.T) {
	mgr := slide.NewMediaManager()

	// Simulate the same logo appearing on every slide.
	const pages = 100
	for i := 0; i < pages; i++ {
		mgr.AddMediaAuto("company_logo.png", []byte("logo_data"))
	}

	// Only one copy must be stored.
	if mgr.Count() != 1 {
		t.Errorf("after adding the same file %d times, Count = %d, want 1", pages, mgr.Count())
	}
}

func TestMediaManager_Deduplication_DifferentFiles(t *testing.T) {
	mgr := slide.NewMediaManager()

	// Add different files.
	rID1, _ := mgr.AddMediaAuto("logo.png", []byte("data1"))
	rID2, _ := mgr.AddMediaAuto("banner.png", []byte("data2"))
	rID3, _ := mgr.AddMediaAuto("icon.png", []byte("data3"))

	// Different files must have different rIDs.
	if rID1 == rID2 || rID2 == rID3 || rID1 == rID3 {
		t.Error("different files must not share an rID")
	}

	// Count must be 3.
	if mgr.Count() != 3 {
		t.Errorf("Count = %d, want 3", mgr.Count())
	}

	// Re-add an existing file.
	rID1Again, _ := mgr.AddMediaAuto("logo.png", []byte("data1"))

	// Must reuse the existing rID.
	if rID1Again != rID1 {
		t.Errorf("re-adding logo.png returned rID = %q, want %q", rID1Again, rID1)
	}

	// Count must still be 3.
	if mgr.Count() != 3 {
		t.Errorf("after dedup Count = %d, want 3", mgr.Count())
	}
}

func TestMediaManager_Deduplication_SameNameDifferentData(t *testing.T) {
	mgr := slide.NewMediaManager()

	// First add.
	rID1, res1 := mgr.AddMediaAuto("image.png", []byte("original_data"))

	// Second add: same filename, different content (hash-based dedup creates a new resource).
	rID2, res2 := mgr.AddMediaAuto("image.png", []byte("different_data"))

	// Different content must produce different rIDs.
	if rID2 == rID1 {
		t.Errorf("different content must not share an rID, got %q == %q", rID2, rID1)
	}

	// Each resource must carry its own data.
	if string(res1.Data()) != "original_data" {
		t.Errorf("res1.Data = %q, want original_data", res1.Data())
	}
	if string(res2.Data()) != "different_data" {
		t.Errorf("res2.Data = %q, want different_data", res2.Data())
	}

	// Count must be 2.
	if mgr.Count() != 2 {
		t.Errorf("Count = %d, want 2", mgr.Count())
	}
}

func TestMediaManager_Deduplication_DifferentNameSameContent(t *testing.T) {
	mgr := slide.NewMediaManager()

	sameData := []byte("identical_logo_bytes")

	// Different filenames, identical content.
	rID1, res1 := mgr.AddMediaAuto("logo_v1.png", sameData)
	rID2, res2 := mgr.AddMediaAuto("logo_v2.png", sameData)
	rID3, res3 := mgr.AddMediaAuto("company_logo.png", sameData)

	// Identical content must share the same rID (hash-based dedup).
	if rID2 != rID1 {
		t.Errorf("same content with different filename should reuse rID, got rID1=%q, rID2=%q", rID1, rID2)
	}
	if rID3 != rID1 {
		t.Errorf("same content with different filename should reuse rID, got rID1=%q, rID3=%q", rID1, rID3)
	}

	// The returned resource object must be identical.
	if res2 != res1 || res3 != res1 {
		t.Error("identical content should return the same resource object")
	}

	// Count must be 1.
	if mgr.Count() != 1 {
		t.Errorf("Count = %d, want 1 (hash dedup)", mgr.Count())
	}
}

func TestMediaManager_Deduplication_HashConsistency(t *testing.T) {
	mgr := slide.NewMediaManager()

	data := []byte("test_content")

	// Add the same content multiple times under different names.
	rID1, res1 := mgr.AddMediaAuto("a.png", data)
	rID2, res2 := mgr.AddMediaAuto("b.jpg", data)
	rID3, res3 := mgr.AddMediaAuto("c.mp4", data)

	// All must return the same rID.
	if rID1 != rID2 || rID2 != rID3 {
		t.Errorf("hash dedup failed: rID1=%q, rID2=%q, rID3=%q", rID1, rID2, rID3)
	}

	// All resource objects must be identical.
	if res1 != res2 || res2 != res3 {
		t.Error("identical content should return the same resource object")
	}

	// Hash must be non-empty and consistent.
	if res1.Hash() == "" {
		t.Error("Hash must not be empty")
	}
	if res1.Hash() != res2.Hash() || res2.Hash() != res3.Hash() {
		t.Error("identical content must produce the same hash")
	}
}
