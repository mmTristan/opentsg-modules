package framecount

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"runtime/debug"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/text"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	frameDemo := Config{FrameCounter: true}
	examplejson.SaveExampleJson(frameDemo, WidgetType, "minimum", true)

	frameDemoMax := Config{FrameCounter: true, TextColour: "#C2A649", BackColour: "#91B645", Font: text.FontTitle, FontSize: 25,
		Imgpos: topLeft}
	examplejson.SaveExampleJson(frameDemoMax, WidgetType, "maximum", true)

	frameDemoStyle := Config{FrameCounter: true, TextColour: "rgb(154,58,115)", BackColour: "rgb12(816,816,816)", Font: text.FontBody}
	examplejson.SaveExampleJson(frameDemoStyle, WidgetType, "styleChange", true)

}

// Wait for august to undo the bug it decided to make
func TestStringGen(t *testing.T) {
	// Check if this is a suitable build of go to run these tests
	bi, _ := debug.ReadBuildInfo()
	c := context.Background()

	// Keep this in the background unti it runs
	if bi.GoVersion[:6] != "go1.18" {
		numberToCheck := []int{0, 12, 134, 5666}
		expecResult := []string{"0000", "0012", "0134", "5666"}
		explanation := []string{"0000", "0012", "0134", "5666"}
		yesFrame := Config{FrameCounter: true, FontSize: 100}
		//	yesFrame.FrameCounter = true

		for i := range numberToCheck {
			// Generate the image and the string
			inter := func() int { return numberToCheck[i] }
			pos = inter
			myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{33, 33}})
			genErr := yesFrame.Generate(myImage, &c)

			examplejson.SaveExampleJson(yesFrame, WidgetType, explanation[i], false)

			// f, _ := os.Create("./testdata/framecount" + expecResult[i] + ".png")
			// png.Encode(f, myImage)

			// Assign the colour to the correct type of image NGRBA64 and replace the colour values
			file, _ := os.Open("./testdata/framecount" + expecResult[i] + ".png")
			// Decode to get the colour values
			baseVals, _ := png.Decode(file)

			// Assign the colour to the correct type of image NGRBA64 and replace the colour values
			readImage := image.NewNRGBA64(baseVals.Bounds())
			colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Src)
			// Make a hash of the pixels of each image
			hnormal := sha256.New()
			htest := sha256.New()
			hnormal.Write(readImage.Pix)
			htest.Write(myImage.Pix)

			// f, _ := os.Create("./testdata/framecount" + expecResult[i] + "2.png")
			// colour.PngEncode(f, myImage)
			// GenResult, genErr := intTo4(numberToCheck[i])
			Convey("Checking the frame count image is generated", t, func() {
				Convey(fmt.Sprintf("using  %v as integer ", numberToCheck[i]), func() {
					Convey(fmt.Sprintf("A 4 digit number of %v is expected and the generated sha256 are identical", expecResult[i]), func() {
						So(genErr, ShouldBeNil)
						So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
					})
				})
			})
		}

		// Delete the files afterwards
		for i := 0; i < len(numberToCheck); i++ {
			os.Remove("framecount" + expecResult[i] + ".png")
		}
	}
}

func TestFonts(t *testing.T) {
	// Test the size -but these aren't here yet with go1.19
	bi, _ := debug.ReadBuildInfo()
	c := context.Background()
	// Keep this in the background unti it runs
	if bi.GoVersion[:6] != "go1.18" {
		pos = func() int { return 567 }
		var mockFrame Config
		mockFrame.FrameCounter = true
		fontType := []string{"header", "", "./testdata/Timmy-Regular.ttf", "title"}
		explanation := []string{"header", "default", "imported", "title"}
		sizes := []float64{12, 22, 40, 24}

		for i, fon := range fontType {
			mockFrame.Font = fon
			mockFrame.FontSize = sizes[i]
			myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{100, 100}})
			genErr := mockFrame.Generate(myImage, &c)

			examplejson.SaveExampleJson(mockFrame, WidgetType, explanation[i], false)
			// Save these images when we can test for them
			//	f, _ := os.Create("./testdata/framecount" + fmt.Sprintf("%v", i) + "2.png")
			//	png.Encode(f, myImage)

			Convey("Checking the frame count image is generated with different fonts and sizes", t, func() {
				Convey(fmt.Sprintf("using  %v as the font", fon), func() {
					Convey("No error is expected", func() {
						So(genErr, ShouldBeNil)

					})
				})
			})
		}
	}
}

func TestErrors(t *testing.T) {
	c := context.Background()

	numberToCheck := []int{99999}
	expecResult := []string{"frame Count greater then 9999"}

	var yesFrame Config
	yesFrame.FrameCounter = true
	yesFrame.FontSize = 90

	for i, n := range numberToCheck {
		// Generate the image and the string
		inter := func() int { return n }
		pos = inter

		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{29, 29}})
		genErr := yesFrame.Generate(myImage, &c)

		Convey("Checking the frame count catches errors", t, func() {
			Convey(fmt.Sprintf("Checking for an error of  %v", expecResult[i]), func() {
				Convey("The expected error is caught", func() {
					So(genErr.Error(), ShouldEqual, expecResult[i])
				})
			})
		})
	}

}

func TestInterpret(t *testing.T) {
	xonly := `{
		"gridPosition":{
			"x":100
		}
	}`
	yonly := `{
		"gridPosition":{
			"y":100
		}
	}`
	both := `{
		"gridPosition":{
			"x" :2.6,
			"y":100
		}
	}`
	alias1 := `{
		"gridPosition":{
			"alias":"bottom right"
		}
	}`
	alias2 := `{
		"gridPosition":{
			"alias":"top right"
		}
	}`
	differentPos := []string{xonly, yonly, both, alias1, alias2}
	expectedX := []int{100, 0, 2, 89, 89}
	expectedY := []int{0, 100, 100, 89, 0}

	for i, testbody := range differentPos {
		body := []byte(testbody)
		var f Config
		json.Unmarshal(body, &f)
		//	fmt.Println(f.Imgpos)
		x, y := userPos(f.Imgpos.(map[string]interface{}), image.Point{100, 100}, image.Point{10, 10})
		// Generate the image and the string

		Convey("Checking the interface to position works", t, func() {
			Convey(fmt.Sprintf("using an input json of  %v", testbody), func() {
				Convey(fmt.Sprintf("An x and y of %v and %v are produced", expectedX[i], expectedY[i]), func() {
					So(x, ShouldEqual, expectedX[i])
					So(y, ShouldEqual, expectedY[i])
				})
			})
		})
	}

}
