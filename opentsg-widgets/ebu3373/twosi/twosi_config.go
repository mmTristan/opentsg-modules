package twosi

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type Config struct {
	//	Type    string      `json:"type" yaml:"type"`
	config.WidgetGrid `yaml:",inline"`
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
}

//go:embed jsonschema/twoschema.json
var Schema []byte

/*
func (t twosiJSON) Alias() string {
	return t.GridLoc.Alias
}

func (t twosiJSON) Location() string {
	return t.GridLoc.Location
}
*/
