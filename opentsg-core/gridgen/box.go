package gridgen

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"math"
	"regexp"
	"strconv"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
)

/*




design ideas

    "grid": {
        "alias": " NotherNameForThisLocation",
        "location": "(200,1700)-(3640,1900)"
    },

	becomes

    "box": {
        "alias": " noiseBox",
        "bounds": {"x":R1, "y":C100, "w":B, "h":3}
    },

	grid coordinates are the default unit of what is used
	explicit coordinate is relative card

    "box": {
        "alias": " noiseBox2",
        "bounds": {"x":5%, "y":20%, "x2":34%, "y2":40%}
    },


	"box": {
        "alias": " noiseBox2",
        "coordinates": {"x":"200px", "y":1700px, "x2":8000px, "y2":4000px}
    },

{"x":200px, "y":1700px, "w":3640px, "h":1900px} implicitly a top-left pinned box because it has 4 properties x, y, w, h
{"x":200, "y":1700, "x2":3640, "y2":1900} implicitly a corner pinned box because it has 4 properties x, y, x2, y2
{"cx":200, "cy":1700, "x2":3640, "y2":1900} do the

{"cx":200, "cy":1700, "radius":20px}

edge antiasliasing questions
inheritance positions questions

format for xy coordinates

it its coordinate then

x,y as pixels. Each value is the grid, no sub grid componenets yet

so 16,16 would then be used. Bin off A1, R1C1?

*/

type Location struct {

	// keep the Alias from last time
	Alias string
	//
	Box Box
}

// implement hsl(0, 100%, 50%);

// keep these ideas in mind https://www.w3schools.com/css/css_boxmodel.asp
/*
remove margin, padding and border as there will not be that mich need

thoughts - sa y the limits so people dont think this is a dierect css import
as fetures will deffo be missing

*/
type Box struct {
	// use a predeclared alias
	// alias must be declared before
	UseAlias   string
	UseGridKey string `json:"useGridKey"`

	// top left coordinates
	// actually any
	X any `json:"x" yaml:"x"`
	Y any `json:"y" yaml:"y"`
	// bottom right
	// if not used then the grid is 1 square
	X2 any `json:"x2" yaml:"x2"`
	Y2 any `json:"y2" yaml:"y2"`

	// width height
	// can they be A or 1 etc. just mix it up
	Width  any `json:"width" yaml:"width"`
	Height any `json:"height" yaml:"width"`

	// centre values
	// width
	XAlignment, YAlignment string // default top left but let them choose
	// or masks like this. Leave masks out for the moment?
	//  mask-image: radial-gradient(circle, black 50%, rgba(0, 0, 0, 0.5) 50%);

	// circle properties
	// border radius - what css uses
	// https://prykhodko.medium.com/css-border-radius-how-does-it-work-bfdf23792ac2
	// taps out at 50% - keep it the simple version to start
	BorderRadius any
}

/*

no shapes apart from square with rounded edges

cx,cy? are these needed when you can still set squares
keep

*/

// https://www.w3schools.com/css/css_boxmodel.asp

/*

pixel, string, percentage or grid for xy
pixel string percentage for height and width

*/

/*
func anyToLength(coordinate any) int {

	coord := string(fmt.Sprintf("%v", coordinate))

	regSpreadX := regexp.MustCompile(`^[a-zA-Z]{1,}$`)
	regCoord := regexp.MustCompile(`^[0-9]{1,}$`)

	regPixels := regexp.MustCompile(`^[0-9]{1,}[Pp][Xx]$`)
	regXY := regexp.MustCompile(`^\(-{0,1}[0-9]{1,5},-{0,1}[0-9]{1,5}\)-\(-{0,1}[0-9]{1,5},-{0,1}[0-9]{1,5}\)$`)
	regRC := regexp.MustCompile(`^[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1})$`)
	regRCArea := regexp.MustCompile(`^[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1}):[Rr]([\d]{2,}|[1-9]{1})[Cc]([\d]{2,}|[1-9]{1})$`)

	switch {
	case true:
	default:
	}

	return 0
}
*/

func (l Location) GridSquareLocatorAndGenerator(c *context.Context) (draw.Image, image.Point, draw.Image, error) {

	alias := core.GetAliasBox(*c)

	if l.Box.UseAlias != "" {
		alias.Mu.Lock()
		item, ok := alias.Data[l.Box.UseAlias]
		alias.Mu.Unlock()

		if ok {
			// just recurse through
			if mid, ok := item.(Location); ok {
				return mid.GridSquareLocatorAndGenerator(c)
			}
		}
	}

	if l.Box.UseGridKey != "" {

		regArt := regexp.MustCompile(`^key:[\w]{3,10}$`)
		if regArt.MatchString(l.Box.UseGridKey) {

			return artToCanvas(l.Box.UseGridKey, c)

		}

	}

	return l.GetAreas(c)
}

func (b Location) GetAreas(c *context.Context) (draw.Image, image.Point, draw.Image, error) {
	if b.Box.X == nil || b.Box.Y == nil {
		//invalid coordiantes recived
	}

	aliasMap := core.GetAliasBox(*c)
	dimensions := (*c).Value(sizekey).(image.Point)
	xUnit := (*c).Value(xkey).(float64)
	yUnit := (*c).Value(ykey).(float64)

	y, err := anyToDist(b.Box.Y, dimensions.Y, yUnit)
	if err != nil {
		return nil, image.Point{}, nil, err
	}

	x, err := anyToDist(b.Box.X, dimensions.X, xUnit)
	if err != nil {
		return nil, image.Point{}, nil, err
	}

	var endY float64

	// switch the width in order of precedence
	switch {
	case b.Box.Y2 != nil:
		endY, err = anyToDist(b.Box.Y2, dimensions.Y, yUnit)
	case b.Box.Height != nil:
		var mid float64
		mid, err = anyToDist(b.Box.Height, dimensions.Y, yUnit)
		endY = y + mid
	default:
		// default is one y unit
		endY = y + yUnit
		// height is one
	}

	if err != nil {
		return nil, image.Point{}, nil, err
	}

	var endX float64

	// switch the width in order of precedence
	switch {
	case b.Box.X2 != nil:
		endX, err = anyToDist(b.Box.X2, dimensions.X, xUnit)
	case b.Box.Width != nil:
		var mid float64
		mid, err = anyToDist(b.Box.Width, dimensions.X, xUnit)
		endX = x + mid
	default:
		// default is one y unit
		endX = x + xUnit
		// height is one
	}

	if err != nil {
		return nil, image.Point{}, nil, err
	}

	width := int(endX) - int(x)
	height := int(endY) - int(y)
	tsgLocation := image.Point{X: int(x), Y: int(y)}

	// get the area that the widget covers

	//ignore the XY coordinate power user
	//	if (((gb.X + generatedGridInfo.X) > maxBounds.X) || (gb.Y+generatedGridInfo.Y) > maxBounds.Y) && !regXY.MatchString(gridString) {
	//
	//			return emptyGrid, fmt.Errorf(errBounds, maxBounds, gb.X+generatedGridInfo.X, gb.Y+generatedGridInfo.Y)
	//	}

	mask := (*c).Value(tilemaskkey)
	var widgMask draw.Image
	if mask != nil {
		mask := mask.(draw.Image)
		widgMask = ImageGenerator(*c, image.Rect(0, 0, width, height))
		colour.Draw(widgMask, widgMask.Bounds(), mask, tsgLocation, draw.Src)
	}

	if b.Box.BorderRadius != nil {

		xSize, dim := xUnit, dimensions.X
		if xSize > yUnit {
			xSize = yUnit
		}

		if dim > dimensions.Y {
			dim = dimensions.Y
		}

		r, err := anyToDist(b.Box.BorderRadius, dim, xSize)
		if err != nil {
			return nil, image.Point{}, nil, err
		}
		midMask := roundedMask(c, image.Rect(0, 0, width, height), int(r))
		if widgMask == nil {
			widgMask = midMask
		} else {
			// mask the tsig mask, with the rounded mask. Only in the bounds of the tsig mask.
			draw.DrawMask(widgMask, widgMask.Bounds(), midMask, image.Point{}, widgMask, image.Point{}, draw.Src)
		} // mask it?

	}

	widgetCanvas := ImageGenerator(*c, image.Rect(0, 0, width, height))

	// log the whole location
	if b.Alias != "" {
		aliasMap.Mu.Lock() // prevent concurrent map writes
		aliasMap.Data[b.Alias] = b
		aliasMap.Mu.Unlock()
	}

	return widgetCanvas, tsgLocation, widgMask, nil
}

func roundedMask(c *context.Context, rect image.Rectangle, radius int) draw.Image {

	base := ImageGenerator(*c, rect)
	draw.Draw(base, base.Bounds(), &image.Uniform{&colour.CNRGBA64{A: 0xffff}}, image.Point{}, draw.Src)

	startPoints := []image.Point{{radius, radius}, {radius, rect.Max.Y - radius},
		{rect.Max.X - radius, radius}, {rect.Max.X - radius, rect.Max.Y - radius}}

	dir := []image.Point{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

	for i, sp := range startPoints {

		for x := 0; x <= radius; x++ {

			for y := 0; y <= radius; y++ {
				//	r := xy
				r := xyToRadius(float64(x), float64(y))
				if r > float64(radius) {
					base.Set(sp.X+(dir[i].X*x), sp.Y+(dir[i].Y*y), &colour.CNRGBA64{})
				}
			}
		}

	}

	return base
}

func xyToRadius(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}

type DistanceField struct {
	Dist any `yaml:",flow"`
}

func (d *DistanceField) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ang any
	err := unmarshal(&ang)
	if err != nil {
		return err
	}

	d.Dist = ang
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

// unit distance
// dimension distance
func anyToDist(a any, dimension int, unitWidth float64) (float64, error) {

	dist := fmt.Sprintf("%v", a)

	pixel := regexp.MustCompile(`^-{0,1}\d{1,}px$`)
	grid := regexp.MustCompile(`^\d{1,}$`)
	pcDefault := regexp.MustCompile(`^-{0,1}\d{0,2}\.{1}\d{0,}%$|^-{0,1}\d{0,2}%$|^-{0,1}(100)%$`)

	/*
		squareX := (*c).Value(xkey).(float64)
		squareY := (*c).Value(ykey).(float64)
	*/

	switch {
	case pixel.MatchString(dist):

		pxDist, err := strconv.Atoi(dist[:len(dist)-2])

		if err != nil {
			err = fmt.Errorf("extracting %s as a integer: %v", dist, err.Error())
			return 0, err
		}
		return float64(pxDist), nil
	case pcDefault.MatchString(dist):

		// trim the %
		dist = dist[:len(dist)-1]

		perc, err := strconv.ParseFloat(dist, 64)

		if err != nil {
			return 0, fmt.Errorf("extracting %s as a percentage : %v", dist, err.Error())
		}
		fmt.Println(perc)
		totalWidth := (perc / 100) * float64(dimension)
		fmt.Println(totalWidth, dimension)
		// @TOOD include the dimensions
		return totalWidth, nil
	case grid.MatchString(dist):
		unit, err := strconv.ParseFloat(dist, 64)
		if err != nil {
			return 0, fmt.Errorf("extracting %s as a percentage : %v", dist, err.Error())

		}
		totalWidth := unit * unitWidth
		// @TOOD include the dimensions
		return totalWidth, nil
	default:
		return 0, fmt.Errorf("unknown coordiante use %s", dist)
	}

}
