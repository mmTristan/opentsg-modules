package bowtie

import (
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-core/parameters"

	_ "embed"
)

type Config struct {
	SegementCount  int      `json:"segmentCount,omitempty" yaml:"segmentCount,omitempty"`
	SegmentColours []string `json:"segmentColors,omitempty" yaml:"segmentColors,omitempty"`
	Blend          string   `json:"blend,omitempty" yaml:"blend,omitempty"`

	//////// defaults
	ColourSpace *colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`

	parameters.RotationAngle `yaml:",inline"`
	parameters.StartAngle    `yaml:",inline"`
	parameters.Offset        `yaml:",inline"`
	config.WidgetGrid        `yaml:",inline"`
}

//go:embed jsonschema/jsonschema.json
var Schema []byte
