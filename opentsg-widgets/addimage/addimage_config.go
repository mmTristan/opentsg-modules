package addimage

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// Addimage definitions
const wName = "addimage"
const wType = "addimage"
const wLibrary = "builtin"
const hooks = ""*/

type addimageJSON struct {
	// Type    string            `json:"type" yaml:"type"`
	Image string `json:"image" yaml:"image"`
	// Imgsize *config.Framesize `json:"imagesize,omitempty" yaml:"imagesize,omitempty"`
	//	Imgpos  *config.Position `json:"position,omitempty" yaml:"position,omitempty"`
	GridLoc     *config.Grid       `json:"grid,omitempty" yaml:"grid,omitempty"`
	ColourSpace *colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	ImgFill     string             `json:"imageFill,omitempty" yaml:"imageFill,omitempty"`
}

//go:embed jsonschema/addimageschema.json
var schemaInit []byte

func (a addimageJSON) Alias() string {
	return a.GridLoc.Alias
}

func (a addimageJSON) Location() string {
	return a.GridLoc.Location
}
