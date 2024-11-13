package noise

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type Config struct {
	//	Type      string       `json:"type" yaml:"type"`
	NoiseType         string            `json:"noiseType" yaml:"noiseType"`
	Minimum           int               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum           int               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	YOffsets          Guillotine        `json:"yOffset,omitempty" yaml:"yOffset,omitempty"`
	ColourSpace       colour.ColorSpace `json:"colorSpace" yaml:"colorSpace"`
	config.WidgetGrid `yaml:",inline"`
}

// go for top then bottom and work from there
type Guillotine struct {
	TopLeft     int `json:"topLeft,omitempty" yaml:"topLeft,omitempty"`
	TopRight    int `json:"topRight,omitempty" yaml:"topRight,omitempty"`
	BottomRight int `json:"bottomRight,omitempty" yaml:"bottomRight,omitempty"`
	BottomLeft  int `json:"bottomLeft,omitempty" yaml:"bottomLeft,omitempty"`
}

//go:embed jsonschema/noiseschema.json
var Schema []byte

/*
func (n noiseJSON) Alias() string {
	return n.GridLoc.Alias
}

func (n noiseJSON) Location() string {
	return n.GridLoc.Location
}*/
