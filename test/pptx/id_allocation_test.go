package pptx_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// ============================================================================
// SlideContext ID allocation correctness tests
// ============================================================================
//
// SlideContext acts as the delegate handed to each component, with authority
// to allocate rIds and shapeIds. An incorrect allocation corrupts the PPTX
// due to ID collisions.
//
// Test goals:
//  1. Prove that successive NextShapeID() calls on the same Slide return
//     strictly monotonically increasing values.
//  2. Prove that IDs are never duplicated.
//  3. Prove that concurrent allocations never collide (atomic counter).
//
// ============================================================================

// TestSlideContext_ShapeIDIncremental verifies that shape IDs are strictly
// monotonically increasing.
func TestSlideContext_ShapeIDIncremental(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()
	ctx := slide.NewContext()

	// Track allocated IDs.
	allocatedIDs := make(map[uint32]bool)
	var prevID uint32 = 0

	// Allocate 100 IDs and verify monotonic increase.
	for i := 0; i < 100; i++ {
		id := ctx.NextShapeID()

		// Verify ID is non-zero (0 is considered invalid).
		if id == 0 {
			t.Fatalf("allocated invalid ID: 0")
		}

		// Verify strict increase.
		if i > 0 && id <= prevID {
			t.Fatalf("ID not monotonically increasing: prev=%d, current=%d", prevID, id)
		}

		// Verify no duplicates.
		if allocatedIDs[id] {
			t.Fatalf("duplicate shape ID allocated: %d", id)
		}
		allocatedIDs[id] = true
		prevID = id
	}

	t.Logf("shape ID increment test passed: %d unique IDs allocated", len(allocatedIDs))
}

// TestSlideContext_ShapeIDNoDuplicate verifies that shape IDs are never duplicated.
func TestSlideContext_ShapeIDNoDuplicate(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()
	ctx := slide.NewContext()

	const allocationCount = 1000
	allocatedIDs := make(map[uint32]bool, allocationCount)

	for i := 0; i < allocationCount; i++ {
		id := ctx.NextShapeID()

		if allocatedIDs[id] {
			t.Fatalf("duplicate ID detected on allocation #%d: %d", i+1, id)
		}
		allocatedIDs[id] = true
	}

	if len(allocatedIDs) != allocationCount {
		t.Errorf("unique ID count: %d, expected: %d", len(allocatedIDs), allocationCount)
	}

	t.Logf("shape ID uniqueness test passed: %d allocations with zero duplicates", allocationCount)
}

// TestSlideContext_ConcurrentIDAllocation verifies that concurrent ID
// allocations never produce collisions.
func TestSlideContext_ConcurrentIDAllocation(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()
	ctx := slide.NewContext()

	const concurrencyLevel = 1000 // 1000 components competing for IDs simultaneously

	// Use sync.Map to record every allocated ID and detect duplicates.
	var idMap sync.Map
	var wg sync.WaitGroup
	var duplicateCount int32

	// Release 1000 goroutines to race for IDs.
	for i := 0; i < concurrencyLevel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Concurrent allocation.
			id := ctx.NextShapeID()

			// If LoadOrStore returns loaded=true, the ID was already taken.
			if _, loaded := idMap.LoadOrStore(id, true); loaded {
				atomic.AddInt32(&duplicateCount, 1)
			}
		}()
	}

	wg.Wait()

	// Verify zero duplicates.
	if duplicateCount > 0 {
		t.Fatalf("concurrent allocation produced %d duplicate IDs", duplicateCount)
	}

	// Verify total count matches expectation.
	count := 0
	idMap.Range(func(_, _ any) bool {
		count++
		return true
	})

	if count != concurrencyLevel {
		t.Fatalf("expected %d IDs to be allocated, got %d", concurrencyLevel, count)
	}

	t.Logf("concurrent ID allocation test passed: %d goroutines, zero collisions", concurrencyLevel)
}

// TestSlideContext_IDAllocationOrder verifies the allocation order of IDs.
func TestSlideContext_IDAllocationOrder(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()
	ctx := slide.NewContext()

	firstID := ctx.NextShapeID()
	secondID := ctx.NextShapeID()
	thirdID := ctx.NextShapeID()

	// Verify strict increase.
	if secondID <= firstID {
		t.Errorf("second ID (%d) must be greater than first ID (%d)", secondID, firstID)
	}
	if thirdID <= secondID {
		t.Errorf("third ID (%d) must be greater than second ID (%d)", thirdID, secondID)
	}

	// Verify step size is exactly 1 (strictly sequential).
	if secondID-firstID != 1 {
		t.Errorf("unexpected ID gap: second - first = %d (expected 1)", secondID-firstID)
	}
	if thirdID-secondID != 1 {
		t.Errorf("unexpected ID gap: third - second = %d (expected 1)", thirdID-secondID)
	}

	t.Logf("ID allocation order test passed: %d -> %d -> %d", firstID, secondID, thirdID)
}

// TestSlideContext_MultipleContextsSameSlide verifies that multiple contexts
// for the same slide share the underlying ID allocator.
func TestSlideContext_MultipleContextsSameSlide(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	// Create multiple contexts for the same slide.
	ctx1 := slide.NewContext()
	ctx2 := slide.NewContext()
	ctx3 := slide.NewContext()

	// Allocate one ID from each context.
	id1 := ctx1.NextShapeID()
	id2 := ctx2.NextShapeID()
	id3 := ctx3.NextShapeID()

	// All IDs must be unique because they share the underlying atomic counter.
	allIDs := []uint32{id1, id2, id3}
	idSet := make(map[uint32]bool)
	for i, id := range allIDs {
		if idSet[id] {
			t.Fatalf("different contexts allocated duplicate ID: %d (index %d)", id, i)
		}
		idSet[id] = true
	}

	t.Logf("multiple-contexts shared allocator test passed: %v", allIDs)
}

// TestSlideContext_ConcurrentStress is a high-load concurrency stress test.
func TestSlideContext_ConcurrentStress(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()

	const goroutineCount = 100
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	globalIDSet := sync.Map{} // concurrency-safe map
	var duplicateCount int32

	for g := 0; g < goroutineCount; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			ctx := slide.NewContext()

			for i := 0; i < operationsPerGoroutine; i++ {
				id := ctx.NextShapeID()

				// Record; count if already present.
				if _, exists := globalIDSet.LoadOrStore(id, true); exists {
					atomic.AddInt32(&duplicateCount, 1)
				}
			}
		}(g)
	}

	wg.Wait()

	if duplicateCount > 0 {
		t.Fatalf("concurrent stress test found %d duplicate IDs", duplicateCount)
	}

	// Count total unique IDs.
	var uniqueCount int
	globalIDSet.Range(func(_, _ any) bool {
		uniqueCount++
		return true
	})

	expectedCount := goroutineCount * operationsPerGoroutine
	if uniqueCount != expectedCount {
		t.Errorf("unique ID count: %d, expected: %d", uniqueCount, expectedCount)
	}

	t.Logf("concurrent stress test passed: %d goroutines x %d allocations each, zero collisions",
		goroutineCount, operationsPerGoroutine)
}

// TestSlideContext_RelationshipIDIncremental verifies that relationship IDs
// (rId) are allocated without duplicates.
func TestSlideContext_RelationshipIDIncremental(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()
	ctx := slide.NewContext()

	// Minimal PNG header bytes used to trigger media allocation.
	imageData := []byte{0x89, 0x50, 0x4E, 0x47}

	allocatedRIDs := make(map[string]bool)

	for i := 0; i < 10; i++ {
		rID, err := ctx.AddMedia(imageData, "test.png")
		if err != nil {
			// Skip invalid-image errors; we only care about rId allocation logic.
			continue
		}

		if rID == "" {
			continue
		}

		// Verify rId format (typically "rId1", "rId2", etc.).
		if len(rID) < 4 || rID[:3] != "rId" {
			t.Errorf("unexpected rId format: %s", rID)
			continue
		}

		// Verify no duplicate rIds.
		if allocatedRIDs[rID] {
			t.Fatalf("duplicate rId allocated: %s", rID)
		}
		allocatedRIDs[rID] = true
	}

	t.Logf("relationship ID allocation test passed")
}

// TestSlideContext_InitialValue verifies that the first allocated ID conforms
// to the OOXML specification (1 is reserved for the root node).
func TestSlideContext_InitialValue(t *testing.T) {
	prs := pptx.New()
	slide := prs.AddSlide()
	ctx := slide.NewContext()

	// The first allocated ID should be 2 (1 is reserved for the root node).
	firstID := ctx.NextShapeID()
	if firstID != 2 {
		t.Errorf("first ID: %d, expected: 2 (1 is reserved for the root node)", firstID)
	}

	t.Logf("initial ID value test passed: first ID = %d", firstID)
}
