package gradients

import (
	"context"
	"image"
	"image/draw"

	"github.com/mmTristan/opentsg-modules/opentsg-core/colour"
	"github.com/mmTristan/opentsg-modules/opentsg-widgets/text"
)

// labels places the label on the stripe based on the angle of the stripe, the text does not change angle
func (txt textObjectJSON) labels(target draw.Image, colourSpace colour.ColorSpace, label, angle string) {

	var canvas draw.Image

	// get the flat row
	switch angle {
	case rotate180, noRotation:
		bounds := target.Bounds()
		bounds.Max.Y = int((txt.TextHeight) * float64(bounds.Max.Y) / 100)
		canvas = colour.NewNRGBA64(colourSpace, bounds)
		// canvas = image.NewNRGBA64(bounds)
	case rotate270, rotate90:
		// canvas = image.NewNRGBA64(image.Rect(0, 0, target.Bounds().Dy(), (con.TextProperties.TextHeight*target.Bounds().Dx())/100))
		canvas = colour.NewNRGBA64(colourSpace, image.Rect(0, 0, target.Bounds().Dy(), int((txt.TextHeight)*float64(target.Bounds().Max.X)/100)))
	}

	mc := context.Background()

	txtBox := text.NewTextboxer(colourSpace,
		text.WithTextColourString(txt.TextColour),
		text.WithXAlignment(txt.TextXPosition),
		text.WithYAlignment(txt.TextYPosition),
		text.WithFont(text.FontPixel),
		text.WithFill(text.FillTypeFull),
	)

	txtBox.DrawString(canvas, &mc, label)

	// gradientBounds := canvas.Bounds().Max

	// var stripeH int
	//stripeH = gradientBounds.Y
	/*
		switch angle {
		case rotate180, noRotation:
			stripeH = gradientBounds.Y
		case rotate270, rotate90:
			stripeH = gradientBounds.X
		}*/

	/*
		lFont := fontGen(t.TextHeight, stripeH)

		// @TODO update so it always draws the same thing then put over the image as a overlay
		// so things are transposed we always draw the rows so they fit
		// 180 is inverse x and y
		// 90 is swap x and y
		// 270 is inverse and swap x and y

		col := colourgen.HexToColour(t.TextColour, colour.ColorSpace{})


		xpos := xPos(b.X, lFont, label, t.TextXPosition)
		ypos := yPos(lFont, t.TextYPosition, b.Y)

		point := fixed.Point26_6{X: fixed.Int26_6(xpos * 64), Y: fixed.Int26_6(ypos * 64)}

		/*
			// Assign the point based on the rotation to ensure the label lines up with the bar
			var point fixed.Point26_6
			switch angle {
			case rotate180, noRotation:
				xpos := xPos(b.X, lFont, label, t.TextXPosition)
				ypos := yPos(lFont, t.TextYPosition, stripeH)
				// else do not change the y
				point = fixed.Point26_6{X: fixed.Int26_6(xpos * 64), Y: fixed.Int26_6(ypos * 64)}
			case rotate270, rotate90:
				xpos := xPos(b.X, lFont, label, t.TextXPosition)
				ypos := yPos(lFont, t.TextYPosition, b.Y)

				point = fixed.Point26_6{X: fixed.Int26_6(xpos * 64), Y: fixed.Int26_6(ypos * 64)}
			}*/
	/*
		d := &font.Drawer{
			Dst:  canvas,
			Src:  image.NewUniform(col),
			Face: lFont,
			Dot:  point,
		}

		d.DrawString(label)*/

	// rotate the text and transpose it on
	// @TODO figure out how to make this more efficent
	b := canvas.Bounds().Max
	var intermediate draw.Image
	intermediate = image.NewNRGBA64(target.Bounds())
	switch angle {
	case rotate90:
		for x := 0; x <= b.X; x++ {
			for y := 0; y <= b.Y; y++ {
				c := canvas.At(x, y)
				intermediate.Set(b.Y-y, x, c)
			}
		}
	case rotate270:
		for x := 0; x <= b.X; x++ {
			for y := 0; y <= b.Y; y++ {
				c := canvas.At(x, y)
				intermediate.Set(y, b.X-x, c)
			}
		}

	case rotate180:

		for x := 0; x <= b.X; x++ {
			for y := 0; y <= b.Y; y++ {
				c := canvas.At(x, y)
				intermediate.Set(b.X-x, b.Y-y, c)
			}
		}

	default:
		intermediate = canvas

	}

	// fmt.Println(intermediate.At(0, 0), txtBox, colourSpace)
	// add the label
	colour.Draw(target, target.Bounds(), intermediate, image.Point{}, draw.Over)
}
