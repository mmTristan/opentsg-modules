package textbox

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"
)

type TextboxJSON struct {
	// Type       string       `json:"type" yaml:"type"`
	Text       []string             `json:"text,omitempty" yaml:"text,omitempty"`
	Border     parameters.HexString `json:"borderColor,omitempty" yaml:"borderColor,omitempty"`
	BorderSize float64              `json:"borderSize,omitempty" yaml:"borderSize,omitempty"`
	Font       string               `json:"font,omitempty" yaml:"font,omitempty"`

	Back       string `json:"backgroundColor,omitempty" yaml:"backgroundColor,omitempty"`
	Textc      string `json:"textColor,omitempty" yaml:"textColor,omitempty"`
	FillType   string `json:"fillType,omitempty" yaml:"fillType,omitempty"`
	XAlignment string `json:"xAlignment,omitempty" yaml:"xAlignment,omitempty"`
	YAlignment string `json:"yAlignment,omitempty" yaml:"yAlignment,omitempty"`
}

//go:embed jsonschema/textBoxSchema.json
var Schema []byte

/*
func (tb TextboxJSON) Alias() string {
	return tb.GridLoc.Alias
}

func (tb TextboxJSON) Location() string {
	return tb.GridLoc.Location
}
*/
