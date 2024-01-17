package qrgen

import (
	_ "embed"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-core/config"
)

/*
// qrcode definitions
const wName = "qrcode"
const wType = "qrcode"
const wLibrary = "builtin"
const hooks = "[framecount]"*/

type qrcodeJSON struct {
	// Type    string             `json:"type" yaml:"type"`
	Code        string             `json:"code" yaml:"code"`
	Imgpos      *config.Position   `json:"gridPosition,omitempty" yaml:"gridPosition,omitempty"`
	Size        *sizeJSON          `json:"size,omitempty" yaml:"size,omitempty"`
	Query       *[]objectQueryJSON `json:"objectQuery,omitempty" yaml:"objectQuery,omitempty"`
	GridLoc     *config.Grid       `json:"grid,omitempty" yaml:"grid,omitempty"`
	ColourSpace colour.ColorSpace  `json:"colorSpace" yaml:"colorSpace"`
}

type sizeJSON struct {
	Width  float64 `json:"width" yaml:"width"`
	Height float64 `json:"height" yaml:"height"`
}

type objectQueryJSON struct {
	Target string   `json:"targetAlias" yaml:"targetAlias"`
	Keys   []string `json:"keys" yaml:"keys"`
}

//go:embed jsonschema/qrgenschema.json
var schemaInit []byte

func (q qrcodeJSON) Alias() string {
	return q.GridLoc.Alias
}

func (q qrcodeJSON) Location() string {
	return q.GridLoc.Location
}
