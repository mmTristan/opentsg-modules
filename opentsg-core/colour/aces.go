package colour

import (
	"image"
	"image/color"
	"math"
)

/////////////
////ACES/////
////////////

// ARGBA is aces RGBA and contains the same properties of a image.Image
type ARGBA struct {
	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// NewARGBA generates an ARGBA image of size r
func NewARGBA(r image.Rectangle) *ARGBA {
	return &ARGBA{
		Pix:    make([]uint8, r.Dx()*r.Dy()*16),
		Stride: 16 * r.Dx(),
		Rect:   r,
	}
}

// Bounds gives the size of the image
func (a *ARGBA) Bounds() image.Rectangle { return a.Rect }

// Set sets the colour of ARGBA at the x,y position
func (a *ARGBA) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(a.Rect)) {
		return
	}
	i := a.PixOffset(x, y)
	var m model
	ogc := m.Convert(c).(RGBA128)
	s := a.Pix[i : i+16 : i+16] // Small cap improves performance, see https://golang.org/issue/27857
	c1 := rGBA128Alias{math.Float32bits(ogc.R), math.Float32bits(ogc.G), math.Float32bits(ogc.B), math.Float32bits(ogc.A)}
	s[0] = byte(c1.R >> 24)
	s[1] = byte(c1.R >> 16)
	s[2] = byte(c1.R >> 8)
	s[3] = byte(c1.R)
	s[4] = byte(c1.G >> 24)
	s[5] = byte(c1.G >> 16)
	s[6] = byte(c1.G >> 8)
	s[7] = byte(c1.G)
	s[8] = byte(c1.B >> 24)
	s[9] = byte(c1.B >> 16)
	s[10] = byte(c1.B >> 8)
	s[11] = byte(c1.B)
	s[12] = byte(c1.A >> 24)
	s[13] = byte(c1.A >> 16)
	s[14] = byte(c1.A >> 8)
	s[15] = byte(c1.A)
}

// PixOffset gives the pixel position at X,Y
func (a *ARGBA) PixOffset(x, y int) int {
	return (y-a.Rect.Min.Y)*a.Stride + (x-a.Rect.Min.X)*16
}

// At returns the colour at x,y
func (a *ARGBA) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(a.Rect)) {
		return RGBA128{}
	}
	i := a.PixOffset(x, y)
	s := a.Pix[i : i+16 : i+16] // Small cap improves performance, see https://golang.org/issue/27857

	return RGBA128{
		math.Float32frombits((uint32(s[0])<<24 | uint32(s[1])<<16 | uint32(s[2])<<8 | uint32(s[3]))),
		math.Float32frombits((uint32(s[4])<<24 | uint32(s[5])<<16 | uint32(s[6])<<8 | uint32(s[7]))),
		math.Float32frombits((uint32(s[8])<<24 | uint32(s[9])<<16 | uint32(s[10])<<8 | uint32(s[11]))),
		math.Float32frombits((uint32(s[12])<<24 | uint32(s[13])<<16 | uint32(s[14])<<8 | uint32(s[15]))),
	}

}

// ColorModel returns the color model of argba, which is used for converting colours
func (a *ARGBA) ColorModel() color.Model {
	m := model{}

	return m
}

///////////////
////COLOUR////
//////////////

// RGBA128 is the ACES colour struct using float 32 as there is no native float 16
type RGBA128 struct {
	R, G, B, A float32
}
type rGBA128Alias struct {
	R, G, B, A uint32
}

// RGBA returns the rgba value of RGBA128
func (c RGBA128) RGBA() (r, g, b, a uint32) {
	// multiply by alpha to keep in line with the rest of
	// the go image.Image library
	r = uint32(c.R)
	r *= uint32(c.A)
	r /= 0xffff
	g = uint32(c.G)
	g *= uint32(c.A)
	g /= 0xffff
	b = uint32(c.B)
	b *= uint32(c.A)
	b /= 0xffff
	a = uint32(c.A)

	return
}

////////////
///MODEL///
///////////

type model struct {
}

// convert converts a color to RGBA128
func (m model) Convert(c color.Color) color.Color {
	if _, ok := c.(RGBA128); ok {
		return c
	}
	/*
		//if n, ok := c.(color.NRGBA64); ok { //adds a slight performance increase as most things are written in this model
		//	return RGBA128{float32(n.R), float32(n.G), float32(n.B), float32(n.A)}
		//}
	*/
	r, g, b, a := c.RGBA()
	if a == 0xffff {
		return RGBA128{float32(uint16(r)), float32(uint16(g)), float32(uint16(b)), 0xffff}
	}
	if a == 0 {
		return RGBA128{0, 0, 0, 0}
	}
	// Since Color.RGBA returns an alpha-premultiplied color, we should have r <= a && g <= a && b <= a.
	r = (r * 0xffff) / a
	g = (g * 0xffff) / a
	b = (b * 0xffff) / a

	return RGBA128{float32(uint16(r)), float32(uint16(g)), float32(uint16(b)), float32(uint16(a))}
}
