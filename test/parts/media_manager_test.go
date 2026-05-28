package parts_test

import (
	"testing"

	"github.com/hurtener/pptx-go/parts"
	"github.com/hurtener/pptx-go/slide"
)

// ============================================================================
// MediaManager basic CRUD tests
// ============================================================================

func TestMediaManager_AddAndGet(t *testing.T) {
	// Initialise an empty MediaManager.
	mgr := slide.NewMediaManager()

	// Add an image.
	rID1, img := mgr.AddMediaAuto("logo.png", []byte("fake_image_data"))
	if rID1 != "rId1" {
		t.Errorf("first add rID = %q, want %q", rID1, "rId1")
	}
	if img.ContentType() != "image/png" {
		t.Errorf("image ContentType = %q, want %q", img.ContentType(), "image/png")
	}
	if img.MediaType() != parts.MediaTypeImage {
		t.Errorf("image MediaType = %v, want %v", img.MediaType(), parts.MediaTypeImage)
	}

	// Add a video.
	rID2, video := mgr.AddMediaAuto("demo.mp4", []byte("fake_video_data"))
	if rID2 != "rId2" {
		t.Errorf("second add rID = %q, want %q", rID2, "rId2")
	}
	if video.ContentType() != "video/mp4" {
		t.Errorf("video ContentType = %q, want %q", video.ContentType(), "video/mp4")
	}
	if video.MediaType() != parts.MediaTypeVideo {
		t.Errorf("video MediaType = %v, want %v", video.MediaType(), parts.MediaTypeVideo)
	}

	// Verify GetMedia retrieves the correct resource.
	retrieved := mgr.GetMedia("rId1")
	if retrieved == nil {
		t.Fatal("GetMedia(\"rId1\") returned nil")
	}
	if string(retrieved.Data()) != "fake_image_data" {
		t.Errorf("Data = %q, want %q", string(retrieved.Data()), "fake_image_data")
	}
	if retrieved.FileName() != "logo.png" {
		t.Errorf("FileName = %q, want %q", retrieved.FileName(), "logo.png")
	}

	// Verify count.
	if mgr.Count() != 2 {
		t.Errorf("Count = %d, want 2", mgr.Count())
	}
}

// ============================================================================
// MIME type inference tests
// ============================================================================

func TestMediaManager_ContentTypeInference(t *testing.T) {
	mgr := slide.NewMediaManager()

	tests := []struct {
		fileName        string
		wantContentType string
		wantMediaType   parts.MediaType
	}{
		{"image.png", "image/png", parts.MediaTypeImage},
		{"photo.jpg", "image/jpeg", parts.MediaTypeImage},
		{"photo.jpeg", "image/jpeg", parts.MediaTypeImage},
		{"animation.gif", "image/gif", parts.MediaTypeImage},
		{"icon.bmp", "image/bmp", parts.MediaTypeImage},
		{"modern.webp", "image/webp", parts.MediaTypeImage},
		{"video.mp4", "video/mp4", parts.MediaTypeVideo},
		{"clip.webm", "video/webm", parts.MediaTypeVideo},
		{"movie.avi", "video/x-msvideo", parts.MediaTypeVideo},
		{"audio.mp3", "audio/mpeg", parts.MediaTypeAudio},
		{"sound.wav", "audio/wav", parts.MediaTypeAudio},
		{"music.aac", "audio/aac", parts.MediaTypeAudio},
		{"unknown.xyz", "application/octet-stream", parts.MediaTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			// Use unique data per file to avoid dedup interference with type inference.
			data := []byte("data_for_" + tt.fileName)
			_, res := mgr.AddMediaAuto(tt.fileName, data)
			if res.ContentType() != tt.wantContentType {
				t.Errorf("ContentType = %q, want %q", res.ContentType(), tt.wantContentType)
			}
			if res.MediaType() != tt.wantMediaType {
				t.Errorf("MediaType = %v, want %v", res.MediaType(), tt.wantMediaType)
			}
		})
	}
}

// ============================================================================
// Index lookup tests
// ============================================================================

func TestMediaManager_GetByIndex(t *testing.T) {
	mgr := slide.NewMediaManager()

	// Add media.
	mgr.AddMediaAuto("logo.png", []byte("img_data"))
	mgr.AddMediaAuto("sound.mp3", []byte("audio_data"))

	t.Run("GetMediaByFileName", func(t *testing.T) {
		res := mgr.GetMediaByFileName("logo.png")
		if res == nil {
			t.Fatal("GetMediaByFileName returned nil")
		}
		if res.ContentType() != "image/png" {
			t.Errorf("ContentType = %q, want image/png", res.ContentType())
		}
	})

	t.Run("GetMediaByTarget", func(t *testing.T) {
		res := mgr.GetMediaByTarget("ppt/media/logo.png")
		if res == nil {
			t.Fatal("GetMediaByTarget returned nil")
		}
		if res.RID() != "rId1" {
			t.Errorf("RID = %q, want rId1", res.RID())
		}
	})

	t.Run("HasMedia", func(t *testing.T) {
		if !mgr.HasMedia("rId1") {
			t.Error("HasMedia(\"rId1\") = false, want true")
		}
		if mgr.HasMedia("rId999") {
			t.Error("HasMedia(\"rId999\") = true, want false")
		}
	})

	t.Run("HasMediaByFileName", func(t *testing.T) {
		if !mgr.HasMediaByFileName("logo.png") {
			t.Error("HasMediaByFileName(\"logo.png\") = false, want true")
		}
	})
}

// ============================================================================
// Remove and Clear tests
// ============================================================================

func TestMediaManager_RemoveAndClear(t *testing.T) {
	t.Run("RemoveMedia", func(t *testing.T) {
		mgr := slide.NewMediaManager()
		mgr.AddMediaAuto("test.png", []byte("data"))

		if !mgr.RemoveMedia("rId1") {
			t.Error("RemoveMedia returned false")
		}
		if mgr.HasMedia("rId1") {
			t.Error("HasMedia still returns true after removal")
		}
		if mgr.Count() != 0 {
			t.Errorf("Count after removal = %d, want 0", mgr.Count())
		}
	})

	t.Run("Clear", func(t *testing.T) {
		mgr := slide.NewMediaManager()
		mgr.AddMediaAuto("a.png", []byte("a"))
		mgr.AddMediaAuto("b.mp3", []byte("b"))
		mgr.AddMediaAuto("c.mp4", []byte("c"))

		mgr.Clear()
		if mgr.Count() != 0 {
			t.Errorf("Count after Clear = %d, want 0", mgr.Count())
		}
		if mgr.HasMedia("rId1") {
			t.Error("HasMedia still returns true after Clear")
		}
	})
}
