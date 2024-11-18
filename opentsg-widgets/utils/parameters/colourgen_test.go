package parameters

import (
	"fmt"
	"testing"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGoodHex(t *testing.T) {

	goodNoAlpha := []string{"#C06090", "#C69", "rgb(192,96,144)"}

	for i := range goodNoAlpha {
		genC := HexToColour(goodNoAlpha[i], colour.ColorSpace{})
		Convey("Checking a known string input", t, func() {
			Convey(fmt.Sprintf("using a %s as the hex colour", goodNoAlpha[i]), func() {
				Convey("A purple colour is returned, of R 192, of G 96 of B144", func() {
					So(genC, ShouldResemble, &colour.CNRGBA64{R: 49152, G: 24576, B: 36864, A: 65535})
				})
			})
		})
	}

}

func TestGoodHexAlpha(t *testing.T) {

	goodAlpha := []string{"#C06090f0", "#C69f", "rgba(192,96,144,240)", "rgb12(4095,1023,255)", "rgba12(4095,1023,255,4095)"}
	expect := []*colour.CNRGBA64{{R: 49152, G: 24576, B: 36864, A: 61440}, {R: 49152, G: 24576, B: 36864, A: 65535},
		{R: 49152, G: 24576, B: 36864, A: 61440}, {R: 65520, G: 16368, B: 4080, A: 0xffff}, {R: 65520, G: 16368, B: 4080, A: 65535}}
	for i := range goodAlpha {
		genC := HexToColour(goodAlpha[i], colour.ColorSpace{})
		Convey("Checking a known string input", t, func() {
			Convey(fmt.Sprintf("using a %s as the hex colour", goodAlpha[i]), func() {
				Convey("A purple colour is returned, a R of 192, a G of 96, a B of 144 and an A of 240", func() {
					So(genC, ShouldResemble, expect[i])
				})
			})
		})
	}
}

func TestBadHex(t *testing.T) {

	badIn := []string{"#CgA649", "realbad", "rgba(243,56,78)", "rgb(20,20,20,20)", "rgba12(20,20,20,4096)"}

	for i := range badIn {
		// these check if they somehow make it through the initial json regex that no value is returned
		genC := HexToColour(badIn[i], colour.ColorSpace{})
		var out *colour.CNRGBA64

		Convey("Checking an invalid hex code is fenced by regex", t, func() {
			Convey(fmt.Sprintf("using a %s as the hex colour", badIn[i]), func() {
				Convey("No Colour is returned as g is an invalid hex code", func() {
					So(genC, ShouldResemble, out)
				})
			})
		})
	}
}

/* everything is in CNRGBA so doesn't need to be changed or tested
func TestConvert(t *testing.T) {

	cToCheck := []*colour.CNRGBA64{{R: 100, G: 88, B: 66, A: 240}, {R: 100, G: 88, B: 66, A: 255}, {R: 194, G: 166, B: 73, A: 255}}
	expec := []*colour.CNRGBA64{{R: 25600, G: 22528, B: 16896, A: 61440}, {R: 25600, G: 22528, B: 16896, A: 65535}, {R: 49664, G: 42496, B: 18688, A: 65535}}
	// {25600, 22528, 16896, 62464}, {25600, 22528, 16896, 65535}}
	for i, c := range cToCheck {
		// these check if they somehow make it through the initial json regex that no value is returned
		genC := ConvertNRGBA64(c)
		fmt.Println(c)
		Convey("Checking rgba are converted to nrgba64", t, func() {
			Convey(fmt.Sprintf("using a %v as the input colour", c), func() {
				Convey("A converted colour is returned", func() {
					So(genC, ShouldResemble, expec[i])
				})
			})
		})
	}
}*/
