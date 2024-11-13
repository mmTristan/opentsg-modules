package geometrytext

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"strings"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/text"
)

func LabelGenerator(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	opts := []any{c}
	conf := widgethandler.GenConf[Config]{Debug: debug, Schema: Schema, WidgetType: "builtin.geometrytext", ExtraOpt: opts}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

var getGeometry = gridgen.GetGridGeometry

// amend so that the number of colours is based off of the input, can be upgraded to 5 or 6 for performance
func (gt Config) Generate(canvas draw.Image, opt ...any) error {
	var c *context.Context

	if len(opt) != 0 {
		var ok bool
		c, ok = opt[0].(*context.Context)
		if !ok {
			return fmt.Errorf("0DEV configuration error when assigning fourcolour context")
		}
	} else {
		return fmt.Errorf("0DEV configuration error when assigning fourcolour context")
	}

	flats, err := getGeometry(c, gt.GridLoc.Location)
	if err != nil {
		return err
	}
	// fmt.Println(len(flats), gt.GridLoc)
	// This is too intensive as text box does way more than this widget needs

	geomBox := text.NewTextboxer(gt.ColourSpace,
		text.WithFont(text.FontPixel),
		text.WithTextColourString(gt.TextColour),
	)

	// extract colours here and text
	/*
		colour := colourgen.HexToColour(gt.TextColour, gt.ColourSpace)
		fontByte := textbox.FontSelector(c, "pixel")

		fontain, err := freetype.ParseFont(fontByte)
		if err != nil {
			return fmt.Errorf("0101 %v", err)
		}

		d := &font.Drawer{
			Dst: canvas,
			Src: image.NewUniform(colour),
		}
	*/
	///	cont := context.Background()
	for _, f := range flats {

		segment := gridgen.ImageGenerator(*c, image.Rect(0, 0, f.Shape.Dx(), f.Shape.Dy()))
		lines := strings.Split(f.Name, " ")
		geomBox.DrawStrings(segment, c, lines)
		colour.Draw(canvas, f.Shape, segment, image.Point{}, draw.Src)
		// geomBox.DrawStrings(f.Shape, cont, lines)

		/*
			//if i%1000 == 0 {
			//	fmt.Println(i)
			//}
			height := (1.1 / 3.0) * (float64(f.Shape.Dy()))
			width := (1.1 / 3.0) * (float64(f.Shape.Dx()))
			if width < height {
				height = width
			}
			// height /= 2

			opt := truetype.Options{Size: height, SubPixelsY: 8, Hinting: 2}
			myFace := truetype.NewFace(fontain, &opt)

			//	textAreaX := float64(f.Shape.Dx())
			//	textAreaY := float64(f.Shape.Dy())
			//	big := true

			/*	for big {

				thresholdX := float64(labelBox.Max.X.Round() + labelBox.Min.X.Round())
				thresholdY := float64(labelBox.Max.Y.Round() + labelBox.Min.Y.Round())
				// Comparre the text width to the width of the text box
				if (thresholdX > textAreaX) || (thresholdY > textAreaY) {

					height *= 0.9
					opt = truetype.Options{Size: height, SubPixelsY: 8, Hinting: 2}
					myFace = truetype.NewFace(fontain, &opt)
					labelBox, _ = font.BoundString(myFace, label)

				} else {
					big = false
				}
			}

			labelBox, _ := font.BoundString(myFace, lines[0])
			xOff := xPos(f.Shape, labelBox)

			for i, line := range lines {
				labelBox, _ := font.BoundString(myFace, line)
				yOff := yPos(f.Shape, labelBox, float64(len(lines)), i)

				//	fmt.Println(xOff, yOff)
				point := fixed.Point26_6{X: fixed.Int26_6(xOff * 64), Y: fixed.Int26_6(yOff * 64)}

				//	myFace := truetype.NewFace(fontain, &opt)
				d.Face = myFace
				d.Dot = point
				d.DrawString(line)
			}
		*/
	}

	return nil
}

/*

func xPos(canvas image.Rectangle, rect fixed.Rectangle26_6) int {
	textWidth := rect.Max.X.Round() - rect.Min.X.Round()

	return canvas.Min.X + (((canvas.Bounds().Dx()) - textWidth) / 2)

}

func yPos(canvas image.Rectangle, rect fixed.Rectangle26_6, lines float64, count int) int {
	yOffset := (rect.Max.Y.Round()) - (rect.Min.Y.Round())
	mid := (float64(canvas.Bounds().Dy()) + float64(yOffset)*0.8) / (2.0 * lines)

	return (canvas.Bounds().Min.Y + int(mid)) + int(canvas.Bounds().Dy()*count)/int(lines)

}*/
