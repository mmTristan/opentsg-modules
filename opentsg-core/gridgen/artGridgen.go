package gridgen

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"regexp"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/credentials"
	"github.com/nfnt/resize"
	"golang.org/x/image/tiff"
)

// artkeyGen set the background of the canvas as the inputted image and generate a map of keys
// that provide the points of the transparent areas with keys.
func artKeyGen(c *context.Context, geomCanvas draw.Image, base string, frame FrameConfiguration) (draw.Image, error) {
	// make the canvas to the user specification
	canvas, err := baseGen(c, geomCanvas, frame)
	if err != nil {
		return canvas, err
	}
	// extract the NRGBA64 image scaled to the canvas
	baseImg, err := keyGen(c, base, canvas.Bounds().Max)
	if err != nil {
		return canvas, err
	}

	// add the image and then extract the locations
	colour.Draw(canvas, canvas.Bounds(), baseImg, image.Point{}, draw.Src)
	keys, err := imageToKeyMap(baseImg, c)
	if err != nil {
		return canvas, err
	}
	// update with the map context
	cmid := context.WithValue(*c, artkey, keys)
	*c = cmid

	return canvas, nil
}

// art to canvas takes  a key and provides the image point, image and mask for a widget to use
func artToCanvas(tag string, c *context.Context) (draw.Image, image.Point, draw.Image, error) {
	// check the keymap has been made
	key := (*c).Value(artkey)
	if key == nil {
		return nil, image.Point{}, nil, fmt.Errorf("0049 no background image with keys has been provided")
	}
	keys := key.(map[string]artGrid)
	location := keys[tag[4:]]
	// check there is a location for the key
	if (location == artGrid{}) {

		return nil, image.Point{}, nil, fmt.Errorf("0050 the key %s was not found", tag)
	}

	return location.canvas, image.Point{location.loc.X, location.loc.Y}, location.mask, nil
}

// key gen extracts the file from a http source then a local source
func keyGen(c *context.Context, base string, bounds image.Point) (*image.NRGBA64, error) {

	file, err := credentials.GetWebBytes(c, base)
	if err == nil {
		return extract(file, base, bounds)
	}

	file, err = os.ReadFile(base)
	if err == nil {

		return extract(file, base, bounds)
	}

	return nil, fmt.Errorf("0042 error opening background image %v", err)

}

// extract converts the bytes to an image and scales it if required
func extract(b []byte, fname string, bounds image.Point) (i *image.NRGBA64, e error) {
	// assign the context to loop it out

	var midI image.Image
	read := bytes.NewReader(b)
	regTIFF := regexp.MustCompile(`^[\w\W]{1,255}\.[tT][iI][fF]{1,2}$`)
	regPNG := regexp.MustCompile(`^[\w\W]{1,255}\.[pP][nN][gG]$`)
	switch {
	case regPNG.MatchString(fname):
		midI, e = png.Decode(read)
	case regTIFF.MatchString(fname):
		midI, e = tiff.Decode(read)
	default:
		e = fmt.Errorf("0043 %v is an invalid file type", fname)
	}
	// resize the image to fit the canvas
	if midI.Bounds().Max.X != bounds.X || midI.Bounds().Max.Y != bounds.Y {
		midI = resize.Resize(uint(bounds.X), uint(bounds.Y), midI, resize.Bicubic)
	}
	i = image.NewNRGBA64(midI.Bounds())
	colour.Draw(i, i.Bounds(), midI, image.Point{}, draw.Src)
	if e != nil {
		e = fmt.Errorf("0042 error opening background image %v", e)
	}

	return
}

// imageToKeyMap finds the areas of transparency in an image
// the assign the locations, mask and canvas in context
func imageToKeyMap(toScan *image.NRGBA64, c *context.Context) (map[string]artGrid, error) {
	// skip the process is the image is empty
	artKey := make(map[string]artGrid)
	if toScan.Opaque() {
		return artKey, nil
	}

	cells := make(map[image.Point]cell)
	scanBounds := toScan.Bounds().Max
	// loop through the image to find groups of transparency
	var groups []map[image.Point]cell

	// process as pixels to stop calling canvas.At thousands of times
	scanPix := toScan.Pix

	for x := 0; x < scanBounds.X; x++ {
		for y := 0; y < scanBounds.Y; y++ {
			location := image.Point{x, y}
			c := cells[location]
			if !cells[location].checked { // skip the cell if it's already been checked
				group := visit(scanPix, location, scanBounds, cells)
				if len(group) != 0 {
					groups = append(groups, group)
				}
			}
			c.checked = true
			cells[location] = c
		}
	}

	// find the areas of each square
	for _, target := range groups {
		minX := scanBounds.X
		minY := scanBounds.Y
		minX, minY, maxX, maxY := keySizeFinder(target, minX, minY)

		// add 1 to include the full range of the image
		// update this to be generic for the different image types
		result := ImageGenerator(*c, image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
		mask := ImageGenerator(*c, image.Rect(0, 0, maxX-minX+1, maxY-minY+1))

		// find the rgba value of the key
		r, g, b, a := visitKey(scanPix, image.Point{minX + int(float64(maxX-minX)/2.0), minY + int(float64(maxY-minY)/2.0)}, scanBounds, target)
		key := rgbaKey(r, g, b, a)

		if key != "" { // if the colour matches a key add to the map
			var aG artGrid
			aG.canvas = result
			aG.loc = image.Point{minX, minY}
			aG.mask = mask
			artKey[key] = aG // assign it to the map of locations for this run
		} else {

			return nil, fmt.Errorf("0044 Transparent area found with no key, aborting key set up")
		}
		// assign a transparency mask to the mask image relative to the pixels that are measured
		for k, v := range target {
			if v.match {
				mask.Set(k.X-minX, k.Y-minY, color.NRGBA64{A: 0xffff}) //- v.opacity})
			}
		}
	}

	return artKey, nil
}

// keySizeFinder extracts the range of the x,y positions of an image key.
// It returns the  minX, minY, maxX, maxY
func keySizeFinder(target map[image.Point]cell, minX, minY int) (int, int, int, int) {
	var maxX, maxY int
	for position, cell := range target {
		if cell.match { // check these are valid coordinates
			if position.X < minX {
				minX = position.X
			} else if position.X > maxX {
				maxX = position.X
			}
			if position.Y < minY {
				minY = position.Y
			} else if position.Y > maxY {
				maxY = position.Y
			}
		}

	}

	return minX, minY, maxX, maxY
}

// art grid contains the info to translate an image onto the grid
type artGrid struct {
	loc    image.Point
	canvas draw.Image
	mask   draw.Image
}

func pixPos(x, y int, b image.Point) int {

	return y*b.X*8 + x*8
}

// cell is used to check if a struct is transparent and if it has been checked
type cell struct {
	checked   bool
	match     bool
	neighbour bool
	opacity   uint16
}

// checks a cell and if it is transparent it searches through the neighbours to find groups of transparency
func visit(canvas []uint8, location, maxB image.Point, cells map[image.Point]cell) map[image.Point]cell {
	var badCell cell
	badCell.checked = true
	if uint16(canvas[pixPos(location.X, location.Y, maxB)+6])<<8|uint16(canvas[pixPos(location.X, location.Y, maxB)]+7) > 49152 {
		cells[location] = badCell

		return nil
	}
	i := 0
	// checker is the map of cells for this group
	checker := make(map[image.Point]cell)
	// dummy variable for quick assigning
	var trueCell cell
	trueCell.checked = true
	trueCell.match = true

	allNeighbours := neighbourGen(location, maxB, cells) // []image.Point{{x + 1, y}, {x - 1, y}, {x, y + 1}, {x, y - 1}}

	checker[location] = trueCell
	for { // search direction
		cellToCheck := allNeighbours[i]
		checkCell := cells[cellToCheck]
		checkCell.checked = true
		pos := pixPos(cellToCheck.X, cellToCheck.Y, maxB)
		opaque := uint16(canvas[pos+6])<<8 | uint16(canvas[pos+7])
		// 19456
		if opaque < 49152 && !cells[cellToCheck].checked {
			// fmt.Println(n, cells[n])
			checkCell.match = true
			checkCell.opacity = opaque
			checker[cellToCheck] = checkCell
			newN := neighbourGen(cellToCheck, maxB, cells)
			allNeighbours = append(allNeighbours, newN...)

		}
		cells[cellToCheck] = checkCell // update the map of all cells to prevent double dipping
		i++
		if i >= len(allNeighbours) {

			break
		}
	}

	return checker
}

// generarte the neighbours for a point
// checking they aren't on the boundary or already been checked
func neighbourGen(loc, bounds image.Point, cells map[image.Point]cell) []image.Point {
	neighbours := [4]image.Point{{loc.X + 1, loc.Y}, {loc.X - 1, loc.Y}, {loc.X, loc.Y + 1}, {loc.X, loc.Y - 1}}
	valid := make([]image.Point, 4)
	vPos := 0
	for _, n := range neighbours {
		// check if any of the neighbour positions are valid
		if !cells[n].checked && !cells[n].neighbour {
			if !(n.X < 0) && !(n.Y < 0) && !(n.X >= bounds.X) && !(n.Y >= bounds.Y) {
				valid[vPos] = n
				vPos++
			}
		}
		c := cells[n]
		c.neighbour = true
		cells[n] = c
	}

	return valid[:vPos]
}

// generate the average colour for the middle key
func visitKey(canvas []uint8, location, maxB image.Point, cells map[image.Point]cell) (uint16, uint16, uint16, uint16) {
	if canvas[pixPos(location.X, location.Y, maxB)+7] == 0 {
		return 0, 0, 0, 0
	}
	// when it's matched spiral through the area of neighbours

	c := cells[location]
	i := 0

	var r, g, b, a int
	var trueCell cell
	trueCell.checked = true
	trueCell.match = true
	var total int

	var reset cell
	reset.match = true
	// reset the fringe cells as checked so they can be checked as edge cases of the key
	for k, v := range cells {
		if v.opacity != 0 {
			reset.opacity = v.opacity
			cells[k] = reset
		}
	}

	allNeighbours := neighbourGen(location, maxB, cells) // []image.Point{{x + 1, y}, {x - 1, y}, {x, y + 1}, {x, y - 1}}

	c.match = true
	for { // search direction
		cellToCheck := allNeighbours[i]

		checkCell := cells[cellToCheck]
		checkCell.checked = true
		pos := pixPos(cellToCheck.X, cellToCheck.Y, maxB)
		if canvas[pos+7] != 0 && !cells[cellToCheck].checked {
			total++
			checkCell.match = true
			checkCell.opacity = 0

			nr := uint16(canvas[pos])<<8 | uint16(canvas[pos+1])
			ng := uint16(canvas[pos+2])<<8 | uint16(canvas[pos+3])
			nb := uint16(canvas[pos+4])<<8 | uint16(canvas[pos+5])
			na := uint16(canvas[pos+6])<<8 | uint16(canvas[pos+7])
			r += int(nr)
			g += int(ng)
			b += int(nb)
			a += int(na)

			newN := neighbourGen(cellToCheck, maxB, cells)
			allNeighbours = append(allNeighbours, newN...)
		}
		cells[cellToCheck] = checkCell
		i++
		if i >= len(allNeighbours) {
			break // break the loop once every neighbour has been checked
		}
	}

	if total == 0 {
		return 0, 0, 0, 0
	}
	cells[location] = c
	// return the average rgba value
	return uint16(r / total), uint16(g / total), uint16(b / total), uint16(a / total)
}

// hardcoded keynames
func rgbaKey(r, g, b, a uint16) string {
	min := uint16(100)
	max := uint16(60000)
	midBot := uint16(25000)
	midTop := uint16(50000)

	switch {
	case r < min && g < min && b > max && a > max:
		return "blue"
	case r < min && g > max && b < min && a > max:
		return "green"
	case r > max && g < min && b < min && a > max:
		return "red"
	case r < min && g < min && b < min && a > max:
		return "black"
	case r > max && g > max && b > max && a > max:
		return "white"
	case r > max && g > max && b < min && a > max:
		return "yellow"
	case r < min && g > max && b > max && a > max:
		return "cyan"
	case r > max && g < min && b > max && a > max:
		return "purple"
	case r < min && g < min && b > midBot && b < midTop && a > max:
		return "dimBlue"
	case r > midBot && r < midTop && g < min && b < min && a > max:
		return "dimRed"
	case r < min && g > midBot && g < midTop && b < min && a > max:
		return "dimGreen"
	case r > midBot && r < midTop && g < min && b > midBot && b < midTop && a > max:
		return "dimPurple"
	case r > midBot && r < midTop && g > midBot && g < midTop && b < min && a > max:
		return "dimYellow"
	case r > midBot && r < midTop && g > midBot && g < midTop && b > midBot && b < midTop && a > max:
		return "dimCyan"
	case r < min && g > midBot && g < midTop && b > midBot && b < midTop && a > max:
		return "gray"
	default:
		return "" // with unknown colours
	}
}
