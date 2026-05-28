package opc

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

// Package represents an OPC package (e.g. a PPTX file).
type Package struct {
	parts          *PartCollection
	relationships  *Relationships
	contentTypes   *ContentTypes
	coreProperties *CoreProperties
	mu             sync.RWMutex
}

// NewPackage creates a new, empty OPC package.
func NewPackage() *Package {
	return &Package{
		parts:         NewPartCollection(),
		relationships: NewRelationships(RootURI()),
		contentTypes:  NewContentTypes(),
	}
}

// Parts returns the part collection.
func (p *Package) Parts() *PartCollection {
	return p.parts
}

// Relationships returns the package-level relationships.
func (p *Package) Relationships() *Relationships {
	return p.relationships
}

// ContentTypes returns the content type definitions.
func (p *Package) ContentTypes() *ContentTypes {
	return p.contentTypes
}

// CoreProperties returns the core properties.
func (p *Package) CoreProperties() *CoreProperties {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.coreProperties
}

// SetCoreProperties sets the core properties.
func (p *Package) SetCoreProperties(props *CoreProperties) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.coreProperties = props
}

// ===== Opening a package =====

// Open opens an OPC package from a ZIP stream.
func Open(r io.ReaderAt, size int64) (*Package, error) {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	pkg := NewPackage()

	if err := pkg.loadContentTypes(zipReader); err != nil {
		return nil, fmt.Errorf("failed to load content types: %w", err)
	}

	if err := pkg.loadParts(zipReader); err != nil {
		return nil, fmt.Errorf("failed to load parts: %w", err)
	}

	if err := pkg.loadRelationships(zipReader); err != nil {
		return nil, fmt.Errorf("failed to load relationships: %w", err)
	}

	for _, part := range pkg.parts.All() {
		part.SetDirty(false)
	}

	return pkg, nil
}

// OpenFile opens an OPC package from a file path.
func OpenFile(path string) (*Package, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return Open(file, stat.Size())
}

func (p *Package) loadContentTypes(zipReader *zip.Reader) error {
	var ctFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == PathContentTypes {
			ctFile = f
			break
		}
	}

	if ctFile == nil {
		return fmt.Errorf("[Content_Types].xml not found")
	}

	rc, err := ctFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open [Content_Types].xml: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read [Content_Types].xml: %w", err)
	}

	return p.contentTypes.FromXML(data)
}

func (p *Package) loadParts(zipReader *zip.Reader) error {
	for _, f := range zipReader.File {
		// Normalize path to handle Windows backslash issues.
		normalizedName := NormalizeZipPath(f.Name)

		if normalizedName == PathContentTypes {
			continue
		}
		if strings.Contains(normalizedName, PathRelsDir+"/") && strings.HasSuffix(normalizedName, ".rels") {
			continue
		}
		if strings.HasSuffix(normalizedName, "/") {
			continue
		}

		uri := NewPackURI("/" + normalizedName)

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", f.Name, err)
		}

		blob, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", f.Name, err)
		}

		contentType := p.contentTypes.GetContentType(uri)
		part := NewPart(uri, contentType, blob)
		part.SetDirty(false)

		if err := p.parts.Add(part); err != nil {
			return fmt.Errorf("failed to add part %s: %w", uri.URI(), err)
		}
	}

	return nil
}

func (p *Package) loadRelationships(zipReader *zip.Reader) error {
	for _, f := range zipReader.File {
		// Normalize path.
		normalizedName := NormalizeZipPath(f.Name)

		if !strings.Contains(normalizedName, PathRelsDir+"/") || !strings.HasSuffix(normalizedName, ".rels") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", normalizedName, err)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", normalizedName, err)
		}

		relURI := NewPackURI("/" + normalizedName)
		sourceURI := relURI.SourceURI()

		rels := NewRelationships(sourceURI)
		if err := rels.FromXML(data); err != nil {
			return fmt.Errorf("failed to parse relationships %s: %w", f.Name, err)
		}

		if relURI.IsPackageRels() {
			p.relationships = rels
		} else {
			part := p.parts.Get(sourceURI)
			if part != nil {
				part.LoadRelationships(data)
			}
		}
	}

	return nil
}

// ===== Part management =====

// AddPart adds a part to the package.
func (p *Package) AddPart(part *Part) error {
	return p.parts.Add(part)
}

// CreatePart creates and adds a new part.
func (p *Package) CreatePart(uri *PackURI, contentType string, blob []byte) (*Part, error) {
	part := NewPart(uri, contentType, blob)
	if err := p.parts.Add(part); err != nil {
		return nil, err
	}
	return part, nil
}

// CreatePartFromReader creates and adds a part from a Reader.
func (p *Package) CreatePartFromReader(uri *PackURI, contentType string, r io.Reader) (*Part, error) {
	part, err := NewPartFromReader(uri, contentType, r)
	if err != nil {
		return nil, err
	}
	if err := p.parts.Add(part); err != nil {
		return nil, err
	}
	return part, nil
}

// CreatePartFromXML creates and adds a part from an XML struct.
func (p *Package) CreatePartFromXML(uri *PackURI, contentType string, v interface{}) (*Part, error) {
	data, err := xml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML: %w", err)
	}
	data = append([]byte(XMLDeclaration), data...)
	return p.CreatePart(uri, contentType, data)
}

// GetPart returns the part with the given URI.
func (p *Package) GetPart(uri *PackURI) *Part {
	return p.parts.Get(uri)
}

// GetPartByStr returns the part with the given string URI.
func (p *Package) GetPartByStr(uri string) *Part {
	return p.parts.GetByStr(uri)
}

// GetPartsByType returns all parts with the given content type.
func (p *Package) GetPartsByType(contentType string) []*Part {
	return p.parts.GetByType(contentType)
}

// ContainsPart reports whether a part with the given URI exists.
func (p *Package) ContainsPart(uri *PackURI) bool {
	return p.parts.Contains(uri)
}

// RemovePart removes a part from the package.
func (p *Package) RemovePart(uri *PackURI) error {
	return p.parts.Remove(uri)
}

// PartCount returns the number of parts.
func (p *Package) PartCount() int {
	return p.parts.Count()
}

// AllParts returns all parts.
func (p *Package) AllParts() []*Part {
	return p.parts.All()
}

// PartURIs returns all part URIs.
func (p *Package) PartURIs() []*PackURI {
	return p.parts.URIs()
}

// DirtyParts returns all modified parts.
func (p *Package) DirtyParts() []*Part {
	return p.parts.DirtyParts()
}

// ===== Relationship management =====

// AddRelationship adds a package-level relationship.
func (p *Package) AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error) {
	return p.relationships.AddNew(relType, targetURI, isExternal)
}

// GetRelationship returns the package-level relationship with the given rID.
func (p *Package) GetRelationship(rID string) *Relationship {
	return p.relationships.Get(rID)
}

// GetRelationshipsByType returns all package-level relationships of the given type.
func (p *Package) GetRelationshipsByType(relType string) []*Relationship {
	return p.relationships.GetByType(relType)
}

// GetPartByRelType returns the target part of the first package-level relationship with the given type.
func (p *Package) GetPartByRelType(relType string) *Part {
	rels := p.relationships.GetByType(relType)
	if len(rels) == 0 {
		return nil
	}
	return p.parts.Get(rels[0].TargetURI())
}

// ResolveRelationship resolves the target part of the first relationship of the given type from source.
func (p *Package) ResolveRelationship(source *Part, relType string) *Part {
	rels := source.Relationships().GetByType(relType)
	if len(rels) == 0 {
		return nil
	}
	return p.parts.Get(rels[0].TargetURI())
}

// ===== Saving a package =====

// Save writes the package as a ZIP archive to w.
func (p *Package) Save(w io.Writer) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	if err := p.writeContentTypes(zipWriter); err != nil {
		return fmt.Errorf("failed to write content types: %w", err)
	}

	if err := p.writePackageRelationships(zipWriter); err != nil {
		return fmt.Errorf("failed to write package relationships: %w", err)
	}

	if err := p.writeParts(zipWriter); err != nil {
		return fmt.Errorf("failed to write parts: %w", err)
	}

	if p.coreProperties != nil {
		if err := p.writeCoreProperties(zipWriter); err != nil {
			return fmt.Errorf("failed to write core properties: %w", err)
		}
	}

	return nil
}

// SaveFile saves the package to a file.
func (p *Package) SaveFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return p.Save(file)
}

// SaveToBytes saves the package to a byte slice.
func (p *Package) SaveToBytes() ([]byte, error) {
	buf := &bytesBuffer{}
	if err := p.Save(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Package) writeContentTypes(zipWriter *zip.Writer) error {
	p.updateContentTypes()

	data, err := p.contentTypes.ToXML()
	if err != nil {
		return err
	}

	return p.writeZipEntry(zipWriter, PathContentTypes, data)
}

func (p *Package) writePackageRelationships(zipWriter *zip.Writer) error {
	if p.relationships.Count() == 0 {
		return nil
	}

	data, err := p.relationships.ToXML()
	if err != nil {
		return err
	}

	relPath := PathRelsDir + "/" + PathRelsFile
	return p.writeZipEntry(zipWriter, relPath, data)
}

func (p *Package) writeParts(zipWriter *zip.Writer) error {
	for _, part := range p.parts.All() {
		filePath := strings.TrimPrefix(part.PartURI().URI(), "/")
		if err := p.writeZipEntry(zipWriter, filePath, part.Blob()); err != nil {
			return fmt.Errorf("failed to write part %s: %w", filePath, err)
		}

		if part.HasRelationships() {
			relPath := p.relFilePath(part.PartURI())
			relData, err := part.RelationshipsBlob()
			if err != nil {
				return fmt.Errorf("failed to serialize relationships for %s: %w", filePath, err)
			}
			if relData != nil {
				if err := p.writeZipEntry(zipWriter, relPath, relData); err != nil {
					return fmt.Errorf("failed to write relationships for %s: %w", filePath, err)
				}
			}
		}
	}

	return nil
}

func (p *Package) writeCoreProperties(zipWriter *zip.Writer) error {
	data, err := p.coreProperties.ToXML()
	if err != nil {
		return err
	}

	return p.writeZipEntry(zipWriter, "docProps/core.xml", data)
}

func (p *Package) writeZipEntry(zipWriter *zip.Writer, path string, data []byte) error {
	writer, err := createZipEntry(zipWriter, path, len(data))
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write zip entry %s: %w", path, err)
	}

	return nil
}

// createZipEntry creates a ZIP entry with correct timestamps and compatibility settings.
// This internal helper ensures all ZIP entries are created in a uniform, safe manner.
func createZipEntry(zipWriter *zip.Writer, path string, size int) (io.Writer, error) {
	// 1. Strip the leading slash to comply with the ZIP spec.
	//    ZIP internal paths are relative and must not start with a slash.
	path = strings.TrimPrefix(path, "/")

	// 2. Build a precise FileHeader.
	header := &zip.FileHeader{
		Name:               path,
		UncompressedSize:   uint32(size),
		UncompressedSize64: uint64(size),
		Modified:           time.Now(), // Set current timestamp (works around a Windows Explorer MS-DOS time parsing bug).
	}

	// 3. Set compression method (Deflate for text, Store for already-compressed data).
	header.Method = zip.Deflate

	// 4. Compatibility flags.
	//    - Use UTF-8 file names (Bit 11).
	//    - Important for file names that contain non-ASCII characters.
	header.Flags |= 0x800 // UTF-8 file name flag

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return nil, fmt.Errorf("failed to create zip entry %s: %w", path, err)
	}

	return writer, nil
}

func (p *Package) relFilePath(uri *PackURI) string {
	dir := path.Dir(strings.TrimPrefix(uri.URI(), "/"))
	filename := path.Base(uri.URI())
	return path.Join(dir, PathRelsDir, filename+".rels")
}

func (p *Package) updateContentTypes() {
	for _, part := range p.parts.All() {
		uri := part.PartURI()
		contentType := part.ContentType()

		if uri.IsRelationshipsPart() {
			contentType = ContentTypeRelationships
		}

		ext := uri.Extension()
		defaultCT := p.contentTypes.GetDefault(ext)

		if contentType != "" && contentType != ContentTypeDefault {
			if defaultCT == "" || defaultCT == ContentTypeDefault || defaultCT != contentType {
				p.contentTypes.AddOverride(uri, contentType)
			}
		}
	}
}

// ===== Other methods =====

// Clone clones the entire package (smart copy: zero-copy for static resources, deep copy for dynamic ones).
func (p *Package) Clone() *Package {
	p.mu.RLock()
	defer p.mu.RUnlock()

	newPkg := NewPackage()

	for _, part := range p.parts.All() {
		var newPart *Part

		// Choose the copy strategy based on the content type.
		if IsImmutableContentType(part.ContentType()) {
			// Immutable resource: use zero-copy.
			newPart = part.CloneShared()
		} else {
			// Mutable resource: use deep copy.
			newPart = part.Clone()
		}
		_ = newPkg.parts.Add(newPart)
	}

	newPkg.relationships = p.relationships.Clone()

	newPkg.contentTypes = &ContentTypes{
		defaults:  make(map[string]string),
		overrides: make(map[string]string),
	}
	for k, v := range p.contentTypes.Defaults() {
		newPkg.contentTypes.AddDefault(k, v)
	}
	for k, v := range p.contentTypes.Overrides() {
		newPkg.contentTypes.AddOverride(NewPackURI(k), v)
	}

	if p.coreProperties != nil {
		newPkg.coreProperties = &CoreProperties{}
		newPkg.coreProperties.SetTitle(p.coreProperties.Title())
		newPkg.coreProperties.SetCreator(p.coreProperties.Creator())
		newPkg.coreProperties.SetSubject(p.coreProperties.Subject())
		newPkg.coreProperties.SetDescription(p.coreProperties.Description())
		newPkg.coreProperties.SetKeywords(p.coreProperties.Keywords())
		newPkg.coreProperties.SetCreated(p.coreProperties.Created())
		newPkg.coreProperties.SetModified(p.coreProperties.Modified())
		newPkg.coreProperties.SetLastModifiedBy(p.coreProperties.LastModifiedBy())
		newPkg.coreProperties.SetRevision(p.coreProperties.Revision())
		newPkg.coreProperties.SetCategory(p.coreProperties.Category())
		newPkg.coreProperties.SetContentType(p.coreProperties.ContentType())
		newPkg.coreProperties.SetLanguage(p.coreProperties.Language())
	}

	return newPkg
}

// CloneDeep clones the entire package using a full deep copy (no zero-copy sharing).
// Use this when a fully independent replica is required.
func (p *Package) CloneDeep() *Package {
	p.mu.RLock()
	defer p.mu.RUnlock()

	newPkg := NewPackage()

	for _, part := range p.parts.All() {
		newPart := part.Clone() // Always deep copy.
		_ = newPkg.parts.Add(newPart)
	}

	newPkg.relationships = p.relationships.Clone()

	newPkg.contentTypes = &ContentTypes{
		defaults:  make(map[string]string),
		overrides: make(map[string]string),
	}
	for k, v := range p.contentTypes.Defaults() {
		newPkg.contentTypes.AddDefault(k, v)
	}
	for k, v := range p.contentTypes.Overrides() {
		newPkg.contentTypes.AddOverride(NewPackURI(k), v)
	}

	if p.coreProperties != nil {
		newPkg.coreProperties = &CoreProperties{}
		newPkg.coreProperties.SetTitle(p.coreProperties.Title())
		newPkg.coreProperties.SetCreator(p.coreProperties.Creator())
		newPkg.coreProperties.SetSubject(p.coreProperties.Subject())
		newPkg.coreProperties.SetDescription(p.coreProperties.Description())
		newPkg.coreProperties.SetKeywords(p.coreProperties.Keywords())
		newPkg.coreProperties.SetCreated(p.coreProperties.Created())
		newPkg.coreProperties.SetModified(p.coreProperties.Modified())
		newPkg.coreProperties.SetLastModifiedBy(p.coreProperties.LastModifiedBy())
		newPkg.coreProperties.SetRevision(p.coreProperties.Revision())
		newPkg.coreProperties.SetCategory(p.coreProperties.Category())
		newPkg.coreProperties.SetContentType(p.coreProperties.ContentType())
		newPkg.coreProperties.SetLanguage(p.coreProperties.Language())
	}

	return newPkg
}

// Close closes the package and releases its resources.
func (p *Package) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parts.Clear()
	p.relationships = nil

	return nil
}

// bytesBuffer is a simple bytes buffer that implements io.Writer.
type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuffer) Bytes() []byte {
	return b.data
}
