// Package mask is used for generating masked images
package mask

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
)

// Generate a mask of correct size

// Rectangle is the border
// Circle x2 and y2 and assign an alpha channel
// Uint16(0xffff) or alpha 16

// Mask generates a square or circular mask of w and h,
// this mask is then placed on an existing image to create an image at a new layer.
func Mask(shape string, w, h, x, y int, oldCanvas image.Image) *image.NRGBA64 {
	// Generate the mask
	mask := maskCanvas(shape, w, h)
	// Create an empty layer to move the image to
	masked := image.NewNRGBA64(image.Rect(0, 0, w, h))

	colour.DrawMask(masked, masked.Bounds(), oldCanvas, image.Point{x, y}, mask, image.Point{}, draw.Src)
	// _ = savefile.Savefile("circle", "png", masked)

	return masked
}

const (
	Circle = "circle"
	Square = "square"
)

func maskCanvas(shape string, w, h int) *image.NRGBA64 {
	c := image.NewNRGBA64(image.Rect(0, 0, w, h))
	opaque := color.Alpha16{A: uint16(0xffff)}
	switch shape {
	case Circle:
		c = circle(c, opaque)
	case Square:
		c = square(c, opaque)
	}

	return c
}

func square(maskc *image.NRGBA64, c color.Color) *image.NRGBA64 {
	// Fill the whole thing as alpha an alpha channel
	for i := 0; i < maskc.Bounds().Max.X; i++ {
		for j := 0; j < maskc.Bounds().Max.Y; j++ {
			maskc.Set(i, j, c)
		}
	}

	return maskc
}

func circle(maskc *image.NRGBA64, c color.Color) *image.NRGBA64 {
	// Implemenent some x y equal checks
	x := float64(maskc.Bounds().Max.X)
	y := float64(maskc.Bounds().Max.Y)
	radiusSq := math.Pow(x/2, 2)
	for i := -x / 2; i < x/2; i++ {
		for j := -y / 2; j < y/2; j++ {
			// If these values are within the circle we treat them as
			// opaque areas
			if radiusSq >= (i*i + j*j) {
				// Set x and y back to 0 minimums
				maskc.Set(int(i+x/2), int(j+y/2), c)
			}
		}
	}

	return maskc
}
