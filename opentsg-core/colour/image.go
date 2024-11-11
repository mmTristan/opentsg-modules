package colour

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"

	"golang.org/x/image/tiff"
)

/*
ColorSpace contains all the information of a colours colourspace.
*/
type ColorSpace struct {
	// Short form to be used for common colour spaces e.g. "rec709"
	ColorSpace string `json:"ColorSpace,omitempty" yaml:"ColorSpace,omitempty"`
	// Preffered transformtype. Not currently in use
	TransformType string `json:"TransformType,omitempty" yaml:"TransformType,omitempty"`
	// for custom colourspaces, the primaries can be declared in XY space
	Primaries Primaries `json:"Primaries,omitempty" yaml:"Primaries,omitempty"`
}

// The Primaries of a colour space consist of the XY coordinates
// of the RGB and whitepoint
type Primaries struct {
	Red        XY `json:"Red,omitempty" yaml:"Red,omitempty"`
	Green      XY `json:"Green,omitempty" yaml:"Green,omitempty"`
	Blue       XY `json:"Blue,omitempty" yaml:"Blue,omitempty"`
	WhitePoint XY `json:"WhitePoint,omitempty" yaml:"WhitePoint,omitempty"`
}

// XY is the XY spae of the CIE colour chart
type XY struct {
	X int `json:"X,omitempty" yaml:"X,omitempty"`
	Y int `json:"Y,omitempty" yaml:"Y,omitempty"`
}

// Image is used for colour space aware images.
type Image interface {
	Space() ColorSpace
	draw.Image // Draw include Set

	// Margins()  // LeftRightTopBottom
}

// NRGBA64 is a wrapped *image.NRGBA64 with a colorspace
type NRGBA64 struct {
	BaseImg     *image.NRGBA64
	ColourSpace ColorSpace
}

// Generate a new NRGBA64 imagethat is wrapped with a colour space
func NewNRGBA64(s ColorSpace, r image.Rectangle) *NRGBA64 {

	base := image.NewNRGBA64(r)

	return &NRGBA64{BaseImg: base, ColourSpace: s}

}

func (n NRGBA64) Bounds() image.Rectangle {
	return n.BaseImg.Bounds()
}

// Space returns the ColorSpace of the Image
func (n NRGBA64) Space() ColorSpace {
	return n.ColourSpace
}

// return the pixels of the base image
func (n NRGBA64) Pix() []uint8 {
	return n.BaseImg.Pix
}

func (n NRGBA64) At(x, y int) color.Color {
	/* can wrap
		NRGBA 64 colour as an tsg.colour
		if we want to preserve colour space
	 //	n.Base.NRGBA64At(x,y)

	*/

	baseCol := n.BaseImg.NRGBA64At(x, y)
	// return a colour space aware colour
	return &CNRGBA64{R: baseCol.R, G: baseCol.G, B: baseCol.B, A: baseCol.A, ColorSpace: n.ColourSpace}

}

// ColorModel retruns the *image.NRGBA64 colorModel
func (n NRGBA64) ColorModel() color.Model {
	return n.BaseImg.ColorModel()
}

// BaseImage returns the *image.NRGBA64
// so it can be used with the go library
func (n NRGBA64) BaseImage() *image.NRGBA64 {
	return n.BaseImg
}

// Set is the same as image.NRGBA64.Set(), but with
// a colour transform before setting the image.
func (n NRGBA64) Set(x int, y int, c color.Color) {

	// update the colour if it has an explicit colour space
	// and the base image is using colour spaces
	if cmid, ok := c.(Color); ok && (n.ColourSpace != ColorSpace{}) {
		c = transform(cmid.GetColorSpace(), n.ColourSpace, c)
	}

	// use SetNRGBA64 where possible to preserve colour
	switch convert := c.(type) {
	case color.NRGBA64:
		n.BaseImg.SetNRGBA64(x, y, convert)
	case *CNRGBA64:

		n.BaseImg.SetNRGBA64(x, y, color.NRGBA64{R: convert.R, G: convert.G, B: convert.B, A: convert.A})
	default:
		n.BaseImg.Set(x, y, convert)
	}
	//	n.base.SetNRGBA64(x, y, color.NRGBA64{R: uint16(R), G: uint16(G), B: uint16(B), A: uint16(A)})
	//
	// fmt.Println(n.base.At(x, y))
}

// PNGEncode is a wrapper of png.Encode, it allows
// the NRGBA64 image to pass the base *image.NRGBA64
// to the encoder so the png is saved properly.
func PngEncode(w io.Writer, m image.Image) error {

	// cut out the NRGB64 wrapper as png
	// doesn't know how to handle it correctly
	// and it changes the expected values when alpha is not 0xffff
	if mid, ok := m.(*NRGBA64); ok {
		m = mid.BaseImg
	}

	return png.Encode(w, m)
}

// TiffEncode is a wrapper of png.Encode, it allows
// the NRGBA64 image to pass the base *image.NRGBA64
// to the encoder so the tiff is saved properly.
func TiffEncode(w io.Writer, m image.Image, opt *tiff.Options) error {

	// cut out the NRGB64 wrapper as tiff
	// doesn't know how to handle it correctly
	// and it changes the expected values when alpha is not 0xffff
	if mid, ok := m.(*NRGBA64); ok {
		m = mid.BaseImg
	}

	return tiff.Encode(w, m, opt)
}
