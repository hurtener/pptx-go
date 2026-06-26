package scene

import "github.com/hurtener/pptx-go/pptx"

// The shared rich-text model (RFC §9). A scene RichText is an ordered list of
// TextRuns, each plain text + an inline style + a color. The renderer maps a
// RichText onto a builder Paragraph with one Run per TextRun, so inline run
// colors honor theme swaps the same way page colors do.

// RichText is an ordered list of styled text runs.
type RichText []TextRun

// TextRun is one styled span of text within a RichText.
type TextRun struct {
	Text  string
	Style RunStyle
	Color TextColor
}

// RunStyle is the inline styling of a TextRun. TypeRole selects the typography
// scale; the booleans are inline toggles. Code is inline code (mono + tint,
// D-013); Link marks the run as a hyperlink with Href as its target.
type RunStyle struct {
	TypeRole  TypeRole
	Bold      bool
	Italic    bool
	Underline bool
	Strike    bool
	Code      bool
	Link      bool
	Href      string
	// Superscript raises the run above the baseline at a reduced size — a footnote
	// marker on a figure/stat (R14.12, D-126). Zero = on the baseline.
	Superscript bool
}

// TextColor is a run color: a TextColorRole token (theme-bound, the default
// path) or a literal RGB (the escape hatch). The zero value is the token
// TextPrimary.
type TextColor struct {
	role    TextColorRole
	literal pptx.RGB
	hasLit  bool
}

// TokenTextColor returns a token color bound to a semantic text role.
func TokenTextColor(role TextColorRole) TextColor { return TextColor{role: role} }

// LiteralColor returns an unbound literal color (a 6-hex string), bypassing the
// theme (RFC §9).
func LiteralColor(hex string) TextColor { return TextColor{literal: pptx.RGB(hex), hasLit: true} }

// IsLiteral reports whether the color is a literal (vs a token).
func (c TextColor) IsLiteral() bool { return c.hasLit }

// Role returns the bound text-color role (valid when IsLiteral is false).
func (c TextColor) Role() TextColorRole { return c.role }

// Literal returns the literal RGB (valid when IsLiteral is true).
func (c TextColor) Literal() pptx.RGB { return c.literal }
