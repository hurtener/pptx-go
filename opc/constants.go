// Package opc provides a Go implementation of the OOXML Open Packaging Convention (OPC),
// used for processing Office Open XML file formats such as PPTX.
package opc

// Content type constants (Content Types)
const (
	// OPC relationship content type
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

	// Theme and styles
	ContentTypeTheme         = "application/vnd.openxmlformats-officedocument.theme+xml"
	ContentTypeThemeOverride = "application/vnd.openxmlformats-officedocument.themeOverride+xml"
	ContentTypeStyles        = "application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"

	// Charts
	ContentTypeChart   = "application/vnd.openxmlformats-officedocument.drawingml.chart+xml"
	ContentTypeChartEx = "application/vnd.ms-office.chartex+xml"

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

	// Default content type (fallback for unknown extensions)
	ContentTypeDefault = "application/octet-stream"
)

// Relationship type constants (Relationship Types)
const (
	// OPC core relationship
	RelTypeCoreProperties = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties"

	// Office document relationship
	RelTypeOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"

	// Extended and custom properties
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

// DefaultContentTypes maps file extensions to their default content types.
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

// ContentTypeToExtension maps content types to their canonical file extensions.
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

// GetContentTypeByExtension returns the content type for the given file extension.
func GetContentTypeByExtension(ext string) string {
	if ct, ok := DefaultContentTypes[ext]; ok {
		return ct
	}
	return ContentTypeDefault
}

// GetExtensionByContentType returns the file extension for the given content type.
func GetExtensionByContentType(contentType string) string {
	if ext, ok := ContentTypeToExtension[contentType]; ok {
		return ext
	}
	return ".bin"
}

// OPC namespaces
const (
	NamespaceOPCPackage      = "http://schemas.openxmlformats.org/package/2006/content-types"
	NamespaceRelationships   = "http://schemas.openxmlformats.org/package/2006/relationships"
	NamespaceRelationshipsNs = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
)

// OPC default paths
const (
	PathContentTypes = "[Content_Types].xml"
	PathRelsDir      = "_rels"
	PathRelsFile     = ".rels"
)

// XMLDeclaration is the standard XML declaration header used in all XML files within an OPC package.
const XMLDeclaration = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`

// IsImmutableContentType reports whether the content type represents an immutable resource.
// Immutable resources can be shared via zero-copy without deep-copying their data.
func IsImmutableContentType(contentType string) bool {
	switch contentType {
	// Image types — binary data, read-only
	case ContentTypePNG, ContentTypeJPEG, ContentTypeGIF,
		ContentTypeBMP, ContentTypeTIFF, ContentTypeWMF,
		ContentTypeEMF, ContentTypeSVG:
		return true

	// Audio and video types — binary data, read-only
	case ContentTypeWAV, ContentTypeMP3, ContentTypeMIDI,
		ContentTypeMP4, ContentTypeAVI, ContentTypeWMV:
		return true

	// Themes and masters — template files that are typically unchanged
	case ContentTypeTheme, ContentTypeThemeOverride,
		ContentTypeSlideMaster, ContentTypeSlideLayout:
		return true

	// Font files — read-only
	case ContentTypeFont:
		return true

	default:
		return false
	}
}

// IsLargeBinaryContentType reports whether the content type represents a large binary payload.
// Used to determine whether zero-copy optimization is worthwhile.
func IsLargeBinaryContentType(contentType string) bool {
	switch contentType {
	case ContentTypePNG, ContentTypeJPEG, ContentTypeGIF,
		ContentTypeBMP, ContentTypeTIFF, ContentTypeWMF,
		ContentTypeEMF, ContentTypeSVG,
		ContentTypeWAV, ContentTypeMP3, ContentTypeMIDI,
		ContentTypeMP4, ContentTypeAVI, ContentTypeWMV,
		ContentTypeFont:
		return true
	default:
		return false
	}
}

// IsImageContentType reports whether the content type represents an image.
func IsImageContentType(contentType string) bool {
	switch contentType {
	case ContentTypePNG, ContentTypeJPEG, ContentTypeGIF,
		ContentTypeBMP, ContentTypeTIFF, ContentTypeWMF,
		ContentTypeEMF, ContentTypeSVG:
		return true
	default:
		return false
	}
}

// IsMediaContentType reports whether the content type represents audio or video.
func IsMediaContentType(contentType string) bool {
	switch contentType {
	case ContentTypeWAV, ContentTypeMP3, ContentTypeMIDI,
		ContentTypeMP4, ContentTypeAVI, ContentTypeWMV:
		return true
	default:
		return false
	}
}
