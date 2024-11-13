package fourcolour

import (
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type Config struct {
	Colourpallette    []string          `json:"colors" yaml:"colors"`
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	config.WidgetGrid `yaml:",inline"`
}

var Schema = []byte(`{
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"$id": "https://example.com/product.schema.json",
	"title": "Allow anything through for tests",
	"description": "An empty schema to allow custom structs to run through",
	"type": "object"
	}`)

/*
func (f fourJSON) Alias() string {
	return f.GridLoc.Alias
}

func (f fourJSON) Location() string {
	return f.GridLoc.Location
}*/
