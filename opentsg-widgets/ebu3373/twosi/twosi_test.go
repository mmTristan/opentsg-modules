package twosi

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	getPostion = func(gridString, alias string, c *context.Context) (draw.Image, image.Point, draw.Image, error) {
		return nil, image.Point{}, nil, nil
	}
	// base example
	twosiDemo := Config{}
	twosiDemo.GridLoc = &config.Grid{Alias: "testlocation"}
	examplejson.SaveExampleJson(twosiDemo, WidgetType, "base", true)

	getPostion = gridgen.GridSquareLocatorAndGenerator

}

func TestChannels(t *testing.T) {

	sizes := [][2]int{{1510, 600}, {755, 300}, {2000, 400}}
	testBase := []string{"testdata/uhd", "testdata/hd", "testdata/obtuse"}
	explanation := []string{"uhd", "hd", "obtuse"}

	for i, size := range sizes {
		mock := Config{}
		mock.GridLoc = &config.Grid{Alias: "testlocation"}
		myImage := image.NewNRGBA64(image.Rect(0, 0, size[0], size[1]))
		examplejson.SaveExampleJson(mock, WidgetType, explanation[i], false)
		// Generate the ramp image
		_ = mock.Generate(myImage)

		offsets := [][2]int{{0, 0}, {2, 0}, {0, 1}, {2, 1}}
		b := myImage.Bounds().Max
		let := []string{"A", "B", "C", "D"}
		for j, off := range offsets {

			chunk := image.NewNRGBA64(myImage.Bounds())
			colour.Draw(chunk, chunk.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

			maskC := mask(b.X, b.Y, off[0], off[1])
			colour.DrawMask(chunk, chunk.Bounds(), myImage, image.Point{}, maskC, image.Point{}, draw.Over)

			// f, _ := os.Create(testBase[i] + let[j] + ".png")
			// png.Encode(f, chunk)

			file, _ := os.Open(testBase[i] + let[j] + ".png")
			// Decode to get the colour values
			baseVals, _ := png.Decode(file)
			// Assign the colour to the correct type of image NGRBA64 and replace the colour values
			readImage := image.NewNRGBA64(baseVals.Bounds())
			colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

			hnormal := sha256.New()
			htest := sha256.New()
			hnormal.Write(readImage.Pix)
			htest.Write(chunk.Pix)

			// f, _ := os.Create(testBase[i] + let[j] + "er.png")
			// colour.PngEncode(f, chunk)

			Convey("Checking the twosi images are generated", t, func() {
				Convey(fmt.Sprintf("Comparing the generated image to the channe, %v%v.png", testBase[i], let[j]), func() {
					Convey("No error is returned and the file matches", func() {
						So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
					})
				})
			})

		}
	}
	// Generate this for other
}
