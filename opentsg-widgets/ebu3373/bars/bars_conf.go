package bars

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type BarJSON struct {
	//	Type    string      `json:"type" yaml:"type"`
	ColourSpace       *colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	config.WidgetGrid `yaml:",inline"`
}

//go:embed jsonschema/barschema.json
var Schema []byte

/*
	func (b barJSON) Alias() string {
		return b.GridLoc.Alias
	}

	func (b barJSON) Location() string {
		return b.GridLoc.Location
	}
*/
func (b BarJSON) Wait() (bool, []string) {
	return false, []string{}
}
