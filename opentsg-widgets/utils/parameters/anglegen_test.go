package parameters

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/yaml.v3"
)

func TestGoodInputs(t *testing.T) {
	goodStrings := []RotationAngle{{"π*2/3"}, {"π*1"}}
	ExpectedRadianResult := []float64{2.0943951023931953, 3.141592653589793}

	for i, angle := range goodStrings {

		res, err := angle.ClockwiseRotationAngle()

		Convey("Checking radians are calculated correctly", t, func() {
			Convey(fmt.Sprintf("using an angle of %v", angle), func() {
				Convey(fmt.Sprintf("No error is returned and the calculated angle is %v", ExpectedRadianResult[i]), func() {
					So(err, ShouldBeNil)
					So(res, ShouldResemble, ExpectedRadianResult[i])
				})
			})
		})
	}

	goodDegree := []RotationAngle{{"45"}, {"180"}}
	ExpectedDegreeResult := []float64{0.7853981633974483, 3.141592653589793}

	for i, angle := range goodDegree {

		res, err := angle.ClockwiseRotationAngle()

		Convey("Checking radians are calculated correctly", t, func() {
			Convey(fmt.Sprintf("using an angle of %v", angle), func() {
				Convey(fmt.Sprintf("No error is returned and the calculated angle is %v", ExpectedDegreeResult[i]), func() {
					So(err, ShouldBeNil)
					So(res, ShouldResemble, ExpectedDegreeResult[i])
				})
			})
		})
	}

}

func TestBadInputs(t *testing.T) {
	goodDegree := []RotationAngle{{"4.5.0"}, {"thirtyfive"}}

	for _, angle := range goodDegree {

		res, err := angle.ClockwiseRotationAngle()

		Convey("Checking incorrect strings are caught", t, func() {
			Convey(fmt.Sprintf("using an angle of %v", angle), func() {
				Convey("An error is returned highlighting the angle in question", func() {
					So(err, ShouldResemble, fmt.Errorf("%s is not a valid angle", angle.CwRotation))
					So(res, ShouldEqual, 0)
				})
			})
		})
	}

}

func TestParameterMarshall(t *testing.T) {
	type testStruct struct {
		Angle *AngleField    `json:"angle,omitempty" yaml:"angle,omitempty"`
		Dist  *DistanceField `json:"distance,omitempty" yaml:"distance,omitempty"`
		Arr   []*AngleField  `json:"array,omitempty" yaml:"array,omitempty"`
	}

	destfile := []string{"testdata/angle", "testdata/distance", "testdata/array"}
	expectedField := []testStruct{
		{Angle: &AngleField{Ang: "π*23/47"}},
		{Dist: &DistanceField{Dist: "25px"}},
		{Arr: []*AngleField{{Ang: 42.01}, {Ang: 53.01}}},
	}

	for i, dest := range destfile {

		b, err := os.ReadFile(dest + ".json")
		var tsj testStruct
		jErr := json.Unmarshal(b, &tsj)

		Convey("Checking that the fields are correctly unmarshalled with json", t, func() {
			Convey(fmt.Sprintf("unmarshalling a json of %s", string(b)), func() {
				Convey("No error is returned and the file is correctly marshalled to the struct", func() {
					So(err, ShouldBeNil)
					So(jErr, ShouldBeNil)
					So(tsj, ShouldResemble, expectedField[i])
				})
			})
		})

		b, err = os.ReadFile(dest + ".yaml")
		var tsy testStruct
		yErr := yaml.Unmarshal(b, &tsy)

		Convey("Checking that the fields are correctly unmarshalled with yaml", t, func() {
			Convey(fmt.Sprintf("unmarshalling a yaml of %s", string(b)), func() {
				Convey("No error is returned and the file is correctly marshalled to the struct", func() {
					So(err, ShouldBeNil)
					So(yErr, ShouldBeNil)
					So(tsy, ShouldResemble, expectedField[i])
				})
			})
		})

	}

	for i, ef := range expectedField {

		jb, jErr := json.MarshalIndent(ef, "", "    ")

		fjb, fjErr := os.ReadFile(destfile[i] + ".json")

		Convey("Checking that the fields are correctly marshalled with json", t, func() {
			Convey(fmt.Sprintf("marshalling a stuct of %v", ef), func() {
				Convey("No error is returned and the struct is correctly marshalled to the json", func() {
					So(jErr, ShouldBeNil)
					So(fjErr, ShouldBeNil)
					So(string(fjb), ShouldResemble, string(jb))
				})
			})
		})

		yb, yErr := yaml.Marshal(ef)

		fyb, fyErr := os.ReadFile(destfile[i] + ".yaml")

		Convey("Checking that the fields are correctly marshalled with json", t, func() {
			Convey(fmt.Sprintf("marshalling a stuct of %v", ef), func() {
				Convey("No error is returned and the struct is correctly marshalled to the json", func() {
					So(yErr, ShouldBeNil)
					So(fyErr, ShouldBeNil)
					So(string(fyb), ShouldResemble, string(yb))
				})
			})
		})

	}

}
