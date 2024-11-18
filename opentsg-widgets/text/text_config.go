package text

import (
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"
)

// TextboxProperties contains all the properties for generating a textbox
type TextboxProperties struct {
	// Type       string       `json:"type" yaml:"type"`
	font             string            // `json:"font" yaml:"font"`
	colourSpace      colour.ColorSpace // `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	backgroundColour *colour.CNRGBA64  // `json:"backgroundcolor" yaml:"backgroundcolor"`
	textColour       *colour.CNRGBA64  // `json:"textcolor" yaml:"textcolor"`
	fillType         string
	xAlignment       string
	yAlignment       string
	verticalText     bool
}

// NewTextBoxer generates a new TextBoxProperties object.
// Tailored to the options provided and the
// order in which the options are specified are the order in which the executed
func NewTextboxer(colourSpace colour.ColorSpace, options ...func(*TextboxProperties)) *TextboxProperties {
	txt := &TextboxProperties{colourSpace: colourSpace}
	for _, opt := range options {
		opt(txt)
	}
	return txt
}

// WithFill sets the fill type
// of full or relaxed.
func WithFill(fill string) func(t *TextboxProperties) {

	return func(t *TextboxProperties) {
		t.fillType = fill
	}
}

// WithFont sets the font, this can be a web font, a locally
// stored font or one of the textbox defaults.
func WithFont(font string) func(t *TextboxProperties) {

	return func(t *TextboxProperties) {
		t.font = font
	}
}

// WithTextColourString sets the text colour as one of the openTSG string colours
func WithTextColourString(colour string) func(t *TextboxProperties) {

	return func(t *TextboxProperties) {
		c := parameters.HexToColour(colour, t.colourSpace)
		t.textColour = c
	}
}

// WithTextColour sets the color as a *colour.CNRGBA64
func WithTextColour(colour *colour.CNRGBA64) func(t *TextboxProperties) {

	return func(t *TextboxProperties) {
		t.textColour = colour
	}
}

// WithBackgroundColourString sets the text colour as one of the openTSG string colours
func WithBackgroundColourString(colour string) func(t *TextboxProperties) {

	return func(t *TextboxProperties) {
		c := parameters.HexToColour(colour, t.colourSpace)
		t.backgroundColour = c
	}
}

// WithBackGroundColour sets the color as a *colour.CNRGBA64
func WithBackgroundColour(colour *colour.CNRGBA64) func(t *TextboxProperties) {

	return func(t *TextboxProperties) {
		t.backgroundColour = colour
	}
}

// WithXAlignment sets the x alignment of the textbox
func WithXAlignment(x string) func(t *TextboxProperties) {
	return func(t *TextboxProperties) {
		t.xAlignment = x
	}
}

// WithYAlignment sets the x alignment of the textbox
func WithYAlignment(y string) func(t *TextboxProperties) {
	return func(t *TextboxProperties) {
		t.yAlignment = y
	}
}

// WithVerticalText sets the vertical text with true,
// if false or not declared the text is horizontal
func WithVerticalText(vertical bool) func(t *TextboxProperties) {
	return func(t *TextboxProperties) {
		t.verticalText = true
	}
}
