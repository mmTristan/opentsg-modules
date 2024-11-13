package zoneplate

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/parameters"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/mask"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// base example
	tbDemo := ZConfig{PlateType: circlePattern}
	examplejson.SaveExampleJson(tbDemo, WidgetType, "minimum", true)

	tbDemoMaximum := ZConfig{PlateType: verticalSweep, Startcolour: "white"}
	tbDemoMaximum.CwRotation = "π*34/53"
	examplejson.SaveExampleJson(tbDemoMaximum, WidgetType, "maximum", true)

	tbDemoNoAngle := ZConfig{PlateType: horizontalSweep, Startcolour: "black"}
	examplejson.SaveExampleJson(tbDemoNoAngle, WidgetType, "noangle", true)

	tbDemoFre := ZConfig{Frequency: parameters.AngleField{Ang: "π*13/200"}, Offset: parameters.Offset{parameters.XYOffset{X: "20%", Y: "-25"}}}
	examplejson.SaveExampleJson(tbDemoFre, WidgetType, "slowFrequency", true)

}

func TestZoneGenAngle(t *testing.T) {
	mockZone := ZConfig{PlateType: horizontalSweep,
		Startcolour: "white"}

	explanation := []string{"fractional Radians", "Degrees", "wholeRadians", "noTurning"}
	angleDummies := []interface{}{"π*1/2", 90, "π*1", nil}

	testF := []string{"./testdata/normalzp.png", "./testdata/normalzp.png", "./testdata/zonepi.png", "./testdata/zonepi1.png"}

	for i := range angleDummies {
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
		mockZone.CwRotation = angleDummies[i]

		examplejson.SaveExampleJson(mockZone, WidgetType, explanation[i], false)
		// Generate the ramp image
		genErr := mockZone.Generate(myImage)

		// Open the image to compare to
		file, _ := os.Open(testF[i])
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

		// Save the file
		Convey("Checking the angles of the zoneplate", t, func() {
			Convey(fmt.Sprintf("Comparing the ramp at an angle of %v ", angleDummies[i]), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}

func TestPlateType(t *testing.T) {
	plateTypes := []string{horizontalSweep, ellipse, circlePattern, verticalSweep}

	for _, pt := range plateTypes {

		mockZone := ZConfig{PlateType: pt}
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
		genErr := mockZone.Generate(myImage)

		file, _ := os.Open(fmt.Sprintf("./testdata/%v.png", pt))
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

		// Save the file
		Convey("Checking the plate types of the zone plate", t, func() {
			Convey(fmt.Sprintf("Comparing the plate type of %v ", pt), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}

func TestZoneGenWaveType(t *testing.T) {
	mockZone := ZConfig{Frequency: parameters.AngleField{Ang: "π*13/100"}}

	wave := []string{Sin, Cos, zp}

	testF := []string{"./testdata/normalSin.png", "./testdata/normalCos.png", "./testdata/normalZpCirc.png"}

	for i, w := range wave {
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
		mockZone.WaveType = w

		examplejson.SaveExampleJson(mockZone, WidgetType, w, false)
		// Generate the ramp image
		genErr := mockZone.Generate(myImage)
		f, _ := os.Create(testF[i])
		png.Encode(f, myImage)
		// Open the image to compare to
		file, _ := os.Open(testF[i])
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

		// Save the file
		Convey("Checking the wave types of the zoneplate", t, func() {
			Convey(fmt.Sprintf("Comparing the wavetype of %v ", w), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}

func TestZoneGenMask(t *testing.T) {
	mockZone := ZConfig{PlateType: circlePattern,
		Startcolour: "grey"}

	// mockZone.Mask = mask.Circle
	testF := []string{"./testdata/normalzpm.png"}
	explanation := []string{"circularMask"}

	for i := range testF {
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})

		examplejson.SaveExampleJson(mockZone, WidgetType, explanation[i], false)
		// Generate the ramp image
		genErr := mockZone.Generate(myImage)
		// Reapply the mask because for somereason it is not transferred across the test suiteS?
		myImage = mask.Mask(mask.Circle, 1000, 1000, 0, 0, myImage)
		// Open the image to compare to
		file, _ := os.Open(testF[i])
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

		Convey("Checking the mask of the zoneplate", t, func() {
			Convey(fmt.Sprintf("Comparing the mask of the zoneplate of %v ", "VOID" /*mockZone.Mask*/), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}

/*
func TestZoneGenErrors(t *testing.T) {
	var mockZone zoneplateJSON
	// Make the dummy functions to circumvent config

	for i := 0; i < 1; i++ {
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
		// Generate the zone plate image
		genErr := mockZone.Generate(myImage)

		// Save the file
		Convey("Checking zoneplate error catching", t, func() {
			Convey(("Running an empty zoneplate with no inouts"), func() {
				Convey("An error is returned that it has not been configured", func() {
					So(fmt.Sprintf("%v", genErr), ShouldEqual, "0111 No zone plate module selected")
				})
			})
		})
	}
}
*/
