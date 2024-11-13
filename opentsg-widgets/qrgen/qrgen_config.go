package qrgen

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-core/parameters"
)

type Config struct {
	// Type    string             `json:"type" yaml:"type"`
	Code              string `json:"code" yaml:"code"`
	parameters.Offset `yaml:",inline"`
	Size              *sizeJSON          `json:"size,omitempty" yaml:"size,omitempty"`
	Query             *[]objectQueryJSON `json:"objectQuery,omitempty" yaml:"objectQuery,omitempty"`
	config.WidgetGrid `yaml:",inline"`
	ColourSpace       colour.ColorSpace `json:"colorSpace" yaml:"colorSpace"`
}

type sizeJSON struct {
	Width  float64 `json:"width" yaml:"width"`
	Height float64 `json:"height" yaml:"height"`
}

type objectQueryJSON struct {
	Target string   `json:"targetAlias" yaml:"targetAlias"`
	Keys   []string `json:"keys" yaml:"keys"`
}

//go:embed jsonschema/qrgenschema.json
var Schema []byte

/*
func (q qrcodeJSON) Alias() string {
	return q.GridLoc.Alias
}

func (q qrcodeJSON) Location() string {
	return q.GridLoc.Location
}*/
