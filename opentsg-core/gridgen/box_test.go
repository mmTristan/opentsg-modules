package gridgen

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBox(t *testing.T) {

	// test empty and bad json and look at the output
	squareX := 100.0
	squareY := 100.0
	c := context.Background()
	cmid := context.WithValue(c, xkey, squareX)
	cmid = context.WithValue(cmid, ykey, squareY)
	cmid = context.WithValue(cmid, sizekey, image.Point{1000, 1000})
	cmid = context.WithValue(cmid, frameKey, FrameConfiguration{
		Rows: 9,
		Cols: 16,
	})
	cmid = InitAliasBox(cmid)
	cPoint := &cmid

	goodSize := []Location{

		{Alias: "test", Box: Box{X: 0, Y: 1}},
		{Box: Box{X: 0, Y: 1, X2: 2, Y2: 3}},
		{Box: Box{UseAlias: "test"}},
		{Box: Box{X: "27px", Y: "27px", X2: "53px", Y2: "53px"}},
		{Box: Box{X: 0, Y: 1, Width: 1, Height: 1}},
		{Box: Box{X: 1, Y: 1, Y2: "100%", X2: "100%"}},
		{Box: Box{X: "-27px", Y: "-27px", X2: "53px", Y2: "53px"}},
	} //, "a1:b2", "test", "(27,27)-(53,53)", "R1C02", "R2C2:R10C10", "(-27,-27)-(53,53)"}
	// alias := []string{"test", "", "", "", "", "", ""}
	expec := []image.Rectangle{image.Rect(0, 0, 100, 100), image.Rect(0, 0, 200, 200), image.Rect(0, 0, 100, 100),
		image.Rect(0, 0, 26, 26), image.Rect(0, 0, 100, 100), image.Rect(0, 0, 900, 900), image.Rect(0, 0, 80, 80)}
	expecP := []image.Point{{0, 100}, {0, 100}, {0, 100}, {27, 27}, {0, 100}, {100, 100}, {-27, -27}}
	// rows = func(context.Context) int { return 9 }
	// cols = func(context.Context) int { return 16 }

	for i, size := range goodSize {
		toCheck, pCheck, _, err := size.GridSquareLocatorAndGenerator(cPoint)
		Convey("Checking the differrent methods of string input make a map", t, func() {
			Convey(fmt.Sprintf("using a %v as the input box", size), func() {
				Convey("The generated images are the correct size", func() {
					So(err, ShouldBeNil)
					So(pCheck, ShouldResemble, expecP[i])
					So(toCheck.Bounds(), ShouldResemble, expec[i])

				})
			})
		})

	}

	// insert a tsig
	msk := image.NewNRGBA64(image.Rect(0, 0, 100, 100))
	for x := 0; x < 50; x++ {
		for y := 0; y < 100; y++ {
			msk.Set(x, y, color.RGBA64{A: 0xffff})
		}
	}
	cmid = context.WithValue(cmid, tilemaskkey, msk)
	cPoint = &cmid
	goodRadius := []string{"25px", "1", "5%"}
	for _, radius := range goodRadius {

		size := Location{Box: Box{BorderRadius: radius, X: 0, Y: 0}}
		toCheck, _, msk, err := size.GridSquareLocatorAndGenerator(cPoint)

		draw.DrawMask(toCheck, toCheck.Bounds(), &image.Uniform{color.RGBA{R: 0x91, G: 0xB6, B: 0x45, A: 0xff}}, image.Point{},
			msk, image.Point{}, draw.Src)

		//	f, _ := os.Create(fmt.Sprintf("./testdata/box/tsigRadius%v.png", radius))
		//	png.Encode(f, toCheck)

		file, _ := os.Open(fmt.Sprintf("./testdata/box/tsigRadius%v.png", radius))
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(toCheck.(*image.NRGBA64).Pix)

		Convey("Checking the rounded corners do not interfere with the tsig tiles mask", t, func() {
			Convey(fmt.Sprintf("using a border radius of %s", radius), func() {
				Convey("The generated images have a combination of the curved edge and tsig masks", func() {
					So(err, ShouldBeNil)
					So(toCheck.Bounds(), ShouldResemble, image.Rect(0, 0, 100, 100))
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}

	radiuses := []int{250, 150, 50}

	for _, r := range radiuses {
		base := image.Rect(0, 0, 500, 500)
		msk := roundedMask(cPoint, base, r)

		genRound := ImageGenerator(*cPoint, base)
		draw.DrawMask(genRound, genRound.Bounds(), &image.Uniform{color.RGBA{R: 0xC2, G: 0xA6, B: 0x49, A: 0xff}}, image.Point{},
			msk, image.Point{}, draw.Src)

		file, _ := os.Open(fmt.Sprintf("./testdata/box/%v.png", r))
		// Decode to get the colour values
		baseVals, _ := png.Decode(file)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		// Make a hash of the pixels of each image
		hnormal := sha256.New()
		htest := sha256.New()
		hnormal.Write(readImage.Pix)
		htest.Write(genRound.(*image.NRGBA64).Pix)

		// Save the file
		Convey("Checking the distances of the border radius and the shape they make", t, func() {
			Convey(fmt.Sprintf("Comparing the border radius at a length of %vpx", r), func() {
				Convey("The file matches exactly", func() {
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})

	}
}

func TestBoxTSIG(t *testing.T) {
	tpigs := "./testdata/tpig/mock.json"

	c := context.Background()
	f := FrameConfiguration{
		FrameSize: image.Point{30, 30},
		Rows:      3,
		Cols:      3,
	}

	c = context.WithValue(c, frameKey, f)
	cp := &c
	dest, _ := flatmap(cp, "./", tpigs)
	baseGen(cp, dest.canvas, f)

	splice(cp, 3, 3, 10, 10)

	gridtarget := []Location{{Box: Box{X: 0, Y: 1}}, {Box: Box{X: 0, Y: 0, Y2: 2}}, {Box: Box{X: 1, Y: 2}},
		{Box: Box{X: 0, Y: 0, X2: 3, Y2: 3}},
	} //"A1", "A0:a2", "r2c3", "R1C1:R3C3"}
	expectedSegment := [][]Segmenter{
		{{Name: "A001", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 10, Y: 10}}, Tags: []string{}, ImportPosition: 1}},
		{{Name: "A000", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 10, Y: 10}}, Tags: []string{}}, {Name: "A001", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 10}, Max: image.Point{X: 10, Y: 20}}, Tags: []string{}, ImportPosition: 1}},
		{},
		// some values are repeated across grids
		{{Name: "A000", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 10, Y: 10}}, Tags: []string{}, ImportPosition: 0},
			{Name: "A001", Shape: image.Rectangle{Min: image.Point{X: 0, Y: 10}, Max: image.Point{X: 10, Y: 20}}, Tags: []string{}, ImportPosition: 1},
			{Name: "A002", Shape: image.Rectangle{Min: image.Point{X: 10, Y: 0}, Max: image.Point{X: 25, Y: 15}}, Tags: []string{}, ImportPosition: 2},
			{Name: "A003", Shape: image.Rectangle{Min: image.Point{X: 28, Y: 0}, Max: image.Point{X: 30, Y: 30}}, Tags: []string{}, ImportPosition: 3},
			{Name: "A004", Shape: image.Rectangle{Min: image.Point{X: 20, Y: 20}, Max: image.Point{X: 30, Y: 30}}, Tags: []string{}, ImportPosition: 4}}, {}}
	for i, gt := range gridtarget {
		s, e := gt.GetGridGeometry(cp)

		Convey("Checking tpig segements are returned from the grids", t, func() {
			Convey(fmt.Sprintf("extracting the values in grid %v", gt), func() {
				Convey("An array of segemnets related to the grid positions is returned", func() {
					So(e, ShouldBeNil)
					So(s, ShouldResemble, expectedSegment[i])
				})
			})
		})
	}
}
