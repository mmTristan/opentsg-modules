package gridgen

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
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
	cmid = core.PutAliasBox(cmid)
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
	rows = func(context.Context) int { return 9 }
	cols = func(context.Context) int { return 16 }
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

	radiuses := []int{250, 150, 50}

	for _, r := range radiuses {
		base := image.Rect(0, 0, 500, 500)
		msk := roundedMask(cPoint, base, r)

		img := ImageGenerator(*cPoint, base)
		draw.DrawMask(img, img.Bounds(), &image.Uniform{color.RGBA{R: 0xC2, G: 0xA6, B: 0x49, A: 0xff}}, image.Point{},
			msk, image.Point{}, draw.Src)

		f, _ := os.Create(fmt.Sprintf("./testdata/box/%v.png", r))
		png.Encode(f, img)

	}
}
