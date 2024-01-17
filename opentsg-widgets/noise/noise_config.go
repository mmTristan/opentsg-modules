package noise

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// noise definitions
const wName = "noise"
const wType = "noise"
const wLibrary = "builtin"
const hooks = ""*/

type noiseJSON struct {
	//	Type      string       `json:"type" yaml:"type"`
	NoiseType   string            `json:"noiseType" yaml:"noiseType"`
	Minimum     int               `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum     int               `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	ColourSpace colour.ColorSpace `json:"colorSpace" yaml:"colorSpace"`
	GridLoc     *config.Grid      `json:"grid,omitempty" yaml:"grid,omitempty"`
}

//go:embed jsonschema/noiseschema.json
var schemaInit []byte

func (n noiseJSON) Alias() string {
	return n.GridLoc.Alias
}

func (n noiseJSON) Location() string {
	return n.GridLoc.Location
}
