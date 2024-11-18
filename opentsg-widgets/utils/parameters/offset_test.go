package parameters

import (
	"fmt"
	"image"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOffset(t *testing.T) {
	goodStrings := []Offset{{Offset: XYOffset{X: 5, Y: 15}}, {Offset: XYOffset{X: "75.3%", Y: "23"}},
		{Offset: XYOffset{X: "1px", Y: "1001px"}}, {Offset: XYOffset{X: "-1px", Y: "-1001px"}}}
	ExpectedRadianResult := [][2]float64{{5, 15}, {75, 23}, {1, 1001}, {-1, -1001}}

	for i, off := range goodStrings {

		res, err := off.CalcOffset(image.Point{100, 100})

		Convey("Checking radians are calculated correctly", t, func() {
			Convey(fmt.Sprintf("using an angle of %v", off), func() {
				Convey(fmt.Sprintf("No error is returned and the calculated angle is %v", ExpectedRadianResult[i]), func() {
					So(err, ShouldBeNil)
					So(res.X, ShouldResemble, ExpectedRadianResult[i][0])
					So(res.Y, ShouldResemble, ExpectedRadianResult[i][1])
				})
			})
		})
	}
}
