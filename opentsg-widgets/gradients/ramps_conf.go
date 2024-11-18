package gradients

import (
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"

	_ "embed"
)

// Ramp is the gradient configuration object
type Ramp struct {
	Gradients         groupContents     `json:"groupsTemplates,omitempty" yaml:"groupsTemplates,omitempty"`
	Groups            []RampProperties  `json:"groups,omitempty" yaml:"groups,omitempty"`
	WidgetProperties  control           `json:"widgetProperties,omitempty" yaml:"widgetProperties,omitempty"`
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	config.WidgetGrid `yaml:",inline"`
}

type groupContents struct {
	GroupSeparator    groupSeparator    `json:"separator,omitempty" yaml:"separator,omitempty"`
	GradientSeparator gradientSeparator `json:"gradientSeparator,omitempty" yaml:"gradientSeparator,omitempty"`
	Gradients         []Gradient        `json:"gradients,omitempty" yaml:"gradients,omitempty"`
}

type textObjectJSON struct {
	TextYPosition string  `json:"textyPosition,omitempty" yaml:"textyPosition,omitempty"`
	TextXPosition string  `json:"textxPosition,omitempty" yaml:"textxPosition,omitempty"`
	TextHeight    float64 `json:"textHeight,omitempty" yaml:"textHeight,omitempty"`
	TextColour    string  `json:"textColor,omitempty" yaml:"textColor,omitempty"`
}

type RampProperties struct {
	Colour            string `json:"color,omitempty" yaml:"color,omitempty"`
	InitialPixelValue int    `json:"initialPixelValue,omitempty" yaml:"initialPixelValue,omitempty"`
	Reverse           bool   `json:"reverse,omitempty" yaml:"reverse,omitempty"`
}
type Gradient struct {
	Height   int    `json:"height,omitempty" yaml:"height,omitempty"`
	BitDepth int    `json:"bitDepth,omitempty" yaml:"bitDepth,omitempty"`
	Label    string `json:"label,omitempty" yaml:"label,omitempty"`

	// things that are added on run throughs
	startPoint int
	reverse    bool

	// Things we generate
	base   control
	colour string
}

type groupSeparator struct {
	Height int    `json:"height" yaml:"height"`
	Colour string `json:"color" yaml:"color"`
}

type gradientSeparator struct {
	Colours []string `json:"colors" yaml:"colors"`
	Height  int      `json:"height" yaml:"height"`
	// things the user does not assign
	base control
	step int
}

type control struct {
	MaxBitDepth int `json:"maxBitDepth" yaml:"maxBitDepth"`

	// CwRotation       anglegen.Angle `json:"cwRotation" yaml:"cwRotation"`
	ObjectFitFill    bool           `json:"objectFitFill" yaml:"objectFitFill"`
	PixelValueRepeat int            `json:"pixelValueRepeat" yaml:"pixelValueRepeat"`
	TextProperties   textObjectJSON `json:"textProperties" yaml:"textProperties"`
	// These are things the user does not set
	// embed the angle
	parameters.RotationAngle
	/*
		fill function - for rotation to automatically translate the fill location
		fill - get stepsize and end goal

		step size - fill or truncate. Add a multiplier


	*/

	angleType      string
	truePixelShift float64
}

//go:embed jsonschema/gradientSchema.json
var Schema []byte

/*
func (r Ramp) Alias() string {
	return r.GridLoc.Alias
}

func (r Ramp) Location() string {
	return r.GridLoc.Location
}*/
