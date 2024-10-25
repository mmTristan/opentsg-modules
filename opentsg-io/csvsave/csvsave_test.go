package csvsave

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"image"
	"image/draw"
	"image/png"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCsv(t *testing.T) { // ("testing the csvs generated against a known value", func() {
	// this is the standard 16 bit image as a tiff
	file, _ := os.Open("./tests/base.png")
	// decode to get the colour values
	baseVals, _ := png.Decode(file)

	readImage := image.NewNRGBA64(baseVals.Bounds())
	// transfer to nrgba64
	draw.Draw(readImage, readImage.Bounds(), baseVals, image.Point{0, 0}, draw.Over)
	f, _ := os.Create("./tests/base.csv")
	Encode(f, readImage)

	// Test the outputs in the next section
	shaGen := func(normal, test string) (hash.Hash, hash.Hash) {
		hnormal := sha256.New()
		htest := sha256.New()
		fnormal, _ := os.ReadFile(normal)
		ftest, _ := os.ReadFile(test)
		hnormal.Write(fnormal)
		htest.Write(ftest)

		return hnormal, htest
	}

	GenCsv := []string{"./tests/base.csv"}
	NormalCsv := []string{"./tests/basetest.csv"}

	for i := range GenCsv {
		normal, test := shaGen(NormalCsv[i], GenCsv[i])
		Convey("Checking the csv files are saved and match the example file exactly", t, func() {
			Convey("using a the base image to generate the csvs", func() {
				Convey(fmt.Sprintf("The contents of %s matches %s", NormalCsv[i], GenCsv[i]), func() {
					So(test.Sum(nil), ShouldResemble, normal.Sum(nil))
				})
			})
		})
	}

	_ = os.Remove("./tests/base.csv")

}
