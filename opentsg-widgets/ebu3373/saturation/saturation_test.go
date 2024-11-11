package saturation

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	satDemo := Config{}
	examplejson.SaveExampleJson(satDemo, WidgetType, "minimum", true)

	satDemoMax := Config{Colours: []string{"red", "green", "blue"}}
	examplejson.SaveExampleJson(satDemoMax, WidgetType, "maximum", true)

	satDemoDiff := Config{Colours: []string{"blue", "red", "green"}}
	examplejson.SaveExampleJson(satDemoDiff, WidgetType, "diff", true)

}

func TestBars(t *testing.T) {
	myImage := image.NewNRGBA64(image.Rect(0, 0, 2330, 600))
	s := Config{}
	colours := [][]string{{"red", "green", "blue"}, {"red", "blue"}, {"blue"}, {}}
	explanation := []string{"redGreenBlue", "redBlue", "blue", "defualt"}

	for i, c := range colours {
		s.Colours = c
		genErr := s.Generate(myImage)
		examplejson.SaveExampleJson(s, WidgetType, explanation[i], false)

		f, _ := os.Open(fmt.Sprintf("./testdata/ordertest%v.png", i))

		baseVals, _ := png.Decode(f)
		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(myImage.Pix)
		// F, _ := os.Create(testF[i] + fmt.Sprintf("%v.png", i))
		// Png.Encode(f, myImage)

		Convey("Checking saturations ramps can be generated for differenent colours", t, func() {
			Convey("Comparing the generated ramp to the base test", func() {
				Convey("No error is returned and the file matches", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}
}
