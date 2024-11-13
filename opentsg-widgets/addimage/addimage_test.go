package addimage

// Run through the file fences once these are made using example 18 and 8 bit versions of then

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"runtime/debug"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	"github.com/mrmxf/opentsg-modules/opentsg-core/parameters"
	examplejson "github.com/mrmxf/opentsg-modules/opentsg-widgets/exampleJson"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDemo(t *testing.T) {
	// minimum example
	ai := Config{Image: "https://opentsg.studio/blog/2023/09/13/2023-09-13-coming-soon/featured-logo-otsg.png"}
	examplejson.SaveExampleJson(ai, WidgetType, "minimum", true)

	// maximum example
	aiMax := Config{Image: "https://opentsg.studio/blog/2023/09/13/2023-09-13-coming-soon/featured-logo-otsg.png", ImgFill: "y scale"}
	examplejson.SaveExampleJson(aiMax, WidgetType, "maximum", true)
}

func TestBadStrings(t *testing.T) {

	mockContext := context.Background()

	// make the test suite translatable to different os and hosts that are not gitpod
	location := os.Getenv("PWD")
	sep := string(os.PathSeparator)
	// Slap some text files here
	badString := []string{"./testdata/badfile.txt", "", "./testdata/bad.dpx"}

	badStrErr := []string{
		fmt.Sprintf("0163 %s%stestdata%sbadfile.txt is an invalid file type", location, sep, sep),
		"0161 No image declared",
		fmt.Sprintf("0163 %s%stestdata%sbad.dpx is an invalid file type", location, sep, sep)}

	for i, bad := range badString {
		mockImg := Config{Image: bad}
		genErr := mockImg.Generate(nil, &mockContext)
		Convey("Checking the regex fence is working", t, func() {
			Convey(fmt.Sprintf("using a %s as the file to open %s", bad, location), func() {
				Convey("An error is returned as the file in invalid", func() {
					So(genErr.Error(), ShouldResemble, badStrErr[i])

				})
			})
		})
	}

}

func Test16files(t *testing.T) {

	goodString := []string{"./testdata/test16bit.tiff", "./testdata/test16bit.png"}

	for _, name := range goodString {
		tfile, _ := os.Open(name)
		_, _, genErr := fToImg(tfile, name)

		Convey("Checking that 16 bit files get through the fence", t, func() {
			Convey(fmt.Sprintf("using a %s as the file to open", name), func() {
				Convey("No error is returned as the files are 16 bit", func() {
					So(genErr, ShouldBeNil)

				})
			})
		})
	}
}

// 8 bit files are now allowed through
func Test8files(t *testing.T) {
	mockContext := context.Background()
	good8String := []string{"./testdata/test8bit.png", "./testdata/test8bit.tiff", "./testdata/squares.png"}

	for _, name := range good8String {
		tfile, _ := os.Open(name)
		_, _, genErr := fToImg(tfile, name)

		canvas := image.NewNRGBA64(image.Rect(0, 0, 1000, 1000))
		mockImg := Config{Image: name}
		_ = mockImg.Generate(canvas, &mockContext)

		//	f, _ := os.Create(fmt.Sprintf("file%v.png", i))
		//	png.Encode(f, canvas)

		Convey("Checking that 8 bit files are filtered through the fence", t, func() {
			Convey(fmt.Sprintf("using a %s as the file to open", name), func() {
				Convey("An error is returned stating the file is not the correct bit depth", func() {
					So(genErr, ShouldBeNil) // , fmt.Errorf("0166 %s colour depth is %v bits not 16 bits. Only 16 bit files are accepted", name, 8))

				})
			})
		})
	}
}

func TestWebsites(t *testing.T) {
	mockContext := context.Background()
	validSite := []string{"https://opentsg.studio/blog/2023/09/13/2023-09-13-coming-soon/featured-logo-otsg.png"}
	expec := []string{"1c9e781fb1ac14c8b292f8fcf68d95a00dd0ee96f1443627e322b4fe7ad9e809"}
	for i, imgToAdd := range validSite {
		ai := Config{Image: imgToAdd}
		// Ai.Image = imgToAdd
		genImg := image.NewNRGBA64(image.Rect(0, 0, 4096, 2160))
		genErr := ai.Generate(genImg, &mockContext)
		htest := sha256.New()
		htest.Write(genImg.Pix)

		//	f, _ := os.Create(fmt.Sprintf("file%v.png", i))
		//	png.Encode(f, genImg)

		Convey("Checking that images sourced from http can generate images", t, func() {
			Convey(fmt.Sprintf("using a %s as the file to open", imgToAdd), func() {
				Convey("No error is returned as the image is of a correct type", func() {
					So(genErr, ShouldBeNil)
					So(fmt.Sprintf("%x", htest.Sum(nil)), ShouldResemble, expec[i])
				})
			})
		})
	}
}

func TestZoneGenMask(t *testing.T) {
	mockContext := context.Background()
	bi, _ := debug.ReadBuildInfo()
	// Keep this in the background unti it runs
	if bi.GoVersion[:6] != "go1.18" {
		var imgMock Config
		var pos config.Position
		pos.X = 0
		pos.Y = 0
		// imgMock.Imgpos = &pos

		sizeDummies := []image.Point{{1000, 1000}, {1000, 500}}

		testF := []string{"../zoneplate/testdata/normalzpm.png", "./testdata/redrawnzp.png"}
		explanation := []string{"mask", "maskResize"}

		for i := range sizeDummies {
			imgMock.Image = testF[0]

			myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, sizeDummies[i]})

			// generate the ramp image
			genErr := imgMock.Generate(myImage, &mockContext)
			examplejson.SaveExampleJson(imgMock, WidgetType, explanation[i], false)

			file, _ := os.Open(testF[i])
			defer file.Close()
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

			// Save the file
			// f, _ := os.Create(testF[i] + ".png")
			// colour.PngEncode(f, myImage)

			Convey("Checking the size of the squeezed zoneplate to fill the canvas", t, func() {
				Convey(fmt.Sprintf("Adding the image to a blank canvas the size of %v", sizeDummies[i]), func() {
					Convey("No error is returned and the file matches exactly", func() {
						So(genErr, ShouldBeNil)
						So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
					})
				})
			})

		}
	}
}

func TestFillTypes(t *testing.T) {
	mockContext := context.Background()

	var imgMock Config

	// imgMock.Imgpos = &c

	//	sizeDummies := [][]int{{0, 0}, {1000, 500}}

	testFile := "./testdata/test16bit.png"
	fillTypes := []string{"x scale", "y scale", "xy scale", "fill", "preserve"}
	explanation := []string{"xScale", "yScale", "xyScale", "fill", "preserve"}

	for i, fill := range fillTypes {
		imgMock.Image = testFile
		imgMock.ImgFill = fill

		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 900}})
		genErr := imgMock.Generate(myImage, &mockContext)

		examplejson.SaveExampleJson(imgMock, WidgetType, explanation[i], false)
		// Open the image to compare to

		file, _ := os.Open(fmt.Sprintf("./testdata/fill%v.png", i))
		defer file.Close()
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)
		readImage := image.NewNRGBA64(baseVals.Bounds())

		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Src)
		// Decode to get the colour values
		_ = png.Encode(file, myImage)
		// Save the file
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(myImage.Pix)
		compare(myImage, readImage)

		//	f, _ := os.Create(fmt.Sprintf("./testdata/fill%v.png", i) + ".png")
		//	colour.PngEncode(f, myImage)

		Convey("Checking the different fill methods of addimage", t, func() {
			Convey(fmt.Sprintf("Adding the image to a blank canvas and using the fill type of %s", fill), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}

	/*
		imgMock = addimageJSON{Image: "./testdata/test16bit.png", ColourSpace: colour.ColorSpace{ColorSpace: "rec709"}}
		base := colour.NewNRGBA64(colour.ColorSpace{ColorSpace: "rec2020"}, image.Rect(0, 0, 1000, 1000))
		cb := context.Background()
		fmt.Println(imgMock.Generate(base, &cb))

		f, _ := os.Create("test709.png")
		png.Encode(f, base) */
}

func TestOffsets(t *testing.T) {
	mockContext := context.Background()

	var imgMock = Config{Image: "./testdata/squares.png", ImgFill: "fill"}
	// imgMock.Imgpos = &c

	//	sizeDummies := [][]int{{0, 0}, {1000, 500}}

	offsets := []parameters.Offset{{Offset: parameters.XYOffset{X: "25px", Y: "100px"}}, {Offset: parameters.XYOffset{X: -20, Y: -20}}}
	explanation := []string{"forward_offset", "backwards_offset"}

	for i, off := range offsets {
		imgMock.Offset = off

		myImage := image.NewNRGBA64(image.Rectangle{image.Point{0, 0}, image.Point{1000, 1000}})
		genErr := imgMock.Generate(myImage, &mockContext)

		examplejson.SaveExampleJson(imgMock, WidgetType, explanation[i], true)
		// Open the image to compare to

		file, _ := os.Open(fmt.Sprintf("./testdata/offset%v.png", i))
		defer file.Close()
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)
		readImage := image.NewNRGBA64(baseVals.Bounds())

		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Src)
		// Decode to get the colour values
		_ = png.Encode(file, myImage)
		// Save the file
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(myImage.Pix)
		compare(myImage, readImage)

		// f, _ := os.Create(fmt.Sprintf("./testdata/offseti%v.png", i))
		// colour.PngEncode(f, myImage)

		Convey("Checking the offset methods of addimage", t, func() {
			Convey(fmt.Sprintf("Adding the image to a blank canvas and using an offset of %v", off), func() {
				Convey("No error is returned and the file matches exactly", func() {
					So(genErr, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}
}

func compare(base, new draw.Image) {

	count := 0
	b := base.Bounds().Max
	for x := 0; x < b.X; x++ {
		for y := 0; y < b.Y; y++ {
			if base.At(x, y) != new.At(x, y) {
				count++
				// fmt.Println(x, y, base.At(x, y), new.At(x, y))
			}

		}

	}

	fmt.Println(count, "non matches")
}
