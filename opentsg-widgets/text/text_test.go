package text

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

	. "github.com/smartystreets/goconvey/convey"
)

func TestTextAlignments(t *testing.T) {
	defaultContext := context.Background()

	baseTextBox := TextboxProperties{font: "title", textColour: &colour.CNRGBA64{R: 194 << 8, G: 166 << 8, B: 73 << 8, A: 0xffff}, backgroundColour: &colour.CNRGBA64{R: 0x00f0, G: 0x00f0, B: 0x00f0, A: 0xffff}}

	xPostitions := []string{AlignmentLeft, AlignmentRight, AlignmentMiddle}
	xResults := []string{"xLeft.png", "xRight.png", "xMiddle.png"}

	xPostionBox := baseTextBox

	for i, p := range xPostitions {

		xPostionBox.xAlignment = p
		base := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		genErr := xPostionBox.DrawStrings(base, &defaultContext, []string{"sample", "text"})

		file, _ := os.Open("testdata/" + xResults[i])
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.Pix)

		Convey("Checking that strings are generated", t, func() {
			Convey(fmt.Sprintf("Generating an image with the following alignment: %v ", p), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

	yPostitions := []string{AlignmentTop, AlignmentBottom, AlignmentMiddle}
	yResults := []string{"yTop.png", "yBottom.png", "yMiddle.png"}

	yPostionBox := baseTextBox

	for i, p := range yPostitions {

		yPostionBox.yAlignment = p
		base := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		genErr := yPostionBox.DrawStrings(base, &defaultContext, []string{"sample", "text"})

		file, _ := os.Open("testdata/" + yResults[i])
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.Pix)

		Convey("Checking that strings are generated", t, func() {
			Convey(fmt.Sprintf("Generating an image with the following alignment: %v ", p), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
	// funky font from https://www.fontspace.com/super-funky-font-f100518
	//
	// https://get.fontspace.co/webfont/XqrG/OGU3MmU3NmQzZGM2NGExZmFhNDY2YTk5MzhlNWMzMjYudHRm/helloween-2.ttf
	fonts := []string{FontHeader, "./testdata/SuperFunky-testFont.ttf",
		"https://get.fontspace.co/webfont/XqrG/OGU3MmU3NmQzZGM2NGExZmFhNDY2YTk5MzhlNWMzMjYudHRm/helloween-2.ttf"}
	fontResults := []string{"builtinFont.png", "localFont.png", "webFont.png"}

	fontPostionBox := baseTextBox

	for i, f := range fonts {

		fontPostionBox.font = f
		base := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		genErr := fontPostionBox.DrawStrings(base, &defaultContext, []string{"sample", "text"})
		//	f, _ := os.Create("testdata/" + fontResults[i])
		//	png.Encode(f, base)

		file, _ := os.Open("testdata/" + fontResults[i])
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.Pix)

		Convey("Checking that strings are generated", t, func() {
			Convey(fmt.Sprintf("Generating an image with the following font: %v ", f), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}

	fillTypes := []string{FillTypeFull, FillTypeRelaxed}
	fillResults := []string{"fullFill.png", "relaxedFill.png"}

	fillBox := baseTextBox

	for i, ft := range fillTypes {

		fillBox.fillType = ft
		base := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		genErr := fillBox.DrawStrings(base, &defaultContext, []string{"sample", "text"})

		file, _ := os.Open("testdata/" + fillResults[i])
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.Pix)

		Convey("Checking that strings are generated", t, func() {
			Convey(fmt.Sprintf("Generating an image with the fill type: %v ", fillTypes), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

}

func TestVerticalText(t *testing.T) {
	baseTextBox := TextboxProperties{font: "title", textColour: &colour.CNRGBA64{R: 194 << 8, G: 166 << 8, B: 73 << 8, A: 0xffff},
		backgroundColour: &colour.CNRGBA64{R: 0x00f0, G: 0x00f0, B: 0x00f0, A: 0xffff},
		verticalText:     true}

	yPostitions := []string{AlignmentTop, AlignmentMiddle, AlignmentBottom}
	xPostitions := []string{AlignmentLeft, AlignmentMiddle, AlignmentRight}
	yResults := []string{"verticalTop.png", "verticalMiddle.png", "verticalBottom.png"}

	defaultContext := context.Background()

	yVerticalBox := baseTextBox

	for i, p := range yPostitions {

		yVerticalBox.yAlignment = p
		yVerticalBox.xAlignment = xPostitions[i]

		base := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		genErr := yVerticalBox.DrawStrings(base, &defaultContext, []string{"sample", "with space"})

		//	f, _ := os.Create("testdata/" + yResults[i])
		//	png.Encode(f, base)

		file, _ := os.Open("testdata/" + yResults[i])
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.Pix)

		Convey("Checking that strings are generated", t, func() {
			Convey(fmt.Sprintf("Generating an image with the following alignment: %v ", p), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}
