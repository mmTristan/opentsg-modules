package resize

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
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"
	"github.com/nfnt/resize"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExample(t *testing.T) {
	getPictureSize = func(c context.Context) image.Point {
		return image.Point{3840, 2160}
	}
	examples := []Config{
		{XDetections: []*parameters.DistanceField{{"3840px"}, {"1920px"}},
			YDetections: []*parameters.DistanceField{{"1080px"}, {"540px"}}},

		{XDetections: []*parameters.DistanceField{{"75%"}, {"50%"}},
			YDetections: []*parameters.DistanceField{{"50%"}, {"25%"}},
			Graticule:   graticule{Position: text.AlignmentMiddle, TextColor: "#C2A649"}},

		{XStep: &parameters.DistanceField{"500px"}, YStep: &parameters.DistanceField{"250px"}},

		{XStep: &parameters.DistanceField{"10%"}, YStep: &parameters.DistanceField{"10%"},
			XStepEnd: &parameters.DistanceField{"50%"}, YStepEnd: &parameters.DistanceField{"50%"}},

		{XStep: &parameters.DistanceField{"10%"}, XStepEnd: &parameters.DistanceField{"50%"},
			Graticule: graticule{Position: text.AlignmentLeft, TextColor: "#C2A649", GraticuleColour: "#9A3A73"}},
	}

	desc := []string{"pixel", "percentage", "stepper", "limitedStepper", "graticule"}

	for i, example := range examples {

		examplejson.SaveExampleJson(example, WidgetType, desc[i], true)
	}

}
func TestResizeDirections(t *testing.T) {

	canvasSize := []image.Point{{X: 4096, Y: 2160}, {X: 1920, Y: 1080}, {X: 1280, Y: 720}}
	destX := []string{"3840px", "1280px", "960px"}
	destXInt := []uint{3840, 1280, 960}

	for i, p := range canvasSize {
		base := image.NewNRGBA64(image.Rectangle{image.Point{}, p})

		c := context.Background()
		getPictureSize = func(c context.Context) image.Point {
			return p
		}
		genErr := Config{XDetections: []*parameters.DistanceField{{destX[i]}}}.Generate(base, &c)

		// f, _ := os.Create(fmt.Sprintf("./testdata/resize%v.png", p.X))
		// png.Encode(f, base)

		file, _ := os.Open(fmt.Sprintf("./testdata/resize%v.png", p.X))
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.Pix)

		Convey("Checking the detectors are generated", t, func() {
			Convey(fmt.Sprintf("generating an x detection of %v for a canvas of %v", destX, canvasSize[i]), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

		down := resize.Resize(destXInt[i], uint(p.Y), base, resize.Lanczos3)
		up := resize.Resize(uint(p.X), uint(p.Y), down, resize.Lanczos3)

		fl, _ := os.Create(fmt.Sprintf("./testdata/resizeLanczos%v.png", p.X))
		png.Encode(fl, up)

		downbi := resize.Resize(destXInt[i], uint(p.Y), base, resize.Bicubic)
		upbi := resize.Resize(uint(p.X), uint(p.Y), downbi, resize.Bicubic)

		fb, _ := os.Create(fmt.Sprintf("./testdata/resizeBicubic%v.png", p.X))
		png.Encode(fb, upbi)

	}

	destY := []string{"1080px"}
	destYint := []uint{1080}

	for i := range destY {
		base := image.NewNRGBA64(image.Rectangle{image.Point{}, canvasSize[i]})

		c := context.Background()
		getPictureSize = func(c context.Context) image.Point {
			return canvasSize[i]
		}
		genErr := Config{YDetections: []*parameters.DistanceField{{destY[i]}}}.Generate(base, &c)
		// f, _ := os.Create(fmt.Sprintf("./testdata/resize%v.png", canvasSize[i].Y))
		// png.Encode(f, base)

		file, _ := os.Open(fmt.Sprintf("./testdata/resize%v.png", canvasSize[i].Y))

		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.Pix)

		Convey("Checking the detectors are generated", t, func() {
			Convey(fmt.Sprintf("generating an y detection of %v for a canvas of %v", destY, canvasSize[i]), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

		down := resize.Resize(uint(canvasSize[i].X), destYint[i], base, resize.Lanczos3)
		up := resize.Resize(uint(canvasSize[i].X), uint(canvasSize[i].Y), down, resize.Lanczos3)

		fl, _ := os.Create(fmt.Sprintf("./testdata/resizeLanczos%v.png", canvasSize[i].Y))
		png.Encode(fl, up)

		downbi := resize.Resize(uint(canvasSize[i].X), destYint[i], base, resize.Bicubic)
		upbi := resize.Resize(uint(canvasSize[i].X), uint(canvasSize[i].Y), downbi, resize.Bicubic)

		fb, _ := os.Create(fmt.Sprintf("./testdata/resizeBicubic%v.png", canvasSize[i].Y))
		png.Encode(fb, upbi)

	}
}

func TestBoxes(t *testing.T) {

	boxSizes := []int{3, 6, 27, 11, 7}
	imageSize := []image.Point{
		{1920, 1080},
		{1080, 1920},
		{1920, 1080},
		{1080, 1920},
		{100, 1000},
	}

	baseImageSize := image.Point{X: 1920, Y: 1080}
	c := context.Background()

	for i, size := range boxSizes {

		getPictureSize = func(c context.Context) image.Point {
			return baseImageSize
		}

		fill := make([]*parameters.DistanceField, size)

		for i := range fill {
			fill[i] = &parameters.DistanceField{"1280px"}
		}
		myImage := image.NewNRGBA64(image.Rectangle{image.Point{}, imageSize[i]})
		genErr := Config{XDetections: fill}.Generate(myImage, &c)

		// f, _ := os.Create(fmt.Sprintf("./testdata/%vbox.png", size))
		// png.Encode(f, myImage)
		// Open the image to compare to
		file, _ := os.Open(fmt.Sprintf("./testdata/%vbox.png", size))
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
		Convey("Checking the boxes fill the space easily", t, func() {
			Convey(fmt.Sprintf("seeing how %v boxes fill a space of %v", size, imageSize[i]), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}

func TestSteps(t *testing.T) {

	steps := []parameters.DistanceField{{10}, {15}, {"100px"}, {"256px"}}
	ends := []parameters.DistanceField{{10}, {0}, {"800px"}, {"512px"}}
	graticules := []string{text.AlignmentLeft, text.AlignmentRight, text.AlignmentTop, text.AlignmentBottom}
	imageSize := image.Point{1920, 1080}

	c := context.Background()

	for i, step := range steps {

		getPictureSize = func(_ context.Context) image.Point {
			return imageSize
		}

		myImage := image.NewNRGBA64(image.Rectangle{image.Point{}, imageSize})
		genErr := Config{XStep: &step, XStepEnd: &ends[i],
			Graticule: graticule{Position: graticules[i]}}.Generate(myImage, &c)

		// f, _ := os.Create(fmt.Sprintf("./testdata/steps/step%v%s.png", step.Dist, graticules[i]))
		// png.Encode(f, myImage)
		// Open the image to compare to
		file, _ := os.Open(fmt.Sprintf("./testdata/steps/step%v%s.png", step.Dist, graticules[i]))
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
		Convey("Checking the boxes fill the space easily", t, func() {
			Convey(fmt.Sprintf("seeing how %v boxes fill a space of %v", step, imageSize), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}
