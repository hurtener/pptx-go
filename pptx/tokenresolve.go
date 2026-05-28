package pptx

// Token resolution (RFC §7.4). Each Resolve* method maps a semantic token to
// the theme's concrete OOXML value. Resolution is deterministic: the same
// token against the same theme always yields the same value (the property the
// theme-swap guarantee rests on). A token absent from the theme falls back to
// a safe neutral value rather than panicking (CLAUDE.md §5 — never panic
// across a public API boundary).

// Fallback values for tokens a theme leaves unset.
const (
	fallbackSurface = RGB("FFFFFF")
	fallbackText    = RGB("000000")
)

var fallbackFont = FontSpec{Family: "Calibri", Size: 14, Weight: 400}

// ResolveColor returns the 6-hex RGB for a surface color role.
func (t *Theme) ResolveColor(role ColorRole) RGB {
	if v, ok := t.Colors.Surfaces[role]; ok {
		return v
	}
	return fallbackSurface
}

// ResolveTextColor returns the 6-hex RGB for a text color role.
func (t *Theme) ResolveTextColor(role TextColorRole) RGB {
	if v, ok := t.Colors.Text[role]; ok {
		return v
	}
	return fallbackText
}

// ResolveType returns the FontSpec for a typography role.
func (t *Theme) ResolveType(role TypeRole) FontSpec {
	if v, ok := t.Typography[role]; ok {
		return v
	}
	return fallbackFont
}

// ResolveSpace returns the EMU for a spacing role.
func (t *Theme) ResolveSpace(role SpaceRole) EMU {
	if v, ok := t.Spacing[role]; ok {
		return v
	}
	return 0
}

// ResolveRadius returns the EMU for a radius role.
func (t *Theme) ResolveRadius(role RadiusRole) EMU {
	if v, ok := t.Radii[role]; ok {
		return v
	}
	return 0
}

// ResolveElevation returns the Elevation for an elevation role.
func (t *Theme) ResolveElevation(role ElevationRole) Elevation {
	if v, ok := t.Elevations[role]; ok {
		return v
	}
	return Elevation{}
}
