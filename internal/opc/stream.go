package opc

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// ===== Data source interfaces and implementations =====

// PartSource is the interface for a part's data source.
type PartSource interface {
	Open() (io.ReadCloser, error)
	Size() int64
}

// ZipFileSource is a data source backed by an entry in a ZIP archive.
type ZipFileSource struct {
	file *zip.File
}

// NewZipFileSource creates a data source from a zip.File.
func NewZipFileSource(f *zip.File) *ZipFileSource {
	return &ZipFileSource{file: f}
}

// Open opens the ZIP archive entry.
func (s *ZipFileSource) Open() (io.ReadCloser, error) {
	return s.file.Open()
}

// Size returns the uncompressed size.
func (s *ZipFileSource) Size() int64 {
	return int64(s.file.UncompressedSize64)
}

// BytesSource is a data source backed by an in-memory byte slice.
type BytesSource struct {
	data []byte
}

// NewBytesSource creates a data source from a byte slice.
func NewBytesSource(data []byte) *BytesSource {
	return &BytesSource{data: data}
}

// Open returns a reader over the byte slice.
func (s *BytesSource) Open() (io.ReadCloser, error) {
	return io.NopCloser(&bytesReaderAt{data: s.data}), nil
}

// Size returns the data size.
func (s *BytesSource) Size() int64 {
	return int64(len(s.data))
}

// ReaderSource is a data source backed by an io.Reader.
type ReaderSource struct {
	reader io.Reader
	size   int64
}

// NewReaderSource creates a data source from an io.Reader.
func NewReaderSource(r io.Reader, size int64) *ReaderSource {
	return &ReaderSource{reader: r, size: size}
}

// Open returns the underlying reader.
func (s *ReaderSource) Open() (io.ReadCloser, error) {
	return io.NopCloser(s.reader), nil
}

// Size returns the data size.
func (s *ReaderSource) Size() int64 {
	return s.size
}

// ===== Streaming write interfaces =====

// StreamWriter is the interface for streaming writers.
type StreamWriter interface {
	StreamWriteTo(w io.Writer) error
}

// XMLStreamer is the interface for XML streaming writers.
type XMLStreamer interface {
	StreamXML(enc *xml.Encoder) error
}

// ===== Streaming ZIP writer =====

// StreamingZipWriter is a streaming ZIP writer.
type StreamingZipWriter struct {
	zipWriter *zip.Writer
}

// NewStreamingZipWriter creates a streaming ZIP writer.
func NewStreamingZipWriter(w io.Writer) *StreamingZipWriter {
	return &StreamingZipWriter{
		zipWriter: zip.NewWriter(w),
	}
}

// createZipHeader creates a ZIP file header with correct timestamps and compatibility settings.
func (sw *StreamingZipWriter) createZipHeader(path string) *zip.FileHeader {
	// Strip the leading slash to comply with the ZIP spec.
	path = strings.TrimPrefix(path, "/")

	header := &zip.FileHeader{
		Name:     path,
		Modified: time.Now(), // Set current timestamp.
		Method:   zip.Deflate,
	}

	// UTF-8 file name flag.
	header.Flags |= 0x800

	return header
}

// Create creates a ZIP entry and returns a writer for it.
func (sw *StreamingZipWriter) Create(path string) (io.Writer, error) {
	header := sw.createZipHeader(path)
	return sw.zipWriter.CreateHeader(header)
}

// WriteFromReader streams a ZIP entry from a Reader.
func (sw *StreamingZipWriter) WriteFromReader(path string, reader io.Reader) error {
	header := sw.createZipHeader(path)
	w, err := sw.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

// WriteFromStreamer streams a ZIP entry using a StreamWriter.
func (sw *StreamingZipWriter) WriteFromStreamer(path string, streamer StreamWriter) error {
	header := sw.createZipHeader(path)
	w, err := sw.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	return streamer.StreamWriteTo(w)
}

// WriteFromXMLStreamer streams a ZIP entry using an XMLStreamer (automatically prepends the XML declaration).
func (sw *StreamingZipWriter) WriteFromXMLStreamer(path string, streamer XMLStreamer) error {
	header := sw.createZipHeader(path)
	w, err := sw.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Write the XML declaration.
	if _, err := w.Write([]byte(XMLDeclaration)); err != nil {
		return err
	}

	encoder := xml.NewEncoder(w)
	if err := streamer.StreamXML(encoder); err != nil {
		return err
	}
	return encoder.Flush()
}

// WriteStreamPart streams a StreamPart to the ZIP archive.
func (sw *StreamingZipWriter) WriteStreamPart(part *StreamPart) error {
	path := part.PartURI().MemberName()
	header := sw.createZipHeader(path)
	w, err := sw.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	// Open the part stream.
	rc, err := part.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Stream-copy the content.
	_, err = io.Copy(w, rc)
	return err
}

// WriteBytes writes a byte slice to a ZIP entry.
func (sw *StreamingZipWriter) WriteBytes(path string, data []byte) error {
	header := sw.createZipHeader(path)
	w, err := sw.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// WriteXML writes XML data to a ZIP entry, automatically prepending the XML declaration.
func (sw *StreamingZipWriter) WriteXML(path string, data []byte) error {
	header := sw.createZipHeader(path)
	w, err := sw.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	// Write the XML declaration.
	if _, err := w.Write([]byte(XMLDeclaration)); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Close closes the underlying ZIP writer.
func (sw *StreamingZipWriter) Close() error {
	return sw.zipWriter.Close()
}

// ===== Streaming parts =====

// StreamPart is a streaming part that supports lazy loading.
type StreamPart struct {
	uri           *PackURI
	contentType   string
	source        PartSource
	relationships *Relationships
	dirty         bool
	loaded        bool
	blob          []byte // Cached data (set once loaded).
	mu            sync.RWMutex
}

// NewStreamPart creates a streaming part.
func NewStreamPart(uri *PackURI, contentType string, source PartSource) *StreamPart {
	return &StreamPart{
		uri:           uri,
		contentType:   contentType,
		source:        source,
		relationships: NewRelationships(uri),
		dirty:         false,
		loaded:        false,
	}
}

// PartURI returns the part URI.
func (p *StreamPart) PartURI() *PackURI {
	return p.uri
}

// ContentType returns the content type.
func (p *StreamPart) ContentType() string {
	return p.contentType
}

// SetContentType sets the content type.
func (p *StreamPart) SetContentType(ct string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.contentType = ct
	p.dirty = true
}

// Open opens the part content stream.
func (p *StreamPart) Open() (io.ReadCloser, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// If already loaded into memory, return an in-memory reader.
	if p.loaded {
		return io.NopCloser(&bytesReaderAt{data: p.blob}), nil
	}

	// Otherwise open from the source.
	if p.source != nil {
		return p.source.Open()
	}

	return nil, nil
}

// Load reads the part content into memory.
func (p *StreamPart) Load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.loaded {
		return nil
	}

	if p.source == nil {
		return nil
	}

	rc, err := p.source.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	p.blob, err = io.ReadAll(rc)
	if err != nil {
		return err
	}

	p.loaded = true
	return nil
}

// Blob returns the content, loading it first if necessary.
func (p *StreamPart) Blob() ([]byte, error) {
	if err := p.Load(); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.blob, nil
}

// SetBlob sets the content.
func (p *StreamPart) SetBlob(data []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.blob = data
	p.loaded = true
	p.dirty = true
}

// SetBlobFromReader sets the content from a Reader.
func (p *StreamPart) SetBlobFromReader(r io.Reader) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	p.blob = data
	p.loaded = true
	p.dirty = true
	return nil
}

// IsLoaded reports whether the content has been loaded into memory.
func (p *StreamPart) IsLoaded() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.loaded
}

// IsDirty reports whether the part has been modified.
func (p *StreamPart) IsDirty() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dirty
}

// SetDirty sets the dirty flag.
func (p *StreamPart) SetDirty(dirty bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dirty = dirty
}

// Relationships returns the relationship collection.
func (p *StreamPart) Relationships() *Relationships {
	return p.relationships
}

// LoadRelationships loads relationships from XML data.
func (p *StreamPart) LoadRelationships(data []byte) error {
	return p.relationships.FromXML(data)
}

// HasRelationships reports whether the part has any relationships.
func (p *StreamPart) HasRelationships() bool {
	return p.relationships.Count() > 0
}

// RelationshipsBlob returns the serialized XML of the relationships.
func (p *StreamPart) RelationshipsBlob() ([]byte, error) {
	if p.relationships.Count() == 0 {
		return nil, nil
	}
	return p.relationships.ToXML()
}

// RelationshipsURI returns the URI of the relationships file for this part.
func (p *StreamPart) RelationshipsURI() *PackURI {
	return p.uri.RelationshipsURI()
}

// Size returns the content size.
func (p *StreamPart) Size() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.loaded {
		return int64(len(p.blob))
	}
	if p.source != nil {
		return p.source.Size()
	}
	return 0
}

// UnmarshalBlob unmarshals the blob as XML into v, loading it first if necessary.
func (p *StreamPart) UnmarshalBlob(v any) error {
	if err := p.Load(); err != nil {
		return err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return xml.Unmarshal(p.blob, v)
}

// Clone returns a deep copy of the part.
func (p *StreamPart) Clone() *StreamPart {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var blobCopy []byte
	if p.loaded && p.blob != nil {
		blobCopy = make([]byte, len(p.blob))
		copy(blobCopy, p.blob)
	}

	return &StreamPart{
		uri:           p.uri.Clone(),
		contentType:   p.contentType,
		source:        p.source,
		relationships: p.relationships.Clone(),
		dirty:         p.dirty,
		loaded:        p.loaded,
		blob:          blobCopy,
	}
}

// ===== Streaming writer implementations =====

// RelationshipsStreamer is a streaming writer for relationships.
type RelationshipsStreamer struct {
	rels *Relationships
}

// NewRelationshipsStreamer creates a streaming writer for relationships.
func NewRelationshipsStreamer(rels *Relationships) *RelationshipsStreamer {
	return &RelationshipsStreamer{rels: rels}
}

// StreamWriteTo implements StreamWriter.
func (rs *RelationshipsStreamer) StreamWriteTo(w io.Writer) error {
	encoder := xml.NewEncoder(w)

	// Write the Relationships root element.
	start := xml.StartElement{
		Name: xml.Name{Local: "Relationships"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns"}, Value: NamespaceRelationships},
		},
	}

	if err := encoder.EncodeToken(start); err != nil {
		return err
	}

	// Write each Relationship element.
	for _, rel := range rs.rels.All() {
		relElem := xml.StartElement{
			Name: xml.Name{Local: "Relationship"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "Id"}, Value: rel.RID()},
				{Name: xml.Name{Local: "Type"}, Value: rel.Type()},
				{Name: xml.Name{Local: "Target"}, Value: rel.TargetRef()},
			},
		}
		if rel.IsExternal() {
			relElem.Attr = append(relElem.Attr, xml.Attr{
				Name:  xml.Name{Local: "TargetMode"},
				Value: "External",
			})
		}

		if err := encoder.EncodeToken(relElem); err != nil {
			return err
		}
		if err := encoder.EncodeToken(relElem.End()); err != nil {
			return err
		}
	}

	// Close the root element.
	if err := encoder.EncodeToken(start.End()); err != nil {
		return err
	}

	return encoder.Flush()
}

// ContentTypesStreamer is a streaming writer for ContentTypes.
type ContentTypesStreamer struct {
	ct *ContentTypes
}

// NewContentTypesStreamer creates a streaming writer for ContentTypes.
func NewContentTypesStreamer(ct *ContentTypes) *ContentTypesStreamer {
	return &ContentTypesStreamer{ct: ct}
}

// StreamWriteTo implements StreamWriter.
func (cs *ContentTypesStreamer) StreamWriteTo(w io.Writer) error {
	encoder := xml.NewEncoder(w)

	// Write the Types root element.
	start := xml.StartElement{
		Name: xml.Name{Local: "Types"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns"}, Value: NamespaceOPCPackage},
		},
	}

	if err := encoder.EncodeToken(start); err != nil {
		return err
	}

	// Write Default elements.
	for ext, ctType := range cs.ct.Defaults() {
		defElem := xml.StartElement{
			Name: xml.Name{Local: "Default"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "Extension"}, Value: ext},
				{Name: xml.Name{Local: "ContentType"}, Value: ctType},
			},
		}
		if err := encoder.EncodeToken(defElem); err != nil {
			return err
		}
		if err := encoder.EncodeToken(defElem.End()); err != nil {
			return err
		}
	}

	// Write Override elements.
	for uri, ctType := range cs.ct.Overrides() {
		overrideElem := xml.StartElement{
			Name: xml.Name{Local: "Override"},
			Attr: []xml.Attr{
				{Name: xml.Name{Local: "PartName"}, Value: uri},
				{Name: xml.Name{Local: "ContentType"}, Value: ctType},
			},
		}
		if err := encoder.EncodeToken(overrideElem); err != nil {
			return err
		}
		if err := encoder.EncodeToken(overrideElem.End()); err != nil {
			return err
		}
	}

	// Close the root element.
	if err := encoder.EncodeToken(start.End()); err != nil {
		return err
	}

	return encoder.Flush()
}

// ===== Concurrent write data structures =====

// PartData carries part data through a channel.
type PartData struct {
	URI         string     // Part URI
	Path        string     // ZIP internal path
	ContentType string     // Content type
	Data        []byte     // Data payload
	Source      PartSource // Data source (for lazy loading)
	Error       error      // Write error, if any
}

// PartDataChannel is the channel type for PartData.
type PartDataChannel chan *PartData

// NewPartDataChannel creates a buffered PartData channel.
func NewPartDataChannel(bufferSize int) PartDataChannel {
	return make(PartDataChannel, bufferSize)
}

// ===== Global resource deduplication pool =====

// ResourceHashKey is the key type used in the deduplication pool.
type ResourceHashKey string

// ResourceEntry is an entry in the deduplication pool.
type ResourceEntry struct {
	URI       string // Part URI
	Hash      string // Content hash (SHA256)
	Size      int64  // Original size
	Reference int    // Reference count
}

// ResourceDedupPool is a globally shared resource deduplication pool.
// It uses sync.Map for concurrent-safe deduplication.
type ResourceDedupPool struct {
	entries sync.Map // map[ResourceHashKey]*ResourceEntry
	mu      sync.RWMutex
}

// globalResourcePool is the singleton resource deduplication pool.
var globalResourcePool = &ResourceDedupPool{}

// GetGlobalResourcePool returns the global resource deduplication pool.
func GetGlobalResourcePool() *ResourceDedupPool {
	return globalResourcePool
}

// NewResourceDedupPool creates a new resource deduplication pool.
func NewResourceDedupPool() *ResourceDedupPool {
	return &ResourceDedupPool{}
}

// ComputeHash computes a hash of the given data.
// Note: this is a fast non-cryptographic FNV-1a hash, not SHA-256.
// It avoids an extra import; callers that need cryptographic strength should use crypto/sha256.
func ComputeHash(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// FNV-1a hash.
	var hash uint32 = 2166136261
	for _, b := range data {
		hash ^= uint32(b)
		hash *= 16777619
	}

	// Include the length to reduce collision probability.
	return fmt.Sprintf("%x-%d", hash, len(data))
}

// Register registers a resource and returns whether it is new.
// If the resource already exists, its reference count is incremented and false is returned.
func (p *ResourceDedupPool) Register(uri string, data []byte) (isNew bool, existingURI string) {
	hash := ComputeHash(data)
	key := ResourceHashKey(hash)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already registered.
	if entry, ok := p.entries.Load(key); ok {
		e := entry.(*ResourceEntry)
		e.Reference++
		return false, e.URI
	}

	// New resource.
	entry := &ResourceEntry{
		URI:       uri,
		Hash:      hash,
		Size:      int64(len(data)),
		Reference: 1,
	}
	p.entries.Store(key, entry)
	return true, uri
}

// RegisterWithHash registers a resource using a pre-computed hash.
func (p *ResourceDedupPool) RegisterWithHash(uri string, hash string, size int64) (isNew bool, existingURI string) {
	key := ResourceHashKey(hash)

	p.mu.Lock()
	defer p.mu.Unlock()

	if entry, ok := p.entries.Load(key); ok {
		e := entry.(*ResourceEntry)
		e.Reference++
		return false, e.URI
	}

	entry := &ResourceEntry{
		URI:       uri,
		Hash:      hash,
		Size:      size,
		Reference: 1,
	}
	p.entries.Store(key, entry)
	return true, uri
}

// Lookup looks up a resource by hash.
func (p *ResourceDedupPool) Lookup(hash string) (*ResourceEntry, bool) {
	if entry, ok := p.entries.Load(ResourceHashKey(hash)); ok {
		return entry.(*ResourceEntry), true
	}
	return nil, false
}

// Release decrements the reference count for the resource with the given hash.
func (p *ResourceDedupPool) Release(hash string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if entry, ok := p.entries.Load(ResourceHashKey(hash)); ok {
		e := entry.(*ResourceEntry)
		e.Reference--
		if e.Reference <= 0 {
			p.entries.Delete(ResourceHashKey(hash))
		}
	}
}

// Clear removes all entries from the pool.
func (p *ResourceDedupPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.entries = sync.Map{}
}

// Stats returns the count and total size of all registered resources.
func (p *ResourceDedupPool) Stats() (count int, totalSize int64) {
	p.entries.Range(func(key, value interface{}) bool {
		count++
		entry := value.(*ResourceEntry)
		totalSize += entry.Size
		return true
	})
	return
}

// ===== Concurrent ZIP collector =====

// ConcurrentZipCollector collects part data from a channel and writes it to a ZIP archive concurrently.
type ConcurrentZipCollector struct {
	zipWriter  *zip.Writer
	dataChan   PartDataChannel
	errorChan  chan error
	doneChan   chan struct{}
	wg         sync.WaitGroup
	bufferSize int
}

// NewConcurrentZipCollector creates a concurrent ZIP collector.
func NewConcurrentZipCollector(w io.Writer, bufferSize int) *ConcurrentZipCollector {
	return &ConcurrentZipCollector{
		zipWriter:  zip.NewWriter(w),
		dataChan:   make(PartDataChannel, bufferSize),
		errorChan:  make(chan error, 1),
		doneChan:   make(chan struct{}),
		bufferSize: bufferSize,
	}
}

// Start launches the collector goroutine.
func (c *ConcurrentZipCollector) Start() {
	c.wg.Add(1)
	go c.collect()
}

// collect is the collector goroutine.
func (c *ConcurrentZipCollector) collect() {
	defer c.wg.Done()

	for data := range c.dataChan {
		if data.Error != nil {
			c.errorChan <- data.Error
			return
		}

		// Write the ZIP entry.
		if err := c.writePart(data); err != nil {
			c.errorChan <- err
			return
		}
	}

	// All data written — close the ZIP.
	if err := c.zipWriter.Close(); err != nil {
		c.errorChan <- err
		return
	}

	close(c.doneChan)
}

// writePart writes a single part to the ZIP archive.
func (c *ConcurrentZipCollector) writePart(data *PartData) error {
	path := strings.TrimPrefix(data.Path, "/")

	// Use FileHeader to set correct timestamps.
	header := &zip.FileHeader{
		Name:     path,
		Modified: time.Now(),
		Method:   zip.Deflate,
	}
	header.Flags |= 0x800 // UTF-8 file name flag

	w, err := c.zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry %s: %w", path, err)
	}

	if data.Data != nil {
		_, err = w.Write(data.Data)
		return err
	}

	if data.Source != nil {
		rc, err := data.Source.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		_, err = io.Copy(w, rc)
		return err
	}

	return nil
}

// Submit submits part data to the collector.
func (c *ConcurrentZipCollector) Submit(data *PartData) error {
	select {
	case c.dataChan <- data:
		return nil
	case err := <-c.errorChan:
		return err
	case <-c.doneChan:
		return fmt.Errorf("collector already finished")
	}
}

// SubmitBytes submits a byte slice as a ZIP entry.
func (c *ConcurrentZipCollector) SubmitBytes(path string, data []byte) error {
	return c.Submit(&PartData{
		Path: path,
		Data: data,
	})
}

// Close signals that no more data will be submitted and waits for the collector to finish.
func (c *ConcurrentZipCollector) Close() error {
	close(c.dataChan)

	select {
	case <-c.doneChan:
		return nil
	case err := <-c.errorChan:
		return err
	}
}

// Wait waits for the collector to finish (alias for Close).
func (c *ConcurrentZipCollector) Wait() error {
	return c.Close()
}

// DataChannel returns the data channel (for external producers).
func (c *ConcurrentZipCollector) DataChannel() PartDataChannel {
	return c.dataChan
}

// ===== Helper types =====

// bytesReaderAt is a simple byte-slice reader.
type bytesReaderAt struct {
	data []byte
	pos  int
}

func (r *bytesReaderAt) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
