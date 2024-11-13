package framecount

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
)

type Config struct {
	//	Type         string            `json:"type" yaml:"type"`
	FrameCounter      bool        `json:"frameCounter,omitempty" yaml:"frameCounter,omitempty"`
	Imgpos            interface{} `json:"gridPosition" yaml:"gridPosition"`
	TextColour        string      `json:"textColor" yaml:"textColor"`
	BackColour        string      `json:"backgroundColor" yaml:"backgroundColor"`
	Font              string      `json:"font" yaml:"font"`
	FontSize          float64     `json:"fontSize" yaml:"fontSize"`
	config.WidgetGrid `yaml:",inline"`
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`

	//	DesignScale  string       `json:"designScale" yaml:"designScale"`
	// This is added in for metadata purposes
	FrameNumber int `json:"frameNumber"`
}

// start the count at -1 as it is incremented before being returned
var framecount = -1

//go:embed jsonschema/framecounter.json
var Schema []byte

func (f *Config) getFrames() bool {
	if f.FrameCounter {
		framecount++
	}

	return f.FrameCounter
}

func framePos() int {
	return framecount
}

/*
func (f frameJSON) Alias() string {
	return f.GridLoc.Alias
}

func (f frameJSON) Location() string {
	return f.GridLoc.Location
}*/
