// Package gridgen generates the images canvases for the widgets to write to and place on the test card
package gridgen

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/fogleman/gg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
)

type gridContextKey string

const (
	xkey        gridContextKey = "x key holder"
	ykey        gridContextKey = "y key holder"
	sizekey     gridContextKey = "size of the canvas"
	artkey      gridContextKey = "art key holder"
	tpigkey     gridContextKey = "tpig key holder, contains the segements of a tpig"
	tilekey     gridContextKey = "tpig key holder for individual tiles, contains the segements of a tpig"
	gridkey     gridContextKey = "tpig key holder for individual grids, shows what values they contain"
	tilemaskkey gridContextKey = "tpig mask representing the shape of a tpig"
	carvekey    gridContextKey = "contains the carving information for tpigs"
	aliasKeyBox gridContextKey = "contains the alias map for the run of openTSG, but with the new box layout"
	aliasKey    gridContextKey = "contains the alias map for the run of openTSG"
	frameKey    gridContextKey = "log of the frame contents for this run"
)

// FrameConfiguration contains the configuration
// options for each frame.
type FrameConfiguration struct {
	// Rows is how many grid rows
	Rows int
	// Cols is how many grid columns
	Cols int
	// What is the width of the lines
	LineWidth float64
	// FrameSize is the size of the frame
	FrameSize image.Point
	// CanvasType - "ACES" or "" for default
	CanvasType string
	// What Color is the canvas
	CanvasFill color.Color
	// what colour are the grid lines
	LineColour color.Color
	// colour space of the frame
	ColorSpace colour.ColorSpace
	// A path to a tsig file
	Geometry string
	// A path to an image file
	BaseImage string
}

// var rows = canvaswidget.GetGridRows
// var cols = canvaswidget.GetGridColumns
// var getWidth = canvaswidget.GetLWidth

// var size = canvaswidget.GetPictureSize
// var imageType = canvaswidget.GetCanvasType

// Colours
// var getFill = canvaswidget.GetFillColour
// var colourSpaceType = canvaswidget.GetBaseColourSpace

type canvasAndMask struct {
	canvas, mask draw.Image
}

func baseGen(c *context.Context, geomCanvas draw.Image, frame FrameConfiguration) (draw.Image, error) {
	var canvas draw.Image
	cmid := context.WithValue(*c, frameKey, frame)
	if geomCanvas == nil {
		// s := size(*c)
		// based on type do this and use aces as increased fidelity?
		// canvasSize := image.Rect(0, 0, s.X, s.Y)
		canvasSize := image.Rect(0, 0, frame.FrameSize.X, frame.FrameSize.Y)
		canvas = ImageGenerator(*c, canvasSize)
	} else {
		canvas = geomCanvas
	}

	// fillColour := getFill(*c)
	var background color.Color = &colour.CNRGBA64{R: 46080, G: 46080, B: 46080, A: 0xffff}
	if frame.CanvasFill == nil { // check for user defined colours
		background = frame.CanvasFill
		// background = colourgen.ConvertNRGBA64(col)
	}

	colour.Draw(canvas, canvas.Bounds(), &image.Uniform{background}, image.Point{}, draw.Src)
	// make the squares sizes
	//x := cols(*c)
	//y := rows(*c)
	if frame.Cols == 0 || frame.Rows == 0 {
		return canvas, fmt.Errorf("0041 No columns or rows declared, got %v rows and %v columns", frame.Rows, frame.Cols)
	}
	// @TODO make these scale, not be whole numbers
	// make sure the number is a whole number etc
	squareX := float64(canvas.Bounds().Max.X) / float64(frame.Cols)
	squareY := float64(canvas.Bounds().Max.Y) / float64(frame.Rows)
	gridToScale(frame.Cols) // Tell the user the valid list of coordinates, not used anymore
	cmid = context.WithValue(*c, xkey, squareX)
	cmid = context.WithValue(cmid, ykey, squareY)
	cmid = context.WithValue(cmid, sizekey, canvas.Bounds().Max)
	*c = cmid

	splice(c, frame.Cols, frame.Rows, squareX, squareY)

	return canvas, nil
}

// ImageGenerator generates an image based off the configuration type.
func ImageGenerator(c context.Context, canvasSize image.Rectangle) draw.Image {
	frameConfig := (c).Value(frameKey)
	if frameConfig == nil {
		frameConfig = FrameConfiguration{}
	}
	frame := frameConfig.(FrameConfiguration)
	// base := imageType(c)
	if frame.CanvasType == "ACES" {

		return colour.NewARGBA(canvasSize)
	}

	// space := colourSpaceType(c)
	switch frame.ColorSpace {
	case colour.ColorSpace{}:
		// if there's no colour space just use the base go images for performance
		return image.NewNRGBA64(canvasSize)
	default:
		return colour.NewNRGBA64(frame.ColorSpace, canvasSize)
	}

}

// var baser = canvaswidget.GetBaseImage
// var geometry = canvaswidget.GetGeometry

// Gridgen generates the base openTSG image for a frame, drawing the gridlines or
// the specified base image. In both instances the grid coordinate system are calculated
// and drawn as lines.
// If an image has been used for the base then colour locations are also calculated.
func GridGen(c *context.Context, dir string, frame FrameConfiguration) (draw.Image, error) {
	// if tsig
	// geom := geometry(*c)
	var geomImg canvasAndMask

	if frame.Geometry != "" {
		// update the context and produce a mask to draw over the main image
		// get it to generate a base image that supersedes the one given in s? This is then used as a base for the other methods so they can combine
		var err error
		geomImg, err = flatmap(c, dir, frame.Geometry)
		if err != nil {
			return nil, err
		}
	}

	// base := baser(*c)
	if frame.BaseImage != "" {

		return artKeyGen(c, geomImg.canvas, frame.BaseImage, frame)
	}

	return gridGen(c, geomImg, frame)
}

// Gridgen generates a canvas using the information found in the config options
func gridGen(c *context.Context, geomCanvas canvasAndMask, frame FrameConfiguration) (draw.Image, error) {
	canvas, err := baseGen(c, geomCanvas.canvas, frame)
	if err != nil {
		return canvas, err
	}

	squareX := (*c).Value(xkey).(float64)
	squareY := (*c).Value(ykey).(float64)
	// make a grid frame for each generated module
	width := frame.LineWidth
	// gImage := maskGen(squareX, squareY, width, c)
	squares := make(map[image.Point]image.Image)

	// make the squares
	x := 0.0

	for x < float64(canvas.Bounds().Max.X) {
		y := 0.0
		for y < float64(canvas.Bounds().Max.Y) {

			size := image.Point{X: int(x+squareX) - int(x), Y: int(y+squareY) - int(y)}
			gImage, ok := squares[size]
			if !ok {
				gImage = maskGen(size.X, size.Y, width, frame)
				squares[size] = gImage
			}

			colour.Draw(canvas, image.Rect(int(x), int(y), int(x+squareX), int(y+squareY)), gImage, image.Point{}, draw.Over)
			y += squareY
		}
		x += squareX
	}

	// if there is a global mask apply it
	if (geomCanvas != canvasAndMask{}) {
		base := ImageGenerator(*c, canvas.Bounds())
		colour.DrawMask(base, base.Bounds(), geomCanvas.canvas, image.Point{}, geomCanvas.mask, image.Point{}, draw.Src)

		return base, nil
	}

	return canvas, nil
}

func maskGen(maxX, maxY int, width float64, frame FrameConfiguration) image.Image {
	// make a canvas and change it to a gg context with the required set up
	maskTailor := image.NewNRGBA64(image.Rect(0, 0, maxX, maxY))
	// this is automaticall changed to rgb
	cd := gg.NewContextForImage(maskTailor)
	var myBorder color.Color = &colour.CNRGBA64{R: 0, G: 0, B: 0, A: 0xffff}
	// colour := canvaswidget.GetLineColour(*c)
	if frame.LineColour == nil { // check for user defined colours
		myBorder = frame.LineColour
		// myBorder = colourgen.ConvertNRGBA64(col)
	}

	cd.SetColor(myBorder)
	cd.SetLineWidth(width)
	cd.SetLineCapSquare()
	var shift float64
	if width > 1 {
		shift = width / 2

	} else {
		shift = 1 - width
	}
	// shift the y coordinates when drawing horizontally
	cd.DrawLine(0, shift, float64(maxX), shift)
	cd.DrawLine(0, float64(maxY)-shift, float64(maxX), float64(maxY)-shift)
	// shift the x coordinates when drawing vertically
	cd.DrawLine(shift, 0, shift, float64(maxY))
	cd.DrawLine(float64(maxX)-shift, 0, float64(maxX)-shift, float64(maxY))

	cd.Stroke()

	// fix the corners where the lines over run
	if width != math.Trunc(width) {
		depth := int(math.Floor(width))
		iner := cd.Image().At(100, depth)
		cd.SetColor(iner)

		// assign each of the four corners for a square
		cd.SetPixel(depth, depth)
		cd.SetPixel(depth, maxY-1-depth)
		cd.SetPixel(maxX-1-depth, depth)
		cd.SetPixel(maxX-1-depth, maxY-1-depth)
		cd.Stroke()
	}

	return cd.Image()
}

// grid contains all the information for a generated grid
type grid struct {
	GImage draw.Image
	GMask  draw.Image
	X, Y   int
	w, h   int
}

// GridSquareLocatorAndGenerator converts the grid and alias string into a canvas the size of the grid,
// the location generated is the upper left most corner of the grid, along with any masks that are required for non square
// grids.
func GridSquareLocatorAndGenerator(gridString, alias string, c *context.Context) (draw.Image, image.Point, draw.Image, error) {
	regArt := regexp.MustCompile(`^key:[\w]{3,10}$`)

	if regArt.MatchString(gridString) {

		return artToCanvas(gridString, c)

	}

	// regex the grid string to either be looking for the key or to go param to canvas, It now needs to return a mask as well
	mid, err := gridSquareLocatorAndGenerator(gridString, alias, c)
	// extract the mask from the parent mask, if there is one
	return mid.GImage, image.Point{mid.X, mid.Y}, mid.GMask, err
}

// errors as variables
var (
	invalidCoordinates = "0045 The grid dimensions of %v are invalid, received coordinates of (%v,%v)-(%v,%v)"
	invalidAlias       = "0046 %v is not a valid grid alias"
	errBounds          = "0047 Area outside of image bounds of %v, received an x value of %v and a y value of %v"
	invalidGrid        = "0048 %v is not a valid grid string"
)

func gridSquareLocatorAndGenerator(gridString, alias string, c *context.Context) (grid, error) {
	var generatedGridInfo grid
	var emptyGrid grid // for sending back empty info with errors
	gridString = strings.ToLower(gridString)

	regSing := regexp.MustCompile("^[a-zA-Z]{1,3}[0-9]{1,3}$")
	regArea := regexp.MustCompile("^[a-zA-Z]{1,3}[0-9]{1,3}:[a-zA-Z]{1,3}[0-9]{1,3}$")
	regAlias := regexp.MustCompile(`^[\w\W]{1,30}$`)
	regXY := regexp.MustCompile(`^\(-{0,1}[0-9]{1,5},-{0,1}[0-9]{1,5}\)-\(-{0,1}[0-9]{1,5},-{0,1}[0-9]{1,5}\)$`)
	regRC := regexp.MustCompile(`^[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1})$`)
	regRCArea := regexp.MustCompile(`^[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1}):[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1})$`)

	squareX := (*c).Value(xkey).(float64)
	squareY := (*c).Value(ykey).(float64)

	// @TODO insert an offset function here
	/*
		this needs to contain the offset in pre canvas sizes as it could lead to slight offsets
		needs to be in raw coordinates and added to the multiplication. Because of
		the rounding nature of finding the coordinates. Then can apply offset even if its 0 to all
		calculations
	*/

	aliasMap := GetAlias(*c)
	// TODO clean the switch statement by making everything a function of grid
	switch {
	case regSing.MatchString(gridString):
		x, y, err := gridSplit(gridString)
		if err != nil {

			return emptyGrid, err
		}
		// get square locations
		generatedGridInfo.X = int(float64(x) * squareX)
		generatedGridInfo.Y = int(float64(y) * squareY)
		// make a 1x1 square
		generatedGridInfo.w, generatedGridInfo.h = int(float64(x+1)*squareX)-generatedGridInfo.X, int(float64(y+1)*squareY)-generatedGridInfo.Y

		// g.GImage = image.NewNRGBA64(image.Rect(0, 0, squareX, squareY))
	case regArea.MatchString(gridString):
		// gridSplit(gridString) //split it around :
		grids := strings.Split(gridString, ":")
		x, y, err := gridSplit(grids[0])
		if err != nil {

			return emptyGrid, err
		}
		xend, yend, err := gridSplit(grids[1])
		if err != nil {

			return emptyGrid, err
		}

		generatedGridInfo.X = int(float64(x) * squareX)
		generatedGridInfo.Y = int(float64(y) * squareY)
		// make sure the coordinates are in a valid direction
		if xend < x || yend < y {

			return emptyGrid, fmt.Errorf(invalidCoordinates, gridString, x, y, xend, yend)

		}
		generatedGridInfo.w, generatedGridInfo.h = int(float64(xend+1)*squareX)-generatedGridInfo.X, int(float64(yend+1)*squareY)-generatedGridInfo.Y

		// g.GImage = image.NewNRGBA64(image.Rect(0, 0, squareX*(xend-x+1), squareY*(yend-y+1)))
	case regXY.MatchString(gridString):
		// remove surronding brackets and replace
		//gridString = strings.ReplaceAll(gridString, "(", "")
		//gridString = strings.ReplaceAll(gridString, ")", "")
		// split at the mid point
		grid := strings.Split(gridString, ")-(")
		for i := range grid {
			//		fmt.Println(g)
			grid[i] = strings.ReplaceAll(grid[i], "(", "")
			grid[i] = strings.ReplaceAll(grid[i], ")", "")
		}

		x, y, xend, yend, err := pointToVal(grid)

		if err != nil {

			return emptyGrid, err
		}

		// make sure the coordinates are in a valid direction
		if xend < x || yend < y {

			return emptyGrid, fmt.Errorf(invalidCoordinates, gridString, x, y, xend, yend)
		}

		generatedGridInfo.X = x
		generatedGridInfo.Y = y

		generatedGridInfo.w, generatedGridInfo.h = xend-x, yend-y

	case regRC.MatchString(gridString):
		gridString = strings.ToUpper(gridString)
		x, y := 0, 0
		fmt.Sscanf(gridString, "R%dC%d", &x, &y)
		// get square locations
		generatedGridInfo.X = int(float64(x-1) * squareX)
		generatedGridInfo.Y = int(float64(y-1) * squareY)
		// make a 1x1 square
		generatedGridInfo.w, generatedGridInfo.h = int(float64(x)*squareX)-generatedGridInfo.X, int(float64(y)*squareY)-generatedGridInfo.Y

	case regRCArea.MatchString(gridString):

		gridString = strings.ToUpper(gridString)
		xs, ys, xe, ye := 0, 0, 0, 0
		fmt.Sscanf(gridString, "R%dC%d:R%dC%d", &xs, &ys, &xe, &ye)

		if xe < xs || ye < ys {

			return emptyGrid, fmt.Errorf(invalidCoordinates, gridString, xs, ys, xe, ye)
		}
		// get square locations
		generatedGridInfo.X = int(float64(xs-1) * squareX)
		generatedGridInfo.Y = int(float64(ys-1) * squareY)
		// make a 1x1 square
		generatedGridInfo.w, generatedGridInfo.h = int(float64(xe-1)*squareX)-generatedGridInfo.X, int(float64(ye-1)*squareY)-generatedGridInfo.Y
		// squareX*(xe-xs), squareY*(ye-ys)
	case regAlias.MatchString(gridString):
		loc := aliasMap.Data[gridString]
		if loc != "" {
			// call the function again but with the required coordinates
			generatedGridInfo, _ = gridSquareLocatorAndGenerator(loc, "", c)
		} else {

			return emptyGrid, fmt.Errorf(invalidAlias, gridString)
		}

	default:
		// panic("No coordinate system assigned, aborting program")

		return generatedGridInfo, fmt.Errorf(invalidGrid, gridString)
	}

	// generate the image based on the user input to ensure continuity
	generatedGridInfo.GImage = ImageGenerator(*c, image.Rect(0, 0, generatedGridInfo.w, generatedGridInfo.h))

	mask := (*c).Value(tilemaskkey)
	if mask != nil {
		mask := mask.(draw.Image)
		maskdest := ImageGenerator(*c, image.Rect(0, 0, generatedGridInfo.w, generatedGridInfo.h))
		colour.Draw(maskdest, maskdest.Bounds(), mask, image.Point{generatedGridInfo.X, generatedGridInfo.Y}, draw.Src)
		generatedGridInfo.GMask = maskdest
	}

	// add the alias to the map after generation
	if alias != "" {
		aliasMap.Mu.Lock() // prevent concurrent map writes
		aliasMap.Data[alias] = gridString
		aliasMap.Mu.Unlock()
	}

	// check the image fits on the target canvas
	maxBounds := (*c).Value(sizekey).(image.Point)
	gb := generatedGridInfo.GImage.Bounds().Max

	//ignore the XY coordinate power user
	if (((gb.X + generatedGridInfo.X) > maxBounds.X) || (gb.Y+generatedGridInfo.Y) > maxBounds.Y) && !regXY.MatchString(gridString) {

		return emptyGrid, fmt.Errorf(errBounds, maxBounds, gb.X+generatedGridInfo.X, gb.Y+generatedGridInfo.Y)
	}

	return generatedGridInfo, nil
}

func divMod(numerator, denominator int) (int, int) {
	quotient := numerator / denominator // integer division, decimals are truncated
	remainder := numerator % denominator

	if remainder == 0 {

		return quotient - 1, remainder + 26
	}

	return quotient, remainder
}

// gridSplits the letter and number section of the grid coordinates into x,y values
func gridSplit(tile string) (int, int, error) {
	splitPoint := 0
	for i, c := range tile {
		if c < rune('a') || c > rune('z') {
			splitPoint = i

			break
		}
	}

	// base := (len(tile[:splitPoint]) - 1) * 26
	if splitPoint == 0 {
		splitPoint = 1
	}
	x := 0
	// loop through addding the excel values
	for i, val := range tile[:splitPoint] {
		if i == splitPoint-1 { // prevent any a value apart from the alst being counted as 0
			x += int(math.Pow(26, float64(splitPoint-i-1))) * int(val-97)
		} else {
			x += int(math.Pow(26, float64(splitPoint-i-1))) * int(val-96)
		}

	}

	// x = base + xpos(rune(tile[splitPoint-1]))

	y, err := strconv.Atoi(tile[splitPoint:])
	if err != nil {

		return 0, 0, err
	}

	return x, y, nil
}

// pointToVal converts the grid strings to xy coordinates
func pointToVal(grid []string) (int, int, int, int, error) {
	if len(grid) != 2 {

		return 0, 0, 0, 0, fmt.Errorf("error invalid coordinates of %v", grid)
	}
	xy := strings.Split(grid[0], ",")
	xyend := strings.Split(grid[1], ",")
	results := make([]int, 4)
	checks := [][]string{xy, xyend}

	for i, points := range checks {
		for j, point := range points {
			val, err := strconv.Atoi(point)
			if err != nil {
				return 0, 0, 0, 0, err
			}
			results[i*2+j] = val
		}
	}

	return results[0], results[1], results[2], results[3], nil
}

// SyncMap  is a map with a sync.Mutex to prevent concurrent writes.
type SyncMap struct {
	Data map[string]string
	Mu   *sync.Mutex
}

// PutAlias inits a map of the alias in a context
func PutAlias(c context.Context) context.Context {
	n := SyncMap{make(map[string]string), &sync.Mutex{}}

	return context.WithValue(c, aliasKey, n)
}

// Get alias returns a map of the locations alias and their grid positions.
func GetAlias(c context.Context) SyncMap {
	Alias := c.Value(aliasKey)
	if Alias != nil {

		return Alias.(SyncMap)
	}
	// else return an empty map
	var newmu sync.Mutex

	return SyncMap{Mu: &newmu, Data: make(map[string]string)}
}
