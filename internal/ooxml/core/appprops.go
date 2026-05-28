package core

// ============================================================================
// AppPropsPart - application properties part
// ============================================================================
//
// Corresponds to /docProps/app.xml
// Carries application-related metadata (company, manager, slide count, etc.)
//
// ============================================================================

import (
	"encoding/xml"
	"fmt"
	"sync"

	"github.com/hurtener/pptx-go/internal/ooxml"
	"github.com/hurtener/pptx-go/internal/opc"
)

// AppPropsPart holds the application properties part (/docProps/app.xml).
type AppPropsPart struct {
	uri      *opc.PackURI
	appProps *XMLAppProps
	mu       sync.RWMutex
}

// NewAppPropsPart creates a new application properties part with defaults.
func NewAppPropsPart() *AppPropsPart {
	return &AppPropsPart{
		uri: opc.NewPackURI("/docProps/app.xml"),
		appProps: &XMLAppProps{
			Application: "Microsoft Office PowerPoint",
			AppVersion:  "15.0000",
		},
	}
}

// PartURI returns the part URI.
func (a *AppPropsPart) PartURI() *opc.PackURI {
	return a.uri
}

// AppProps returns the application properties data.
func (a *AppPropsPart) AppProps() *XMLAppProps {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.appProps
}

// ============================================================================
// Property getter methods
// ============================================================================

// GetAppApplication returns the application name.
func (a *AppPropsPart) GetAppApplication() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.appProps == nil {
		return ""
	}
	return a.appProps.Application
}

// GetAppVersion returns the application version.
func (a *AppPropsPart) GetAppVersion() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.appProps == nil {
		return ""
	}
	return a.appProps.AppVersion
}

// GetAppCompany returns the company name.
func (a *AppPropsPart) GetAppCompany() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.appProps == nil {
		return ""
	}
	return a.appProps.Company
}

// GetAppManager returns the manager name.
func (a *AppPropsPart) GetAppManager() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.appProps == nil {
		return ""
	}
	return a.appProps.Manager
}

// GetAppSlideCount returns the slide count.
func (a *AppPropsPart) GetAppSlideCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.appProps == nil || a.appProps.Slides == nil {
		return 0
	}
	return *a.appProps.Slides
}

// GetAppWordCount returns the word count.
func (a *AppPropsPart) GetAppWordCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.appProps == nil || a.appProps.Words == nil {
		return 0
	}
	return *a.appProps.Words
}

// GetAppTotalTime returns the total editing time in minutes.
func (a *AppPropsPart) GetAppTotalTime() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.appProps == nil || a.appProps.TotalTime == nil {
		return 0
	}
	return *a.appProps.TotalTime
}

// ============================================================================
// Property setter methods
// ============================================================================

// SetAppCompany sets the company name.
func (a *AppPropsPart) SetAppCompany(company string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Company = company
}

// SetAppManager sets the manager name.
func (a *AppPropsPart) SetAppManager(manager string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Manager = manager
}

// SetAppApplication sets the application name.
func (a *AppPropsPart) SetAppApplication(app string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Application = app
}

// SetAppVersion sets the application version.
func (a *AppPropsPart) SetAppVersion(version string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.AppVersion = version
}

// SetAppSlideCount sets the slide count.
func (a *AppPropsPart) SetAppSlideCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Slides = &count
}

// SetAppWordCount sets the word count.
func (a *AppPropsPart) SetAppWordCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Words = &count
}

// SetAppTotalTime sets the total editing time in minutes.
func (a *AppPropsPart) SetAppTotalTime(minutes int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.TotalTime = &minutes
}

// SetAppNoteCount sets the notes count.
func (a *AppPropsPart) SetAppNoteCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Notes = &count
}

// SetAppHiddenSlideCount sets the hidden slide count.
func (a *AppPropsPart) SetAppHiddenSlideCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.HiddenSlides = &count
}

// SetAppMMClipCount sets the multimedia clip count.
func (a *AppPropsPart) SetAppMMClipCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.MMClips = &count
}

// SetAppParagraphCount sets the paragraph count.
func (a *AppPropsPart) SetAppParagraphCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Paragraphs = &count
}

// SetAppCharacterCount sets the character count.
func (a *AppPropsPart) SetAppCharacterCount(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.Characters = &count
}

// SetAppHyperlinkBase sets the hyperlink base URL.
func (a *AppPropsPart) SetAppHyperlinkBase(base string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.HyperlinkBase = base
}

// SetAppLinksUpToDate sets whether links are up to date.
func (a *AppPropsPart) SetAppLinksUpToDate(upToDate bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.LinksUpToDate = &upToDate
}

// SetAppSharedDoc sets whether the document is shared.
func (a *AppPropsPart) SetAppSharedDoc(shared bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.SharedDoc = &shared
}

// SetAppHeadingPairs sets the heading pairs.
func (a *AppPropsPart) SetAppHeadingPairs(headingPairs *XMLHeadingPairs) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.HeadingPairs = headingPairs
}

// SetAppTitlesOfParts sets the titles of parts.
func (a *AppPropsPart) SetAppTitlesOfParts(titlesOfParts *XMLTitlesOfParts) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ensureAppProps()
	a.appProps.TitlesOfParts = titlesOfParts
}

// ensureAppProps ensures the app properties struct is initialized (must be called with the lock held).
func (a *AppPropsPart) ensureAppProps() {
	if a.appProps == nil {
		a.appProps = &XMLAppProps{
			Application: "Microsoft Office PowerPoint",
			AppVersion:  "15.0000",
		}
	}
}

// ============================================================================
// XML serialization / deserialization
// ============================================================================

// ToXML serializes the application properties to XML.
func (a *AppPropsPart) ToXML() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.appProps == nil {
		return nil, fmt.Errorf("app data is nil")
	}

	// ensure namespace declarations are present
	a.appProps.XmlnsProp = NamespaceExtendedProperties
	a.appProps.XmlnsVt = NamespaceDocPropsVTypes

	output, err := xml.MarshalIndent(a.appProps, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(ooxml.XMLDeclaration), output...), nil
}

// FromXML deserializes application properties from XML.
func (a *AppPropsPart) FromXML(data []byte) error {
	var app XMLAppProps
	if err := xml.Unmarshal(data, &app); err != nil {
		return fmt.Errorf("failed to unmarshal app XML: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.appProps = &app

	return nil
}

// ParseAppProps parses application properties from XML bytes.
func ParseAppProps(data []byte) (*XMLAppProps, error) {
	var app XMLAppProps
	if err := xml.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("failed to parse app: %w", err)
	}
	return &app, nil
}
