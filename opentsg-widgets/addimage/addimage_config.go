package addimage

import (
	_ "embed"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"
)

type Config struct {
	// Type    string            `json:"type" yaml:"type"`
	Image string `json:"image" yaml:"image"`
	// Imgsize *config.Framesize `json:"imagesize,omitempty" yaml:"imagesize,omitempty"`
	//	Imgpos  *config.Position `json:"position,omitempty" yaml:"position,omitempty"`
	config.WidgetGrid `yaml:",inline"`
	ColourSpace       *colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	ImgFill           string             `json:"imageFill,omitempty" yaml:"imageFill,omitempty"`
	parameters.Offset `yaml:",inline"`
	// Position field
	/*
		centroid offset
		Offset interface
	*/
}

//go:embed jsonschema/addimageschema.json
var Schema []byte

/*
func (a addimageJSON) Alias() string {
	return a.GridLoc.Alias
}

func (a addimageJSON) Location() string {
	return a.GridLoc.Location
}
*/
