package luma

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type LumaJSON struct {
	// Type    string      `json:"type" yaml:"type"`
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	config.WidgetGrid `yaml:",inline"`
}

//go:embed jsonschema/lumaschema.json
var Schema []byte

/*
func (l lumaJSON) Alias() string {
	return l.GridLoc.Alias
}

func (l lumaJSON) Location() string {
	return l.GridLoc.Location
}*/
