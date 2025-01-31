// Package zoneplate is used to generate a square zoneplate
package zoneplate

import (
	"fmt"
	"math"
	"strings"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
)

const (
	WidgetType = "builtin.zoneplate"
)

func (z ZConfig) Handle(resp tsg.Response, req *tsg.Request) {
	frequency, _ := z.Frequency.GetAngle()
	if frequency > math.Pi {
		frequency = math.Pi
	} else if frequency == 0 {
		frequency = 0.8 * math.Pi
	}

	// set up constants for the zone plate
	b := resp.BaseImage().Bounds().Max
	rm := float64(b.X)
	w := rm / 5

	// set up the offset, this is centred in the middle of the box
	off, err := z.CalcOffset(resp.BaseImage().Bounds().Max)
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}
	xOffset := b.X/2 + off.X
	yOffset := b.Y/2 + off.Y

	yMagnitude := 1.0
	// set xy to radius function
	extractFunc := xyToRadius
	switch z.PlateType {
	case verticalSweep, sweepPattern:
		extractFunc = xyToVerticalRadius
	case horizontalSweep:
		extractFunc = xyToHorizontalRadius
		//	pattern, _, _ = createWeights16(b.X, 6, z.baseX/z.destX, lan)
	case circlePattern, "":
	case ellipse:
		yMagnitude = 0.5
	default:
		resp.Write(tsg.WidgetError, fmt.Sprintf("unknown plateType \"%v\"", z.PlateType))
		return
	}

	// set zone plate function
	zplate := zPlate
	switch z.WaveType {
	case Sin:
		zplate = sPlate
	case Cos:
		zplate = cPlate
	case "cos*sin^2":
		zplate = tPlate
	}

	ztc := zoneToColour
	if len(z.Colors) > 0 {
		colours := make([]colour.CNRGBA64, len(z.Colors))
		for i, c := range z.Colors {
			colours[i] = *c.ToColour(req.PatchProperties.ColourSpace)
		}

		ztc = func(zone float64) colour.CNRGBA64 {
			//	fmt.Println(((zone+1)/2)*(float64(len(colours))), zone)
			pos := int(((zone+1)/2)*(float64(len(colours)))) % len(colours)
			return colours[pos]
		}
	}

	rotation, err := z.ClockwiseRotationAngle()
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	rotationOffset := startOffset(z.Startcolour)

	for x := 0; x < b.X; x++ {
		for y := 0; y < b.Y; y++ {

			xp, yp := rotate(float64(x-xOffset), float64(y-yOffset)*yMagnitude, rotation)
			r := extractFunc(xp, yp)

			//	zone := math.Sin((z.km*r*r)/(2*rm)+offset) * (0.5*math.Tanh((rm-r)/w) + 0.5)
			zone := zplate(r, frequency, rm, w, rotationOffset)

			// assign the colour and draw the canvas
			fill := ztc(zone)
			resp.BaseImage().Set(x, y, &fill)

		}
	}

	resp.Write(tsg.WidgetSuccess, "success")
}

const (
	sweepPattern    = "sweep"
	verticalSweep   = "verticalSweep"
	horizontalSweep = "horizontalSweep"
	circlePattern   = "circular"
	ellipse         = "ellipse"

	// Plate Types
	Sin = "sin"
	Cos = "cos"
	zp  = "zonePlate"
)

func zoneToColour(zone float64) colour.CNRGBA64 {
	//colourPos := 8192 + uint16(49151*(zone+1)/2)
	colourPos := uint16(0xffff * (zone + 1) / 2)
	return colour.CNRGBA64{R: colourPos, G: colourPos, B: colourPos, A: 0xffff}
}

func zPlate(r, km, rm, w, rotationOffset float64) float64 {
	return math.Sin((km*r*r)/(2*rm)+rotationOffset) * (0.5*math.Tanh((rm-r)/w) + 0.5)
}

func sPlate(r, km, rm, w, rotationOffset float64) float64 {

	return math.Sin(r*km + rotationOffset)
}

func cPlate(r, km, rm, w, rotationOffset float64) float64 {

	return math.Cos(r*km + rotationOffset)
}

func tPlate(r, km, rm, w, rotationOffset float64) float64 {

	return math.Cos(r*km) * math.Sin(r*km) * math.Sin(r*km) // * 2.598
}

// @TODO add a non decay version as well

func rotate(x, y, angle float64) (float64, float64) {

	if angle == 0 {
		return x, y
	}

	xp := x*math.Cos(angle) - y*math.Sin(angle)
	yp := x*math.Sin(angle) + y*math.Cos(angle)

	return xp, yp
}

func xyToAngle(x, y float64) float64 {

	ang := math.Atan2(y, x)

	// add 2 pi by the inverse to keep the angle
	// incrementinh
	if ang < 0 {
		return ang + math.Pi*2
	}

	return ang
}

func startOffset(start string) float64 {
	// Set the phi for sin to move the base colour from 0 to 1 or -1
	switch strings.ToLower(start) {
	case "white":
		return (math.Pi / 2)
	case "black":
		return -1 * (math.Pi / 2)
	default:
		return 0
	}
}

func xyToRadius(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}

func xyToHorizontalRadius(_, y float64) float64 {
	return y
}

func xyToVerticalRadius(x, _ float64) float64 {

	return x // math.Sqrt(x*x + x*x)
}
