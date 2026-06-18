package opc

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
)

// StreamPackage is a streaming OPC package that supports lazy loading and streaming writes.
type StreamPackage struct {
	parts          map[string]*StreamPart // URI -> StreamPart
	partOrder      []string               // Preserves insertion order.
	relationships  *Relationships
	contentTypes   *ContentTypes
	coreProperties *CoreProperties
	zipReader      *zip.Reader
	zipFile        *os.File // Kept open to support lazy loading.
	mu             sync.RWMutex
}

// NewStreamPackage creates a new, empty streaming package.
func NewStreamPackage() *StreamPackage {
	return &StreamPackage{
		parts:         make(map[string]*StreamPart),
		partOrder:     make([]string, 0),
		relationships: NewRelationships(RootURI()),
		contentTypes:  NewContentTypes(),
	}
}

// OpenStream opens an OPC package from a file path using streaming (lazy loading).
// The file handle remains open to support on-demand part loading.
func OpenStream(path string, opts ...OpenOption) (*StreamPackage, error) {
	cfg := resolveOpenConfig(opts)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	zipReader, err := zip.NewReader(file, stat.Size())
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	pkg := &StreamPackage{
		parts:         make(map[string]*StreamPart),
		partOrder:     make([]string, 0),
		relationships: NewRelationships(RootURI()),
		contentTypes:  NewContentTypes(),
		zipReader:     zipReader,
		zipFile:       file,
	}

	// Load only metadata; do not read part content yet.
	if err := pkg.loadContentTypes(cfg); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to load content types: %w", err)
	}

	if err := pkg.loadPartMetadata(cfg); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to load part metadata: %w", err)
	}

	if err := pkg.loadRelationships(cfg); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to load relationships: %w", err)
	}

	return pkg, nil
}

// OpenStreamFromReader opens an OPC package from an io.ReaderAt using streaming.
// Note: the caller is responsible for keeping the ReaderAt valid for the lifetime of the package.
func OpenStreamFromReader(r io.ReaderAt, size int64, opts ...OpenOption) (*StreamPackage, error) {
	cfg := resolveOpenConfig(opts)

	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	pkg := &StreamPackage{
		parts:         make(map[string]*StreamPart),
		partOrder:     make([]string, 0),
		relationships: NewRelationships(RootURI()),
		contentTypes:  NewContentTypes(),
		zipReader:     zipReader,
	}

	if err := pkg.loadContentTypes(cfg); err != nil {
		return nil, fmt.Errorf("failed to load content types: %w", err)
	}

	if err := pkg.loadPartMetadata(cfg); err != nil {
		return nil, fmt.Errorf("failed to load part metadata: %w", err)
	}

	if err := pkg.loadRelationships(cfg); err != nil {
		return nil, fmt.Errorf("failed to load relationships: %w", err)
	}

	return pkg, nil
}

// loadContentTypes loads content types (must be done eagerly).
func (p *StreamPackage) loadContentTypes(cfg openConfig) error {
	for _, f := range p.zipReader.File {
		// Normalize path.
		normalizedName := NormalizeZipPath(f.Name)
		if normalizedName == PathContentTypes {
			data, err := readZipEntry(f, cfg.maxPartBytes)
			if err != nil {
				return err
			}
			return p.contentTypes.FromXML(data)
		}
	}
	return fmt.Errorf("[Content_Types].xml not found")
}

// loadPartMetadata loads only part metadata, deferring content loading. The
// part body is read lazily, but the declared size and entry path are validated
// here so an oversized or zip-slip part is rejected at open (CLAUDE.md §7).
func (p *StreamPackage) loadPartMetadata(cfg openConfig) error {
	for _, f := range p.zipReader.File {
		// Normalize path.
		normalizedName := NormalizeZipPath(f.Name)

		// Skip special files.
		if normalizedName == PathContentTypes {
			continue
		}
		if strings.Contains(normalizedName, PathRelsDir+"/") && strings.HasSuffix(normalizedName, ".rels") {
			continue
		}
		if normalizedName == "" || strings.HasSuffix(normalizedName, "/") {
			continue
		}
		if err := safePartPath(normalizedName); err != nil {
			return err
		}
		if cfg.maxPartBytes > 0 && f.UncompressedSize64 > uint64(cfg.maxPartBytes) {
			return fmt.Errorf("%w: %q declares %d bytes (limit %d)", ErrPartTooLarge, f.Name, f.UncompressedSize64, cfg.maxPartBytes)
		}

		uri := NewPackURI("/" + normalizedName)
		contentType := p.contentTypes.GetContentType(uri)

		// Create a streaming part backed by a ZipFileSource for lazy loading.
		part := NewStreamPart(uri, contentType, NewZipFileSource(f))

		p.parts[uri.URI()] = part
		p.partOrder = append(p.partOrder, uri.URI())
	}

	return nil
}

// loadRelationships loads all relationship files.
func (p *StreamPackage) loadRelationships(cfg openConfig) error {
	for _, f := range p.zipReader.File {
		// Normalize path.
		normalizedName := NormalizeZipPath(f.Name)

		if !strings.Contains(normalizedName, PathRelsDir+"/") || !strings.HasSuffix(normalizedName, ".rels") {
			continue
		}
		if err := safePartPath(normalizedName); err != nil {
			return err
		}

		data, err := readZipEntry(f, cfg.maxPartBytes)
		if err != nil {
			return err
		}

		relURI := NewPackURI("/" + normalizedName)
		sourceURI := relURI.SourceURI()

		rels := NewRelationships(sourceURI)
		if err := rels.FromXML(data); err != nil {
			return err
		}

		if relURI.IsPackageRels() {
			p.relationships = rels
		} else {
			part := p.parts[sourceURI.URI()]
			if part != nil {
				part.LoadRelationships(data)
			}
		}
	}

	return nil
}

// ===== Part access =====

// GetPart returns the part with the given URI (content is loaded on demand).
func (p *StreamPackage) GetPart(uri *PackURI) *StreamPart {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.parts[uri.URI()]
}

// GetPartByStr returns the part with the given string URI.
func (p *StreamPackage) GetPartByStr(uri string) *StreamPart {
	return p.GetPart(NewPackURI(uri))
}

// ContainsPart reports whether a part with the given URI exists.
func (p *StreamPackage) ContainsPart(uri *PackURI) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, exists := p.parts[uri.URI()]
	return exists
}

// AllParts returns all parts in insertion order.
func (p *StreamPackage) AllParts() []*StreamPart {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*StreamPart, 0, len(p.partOrder))
	for _, uri := range p.partOrder {
		result = append(result, p.parts[uri])
	}
	return result
}

// PartCount returns the number of parts.
func (p *StreamPackage) PartCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.parts)
}

// PartURIs returns all part URIs in insertion order.
func (p *StreamPackage) PartURIs() []*PackURI {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*PackURI, 0, len(p.partOrder))
	for _, uri := range p.partOrder {
		result = append(result, NewPackURI(uri))
	}
	return result
}

// GetPartsByType returns all parts with the given content type.
func (p *StreamPackage) GetPartsByType(contentType string) []*StreamPart {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*StreamPart
	for _, uri := range p.partOrder {
		if part := p.parts[uri]; part.ContentType() == contentType {
			result = append(result, part)
		}
	}
	return result
}

// ===== Part management =====

// AddPart adds a streaming part to the package.
func (p *StreamPackage) AddPart(part *StreamPart) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	uri := part.PartURI().URI()
	if _, exists := p.parts[uri]; exists {
		return fmt.Errorf("part with URI %s already exists", uri)
	}

	p.parts[uri] = part
	p.partOrder = append(p.partOrder, uri)
	return nil
}

// CreateStreamPart creates and adds a streaming part.
func (p *StreamPackage) CreateStreamPart(uri *PackURI, contentType string, source PartSource) (*StreamPart, error) {
	part := NewStreamPart(uri, contentType, source)
	if err := p.AddPart(part); err != nil {
		return nil, err
	}
	return part, nil
}

// CreatePartFromBytes creates a part from a byte slice, loading it immediately into memory.
func (p *StreamPackage) CreatePartFromBytes(uri *PackURI, contentType string, data []byte) (*StreamPart, error) {
	part := NewStreamPart(uri, contentType, NewBytesSource(data))
	part.SetBlob(data) // Load into memory immediately.
	if err := p.AddPart(part); err != nil {
		return nil, err
	}
	return part, nil
}

// RemovePart removes a part from the package.
func (p *StreamPackage) RemovePart(uri *PackURI) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := uri.URI()
	if _, exists := p.parts[key]; !exists {
		return fmt.Errorf("part with URI %s not found", key)
	}

	delete(p.parts, key)

	// Remove from the order slice.
	for i, u := range p.partOrder {
		if u == key {
			p.partOrder = append(p.partOrder[:i], p.partOrder[i+1:]...)
			break
		}
	}

	return nil
}

// ===== Relationship management =====

// Relationships returns the package-level relationships.
func (p *StreamPackage) Relationships() *Relationships {
	return p.relationships
}

// AddRelationship adds a package-level relationship.
func (p *StreamPackage) AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error) {
	return p.relationships.AddNew(relType, targetURI, isExternal)
}

// GetPartByRelType returns the target part of the first relationship of the given type.
func (p *StreamPackage) GetPartByRelType(relType string) *StreamPart {
	rels := p.relationships.GetByType(relType)
	if len(rels) == 0 {
		return nil
	}
	return p.parts[rels[0].TargetURI().URI()]
}

// ===== Streaming save =====

// StreamSave writes the package as a ZIP archive to w using streaming.
func (p *StreamPackage) StreamSave(w io.Writer) error {
	sw := NewStreamingZipWriter(w)

	// 1. Write [Content_Types].xml (streaming).
	if err := p.streamWriteContentTypes(sw); err != nil {
		return fmt.Errorf("failed to write content types: %w", err)
	}

	// 2. Write package-level relationships (streaming).
	if err := p.streamWritePackageRelationships(sw); err != nil {
		return fmt.Errorf("failed to write package relationships: %w", err)
	}

	// 3. Stream all parts.
	if err := p.streamWriteParts(sw); err != nil {
		return fmt.Errorf("failed to write parts: %w", err)
	}

	// 4. Write core properties if present.
	if p.coreProperties != nil {
		if err := p.streamWriteCoreProperties(sw); err != nil {
			return fmt.Errorf("failed to write core properties: %w", err)
		}
	}

	return sw.Close()
}

// StreamSaveFile writes the package to a file using streaming.
func (p *StreamPackage) StreamSaveFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return p.StreamSave(file)
}

func (p *StreamPackage) streamWriteContentTypes(sw *StreamingZipWriter) error {
	p.updateContentTypes()
	streamer := NewContentTypesStreamer(p.contentTypes)
	return sw.WriteFromStreamer(PathContentTypes, streamer)
}

func (p *StreamPackage) streamWritePackageRelationships(sw *StreamingZipWriter) error {
	if p.relationships.Count() == 0 {
		return nil
	}

	streamer := NewRelationshipsStreamer(p.relationships)
	relPath := PathRelsDir + "/" + PathRelsFile
	return sw.WriteFromStreamer(relPath, streamer)
}

func (p *StreamPackage) streamWriteParts(sw *StreamingZipWriter) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, uri := range p.partOrder {
		part := p.parts[uri]

		// Stream the part content.
		if err := sw.WriteStreamPart(part); err != nil {
			return err
		}

		// Stream the relationships if any.
		if part.HasRelationships() {
			relPath := p.relFilePath(part.PartURI())
			streamer := NewRelationshipsStreamer(part.Relationships())
			if err := sw.WriteFromStreamer(relPath, streamer); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *StreamPackage) streamWriteCoreProperties(sw *StreamingZipWriter) error {
	data, err := p.coreProperties.ToXML()
	if err != nil {
		return err
	}
	return sw.WriteFromReader("docProps/core.xml", &bytesReaderAt{data: data})
}

func (p *StreamPackage) relFilePath(uri *PackURI) string {
	dir := path.Dir(strings.TrimPrefix(uri.URI(), "/"))
	filename := path.Base(uri.URI())
	return path.Join(dir, PathRelsDir, filename+".rels")
}

func (p *StreamPackage) updateContentTypes() {
	for _, uri := range p.partOrder {
		part := p.parts[uri]
		contentType := part.ContentType()
		packURI := part.PartURI()

		if packURI.IsRelationshipsPart() {
			contentType = ContentTypeRelationships
		}

		ext := packURI.Extension()
		defaultCT := p.contentTypes.GetDefault(ext)

		if contentType != "" && contentType != ContentTypeDefault {
			if defaultCT == "" || defaultCT == ContentTypeDefault || defaultCT != contentType {
				p.contentTypes.AddOverride(packURI, contentType)
			}
		}
	}
}

// ===== Other methods =====

// ContentTypes returns the content type definitions.
func (p *StreamPackage) ContentTypes() *ContentTypes {
	return p.contentTypes
}

// CoreProperties returns the core properties.
func (p *StreamPackage) CoreProperties() *CoreProperties {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.coreProperties
}

// SetCoreProperties sets the core properties.
func (p *StreamPackage) SetCoreProperties(props *CoreProperties) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.coreProperties = props
}

// Close closes the package and releases its resources.
func (p *StreamPackage) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close the underlying file.
	if p.zipFile != nil {
		err := p.zipFile.Close()
		p.zipFile = nil
		return err
	}

	return nil
}

// ===== Lazy-loading iterator =====

// PartIterator iterates over the parts of a StreamPackage.
type PartIterator struct {
	pkg    *StreamPackage
	index  int
	filter func(*StreamPart) bool
}

// NewPartIterator creates a part iterator for the package.
func (p *StreamPackage) NewPartIterator() *PartIterator {
	return &PartIterator{
		pkg:    p,
		index:  0,
		filter: nil,
	}
}

// FilterByType restricts the iterator to parts with the given content type.
func (it *PartIterator) FilterByType(contentType string) *PartIterator {
	it.filter = func(part *StreamPart) bool {
		return part.ContentType() == contentType
	}
	return it
}

// Next advances to the next matching part and reports whether one was found.
func (it *PartIterator) Next() bool {
	it.pkg.mu.RLock()
	defer it.pkg.mu.RUnlock()

	for it.index < len(it.pkg.partOrder) {
		uri := it.pkg.partOrder[it.index]
		it.index++
		part := it.pkg.parts[uri]
		if it.filter == nil || it.filter(part) {
			return true
		}
	}
	return false
}

// Part returns the current part.
func (it *PartIterator) Part() *StreamPart {
	if it.index <= 0 || it.index > len(it.pkg.partOrder) {
		return nil
	}
	uri := it.pkg.partOrder[it.index-1]
	return it.pkg.parts[uri]
}

// Open opens the content stream of the current part.
func (it *PartIterator) Open() (io.ReadCloser, error) {
	part := it.Part()
	if part == nil {
		return nil, fmt.Errorf("no current part")
	}
	return part.Open()
}

// ===== Concurrent streaming save =====

// ConcurrentStreamSave saves the package concurrently using a goroutine collector.
// workerCount controls the number of concurrent worker goroutines.
// bufferSize controls the channel buffer size.
func (p *StreamPackage) ConcurrentStreamSave(w io.Writer, workerCount, bufferSize int) error {
	// Create the concurrent collector.
	collector := NewConcurrentZipCollector(w, bufferSize)
	collector.Start()

	// Create wait group for workers.
	var wg sync.WaitGroup
	errChan := make(chan error, workerCount+1)

	// Producer: submit all parts to the channel.
	go func() {
		p.mu.RLock()
		defer p.mu.RUnlock()

		// 1. Write ContentTypes.
		p.updateContentTypes()
		ctData, err := p.contentTypes.ToXML()
		if err != nil {
			errChan <- fmt.Errorf("failed to serialize content types: %w", err)
			return
		}
		if err := collector.Submit(&PartData{
			Path: PathContentTypes,
			Data: ctData,
		}); err != nil {
			errChan <- err
			return
		}

		// 2. Write package-level relationships.
		if p.relationships.Count() > 0 {
			relData, err := p.relationships.ToXML()
			if err != nil {
				errChan <- fmt.Errorf("failed to serialize relationships: %w", err)
				return
			}
			relPath := PathRelsDir + "/" + PathRelsFile
			if err := collector.Submit(&PartData{
				Path: relPath,
				Data: relData,
			}); err != nil {
				errChan <- err
				return
			}
		}

		// 3. Submit parts concurrently using a worker pool.
		partChan := make(chan *StreamPart, workerCount*2)

		// Start worker goroutines to read part data.
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for part := range partChan {
					data, err := part.Blob()
					if err != nil {
						errChan <- fmt.Errorf("failed to read part %s: %w", part.PartURI().URI(), err)
						return
					}

					// Submit the part data.
					if err := collector.Submit(&PartData{
						URI:         part.PartURI().URI(),
						Path:        part.PartURI().MemberName(),
						ContentType: part.ContentType(),
						Data:        data,
					}); err != nil {
						errChan <- err
						return
					}

					// Submit relationship data if present.
					if part.HasRelationships() {
						relData, err := part.RelationshipsBlob()
						if err != nil {
							errChan <- fmt.Errorf("failed to serialize relationships for %s: %w", part.PartURI().URI(), err)
							return
						}
						relPath := p.relFilePath(part.PartURI())
						if err := collector.Submit(&PartData{
							Path: relPath,
							Data: relData,
						}); err != nil {
							errChan <- err
							return
						}
					}
				}
			}()
		}

		// Distribute parts to workers.
		for _, uri := range p.partOrder {
			partChan <- p.parts[uri]
		}
		close(partChan)

		// Wait for all workers to finish.
		wg.Wait()

		// 4. Write core properties if present.
		if p.coreProperties != nil {
			cpData, err := p.coreProperties.ToXML()
			if err != nil {
				errChan <- fmt.Errorf("failed to serialize core properties: %w", err)
				return
			}
			if err := collector.Submit(&PartData{
				Path: "docProps/core.xml",
				Data: cpData,
			}); err != nil {
				errChan <- err
				return
			}
		}

		// Signal that all data has been submitted.
		if err := collector.Close(); err != nil {
			errChan <- err
		}
	}()

	// Wait for completion or error.
	select {
	case err := <-errChan:
		return err
	case <-collector.doneChan:
		return nil
	}
}

// ConcurrentStreamSaveFile saves the package to a file using concurrent streaming.
func (p *StreamPackage) ConcurrentStreamSaveFile(path string, workerCount, bufferSize int) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return p.ConcurrentStreamSave(file, workerCount, bufferSize)
}

// ===== Resource deduplication =====

// RegisterMediaWithDedup registers a media resource and deduplicates it.
// Returns isNew (true if this is the first registration) and existingURI (the canonical URI).
func (p *StreamPackage) RegisterMediaWithDedup(uri string, data []byte) (isNew bool, existingURI string) {
	pool := GetGlobalResourcePool()
	return pool.Register(uri, data)
}

// AddMediaPartWithDedup adds a media part with deduplication.
// If the resource already exists, returns the existing part's URI without creating a new part.
func (p *StreamPackage) AddMediaPartWithDedup(uri *PackURI, contentType string, data []byte) (actualURI *PackURI, isNew bool, err error) {
	// Check whether the resource already exists.
	pool := GetGlobalResourcePool()
	isNew, existingURI := pool.Register(uri.URI(), data)

	if !isNew {
		// Resource already registered — return the existing URI.
		return NewPackURI(existingURI), false, nil
	}

	// Create a new part.
	part, err := p.CreatePartFromBytes(uri, contentType, data)
	if err != nil {
		pool.Release(ComputeHash(data))
		return nil, false, err
	}

	return part.PartURI(), true, nil
}

// GetMediaDedupStats returns deduplication statistics for media resources.
func (p *StreamPackage) GetMediaDedupStats() (count int, totalSize int64) {
	return GetGlobalResourcePool().Stats()
}

// ClearMediaDedupPool clears the media resource deduplication pool.
func (p *StreamPackage) ClearMediaDedupPool() {
	GetGlobalResourcePool().Clear()
}
