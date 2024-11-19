package gridgen

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"math"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
)

/*
Location is the location information of a widget.

It contains the box properties in css style for drawing shapes.
*/
type Location struct {

	// What is the alias for this location
	Alias string `json:"alias" yaml:"alias"`
	// CSS style fields for drawing the box
	Box Box `json:"box,omitempty" yaml:"box,omitempty"`
}

/*
Box implements the box properties of a widget.
The XY coordinates start at 0,0 in the top left of the canvas.

It tries to adhere to css style of a box https://www.w3schools.com/css/css_boxmodel.asp
Not every property is implemented however.
*/
type Box struct {
	// use a predeclared alias
	// alias must be declared before
	UseAlias   string `json:"useAlias,omitempty" yaml:"useAlias,omitempty"`
	UseGridKey string `json:"useGridKey,omitempty" yaml:"useGridKey,omitempty"`

	// top left coordinates
	X any `json:"x,omitempty" yaml:"x,omitempty"`
	Y any `json:"y,omitempty" yaml:"y,omitempty"`
	// bottom right
	// if not used then the grid is 1 square
	X2 any `json:"x2,omitempty" yaml:"x2,omitempty"`
	Y2 any `json:"y2,omitempty" yaml:"y2,omitempty"`

	// Width and Height take second priority to
	// x2 ans y2 if they are called
	Width  any `json:"width,omitempty" yaml:"width,omitempty"`
	Height any `json:"height,omitempty" yaml:"height,omitempty"`

	// xAlignment and yAlignemnt values are not implemented yet
	XAlignment string `json:"xAlignment,omitempty" yaml:"xAlignment,omitempty"`
	YAlignment string `json:"yAlignment,omitempty" yaml:"yAlignment,omitempty"` // default top left but let them choose

	// border radius - follows the simple layout of
	// https://prykhodko.medium.com/css-border-radius-how-does-it-work-bfdf23792ac2
	// taps out at 50% of the shortest dimension.
	BorderRadius any `json:"border-radius,omitempty" yaml:"border-radius,omitempty"`
}

// InitAliasBox inits a map of the alias for handlers in a context
func InitAliasBox(c context.Context) context.Context {
	n := SyncMapBox{make(map[string]any), &sync.Mutex{}}

	return context.WithValue(c, aliasKeyBox, n)
}

// SyncMap  is a map with a sync.Mutex to prevent concurrent writes.
type SyncMapBox struct {
	Data map[string]any
	Mu   *sync.Mutex
}

// Get alias returns a map of the locations alias and their grid positions.
func GetAliasBox(c context.Context) SyncMapBox {
	Alias := c.Value(aliasKeyBox)
	if Alias != nil {

		return Alias.(SyncMapBox)
	}
	// else return an empty map
	var newmu sync.Mutex

	return SyncMapBox{Mu: &newmu, Data: make(map[string]any)}
}

// GridSquareLocatorAndGenerator converts the box struct into a canvas the size of the grid,
// the location generated is the upper left most corner of the grid, along with any masks that are required for non square
// grids.
func (l Location) GridSquareLocatorAndGenerator(c *context.Context) (draw.Image, image.Point, draw.Image, error) {

	alias := GetAliasBox(*c)

	if l.Box.UseAlias != "" {
		alias.Mu.Lock()
		item, ok := alias.Data[l.Box.UseAlias]
		alias.Mu.Unlock()

		if ok {
			// just recurse through
			if mid, ok := item.(Location); ok {
				return mid.GridSquareLocatorAndGenerator(c)
			}
		} else {
			return nil, image.Point{}, nil, fmt.Errorf("\"%s\" is not a valid grid alias", l.Box.UseAlias)
		}
	}

	if l.Box.UseGridKey != "" {

		regArt := regexp.MustCompile(`^key:[\w]{3,10}$`)
		if regArt.MatchString(l.Box.UseGridKey) {

			return artToCanvas(l.Box.UseGridKey, c)

		}

	}

	return l.gridSquareLocatorAndGenerator(c)
}

// gridSquareLocatorAndGenerator
func (b Location) gridSquareLocatorAndGenerator(c *context.Context) (draw.Image, image.Point, draw.Image, error) {

	dims, tsgLocation, err := b.CalcArea(c)

	if err != nil {
		return nil, image.Point{}, nil, err
	}

	mask := (*c).Value(tilemaskkey)
	var widgMask draw.Image
	if mask != nil {
		mask := mask.(draw.Image)
		widgMask = ImageGenerator(*c, dims)
		colour.Draw(widgMask, widgMask.Bounds(), mask, tsgLocation, draw.Src)
	}

	xUnit := (*c).Value(xkey).(float64)
	yUnit := (*c).Value(ykey).(float64)
	if b.Box.BorderRadius != nil {

		xSize, dim := xUnit, dims.Max.X
		if xSize > yUnit {
			xSize = yUnit
		}

		if dim > dims.Max.Y {
			dim = dims.Max.Y
		}

		rad, err := anyToDist(b.Box.BorderRadius, dim, xSize)
		r := int(rad)
		if err != nil {
			return nil, image.Point{}, nil, err
		}

		if r > dims.Max.X/2 {
			r = dims.Max.X / 2
		}

		if r > dims.Max.Y/2 {
			r = dims.Max.Y / 2
		}

		midMask := roundedMask(c, dims, int(r))
		if widgMask == nil {
			widgMask = midMask
		} else {
			// mask the tsig mask, with the rounded mask. Only in the bounds of the tsig mask.
			draw.DrawMask(widgMask, widgMask.Bounds(), midMask, image.Point{}, widgMask, image.Point{}, draw.Src)
		} // mask it?

	}

	widgetCanvas := ImageGenerator(*c, dims)

	// log the whole location
	if b.Alias != "" {
		aliasMap := GetAliasBox(*c)
		aliasMap.Mu.Lock() // prevent concurrent map writes
		aliasMap.Data[b.Alias] = b
		aliasMap.Mu.Unlock()
	}

	return widgetCanvas, tsgLocation, widgMask, nil
}

// CalcArea calculates the dimension of the box and the coordinate the top left is placed at.
func (l Location) CalcArea(c *context.Context) (image.Rectangle, image.Point, error) {
	if l.Box.X == nil || l.Box.Y == nil {
		//invalid coordinates received
		return image.Rectangle{}, image.Point{}, fmt.Errorf("invalid coordinates of x %v and y %v received", l.Box.X, l.Box.Y)
	}

	dimensions := (*c).Value(sizekey).(image.Point)
	xUnit := (*c).Value(xkey).(float64)
	yUnit := (*c).Value(ykey).(float64)

	y, err := anyToDist(l.Box.Y, dimensions.Y, yUnit)
	if err != nil {
		return image.Rectangle{}, image.Point{}, err
	}

	x, err := anyToDist(l.Box.X, dimensions.X, xUnit)
	if err != nil {
		return image.Rectangle{}, image.Point{}, err
	}

	var endY float64

	// switch the width in order of precedence
	switch {
	case l.Box.Y2 != nil:
		endY, err = anyToDist(l.Box.Y2, dimensions.Y, yUnit)
	case l.Box.Height != nil:
		var mid float64
		mid, err = anyToDist(l.Box.Height, dimensions.Y, yUnit)
		endY = y + mid
	default:
		// default is one y unit
		endY = y + yUnit
		// height is one
	}

	if err != nil {
		return image.Rectangle{}, image.Point{}, err
	}

	var endX float64

	// switch the width in order of precedence
	switch {
	case l.Box.X2 != nil:
		endX, err = anyToDist(l.Box.X2, dimensions.X, xUnit)
	case l.Box.Width != nil:
		var mid float64
		mid, err = anyToDist(l.Box.Width, dimensions.X, xUnit)
		endX = x + mid
	default:
		// default is one y unit
		endX = x + xUnit
		// height is one
	}

	if err != nil {
		return image.Rectangle{}, image.Point{}, err
	}

	width := int(endX) - int(x)
	height := int(endY) - int(y)
	tsgLocation := image.Point{X: int(x), Y: int(y)}

	return image.Rect(0, 0, width, height), tsgLocation, nil
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
		// fmt.Println(perc)
		totalWidth := (perc / 100) * float64(dimension)
		//	fmt.Println(totalWidth, dimension)
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
		return 0, fmt.Errorf("unknown coordinate use %s", dist)
	}

}

// Get GridGeometry returns the geometry of a grid coordiante, localised to those grid coordinates.
// This is for use with widgets that utilise geometry.
// if not tsigs were called in the set up phase then a tsig that
// fills the area is given
func (l Location) GetGridGeometry(c *context.Context) ([]Segmenter, error) {

	bounds, tsgPoint, err := l.CalcArea(c)
	if err != nil {

		return nil, err
	}

	// set the bounds to cover the area
	bounds = bounds.Add(tsgPoint)
	geometry := (*c).Value(tilekey)

	var geometryHolder []*Segmenter
	if geometry != nil {
		geometryHolder = geometry.([]*Segmenter)
	} else {
		return []Segmenter{{Shape: bounds, Tags: []string{"autogenerated"}}}, nil
	}
	matches := []*Segmenter{}
	for _, g := range geometryHolder {
		if bounds.Overlaps(g.Shape) {
			matches = append(matches, g)
		}
	}

	offsetMatches := make([]Segmenter, len(matches))

	for i, match := range matches {
		mid := *match
		// go for negative offsets
		mid.Shape = mid.Shape.Add(image.Point{-tsgPoint.X, -tsgPoint.Y})
		offsetMatches[i] = mid
	}

	sort.Slice(offsetMatches, func(i, j int) bool {
		return offsetMatches[i].ImportPosition < offsetMatches[j].ImportPosition
	})

	return offsetMatches, err

}
