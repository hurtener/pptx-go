package opc

import (
	"path"
	"strings"
)

// PackURI represents a URI within a package (e.g. /ppt/slides/slide1.xml).
// It follows the URI rules defined in the OPC specification.
type PackURI struct {
	uri      string
	segments []string
}

// NewPackURI creates a new PackURI.
func NewPackURI(uri string) *PackURI {
	// Normalize: ensure the URI starts with /.
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}
	// Collapse duplicate slashes.
	for strings.Contains(uri, "//") {
		uri = strings.ReplaceAll(uri, "//", "/")
	}

	p := &PackURI{
		uri: uri,
	}
	p.segments = p.parseSegments()
	return p
}

// parseSegments splits the URI into its path segments.
func (p *PackURI) parseSegments() []string {
	trimmed := strings.Trim(p.uri, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

// URI returns the raw URI string.
func (p *PackURI) URI() string {
	return p.uri
}

// String returns the URI string representation.
func (p *PackURI) String() string {
	return p.uri
}

// BaseName returns the base name of the URI without its extension.
// Example: /ppt/slides/slide1.xml -> slide1
func (p *PackURI) BaseName() string {
	filename := p.FileName()
	ext := path.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// FileName returns the file name including the extension.
// Example: /ppt/slides/slide1.xml -> slide1.xml
func (p *PackURI) FileName() string {
	return path.Base(p.uri)
}

// Extension returns the file extension.
// Example: /ppt/slides/slide1.xml -> .xml
func (p *PackURI) Extension() string {
	return path.Ext(p.uri)
}

// DirName returns the parent directory path.
// Example: /ppt/slides/slide1.xml -> /ppt/slides
func (p *PackURI) DirName() string {
	return path.Dir(p.uri)
}

// Segments returns the individual path segments of the URI.
func (p *PackURI) Segments() []string {
	return p.segments
}

// MemberName returns the member name relative to the package root (suitable for use in ZIP archives).
// The leading / is removed.
func (p *PackURI) MemberName() string {
	return strings.TrimPrefix(p.uri, "/")
}

// RelPath returns the relative path (used as a relationship target).
func (p *PackURI) RelPath() string {
	return p.MemberName()
}

// Join resolves a relative path against this URI and returns a new PackURI.
func (p *PackURI) Join(relativePath string) *PackURI {
	// Absolute path: use as-is.
	if strings.HasPrefix(relativePath, "/") {
		return NewPackURI(relativePath)
	}

	// Resolve relative path.
	result := p.DirName()
	parts := strings.Split(relativePath, "/")

	for _, part := range parts {
		if part == ".." {
			// Go up one directory level.
			if result != "/" && result != "" {
				result = path.Dir(result)
			}
		} else if part != "." && part != "" {
			// Append the path segment.
			result = path.Join(result, part)
		}
	}

	return NewPackURI(result)
}

// RelPathFrom computes the relative path from another URI to this URI.
func (p *PackURI) RelPathFrom(other *PackURI) string {
	fromSegs := other.DirSegments()
	toSegs := p.DirSegments()

	// Find the common prefix length.
	commonLen := 0
	minLen := len(fromSegs)
	if len(toSegs) < minLen {
		minLen = len(toSegs)
	}

	for i := 0; i < minLen; i++ {
		if fromSegs[i] == toSegs[i] {
			commonLen++
		} else {
			break
		}
	}

	// Build the relative path.
	var upLevels []string
	for i := commonLen; i < len(fromSegs); i++ {
		upLevels = append(upLevels, "..")
	}

	var downPath []string
	for i := commonLen; i < len(toSegs); i++ {
		downPath = append(downPath, toSegs[i])
	}
	downPath = append(downPath, p.FileName())

	relativeParts := append(upLevels, downPath...)
	if len(relativeParts) == 0 {
		return p.FileName()
	}
	return strings.Join(relativeParts, "/")
}

// DirSegments returns the directory segments of the URI (excluding the file name).
func (p *PackURI) DirSegments() []string {
	dir := p.DirName()
	if dir == "/" || dir == "." {
		return []string{}
	}
	return strings.Split(strings.Trim(dir, "/"), "/")
}

// IsRelationshipsPart reports whether this URI refers to a relationships part.
func (p *PackURI) IsRelationshipsPart() bool {
	return strings.HasSuffix(p.FileName(), ".rels")
}

// RelationshipsURI returns the URI of the relationships file for this part.
// Example: /ppt/slides/slide1.xml -> /ppt/slides/_rels/slide1.xml.rels
func (p *PackURI) RelationshipsURI() *PackURI {
	if p.IsRelationshipsPart() {
		return p
	}

	dir := p.DirName()
	filename := p.FileName()

	// Build the relationships file path.
	relPath := path.Join(dir, PathRelsDir, filename+".rels")
	return NewPackURI(relPath)
}

// SourceURI returns the source part URI from a relationships file path.
// Example: /ppt/slides/_rels/slide1.xml.rels -> /ppt/slides/slide1.xml
func (p *PackURI) SourceURI() *PackURI {
	if !p.IsRelationshipsPart() {
		return p
	}

	// Parse the relationships file path.
	dir := p.DirName()
	filename := p.FileName()

	// Verify the file is inside a _rels directory.
	if path.Base(dir) != PathRelsDir {
		return p
	}

	// Remove the _rels directory and the .rels extension.
	parentDir := path.Dir(dir)
	sourceFilename := strings.TrimSuffix(filename, ".rels")

	return NewPackURI(path.Join(parentDir, sourceFilename))
}

// Equals reports whether two PackURIs are equal.
func (p *PackURI) Equals(other *PackURI) bool {
	if other == nil {
		return false
	}
	return p.uri == other.uri
}

// EqualsStr reports whether the URI equals the given string URI.
func (p *PackURI) EqualsStr(uri string) bool {
	other := NewPackURI(uri)
	return p.Equals(other)
}

// IsAbsolute reports whether the URI is an absolute path.
func (p *PackURI) IsAbsolute() bool {
	return strings.HasPrefix(p.uri, "/")
}

// Clone returns a copy of this PackURI.
func (p *PackURI) Clone() *PackURI {
	return NewPackURI(p.uri)
}

// MarshalText implements encoding.TextMarshaler.
func (p PackURI) MarshalText() ([]byte, error) {
	return []byte(p.uri), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *PackURI) UnmarshalText(data []byte) error {
	p.uri = string(data)
	if !strings.HasPrefix(p.uri, "/") {
		p.uri = "/" + p.uri
	}
	p.segments = p.parseSegments()
	return nil
}

// RootURI returns the package root URI.
func RootURI() *PackURI {
	return NewPackURI("/")
}

// ContentTypesURI returns the URI of [Content_Types].xml.
func ContentTypesURI() *PackURI {
	return NewPackURI("/" + PathContentTypes)
}

// PackageRelsURI returns the URI of the package-level relationships file.
func PackageRelsURI() *PackURI {
	return NewPackURI("/" + PathRelsDir + "/" + PathRelsFile)
}

// IsPackageRels reports whether this URI refers to the package-level relationships file.
func (p *PackURI) IsPackageRels() bool {
	return p.uri == "/"+PathRelsDir+"/"+PathRelsFile
}

// IsValidPackURI reports whether the given string is a valid pack URI.
func IsValidPackURI(uri string) bool {
	// Basic validation.
	if uri == "" {
		return false
	}

	// Must start with / (absolute path).
	if !strings.HasPrefix(uri, "/") {
		return false
	}

	// Check for illegal characters.
	illegalChars := []string{"\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range illegalChars {
		if strings.Contains(uri, char) {
			return false
		}
	}

	return true
}

// NormalizeURI normalizes a URI string.
func NormalizeURI(uri string) string {
	// Convert backslashes to forward slashes. Use an explicit replace rather
	// than filepath.ToSlash, which is a no-op on non-Windows platforms.
	uri = strings.ReplaceAll(uri, "\\", "/")

	// Ensure a leading slash.
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}

	// Collapse duplicate slashes.
	for strings.Contains(uri, "//") {
		uri = strings.ReplaceAll(uri, "//", "/")
	}

	// Remove trailing slash (except for the root).
	if len(uri) > 1 && strings.HasSuffix(uri, "/") {
		uri = strings.TrimSuffix(uri, "/")
	}

	return uri
}

// NormalizeZipPath normalizes an internal ZIP path.
// Designed specifically for paths read from ZIP archives, handling Windows backslash issues.
// Unlike NormalizeURI, this function does not add a leading slash — it preserves the relative form.
func NormalizeZipPath(path string) string {
	// 1. Convert Windows backslashes to forward slashes.
	//    The ZIP spec requires forward slashes, but some tools on Windows may produce backslashes.
	path = strings.ReplaceAll(path, "\\", "/")

	// 2. Collapse duplicate slashes (e.g. // or ///).
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// 3. Remove any leading slash (ZIP internal paths are relative).
	path = strings.TrimPrefix(path, "/")

	// 4. Remove any trailing slash (unless the path is empty).
	if len(path) > 0 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	return path
}
