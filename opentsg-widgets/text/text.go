// package text is the inbuilt OpenTSG text handler.
// All text should be handled through this module
package text

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"
	"strings"

	_ "embed"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/core"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	FillTypeFull    = "full"
	FillTypeRelaxed = "relaxed"

	AlignmentLeft   = "left"
	AlignmentRight  = "right"
	AlignmentMiddle = "middle"
	AlignmentTop    = "top"
	AlignmentBottom = "bottom"

	FontTitle  = "title"
	FontBody   = "body"
	FontPixel  = "pixel"
	FontHeader = "header"
)

//go:embed MavenPro-Bold.ttf
var Title []byte

//go:embed MavenPro-Regular.ttf
var Header []byte

//go:embed Marvel-Regular.ttf
var Body []byte

//go:embed PixeloidSans.ttf
var Pixel []byte

// DrawString draws a single string in a textbox
func (t TextboxProperties) DrawString(canvas draw.Image, tsgContext *context.Context, label string) error {
	return t.DrawStrings(canvas, tsgContext, []string{label})
}

// drawstring draws multiple lines of text in a textbox
func (t TextboxProperties) DrawStringsHandler(canvas draw.Image, req *tsg.Request, labels []string) error {

	// check somethings been assigned first
	if t.backgroundColour != nil {
		// draw the background first
		if t.backgroundColour.A != 0 {
			colour.Draw(canvas, canvas.Bounds(), &image.Uniform{t.backgroundColour}, image.Point{}, draw.Over)
		}
	}

	// only do the text calculations if there's any
	// text colour
	// or any text
	if t.textColour != nil && len(labels) > 0 {
		if t.textColour.A != 0 {
			fontByte := fontSelectorHandler(req, t.font)

			fontain, err := freetype.ParseFont(fontByte)
			if err != nil {
				return fmt.Errorf("0101 %v", err)
			}

			lines := len(labels)
			fontBounds := fontain.Bounds(fixed.Int26_6(1 * 64))

			var label string
			for _, labl := range labels {
				if len(labl) > len(label) {
					label = labl
				}
			}

			// scale the text to which ever dimension is smaller
			// (y / lines) / y scale
			scale := 64 * (float64(canvas.Bounds().Max.Y) / float64(lines)) / (float64(fontBounds.Max.Y - fontBounds.Min.Y))
			// x to y ratio * x / (x scale * number of letters)
			width := 2 * 64 * (float64(canvas.Bounds().Max.X)) / (float64(fontBounds.Max.X-fontBounds.Min.X) * float64(len(label)))

			if t.verticalText {
				// scale in the opposite directions
				scale = 2 * 64 * (float64(canvas.Bounds().Max.Y)) / (float64(fontBounds.Max.Y-fontBounds.Min.Y) * float64(len(label)))
				width = 64 * (float64(canvas.Bounds().Max.Y) / float64(lines)) / (float64(fontBounds.Max.Y - fontBounds.Min.Y))
			}

			// @TODO flip directions for vertical text

			if width < scale {
				scale = width
			}

			opt := truetype.Options{Size: scale, SubPixelsY: 8, Hinting: 2}
			myFace := truetype.NewFace(fontain, &opt)

			bounds := canvas.Bounds().Max

			if t.verticalText {
				bounds.X /= lines
			} else {
				bounds.Y /= lines
			}

			switch t.fillType {
			case FillTypeFull:
				myFace, _ = fullFill(bounds, myFace, fontain, t.verticalText, scale, label)
			default:
				myFace, _ = relaxedFill(bounds, myFace, fontain, t.verticalText, scale, label)
			}

			// fix point for things like framecount or is this unlikely to matter?
			for i, label := range labels {

				lab := []string{label}
				// this is the y height
				var verticalBox fixed.Rectangle26_6
				// set up the vertical text
				if t.verticalText {
					// split per letter
					lab = strings.Split(label, "")

					verticalBox, _ = getBoundBox(myFace, t.verticalText, label)
				}
				var prevYOff fixed.Int26_6

				for _, l := range lab {
					labelBox, _ := font.BoundString(myFace, l)
					var xOff, yOff int
					if !t.verticalText {
						xOff = xPos(canvas, labelBox, t.xAlignment, 1, 0)
						yOff = yPos(canvas, labelBox, t.yAlignment, float64(lines), i)
					} else {

						// replace the space with something that has dimensions
						if l == " " {
							labelBox, _ = font.BoundString(myFace, "<")
						}
						/*
							go through every letter
							using the minimum to set the y position
							then take away the height of the previous letter to ensure the text moves
						*/
						xOff = xPos(canvas, labelBox, t.xAlignment, float64(lines), i)
						yOff = yPos(canvas, verticalBox, t.yAlignment, 1, 0) + int(prevYOff.Round()) - labelBox.Min.Y.Round()
						prevYOff += (labelBox.Max.Y - labelBox.Min.Y)

					}

					point := fixed.Point26_6{X: fixed.Int26_6(xOff * 64), Y: fixed.Int26_6(yOff * 64)}

					//	myFace := truetype.NewFace(fontain, &opt)
					d := &font.Drawer{
						Dst:  canvas,
						Src:  image.NewUniform(t.textColour),
						Face: myFace,
						Dot:  point,
					}
					d.DrawString(l)
				}
			}
		}
	}

	return nil
}

// drawstring draws multiple lines of text in a textbox
func (t TextboxProperties) DrawStrings(canvas draw.Image, tsgContext *context.Context, labels []string) error {

	// check somethings been assigned first
	if t.backgroundColour != nil {
		// draw the background first
		if t.backgroundColour.A != 0 {
			colour.Draw(canvas, canvas.Bounds(), &image.Uniform{t.backgroundColour}, image.Point{}, draw.Over)
		}
	}

	// only do the text calculations if there's any
	// text colour
	// or any text
	if t.textColour != nil && len(labels) > 0 {
		if t.textColour.A != 0 {
			fontByte := fontSelector(tsgContext, t.font)

			fontain, err := freetype.ParseFont(fontByte)
			if err != nil {
				return fmt.Errorf("0101 %v", err)
			}

			lines := len(labels)
			fontBounds := fontain.Bounds(fixed.Int26_6(1 * 64))

			var label string
			for _, labl := range labels {
				if len(labl) > len(label) {
					label = labl
				}
			}

			// scale the text to which ever dimension is smaller
			// (y / lines) / y scale
			scale := 64 * (float64(canvas.Bounds().Max.Y) / float64(lines)) / (float64(fontBounds.Max.Y - fontBounds.Min.Y))
			// x to y ratio * x / (x scale * number of letters)
			width := 2 * 64 * (float64(canvas.Bounds().Max.X)) / (float64(fontBounds.Max.X-fontBounds.Min.X) * float64(len(label)))

			if t.verticalText {
				// scale in the opposite directions
				scale = 2 * 64 * (float64(canvas.Bounds().Max.Y)) / (float64(fontBounds.Max.Y-fontBounds.Min.Y) * float64(len(label)))
				width = 64 * (float64(canvas.Bounds().Max.Y) / float64(lines)) / (float64(fontBounds.Max.Y - fontBounds.Min.Y))
			}

			// @TODO flip directions for vertical text

			if width < scale {
				scale = width
			}

			opt := truetype.Options{Size: scale, SubPixelsY: 8, Hinting: 2}
			myFace := truetype.NewFace(fontain, &opt)

			bounds := canvas.Bounds().Max

			if t.verticalText {
				bounds.X /= lines
			} else {
				bounds.Y /= lines
			}

			switch t.fillType {
			case FillTypeFull:
				myFace, _ = fullFill(bounds, myFace, fontain, t.verticalText, scale, label)
			default:
				myFace, _ = relaxedFill(bounds, myFace, fontain, t.verticalText, scale, label)
			}

			// fix point for things like framecount or is this unlikely to matter?
			for i, label := range labels {

				lab := []string{label}
				// this is the y height
				var verticalBox fixed.Rectangle26_6
				// set up the vertical text
				if t.verticalText {
					// split per letter
					lab = strings.Split(label, "")

					verticalBox, _ = getBoundBox(myFace, t.verticalText, label)
				}
				var prevYOff fixed.Int26_6

				for _, l := range lab {
					labelBox, _ := font.BoundString(myFace, l)
					var xOff, yOff int
					if !t.verticalText {
						xOff = xPos(canvas, labelBox, t.xAlignment, 1, 0)
						yOff = yPos(canvas, labelBox, t.yAlignment, float64(lines), i)
					} else {

						// replace the space with something that has dimensions
						if l == " " {
							labelBox, _ = font.BoundString(myFace, "<")
						}
						/*
							go through every letter
							using the minimum to set the y position
							then take away the height of the previous letter to ensure the text moves
						*/
						xOff = xPos(canvas, labelBox, t.xAlignment, float64(lines), i)
						yOff = yPos(canvas, verticalBox, t.yAlignment, 1, 0) + int(prevYOff.Round()) - labelBox.Min.Y.Round()
						prevYOff += (labelBox.Max.Y - labelBox.Min.Y)

					}

					point := fixed.Point26_6{X: fixed.Int26_6(xOff * 64), Y: fixed.Int26_6(yOff * 64)}

					//	myFace := truetype.NewFace(fontain, &opt)
					d := &font.Drawer{
						Dst:  canvas,
						Src:  image.NewUniform(t.textColour),
						Face: myFace,
						Dot:  point,
					}
					d.DrawString(l)
				}
			}
		}
	}

	return nil
}

func fullFill(area image.Point, sizeFont font.Face, fontain *truetype.Font, verticalText bool, height float64, label string) (font.Face, fixed.Rectangle26_6) {
	// labelBox, adv := font.BoundString(sizeFont, label)
	labelBox, adv := getBoundBox(sizeFont, verticalText, label)
	textAreaX := float64(area.X)
	textAreaY := float64(area.Y) // Both side

	big := true
	prevFont := sizeFont
	prevBox := labelBox
	// change the font when the initial bit is already too big
	if adv.Round() > int(textAreaX) || math.Abs(float64(labelBox.Max.Y.Round()-labelBox.Min.Y.Round())) > float64(textAreaY) {
		return relaxedFill(area, sizeFont, fontain, verticalText, height, label)
	}

	// scale the text down to fix the box
	for big {
		// the base is always 0
		thresholdX := float64(labelBox.Max.X.Round()) //+ labelBox.Min.X.Round())
		thresholdY := math.Abs(float64(labelBox.Max.Y.Round())) + math.Abs(float64(labelBox.Min.Y.Round()))
		// fmt.Println(thresholdX, thresholdY, labelBox, label, height, textAreaX, textAreaY)
		// Compare the text width to the width of the text box
		if (thresholdX < textAreaX) && (thresholdY < textAreaY) {

			height *= 1.1
			opt := truetype.Options{Size: height, SubPixelsY: 8, Hinting: 2}
			prevFont = sizeFont
			prevBox = labelBox

			sizeFont = truetype.NewFace(fontain, &opt)
			// var adv fixed.Int26_6
			labelBox, _ = getBoundBox(sizeFont, verticalText, label)
			// fmt.Println(adv.Round())

		} else {
			big = false
		}
	}

	return prevFont, prevBox
}
func getBoundBox(face font.Face, verticalText bool, label string) (fixed.Rectangle26_6, fixed.Int26_6) {
	if verticalText {
		lab := strings.Split(label, "")
		var maxWidth fixed.Int26_6
		var totalHeight fixed.Int26_6
		for _, l := range lab {
			if l == " " {
				l = "<"
			}
			b, _ := font.BoundString(face, l)

			totalHeight += b.Max.Y - b.Min.Y
			width := b.Max.X - b.Min.X
			if width > maxWidth {
				maxWidth = width
			}
		}

		/*
			get the count of the spaces
		*/
		return fixed.Rectangle26_6{Max: fixed.Point26_6{X: maxWidth, Y: totalHeight}}, maxWidth + 1
	}

	labelBox, adv := font.BoundString(face, label)
	return labelBox, adv

}

func relaxedFill(area image.Point, sizeFont font.Face, fontain *truetype.Font, verticalText bool, height float64, label string) (font.Face, fixed.Rectangle26_6) {
	labelBox, _ := getBoundBox(sizeFont, verticalText, label)
	textAreaX := float64(area.X)
	textAreaY := float64(area.Y) // Both side

	big := true

	// scale the text down to fix the box
	for big {

		thresholdX := float64(labelBox.Max.X.Round()) // + labelBox.Min.X.Round())
		thresholdY := float64(labelBox.Max.Y.Round() - labelBox.Min.Y.Round())
		// fmt.Println(thresholdX, thresholdY, labelBox, label, height)
		// Compare the text width to the width of the text box
		if (thresholdX > textAreaX) || (thresholdY > textAreaY) {

			height *= 0.9
			opt := truetype.Options{Size: height, SubPixelsY: 8, Hinting: 2}
			sizeFont = truetype.NewFace(fontain, &opt)
			labelBox, _ = getBoundBox(sizeFont, verticalText, label)

		} else {
			big = false
		}
	}

	return sizeFont, labelBox
}

// place in the middle
func xPos(canvas image.Image, rect fixed.Rectangle26_6, position string, lines float64, count int) int {
	textWidth := rect.Max.X.Round() - rect.Min.X.Round()
	// textWidth := rect.Max.X.Ceil() - rect.Min.X.Ceil()
	// account for the minimum is where the text is started to be drawn

	switch position {
	case AlignmentLeft:
		return (canvas.Bounds().Max.X*count)/int(lines) - rect.Min.X.Round()
	case AlignmentRight:
		// get the start point, then account for the
		// start postion of the text box
		return (canvas.Bounds().Max.X*(count+1))/int(lines) - textWidth - rect.Min.X.Round()
	default:
		barWidth := (canvas.Bounds().Max.X) / int(lines)
		textWidther := math.Abs(float64(rect.Max.X.Round() - rect.Min.X.Round()))
		halfGap := (barWidth - int(textWidther)) / 2

		return ((canvas.Bounds().Max.X * (count + 1)) / int(lines)) - textWidth - halfGap - rect.Min.X.Round()
	}

}

// ypos calculates the yposition for the text
func yPos(canvas image.Image, rect fixed.Rectangle26_6, position string, lines float64, count int) int {

	// mid := (float64(canvas.Bounds().Max.Y) + float64(yOffset)) / (2.0 * lines)
	//	fmt.Println(rect, rect.Min, int(mid)+(canvas.Bounds().Max.Y*count)/int(lines))
	switch position {
	case AlignmentBottom:
		// fmt.Println((canvas.Bounds().Max.Y*count)/int(lines) - yOffset)
		return (canvas.Bounds().Max.Y*(count+1))/int(lines) - rect.Max.Y.Round()
	case AlignmentTop:
		return (canvas.Bounds().Max.Y*count)/int(lines) - rect.Min.Y.Round()
	default:
		// total length is  rect.Max.Y.Round() - rect.Max.Y.Round()
		barHeight := (canvas.Bounds().Max.Y) / int(lines)
		textHeight := math.Abs(float64(rect.Max.Y.Round() - rect.Min.Y.Round()))
		halfGap := (barHeight - int(textHeight)) / 2

		// return int((float64(canvas.Bounds().Max.Y)*float64(count)+0.5)/lines) - rect.Max.Y.Round()
		return (canvas.Bounds().Max.Y*(count+1))/int(lines) - rect.Max.Y.Round() - halfGap
	}

}

// font selector enumerates through the different sources of http,
// local files,
// then predetermined embedded fonts and returns the font based on the input string.
func fontSelector(c *context.Context, fontLocation string) []byte {

	font, err := core.GetWebBytes(c, fontLocation)

	if err == nil {
		return font
	}

	font, err = os.ReadFile(fontLocation)
	if err == nil {
		return font
	}

	switch fontLocation {
	case "title":
		return Title
	case "body":
		return Body
	case "pixel":
		return Pixel
	default:
		return Header
	}
}

// font selector enumerates through the different sources of http,
// local files,
// then predetermined embedded fonts and returns the font based on the input string.
func fontSelectorHandler(req *tsg.Request, fontLocation string) []byte {

	font, err := req.SearchWithCredentials(req.Context, fontLocation)

	if err == nil {
		return font
	}

	font, err = os.ReadFile(fontLocation)
	if err == nil {
		return font
	}

	switch fontLocation {
	case "title":
		return Title
	case "body":
		return Body
	case "pixel":
		return Pixel
	default:
		return Header
	}
}
