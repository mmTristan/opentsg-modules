package bars

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	barsDemo := BarJSON{}
	examplejson.SaveExampleJson(barsDemo, WidgetType, "base", true)
}

func TestBars(t *testing.T) {

	sizes := [][2]int{{3840, 1160}, {1920, 580}, {500, 1900}}
	testBase := []string{"testdata/uhd.png", "testdata/hd.png", "testdata/obtuse.png"}
	explanation := []string{"uhd", "hd", "obtuse"}

	for i, size := range sizes {
		mock := BarJSON{}
		mock.GridLoc = &config.Grid{Alias: "testlocation"}
		myImage := image.NewNRGBA64(image.Rect(0, 0, size[0], size[1]))

		examplejson.SaveExampleJson(mock, WidgetType, explanation[i], false)
		// Generate the ramp image
		genErr := mock.Generate(myImage)
		// Open the image to compare to
		file, _ := os.Open(testBase[i])
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
		compare(readImage, myImage)
		Convey("Checking the bar functions are generated correctly", t, func() {
			Convey(fmt.Sprintf("Comparing the generated ramp to %v", testBase[i]), func() {
				Convey("No error is returned and the file matches", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}

func compare(base, new draw.Image) {
	count := 0
	b := base.Bounds().Max
	for x := 0; x < b.X; x++ {
		for y := 0; y < b.Y; y++ {
			if base.At(x, y) != new.At(x, y) {
				count++
			}

		}

	}

	fmt.Println(count, "non matches")
}
