package bars

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
)

func bar(target draw.Image, fills [][]bars, heights []float64) {
	// 2d array of shapes
	// heights := []float64{7 * b, 1 * b, 1 * b, 3 * b}
	// treat the widths and heights as floats until the rectangle is formed
	b := target.Bounds().Max
	y := 0.0
	for i, h := range heights {

		w := 0.0
		twidth := 0.0
		for _, f := range fills[i] {
			twidth += f.width * float64(b.X)
			area := image.Rect(int(w), int(y), int(w+f.width*float64(b.X)), int(y+h*float64(b.Y)))
			var fill image.Image

			if f.color != nil {
				fill = &image.Uniform{f.color}
			} else {
				fill = f.fill(image.Rect(0, 0, area.Dx(), area.Dy()))
			}

			colour.Draw(target, area, fill, image.Point{}, draw.Src)

			w += f.width * float64(b.X)
		}

		y += h * float64(b.Y)

	}
}

type bars struct {
	width float64
	color color.Color
	fill  func(image.Rectangle) draw.Image
}

const (
	//widths
	d = 240 / 1920.0
	f = 205 / 1920.0
	c = 206 / 1920.0
	b = 1 / 12.0
	k = 309 / 1920.0
	g = 411 / 1920.0
	h = 171 / 1920.0

	sdwidth = 1 / 7.0
	qwidth  = (5 * sdwidth) / 4.0
)

var (
	gray40   = color.NRGBA64{R: 414 << 6, G: 414 << 6, B: 414 << 6, A: 0xffff}
	white75  = color.NRGBA64{R: 721 << 6, G: 721 << 6, B: 721 << 6, A: 0xffff}
	yellow75 = color.NRGBA64{R: 721 << 6, G: 721 << 6, B: 64 << 6, A: 0xffff}
	cyan75   = color.NRGBA64{R: 64 << 6, G: 721 << 6, B: 721 << 6, A: 0xffff}
	green75  = color.NRGBA64{R: 64 << 6, G: 721 << 6, B: 64 << 6, A: 0xffff}
	mag75    = color.NRGBA64{R: 721 << 6, G: 64 << 6, B: 721 << 6, A: 0xffff}
	red75    = color.NRGBA64{R: 721 << 6, G: 64 << 6, B: 64 << 6, A: 0xffff}
	blue75   = color.NRGBA64{R: 64 << 6, G: 64 << 6, B: 721 << 6, A: 0xffff}

	gray40SD   = color.NRGBA{R: 104, G: 104, B: 104, A: 0xff}
	white75SD  = color.NRGBA{R: 180, G: 180, B: 180, A: 0xff}
	yellow75SD = color.NRGBA{R: 180, G: 180, B: 16, A: 0xff}
	cyan75SD   = color.NRGBA{R: 16, G: 180, B: 180, A: 0xff}
	green75SD  = color.NRGBA{R: 16, G: 180, B: 16, A: 0xff}
	mag75SD    = color.NRGBA{R: 180, G: 16, B: 180, A: 0xff}
	red75SD    = color.NRGBA{R: 180, G: 16, B: 16, A: 0xff}
	blue75SD   = color.NRGBA{R: 16, G: 16, B: 180, A: 0xff}
)

var (
	cyan100   = color.NRGBA64{R: 64 << 6, G: 940 << 6, B: 940 << 6, A: 0xffff}
	white100  = color.NRGBA64{R: 940 << 6, G: 940 << 6, B: 940 << 6, A: 0xffff}
	blue100   = color.NRGBA64{R: 64 << 6, G: 64 << 6, B: 940 << 6, A: 0xffff}
	yellow100 = color.NRGBA64{R: 940 << 6, G: 940 << 6, B: 64 << 6, A: 0xffff}
	black0    = color.NRGBA64{R: 64 << 6, G: 64 << 6, B: 64 << 6, A: 0xffff}
	red100    = color.NRGBA64{R: 940 << 6, G: 64 << 6, B: 64 << 6, A: 0xffff}

	cyan100SD   = color.NRGBA{R: 16, G: 235, B: 235, A: 0xff}
	white100SD  = color.NRGBA{R: 235, G: 235, B: 235, A: 0xff}
	blue100SD   = color.NRGBA{R: 16, G: 16, B: 235, A: 0xff}
	yellow100SD = color.NRGBA{R: 235, G: 235, B: 16, A: 0xff}
	black0SD    = color.NRGBA{R: 16, G: 16, B: 16, A: 0xff}
	red100SD    = color.NRGBA{R: 235, G: 16, B: 16, A: 0xff}
)

var (
	gray15    = color.NRGBA64{R: 195 << 6, G: 195 << 6, B: 195 << 6, A: 0xffff}
	black2Neg = color.NRGBA64{R: 46 << 6, G: 46 << 6, B: 46 << 6, A: 0xffff}
	black2Pos = color.NRGBA64{R: 82 << 6, G: 82 << 6, B: 82 << 6, A: 0xffff}
	black4Pos = color.NRGBA64{R: 99 << 6, G: 99 << 6, B: 99 << 6, A: 0xffff}

	gray15SD    = color.NRGBA{R: 49, G: 49, B: 49, A: 0xff}
	black2NegSD = color.NRGBA{R: 12, G: 12, B: 12, A: 0xff}
	black4NegSD = color.NRGBA{R: 7, G: 7, B: 7, A: 0xff}
	black2PosSD = color.NRGBA{R: 20, G: 20, B: 20, A: 0xff}
	black4PosSD = color.NRGBA{R: 25, G: 25, B: 25, A: 0xff}
)

var (
	ISD = color.NRGBA{R: 16, G: 70, B: 106, A: 0xff}
	QSD = color.NRGBA{R: 72, G: 16, B: 118, A: 0xff}
)

func yRamp(bounds image.Rectangle) draw.Image {
	base := image.NewNRGBA64(bounds)

	start := 65
	end := 939
	step := float64(end-start) / float64(bounds.Dx())

	for x := 0; x <= bounds.Dx(); x++ {

		col := uint16(start + int(float64(x)*step))

		colour.Draw(base, image.Rect(x, bounds.Min.Y, x+1, bounds.Max.Y), &image.Uniform{color.NRGBA64{R: col << 6, G: col << 6, B: col << 6, A: 0xffff}}, image.Point{}, draw.Src)
	}

	return base
}

func yRampSD(bounds image.Rectangle) draw.Image {
	base := image.NewNRGBA(bounds)

	start := 17
	end := 234
	step := float64(end-start) / float64(bounds.Dx())

	for x := 0; x <= bounds.Dx(); x++ {

		col := uint8(start + int(float64(x)*step))

		colour.Draw(base, image.Rect(x, bounds.Min.Y, x+1, bounds.Max.Y), &image.Uniform{color.NRGBA{R: col, G: col, B: col, A: 0xff}}, image.Point{}, draw.Src)
	}

	return base
}

func superBlack(bounds image.Rectangle) draw.Image {
	base := image.NewNRGBA64(bounds)

	colour.Draw(base, base.Bounds(), &image.Uniform{black0}, image.Point{}, draw.Src)

	start := 64
	end := 4
	step := 2 * float64(end-start) / float64(bounds.Dx())
	rolling := float64(start)
	for x := 0; x <= bounds.Dx(); x++ {

		if x < bounds.Dx()/2 {
			rolling += step

		} else {
			rolling -= step
		}
		col := uint16(rolling)
		colour.Draw(base, image.Rect(x, bounds.Dy()/3, x+1, 2*bounds.Dy()/3), &image.Uniform{color.NRGBA64{R: col << 6, G: col << 6, B: col << 6, A: 0xffff}}, image.Point{}, draw.Src)

	}

	return base
}

func superBlackSD(bounds image.Rectangle) draw.Image {
	base := image.NewNRGBA(bounds)

	colour.Draw(base, base.Bounds(), &image.Uniform{black0}, image.Point{}, draw.Src)

	start := 16
	end := 1
	step := 2 * float64(end-start) / float64(bounds.Dx())
	rolling := float64(start)
	for x := 0; x <= bounds.Dx(); x++ {

		if x < bounds.Dx()/2 {
			rolling += step

		} else {
			rolling -= step
		}
		col := uint8(rolling)
		colour.Draw(base, image.Rect(x, bounds.Dy()/3, x+1, 2*bounds.Dy()/3), &image.Uniform{color.NRGBA{R: col, G: col, B: col, A: 0xff}}, image.Point{}, draw.Src)

	}

	return base
}

func superWhite(bounds image.Rectangle) draw.Image {
	base := image.NewNRGBA64(bounds)

	colour.Draw(base, base.Bounds(), &image.Uniform{white100}, image.Point{}, draw.Src)

	start := 940
	end := 1020
	step := 2 * float64(end-start) / float64(bounds.Dx())
	rolling := float64(start)
	for x := 0; x <= bounds.Dx(); x++ {

		if x < bounds.Dx()/2 {
			rolling += step

		} else {
			rolling -= step
		}
		col := uint16(rolling)
		colour.Draw(base, image.Rect(x, bounds.Dy()/3, x+1, 2*bounds.Dy()/3), &image.Uniform{color.NRGBA64{R: col << 6, G: col << 6, B: col << 6, A: 0xffff}}, image.Point{}, draw.Src)

	}

	return base
}

func superWhiteSD(bounds image.Rectangle) draw.Image {
	base := image.NewNRGBA(bounds)

	colour.Draw(base, base.Bounds(), &image.Uniform{white100}, image.Point{}, draw.Src)

	start := 235
	end := 255
	step := 2 * float64(end-start) / float64(bounds.Dx())
	rolling := float64(start)
	for x := 0; x <= bounds.Dx(); x++ {

		if x < bounds.Dx()/2 {
			rolling += step

		} else {
			rolling -= step
		}
		col := uint8(rolling)
		colour.Draw(base, image.Rect(x, bounds.Dy()/3, x+1, 2*bounds.Dy()/3), &image.Uniform{color.NRGBA{R: col, G: col, B: col, A: 0xff}}, image.Point{}, draw.Src)

	}

	return base
}
