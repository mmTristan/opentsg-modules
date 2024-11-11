package nearblack

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type Config struct {
	// Type    string      `json:"type" yaml:"type"`
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	config.WidgetGrid `yaml:",inline"`
}

//go:embed jsonschema/nbschema.json
var Schema []byte

/*
func (nb nearblackJSON) Alias() string {
	return nb.GridLoc.Alias
}

func (nb nearblackJSON) Location() string {
	return nb.GridLoc.Location
}
*/
