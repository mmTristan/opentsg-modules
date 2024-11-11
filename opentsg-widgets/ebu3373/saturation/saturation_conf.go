package saturation

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type Config struct {
	// Type    string       `json:"type" yaml:"type"`
	Colours           []string          `json:"colors,omitempty" yaml:"colors,omitempty"`
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	config.WidgetGrid `yaml:",inline"`
}

//go:embed jsonschema/satschema.json
var Schema []byte

/*
func (s saturationJSON) Alias() string {
	return s.GridLoc.Alias
}

func (s saturationJSON) Location() string {
	return s.GridLoc.Location
}*/
