package nearblack

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// Ebu3373/nearblack definitions
const wName = "nearblack"
const wType = "ebu3373/nearblack"
const wLibrary = "builtin"
const hooks = ""*/

type nearblackJSON struct {
	// Type    string      `json:"type" yaml:"type"`
	ColourSpace colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	GridLoc     config.Grid       `json:"grid,omitempty" yaml:"grid,omitempty"`
}

//go:embed jsonschema/nbschema.json
var schemaInit []byte

func (nb nearblackJSON) Alias() string {
	return nb.GridLoc.Alias
}

func (nb nearblackJSON) Location() string {
	return nb.GridLoc.Location
}
