// Package pptx provides a high-level interface for working with PPTX files.
package pptx

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/hurtener/pptx-go/parts"
)

// ============================================================================
// MediaManager - media resource manager (concurrency-safe cache + cross-slide dedup)
// ============================================================================
//
// Design principles:
//  1. Write once, read everywhere - after initialization the dominant operation is reads.
//  2. Read-optimized - uses sync.Map; reads do not block.
//  3. Multi-index - fast lookup by rID, fileName, target, or content hash.
//  4. Content deduplication - automatic dedup based on content hash; identical
//     media is stored only once.
//  5. Cross-slide references - the same media can be referenced by multiple
//     slides, each with its own independent rId.
//
// Typical use case:
//   - The same company logo is inserted on 10 slides.
//   - After hash-based dedup, the ZIP archive contains a single image1.png.
//   - Each slide still receives its own rId (e.g. rId5, rId12, rId23 …).
//   - This greatly reduces the size of the generated file.
//
// ============================================================================

// MediaManager maintains a concurrency-safe cache of all media resources in a
// PPTX file.
type MediaManager struct {
	// Primary store: rID -> *MediaResource
	byRID sync.Map

	// Secondary index: fileName -> rID (lookup by file name)
	byName sync.Map

	// Secondary index: target -> rID (lookup by path)
	byTarget sync.Map

	// Secondary index: contentHash -> rID (content deduplication)
	byHash sync.Map

	// Total resource count.
	count int64

	// Auto-increment ID counter for generating rId1, rId2, …
	nextID int64

	// Media file counter for generating image1.png, image2.png, …
	mediaFileID int64

	// Initialization guard (ensures one-time load).
	once sync.Once

	// Write mutex (used only for batch operations requiring mutual exclusion).
	writeMu sync.Mutex

	// ============================================================================
	// Cross-slide reference support
	// ============================================================================

	// Per-slide relationship map: slideIndex -> *SlideMediaIndex.
	// Each slide has its own rId namespace.
	slideRelations sync.Map // map[int]*SlideMediaIndex

	// Global media store: hash -> *MediaResource.
	// Holds deduplicated media resources.
	globalMedia sync.Map
}

// SlideMediaIndex manages the media references for a single slide.
type SlideMediaIndex struct {
	// slideIndex is the zero-based slide index.
	slideIndex int

	// localToHash maps a local rID to the global content hash.
	localToHash sync.Map

	// hashToLocal maps a global content hash to a local rID (reverse index).
	hashToLocal sync.Map

	// nextLocalID is the per-slide rID counter.
	nextLocalID int64
}

// NewSlideMediaIndex creates a new SlideMediaIndex for the given slide.
func NewSlideMediaIndex(slideIndex int) *SlideMediaIndex {
	return &SlideMediaIndex{
		slideIndex: slideIndex,
	}
}

// NewMediaManager creates a new media resource manager.
func NewMediaManager() *MediaManager {
	return &MediaManager{}
}

// ============================================================================
// Write methods
// ============================================================================

// AddMedia adds a media resource to the cache.
// Returns the resource rID; if the resource already exists the existing rID is returned.
func (m *MediaManager) AddMedia(resource *parts.MediaResource) string {
	if resource == nil {
		return ""
	}

	rID := resource.RID()
	if rID == "" {
		return ""
	}

	// Store only if not already present.
	if _, loaded := m.byRID.LoadOrStore(rID, resource); loaded {
		return rID // already exists
	}

	// Maintain secondary indexes.
	if resource.FileName() != "" {
		m.byName.Store(resource.FileName(), rID)
	}
	if resource.Target() != "" {
		m.byTarget.Store(resource.Target(), rID)
	}

	// Increment count (atomic).
	atomic.AddInt64(&m.count, 1)

	return rID
}

// AddMediaWithBytes creates a media resource from raw bytes and adds it to the cache.
func (m *MediaManager) AddMediaWithBytes(rID, fileName, contentType, target string, data []byte) *parts.MediaResource {
	resource := parts.NewMediaResourceFromBytes(fileName, contentType, target, data)
	resource.SetRID(rID)
	m.AddMedia(resource)
	return resource
}

// AddMediaWithReader creates a media resource from an io.Reader and adds it to the cache.
func (m *MediaManager) AddMediaWithReader(rID, fileName, contentType, target string, reader io.Reader, size int64) *parts.MediaResource {
	resource := parts.NewMediaResourceFromReader(fileName, contentType, target, reader, size)
	resource.SetRID(rID)
	m.AddMedia(resource)
	return resource
}

// AddMediaAuto infers the MIME type and generates an auto-incremented rID.
// If identical content already exists (based on hash), the existing resource is
// returned (deduplication). Returns the generated rID and the MediaResource.
func (m *MediaManager) AddMediaAuto(fileName string, data []byte) (string, *parts.MediaResource) {
	// Compute content hash.
	contentHash := computeHash(data)

	// Deduplication check: return existing resource if content already stored.
	if existing := m.GetMediaByHash(contentHash); existing != nil {
		return existing.RID(), existing
	}

	// Generate auto-incremented rID.
	id := atomic.AddInt64(&m.nextID, 1)
	rID := formatRID(id)

	// Infer MIME type.
	contentType := inferContentType(fileName)

	// Build target path.
	target := "ppt/media/" + fileName

	// Create and register the resource.
	resource := parts.NewMediaResourceFromBytes(fileName, contentType, target, data)
	resource.SetRID(rID)
	resource.SetHash(contentHash)

	// Add to cache.
	m.AddMedia(resource)

	// Maintain hash index.
	m.byHash.Store(contentHash, rID)

	return rID, resource
}

// computeHash computes the MD5 hash of data as a hex string.
func computeHash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// formatRID formats a numeric ID as an rId string (e.g. rId1, rId2).
func formatRID(id int64) string {
	return "rId" + strconv.FormatInt(id, 10)
}

// inferContentType infers the MIME type from the file extension.
func inferContentType(fileName string) string {
	ext := filepath.Ext(fileName)
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".bmp":
		return "image/bmp"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".avi":
		return "video/x-msvideo"
	case ".wmv":
		return "video/x-ms-wmv"
	case ".mov":
		return "video/quicktime"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".aac":
		return "audio/aac"
	case ".ogg":
		return "audio/ogg"
	default:
		return "application/octet-stream"
	}
}

// RemoveMedia removes a media resource from the cache.
// Returns true if the resource was found and removed.
func (m *MediaManager) RemoveMedia(rID string) bool {
	if rID == "" {
		return false
	}

	// Load the resource first so we can clean up the secondary indexes.
	val, ok := m.byRID.Load(rID)
	if !ok {
		return false
	}

	resource := val.(*parts.MediaResource)

	// Remove secondary indexes.
	if resource.FileName() != "" {
		m.byName.Delete(resource.FileName())
	}
	if resource.Target() != "" {
		m.byTarget.Delete(resource.Target())
	}

	// Remove from primary store.
	m.byRID.Delete(rID)

	// Decrement count (atomic).
	atomic.AddInt64(&m.count, -1)

	return true
}

// Clear removes all media resources from the cache.
func (m *MediaManager) Clear() {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	m.byRID = sync.Map{}
	m.byName = sync.Map{}
	m.byTarget = sync.Map{}
	m.byHash = sync.Map{}
	atomic.StoreInt64(&m.count, 0)
}

// ============================================================================
// Read methods (concurrency-safe, lock-free)
// ============================================================================

// GetMedia returns the media resource for the given rID.
func (m *MediaManager) GetMedia(rID string) *parts.MediaResource {
	if rID == "" {
		return nil
	}

	val, ok := m.byRID.Load(rID)
	if !ok {
		return nil
	}
	return val.(*parts.MediaResource)
}

// GetMediaByFileName returns the media resource for the given file name.
func (m *MediaManager) GetMediaByFileName(fileName string) *parts.MediaResource {
	if fileName == "" {
		return nil
	}

	ridVal, ok := m.byName.Load(fileName)
	if !ok {
		return nil
	}

	return m.GetMedia(ridVal.(string))
}

// GetMediaByTarget returns the media resource for the given target path.
func (m *MediaManager) GetMediaByTarget(target string) *parts.MediaResource {
	if target == "" {
		return nil
	}

	ridVal, ok := m.byTarget.Load(target)
	if !ok {
		return nil
	}

	return m.GetMedia(ridVal.(string))
}

// GetMediaByHash returns the media resource for the given content hash (for deduplication).
func (m *MediaManager) GetMediaByHash(hash string) *parts.MediaResource {
	if hash == "" {
		return nil
	}

	ridVal, ok := m.byHash.Load(hash)
	if !ok {
		return nil
	}

	return m.GetMedia(ridVal.(string))
}

// HasMedia reports whether a media resource with the given rID exists.
func (m *MediaManager) HasMedia(rID string) bool {
	_, ok := m.byRID.Load(rID)
	return ok
}

// HasMediaByFileName reports whether a resource with the given file name exists.
func (m *MediaManager) HasMediaByFileName(fileName string) bool {
	_, ok := m.byName.Load(fileName)
	return ok
}

// ============================================================================
// Bulk read methods
// ============================================================================

// AllMedia returns all media resources as a new slice (thread-safe).
func (m *MediaManager) AllMedia() []*parts.MediaResource {
	result := make([]*parts.MediaResource, 0, m.Count())
	m.byRID.Range(func(key, value interface{}) bool {
		result = append(result, value.(*parts.MediaResource))
		return true
	})
	return result
}

// AllMediaByType returns all media resources of the specified type.
func (m *MediaManager) AllMediaByType(mediaType parts.MediaType) []*parts.MediaResource {
	result := make([]*parts.MediaResource, 0)
	m.byRID.Range(func(key, value interface{}) bool {
		res := value.(*parts.MediaResource)
		if res.MediaType() == mediaType {
			result = append(result, res)
		}
		return true
	})
	return result
}

// AllImages returns all image resources.
func (m *MediaManager) AllImages() []*parts.MediaResource {
	return m.AllMediaByType(parts.MediaTypeImage)
}

// AllAudio returns all audio resources.
func (m *MediaManager) AllAudio() []*parts.MediaResource {
	return m.AllMediaByType(parts.MediaTypeAudio)
}

// AllVideo returns all video resources.
func (m *MediaManager) AllVideo() []*parts.MediaResource {
	return m.AllMediaByType(parts.MediaTypeVideo)
}

// ============================================================================
// Statistics
// ============================================================================

// Count returns the total number of media resources.
func (m *MediaManager) Count() int64 {
	return atomic.LoadInt64(&m.count)
}

// CountByType returns the number of media resources of the specified type.
func (m *MediaManager) CountByType(mediaType parts.MediaType) int64 {
	var count int64
	m.byRID.Range(func(key, value interface{}) bool {
		if value.(*parts.MediaResource).MediaType() == mediaType {
			count++
		}
		return true
	})
	return count
}

// CountImages returns the number of image resources.
func (m *MediaManager) CountImages() int64 {
	return m.CountByType(parts.MediaTypeImage)
}

// CountAudio returns the number of audio resources.
func (m *MediaManager) CountAudio() int64 {
	return m.CountByType(parts.MediaTypeAudio)
}

// CountVideo returns the number of video resources.
func (m *MediaManager) CountVideo() int64 {
	return m.CountByType(parts.MediaTypeVideo)
}

// ============================================================================
// List methods
// ============================================================================

// ListRIDs returns all rIDs.
func (m *MediaManager) ListRIDs() []string {
	result := make([]string, 0, m.Count())
	m.byRID.Range(func(key, value interface{}) bool {
		result = append(result, key.(string))
		return true
	})
	return result
}

// ListFileNames returns all file names.
func (m *MediaManager) ListFileNames() []string {
	result := make([]string, 0, m.Count())
	m.byName.Range(func(key, value interface{}) bool {
		result = append(result, key.(string))
		return true
	})
	return result
}

// ListTargets returns all target paths.
func (m *MediaManager) ListTargets() []string {
	result := make([]string, 0, m.Count())
	m.byTarget.Range(func(key, value any) bool {
		result = append(result, key.(string))
		return true
	})
	return result
}

// ============================================================================
// Cross-slide media reference methods
// ============================================================================
//
// Core feature: the same media can be referenced by multiple slides, each with
// its own independent rId.
//
// Typical use case:
//   - The same company logo is inserted on 10 slides.
//   - After hash-based dedup, the ZIP archive contains a single image1.png.
//   - Each slide still receives its own rId (e.g. rId5, rId12, rId23 …).
//
// ============================================================================

// AddMediaForSlide adds media for a specific slide with cross-slide deduplication.
// Returns the slide-local rId and the global media resource.
//
// Example:
//
//	// Insert logo on slide 0
//	rId1, _ := mediaManager.AddMediaForSlide(0, logoData, "logo.png")
//	// Returns: rId1="rId1", global storage: image1.png
//
//	// Insert the same logo on slide 1
//	rId2, _ := mediaManager.AddMediaForSlide(1, logoData, "logo.png")
//	// Returns: rId2="rId1" (local rId for that slide), reuses image1.png
//
//	// The final ZIP contains a single image1.png; both slides reference it
//	// via their own local rIds.
func (m *MediaManager) AddMediaForSlide(slideIndex int, data []byte, fileName string) (string, *parts.MediaResource) {
	// 1. Compute content hash.
	contentHash := computeHash(data)

	// 2. Get or create the global media resource.
	globalResource, _ := m.getOrCreateGlobalMedia(contentHash, data, fileName)

	// 3. Get or create the per-slide media index.
	slideIndexer := m.getOrCreateSlideIndex(slideIndex)

	// 4. Get or create the slide-local rId for this media.
	localRID, _ := slideIndexer.getOrCreateLocalRID(contentHash, globalResource)

	return localRID, globalResource
}

// getOrCreateGlobalMedia returns the existing global media resource for the given
// hash, or creates a new one.
func (m *MediaManager) getOrCreateGlobalMedia(contentHash string, data []byte, fileName string) (*parts.MediaResource, bool) {
	// Fast path: already exists.
	if val, ok := m.globalMedia.Load(contentHash); ok {
		return val.(*parts.MediaResource), true
	}

	// Slow path: create under lock.
	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	// Double-checked locking.
	if val, ok := m.globalMedia.Load(contentHash); ok {
		return val.(*parts.MediaResource), true
	}

	// Generate a media file name (image1.png, image2.png, …).
	mediaFileID := atomic.AddInt64(&m.mediaFileID, 1)
	ext := filepath.Ext(fileName)
	mediaFileName := "image" + strconv.FormatInt(mediaFileID, 10) + ext

	// Infer MIME type.
	contentType := inferContentType(mediaFileName)

	// Build target path.
	target := "ppt/media/" + mediaFileName

	// Create the resource.
	resource := parts.NewMediaResourceFromBytes(mediaFileName, contentType, target, data)
	resource.SetHash(contentHash)

	// Store in the global media pool.
	m.globalMedia.Store(contentHash, resource)

	// Update the legacy index for backward compatibility.
	m.byHash.Store(contentHash, contentHash)

	return resource, false
}

// getOrCreateSlideIndex returns the existing media index for the given slide, or
// creates a new one.
func (m *MediaManager) getOrCreateSlideIndex(slideIndex int) *SlideMediaIndex {
	if val, ok := m.slideRelations.Load(slideIndex); ok {
		return val.(*SlideMediaIndex)
	}

	m.writeMu.Lock()
	defer m.writeMu.Unlock()

	// Double-checked locking.
	if val, ok := m.slideRelations.Load(slideIndex); ok {
		return val.(*SlideMediaIndex)
	}

	indexer := NewSlideMediaIndex(slideIndex)
	m.slideRelations.Store(slideIndex, indexer)
	return indexer
}

// GetSlideMediaIndex returns the media index for the given slide, or nil if none exists.
func (m *MediaManager) GetSlideMediaIndex(slideIndex int) *SlideMediaIndex {
	if val, ok := m.slideRelations.Load(slideIndex); ok {
		return val.(*SlideMediaIndex)
	}
	return nil
}

// GetGlobalMediaByHash returns the global media resource for the given hash.
func (m *MediaManager) GetGlobalMediaByHash(hash string) *parts.MediaResource {
	if val, ok := m.globalMedia.Load(hash); ok {
		return val.(*parts.MediaResource)
	}
	return nil
}

// AllGlobalMedia returns all deduplicated global media resources.
func (m *MediaManager) AllGlobalMedia() []*parts.MediaResource {
	result := make([]*parts.MediaResource, 0)
	m.globalMedia.Range(func(key, value any) bool {
		result = append(result, value.(*parts.MediaResource))
		return true
	})
	return result
}

// GlobalMediaCount returns the number of deduplicated global media resources.
func (m *MediaManager) GlobalMediaCount() int64 {
	var count int64
	m.globalMedia.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// SlideCount returns the number of slides that reference at least one media resource.
func (m *MediaManager) SlideCount() int64 {
	var count int64
	m.slideRelations.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// ============================================================================
// SlideMediaIndex methods
// ============================================================================

// getOrCreateLocalRID returns the existing slide-local rId for the given hash,
// or generates a new one.
func (smi *SlideMediaIndex) getOrCreateLocalRID(contentHash string, resource *parts.MediaResource) (string, bool) {
	// Check whether this slide already references the media.
	if val, ok := smi.hashToLocal.Load(contentHash); ok {
		return val.(string), true
	}

	// Generate a new local rId.
	id := atomic.AddInt64(&smi.nextLocalID, 1)
	localRID := "rId" + strconv.FormatInt(id, 10)

	// Establish the bidirectional mapping.
	smi.hashToLocal.Store(contentHash, localRID)
	smi.localToHash.Store(localRID, contentHash)

	// Assign the slide-local rId to the resource.
	resource.SetRID(localRID)

	return localRID, false
}

// GetLocalRIDByHash returns the slide-local rId for the given content hash.
func (smi *SlideMediaIndex) GetLocalRIDByHash(hash string) string {
	if val, ok := smi.hashToLocal.Load(hash); ok {
		return val.(string)
	}
	return ""
}

// GetHashByLocalRID returns the content hash for the given slide-local rId.
func (smi *SlideMediaIndex) GetHashByLocalRID(localRID string) string {
	if val, ok := smi.localToHash.Load(localRID); ok {
		return val.(string)
	}
	return ""
}

// AllLocalRIDs returns all slide-local rIDs.
func (smi *SlideMediaIndex) AllLocalRIDs() []string {
	result := make([]string, 0)
	smi.hashToLocal.Range(func(key, value any) bool {
		result = append(result, value.(string))
		return true
	})
	return result
}

// LocalRefCount returns the number of media references for this slide.
func (smi *SlideMediaIndex) LocalRefCount() int64 {
	var count int64
	smi.hashToLocal.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// ============================================================================
// Statistics and diagnostics
// ============================================================================

// DeduplicationStats holds deduplication statistics.
type DeduplicationStats struct {
	// GlobalMediaCount is the number of unique media resources actually stored.
	GlobalMediaCount int64

	// TotalReferences is the total number of references across all slides.
	TotalReferences int64

	// SlideCount is the number of slides that reference media.
	SlideCount int64

	// SavedBytes is the estimated number of bytes saved by deduplication.
	SavedBytes int64

	// DeduplicationRate is the deduplication ratio (0.0 – 1.0).
	DeduplicationRate float64
}

// GetDeduplicationStats returns deduplication statistics.
func (m *MediaManager) GetDeduplicationStats() DeduplicationStats {
	stats := DeduplicationStats{}

	// Count global media resources.
	stats.GlobalMediaCount = m.GlobalMediaCount()

	// Count slides.
	stats.SlideCount = m.SlideCount()

	// Count total references.
	m.slideRelations.Range(func(key, value any) bool {
		smi := value.(*SlideMediaIndex)
		stats.TotalReferences += smi.LocalRefCount()
		return true
	})

	// Compute deduplication rate.
	if stats.TotalReferences > 0 {
		stats.DeduplicationRate = 1.0 - float64(stats.GlobalMediaCount)/float64(stats.TotalReferences)
	}

	// Compute bytes saved.
	m.globalMedia.Range(func(key, value any) bool {
		res := value.(*parts.MediaResource)
		// Count how many slides reference this media resource.
		refCount := int64(0)
		m.slideRelations.Range(func(k, v any) bool {
			smi := v.(*SlideMediaIndex)
			if smi.GetLocalRIDByHash(key.(string)) != "" {
				refCount++
			}
			return true
		})
		if refCount > 1 {
			stats.SavedBytes += int64(len(res.Data())) * (refCount - 1)
		}
		return true
	})

	return stats
}
