package bars

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// Ebu3373/bars definitions
const wName = "bars"
const wType = "ebu3373/bars"
const wLibrary = "builtin"
const hooks = ""*/

type barJSON struct {
	//	Type    string      `json:"type" yaml:"type"`
	ColourSpace *colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	GridLoc     config.Grid        `json:"grid,omitempty" yaml:"grid,omitempty"`
}

//go:embed jsonschema/barschema.json
var schemaInit []byte

func (b barJSON) Alias() string {
	return b.GridLoc.Alias
}

func (b barJSON) Location() string {
	return b.GridLoc.Location
}

func (b barJSON) Wait() (bool, []string) {
	return false, []string{}
}
