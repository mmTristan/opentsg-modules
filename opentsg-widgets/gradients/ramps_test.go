package gradients

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
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/text"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	tbDemo := Ramp{}
	examplejson.SaveExampleJson(tbDemo, WidgetType, "minimum", false)

	mockFull := Ramp{Groups: []RampProperties{{Colour: "green", InitialPixelValue: 960, Reverse: true}, {Colour: "gray", InitialPixelValue: 960, Reverse: true}, {Colour: "blue", InitialPixelValue: 0}, {Colour: "red", InitialPixelValue: 0}},
		Gradients: groupContents{GroupSeparator: groupSeparator{Height: 2, Colour: "white"},
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6",
			TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}, PixelValueRepeat: 1}}

	// set the angle later
	mockFull.WidgetProperties.CwRotation = "π*31/20"
	examplejson.SaveExampleJson(mockFull, WidgetType, "maximum", true)

	mockNoGroupDiv := Ramp{Groups: []RampProperties{{Colour: "green"}, {Colour: "blue", InitialPixelValue: 0}, {Colour: "red", InitialPixelValue: 0}},
		Gradients: groupContents{
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6",
			TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}, PixelValueRepeat: 1, ObjectFitFill: true}}

	examplejson.SaveExampleJson(mockNoGroupDiv, WidgetType, "noGroupSeparator", true)

	mockNoGradDiv := Ramp{Groups: []RampProperties{{Colour: "green"}, {Colour: "blue", InitialPixelValue: 0}, {Colour: "red", InitialPixelValue: 0}},
		Gradients: groupContents{
			GroupSeparator: groupSeparator{Height: 4, Colour: "grey"},
			Gradients:      []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}}},
		WidgetProperties: control{MaxBitDepth: 8, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6"}, PixelValueRepeat: 1, ObjectFitFill: false}}

	examplejson.SaveExampleJson(mockNoGradDiv, WidgetType, "noGradientSeparator", true)

	mockNoText := Ramp{Groups: []RampProperties{{Colour: "red", InitialPixelValue: 128}, {Colour: "green", InitialPixelValue: 128}, {Colour: "blue", InitialPixelValue: 128}, {Colour: "grey", InitialPixelValue: 128}},
		Gradients: groupContents{
			Gradients: []Gradient{{Height: 5, BitDepth: 4}, {Height: 5, BitDepth: 6}, {Height: 5, BitDepth: 8}}},
		WidgetProperties: control{MaxBitDepth: 8, PixelValueRepeat: 1, ObjectFitFill: false}}

	examplejson.SaveExampleJson(mockNoText, WidgetType, "noText", true)

}

func TestTemp(t *testing.T) {
	mock := Ramp{Groups: []RampProperties{{Colour: "green", InitialPixelValue: 960}, {Colour: "gray", InitialPixelValue: 960}},
		Gradients: groupContents{GroupSeparator: groupSeparator{Height: 0, Colour: "white"},
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black", "red", "blue"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6", TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}}}
	tester := image.NewNRGBA64(image.Rect(0, 0, 1024, 1000)) // 960))
	mock.Generate(tester)

	examplejson.SaveExampleJson(mock, WidgetType, "demo", false)

	f, _ := os.Create("./testdata/tester.png")
	png.Encode(f, tester)

}

func TestRotation(t *testing.T) {

	mock := Ramp{Groups: []RampProperties{{Colour: "green", InitialPixelValue: 960}, {Colour: "gray", InitialPixelValue: 960}},
		Gradients: groupContents{GroupSeparator: groupSeparator{Height: 0, Colour: "white"},
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black", "red", "blue"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6", TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}}}

	explanationRight := []string{"flat", "90degrees", "180degrees", "270degrees"}
	anglesRight := []string{"", "π*1/2", "π*1", "π*3/2"}
	testFRight := []string{"./testdata/test.png", "./testdata/test90.png", "./testdata/test180.png", "./testdata/test270.png"}

	for i, angle := range anglesRight {

		mock.WidgetProperties.CwRotation = angle

		angleImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{4096, 2000}})
		examplejson.SaveExampleJson(mock, WidgetType, explanationRight[i], false)
		genErr := mock.Generate(angleImage)

		// Generate the ramp image
		// genErr := mock.Generate(myImage)
		// Open the image to compare to
		file, _ := os.Open(testFRight[i])
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)
		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)
		png.Encode(file, angleImage)
		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(angleImage.Pix)
		// f, _ := os.Create(testFRight[i] + ".png")
		// png.Encode(f, angleImage)

		Convey("Checking the ramps are generated at 90 degree angles", t, func() {
			Convey(fmt.Sprintf("Comparing the generated ramp to %v with an angle of %v", testFRight[i], angle), func() {
				Convey("No error is returned and the file matches", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

	anglesOffRight := []string{"π*1/20", "π*5/12", "π*9/10", "π*31/20"}
	explanation := []string{"9degrees", "75degrees", "162degrees", "279degrees"}
	testFRightOff := []string{"./testdata/angLinear.png", "./testdata/ang90.png", "./testdata/ang180.png", "./testdata/ang270.png"}

	for i, angle := range anglesOffRight {
		mock.WidgetProperties.CwRotation = angle
		angleImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{4096, 2000}})
		examplejson.SaveExampleJson(mock, WidgetType, explanation[i], false)
		// Generate the ramp image
		genErr := mock.Generate(angleImage)
		// Open the image to compare to
		file, _ := os.Open(testFRightOff[i])

		png.Encode(file, angleImage)
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)
		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(angleImage.Pix)

		// f, _ := os.Create(testFRightOff[i] + ".png")
		// 	png.Encode(f, angleImage)

		Convey("Checking the ramps are generated at angles other than 90 degrees", t, func() {
			Convey(fmt.Sprintf("Comparing the generated ramp to %v with an angle of %v", testFRightOff[i], angle), func() {
				Convey("No error is returned and the file matches", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

}
