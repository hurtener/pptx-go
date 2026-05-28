# Go-PPTX

A Go library for creating, reading, and modifying PowerPoint (PPTX) files with streaming support for large files.

### Features

- **Full PPTX Support**: Create, read, and modify PPTX files
- **Streaming I/O**: Handle large files efficiently with lazy loading
- **OPC Implementation**: Complete Open Packaging Convention implementation
- **Parts Layer**: High-level PPTX content handling
  - Presentation, Slide, Master, Layout, Media parts
  - XML namespace handling for OOXML compatibility
  - Shape ID allocation and relationship management
- **Thread Safe**: Safe for concurrent use
  - `sync/atomic` for relationship ID allocation
  - `sync.RWMutex` for thread-safe operations
  - `sync.Map` for resource deduplication
- **Zero Dependencies**: Only uses Go standard library

### Installation

```bash
go get github.com/hurtener/pptx-go
```

### Quick Start

#### Traditional Usage (Small Files)

```go
package main

import (
    "github.com/hurtener/pptx-go/opc"
)

func main() {
    // Open existing file
    pkg, err := opc.OpenFile("presentation.pptx")
    if err != nil {
        panic(err)
    }
    defer pkg.Close()

    // Access parts
    slides := pkg.GetPartsByType(opc.ContentTypeSlide)

    // Save changes
    pkg.SaveFile("output.pptx")
}
```

#### Streaming Usage (Large Files)

```go
package main

import (
    "github.com/hurtener/pptx-go/opc"
)

func main() {
    // Open with lazy loading - only metadata is loaded
    pkg, err := opc.OpenStream("large.presentation.pptx")
    if err != nil {
        panic(err)
    }
    defer pkg.Close()

    // Get a part - content not loaded yet
    slide := pkg.GetPart(slideURI)

    // Load only when needed
    if needsModification {
        blob, _ := slide.Blob()  // Now loaded
        // ... modify blob
        slide.SetBlob(modifiedBlob)
    }

    // Stream save - no buffering of complete XML
    pkg.StreamSaveFile("output.pptx")
}
```

### When to Use Which Mode

| Scenario | Recommended Mode |
|----------|-----------------|
| File size < 10MB | Traditional |
| File size > 50MB | Streaming |
| Only reading metadata | Streaming |
| Modifying many parts | Traditional |
| Modifying few parts | Streaming |
| Random access to all content | Traditional |

### Thread-Safe Relationship ID Allocation

`Relationships` uses `sync/atomic.Int32` for thread-safe relationship ID allocation when multiple goroutines call `AddRelationship()` concurrently.

```go
// Automatic atomic ID allocation
rels := opc.NewRelationships(sourceURI)
rel1, _ := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide1.xml", false)  // rId1
rel2, _ := rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide2.xml", false)  // rId2

// Preview next ID without consuming
nextID := rels.NextRID()  // "rId3"

// Thread-safe for concurrent calls
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        rels.AddNew(opc.RelTypeSlide, "/ppt/slides/slide.xml", false)
    }()
}
wg.Wait()
// All IDs are unique, no duplicates
```

**Key features:**
- `AddNew()` uses atomic operations, safe for concurrent calls
- Counter auto-initializes from existing relationships when loading from XML
- `NextRID()` previews the next ID without consuming

### Concurrent Streaming (Advanced)

For high-performance scenarios, the library provides concurrent streaming capabilities:

| Feature | Description |
|---------|-------------|
| `PartDataChannel` | Channel-based concurrent part writing |
| `ResourceDedupPool` | `sync.Map` based image/media deduplication |
| `ConcurrentZipCollector` | Goroutine-based ZIP writing |
| `ConcurrentStreamSave` | Worker-based concurrent save |

See [Streaming Design](docs/streaming-design.md) for detailed documentation.

### Documentation

- [Streaming Design](docs/streaming-design.md) - Detailed streaming architecture
- [OPC Package API](docs/opc/README.md) - OPC API reference (go doc)
- [OPC Relationship Resolution](docs/opc/relationship_resolution.md) - How relative paths are resolved
- [Presentation Part](docs/parts/presentation.md) - Presentation.xml handling
- [Slide Part](docs/parts/slide.md) - Slide content and shapes
- [Master Part](docs/parts/master.md) - Slide master and layouts
- [Media Manager](docs/parts/media.md) - Media handling and deduplication
- [XML Utilities](docs/parts/xmlutils.md) - XML namespace handling for OOXML

### Project Structure

```
go-pptx/
├── opc/                    # Open Packaging Convention implementation
│   ├── constants.go        # Content types and relationship types
│   ├── packuri.go          # Pack URI handling
│   ├── part.go             # Part and PartCollection
│   ├── package.go          # Traditional Package
│   ├── contenttypes.go     # [Content_Types].xml
│   ├── coreprops.go        # Core properties
│   ├── relation.go         # Relationships
│   ├── stream.go           # Streaming types
│   └── streampkg.go        # Streaming Package
├── parts/                  # PPTX content parts
│   ├── presentation.go     # Presentation part (presentation.xml)
│   ├── slide.go            # Slide part (slideN.xml)
│   ├── slide_types.go      # Slide XML structures
│   ├── master.go           # Slide master
│   ├── master_types.go     # Master XML structures
│   ├── master_cache.go     # Thread-safe master cache
│   ├── media.go            # Media handling
│   ├── media_manager.go    # Media manager with deduplication
│   ├── coreprops.go        # Core properties part
│   ├── relationship.go     # XML relationship structures
│   └── xmlutils.go         # XML namespace utilities
├── test/
│   ├── parts/              # Parts tests
│   ├── utils/              # Test utilities and examples
│   └── pipeline_test.go    # Integration tests
└── docs/
    ├── streaming-design.md # Streaming design documentation
    ├── opc/                # OPC documentation
    │   ├── README.md       # OPC API reference
    │   └── relationship_resolution.md
    └── parts/              # Parts documentation
        ├── presentation.md
        ├── slide.md
        ├── master.md
        ├── media.md
        ├── relationship.md
        └── xmlutils.md
```

### License

MIT License
