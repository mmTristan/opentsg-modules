package qrgen

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/boombuler/barcode/qr"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/parameters"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	qrDemo := Config{Code: "https://opentsg.studio/"}
	examplejson.SaveExampleJson(qrDemo, WidgetType, "minimum", true)

	qrDemoMax := Config{Code: "https://opentsg.studio/", Size: &sizeJSON{Width: 100, Height: 100}}
	examplejson.SaveExampleJson(qrDemoMax, WidgetType, "maximum", true)

	qrDemoMiddle := Config{Code: "https://opentsg.studio/", Size: &sizeJSON{Width: 50, Height: 50}}
	qrDemoMiddle.Offset = parameters.Offset{Offset: parameters.XYOffset{X: 50, Y: 50}}
	examplejson.SaveExampleJson(qrDemoMiddle, WidgetType, "middlepic", true)
}

func TestQrGen(t *testing.T) {
	// Run this so the qr code is not placed, placed in the middle and bottom right
	var qrmock Config

	numberToCheck := [][]float64{{0, 0}, {50, 50}, {97.1, 97.1}}
	fileCheck := []string{"./testdata/topleft.png", "./testdata/middle.png", "./testdata/bottomright.png"}
	explanation := []string{"topleft", "middle", "topright"}
	qrmock.Code = "https://mrmxf.io/"
	code, _ := qr.Encode("https://mrmxf.io/", qr.H, qr.Auto)
	fmt.Println(code.Bounds())

	for i, num := range numberToCheck {
		// Get file to place the qr code on
		file, _ := os.Open("./testdata/zonepi.png")
		baseVals, _ := png.Decode(file)
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{}, draw.Over)
		// Get the image to compare against
		fileCont, _ := os.Open(fileCheck[i])
		baseCont, _ := png.Decode(fileCont)
		control := image.NewNRGBA64(baseCont.Bounds())
		colour.Draw(control, control.Bounds(), baseCont, image.Point{}, draw.Over)
		// Generate the image and the string

		qrmock.Offset = parameters.Offset{Offset: parameters.XYOffset{X: num[0], Y: num[1]}}

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		c := context.Background()
		genErr := qrmock.Generate(readImage, &c)
		examplejson.SaveExampleJson(qrmock, WidgetType, explanation[i], false)
		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(control.Pix)
		htest.Write(readImage.Pix)

		// GenResult, genErr := intTo4(numberToCheck[i])
		Convey("Checking the qr code is added to an image is generated", t, func() {
			Convey(fmt.Sprintf("using a location of x:%v, y:%v  as integer ", numberToCheck[i][0], numberToCheck[i][1]), func() {
				Convey("A qr code is added and the generated sha256 is identical", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

	qrmock.Offset = parameters.Offset{}
	max := sizeJSON{Width: 100, Height: 100}
	qrmock.Size = &max

	base := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
	c := context.Background()
	genErr := qrmock.Generate(base, &c)
	examplejson.SaveExampleJson(qrmock, WidgetType, "full", false)

	file, _ := os.Open("./testdata/full.png")
	baseVals, _ := png.Decode(file)
	readImage := image.NewNRGBA64(baseVals.Bounds())
	colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{}, draw.Over)

	hnormal := sha256.New()
	htest := sha256.New()
	hnormal.Write(readImage.Pix)
	htest.Write(base.Pix)

	Convey("Checking the qr code is added to fill a space", t, func() {
		Convey(fmt.Sprintf("using a size of width:%v, height:%v  as integer ", 100, 100), func() {
			Convey("A qr code is added and the generated sha256 is identical", func() {
				So(genErr, ShouldBeNil)
				So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
			})
		})
	})
}

func TestErr(t *testing.T) {
	var qrmock Config

	// Run this so the qr code is not placed, placed in the middle and bottom right
	numberToCheck := [][]float64{{100, 0}, {98, 100}, {0, 100}, {40, 80}, {0, 0}, {0, 0}}
	numberToResize := [][]float64{{0, 0}, {0, 0}, {0, 0}, {0, 0}, {2, 2}, {1, 4}}
	expecErr := []string{"0133 the x position 100 is greater than the x boundary of 100",
		"0133 the x position 98 is greater than the x boundary of 100",
		"0133 the y position 100 is greater than the y boundary of 100",
		"0133 the y position 80 is greater than the y boundary of 100",
		"0132 can not scale barcode to an image smaller than 29x29",
		"0132 can not scale barcode to an image smaller than 29x29"}
	qrmock.Code = "https://mrmxf.io/"
	code, _ := qr.Encode("https://mrmxf.io/", qr.H, qr.Auto)
	fmt.Println(code.Bounds())

	for i, check := range numberToCheck {

		dummy := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{100, 100}})

		// Generate the image and the string
		qrmock.Offset = parameters.Offset{Offset: parameters.XYOffset{X: check[0], Y: check[1]}}

		var s sizeJSON
		s.Width = numberToResize[i][0]
		s.Height = numberToResize[i][1]
		qrmock.Size = &s
		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		c := context.Background()
		genErr := qrmock.Generate(dummy, &c)
		fmt.Println(genErr)
		// GenResult, genErr := intTo4(numberToCheck[i])
		Convey("Checking that x and y errors are caught", t, func() {
			Convey(fmt.Sprintf("using a location of x:%v, y:%v  as integer and a resize of x:%v, y:%v ", check[0], check[1], numberToResize[i][0], numberToResize[i][1]), func() {
				Convey("A qr code is added and the generated sha256 is identical", func() {
					So(genErr.Error(), ShouldEqual, expecErr[i])
				})
			})
		})
	}
}

func TestQrResize(t *testing.T) {
	// Run this so the qr code is not placed, placed in the middle and bottom right
	var qrmock Config

	numberToCheck := [][]float64{{58, 58}, {100, 100}, {75, 75}}
	// FileCheck := []string{"./testdata/topleftr.png", "./testdata/middler.png", "./testdata/bottomrightr.png"}
	qrmock.Code = "https://mrmxf.io/"

	// Just check the error sizes are passed through
	for _, check := range numberToCheck {
		mock := image.NewNRGBA64(image.Rect(0, 0, 200, 200))
		var s sizeJSON
		s.Width = check[0]
		s.Height = check[1]
		qrmock.Size = &s
		fmt.Println(qrmock)
		c := context.Background()
		genErr := qrmock.Generate(mock, &c)

		Convey("Checking that the qr code can be resized", t, func() {
			Convey(fmt.Sprintf("using a resize value of x:%v, y:%v  as integer ", check[0], check[1]), func() {
				Convey("A qr code is resized and no error is returneds", func() {
					So(genErr, ShouldBeNil)
				})
			})
		})

	}
}
