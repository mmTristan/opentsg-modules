package textbox

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/text"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	tbDemo := TextboxJSON{}
	examplejson.SaveExampleJson(tbDemo, WidgetType, "minimum", false)

	tbDemoMax := TextboxJSON{Text: []string{"sample", "text"}, Border: "#91B645", BorderSize: 5,
		Back: "#eeeeee", Textc: "#C2A649", Font: text.FontPixel, FillType: text.FillTypeRelaxed,
		XAlignment: text.AlignmentLeft, YAlignment: text.AlignmentBottom}
	examplejson.SaveExampleJson(tbDemoMax, WidgetType, "maximum", true)

	tbDemoNoborder := TextboxJSON{Text: []string{"sample", "text"},
		Back: "#333333", Textc: "#C2A649"}
	examplejson.SaveExampleJson(tbDemoNoborder, WidgetType, "noborder", true)

	tbDemoNoText := TextboxJSON{Border: "#EEE", BorderSize: 25,
		Back: "rgb12(3566,2312,34)"}
	examplejson.SaveExampleJson(tbDemoNoText, WidgetType, "notext", true)

}

func TestLines(t *testing.T) {
	mockContext := context.Background()

	mockTB := TextboxJSON{
		Textc: "#C2A649", Border: "#f0f0f0", Back: "#ffffff", BorderSize: 20,
		YAlignment: text.AlignmentMiddle}
	stringsToCheck := [][]string{{"sample text"}, {"sample", "text"}}
	original := []string{"./testdata/singleline.png", "./testdata/multiline.png"}
	explanation := []string{"singleline", "multiline"}

	for i, str := range stringsToCheck {

		myImage := colour.NewNRGBA64(colour.ColorSpace{}, image.Rectangle{image.Point{0, 0}, image.Point{1024, 240}})
		mockTB.Text = str
		genErr := mockTB.Generate(myImage, &mockContext)
		examplejson.SaveExampleJson(mockTB, WidgetType, explanation[i], false)

		// f, _ := os.Create(original[i])
		// png.Encode(f, myImage)

		file, _ := os.Open(original[i])
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(myImage.Pix())

		// f, _ := os.Create("./testdata/" + fmt.Sprintf("%v", i) + ".png")
		// colour.PngEncode(f, myImage)
		// Save the file
		Convey("Checking that strings are generated", t, func() {
			Convey(fmt.Sprintf("Generating an image with the following strings: %v ", str), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}
func TestFontImport(t *testing.T) {

	/*

		set up some demos

		empty background
		no border - colour or width
		no text

	*/

	// these are some example jsons
	tests := []TextboxJSON{{Textc: "rgb12(3000,1401,1116)", Font: text.FontHeader, Text: []string{"sample", "text"}, FillType: text.FillTypeFull},
		{Textc: "rgb12(3000,1401,1116)", Back: "#99AD49", Font: text.FontBody, Text: []string{"sample", "text"}, XAlignment: text.AlignmentLeft, YAlignment: text.AlignmentTop},
		{Textc: "rgb12(3000,1401,1116)", Back: "#99AD49", Border: "#000000", BorderSize: 5, Font: text.FontTitle, Text: []string{"sample", "text"}, XAlignment: text.AlignmentRight, YAlignment: text.AlignmentBottom},
		{Back: "rgb(255,255,0)", Font: text.FontHeader, Text: []string{"sample", "text"}},
		{Border: "rgb(134,24,180)", Text: []string{"sample", "text"}, BorderSize: 10},
		{Border: "rgb(134,24,180)", Font: text.FontHeader, Text: []string{"sample", "text"}, BorderSize: 0},
		{Textc: "rgb(134,24,180)", Back: "rgb(255,255,0)", Border: "rgb(134,24,180)", Font: text.FontPixel, Text: []string{"example space", "rec2020"}, BorderSize: 5, ColourSpace: colour.ColorSpace{ColorSpace: "rec2020"}, XAlignment: text.AlignmentMiddle, YAlignment: text.AlignmentMiddle},
	}

	explanation := []string{"text-only", "text-background", "text-background-border", "background", "border", "nothing", "rec2020"}

	// generate the jsons as a list of examples
	for i, e := range explanation {
		//	bc := context.Background()
		//	vase := image.NewNRGBA64(image.Rect(0, 0, 1000, 100))
		//	tests[i].Generate(vase, bc)
		examplejson.SaveExampleJson(tests[i], WidgetType, e, false)
		// f, _ := os.Create(e + ".png")
		// colour.PngEncode(f, vase)
	}
	//	mockContext := context.Background()

	base := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
	//	text := texter.TextboxJSON{Textc: "#260498", Back: "#980609"}
	genErr := TextboxJSON{Border: "#800080", BorderSize: 5, Textc: "#260498", Back: "#980609", Text: []string{"The quick",
		"brown dog jumped", "over the lazy gray fox"}, Font: `https://get.fontspace.co/webfont/XqrG/OGU3MmU3NmQzZGM2NGExZmFhNDY2YTk5MzhlNWMzMjYudHRm/helloween-2.ttf`}.Generate(base)

	// f, _ := os.Create("testdata/multiLongLines.png")
	// png.Encode(f, base)
	file, _ := os.Open("testdata/multiLongLines.png")
	// Decode to get the colour values
	baseVals, _ := png.Decode(file)

	// Assign the colour to the correct type of image NGRBA64 and replace the colour values
	readImage := image.NewNRGBA64(baseVals.Bounds())
	colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

	// Make a hash of the pixels of each image
	hnormal := sha256.New()
	htest := sha256.New()
	hnormal.Write(readImage.Pix)
	htest.Write(base.Pix)

	// f, _ := os.Create("./testdata/" + fmt.Sprintf("%v", i) + ".png")
	// colour.PngEncode(f, myImage)
	// Save the file
	Convey("Checking that multiple lines of small text are included", t, func() {
		Convey("Generating an image with an imported string", func() {
			Convey("No error is returned and the file matches exactly", func() {
				So(genErr, ShouldBeNil)
				So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
			})
		})
	})

}
