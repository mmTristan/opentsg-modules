package geometrytext

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	geometrymock "github.com/mrmxf/opentsg-modules/opentsg-widgets/geometryMock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFillMethod(t *testing.T) {
	rand.Seed(1320)
	/*mg := geometrymock.Mockgeom(1000, 1000)
	getGeometry = func(c *context.Context, coordinate string) ([]gridgen.Segmenter, error) {
		return mg, nil
	}*/

	mockG := config.Grid{Location: "Nothing"}
	mockJson4 := Config{GridLoc: &mockG, TextColour: "#C2A649"}

	nameLength := []int{8, 12, 16, 18}

	for _, n := range nameLength {

		mg := geometrymock.Mockgeom(1000, 1000, n)
		getGeometry = func(c *context.Context, coordinate string) ([]gridgen.Segmenter, error) {
			return mg, nil
		}

		canvas := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		colour.Draw(canvas, canvas.Bounds(), &image.Uniform{color.NRGBA64{R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}}, image.Point{0, 0}, draw.Over)

		c := context.Background()
		genErr := mockJson4.Generate(canvas, &c)
		f, _ := os.Create(fmt.Sprintf("./testdata/generatecheck%v.png", n))
		png.Encode(f, canvas)

		file, _ := os.Open(fmt.Sprintf("./testdata/generatecheck%v.png", n))
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// de, _ := os.Create(fmt.Sprintf("./testdata/generatecheck%v.png.png", i))
		// png.Encode(de, readImage)

		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(canvas.Pix)

		for i, p := range readImage.Pix {
			if p != canvas.Pix[i] {
				//	fmt.Println(i, p, canvas.Pix[i], reflect.TypeOf(canvas), reflect.TypeOf(baseVals))
			}
		}
		// f, _ := os.Create(testFRight[i] + ".png")
		// png.Encode(f, angleImage)

		Convey("Checking the ramps are generated at 90 degree angles", t, func() {
			Convey(fmt.Sprintf("Comparing the generated ramp to %v with an angle of %v", "testFRight[i]", "angle"), func() {
				Convey("No error is returned and the file matches", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

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
