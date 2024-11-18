package bowtie

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
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBowties(t *testing.T) {

	simple := Config{SegementCount: 8}
	corners := Config{SegementCount: 12}
	colours := Config{SegementCount: 12, SegmentColours: []parameters.HexString{"#C2A649", "#9A3A73", "#91B645", "#433F87"}} //, "#433F87"}} //, "#433F87"}}
	all := Config{SegementCount: 32, SegmentColours: []parameters.HexString{"#C2A649", "#9A3A73", "#91B645", "#433F87"}}
	all.CwRotation = "π*23/47"

	explanation := []string{"8Segment", "12Segment", "colourSegments", "32Segments"}
	testF := []string{"./testdata/swirl0.png", "./testdata/swirl1.png", "./testdata/swirl2.png", "./testdata/swirl3.png"}

	bowties := []Config{simple, corners, colours, all} //, all}

	for i, s := range bowties {

		cb := context.Background()
		img := image.NewNRGBA64(image.Rect(0, 0, 200, 160))

		examplejson.SaveExampleJson(s, WidgetType, explanation[i], true)

		genErr := s.Generate(img, &cb)

		//	f, _ := os.Create(fmt.Sprintf("./testdata/swirl%v.png", i))
		//	png.Encode(f, img)

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
		htest.Write(img.Pix)

		Convey("Generating bowties, comparing them with the defaults", t, func() {
			Convey(fmt.Sprintf("Making the bowties with the parameters %v ", s), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
}

func TestOffsets(t *testing.T) {

	all := Config{SegementCount: 32, SegmentColours: []parameters.HexString{"#C2A649", "#9A3A73", "#91B645", "#433F87"}}

	left := parameters.Offset{Offset: parameters.XYOffset{X: "-50"}}
	offRight := parameters.Offset{Offset: parameters.XYOffset{X: "50px", Y: "-70"}}
	offUp := parameters.Offset{Offset: parameters.XYOffset{Y: 20}}

	explanation := []string{"offsetXLeft", "offsetXAndY", "offSetY"}

	offsets := []parameters.Offset{left, offRight, offUp} //, all}

	for i, off := range offsets {

		cb := context.Background()
		img := image.NewNRGBA64(image.Rect(0, 0, 200, 160))

		//examplejson.SaveExampleJson(s, widgetType, explanation[i], true)
		all.Offset = off
		genErr := all.Generate(img, &cb)
		examplejson.SaveExampleJson(all, WidgetType, explanation[i], true)

		file, _ := os.Open(fmt.Sprintf("./testdata/offset%v.png", i))
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(img.Pix)

		Convey("Generating bowties, and offsetting the origin", t, func() {
			Convey(fmt.Sprintf("Offsetting the bowties with the parameters %v ", off), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

		//f, _ := os.Create(fmt.Sprintf("./testdata/offset%v.png", i))
		//png.Encode(f, img)
	}
}

func TestBlends(t *testing.T) {

	simple := Config{SegementCount: 4}
	corners := Config{SegementCount: 8}
	colours := Config{SegementCount: 8, SegmentColours: []parameters.HexString{"#C2A649", "#9A3A73", "#91B645", "#433F87"}}

	explanation := []string{"SinBowtie", "Sin8Segment", "SinColours"}

	sins := []Config{simple, corners, colours}

	for i, s := range sins {
		s.Blend = "sin"
		cb := context.Background()
		img := image.NewNRGBA64(image.Rect(0, 0, 200, 160))

		genErr := s.Generate(img, &cb)
		// f, _ := os.Create(fmt.Sprintf("./testdata/blendSin%v.png", i))
		// png.Encode(f, img)
		examplejson.SaveExampleJson(s, WidgetType, explanation[i], true)

		file, _ := os.Open(fmt.Sprintf("./testdata/blendSin%v.png", i))
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(img.Pix)

		Convey("Generating bowties with blends", t, func() {
			Convey(fmt.Sprintf("Generating a bow tie with the following properties %v", s), func() {
				Convey("No error is returned and a blended bowtie is generated", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}

}

func TestErrors(t *testing.T) {
	simple := Config{SegementCount: 3}
	badAng := Config{SegementCount: 300}
	badAng.CwRotation = "math.Pi"

	bowties := []Config{simple, badAng} //, all}
	errs := []string{"0DEV 4 or more segments required, received 3", "0DEV error calculating the rotational angle math.Pi is not a valid angle"}

	for i, s := range bowties {

		cb := context.Background()
		img := image.NewNRGBA64(image.Rect(0, 0, 200, 160))
		genErr := s.Generate(img, &cb)

		Convey("Generating bowties that deliberately generate errors", t, func() {
			Convey(fmt.Sprintf("Generating the error %v ", errs[i]), func() {
				Convey("The error is correctly generated", func() {
					So(genErr, ShouldResemble, fmt.Errorf(errs[i]))

				})
			})
		})
	}
}

func TestRotate(t *testing.T) {

	simple := Config{SegementCount: 4, Blend: "sin"}
	all := Config{SegementCount: 32, SegmentColours: []parameters.HexString{"#C2A649", "#9A3A73", "#91B645", "#433F87"}}
	allOff := Config{SegementCount: 32, SegmentColours: []parameters.HexString{"#C2A649", "#9A3A73", "#91B645", "#433F87"}}
	allOff.Offset = parameters.Offset{Offset: parameters.XYOffset{X: "50px", Y: "-70"}}
	startAng := Config{SegementCount: 4, Blend: "sin"}
	startAng.StartAng = 180

	rotates := []Config{simple, all, allOff, startAng}

	for i, rot := range rotates {

		rot.CwRotation = "π*15/407"
		cb := context.Background()
		img := image.NewNRGBA64(image.Rect(0, 0, 200, 160))

		framePos = func(_ context.Context) int {

			return i + 17
		}

		genErr := rot.Generate(img, &cb)
		//	f, _ := os.Create(fmt.Sprintf("./testdata/rotate%v.png", i))
		//	png.Encode(f, img)

		file, _ := os.Open(fmt.Sprintf("./testdata/rotate%v.png", i))
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(img.Pix)

		Convey("Generating bowties and mocking the rotation", t, func() {
			Convey(fmt.Sprintf("Generating a bow tie that has the following properties %v on frame %v", rot, framePos(nil)), func() {
				Convey("No error is returned and the bowtie has rotated", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}

}
