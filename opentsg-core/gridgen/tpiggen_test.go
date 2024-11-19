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

func TestTpigGeometry(t *testing.T) {
	// get my picture size
	//// check the lines of halves and fulls
	/*	size = func(context.Context) image.Point { return image.Point{30, 30} }
		rows = func(c context.Context) int { return 3 }
		cols = func(c context.Context) int { return 3 }*/

	// repeat for the  input being a tpig and not being a tpig
	tpigs := "./testdata/tpig/mock.json"

	c := context.Background()
	f := FrameConfiguration{
		FrameSize: image.Point{30, 30},
		Rows:      3,
		Cols:      3,
	}

	c = context.WithValue(c, frameKey, f)
	cp := &c
	dest, e := flatmap(cp, "./", tpigs)
	// the contents will be cheked throughout
	Convey("Checking the tpig can be imported and read", t, func() {
		Convey(fmt.Sprintf("using a %v as the input file", tpigs), func() {
			Convey("No error is generated extracting the file", func() {
				So(e, ShouldBeNil)
				So(dest.canvas.Bounds(), ShouldResemble, image.Rect(0, 0, 30, 30))
			})
		})
	})
	canvas, e := baseGen(cp, dest.canvas, f)
	Convey("Checking the tpig context is incorporated into the base generation", t, func() {
		Convey("using the tpig context in base along with the tpig image", func() {
			Convey("No error is generated making the base image", func() {
				So(e, ShouldBeNil)
				So(canvas.Bounds(), ShouldResemble, image.Rect(0, 0, 30, 30))
			})
		})
	})

	/* loop through the different variations ensuring each method works
	 */

	splice(cp, 3, 3, 10, 10)

	gridtarget := []string{"A1", "A0:a2", "r2c3", "R1C1:R3C3"}
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
		s, e := GetGridGeometry(cp, gt)

		Convey("Checking tpig segements are returned from the grids", t, func() {
			Convey(fmt.Sprintf("extracting the values in grid %v", gt), func() {
				Convey("An array of segemnets related to the grid positions is returned", func() {
					So(e, ShouldBeNil)
					So(s, ShouldResemble, expectedSegment[i])
				})
			})
		})
	}

	// create a filler image with generic checkerboard
	filler := colour.NewNRGBA64(colour.ColorSpace{}, image.Rect(0, 0, 30, 30))
	colours := []color.Color{color.NRGBA64{R: 0xffff, A: 0xffff}, color.NRGBA64{G: 0xffff, A: 0xffff}}
	for x := 0; x < 30; x += 10 {
		for y := 0; y < 30; y += 10 {
			colour.Draw(filler, image.Rect(x, y, x+10, y+10), &image.Uniform{colours[((x+y)/10)%2]}, image.Point{}, draw.Src)
		}
	}

	// Carve the image up

	res := Carve(cp, filler, []string{"./testdata/tpig/full.png"})
	for _, v := range res {
		expectedBytes, _ := os.Open(v.Location[0])
		baseVals, _ := png.Decode(expectedBytes)

		// Assign the colour to the correct type of image NGRBA64 and replace the colour values
		readImage := image.NewNRGBA64(baseVals.Bounds())
		colour.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)

		hnormal := sha256.New()
		htest := sha256.New()

		hnormal.Write(readImage.Pix)
		switch img := v.Image.(type) {
		case *image.NRGBA64:
			htest.Write(img.Pix)
		case *colour.NRGBA64:
			htest.Write(img.Pix())
		}
		Convey("Checking the carved images match their expected tpig carving", t, func() {
			Convey(fmt.Sprintf("comparing the result to %v", v.Location[0]), func() {
				Convey("The hashes of the two images match exactly", func() {
					So(htest.Sum(nil), ShouldResemble, hnormal.Sum(nil))
				})
			})
		})
	}
	// make the image
	// call the carve
	// loop through and check the image matches the expected one already there
	// ADD carving here
}

func TestGridGeometry(t *testing.T) {
	// get my picture size
	//// check the lines of halves and fulls
	//size = func(context.Context) image.Point { return image.Point{30, 30} }
	//rows = func(c context.Context) int { return 3 }
	//cols = func(c context.Context) int { return 3 }

	// repeat for the  input being a tpig and not being a tpig
	tpigs := "./testdata/tpig/mock.json"

	c := context.Background()
	f := FrameConfiguration{
		FrameSize: image.Point{30, 30},
		Rows:      3,
		Cols:      3,
	}

	c = context.WithValue(c, frameKey, f)
	cp := &c
	dest, e := flatmap(cp, "./", tpigs)
	// the contents will be cheked throughout
	Convey("Checking the tpig can be imported and read", t, func() {
		Convey(fmt.Sprintf("using a %v as the input file", tpigs), func() {
			Convey("No error is generated extracting the file", func() {
				So(e, ShouldBeNil)
				So(dest.canvas.Bounds(), ShouldResemble, image.Rect(0, 0, 30, 30))
			})
		})
	})

	baseGen(cp, nil, f)
	splice(cp, 3, 3, 10, 10)

	fmt.Println(e)
	cd := context.Background()
	cpp := &cd
	baseGen(cpp, nil, f)
	splice(cpp, 3, 3, 10, 10)

	s, e := GetGridGeometry(cpp, "A0:A2")
	fmt.Println(s, e)

}

/*
func TestTpigGeometryHouse(t *testing.T) {
	c := context.Background()

	in, err := flatmap(&c, "./testdata/tpig/house.json")
	fmt.Println(err)

	draw.DrawMask(in.canvas, in.canvas.Bounds(), &image.Uniform{color.NRGBA64{B: 0xf0f0, A: 0xffff}}, image.Point{}, in.mask, image.Point{}, draw.Over)

	segments := c.Value(tilekey).([]Segmenter)
	fmt.Println(segments)

	textbox := texter.NewTextboxer(colour.ColorSpace{},

		texter.WithTextColour(&colour.CNRGBA64{A: 0xffff}),
	)

	for _, s := range segments {
		border := 5
		area := image.Rectangle{Min: s.Shape.Min.Add(image.Point{border, border}), Max: s.Shape.Max.Add(image.Point{-border, -border})}

		fmt.Println(area.Add(image.Point{-border - s.Shape.Min.X, -border - s.Shape.Min.Y}))
		box := image.NewNRGBA64(area.Add(image.Point{-border - s.Shape.Min.X, -border - s.Shape.Min.Y}))

		draw.Draw(box, box.Bounds(), &image.Uniform{color.NRGBA64{R: 0xf0f0, A: 0xffff}}, image.Point{}, draw.Src)
		fmt.Println(s.Name)
		err := textbox.DrawStrings(box, &c, []string{s.Name})
		fmt.Println(err)

		fmt.Println(area, area.Min)
		draw.Draw(in.canvas, area, box, image.Point{}, draw.Src)
	}

	f, _ := os.Create("./testdata/tpig/house.jpeg")
	jpeg.Encode(f, in.canvas, nil)
}
*/
