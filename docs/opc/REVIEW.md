# OPC Layer Feature Review

> Review date: 2026-03-30
> Review scope: opc package core functionality

---

## 1. PPTX ZIP Packaging and Unpacking

### Status: ✅ Complete

| Feature | Implementation | File |
|---------|---------------|------|
| ZIP reading | `archive/zip.Reader` | `package.go:64-89` |
| ZIP writing | `archive/zip.Writer` + `CreateHeader` | `package.go:454-485` |
| Path normalization | `NormalizeZipPath()` | `packuri.go:318-340` |
| Timestamp handling | `time.Now()` + MS-DOS format | `package.go:468` |

### Key Implementation

```go
// Use FileHeader when creating ZIP entries to ensure correct timestamps
func createZipEntry(zipWriter *zip.Writer, path string, size int) (io.Writer, error) {
    path = strings.TrimPrefix(path, "/")  // Strip leading slash

    header := &zip.FileHeader{
        Name:               path,
        UncompressedSize:   uint32(size),
        UncompressedSize64: uint64(size),
        Modified:           time.Now(),  // Fixes Windows Explorer MS-DOS time bug
        Method:             zip.Deflate,
    }
    header.Flags |= 0x800  // UTF-8 filename flag

    return zipWriter.CreateHeader(header)
}
```

### Issues Resolved

- ✅ Windows backslash path compatibility
- ✅ MS-DOS timestamps (fixes Windows Explorer display issue)
- ✅ UTF-8 filename support
- ✅ Leading slash stripping (conforms to ZIP specification)

---

## 2. .rels Relationship Management

### Status: ✅ Complete

| Feature | Implementation | File |
|---------|---------------|------|
| Automatic rId allocation | `atomic.Int32` atomic counter | `relation.go:110` |
| Monotonically increasing, no duplicates | `allocateRID()` | `relation.go:246-248` |
| Counter initialization | `InitRIDCounter()` | `relation.go:252-256` |
| Thread safety | `sync.RWMutex` | `relation.go:108` |

### Key Implementation

```go
type Relationships struct {
    relationships map[string]*Relationship
    order        []string
    mu           sync.RWMutex
    sourceURI    *PackURI
    rIDCounter   atomic.Int32  // Atomic counter
}

// Thread-safe ID allocation
func (rs *Relationships) allocateRID() string {
    return fmt.Sprintf("rId%d", rs.rIDCounter.Add(1))
}

// Initialize the counter after loading from XML to avoid conflicts
func (rs *Relationships) initRIDCounterLocked() {
    maxNum := int32(0)
    for rID := range rs.relationships {
        if strings.HasPrefix(rID, "rId") {
            var num int
            fmt.Sscanf(rID, "rId%d", &num)
            if int32(num) > maxNum {
                maxNum = int32(num)
            }
        }
    }
    rs.rIDCounter.Store(maxNum)
}
```

---

## 3. [Content_Types].xml Management

### Status: ✅ Complete

| Feature | Implementation | File |
|---------|---------------|------|
| Automatic registration | `updateContentTypes()` | `package.go:495-511` |
| Default types | `DefaultContentTypes` | `constants.go:109-131` |
| Override management | `AddOverride()` | `contenttypes.go:39-43` |
| Smart detection | By extension / content type | `contenttypes.go:47-63` |

### Workflow

```
Package.Save()
    └─> writeContentTypes()
        └─> updateContentTypes()  // Iterates all Parts
            ├─> Get URI and ContentType
            ├─> Check for default mapping
            └─> Add Override if no default or type differs
```

---

## 4. Clone() Smart Copy Method

### Status: ✅ Complete

| Type | Copy Strategy | Method | Criterion |
|------|--------------|--------|-----------|
| Images (PNG/JPEG/GIF/...) | Shallow copy (zero-copy) | `CloneShared()` | `IsImmutableContentType()` |
| Audio/video (MP4/WAV/...) | Shallow copy | `CloneShared()` | `IsImmutableContentType()` |
| Theme/master (Theme/Master) | Shallow copy | `CloneShared()` | `IsImmutableContentType()` |
| Font | Shallow copy | `CloneShared()` | `IsImmutableContentType()` |
| Slide | Deep copy | `Clone()` | Mutable by default |
| Presentation | Deep copy | `Clone()` | Mutable by default |

### Key Implementation

```go
// package.go:479-529
func (p *Package) Clone() *Package {
    newPkg := NewPackage()

    for _, part := range p.parts.All() {
        var newPart *Part

        if IsImmutableContentType(part.ContentType()) {
            newPart = part.CloneShared()  // zero-copy
        } else {
            newPart = part.Clone()  // deep copy
        }
        _ = newPkg.parts.Add(newPart)
    }

    // Clone relationships and content types
    newPkg.relationships = p.relationships.Clone()
    // ...

    return newPkg
}
```

### Part Copy Implementation

```go
// part.go:219-241 - Shallow copy
func (p *Part) CloneShared() *Part {
    return &Part{
        uri:          p.uri,              // Shared pointer
        contentType:  p.contentType,
        sharedBlob:   p.blob,             // zero-copy!
        relationships: p.relationships,   // Shared
        immutable:    true,
    }
}

// part.go:194-217 - Deep copy
func (p *Part) Clone() *Part {
    blobCopy := make([]byte, len(p.blob))
    copy(blobCopy, p.blob)  // Independent copy

    return &Part{
        uri:          p.uri.Clone(),
        blob:         blobCopy,
        relationships: p.relationships.Clone(),
        immutable:    false,
    }
}
```

---

## 5. Other Features

### 5.1 Streaming

| Component | Purpose | File |
|-----------|---------|------|
| `StreamPackage` | Lazy loading, streaming reads and writes | `streampkg.go` |
| `StreamPart` | On-demand content loading | `stream.go:206-421` |
| `StreamingZipWriter` | Streaming ZIP writes | `stream.go:95-203` |
| `PartIterator` | Lazy-loading iterator | `streampkg.go:501-557` |

### 5.2 Concurrency

| Component | Purpose | File |
|-----------|---------|------|
| `ConcurrentZipCollector` | goroutine + channel collection | `stream.go:730-831` |
| `ConcurrentStreamSave()` | Concurrent save | `streampkg.go:564-705` |
| `sync.RWMutex` | Read-write lock protection | All structs |
| `atomic.Int32` | Atomic counter | `relation.go:110` |

### 5.3 Resource Management

| Component | Purpose | File |
|-----------|---------|------|
| `ResourcePool` | Global resource pool + reference counting | `resource_pool.go` |
| `ResourceDedupPool` | Hash-based deduplication pool | `stream.go:586-712` |
| `GetGlobalPool()` | Global singleton | `resource_pool.go:32` |

### 5.4 Core Properties

| Component | Purpose | File |
|-----------|---------|------|
| `CoreProperties` | Dublin Core metadata | `coreprops.go` |
| Title/author/timestamps/etc. | 12 properties | `coreprops.go:11-23` |
| XML serialization | `ToXML()` / `FromXML()` | `coreprops.go:263-300` |

### 5.5 URI Handling (PackURI)

| Feature | Method | File |
|---------|--------|------|
| Path resolution | `Join()`, `RelPathFrom()` | `packuri.go:96-158` |
| Relationships file | `RelationshipsURI()`, `SourceURI()` | `packuri.go:174-211` |
| Normalization | `NormalizeURI()`, `NormalizeZipPath()` | `packuri.go:295-340` |

---

## 6. Test Coverage

### Test Statistics

| Directory | Test Count | Status |
|-----------|-----------|--------|
| `test/opc` | 5 | ✅ All passing |
| `test/utils` | 97+ | ✅ All passing |

### Key Tests

| Test | What It Verifies |
|------|-----------------|
| `TestResourcePool_*` | Resource pool functionality |
| `TestPackage_Clone_SmartCloning` | Smart copy strategy |
| `TestZipEntry_Timestamp` | Timestamp correctness |
| `TestZipEntry_TimestampNotZero` | Windows compatibility |
| `TestNormalizeZipPath_*` | Path normalization |

---

## 7. Summary

### Feature Completeness

| Feature | Status |
|---------|--------|
| ZIP packaging/unpacking | ✅ Complete |
| Windows slash handling | ✅ Complete |
| Timestamp handling | ✅ Complete |
| .rels relationship management | ✅ Complete |
| ContentTypes management | ✅ Complete |
| Clone() smart copy | ✅ Complete |
| Streaming | ✅ Complete |
| Concurrency safety | ✅ Complete |
| Resource pool / deduplication | ✅ Complete |

### Design Highlights

1. **Zero-copy optimization**: Immutable resources share their underlying data, reducing memory usage.
2. **Atomic operations**: rId allocation uses `atomic.Int32`, eliminating lock contention.
3. **Lazy loading**: `StreamPackage` supports on-demand loading, making large-file handling more efficient.
4. **Concurrent collection**: `ConcurrentZipCollector` uses goroutines for parallel writes.
5. **Resource deduplication**: Hash-based deduplication pool prevents redundant storage of identical resources.

---

## 8. File Structure

```
opc/
├── constants.go      # Constants (content types, relationship types, namespaces)
├── packuri.go        # PackURI path handling
├── package.go        # Package core implementation
├── streampkg.go      # StreamPackage streaming support
├── stream.go         # Streaming writers, data sources
├── part.go           # Part implementation
├── contenttypes.go   # ContentTypes management
├── relation.go       # Relationships management
├── coreprops.go      # CoreProperties
└── resource_pool.go  # ResourcePool
```
