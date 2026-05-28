# Media Module — Interface Documentation

> Unified handling of images, audio, video, and other media files in PPTX

---

## Enum Types

### MediaType

Media type enum.

| Constant | Value | Description |
|------|-----|------|
| `MediaTypeUnknown` | `0` | Unknown type |
| `MediaTypeImage` | `1` | Image |
| `MediaTypeAudio` | `2` | Audio |
| `MediaTypeVideo` | `3` | Video |

#### Methods

```go
func (mt MediaType) String() string
```
Returns the string representation of the media type: `"image"` / `"audio"` / `"video"` / `"unknown"`

---

## MediaResource

Media resource struct (read-only), for unified handling of images, audio, and video.

### Fields

| Field | Type | Accessor | Description |
|------|------|--------|------|
| `fileName` | `string` | `FileName()` | File name (e.g. `image1.png`) |
| `contentType` | `string` | `ContentType()` | MIME type (e.g. `image/png`) |
| `mediaType` | `MediaType` | `MediaType()` | Media type enum |
| `target` | `string` | `Target()` | Full path within the ZIP (e.g. `ppt/media/image1.png`) |
| `data` | `[]byte` | `Data()` | Byte data for small files (may be nil) |
| `dataSize` | `int64` | `DataSize()` | Data size in bytes |
| `reader` | `io.Reader` | `Reader()` | Reader for large files (may be nil) |
| `rId` | `string` | `RID()` | Relationship ID |
| `extension` | `string` | `Extension()` | File extension (e.g. `.png`) |
| `hash` | `string` | `Hash()` | Content hash (MD5) |

### Type-check Methods

```go
func (m *MediaResource) HasData() bool      // Returns true if byte data is present
func (m *MediaResource) HasReader() bool     // Returns true if a Reader is present
func (m *MediaResource) IsImage() bool       // Returns true if the resource is an image
func (m *MediaResource) IsAudio() bool       // Returns true if the resource is audio
func (m *MediaResource) IsVideo() bool       // Returns true if the resource is video
```

### Setter Methods

```go
func (m *MediaResource) SetRID(rId string)   // Sets the relationship ID
func (m *MediaResource) SetHash(hash string)  // Sets the content hash
```

---

## Constructors

### NewMediaResourceFromBytes

```go
func NewMediaResourceFromBytes(fileName, contentType, target string, data []byte) *MediaResource
```

Creates a media resource from byte data; suitable for small files (e.g. small images).

### NewMediaResourceFromReader

```go
func NewMediaResourceFromReader(fileName, contentType, target string, reader io.Reader, size int64) *MediaResource
```

Creates a media resource from a Reader; suitable for large files (e.g. video, large images).

---

## MIME Type Helper Functions

### Supported Image Types

```go
image/png, image/jpeg, image/gif, image/bmp, image/tiff,
image/svg+xml, image/webp, image/x-emf, image/x-wmf
```

### Supported Audio Types

```go
audio/mpeg, audio/wav, audio/ogg, audio/aac, audio/mp4
```

### Supported Video Types

```go
video/mp4, video/webm, video/ogg, video/quicktime,
video/x-msvideo, video/x-ms-wmv
```

---

## MediaManager

Media resource manager (concurrency-safe cache).

### Design Principles

1. Write once, read everywhere — after initialisation the primary operation is reads.
2. Read-optimised — uses `sync.RWMutex`; read operations do not block.
3. Dual index — fast lookup by both rID and fileName.

### Index Structure

| Index | Key | Value |
|------|-----|-------|
| `byRID` | rID | `*MediaResource` |
| `byName` | fileName | rID |
| `byTarget` | target | rID |
| `byHash` | contentHash | rID |

### Global Instance

```go
func DefaultMediaManager() *MediaManager
var defaultMediaManager *MediaManager
```

### Constructor

```go
func NewMediaManager() *MediaManager
```

---

## MediaManager Write Methods

### AddMedia

```go
func (m *MediaManager) AddMedia(resource *MediaResource) string
```

Adds a media resource to the cache and returns its rID. Returns the existing rID if the resource is already present.

### AddMediaWithBytes

```go
func (m *MediaManager) AddMediaWithBytes(rID, fileName, contentType, target string, data []byte) *MediaResource
```

Adds a media resource from byte data.

### AddMediaWithReader

```go
func (m *MediaManager) AddMediaWithReader(rID, fileName, contentType, target string, reader io.Reader, size int64) *MediaResource
```

Adds a media resource from a Reader.

### AddMediaAuto

```go
func (m *MediaManager) AddMediaAuto(fileName string, data []byte) (string, *MediaResource)
```

Automatically infers the MIME type and generates an auto-incrementing rID. If the same content already exists (based on hash), returns the existing resource (deduplication).

### RemoveMedia

```go
func (m *MediaManager) RemoveMedia(rID string) bool
```

Removes a media resource.

### Clear

```go
func (m *MediaManager) Clear()
```

Clears all media resources.

---

## MediaManager Read Methods

### GetMedia

```go
func (m *MediaManager) GetMedia(rID string) *MediaResource
```

Retrieves a media resource by rID.

### GetMediaByFileName

```go
func (m *MediaManager) GetMediaByFileName(fileName string) *MediaResource
```

Retrieves a media resource by file name.

### GetMediaByTarget

```go
func (m *MediaManager) GetMediaByTarget(target string) *MediaResource
```

Retrieves a media resource by target path.

### GetMediaByHash

```go
func (m *MediaManager) GetMediaByHash(hash string) *MediaResource
```

Retrieves a media resource by content hash (used for deduplication).

### HasMedia

```go
func (m *MediaManager) HasMedia(rID string) bool
```

Checks whether a media resource exists.

### HasMediaByFileName

```go
func (m *MediaManager) HasMediaByFileName(fileName string) bool
```

Checks whether a file name exists.

---

## MediaManager Bulk Read Methods

### AllMedia

```go
func (m *MediaManager) AllMedia() []*MediaResource
```

Returns all media resources.

### AllMediaByType

```go
func (m *MediaManager) AllMediaByType(mediaType MediaType) []*MediaResource
```

Returns all media resources of the specified type.

### AllImages

```go
func (m *MediaManager) AllImages() []*MediaResource
```

Returns all image resources.

### AllAudio

```go
func (m *MediaManager) AllAudio() []*MediaResource
```

Returns all audio resources.

### AllVideo

```go
func (m *MediaManager) AllVideo() []*MediaResource
```

Returns all video resources.

---

## MediaManager Statistics Methods

### Count

```go
func (m *MediaManager) Count() int64
```

Returns the total number of media resources.

### CountByType

```go
func (m *MediaManager) CountByType(mediaType MediaType) int64
```

Returns the number of media resources of the specified type.

### CountImages

```go
func (m *MediaManager) CountImages() int64
```

Returns the number of images.

### CountAudio

```go
func (m *MediaManager) CountAudio() int64
```

Returns the number of audio resources.

### CountVideo

```go
func (m *MediaManager) CountVideo() int64
```

Returns the number of video resources.

---

## MediaManager List Methods

### ListRIDs

```go
func (m *MediaManager) ListRIDs() []string
```

Returns all rIDs.

### ListFileNames

```go
func (m *MediaManager) ListFileNames() []string
```

Returns all file names.

### ListTargets

```go
func (m *MediaManager) ListTargets() []string
```

Returns all target paths.

---

## Global Convenience Functions

```go
func AddMedia(resource *MediaResource) string
func GetMedia(rID string) *MediaResource
func GetMediaByFileName(fileName string) *MediaResource
func GetMediaByTarget(target string) *MediaResource
func ClearMedia()
```

These operate on the global default manager `defaultMediaManager`.
