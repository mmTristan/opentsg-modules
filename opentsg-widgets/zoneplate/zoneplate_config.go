package zoneplate

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// zoneplate definitions
const wName = "zone plate"
const wType = "zoneplate"
const wLibrary = "builtin"
const hooks = ""*/

type zoneplateJSON struct {
	Platetype   string            `json:"plateType,omitempty" yaml:"plateType,omitempty"`
	ColourSpace colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	Startcolour string            `json:"startColor,omitempty" yaml:"startColor,omitempty"`
	Angle       interface{}       `json:"angle,omitempty" yaml:"angle,omitempty"`
	// Mask        string            `json:"mask,omitempty" yaml:"mask,omitempty"`
	GridLoc *config.Grid `json:"grid,omitempty" yaml:"grid,omitempty"`
}

//go:embed jsonschema/zoneplateschema.json
var schemaInit []byte

func (z zoneplateJSON) Alias() string {
	return z.GridLoc.Alias
}

func (z zoneplateJSON) Location() string {
	return z.GridLoc.Location
}
