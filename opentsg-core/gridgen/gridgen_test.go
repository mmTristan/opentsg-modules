package gridgen

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

// put some inputs in and check the bounds of the image we get out

func TestGrids(t *testing.T) {

	//getWidth = func(c context.Context) float64 { return 1 }
	//size = func(c context.Context) image.Point { return image.Point{1000, 1000} }
	// imageType = func(c context.Context) string { return "" }

	// Colours
	// getFill = func(c context.Context) string { return "#ffffff" }

	f := FrameConfiguration{
		ImageSize:  image.Point{1000, 1000},
		LineWidth:  1,
		CanvasType: "",
		CanvasFill: "#ffffff",
		Rows:       3,
		Cols:       3,
	}

	sizes := []image.Point{{7, 23}, {8, 25}}
	gridType := []string{"SqueezeGrid", "UniformGrid"}

	for i, size := range sizes {

		//	rows = func(c context.Context) int { return size.Y }
		//	cols = func(c context.Context) int { return size.X }
		f.Rows = size.Y
		f.Cols = size.X
		c := context.Background()
		cPoint := &c

		base, err := GridGen(cPoint, "./", f)

		f, _ := os.Open("testdata/grids/" + gridType[i] + ".png")
		baseVals, _ := png.Decode(f)

		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Src)
		// make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(base.(*image.NRGBA64).Pix)

		Convey("Checking the grids fit on the base testcard", t, func() {
			Convey(fmt.Sprintf("using a %v, %v as the row and column counts", size.Y, size.X), func() {
				Convey("The generated images have grids that fill the test card", func() {
					So(err, ShouldBeNil)
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}

	// reset everything for use with the other tests
	/*
		rows = canvaswidget.GetGridRows
		cols = canvaswidget.GetGridColumns
		getWidth = canvaswidget.GetLWidth
		size = canvaswidget.GetPictureSize
		imageType = canvaswidget.GetCanvasType
		// Colours
		getFill = canvaswidget.GetFillColour
	*/
}

// make a test for the json init stage

func TestPtoCanvas(t *testing.T) { // test the way [{}] are read etc
	// test empty and bad json and look at the output
	squareX := 100.0
	squareY := 100.0
	c := context.Background()
	cmid := context.WithValue(c, xkey, squareX)
	cmid = context.WithValue(cmid, ykey, squareY)
	cmid = context.WithValue(cmid, sizekey, image.Point{1600, 900})
	f := FrameConfiguration{
		Rows: 9,
		Cols: 16,
	}

	cmid = context.WithValue(cmid, frameKey, f)
	cmid = PutAlias(cmid)
	cPoint := &cmid

	// check size of single area then put through the alias and we get some sizes
	goodSize := []string{"a1", "a1:b2", "test", "(27,27)-(53,53)", "R1C02", "R2C2:R10C10", "(-27,-27)-(53,53)"}
	alias := []string{"test", "", "", "", "", "", ""}
	expec := []image.Rectangle{image.Rect(0, 0, 100, 100), image.Rect(0, 0, 200, 200), image.Rect(0, 0, 100, 100),
		image.Rect(0, 0, 26, 26), image.Rect(0, 0, 100, 100), image.Rect(0, 0, 800, 800), image.Rect(0, 0, 80, 80)}
	expecP := []image.Point{{0, 100}, {0, 100}, {0, 100}, {27, 27}, {0, 100}, {100, 100}, {-27, -27}}
	//	rows = func(context.Context) int { return 9 }
	//	cols = func(context.Context) int { return 16 }

	for i, size := range goodSize {
		toCheck, pCheck, _, err := GridSquareLocatorAndGenerator(size, alias[i], cPoint)
		Convey("Checking the differrent methods of string input make a map", t, func() {
			Convey(fmt.Sprintf("using a %v as the input coordinates", size), func() {
				Convey("The generated images are the correct size", func() {
					So(err, ShouldBeNil)
					So(pCheck, ShouldResemble, expecP[i])
					So(toCheck.Bounds(), ShouldResemble, expec[i])

				})
			})
		})
	}

	badSize := []string{"a19:", "b2:a1", "fake"}
	badAlias := []string{"test", "", ""}
	badE := []string{"0046 a19: is not a valid grid alias",
		"0045 The grid dimensions of b2:a1 are invalid, received coordinates of (1,2)-(0,1)",
		"0046 fake is not a valid grid alias",
	}

	for i, size := range badSize {
		toCheck, pCheck, _, err := GridSquareLocatorAndGenerator(size, badAlias[i], cPoint)
		// _, pCheck, _, err := GridSquareLocatorAndGenerator(size, badAlias[i], cPoint)
		Convey("Checking the differrent methods of bad string input to make a map", t, func() {
			Convey(fmt.Sprintf("using a %v as the input coordinates", size), func() {
				Convey(fmt.Sprintf("An error of %v is returned as these are invalid coordinates", badE[i]), func() {
					So(toCheck, ShouldBeNil)
					So(pCheck, ShouldResemble, image.Point{})
					So(err.Error(), ShouldEqual, badE[i])
				})
			})
		})
	}

	tooLarge := []string{"a5:q6", "t6:at20"} //"(200,200)-(500,901)", "t6:at20"}
	largeErr := []string{"0047 Area outside of image bounds of (1600,900), received an x value of 1700 and a y value of 700",
		//	"0047 Area outside of image bounds of (1600,900), received an x value of 500 and a y value of 901",
		"0047 Area outside of image bounds of (1600,900), received an x value of 4600 and a y value of 2100",
	}

	for i, size := range tooLarge {
		toCheck, pCheck, _, err := GridSquareLocatorAndGenerator(size, "", cPoint)
		// _, pCheck, _, err := GridSquareLocatorAndGenerator(size, "", cPoint)
		Convey("Checking the differrent methods of bad string input to make a map", t, func() {
			Convey(fmt.Sprintf("using a %v as the input coordinates", size), func() {
				Convey(fmt.Sprintf("An error of %v is returned as these are invalid coordinates", badE[i]), func() {
					So(toCheck, ShouldBeNil)
					So(pCheck, ShouldResemble, image.Point{})
					So(err.Error(), ShouldEqual, largeErr[i])
				})
			})
		})
	}

}

// check the image and mask set ups -> problem for the last week of july

func TestGridGen(t *testing.T) {
	// get my picture size
	//// check the lines of halves and fulls
	//size = func(context.Context) image.Point { return image.Point{1600, 900} }
	widths := []float64{0.5, 1, 5}
	targets := []string{"./testdata/halfgrid.png", "./testdata/onegrid.png", "./testdata/fivegrid.png"}

	for i, w := range widths {
		// getWidth = func(context.Context) float64 { return w }
		valC := context.Background()
		myImage, _ := GridGen(&valC, "./", FrameConfiguration{
			ImageSize: image.Point{1600, 900},
			LineWidth: w,
			Rows:      9,
			Cols:      16,
		})

		f, _ := os.Open(targets[i])
		baseVals, _ := png.Decode(f)

		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Src)
		// make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(myImage.(*image.NRGBA64).Pix)

		Convey("Checking the widths of the lines are generated", t, func() {
			Convey(fmt.Sprintf("using a width of %v linewidth", w), func() {
				Convey("The hash of the generated image matches the one of the expected file", func() {
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}

}

func TestArtKey(t *testing.T) {
	filename := "./testdata/base4k.png"
	// get my picture size
	//// check the lines of halves and fulls
	// check the base is the same as the 4k image that goes in as an init
	// check the points and image sizes are correct using art param afterwards
	// 	size = func(context.Context) image.Point { return image.Point{1600, 900} }
	//	widths := []float64{0.5, 1, 5}
	targets := []image.Point{{3840, 2160}, {1920, 1080}, {7680, 4320}}
	base := []string{"./testdata/base4k.png", "./testdata/hdresize.png", "./testdata/resize8k.png"}
	// rows = func(context.Context) int { return 9 }
	// cols = func(context.Context) int { return 16 }

	expectedPoints := []image.Point{{156, 141}, {78, 70}, {312, 282}}
	expectedBounds := []image.Rectangle{image.Rect(0, 0, 1129, 545), image.Rect(0, 0, 565, 273), image.Rect(0, 0, 2258, 1090)}
	fmt.Println(targets)
	for i, tar := range targets {
		fmt.Println(i, tar)
		//	size = func(context.Context) image.Point { return tar }
		f, _ := os.Open("./testdata/base4k.png")
		baseVals, _ := png.Decode(f)
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Src)
		// getWidth = func() float64 { return w }
		fmt.Println(readImage.At(178, 1240))
		valC := context.Background()
		var mockgeom draw.Image
		myImage, _ := artKeyGen(&valC, mockgeom, filename,
			FrameConfiguration{
				ImageSize: tar,
				Rows:      9,
				Cols:      16,
			})

		// make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()

		fb, _ := os.Open(base[i])
		basetest, _ := png.Decode(fb)
		testImage := image.NewNRGBA64(basetest.Bounds())
		colour.Draw(testImage, testImage.Bounds(), basetest, image.Point{0, 0}, draw.Src)
		hnormal.Write(testImage.Pix)
		htest.Write(myImage.(*image.NRGBA64).Pix)

		Convey("Checking the background image is scaled", t, func() {
			Convey(fmt.Sprintf("using %s as the base image", filename), func() {
				Convey("The hash of the generated image matches the one of the expected file", func() {
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

		add, loc, _, _ := artToCanvas("key:red", &valC)
		// test the bound results here as part of the loop

		Convey("Checking the keys are located", t, func() {
			Convey(fmt.Sprintf("using %s as the base image and searching for \"key:red\"", filename), func() {
				Convey("The size and location of the file match the expected one", func() {
					So(loc, ShouldResemble, expectedPoints[i])
					So(add.Bounds(), ShouldResemble, expectedBounds[i])
				})
			})
		})
	}

}
