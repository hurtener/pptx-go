# OPC Package Documentation

> **[Feature Review Report](./REVIEW.md)** - View the complete feature review results for the OPC layer

---

package opc // import "github.com/hurtener/pptx-go/opc"

Package opc provides a Go implementation of the OOXML Open Packaging Convention (OPC)
for working with Office Open XML file formats such as PPTX.

CONSTANTS

const (
	// OPC relationship type
	ContentTypeRelationships = "application/vnd.openxmlformats-package.relationships+xml"

	// PPTX core content types
	ContentTypePresentation   = "application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"
	ContentTypeSlide          = "application/vnd.openxmlformats-officedocument.presentationml.slide+xml"
	ContentTypeSlideLayout    = "application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"
	ContentTypeSlideMaster    = "application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"
	ContentTypeNotesSlide     = "application/vnd.openxmlformats-officedocument.presentationml.notesSlide+xml"
	ContentTypeHandoutMaster  = "application/vnd.openxmlformats-officedocument.presentationml.handoutMaster+xml"
	ContentTypeNotesMaster    = "application/vnd.openxmlformats-officedocument.presentationml.notesMaster+xml"
	ContentTypePresentationML = "application/vnd.openxmlformats-officedocument.presentationml.template.main+xml"

	// Themes and styles
	ContentTypeTheme         = "application/vnd.openxmlformats-officedocument.theme+xml"
	ContentTypeThemeOverride = "application/vnd.openxmlformats-officedocument.themeOverride+xml"
	ContentTypeStyles        = "application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"

	// Core properties
	ContentTypeCoreProperties = "application/vnd.openxmlformats-package.core-properties+xml"

	// Extended properties
	ContentTypeExtendedProperties = "application/vnd.openxmlformats-officedocument.extended-properties+xml"

	// Custom properties
	ContentTypeCustomProperties = "application/vnd.openxmlformats-officedocument.custom-properties+xml"

	// Image content types
	ContentTypePNG  = "image/png"
	ContentTypeJPEG = "image/jpeg"
	ContentTypeGIF  = "image/gif"
	ContentTypeBMP  = "image/bmp"
	ContentTypeTIFF = "image/tiff"
	ContentTypeWMF  = "image/x-wmf"
	ContentTypeEMF  = "image/x-emf"
	ContentTypeSVG  = "image/svg+xml"

	// Audio content types
	ContentTypeWAV  = "audio/wav"
	ContentTypeMP3  = "audio/mpeg"
	ContentTypeMIDI = "audio/midi"

	// Video content types
	ContentTypeMP4 = "video/mp4"
	ContentTypeAVI = "video/x-msvideo"
	ContentTypeWMV = "video/x-ms-wmv"

	// Other
	ContentTypeXML  = "application/xml"
	ContentTypeFont = "application/x-font"

	// Default content type mapping (extension-based)
	ContentTypeDefault = "application/octet-stream"
)
    Content type constants (Content Types)

const (
	// OPC core relationships
	RelTypeCoreProperties = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"

	// Office document relationships
	RelTypeOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"

	// Extended properties
	RelTypeExtendedProperties = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties"
	RelTypeCustomProperties   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/custom-properties"

	// Slide relationships
	RelTypeSlide         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide"
	RelTypeSlideLayout   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout"
	RelTypeSlideMaster   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster"
	RelTypeNotesSlide    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesSlide"
	RelTypeNotesMaster   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/notesMaster"
	RelTypeHandoutMaster = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/handoutMaster"

	// Theme relationships
	RelTypeTheme         = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme"
	RelTypeThemeOverride = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/themeOverride"

	// Media relationships
	RelTypeImage = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"
	RelTypeAudio = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/audio"
	RelTypeVideo = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/video"
	RelTypeMedia = "http://schemas.microsoft.com/office/2007/relationships/media"

	// Hyperlinks
	RelTypeHyperlink = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"

	// Fonts
	RelTypeFont = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/font"

	// OLE objects
	RelTypeOLEObject = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/oleObject"

	// Thumbnails
	RelTypeThumbnail = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/thumbnail"

	// Styles
	RelTypeStyles = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
)
    Relationship type constants (Relationship Types)

const (
	NamespaceOPCPackage      = "http://schemas.openxmlformats.org/package/2006/content-types"
	NamespaceRelationships   = "http://schemas.openxmlformats.org/package/2006/relationships"
	NamespaceRelationshipsNs = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
)
    OPC namespaces

const (
	PathContentTypes = "[Content_Types].xml"
	PathRelsDir      = "_rels"
	PathRelsFile     = ".rels"
)
    OPC default paths


VARIABLES

var ContentTypeToExtension = map[string]string{
	ContentTypePNG:           ".png",
	ContentTypeJPEG:          ".jpg",
	ContentTypeGIF:           ".gif",
	ContentTypeBMP:           ".bmp",
	ContentTypeTIFF:          ".tiff",
	ContentTypeWMF:           ".wmf",
	ContentTypeEMF:           ".emf",
	ContentTypeSVG:           ".svg",
	ContentTypeWAV:           ".wav",
	ContentTypeMP3:           ".mp3",
	ContentTypeMIDI:          ".mid",
	ContentTypeMP4:           ".mp4",
	ContentTypeAVI:           ".avi",
	ContentTypeWMV:           ".wmv",
	ContentTypeRelationships: ".rels",
	ContentTypeXML:           ".xml",
}
    Reverse mapping from content type to file extension

var DefaultContentTypes = map[string]string{
	".xml":   ContentTypeXML,
	".rels":  ContentTypeRelationships,
	".png":   ContentTypePNG,
	".jpg":   ContentTypeJPEG,
	".jpeg":  ContentTypeJPEG,
	".gif":   ContentTypeGIF,
	".bmp":   ContentTypeBMP,
	".tiff":  ContentTypeTIFF,
	".tif":   ContentTypeTIFF,
	".wmf":   ContentTypeWMF,
	".emf":   ContentTypeEMF,
	".svg":   ContentTypeSVG,
	".wav":   ContentTypeWAV,
	".mp3":   ContentTypeMP3,
	".mid":   ContentTypeMIDI,
	".midi":  ContentTypeMIDI,
	".mp4":   ContentTypeMP4,
	".avi":   ContentTypeAVI,
	".wmv":   ContentTypeWMV,
	".font":  ContentTypeFont,
	".odttf": ContentTypeFont,
}
    Default content type mapping (extension -> content type)


FUNCTIONS

func ComputeHash(data []byte) string
    ComputeHash computes the SHA256 hash of data

func GetContentTypeByExtension(ext string) string
    GetContentTypeByExtension returns the content type for a file extension

func GetExtensionByContentType(contentType string) string
    GetExtensionByContentType returns the file extension for a content type

func IsImageContentType(contentType string) bool
    IsImageContentType reports whether the content type is an image type

func IsImmutableContentType(contentType string) bool
    IsImmutableContentType reports whether the content type represents an
    immutable resource. Immutable resources may be shared via zero-copy without
    deep copying.

func IsLargeBinaryContentType(contentType string) bool
    IsLargeBinaryContentType reports whether the content type represents a large
    binary blob. Used to determine whether zero-copy optimization is worthwhile.

func IsMediaContentType(contentType string) bool
    IsMediaContentType reports whether the content type is an audio or video type

func IsValidPackURI(uri string) bool
    IsValidPackURI reports whether the URI is a valid pack URI

func NormalizeURI(uri string) string
    NormalizeURI normalizes a URI


TYPES

type BytesReader struct {
	// Has unexported fields.
}
    BytesReader is a simple bytes reader implementation

func NewBytesReader(data []byte) *BytesReader
    NewBytesReader creates a new BytesReader

func (r *BytesReader) Read(p []byte) (n int, err error)
    Read implements io.Reader

type BytesSource struct {
	// Has unexported fields.
}
    BytesSource is an in-memory byte data source

func NewBytesSource(data []byte) *BytesSource
    NewBytesSource creates a data source from a byte slice

func (s *BytesSource) Open() (io.ReadCloser, error)
    Open returns a bytes.Reader

func (s *BytesSource) Size() int64
    Size returns the data size

type ConcurrentZipCollector struct {
	// Has unexported fields.
}
    ConcurrentZipCollector is a concurrent ZIP collector. It uses a goroutine
    to collect part data from a channel and write it to a ZIP archive.

func NewConcurrentZipCollector(w io.Writer, bufferSize int) *ConcurrentZipCollector
    NewConcurrentZipCollector creates a concurrent ZIP collector

func (c *ConcurrentZipCollector) Close() error
    Close closes the collector and waits for all data to be written

func (c *ConcurrentZipCollector) DataChannel() PartDataChannel
    DataChannel returns the data channel (for external producers)

func (c *ConcurrentZipCollector) Start()
    Start launches the collector goroutine

func (c *ConcurrentZipCollector) Submit(data *PartData) error
    Submit submits part data to the collector

func (c *ConcurrentZipCollector) SubmitBytes(path string, data []byte) error
    SubmitBytes submits byte data

func (c *ConcurrentZipCollector) Wait() error
    Wait waits for the collector to finish

type ContentTypes struct {
	// Has unexported fields.
}
    ContentTypes represents the content of [Content_Types].xml.
    It defines the content type for every part in the package.

func NewContentTypes() *ContentTypes
    NewContentTypes creates a new content type definition

func (ct *ContentTypes) AddDefault(extension, contentType string)
    AddDefault adds a default content type mapping

func (ct *ContentTypes) AddOverride(uri *PackURI, contentType string)
    AddOverride adds a content type override for a specific URI

func (ct *ContentTypes) Defaults() map[string]string
    Defaults returns all default content type mappings

func (ct *ContentTypes) FromXML(data []byte) error
    FromXML parses content types from XML

func (ct *ContentTypes) GetContentType(uri *PackURI) string
    GetContentType returns the content type for the given URI.
    Overrides are checked first, then defaults.

func (ct *ContentTypes) GetDefault(extension string) string
    GetDefault returns the default content type for a file extension

func (ct *ContentTypes) GetOverride(uri *PackURI) string
    GetOverride returns the content type override for a URI

func (ct *ContentTypes) Overrides() map[string]string
    Overrides returns all content type overrides

func (ct *ContentTypes) RemoveOverride(uri *PackURI)
    RemoveOverride removes a content type override

func (ct *ContentTypes) ToXML() ([]byte, error)
    ToXML serializes the content types to XML

type ContentTypesStreamer struct {
	// Has unexported fields.
}
    ContentTypesStreamer is a streaming writer for ContentTypes

func NewContentTypesStreamer(ct *ContentTypes) *ContentTypesStreamer
    NewContentTypesStreamer creates a ContentTypes streaming writer

func (cs *ContentTypesStreamer) StreamWriteTo(w io.Writer) error
    StreamWriteTo implements StreamWriter

type CoreProperties struct {
	// Has unexported fields.
}
    CoreProperties represents the core properties of a package (Dublin Core metadata)

func (cp *CoreProperties) Category() string
    Category returns the category

func (cp *CoreProperties) ContentType() string
    ContentType returns the content type

func (cp *CoreProperties) Created() string
    Created returns the creation timestamp

func (cp *CoreProperties) Creator() string
    Creator returns the creator

func (cp *CoreProperties) Description() string
    Description returns the description

func (cp *CoreProperties) FromXML(data []byte) error
    FromXML parses core properties from XML

func (cp *CoreProperties) Keywords() string
    Keywords returns the keywords

func (cp *CoreProperties) Language() string
    Language returns the language

func (cp *CoreProperties) LastModifiedBy() string
    LastModifiedBy returns the last-modified-by user

func (cp *CoreProperties) Modified() string
    Modified returns the last-modified timestamp

func (cp *CoreProperties) Revision() string
    Revision returns the revision number

func (cp *CoreProperties) SetCategory(category string)
    SetCategory sets the category

func (cp *CoreProperties) SetContentType(contentType string)
    SetContentType sets the content type

func (cp *CoreProperties) SetCreated(created string)
    SetCreated sets the creation timestamp

func (cp *CoreProperties) SetCreator(creator string)
    SetCreator sets the creator

func (cp *CoreProperties) SetDescription(description string)
    SetDescription sets the description

func (cp *CoreProperties) SetKeywords(keywords string)
    SetKeywords sets the keywords

func (cp *CoreProperties) SetLanguage(language string)
    SetLanguage sets the language

func (cp *CoreProperties) SetLastModifiedBy(lastModifiedBy string)
    SetLastModifiedBy sets the last-modified-by user

func (cp *CoreProperties) SetModified(modified string)
    SetModified sets the last-modified timestamp

func (cp *CoreProperties) SetRevision(revision string)
    SetRevision sets the revision number

func (cp *CoreProperties) SetSubject(subject string)
    SetSubject sets the subject

func (cp *CoreProperties) SetTitle(title string)
    SetTitle sets the title

func (cp *CoreProperties) Subject() string
    Subject returns the subject

func (cp *CoreProperties) Title() string
    Title returns the title

func (cp *CoreProperties) ToXML() ([]byte, error)
    ToXML serializes the core properties to XML

type DefaultPartFactory struct{}
    DefaultPartFactory is the default part factory

func (f *DefaultPartFactory) CreatePart(uri *PackURI, contentType string, blob []byte) (*Part, error)
    CreatePart implements PartFactory

type PackURI struct {
	// Has unexported fields.
}
    PackURI represents a URI within the package (e.g. /ppt/slides/slide1.xml).
    It follows the URI rules defined in the OPC specification.

func ContentTypesURI() *PackURI
    ContentTypesURI returns the URI for [Content_Types].xml

func NewPackURI(uri string) *PackURI
    NewPackURI creates a new PackURI

func PackageRelsURI() *PackURI
    PackageRelsURI returns the URI of the package-level relationships file

func RootURI() *PackURI
    RootURI returns the package root URI

func (p *PackURI) BaseName() string
    BaseName returns the base name of the URI (without extension).
    Example: /ppt/slides/slide1.xml -> slide1

func (p *PackURI) Clone() *PackURI
    Clone creates a copy of the PackURI

func (p *PackURI) DirName() string
    DirName returns the parent directory path.
    Example: /ppt/slides/slide1.xml -> /ppt/slides

func (p *PackURI) DirSegments() []string
    DirSegments returns a slice of directory segments (excluding the file name)

func (p *PackURI) Equals(other *PackURI) bool
    Equals reports whether two PackURIs are equal

func (p *PackURI) EqualsStr(uri string) bool
    EqualsStr reports whether the PackURI equals a string URI

func (p *PackURI) Extension() string
    Extension returns the file extension.
    Example: /ppt/slides/slide1.xml -> .xml

func (p *PackURI) FileName() string
    FileName returns the file name including extension.
    Example: /ppt/slides/slide1.xml -> slide1.xml

func (p *PackURI) IsAbsolute() bool
    IsAbsolute reports whether the URI is an absolute path

func (p *PackURI) IsPackageRels() bool
    IsPackageRels reports whether this is the package-level relationships file

func (p *PackURI) IsRelationshipsPart() bool
    IsRelationshipsPart reports whether this is a relationships part

func (p *PackURI) Join(relativePath string) *PackURI
    Join resolves a relative path against this URI and returns a new PackURI

func (p PackURI) MarshalText() ([]byte, error)
    MarshalText implements encoding.TextMarshaler

func (p *PackURI) MemberName() string
    MemberName returns the member name relative to the package root (used for
    ZIP entries). Strips the leading /.

func (p *PackURI) RelPath() string
    RelPath returns the relative path (used as a relationship target)

func (p *PackURI) RelPathFrom(other *PackURI) string
    RelPathFrom computes the relative path from another URI to this URI

func (p *PackURI) RelationshipsURI() *PackURI
    RelationshipsURI returns the URI of the relationships file for this part.
    Example: /ppt/slides/slide1.xml -> /ppt/slides/_rels/slide1.xml.rels

func (p *PackURI) Segments() []string
    Segments returns the individual segments of the URI

func (p *PackURI) SourceURI() *PackURI
    SourceURI derives the source part URI from a relationships file path.
    Example: /ppt/slides/_rels/slide1.xml.rels -> /ppt/slides/slide1.xml

func (p *PackURI) String() string
    String returns the string representation of the URI

func (p *PackURI) URI() string
    URI returns the raw URI string

func (p *PackURI) UnmarshalText(data []byte) error
    UnmarshalText implements encoding.TextUnmarshaler

type Package struct {
	// Has unexported fields.
}
    Package represents an OPC package (e.g. a PPTX file)

func NewPackage() *Package
    NewPackage creates a new empty OPC package

func Open(r io.ReaderAt, size int64) (*Package, error)
    Open opens an OPC package from a ZIP stream

func OpenFile(path string) (*Package, error)
    OpenFile opens an OPC package from a file path

func (p *Package) AddPart(part *Part) error
    AddPart adds a part to the package

func (p *Package) AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error)
    AddRelationship adds a package-level relationship

func (p *Package) AllParts() []*Part
    AllParts returns all parts

func (p *Package) Clone() *Package
    Clone clones the entire package (smart copy: immutable resources use
    zero-copy, mutable resources use deep copy). Uses CloneShared for
    immutable content types (images, media, themes, masters, etc.) and deep
    copy for mutable content types (slides, presentation, etc.).

func (p *Package) CloneDeep() *Package
    CloneDeep clones the entire package using a full deep copy (no zero-copy).
    Use when a completely independent copy is required.

func (p *Package) Close() error
    Close closes the package and releases resources

func (p *Package) ContainsPart(uri *PackURI) bool
    ContainsPart reports whether the part exists

func (p *Package) ContentTypes() *ContentTypes
    ContentTypes returns the content type definitions

func (p *Package) CoreProperties() *CoreProperties
    CoreProperties returns the core properties

func (p *Package) CreatePart(uri *PackURI, contentType string, blob []byte) (*Part, error)
    CreatePart creates and adds a new part

func (p *Package) CreatePartFromReader(uri *PackURI, contentType string, r io.Reader) (*Part, error)
    CreatePartFromReader creates and adds a part from a Reader

func (p *Package) CreatePartFromXML(uri *PackURI, contentType string, v interface{}) (*Part, error)
    CreatePartFromXML creates and adds a part from an XML struct

func (p *Package) DirtyParts() []*Part
    DirtyParts returns all modified parts

func (p *Package) GetPart(uri *PackURI) *Part
    GetPart returns the part for the given URI

func (p *Package) GetPartByRelType(relType string) *Part
    GetPartByRelType returns the target part for a relationship type

func (p *Package) GetPartByStr(uri string) *Part
    GetPartByStr returns the part for a string URI

func (p *Package) GetPartsByType(contentType string) []*Part
    GetPartsByType returns all parts with the given content type

func (p *Package) GetRelationship(rID string) *Relationship
    GetRelationship returns the package-level relationship for a given rID

func (p *Package) GetRelationshipsByType(relType string) []*Relationship
    GetRelationshipsByType returns all package-level relationships for a type

func (p *Package) PartCount() int
    PartCount returns the number of parts

func (p *Package) PartURIs() []*PackURI
    PartURIs returns all part URIs

func (p *Package) Parts() *PartCollection
    Parts returns the part collection

func (p *Package) Relationships() *Relationships
    Relationships returns the package-level relationships

func (p *Package) RemovePart(uri *PackURI) error
    RemovePart removes a part from the package

func (p *Package) ResolveRelationship(source *Part, relType string) *Part
    ResolveRelationship resolves a relationship between parts and returns the
    target part

func (p *Package) Save(w io.Writer) error
    Save writes the package as a ZIP archive to w

func (p *Package) SaveFile(path string) error
    SaveFile saves the package to a file

func (p *Package) SaveToBytes() ([]byte, error)
    SaveToBytes saves the package to a byte slice

func (p *Package) SetCoreProperties(props *CoreProperties)
    SetCoreProperties sets the core properties

type Part struct {
	// Has unexported fields.
}
    Part represents a part within the package. It supports two data modes:
    exclusive mode (blob) and shared mode (sharedBlob). Shared mode is used
    for zero-copy optimization of immutable resources.

func NewPart(uri *PackURI, contentType string, blob []byte) *Part
    NewPart creates a new part (exclusive data mode)

func NewPartFromReader(uri *PackURI, contentType string, r io.Reader) (*Part, error)
    NewPartFromReader creates a part from a Reader

func NewSharedPart(uri *PackURI, contentType string, sharedBlob []byte) *Part
    NewSharedPart creates a part with shared data (zero-copy, for immutable
    resources). The caller must ensure sharedBlob is not modified for the
    lifetime of the Part.

func (p *Part) AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error)
    AddRelationship adds a relationship

func (p *Part) Blob() []byte
    Blob returns the raw content. For immutable resources this returns a shared
    read-only slice; otherwise it returns an independent copy.

func (p *Part) Clone() *Part
    Clone clones the part (deep copy, for mutable content). If the original was
    shared, an independent copy is created.

func (p *Part) CloneShared() *Part
    CloneShared clones the part (zero-copy, for immutable resources). The
    caller is responsible for ensuring the Part's content will never be modified.

func (p *Part) ContentType() string
    ContentType returns the content type

func (p *Part) GetRelatedPart(rID string) *PackURI
    GetRelatedPart returns the target part URI via relationship (requires
    Package context)

func (p *Part) HasRelationships() bool
    HasRelationships reports whether the part has relationships

func (p *Part) IsDirty() bool
    IsDirty reports whether the part has been modified

func (p *Part) IsImmutable() bool
    IsImmutable reports whether this is an immutable resource

func (p *Part) LoadRelationships(data []byte) error
    LoadRelationships loads relationships from XML

func (p *Part) MarshalToBlob(v interface{}) error
    MarshalToBlob serializes v to XML and stores it in the blob

func (p *Part) PartURI() *PackURI
    PartURI returns the part URI

func (p *Part) Reader() io.Reader
    Reader returns a Reader over the content

func (p *Part) Relationships() *Relationships
    Relationships returns the relationships collection

func (p *Part) RelationshipsBlob() ([]byte, error)
    RelationshipsBlob returns the XML content of the relationships

func (p *Part) RelationshipsURI() *PackURI
    RelationshipsURI returns the URI of the relationships file

func (p *Part) RemoveRelationship(rID string) error
    RemoveRelationship removes a relationship

func (p *Part) SetBlob(blob []byte)
    SetBlob sets the content

func (p *Part) SetBlobFromReader(r io.Reader) error
    SetBlobFromReader sets the content from a Reader

func (p *Part) SetContentType(ct string)
    SetContentType sets the content type

func (p *Part) SetDirty(dirty bool)
    SetDirty sets the dirty flag

func (p *Part) SetImmutable(immutable bool)
    SetImmutable marks the part as an immutable resource

func (p *Part) Size() int
    Size returns the content size

func (p *Part) UnmarshalBlob(v interface{}) error
    UnmarshalBlob parses the XML content of the blob into v

type PartCollection struct {
	// Has unexported fields.
}
    PartCollection is a collection of parts

func NewPartCollection() *PartCollection
    NewPartCollection creates a new part collection

func (c *PartCollection) Add(part *Part) error
    Add adds a part

func (c *PartCollection) All() []*Part
    All returns all parts in insertion order

func (c *PartCollection) Clear()
    Clear clears the collection

func (c *PartCollection) Contains(uri *PackURI) bool
    Contains reports whether the collection contains the specified part

func (c *PartCollection) Count() int
    Count returns the number of parts

func (c *PartCollection) DirtyParts() []*Part
    DirtyParts returns all modified parts

func (c *PartCollection) Get(uri *PackURI) *Part
    Get returns the part for a URI

func (c *PartCollection) GetByStr(uri string) *Part
    GetByStr returns the part for a string URI

func (c *PartCollection) GetByType(contentType string) []*Part
    GetByType returns parts by content type

func (c *PartCollection) Remove(uri *PackURI) error
    Remove removes a part

func (c *PartCollection) URIs() []*PackURI
    URIs returns all part URIs

type PartData struct {
	URI         string     // Part URI
	Path        string     // Path within the ZIP archive
	ContentType string     // Content type
	Data        []byte     // Data content
	Source      PartSource // Data source (for lazy loading)
	Error       error      // Write error (if any)
}
    PartData is part data used for channel transmission

type PartDataChannel chan *PartData
    PartDataChannel is a channel type for part data

func NewPartDataChannel(bufferSize int) PartDataChannel
    NewPartDataChannel creates a part data channel

type PartFactory interface {
	// CreatePart creates a part
	CreatePart(uri *PackURI, contentType string, blob []byte) (*Part, error)
}
    PartFactory is the part factory interface

type PartIterator struct {
	// Has unexported fields.
}
    PartIterator is a part iterator

func (it *PartIterator) FilterByType(contentType string) *PartIterator
    FilterByType filters by content type

func (it *PartIterator) Next() bool
    Next advances to the next part

func (it *PartIterator) Open() (io.ReadCloser, error)
    Open opens the content stream of the current part

func (it *PartIterator) Part() *StreamPart
    Part returns the current part

type PartSource interface {
	Open() (io.ReadCloser, error)
	Size() int64
}
    PartSource is the part data source interface

type ReaderSource struct {
	// Has unexported fields.
}
    ReaderSource is an io.Reader data source

func NewReaderSource(r io.Reader, size int64) *ReaderSource
    NewReaderSource creates a data source from an io.Reader

func (s *ReaderSource) Open() (io.ReadCloser, error)
    Open returns the reader

func (s *ReaderSource) Size() int64
    Size returns the data size

type RelTypeCollection struct {
	// Has unexported fields.
}
    RelTypeCollection is a collection of relationships grouped by type

func NewRelTypeCollection() *RelTypeCollection
    NewRelTypeCollection creates a new relationship type collection

func (c *RelTypeCollection) Add(rel *Relationship)
    Add adds a relationship to the type collection

func (c *RelTypeCollection) GetByType(relType string) []*Relationship
    GetByType returns relationships by type

func (c *RelTypeCollection) Types() []string
    Types returns all relationship types

type Relatable interface {
	// PartURI returns the URI of the part
	PartURI() *PackURI
	// Relationships returns the relationships collection of the part
	Relationships() *Relationships
	// AddRelationship adds a relationship
	AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error)
}
    Relatable is the interface for parts that can hold relationships

type Relationship struct {
	// Has unexported fields.
}
    Relationship represents a relationship between two parts

func NewRelationship(rID, relType, targetURI string, isExternal bool, source *PackURI) *Relationship
    NewRelationship creates a new relationship

func (r *Relationship) Equals(other *Relationship) bool
    Equals reports whether two relationships are equal

func (r *Relationship) IsExternal() bool
    IsExternal reports whether this is an external relationship

func (r *Relationship) RID() string
    RID returns the relationship ID

func (r *Relationship) SetSource(source *PackURI)
    SetSource sets the source URI

func (r *Relationship) SourceURI() *PackURI
    SourceURI returns the source URI

func (r *Relationship) TargetMode() string
    TargetMode returns the target mode

func (r *Relationship) TargetRef() string
    TargetRef returns the target reference (relative or absolute). If a source
    part is set, returns the relative path from the source part to the target.

func (r *Relationship) TargetURI() *PackURI
    TargetURI returns the target URI

func (r *Relationship) Type() string
    Type returns the relationship type

type Relationships struct {
	// Has unexported fields.
}
    Relationships represents a collection of relationships

func NewRelationships(sourceURI *PackURI) *Relationships
    NewRelationships creates a new relationships collection

func ParseRelationshipsFromXML(data []byte, sourceURI *PackURI) (*Relationships, error)
    ParseRelationshipsFromXML parses relationships from XML data

func (rs *Relationships) Add(rel *Relationship) error
    Add adds a relationship

func (rs *Relationships) AddNew(relType, targetURI string, isExternal bool) (*Relationship, error)
    AddNew creates and adds a new relationship (allocates an ID atomically)

func (rs *Relationships) All() []*Relationship
    All returns all relationships in insertion order

func (rs *Relationships) Clone() *Relationships
    Clone clones the relationships collection

func (rs *Relationships) Contains(rID string) bool
    Contains reports whether the collection contains a relationship with the
    given rID

func (rs *Relationships) Count() int
    Count returns the number of relationships

func (rs *Relationships) FromXML(data []byte) error
    FromXML parses the relationships collection from XML

func (rs *Relationships) Get(rID string) *Relationship
    Get returns the relationship for a given rID

func (rs *Relationships) GetByTarget(targetURI *PackURI) *Relationship
    GetByTarget returns the relationship for a target URI

func (rs *Relationships) GetByType(relType string) []*Relationship
    GetByType returns all relationships for a given type

func (rs *Relationships) InitRIDCounter()
    InitRIDCounter initializes the rID counter from the maximum value found
    among existing relationships. Call this after loading relationships from XML
    to ensure newly allocated IDs do not conflict with existing ones.

func (rs *Relationships) MarshalXML(e *xml.Encoder, start xml.StartElement) error
    MarshalXML implements xml.Marshaler

func (rs *Relationships) NextRID() string
    NextRID returns the next relationship ID (preview, does not consume). This
    method is idempotent: multiple calls return the same value until AddNew
    actually uses it.

func (rs *Relationships) Remove(rID string) error
    Remove removes a relationship

func (rs *Relationships) SetSourceURI(sourceURI *PackURI)
    SetSourceURI sets the source URI

func (rs *Relationships) SourceURI() *PackURI
    SourceURI returns the source URI

func (rs *Relationships) ToXML() ([]byte, error)
    ToXML serializes the relationships collection to XML

func (rs *Relationships) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error
    UnmarshalXML implements xml.Unmarshaler

type RelationshipsStreamer struct {
	// Has unexported fields.
}
    RelationshipsStreamer is a streaming writer for relationships

func NewRelationshipsStreamer(rels *Relationships) *RelationshipsStreamer
    NewRelationshipsStreamer creates a relationships streaming writer

func (rs *RelationshipsStreamer) StreamWriteTo(w io.Writer) error
    StreamWriteTo implements StreamWriter

type ResourceDedupPool struct {
	// Has unexported fields.
}
    ResourceDedupPool is a global resource deduplication pool. It uses sync.Map
    for concurrency-safe resource deduplication.

func GetGlobalResourcePool() *ResourceDedupPool
    GetGlobalResourcePool returns the global resource pool (note: this is a
    deduplication pool, not a shared object pool)

func NewResourceDedupPool() *ResourceDedupPool
    NewResourceDedupPool creates a new resource deduplication pool

type ResourcePool struct {
	// Has unexported fields.
}
    ResourcePool is a global resource pool that manages shareable static
    resources. It uses a zero-copy strategy so that multiple Packages can share
    the same binary data. Reference counting is supported for tracking resource
    usage.

func GetGlobalPool() *ResourcePool
    GetGlobalPool returns the global resource pool

func (p *ResourcePool) CreateSharedPart(uri *PackURI, contentType string, loader func() ([]byte, error)) (*Part, error)
    CreateSharedPart creates a shared part from the resource pool. If the
    resource is not in the pool it is loaded using the loader function.

func (p *ResourcePool) GetOrLoad(uri string, contentType string, loader func() ([]byte, error)) ([]byte, error)
    GetOrLoad retrieves or loads a resource (globally unique instance,
    zero-copy). The loader function is called at most once when the resource
    does not yet exist.

func (p *ResourcePool) Prefetch(resources map[string]func() ([]byte, error)) error
    Prefetch pre-loads resources into the pool. Use this to load known resources
    ahead of time and avoid latency at render time.

func (p *ResourcePool) Release(uri string)
    Release decrements the reference count for a resource. When the count
    reaches zero the resource is removed.

func (p *ResourcePool) ReleaseAll()
    ReleaseAll releases all resources (use with caution)

func (p *ResourcePool) Stats() map[string]int
    Stats returns resource pool statistics. Returns counts for themes, masters,
    layouts, media, fonts, and total.

func (p *ResourceDedupPool) Clear()
    Clear clears the resource pool

func (p *ResourceDedupPool) Lookup(hash string) (*ResourceEntry, bool)
    Lookup looks up a resource by hash

func (p *ResourceDedupPool) Register(uri string, data []byte) (isNew bool, existingURI string)
    Register registers a resource and returns whether it is new. If the
    resource already exists, its reference count is incremented and false is
    returned.

func (p *ResourceDedupPool) RegisterWithHash(uri string, hash string, size int64) (isNew bool, existingURI string)
    RegisterWithHash registers a resource using a pre-computed hash

func (p *ResourceDedupPool) Release(hash string)
    Release decrements the reference count for a resource

func (p *ResourceDedupPool) Stats() (count int, totalSize int64)
    Stats returns resource pool statistics

type ResourceEntry struct {
	URI       string // Part URI
	Hash      string // Content hash (SHA256)
	Size      int64  // Original size
	Reference int    // Reference count
}
    ResourceEntry is a resource pool entry

type ResourceHashKey string
    ResourceHashKey is a resource hash key

type StreamPackage struct {
	// Has unexported fields.
}
    StreamPackage is a streaming OPC package with lazy loading and streaming writes

func NewStreamPackage() *StreamPackage
    NewStreamPackage creates a new streaming package

func OpenStream(path string) (*StreamPackage, error)
    OpenStream opens an OPC package in streaming mode. The file handle remains
    open to support lazy loading.

func OpenStreamFromReader(r io.ReaderAt, size int64) (*StreamPackage, error)
    OpenStreamFromReader opens a streaming package from an io.ReaderAt. The
    caller is responsible for keeping the ReaderAt valid.

func (p *StreamPackage) AddMediaPartWithDedup(uri *PackURI, contentType string, data []byte) (actualURI *PackURI, isNew bool, err error)
    AddMediaPartWithDedup adds a media part with deduplication. If the resource
    already exists, the URI of the existing part is returned and no new part is
    created.

func (p *StreamPackage) AddPart(part *StreamPart) error
    AddPart adds a streaming part

func (p *StreamPackage) AddRelationship(relType, targetURI string, isExternal bool) (*Relationship, error)
    AddRelationship adds a package-level relationship

func (p *StreamPackage) AllParts() []*StreamPart
    AllParts returns all parts

func (p *StreamPackage) ClearMediaDedupPool()
    ClearMediaDedupPool clears the media resource deduplication pool

func (p *StreamPackage) Close() error
    Close closes the package and releases resources

func (p *StreamPackage) ConcurrentStreamSave(w io.Writer, workerCount, bufferSize int) error
    ConcurrentStreamSave saves the package concurrently using a goroutine
    collector. workerCount is the number of concurrent worker goroutines;
    bufferSize is the channel buffer size.

func (p *StreamPackage) ConcurrentStreamSaveFile(path string, workerCount, bufferSize int) error
    ConcurrentStreamSaveFile saves the package to a file using concurrent streaming

func (p *StreamPackage) ContainsPart(uri *PackURI) bool
    ContainsPart reports whether the part exists

func (p *StreamPackage) ContentTypes() *ContentTypes
    ContentTypes returns the content type definitions

func (p *StreamPackage) CoreProperties() *CoreProperties
    CoreProperties returns the core properties

func (p *StreamPackage) CreatePartFromBytes(uri *PackURI, contentType string, data []byte) (*StreamPart, error)
    CreatePartFromBytes creates a part from bytes (loaded immediately into memory)

func (p *StreamPackage) CreateStreamPart(uri *PackURI, contentType string, source PartSource) (*StreamPart, error)
    CreateStreamPart creates and adds a streaming part

func (p *StreamPackage) GetMediaDedupStats() (count int, totalSize int64)
    GetMediaDedupStats returns media resource deduplication statistics

func (p *StreamPackage) GetPart(uri *PackURI) *StreamPart
    GetPart returns a part (content is loaded on demand)

func (p *StreamPackage) GetPartByRelType(relType string) *StreamPart
    GetPartByRelType returns the target part for a relationship type

func (p *StreamPackage) GetPartByStr(uri string) *StreamPart
    GetPartByStr returns the part for a string URI

func (p *StreamPackage) GetPartsByType(contentType string) []*StreamPart
    GetPartsByType returns parts by content type

func (p *StreamPackage) NewPartIterator() *PartIterator
    NewPartIterator creates a part iterator

func (p *StreamPackage) PartCount() int
    PartCount returns the number of parts

func (p *StreamPackage) PartURIs() []*PackURI
    PartURIs returns all part URIs

func (p *StreamPackage) RegisterMediaWithDedup(uri string, data []byte) (isNew bool, existingURI string)
    RegisterMediaWithDedup registers a media resource with deduplication.
    Returns: isNew indicates whether it is a new resource; existingURI is the
    URI of the already-existing resource.

func (p *StreamPackage) Relationships() *Relationships
    Relationships returns the package-level relationships

func (p *StreamPackage) RemovePart(uri *PackURI) error
    RemovePart removes a part

func (p *StreamPackage) SetCoreProperties(props *CoreProperties)
    SetCoreProperties sets the core properties

func (p *StreamPackage) StreamSave(w io.Writer) error
    StreamSave streams the package to an io.Writer

func (p *StreamPackage) StreamSaveFile(path string) error
    StreamSaveFile streams the package to a file

type StreamPart struct {
	// Has unexported fields.
}
    StreamPart is a streaming part with lazy loading support

func NewStreamPart(uri *PackURI, contentType string, source PartSource) *StreamPart
    NewStreamPart creates a streaming part

func (p *StreamPart) Blob() ([]byte, error)
    Blob returns the content (loading it first if not already loaded)

func (p *StreamPart) Clone() *StreamPart
    Clone clones the part

func (p *StreamPart) ContentType() string
    ContentType returns the content type

func (p *StreamPart) HasRelationships() bool
    HasRelationships reports whether the part has relationships

func (p *StreamPart) IsDirty() bool
    IsDirty reports whether the part has been modified

func (p *StreamPart) IsLoaded() bool
    IsLoaded reports whether the content has been loaded into memory

func (p *StreamPart) Load() error
    Load loads the content into memory

func (p *StreamPart) LoadRelationships(data []byte) error
    LoadRelationships loads relationships from XML

func (p *StreamPart) Open() (io.ReadCloser, error)
    Open opens the part content stream

func (p *StreamPart) PartURI() *PackURI
    PartURI returns the part URI

func (p *StreamPart) Relationships() *Relationships
    Relationships returns the relationships collection

func (p *StreamPart) RelationshipsBlob() ([]byte, error)
    RelationshipsBlob returns the XML content of the relationships

func (p *StreamPart) RelationshipsURI() *PackURI
    RelationshipsURI returns the URI of the relationships file

func (p *StreamPart) SetBlob(data []byte)
    SetBlob sets the content

func (p *StreamPart) SetBlobFromReader(r io.Reader) error
    SetBlobFromReader sets the content from a Reader

func (p *StreamPart) SetContentType(ct string)
    SetContentType sets the content type

func (p *StreamPart) SetDirty(dirty bool)
    SetDirty sets the dirty flag

func (p *StreamPart) Size() int64
    Size returns the content size

func (p *StreamPart) UnmarshalBlob(v any) error
    UnmarshalBlob parses the XML content from the blob

type StreamWriter interface {
	StreamWriteTo(w io.Writer) error
}
    StreamWriter is the streaming writer interface

type StreamingZipWriter struct {
	// Has unexported fields.
}
    StreamingZipWriter is a streaming ZIP writer

func NewStreamingZipWriter(w io.Writer) *StreamingZipWriter
    NewStreamingZipWriter creates a streaming ZIP writer

func (sw *StreamingZipWriter) Close() error
    Close closes the ZIP writer

func (sw *StreamingZipWriter) Create(path string) (io.Writer, error)
    Create creates a ZIP entry and returns a writer

func (sw *StreamingZipWriter) WriteBytes(path string, data []byte) error
    WriteBytes writes byte data

func (sw *StreamingZipWriter) WriteFromReader(path string, reader io.Reader) error
    WriteFromReader streams a ZIP entry from an io.Reader

func (sw *StreamingZipWriter) WriteFromStreamer(path string, streamer StreamWriter) error
    WriteFromStreamer streams a ZIP entry from a StreamWriter

func (sw *StreamingZipWriter) WriteFromXMLStreamer(path string, streamer XMLStreamer) error
    WriteFromXMLStreamer streams a ZIP entry from an XMLStreamer

func (sw *StreamingZipWriter) WriteStreamPart(part *StreamPart) error
    WriteStreamPart streams a StreamPart into a ZIP entry

func (sw *StreamingZipWriter) WriteXML(path string, data []byte) error
    WriteXML writes XML data (adds XML declaration automatically)

type XContentTypes struct {
	XMLName   xml.Name    `xml:"Types"`
	Xmlns     string      `xml:"xmlns,attr"`
	Defaults  []XDefault  `xml:"Default"`
	Overrides []XOverride `xml:"Override"`
}
    XContentTypes is the XML-serialisable root element for content types

type XCoreProperties struct {
	XMLName        xml.Name   `xml:"coreProperties"`
	XmlnsDc        string     `xml:"xmlns:dc,attr"`
	XmlnsDcterms   string     `xml:"xmlns:dcterms,attr"`
	XmlnsDcmitype  string     `xml:"xmlns:dcmitype,attr"`
	XmlnsXsi       string     `xml:"xmlns:xsi,attr"`
	XmlnsCore      string     `xml:"xmlns,attr"`
	Title          string     `xml:"dc:title"`
	Creator        string     `xml:"dc:creator"`
	Subject        string     `xml:"dc:subject"`
	Description    string     `xml:"dc:description"`
	Keywords       *XKeywords `xml:"cp:keywords"`
	Created        *XDate     `xml:"dcterms:created"`
	Modified       *XDate     `xml:"dcterms:modified"`
	LastModifiedBy string     `xml:"cp:lastModifiedBy"`
	Revision       string     `xml:"cp:revision"`
	Category       string     `xml:"cp:category"`
	ContentType    string     `xml:"cp:contentType"`
}
    XCoreProperties is the XML-serialisable core properties element

type XDate struct {
	Type  string `xml:"xsi:type,attr"`
	Value string `xml:",chardata"`
}
    XDate is a date element

type XDefault struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}
    XDefault is the XML-serialisable default content type element

type XKeywords struct {
	Value string `xml:",chardata"`
}
    XKeywords is the keywords element

type XMLStreamer interface {
	StreamXML(enc *xml.Encoder) error
}
    XMLStreamer is the XML streaming writer interface

type XOverride struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}
    XOverride is the XML-serialisable content type override element

type XRelationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr,omitempty"`
}
    XRelationship is the XML-serialisable relationship element

type XRelationships struct {
	XMLName       xml.Name        `xml:"Relationships"`
	Xmlns         string          `xml:"xmlns,attr"`
	Relationships []XRelationship `xml:"Relationship"`
}
    XRelationships is the XML-serialisable root element for relationships

type ZipFileSource struct {
	// Has unexported fields.
}
    ZipFileSource is a data source for a part within a ZIP file

func NewZipFileSource(f *zip.File) *ZipFileSource
    NewZipFileSource creates a data source from a zip.File

func (s *ZipFileSource) Open() (io.ReadCloser, error)
    Open opens the ZIP file entry

func (s *ZipFileSource) Size() int64
    Size returns the uncompressed size
