// package parameters converts strings to floats for the opentsg angles
package parameters

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
)

// RotationAngle is a string for allowing multiple input formats
// that can then be parsed to radians.
type RotationAngle struct {
	CwRotation any `json:"cwRotation,omitempty" yaml:"cwRotation,omitempty"`
}

/*
AngleField is a struct that wraps an any.
It is to be used as a type any for calculating the angle of
anything that is not got with parameters.

e.g.

	type testStruct struct {
		MyNewAngle AngleField `json:"myNewAngle"`
	}

This can then be parsed as a first level object in json or yaml.

Encoding is also provided
*/
type AngleField struct {
	Ang any `yaml:",flow"`
}

func (a *AngleField) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ang any
	err := unmarshal(&ang)
	if err != nil {
		return err
	}

	a.Ang = ang
	return nil
}

func (a AngleField) MarshalYAML() (interface{}, error) {
	return a.Ang, nil
}

func (a *AngleField) UnmarshalJSON(data []byte) error {
	var ang any
	err := json.Unmarshal(data, &ang)

	a.Ang = ang
	return err
}

func (a AngleField) MarshalJSON() ([]byte, error) {

	return json.Marshal(a.Ang)

}

func (a AngleField) GetAngle() (float64, error) {

	return anyToAng(a.Ang)
}

// RotationAngle is a string for allowing multiple input formats
// that can then be parsed to radians.
type StartAngle struct {
	StartAng any `json:"startAngle,omitempty" yaml:"startAngle,omitempty"`
}

func (s *StartAngle) GetStartAngle() (float64, error) {
	// piUnit := []rune{960} //[]rune{960} this is how the pi image is changed

	if s.StartAng == nil {
		return 0, nil
	}
	return anyToAng(s.StartAng)
}

// ClockwiseRotationAngle returns the float equivalent of a pi string e.g. π*1/4 or degree values as a string, converted to radians.
// patterns to be used are :
//
// π\*(\d){1,4}/{1}(\d){1,4}$
//
// ^π\*(\d){1,4}$
//
// ^(\d){0,}(\.){0,1}(\d){0,}$
func (a *RotationAngle) ClockwiseRotationAngle() (float64, error) {
	// piUnit := []rune{960} //[]rune{960} this is how the pi image is changed

	if a.CwRotation == nil {
		return 0, nil
	}
	return anyToAng(a.CwRotation)
}

func anyToAng(anyAng any) (float64, error) {
	ang := fmt.Sprintf("%v", anyAng)
	/*regex here the types of radian and degree*/
	raidanFrac := regexp.MustCompile(`π\*(\d){1,4}/{1}(\d){1,4}$`)
	radian := regexp.MustCompile(`^π\*(\d){1,4}$`)
	degree := regexp.MustCompile(`^(\d){0,}(\.){0,1}(\d){0,}$`)

	// check angle type
	switch {
	case radian.MatchString(ang) || raidanFrac.MatchString(ang):
		// convert the string to a float64
		rad := stringTofraction(ang[3:])

		return (math.Pi * rad), nil
	case degree.MatchString(ang):
		degree, err := strconv.ParseFloat(ang, 64)

		return (math.Pi * (degree / 180)), err
	default:

		return 0, fmt.Errorf("%s is not a valid angle", anyAng)
	}
}

func stringTofraction(form string) float64 {
	var angle float64
	var pos int
	div := false

	for i, c := range form {
		if c == 47 { // 47 is "/"
			div = true
			pos = i
		}
	}

	if div {
		num, _ := strconv.ParseFloat(form[:pos], 64)
		dom, _ := strconv.ParseFloat(form[pos+1:], 64)
		angle = num / dom

	} else {
		angle, _ = strconv.ParseFloat(form, 64)
	}

	return angle
}
