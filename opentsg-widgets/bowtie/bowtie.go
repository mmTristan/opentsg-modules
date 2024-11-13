package bowtie

import (
	"context"
	"fmt"
	"image/draw"
	"math"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colourgen"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
)

const (
	WidgetType = "builtin.bowtie"
)

// SwirlGen takes a canvas and then returns an image of the swirl layered on top of the image/
// the angular rotation is set by the frame number, to ensure the image rotates.
func SwirlGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	opts := []any{c}
	conf := widgethandler.GenConf[Config]{Debug: debug, Schema: Schema, WidgetType: WidgetType, ExtraOpt: opts}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

var (
	defaultColours = []*colour.CNRGBA64{{A: 0xffff}, {R: 0xffff, G: 0xffff, B: 0xffff, A: 0xffff}}
	framePos       = core.GetFramePosition
)

type segment struct {
	colourPos  int
	angleValid func(float64) bool

	startAng, endAng float64
	angStep          float64
	startN, endN     int
}

func (c Config) Handle(resp tsg.Response, req *tsg.Request) {

	if c.SegementCount < 4 {
		resp.Write(tsg.WidgetError, fmt.Sprintf("0DEV 4 or more segments required, received %v", c.SegementCount))
		return
	}

	var colours []*colour.CNRGBA64

	if len(c.SegmentColours) == 0 {
		colours = defaultColours
	} else {
		// let the users declar their own colours
		colours = make([]*colour.CNRGBA64, len(c.SegmentColours))

		for i, cstring := range c.SegmentColours {
			col := colourgen.HexToColour(cstring, req.PatchProperties.ColourSpace)
			colours[i] = col
		}

	}

	bounds := resp.BaseImage().Bounds().Max

	angleStep := (2 * math.Pi) / float64(c.SegementCount)

	ang, err := c.ClockwiseRotationAngle()
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	// get the frame count to get the angular rotation
	frame := req.FrameProperties.FrameNumber

	startAng, err := c.GetStartAngle()
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	startPoint := (math.Pi * 2) - (ang * float64(frame)) - startAng
	// reset the angle to be as close to 2pi as possible.
	// these steps do not change the angle of rotation
	for startPoint < 0 {
		startPoint += (2 * math.Pi)
	}

	segments := make([]segment, c.SegementCount)

	for i := 0; i < c.SegementCount; i++ {
		endAng := startPoint - angleStep

		// make sure the start points are always positive
		for startPoint < 0 {
			startPoint += (2 * math.Pi)
		}

		// make it the start point to stop any
		// float issues meaning a line of angles are missed
		if i == c.SegementCount-1 {

			endAng = (math.Pi * 2) - (ang * float64(frame)) - startAng
		}

		for endAng < 0 {
			endAng += (2 * math.Pi)
		}
		// set the start point for the compare fucntions
		funcStart := startPoint

		if endAng < startPoint {
			segments[i] = segment{colourPos: i % len(colours),
				startAng: startPoint,
				endAng:   endAng,
				angStep:  angleStep,
				startN:   (i + 1%len(colours) + len(colours)) % len(colours),
				endN:     (i - 1%len(colours) + len(colours)) % len(colours),
				angleValid: func(ang float64) bool {

					return (ang < funcStart) && (ang >= endAng)
				},
			}
		} else {
			segments[i] = segment{colourPos: i % len(colours),
				startAng: startPoint,
				endAng:   endAng,
				angStep:  angleStep,
				startN:   (i + 1%len(colours) + len(colours)) % len(colours),
				endN:     (i - 1%len(colours) + len(colours)) % len(colours),
				angleValid: func(ang float64) bool {
					return (ang <= funcStart) || (ang >= endAng)
				},
			}
		}
		startPoint = endAng

	}

	out, err := c.CalcOffset(bounds)
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	fill(resp.BaseImage(), colours, segments, float64(bounds.X)/2+float64(out.X), float64(bounds.Y)/2+float64(out.Y), c.Blend)

	resp.Write(tsg.WidgetSuccess, "success")
}

func (s Config) Generate(canvas draw.Image, opts ...any) error {

	c, ok := opts[0].(*context.Context)
	if !ok {
		return fmt.Errorf("0155 configuration error when assigning swirl context")
	}

	if s.SegementCount < 4 {
		return fmt.Errorf("0DEV 4 or more segments required, received %v", s.SegementCount)
	}

	if s.ColourSpace == nil {
		s.ColourSpace = &colour.ColorSpace{}
	}

	var colours []*colour.CNRGBA64

	if len(s.SegmentColours) == 0 {
		colours = defaultColours
	} else {
		// let the users declar their own colours
		colours = make([]*colour.CNRGBA64, len(s.SegmentColours))

		for i, cstring := range s.SegmentColours {
			col := colourgen.HexToColour(cstring, *s.ColourSpace)
			colours[i] = col
		}

	}

	bounds := canvas.Bounds().Max

	angleStep := (2 * math.Pi) / float64(s.SegementCount)

	ang, err := s.ClockwiseRotationAngle()
	if err != nil {
		return fmt.Errorf("0DEV error calculating the rotational angle %v", err)
	}

	// get the frame count to get the angular rotation
	frame := framePos(*c)

	startAng, err := s.GetStartAngle()
	if err != nil {
		return fmt.Errorf("0DEV error calculating the startl angle %v", err)
	}

	startPoint := (math.Pi * 2) - (ang * float64(frame)) - startAng
	// reset the angle to be as close to 2pi as possible.
	// these steps do not change the angle of rotation
	for startPoint < 0 {
		startPoint += (2 * math.Pi)
	}

	segments := make([]segment, s.SegementCount)

	for i := 0; i < s.SegementCount; i++ {
		endAng := startPoint - angleStep

		// make sure the start points are always positive
		for startPoint < 0 {
			startPoint += (2 * math.Pi)
		}

		// make it the start point to stop any
		// float issues meaning a line of angles are missed
		if i == s.SegementCount-1 {

			endAng = (math.Pi * 2) - (ang * float64(frame)) - startAng
		}

		for endAng < 0 {
			endAng += (2 * math.Pi)
		}
		// set the start point for the compare fucntions
		funcStart := startPoint

		if endAng < startPoint {
			segments[i] = segment{colourPos: i % len(colours),
				startAng: startPoint,
				endAng:   endAng,
				angStep:  angleStep,
				startN:   (i + 1%len(colours) + len(colours)) % len(colours),
				endN:     (i - 1%len(colours) + len(colours)) % len(colours),
				angleValid: func(ang float64) bool {

					return (ang < funcStart) && (ang >= endAng)
				},
			}
		} else {
			segments[i] = segment{colourPos: i % len(colours),
				startAng: startPoint,
				endAng:   endAng,
				angStep:  angleStep,
				startN:   (i + 1%len(colours) + len(colours)) % len(colours),
				endN:     (i - 1%len(colours) + len(colours)) % len(colours),
				angleValid: func(ang float64) bool {
					return (ang <= funcStart) || (ang >= endAng)
				},
			}
		}
		startPoint = endAng

	}

	out, _ := s.CalcOffset(bounds)

	fill(canvas, colours, segments, float64(bounds.X)/2+float64(out.X), float64(bounds.Y)/2+float64(out.Y), s.Blend)

	return nil
}

func fill(canvas draw.Image, colours []*colour.CNRGBA64, segments []segment, originX, originY float64, blend string) {

	// origin to the right - take away the distance from the edge
	// offset positive
	// else add it

	if originX > 0 {
		originX *= -1
	}

	if originY < 0 {
		originY *= -1
	}

	b := canvas.Bounds().Max
	// go through every pixel
	for x := 0.0; x < float64(b.X); x++ {

		for y := 0.0; y < float64(b.Y); y++ {
			ang := xyToAngle(x+originX, float64(b.Y)-y-originY)
			// find the segment
			// get the angle offset
			//	fmt.Println(ang)
			colourPos := 0
			var segm segment
			for _, seg := range segments {

				if seg.angleValid(ang) {

					segm = seg
					break
				}
			}

			colourPos = segm.colourPos
			// set the blur function here
			switch blend {
			case "sin":
				// blend the neighbouring colours

				leftDiff := ang - segm.endAng
				rightDiff := segm.startAng - ang

				if rightDiff < 0 {
					rightDiff += 2 * math.Pi
				}

				if leftDiff < 0 {
					leftDiff += 2 * math.Pi
				}

				neighCol := segm.startN
				if rightDiff < leftDiff {
					leftDiff = rightDiff
					neighCol = segm.endN
				}

				/*
					start at 0.5 which is at pi/6
					then add the angle step as a percentage * pi/3
					to get to the max of pi/2 (which is 1 as a sin function)
				*/
				blendOrder := math.Sin((math.Pi / (6)) + (math.Pi/3)*((2*leftDiff)/segm.angStep)) // difference from neighbour

				//blendOrder := math.Sin((math.Pi * diff) / segm.angStep)
				oc := colours[colourPos]
				nc := colours[neighCol]

				newColour := colour.CNRGBA64{
					R: blender(oc.R, nc.R, blendOrder),
					G: blender(oc.G, nc.G, blendOrder),
					B: blender(oc.B, nc.B, blendOrder),
					A: 0xffff,
				}

				canvas.Set(int(x), int(y), &newColour)
			default:
				canvas.Set(int(x), int(y), colours[colourPos])
			}
			// get segment

		}
	}
	/*
		for each x,y make relative to the origin.

		Get the r and angle.

		find which segment it falls in
		find segment by flooring
	*/

}

func blender(in, neigh uint16, blendStrength float64) uint16 {
	return uint16(float64(in)*blendStrength + float64(neigh)*(1-blendStrength))
}

// s = rcosphi
// y = r sin phi
func xyToAngle(x, y float64) float64 {

	ang := math.Atan2(y, x)

	// add 2 pi by the inverse to keep the angle
	// incrementinh
	if ang < 0 {
		return ang + math.Pi*2
	}

	return ang
}

// r = sqrt(x2 +y2)
