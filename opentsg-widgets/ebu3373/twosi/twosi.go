// package twosi generates the ebu3373 two sample interleave text
package twosi

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/text"
)

const (
	WidgetType = "builtin.ebu3373/twosi"
)

// Colour "constants"
var (
	grey       = colour.CNRGBA64{R: 26496, G: 26496, B: 26496, A: 0xffff}
	letterFill = colour.CNRGBA64{R: 41470, G: 41470, B: 41470, A: 0xffff}
	// letterFill = colour.CNRGBA64{R: 41470, G: 0, B: 0, A: 0xffff}
)

// Each abcd channel follows this format
type channel struct {
	yOff, xOff int
	Letter     string
	mask       draw.Image
}

func (t Config) Handle(resp tsg.Response, req *tsg.Request) {
	// Kick off with filling it all in as grey
	backFill := grey
	backFill.UpdateColorSpace(req.PatchProperties.ColourSpace)
	colour.Draw(resp.BaseImage(), resp.BaseImage().Bounds(), &image.Uniform{&backFill}, image.Point{}, draw.Src)

	// Flexible option to get figure out where the image is to be placed
	// this then adds an offset to the genertaed image so it all lines up.

	canvasLocation := req.PatchProperties.TSGLocation

	// Apply the offset
	xOff := -canvasLocation.X % 4
	yOff := -canvasLocation.Y % 4

	b := resp.BaseImage().Bounds().Max

	if b.In(image.Rect(0, 0, 600, 300)) { // Minimum size box we are going with
		resp.Write(tsg.WidgetError, fmt.Sprintf("0171 the minimum size is 600 by 300, received an image of %v by %v", b.X, b.Y))
		return
	}

	// Calculate relevant scale here
	// Relevant to 1510 and 600 (the size of ebu 3373)
	xScale := float64(b.X) / 1510.0
	yScale := float64(b.Y) / 600.0

	letterSize := aPos(int(math.Round(72 * xScale)))

	// Get the title font to be used

	connections := make(map[string]channel)
	connections["A"] = channel{yOff: 0, xOff: 0, Letter: "A"}
	connections["B"] = channel{yOff: 0, xOff: 2, Letter: "B"}
	connections["C"] = channel{yOff: 1, xOff: 0, Letter: "C"}
	connections["D"] = channel{yOff: 1, xOff: 2, Letter: "D"}

	letterColour := letterFill
	letterColour.UpdateColorSpace(req.PatchProperties.ColourSpace)

	// Generate the letter that is only relevant to its channel
	for k, v := range connections {
		// Generate the mask and the canvas
		mid := mask(letterSize, letterSize, v.xOff, v.yOff)
		v.mask = req.GenerateSubImage(resp.BaseImage(), image.Rect(0, 0, letterSize, letterSize))

		// generate a textbox of A
		txtBox := text.NewTextboxer(req.PatchProperties.ColourSpace,
			text.WithFont(text.FontTitle),
			text.WithFill(text.FillTypeFull),
			text.WithTextColour(&letterFill),
		)

		txtBox.DrawString(v.mask, nil, v.Letter)

		colour.DrawMask(v.mask, v.mask.Bounds(), v.mask, image.Point{}, mid, image.Point{}, draw.Src)
		connections[k] = v
	}

	letterOrder := [][2]string{{"A", "B"}, {"A", "C"}, {"A", "D"}, {"B", "C"}, {"B", "D"}, {"C", "D"}}

	xLength := aPos(int(164 * xScale))
	yDepth := aPos(int(164 * yScale))

	lineOff := 24

	letterGap := aPos(int(24 * xScale))
	channelGap := aPos(int(48 * xScale))
	objectWidth := (letterSize*2+letterGap)*6 + 5*channelGap
	startPoint := (b.X - objectWidth) / 2

	// Check start point for being in a  "A" channel start position and configure the numbers so everything lines up

	objectHeight := (letterSize + lineOff + yDepth)
	yStart := aPos((b.Y-objectHeight)/2) + yOff

	if yStart < 0 || startPoint < 0 { // 0 means they're outside the box
		resp.Write(tsg.WidgetError, fmt.Sprintf("0172 irregular sized box, the two sample interleave pattern will not fit within the constraints of %v, %v", b.X, b.Y))
		return
	}

	// If either of these are negative just error and leave the or return a gray canvas? Consult Bruce
	/*	letterSize, startPoint,
		lineOff, xOff, yOff,
		yDepth, xLength int
		yScale float64*/

	letterProperties := letterMetrics{letterSize: letterSize, startPoint: startPoint,
		yOff: yOff, yScale: yScale, yDepth: yDepth,
		xOff: xOff, xLength: xLength, lineOff: lineOff}

	letterProperties.letterDrawer(resp.BaseImage(), letterColour, letterOrder, connections, letterGap, channelGap, yStart)

	resp.Write(tsg.WidgetSuccess, "success")
}

type letterMetrics struct {
	letterSize, startPoint,
	lineOff, xOff, yOff,
	yDepth, xLength int
	yScale float64
}

// letterdrawer loops through the letters and lines drawing them on the canvas
// moving horizontally along each time when drawing a letter
func (lm letterMetrics) letterDrawer(canvas draw.Image, letterColour colour.CNRGBA64, letterOrder [][2]string, connections map[string]channel, letterGap, channelGap, yStart int) {

	position := aPos(lm.startPoint) + lm.xOff
	realY := aPos(yStart + lm.letterSize + lm.lineOff) // Y start for some of the lines
	// make a struct of all the information and give this as apointer
	for _, letter := range letterOrder { // Draw the lines for every letter combination
		// through the three types of lines drawing where required
		left := connections[letter[0]]
		right := connections[letter[1]]

		if left.xOff != right.xOff {
			// then draw the vertical lines
			verticalLines(canvas, letterColour, left, right, position, realY, lm.yDepth, lm.lineOff, lm.xOff, lm.yOff)
		}

		if left.yOff != right.yOff {
			// draw the horizontal lines
			horizontalLines(canvas, letterColour, left, right, position, realY, lm.xLength, lm.xOff, lm.yOff)
		}

		// Draw diagonal lines regardless of the offsets
		diagonalLines(canvas, letterColour, left, right, position, realY, lm.xLength, lm.yDepth, lm.lineOff, lm.yScale, lm.xOff, lm.yOff)

		// draw the letters last
		// @TODO fix the draw locations so there isn't as much empty drawing
		colour.Draw(canvas, canvas.Bounds(), left.mask, image.Point{-position, -yStart}, draw.Over)
		position += lm.letterSize + letterGap // 72+24

		colour.Draw(canvas, canvas.Bounds(), right.mask, image.Point{-position, -yStart}, draw.Over)
		position += lm.letterSize + channelGap // 72+48

	}
}

func verticalLines(canvas draw.Image, letterColour colour.CNRGBA64, left, right channel, position, realY, yDepth, lineOff int, xoff, yoff int) {
	relativePos := position / 4
	leftShift := 1
	rightShift := 0
	if left.xOff > right.xOff { // Reverse the shifts for the one instance B C channel is used
		leftShift = 0
		rightShift = 1
	}

	for y := realY + lineOff; y < realY+yDepth; y += 2 { // Set the x positions all along y

		canvas.Set(4*relativePos+leftShift+left.xOff+xoff, y+left.yOff+yoff, &letterColour)
		canvas.Set(4*relativePos+8+leftShift+left.xOff+xoff, y+left.yOff+yoff, &letterColour)
		canvas.Set(4*relativePos+8+rightShift+right.xOff+xoff, y+right.yOff+yoff, &letterColour)
		canvas.Set(4*relativePos+16+rightShift+right.xOff+xoff, y+right.yOff+yoff, &letterColour)

	}
}

func horizontalLines(canvas draw.Image, letterColour colour.CNRGBA64, left, right channel, startPosition, realY, xLength, xoff, yoff int) {
	m := (realY) / 2
	ys := []int{2*m + left.yOff, 2*m + 6 + right.yOff, 2*m + 8 + left.yOff, 2*m + 14 + right.yOff}
	offsets := []int{left.xOff, right.xOff, left.xOff, right.xOff}
	// Draw each line along the Y
	for i, y := range ys {
		for x := (startPosition + 6) / 4; x < (startPosition+xLength)/4; x++ {

			canvas.Set(4*x+offsets[i]+xoff, y+yoff, &letterColour)
			canvas.Set(4*x+offsets[i]+1+xoff, y+yoff, &letterColour)

		}
	}
}

func diagonalLines(canvas draw.Image, letterColour colour.CNRGBA64, left, right channel, position, realY, xLength, yDepth, lineOff int, yScale float64, xoff, yoff int) {
	pos := (position + xLength) / 4
	if (pos-left.xOff)%4 != 0 {
		pos += 4 - ((pos - left.xOff) % 4)
	}
	ystart := left.yOff
	count := 0
	max := position + xLength
	min := position + 40
	// - int(40*yScale)

	for y := realY + lineOff; y < realY+yDepth; y += 2 {
		count++
		xshift := 1 + left.xOff
		yshift := 0

		if count%2 == 0 {
			xshift = 10 + left.xOff
			// amend positioning for
			if right.yOff == left.yOff {
				xshift++
			} else {
				yshift = 1
			}
		}

		if right.xOff == left.xOff {
			xshift--
		}
		x := []int{4*pos + xshift, 4*pos + xshift + 12}

		// Place the x position
		for _, xp := range x {
			if xp > min && xp < max && (ystart+y+yshift) > realY+lineOff+int(12*yScale) {

				canvas.Set(xp+xoff, ystart+y+yshift+yoff, &letterColour)
			}
		}

		if count%2 == 0 {
			pos--
		}
	}
}

func aPos(p int) int {
	// Basic loop increasing until it is an A channel start int
	for {
		if p%4 == 0 {

			return p
		}
		p++
	}
}

func mask(x, y, xOff, yOff int) *image.NRGBA64 {
	img := image.NewNRGBA64(image.Rect(0, 0, x, y))
	b := img.Bounds().Max

	for m := 0; m <= (b.X / 4); m++ {
		for n := 0; n <= (b.Y / 2); n++ {
			img.SetNRGBA64(4*m+xOff, 2*n+yOff, color.NRGBA64{0, 0, 0, 0xffff})
			img.SetNRGBA64(4*m+xOff+1, 2*n+yOff, color.NRGBA64{0, 0, 0, 0xffff})
		}
	}

	return img
}

func maskOffset(x, y, startX, startY, lettXOff, lettYOff int) *image.NRGBA64 {
	img := image.NewNRGBA64(image.Rect(0, 0, x, y))
	b := img.Bounds().Max

	for m := startX; m <= (b.X); m += 4 {
		for n := startY; n <= (b.Y); n += 2 {

			img.SetNRGBA64(m+lettXOff, n+lettYOff, color.NRGBA64{0, 0, 0, 0xffff})
			img.SetNRGBA64(m+lettXOff+1, n+lettYOff, color.NRGBA64{0, 0, 0, 0xffff})
		}
	}

	return img
}
