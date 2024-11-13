// Package gradients generates pixel gradient ramps
package gradients

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"math"
	"strings"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
)

const (
	rotate180  = "rotate180"
	rotate90   = "rotate90 "
	rotate270  = "rotate270"
	noRotation = "xy"

	WidgetType = "builtin.gradients"
)

// TextBoxGen generates text boxes on a given image based on config values
func RampGenerate(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()

	conf := widgethandler.GenConf[Ramp]{Debug: debug, Schema: Schema, WidgetType: WidgetType}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

func (r Ramp) Handle(resp tsg.Response, req *tsg.Request) {
	// calculate the whole height of each one
	// holderc := context.Background()

	rotation, err := setBase(&r.WidgetProperties, resp.BaseImage().Bounds().Max)

	if err != nil {

		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	// validate the control here
	//	input.StripeGroup.InterStripe.base = input.WidgetProperties

	totalHeight := r.Gradients.GroupSeparator.Height + ((len(r.Gradients.Gradients) - 1) * r.Gradients.GradientSeparator.Height)
	for _, r := range r.Gradients.Gradients {
		totalHeight += r.Height
	}

	totalHeight *= len(r.Groups)
	rowHeight := r.WidgetProperties.rowDimension(resp.BaseImage().Bounds())

	groupStep := float64(rowHeight) / float64(totalHeight)

	position := 0.0
	// posPoint := image.Point{}
	for _, str := range r.Groups {

		if r.Gradients.GroupSeparator.Height != 0 {

			rowHeight := r.Gradients.GroupSeparator.Height
			// draw the header
			end := int(position + groupStep*float64(rowHeight))
			rowCut := r.WidgetProperties.rowOrColumn(resp.BaseImage().Bounds(), end, position)
			row := colour.NewNRGBA64(req.PatchProperties.ColourSpace, rowCut)
			//	row := image.NewNRGBA64(rowCut)
			posPoint := r.WidgetProperties.positionPoint(resp.BaseImage().Bounds().Max, end-int(position), int(position))
			drawSegment(resp.BaseImage(), row, req.PatchProperties.ColourSpace, posPoint, r.Gradients.GroupSeparator)

			position += groupStep * float64(rowHeight)

		}

		for i, ramp := range r.Gradients.Gradients {

			end := int(position + groupStep*float64(ramp.Height))
			rowCut := r.WidgetProperties.rowOrColumn(resp.BaseImage().Bounds(), end, position)
			rrow := colour.NewNRGBA64(req.PatchProperties.ColourSpace, rowCut)
			// rrow := image.NewNRGBA64(rowCut)

			ramp.colour = str.Colour
			ramp.startPoint = str.InitialPixelValue
			ramp.reverse = str.Reverse

			ramp.base = r.WidgetProperties
			posPoint := r.WidgetProperties.positionPoint(resp.BaseImage().Bounds().Max, end-int(position), int(position))
			drawSegment(resp.BaseImage(), rrow, req.PatchProperties.ColourSpace, posPoint, ramp)

			position += groupStep * float64(ramp.Height)
			//	posPoint = r.base.positionPoint(target.Bounds().Max, int(position))
			if i+1 < len(r.Gradients.Gradients) {
				interHeight := r.Gradients.GradientSeparator.Height
				// accounts for jumps in floats and ints
				end := int(position + groupStep*float64(interHeight))

				rowCut := r.WidgetProperties.rowOrColumn(resp.BaseImage().Bounds(), end, position)
				irow := colour.NewNRGBA64(req.PatchProperties.ColourSpace, rowCut)
				// irow := image.NewNRGBA64(rowCut)
				altCopy := r.Gradients.GradientSeparator
				altCopy.base = r.WidgetProperties
				altCopy.step = r.Gradients.Gradients[i+1].BitDepth
				posPoint := r.WidgetProperties.positionPoint(resp.BaseImage().Bounds().Max, end-int(position), int(position))
				drawSegment(resp.BaseImage(), irow, req.PatchProperties.ColourSpace, posPoint, altCopy)

				position += groupStep * float64(interHeight)
				//	posPoint = input.base.positionPoint(target.Bounds().Max, int(position))
				// calculate segments here
			}

		}

	}

	// rotate if required
	// this is not pixel accurate
	if rotation != 0 {
		rotate(resp.BaseImage(), rotation)
	}

	resp.Write(tsg.WidgetSuccess, "success")
}

// Generate runs the ramps widget.
func (r Ramp) Generate(target draw.Image, opts ...any) error {
	// calculate the whole height of each one
	// holderc := context.Background()

	rotation, err := setBase(&r.WidgetProperties, target.Bounds().Max)

	if err != nil {
		return err
	}

	// validate the control here
	//	input.StripeGroup.InterStripe.base = input.WidgetProperties

	totalHeight := r.Gradients.GroupSeparator.Height + ((len(r.Gradients.Gradients) - 1) * r.Gradients.GradientSeparator.Height)
	for _, r := range r.Gradients.Gradients {
		totalHeight += r.Height
	}

	totalHeight *= len(r.Groups)
	rowHeight := r.WidgetProperties.rowDimension(target.Bounds())

	groupStep := float64(rowHeight) / float64(totalHeight)

	position := 0.0
	// posPoint := image.Point{}
	for _, str := range r.Groups {

		if r.Gradients.GroupSeparator.Height != 0 {

			rowHeight := r.Gradients.GroupSeparator.Height
			// draw the header
			end := int(position + groupStep*float64(rowHeight))
			rowCut := r.WidgetProperties.rowOrColumn(target.Bounds(), end, position)
			row := colour.NewNRGBA64(r.ColourSpace, rowCut)
			//	row := image.NewNRGBA64(rowCut)
			posPoint := r.WidgetProperties.positionPoint(target.Bounds().Max, end-int(position), int(position))
			drawSegment(target, row, r.ColourSpace, posPoint, r.Gradients.GroupSeparator)

			position += groupStep * float64(rowHeight)

		}

		for i, ramp := range r.Gradients.Gradients {

			end := int(position + groupStep*float64(ramp.Height))
			rowCut := r.WidgetProperties.rowOrColumn(target.Bounds(), end, position)
			rrow := colour.NewNRGBA64(r.ColourSpace, rowCut)
			// rrow := image.NewNRGBA64(rowCut)

			ramp.colour = str.Colour
			ramp.startPoint = str.InitialPixelValue
			ramp.reverse = str.Reverse

			ramp.base = r.WidgetProperties
			posPoint := r.WidgetProperties.positionPoint(target.Bounds().Max, end-int(position), int(position))
			drawSegment(target, rrow, r.ColourSpace, posPoint, ramp)

			position += groupStep * float64(ramp.Height)
			//	posPoint = r.base.positionPoint(target.Bounds().Max, int(position))
			if i+1 < len(r.Gradients.Gradients) {
				interHeight := r.Gradients.GradientSeparator.Height
				// accounts for jumps in floats and ints
				end := int(position + groupStep*float64(interHeight))

				rowCut := r.WidgetProperties.rowOrColumn(target.Bounds(), end, position)
				irow := colour.NewNRGBA64(r.ColourSpace, rowCut)
				// irow := image.NewNRGBA64(rowCut)
				altCopy := r.Gradients.GradientSeparator
				altCopy.base = r.WidgetProperties
				altCopy.step = r.Gradients.Gradients[i+1].BitDepth
				posPoint := r.WidgetProperties.positionPoint(target.Bounds().Max, end-int(position), int(position))
				drawSegment(target, irow, r.ColourSpace, posPoint, altCopy)

				position += groupStep * float64(interHeight)
				//	posPoint = input.base.positionPoint(target.Bounds().Max, int(position))
				// calculate segments here
			}

		}

	}

	// rotate if required
	// this is not pixel accurate
	if rotation != 0 {
		rotate(target, rotation)
	}

	return nil
}

// setBase sets the starting positions and angles of the base image.
func setBase(target *control, dims image.Point) (float64, error) {
	radian := 0.0
	target.angleType = noRotation

	if target.CwRotation != "" {
		// angString := fmt.Sprintf("%v", target.CwRotation)
		var err error
		radian, err = target.ClockwiseRotationAngle()
		if err != nil {
			return 0, err
		}
	}

	angDiff, angleOffset := diff(radian, 1.571, 4.712, 3.142, 0.0)
	rads := fmt.Sprintf("%.3f", angDiff)

	rowLength := dims.X
	switch rads {
	case "1.571":
		target.angleType = rotate90
		rowLength = dims.Y
	case "4.712":
		target.angleType = rotate270
		rowLength = dims.Y
	case "3.142":
		target.angleType = rotate180
	}

	/*
		calculate shift here

	*/
	if target.MaxBitDepth == 0 {
		target.MaxBitDepth = 16
	} else if target.MaxBitDepth > 16 {
		return 0.0, fmt.Errorf("0DEV The maximum bit depth is 16 bits, received a value of %v", target.MaxBitDepth)
	}
	//	stepLength := math.Pow(2, float64(target.MaxBitDepth))
	// step := float64(rowLength) / stepLength

	if target.ObjectFitFill {

		stepLength := math.Pow(2, float64(target.MaxBitDepth))
		step := float64(rowLength) / stepLength
		target.truePixelShift = step
	} else {

		if target.PixelValueRepeat == 0 {
			target.truePixelShift = 1
		} else {
			target.truePixelShift = float64(target.PixelValueRepeat)
		}

	}

	return angleOffset, nil
}

func diff(angle float64, targets ...float64) (target float64, diff float64) {
	diff = math.Pi * 2
	target = 0.0
	for _, t := range targets {
		calcDiff := math.Abs(angle - t)
		if calcDiff < diff {
			diff = calcDiff
			target = t
		}
	}

	return
}

func (h groupSeparator) Generate(img draw.Image, cspace colour.ColorSpace) {
	if h.Height == 0 {
		return
	}

	c, _ := assignRGBValues(h.Colour, 65535, 0, 65535)
	c.UpdateColorSpace(cspace)
	colour.Draw(img, img.Bounds(), &image.Uniform{&c}, image.Point{}, draw.Over)
}

func (a gradientSeparator) Generate(img draw.Image, cspace colour.ColorSpace) {

	if a.Height == 0 {
		return
	}

	bitStep := int(math.Pow(2, float64((a.base.MaxBitDepth - a.step))))
	shiftStep := a.base.truePixelShift * float64(bitStep)

	altCount := 0
	end := a.base.getLoop(img.Bounds())
	xPosition := 0.0

	for xPosition <= float64(end) {
		stepEnd := int(xPosition + shiftStep)
		c, _ := assignRGBValues(a.Colours[altCount%len(a.Colours)], 65535, 0, 65535)
		target := a.base.set(int(xPosition), stepEnd-int(xPosition), img.Bounds().Max)
		c.UpdateColorSpace(cspace)
		colour.Draw(img, target, &image.Uniform{&c}, image.Point{}, draw.Over)
		altCount++
		xPosition += shiftStep
	}

}

func (s Gradient) Generate(img draw.Image, cspace colour.ColorSpace) {
	shift16 := 1 << (16 - s.BitDepth)

	// set the steps relative to the max bitdepth
	bitStep := int(math.Pow(2, float64((s.base.MaxBitDepth - s.BitDepth))))
	// multiply the step by the shift factor
	shiftStep := s.base.truePixelShift * float64(bitStep)

	// generate a start point in 16 bit
	// sanity check the start point is within the bitdepth
	startPoint := s.startPoint << (16 - s.base.MaxBitDepth)

	overRun := startPoint % shift16

	if overRun != 0 {
		/*
			tow options line up when we can let them lineup when required
			or have them always lineup by shifting the value to match the lowest bit closest bitdepth
			e.g. would shift to 0?@TODO fix eveything else before tackling this
		*/
		if !s.reverse {
			startPoint += (shift16 - overRun)
		} else {
			startPoint -= overRun
		}
	}

	altCount := 0
	// have the mover position and the bounds
	end := s.base.getLoop(img.Bounds())
	xPosition := 0.0
	for xPosition <= float64(end) {
		stepEnd := int(xPosition + shiftStep)

		c, _ := assignRGBValues(s.colour, float64(startPoint), 0, 65535)
		target := s.base.set(int(xPosition), stepEnd-int(xPosition), img.Bounds().Max)

		// colour.Draw(img, image.Rect(x, img.Bounds().Min.Y, x+step, img.Bounds().Max.Y), &image.Uniform{c}, image.Point{}, draw.Over)
		c.UpdateColorSpace(cspace)
		colour.Draw(img, target, &image.Uniform{&c}, image.Point{}, draw.Over)
		altCount++

		// make the colour steps 16 bit
		if !s.reverse {
			startPoint += shift16 // 1 << shift16
		} else {
			startPoint -= shift16 // 1 << shift16
		}

		xPosition += shiftStep
	}

	// generate the label if needed
	if s.Label != "" {
		s.base.TextProperties.labels(img, cspace, s.Label, s.base.angleType)
	}

	// run the labels here - use the other label code
}

func (c control) getLoop(bounds image.Rectangle) (end int) {

	if c.angleType == noRotation || c.angleType == rotate180 {
		end = bounds.Max.X
	} else {
		end = bounds.Max.Y
	}
	return
}

func (c control) rowDimension(bounds image.Rectangle) (end int) {

	if c.angleType == noRotation || c.angleType == rotate180 {
		end = bounds.Max.Y
	} else {
		end = bounds.Max.X
	}
	return
}

func (c control) rowOrColumn(bounds image.Rectangle, end int, position float64) image.Rectangle {
	if c.angleType == noRotation || c.angleType == rotate180 {
		return image.Rect(0, 0, bounds.Dx(), end-int(position))
	} else {
		return image.Rect(0, 0, end-int(position), bounds.Dy())
	}

}

func (c control) positionPoint(bounds image.Point, rowSize, shift int) image.Point {

	switch c.angleType {
	case noRotation:
		return image.Point{Y: shift}
	case rotate180:
		return image.Point{Y: bounds.Y - shift - rowSize}
	case rotate90:
		return image.Point{X: bounds.X - shift - rowSize}
	default: // rotate270
		return image.Point{X: shift}
	}

}

// func set sets the canvas position based on the rotation without running a transformation
func (c control) set(position, step int, bounds image.Point) image.Rectangle {

	switch c.angleType {
	case noRotation:
		return image.Rect(position, 0, position+step, bounds.Y)
	case rotate180:
		return image.Rect(bounds.X-(position), 0, bounds.X-(position+step), bounds.Y)
	case rotate90:

		return image.Rect(0, position, bounds.X, position+step)
	case rotate270:

		return image.Rect(0, bounds.Y-(position), bounds.X, bounds.Y-(position+step))
	}

	return image.Rectangle{}
}

type maker interface {
	Generate(img draw.Image, cspace colour.ColorSpace)
}

// drawSegment runs the segment drawer
func drawSegment(base, img draw.Image, cspace colour.ColorSpace, start image.Point, g maker) {

	g.Generate(img, cspace) // add optional parameterss?

	colour.Draw(base, img.Bounds().Add(start), img, image.Point{}, draw.Over)
}

func assignRGBValues(colourString string, rgb float64, maxBlack, maxWhite uint16) (colour.CNRGBA64, error) {
	switch strings.ToLower(colourString) {
	case "grey", "gray": // "black", "white",
		return colour.CNRGBA64{R: uint16(rgb), G: uint16(rgb), B: uint16(rgb), A: 0xffff}, nil
	case "black":
		return colour.CNRGBA64{R: maxBlack, G: maxBlack, B: maxBlack, A: 0xffff}, nil
	case "white":
		return colour.CNRGBA64{R: maxWhite, G: maxWhite, B: maxWhite, A: 0xffff}, nil
	case "red":
		return colour.CNRGBA64{R: uint16(rgb), A: 0xffff}, nil
	case "green":
		return colour.CNRGBA64{G: uint16(rgb), A: 0xffff}, nil
	case "blue":
		return colour.CNRGBA64{B: uint16(rgb), A: uint16(0xffff)}, nil
	default:
		return colour.CNRGBA64{}, fmt.Errorf("%s Non specific colour called, rgb values set at 0", colourString) // Unused error
	}
}

func rotate(canvas draw.Image, radian float64) {

	// Take n as 10 for the moment
	// Math.ceil x, y and floor each one

	size := canvas.Bounds().Max
	x0, y0 := float64((size.X / 2)), float64((size.Y / 2))
	// Calculate these on initialisation
	// Use base as a method of calculating it all without changing the canvas
	base := image.NewNRGBA64(canvas.Bounds())
	colour.Draw(base, base.Bounds(), canvas, image.Point{}, draw.Src)
	N := int(10)

	rgbs := make([][]uint32, 4)
	for i := range rgbs {
		rgbs[i] = make([]uint32, 4)
	}

	val := make([]uint16, 4)

	for i := 0; i < base.Bounds().Max.X; i++ {
		for j := 0; j < base.Bounds().Max.Y; j++ {
			// Calculate the pixel location to extract from
			xp := math.Cos(-radian)*(float64(i)-x0) + math.Sin(-radian)*(float64(j)-y0) + x0
			yp := -1*math.Sin(-radian)*(float64(i)-x0) + math.Cos(-radian)*(float64(j)-y0) + y0
			_, xFrac := math.Modf(xp)
			_, yFrac := math.Modf(yp)
			x := int(xFrac * 10)
			y := int(yFrac * 10)

			xpos, ypos := int(math.Floor(xp)), int(math.Floor(yp))
			locs := [][]int{{xpos, ypos}, {xpos + 1, ypos}, {xpos, ypos + 1}, {xpos + 1, ypos + 1}}
			// Overwrite the rgb values each time instead of making a new array for each loop
			for i, loc := range locs {
				rgbs[i][0], rgbs[i][1], rgbs[i][2], rgbs[i][3] = base.At(loc[0], loc[1]).RGBA()
			}
			for k := 0; k < 4; k++ {
				val[k] = uint16((1.0 / (float64(N * N))) * float64(((N-x)*(N-y)*int(rgbs[0][k]) + x*(N-y)*int(rgbs[1][k]) + y*(N-x)*int(rgbs[2][k]) + x*y*int(rgbs[3][k]))))
			}

			// If not empty then assign the value to ignore the black background
			// colourspace is not required as it was already changed during the generation
			// of the ramps
			if val[3] != 0 {
				canvas.Set(i, j, &colour.CNRGBA64{R: val[0], G: val[1], B: val[2], A: 0xffff})
			} else {
				canvas.Set(i, j, &colour.CNRGBA64{})
			}
		}
	}
}
