package pptx

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hurtener/pptx-go/internal/ooxml/theme"
	"github.com/hurtener/pptx-go/internal/opc"
)

// ErrThemeNotFound is returned by LoadTheme when the package has no theme part.
var ErrThemeNotFound = errors.New("pptx: no theme part in package")

// Theme ↔ theme1.xml mapping (RFC §7.3). PowerPoint's theme is a positional
// 12-color scheme (dk1/lt1/dk2/lt2/accent1..6/hlink/folHlink) plus a
// major/minor font pair. pptx's palette is semantic, so the mapping is by
// convention (documented in docs/design/THEME.md): each OOXML slot has one
// canonical semantic owner for writing, and each semantic role reads back
// from its slot. Roles without a slot (e.g. TextMuted) keep their default.

// writeSlots is the canonical slot → value source used when emitting OOXML.
func (t *Theme) writeSlots() map[theme.ColorRole]RGB {
	return map[theme.ColorRole]RGB{
		theme.ColorRoleLight1:            t.ResolveColor(ColorSurface),
		theme.ColorRoleLight2:            t.ResolveColor(ColorSurfaceAlt),
		theme.ColorRoleDark1:             RGB(t.ResolveTextColor(TextPrimary)),
		theme.ColorRoleDark2:             RGB(t.ResolveTextColor(TextSecondary)),
		theme.ColorRoleAccent1:           t.ResolveColor(ColorAccent),
		theme.ColorRoleAccent2:           t.ResolveColor(ColorAccentAlt),
		theme.ColorRoleAccent3:           t.ResolveColor(ColorAccentWarm),
		theme.ColorRoleAccent4:           t.ResolveColor(ColorSuccess),
		theme.ColorRoleAccent5:           t.ResolveColor(ColorWarning),
		theme.ColorRoleAccent6:           t.ResolveColor(ColorError),
		theme.ColorRoleHyperlink:         t.ResolveColor(ColorInfo),
		theme.ColorRoleFollowedHyperlink: RGB(t.ResolveTextColor(TextAccentAlt)),
	}
}

// toThemePart builds an OOXML ThemePart from the semantic theme.
func (t *Theme) toThemePart() *theme.ThemePart {
	tp := theme.NewThemePart(1)
	for slot, rgb := range t.writeSlots() {
		tp.SetThemeColorRGB(slot, string(rgb))
	}
	tp.SetThemeMajorFont(t.HeadingFont, "", "")
	tp.SetThemeMinorFont(t.BodyFont, "", "")
	return tp
}

// ThemeXML serializes the theme to a theme1.xml byte slice.
func (t *Theme) ThemeXML() ([]byte, error) {
	b, err := t.toThemePart().ToXML()
	if err != nil {
		return nil, fmt.Errorf("encode theme: %w", err)
	}
	return b, nil
}

// themeFromPart maps a parsed OOXML ThemePart onto the semantic taxonomy,
// starting from the default theme so roles without an OOXML slot keep sane
// values.
func themeFromPart(tp *theme.ThemePart) *Theme {
	t := DefaultTheme()
	set := func(role ColorRole, slot theme.ColorRole) {
		if v := tp.GetThemeColorRGB(slot); v != "" {
			t.Colors.Surfaces[role] = RGB(v)
		}
	}
	setText := func(role TextColorRole, slot theme.ColorRole) {
		if v := tp.GetThemeColorRGB(slot); v != "" {
			t.Colors.Text[role] = RGB(v)
		}
	}
	set(ColorCanvas, theme.ColorRoleLight1)
	set(ColorSurface, theme.ColorRoleLight1)
	set(ColorSurfaceAlt, theme.ColorRoleLight2)
	set(ColorAccent, theme.ColorRoleAccent1)
	set(ColorAccentAlt, theme.ColorRoleAccent2)
	set(ColorAccentWarm, theme.ColorRoleAccent3)
	set(ColorSuccess, theme.ColorRoleAccent4)
	set(ColorWarning, theme.ColorRoleAccent5)
	set(ColorError, theme.ColorRoleAccent6)
	set(ColorInfo, theme.ColorRoleHyperlink)
	setText(TextPrimary, theme.ColorRoleDark1)
	setText(TextSecondary, theme.ColorRoleDark2)
	setText(TextInverse, theme.ColorRoleLight1)
	setText(TextAccent, theme.ColorRoleAccent1)
	setText(TextAccentAlt, theme.ColorRoleFollowedHyperlink)

	if maj := tp.MajorLatinFont(); maj != "" {
		t.HeadingFont = maj
		for role, spec := range t.Typography {
			if role <= TypeH5 {
				spec.Family = maj
				t.Typography[role] = spec
			}
		}
	}
	if minor := tp.MinorLatinFont(); minor != "" {
		t.BodyFont = minor
		for role := TypeBody; role <= TypeCaption; role++ {
			spec := t.Typography[role]
			spec.Family = minor
			t.Typography[role] = spec
		}
	}
	if name := tp.Theme().Name; name != "" {
		t.Name = name
	}
	return t
}

// LoadThemeFromBytes extracts the Theme from a .pptx file's theme1.xml.
func LoadThemeFromBytes(data []byte) (*Theme, error) {
	pkg, err := opc.Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open package: %w", err)
	}
	defer func() { _ = pkg.Close() }()
	return loadThemeFromPackage(pkg)
}

// LoadTheme extracts the Theme from a .pptx template file's theme1.xml.
func LoadTheme(path string) (*Theme, error) {
	pkg, err := opc.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = pkg.Close() }()
	return loadThemeFromPackage(pkg)
}

func loadThemeFromPackage(pkg *opc.Package) (*Theme, error) {
	parts := pkg.GetPartsByType(opc.ContentTypeTheme)
	if len(parts) == 0 {
		return nil, fmt.Errorf("no theme part in package: %w", ErrThemeNotFound)
	}
	tp := theme.NewThemePart(1)
	if err := tp.FromXML(parts[0].Blob()); err != nil {
		return nil, fmt.Errorf("parse theme1.xml: %w", err)
	}
	return themeFromPart(tp), nil
}
