package luma

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// Ebu3373/luma definitions
const wName = "luma"
const wType = "ebu3373/luma"
const wLibrary = "builtin"
const hooks = ""*/

type lumaJSON struct {
	// Type    string      `json:"type" yaml:"type"`
	ColourSpace colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	GridLoc     config.Grid       `json:"grid,omitempty" yaml:"grid,omitempty"`
}

//go:embed jsonschema/lumaschema.json
var schemaInit []byte

func (l lumaJSON) Alias() string {
	return l.GridLoc.Alias
}

func (l lumaJSON) Location() string {
	return l.GridLoc.Location
}
