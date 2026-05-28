package parts_test

import (
	"os"
	"sync"
	"testing"

	"github.com/hurtener/pptx-go/internal/ooxml/slide"
	"github.com/hurtener/pptx-go/pptx"
)

// ============================================================================
// MasterCache concurrent-read tests
// ============================================================================

func TestMasterCache_ConcurrentRead(t *testing.T) {
	// Initialize the cache from real files.
	cache := pptx.NewMasterCache()
	masters, layouts := loadTestMasterData()
	cache.Init(masters, layouts)

	// Launch concurrent readers.
	const readers = 100
	const readsPerGoroutine = 50
	var wg sync.WaitGroup
	wg.Add(readers)

	for i := 0; i < readers; i++ {
		go func(readerID int) {
			defer wg.Done()

			for j := 0; j < readsPerGoroutine; j++ {
				// Read a layout.
				if layout, ok := cache.GetLayout("layout_1"); ok {
					_ = layout.Placeholders()
					_ = layout.PlaceholderByType(slide.PlaceholderTypeTitle)
					_ = layout.PlaceholderByType(slide.PlaceholderTypeBody)
				}

				// Read a master.
				if master, ok := cache.GetMaster("master_1"); ok {
					_ = master.Placeholders()
					_ = master.Background()
				}

				// Read a placeholder.
				if ph, ok := cache.GetPlaceholder("layout_1", "title"); ok {
					_ = ph.X()
					_ = ph.Y()
					_ = ph.Cx()
					_ = ph.Cy()
				}

				// Read statistics.
				_ = cache.LayoutCount()
				_ = cache.MasterCount()
				_ = cache.LayoutExists("layout_1")
				_ = cache.MasterExists("master_1")

				// Read lists.
				_ = cache.ListLayoutIDs()
				_ = cache.ListMasterIDs()
			}
		}(i)
	}

	wg.Wait()
}

// ============================================================================
// MasterCache high-concurrency stress test
// ============================================================================

func TestMasterCache_HighConcurrency(t *testing.T) {
	cache := pptx.NewMasterCache()
	masters, layouts := loadTestMasterData()
	cache.Init(masters, layouts)

	// 100 goroutines, each performing 100 reads.
	const totalGoroutines = 100
	const iterations = 100
	var wg sync.WaitGroup
	wg.Add(totalGoroutines)

	for i := 0; i < totalGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// Read all data concurrently.
				layout, _ := cache.GetLayout("layout_1")
				if layout != nil {
					_ = layout.PlaceholderCount()
					_ = layout.Placeholders()
				}

				master, _ := cache.GetMaster("master_1")
				if master != nil {
					_ = master.PlaceholderCount()
				}

				// Index lookups.
				if _, ok := cache.GetLayoutByName("13_Title Slide"); ok {
				}
				if _, ok := cache.GetPlaceholderByID("layout_1", "ph_8"); ok {
				}
				if _, ok := cache.GetPlaceholder("layout_1", "title"); ok {
				}

				// Bulk reads.
				_ = cache.AllLayouts()
				_ = cache.AllMasters()
			}
		}(i)
	}

	wg.Wait()
}

// ============================================================================
// MasterCache helper functions
// ============================================================================

func loadTestMasterData() ([]*slide.SlideMasterData, []*slide.SlideLayoutData) {
	var masters []*slide.SlideMasterData
	var layouts []*slide.SlideLayoutData

	// Load a master.
	masterXML, _ := os.ReadFile("../test-data/test/ppt/slideMasters/slideMaster2.xml")
	if master, err := slide.ParseMaster(masterXML); err == nil {
		masters = append(masters, master)
	}

	// Load a layout.
	layoutXML, _ := os.ReadFile("../test-data/test/ppt/slideLayouts/slideLayout5.xml")
	if layout, err := slide.ParseLayout(layoutXML); err == nil {
		layouts = append(layouts, layout)
	}

	return masters, layouts
}
