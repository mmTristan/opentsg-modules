package noise

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	examplejson "github.com/mmTristan/opentsg-modules/opentsg-widgets/exampleJson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	noiseDemo := noiseJSON{NoiseType: whiteNoise}
	examplejson.SaveExampleJson(noiseDemo, widgetType, "minimum", true)

	noiseDemoMax := noiseJSON{NoiseType: whiteNoise, Minimum: 2000, Maximum: 3000}
	examplejson.SaveExampleJson(noiseDemoMax, widgetType, "maximum", true)
}

func TestZoneGenAngle(t *testing.T) {
	var mockNoise noiseJSON

	mockNoise.NoiseType = whiteNoise
	randnum = func() int64 { return 27 }

	testF := []string{"./testdata/whitenoise.png"}
	explanation := []string{"whitenoise"}

	for i, compare := range testF {
		mockNoise.Maximum = 4095
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
		// Generate the noise image
		genErr := mockNoise.Generate(myImage)
		examplejson.SaveExampleJson(mockNoise, widgetType, explanation[i], false)
		// Open the image to compare to
		file, _ := os.Open(compare)
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(myImage.Pix)
		// Save the file
		Convey("Checking that the noise is generated", t, func() {
			Convey(fmt.Sprintf("Comparing the generated image to %v ", compare), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}
