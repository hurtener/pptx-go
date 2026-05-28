// Package ooxml holds shared OOXML helpers used across the per-part-family
// subpackages (presentation, slide, theme, core, chart, relations, media).
//
// Per RFC §6.2 the part families stay independent: a subpackage may import
// this root package for shared helpers (canonical namespace URIs,
// StripNamespacePrefixes, the XML declaration) but must not import another
// family's XML types. This package imports no subpackage, keeping the
// dependency graph acyclic.
package ooxml

// Canonical OOXML / OPC namespace URIs (ISO/IEC 29500 transitional profile,
// the shape PowerPoint emits — RFC §6.3). These are the single source of
// truth for namespace strings; codecs reference them as they are migrated
// off hard-coded literals.
const (
	// OPC package-level namespaces.
	NSPackageRelationships = "http://schemas.openxmlformats.org/package/2006/relationships"
	NSPackageContentTypes  = "http://schemas.openxmlformats.org/package/2006/content-types"
	NSCoreProperties       = "http://schemas.openxmlformats.org/package/2006/metadata/core-properties"

	// officeDocument shared namespaces.
	NSRelationships     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	NSExtendedProps     = "http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"
	NSDocPropsVTypes    = "http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"
	NSDublinCore        = "http://purl.org/dc/elements/1.1/"
	NSDublinCoreTerms   = "http://purl.org/dc/terms/"
	NSDublinCoreMIType  = "http://purl.org/dc/dcmitype/"
	NSXMLSchemaInstance = "http://www.w3.org/2001/XMLSchema-instance"

	// DrawingML.
	NSDrawingML      = "http://schemas.openxmlformats.org/drawingml/2006/main"
	NSDrawingMLChart = "http://schemas.openxmlformats.org/drawingml/2006/chart"

	// PresentationML.
	NSPresentationML = "http://schemas.openxmlformats.org/presentationml/2006/main"
)

// Relationship-type URIs are owned by the relations subpackage
// (internal/ooxml/relations) for now; codecs reference them there. They
// consolidate here as families migrate off their local copies.
