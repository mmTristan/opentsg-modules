// Package framecount adds a framecounter to a user specified location
package framecount

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/gridgen"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
	"github.com/mrmxf/opentsg-modules/opentsg-widgets/text"
)

const (
	WidgetType = "builtin.frameCounter"
)

func CountGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	opts := []any{c}
	conf := widgethandler.GenConf[Config]{Debug: debug, Schema: Schema, WidgetType: WidgetType, ExtraOpt: opts}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

// Handler has some updates over previous versions
// it now runs regardless of FrameCounter, if you are calling it you want a frame counter
func (c Config) Handle(resp tsg.Response, req *tsg.Request) {
	b := resp.BaseImage().Bounds().Max

	if c.Font == "" {
		c.Font = text.FontPixel
	}

	// stop errors happening when font is not declared
	if c.FontSize == 0 {
		c.FontSize = 100
	}

	// Size of the text in pixels to font
	c.FontSize = (float64(b.Y) * (c.FontSize / 100)) // keep as pixels

	if b.Y > b.X {
		c.FontSize *= (float64(b.X) / float64(b.Y)) // Scale the font size for narrow grids
	}

	if c.FontSize < 7 {
		resp.Write(tsg.WidgetError, fmt.Sprintf("0DDEV The font size %v pixels is smaller thant the minimum value of 7 pixels", c.FontSize))
		return
	}

	square := image.Point{int(c.FontSize), int(c.FontSize)}

	frame := req.GenerateSubImage(resp.BaseImage(), image.Rect(0, 0, square.X, square.Y))

	defaultBackground := colour.CNRGBA64{R: uint16(195) << 8, G: uint16(195) << 8, B: uint16(195) << 8, A: uint16(195) << 8, ColorSpace: c.ColourSpace}
	defaulText := colour.CNRGBA64{A: 65535, ColorSpace: c.ColourSpace}

	txtBox := text.NewTextboxer(c.ColourSpace,
		text.WithFill(text.FillTypeFull),
		text.WithFont(c.Font),
		text.WithBackgroundColour(&defaultBackground),
		text.WithTextColour(&defaulText),
	)

	// update the colours if required
	if c.BackColour != "" {
		text.WithBackgroundColourString(c.BackColour)(txtBox)
	}

	if c.TextColour != "" {
		text.WithTextColourString(c.TextColour)(txtBox)
	}
	// MyFont.Advance
	mes, err := intTo4(req.FrameProperties.FrameNumber)
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	err = txtBox.DrawStringsHandler(frame, req, []string{mes})
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
		return
	}

	fb := frame.Bounds().Max
	// If pos not given then draw it here

	var x, y int
	switch imgpos := c.Imgpos.(type) {
	case map[string]interface{}:
		x, y = userPos(imgpos, b, fb)
	default:
		x, y = 0, 0

	}

	if x > (b.X - fb.X) {
		resp.Write(tsg.WidgetError, fmt.Sprintf("_0153 the x position %v is greater than the x boundary of %v with frame width of %v", x, resp.BaseImage().Bounds().Max.X, fb.X))
		return
	} else if y > b.Y-fb.Y {
		resp.Write(tsg.WidgetError, fmt.Sprintf("_0153 the y position %v is greater than the y boundary of %v with frame height of %v", y, resp.BaseImage().Bounds().Max.Y, fb.Y))
		return
	}

	// Corner := image.Point{-1 * (canvas.Bounds().Max.X - height - 1), -1 * (canvas.Bounds().Max.Y - height - 1)}
	colour.Draw(resp.BaseImage(), image.Rect(x, y, x+int(c.FontSize), y+int(c.FontSize)), frame, image.Point{}, draw.Over)

	resp.Write(tsg.WidgetSuccess, "success")

}

var pos = framePos

func (f Config) Helper(key string, c *context.Context) {
	// Update the frame number add amend context with it
	f.FrameNumber = pos()
	fc := make(map[string]Config)
	fc[key] = f

	// Widgethandler.Put(fc, c)

}

func (f Config) Generate(canvas draw.Image, extraOpts ...any) error {

	b := canvas.Bounds().Max
	if !f.getFrames() {
		return fmt.Errorf("0DEV frame counter not enabled for this frame. Ensure frameCounter is set to true")
	}
	if f.Font == "" {
		f.Font = text.FontPixel
	}

	if len(extraOpts) != 1 {
		return fmt.Errorf("0154 framecount configuration error")
	}

	c, ok := extraOpts[0].(*context.Context)
	if !ok {
		return fmt.Errorf("0155 configuration error when assiging framecount context")
	}

	// stop errors happening when font is not declared
	if f.FontSize == 0 {
		f.FontSize = 100
	}

	// Size of the text in pixels to font
	f.FontSize = (float64(b.Y) * (f.FontSize / 100)) // keep as pixels

	if b.Y > b.X {
		f.FontSize *= (float64(b.X) / float64(b.Y)) // Scale the font size for narrow grids
	}

	if f.FontSize < 7 {
		return fmt.Errorf("0DDEV The font size %v pixels is smaller thant the minimum value of 7 pixels", f.FontSize)
	}

	square := image.Point{int(f.FontSize), int(f.FontSize)}

	frame := gridgen.ImageGenerator(*c, image.Rect(0, 0, square.X, square.Y))

	defaultBackground := colour.CNRGBA64{R: uint16(195) << 8, G: uint16(195) << 8, B: uint16(195) << 8, A: uint16(195) << 8, ColorSpace: f.ColourSpace}
	defaulText := colour.CNRGBA64{A: 65535, ColorSpace: f.ColourSpace}

	txtBox := text.NewTextboxer(f.ColourSpace,
		text.WithFill(text.FillTypeFull),
		text.WithFont(f.Font),
		text.WithBackgroundColour(&defaultBackground),
		text.WithTextColour(&defaulText),
	)

	// update the colours if required
	if f.BackColour != "" {
		text.WithBackgroundColourString(f.BackColour)(txtBox)
	}

	if f.TextColour != "" {
		text.WithTextColourString(f.TextColour)(txtBox)
	}
	// MyFont.Advance
	mes, err := intTo4(pos())
	if err != nil {
		return err
	}

	err = txtBox.DrawString(frame, c, mes)
	if err != nil {
		return err
	}

	fb := frame.Bounds().Max
	// If pos not given then draw it here

	var x, y int
	switch imgpos := f.Imgpos.(type) {
	case map[string]interface{}:
		x, y = userPos(imgpos, b, fb)
	default:
		x, y = 0, 0

	}

	if x > (b.X - fb.X) {
		return fmt.Errorf("_0153 the x position %v is greater than the x boundary of %v with frame width of %v", x, canvas.Bounds().Max.X, fb.X)
	} else if y > b.Y-fb.Y {
		return fmt.Errorf("_0153 the y position %v is greater than the y boundary of %v with frame height of %v", y, canvas.Bounds().Max.Y, fb.Y)
	}

	// Corner := image.Point{-1 * (canvas.Bounds().Max.X - height - 1), -1 * (canvas.Bounds().Max.Y - height - 1)}
	colour.Draw(canvas, image.Rect(x, y, x+int(f.FontSize), y+int(f.FontSize)), frame, image.Point{}, draw.Over)

	return nil
}

func intTo4(num int) (string, error) {
	s := strconv.Itoa(num)
	if len(s) > 4 {
		return "", fmt.Errorf("frame Count greater then 9999")
	}

	buf0 := strings.Repeat("0", 4-len(s))

	s = buf0 + s

	return s, nil
}

const (
	bottomLeft  = "bottom left"
	bottomRight = "bottom right"
	topRight    = "top right"
	topLeft     = "top left"
)

func userPos(location map[string]interface{}, canSize, frameSize image.Point) (int, int) {
	if location["alias"] != nil {
		// Process as simple location
		// The minus one is inluded to compensate for canvas startnig at 0
		switch location["alias"].(string) {
		case bottomLeft:
			return 0, canSize.Y - frameSize.Y - 1
		case bottomRight:
			return canSize.X - frameSize.X - 1, canSize.Y - frameSize.Y - 1
		case topRight:
			return canSize.X - frameSize.X - 1, 0
		default:
			return 0, 0
		}
	} else {
		var x, y int
		if mid := location["x"]; mid != nil { // Make a percentage of the canvas
			var percent float64
			switch val := mid.(type) {
			case float64:
				percent = val
			case int:
				percent = float64(val)
			}
			x = int(math.Floor(percent * (float64(canSize.X) / 100)))
		}
		if mid := location["y"]; mid != nil {
			var percent float64
			switch val := mid.(type) {
			case float64:
				percent = val
			case int:
				percent = float64(val)
			}
			y = int(math.Floor(percent * (float64(canSize.Y) / 100)))
		}

		return x, y
	}
}
