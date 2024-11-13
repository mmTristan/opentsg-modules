package gridgen

import (
	"context"
	"fmt"
	"image"
	"image/draw"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
)

type Box interface {
	GenerateBox(c *context.Context) (canvas draw.Image, loc image.Point, mask draw.Image, err error)
}

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

type idea struct {

	// keep the alias from last time
	alias string
	//
	bounds bounds
}

// implement hsl(0, 100%, 50%);

// keep these ideas in mind https://www.w3schools.com/css/css_boxmodel.asp
/*
remove margin, padding and border as there will not be that mich need

thoughts - sa y the limits so people dont think this is a dierect css import
as fetures will deffo be missing

*/
type bounds struct {
	// use a predeclared alias
	// alias must be declared before
	useAlias string

	// top left coordinates
	// actually any
	x, y any
	// bottom right
	// if not used then the grid is 1 square
	x2, y2 any

	// width height
	// can they be A or 1 etc. just mix it up
	w, h any // width: 200px; height: 150px;

	// centre values
	// width
	XAlignment, YAlignment string // default top left but let them choose
	// or masks like this. Leave masks out for the moment?
	//  mask-image: radial-gradient(circle, black 50%, rgba(0, 0, 0, 0.5) 50%);

	// circle properties
	// border radius - what css uses
	// https://prykhodko.medium.com/css-border-radius-how-does-it-work-bfdf23792ac2
	// taps out at 50% - keep it the simple version to start
	radius int
}

/*

no shapes apart from square with rounded edges

cx,cy? are these needed when you can still set squares
keep

*/

// https://www.w3schools.com/css/css_boxmodel.asp
func (b bounds) generateBox(c *context.Context) (draw.Image, image.Point, draw.Image, error) {

	// get the start point
	/*
		either xy or cx cy
		 or useALias
	*/
	aliasMap := core.GetAlias(*c)
	switch {
	case b.useAlias != "":
		// @TODO update the alias to be the
		// the image.Point, canvas size and a mask,
		// if applicable
		loc := aliasMap.Data[b.useAlias]
		if loc != "" {
			// call the function again but with the required coordinates
			mid, _ := gridSquareLocatorAndGenerator(loc, "", c)
			return mid.GImage, image.Point{mid.X, mid.Y}, mid.GMask, nil
		} else {

			return nil, image.Point{}, nil, fmt.Errorf(invalidAlias, b.useAlias)
		}
	case b.x != nil || b.y != nil:

	default:
		// return no coordinate postion used
	}

	/*
		get end locatoin

		wh, r, or x2y2 for xy
		wh, r, for cxcy
	*/

	/*
		now we have the square image the mask is calculated
		which is if radius or offsets (or both)
	*/

	// get the end point

	// returns the mask, coordinate and base image and error
	return nil, image.Point{}, nil, nil
}

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
