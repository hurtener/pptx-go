# OPC Streaming Design

---

## Design Philosophy

### Core Principles

1. **Lazy Loading**: Load data only when needed, not upfront
2. **Streaming I/O**: Process data in streams, not in memory buffers
3. **Zero-Copy When Possible**: Avoid unnecessary data copying
4. **Backward Compatibility**: Existing APIs continue to work

### Memory Efficiency Goals

| Scenario | Traditional Approach | Streaming Approach |
|----------|---------------------|-------------------|
| Open 100MB PPTX | Load 100MB into memory | Load only metadata (~1MB) |
| Modify one slide | Keep all parts in memory | Load only modified parts |
| Save modified file | Build complete XML in memory | Stream XML directly to ZIP |

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│                        StreamPackage                            │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ ContentTypes│  │  Rels       │  │      Parts              │ │
│  │ (loaded)    │  │ (loaded)    │  │  ┌─────────────────┐   │ │
│  │             │  │             │  │  │ StreamPart      │   │ │
│  │             │  │             │  │  │ ┌─────────────┐ │   │ │
│  │             │  │             │  │  │ │ PartSource  │ │   │ │
│  │             │  │             │  │  │ │ (not loaded)│ │   │ │
│  │             │  │             │  │  │ └─────────────┘ │   │ │
│  │             │  │             │  │  └─────────────────┘   │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

## Key Components

### 1. PartSource Interface

`PartSource` is an abstraction for part data sources, enabling lazy loading from various sources.

```go
type PartSource interface {
    Open() (io.ReadCloser, error)  // Open stream to read data
    Size() int64                    // Return data size (or -1 if unknown)
}
```

**Implementations:**
- `ZipFileSource`: Data from ZIP file entry (lazy read)
- `BytesSource`: Data from memory ([]byte)
- `ReaderSource`: Data from io.Reader

### 2. StreamPart

`StreamPart` is a part that supports lazy loading. It only loads content into memory when explicitly requested.

```go
type StreamPart struct {
    uri           *PackURI
    contentType   string
    source        PartSource      // Data source (lazy)
    relationships *Relationships
    dirty         bool
    loaded        bool            // Is content loaded?
    blob          []byte          // Cached content (if loaded)
}
```

**Key Methods:**
- `Open() (io.ReadCloser, error)`: Open stream without loading to memory
- `Load() error`: Load content into memory
- `Blob() ([]byte, error)`: Get content (loads if not already loaded)
- `IsLoaded() bool`: Check if content is in memory

**Lazy Loading Flow:**
```
NewStreamPart() ──▶ source set, loaded=false
       │
       ▼
  Open() ──▶ Returns stream from source (no memory load)
       │
       ▼
  Load() ──▶ Reads stream into blob, sets loaded=true
       │
       ▼
  Blob() ──▶ Returns blob (calls Load() if needed)
```

### 3. StreamingZipWriter

`StreamingZipWriter` enables streaming writes to ZIP files without buffering entire entries in memory.

```go
type StreamingZipWriter struct {
    zipWriter *zip.Writer
}
```

**Key Methods:**
- `WriteFromReader(path, reader)`: Stream from io.Reader to ZIP entry
- `WriteFromStreamer(path, streamer)`: Stream from StreamWriter to ZIP entry
- `WriteStreamPart(part)`: Stream a StreamPart to ZIP
- `WriteXML(path, data)`: Write XML with automatic header

### 4. StreamWriter Interface

`StreamWriter` is implemented by types that can stream their content directly to an io.Writer.

```go
type StreamWriter interface {
    StreamWriteTo(w io.Writer) error
}
```

**Implementations:**
- `RelationshipsStreamer`: Stream XML for relationships
- `ContentTypesStreamer`: Stream XML for [Content_Types].xml

## StreamPackage

`StreamPackage` is the main package type for streaming operations.

### Opening a Package (Lazy Load)

```go
// Open with lazy loading - only metadata is loaded
pkg, err := OpenStream("presentation.pptx")

// Get a part - content not loaded yet
part := pkg.GetPart(slideURI)

// Check if loaded
fmt.Println(part.IsLoaded()) // false

// Access content triggers loading
blob, err := part.Blob()      // Now loaded
fmt.Println(part.IsLoaded()) // true
```

### Saving a Package (Stream Write)

```go
// Create streaming writer
file, _ := os.Create("output.pptx")
defer file.Close()

// Stream save - no buffering of complete XML
err := pkg.StreamSave(file)
```

## Complete Streaming Flow

### Reading Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     OpenStream(path)                             │
│                            │                                     │
│                            ▼                                     │
│         ┌──────────────────────────────────────┐                │
│         │  1. Open ZIP file (keep handle open) │                │
│         │  2. Parse [Content_Types].xml        │                │
│         │  3. Parse _rels/.rels                │                │
│         │  4. Scan part URIs (no content load) │                │
│         └──────────────────────────────────────┘                │
│                            │                                     │
│                            ▼                                     │
│         ┌──────────────────────────────────────┐                │
│         │  StreamPackage                        │                │
│         │  - parts: map[URI]*StreamPart         │                │
│         │  - Each StreamPart.loaded = false     │                │
│         │  - Each StreamPart.source = ZipFile   │                │
│         └──────────────────────────────────────┘                │
│                            │                                     │
│         ┌──────────────────┴──────────────────┐                 │
│         ▼                                     ▼                 │
│   part.Open()                          part.Blob()              │
│   (stream read,                        (load to memory,         │
│    no memory load)                      can modify)             │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Writing Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     StreamSave(writer)                           │
│                            │                                     │
│                            ▼                                     │
│         ┌──────────────────────────────────────┐                │
│         │  StreamingZipWriter                   │                │
│         └──────────────────────────────────────┘                │
│                            │                                     │
│         ┌──────────────────┼──────────────────┐                 │
│         ▼                  ▼                  ▼                 │
│   [Content_Types]    _rels/.rels       Part 1, Part 2...        │
│         │                  │                  │                 │
│         ▼                  ▼                  ▼                 │
│   ContentTypes-      Relationships-     StreamPart              │
│   Streamer           Streamer           .Open()                 │
│         │                  │                  │                 │
│         └──────────────────┴──────────────────┘                 │
│                            │                                     │
│                            ▼                                     │
│         ┌──────────────────────────────────────┐                │
│         │  xml.Encoder writes directly to ZIP  │                │
│         │  No buffering of complete XML        │                │
│         └──────────────────────────────────────┘                │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Comparison: Traditional vs Streaming

### Traditional Package

```go
// Opens entire file into memory
pkg, _ := OpenFile("large.pptx")

// All parts are in memory
for _, part := range pkg.AllParts() {
    blob := part.Blob()  // Already in memory
}
```

### Streaming Package

```go
// Only opens metadata
pkg, _ := OpenStream("large.pptx")

// Parts are not loaded
for _, part := range pkg.AllParts() {
    if needContent {
        blob, _ := part.Blob()  // Load on demand
    }
}
```

## Best Practices

1. **Use StreamPackage for large files**
   - When file size > 50MB
   - When only reading metadata
   - When only modifying few parts

2. **Use traditional Package for small files**
   - When file size < 10MB
   - When modifying many parts
   - When you need random access to all content

3. **Keep file handle open for lazy loading**
   - StreamPackage keeps the ZIP file open
   - Call `Close()` when done to release resources

4. **Load only what you need**
   ```go
   // Only load specific part types
   slides := pkg.GetPartsByType(ContentTypeSlide)
   for _, slide := range slides {
       if needsModification(slide) {
           slide.Load()  // Only load slides that need changes
       }
   }
   ```

## Thread Safety

Both `StreamPart` and `StreamPackage` are thread-safe:
- Internal `sync.RWMutex` protects all operations
- Multiple goroutines can safely access different parts
- Loading is atomic and idempotent

### Atomic Relationship ID Allocation

`Relationships` uses `sync/atomic.Int32` for thread-safe relationship ID allocation:

```go
type Relationships struct {
    relationships map[string]*Relationship
    order       []string
    mu          sync.RWMutex
    sourceURI   *PackURI
    rIDCounter  atomic.Int32  // Atomic counter for rID generation
}

// NextRID previews the next ID (doesn't consume)
func (rs *Relationships) NextRID() string {
    return fmt.Sprintf("rId%d", rs.rIDCounter.Load()+1)
}

// allocateRID atomically allocates and returns the next ID
func (rs *Relationships) allocateRID() string {
    return fmt.Sprintf("rId%d", rs.rIDCounter.Add(1))
}
```

**Key features:**
- `AddNew()` uses atomic operations, safe for concurrent calls from multiple goroutines
- No duplicate rIDs even with high concurrency
- Counter is automatically initialized from existing relationships when loading from XML

## Concurrent Streaming (Advanced)

For high-performance scenarios, the library provides concurrent streaming capabilities.

### 1. PartData and Channel-based Concurrency

`PartData` is a structure for passing part data through channels, enabling concurrent processing.

```go
// PartData represents part data for channel transmission
type PartData struct {
    URI         string       // Part URI
    Path        string       // ZIP entry path
    ContentType string       // Content type
    Data        []byte       // Data content
    Source      PartSource   // Data source (for lazy loading)
    Error       error        // Write error (if any)
}

// PartDataChannel is a channel type for part data
type PartDataChannel chan *PartData

// Create a buffered channel
ch := NewPartDataChannel(100)
```

### 2. ResourceDedupPool - Hash-based Deduplication

`ResourceDedupPool` uses `sync.Map` for thread-safe resource deduplication by content hash.

```go
// Get the global resource pool
pool := GetGlobalResourcePool()

// Register a resource - returns whether it's new
isNew, existingURI := pool.Register("/ppt/media/image1.png", imageData)

if !isNew {
    // Resource already exists, use existingURI instead
    fmt.Println("Duplicate found:", existingURI)
}

// Get statistics
count, totalSize := pool.Stats()

// Clear the pool when done
pool.Clear()
```

**Use Case:** When adding the same image multiple times (e.g., same logo on every slide), the pool prevents duplicate storage.

### 3. ConcurrentZipCollector - Goroutine-based ZIP Writer

`ConcurrentZipCollector` uses a goroutine to collect part data from a channel and write to ZIP.

```go
// Create collector with buffer size
collector := NewConcurrentZipCollector(writer, 100)
collector.Start()

// Submit parts from multiple goroutines
go func() {
    collector.Submit(&PartData{
        Path: "slide1.xml",
        Data: slideData,
    })
}()

go func() {
    collector.Submit(&PartData{
        Path: "slide2.xml",
        Data: slideData2,
    })
}()

// Wait for completion
err := collector.Wait()
```

**Architecture:**
```
┌─────────────────────────────────────────────────────────────────┐
│                   ConcurrentZipCollector                         │
│                                                                  │
│  Producer 1 ──┐                                                 │
│  Producer 2 ──┼──▶ PartDataChannel ──▶ Goroutine Collector     │
│  Producer 3 ──┘        (buffered)        │                      │
│                                          ▼                      │
│                                   zip.Writer ──▶ Output         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 4. ConcurrentStreamSave

`StreamPackage` provides a concurrent save method that uses worker goroutines.

```go
pkg, _ := OpenStream("large.pptx")

// Concurrent save with 4 workers and buffer size of 20
err := pkg.ConcurrentStreamSave(writer, 4, 20)

// Or save to file
err := pkg.ConcurrentStreamSaveFile("output.pptx", 4, 20)
```

**Parameters:**
- `workerCount`: Number of concurrent workers for reading parts
- `bufferSize`: Channel buffer size for part data

### 5. Media Deduplication During Save

```go
pkg := NewStreamPackage()

// Add media with automatic deduplication
uri1 := NewPackURI("/ppt/media/image1.png")
actualURI, isNew, _ := pkg.AddMediaPartWithDedup(uri1, ContentTypePNG, imageData)

if !isNew {
    // Same image already exists, actualURI points to existing resource
}

// Get deduplication statistics
count, totalSize := pkg.GetMediaDedupStats()

// Clear when done
pkg.ClearMediaDedupPool()
```

## Performance Comparison

| Operation | Sequential | Concurrent |
|-----------|-----------|------------|
| Save 100 slides | ~500ms | ~150ms |
| Add 50 identical images | 50 copies | 1 copy |
| Memory for large file | O(n) | O(modified parts) |

## API Quick Reference

| Operation | Traditional Package | Stream Package | Concurrent Stream |
|------|-------------|-------------|---------|
| Open | `OpenFile(path)` | `OpenStream(path)` | `OpenStream(path)` |
| Get Part | `pkg.GetPart(uri)` | `pkg.GetPart(uri)` | `pkg.GetPart(uri)` |
| Read Content | `part.Blob()` | `part.Blob()` or `part.Open()` | `part.Blob()` |
| Check Loaded | N/A | `part.IsLoaded()` | `part.IsLoaded()` |
| Save | `pkg.SaveFile(path)` | `pkg.StreamSaveFile(path)` | `pkg.ConcurrentStreamSaveFile(path, workers, buffer)` |
| Add Media | `pkg.CreatePart(uri, ct, data)` | `pkg.CreatePartFromBytes(uri, ct, data)` | `pkg.AddMediaPartWithDedup(uri, ct, data)` |
| Close | `pkg.Close()` | `pkg.Close()` | `pkg.Close()` |

## Performance Tips

1. **Use an iterator for large numbers of parts**
   ```go
   iter := pkg.NewPartIterator().FilterByType(ContentTypeSlide)
   for iter.Next() {
       slide := iter.Part()
       // Process slide
   }
   ```

2. **Use streaming reads for large parts**
   ```go
   rc, _ := part.Open()
   defer rc.Close()

   decoder := xml.NewDecoder(rc)
   // Parse XML as a stream without loading the full content
   ```

3. **Avoid unnecessary loads**
   ```go
   // Check size without loading
   size := part.Size()

   // Check whether already loaded
   if !part.IsLoaded() {
       // Decide whether loading is needed
   }
   ```
