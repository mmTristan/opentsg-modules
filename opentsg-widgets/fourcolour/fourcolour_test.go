package fourcolour

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	geometrymock "github.com/mrmxf/opentsg-modules/opentsg-widgets/geometryMock"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFillMethod(t *testing.T) {
	//rand.Seed(1320)
	randSrc := rand.New(rand.NewSource(1320))
	mg := geometrymock.Mockgeom(randSrc, 1000, 1000, 8)

	// mockG := config.Grid{Location: "Nothing"}
	mockJson4 := Config{Colourpallette: []parameters.HexString{"#FF0000", "#00FF00", "#0000FF", "#FFFF00", "#FF00FF"}}
	mockJson5 := Config{Colourpallette: []parameters.HexString{"#FF0000", "#00FF00", "#0000FF", "#FFFF00"}}
	mockJsons := []Config{mockJson4, mockJson5}

	explanation := []string{"fiveColour", "fourColour"}

	for i, mj := range mockJsons {

		canvas := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))

		out := tsg.TestResponder{BaseImg: canvas}
		mj.Handle(&out, &tsg.Request{PatchProperties: tsg.PatchProperties{Geometry: mg}})

		examplejson.SaveExampleJson(mj, WidgetType, explanation[i], false)

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
					So(out.Message, ShouldResemble, "success")
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}
	// save the image for four and five colour comparisons
}

func BenchmarkNRGBA64ACESColour(b *testing.B) {
	// decode to get the colour values
	randSrc := rand.New(rand.NewSource(1320))
	mg := geometrymock.Mockgeom(randSrc, 1000, 1000, 8)

	//	mockG := config.Grid{Location: "Nothing"}
	// mockJson := fourJSON{GridLoc: &mockG, Colourpallette: []string{"#FF0000", "#00FF00", "#0000FF", "#FFFF00", "#FF00FF"}}
	mockJson := Config{Colourpallette: []parameters.HexString{"#FF0000", "#00FF00", "#0000FF", "#FFFF00"}}
	canvas := image.NewNRGBA64(image.Rect(0, 0, 1, 1))

	out := tsg.TestResponder{BaseImg: canvas}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		mockJson.Handle(&out, &tsg.Request{PatchProperties: tsg.PatchProperties{Geometry: mg}})
	}
}

func BenchmarkNRGBA64ACESOTher(b *testing.B) {
	// decode to get the colour values
	randSrc := rand.New(rand.NewSource(1320))
	mg := geometrymock.Mockgeom(randSrc, 1000, 1000, 8)
	//	mockG := config.Grid{Location: "Nothing"}
	mockJson := Config{Colourpallette: []parameters.HexString{"#FF0000", "#00FF00", "#0000FF", "#FFFF00", "#FF00FF"}}

	canvas := image.NewNRGBA64(image.Rect(0, 0, 1, 1))
	out := tsg.TestResponder{BaseImg: canvas}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		mockJson.Handle(&out, &tsg.Request{PatchProperties: tsg.PatchProperties{Geometry: mg}})
	}
}
