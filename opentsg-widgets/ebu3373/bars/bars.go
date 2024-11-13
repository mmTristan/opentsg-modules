// Package bars generates the ebu colour bars
package bars

import (
	"context"
	"image"
	"image/draw"
	"math"
	"sync"

	"github.com/mrmxf/opentsg-modules/opentsg-core/colour"
	errhandle "github.com/mrmxf/opentsg-modules/opentsg-core/errHandle"
	"github.com/mrmxf/opentsg-modules/opentsg-core/tsg"
	"github.com/mrmxf/opentsg-modules/opentsg-core/widgethandler"
)

const (
	WidgetType = "builtin.ebu3373/bars"
)

func BarGen(canvasChan chan draw.Image, debug bool, c *context.Context, wg, wgc *sync.WaitGroup, logs *errhandle.Logger) {
	defer wg.Done()
	conf := widgethandler.GenConf[BarJSON]{Debug: debug, Schema: Schema, WidgetType: WidgetType}
	widgethandler.WidgetRunner(canvasChan, conf, c, logs, wgc) // Update this to pass an error which is then formatted afterwards

}

// these colours should all be relative to BT 2100 (rec 2020)
// they were given as 10 bit but moved to 16 bit for running with go
var (
	white   = colour.CNRGBA64{R: 60160, G: 60160, B: 60160, A: 0xffff}
	yellow  = colour.CNRGBA64{R: 60160, G: 60160, B: 4096, A: 0xffff}
	cyan    = colour.CNRGBA64{R: 4096, G: 60160, B: 60160, A: 0xffff}
	green   = colour.CNRGBA64{R: 4096, G: 60160, B: 4096, A: 0xffff}
	magenta = colour.CNRGBA64{R: 60160, G: 4096, B: 60160, A: 0xffff}
	red     = colour.CNRGBA64{R: 60160, G: 4096, B: 4096, A: 0xffff}
	blue    = colour.CNRGBA64{R: 4096, G: 4096, B: 60160, A: 0xffff}
	grey    = colour.CNRGBA64{R: 26496, G: 26496, B: 26496, A: 0xffff}
)

var (
	white40   = colour.CNRGBA64{R: 46144, G: 46144, B: 46144, A: 0xffff}
	yellow40  = colour.CNRGBA64{R: 46144, G: 46144, B: 4096, A: 0xffff}
	cyan40    = colour.CNRGBA64{R: 4096, G: 46144, B: 46144, A: 0xffff}
	green40   = colour.CNRGBA64{R: 4096, G: 46144, B: 4096, A: 0xffff}
	magenta40 = colour.CNRGBA64{R: 46144, G: 4096, B: 46144, A: 0xffff}
	red40     = colour.CNRGBA64{R: 46144, G: 4096, B: 4096, A: 0xffff}
	blue40    = colour.CNRGBA64{R: 4096, G: 4096, B: 46144, A: 0xffff}
)
var (
	dLWhite   = colour.CNRGBA64{R: 38528, G: 38528, B: 38528, A: 0xffff}
	dLYellow  = colour.CNRGBA64{R: 594 << 6, G: 601 << 6, B: 246 << 6, A: 0xffff}
	dLCyan    = colour.CNRGBA64{R: 408 << 6, G: 591 << 6, B: 601 << 6, A: 0xffff}
	dLGreen   = colour.CNRGBA64{R: 388 << 6, G: 589 << 6, B: 232 << 6, A: 0xffff}
	dLMagenta = colour.CNRGBA64{R: 534 << 6, G: 227 << 6, B: 595 << 6, A: 0xffff}
	dLRed     = colour.CNRGBA64{R: 522 << 6, G: 216 << 6, B: 138 << 6, A: 0xffff}
	dLBlue    = colour.CNRGBA64{R: 187 << 6, G: 127 << 6, B: 602 << 6, A: 0xffff}
)

var (
	sLWhite   = colour.CNRGBA64{R: 39552, G: 39552, B: 39552, A: 0xffff}
	sLYellow  = colour.CNRGBA64{R: 610 << 6, G: 616 << 6, B: 253 << 6, A: 0xffff}
	sLCyan    = colour.CNRGBA64{R: 422 << 6, G: 605 << 6, B: 615 << 6, A: 0xffff}
	sLGreen   = colour.CNRGBA64{R: 400 << 6, G: 603 << 6, B: 238 << 6, A: 0xffff}
	sLMagenta = colour.CNRGBA64{R: 541 << 6, G: 230 << 6, B: 601 << 6, A: 0xffff}
	sLRed     = colour.CNRGBA64{R: 527 << 6, G: 218 << 6, B: 139 << 6, A: 0xffff}
	sLBlue    = colour.CNRGBA64{R: 186 << 6, G: 126 << 6, B: 598 << 6, A: 0xffff}
)

func (bar BarJSON) Generate(canvas draw.Image, opt ...any) error {
	b := canvas.Bounds().Max

	if bar.ColourSpace == nil {
		bar.ColourSpace = &colour.ColorSpace{}
	}

	colour.Draw(canvas, canvas.Bounds(), &image.Uniform{&grey}, image.Point{}, draw.Src)
	wScale := (float64(b.X) / 3840.0)
	barWidth := wScale * 412

	heights := []float64{200, 560, 200, 200}
	cbars := [][]colour.CNRGBA64{
		{white, yellow, cyan, green, magenta, red, blue},
		{white40, yellow40, cyan40, green40, magenta40, red40, blue40},
		{dLWhite, dLYellow, dLCyan, dLGreen, dLMagenta, dLRed, dLBlue},
		{sLWhite, sLYellow, sLCyan, sLGreen, sLMagenta, sLRed, sLBlue},
	}
	hOff := 0.0
	hScale := (float64(b.Y) / 1160.0)
	for i, h := range heights {

		barHeight := hScale * h

		boxHeight := hOff + barHeight

		// accounting for rounding errors at the end of the bars
		if math.Abs(float64(b.Y)-boxHeight) < 0.002 {
			boxHeight = float64(b.Y)
		}

		off := wScale * 480
		for _, c := range cbars[i] {
			area := image.Rect(int(off), int(hOff), int(off+barWidth), int(boxHeight))

			fill := c
			fill.UpdateColorSpace(*bar.ColourSpace)
			colour.Draw(canvas, area, &image.Uniform{&fill}, image.Point{}, draw.Over)
			off += barWidth
		}

		hOff = boxHeight
	}

	return nil
}

func (bar BarJSON) Handle(resp tsg.Response, _ *tsg.Request) {
	b := resp.BaseImage().Bounds().Max

	if bar.ColourSpace == nil {
		bar.ColourSpace = &colour.ColorSpace{}
	}

	colour.Draw(resp.BaseImage(), resp.BaseImage().Bounds(), &image.Uniform{&grey}, image.Point{}, draw.Src)
	wScale := (float64(b.X) / 3840.0)
	barWidth := wScale * 412

	heights := []float64{200, 560, 200, 200}
	cbars := [][]colour.CNRGBA64{
		{white, yellow, cyan, green, magenta, red, blue},
		{white40, yellow40, cyan40, green40, magenta40, red40, blue40},
		{dLWhite, dLYellow, dLCyan, dLGreen, dLMagenta, dLRed, dLBlue},
		{sLWhite, sLYellow, sLCyan, sLGreen, sLMagenta, sLRed, sLBlue},
	}
	hOff := 0.0
	hScale := (float64(b.Y) / 1160.0)
	for i, h := range heights {

		barHeight := hScale * h

		boxHeight := hOff + barHeight

		// accounting for rounding errors at the end of the bars
		if math.Abs(float64(b.Y)-boxHeight) < 0.002 {
			boxHeight = float64(b.Y)
		}

		off := wScale * 480
		for _, c := range cbars[i] {
			area := image.Rect(int(off), int(hOff), int(off+barWidth), int(boxHeight))

			fill := c
			fill.UpdateColorSpace(*bar.ColourSpace)
			colour.Draw(resp.BaseImage(), area, &image.Uniform{&fill}, image.Point{}, draw.Over)
			off += barWidth
		}

		hOff = boxHeight
	}

	resp.Write(tsg.WidgetSuccess, "success")
}
