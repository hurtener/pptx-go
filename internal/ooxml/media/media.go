package media

import (
	"io"
)

// ============================================================================
// MediaResource - unified handling for images, audio, and video
// ============================================================================
//
// Design principles:
// 1. All fields are read-only (lowercase fields initialized via constructors).
// 2. Small files store []byte directly; large files use io.Reader.
// 3. Optimized for high-concurrency reads — no locking required.
// ============================================================================

// MediaType enumerates the supported media kinds.
type MediaType int8

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeImage             // image
	MediaTypeAudio             // audio
	MediaTypeVideo             // video
)

// MediaResource is a read-only representation of a media file (image, audio,
// or video) inside a PPTX package.
type MediaResource struct {
	// basic attributes
	fileName    string    // file name (e.g. image1.png, audio1.mp3)
	contentType string    // MIME type (e.g. image/png, audio/mpeg)
	mediaType   MediaType // media kind
	target      string    // full path inside the ZIP (e.g. ppt/media/image1.png)

	// data storage — one of these is set
	data     []byte    // small file: raw bytes
	dataSize int64     // data size in bytes
	reader   io.Reader // large file: lazy-loaded reader

	// association info
	rId       string // relationship ID (referenced by slide/slideLayout/slideMaster)
	extension string // file extension (e.g. .png, .mp3)
	hash      string // content hash (MD5, used for deduplication)
}

// ============================================================================
// Constructors
// ============================================================================

// NewMediaResourceFromBytes creates a MediaResource from raw bytes.
// Suitable for small files (e.g. small images).
func NewMediaResourceFromBytes(fileName, contentType, target string, data []byte) *MediaResource {
	return &MediaResource{
		fileName:    fileName,
		contentType: contentType,
		mediaType:   detectMediaType(contentType),
		target:      target,
		data:        data,
		dataSize:    int64(len(data)),
		extension:   extractExtension(fileName),
	}
}

// NewMediaResourceFromReader creates a MediaResource from an io.Reader.
// Suitable for large files (e.g. videos, large images).
func NewMediaResourceFromReader(fileName, contentType, target string, reader io.Reader, size int64) *MediaResource {
	return &MediaResource{
		fileName:    fileName,
		contentType: contentType,
		mediaType:   detectMediaType(contentType),
		target:      target,
		reader:      reader,
		dataSize:    size,
		extension:   extractExtension(fileName),
	}
}

// ============================================================================
// Getter methods
// ============================================================================

// FileName returns the file name (e.g. image1.png).
func (m *MediaResource) FileName() string { return m.fileName }

// ContentType returns the MIME type (e.g. image/png).
func (m *MediaResource) ContentType() string { return m.contentType }

// MediaType returns the media kind.
func (m *MediaResource) MediaType() MediaType { return m.mediaType }

// Target returns the full path inside the ZIP (e.g. ppt/media/image1.png).
func (m *MediaResource) Target() string { return m.target }

// Data returns the raw bytes, or nil for reader-backed resources.
func (m *MediaResource) Data() []byte { return m.data }

// DataSize returns the data size in bytes.
func (m *MediaResource) DataSize() int64 { return m.dataSize }

// Reader returns the data reader, or nil for byte-backed resources.
func (m *MediaResource) Reader() io.Reader { return m.reader }

// RID returns the relationship ID.
func (m *MediaResource) RID() string { return m.rId }

// Extension returns the file extension (e.g. .png).
func (m *MediaResource) Extension() string { return m.extension }

// HasData reports whether raw bytes are available.
func (m *MediaResource) HasData() bool { return m.data != nil }

// HasReader reports whether a reader is available.
func (m *MediaResource) HasReader() bool { return m.reader != nil }

// IsImage reports whether the media is an image.
func (m *MediaResource) IsImage() bool { return m.mediaType == MediaTypeImage }

// IsAudio reports whether the media is audio.
func (m *MediaResource) IsAudio() bool { return m.mediaType == MediaTypeAudio }

// IsVideo reports whether the media is video.
func (m *MediaResource) IsVideo() bool { return m.mediaType == MediaTypeVideo }

// ============================================================================
// Setter methods (initialization only)
// ============================================================================

// SetRID sets the relationship ID.
func (m *MediaResource) SetRID(rId string) {
	m.rId = rId
}

// SetHash sets the content hash.
func (m *MediaResource) SetHash(hash string) {
	m.hash = hash
}

// Hash returns the content hash.
func (m *MediaResource) Hash() string { return m.hash }

// ============================================================================
// Helper functions
// ============================================================================

// detectMediaType infers the MediaType from a MIME content type string.
func detectMediaType(contentType string) MediaType {
	switch {
	case isImageContentType(contentType):
		return MediaTypeImage
	case isAudioContentType(contentType):
		return MediaTypeAudio
	case isVideoContentType(contentType):
		return MediaTypeVideo
	default:
		return MediaTypeUnknown
	}
}

// isImageContentType reports whether ct is an image MIME type.
func isImageContentType(ct string) bool {
	prefixes := []string{
		"image/png",
		"image/jpeg",
		"image/gif",
		"image/bmp",
		"image/tiff",
		"image/svg+xml",
		"image/webp",
		"image/x-emf",
		"image/x-wmf",
	}
	for _, p := range prefixes {
		if ct == p {
			return true
		}
	}
	return false
}

// isAudioContentType reports whether ct is an audio MIME type.
func isAudioContentType(ct string) bool {
	prefixes := []string{
		"audio/mpeg",
		"audio/wav",
		"audio/ogg",
		"audio/aac",
		"audio/mp4",
	}
	for _, p := range prefixes {
		if ct == p {
			return true
		}
	}
	return false
}

// isVideoContentType reports whether ct is a video MIME type.
func isVideoContentType(ct string) bool {
	prefixes := []string{
		"video/mp4",
		"video/webm",
		"video/ogg",
		"video/quicktime",
		"video/x-msvideo",
		"video/x-ms-wmv",
	}
	for _, p := range prefixes {
		if ct == p {
			return true
		}
	}
	return false
}

// extractExtension extracts the file extension from a file name.
func extractExtension(fileName string) string {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[i:]
		}
	}
	return ""
}

// ============================================================================
// MediaType String method
// ============================================================================

// String returns a human-readable name for the media type.
func (mt MediaType) String() string {
	switch mt {
	case MediaTypeImage:
		return "image"
	case MediaTypeAudio:
		return "audio"
	case MediaTypeVideo:
		return "video"
	default:
		return "unknown"
	}
}
