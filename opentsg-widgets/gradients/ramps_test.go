package gradients

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
	"github.com/mmTristan/opentsg-modules/opentsg-widgets/text"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	tbDemo := Ramp{}
	examplejson.SaveExampleJson(tbDemo, widgetType, "minimum", false)

	mockFull := Ramp{Groups: []RampProperties{{Colour: "green", InitialPixelValue: 960, Reverse: true}, {Colour: "gray", InitialPixelValue: 960, Reverse: true}, {Colour: "blue", InitialPixelValue: 0}, {Colour: "red", InitialPixelValue: 0}},
		Gradients: groupContents{GroupSeparator: groupSeparator{Height: 2, Colour: "white"},
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6",
			TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}, CwRotation: "π*31/20", PixelValueRepeat: 1}}
	examplejson.SaveExampleJson(mockFull, widgetType, "maximum", true)

	mockNoGroupDiv := Ramp{Groups: []RampProperties{{Colour: "green"}, {Colour: "blue", InitialPixelValue: 0}, {Colour: "red", InitialPixelValue: 0}},
		Gradients: groupContents{
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6",
			TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}, PixelValueRepeat: 1, ObjectFitFill: true}}

	examplejson.SaveExampleJson(mockNoGroupDiv, widgetType, "noGroupSeparator", true)

	mockNoGradDiv := Ramp{Groups: []RampProperties{{Colour: "green"}, {Colour: "blue", InitialPixelValue: 0}, {Colour: "red", InitialPixelValue: 0}},
		Gradients: groupContents{
			GroupSeparator: groupSeparator{Height: 4, Colour: "grey"},
			Gradients:      []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}}},
		WidgetProperties: control{MaxBitDepth: 8, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6"}, PixelValueRepeat: 1, ObjectFitFill: false}}

	examplejson.SaveExampleJson(mockNoGradDiv, widgetType, "noGradientSeparator", true)

	mockNoText := Ramp{Groups: []RampProperties{{Colour: "red", InitialPixelValue: 128}, {Colour: "green", InitialPixelValue: 128}, {Colour: "blue", InitialPixelValue: 128}, {Colour: "grey", InitialPixelValue: 128}},
		Gradients: groupContents{
			Gradients: []Gradient{{Height: 5, BitDepth: 4}, {Height: 5, BitDepth: 6}, {Height: 5, BitDepth: 8}}},
		WidgetProperties: control{MaxBitDepth: 8, PixelValueRepeat: 1, ObjectFitFill: false}}

	examplejson.SaveExampleJson(mockNoText, widgetType, "noText", true)

}

func TestTemp(t *testing.T) {
	mock := Ramp{Groups: []RampProperties{{Colour: "green", InitialPixelValue: 960}, {Colour: "gray", InitialPixelValue: 960}},
		Gradients: groupContents{GroupSeparator: groupSeparator{Height: 0, Colour: "white"},
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black", "red", "blue"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6", TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}}}
	tester := image.NewNRGBA64(image.Rect(0, 0, 1024, 1000)) //960))
	mock.Generate(tester)

	examplejson.SaveExampleJson(mock, widgetType, "demo", false)

	f, _ := os.Create("./testdata/tester.png")
	png.Encode(f, tester)

	/*
		rotates := []string{"π*1/2", "π*3/2", "π*2/2", "π*5/8"}
		names := []string{"tester90.png", "tester270.png", "tester180.png", "testerwonk.png"}

		for i, ang := range rotates {
			mockAngle := mock
			mockAngle.WidgetProperties.CwRotation = ang
			// mock.WidgetProperties.TextProperties = textObjectJSON{TextColour: "#F32399"}

			testerAng := image.NewNRGBA64(image.Rect(0, 0, 4000, 4000))
			firstrun(testerAng, mockAngle)

			fang, _ := os.Create(names[i])
			png.Encode(fang, testerAng)
		}

		mock.Groups = []RampProperties{{Colour: "gray", InitialPixelValue: 1023, Reverse: true}}
		mock.WidgetProperties = control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextColour: "#F32399"}}
		mock.Gradients.Gradients = []Gradient{{Height: 20, BitDepth: 8}, {Height: 20, BitDepth: 4}}
		tester2 := image.NewNRGBA64(image.Rect(0, 0, 5000, 1000))
		firstrun(tester2, mock)

		f2, _ := os.Create("tester2.png")
		png.Encode(f2, tester2)

		mock = Ramp{Groups: []RampProperties{{Colour: "green", InitialPixelValue: 960}, {Colour: "gray", InitialPixelValue: 960}},
			Gradients: groupContents{GroupSeparator: groupSeparator{Height: 0, Colour: "white"},
				GradientSeparator: gradientSeparator{Colours: []string{"white", "black", "red", "blue"}, Height: 1},
				Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
			WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextColour: "#345AB6", TextHeight: 70}}}
		Squeezer := image.NewNRGBA64(image.Rect(0, 0, 5000, 1000)) //960))
		mock.WidgetProperties.ObjectFitFill = true
		firstrun(Squeezer, mock)
		fsqueeze, _ := os.Create("testerSqu.png")
		png.Encode(fsqueeze, Squeezer)

		twoer := image.NewNRGBA64(image.Rect(0, 0, 5000, 1000)) //960))
		mock.WidgetProperties.ObjectFitFill = false
		mock.WidgetProperties.PixelValueRepeat = 2
		firstrun(twoer, mock)
		fstwo, _ := os.Create("testerTwo.png")
		png.Encode(fstwo, twoer)
	*/
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
		examplejson.SaveExampleJson(mock, widgetType, explanationRight[i], false)
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
		examplejson.SaveExampleJson(mock, widgetType, explanation[i], false)
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

		//f, _ := os.Create(testFRightOff[i] + ".png")
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

/*
this was a test to ensure colour spaces were being taken into account with the widgets.
func TestCspace(t *testing.T) {
	mockSpace := Ramp{Groups: []RampProperties{{Colour: "green", InitialPixelValue: 960}, {Colour: "red", InitialPixelValue: 960}},
		Gradients: groupContents{GroupSeparator: groupSeparator{Height: 0, Colour: "white"},
			GradientSeparator: gradientSeparator{Colours: []string{"white", "black", "red", "blue"}, Height: 1},
			Gradients:         []Gradient{{Height: 5, BitDepth: 4, Label: "4b"}, {Height: 5, BitDepth: 6, Label: "6b"}, {Height: 5, BitDepth: 8, Label: "8b"}, {Height: 5, BitDepth: 10, Label: "10b"}}},
		WidgetProperties: control{MaxBitDepth: 10, TextProperties: textObjectJSON{TextHeight: 30, TextColour: "#345AB6", TextXPosition: text.AlignmentLeft, TextYPosition: text.AlignmentTop}}}

	spaces := []colour.ColorSpace{{ColorSpace: "rec709"}, {ColorSpace: "rec2020"}}
	baseImage := colour.NewNRGBA64(colour.ColorSpace{ColorSpace: "rec2020"}, image.Rect(0, 0, 2000, 2000))
	for i, s := range spaces {
		mockSpace.ColourSpace = s

		base := colour.NewNRGBA64(colour.ColorSpace{ColorSpace: "rec2020"}, image.Rect(0, 0, 2000, 1000))
		// base = colour.NewNRGBA64(colour.ColorSpace{}, image.Rect(0, 0, 2000, 1000))
		fmt.Println(mockSpace.Generate(base))

	//	fmt.Println(base.At(500, 500))
		//base.Set(500, 500, &colour.CNRGBA64{R: 65335, A: 0xffff, Space: colour.ColorSpace{ColorSpace: "rec709"}})

	//	colour.Draw(base, image.Rect(400, 400, 600, 600), &image.Uniform{&colour.CNRGBA64{R: 65335, A: 0xffff, Space: s}}, image.Point{}, draw.Over)
	//	fmt.Println(base.At(500, 500))
		f, _ := os.Create(fmt.Sprintf("test%v.png", i))
		png.Encode(f, base)

		colour.Draw(baseImage, image.Rect(0, 1000*i, 2000, 10000*(i+1)), base, image.Point{}, draw.Over)
	}

	f, _ := os.Create("test.png")
	png.Encode(f, baseImage)
}
*/
