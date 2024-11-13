package noise

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
	noiseDemo := Config{NoiseType: whiteNoise}
	examplejson.SaveExampleJson(noiseDemo, WidgetType, "minimum", true)

	noiseDemoMax := Config{NoiseType: whiteNoise, Minimum: 2000, Maximum: 3000}
	examplejson.SaveExampleJson(noiseDemoMax, WidgetType, "maximum", true)
}

func TestWhiteNoise(t *testing.T) {
	var mockNoise Config

	mockNoise.NoiseType = whiteNoise
	randnum = func() int64 { return 27 }

	testF := []string{"./testdata/whitenoise.png"}
	explanation := []string{"whitenoise"}

	for i, compare := range testF {
		mockNoise.Maximum = 4095
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
		// Generate the noise image
		genErr := mockNoise.Generate(myImage)
		examplejson.SaveExampleJson(mockNoise, WidgetType, explanation[i], false)
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

	explanationG := []string{"leftGuillotine", "rightGuillotine", "TopLeftGuillotine", "TopRightGuillotine", "Diag"}
	offsets := []Guillotine{{BottomRight: 100}, {BottomLeft: 100}, {TopRight: 100}, {TopLeft: 100}, {TopLeft: 50, BottomRight: 50}}

	for i, off := range offsets {
		mockNoise.Maximum = 4095
		mockNoise.YOffsets = off
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 200}})
		// Generate the noise image
		genErr := mockNoise.Generate(myImage)
		examplejson.SaveExampleJson(mockNoise, WidgetType, explanationG[i], true)
		// Open the image to compare to
		file, _ := os.Open(fmt.Sprintf("./testdata/%s.png", explanationG[i]))
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

		// f, _ := os.Create(fmt.Sprintf("test%v.png", i))
		// png.Encode(f, myImage)
		// Save the file
		Convey("Checking that the noise is generated", t, func() {
			Convey(fmt.Sprintf("Comparing the generated image to %v ", explanationG[i]), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

	explanationBoth := []string{"bothBottom", "bothTop"}
	offsetBoth := []Guillotine{{BottomRight: 100, BottomLeft: 50}, {TopLeft: 100, TopRight: 50}}

	for i, off := range offsetBoth {
		mockNoise.Maximum = 4095
		mockNoise.YOffsets = off
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 200}})
		// Generate the noise image
		genErr := mockNoise.Generate(myImage)

		// Open the image to compare to

		file, _ := os.Open(fmt.Sprintf("./testdata/%s.png", explanationBoth[i]))
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

		//	f, _ := os.Create(fmt.Sprintf("./testdata/%s.png", explanationG[i]))
		//	png.Encode(f, myImage)
		// Save the file

		Convey("Checking that the noise is generated", t, func() {
			Convey(fmt.Sprintf("Comparing the generated image to %v ", explanationBoth[i]), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

}
