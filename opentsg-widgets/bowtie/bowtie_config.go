package bowtie

import (
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"

	_ "embed"
)

type Config struct {
	SegementCount  int                    `json:"segmentCount,omitempty" yaml:"segmentCount,omitempty"`
	SegmentColours []parameters.HexString `json:"segmentColors,omitempty" yaml:"segmentColors,omitempty"`
	Blend          string                 `json:"blend,omitempty" yaml:"blend,omitempty"`

	parameters.RotationAngle `yaml:",inline"`
	parameters.StartAngle    `yaml:",inline"`
	parameters.Offset        `yaml:",inline"`
}

//go:embed jsonschema/jsonschema.json
var Schema []byte
