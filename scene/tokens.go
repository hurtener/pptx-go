package scene

import "github.com/hurtener/pptx-go/pptx"

// Token enums are aliases of pptx's, with const re-exports, so a scene caller
// uses the same vocabulary as the builder (RFC §9; P2). These are the design
// tokens the IR references; they resolve against the active theme at render
// time.

// Surface color roles.
type ColorRole = pptx.ColorRole

const (
	ColorCanvas     = pptx.ColorCanvas
	ColorSurface    = pptx.ColorSurface
	ColorSurfaceAlt = pptx.ColorSurfaceAlt
	ColorAccent     = pptx.ColorAccent
	ColorAccentAlt  = pptx.ColorAccentAlt
	ColorAccentWarm = pptx.ColorAccentWarm
	ColorSuccess    = pptx.ColorSuccess
	ColorWarning    = pptx.ColorWarning
	ColorError      = pptx.ColorError
	ColorInfo       = pptx.ColorInfo
)

// Text color roles (inline runs).
type TextColorRole = pptx.TextColorRole

const (
	TextPrimary   = pptx.TextPrimary
	TextSecondary = pptx.TextSecondary
	TextTertiary  = pptx.TextTertiary
	TextInverse   = pptx.TextInverse
	TextMuted     = pptx.TextMuted
	TextAccent    = pptx.TextAccent
	TextAccentAlt = pptx.TextAccentAlt
	TextSuccess   = pptx.TextSuccess
	TextWarning   = pptx.TextWarning
	TextError     = pptx.TextError
)

// Typography roles.
type TypeRole = pptx.TypeRole

const (
	TypeDisplay   = pptx.TypeDisplay
	TypeH1        = pptx.TypeH1
	TypeH2        = pptx.TypeH2
	TypeH3        = pptx.TypeH3
	TypeH4        = pptx.TypeH4
	TypeH5        = pptx.TypeH5
	TypeBody      = pptx.TypeBody
	TypeBodySmall = pptx.TypeBodySmall
	TypeCaption   = pptx.TypeCaption
	TypeMono      = pptx.TypeMono
	TypeCode      = pptx.TypeCode
)

// Spacing roles.
type SpaceRole = pptx.SpaceRole

const (
	SpaceXS  = pptx.SpaceXS
	SpaceSM  = pptx.SpaceSM
	SpaceMD  = pptx.SpaceMD
	SpaceLG  = pptx.SpaceLG
	SpaceXL  = pptx.SpaceXL
	Space2XL = pptx.Space2XL
)

// Corner-radius roles.
type RadiusRole = pptx.RadiusRole

const (
	RadiusNone = pptx.RadiusNone
	RadiusSM   = pptx.RadiusSM
	RadiusMD   = pptx.RadiusMD
	RadiusLG   = pptx.RadiusLG
	RadiusFull = pptx.RadiusFull
)

// Elevation roles.
type ElevationRole = pptx.ElevationRole

const (
	ElevationFlat     = pptx.ElevationFlat
	ElevationRaised   = pptx.ElevationRaised
	ElevationElevated = pptx.ElevationElevated
)

// Anchor is a reference point reused from the builder for decoration/connector
// placement.
type Anchor = pptx.Anchor

const (
	AnchorTopLeft      = pptx.AnchorTopLeft
	AnchorTopCenter    = pptx.AnchorTopCenter
	AnchorTopRight     = pptx.AnchorTopRight
	AnchorCenterLeft   = pptx.AnchorCenterLeft
	AnchorCenter       = pptx.AnchorCenter
	AnchorCenterRight  = pptx.AnchorCenterRight
	AnchorBottomLeft   = pptx.AnchorBottomLeft
	AnchorBottomCenter = pptx.AnchorBottomCenter
	AnchorBottomRight  = pptx.AnchorBottomRight
)

// Position and Size are EMU geometry types reused from the builder, carried on
// the Decoration node (offset + size).
type Position = pptx.Position
type Size = pptx.Size
