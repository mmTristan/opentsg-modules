package parameters

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"regexp"
	"strconv"
)

/*
DistanceField is a struct that wraps an any.
It is to be used as a type any for calculating the angle of
anything that is not got with parameters.

e.g.

	type testStruct struct {
		MyNewOffset DistanceField `json:"DirectionOffset"`
	}

This can then be parsed as a first level object in json or yaml.

Encoding is also provided
*/
type DistanceField struct {
	Dist any `yaml:",flow"`
}

func (d *DistanceField) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var dist any
	err := unmarshal(&dist)
	if err != nil {
		return err
	}

	d.Dist = dist
	return nil
}

func (d DistanceField) MarshalYAML() (interface{}, error) {
	return d.Dist, nil
}

func (d *DistanceField) UnmarshalJSON(data []byte) error {
	var dist any
	err := json.Unmarshal(data, &dist)

	d.Dist = dist
	return err
}

func (d DistanceField) MarshalJSON() ([]byte, error) {

	return json.Marshal(d.Dist)
}

func (d DistanceField) CalcOffset(bound int) (int, error) {

	return offsetToPixels(fmt.Sprintf("%v", d.Dist), bound)
}

// Offset is a string for allowing multiple input formats
// that can then be parsed to radians.
type Offset struct {
	Offset XYOffset `json:"offset,omitempty" yaml:"offset,omitempty"`
}

type XYOffset struct {
	X any `json:"x,omitempty" yaml:"x,omitempty"`
	Y any `json:"y,omitempty" yaml:"y,omitempty"`
}

func (xy Offset) CalcOffset(max image.Point) (offPoint image.Point, err error) {

	if xy.Offset.X != nil {
		offPoint.X, err = offsetToPixels(fmt.Sprintf("%v", xy.Offset.X), max.X)

		if err != nil {
			return
		}
	}

	if xy.Offset.Y != nil {
		offPoint.Y, err = offsetToPixels(fmt.Sprintf("%v", xy.Offset.Y), max.Y)

		if err != nil {
			return
		}
	}

	return
}

func offsetToPixels(val string, max int) (int, error) {

	pixel := regexp.MustCompile(`^-{0,1}\d{1,}px$`)
	percent := regexp.MustCompile(`^-{0,1}\d{0,2}\.{1}\d{0,}$|^-{0,1}\d{0,2}$|^-{0,1}(100)$`)
	pcDefault := regexp.MustCompile(`^-{0,1}\d{0,2}\.{1}\d{0,}%$|^-{0,1}\d{0,2}%$|^-{0,1}(100)%$`)

	switch {
	case pixel.MatchString(val):
		// convert the string to a float64
		i, err := strconv.Atoi(val[:len(val)-2])

		if err != nil {
			err = fmt.Errorf("extracting %s as a integer: %v", val, err.Error())
		}

		return i, err
	case pcDefault.MatchString(val), percent.MatchString(val):

		// trim the percent if its there
		if string(val[len(val)-1]) == "%" {
			val = val[:len(val)-1]
		}

		offset, err := strconv.ParseFloat(val, 64)

		if err != nil {
			return 0, fmt.Errorf("extracting %s as a percentage : %v", val, err.Error())
		}

		// calculate the percentage rounded
		return int(math.Round((offset / 100) * float64(max))), nil
	default:

		return 0, fmt.Errorf("%s is not a valid offset parameter", val)
	}
}
