package saturation

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// Ebu3373/saturation definitions
const wName = "saturation"
const wType = "ebu3373/saturation"
const wLibrary = "builtin"
const hooks = ""*/

type saturationJSON struct {
	// Type    string       `json:"type" yaml:"type"`
	Colours     []string          `json:"colors,omitempty" yaml:"colors,omitempty"`
	ColourSpace colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	GridLoc     *config.Grid      `json:"grid,omitempty" yaml:"grid,omitempty"`
}

//go:embed jsonschema/satschema.json
var schemaInit []byte

func (s saturationJSON) Alias() string {
	return s.GridLoc.Alias
}

func (s saturationJSON) Location() string {
	return s.GridLoc.Location
}
