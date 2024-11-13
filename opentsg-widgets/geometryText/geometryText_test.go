package geometrytext

import (
	"context"
	"image"
	"image/png"
	"math/rand"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	geometrymock "github.com/mrmxf/opentsg-modules/opentsg-widgets/geometryMock"
	// . "github.com/smartystreets/goconvey/convey"
)

func TestFillMethod(t *testing.T) {
	rand.Seed(1320)
	mg := geometrymock.Mockgeom(1000, 1000)
	getGeometry = func(c *context.Context, coordinate string) ([]gridgen.Segmenter, error) {
		return mg, nil
	}
	mockG := config.Grid{Location: "Nothing"}
	mockJson4 := Config{GridLoc: &mockG, TextColour: "#C2A649"}
	canvas := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
	c := context.Background()
	mockJson4.Generate(canvas, &c)
	f, _ := os.Create("./testdata/generatecheck2" + ".png")
	png.Encode(f, canvas)

	//	mockJson5 := fourJSON{GridLoc: &mockG, Colourpallette: []string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00"}}

	/*for _, mj := range mockJsons {
		// check the rectangle matches init
		canvas := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		c := context.Background()
		genErr := mj.Generate(canvas, &c)

		f, _ := os.Open("./testdata/generatecheck" + fmt.Sprint(len(mj.Colourpallette)) + ".png")
		baseVals, _ := png.Decode(f)

		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		hnormal := sha256.New()
		htest := sha256.New()

		hnormal.Write(readImage.Pix)
		htest.Write(canvas.Pix)
		//	for mock

		Convey("Checking the algorthim fills in the sqaures without error", t, func() {
			Convey(fmt.Sprintf("Using a colour pallette of %v colours", len(mj.Colourpallette)), func() {
				Convey("No error is generated and the image matches the expected one", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}*/
	// save the image for four and five colour comparisons
}
