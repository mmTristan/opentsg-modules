// package textbox generates textboxes.
package textbox

import (
	"context"
	"image"
	"image/draw"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"

	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
	texter "github.com/mrmxf/opentsg-modules/opentsg-widgets/text"
)

const (
	WidgetType = "builtin.textbox"
)

// TBGenerate generates text boxes on a given image based on config values
func TBGenerate(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	opts := []any{c}
	conf := widgethandler.GenConf[TextboxJSON]{Debug: debug, Schema: Schema, WidgetType: WidgetType, ExtraOpt: opts}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards
}

func (tb TextboxJSON) Handle(resp tsg.Response, req *tsg.Request) {
	// calculate the border here

	bounds := resp.BaseImage().Bounds().Max
	var width, height float64
	if tb.BorderSize > 0 { // prevent div 0 errors
		width, height = (float64(bounds.X)*tb.BorderSize)/100, (float64(bounds.Y)*tb.BorderSize)/100
	} // else leave as 0

	// draw the borders
	borderwidth := int(height)
	if width < height {
		borderwidth = int(width)
	}

	borderColour := tb.Border.ToColour(tb.ColourSpace)
	colour.Draw(resp.BaseImage(), image.Rect(0, 0, borderwidth, resp.BaseImage().Bounds().Max.Y), &image.Uniform{borderColour}, image.Point{}, draw.Src)
	colour.Draw(resp.BaseImage(), image.Rect(0, 0, resp.BaseImage().Bounds().Max.X, borderwidth), &image.Uniform{borderColour}, image.Point{}, draw.Src)
	colour.Draw(resp.BaseImage(), image.Rect(resp.BaseImage().Bounds().Max.X-borderwidth, 0, resp.BaseImage().Bounds().Max.X, resp.BaseImage().Bounds().Max.Y), &image.Uniform{borderColour}, image.Point{}, draw.Src)
	colour.Draw(resp.BaseImage(), image.Rect(0, resp.BaseImage().Bounds().Max.Y-borderwidth, resp.BaseImage().Bounds().Max.X, resp.BaseImage().Bounds().Max.Y), &image.Uniform{borderColour}, image.Point{}, draw.Src)

	// get the text and background
	c := colour.NewNRGBA64(tb.ColourSpace, image.Rect(0, 0, resp.BaseImage().Bounds().Max.X-borderwidth*2, resp.BaseImage().Bounds().Max.Y-borderwidth*2))

	textbox := texter.NewTextboxer(tb.ColourSpace,

		texter.WithBackgroundColourString(tb.Back),
		texter.WithTextColourString(tb.Textc),
		texter.WithFill(tb.FillType),
		texter.WithXAlignment(tb.XAlignment),
		texter.WithYAlignment(tb.YAlignment),
		texter.WithFont(tb.Font),
	)

	err := textbox.DrawStringsHandler(c, req, tb.Text)
	if err != nil {
		resp.Write(tsg.WidgetError, err.Error())
	}

	// apply the text
	colour.Draw(resp.BaseImage(), image.Rect(borderwidth, borderwidth, resp.BaseImage().Bounds().Max.X-borderwidth, resp.BaseImage().Bounds().Max.Y-borderwidth), c, image.Point{}, draw.Src)

	resp.Write(tsg.WidgetSuccess, "success")
}

func (tb TextboxJSON) Generate(canvas draw.Image, opts ...any) error {
	// calculate the border here

	bounds := canvas.Bounds().Max
	var width, height float64
	if tb.BorderSize > 0 { // prevent div 0 errors
		width, height = (float64(bounds.X)*tb.BorderSize)/100, (float64(bounds.Y)*tb.BorderSize)/100
	} // else leave as 0

	// draw the borders
	borderwidth := int(height)
	if width < height {
		borderwidth = int(width)
	}

	borderColour := tb.Border.ToColour(tb.ColourSpace)
	colour.Draw(canvas, image.Rect(0, 0, borderwidth, canvas.Bounds().Max.Y), &image.Uniform{borderColour}, image.Point{}, draw.Src)
	colour.Draw(canvas, image.Rect(0, 0, canvas.Bounds().Max.X, borderwidth), &image.Uniform{borderColour}, image.Point{}, draw.Src)
	colour.Draw(canvas, image.Rect(canvas.Bounds().Max.X-borderwidth, 0, canvas.Bounds().Max.X, canvas.Bounds().Max.Y), &image.Uniform{borderColour}, image.Point{}, draw.Src)
	colour.Draw(canvas, image.Rect(0, canvas.Bounds().Max.Y-borderwidth, canvas.Bounds().Max.X, canvas.Bounds().Max.Y), &image.Uniform{borderColour}, image.Point{}, draw.Src)

	// get the text and background
	c := colour.NewNRGBA64(tb.ColourSpace, image.Rect(0, 0, canvas.Bounds().Max.X-borderwidth*2, canvas.Bounds().Max.Y-borderwidth*2))

	cb := context.Background()
	textbox := texter.NewTextboxer(tb.ColourSpace,

		texter.WithBackgroundColourString(tb.Back),
		texter.WithTextColourString(tb.Textc),
		texter.WithFill(tb.FillType),
		texter.WithXAlignment(tb.XAlignment),
		texter.WithYAlignment(tb.YAlignment),
		texter.WithFont(tb.Font),
	)

	err := textbox.DrawStrings(c, &cb, tb.Text)
	if err != nil {
		return err
	}

	// apply the text
	colour.Draw(canvas, image.Rect(borderwidth, borderwidth, canvas.Bounds().Max.X-borderwidth, canvas.Bounds().Max.Y-borderwidth), c, image.Point{}, draw.Src)

	return nil
}
