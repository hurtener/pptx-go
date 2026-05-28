# Media - Media Management

The media manager maintains a concurrency-safe cache of all media assets in a PPTX file, with automatic deduplication for images, audio, video, and other resources.

## MediaManager

Media asset manager.

```go
type MediaManager struct {
    // Has unexported fields.
}
```

### Constructor

```go
func NewMediaManager() *MediaManager
```

### Basic Operations

#### AddMedia

Adds a media asset to the cache and returns its rID. If the asset already exists, the existing rID is returned.

```go
func (m *MediaManager) AddMedia(resource *parts.MediaResource) string
```

#### AddMediaAuto

Automatically infers the MIME type and generates an auto-incremented rID. If the same content already exists (based on its hash), the existing resource is returned (deduplication).

```go
func (m *MediaManager) AddMediaAuto(fileName string, data []byte) (string, *parts.MediaResource)
```

**Parameters:**
- `fileName`: file name (used to infer the MIME type)
- `data`: media data

**Returns:**
- The generated rID and the created MediaResource

**Example:**

```go
mm := pptx.NewMediaManager()
data, _ := os.ReadFile("logo.png")

rId, resource := mm.AddMediaAuto("logo.png", data)
fmt.Printf("Added media: rId=%s, type=%s\n", rId, resource.ContentType)
```

#### AddMediaForSlide

Adds media for a specific slide (supports cross-slide deduplication).

```go
func (m *MediaManager) AddMediaForSlide(slideIndex int, data []byte, fileName string) (string, *parts.MediaResource)
```

**Parameters:**
- `slideIndex`: slide index
- `data`: media data
- `fileName`: file name

**Returns:**
- The slide-local rId and the global media resource

**Example:**

```go
// Insert logo on slide 1
rId1, _ := mm.AddMediaForSlide(0, logoData, "logo.png")
// Returns: rId1="rId1", stores image1.png globally

// Insert the same logo on slide 2
rId2, _ := mm.AddMediaForSlide(1, logoData, "logo.png")
// Returns: rId2="rId1" (local rId for that slide), reuses image1.png

// The final ZIP contains only one copy of image1.png,
// but both slides have their own rId references
```

#### AddMediaWithBytes

Adds a media asset from byte data.

```go
func (m *MediaManager) AddMediaWithBytes(rID, fileName, contentType, target string, data []byte) *parts.MediaResource
```

**Parameters:**
- `rID`: relationship ID
- `fileName`: file name
- `contentType`: MIME type
- `target`: target path
- `data`: media data

#### AddMediaWithReader

Adds a media asset from a Reader.

```go
func (m *MediaManager) AddMediaWithReader(rID, fileName, contentType, target string, reader io.Reader, size int64) *parts.MediaResource
```

### Query Operations

#### GetMedia

Retrieves a media asset by rID.

```go
func (m *MediaManager) GetMedia(rID string) *parts.MediaResource
```

#### GetMediaByFileName

Retrieves a media asset by file name.

```go
func (m *MediaManager) GetMediaByFileName(fileName string) *parts.MediaResource
```

#### GetMediaByHash

Retrieves a media asset by content hash (used for deduplication).

```go
func (m *MediaManager) GetMediaByHash(hash string) *parts.MediaResource
```

#### GetMediaByTarget

Retrieves a media asset by target path.

```go
func (m *MediaManager) GetMediaByTarget(target string) *parts.MediaResource
```

#### GetGlobalMediaByHash

Retrieves a global media asset by hash.

```go
func (m *MediaManager) GetGlobalMediaByHash(hash string) *parts.MediaResource
```

#### GetSlideMediaIndex

Returns the media index for a slide.

```go
func (m *MediaManager) GetSlideMediaIndex(slideIndex int) *SlideMediaIndex
```

### List Operations

#### AllMedia

Returns all media assets (returns a new slice; thread-safe).

```go
func (m *MediaManager) AllMedia() []*parts.MediaResource
```

#### AllGlobalMedia

Returns all global media assets (after deduplication).

```go
func (m *MediaManager) AllGlobalMedia() []*parts.MediaResource
```

#### AllImages

Returns all image assets.

```go
func (m *MediaManager) AllImages() []*parts.MediaResource
```

#### AllVideo

Returns all video assets.

```go
func (m *MediaManager) AllVideo() []*parts.MediaResource
```

#### AllAudio

Returns all audio assets.

```go
func (m *MediaManager) AllAudio() []*parts.MediaResource
```

#### AllMediaByType

Returns all media assets of the specified type.

```go
func (m *MediaManager) AllMediaByType(mediaType parts.MediaType) []*parts.MediaResource
```

#### ListRIDs

Returns all rIDs.

```go
func (m *MediaManager) ListRIDs() []string
```

#### ListFileNames

Returns all file names.

```go
func (m *MediaManager) ListFileNames() []string
```

#### ListTargets

Returns all target paths.

```go
func (m *MediaManager) ListTargets() []string
```

### Statistics Operations

#### Count

Returns the total number of media assets.

```go
func (m *MediaManager) Count() int64
```

#### GlobalMediaCount

Returns the number of global media assets (after deduplication).

```go
func (m *MediaManager) GlobalMediaCount() int64
```

#### CountImages

Returns the number of image assets.

```go
func (m *MediaManager) CountImages() int64
```

#### CountVideo

Returns the number of video assets.

```go
func (m *MediaManager) CountVideo() int64
```

#### CountAudio

Returns the number of audio assets.

```go
func (m *MediaManager) CountAudio() int64
```

#### CountByType

Returns the number of media assets of the specified type.

```go
func (m *MediaManager) CountByType(mediaType parts.MediaType) int64
```

#### SlideCount

Returns the number of slides that reference media.

```go
func (m *MediaManager) SlideCount() int64
```

### Other Operations

#### HasMedia

Checks whether a media asset exists.

```go
func (m *MediaManager) HasMedia(rID string) bool
```

#### HasMediaByFileName

Checks whether a file name exists.

```go
func (m *MediaManager) HasMediaByFileName(fileName string) bool
```

#### RemoveMedia

Removes a media asset.

```go
func (m *MediaManager) RemoveMedia(rID string) bool
```

#### Clear

Clears all media assets.

```go
func (m *MediaManager) Clear()
```

### Deduplication Statistics

#### GetDeduplicationStats

Returns deduplication statistics.

```go
func (m *MediaManager) GetDeduplicationStats() DeduplicationStats
```

## DeduplicationStats

Deduplication statistics.

```go
type DeduplicationStats struct {
    // GlobalMediaCount is the number of globally stored media assets (actual storage)
    GlobalMediaCount int64

    // TotalReferences is the total number of references across all slides
    TotalReferences int64

    // SlideCount is the number of slides
    SlideCount int64

    // SavedBytes is the storage saved by deduplication (in bytes)
    SavedBytes int64

    // DeduplicationRate is the deduplication ratio (0.0 - 1.0)
    DeduplicationRate float64
}
```

**Example:**

```go
stats := mm.GetDeduplicationStats()
fmt.Printf("Global media count: %d\n", stats.GlobalMediaCount)
fmt.Printf("Total references: %d\n", stats.TotalReferences)
fmt.Printf("Bytes saved: %d\n", stats.SavedBytes)
fmt.Printf("Deduplication rate: %.2f%%\n", stats.DeduplicationRate*100)
```

## SlideMediaIndex

A slide media index that manages the media references for a single slide.

```go
type SlideMediaIndex struct {
    // Has unexported fields.
}
```

### Constructor

```go
func NewSlideMediaIndex(slideIndex int) *SlideMediaIndex
```

### Methods

#### GetLocalRIDByHash

Returns the local rId for a given hash.

```go
func (smi *SlideMediaIndex) GetLocalRIDByHash(hash string) string
```

#### GetHashByLocalRID

Returns the hash for a given local rId.

```go
func (smi *SlideMediaIndex) GetHashByLocalRID(localRID string) string
```

#### AllLocalRIDs

Returns all local rIDs.

```go
func (smi *SlideMediaIndex) AllLocalRIDs() []string
```

#### LocalRefCount

Returns the number of local references.

```go
func (smi *SlideMediaIndex) LocalRefCount() int64
```

---

# MasterManager - Slide Master/Layout Manager

Slide master and slide layout manager.

```go
type MasterManager struct {
    // Has unexported fields.
}
```

### Constructors

```go
func NewMasterManager() *MasterManager

func NewMasterManagerWithCache(cache *MasterCache) *MasterManager
```

### Load Methods

#### LoadFromZipFile

Loads from a ZIP file path.

```go
func (m *MasterManager) LoadFromZipFile(filePath string) error
```

#### LoadFromZip

Loads slide masters and layouts from a ZIP Reader.

```go
func (m *MasterManager) LoadFromZip(zipReader *zip.Reader) error
```

**Note:** Iterates over `/ppt/slideMasters/` and `/ppt/slideLayouts/` directories inside the ZIP.

### Query Methods

#### GetMaster

Returns a slide master.

```go
func (m *MasterManager) GetMaster(masterID string) (*parts.SlideMasterData, bool)
```

#### GetMasterByName

Returns a slide master by name.

```go
func (m *MasterManager) GetMasterByName(name string) (*parts.SlideMasterData, bool)
```

#### GetLayout

Returns a slide layout.

```go
func (m *MasterManager) GetLayout(layoutID string) (*parts.SlideLayoutData, bool)
```

#### GetLayoutByName

Returns a slide layout by name.

```go
func (m *MasterManager) GetLayoutByName(name string) (*parts.SlideLayoutData, bool)
```

#### GetPlaceholder

Returns a placeholder.

```go
func (m *MasterManager) GetPlaceholder(layoutID, phType string) (*parts.Placeholder, bool)
```

### List Methods

#### AllMasters

Returns all slide masters.

```go
func (m *MasterManager) AllMasters() map[string]*parts.SlideMasterData
```

#### AllLayouts

Returns all slide layouts.

```go
func (m *MasterManager) AllLayouts() map[string]*parts.SlideLayoutData
```

#### ListLayoutIDs

Lists all layout IDs.

```go
func (m *MasterManager) ListLayoutIDs() []string
```

#### ListLayoutNames

Lists all layout names.

```go
func (m *MasterManager) ListLayoutNames() []string
```

### Statistics Methods

#### MasterCount

Returns the number of slide masters.

```go
func (m *MasterManager) MasterCount() int
```

#### LayoutCount

Returns the number of slide layouts.

```go
func (m *MasterManager) LayoutCount() int
```

#### Cache

Returns the internal cache (read-only).

```go
func (m *MasterManager) Cache() *MasterCache
```

---

# MasterCache - Slide Master Cache

A read-only cache of slide masters and layouts. After initialization all fields are read-only, enabling lock-free concurrent access.

```go
type MasterCache struct {
    // Has unexported fields.
}
```

### Constructor

```go
func NewMasterCache() *MasterCache
```

### Initialization Methods

#### Init

Initializes the cache with the provided data (runs only once; subsequent calls are ignored).

```go
func (c *MasterCache) Init(masters []*parts.SlideMasterData, layouts []*parts.SlideLayoutData)
```

#### InitFunc

Lazy initialization: accepts an initialization function that runs only on the first access.

```go
func (c *MasterCache) InitFunc(initFn func() ([]*parts.SlideMasterData, []*parts.SlideLayoutData))
```

### Query Methods

#### GetMaster

Returns a slide master by ID.

```go
func (c *MasterCache) GetMaster(masterID string) (*parts.SlideMasterData, bool)
```

#### GetMasterByName

Returns a slide master by name.

```go
func (c *MasterCache) GetMasterByName(name string) (*parts.SlideMasterData, bool)
```

#### GetLayout

Returns a slide layout by ID.

```go
func (c *MasterCache) GetLayout(layoutID string) (*parts.SlideLayoutData, bool)
```

#### GetLayoutByName

Returns a slide layout by name.

```go
func (c *MasterCache) GetLayoutByName(name string) (*parts.SlideLayoutData, bool)
```

#### GetPlaceholder

Returns a placeholder by layout ID and placeholder type.

```go
func (c *MasterCache) GetPlaceholder(layoutID, phType string) (*parts.Placeholder, bool)
```

**Parameters:**
- `phType`: a value returned by `PlaceholderType.String()`, e.g. "title", "body"

#### GetPlaceholderByID

Returns a placeholder by layout ID and placeholder ID.

```go
func (c *MasterCache) GetPlaceholderByID(layoutID, placeholderID string) (*parts.Placeholder, bool)
```

#### GetMasterPlaceholder

Returns a placeholder by master ID and placeholder type.

```go
func (c *MasterCache) GetMasterPlaceholder(masterID, phType string) (*parts.Placeholder, bool)
```

### List Methods

#### AllMasters

Returns all slide masters (read-only).

```go
func (c *MasterCache) AllMasters() map[string]*parts.SlideMasterData
```

#### AllLayouts

Returns all slide layouts (read-only).

```go
func (c *MasterCache) AllLayouts() map[string]*parts.SlideLayoutData
```

#### ListMasterIDs

Lists all master IDs.

```go
func (c *MasterCache) ListMasterIDs() []string
```

#### ListLayoutIDs

Lists all layout IDs.

```go
func (c *MasterCache) ListLayoutIDs() []string
```

#### ListLayoutNames

Lists all layout names.

```go
func (c *MasterCache) ListLayoutNames() []string
```

### Existence Check Methods

#### MasterExists

Checks whether a slide master exists.

```go
func (c *MasterCache) MasterExists(masterID string) bool
```

#### LayoutExists

Checks whether a slide layout exists.

```go
func (c *MasterCache) LayoutExists(layoutID string) bool
```

### Statistics Methods

#### MasterCount

Returns the number of slide masters.

```go
func (c *MasterCache) MasterCount() int
```

#### LayoutCount

Returns the number of slide layouts.

```go
func (c *MasterCache) LayoutCount() int
```

## Usage Examples

### Basic Media Management

```go
// Get the media manager
mm := pres.MediaManager()

// Add an image
data, _ := os.ReadFile("logo.png")
rId, resource := mm.AddMediaAuto("logo.png", data)

// Use it on a slide
slide := pres.AddSlide()
slide.AddPicture(100, 100, 200, 150, rId)
```

### Cross-Slide Deduplication

```go
// Use the same image on multiple slides
logoData, _ := os.ReadFile("logo.png")

// Add the same logo to multiple slides
for i := 0; i < 5; i++ {
    slide := pres.AddSlide()
    rId, _ := mm.AddMediaForSlide(i, logoData, "logo.png")
    slide.AddPicture(100, 100, 200, 150, rId)
}

// Only one copy of the logo is stored
stats := mm.GetDeduplicationStats()
fmt.Printf("Bytes saved: %d\n", stats.SavedBytes)
```

### Querying Media Information

```go
// Get media by type
images := mm.AllImages()
videos := mm.AllVideo()
audio := mm.AllAudio()

// Statistics
fmt.Printf("Images: %d, Videos: %d, Audio: %d\n",
    mm.CountImages(), mm.CountVideo(), mm.CountAudio())
```

### Using the Master Cache

```go
// Get the master cache
cache := pres.MasterCache()

// Get a layout
layout, ok := cache.GetLayoutByName("title")
if ok {
    fmt.Println("Found title layout:", layout.Name)
}

// Get a placeholder
ph, ok := cache.GetPlaceholder(layout.ID, "title")
if ok {
    fmt.Printf("Title placeholder position: (%d, %d)\n", ph.X, ph.Y)
}

// List all layouts
for _, name := range cache.ListLayoutNames() {
    fmt.Println("Layout:", name)
}
```
