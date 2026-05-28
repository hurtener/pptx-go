package parts_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// ============================================================================
// MediaManager concurrency safety tests
// ============================================================================

func TestMediaManager_Concurrency(t *testing.T) {
	mgr := pptx.NewMediaManager()
	const goroutines = 100
	const opsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Pre-populate some media for reads.
	for i := 0; i < 20; i++ {
		rID := "rId" + strconv.Itoa(i+1)
		fileName := "preset_" + strconv.Itoa(i) + ".png"
		mgr.AddMediaWithBytes(rID, fileName, "image/png", "ppt/media/"+fileName, []byte("preset_data"))
	}

	// Launch concurrent goroutines.
	for g := 0; g < goroutines; g++ {
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < opsPerGoroutine; i++ {
				op := rand.Intn(3) // 0: Add, 1: Get, 2: Remove

				switch op {
				case 0:
					// Random add.
					n := rand.Intn(1000)
					fileName := "img_" + strconv.Itoa(n) + ".png"
					mgr.AddMediaAuto(fileName, []byte("data_"+strconv.Itoa(n)))

				case 1:
					// Random read.
					rID := "rId" + strconv.Itoa(rand.Intn(50)+1)
					mgr.GetMedia(rID)
					mgr.GetMediaByFileName("preset_" + strconv.Itoa(rand.Intn(20)) + ".png")
					mgr.HasMedia(rID)

				case 2:
					// Random remove.
					rID := "rId" + strconv.Itoa(rand.Intn(30)+1)
					mgr.RemoveMedia(rID)
				}
			}
		}(g)
	}

	wg.Wait()

	// Verify the manager is still usable.
	_ = mgr.Count()
	_ = mgr.AllMedia()
	_ = mgr.ListRIDs()
}

// ============================================================================
// Concurrent read/write separation test
// ============================================================================

func TestMediaManager_ConcurrentReadWrite(t *testing.T) {
	mgr := pptx.NewMediaManager()
	const writers = 20
	const readers = 80
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// Writer goroutines.
	for w := 0; w < writers; w++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				fileName := "writer_" + strconv.Itoa(id) + "_" + strconv.Itoa(i) + ".png"
				mgr.AddMediaAuto(fileName, []byte("data"))
			}
		}(w)
	}

	// Reader goroutines.
	for r := 0; r < readers; r++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Exercise various read methods.
				rID := "rId" + strconv.Itoa(rand.Intn(100)+1)
				mgr.GetMedia(rID)
				mgr.HasMedia(rID)
				mgr.Count()
				mgr.AllImages()
				mgr.ListFileNames()
			}
		}(r)
	}

	wg.Wait()
}

// ============================================================================
// Concurrent counter test
// ============================================================================

func TestMediaManager_ConcurrentCount(t *testing.T) {
	mgr := pptx.NewMediaManager()
	const ops = 1000

	var wg sync.WaitGroup
	wg.Add(ops)

	for i := 0; i < ops; i++ {
		go func(n int) {
			defer wg.Done()
			fileName := "count_test_" + strconv.Itoa(n) + ".png"
			// Use unique data to prevent dedup, ensuring each operation adds a new resource.
			data := []byte("data_for_count_test_" + strconv.Itoa(n))
			mgr.AddMediaAuto(fileName, data)
		}(i)
	}

	wg.Wait()

	// Final count must equal the number of operations.
	if mgr.Count() != ops {
		t.Errorf("Count = %d, want %d", mgr.Count(), ops)
	}
}
