// Package zoneplate is used to generate a square zoneplate
package zoneplate

import (
	"context"
	"fmt"
	"image/draw"
	"math"
	"strings"
	"sync"

	"github.com/mmTristan/opentsg-modules/opentsg-core/anglegen"
	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mmTristan/opentsg-modules/opentsg-core/errHandle"
	"github.com/mmTristan/opentsg-modules/opentsg-core/widgethandler"
)

const (
	widgetType = "builtin.zoneplate"
)

// zoneGen takes a canvas and then returns an image of the zone plate layered ontop of the image
func ZoneGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := widgethandler.GenConf[zoneplateJSON]{Debug: debug, Schema: schemaInit, WidgetType: widgetType}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

func (z zoneplateJSON) Generate(canvas draw.Image, opts ...any) error {
	// Get the config for the zoneplate
	platetype := z.Platetype
	// Check zoneplates have been called
	if platetype == "" {
		return fmt.Errorf("0111 No zone plate module selected")
	}
	w, h := canvas.Bounds().Max.X, canvas.Bounds().Max.Y // PlateDime()

	x := float64(w) // Float64(canvas.Bounds().Max.X)
	y := float64(h) // float64(canvas.Bounds().Max.Y)
	zv := zoneConst(x, y)

	// Offset is 0 to pi to move from black to white
	offset := startOffset(z.Startcolour)
	//	platetilt := z.Angle
	var angle float64
	if z.Angle != nil {
		var err error
		angle, err = anglegen.AngleCalc(fmt.Sprintf("%v", z.Angle))
		if err != nil {
			return err
		}
	}

	for i := zv.yNeg; i < zv.yPos; i++ {

		for j := zv.xNeg; j < zv.xPos; j++ {
			r := radialCalc(platetype, float64(j), float64(i), angle, float64(w), float64(h))

			zone := math.Sin((zv.km*r*r)/(2*zv.rm)+offset) * (0.5*math.Tanh((zv.rm-r)/zv.w) + 0.5)

			// Assign colour as an integer between 0 and 4095 as g scaled out of
			colourPos := uint16(4095*((zone+1)/2)) << 4 // Uint16 acts as a floor function

			fill := colour.CNRGBA64{R: colourPos, G: colourPos, B: colourPos, A: 0xffff, ColorSpace: z.ColourSpace}
			canvas.Set(int(j+zv.xPos), int(i+zv.yPos), &fill)
		}
	}
	//@TODO fix this
	/*
		// Check if needs to be masked and apply it
		if maskShape := z.Mask; maskShape != "" {
			// At the moment just make a mask around the zoneplate
			canvas = mask.Mask(maskShape, w, h, 0, 0, canvas)

		}*/

	return nil
}

type zoneVars struct {
	xNeg float64
	xPos float64
	yNeg float64
	yPos float64
	km   float64
	rm   float64
	w    float64
}

const (
	sweepPattern  = "sweep"
	circlePattern = "circular"
)

func radialCalc(plateType string, x, y float64, radian float64, w, h float64) float64 {

	// To radians
	// Generate the angle as a string and then extract the value in radians

	// Calculate new x and y values based off of the chosen angle
	xp := x*math.Cos(radian) - y*math.Sin(radian)
	yp := x*math.Sin(radian) + y*math.Cos(radian)
	if plateType == sweepPattern {
		return math.Abs(yp)
		// } else if plateType == "ellipse" {
		//	return math.Sqrt((2 * xp * xp) + (yp * yp))
	}

	switch {
	case w > h:
		// Fmt.Println(float64(w / h))
		return math.Sqrt((xp * xp) + ((w / h) * (w / h) * yp * yp))
	case h > w:
		// Fmt.Println(float64(h / w))
		return math.Sqrt(((h / w) * (h / w) * xp * xp) + (yp * yp))
	default:
		return math.Sqrt((xp * xp) + (yp * yp))
	}
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

func zoneConst(x, y float64) (zv zoneVars) {

	zv.km = 0.8 * math.Pi

	if int(x)%2 == 1 {
		zv.cartesian(x-1, y-1)
	} else {
		zv.cartesian(x, y)
	}
	zv.w = zv.rm / 5

	return zv
}

// set x and y to be 0,0 in the middle
func (zv *zoneVars) cartesian(x, y float64) {
	zv.xNeg = -1 * x / 2
	zv.yNeg = -1 * y / 2
	zv.xPos = -1 * zv.xNeg
	zv.yPos = -1 * zv.yNeg
	zv.rm = float64(x)
}
