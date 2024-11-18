// Package colourgen generates rgb values
package parameters

import (
	"fmt"
	"image/color"
	"regexp"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
)

// HexToColour takes a string and returns a colour value, extracting the rgba values from the string. When no alpha channel
// is found the alpha is set to be the max 16 bit value.
//
// Acceptable formats are #rgb, #rgba, #rrggbb, ##rrggbbaa, rgb(r,g,b), rgba(r,g,b,a), rgb12(r,g,b) and rgba12(r,g,b,a)
//
// The resulting value is either color.NRGBA or color.NRBGA64, 12 bit RGB values are represented in 16 bit NRGBA64.
// If the alpha channel is found and its the maximum value the maximum 16 bit value is used.
func HexToColour(colorCode string, space colour.ColorSpace) *colour.CNRGBA64 {
	var base *colour.CNRGBA64
	regRRGGBB := regexp.MustCompile(`^#[A-Fa-f0-9]{6}$`)
	regRGB := regexp.MustCompile(`^#[A-Fa-f0-9]{3}$`)
	regRRGGBBAA := regexp.MustCompile(`^#[A-Fa-f0-9]{8}$`)
	regRGBA := regexp.MustCompile(`^#[A-Fa-f0-9]{4}$`)
	regcssRGBA := regexp.MustCompile(`^(rgba\()\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\)$`)
	regcssRGB := regexp.MustCompile(`^(rgb\()\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5]),\b([01]?[0-9][0-9]?|2[0-4][0-9]|25[0-5])\)$`)
	regcssRGB12 := regexp.MustCompile(`^rgb12\(([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5])\)$`)
	regcssRGBA12 := regexp.MustCompile(`^rgba12\(([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5]),([0-3]?[0-9]{1,3}|40[0-9][0-5])\)$`)
	// break down the string an attribute each 2 hex value to the rgb

	// check length as all are unqiue>
	switch {
	case regRRGGBB.MatchString(colorCode):

		base = rrggbb(colorCode)
	case regRGB.MatchString(colorCode):

		base = rgb(colorCode)
	case regRRGGBBAA.MatchString(colorCode):

		base = rrggbbaa(colorCode)
	case regRGBA.MatchString(colorCode):

		base = rgba(colorCode)
	case regcssRGBA.MatchString(colorCode):

		base = cssrgba(colorCode)
	case regcssRGB.MatchString(colorCode):

		base = cssrgb(colorCode)
	case regcssRGB12.MatchString(colorCode):

		base = cssrgb12(colorCode)
	case regcssRGBA12.MatchString(colorCode):

		return cssrgba12(colorCode)
	default:
		// base = &colour.CNRGBA64{}
		return nil
	}

	base.ColorSpace = space
	return base
}

type HexString string

func (h HexString) ToColour(space colour.ColorSpace) *colour.CNRGBA64 {
	return HexToColour(string(h), space)
}

func rrggbb(hex string) *colour.CNRGBA64 {
	var R, G, B uint16
	fmt.Sscanf(hex, "#%02x%02x%02x", &R, &G, &B)

	return &colour.CNRGBA64{R: R << 8, G: G << 8, B: B << 8, A: 0xffff}
}

func rgb(hex string) *colour.CNRGBA64 {
	var R, G, B uint16
	fmt.Sscanf(hex, "#%01x%01x%01x", &R, &G, &B)

	return &colour.CNRGBA64{R: R << 12, G: G << 12, B: B << 12, A: 0xffff}
}

func rrggbbaa(hex string) *colour.CNRGBA64 {
	var R, G, B, A uint16
	fmt.Sscanf(hex, "#%02x%02x%02x%02x", &R, &G, &B, &A)
	if A == 0xff {
		A = 0xffff
	} else {
		A <<= 8
	}

	return &colour.CNRGBA64{R: R << 8, G: G << 8, B: B << 8, A: A}
}

func rgba(hex string) *colour.CNRGBA64 {
	var R, G, B, A uint16
	fmt.Sscanf(hex, "#%01x%01x%01x%01x", &R, &G, &B, &A)

	if A == 0xf {
		A = 0xffff
	} else {
		A <<= 12
	}

	return &colour.CNRGBA64{R: R << 12, G: G << 12, B: B << 12, A: A}
}

func cssrgba(css string) *colour.CNRGBA64 {
	var R, G, B, A uint16
	fmt.Sscanf(css, "rgba(%v,%v,%v,%v)", &R, &G, &B, &A)
	if A == 0xff {
		A = 0xffff
	} else {
		A <<= 8
	}

	return &colour.CNRGBA64{R: R << 8, G: G << 8, B: B << 8, A: A}
}

func cssrgb(css string) *colour.CNRGBA64 {
	var R, G, B uint16
	fmt.Sscanf(css, "rgb(%v,%v,%v)", &R, &G, &B)

	return &colour.CNRGBA64{R: R << 8, G: G << 8, B: B << 8, A: 0xffff}
}

func cssrgb12(css string) *colour.CNRGBA64 {
	var R, G, B uint16
	fmt.Sscanf(css, "rgb12(%v,%v,%v)", &R, &G, &B)

	return &colour.CNRGBA64{R: R << 4, G: G << 4, B: B << 4, A: 0xffff}
}

func cssrgba12(css string) *colour.CNRGBA64 {
	var R, G, B, A uint16
	fmt.Sscanf(css, "rgba12(%v,%v,%v,%v)", &R, &G, &B, &A)
	if A == 4095 {
		A = 0xffff
	} else {
		A <<= 4
	}

	return &colour.CNRGBA64{R: R << 4, G: G << 4, B: B << 4, A: 0xffff}
}

// ConvertNRGBA64 converts any colour into an NRGBA64 colour.
// The colours are returned as 8 bit colours shifted to 16 bit, unless already a color.NRGBA64 then that is returned without change.
// This is designed to preserve the colours 8 bit representation and not change the original value to be a warped 16 bit
// representation.
//
// Max 8 bit alpha channel values of 255 are set to max 16 bit values of 65535. This is to make solid images.
//
// e.g. rgb(250,4,100) becomes rgb(64000,1024,25600) in NRGBA64 form
func ConvertNRGBA64(original color.Color) color.NRGBA64 {
	if col, ok := original.(color.NRGBA64); ok {
		return col
	}

	nrgba := color.NRGBAModel.Convert(original).(color.NRGBA)

	convert := color.NRGBA64{R: uint16(nrgba.R) << 8, G: uint16(nrgba.G) << 8, B: uint16(nrgba.B) << 8}
	if nrgba.A == 0xff { // If alpha is the max value for an 8 bit number
		// round up to make solid for 16 bit images
		convert.A = 0xffff
	} else {
		convert.A = uint16(nrgba.A) << 8
	}

	return convert
}
