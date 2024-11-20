// Package colour contains all the functions for utilising colour spaces in OpenTSG
package colour

import (
	"image"
	"image/color"
	"image/draw"
)

// maxAlpha is the maximum color value returned by image.Color.RGBA.
const maxAlpha = 1<<16 - 1

/*
CNRGBA64 behaves in the same way as color.NRGBA64
with the added feature of a color space
*/
type CNRGBA64 struct {
	// color.NRGBA64
	R, G, B, A uint16
	ColorSpace ColorSpace
}

// The Color interface is the same as color.Color
// with the ability to find the colour space the
// colour is based in.
type Color interface {
	color.Color
	GetColorSpace() ColorSpace
	UpdateColorSpace(ColorSpace)
}

func (c *CNRGBA64) GetColorSpace() ColorSpace {
	return c.ColorSpace
}

func (c *CNRGBA64) UpdateColorSpace(s ColorSpace) {
	c.ColorSpace = s
}

func (c *CNRGBA64) RGBA() (r, g, b, a uint32) {
	return color.NRGBA64{R: c.R, G: c.G, B: c.B, A: c.A}.RGBA()
}

// CyCbCr is a demo version of a color.CyCbCr
type CyCbCr struct {
	Y, Cb, Cr uint8
	Space     ColorSpace
}

func (c *CyCbCr) GetSpace() ColorSpace {
	return c.Space
}

func (c *CyCbCr) UpdateSpace(s ColorSpace) {
	c.Space = s
}

func (c *CyCbCr) RGBA() (r, g, b, a uint32) {
	return color.YCbCr{Y: c.Y, Cb: c.Cb, Cr: c.Cr}.RGBA()
}

// Draw calls DrawMask with a nil mask.
func Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point, op draw.Op) {
	DrawMask(dst, r, src, sp, nil, image.Point{}, op)
}

// DrawMask aligns r.Min in dst with sp in src and mp in mask and then replaces the rectangle r
// in dst with the result of a Porter-Duff composition. A nil mask is treated as opaque.
/*
This version has been made to include the transform options for when colour space aware colours
are used. This works by using the transform function before placing the colour on the destination
image

Further more this uses NRGBA64 as a base to set colours, rather than the RGBA64 model
favoured by go. This means that the non alpha multiplied RGB values are used unless
required when alpha is neither 0 or maxAlpha. This leads to slight discrepancies
with the go value, but this is more accurate.

This function is only recommended when using NRGB64 images that are colour space aware and
you drawing with the same base images.
*/
func DrawMask(dest draw.Image, r image.Rectangle, src image.Image, sp image.Point, mask image.Image, mp image.Point, op draw.Op) {

	switch dst := dest.(type) {
	case *NRGBA64:

		// follow the draw.DrawMask generic code
		// with a few slight differences to ensure colour space is preserved.

		clip(dest, &r, src, &sp, mask, &mp)
		if r.Empty() {
			return
		}

		x0, x1, dx := r.Min.X, r.Max.X, 1
		y0, y1, dy := r.Min.Y, r.Max.Y, 1
		// set the colour to be colour space aware
		var out CNRGBA64
		sy := sp.Y + y0 - r.Min.Y
		my := mp.Y + y0 - r.Min.Y
		c := 0
		for y := y0; y != y1; y, sy, my = y+dy, sy+dy, my+dy {
			c++
			sx := sp.X + x0 - r.Min.X
			mx := mp.X + x0 - r.Min.X
			for x := x0; x != x1; x, sx, mx = x+dx, sx+dx, mx+dx {

				ma := uint32(maxAlpha)
				if mask != nil {
					_, _, _, ma = mask.At(mx, my).RGBA()
				}

				switch {
				// case op == draw.Over:
				// this differs from the go code
				// as it sets straight on top as does not change to rGBA64 like
				// teh draw.Drawmask function does
				//	dst.Set(x, y, src.At(sx, sy))

				case ma == 0:
					if op == draw.Over {
						// No-op.
					} else { // reset to a transparent pixel
						dest.Set(x, y, color.Transparent)
					}
				case ma == maxAlpha && op == draw.Src:
					dest.Set(x, y, src.At(sx, sy))
				default:

					/*


						we know the base is NRGBA64


					*/

					srCol := src.At(sx, sy)
					var sr, sg, sb, sa uint32

					// this works in nrgb64, so we treat every colour as NRGB64 by getting the non multiplied values
					// and only alpha multiplying them if we're required
					if ncol, ok := srCol.(color.NRGBA64); ok {
						sr, sg, sb, sa = uint32(ncol.R), uint32(ncol.G), uint32(ncol.B), uint32(ncol.A)
					} else if cspace, ok := src.At(sx, sy).(*CNRGBA64); ok {

						// transform the colour before applying it
						tCol := transform(cspace.ColorSpace, dst.ColourSpace, src.At(sx, sy))
						ncol := tCol.(*CNRGBA64)
						// making sure to cut out alpha multiplied values
						//	atc := tCol.(*CNRGBA64)
						//	sr, sg, sb, sa = uint32(atc.R), uint32(atc.G), uint32(atc.B), uint32(atc.A)
						sr, sg, sb, sa = uint32(ncol.R), uint32(ncol.G), uint32(ncol.B), uint32(ncol.A)
						// sr, sg, sb, sa = tCol.RGBA()
					} else {
						// convert these into non alpha multiplied values
						// this will be lossy so stick to NRGBA64 or CNRGBA64
						// if you want accurate values
						// @TODO figure out how to stop RGBA 64 causing overflow issues
						//	sr, sg, sb, sa = srCol.RGBA()

						nrgbCol := color.NRGBA64Model.Convert(srCol).(color.NRGBA64)

						sr, sg, sb, sa = uint32(nrgbCol.R), uint32(nrgbCol.G), uint32(nrgbCol.B), uint32(nrgbCol.A)
					}
					// NRGBA64 is non alpha multiplied

					if op == draw.Src || sa == maxAlpha {
						out.R = uint16(sr * ma / maxAlpha)
						out.G = uint16(sg * ma / maxAlpha)
						out.B = uint16(sb * ma / maxAlpha)
						out.A = uint16(sa * ma / maxAlpha)

					} else {
						dstCol := dest.At(x, y).(*CNRGBA64)
						dr, dg, db, da := uint32(dstCol.R), uint32(dstCol.G), uint32(dstCol.B), uint32(dstCol.A)

						switch {
						case da == 0:
							out = CNRGBA64{R: uint16(sr), G: uint16(sg), B: uint16(sb), A: uint16(sa)}
						case sa == 0:

							// don't draw anything as its transparent
							continue
						default:

							// else get the alpha multiplied version of the dst and src RGBA values
							sr, sg, sb = ((sr * sa) / maxAlpha), ((sg * sa) / maxAlpha), ((sb * sa) / maxAlpha)
							dr, dg, db = ((dr * da) / maxAlpha), ((dg * da) / maxAlpha), ((db * da) / maxAlpha)

							// the alpha weighting can be changed
							// to a function for when different
							// clour spaces are usef
							a := maxAlpha - (sa * ma / maxAlpha)

							out.A = uint16((da*a + sa*ma) / maxAlpha)
							// divide by alpha to get th eno alpha multiplied version
							out.R = uint16((dr*a + sr*ma) / uint32(out.A))
							out.G = uint16((dg*a + sg*ma) / uint32(out.A))
							out.B = uint16((db*a + sb*ma) / uint32(out.A))

							// out.G = uint16((dg*da + sg*sa) / (da + sa))
							// out.B = uint16((db*da + sb*sa) / (da + sa))
							// out.A = uint16((da*a + sa*ma) / maxAlpha)

							// midOut := color.NRGBA64Model.Convert(tempout).(color.NRGBA64)
							// out.R, out.G, out.B, out.A = midOut.R, midOut.G, midOut.B, midOut.A

							/*
								var tempout color.RGBA64
								dr, dg, db, da := dst.At(x, y).RGBA()
								a := maxAlpha - (sa * ma / maxAlpha)
								tempout.R = uint16((dr*a + sr*ma) / maxAlpha)
								tempout.G = uint16((dg*a + sg*ma) / maxAlpha)
								tempout.B = uint16((db*a + sb*ma) / maxAlpha)
								tempout.A = uint16((da*a + sa*ma) / maxAlpha)

								midOut := color.NRGBA64Model.Convert(tempout).(color.NRGBA64)
								out.R, out.G, out.B, out.A = midOut.R, midOut.G, midOut.B, midOut.A*/
						}
					}

					// @TODO double check if this is needed
					// or double dipping transformations
					// assign the colour space if there is one
					//		if cspace, ok := src.At(sx, sy).(*CNRGBA64); ok {
					//			out.ColorSpace = cspace.ColorSpace
					//		}
					// The third argument is &out instead of out (and out is
					// declared outside of the inner loop) to avoid the implicit
					// conversion to color.Color here allocating memory in the
					// inner loop if sizeof(color.RGBA64) > sizeof(uintptr).

					dest.Set(x, y, &out)
				}
			}
		}

	default:
		draw.DrawMask(dest, r, src, sp, mask, mp, op)

	}

}

// clip clips r against each image's bounds (after translating into the
// destination image's coordinate space) and shifts the points sp and mp by
// the same amount as the change in r.Min.
// This the same as the draw standard library
func clip(dst draw.Image, r *image.Rectangle, src image.Image, sp *image.Point, mask image.Image, mp *image.Point) {
	orig := r.Min
	*r = r.Intersect(dst.Bounds())
	*r = r.Intersect(src.Bounds().Add(orig.Sub(*sp)))
	if mask != nil {
		*r = r.Intersect(mask.Bounds().Add(orig.Sub(*mp)))
	}
	dx := r.Min.X - orig.X
	dy := r.Min.Y - orig.Y
	if dx == 0 && dy == 0 {
		return
	}
	sp.X += dx
	sp.Y += dy
	if mp != nil {
		mp.X += dx
		mp.Y += dy
	}
}
