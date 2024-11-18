package resize

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/canvaswidget"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/text"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/utils/parameters"
	"gonum.org/v1/gonum/mat"

	_ "embed"
)

const (
	WidgetType = "builtin.resize"
)

var Schema = []byte(`{}`)

// zoneGen takes a canvas and then returns an image of the zone plate layered ontop of the image
func Gen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := widgethandler.GenConf[Config]{Debug: debug, Schema: []byte("{}"), WidgetType: WidgetType, ExtraOpt: []any{c}}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

type Config struct {
	XDetections []*parameters.DistanceField `json:"xDetections,omitempty" yaml:"xDetections,omitempty"`
	YDetections []*parameters.DistanceField `json:"yDetections,omitempty" yaml:"yDetections,omitempty"`
	XStep       *parameters.DistanceField   `json:"xStep,omitempty" yaml:"xStep,omitempty"` // set to be percentage
	YStep       *parameters.DistanceField   `json:"yStep,omitempty" yaml:"yStep,omitempty"`
	XStepEnd    *parameters.DistanceField   `json:"xStepEnd,omitempty" yaml:"xStepEnd,omitempty"` // set to be percentage
	YStepEnd    *parameters.DistanceField   `json:"yStepEnd,omitempty" yaml:"yStepEnd,omitempty"`
	//
	Graticule graticule `json:"graticule,omitempty"`
	// insert generic location ones
	ColourSpace       colour.ColorSpace `json:"colorSpace,omitempty" yaml:"colorSpace,omitempty"`
	config.WidgetGrid `yaml:",inline"`
}

type graticule struct {
	TextColor       string               `json:"textColor,omitempty"`
	GraticuleColour parameters.HexString `json:"graticuleColor,omitempty"`
	Position        string               `json:"position,omitempty"`
	// Width           anglegen.Parameter is this needed?

}

// detection is used by the processing
type detection struct {
	direction string
	size      int
}

// update getPicture size with tests for functions
var getPictureSize = canvaswidget.GetPictureSize

func (c Config) Handle(resp tsg.Response, req *tsg.Request) {

	dest := req.FrameProperties.FrameDimensions

	// get the resizes
	detections, err := c.ExtractResizes(dest)

	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	// set up the sub widget boxes that we draw in
	b := resp.BaseImage().Bounds().Max
	totalBoxes := len(detections)
	xBoxCount, yBoxCount := boxSetUp(totalBoxes, b)

	xSize := float64(b.X) / float64(xBoxCount)
	ySize := float64(b.Y) / float64(yBoxCount)

	for i, dir := range detections {

		// calculate the size of the box to draw the resize in
		// including scales to avoid any gaps
		x := i % (xBoxCount)
		y := i / (xBoxCount)

		// fmt.Println(x, y, xBoxCount, yBoxCount, i)
		// rearrange the order for narrow images
		if float64(b.X)/float64(b.Y) < 1 {
			x = i / (yBoxCount)
			y = i % (yBoxCount)
		}

		boxX := int(xSize*float64(x+1)) - int((xSize)*float64(x))
		boxY := int(ySize*float64(y+1)) - int((ySize)*float64(y))
		box := image.NewNRGBA64(image.Rect(0, 0, boxX, boxY))

		baseX := 0.0
		boxDirSize := 0.0
		switch dir.direction {
		case "y":
			boxDirSize = float64(boxY)
			baseX = float64(dest.Y)
		case "x":
			boxDirSize = float64(boxX)
			baseX = float64(dest.X)

		}
		// get the scale of the final box

		boxSize := int(math.Round((float64(dir.size) / baseX) * boxDirSize))

		if boxSize < 1 {
			boxSize = 1
		}

		// get the ratio of the resize
		ratio := baseX / float64(dir.size)

		var colIntesity []uint16

		// for small changes it starts at grey
		// so we use alternating lines that will induce
		// banding or a third colour.
		if ratio < 1.35 {
			// get it so it scales to 1
			// higher ratios scale to 1 for more contrast
			contrast := (0.35 - (1.35 - ratio)) / 0.35
			intense := 0x2000 * contrast
			colIntesity = make([]uint16, int(baseX))

			for i := range colIntesity {
				if i%2 == 0 {
					colIntesity[i] = uint16(0x9000 - intense)
				} else {
					colIntesity[i] = uint16(0xC000 + intense)
				}
			}

		} else {

			// if the size is a multiple of a quarter of the
			// width, then it starts as grey so we need to intervene
			quarter := int(baseX / 4)
			if dir.size%quarter == 0 {
				ratio -= 0.05
				// resize the box size to avoid crashing the simultaneous equations later
				boxSize = int(math.Round((float64(1/ratio) * boxDirSize)))
				if boxSize < 1 {
					boxSize = 1
				}
			}

			// calculate and solve the coefficients
			coeffs := createCoefficients(boxSize, 6, ratio, lan)
			var err error
			colIntesity, err = coeffSolver(coeffs, int(boxDirSize))
			if err != nil {
				resp.Write(tsg.WidgetError, err.Error())
				return
			}
		}

		cols := make([]color.NRGBA64, len(colIntesity))
		for i, c := range colIntesity {
			cols[i] = color.NRGBA64{R: c, G: c, B: c, A: 0xffff}
		}

		// draw the lines
		for x := 0; x < boxX; x++ {
			for y := 0; y < boxY; y++ {
				r := 0
				switch dir.direction {
				case "y":
					r = y
				case "x":
					r = x
				}
				//	fmt.Println(cycle)
				box.Set(x, y, cols[r])
			}
		}

		// generate the text and the graticule
		err := c.generateText(box, xSize, ySize, fmt.Sprintf("%s to %v", dir.direction, dir.size))
		if err != nil {
			resp.Write(tsg.WidgetError, err.Error())
			return
		}
		// draw the box on the whole canvas
		draw.Draw(resp.BaseImage(), image.Rect(int((xSize)*float64(x)), int((ySize)*float64(y)), int(xSize*float64(x+1)), int(ySize*float64(y+1))), box, image.Point{}, draw.Over)

	}

	resp.Write(tsg.WidgetSuccess, "success")
}

func (r Config) Generate(canvas draw.Image, extras ...any) error {

	if len(extras) != 1 {
		return fmt.Errorf("configuration error, need an context input")
	}

	c, ok := extras[0].(*context.Context)

	if !ok {
		return fmt.Errorf("invalid context used as additional input")
	}

	dest := getPictureSize(*c)

	// get the resizes
	detections, err := r.ExtractResizes(dest)

	if err != nil {
		return err
	}

	// set up the sub widget boxes that we draw in
	b := canvas.Bounds().Max
	totalBoxes := len(detections)
	xBoxCount, yBoxCount := boxSetUp(totalBoxes, b)

	xSize := float64(b.X) / float64(xBoxCount)
	ySize := float64(b.Y) / float64(yBoxCount)

	for i, dir := range detections {

		// calculate the size of the box to draw the resize in
		// including scales to avoid any gaps
		x := i % (xBoxCount)
		y := i / (xBoxCount)

		// fmt.Println(x, y, xBoxCount, yBoxCount, i)
		// rearrange the order for narrow images
		if float64(b.X)/float64(b.Y) < 1 {
			x = i / (yBoxCount)
			y = i % (yBoxCount)
		}

		boxX := int(xSize*float64(x+1)) - int((xSize)*float64(x))
		boxY := int(ySize*float64(y+1)) - int((ySize)*float64(y))
		box := image.NewNRGBA64(image.Rect(0, 0, boxX, boxY))

		baseX := 0.0
		boxDirSize := 0.0
		switch dir.direction {
		case "y":
			boxDirSize = float64(boxY)
			baseX = float64(dest.Y)
		case "x":
			boxDirSize = float64(boxX)
			baseX = float64(dest.X)

		}
		// get the scale of the final box

		boxSize := int(math.Round((float64(dir.size) / baseX) * boxDirSize))

		if boxSize < 1 {
			boxSize = 1
		}

		// get the ratio of the resize
		ratio := baseX / float64(dir.size)

		var colIntesity []uint16

		// for small changes it starts at grey
		// so we use alternating lines that will induce
		// banding or a third colour.
		if ratio < 1.35 {
			// get it so it scales to 1
			// higher ratios scale to 1 for more contrast
			contrast := (0.35 - (1.35 - ratio)) / 0.35
			intense := 0x2000 * contrast
			colIntesity = make([]uint16, int(baseX))

			for i := range colIntesity {
				if i%2 == 0 {
					colIntesity[i] = uint16(0x9000 - intense)
				} else {
					colIntesity[i] = uint16(0xC000 + intense)
				}
			}

		} else {

			// if the size is a multiple of a quarter of the
			// width, then it starts as grey so we need to intervene
			quarter := int(baseX / 4)
			if dir.size%quarter == 0 {
				ratio -= 0.05
				// resize the box size to avoid crashing the simultaneous equations later
				boxSize = int(math.Round((float64(1/ratio) * boxDirSize)))
				if boxSize < 1 {
					boxSize = 1
				}
			}

			// calculate and solve the coefficients
			coeffs := createCoefficients(boxSize, 6, ratio, lan)
			var err error
			colIntesity, err = coeffSolver(coeffs, int(boxDirSize))
			if err != nil {
				return err
			}
		}

		cols := make([]color.NRGBA64, len(colIntesity))
		for i, c := range colIntesity {
			cols[i] = color.NRGBA64{R: c, G: c, B: c, A: 0xffff}
		}

		// draw the lines
		for x := 0; x < boxX; x++ {
			for y := 0; y < boxY; y++ {
				r := 0
				switch dir.direction {
				case "y":
					r = y
				case "x":
					r = x
				}
				//	fmt.Println(cycle)
				box.Set(x, y, cols[r])
			}
		}

		// generate the text and the graticule
		err := r.generateText(box, xSize, ySize, fmt.Sprintf("%s to %v", dir.direction, dir.size))
		if err != nil {
			return err
		}
		// draw the box on the whole canvas
		draw.Draw(canvas, image.Rect(int((xSize)*float64(x)), int((ySize)*float64(y)), int(xSize*float64(x+1)), int(ySize*float64(y+1))), box, image.Point{}, draw.Over)

	}

	return nil
}

func (r Config) generateText(box draw.Image, xSize, ySize float64, label string) error {

	// calculate a basic graticule
	graticuleWidth := int(math.Min(math.Ceil(xSize/100), math.Ceil(ySize/100)))

	if r.Graticule.TextColor == "" {
		r.Graticule.TextColor = "#888888"
	}

	if r.Graticule.GraticuleColour == "" {
		r.Graticule.GraticuleColour = "#f0f0f0"
	}

	gratColour := r.Graticule.GraticuleColour.ToColour(r.ColourSpace)

	// draw text here
	var textBox *image.NRGBA64
	textBoxer := text.NewTextboxer(r.ColourSpace,
		text.WithTextColourString(r.Graticule.TextColor))

	var tbOffset image.Point

	// set the textbox and graticule shape

	switch r.Graticule.Position {
	case text.AlignmentLeft, text.AlignmentRight:
		// set the text to be vertical
		text.WithVerticalText(true)(textBoxer)
		gWidth := int(math.Min(math.Ceil(xSize/4), math.Ceil(ySize/4)))

		textBox = image.NewNRGBA64(image.Rect(0, 0, box.Bounds().Dx()/4, box.Bounds().Dy()-graticuleWidth*2))
		if r.Graticule.Position == text.AlignmentRight {
			// set the position
			tbOffset = image.Point{X: (3*box.Bounds().Dx())/4 - graticuleWidth, Y: graticuleWidth}
			draw.Draw(box, image.Rect(box.Bounds().Dx()-graticuleWidth, 0, box.Bounds().Dx(), box.Bounds().Dy()), &image.Uniform{gratColour}, image.Point{}, draw.Over)

			// draw the side lines
			draw.Draw(box, image.Rect(box.Bounds().Dx()-gWidth, 0, box.Bounds().Dx(), graticuleWidth), &image.Uniform{gratColour}, image.Point{}, draw.Over)
			draw.Draw(box, image.Rect(box.Bounds().Dx()-gWidth, box.Bounds().Dy()-graticuleWidth, box.Bounds().Dx(), box.Bounds().Dy()), &image.Uniform{gratColour}, image.Point{}, draw.Over)
		} else {
			tbOffset = image.Point{X: graticuleWidth, Y: graticuleWidth}
			draw.Draw(box, image.Rect(0, 0, graticuleWidth, box.Bounds().Dy()), &image.Uniform{gratColour}, image.Point{}, draw.Over)

			draw.Draw(box, image.Rect(0, 0, gWidth, graticuleWidth), &image.Uniform{gratColour}, image.Point{}, draw.Over)
			draw.Draw(box, image.Rect(0, box.Bounds().Dy()-graticuleWidth, gWidth, box.Bounds().Dy()), &image.Uniform{gratColour}, image.Point{}, draw.Over)
		}

		// set the x alignment of the text to be flush
		// to the graticule
		text.WithXAlignment(r.Graticule.Position)(textBoxer)
	case text.AlignmentTop, text.AlignmentBottom:
		textBox = image.NewNRGBA64(image.Rect(0, 0, box.Bounds().Dx()-graticuleWidth*2, box.Bounds().Dy()/4))
		gWidth := int(math.Min(math.Ceil(xSize/4), math.Ceil(ySize/4)))

		if r.Graticule.Position == text.AlignmentBottom {
			tbOffset = image.Point{Y: (3*box.Bounds().Dy())/4 - graticuleWidth, X: graticuleWidth}
			draw.Draw(box, image.Rect(0, box.Bounds().Dy()-graticuleWidth, box.Bounds().Dx(), box.Bounds().Dy()), &image.Uniform{gratColour}, image.Point{}, draw.Over)

			draw.Draw(box, image.Rect(0, box.Bounds().Dy()-gWidth, graticuleWidth, box.Bounds().Dy()), &image.Uniform{gratColour}, image.Point{}, draw.Over)
			draw.Draw(box, image.Rect(box.Bounds().Dx()-graticuleWidth, box.Bounds().Dy()-gWidth, box.Bounds().Dx(), box.Bounds().Dy()), &image.Uniform{gratColour}, image.Point{}, draw.Over)
		} else {
			tbOffset = image.Point{Y: graticuleWidth, X: graticuleWidth}
			draw.Draw(box, image.Rect(0, 0, box.Bounds().Dx(), graticuleWidth), &image.Uniform{gratColour}, image.Point{}, draw.Over)

			draw.Draw(box, image.Rect(0, 0, graticuleWidth, gWidth), &image.Uniform{gratColour}, image.Point{}, draw.Over)
			draw.Draw(box, image.Rect(box.Bounds().Dx()-graticuleWidth, 0, box.Bounds().Dx(), gWidth), &image.Uniform{gratColour}, image.Point{}, draw.Over)

			// image.NewNRGBA64(box.Bounds())
		}

		// set the y alignment of the text to be flush with the graticule
		text.WithYAlignment(r.Graticule.Position)(textBoxer)
	default:
		// default is just fill, with no graticule
		textBox = image.NewNRGBA64(box.Bounds())
		// if image is portrait set the text to be vertical
		if ySize/xSize > 1.3 {
			text.WithVerticalText(true)(textBoxer)
		}
	}

	err := textBoxer.DrawStrings(textBox, nil, []string{label})
	if err != nil {
		return err
	}

	// draw the text over the lines
	draw.Draw(box, image.Rectangle{tbOffset, box.Bounds().Max}, textBox, image.Point{}, draw.Over)

	return nil
}

func (r Config) ExtractResizes(dest image.Point) ([]detection, error) {
	xDetections := make([]int, len(r.XDetections))

	// convert the measurements into integers
	for i, dist := range r.XDetections {
		d, err := dist.CalcOffset(dest.X)

		if err != nil {
			return nil, err
		}

		if d > dest.X {
			return nil, fmt.Errorf("%v resulted in a width of %v, which is greater than the width of the testcard (%v)",
				dist.Dist, d, dest.X)
		}

		xDetections[i] = d
	}

	yDetections := make([]int, len(r.YDetections))

	for i, dist := range r.YDetections {
		d, err := dist.CalcOffset(dest.Y)

		if err != nil {
			return nil, err
		}

		if d > dest.Y {
			return nil, fmt.Errorf("%v resulted in a height of %v, which is greater than the height of the testcard (%v)",
				dist.Dist, d, dest.Y)
		}

		yDetections[i] = d
	}

	// set up the steppers
	// these will always be >0 when the schema is set up
	if r.XStep != nil {

		if r.XStepEnd == nil {
			r.XStepEnd = &parameters.DistanceField{}
		}

		xSteps, err := stepGenerator(*r.XStep, *r.XStepEnd, dest.X)
		if err != nil {
			return nil, err
		}

		if xSteps[0] > dest.X {
			return nil, fmt.Errorf("%v resulted in a width of %v, which is greater than the width of the testcard (%v)",
				r.XStep.Dist, xSteps[0], dest.X)
		}

		xDetections = append(xDetections, xSteps...)
	}

	if r.YStep != nil {
		if r.YStepEnd == nil {
			r.YStepEnd = &parameters.DistanceField{}
		}

		ySteps, err := stepGenerator(*r.YStep, *r.YStepEnd, dest.Y)
		if err != nil {
			return nil, err
		}

		if ySteps[0] > dest.X {
			return nil, fmt.Errorf("%v resulted in a height of %v, which is greater than the height of the testcard (%v)",
				r.YStep.Dist, ySteps[0], dest.Y)
		}

		yDetections = append(yDetections, ySteps...)
	}

	// set up the detections to be handled in in the same loop
	totalBoxes := len(xDetections) + len(yDetections)
	detections := make([]detection, totalBoxes)
	for i, x := range xDetections {
		detections[i] = detection{size: x, direction: "x"}
	}
	for i, y := range yDetections {
		detections[i+len(xDetections)] = detection{size: y, direction: "y"}
	}

	return detections, nil
}

// get the x count and y count of the box layout
func boxSetUp(totalBoxes int, b image.Point) (xBoxCount int, yBoxCount int) {

	xStart := 1
	yStart := totalBoxes + 1

	diff := 0xffff
	squareDiff := float64(0xffff)

	// find which xy fits "best"
	for x := xStart; x <= totalBoxes/2+1; x++ {

		for y := yStart - x; y >= 1; y-- {

			if x*y < totalBoxes {
				continue
			}

			// calculate and see what shape the segment
			// takes, we want a more square one
			xSize := float64(b.X) / float64(x)
			ySize := float64(b.Y) / float64(y)

			// if its portrait then flip the x and the y
			if b.Y < b.X {
				xSize = float64(b.X) / float64(y)
				ySize = float64(b.Y) / float64(x)
			}
			sqDiff := math.Abs(xSize - ySize)

			if x*y-totalBoxes <= diff {
				if sqDiff < squareDiff {
					xBoxCount = x
					yBoxCount = y
					diff = x*y - totalBoxes
					squareDiff = sqDiff
				}

			}
		}
	}

	// set the x and y based on the if the canvas is landscape or portrait
	if b.Y < b.X {
		if xBoxCount < yBoxCount {
			xBoxCountMid := xBoxCount
			xBoxCount = yBoxCount
			yBoxCount = xBoxCountMid
		}
	} else {
		if yBoxCount < xBoxCount {
			xBoxCountMid := xBoxCount
			xBoxCount = yBoxCount
			yBoxCount = xBoxCountMid
		}
	}

	return xBoxCount, yBoxCount
}

// stepGenerator generates the steps of canvas
// to be checked for
func stepGenerator(stepper, endPos parameters.DistanceField, baseSize int) ([]int, error) {

	step, err := stepper.CalcOffset(baseSize)
	if err != nil {
		return nil, err
	}

	ep, err := endPos.CalcOffset(baseSize)

	if err != nil {

	}

	stepCount := int(math.Ceil(float64(baseSize)-float64(ep)) / float64(step))

	steps := make([]int, stepCount)
	for i := 1; i <= stepCount; i++ {
		//	pos := baseSize - step*i
		steps[i-1] = baseSize - step*i

	}

	return steps, nil
}

type pixel struct {
	weight float64
	pixel  int
}

// create coefficients generates the coeffs for a resize
func createCoefficients(destX int, filterLength int, scale float64, kernel func(float64) float64) [][]pixel {

	blur := 1.0
	filterLength = filterLength * int(math.Max(math.Ceil(blur*scale), 1))
	filterFactor := math.Min(1./(blur*scale), 1)

	coeffs := make([][]pixel, destX)
	// for each x pixel we caclulate the coefficients and
	// the source pixel positions it samples from

	for x := 0; x < destX; x++ {

		coeffs[x] = make([]pixel, filterLength)

		// get the source position of the pixe;
		interpX := scale*(float64(x)+0.5) - 0.5
		startX := int(interpX) - filterLength/2 + 1
		interpX -= float64(startX)

		for i := 0; i < filterLength; i++ {
			in := (interpX - float64(i)) * filterFactor
			val := math.Abs((kernel(in) * 65536))

			coeffs[x][i] = pixel{weight: val, pixel: startX + i}
		}
	}

	return coeffs
}

// coeffSolver solves the coefficients of the resizing,
// where that the pixels are values that will collapse to grey.
func coeffSolver(coeffs [][]pixel, width int) ([]uint16, error) {

	// set up the parameters
	destWidth := len(coeffs)
	source := make([]float64, width*destWidth)
	dest := make([]float64, destWidth)

	// 2d matrix of source pixels x dest pixels
	// this is then solved against a
	// matrix  of dest pixels x 1
	// This leaves us with an array
	// of the destination pixels that give us
	// a grey background, when they are transformed.

	for i, coeff := range coeffs {

		// get the total coefficient for this
		// pixel
		coeffTotal := 0.0
		for _, c := range coeff {
			coeffTotal += c.weight
		}

		//
		for _, c := range coeff {
			xi := c.pixel
			if xi < 0 {
				xi = 0
			} else if xi >= width {
				xi = width - 1
			}

			// get the x offset in the 2d array
			offset := xi*destWidth + i

			// update the
			source[offset] += c.weight

		}

		// set the destination as the combination of coefficients
		dest[i] = ((float64((0xB000))) * coeffTotal)

	}
	// lets solve that equation
	sourceM := mat.NewDense(width, destWidth, source)
	destB := mat.NewVecDense(destWidth, dest)

	result := mat.NewDense(width, 1, nil)
	var qr mat.QR
	qr.Factorize(sourceM)
	err := qr.SolveTo(result, true, destB)
	if err != nil {
		return nil, fmt.Errorf("error calculating pixel values %v", err)
	}

	// convert the results to uint16 for making colours
	data := result.RawMatrix().Data
	out := make([]uint16, len(data))
	for i, d := range data {
		out[i] = uint16(d)
	}

	return out, nil
}

// A matrix mask that makes a mask on the matrix.
// values are currently 0 or 1, anything else can lead to unexpected consequences
func matrixMask(bounds image.Rectangle, matrix [][]int) draw.Image {
	xStep := len(matrix[0])
	yStep := len(matrix)
	base := image.NewNRGBA64(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {

		for y := bounds.Min.X; y < bounds.Max.Y; y++ {

			col := uint16(0xffff * matrix[y%yStep][x%xStep])

			base.Set(x, y, color.NRGBA64{A: col})
		}

	}

	return base
}

// lanczos sampling algorithm
func lan(in float64) float64 {
	if in > -3 && in < 3 {
		return sinc(in) * sinc(in*0.3333333333333333)
	}
	return 0
}

func sinc(x float64) float64 {
	x = math.Abs(x) * math.Pi
	if x >= 1.220703e-4 {
		return math.Sin(x) / x
	}
	return 1
}

/*
// cubic sampling algorthim
// saved for use later
func cubic(in float64) float64 {
	in = math.Abs(in)
	if in <= 1 {
		return in*in*(1.5*in-2.5) + 1.0
	}
	if in <= 2 {
		return in*(in*(2.5-0.5*in)-4.0) + 2.0
	}
	return 0
}*/
