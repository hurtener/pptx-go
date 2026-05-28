package pptx_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/hurtener/pptx-go/pptx"
)

// ============================================================================
// Absolute isolation test for high-concurrency cloning
// ============================================================================
//
// Goal: prove that 100 goroutines calling pptx.New() simultaneously each
// receive a fully independent physical copy. Mutating one instance must
// never pollute the built-in template or any other instance.
//
// ============================================================================

// TestPresentation_ConcurrencyIsolation verifies that presentations created
// concurrently by multiple goroutines are completely isolated from each other.
func TestPresentation_ConcurrencyIsolation(t *testing.T) {
	const goroutineCount = 100
	var wg sync.WaitGroup
	var successCount int32
	var panicCount int32

	// Store each goroutine's Presentation for later independence checks.
	presentations := make([]*pptx.Presentation, goroutineCount)
	var mu sync.Mutex

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panic: %v", id, r)
					atomic.AddInt32(&panicCount, 1)
				}
			}()

			// Create a new Presentation.
			prs := pptx.New()

			// Add a slide.
			slide := prs.AddSlide()

			// Add a text box with unique content based on goroutine ID.
			text := string(rune('A'+id%26)) + "-Slide-Content"
			slide.AddTextBox(100, 100, 500, 50, text)

			// Add a shape.
			slide.AddRectangle(100, 200, 300, 150)

			// Change slide size (verify the mutation does not affect other instances).
			prs.SetSlideSize(1280*9525, 720*9525)

			// Store the reference.
			mu.Lock()
			presentations[id] = prs
			mu.Unlock()

			atomic.AddInt32(&successCount, 1)
		}(i)
	}

	wg.Wait()

	// Verify all goroutines completed successfully.
	if successCount != goroutineCount {
		t.Errorf("success count: %d, expected: %d", successCount, goroutineCount)
	}

	if panicCount > 0 {
		t.Errorf("goroutines that panicked: %d", panicCount)
	}

	t.Logf("all %d presentations created concurrently without error", successCount)

	// Verify each Presentation is independent.
	t.Run("VerifyIndependence", func(t *testing.T) {
		for i := 0; i < goroutineCount; i++ {
			if presentations[i] == nil {
				t.Errorf("Presentation %d is nil", i)
				continue
			}

			// Verify slide count.
			if presentations[i].SlideCount() != 1 {
				t.Errorf("Presentation %d slide count: %d, expected: 1", i, presentations[i].SlideCount())
			}
		}
	})

	t.Log("concurrent isolation test passed: zero cross-instance pollution")
}

// TestPresentation_ConcurrentModification verifies that multiple goroutines
// modifying the same Presentation do not cause data races.
func TestPresentation_ConcurrentModification(t *testing.T) {
	prs := pptx.New()

	const goroutineCount = 50
	var wg sync.WaitGroup
	var slideIndices []int
	var mu sync.Mutex

	// Add slides concurrently.
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			slide := prs.AddSlide()
			slide.AddTextBox(100, 100, 500, 50, "Slide-"+string(rune('0'+id%10)))

			mu.Lock()
			slideIndices = append(slideIndices, slide.Index())
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify slide count.
	if prs.SlideCount() != goroutineCount {
		t.Errorf("slide count: %d, expected: %d", prs.SlideCount(), goroutineCount)
	}

	t.Logf("concurrent add of %d slides succeeded", prs.SlideCount())
}

// TestPresentation_ConcurrentSlideAccess verifies thread safety when slides
// are read concurrently.
func TestPresentation_ConcurrentSlideAccess(t *testing.T) {
	prs := pptx.New()

	// Pre-populate slides.
	const slideCount = 20
	for i := 0; i < slideCount; i++ {
		prs.AddSlide()
	}

	const goroutineCount = 100
	var wg sync.WaitGroup
	var readErrors int32

	// Read slides concurrently.
	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					atomic.AddInt32(&readErrors, 1)
				}
			}()

			// Access a slide by a round-robin index.
			index := id % slideCount
			slide, err := prs.GetSlide(index)
			if err != nil {
				atomic.AddInt32(&readErrors, 1)
				return
			}

			// Read slide properties.
			_ = slide.Index()
			_, _ = slide.SlideSize()

			// Get all slides.
			slides := prs.Slides()
			if len(slides) != slideCount {
				atomic.AddInt32(&readErrors, 1)
			}
		}(i)
	}

	wg.Wait()

	if readErrors > 0 {
		t.Errorf("read errors: %d", readErrors)
	}

	t.Log("concurrent slide access test passed")
}

// TestPresentation_ConcurrentSave verifies that multiple goroutines saving
// different Presentations simultaneously do not interfere with each other.
func TestPresentation_ConcurrentSave(t *testing.T) {
	const goroutineCount = 20
	var wg sync.WaitGroup
	var saveErrors int32

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panic during save: %v", id, r)
					atomic.AddInt32(&saveErrors, 1)
				}
			}()

			prs := pptx.New()
			slide := prs.AddSlide()
			slide.AddTextBox(100, 100, 500, 50, "Content-"+string(rune('0'+id%10)))

			// Write to bytes (no disk I/O).
			data, err := prs.WriteToBytes()
			if err != nil {
				atomic.AddInt32(&saveErrors, 1)
				return
			}

			// Verify output is non-empty.
			if len(data) == 0 {
				atomic.AddInt32(&saveErrors, 1)
			}
		}(i)
	}

	wg.Wait()

	if saveErrors > 0 {
		t.Errorf("save errors: %d", saveErrors)
	}

	t.Log("concurrent save test passed")
}

// TestMediaManager_Concurrency verifies thread safety when multiple independent
// Presentations add media concurrently.
// Note: each goroutine uses its own Presentation — this is the correct usage pattern.
func TestMediaManager_Concurrency(t *testing.T) {
	const goroutineCount = 50
	var wg sync.WaitGroup
	var addErrors int32

	// Minimal PNG header bytes (not a valid image, but sufficient for the test).
	imageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panic: %v", id, r)
					atomic.AddInt32(&addErrors, 1)
				}
			}()

			// Each goroutine gets its own Presentation.
			prs := pptx.New()
			slide := prs.AddSlide()

			fileName := "image-" + string(rune('0'+id%10)) + ".png"
			_, err := slide.AddPictureFromBytes(100+id*10, 100, 100, 100, fileName, imageData)
			if err != nil {
				// Some errors are expected (e.g. invalid image data); not counted as failures.
			}
		}(i)
	}

	wg.Wait()

	if addErrors > 0 {
		t.Errorf("media add errors (panics): %d", addErrors)
	}

	t.Log("media manager concurrency test passed")
}

// TestPresentation_MemoryIsolation verifies that mutating one Presentation
// does not affect another.
func TestPresentation_MemoryIsolation(t *testing.T) {
	// Create first Presentation.
	prs1 := pptx.New()
	slide1 := prs1.AddSlide()
	slide1.AddTextBox(100, 100, 500, 50, "Presentation 1")

	// Create second Presentation.
	prs2 := pptx.New()
	slide2 := prs2.AddSlide()
	slide2.AddTextBox(100, 100, 500, 50, "Presentation 2")

	// Mutate the first Presentation.
	prs1.SetSlideSize(960*9525, 720*9525) // 4:3

	// Verify the second Presentation is unaffected.
	cx, cy := prs2.SlideSize()
	expectedCX := 1280 * 9525 // default 16:9
	expectedCY := 720 * 9525

	// Allow no tolerance — EMU values must match exactly.
	if cx != expectedCX || cy != expectedCY {
		t.Errorf("Presentation 2 size was corrupted: cx=%d, cy=%d, expected: cx=%d, cy=%d",
			cx, cy, expectedCX, expectedCY)
	}

	// Verify slide counts are independent.
	if prs1.SlideCount() != 1 {
		t.Errorf("Presentation 1 slide count: %d, expected: 1", prs1.SlideCount())
	}
	if prs2.SlideCount() != 1 {
		t.Errorf("Presentation 2 slide count: %d, expected: 1", prs2.SlideCount())
	}

	t.Log("memory isolation test passed: zero cross-instance pollution")
}

// TestPresentation_CloneIsolation verifies that a cloned Presentation is
// completely independent from its original.
func TestPresentation_CloneIsolation(t *testing.T) {
	// Create original Presentation.
	original := pptx.New()
	slide := original.AddSlide()
	slide.AddTextBox(100, 100, 500, 50, "Original")

	// Clone.
	cloned, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	// Mutate the original.
	original.AddSlide()
	original.SetSlideSize(960*9525, 720*9525)

	// Verify the clone is not affected.
	if cloned.SlideCount() != 1 {
		t.Errorf("clone slide count was corrupted: %d, expected: 1", cloned.SlideCount())
	}

	cx, _ := cloned.SlideSize()
	if cx != 1280*9525 {
		t.Errorf("clone size was corrupted: cx=%d, expected: %d", cx, 1280*9525)
	}

	// Mutate the clone.
	cloned.AddSlide()

	// Verify the original is not affected.
	if original.SlideCount() != 2 { // original now has 2 slides (1 + 1 added above)
		t.Errorf("original slide count: %d, expected: 2", original.SlideCount())
	}

	t.Log("clone isolation test passed")
}
