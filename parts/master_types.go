package parts

// ============================================================================
// Core data structures for PPT slide masters and layouts
// ============================================================================
//
// Design principles:
// 1. All struct fields are read-only (lowercase fields initialized via constructors,
//    uppercase fields are immutable values).
// 2. Optimized for high-concurrency reads; safe to read without locking.
// 3. Data is built once during parsing and never modified afterward.
// ============================================================================

// PlaceholderType enumerates placeholder types.
// Corresponds to XML: <p:ph type="...">
type PlaceholderType int8

const (
	PlaceholderTypeNone        PlaceholderType = iota // unspecified
	PlaceholderTypeTitle                              // title
	PlaceholderTypeBody                               // body/content
	PlaceholderTypeCenterTitle                        // centered title
	PlaceholderTypeSubTitle                           // subtitle
	PlaceholderTypeDateTime                           // date and time
	PlaceholderTypeSlideNumber                        // slide number
	PlaceholderTypeFooter                             // footer
	PlaceholderTypeHeader                             // header
	PlaceholderTypeObject                             // object
	PlaceholderTypeChart                              // chart
	PlaceholderTypeTable                              // table
	PlaceholderTypeClipArt                            // clip art
	PlaceholderTypeOrgChart                           // org chart
	PlaceholderTypeMedia                              // media
	PlaceholderTypeSlideImage                         // slide image
	PlaceholderTypePicture                            // picture
)

// String returns the string representation of the placeholder type.
func (t PlaceholderType) String() string {
	switch t {
	case PlaceholderTypeTitle:
		return "title"
	case PlaceholderTypeBody:
		return "body"
	case PlaceholderTypeCenterTitle:
		return "ctrTitle"
	case PlaceholderTypeSubTitle:
		return "subTitle"
	case PlaceholderTypeDateTime:
		return "dt"
	case PlaceholderTypeSlideNumber:
		return "sldNum"
	case PlaceholderTypeFooter:
		return "ftr"
	case PlaceholderTypeHeader:
		return "hdr"
	case PlaceholderTypeObject:
		return "obj"
	case PlaceholderTypeChart:
		return "chart"
	case PlaceholderTypeTable:
		return "tbl"
	case PlaceholderTypeClipArt:
		return "clipArt"
	case PlaceholderTypeOrgChart:
		return "dgm"
	case PlaceholderTypeMedia:
		return "media"
	case PlaceholderTypeSlideImage:
		return "sldImg"
	case PlaceholderTypePicture:
		return "pic"
	default:
		return ""
	}
}

// TextStyle holds the default text style for a placeholder (read-only).
// Defines default font, size, color, etc. for text inside a placeholder.
type TextStyle struct {
	fontName  string // font name
	fontSize  int32  // font size (hundredths of a point; 100 = 1pt)
	bold      bool   // bold
	italic    bool   // italic
	underline bool   // underline
	colorRGB  string // text color as RGB hex (e.g. "FF0000")
}

// Placeholder is a fillable region defined in a master or layout (read-only).
// Corresponds to XML: <p:sp> with <p:nvSpPr><p:nvPr><p:ph ...>
type Placeholder struct {
	id              string          // unique placeholder identifier (idx from XML or internally generated)
	placeholderType PlaceholderType // placeholder type
	x               int64           // X coordinate (EMU)
	y               int64           // Y coordinate (EMU)
	cx              int64           // width (EMU)
	cy              int64           // height (EMU)
	rotation        int32           // rotation angle (1/60000 of a degree)
	defaultStyle    *TextStyle      // default text style (may be nil)
}

// ============================================================================
// Background-related structures
// ============================================================================

// BackgroundType enumerates background types.
// Corresponds to the different child elements under XML: <p:bg>
type BackgroundType int8

const (
	BackgroundTypeNone       BackgroundType = iota // no background
	BackgroundTypeSolidColor                       // solid color
	BackgroundTypeGradient                         // gradient
	BackgroundTypePattern                          // pattern fill
	BackgroundTypePicture                          // picture background
	BackgroundTypeThemeColor                       // theme color (e.g. bg1, tx1)
)

// Background holds background definition data (read-only).
// Corresponds to XML: <p:bg> or <p:cSld><p:bg>
// Design note: separate fields for each background variant avoid interface allocations.
type Background struct {
	backgroundType BackgroundType // background type

	// Solid color background (valid when backgroundType == BackgroundTypeSolidColor)
	solidColorRGB string // RGB hex color value, e.g. "FFFFFF"

	// Gradient background (valid when backgroundType == BackgroundTypeGradient)
	gradientAngle  int32          // gradient angle (degrees)
	gradientColors []GradientStop // gradient stop list

	// Picture background (valid when backgroundType == BackgroundTypePicture)
	pictureRId string // picture relationship ID (pointing to a media resource)
	pictureURI string // picture internal URI path

	// Common
	opacity float32 // opacity (0.0–1.0); default 1.0
}

// GradientStop holds a gradient color stop (read-only).
// Corresponds to XML: <a:gs>
type GradientStop struct {
	position float32 // position (0.0–1.0) within the gradient
	colorRGB string  // RGB hex color value
}

// ============================================================================
// SlideLayout-related structures (read-only data)
// ============================================================================
//
// Note: SlideLayoutType is defined in slide_types.go and used here directly.
// ============================================================================

// SlideLayoutData holds read-only layout data (used by the template system).
// Corresponds to XML: /ppt/slideLayouts/slideLayoutN.xml
// Unlike SlideLayoutPart, this is a pure data structure with no XML read/write capability.
type SlideLayoutData struct {
	id           string                  // unique layout identifier (internally generated)
	name         string                  // layout name (shown in the PowerPoint layout picker)
	layoutType   SlideLayoutType         // layout type (reuses the definition from slide_types.go)
	background   *Background             // background (nil means inherit from master)
	masterId     string                  // ID of the owning master
	placeholders map[string]*Placeholder // placeholder set, keyed by placeholder ID
}

// ============================================================================
// SlideMaster-related structures
// ============================================================================

// SlideMasterData holds read-only master data (used by the template system).
// Corresponds to XML: /ppt/slideMasters/slideMasterN.xml
// The master is the top-level container for slide templates and owns one or more layouts.
type SlideMasterData struct {
	id           string                  // unique master identifier (internally generated)
	name         string                  // master name
	background   *Background             // master-level background (may be nil)
	placeholders map[string]*Placeholder // master-level placeholders (may be nil); defines global placeholder styles
	layouts      []*SlideLayoutData      // contained layout list
}

// ============================================================================
// Placeholder accessor methods
// ============================================================================

// ID returns the unique placeholder identifier.
func (p *Placeholder) ID() string { return p.id }

// Type returns the placeholder type.
func (p *Placeholder) Type() PlaceholderType { return p.placeholderType }

// X returns the X coordinate (EMU).
func (p *Placeholder) X() int64 { return p.x }

// Y returns the Y coordinate (EMU).
func (p *Placeholder) Y() int64 { return p.y }

// Cx returns the width (EMU).
func (p *Placeholder) Cx() int64 { return p.cx }

// Cy returns the height (EMU).
func (p *Placeholder) Cy() int64 { return p.cy }

// Rotation returns the rotation angle (1/60000 of a degree).
func (p *Placeholder) Rotation() int32 { return p.rotation }

// DefaultStyle returns the default text style (may be nil).
func (p *Placeholder) DefaultStyle() *TextStyle { return p.defaultStyle }

// Bounds returns the bounding rectangle (x, y, cx, cy).
func (p *Placeholder) Bounds() (x, y, cx, cy int64) {
	return p.x, p.y, p.cx, p.cy
}

// ============================================================================
// TextStyle accessor methods
// ============================================================================

// FontName returns the font name.
func (s *TextStyle) FontName() string { return s.fontName }

// FontSize returns the font size (hundredths of a point; 100 = 1pt).
func (s *TextStyle) FontSize() int32 { return s.fontSize }

// Bold reports whether the text is bold.
func (s *TextStyle) Bold() bool { return s.bold }

// Italic reports whether the text is italic.
func (s *TextStyle) Italic() bool { return s.italic }

// Underline reports whether the text is underlined.
func (s *TextStyle) Underline() bool { return s.underline }

// ColorRGB returns the text color as an RGB hex string.
func (s *TextStyle) ColorRGB() string { return s.colorRGB }

// ============================================================================
// Background accessor methods
// ============================================================================

// Type returns the background type.
func (b *Background) Type() BackgroundType { return b.backgroundType }

// SolidColorRGB returns the RGB value for a solid-color background
// (only valid when Type == BackgroundTypeSolidColor).
func (b *Background) SolidColorRGB() string { return b.solidColorRGB }

// GradientAngle returns the gradient angle
// (only valid when Type == BackgroundTypeGradient).
func (b *Background) GradientAngle() int32 { return b.gradientAngle }

// GradientColors returns the gradient stop list
// (only valid when Type == BackgroundTypeGradient).
func (b *Background) GradientColors() []GradientStop { return b.gradientColors }

// PictureRId returns the picture relationship ID
// (only valid when Type == BackgroundTypePicture).
func (b *Background) PictureRId() string { return b.pictureRId }

// PictureURI returns the picture's internal URI
// (only valid when Type == BackgroundTypePicture).
func (b *Background) PictureURI() string { return b.pictureURI }

// Opacity returns the opacity (0.0–1.0).
func (b *Background) Opacity() float32 { return b.opacity }

// ============================================================================
// GradientStop accessor methods
// ============================================================================

// Position returns the stop position (0.0–1.0).
func (g *GradientStop) Position() float32 { return g.position }

// ColorRGB returns the stop color as an RGB hex string.
func (g *GradientStop) ColorRGB() string { return g.colorRGB }

// ============================================================================
// SlideLayoutData accessor methods
// ============================================================================

// ID returns the unique layout identifier.
func (l *SlideLayoutData) ID() string { return l.id }

// Name returns the layout name.
func (l *SlideLayoutData) Name() string { return l.name }

// LayoutType returns the layout type.
func (l *SlideLayoutData) LayoutType() SlideLayoutType { return l.layoutType }

// Background returns the background (may be nil).
func (l *SlideLayoutData) Background() *Background { return l.background }

// MasterID returns the ID of the owning master.
func (l *SlideLayoutData) MasterID() string { return l.masterId }

// Placeholders returns the placeholder set.
func (l *SlideLayoutData) Placeholders() map[string]*Placeholder { return l.placeholders }

// PlaceholderByID returns the placeholder with the given ID (may be nil).
func (l *SlideLayoutData) PlaceholderByID(id string) *Placeholder {
	return l.placeholders[id]
}

// PlaceholderCount returns the number of placeholders.
func (l *SlideLayoutData) PlaceholderCount() int { return len(l.placeholders) }

// PlaceholderByType returns the first placeholder matching the given type.
func (l *SlideLayoutData) PlaceholderByType(phType PlaceholderType) *Placeholder {
	for _, ph := range l.placeholders {
		if ph.placeholderType == phType {
			return ph
		}
	}
	return nil
}

// TitlePlaceholder returns the title placeholder (convenience method).
func (l *SlideLayoutData) TitlePlaceholder() *Placeholder {
	return l.PlaceholderByType(PlaceholderTypeTitle)
}

// BodyPlaceholder returns the body placeholder (convenience method).
func (l *SlideLayoutData) BodyPlaceholder() *Placeholder {
	return l.PlaceholderByType(PlaceholderTypeBody)
}

// ============================================================================
// SlideMasterData accessor methods
// ============================================================================

// ID returns the unique master identifier.
func (m *SlideMasterData) ID() string { return m.id }

// Name returns the master name.
func (m *SlideMasterData) Name() string { return m.name }

// Background returns the background (may be nil).
func (m *SlideMasterData) Background() *Background { return m.background }

// Placeholders returns the master-level placeholder set.
func (m *SlideMasterData) Placeholders() map[string]*Placeholder { return m.placeholders }

// PlaceholderByID returns the placeholder with the given ID (may be nil).
func (m *SlideMasterData) PlaceholderByID(id string) *Placeholder {
	return m.placeholders[id]
}

// PlaceholderCount returns the number of placeholders.
func (m *SlideMasterData) PlaceholderCount() int { return len(m.placeholders) }

// Layouts returns the layout list.
func (m *SlideMasterData) Layouts() []*SlideLayoutData { return m.layouts }

// LayoutCount returns the number of layouts.
func (m *SlideMasterData) LayoutCount() int { return len(m.layouts) }

// LayoutByID returns the layout with the given ID (may be nil).
func (m *SlideMasterData) LayoutByID(id string) *SlideLayoutData {
	for _, layout := range m.layouts {
		if layout.id == id {
			return layout
		}
	}
	return nil
}
